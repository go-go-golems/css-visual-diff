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

## JS-first visual diff workflows

The old native YAML runner has been removed. For project-scale workflows, write JavaScript verbs and load any project-specific YAML or JSON data from userland with `cvd.objectFromFile()` or normal script helpers. This keeps orchestration, preparation, capture, comparison, and review-site output in one programmable surface.

Use the built-in direct commands for focused one-off work:

```bash
GOWORK=off go run ./cmd/css-visual-diff compare \
  --url1 http://localhost:7070/prototype.html \
  --selector1 '#capture-root' \
  --url2 http://localhost:6006/iframe.html?id=button--primary \
  --selector2 '[data-component="button"]' \
  --out /tmp/cssvd/button-primary

GOWORK=off go run ./cmd/css-visual-diff verbs --help
```

For larger suites, expose a verb under `css-visual-diff verbs ...`. See `examples/verbs/review-sweep.js` for a JS-managed review-site sweep that reads a project spec, performs comparisons, and emits `summary.json`.
