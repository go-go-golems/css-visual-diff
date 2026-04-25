package service

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareSelectionsDiffsStylesAttributesTextAndBounds(t *testing.T) {
	left := sampleSelection("left")
	right := sampleSelection("right")
	right.Text = "Book later"
	right.Bounds = &Bounds{X: 12, Y: 20, Width: 101, Height: 40}
	right.ComputedStyles["color"] = "rgb(255, 0, 0)"
	right.ComputedStyles["font-size"] = "18px"
	right.ComputedStyles["z-index"] = "2"
	right.Attributes["class"] = "secondary"
	right.Attributes["data-state"] = "active"

	comparison, err := CompareSelections(left, right, CompareSelectionOptions{
		Name:              "cta-compare",
		StyleProps:        []string{"color", "font-size", "display"},
		Attributes:        []string{"class", "data-state"},
		ExcludeAttributes: []string{"data-state"},
	})
	require.NoError(t, err)
	require.Equal(t, SelectionComparisonSchemaVersion, comparison.SchemaVersion)
	require.Equal(t, "cta-compare", comparison.Name)
	require.Equal(t, "#cta", comparison.Left.Selector)
	require.Equal(t, "#cta", comparison.Right.Selector)
	require.True(t, comparison.Bounds.Changed)
	require.NotNil(t, comparison.Bounds.Delta)
	require.Equal(t, 2.0, comparison.Bounds.Delta.X)
	require.Equal(t, 1.0, comparison.Bounds.Delta.Width)
	require.True(t, comparison.Text.Changed)
	require.Equal(t, "Book now", comparison.Text.Left)
	require.Equal(t, "Book later", comparison.Text.Right)
	require.Equal(t, []MapValueDiff{
		{Name: "color", Left: "rgb(0, 0, 0)", Right: "rgb(255, 0, 0)", Changed: true},
		{Name: "font-size", Left: "16px", Right: "18px", Changed: true},
	}, comparison.Styles)
	require.Equal(t, []MapValueDiff{
		{Name: "class", Left: "primary", Right: "secondary", Changed: true},
	}, comparison.Attributes)
}

func TestCompareSelectionsDeterministicOrderingAndExclude(t *testing.T) {
	left := sampleSelection("left")
	right := sampleSelection("right")
	left.ComputedStyles = map[string]string{"z": "1", "a": "1", "m": "1", "skip": "left"}
	right.ComputedStyles = map[string]string{"z": "2", "a": "2", "m": "1", "skip": "right"}
	left.Attributes = map[string]string{"data-z": "1", "data-a": "1"}
	right.Attributes = map[string]string{"data-z": "2", "data-a": "2"}

	comparison, err := CompareSelections(left, right, CompareSelectionOptions{ExcludeStyleProps: []string{"skip"}})
	require.NoError(t, err)
	require.Equal(t, []MapValueDiff{
		{Name: "a", Left: "1", Right: "2", Changed: true},
		{Name: "z", Left: "1", Right: "2", Changed: true},
	}, comparison.Styles)
	require.Equal(t, []MapValueDiff{
		{Name: "data-a", Left: "1", Right: "2", Changed: true},
		{Name: "data-z", Left: "1", Right: "2", Changed: true},
	}, comparison.Attributes)
}

func TestCompareSelectionsPixelDiffIntegration(t *testing.T) {
	tmp := t.TempDir()
	leftPath := filepath.Join(tmp, "left.png")
	rightPath := filepath.Join(tmp, "right.png")
	require.NoError(t, WritePNG(leftPath, solidNRGBA(2, 2, color.NRGBA{A: 255})))
	rightImg := solidNRGBA(2, 2, color.NRGBA{A: 255})
	rightImg.Pix[0] = 255
	require.NoError(t, WritePNG(rightPath, rightImg))

	left := sampleSelection("left")
	right := sampleSelection("right")
	left.Screenshot = &ScreenshotDescriptor{Path: leftPath}
	right.Screenshot = &ScreenshotDescriptor{Path: rightPath}

	diffOnlyPath := filepath.Join(tmp, "artifacts", "diff_only.png")
	diffComparisonPath := filepath.Join(tmp, "artifacts", "diff_comparison.png")
	comparison, err := CompareSelections(left, right, CompareSelectionOptions{
		Threshold:          0,
		DiffOnlyPath:       diffOnlyPath,
		DiffComparisonPath: diffComparisonPath,
	})
	require.NoError(t, err)
	require.NotNil(t, comparison.Pixel)
	require.Equal(t, 1, comparison.Pixel.ChangedPixels)
	require.Equal(t, diffOnlyPath, comparison.Pixel.DiffOnlyPath)
	require.Equal(t, diffComparisonPath, comparison.Pixel.DiffComparisonPath)
	require.Equal(t, []SelectionArtifact{
		{Name: "diffOnly", Path: filepath.Clean(diffOnlyPath), Kind: "png"},
		{Name: "diffComparison", Path: filepath.Clean(diffComparisonPath), Kind: "png"},
	}, comparison.Artifacts)
	for _, path := range []string{diffOnlyPath, diffComparisonPath} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Greater(t, info.Size(), int64(0))
	}
}

func TestCompareSelectionsJSONSerialization(t *testing.T) {
	left := sampleSelection("left")
	right := sampleSelection("right")
	right.Text = "Changed"

	comparison, err := CompareSelections(left, right, CompareSelectionOptions{})
	require.NoError(t, err)
	payload, err := json.Marshal(comparison)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"schemaVersion":"cssvd.selectionComparison.v1"`)
	require.Contains(t, string(payload), `"changed":true`)
	require.NotContains(t, string(payload), `schema_version`)

	var roundTrip SelectionComparisonData
	require.NoError(t, json.Unmarshal(payload, &roundTrip))
	require.Equal(t, SelectionComparisonSchemaVersion, roundTrip.SchemaVersion)
	require.True(t, roundTrip.Text.Changed)
}

func TestCompareSelectionsRejectsInvalidThreshold(t *testing.T) {
	_, err := CompareSelections(sampleSelection("left"), sampleSelection("right"), CompareSelectionOptions{Threshold: 300})
	require.Error(t, err)
}

func sampleSelection(name string) SelectionData {
	return SelectionData{
		SchemaVersion: CollectedSelectionSchemaVersion,
		Name:          name,
		URL:           "http://example.com/" + name,
		Selector:      "#cta",
		Source:        "test",
		Status:        SelectorStatus{Name: name, Selector: "#cta", Exists: true, Visible: true},
		Exists:        true,
		Visible:       true,
		Bounds:        &Bounds{X: 10, Y: 20, Width: 100, Height: 40},
		Text:          "Book now",
		ComputedStyles: map[string]string{
			"color":     "rgb(0, 0, 0)",
			"display":   "block",
			"font-size": "16px",
		},
		Attributes: map[string]string{
			"class":      "primary",
			"data-state": "idle",
		},
	}
}
