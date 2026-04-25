---
Title: Elegant JavaScript API additions for pixel comparison workflows
Ticket: CSSVD-JSAPI-PIXEL-WORKFLOWS
Status: active
Topics:
    - frontend
    - visual-regression
    - browser-automation
    - tooling
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/doc/topics/javascript-api.md
      Note: |-
        Public JS API reference that should document any new supported module surface
        Embedded public JS API reference to update with cvd.compare.region
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: |-
        Registers the internal require("diff").compareRegion helper and maps it into compare mode settings
        Native require modules that bridge built-in scripts to Go compare mode
    - Path: internal/cssvisualdiff/dsl/scripts/compare.js
      Note: |-
        Built-in JS compare-region verb that exposes the current workflow through an internal helper module
        Current built-in compare-region JS verb and legacy API shape
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: |-
        Existing strict page-bound locator API that should be reused by the public compare-region API
        Strict locator handle proposed as the public compare-region input
    - Path: internal/cssvisualdiff/modes/compare.go
      Note: |-
        Current browser, screenshot, computed-style, matched-style, and pixel-diff implementation used by compare region
        Current region screenshot
    - Path: internal/cssvisualdiff/service/diff.go
      Note: Existing structural diff service; new pixel compare API must not confuse this with image diffing
ExternalSources:
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
Summary: Design for exposing elegant, strict, composable JavaScript APIs for region pixel comparison, stable compare schemas, and future workflow orchestration based on the first real Pyxis user feedback.
LastUpdated: 2026-04-25T09:15:00-04:00
WhatFor: Use this as the implementation guide for turning the internal compare-region machinery into a public css-visual-diff JavaScript API.
WhenToUse: Use before implementing cvd.compare.region, image diff primitives, config job bridges, or compare-result schema documentation.
---


# Elegant JavaScript API Additions for Pixel Comparison Workflows

## 1. Executive summary

Pyxis is the first serious downstream user of the new `css-visual-diff` JavaScript API. Their feedback is valuable because it comes from an actual workflow: comparing standalone prototype pages against React/Storybook implementations, migrating repetitive YAML visual-diff configs toward programmable JavaScript workflows, and trying to produce useful authoring and review artifacts.

The user request can be summarized as:

> “The JavaScript API lets us load pages, locate elements, preflight selectors, inspect CSS, build snapshots, and write reports. But we cannot call the same pixel/region comparison machinery that powers the built-in `verbs script compare region` command. Please expose that as a public JS API so project scripts do not need to shell out or duplicate internals.”

That request is correct. The best answer is not to document `require("diff")` as a public API and not to add a single loose `cvd.comparePixels(rawObject)` helper. The best answer is to expose a small, composable compare namespace under the existing public module:

```js
const cvd = require("css-visual-diff")

await cvd.compare.region({
  name: "archive-content",
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
  outDir,
  artifacts: ["screenshots", "diff", "json", "markdown"],
  evidence: {
    computed: ["font-family", "font-size", "line-height", "color", "background-color"],
    attributes: ["id", "class"],
    matched: true,
  },
})
```

The important design principles are:

1. **One public module name remains `require("css-visual-diff")`.** Do not promote internal `require("diff")` or `require("report")` as the main user API.
2. **Region comparison should accept page-bound locator/region handles, not arbitrary raw page objects.** This preserves the strict handle model and better error messages.
3. **Pixel comparison should be implemented in reusable Go services.** CLI modes, built-in JS verbs, YAML runner, and public JS API should converge on one implementation.
4. **Results should be plain serializable data with a stable schema version.** Authoring handles are Go-backed; outputs are JSON-like values.
5. **Expose the primitive, not a Pyxis framework.** Pyxis page registries, policy bands, report templates, and accepted-difference lists should remain userland until they prove general.

## 2. Background: what already exists

The current JavaScript stack already has two layers.

The documented public module is:

```js
const cvd = require("css-visual-diff")
```

It includes browser/page/locator/probe/extractor/snapshot/structural-diff/catalog helpers.

The built-in compare-region verb is implemented separately in:

```text
internal/cssvisualdiff/dsl/scripts/compare.js
```

It calls an internal runtime helper:

```js
require("diff").compareRegion({ ... })
```

That helper is registered in:

```text
internal/cssvisualdiff/dsl/registrar.go
```

and ultimately calls:

```go
modes.GenerateCompareResult(ctx, settings)
modes.WriteCompareArtifacts(result, settings.WriteJSON, settings.WriteMarkdown)
```

