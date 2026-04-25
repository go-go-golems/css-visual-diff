---
Title: 'JavaScript API: require("css-visual-diff")'
Slug: javascript-api
Short: Use the Promise-first css-visual-diff JavaScript module for browsers, pages, locators, extraction, snapshots, diffs, catalogs, and config loading.
Topics:
- javascript
- goja
- visual-regression
- browser-automation
Commands:
- verbs
Flags:
- repository
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

`css-visual-diff` exposes a Promise-first JavaScript API for repository-scanned verbs. Scripts use it to drive Chromium pages, prepare targets, preflight selectors, inspect artifacts, and write visual catalog manifests.

This API is available inside scripts executed by:

```bash
css-visual-diff verbs ...
```

It is intentionally asynchronous from day one. Any operation that touches Chromium, timers, files, or catalog writes returns a Promise.

## Quick example

```js
async function inspect(url, selector, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()

  try {
    const page = await browser.page(url, {
      viewport: { width: 1280, height: 720 },
      waitMs: 250,
      name: "homepage",
    })

    const probes = [{
      name: "cta",
      selector,
      props: ["display", "color", "background-color", "font-size"],
      attributes: ["class"],
    }]

    const preflight = await page.preflight(probes)
    if (!preflight[0].exists) {
      throw new cvd.SelectorError(`selector did not match: ${selector}`)
    }

    const result = await page.inspectAll(probes, {
      outDir,
      artifacts: "css-json",
    })

    return {
      ok: true,
      outputDir: result.outputDir,
      resultCount: result.results.length,
    }
  } finally {
    await browser.close()
  }
}
```

## Module exports

```js
const cvd = require("css-visual-diff")
```

Exports:

- `cvd.browser(options?)`
- `cvd.catalog(options)`
- `cvd.loadConfig(path)`
- `cvd.viewport(width, height)` and named viewport helpers
- `cvd.target(name)`
- `cvd.probe(name)`
- `cvd.extractors.*`
- `cvd.extract(locator, extractors)`
- `cvd.snapshot(page, probes, options?)`
- `locator.collect(options?)` and `cvd.collect.selection(locator, options?)` for immutable collected selector data.
- `cvd.compare.selections(left, right, options?)` for comparing collected selector data.
- `cvd.diff(before, after, options?)` for the current structural JSON diff
- Upcoming canonical names from `CSSVD-JSAPI-PIXEL-WORKFLOWS`: `cvd.diff.structural(before, after, options?)` and `cvd.image.diff(options)`
- `cvd.write.json(path, value)`
- `cvd.write.markdown(path, markdown)`
- `cvd.CvdError`
- `cvd.SelectorError`
- `cvd.PrepareError`
- `cvd.BrowserError`
- `cvd.ArtifactError`

## Browser and page API

### `await cvd.browser(options?)`

Creates a Chromium-backed browser service.

```js
const browser = await cvd.browser()
```

Current options are reserved for future use.

### `await browser.page(url, options?)`

Creates a new page, sets the viewport, navigates to `url`, waits if requested, and applies no prepare step unless the target has one.

```js
const page = await browser.page("http://localhost:3000", {
  viewport: { width: 1280, height: 720 },
  waitMs: 500,
  name: "prototype",
})
```

Options:

- `viewport.width` — viewport width, default `1280`
- `viewport.height` — viewport height, default `720`
- `waitMs` — wait after navigation in milliseconds
- `name` — target/page name used in metadata

### `await browser.newPage()`

Creates a blank page. Use `page.goto(...)` afterwards.

```js
const page = await browser.newPage()
await page.goto(url, { viewport: { width: 800, height: 600 } })
```

### `await browser.close()`

Closes the browser service.

Always close browsers in `finally` blocks.

## Page methods

### `await page.goto(url, options?)`

Navigates an existing page.

```js
const target = await page.goto(url, {
  viewport: { width: 1024, height: 768 },
  waitMs: 200,
  name: "target-name",
})
```

Returns:

```js
{
  name: "target-name",
  url: "...",
  waitMs: 200,
  viewport: { width: 1024, height: 768 }
}
```

