package modes

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func readPNG(path string) (image.Image, error) {
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

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func toNRGBA(img image.Image) *image.NRGBA {
	if n, ok := img.(*image.NRGBA); ok {
		return n
	}
	b := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

func padToSameSize(a, b image.Image) (*image.NRGBA, *image.NRGBA) {
	na := toNRGBA(a)
	nb := toNRGBA(b)

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

func computePixelDiff(url1, url2 *image.NRGBA, threshold int) (PixelDiffStats, *image.NRGBA) {
	w := url1.Bounds().Dx()
	h := url1.Bounds().Dy()
	if url2.Bounds().Dx() != w || url2.Bounds().Dy() != h {
		// Should not happen if padded.
		if url2.Bounds().Dx() < w {
			w = url2.Bounds().Dx()
		}
		if url2.Bounds().Dy() < h {
			h = url2.Bounds().Dy()
		}
	}

	thr2 := threshold * threshold
	changed := 0
	total := w * h

	overlay := image.NewNRGBA(image.Rect(0, 0, w, h))
	copy(overlay.Pix, url2.Pix)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*url1.Stride + x*4
			r1 := int(url1.Pix[i+0])
			g1 := int(url1.Pix[i+1])
			b1 := int(url1.Pix[i+2])

			r2 := int(url2.Pix[i+0])
			g2 := int(url2.Pix[i+1])
			b2 := int(url2.Pix[i+2])

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

	return PixelDiffStats{
		Threshold:        threshold,
		TotalPixels:      total,
		ChangedPixels:    changed,
		ChangedPercent:   percent,
		NormalizedWidth:  w,
		NormalizedHeight: h,
	}, overlay
}

func combineSideBySide(url1, url2, diff *image.NRGBA) *image.NRGBA {
	w := url1.Bounds().Dx()
	h := url1.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w*3, h))
	draw.Draw(dst, image.Rect(0, 0, w, h), url1, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(w, 0, w*2, h), url2, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(w*2, 0, w*3, h), diff, image.Point{}, draw.Src)
	return dst
}
