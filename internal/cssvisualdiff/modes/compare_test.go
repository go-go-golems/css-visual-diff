package modes

import (
	"image"
	"image/color"
	"testing"
)

func TestComputePixelDiff_NoOverflow(t *testing.T) {
	img1 := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img2 := image.NewNRGBA(image.Rect(0, 0, 2, 2))

	fill := func(img *image.NRGBA, c color.NRGBA) {
		for y := 0; y < 2; y++ {
			for x := 0; x < 2; x++ {
				i := y*img.Stride + x*4
				img.Pix[i+0] = c.R
				img.Pix[i+1] = c.G
				img.Pix[i+2] = c.B
				img.Pix[i+3] = c.A
			}
		}
	}

	fill(img1, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	fill(img2, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	// Create a large per-channel delta (255) that would overflow if squared in uint8.
	img2.Pix[0] = 255
	img2.Pix[1] = 255
	img2.Pix[2] = 255
	img2.Pix[3] = 255

	stats, overlay := computePixelDiff(img1, img2, 30)
	if stats.TotalPixels != 4 {
		t.Fatalf("expected total pixels 4, got %d", stats.TotalPixels)
	}
	if stats.ChangedPixels != 1 {
		t.Fatalf("expected changed pixels 1, got %d", stats.ChangedPixels)
	}

	// Changed pixel should be painted red.
	if overlay.Pix[0] != 255 || overlay.Pix[1] != 0 || overlay.Pix[2] != 0 || overlay.Pix[3] != 255 {
		t.Fatalf("expected overlay pixel red, got rgba=(%d,%d,%d,%d)", overlay.Pix[0], overlay.Pix[1], overlay.Pix[2], overlay.Pix[3])
	}
}

func TestComputePixelDiff_Identical(t *testing.T) {
	img1 := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img2 := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	for i := 0; i < len(img1.Pix); i++ {
		img1.Pix[i] = 42
		img2.Pix[i] = 42
	}

	stats, _ := computePixelDiff(img1, img2, 0)
	if stats.ChangedPixels != 0 {
		t.Fatalf("expected changed pixels 0, got %d", stats.ChangedPixels)
	}
}