from:

```text
internal/cssvisualdiff/modes/compare.go
```

The existing compare mode already does most of what Pyxis needs:

- open two pages,
- set viewports,
- navigate to URLs,
- wait,
- screenshot full pages,
- screenshot selected regions,
- extract computed styles,
- extract matched-style winners,
- pixel-diff the selected region screenshots,
- write optional JSON and Markdown artifacts,
- return structured data.

The problem is packaging and public API boundary. Project scripts that already have open pages cannot call the region/pixel primitive through `require("css-visual-diff")`.

## 3. Critical read of the Pyxis request

The Pyxis request contains a proposed minimal API:

```js
await cvd.comparePixels({
  left: { page: leftPage, selector: '#root > *' },
  right: { page: rightPage, selector: "[data-page='archive']" },
  threshold: 30,
  outDir,
  writeJson: true,
  writeMarkdown: true,
  writePngs: true,
})
```

This expresses a real need, but the shape should be refined.

### 3.1 What the request gets right

- It works with already-open pages.
- It avoids subprocess calls from Goja.
- It exposes an existing core capability rather than asking users to reimplement image diffing.
- It lets project-level scripts handle registries, loops, policies, and reports.

### 3.2 What should change

#### Name: `comparePixels` is too narrow

The current compare-region operation is not only pixel diffing. It can also write screenshots, include artifact paths, extract computed styles, and explain matched CSS rules. A better public namespace is:

```js
cvd.compare.region(...)
```

Then a truly lower-level image API can exist separately:

```js
cvd.image.diff(...)
```

#### Arguments: raw `{ page, selector }` objects are too loose

The newer JS API intentionally uses Go-backed handles/builders for domain objects. We should not regress to accepting arbitrary raw page-like objects for the central new primitive.

Prefer:

```js
left: leftPage.locator("#root > *")
right: rightPage.locator("[data-page='archive']")
```

or, if we introduce a named region handle:

```js
left: leftPage.region("#root > *", { name: "prototype-content" })
right: rightPage.region("[data-page='archive']", { name: "react-content" })
```

#### Output options: booleans should become artifact intent

This is serviceable:

```js
writeJson: true,
writeMarkdown: true,
writePngs: true,
```

but this is more scalable:

```js
artifacts: ["screenshots", "diff", "json", "markdown"]
```

or:

```js
artifacts: {
  screenshots: true,
  diff: true,
  json: true,
  markdown: true,
}
```

For initial implementation, accepting both is reasonable, but the reference docs should teach `artifacts`.

#### Result schema should be JS-native and versioned

Current Go compare result fields are snake_case because they are Go JSON structs. The JS API should return lowerCamel and a schema version:

```js
{
  schemaVersion: "cssvd.compare.region.v1",
  name: "archive-content",
  equal: false,
  threshold: 30,
  changedPercent: 7.128146453089244,
  changedPixels: 102172,
  totalPixels: 1433360,
  normalizedWidth: 920,
  normalizedHeight: 1558,
  artifacts: { ... },
  left: { ... },
  right: { ... },
  computedDiffs: [...],
  winnerDiffs: [...]
}
```

Persisted `compare.json` written by the new JS API should use the same schema unless explicitly asked for legacy output.

## 4. Recommended public API

### 4.1 P0 API: `cvd.compare.region(...)`

Primary form:

```js
const result = await cvd.compare.region({
  name: "archive-content",
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
  outDir: "out/archive-content",
  artifacts: ["screenshots", "diff", "json", "markdown"],
  evidence: {
    computed: [
      "font-family",
      "font-size",
      "font-weight",
      "line-height",
      "padding-top",
      "padding-bottom",
      "color",
      "background-color",
      "box-shadow",
    ],
    attributes: ["id", "class"],
    matched: true,
  },
})
```

The operation returns a Promise because it touches browser pages and may compute screenshots/pixel differences.

### 4.2 Refined direction: return a comparison result object

A better API may be to make `cvd.compare.region(...)` return a single **comparison result object** rather than immediately forcing every artifact to be selected up front. This fits the user's composability concern: first create the comparison, then ask that comparison for the specific representation or artifact needed by the current workflow.

Recommended shape:

```js
const comparison = await cvd.compare.region({
  name: "archive-content",
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,

  // Optional: choose how much browser-side inspection data the comparison captures.
  // "rich" should be the default for script-facing comparisons; callers filter later.
  inspect: "rich",
})

// Cheap structured facts.
const summary = comparison.summary()
const data = comparison.toJSON()

// Human-facing report derived from the same comparison.
const markdown = comparison.report.markdown()
await comparison.report.writeMarkdown("out/archive-content/compare.md")

// Artifact materialization happens on demand.
await comparison.artifacts.write("out/archive-content", [
  "leftRegion",
  "rightRegion",
  "diffOnly",
  "diffComparison",
  "json",
])

// Individual artifact queries are possible too.
const diffPath = await comparison.artifact("diffComparison").write("out/archive-content/diff_comparison.png")
```

This design makes artifacts **views over one materialized comparison**, not separate operations that may accidentally recapture or recompute different evidence. The comparison object can memoize screenshots, image buffers, diff stats, inspection data, style evidence, and report text. It also gives userland scripts a cleaner choice:

- return `comparison.toJSON()` for CLI JSON rows,
- write only Markdown for human reports,
- write only PNG diff artifacts for CI uploads,
- query and filter rich inspection data after the fact,
- build multi-section workflows by storing an array of comparison objects and lowering them at the end.

This also changes the recommendation for the earlier `evidence` field. If browser-side collection is cheap relative to screenshots and pixel diffing, then the comparison object should collect a rich inspection payload by default and provide filtered views. The API should not force users to know every useful CSS property before comparing. Instead, it should let them ask questions after the comparison exists:

```js
const typography = comparison.styles.diff([
  "font-family",
  "font-size",
  "font-weight",
  "line-height",
  "color",
])

const layout = comparison.styles.diff(cvd.styles.presets.layout)
const attrs = comparison.attributes.diff(["class", "data-page"])
const bounds = comparison.bounds.diff()
```

The option should therefore be named `inspect`, not `evidence`, because it controls the browser inspection profile used to enrich the comparison. Suggested shapes:

```js
inspect: "rich"      // default: bounds, attributes, computed style map, and practical rule evidence
inspect: "minimal"   // only what is required for pixel comparison and summary stats
inspect: {
  styles: "all",     // or a string array when users need a smaller capture
  attributes: "all", // or a string array
  matchedStyles: true,
  bounds: true,
  text: false,
}
```

`rich` should mean “collect enough browser facts for flexible scripting,” not “write all of it to every report.” Reports, JSON output, and Markdown tables can filter aggressively from the already-collected comparison object.

The important caveat is that command output must remain serializable. A Go-backed comparison object with methods is useful inside a script, but jsverbs output processors should receive plain data. Therefore the object must provide an explicit `toJSON()`/`summary()` lowering path, and documentation should teach returning plain data from verbs:

```js
async function compareArchive(outDir) {
  const comparison = await runComparison(outDir)
  await comparison.artifacts.write(outDir, ["diffComparison", "json", "markdown"])
  return comparison.toJSON()
}
```

This slightly relaxes the earlier rule that final results are always plain data. The revised rule should be:

> Long-lived authoring handles and lazy artifact handles can be Go-backed. Anything returned to CLI output, written as JSON, or used as an interchange format must be lowered to a plain serializable value.

### 4.3 Short form

For quick scripts:

```js
const result = await cvd.compare.region(leftLocator, rightLocator, {
  threshold: 30,
  outDir,
  artifacts: ["diff", "json"],
})
```

This is ergonomic, but implementation can wait. The object form is easier to validate and document.

### 4.4 Optional `page.region(...)` handle

A locator can be enough for P0. A later API may introduce named visual regions:

```js
const leftContent = leftPage.region("#root > *", { name: "prototype-content" })
const rightContent = rightPage.region("[data-page='archive']", { name: "react-content" })

await cvd.compare.region({ left: leftContent, right: rightContent })
```

A `cvd.region` or `page.region` handle is useful if we want region-specific methods:

```js
await region.screenshot({ path: "content.png" })
await region.capture({ artifacts: ["screenshot", "bounds", "styles"] })
```

But adding a new handle is not required to solve Pyxis P0. A locator is already page-bound and selector-aware.

### 4.5 Low-level image API: `cvd.image.diff(...)`

Pixel comparison has a pure image layer. Exposing it separately makes the system more composable:

```js
const diff = await cvd.image.diff({
  left: "out/url1_screenshot.png",
  right: "out/url2_screenshot.png",
  threshold: 30,
  outDir: "out/diff",
  artifacts: ["diff", "comparison"],
})
```

Result:

