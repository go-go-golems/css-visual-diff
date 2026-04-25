# css-visual-diff

`css-visual-diff` is a browser-driven visual validation toolkit for comparing rendered web pages, component states, and design implementations. It combines a Go CLI, Chromium automation, pixel-diff image artifacts, computed CSS inspection, YAML configuration workflows, and a programmable JavaScript API for project-specific validation scripts.

The project is useful when a screenshot alone is not enough. A screenshot can show that something changed; `css-visual-diff` helps explain what changed by collecting evidence from the browser:

- element existence and visibility,
- selector bounds and layout deltas,
- rendered text,
- computed CSS values,
- attributes and semantic state,
- region screenshots and pixel-diff PNGs,
- JSON and Markdown reports,
- catalog manifests and indexes for multi-section review.

The current implementation is rebuilt from the proven `sbcap` comparison engine and uses the standard `go-go-golems` project layout.

## Core workflows

`css-visual-diff` supports three complementary workflows.

| Workflow | Use it when | Main entry point |
|---|---|---|
| YAML/config comparison | You have repeatable baseline/current targets, sections, prepare steps, and output modes. | `css-visual-diff run --config ...` |
| Inspection and tuning | You need to debug one URL, side, selector, screenshot, or computed CSS probe. | `inspect`, `screenshot`, `css-md`, `css-json`, `html` |
| JavaScript visual scripts | You need project-specific loops, selectors, policies, catalogs, or CI commands. | `css-visual-diff verbs ...` and `require("css-visual-diff")` |

The JavaScript workflow is now the most flexible API surface. It is designed for designers, frontend engineers, and coding agents who want to build an implementation loop: change UI code, run a validation command, inspect structured evidence, fix the implementation, and run again.

## Install and develop

```bash
GOWORK=off go mod tidy
GOWORK=off go test ./...
GOWORK=off go build ./cmd/css-visual-diff
```

Run the CLI during development:

```bash
GOWORK=off go run ./cmd/css-visual-diff --help
GOWORK=off go run ./cmd/css-visual-diff run --help
GOWORK=off go run ./cmd/css-visual-diff inspect --help
GOWORK=off go run ./cmd/css-visual-diff verbs --help
```

## Quick JavaScript region comparison

The fastest programmable path is `cvd.compare.region(...)`. It compares two loaded page regions, captures screenshots, computes a pixel diff, collects selector facts, and returns a `SelectionComparison` handle.

```js
async function compareCTA(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  let leftPage, rightPage

  try {
    leftPage = await browser.page(leftUrl, {
      viewport: { width: 1280, height: 720 },
      waitMs: 250,
      name: "reference",
    })
    rightPage = await browser.page(rightUrl, {
      viewport: { width: 1280, height: 720 },
      waitMs: 250,
      name: "implementation",
    })

    const comparison = await cvd.compare.region({
      name: "primary-cta",
      left: leftPage.locator("[data-testid='primary-cta']"),
      right: rightPage.locator("[data-testid='primary-cta']"),
      outDir,
      threshold: 30,
      styleProps: ["font-size", "line-height", "color", "background-color", "border-radius"],
      attributes: ["class", "aria-label"],
    })

    await comparison.artifacts.write(outDir, ["json", "markdown"])
    return comparison.summary()
  } finally {
    if (leftPage) await leftPage.close()
    if (rightPage) await rightPage.close()
    await browser.close()
  }
}
```

That one comparison can produce:

```text
left_region.png
right_region.png
diff_only.png
diff_comparison.png
compare.json
compare.md
```

`diff_comparison.png` is the review image. `compare.json` is the automation record. `compare.md` is the human-readable explanation.

## JavaScript API mental model

The API is Promise-first and organized around a small object model:

| Object | Meaning | Example |
|---|---|---|
| `Browser` | Chromium-backed browser service. | `const browser = await cvd.browser()` |
| `Page` | One loaded browser page. | `await browser.page(url, options)` |
| `Locator` | A selector bound to one page. | `page.locator("#cta")` |
| `CollectedSelection` | Stable browser evidence for one selector. | `await locator.collect()` |
| `SelectionComparison` | Analysis of two collected selections. | `await cvd.compare.selections(left, right)` |
| `Catalog` | Durable manifest/index for many artifacts and comparisons. | `cvd.catalog.create(...)` |

A locator is live browser state. A collected selection is immutable evidence. A comparison is deterministic analysis over collected evidence. This separation makes scripts easier to reason about and easier to debug.

Canonical namespaces:

```js
cvd.collect.selection(locator, options?)
cvd.compare.region({ left, right, ... })
cvd.compare.selections(leftSelection, rightSelection, options?)
cvd.image.diff(options)
cvd.diff.structural(before, after, options?)
cvd.snapshot.page(page, probes, options?)
cvd.catalog.create(options)
cvd.config.load(path)
```

Use `summary()` for compact CLI output and `toJSON()` for durable machine-readable data.

## Quick path vs primitive path

Use the quick path when you want a low-effort visual answer:

```js
const comparison = await cvd.compare.region({
  left: leftPage.locator("#cta"),
  right: rightPage.locator("#cta"),
  outDir: "artifacts/cta",
})
return comparison.summary()
```

Use the primitive path when you want project-specific policy logic:

```js
const left = await leftPage.locator("#cta").collect({
  inspect: "rich",
  styles: ["font-size", "line-height", "color"],
  attributes: ["class", "aria-label"],
})
const right = await cvd.collect.selection(rightPage.locator("#cta"), {
  inspect: "rich",
  styles: ["font-size", "line-height", "color"],
  attributes: ["class", "aria-label"],
})

const comparison = await cvd.compare.selections(left, right, {
  styleProps: ["font-size", "line-height", "color"],
  attributes: ["class", "aria-label"],
})

return {
  ok: comparison.styles.diff(["font-size", "line-height"]).length === 0,
  summary: comparison.summary(),
  bounds: comparison.bounds.diff(),
}
```

The quick path is implemented as a composition of the primitives: collect left, collect right, capture regions, compare selections, and return a comparison handle.

## Packaging scripts into a reusable project CLI

One-off scripts are useful while experimenting, but the real power comes from packaging validation scripts as project-local CLI commands. That turns visual validation into a core development loop instead of a manual QA event.

A typical structure is:

```text
visual-verbs/
  shared.js
  homepage.js
  checkout.js
```

A project-local verb can compare key sections and write a catalog:

```js
// visual-verbs/homepage.js
async function validateHomepage(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  const catalog = cvd.catalog.create({
    title: "Homepage visual validation",
    outDir,
    artifactRoot: "artifacts",
  })

  const sections = [
    { name: "hero", selector: "[data-section='hero']" },
    { name: "primary-cta", selector: "[data-testid='primary-cta']" },
    { name: "footer", selector: "footer" },
  ]

  const summaries = []
  let leftPage, rightPage
  try {
    leftPage = await browser.page(leftUrl, { viewport: cvd.viewport.desktop(), waitMs: 500 })
    rightPage = await browser.page(rightUrl, { viewport: cvd.viewport.desktop(), waitMs: 500 })

    for (const section of sections) {
      const artifactDir = catalog.artifactDir(section.name)
      const comparison = await cvd.compare.region({
        name: section.name,
        left: leftPage.locator(section.selector),
        right: rightPage.locator(section.selector),
        outDir: artifactDir,
        styleProps: ["font-size", "line-height", "color", "background-color", "padding"],
        attributes: ["class", "aria-label"],
      })
      await comparison.artifacts.write(artifactDir, ["json", "markdown"])
      catalog.record(comparison, {
        slug: section.name,
        name: section.name,
        url: leftUrl,
        selector: section.selector,
      })
      summaries.push(comparison.summary())
    }

    return {
      ok: summaries.every(s => !s.pixel || s.pixel.changedPercent < 2.0),
      manifestPath: await catalog.writeManifest(),
      indexPath: await catalog.writeIndex(),
      summaries,
    }
  } finally {
    if (leftPage) await leftPage.close()
    if (rightPage) await rightPage.close()
    await browser.close()
  }
}

__verb__("validateHomepage", {
  parents: ["site"],
  short: "Validate homepage visual implementation against a reference URL",
  fields: {
    leftUrl: { argument: true, required: true, help: "Reference/baseline URL" },
    rightUrl: { argument: true, required: true, help: "Implementation/current URL" },
    outDir: { argument: true, required: true, help: "Artifact output directory" },
  },
})
```

