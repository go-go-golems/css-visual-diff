package modes

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestRunPixelDiff_WritesOutputs(t *testing.T) {
	tmp := t.TempDir()

	mk := func(path string, c color.NRGBA) {
		img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
		for y := 0; y < 2; y++ {
			for x := 0; x < 2; x++ {
				i := y*img.Stride + x*4
				img.Pix[i+0] = c.R
				img.Pix[i+1] = c.G
				img.Pix[i+2] = c.B
				img.Pix[i+3] = c.A
			}
		}
		if err := writePNG(path, img); err != nil {
			t.Fatalf("writePNG(%s): %v", path, err)
		}
	}

	orig1 := filepath.Join(tmp, "original-section1.png")
	react1 := filepath.Join(tmp, "react-section1.png")
	mk(orig1, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	mk(react1, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	// One pixel difference.
	img, err := readPNG(react1)
	if err != nil {
		t.Fatalf("readPNG: %v", err)
	}
	n := toNRGBA(img)
	n.Pix[0] = 255
	n.Pix[1] = 255
	n.Pix[2] = 255
	n.Pix[3] = 255
	if err := writePNG(react1, n); err != nil {
		t.Fatalf("writePNG: %v", err)
	}

	capture := CaptureResult{
		Original: PageResult{
			Name: "original",
			URL:  "http://example.com/a",
			Sections: []SectionResult{
				{Name: "section1", Screenshot: orig1, Exists: true, Visible: true},
			},
		},
		React: PageResult{
			Name: "react",
			URL:  "http://example.com/b",
			Sections: []SectionResult{
				{Name: "section1", Screenshot: react1, Exists: true, Visible: true},
			},
		},
	}

	captureJSON, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		t.Fatalf("marshal capture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "capture.json"), captureJSON, 0o644); err != nil {
		t.Fatalf("write capture.json: %v", err)
	}

	if err := RunPixelDiff(context.Background(), tmp, 30); err != nil {
		t.Fatalf("RunPixelDiff: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "pixeldiff.json")); err != nil {
		t.Fatalf("expected pixeldiff.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "pixeldiff.md")); err != nil {
		t.Fatalf("expected pixeldiff.md: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "pixeldiff.json"))
	if err != nil {
		t.Fatalf("read pixeldiff.json: %v", err)
	}
	var out PixelDiffResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal pixeldiff.json: %v", err)
	}
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Skipped {
		t.Fatalf("expected not skipped, got skipped: %s", out.Entries[0].SkipReason)
	}
	if out.Entries[0].ChangedPixels != 1 {
		t.Fatalf("expected changed pixels 1, got %d", out.Entries[0].ChangedPixels)
	}
	if _, err := os.Stat(out.Entries[0].DiffOnlyPath); err != nil {
		t.Fatalf("expected diff-only png: %v", err)
	}
	if _, err := os.Stat(out.Entries[0].DiffComparisonPath); err != nil {
		t.Fatalf("expected diff-comparison png: %v", err)
	}
}