### `await page.prepare(spec)`

Runs a prepare step on the already-loaded page.

Script prepare:

```js
await page.prepare({
  type: "script",
  waitFor: "window.React && window.ReactDOM",
  waitForTimeoutMs: 5000,
  script: `document.body.dataset.ready = "true"`,
  afterWaitMs: 250,
})
```

Direct React global prepare:

```js
await page.prepare({
  type: "directReactGlobal",
  waitFor: "window.React && window.ReactDOM && window.PPXDesktop",
  component: "PPXDesktop",
  props: { page: "shows" },
  rootSelector: "#capture-root",
  width: 920,
  minHeight: 1200,
  background: "#fff",
})
```

`directReactGlobal` is a prepare/rendering mode, not a selector mode. It creates or targets a controlled root and renders a global React component into it before inspection.

### `await page.preflight(probes)`

Checks selectors before expensive extraction.

```js
const statuses = await page.preflight([
  { name: "cta", selector: "#cta", source: "styles", required: true },
])
```

Result entries are lowerCamel objects:

```js
{
  name: "cta",
  selector: "#cta",
  source: "styles",
  exists: true,
  visible: true,
  bounds: { x: 10, y: 20, width: 120, height: 40 },
  textStart: "Book now",
  error: ""
}
```

Use preflight to decide authoring vs CI policy:

- authoring mode: record misses and continue,
- CI mode: throw or rethrow on missing selectors.

### `await page.inspect(probe, options)`

Inspects one probe and writes artifacts.

```js
const artifact = await page.inspect(
  { name: "cta", selector: "#cta", props: ["color"] },
  { outDir: "/tmp/cssvd/cta", artifacts: "css-json" }
)
```

Result:

```js
{
  metadata: {
    side: "script",
    targetName: "prototype",
    url: "http://...",
    viewport: { width: 1280, height: 720 },
    name: "cta",
    selector: "#cta",
    selectorSource: "styles",
    rootSelector: "",
    prepareType: "",
    format: "css-json",
    createdAt: "2026-04-24T...Z"
  },
  style: {
    exists: true,
    computed: { color: "rgb(255, 0, 0)" },
    bounds: { x: 0, y: 0, width: 100, height: 32 },
    attributes: { class: "button" }
  },
  screenshot: "",
  html: "",
  inspectJson: ""
}
```

### `await page.inspectAll(probes, options)`

Inspects multiple probes against one already-loaded/prepared page.

```js
const result = await page.inspectAll(probes, {
  outDir: "/tmp/cssvd/catalog/artifacts/homepage",
  artifacts: "bundle",
})
```

Result:

```js
{
  outputDir: "/tmp/cssvd/catalog/artifacts/homepage",
  results: [ /* inspect artifacts */ ]
}
```

For multiple probes, artifacts are written below per-probe subdirectories. For one probe, artifacts are written directly into `outDir`.

### `page.locator(selector)`

Creates a synchronous page-bound locator handle. Creating a locator does not query the browser. The async locator methods do.

```js
const cta = page.locator("#cta")
const exists = await cta.exists()
```

Locator methods:

- `await locator.status()` — returns selector status with existence, visibility, bounds, text start, and selector error.
- `await locator.exists()` — returns a boolean.
- `await locator.visible()` — returns a boolean.
- `await locator.text(options?)` — returns text content. Use `{ normalizeWhitespace: true, trim: true }` for stable comparisons.
- `await locator.bounds()` — returns `{ x, y, width, height }` or `null` for a missing selector.
- `await locator.computedStyle(props)` — returns a map of CSS property values.
- `await locator.attributes(names)` — returns a map of attribute values.

Example:

```js
const cta = page.locator("#cta")
const [text, bounds, styles] = await Promise.all([
  cta.text({ normalizeWhitespace: true, trim: true }),
  cta.bounds(),
  cta.computedStyle(["height", "color", "background-color"]),
])
```

Operations on one page are serialized internally, so `Promise.all` is safe for page-bound reads.