```js
{
  schemaVersion: "cssvd.image.diff.v1",
  threshold: 30,
  changedPercent: 7.128146453089244,
  changedPixels: 102172,
  totalPixels: 1433360,
  normalizedWidth: 920,
  normalizedHeight: 1558,
  artifacts: {
    diffOnly: "out/diff/diff_only.png",
    diffComparison: "out/diff/diff_comparison.png"
  }
}
```

This should probably be implemented after `compare.region`, because `compare.region` is the user blocker.

### 4.6 Future helper: `cvd.compare.sections(...)`

A multi-section helper is useful but should not be the first primitive. Once `compare.region` exists, userland can already write:

```js
const results = []
for (const section of sections) {
  results.push(await cvd.compare.region({
    name: section.name,
    left: leftPage.locator(section.leftSelector),
    right: rightPage.locator(section.rightSelector),
    outDir: `${outDir}/${section.name}`,
    threshold: 30,
    artifacts: ["screenshots", "diff", "json", "markdown"],
  }))
}
```

If this pattern repeats, core can add:

```js
await cvd.compare.sections({
  leftPage,
  rightPage,
  sections: [
    { name: "page", leftSelector: "#root", rightSelector: "[data-story-frame='pyxis-page-shell']" },
    { name: "content", leftSelector: "#root > *", rightSelector: "[data-page='archive']" },
  ],
  threshold: 30,
  outDir,
})
```

But the helper should be written on top of the same service primitives.

## 5. Result schema proposal

There are now two related concepts:

1. **`ComparisonHandle`** — a Go-backed, script-local object with methods for querying summaries, reports, and artifacts.
2. **`CompareRegionResultV1`** — a plain serializable data object returned by `comparison.toJSON()` and written to `compare.json`.

The handle is for composition inside JavaScript. The data object is for CLI output, JSON files, catalogs, and downstream tools.

### 5.1 `ComparisonHandle` interface

```ts
type ComparisonHandle = {
  kind: "cvd.comparison";
  schemaVersion: "cssvd.compare.region.v1";

  summary(): CompareSummaryV1;
  toJSON(options?: CompareJSONOptions): CompareRegionResultV1;

  bounds: {
    left(): Bounds | null;
    right(): Bounds | null;
    diff(): BoundsDiff;
  };

  styles: {
    left(props?: string[]): Record<string, string>;
    right(props?: string[]): Record<string, string>;
    diff(props?: string[] | StylePreset): StyleDiffV1[];
  };

  attributes: {
    left(names?: string[]): Record<string, string>;
    right(names?: string[]): Record<string, string>;
    diff(names?: string[]): AttributeDiffV1[];
  };

  report: {
    markdown(options?: ReportOptions): string;
    writeMarkdown(path: string, options?: ReportOptions): Promise<string>;
  };

  artifact(name: CompareArtifactName): ComparisonArtifactHandle;

  artifacts: {
    list(): ComparisonArtifactDescriptor[];
    write(outDir: string, names?: CompareArtifactName[]): Promise<CompareRegionResultV1>;
  };
};

type CompareArtifactName =
  | "leftFull"
  | "rightFull"
  | "leftRegion"
  | "rightRegion"
  | "diffOnly"
  | "diffComparison"
  | "json"
  | "markdown";

type ComparisonArtifactHandle = {
  name: CompareArtifactName;
  exists(): boolean;
  path(): string | undefined;
  write(path?: string): Promise<string>;
};
```

`comparison.summary()` should be small and stable enough for tables:

```js
{
  schemaVersion: "cssvd.compare.summary.v1",
  name: "archive-content",
  equal: false,
  threshold: 30,
  changedPercent: 7.128146453089244,
  changedPixels: 102172,
  totalPixels: 1433360,
  artifactsWritten: ["diffComparison", "json"]
}
```

### 5.2 `CompareRegionResultV1`

```ts
type CompareRegionResultV1 = {
  schemaVersion: "cssvd.compare.region.v1";
  name?: string;
  equal: boolean;
  threshold: number;
  changedPercent: number;
  changedPixels: number;
  totalPixels: number;
  normalizedWidth: number;
  normalizedHeight: number;

  artifacts: {
    outDir?: string;
    leftFull?: string;
    rightFull?: string;
    leftRegion?: string;
    rightRegion?: string;
    diffComparison?: string;
    diffOnly?: string;
    json?: string;
    markdown?: string;
  };

  left: CompareRegionSideV1;
  right: CompareRegionSideV1;

  computedDiffs?: StyleDiffV1[];
  winnerDiffs?: WinnerDiffV1[];
};

type CompareRegionSideV1 = {
  name?: string;
  url?: string;
  selector: string;
  bounds?: { x: number; y: number; width: number; height: number };
  screenshot?: string;
  fullScreenshot?: string;
  computed?: Record<string, string>;
  attributes?: Record<string, string>;
};
```

