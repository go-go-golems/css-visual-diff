package service

import (
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffImagesIdentical(t *testing.T) {
	left := solidNRGBA(3, 2, color.NRGBA{R: 42, G: 42, B: 42, A: 255})
	right := solidNRGBA(3, 2, color.NRGBA{R: 42, G: 42, B: 42, A: 255})

	result, diffOnly, nLeft, nRight, err := DiffImages(left, right, PixelDiffOptions{Threshold: 0})
	require.NoError(t, err)
	require.Equal(t, 6, result.TotalPixels)
	require.Equal(t, 0, result.ChangedPixels)
	require.Equal(t, 0.0, result.ChangedPercent)
	require.Equal(t, 3, result.NormalizedWidth)
	require.Equal(t, 2, result.NormalizedHeight)
	require.NotNil(t, diffOnly)
	require.Equal(t, nLeft.Bounds(), nRight.Bounds())
}

func TestDiffImagesChangedPixelAndThreshold(t *testing.T) {
	left := solidNRGBA(2, 2, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	right := solidNRGBA(2, 2, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	right.Pix[0] = 255
	right.Pix[1] = 255
	right.Pix[2] = 255
	right.Pix[3] = 255

	result, diffOnly, _, _, err := DiffImages(left, right, PixelDiffOptions{Threshold: 30})
	require.NoError(t, err)
	require.Equal(t, 4, result.TotalPixels)
	require.Equal(t, 1, result.ChangedPixels)
	require.Equal(t, 25.0, result.ChangedPercent)
	require.Equal(t, uint8(255), diffOnly.Pix[0])
	require.Equal(t, uint8(0), diffOnly.Pix[1])
	require.Equal(t, uint8(0), diffOnly.Pix[2])
	require.Equal(t, uint8(255), diffOnly.Pix[3])

	belowThreshold, _, _, _, err := DiffImages(left, right, PixelDiffOptions{Threshold: 255})
	require.NoError(t, err)
	require.Equal(t, 1, belowThreshold.ChangedPixels, "3-channel distance is still above 255^2")

	smallDelta := solidNRGBA(1, 1, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	smallDelta.Pix[0] = 3
	smallDelta.Pix[1] = 4
	ignored, _, _, _, err := DiffImages(solidNRGBA(1, 1, color.NRGBA{A: 255}), smallDelta, PixelDiffOptions{Threshold: 5})
	require.NoError(t, err)
	require.Equal(t, 0, ignored.ChangedPixels, "strictly greater than threshold squared changes pixels")
}

func TestDiffImagesDifferentSizesNormalize(t *testing.T) {
	left := solidNRGBA(1, 1, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	right := solidNRGBA(2, 3, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	result, _, nLeft, nRight, err := DiffImages(left, right, PixelDiffOptions{})
	require.NoError(t, err)
	require.Equal(t, 2, result.NormalizedWidth)
	require.Equal(t, 3, result.NormalizedHeight)
	require.Equal(t, image.Rect(0, 0, 2, 3), nLeft.Bounds())
	require.Equal(t, image.Rect(0, 0, 2, 3), nRight.Bounds())
	require.Equal(t, 6, result.TotalPixels)
}

func TestWritePixelDiffImagesWritesParentsAndJSONShape(t *testing.T) {
	tmp := t.TempDir()
	leftPath := filepath.Join(tmp, "left.png")
	rightPath := filepath.Join(tmp, "right.png")
	require.NoError(t, WritePNG(leftPath, solidNRGBA(2, 2, color.NRGBA{A: 255})))
	right := solidNRGBA(2, 2, color.NRGBA{A: 255})
	right.Pix[0] = 255
	require.NoError(t, WritePNG(rightPath, right))

	diffOnlyPath := filepath.Join(tmp, "nested", "diff_only.png")
	diffComparisonPath := filepath.Join(tmp, "nested", "comparison", "diff_comparison.png")
	result, err := WritePixelDiffImages(leftPath, rightPath, diffComparisonPath, diffOnlyPath, PixelDiffOptions{Threshold: 0})
	require.NoError(t, err)
	require.Equal(t, diffOnlyPath, result.DiffOnlyPath)
	require.Equal(t, diffComparisonPath, result.DiffComparisonPath)
	require.Equal(t, 1, result.ChangedPixels)

	for _, path := range []string{diffOnlyPath, diffComparisonPath} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Greater(t, info.Size(), int64(0))
	}

	payload, err := json.Marshal(result)
	require.NoError(t, err)
	require.Contains(t, string(payload), `"changedPercent"`)
	require.Contains(t, string(payload), `"normalizedWidth"`)
	require.NotContains(t, string(payload), `changed_percent`)
}

func TestValidatePixelThreshold(t *testing.T) {
	require.NoError(t, ValidatePixelThreshold(0))
	require.NoError(t, ValidatePixelThreshold(255))
	require.Error(t, ValidatePixelThreshold(-1))
	require.Error(t, ValidatePixelThreshold(256))
}

func solidNRGBA(width, height int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*img.Stride + x*4
			img.Pix[i+0] = c.R
			img.Pix[i+1] = c.G
			img.Pix[i+2] = c.B
			img.Pix[i+3] = c.A
		}
	}
	return img
}
