package modes

import (
	"image"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

func readPNG(path string) (image.Image, error) {
	return service.ReadPNG(path)
}

func writePNG(path string, img image.Image) error {
	return service.WritePNG(path, img)
}

func toNRGBA(img image.Image) *image.NRGBA {
	return service.ToNRGBA(img)
}

func computePixelDiff(url1, url2 *image.NRGBA, threshold int) (PixelDiffStats, *image.NRGBA) {
	result, overlay := service.ComputePixelDiff(url1, url2, threshold)
	return pixelDiffStatsFromService(result), overlay
}

func pixelDiffStatsFromService(result service.PixelDiffResult) PixelDiffStats {
	return PixelDiffStats{
		Threshold:          result.Threshold,
		TotalPixels:        result.TotalPixels,
		ChangedPixels:      result.ChangedPixels,
		ChangedPercent:     result.ChangedPercent,
		NormalizedWidth:    result.NormalizedWidth,
		NormalizedHeight:   result.NormalizedHeight,
		DiffComparisonPath: result.DiffComparisonPath,
		DiffOnlyPath:       result.DiffOnlyPath,
	}
}
