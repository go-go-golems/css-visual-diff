package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

type htmlReportData struct {
	Capture       *CaptureResult
	CSSDiff       *CSSDiffResult
	MatchedStyles *MatchedStylesResult
	PixelDiff     *PixelDiffResult
	Files         []reportFile
}

type reportFile struct {
	Name string
	Path string
	Kind string
}

func HTMLReport(ctx context.Context, cfg *config.Config) error {
	if err := os.MkdirAll(cfg.Output.Dir, 0o755); err != nil {
		return err
	}
	data := loadHTMLReportData(cfg.Output.Dir)
	content := renderHTMLReport(cfg, data)
	_ = ctx
	return os.WriteFile(filepath.Join(cfg.Output.Dir, "index.html"), []byte(content), 0o644)
}

func loadHTMLReportData(outDir string) htmlReportData {
	data := htmlReportData{}
	if v, ok := readJSONFile[CaptureResult](filepath.Join(outDir, "capture.json")); ok {
		data.Capture = &v
	}
	if v, ok := readJSONFile[CSSDiffResult](filepath.Join(outDir, "cssdiff.json")); ok {
		data.CSSDiff = &v
	}
	if v, ok := readJSONFile[MatchedStylesResult](filepath.Join(outDir, "matched-styles.json")); ok {
		data.MatchedStyles = &v
	}
	if v, ok := readJSONFile[PixelDiffResult](filepath.Join(outDir, "pixeldiff.json")); ok {
		data.PixelDiff = &v
	}
	data.Files = collectReportFiles(outDir)
	return data
}

func readJSONFile[T any](path string) (T, bool) {
	var zero T
	b, err := os.ReadFile(path)
	if err != nil {
		return zero, false
	}
	if err := json.Unmarshal(b, &zero); err != nil {
		return zero, false
	}
	return zero, true
}

func collectReportFiles(outDir string) []reportFile {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil
	}
	files := []reportFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		kind := "other"
		switch strings.ToLower(filepath.Ext(name)) {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif":
			kind = "image"
		case ".json":
			kind = "json"
		case ".md":
			kind = "markdown"
		case ".html", ".htm":
			kind = "html"
		}
		files = append(files, reportFile{Name: name, Path: name, Kind: kind})
	}
	return files
}

