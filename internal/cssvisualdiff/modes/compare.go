package modes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

type CompareSettings struct {
	URL1      string
	Selector1 string
	WaitMS1   int

	URL2      string
	Selector2 string
	WaitMS2   int

	ViewportW int
	ViewportH int

	Props      []string
	Attributes []string

	OutDir string

	WriteJSON     bool
	WriteMarkdown bool
	WritePNGs     bool

	PixelDiffThreshold int
}

type CompareResult struct {
	Inputs CompareInputs `json:"inputs"`

	URL1 CompareSideResult `json:"url1"`
	URL2 CompareSideResult `json:"url2"`

	ComputedDiffs []StyleDiff  `json:"computed_diffs"`
	WinnerDiffs   []WinnerDiff `json:"winner_diffs"`

	PixelDiff PixelDiffStats `json:"pixel_diff"`
}

type CompareInputs struct {
	URL1      string   `json:"url1"`
	Selector1 string   `json:"selector1"`
	WaitMS1   int      `json:"wait_ms1"`
	URL2      string   `json:"url2"`
	Selector2 string   `json:"selector2"`
	WaitMS2   int      `json:"wait_ms2"`
	ViewportW int      `json:"viewport_w"`
	ViewportH int      `json:"viewport_h"`
	Props     []string `json:"props"`
	Attrs     []string `json:"attrs"`
	OutDir    string   `json:"out_dir"`
}

type CompareSideResult struct {
	URL      string `json:"url"`
	Selector string `json:"selector"`

	FullScreenshot    string `json:"full_screenshot"`
	ElementScreenshot string `json:"element_screenshot"`

	Computed StyleSnapshot   `json:"computed"`
	Matched  MatchedSnapshot `json:"matched"`
}

type PixelDiffStats struct {
	Threshold      int     `json:"threshold"`
	TotalPixels    int     `json:"total_pixels"`
	ChangedPixels  int     `json:"changed_pixels"`
	ChangedPercent float64 `json:"changed_percent"`

	NormalizedWidth  int `json:"normalized_width"`
	NormalizedHeight int `json:"normalized_height"`

	DiffComparisonPath string `json:"diff_comparison_path"`
	DiffOnlyPath       string `json:"diff_only_path"`
}

func Compare(ctx context.Context, settings CompareSettings) error {
	result, err := GenerateCompareResult(ctx, settings)
	if err != nil {
		return err
	}

	return WriteCompareArtifacts(result, settings.WriteJSON, settings.WriteMarkdown)
}

func WriteCompareArtifacts(result CompareResult, writeJSONFile bool, writeMarkdownFile bool) error {
	if writeJSONFile {
		if err := writeJSON(filepath.Join(result.Inputs.OutDir, "compare.json"), result); err != nil {
			return err
		}
	}
	if writeMarkdownFile {
		if err := writeCompareMarkdown(filepath.Join(result.Inputs.OutDir, "compare.md"), result); err != nil {
			return err
		}
	}
	return nil
}