### 5.3 Equality semantics

For pixel comparison, `equal` should mean:

```text
changedPixels === 0
```

under the selected threshold. It should not mean “acceptable under project policy.” Policy bands belong in userland:

```js
const status = result.changedPercent < 1 ? "accepted" : "review"
```

A future option can support a tolerance:

```js
maxChangedPercent: 0.5
```

but do not overload `equal` in v1.

### 5.4 Artifact naming

Default file names for one region:

```text
left_full.png
left_region.png
right_full.png
right_region.png
diff_only.png
diff_comparison.png
compare.json
compare.md
```

If `name` is provided and multiple region results share one directory, names may be prefixed:

```text
archive-content_left_region.png
archive-content_right_region.png
archive-content_diff_only.png
```

The implementation should pick a safe deterministic scheme and document it.

### 5.5 Legacy mapping

The existing compare mode returns fields like:

```text
pixel_diff.changed_percent
pixel_diff.diff_comparison_path
url1.element_screenshot
url2.element_screenshot
computed_diffs
winner_diffs
```

The public JS result should not expose these as the primary shape. If needed, include:

```js
legacy?: { ... }
```

or provide a conversion function for compatibility. Prefer not to expose legacy unless a real migration need appears.

## 6. Options schema proposal

```ts
type CompareRegionOptionsV1 = {
  name?: string;
  left: LocatorHandle | RegionHandle;
  right: LocatorHandle | RegionHandle;

  threshold?: number; // default 30
  outDir?: string;
  artifacts?: CompareArtifactOption[] | CompareArtifactOptions;

  inspect?:
    | "rich"
    | "minimal"
    | {
        styles?: "all" | string[] | false;
        attributes?: "all" | string[] | false;
        matchedStyles?: boolean;
        bounds?: boolean;
        text?: boolean;
        fullScreenshots?: boolean;
      };

  failOnMissing?: boolean; // default true for compare, because screenshot needs element
};

type CompareArtifactOption =
  | "screenshots"
  | "fullScreenshots"
  | "regionScreenshots"
  | "diff"
  | "json"
  | "markdown";
```

Default inspection profile:

```js
inspect: "rich"
```

`rich` should collect broad browser facts because the marginal browser cost is usually small compared with navigation, screenshots, image normalization, and pixel diffing. The noisy part is not collecting the data; the noisy part is dumping all of it into every report. The comparison object should therefore keep rich data available and let reports/JSON/views filter it after the fact.

A `minimal` profile remains useful for very large batch runs or CI jobs where memory and artifact size matter more than exploratory flexibility.

Default artifacts:

```js
["regionScreenshots", "diff"]
```

If `outDir` is omitted, return in-memory stats and temporary paths only if screenshots are necessary; however, because pixel diff currently operates on PNG files, the first implementation can require `outDir` whenever image artifacts are requested. Keep the error explicit:

```text
cvd.compare.region: outDir is required when artifacts include screenshots or diff
```

## 7. Service-layer design

Do not implement this only in `jsapi`. Extract reusable services first.

Proposed files:

```text
internal/cssvisualdiff/service/pixel.go
internal/cssvisualdiff/service/region_compare.go
internal/cssvisualdiff/service/region_compare_test.go
internal/cssvisualdiff/jsapi/compare.go
internal/cssvisualdiff/jsapi/compare_test.go
```

### 7.1 Service types

```go
type RegionRef struct {
    Name     string
    Selector string
    Source   string
}

type CompareRegionOptions struct {
    Name       string
    Threshold  int
    OutDir     string
    Artifacts  CompareArtifacts
    Evidence   CompareEvidence
}

type CompareArtifacts struct {
    RegionScreenshots bool
    FullScreenshots   bool
    Diff              bool
    JSON              bool
    Markdown          bool
}

type CompareEvidence struct {
    Bounds     bool
    Computed   []string
    Attributes []string
    Matched    bool
}

type CompareRegionResult struct {
    SchemaVersion string
    Name          string
    Equal         bool
    Threshold     int
    ChangedPercent float64
    ChangedPixels  int
    TotalPixels    int
    NormalizedWidth int
    NormalizedHeight int
    Artifacts CompareRegionArtifacts
    Left      CompareRegionSide
    Right     CompareRegionSide
    ComputedDiffs []StyleDiff
    WinnerDiffs []WinnerDiff
}
```