Run it directly:

```bash
css-visual-diff verbs --repository ./visual-verbs site validate-homepage \
  http://localhost:4100 \
  http://localhost:4200 \
  ./artifacts/visual/homepage/latest \
  --output json
```

Or wrap it in project-native commands:

```json
{
  "scripts": {
    "visual:homepage": "css-visual-diff verbs --repository ./visual-verbs site validate-homepage",
    "visual:homepage:local": "npm run visual:homepage -- http://localhost:4100 http://localhost:4200 ./artifacts/visual/homepage/latest --output json"
  }
}
```

The reusable loop becomes:

```text
edit CSS/component
run visual command
open index.md and diff_comparison.png
inspect JSON/Markdown evidence
fix implementation
run again
```

For CI, run the same project command, upload the artifact directory, and use the JSON result as the pass/fail or warning signal. CI should not invent a separate visual workflow; it should run the same reusable command developers run locally.

## JavaScript verbs

`css-visual-diff` scans annotated JavaScript files and exposes them as typed CLI commands under the lazy `verbs` namespace:

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
GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples compare region \
  http://localhost:3000/original \
  http://localhost:3000/react \
  '#cta' \
  /tmp/cssvd-example \
  --output json
```

Generated JavaScript commands are intentionally not injected at the CLI root. Use `css-visual-diff verbs ...` so repository scan errors, duplicate script command paths, and script-specific flags stay scoped to the verbs subtree.

## Embedded help

The built binary includes Glazed help entries:

```bash
css-visual-diff help javascript-api
css-visual-diff help javascript-verbs
css-visual-diff help pixel-accuracy-scripting-guide
```

They cover the Promise-first module, browser/page/locator APIs, collection and comparison objects, pixel workflows, structural snapshots, catalogs, repository-scanned verbs, generated flags, output modes, and project CLI packaging patterns.

## YAML config workflows

YAML configs remain the right shape for repeatable baseline/current comparison jobs. A config can define targets, prepare steps, screenshot sections, CSS probes, output modes, and validation options.

Run one config:

```bash
GOWORK=off go run ./cmd/css-visual-diff run \
  --config examples/pyxis-public-shows.yaml \
  --modes capture,cssdiff,matched-styles,pixeldiff,html-report
```

Scan a directory for co-located configs:

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

It skips common generated/vendor directories such as `node_modules`, `.git`, `dist`, `build`, and `.css-visual-diff`.

## Inspect one side before comparing

When tuning selectors, inspect one side before running a full comparison. This helps verify the URL, viewport, prepare hook, root selector, section selector, and CSS probe selector.

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

For tight selector-tuning loops, use single-artifact commands:

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

Available single-artifact commands are `screenshot`, `css-md`, `css-json`, `html`, and `inspect-json`. They share the same selector flags as `inspect`: `--root`, `--section`, `--style`, or `--selector`.

## Prepared targets

Some design sources do not render the comparison target directly. They may open a pan/zoom design canvas, prototype shell, or development board first. Config-driven runs can prepare each target after navigation and before capture/CSS analysis.

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

When `root_selector` is set, capture mode uses an element screenshot for the full baseline PNG instead of a browser full-page screenshot. This avoids including prototype shell or design-canvas chrome in the exported baseline.

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

To generate a static artifact browser for the run, include `html-report` after the modes that produce artifacts. The report is written to `<output.dir>/index.html`.

## Legacy prototype

The previous Python prototype is preserved under:

```text
legacy/python-prototype/
```