func GenerateCompareResult(ctx context.Context, settings CompareSettings) (CompareResult, error) {
	if settings.URL1 == "" || settings.URL2 == "" {
		return CompareResult{}, fmt.Errorf("--url1 and --url2 are required")
	}
	if strings.TrimSpace(settings.Selector1) == "" {
		return CompareResult{}, fmt.Errorf("--selector1 is required")
	}
	if strings.TrimSpace(settings.Selector2) == "" {
		settings.Selector2 = settings.Selector1
	}
	if settings.ViewportW <= 0 || settings.ViewportH <= 0 {
		return CompareResult{}, fmt.Errorf("viewport must be positive")
	}
	if settings.PixelDiffThreshold < 0 || settings.PixelDiffThreshold > 255 {
		return CompareResult{}, fmt.Errorf("--threshold must be between 0 and 255")
	}

	if settings.OutDir == "" {
		settings.OutDir = fmt.Sprintf("./css-visual-diff-compare-%s", time.Now().Format("20060102_150405"))
	}
	settings.OutDir = filepath.Clean(settings.OutDir)
	if err := os.MkdirAll(settings.OutDir, 0o755); err != nil {
		return CompareResult{}, err
	}

	browser, err := driver.NewBrowser(ctx)
	if err != nil {
		return CompareResult{}, err
	}
	defer browser.Close()

	page1, err := browser.NewPage()
	if err != nil {
		return CompareResult{}, err
	}
	defer page1.Close()

	page2, err := browser.NewPage()
	if err != nil {
		return CompareResult{}, err
	}
	defer page2.Close()

	if err := page1.SetViewport(settings.ViewportW, settings.ViewportH); err != nil {
		return CompareResult{}, err
	}
	if err := page2.SetViewport(settings.ViewportW, settings.ViewportH); err != nil {
		return CompareResult{}, err
	}

	side1, err := captureCompareSide(
		page1,
		settings.URL1,
		settings.Selector1,
		settings.WaitMS1,
		settings.OutDir,
		"url1",
		settings.Props,
		settings.Attributes,
	)
	if err != nil {
		return CompareResult{}, err
	}

	side2, err := captureCompareSide(
		page2,
		settings.URL2,
		settings.Selector2,
		settings.WaitMS2,
		settings.OutDir,
		"url2",
		settings.Props,
		settings.Attributes,
	)
	if err != nil {
		return CompareResult{}, err
	}

	computedDiffs := buildDiffs(settings.Props, side1.Computed, side2.Computed)
	winnerDiffs := buildWinnerDiffs(settings.Props, side1.Matched, side2.Matched)

	diffComparisonPath := filepath.Join(settings.OutDir, "diff_comparison.png")
	diffOnlyPath := filepath.Join(settings.OutDir, "diff_only.png")
	pixelDiffStats := PixelDiffStats{
		Threshold:          settings.PixelDiffThreshold,
		DiffComparisonPath: diffComparisonPath,
		DiffOnlyPath:       diffOnlyPath,
	}
	if settings.WritePNGs {
		stats, err := writePixelDiffImages(side1.ElementScreenshot, side2.ElementScreenshot, diffComparisonPath, diffOnlyPath, settings.PixelDiffThreshold)
		if err != nil {
			return CompareResult{}, err
		}
		pixelDiffStats = stats
	}

	result := CompareResult{
		Inputs: CompareInputs{
			URL1:      settings.URL1,
			Selector1: settings.Selector1,
			WaitMS1:   settings.WaitMS1,
			URL2:      settings.URL2,
			Selector2: settings.Selector2,
			WaitMS2:   settings.WaitMS2,
			ViewportW: settings.ViewportW,
			ViewportH: settings.ViewportH,
			Props:     settings.Props,
			Attrs:     settings.Attributes,
			OutDir:    settings.OutDir,
		},
		URL1:          side1,
		URL2:          side2,
		ComputedDiffs: computedDiffs,
		WinnerDiffs:   winnerDiffs,
		PixelDiff:     pixelDiffStats,
	}

	return result, nil
}

func captureCompareSide(
	page *driver.Page,
	url string,
	selector string,
	waitMS int,
	outDir string,
	prefix string,
	props []string,
	attrs []string,
) (CompareSideResult, error) {
	if err := page.Goto(url); err != nil {
		return CompareSideResult{}, err
	}
	if waitMS > 0 {
		page.Wait(time.Duration(waitMS) * time.Millisecond)
	}

	fullPath := filepath.Join(outDir, fmt.Sprintf("%s_full.png", prefix))
	if err := page.FullScreenshot(fullPath); err != nil {
		return CompareSideResult{}, err
	}

	elementPath := filepath.Join(outDir, fmt.Sprintf("%s_screenshot.png", prefix))
	if err := screenshotSelector(page, selector, elementPath); err != nil {
		return CompareSideResult{}, err
	}

	styleSpec := config.StyleSpec{
		Name:          prefix,
		Selector:      selector,
		Props:         props,
		IncludeBounds: true,
		Attributes:    attrs,
	}
	computed, err := evaluateStyle(page, styleSpec)
	if err != nil {
		return CompareSideResult{}, err
	}

	matchedSpec := config.StyleSpec{
		Name:     prefix,
		Selector: selector,
		Props:    props,
		Report:   []string{"box_model"},
	}
	matched, err := evaluateMatched(page, matchedSpec)
	if err != nil {
		return CompareSideResult{}, err
	}

	return CompareSideResult{
		URL:               url,
		Selector:          selector,
		FullScreenshot:    fullPath,
		ElementScreenshot: elementPath,
		Computed:          computed,
		Matched:           matched,
	}, nil
}

