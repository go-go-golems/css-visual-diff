package service

import (
	"context"
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
