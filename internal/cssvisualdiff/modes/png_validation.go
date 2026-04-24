package modes

import (
	"fmt"
	"image"
	"math"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

type ValidationResult struct {
	Target  string    `json:"target"`
	Section string    `json:"section"`
	Status  string    `json:"status"`
	Issues  []string  `json:"issues,omitempty"`
	PNG     *PNGStats `json:"png,omitempty"`
	Bounds  *Bounds   `json:"bounds,omitempty"`
}

type PNGStats struct {
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	TopStripAverage    [3]int `json:"top_strip_average"`
	BottomStripAverage [3]int `json:"bottom_strip_average"`
}

func analyzePNG(path string, stripHeight int) (PNGStats, error) {
	img, err := readPNG(path)
	if err != nil {
		return PNGStats{}, err
	}
	bounds := img.Bounds()
	if stripHeight <= 0 {
		stripHeight = 20
	}
	if stripHeight > bounds.Dy() {
		stripHeight = bounds.Dy()
	}
	return PNGStats{
		Width:              bounds.Dx(),
		Height:             bounds.Dy(),
		TopStripAverage:    averageStrip(img, bounds.Min.Y, bounds.Min.Y+stripHeight),
		BottomStripAverage: averageStrip(img, bounds.Max.Y-stripHeight, bounds.Max.Y),
	}, nil
}

func averageStrip(img image.Image, y0, y1 int) [3]int {
	bounds := img.Bounds()
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if y1 > bounds.Max.Y {
		y1 = bounds.Max.Y
	}
	if y1 <= y0 || bounds.Dx() <= 0 {
		return [3]int{}
	}

	var rSum, gSum, bSum, count int64
	for y := y0; y < y1; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rSum += int64(r >> 8)
			gSum += int64(g >> 8)
			bSum += int64(b >> 8)
			count++
		}
	}
	if count == 0 {
		return [3]int{}
	}
	return [3]int{int(rSum / count), int(gSum / count), int(bSum / count)}
}

func validatePNGStats(stats PNGStats, expectations *config.PNGExpectations) []string {
	if expectations == nil {
		return nil
	}
	var issues []string
	if expectations.Width > 0 && stats.Width != expectations.Width {
		issues = append(issues, fmt.Sprintf("expected width %d, got %d", expectations.Width, stats.Width))
	}
	if expectations.Height > 0 && stats.Height != expectations.Height {
		issues = append(issues, fmt.Sprintf("expected height %d, got %d", expectations.Height, stats.Height))
	}
	if expectations.MinWidth > 0 && stats.Width < expectations.MinWidth {
		issues = append(issues, fmt.Sprintf("expected width >= %d, got %d", expectations.MinWidth, stats.Width))
	}
	if expectations.MinHeight > 0 && stats.Height < expectations.MinHeight {
		issues = append(issues, fmt.Sprintf("expected height >= %d, got %d", expectations.MinHeight, stats.Height))
	}
	if expectations.MaxWidth > 0 && stats.Width > expectations.MaxWidth {
		issues = append(issues, fmt.Sprintf("expected width <= %d, got %d", expectations.MaxWidth, stats.Width))
	}
	if expectations.MaxHeight > 0 && stats.Height > expectations.MaxHeight {
		issues = append(issues, fmt.Sprintf("expected height <= %d, got %d", expectations.MaxHeight, stats.Height))
	}
	issues = append(issues, validateColorExpectation("top strip", stats.TopStripAverage, expectations.TopStripNear, true)...)
	issues = append(issues, validateColorExpectation("top strip", stats.TopStripAverage, expectations.TopStripNotNear, false)...)
	issues = append(issues, validateColorExpectation("bottom strip", stats.BottomStripAverage, expectations.BottomStripNear, true)...)
	issues = append(issues, validateColorExpectation("bottom strip", stats.BottomStripAverage, expectations.BottomStripNotNear, false)...)
	return issues
}

func validateColorExpectation(label string, got [3]int, expectation *config.ColorExpectation, shouldBeNear bool) []string {
	if expectation == nil {
		return nil
	}
	expected, ok := rgbExpectation(expectation)
	if !ok {
		return []string{fmt.Sprintf("%s color expectation must include exactly 3 rgb values", label)}
	}
	tolerance := expectation.Tolerance
	if tolerance < 0 {
		tolerance = 0
	}
	near := colorDistance(got, expected) <= float64(tolerance)
	if shouldBeNear && !near {
		return []string{fmt.Sprintf("%s average %v is not within %d of %v", label, got, tolerance, expected)}
	}
	if !shouldBeNear && near {
		return []string{fmt.Sprintf("%s average %v should not be within %d of %v", label, got, tolerance, expected)}
	}
	return nil
}

func rgbExpectation(expectation *config.ColorExpectation) ([3]int, bool) {
	if expectation == nil || len(expectation.RGB) != 3 {
		return [3]int{}, false
	}
	return [3]int{expectation.RGB[0], expectation.RGB[1], expectation.RGB[2]}, true
}

func colorDistance(a, b [3]int) float64 {
	dr := float64(a[0] - b[0])
	dg := float64(a[1] - b[1])
	db := float64(a[2] - b[2])
	return math.Sqrt(dr*dr + dg*dg + db*db)
}