func screenshotSelector(page *driver.Page, selector, outPath string) error {
	_ = chromedp.Run(page.Context(), chromedp.ScrollIntoView(selector, chromedp.ByQuery))
	return page.Screenshot(selector, outPath)
}

func writeCompareMarkdown(path string, result CompareResult) error {
	content := "# css-visual-diff Compare Report\n\n"
	content += "## Inputs\n\n"
	content += fmt.Sprintf("- URL 1: %s\n", result.Inputs.URL1)
	content += fmt.Sprintf("- Selector 1: `%s`\n", result.Inputs.Selector1)
	content += fmt.Sprintf("- URL 2: %s\n", result.Inputs.URL2)
	content += fmt.Sprintf("- Selector 2: `%s`\n", result.Inputs.Selector2)
	content += fmt.Sprintf("- Viewport: %dx%d\n", result.Inputs.ViewportW, result.Inputs.ViewportH)
	content += fmt.Sprintf("- Pixel threshold: %d\n", result.PixelDiff.Threshold)
	content += "\n"

	content += "## Artifacts\n\n"
	content += fmt.Sprintf("- URL1 full: %s\n", result.URL1.FullScreenshot)
	content += fmt.Sprintf("- URL2 full: %s\n", result.URL2.FullScreenshot)
	content += fmt.Sprintf("- URL1 element: %s\n", result.URL1.ElementScreenshot)
	content += fmt.Sprintf("- URL2 element: %s\n", result.URL2.ElementScreenshot)
	if result.PixelDiff.DiffComparisonPath != "" {
		content += fmt.Sprintf("- Diff comparison: %s\n", result.PixelDiff.DiffComparisonPath)
	}
	if result.PixelDiff.DiffOnlyPath != "" {
		content += fmt.Sprintf("- Diff only: %s\n", result.PixelDiff.DiffOnlyPath)
	}
	content += "\n"

	content += "## Pixel Diff Stats\n\n"
	content += fmt.Sprintf("- Total pixels: %d\n", result.PixelDiff.TotalPixels)
	content += fmt.Sprintf("- Changed pixels: %d\n", result.PixelDiff.ChangedPixels)
	content += fmt.Sprintf("- Changed percent: %.4f%%\n", result.PixelDiff.ChangedPercent)
	content += fmt.Sprintf("- Normalized size: %dx%d\n", result.PixelDiff.NormalizedWidth, result.PixelDiff.NormalizedHeight)
	content += "\n"

	content += "## Computed Style Diffs\n\n"
	if len(result.ComputedDiffs) == 0 {
		content += "No computed diffs for selected props.\n\n"
	} else {
		content += "| Property | URL1 | URL2 |\n"
		content += "| --- | --- | --- |\n"
		for _, d := range result.ComputedDiffs {
			content += fmt.Sprintf("| %s | %s | %s |\n", d.Property, d.Original, d.React)
		}
		content += "\n"
	}

	content += "## Winner Diffs (Matched Styles)\n\n"
	content += "| Property | URL1 Winner | URL2 Winner |\n"
	content += "| --- | --- | --- |\n"
	for _, w := range result.WinnerDiffs {
		content += fmt.Sprintf("| %s | %s | %s |\n", w.Property, formatWinner(w.Original), formatWinner(w.React))
	}
	content += "\n"

	return os.WriteFile(path, []byte(content), 0o644)
}

func writePixelDiffImages(url1Path, url2Path, diffComparisonPath, diffOnlyPath string, threshold int) (PixelDiffStats, error) {
	img1, err := readPNG(url1Path)
	if err != nil {
		return PixelDiffStats{}, err
	}
	img2, err := readPNG(url2Path)
	if err != nil {
		return PixelDiffStats{}, err
	}

	n1, n2 := padToSameSize(img1, img2)
	stats, diffOnly := computePixelDiff(n1, n2, threshold)
	stats.DiffComparisonPath = diffComparisonPath
	stats.DiffOnlyPath = diffOnlyPath

	if err := writePNG(diffOnlyPath, diffOnly); err != nil {
		return PixelDiffStats{}, err
	}

	diffComparison := combineSideBySide(n1, n2, diffOnly)
	if err := writePNG(diffComparisonPath, diffComparison); err != nil {
		return PixelDiffStats{}, err
	}

	return stats, nil
}