## Collected selection model

`CSSVD-JSAPI-PIXEL-WORKFLOWS` introduces a JavaScript-first collection primitive. A locator is live: it says “find this element on this loaded page.” A collected selection is immutable data: it says “these were the browser facts for this selector at this moment.”

The Go service model is implemented as `CollectedSelectionData` with schema version `cssvd.collectedSelection.v1`. The JavaScript handle is exposed as:

```js
const selected = await page.locator("#cta").collect({ inspect: "rich" })
// equivalent namespace form:
const same = await cvd.collect.selection(page.locator("#cta"), { inspect: "rich" })
```

Collection profiles:

- `inspect: "minimal"` — selector status, existence, visibility, and bounds.
- `inspect: "rich"` — default profile for scripts; includes normalized text, common computed styles, common attributes, and status/bounds.
- `inspect: "debug"` — intended for deeper diagnostics; includes HTML, all computed styles, and all attributes.
- object form — custom profile equivalent to the Go `CollectOptions` fields.

Collected data lowers to plain JSON:

```js
{
  schemaVersion: "cssvd.collectedSelection.v1",
  name: "cta",
  url: "http://localhost:3000/",
  selector: "#cta",
  source: "script",
  exists: true,
  visible: true,
  bounds: { x: 32, y: 48, width: 120, height: 40 },
  text: "Book now",
  computedStyles: { color: "rgb(255, 0, 0)", display: "inline-block" },
  attributes: { id: "cta", class: "primary" }
}
```

Use collected selections when later JavaScript wants to compare, filter, serialize, or report on browser facts without re-querying the page. The comparison API builds on this model with `cvd.compare.selections(left, right)` and the low-effort `cvd.compare.region({ left, right })` helper.

## Selection comparison model

A `SelectionComparison` compares two collected selections. It is data-centered: it does not re-query the browser. It compares the immutable facts already captured in the left and right `CollectedSelection` values.

The Go service model is implemented as `SelectionComparisonData` with schema version `cssvd.selectionComparison.v1`. The JavaScript handle is exposed as:

```js
const left = await leftPage.locator("#cta").collect({ inspect: "rich" })
const right = await rightPage.locator("#cta").collect({ inspect: "rich" })

const comparison = await cvd.compare.selections(left, right, {
  threshold: 30,
  styleProps: ["font-size", "line-height", "color"],
  attributes: ["class", "data-state"],
})
```

A lowered comparison has stable lowerCamel data:

```js
{
  schemaVersion: "cssvd.selectionComparison.v1",
  name: "cta",
  left: { selector: "#cta", exists: true, visible: true, bounds: { /* ... */ } },
  right: { selector: "#cta", exists: true, visible: true, bounds: { /* ... */ } },
  pixel: {
    threshold: 30,
    changedPixels: 713,
    changedPercent: 7.13,
    normalizedWidth: 500,
    normalizedHeight: 20
  },
  bounds: {
    changed: true,
    delta: { x: 0, y: 2, width: 0, height: 4 }
  },
  text: { changed: false, left: "Book now", right: "Book now" },
  styles: [
    { name: "font-size", left: "16px", right: "18px", changed: true }
  ],
  attributes: [
    { name: "class", left: "primary", right: "secondary", changed: true }
  ],
  artifacts: [
    { name: "diffComparison", kind: "png", path: "artifacts/diff_comparison.png" }
  ]
}
```

The JavaScript handle exposes query methods over this data rather than forcing users to parse the full JSON every time:

```js
comparison.pixel.summary()
comparison.bounds.diff()
comparison.styles.diff(["font-size", "color"])
comparison.attributes.diff(["class"])
comparison.report.markdown()
comparison.artifacts.write(outDir, ["diffComparison", "json", "markdown"])
```

Reports and artifacts are views over comparison data. The comparison data is the durable source of truth; Markdown, PNGs, and catalog entries are outputs derived from it.

### `await page.close()`

Closes the page.

## Lower-level builders, extraction, snapshots, and diffs

