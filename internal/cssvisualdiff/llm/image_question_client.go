package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/ai"
	geppettoengine "github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/geppetto/pkg/turns"
)

const defaultImageQuestionSystemPrompt = "You are an expert frontend engineer and visual QA analyst. Answer the user's question from the provided screenshot. Be concise, concrete, and mention visible UI evidence."

type ImageQuestionClient struct {
	Bootstrap    *BootstrapResult
	SystemPrompt string
	Detail       string
}

func NewImageQuestionClient(bootstrap *BootstrapResult) *ImageQuestionClient {
	return &ImageQuestionClient{
		Bootstrap: bootstrap,
		Detail:    "high",
	}
}

func (c *ImageQuestionClient) AnswerQuestion(ctx context.Context, imagePath string, question string) (ai.Answer, error) {
	if c == nil || c.Bootstrap == nil {
		return ai.Answer{}, fmt.Errorf("llm image question client is not configured")
	}

	eng, err := c.Bootstrap.BuildEngine()
	if err != nil {
		return ai.Answer{}, err
	}

	image, _, err := BuildImagePayload(imagePath, true)
	if err != nil {
		return ai.Answer{}, err
	}
	if detail := strings.TrimSpace(c.Detail); detail != "" {
		image["detail"] = detail
	}

	prompt := strings.TrimSpace(question)
	if prompt == "" {
		prompt = "Describe the visible UI and any likely frontend/CSS issue."
	}

	systemPrompt := strings.TrimSpace(c.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = defaultImageQuestionSystemPrompt
	}

	turn := &turns.Turn{}
	turns.AppendBlock(turn, turns.NewSystemTextBlock(systemPrompt))
	turns.AppendBlock(turn, turns.NewUserMultimodalBlock(prompt, []map[string]any{image}))

	out, inferenceResult, err := geppettoengine.RunInferenceWithResult(ctx, eng, turn)
	if err != nil {
		return ai.Answer{}, err
	}

	text := ExtractAssistantText(out)
	if strings.TrimSpace(text) == "" {
		return ai.Answer{}, fmt.Errorf("inference returned no assistant text")
	}

	confidence := 1.0
	if inferenceResult == nil {
		confidence = 0.9
	}
	return ai.Answer{Text: text, Confidence: confidence}, nil
}