func renderHTMLReport(cfg *config.Config, data htmlReportData) string {
	var b strings.Builder
	b.WriteString(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>css-visual-diff report</title>
<style>
:root { color-scheme: light; --bg: #f6f4ef; --panel: #fffdfa; --ink: #191815; --muted: #69645d; --line: #d9d2c5; --accent: #c8270d; --ok: #116b3a; --bad: #a21717; }
* { box-sizing: border-box; }
body { margin: 0; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--ink); }
header { padding: 28px 32px 18px; border-bottom: 1px solid var(--line); background: #fffdfa; position: sticky; top: 0; z-index: 5; }
h1 { margin: 0 0 8px; font-size: 24px; letter-spacing: -0.02em; }
.meta { color: var(--muted); font-size: 13px; display: flex; gap: 14px; flex-wrap: wrap; }
main { padding: 24px 32px 56px; max-width: 1800px; margin: 0 auto; }
nav { display: flex; gap: 8px; flex-wrap: wrap; margin: 12px 0 0; }
nav a { color: var(--ink); text-decoration: none; border: 1px solid var(--line); border-radius: 999px; padding: 6px 10px; font-size: 13px; background: #fff; }
section { margin: 0 0 28px; padding: 18px; background: var(--panel); border: 1px solid var(--line); border-radius: 16px; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
h2 { margin: 0 0 14px; font-size: 19px; }
h3 { margin: 0 0 10px; font-size: 15px; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 14px; }
.card { border: 1px solid var(--line); border-radius: 12px; padding: 12px; background: #fff; overflow: hidden; }
.card img { display: block; max-width: 100%; height: auto; border: 1px solid var(--line); border-radius: 8px; background: white; }
.card .caption { font-size: 12px; color: var(--muted); margin: 8px 0 0; overflow-wrap: anywhere; }
table { width: 100%; border-collapse: collapse; font-size: 13px; }
th, td { border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; padding: 8px; }
th { color: var(--muted); font-weight: 650; }
pre { max-height: 360px; overflow: auto; padding: 12px; background: #201f1d; color: #f7f3ea; border-radius: 10px; font-size: 12px; line-height: 1.45; }
.badge { display: inline-block; border-radius: 999px; padding: 2px 8px; font-size: 12px; border: 1px solid var(--line); }
.ok { color: var(--ok); border-color: rgba(17,107,58,.35); background: rgba(17,107,58,.08); }
.bad { color: var(--bad); border-color: rgba(162,23,23,.35); background: rgba(162,23,23,.08); }
.files { columns: 2 280px; }
.files li { break-inside: avoid; margin: 0 0 6px; }
a { color: var(--accent); }
.small { color: var(--muted); font-size: 12px; }
</style>
</head>
<body>
<header>
<h1>css-visual-diff artifact browser</h1>
<div class="meta">`)
	b.WriteString(fmt.Sprintf("<span>slug: <strong>%s</strong></span>", esc(cfg.Metadata.Slug)))
	b.WriteString(fmt.Sprintf("<span>output: <code>%s</code></span>", esc(cfg.Output.Dir)))
	if cfg.Original.URL != "" {
		b.WriteString(fmt.Sprintf("<span>original: <code>%s</code></span>", esc(cfg.Original.URL)))
	}
	if cfg.React.URL != "" {
		b.WriteString(fmt.Sprintf("<span>react: <code>%s</code></span>", esc(cfg.React.URL)))
	}
	b.WriteString(`</div>
<nav><a href="#screenshots">Screenshots</a><a href="#validation">Validation</a><a href="#cssdiff">CSS diff</a><a href="#matched">Matched styles</a><a href="#files">Files</a></nav>
</header>
<main>
`)

	renderScreenshotSection(&b, data)
	renderValidationSection(&b, data)
	renderCSSDiffSection(&b, data)
	renderMatchedStylesSection(&b, data)
	renderFilesSection(&b, data)

	b.WriteString(`</main>
</body>
</html>
`)
	return b.String()
}

func renderScreenshotSection(b *strings.Builder, data htmlReportData) {
	b.WriteString(`<section id="screenshots"><h2>Screenshots and pixel diffs</h2>`)
	if data.Capture == nil {
		b.WriteString(`<p>No capture.json found yet.</p></section>`)
		return
	}
	b.WriteString(`<div class="grid">`)
	renderImageCard(b, "Original full", rel(data.Capture.Original.FullScreenshot), "Full/root screenshot for the original target")
	renderImageCard(b, "React full", rel(data.Capture.React.FullScreenshot), "Full/root screenshot for the React target")
	b.WriteString(`</div>`)
	for i, original := range data.Capture.Original.Sections {
		var react SectionResult
		if i < len(data.Capture.React.Sections) {
			react = data.Capture.React.Sections[i]
		}
		b.WriteString(fmt.Sprintf(`<h3>%s</h3><div class="grid">`, esc(original.Name)))
		renderImageCard(b, "Original", rel(original.Screenshot), original.Selector)
		renderImageCard(b, "React", rel(react.Screenshot), react.Selector)
		if diff := pixelDiffForSection(data.PixelDiff, original.Name); diff != nil && diff.DiffComparisonPath != "" {
			renderImageCard(b, "Pixel diff comparison", rel(diff.DiffComparisonPath), fmt.Sprintf("changed %.4f%% (%d/%d)", diff.ChangedPercent, diff.ChangedPixels, diff.TotalPixels))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</section>`)
}

func renderValidationSection(b *strings.Builder, data htmlReportData) {
	b.WriteString(`<section id="validation"><h2>Validation</h2>`)
	if data.Capture == nil || len(data.Capture.Validation) == 0 {
		b.WriteString(`<p>No validation entries found.</p></section>`)
		return
	}
	b.WriteString(`<table><thead><tr><th>Target</th><th>Section</th><th>Status</th><th>PNG</th><th>Issues</th></tr></thead><tbody>`)
	for _, v := range data.Capture.Validation {
		class := "ok"
		if v.Status != "ok" {
			class = "bad"
		}
		png := ""
		if v.PNG != nil {
			png = fmt.Sprintf("%dx%d<br><span class=\"small\">top %v<br>bottom %v</span>", v.PNG.Width, v.PNG.Height, v.PNG.TopStripAverage, v.PNG.BottomStripAverage)
		}
		b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td><span class="badge %s">%s</span></td><td>%s</td><td>%s</td></tr>`, esc(v.Target), esc(v.Section), class, esc(v.Status), png, esc(strings.Join(v.Issues, "; "))))
	}
	b.WriteString(`</tbody></table></section>`)
}

func renderCSSDiffSection(b *strings.Builder, data htmlReportData) {
	b.WriteString(`<section id="cssdiff"><h2>Computed CSS diff</h2>`)
	if data.CSSDiff == nil || len(data.CSSDiff.Styles) == 0 {
		b.WriteString(`<p>No cssdiff.json found yet.</p></section>`)
		return
	}
	for _, style := range data.CSSDiff.Styles {
		b.WriteString(fmt.Sprintf(`<h3>%s</h3><p class="small">original: <code>%s</code> · react: <code>%s</code></p>`, esc(style.Name), esc(style.OriginalSelector), esc(style.ReactSelector)))
		if len(style.Diffs) == 0 {
			b.WriteString(`<p><span class="badge ok">no configured property diffs</span></p>`)
			continue
		}
		b.WriteString(`<table><thead><tr><th>Property</th><th>Original</th><th>React</th></tr></thead><tbody>`)
		for _, d := range style.Diffs {
			b.WriteString(fmt.Sprintf(`<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>`, esc(d.Property), esc(d.Original), esc(d.React)))
		}
		b.WriteString(`</tbody></table>`)
	}
	b.WriteString(`</section>`)
}

func renderMatchedStylesSection(b *strings.Builder, data htmlReportData) {
	b.WriteString(`<section id="matched"><h2>Matched style winners</h2>`)
	if data.MatchedStyles == nil || len(data.MatchedStyles.Styles) == 0 {
		b.WriteString(`<p>No matched-styles.json found yet.</p></section>`)
		return
	}
	for _, style := range data.MatchedStyles.Styles {
		b.WriteString(fmt.Sprintf(`<h3>%s</h3>`, esc(style.Name)))
		if len(style.Winners) == 0 {
			b.WriteString(`<p>No winner diffs recorded.</p>`)
			continue
		}
		b.WriteString(`<table><thead><tr><th>Property</th><th>Original winner</th><th>React winner</th></tr></thead><tbody>`)
		for _, w := range style.Winners {
			b.WriteString(fmt.Sprintf(`<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>`, esc(w.Property), esc(formatReportWinner(w.Original)), esc(formatReportWinner(w.React))))
		}
		b.WriteString(`</tbody></table>`)
	}
	b.WriteString(`</section>`)
}

func renderFilesSection(b *strings.Builder, data htmlReportData) {
	b.WriteString(`<section id="files"><h2>All generated files</h2><ul class="files">`)
	for _, f := range data.Files {
		if f.Name == "index.html" {
			continue
		}
		b.WriteString(fmt.Sprintf(`<li><span class="badge">%s</span> <a href="%s">%s</a></li>`, esc(f.Kind), esc(f.Path), esc(f.Name)))
	}
	b.WriteString(`</ul></section>`)
}

func renderImageCard(b *strings.Builder, title, path, caption string) {
	b.WriteString(`<div class="card">`)
	b.WriteString(fmt.Sprintf(`<h3>%s</h3>`, esc(title)))
	if strings.TrimSpace(path) == "" {
		b.WriteString(`<p class="small">No image.</p>`)
	} else {
		b.WriteString(fmt.Sprintf(`<a href="%s"><img src="%s" loading="lazy" /></a>`, esc(path), esc(path)))
	}
	if caption != "" {
		b.WriteString(fmt.Sprintf(`<div class="caption">%s</div>`, esc(caption)))
	}
	b.WriteString(`</div>`)
}

func pixelDiffForSection(result *PixelDiffResult, name string) *PixelDiffEntry {
	if result == nil {
		return nil
	}
	for i := range result.Entries {
		if result.Entries[i].Section == name {
			return &result.Entries[i]
		}
	}
	return nil
}

func formatReportWinner(w Winner) string {
	parts := []string{}
	if w.Selector != "" {
		parts = append(parts, w.Selector)
	}
	if w.Value != "" {
		parts = append(parts, "value: "+w.Value)
	}
	if w.Important {
		parts = append(parts, "!important")
	}
	if w.Origin != "" {
		parts = append(parts, "origin: "+string(w.Origin))
	}
	return strings.Join(parts, " · ")
}

func rel(path string) string {
	return filepath.Base(path)
}

func esc(s string) string {
	return html.EscapeString(s)
}