The lower-level API is for script-native visual checks that do not always need inspect artifacts. Builder and handle objects are Go-backed values with controlled methods. If you call a method on the wrong object, the API reports which object owns that method.

### `cvd.viewport(...)`

```js
cvd.viewport(1280, 720)
cvd.viewport({ width: 1280, height: 720 })
cvd.viewport.desktop()
cvd.viewport.tablet()
cvd.viewport.mobile()
```

### `cvd.target(name)`

Builds a page target definition.

```js
const target = cvd.target("homepage")
  .url("http://localhost:3000")
  .viewport(cvd.viewport.desktop())
  .waitMs(250)
  .root("#app")
  .build()
```

### `cvd.probe(name)`

Builds a reusable inspection recipe.

```js
const probe = cvd.probe("cta")
  .selector("#cta")
  .required()
  .text()
  .bounds()
  .styles(["color", "font-size", "background-color"])
  .attributes(["class"])
```

Use probes with `cvd.snapshot(...)`. Use locators when you want to inspect one already-loaded page directly.

### `cvd.extractors.*`

Extractor handles describe which facts to read from a locator.

```js
const extractors = [
  cvd.extractors.exists(),
  cvd.extractors.visible(),
  cvd.extractors.text(),
  cvd.extractors.bounds(),
  cvd.extractors.computedStyle(["color"]),
  cvd.extractors.attributes(["id", "class"]),
]
```

### `await cvd.extract(locator, extractors)`

Strictly extracts facts from a page-bound locator. The first argument must be a locator returned by `page.locator(...)`. The second argument must be an array of `cvd.extractors.*` handles.

```js
const snapshot = await cvd.extract(page.locator("#cta"), [
  cvd.extractors.exists(),
  cvd.extractors.text(),
  cvd.extractors.computedStyle(["color"]),
])
```

Raw object locators are rejected. Use `page.locator("#selector")` instead.

### `await cvd.snapshot(page, probes, options?)`

Strictly evaluates probe builders against a page.

```js
const snapshot = await cvd.snapshot(page, [
  cvd.probe("title").selector("h1").text().styles(["font-size", "color"]),
  cvd.probe("cta").selector("#cta").text().bounds().styles(["background-color"]),
])
```

Raw object probes are rejected. Use `cvd.probe("name").selector("...")` builders.

### Structural diffs and image/pixel diffs

There are two different kinds of diffing in css-visual-diff:

1. **Structural JSON diffs** compare plain values, snapshots, and extracted data.
2. **Image/pixel diffs** compare rendered PNGs or screenshots.

The current public function is the older structural diff helper:

```js
const diff = cvd.diff(before, after, {
  ignorePaths: ["results[0].snapshot.bounds.x"],
})

const markdown = cvd.report(diff).markdown()
await cvd.write.json("out/diff.json", diff)
await cvd.report(diff).writeMarkdown("out/diff.md")
```

Because this ticket does not require backward compatibility, the canonical future names should be explicit:

```js
const structural = cvd.diff.structural(before, after)
const pixels = await cvd.image.diff({
  left: "artifacts/left.png",
  right: "artifacts/right.png",
  threshold: 30,
  diffOnlyPath: "artifacts/diff_only.png",
  diffComparisonPath: "artifacts/diff_comparison.png",
})
```

The Go service primitive for image diffs now returns lowerCamel data shaped like:

```js
{
  threshold: 30,
  totalPixels: 10000,
  changedPixels: 713,
  changedPercent: 7.13,
  normalizedWidth: 500,
  normalizedHeight: 20,
  diffOnlyPath: "artifacts/diff_only.png",
  diffComparisonPath: "artifacts/diff_comparison.png"
}
```

`threshold` is an RGB distance threshold from `0` to `255`. Images with different sizes are normalized by padding to the larger width and height before pixels are compared. Writer helpers create parent directories before writing diff PNG artifacts.

## Artifact formats

Supported `artifacts` / `format` values:

- `bundle` — metadata, screenshot, prepared HTML, CSS JSON/Markdown, DOM inspect JSON
- `png` — selector screenshot
- `html` — prepared HTML for selector/root
- `css-json` — computed CSS JSON
- `css-md` — computed CSS Markdown
- `inspect-json` — DOM inspection JSON
- `metadata-json` — metadata JSON only

## Catalog API

The catalog implementation is Go-backed. JavaScript orchestrates workflows, while Go owns schema versioning, path normalization, summaries, manifest writing, and index writing.

```js
const catalog = cvd.catalog({
  title: "Prototype Catalog",
  outDir: "/tmp/catalog",
  artifactRoot: "artifacts",
  indexName: "index.md",
})
```

### `catalog.artifactDir(slug)`

Returns a normalized artifact directory below `outDir/artifactRoot`.

```js
catalog.artifactDir("../Prototype Public Shows!")
// /tmp/catalog/artifacts/prototype-public-shows
```

### `catalog.addTarget(target)`

Adds a target record.

```js
catalog.addTarget({
  slug: "homepage",
  name: "Homepage",
  url: "http://localhost:3000",
  selector: "#root",
  viewport: { width: 1280, height: 720 },
  metadata: { source: "storybook" },
})
```

### `catalog.recordPreflight(target, statuses)`

Records selector preflight results.

### `catalog.addResult(target, inspectResult)`

Records a `page.inspectAll(...)`-shaped result.

### `catalog.addFailure(target, error)`

Records a failure. If passed a typed `cvd.*Error`, the catalog captures name/code/operation/message.

### `catalog.summary()`

Returns lowerCamel summary counts:

```js
{
  targetCount: 1,
  preflightCount: 1,
  resultCount: 1,
  failureCount: 0,
  artifactCount: 1
}
```

### `catalog.manifest()`

Returns the current manifest object in lowerCamel form.

### `await catalog.writeManifest()`

Writes:

```text
<outDir>/manifest.json
```

Returns the manifest path.

### `await catalog.writeIndex()`

Writes:

```text
<outDir>/<indexName>
```

Returns the index path.

## YAML interop

`cvd.loadConfig(path)` loads an existing `*.css-visual-diff.yml` file through the Go config loader and returns a lowerCamel JS object.

```js
const cfg = await cvd.loadConfig("examples/pyxis-public-shows.yaml")
const target = cfg.original
const probes = cfg.styles.map((style) => ({
  name: style.name,
  selector: style.selectorOriginal || style.selector,
  props: style.props,
  attributes: style.attributes,
}))
```

The built-in command below uses this helper:

```bash
css-visual-diff verbs catalog inspect-config \
  ./example.css-visual-diff.yml original /tmp/catalog \
  --artifacts css-json \
  --output json
```

## Error model

Promise rejections use JS-visible typed errors:

```js
try {
  await page.inspect({ name: "missing", selector: "#missing" }, { outDir, artifacts: "html" })
} catch (err) {
  if (err instanceof cvd.SelectorError) {
    // selector/preflight failure
  }
  if (err instanceof cvd.CvdError) {
    console.error(err.code, err.operation, err.message)
  }
}
```

Classes:

- `CvdError` — base class
- `SelectorError` — selector/preflight failures
- `PrepareError` — prepare/wait failures
- `BrowserError` — browser/page/navigation failures
- `ArtifactError` — inspect/artifact writing failures

Each error has:

- `name`
- `code`
- `operation`
- `details`
- `message`

## Concurrency guidance

Prefer coarse target/page-level parallelism. For example, process independent catalog targets with a worker limit and give each worker its own page.

Avoid assuming per-page CDP operations are meaningfully parallel. Operations on one Chromium page are effectively serialized by chromedp/CDP and should be awaited in order:

```js
for (const target of targets) {
  const page = await browser.page(target.url, { viewport: target.viewport })
  try {
    await page.prepare(target.prepare)
    const preflight = await page.preflight(target.probes)
    const result = await page.inspectAll(target.probes, { outDir: catalog.artifactDir(target.slug) })
    catalog.addResult(target, result)
  } finally {
    await page.close()
  }
}
```

Add explicit worker limits before scaling this pattern to many pages.