### 7.2 Service function

```go
func CompareRegions(
    ctx context.Context,
    leftPage *driver.Page,
    left service.LocatorSpec,
    rightPage *driver.Page,
    right service.LocatorSpec,
    opts CompareRegionOptions,
) (CompareRegionResult, error)
```

This is page-aware but Goja-free. It should be usable by:

- `internal/cssvisualdiff/jsapi/compare.go`,
- `internal/cssvisualdiff/dsl/registrar.go` internal helper,
- `internal/cssvisualdiff/modes/compare.go` eventually,
- future YAML runner bridge.

### 7.3 Pixel image function

Move image diffing out of `modes/compare.go`:

```go
func DiffPNGFiles(leftPath, rightPath string, opts PixelDiffOptions) (PixelDiffResult, error)
```

This isolates image-level testing from browser-level testing.

### 7.4 Artifact writer behavior

All artifact writers should create parent directories:

```go
if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { ... }
```

This addresses the Pyxis redirection/output ergonomics pain indirectly and makes JS helpers safer.

## 8. JavaScript adapter design

File:

```text
internal/cssvisualdiff/jsapi/compare.go
```

Install under existing module exports:

```go
func installCompareAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
    compare := vm.NewObject()
    _ = compare.Set("region", func(call goja.FunctionCall) goja.Value { ... })
    _ = exports.Set("compare", compare)
}
```

### 8.1 Strict unwrapping

The adapter should unwrap `left` and `right` as locator handles:

```go
leftHandle, err := unwrapProxyBacking[locatorHandle](vm, defaultProxyRegistry, "css-visual-diff.compare.region", leftValue, "cvd.locator")
```

If we later add `regionHandle`, the API can accept either:

```go
left := unwrapLocatorOrRegion(...)
```

### 8.2 Page locking

Each locator handle already references a `pageState`. Region comparison touches two pages. Avoid deadlocks.

Do **not** lock both pages with nested locks unless lock ordering is carefully defined. Instead:

1. capture left side under `left.page.runExclusive`,
2. capture right side under `right.page.runExclusive`,
3. run pure image diff after both captures are complete.

This serializes per-page CDP operations while allowing the implementation to remain simple and safe.

If left and right are on the same page, this still works because the operations are sequential.

### 8.3 Promise behavior

`cvd.compare.region(...)` returns a Promise. Errors should be JS-visible and ideally use existing typed error classes:

- selector missing/invalid → `SelectorError`,
- screenshot/artifact failure → `ArtifactError`,
- browser/CDP failure → `BrowserError`,
- invalid options/argument types → `TypeError` or future `CvdError` subclass.

### 8.4 Comparison result handle and lowering

`cvd.compare.region(...)` should return a Go-backed `cvd.comparison` handle rather than a raw Go struct. The handle owns the immutable comparison data plus any cached image buffers/temp paths needed to materialize artifacts.

Methods on the handle should include:

```go
comparison.summary()
comparison.toJSON()
comparison.report.markdown()
comparison.report.writeMarkdown(path)
comparison.artifact(name).write(path)
comparison.artifacts.write(outDir, names)
```

The handle must still lower to plain lowerCamel data for output and persisted JSON. Do not expose raw Go structs through `vm.ToValue` if they expose snake_case fields.

Add lowering helpers:

```go
func lowerCompareSummary(result service.CompareRegionResult) map[string]any
func lowerCompareRegionResult(result service.CompareRegionResult) map[string]any
```

If possible, implement a `toJSON` method so `JSON.stringify(comparison)` behaves sensibly inside scripts. Documentation should still teach explicit `comparison.toJSON()` for command returns because jsverbs/glazed output conversion may not call JavaScript `toJSON` in every path.

## 9. Migration of existing built-ins

The builtin script currently says:

```js
return require("diff").compareRegion({ ... })
```

There are three options.

### Option A: Keep internal helper and add public API separately

Pros:

- Minimum breakage.
- Fastest P0 implementation.

Cons:

- Two paths can drift.

### Option B: Re-implement builtin script using `require("css-visual-diff")`

Example:

```js
async function region(targets, viewport, output, selectors) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  try {
    const left = await browser.page(targets.leftUrl, { viewport, waitMs: targets.leftWaitMs })
    const right = await browser.page(targets.rightUrl, { viewport, waitMs: targets.rightWaitMs })
    return await cvd.compare.region({
      left: left.locator(selectors.leftSelector),
      right: right.locator(selectors.rightSelector || selectors.leftSelector),
      threshold: output.threshold,
      outDir: output.outDir,
      artifacts: outputToArtifacts(output),
      evidence: legacyEvidenceDefaults(),
    })
  } finally {
    await browser.close()
  }
}
```

Pros:

- Builtin scripts dogfood the public API.
- Docs and implementation converge.

Cons:

- Requires async builtin verb behavior to be fully supported and tested for this command.
- Output shape may change unless legacy lowering is preserved.

### Option C: Both paths call the same service

Pros:

- Good short-term compromise.
- No public/internal drift at the browser/image layer.
- CLI output compatibility can be preserved while JS API returns the new schema.

Cons:

- Still leaves `require("diff")` around as an internal helper.

Recommendation: implement Option C first, then consider Option B once compatibility questions are settled.

## 10. YAML config/job bridge design

Pyxis also asks for:

```js
const job = await cvd.jobFromConfig(path)
await job.preflight({ side: 'react' })
await job.inspect({ side: 'original', artifacts: 'css-json', outDir })
await job.compareAll({ outDir })
```

This is valuable but larger than P0.

### Recommended direction

Add a `cvd.config.job(...)` or `cvd.job.fromConfig(...)` namespace later:

```js
const job = await cvd.job.fromConfig("archive-desktop.css-visual-diff.yml")

const plan = job.plan()
const preflight = await job.preflight({ side: "react" })
const result = await job.run({ modes: ["capture", "pixeldiff", "cssdiff"], outDir })
```

But do not build this until compare-region service extraction is done. The job bridge should call official runner/service logic, not reimplement YAML semantics in JS.

### Why not do this first

The immediate blocker is pixel comparison from JS. A job bridge is migration ergonomics. It also has a larger compatibility surface: modes, config variants, output directories, prepare behavior, reports, and runner semantics.

## 11. Structural diff tolerances and CSS normalization

Pyxis asks for tolerances:

```js
cvd.diff(before, after, {
  tolerances: {
    'results[*].snapshot.bounds.y': 4,
  },
})
```

and CSS normalization:

```js
cvd.normalize.css(styleMap, {
  colors: true,
  zeroUnits: true,
  fontFamilies: 'primary-family-only',
})
```

These are good additions, but separate from pixel compare.

### Tolerance design

Extend structural diff options:

```go
type DiffOptions struct {
    IgnorePaths []string `json:"ignorePaths,omitempty"`
    Tolerances  []PathTolerance `json:"tolerances,omitempty"`
}

type PathTolerance struct {
    Path string  `json:"path"`
    Abs  float64 `json:"abs"`
}
```

Use explicit arrays rather than object maps if wildcard matching becomes richer.

### CSS normalization design

Start with pure helpers:

```js
const normalized = cvd.normalize.css(styleMap, {
  colors: "rgb",
  zeroUnits: true,
  fontFamilies: "primary",
})
```

Then `cvd.diffStyles` can be a convenience later.

Do not add JavaScript callback comparators in Goja v1 of this feature. Callback-based comparison is powerful, but it complicates deterministic reports and error handling.

## 12. Documentation plan

