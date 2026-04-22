package llm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/modes"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/stretchr/testify/require"
)

func TestBuildReviewPromptTextIncludesQuestionAndEvidence(t *testing.T) {
	evidence := sampleCompareResult(t)
	text := BuildReviewPromptText(ReviewOptions{
		Question:      "What changed?",
		Evidence:      evidence,
		MaxProperties: 4,
	})

	require.Contains(t, text, "Question:")
	require.Contains(t, text, "What changed?")
	require.Contains(t, text, "Changed pixels: 123 / 1000 (12.30%)")
	require.Contains(t, text, "padding-top: 12px -> 20px")
	require.Contains(t, text, ".btn-old -> .btn-new")
}

func TestBuildReviewImagesLoadsRequiredAndOptionalArtifacts(t *testing.T) {
	evidence := sampleCompareResult(t)
	images, artifacts, err := BuildReviewImages(evidence)
	require.NoError(t, err)
	require.Len(t, images, 3)
	require.Equal(t, evidence.URL1.ElementScreenshot, artifacts["left"])
	require.Equal(t, evidence.URL2.ElementScreenshot, artifacts["right"])
	require.Equal(t, evidence.PixelDiff.DiffComparisonPath, artifacts["comparison"])
	require.Equal(t, "image/png", images[0]["media_type"])
	require.NotEmpty(t, images[0]["content"])
}

func TestExtractAssistantTextReturnsLatestLLMText(t *testing.T) {
	turn := &turns.Turn{}
	turns.AppendBlock(turn, turns.NewUserTextBlock("hello"))
	turns.AppendBlock(turn, turns.NewAssistantTextBlock("first"))
	turns.AppendBlock(turn, turns.NewAssistantTextBlock("second"))

	require.Equal(t, "second", ExtractAssistantText(turn))
}

func sampleCompareResult(t *testing.T) modes.CompareResult {
	t.Helper()
	tmp := t.TempDir()
	left := filepath.Join(tmp, "left.png")
	right := filepath.Join(tmp, "right.png")
	diff := filepath.Join(tmp, "diff.png")
	require.NoError(t, os.WriteFile(left, []byte("left"), 0o644))
	require.NoError(t, os.WriteFile(right, []byte("right"), 0o644))
	require.NoError(t, os.WriteFile(diff, []byte("diff"), 0o644))

	return modes.CompareResult{
		Inputs: modes.CompareInputs{
			ViewportW: 390,
			ViewportH: 844,
		},
		URL1: modes.CompareSideResult{
			URL:               "http://left.test",
			Selector:          "#cta",
			ElementScreenshot: left,
		},
		URL2: modes.CompareSideResult{
			URL:               "http://right.test",
			Selector:          "#cta",
			ElementScreenshot: right,
		},
		ComputedDiffs: []modes.StyleDiff{
			{Property: "padding-top", Original: "12px", React: "20px"},
			{Property: "background-color", Original: "rgb(10, 10, 10)", React: "rgb(200, 100, 100)"},
		},
		WinnerDiffs: []modes.WinnerDiff{
			{Property: "padding-top", Original: modes.Winner{Selector: ".btn-old"}, React: modes.Winner{Selector: ".btn-new"}},
		},
		PixelDiff: modes.PixelDiffStats{
			Threshold:          30,
			TotalPixels:        1000,
			ChangedPixels:      123,
			ChangedPercent:     12.3,
			DiffComparisonPath: diff,
		},
	}
}
