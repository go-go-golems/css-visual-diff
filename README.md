# css-visual-diff

`css-visual-diff` is a Go CLI for comparing rendered HTML/CSS across two browser targets.
It is being rebuilt from the proven `sbcap` comparison engine and now uses a standard
`go-go-golems` project layout.

## Current shape

The live tool now centers on:

- browser-driven capture via `chromedp`
- element-level compare workflows
- computed CSS diffs
- matched-style / cascade inspection
- pixel diff artifacts

The previous Python prototype has been preserved under:

- `legacy/python-prototype/`

## Development

```bash
GOWORK=off go mod tidy
GOWORK=off go test ./...
GOWORK=off go build ./cmd/css-visual-diff
```

## CLI

```bash
GOWORK=off go run ./cmd/css-visual-diff --help
GOWORK=off go run ./cmd/css-visual-diff compare --help
GOWORK=off go run ./cmd/css-visual-diff chromedp-probe --help
```

## Prepared targets

Some design sources do not render the comparison target directly. They may open a
pan/zoom design canvas, prototype shell, or development board first. Config-driven
runs can now prepare each target after navigation and before capture/CSS analysis.

Minimal script prepare:

```yaml
original:
  name: prototype
  url: http://localhost:7070/Pyxis%20Public%20Site.html
  wait_ms: 1000
  viewport: { width: 1200, height: 2200 }
  root_selector: "#capture-root"
  prepare:
    type: script
    wait_for: "window.React && window.ReactDOM && window.PPXDesktop"
    script: |
      document.body.innerHTML = '<div id="capture-root"></div>';
      document.body.style.margin = '0';
      document.body.style.background = '#fff';
      const root = document.getElementById('capture-root');
      root.style.width = '920px';
      ReactDOM.createRoot(root).render(React.createElement(PPXDesktop, { page: 'shows' }));
    after_wait_ms: 1000
```

Convenience React-global prepare:

```yaml
original:
  name: prototype
  url: http://localhost:7070/Pyxis%20Public%20Site.html
  viewport: { width: 1200, height: 2200 }
  prepare:
    type: direct-react-global
    wait_for: "window.React && window.ReactDOM && window.PPXDesktop"
    component: PPXDesktop
    props: { page: shows }
    root_selector: "#capture-root"
    width: 920
    background: "#fff"
```

When `root_selector` is set, capture mode uses an element screenshot for the full
baseline PNG instead of a browser full-page screenshot. This avoids including
prototype shell or design-canvas chrome in the exported baseline.

Optional capture artifacts and validation:

```yaml
sections:
  - name: full
    selector_original: "#capture-root"
    selector_react: "[data-page='shows']"
    expect_text_original:
      includes: ["ppxis", "Upcoming shows", "Instagram"]
      excludes: ["01 · Desktop", "Poster-grid shell"]
    expect_png_original:
      width: 920
      min_height: 1700
      top_strip_near: { rgb: [255, 255, 255], tolerance: 8 }
      top_strip_not_near: { rgb: [240, 238, 233], tolerance: 8 }

output:
  dir: ./out/pyxis-public-shows
  write_json: true
  write_markdown: true
  write_pngs: true
  write_prepared_html: true
  write_inspect_json: true
  validate_pngs: true
```

To generate a static artifact browser for the run, include `html-report` after the
modes that produce artifacts:

```bash
GOWORK=off go run ./cmd/css-visual-diff run \
  --config examples/pyxis-public-shows.yaml \
  --modes capture,cssdiff,matched-styles,pixeldiff,html-report
```

The report is written to:

```text
<output.dir>/index.html
```

### Prototype-only inspection

The current config shape still requires both `original` and `react` targets because
`css-visual-diff` is a comparator. To inspect only a prototype/export target, point
both targets at the same URL and use the same prepare hook, then run only capture and
report modes:

```bash
GOWORK=off go run ./cmd/css-visual-diff run \
  --config examples/pyxis-prototype-only.yaml \
  --modes capture,html-report
```

In that workflow the second target is just a mirror so the report renderer can reuse
its existing two-column UI. Do not interpret pixel diffs from a prototype-only run.
Use it to validate the prepared DOM, PNGs, prepared HTML, inspect JSON, and capture
validation before comparing against Storybook.

Validation is intentionally layered: inspect DOM text/selectors first, then PNG
structure and color-strip statistics. Use human visual review or `understand_image`
for semantic questions such as cutoff/footer visibility; OCR should be a later
fallback, not the first validation tool.
