package modes

import (
	"context"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func Capture(ctx context.Context, cfg *config.Config) error {
	return RunCapture(ctx, cfg)
}

func MatchedStyles(ctx context.Context, cfg *config.Config) error {
	return RunMatchedStyles(ctx, cfg)
}

func AIReview(ctx context.Context, cfg *config.Config) error {
	return RunAIReview(ctx, cfg)
}

func PixelDiff(ctx context.Context, cfg *config.Config, threshold int) error {
	return RunPixelDiff(ctx, cfg.Output.Dir, threshold)
}