Update:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
internal/cssvisualdiff/doc/topics/javascript-verbs.md
internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md
```

Required doc changes:

1. Add `cvd.compare.region(...)` to module exports.
2. Explain that `cvd.diff(...)` is structural JSON diff, while `cvd.compare.region(...)` is pixel/region comparison.
3. Document `CompareRegionResultV1` schema.
4. Add a Pyxis-like example comparing prototype and Storybook pages.
5. In `javascript-verbs.md`, explain that built-in `script compare region` is backed by internal compatibility helpers and that custom scripts should prefer `require("css-visual-diff")`.
6. Document artifact output directory behavior and parent-directory creation.

## 13. Test plan

### Service tests

Add tests for:

- equal PNG files produce zero changed pixels,
- scalar threshold behavior,
- different-size images normalize/pad correctly,
- missing selectors return a selector/artifact error,
- result schema fields are populated,
- output artifact paths are created under `outDir`.

### JS API tests

Add tests for:

- `cvd.compare.region({ left: page.locator(...), right: page.locator(...) })`,
- rejection of raw `{ page, selector }` objects in strict mode,
- helpful type mismatch message mentioning `page.locator(selector)`,
- result uses lowerCamel `changedPercent`, not snake_case,
- artifacts include JSON/Markdown paths when requested,
- parent directories are created for writes.

### Binary smoke

Add ticket script:

```text
scripts/001-compare-region-js-api-smoke.sh
```

It should:

1. build or run the binary,
2. start two tiny local HTML pages,
3. run a repository-scanned JS verb that uses `cvd.compare.region`,
4. assert `compare.json`, `compare.md`, `diff_only.png`, and `diff_comparison.png` exist,
5. assert output JSON contains `schemaVersion`, `changedPercent`, and artifact paths.

## 14. Implementation phases

### Phase 0 — Documentation/design ticket setup

- [x] Read Pyxis source feedback.
- [x] Create this docmgr ticket.
- [x] Write source analysis and design docs.
- [ ] Relate relevant source files.
- [ ] Run docmgr doctor.

### Phase 1 — Service extraction for image diffing

- [ ] Move image diff primitives out of `modes/compare.go` into `internal/cssvisualdiff/service`.
- [ ] Preserve existing compare command behavior.
- [ ] Add service tests for PNG diff behavior.

### Phase 2 — Service-level region comparison

- [ ] Add `service.CompareRegions` using existing driver/page/style helpers.
- [ ] Support region screenshots, optional full screenshots, pixel diff, optional computed styles, optional matched-style winners.
- [ ] Add stable service result type independent of legacy `modes.CompareResult`.

### Phase 3 — JS API `cvd.compare.region`

- [ ] Add `internal/cssvisualdiff/jsapi/compare.go`.
- [ ] Install `cvd.compare.region` under `require("css-visual-diff")`.
- [ ] Require locator handles for left/right.
- [ ] Return Promise and lowerCamel schema.
- [ ] Add JS API tests.

### Phase 4 — Built-in and CLI convergence

- [ ] Route internal `require("diff").compareRegion` through the new service.
- [ ] Preserve existing `verbs script compare region` output compatibility or document any intentional change.
- [ ] Consider migrating builtin script to public `cvd.compare.region`.

### Phase 5 — Docs and smoke tests

- [ ] Update embedded JS API docs.
- [ ] Add pixel-accuracy guide example.
- [ ] Add ticket smoke script.
- [ ] Run `go test ./...` and smoke scripts.

### Phase 6 — Follow-up API work

- [ ] Evaluate `cvd.compare.sections(...)` after userland proves shape.
- [ ] Design `cvd.job.fromConfig(...)` / `cvd.runConfig(...)` with runner parity.
- [ ] Add structural diff tolerances.
- [ ] Add CSS normalization helpers.

## 15. Preferred minimal implementation for P0

If we must keep the first implementation small, do this and no more:

```js
await cvd.compare.region({
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
  outDir,
  artifacts: ["screenshots", "diff", "json", "markdown"],
})
```

Return:

```js
{
  schemaVersion: "cssvd.compare.region.v1",
  equal: false,
  threshold: 30,
  changedPercent: 7.128146453089244,
  changedPixels: 102172,
  totalPixels: 1433360,
  normalizedWidth: 920,
  normalizedHeight: 1558,
  artifacts: {
    outDir,
    leftRegion: ".../left_region.png",
    rightRegion: ".../right_region.png",
    diffOnly: ".../diff_only.png",
    diffComparison: ".../diff_comparison.png",
    json: ".../compare.json",
    markdown: ".../compare.md"
  },
  left: { selector: "#root > *", url: "..." },
  right: { selector: "[data-page='archive']", url: "..." }
}
```

Computed and matched-style evidence can be added immediately if service extraction is easy, but should not block the pixel primitive.

## 16. Non-goals

Do not implement these in the first pass:

- Pyxis page registry DSL.
- Storybook URL builder.
- accepted-difference policy store.
- CI policy bands.
- custom report templates.
- JavaScript callback comparators for structural diffs.
- full YAML runner job bridge.

They are useful, but they should build on the core primitive.

## 17. Final recommendation

Implement `cvd.compare.region(...)` as the public, strict, Promise-returning region pixel comparison primitive under `require("css-visual-diff")`. Build it on extracted Go services so existing CLI modes and internal built-in verbs converge with the public API. Return a versioned lowerCamel schema and write artifacts with predictable paths.

This is the smallest change that unblocks Pyxis while making the API better for future users. It gives userland enough power to build registries, multi-section loops, CI policy, and reports without pushing project-specific concepts into core.
