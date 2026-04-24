package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type PixelDiffEntry struct {
	Section string `json:"section"`

	OriginalScreenshot string `json:"original_screenshot"`
	ReactScreenshot    string `json:"react_screenshot"`

	DiffComparisonPath string `json:"diff_comparison_path,omitempty"`
	DiffOnlyPath       string `json:"diff_only_path,omitempty"`

	Threshold      int     `json:"threshold"`
	TotalPixels    int     `json:"total_pixels"`
	ChangedPixels  int     `json:"changed_pixels"`
	ChangedPercent float64 `json:"changed_percent"`

	NormalizedWidth  int `json:"normalized_width"`
	NormalizedHeight int `json:"normalized_height"`

	Skipped    bool   `json:"skipped"`
	SkipReason string `json:"skip_reason,omitempty"`
}

type PixelDiffResult struct {
	Threshold int              `json:"threshold"`
	Entries   []PixelDiffEntry `json:"entries"`
}

func RunPixelDiff(ctx context.Context, cfgDir string, threshold int) error {
	if threshold < 0 || threshold > 255 {
		return fmt.Errorf("pixeldiff threshold must be between 0 and 255")
	}

	capturePath := filepath.Join(cfgDir, "capture.json")
	data, err := os.ReadFile(capturePath)
	if err != nil {
		return fmt.Errorf("pixeldiff requires capture.json: %w", err)
	}

	var capture CaptureResult
	if err := json.Unmarshal(data, &capture); err != nil {
		return err
	}

	result := PixelDiffResult{Threshold: threshold}
	sections := capture.Original.Sections
	for i := range sections {
		entry := PixelDiffEntry{
			Section:            sections[i].Name,
			OriginalScreenshot: sections[i].Screenshot,
			Threshold:          threshold,
		}
		if i < len(capture.React.Sections) {
			entry.ReactScreenshot = capture.React.Sections[i].Screenshot
		}

		if entry.OriginalScreenshot == "" || entry.ReactScreenshot == "" {
			entry.Skipped = true
			entry.SkipReason = "missing screenshot on one or both sides"
			result.Entries = append(result.Entries, entry)
			continue
		}

		img1, err := readPNG(entry.OriginalScreenshot)
		if err != nil {
			entry.Skipped = true
			entry.SkipReason = fmt.Sprintf("failed to read original screenshot: %v", err)
			result.Entries = append(result.Entries, entry)
			continue
		}
		img2, err := readPNG(entry.ReactScreenshot)
		if err != nil {
			entry.Skipped = true
			entry.SkipReason = fmt.Sprintf("failed to read react screenshot: %v", err)
			result.Entries = append(result.Entries, entry)
			continue
		}

		n1, n2 := padToSameSize(img1, img2)
		stats, diffOnly := computePixelDiff(n1, n2, threshold)

		diffOnlyPath := filepath.Join(cfgDir, fmt.Sprintf("pixeldiff_%s_diff_only.png", sanitizeName(entry.Section)))
		diffComparisonPath := filepath.Join(cfgDir, fmt.Sprintf("pixeldiff_%s_diff_comparison.png", sanitizeName(entry.Section)))

		if err := writePNG(diffOnlyPath, diffOnly); err != nil {
			entry.Skipped = true
			entry.SkipReason = fmt.Sprintf("failed to write diff-only: %v", err)
			result.Entries = append(result.Entries, entry)
			continue
		}

		diffComparison := combineSideBySide(n1, n2, diffOnly)
		if err := writePNG(diffComparisonPath, diffComparison); err != nil {
			entry.Skipped = true
			entry.SkipReason = fmt.Sprintf("failed to write diff-comparison: %v", err)
			result.Entries = append(result.Entries, entry)
			continue
		}

		entry.DiffOnlyPath = diffOnlyPath
		entry.DiffComparisonPath = diffComparisonPath
		entry.TotalPixels = stats.TotalPixels
		entry.ChangedPixels = stats.ChangedPixels
		entry.ChangedPercent = stats.ChangedPercent
		entry.NormalizedWidth = stats.NormalizedWidth
		entry.NormalizedHeight = stats.NormalizedHeight

		result.Entries = append(result.Entries, entry)
	}

	sort.SliceStable(result.Entries, func(i, j int) bool {
		if result.Entries[i].Skipped != result.Entries[j].Skipped {
			return !result.Entries[i].Skipped && result.Entries[j].Skipped
		}
		return result.Entries[i].ChangedPercent > result.Entries[j].ChangedPercent
	})

	if err := writeJSON(filepath.Join(cfgDir, "pixeldiff.json"), result); err != nil {
		return err
	}
	if err := writePixelDiffMarkdown(filepath.Join(cfgDir, "pixeldiff.md"), result); err != nil {
		return err
	}

	_ = ctx
	return nil
}

func writePixelDiffMarkdown(path string, result PixelDiffResult) error {
	content := "# css-visual-diff Pixel Diff Report\n\n"
	content += fmt.Sprintf("- Threshold: %d\n\n", result.Threshold)
	content += "| Section | Changed % | Changed/Total | Diff |\n"
	content += "| --- | ---: | ---: | --- |\n"
	for _, e := range result.Entries {
		if e.Skipped {
			content += fmt.Sprintf("| %s | (skipped) |  | %s |\n", e.Section, e.SkipReason)
			continue
		}
		content += fmt.Sprintf("| %s | %.4f%% | %d/%d | %s |\n", e.Section, e.ChangedPercent, e.ChangedPixels, e.TotalPixels, e.DiffComparisonPath)
	}
	content += "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func sanitizeName(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}
