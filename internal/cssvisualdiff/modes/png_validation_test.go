package modes

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func TestAnalyzePNGComputesStripAverages(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 2; x++ {
			idx := y*img.Stride + x*4
			if y < 2 {
				img.Pix[idx+0] = 255
				img.Pix[idx+1] = 255
				img.Pix[idx+2] = 255
				img.Pix[idx+3] = 255
			} else {
				img.Pix[idx+0] = 10
				img.Pix[idx+1] = 20
				img.Pix[idx+2] = 30
				img.Pix[idx+3] = 255
			}
		}
	}
	path := filepath.Join(t.TempDir(), "sample.png")
	if err := writePNG(path, img); err != nil {
		t.Fatalf("writePNG: %v", err)
	}

	stats, err := analyzePNG(path, 2)
	if err != nil {
		t.Fatalf("analyzePNG: %v", err)
	}
	if stats.Width != 2 || stats.Height != 4 {
		t.Fatalf("expected 2x4, got %dx%d", stats.Width, stats.Height)
	}
	if stats.TopStripAverage != [3]int{255, 255, 255} {
		t.Fatalf("unexpected top average: %v", stats.TopStripAverage)
	}
	if stats.BottomStripAverage != [3]int{10, 20, 30} {
		t.Fatalf("unexpected bottom average: %v", stats.BottomStripAverage)
	}
}

func TestValidatePNGStats(t *testing.T) {
	stats := PNGStats{
		Width:              920,
		Height:             1801,
		TopStripAverage:    [3]int{255, 255, 255},
		BottomStripAverage: [3]int{254, 254, 253},
	}
	issues := validatePNGStats(stats, &config.PNGExpectations{
		Width:     920,
		MinHeight: 1700,
		TopStripNear: &config.ColorExpectation{
			RGB:       []int{255, 255, 255},
			Tolerance: 5,
		},
		TopStripNotNear: &config.ColorExpectation{
			RGB:       []int{240, 238, 233},
			Tolerance: 5,
		},
	})
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}

	issues = validatePNGStats(stats, &config.PNGExpectations{
		Width:     921,
		MinHeight: 1900,
		TopStripNear: &config.ColorExpectation{
			RGB:       []int{240, 238, 233},
			Tolerance: 5,
		},
	})
	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d: %v", len(issues), issues)
	}
}

func TestColorDistanceUsesRGB(t *testing.T) {
	_ = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	if got := colorDistance([3]int{255, 255, 255}, [3]int{255, 255, 255}); got != 0 {
		t.Fatalf("expected distance 0, got %f", got)
	}
}
