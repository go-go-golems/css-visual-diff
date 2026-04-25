package service

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

type PixelDiffOptions struct {
	Threshold int `json:"threshold,omitempty"`
}

type PixelDiffResult struct {
	Threshold      int     `json:"threshold"`
	TotalPixels    int     `json:"totalPixels"`
	ChangedPixels  int     `json:"changedPixels"`
	ChangedPercent float64 `json:"changedPercent"`

	NormalizedWidth  int `json:"normalizedWidth"`
	NormalizedHeight int `json:"normalizedHeight"`

	DiffComparisonPath string `json:"diffComparisonPath,omitempty"`
	DiffOnlyPath       string `json:"diffOnlyPath,omitempty"`
}

func ValidatePixelThreshold(threshold int) error {
	if threshold < 0 || threshold > 255 {
		return fmt.Errorf("pixel diff threshold must be between 0 and 255")
	}
	return nil
}

func ReadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func WritePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func ToNRGBA(img image.Image) *image.NRGBA {
	if n, ok := img.(*image.NRGBA); ok {
		return n
	}
	b := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

func PadToSameSize(a, b image.Image) (*image.NRGBA, *image.NRGBA) {
	na := ToNRGBA(a)
	nb := ToNRGBA(b)

	w := na.Bounds().Dx()
	h := na.Bounds().Dy()
	if nb.Bounds().Dx() > w {
		w = nb.Bounds().Dx()
	}
	if nb.Bounds().Dy() > h {
		h = nb.Bounds().Dy()
	}

	bg := &image.Uniform{C: color.NRGBA{R: 255, G: 255, B: 255, A: 255}}
	outA := image.NewNRGBA(image.Rect(0, 0, w, h))
	outB := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(outA, outA.Bounds(), bg, image.Point{}, draw.Src)
	draw.Draw(outB, outB.Bounds(), bg, image.Point{}, draw.Src)
	draw.Draw(outA, na.Bounds(), na, image.Point{}, draw.Over)
	draw.Draw(outB, nb.Bounds(), nb, image.Point{}, draw.Over)
	return outA, outB
}

func DiffImages(left, right image.Image, opts PixelDiffOptions) (PixelDiffResult, *image.NRGBA, *image.NRGBA, *image.NRGBA, error) {
	if err := ValidatePixelThreshold(opts.Threshold); err != nil {
		return PixelDiffResult{}, nil, nil, nil, err
	}
	nLeft, nRight := PadToSameSize(left, right)
	result, diffOnly := ComputePixelDiff(nLeft, nRight, opts.Threshold)
	return result, diffOnly, nLeft, nRight, nil
}

func DiffPNGFiles(leftPath, rightPath string, opts PixelDiffOptions) (PixelDiffResult, *image.NRGBA, *image.NRGBA, *image.NRGBA, error) {
	left, err := ReadPNG(leftPath)
	if err != nil {
		return PixelDiffResult{}, nil, nil, nil, fmt.Errorf("read left PNG %q: %w", leftPath, err)
	}
	right, err := ReadPNG(rightPath)
	if err != nil {
		return PixelDiffResult{}, nil, nil, nil, fmt.Errorf("read right PNG %q: %w", rightPath, err)
	}
	return DiffImages(left, right, opts)
}

func ComputePixelDiff(left, right *image.NRGBA, threshold int) (PixelDiffResult, *image.NRGBA) {
	w := left.Bounds().Dx()
	h := left.Bounds().Dy()
	if right.Bounds().Dx() != w || right.Bounds().Dy() != h {
		if right.Bounds().Dx() < w {
			w = right.Bounds().Dx()
		}
		if right.Bounds().Dy() < h {
			h = right.Bounds().Dy()
		}
	}

	thr2 := threshold * threshold
	changed := 0
	total := w * h

	overlay := image.NewNRGBA(image.Rect(0, 0, w, h))
	copy(overlay.Pix, right.Pix)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*left.Stride + x*4
			r1 := int(left.Pix[i+0])
			g1 := int(left.Pix[i+1])
			b1 := int(left.Pix[i+2])

			r2 := int(right.Pix[i+0])
			g2 := int(right.Pix[i+1])
			b2 := int(right.Pix[i+2])

			dr := r1 - r2
			dg := g1 - g2
			db := b1 - b2
			mag2 := dr*dr + dg*dg + db*db
			if mag2 > thr2 {
				changed++
				overlay.Pix[i+0] = 255
				overlay.Pix[i+1] = 0
				overlay.Pix[i+2] = 0
				overlay.Pix[i+3] = 255
			}
		}
	}

	percent := 0.0
	if total > 0 {
		percent = (float64(changed) / float64(total)) * 100
	}

	return PixelDiffResult{
		Threshold:        threshold,
		TotalPixels:      total,
		ChangedPixels:    changed,
		ChangedPercent:   percent,
		NormalizedWidth:  w,
		NormalizedHeight: h,
	}, overlay
}

func CombineSideBySide(left, right, diff *image.NRGBA) *image.NRGBA {
	w := left.Bounds().Dx()
	h := left.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w*3, h))
	draw.Draw(dst, image.Rect(0, 0, w, h), left, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(w, 0, w*2, h), right, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(w*2, 0, w*3, h), diff, image.Point{}, draw.Src)
	return dst
}

func WritePixelDiffImages(leftPath, rightPath, diffComparisonPath, diffOnlyPath string, opts PixelDiffOptions) (PixelDiffResult, error) {
	result, diffOnly, nLeft, nRight, err := DiffPNGFiles(leftPath, rightPath, opts)
	if err != nil {
		return PixelDiffResult{}, err
	}
	result.DiffComparisonPath = diffComparisonPath
	result.DiffOnlyPath = diffOnlyPath

	if err := WritePNG(diffOnlyPath, diffOnly); err != nil {
		return PixelDiffResult{}, fmt.Errorf("write diff-only PNG %q: %w", diffOnlyPath, err)
	}

	diffComparison := CombineSideBySide(nLeft, nRight, diffOnly)
	if err := WritePNG(diffComparisonPath, diffComparison); err != nil {
		return PixelDiffResult{}, fmt.Errorf("write diff-comparison PNG %q: %w", diffComparisonPath, err)
	}

	return result, nil
}
