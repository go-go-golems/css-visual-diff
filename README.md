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

## JavaScript verbs and programmable catalogs

`css-visual-diff` can scan annotated JavaScript files and expose them as CLI verbs under the lazy `verbs` namespace:

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs --help
GOWORK=off go run ./cmd/css-visual-diff verbs catalog inspect-page --help
```

Built-in workflow verbs include:

```bash
# Compare one region through the script-backed compare workflow.
GOWORK=off go run ./cmd/css-visual-diff verbs script compare region \
  --leftUrl http://localhost:3000/original \
  --rightUrl http://localhost:3000/react \
  --leftSelector '#cta' \
  --outDir /tmp/cssvd-compare

# Inspect one URL/selector into a Go-backed catalog manifest and index.
GOWORK=off go run ./cmd/css-visual-diff verbs catalog inspect-page \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-page \
  --slug cta \
  --artifacts css-json \
  --output json

# Inspect selectors from an existing YAML config into a catalog.
GOWORK=off go run ./cmd/css-visual-diff verbs catalog inspect-config \
  examples/pyxis-atoms-prototype-vs-storybook.yaml react /tmp/cssvd-config \
  --artifacts css-json \
  --output json
```

External verb repositories can be supplied at runtime:

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-example \
  --output json
```

For repeatable project workflows, check in a local repository config:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./verbs
```

Relative paths are resolved from the config file that declares them, so `./verbs` still points at the project repository when commands are run from nested directories. Use `.css-visual-diff.override.yml` for private local repositories and keep it gitignored.

JavaScript verbs can use the Promise-first native module:

```js
const cvd = require("css-visual-diff")
const browser = await cvd.browser()
const page = await browser.page(url, { viewport: { width: 1280, height: 720 } })
const statuses = await page.preflight([{ name: "cta", selector: "#cta" }])
const result = await page.inspectAll([{ name: "cta", selector: "#cta", props: ["color"] }], {
  outDir: "/tmp/cssvd/artifacts/cta",
  artifacts: "css-json",
})
```

See the embedded Glazed help entries:

```bash
css-visual-diff help javascript-api
css-visual-diff help javascript-verbs
css-visual-diff help pixel-accuracy-scripting-guide
```

These entries cover `require("css-visual-diff")`, Promise behavior, locators, extraction, snapshots, diffs, typed errors, catalog APIs, `__verb__`, repositories, generated flags, output modes, duplicate command paths, and migration notes.

Migration note: generated JavaScript commands are intentionally no longer injected at the CLI root. Use `css-visual-diff verbs ...` so repository scan errors and duplicate script command paths remain scoped to the verbs subtree.

## Run co-located configs from a directory

You can keep small comparison configs next to the component or page they cover and
run all of them through the existing `run` verb:

```bash
GOWORK=off go run ./cmd/css-visual-diff run \
  --config-dir web/packages/pyxis-components/src \
  --dry-run
```

`--config-dir` scans recursively for:

```text
*.css-visual-diff.yml
*.css-visual-diff.yaml
```

It intentionally does not load every YAML file, and it skips common generated/vendor
directories such as `node_modules`, `.git`, `dist`, `build`, and `.css-visual-diff`.
Use explicit `--config path/to/file.yaml` when you want to run one config file.

## Inspect one side before comparing

When tuning a `*.css-visual-diff.yml` file, first inspect one side and one selector
before running a full comparison. This helps verify that the URL, viewport, prepare
hook, root selector, section selector, and CSS probe selector are correct.

A config can contain many screenshot regions under `sections` and many computed-CSS
probes under `styles`. Inspect a named CSS probe and write a bundle of artifacts:

```bash
GOWORK=off go run ./cmd/css-visual-diff inspect \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --out /tmp/css-visual-diff-inspect/button-primary
```

The bundle contains:

```text
metadata.json
screenshot.png
prepared.html
computed-css.json
computed-css.md
inspect.json
```

For tight selector-tuning loops, use the single-artifact verbs:

```bash
# Check the crop/element screenshot.
GOWORK=off go run ./cmd/css-visual-diff screenshot \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary.png

# Check computed CSS in human-readable Markdown.
GOWORK=off go run ./cmd/css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary-css.md

# Debug prepared DOM/root output.
GOWORK=off go run ./cmd/css-visual-diff html \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side original \
  --root \
  --output-file /tmp/original-root.html
```

Available single-artifact verbs are `screenshot`, `css-md`, `css-json`, `html`, and
`inspect-json`. They share the same selector flags as `inspect`: `--root`,
`--section`, `--style`, or `--selector`.

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
