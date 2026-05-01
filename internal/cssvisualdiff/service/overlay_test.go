package service

import (
	"context"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayScreenshotWritesAnnotatedPNG(t *testing.T) {
	browser, err := NewBrowserService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	defer page.Close()

	html := `<!doctype html><html><head><style>body{margin:0;font-family:sans-serif}.hero{margin:40px;width:220px;height:120px;background:#eee}.cta{display:inline-block;margin:20px;padding:12px;background:#333;color:white}</style></head><body><section class="hero"><h1>Hello</h1><a class="cta">Go</a></section></body></html>`
	if err := page.Page().SetViewport(400, 300); err != nil {
		t.Fatal(err)
	}
	if err := page.Page().Goto("data:text/html," + url.PathEscape(html)); err != nil {
		t.Fatal(err)
	}

	red, _ := ParseColor("#ff6347")
	out := filepath.Join(t.TempDir(), "annotated.png")
	result, err := OverlayScreenshot(page.Page(), OverlaySpec{
		Legend:  true,
		Targets: []OverlayTarget{{Name: "Hero", Selector: ".hero", Style: TargetOverlayStyle{BorderColor: &red}}},
	}, out)
	if err != nil {
		t.Fatal(err)
	}
	if result.OutputPath != out {
		t.Fatalf("expected output path %q, got %q", out, result.OutputPath)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non-empty png")
	}
	if result.Colors["Hero"] != "#ff6347" {
		t.Fatalf("expected hero color, got %#v", result.Colors)
	}
}

func TestOverlayScreenshotCropSelectorPaddingAndFiltering(t *testing.T) {
	browser, err := NewBrowserService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	defer page.Close()

	html := `<!doctype html><html><head><style>body{margin:0;height:800px}.hero{margin:40px;width:220px;height:120px;background:#eee}.other{position:absolute;left:20px;top:360px;width:80px;height:60px;background:#ccc}</style></head><body><section class="hero"><h1>Hello</h1></section><div class="other">Other</div></body></html>`
	if err := page.Page().SetViewport(400, 500); err != nil {
		t.Fatal(err)
	}
	if err := page.Page().Goto("data:text/html," + url.PathEscape(html)); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "hero-crop.png")
	result, err := OverlayScreenshot(page.Page(), OverlaySpec{
		Legend: false,
		Crop:   &OverlayCrop{Selector: ".hero", Padding: Insets{Top: 10, Right: 10, Bottom: 10, Left: 10}},
		Targets: []OverlayTarget{
			{Name: "Hero", Selector: ".hero"},
			{Name: "Other", Selector: ".other"},
		},
	}, out)
	if err != nil {
		t.Fatal(err)
	}
	if result.Width != 240 || result.Height != 140 {
		t.Fatalf("expected crop dimensions 240x140, got %dx%d", result.Width, result.Height)
	}
	if len(result.Targets) != 1 || result.Targets[0].Name != "Hero" {
		t.Fatalf("expected only Hero target in crop result, got %#v", result.Targets)
	}
	if _, ok := result.Colors["Other"]; ok {
		t.Fatalf("expected outside target Other to be omitted from colors: %#v", result.Colors)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != result.Width || img.Bounds().Dy() != result.Height {
		t.Fatalf("png dimensions do not match result")
	}
}

func TestOverlayScreenshotCropMissingSelector(t *testing.T) {
	browser, err := NewBrowserService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	defer page.Close()

	if err := page.Page().SetViewport(200, 200); err != nil {
		t.Fatal(err)
	}
	if err := page.Page().Goto("data:text/html," + url.PathEscape(`<!doctype html><div class="hero">Hero</div>`)); err != nil {
		t.Fatal(err)
	}

	_, err = OverlayScreenshot(page.Page(), OverlaySpec{
		Crop:    &OverlayCrop{Selector: ".missing"},
		Targets: []OverlayTarget{{Name: "Hero", Selector: ".hero"}},
	}, filepath.Join(t.TempDir(), "missing.png"))
	if err == nil {
		t.Fatalf("expected missing crop selector error")
	}
}
