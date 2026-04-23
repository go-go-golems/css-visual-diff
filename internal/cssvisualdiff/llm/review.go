package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/modes"
	geppettoengine "github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/geppetto/pkg/turns"
)

const defaultSystemPrompt = "You are an expert frontend engineer and visual QA analyst. Use the provided screenshots and structured CSS/layout evidence to answer the user's question. Be concrete about visual changes, likely CSS causes, and user-facing impact."

type ReviewOptions struct {
	Question       string
	Evidence       modes.CompareResult
	MaxProperties  int
	MaxWinnerDiffs int
	SystemPrompt   string
}

type ReviewResult struct {
	Question        string                          `json:"question"`
	Answer          string                          `json:"answer"`
	Model           string                          `json:"model,omitempty"`
	Provider        string                          `json:"provider,omitempty"`
	APIType         string                          `json:"apiType,omitempty"`
	Profile         string                          `json:"profile,omitempty"`
	Registry        string                          `json:"registry,omitempty"`
	PromptSummary   string                          `json:"promptSummary,omitempty"`
	Artifacts       map[string]string               `json:"artifacts,omitempty"`
	InferenceResult *geppettoengine.InferenceResult `json:"inferenceResult,omitempty"`
}

func ReviewCompare(ctx context.Context, bootstrap *BootstrapResult, opts ReviewOptions) (ReviewResult, error) {
	if bootstrap == nil {
		return ReviewResult{}, fmt.Errorf("bootstrap result is nil")
	}

	eng, err := bootstrap.BuildEngine()
	if err != nil {
		return ReviewResult{}, err
	}

	promptText := BuildReviewPromptText(opts)
	images, artifacts, err := BuildReviewImages(opts.Evidence)
	if err != nil {
		return ReviewResult{}, err
	}

	systemPrompt := strings.TrimSpace(opts.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	turn := &turns.Turn{}
	turns.AppendBlock(turn, turns.NewSystemTextBlock(systemPrompt))
	turns.AppendBlock(turn, turns.NewUserMultimodalBlock(promptText, images))

	out, inferenceResult, err := geppettoengine.RunInferenceWithResult(ctx, eng, turn)
	if err != nil {
		return ReviewResult{}, err
	}

	answer := ExtractAssistantText(out)
	if strings.TrimSpace(answer) == "" {
		return ReviewResult{}, fmt.Errorf("inference returned no assistant text")
	}

	model := SelectedModel(bootstrap)
	provider := ""
	if inferenceResult != nil {
		if strings.TrimSpace(inferenceResult.Model) != "" {
			model = strings.TrimSpace(inferenceResult.Model)
		}
		provider = strings.TrimSpace(inferenceResult.Provider)
	}

	return ReviewResult{
		Question:        normalizedQuestion(opts.Question),
		Answer:          answer,
		Model:           model,
		Provider:        provider,
		APIType:         SelectedAPIType(bootstrap),
		Profile:         SelectedProfile(bootstrap),
		Registry:        SelectedRegistry(bootstrap),
		PromptSummary:   summarizePrompt(opts),
		Artifacts:       artifacts,
		InferenceResult: inferenceResult,
	}, nil
}

func BuildReviewPromptText(opts ReviewOptions) string {
	question := normalizedQuestion(opts.Question)
	maxProperties := opts.MaxProperties
	if maxProperties <= 0 {
		maxProperties = 12
	}
	maxWinnerDiffs := opts.MaxWinnerDiffs
	if maxWinnerDiffs <= 0 {
		maxWinnerDiffs = 6
	}

	var lines []string
	lines = append(lines, "Compare these two rendered UI regions and answer the question using both the screenshots and the structured evidence below.")
	lines = append(lines, "")
	lines = append(lines, "Question:")
	lines = append(lines, question)
	lines = append(lines, "")
	lines = append(lines, "Targets:")
	lines = append(lines, fmt.Sprintf("- Left: %s (%s)", opts.Evidence.URL1.URL, opts.Evidence.URL1.Selector))
	lines = append(lines, fmt.Sprintf("- Right: %s (%s)", opts.Evidence.URL2.URL, opts.Evidence.URL2.Selector))
	lines = append(lines, fmt.Sprintf("- Viewport: %dx%d", opts.Evidence.Inputs.ViewportW, opts.Evidence.Inputs.ViewportH))
	lines = append(lines, "")

	lines = append(lines, "Pixel diff summary:")
	lines = append(lines, fmt.Sprintf("- Changed pixels: %d / %d (%.2f%%) at threshold %d", opts.Evidence.PixelDiff.ChangedPixels, opts.Evidence.PixelDiff.TotalPixels, opts.Evidence.PixelDiff.ChangedPercent, opts.Evidence.PixelDiff.Threshold))
	lines = append(lines, "")

	lines = append(lines, "Computed property changes:")
	computedAdded := 0
	for _, diff := range opts.Evidence.ComputedDiffs {
		if strings.TrimSpace(diff.Original) == strings.TrimSpace(diff.React) {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s -> %s", diff.Property, safePromptValue(diff.Original), safePromptValue(diff.React)))
		computedAdded++
		if computedAdded >= maxProperties {
			break
		}
	}
	if computedAdded == 0 {
		lines = append(lines, "- No computed property changes were recorded in the selected property set.")
	}
	lines = append(lines, "")

	lines = append(lines, "Winning rule changes:")
	winnerAdded := 0
	for _, diff := range opts.Evidence.WinnerDiffs {
		left := strings.TrimSpace(diff.Original.Selector)
		right := strings.TrimSpace(diff.React.Selector)
		if left == right {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s -> %s", diff.Property, safePromptValue(left), safePromptValue(right)))
		winnerAdded++
		if winnerAdded >= maxWinnerDiffs {
			break
		}
	}
	if winnerAdded == 0 {
		lines = append(lines, "- No winning-rule selector changes were recorded in the selected property set.")
	}
	lines = append(lines, "")
	lines = append(lines, "Answer in concise engineering prose. Mention the biggest visual shifts, likely CSS causes, and any important UX impact.")

	return strings.Join(lines, "\n")
}

func BuildReviewImages(evidence modes.CompareResult) ([]map[string]any, map[string]string, error) {
	images := make([]map[string]any, 0, 3)
	artifacts := map[string]string{}

	left, leftPath, err := BuildImagePayload(evidence.URL1.ElementScreenshot, true)
	if err != nil {
		return nil, nil, fmt.Errorf("left screenshot: %w", err)
	}
	images = append(images, left)
	artifacts["left"] = leftPath

	right, rightPath, err := BuildImagePayload(evidence.URL2.ElementScreenshot, true)
	if err != nil {
		return nil, nil, fmt.Errorf("right screenshot: %w", err)
	}
	images = append(images, right)
	artifacts["right"] = rightPath

	if diff, diffPath, err := BuildImagePayload(evidence.PixelDiff.DiffComparisonPath, false); err != nil {
		return nil, nil, fmt.Errorf("diff comparison screenshot: %w", err)
	} else if diff != nil {
		images = append(images, diff)
		artifacts["comparison"] = diffPath
	}

	return images, artifacts, nil
}

func ExtractAssistantText(turn *turns.Turn) string {
	if turn == nil {
		return ""
	}
	for i := len(turn.Blocks) - 1; i >= 0; i-- {
		block := turn.Blocks[i]
		if block.Kind != turns.BlockKindLLMText {
			continue
		}
		if text, ok := block.Payload[turns.PayloadKeyText].(string); ok {
			return strings.TrimSpace(text)
		}
		if raw, ok := block.Payload[turns.PayloadKeyText]; ok && raw != nil {
			if bb, err := json.Marshal(raw); err == nil {
				return strings.TrimSpace(string(bb))
			}
		}
	}
	return ""
}

func WriteReviewJSON(path string, result ReviewResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func WriteReviewMarkdown(path string, result ReviewResult) error {
	var lines []string
	lines = append(lines, "# css-visual-diff LLM Review")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("**Question:** %s", result.Question))
	if result.Profile != "" {
		lines = append(lines, fmt.Sprintf("**Profile:** %s", result.Profile))
	}
	if result.Registry != "" {
		lines = append(lines, fmt.Sprintf("**Registry:** %s", result.Registry))
	}
	if result.Provider != "" || result.Model != "" {
		parts := []string{}
		if result.Provider != "" {
			parts = append(parts, result.Provider)
		}
		if result.Model != "" {
			parts = append(parts, result.Model)
		}
		lines = append(lines, fmt.Sprintf("**Inference:** %s", strings.Join(parts, " / ")))
	}
	lines = append(lines, "")
	lines = append(lines, "## Answer")
	lines = append(lines, "")
	lines = append(lines, result.Answer)
	lines = append(lines, "")
	if usageLines := ReviewUsageMarkdownLines(result); len(usageLines) > 0 {
		lines = append(lines, "## Token Usage")
		lines = append(lines, "")
		lines = append(lines, usageLines...)
		lines = append(lines, "")
	}
	if len(result.Artifacts) > 0 {
		keys := make([]string, 0, len(result.Artifacts))
		for key := range result.Artifacts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines = append(lines, "## Artifacts")
		lines = append(lines, "")
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("- **%s:** `%s`", key, result.Artifacts[key]))
		}
		lines = append(lines, "")
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func ReviewUsageMarkdownLines(result ReviewResult) []string {
	if result.InferenceResult == nil || result.InferenceResult.Usage == nil {
		return nil
	}
	usage := result.InferenceResult.Usage
	total := usage.InputTokens + usage.OutputTokens
	lines := []string{
		fmt.Sprintf("- **Input tokens:** %d", usage.InputTokens),
		fmt.Sprintf("- **Output tokens:** %d", usage.OutputTokens),
		fmt.Sprintf("- **Total tokens:** %d", total),
	}
	if usage.CachedTokens != 0 {
		lines = append(lines, fmt.Sprintf("- **Cached tokens:** %d", usage.CachedTokens))
	}
	if usage.CacheCreationInputTokens != 0 {
		lines = append(lines, fmt.Sprintf("- **Cache creation input tokens:** %d", usage.CacheCreationInputTokens))
	}
	if usage.CacheReadInputTokens != 0 {
		lines = append(lines, fmt.Sprintf("- **Cache read input tokens:** %d", usage.CacheReadInputTokens))
	}
	if reasoningTokens, ok := reviewExtraInt(result, "reasoning_tokens"); ok {
		lines = append(lines, fmt.Sprintf("- **Reasoning tokens:** %d", reasoningTokens))
	}
	if result.InferenceResult.MaxTokens != nil {
		lines = append(lines, fmt.Sprintf("- **Max tokens:** %d", *result.InferenceResult.MaxTokens))
	}
	if result.InferenceResult.DurationMs != nil {
		lines = append(lines, fmt.Sprintf("- **Duration:** %d ms", *result.InferenceResult.DurationMs))
	}
	return lines
}

func ReviewUsageConsoleText(result ReviewResult) string {
	lines := ReviewUsageMarkdownLines(result)
	if len(lines) == 0 {
		return ""
	}
	for i, line := range lines {
		line = strings.TrimPrefix(line, "- **")
		line = strings.Replace(line, "**", "", 1)
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func reviewExtraInt(result ReviewResult, key string) (int, bool) {
	if result.InferenceResult == nil || result.InferenceResult.Extra == nil {
		return 0, false
	}
	raw, ok := result.InferenceResult.Extra[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i), true
		}
	}
	return 0, false
}

func BuildImagePayload(path string, required bool) (map[string]any, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		if required {
			return nil, "", fmt.Errorf("path is empty")
		}
		return nil, "", nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if required {
			return nil, "", err
		}
		return nil, "", nil
	}
	if info.IsDir() {
		return nil, "", fmt.Errorf("%s is a directory", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return map[string]any{
		"media_type": detectImageMediaType(path),
		"content":    base64.StdEncoding.EncodeToString(content),
	}, path, nil
}

func detectImageMediaType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/png"
	}
}

func summarizePrompt(opts ReviewOptions) string {
	question := normalizedQuestion(opts.Question)
	return fmt.Sprintf("question=%q computedDiffs=%d winnerDiffs=%d pixelDiff=%.2f%%", question, len(opts.Evidence.ComputedDiffs), len(opts.Evidence.WinnerDiffs), opts.Evidence.PixelDiff.ChangedPercent)
}

func normalizedQuestion(question string) string {
	question = strings.TrimSpace(question)
	if question == "" {
		return "What are the main visual differences and their likely CSS causes?"
	}
	return question
}

func safePromptValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	return value
}
