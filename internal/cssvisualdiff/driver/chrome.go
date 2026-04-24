package driver

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type Browser struct {
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	browserCtx    context.Context
	browserCancel context.CancelFunc
}

type Page struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func chromeAllocatorOptions() []chromedp.ExecAllocatorOption {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Headless,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}
	if shouldDisableChromeSandbox() {
		log.Debug().Msg("css-visual-diff chromedp: disabling Chrome sandbox")
		opts = append(opts, chromedp.NoSandbox)
	}
	return opts
}

func shouldDisableChromeSandbox() bool {
	if value, ok := os.LookupEnv("CSS_VISUAL_DIFF_CHROME_NO_SANDBOX"); ok {
		return envBool(value)
	}
	if envBool(os.Getenv("CI")) || envBool(os.Getenv("GITHUB_ACTIONS")) {
		return true
	}
	return os.Geteuid() == 0
}

func envBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func NewBrowser(parent context.Context) (*Browser, error) {
	log.Info().Msg("css-visual-diff chromedp: initializing browser")
	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, chromeAllocatorOptions()...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	log.Info().Msg("css-visual-diff chromedp: browser context created")
	return &Browser{
		allocCtx:      allocCtx,
		allocCancel:   allocCancel,
		browserCtx:    browserCtx,
		browserCancel: browserCancel,
	}, nil
}

func (b *Browser) Close() {
	log.Info().Msg("css-visual-diff chromedp: closing browser")
	if b.browserCancel != nil {
		b.browserCancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
}

func (b *Browser) NewPage() (*Page, error) {
	log.Info().Msg("css-visual-diff chromedp: creating page")
	ctx, cancel := chromedp.NewContext(b.browserCtx)
	return &Page{ctx: ctx, cancel: cancel}, nil
}

func (p *Page) Close() {
	log.Info().Msg("css-visual-diff chromedp: closing page")
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *Page) Context() context.Context {
	return p.ctx
}

func (p *Page) SetViewport(width, height int) error {
	log.Info().Int("width", width).Int("height", height).Msg("css-visual-diff chromedp: set viewport")
	if err := chromedp.Run(p.ctx, emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1, false)); err != nil {
		log.Error().Err(err).Msg("css-visual-diff chromedp: set viewport failed")
		return err
	}
	return nil
}

func (p *Page) Goto(url string) error {
	log.Info().Str("url", url).Msg("css-visual-diff chromedp: navigate")
	if err := chromedp.Run(p.ctx, chromedp.Navigate(url)); err != nil {
		log.Error().Err(err).Str("url", url).Msg("css-visual-diff chromedp: navigate failed")
		return err
	}
	return nil
}

func (p *Page) Wait(d time.Duration) {
	log.Info().Dur("duration", d).Msg("css-visual-diff chromedp: wait")
	_ = chromedp.Run(p.ctx, chromedp.Sleep(d))
}

func (p *Page) FullScreenshot(path string) error {
	log.Info().Str("path", path).Msg("css-visual-diff chromedp: full screenshot")
	var buf []byte
	if err := chromedp.Run(p.ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		log.Error().Err(err).Str("path", path).Msg("css-visual-diff chromedp: full screenshot failed")
		return err
	}
	return os.WriteFile(path, buf, 0o644)
}

func (p *Page) Screenshot(selector, path string) error {
	log.Info().Str("selector", selector).Str("path", path).Msg("css-visual-diff chromedp: screenshot")
	var buf []byte
	if err := chromedp.Run(p.ctx, chromedp.Screenshot(selector, &buf, chromedp.ByQuery)); err != nil {
		log.Error().Err(err).Str("selector", selector).Str("path", path).Msg("css-visual-diff chromedp: screenshot failed")
		return err
	}
	return os.WriteFile(path, buf, 0o644)
}

func (p *Page) Evaluate(script string, out any) error {
	log.Info().Msg("css-visual-diff chromedp: evaluate script")
	if err := chromedp.Run(p.ctx, chromedp.Evaluate(script, out)); err != nil {
		log.Error().Err(err).Msg("css-visual-diff chromedp: evaluate failed")
		return err
	}
	return nil
}

func (p *Page) Eval(script string) error {
	var out any
	return p.Evaluate(script, &out)
}

func (p *Page) WaitForFunction(expr string, timeout time.Duration) error {
	if expr == "" {
		return nil
	}

	deadline := time.Now().Add(timeout)
	for {
		var ok bool
		script := fmt.Sprintf(`Boolean(%s)`, expr)
		if err := p.Evaluate(script, &ok); err != nil {
			return fmt.Errorf("evaluate wait_for expression %q: %w", expr, err)
		}
		if ok {
			return nil
		}
		if timeout > 0 && time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for JavaScript expression %q after %s", expr, timeout)
		}
		p.Wait(100 * time.Millisecond)
	}
}
