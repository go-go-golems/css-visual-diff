---
Title: Origin main port guide for hair booking branch primitives
Ticket: CSSVD-PORT-HAIR-PRIMITIVES
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - merge-planning
    - pyxis
    - visual-diff
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/compare-region.js
      Note: Pyxis userland file that must migrate away from cvd.compare.region
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: |-
        Registers require("diff").compareRegion and should not be bypassed by adding a competing public compare API.
        registers require(diff).compareRegion
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: |-
        Candidate destination for the hair-booking locator.waitFor primitive.
        destination for locator.waitFor port
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: |-
        Shows the clean origin/main JavaScript module surface and where any accepted primitive must be registered.
        origin/main JavaScript API registration point
    - Path: internal/cssvisualdiff/modes/compare.go
      Note: |-
        Existing origin/main URL/selector compare-region engine; should remain the public high-level compare path.
        existing origin/main compare-region engine
    - Path: internal/cssvisualdiff/service/dom.go
      Note: |-
        Candidate destination for selector wait service support.
        selector status and wait service implementation area
    - Path: internal/cssvisualdiff/service/pixel.go
      Note: Proposed new internal service file copied/adapted from hair-booking.
    - Path: prototype-design/visual-diff/userland/lib/compare-region.js
      Note: Pyxis userland currently targets the old cvd.compare.region API and needs adaptation to origin/main.
ExternalSources: []
Summary: Guidance for origin/main maintainers on which hair-booking branch pieces to port, which to reject, and how to implement the accepted ports without reintroducing bespoke comparison APIs.
LastUpdated: 2026-05-04T17:05:00-04:00
WhatFor: Give colleagues and interns a concrete implementation guide for salvaging only the useful primitives from the divergent hair-booking branch.
WhenToUse: Before touching the merge conflict, porting branch code, updating Pyxis visual-diff userland, or reviewing a PR that brings hair-booking code into origin/main.
---


# Origin/main Port Guide for Hair-booking Branch Primitives

## Executive Summary

The `bookmark/2026-05-01/hair-booking` branch contains real engineering work, but most of it should not be merged into `origin/main`. The branch developed a convenience layer around `cvd.compare.region`, `cvd.compare.selections`, `locator.collect`, and selection comparison objects. That layer was useful for early Pyxis visual-diff experiments, but it conflicts with the current `origin/main` direction: a smaller, explicit, fluent JavaScript API made of browser, page, locator, probe, extractor, snapshot, diff, report, and catalog primitives.

The porting strategy should therefore be conservative:

- **Port primitives that strengthen the origin/main model.** The best candidates are `locator.waitFor(...)`, the extracted `service/pixel.go` internals, and stable artifact path behavior if main does not already return the paths needed by operators.
- **Do not port bespoke public workflow APIs.** Avoid reintroducing `cvd.compare.region`, `cvd.compare.selections`, `locator.collect`, `cvd.collect.selection`, and the old YAML-native config/run pipeline.
- **Adapt Pyxis userland to main, not main to Pyxis.** Pyxis currently calls APIs from the hair-booking branch (`cvd.compare.region`, `cvd.catalog.create`, `catalog.record(comparison)`). That is a migration task for Pyxis. Main already has `require("diff").compareRegion` for high-level URL/selector pixel comparisons and `cvd.extract`/`cvd.snapshot`/`cvd.diff` for fluent semantic checks.

This document explains the architecture, gives file references, and lays out a phased implementation plan for an intern working on `origin/main`.

---

## 1. Problem Statement and Scope

The repository has a divergent branch named `bookmark/2026-05-01/hair-booking`. It diverged before `origin/main` settled on its current JavaScript-first API. A previous merge attempt left the worktree conflicted. The task is not to resolve that merge. The task is to decide what, if anything, should be ported cleanly onto `origin/main`.

The central design question is:

> Should origin/main regain higher-level helper APIs such as `cvd.compare.region`, or should it preserve the smaller fluent primitives it currently exposes?

The answer recommended here is:

> Preserve the origin/main primitive model. Port only low-level, reusable pieces that make that model better.

### In scope

This document covers:

- The current `origin/main` JS API model.
- The hair-booking branch API additions.
- What to port, what to reject, and why.
- How to implement the accepted ports.
- How Pyxis userland should be adapted.
- Test and validation steps.

### Out of scope

This document does not ask an intern to:

- Complete the existing broken merge.
- Reintroduce the old YAML `run --config` pipeline.
- Replace `require("diff").compareRegion` with `cvd.compare.region`.
- Rewrite the React review site.
- Make Go core know anything Pyxis-specific.

---

## 2. Current-state Architecture in `origin/main`

`origin/main` is built around a small JavaScript module exposed as:

```javascript
const cvd = require("css-visual-diff")
```

The module registers browser and artifact primitives, not project-specific workflows. The public exports in the current docs are:

```text
cvd.browser(options?)
cvd.catalog(options)
cvd.viewport(width, height)
cvd.target(name)
cvd.probe(name)
cvd.extractors.*
cvd.extract(locator, extractors)
cvd.snapshot(page, probes, options?)
cvd.diff(before, after, options?)
cvd.report(diff)
cvd.write.json(path, value)
cvd.write.markdown(path, markdown)
cvd.CvdError / cvd.SelectorError / cvd.PrepareError / cvd.BrowserError / cvd.ArtifactError
```

Evidence:

- `origin/main:internal/cssvisualdiff/jsapi/module.go` registers the module and installs target, probe, extractor, extract, snapshot, diff, catalog, and browser APIs.
- `origin/main:internal/cssvisualdiff/doc/topics/javascript-api.md` documents the exports and does not list `cvd.compare.region`.
- `origin/main:examples/verbs/low-level-inspect.js` demonstrates direct locator/extractor/snapshot usage.
- `origin/main:examples/verbs/catalog-inspect-page.js` demonstrates `cvd.catalog(...)`, `page.preflight(...)`, `page.inspect(...)`, and `catalog.addResult(...)`.

### 2.1 The fluent primitive model

The current mainline mental model is:

```text
Browser opens pages.
Page exposes locators.
Locator reads facts from one loaded DOM element.
Probe is a reusable recipe for later snapshotting.
Extractor is a strict typed request for one kind of element fact.
Snapshot applies probes to a page and returns plain data.
Diff compares plain data.
Report renders a diff.
Catalog organizes durable artifacts and manifests.
```

A typical script in this model looks like:

```javascript
async function inspect(url, selector, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  try {
    const page = await browser.page(url, {
      viewport: cvd.viewport(800, 600),
      waitMs: 250,
      name: "inspect-target",
    })

    const element = await cvd.extract(page.locator(selector), [
      cvd.extractors.exists(),
      cvd.extractors.visible(),
      cvd.extractors.text(),
      cvd.extractors.bounds(),
      cvd.extractors.computedStyle(["display", "color", "font-size"]),
      cvd.extractors.attributes(["id", "class"]),
    ])

    await cvd.write.json(`${outDir}/element.json`, element)
    return element
  } finally {
    await browser.close()
  }
}
```

This design is explicit. It is a little more verbose than a one-call comparison helper, but the script makes every step visible: which page, which locator, which facts, and where evidence is written.

### 2.2 High-level region comparison already exists outside the `cvd` module

`origin/main` also has a native DSL module:

```javascript
const diff = require("diff")
const result = diff.compareRegion({ ... })
```

This is registered by:

```text
origin/main:internal/cssvisualdiff/dsl/registrar.go
```

The native function takes URLs, selectors, viewport, output options, computed CSS properties, and attributes. It then calls:

```go
modes.GenerateCompareResult(ctx.Context, settings)
modes.WriteCompareArtifacts(result, settings.WriteJSON, settings.WriteMarkdown)
```

The key point is that `origin/main` is not missing the ability to compare two regions. It already has a URL/selector comparison command path. What it intentionally does not have is a second public `cvd.compare.region` API inside the fluent `css-visual-diff` module.

### 2.3 Why this matters

If we port `cvd.compare.region`, we create two public ways to perform a region comparison:

```text
require("diff").compareRegion({ left: { url, selector }, right: { url, selector }, ... })
cvd.compare.region({ left: leftPage.locator(...), right: rightPage.locator(...), ... })
```

That may be useful, but it also fragments the API. The mainline project appears to have chosen explicit primitives plus one URL/selector compare operation. A port should respect that decision unless maintainers explicitly decide to add a convenience layer.

---

## 3. Hair-booking Branch Additions

The hair-booking branch added several APIs and services. They fall into three categories.

### 3.1 Clean primitives

These fit the current mainline model:

- `locator.waitFor(...)`
- `service/pixel.go`
- stable artifact path behavior, if not already present in main

### 3.2 Useful internals with questionable public API shape

These contain useful logic but should not necessarily be ported as public APIs:

- deterministic bounds/text/style/attribute diff helpers from `service/selection_compare.go`
- some tests around pixel and semantic comparison behavior

### 3.3 Bespoke workflow APIs

These are convenient but cut against the mainline API direction:

- `cvd.compare.region`
- `cvd.compare.selections`
- `locator.collect`
- `cvd.collect.selection`
- `service/collection.go` with `minimal` / `rich` / `debug` profiles
- `catalog.record(comparison)`

### 3.4 Legacy architecture that must stay out

These were removed or superseded in `origin/main` and should not return:

- `internal/cssvisualdiff/config/*`
- `internal/cssvisualdiff/jsapi/config.go`
- `internal/cssvisualdiff/runner/runner.go`
- old standalone native modes such as `capture.go`, `inspect.go`, `pixeldiff.go`, `prepare.go`, `stories.go`
- old Pyxis YAML examples in the css-visual-diff repo

---

## 4. Decision Matrix: What to Port

| Candidate | Port? | Why |
|---|---:|---|
| `locator.waitFor(...)` | **Yes** | Clean locator primitive, useful for authoring and CI, aligns with explicit page/locator model. |
| `service/pixel.go` | **Yes** | Internal extraction of pixel operations; improves tests and reuse without changing public API. |
| Stable artifact paths | **Maybe** | Good behavior, but only port gaps not already covered by `modes.CompareResult`. |
| Small semantic diff helpers | **Maybe** | Useful if extracted as internal utilities, not as a new public comparison object model. |
| `cvd.compare.region` | **No by default** | Duplicates/competes with `require("diff").compareRegion` and bundles too much policy/workflow into `cvd`. |
| `cvd.compare.selections` | **No by default** | Parallel abstraction to `extract`/`snapshot`/`diff`; only useful if maintainers choose a convenience layer. |
| `locator.collect` / `cvd.collect.selection` | **No by default** | Parallel abstraction to explicit extractors and probes. Profiles can live in userland helpers. |
| `service/collection.go` | **No wholesale** | Tied to collect/profile concept; prefer explicit extractors or JS-side presets. |
| `catalog.record(comparison)` | **No as-is** | Tied to comparison object API; consider a generic artifact/result record later if needed. |
| Old YAML config/native modes | **No** | Directly contradicts mainline direction. |

---

## 5. Recommended Port 1: `locator.waitFor(...)`

### 5.1 Why it should be ported

Waiting for selectors is not a project-specific workflow. It is a basic locator operation. Scripts already need to do this before extraction, screenshotting, or assertions. Without `locator.waitFor`, userland must either poll manually or rely on implicit waits inside heavier operations.

The API should live exactly where the branch placed it: on `page.locator(selector)`.

Target public API:

```javascript
await page.locator('[data-section="hero"]').waitFor({
  timeoutMs: 30000,
  pollIntervalMs: 100,
  visible: true,
  afterWaitMs: 500,
})
```

Expected behavior:

- Poll the locator until it exists.
- If `visible !== false`, require visibility as well.
- Stop after `timeoutMs`.
- Sleep `afterWaitMs` after success to allow layout/animation stabilization.
- Return a status object or `true`. Prefer returning the final status object because that is more useful to scripts.

### 5.2 Files to inspect and port from

Branch source:

```text
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/jsapi/locator.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/dom.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/dom_test.go
```

Destination files in `origin/main`:

```text
internal/cssvisualdiff/jsapi/locator.go
internal/cssvisualdiff/service/dom.go
internal/cssvisualdiff/service/dom_test.go
internal/cssvisualdiff/doc/topics/javascript-api.md
internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md
examples/verbs/low-level-inspect.js (optional example update)
```

### 5.3 Go service pseudocode

```go
type WaitForSelectorOptions struct {
    TimeoutMS      int  `json:"timeoutMs,omitempty"`
    PollIntervalMS int `json:"pollIntervalMs,omitempty"`
    Visible        bool `json:"visible,omitempty"`
    AfterWaitMS    int `json:"afterWaitMs,omitempty"`
}

func WaitForLocator(ctx context.Context, page *driver.Page, locator LocatorSpec, opts WaitForSelectorOptions) (SelectorStatus, error) {
    timeout := defaultDuration(opts.TimeoutMS, 30*time.Second)
    poll := defaultDuration(opts.PollIntervalMS, 100*time.Millisecond)
    deadline := time.Now().Add(timeout)

    for {
        status, err := LocatorStatus(page, locator)
        if err == nil && status.Error == "" && status.Exists {
            if opts.Visible == false || status.Visible {
                if opts.AfterWaitMS > 0 {
                    page.Wait(time.Duration(opts.AfterWaitMS) * time.Millisecond)
                }
                return status, nil
            }
        }

        if time.Now().After(deadline) {
            return status, fmt.Errorf("timed out waiting for selector %q", locator.Selector)
        }
        time.Sleep(poll)
    }
}
```

### 5.4 JSAPI pseudocode

```go
func (l *locatorHandle) waitFor(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
    return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
        rawOptions := exportOptionalObject(vm, "css-visual-diff.locator.waitFor", call.Argument(0))
        return promiseValue(ctx, vm, "css-visual-diff.locator.waitFor", func() (any, error) {
            return l.page.runExclusive(func() (any, error) {
                opts, err := decodeInto[service.WaitForSelectorOptions](rawOptions)
                if err != nil { return nil, err }
                status, err := service.WaitForLocator(ctx.Context, l.page.page.Page(), l.spec(), opts)
                if err != nil { return nil, err }
                return lowerSelectorStatus(status), nil
            })
        }, nil)
    }
}
```

### 5.5 Tests

Add or port tests for:

- selector exists immediately;
- selector appears after a short delay;
- selector exists but hidden when `visible: true`;
- selector exists but hidden when `visible: false`;
- invalid selector returns a useful error;
- timeout behavior is deterministic enough for CI.

Validation command:

```bash
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli
```

---

## 6. Recommended Port 2: Internal Pixel Service

### 6.1 Why it should be ported

Pixel comparison is internal machinery. `origin/main` already does pixel comparison through `modes.GenerateCompareResult`. The branch improves the implementation by extracting PNG and pixel math into a reusable service package.

This is not an API design change. It is internal cleanup.

### 6.2 Branch source

```text
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/pixel.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/pixel_test.go
```

Important functions:

```go
ValidatePixelThreshold(threshold int) error
ReadPNG(path string) (image.Image, error)
WritePNG(path string, img image.Image) error
ToNRGBA(img image.Image) *image.NRGBA
PadToSameSize(a, b image.Image) (*image.NRGBA, *image.NRGBA)
ComputePixelDiff(left, right *image.NRGBA, threshold int) (PixelDiffResult, *image.NRGBA)
DiffImages(left, right image.Image, opts PixelDiffOptions) (...)
DiffPNGFiles(leftPath, rightPath string, opts PixelDiffOptions) (...)
CombineSideBySide(left, right, diff *image.NRGBA) *image.NRGBA
WritePixelDiffImages(leftPath, rightPath, diffComparisonPath, diffOnlyPath string, opts PixelDiffOptions) (PixelDiffResult, error)
```

### 6.3 Destination files

Add:

```text
internal/cssvisualdiff/service/pixel.go
internal/cssvisualdiff/service/pixel_test.go
```

Then refactor:

```text
internal/cssvisualdiff/modes/pixeldiff_util.go
internal/cssvisualdiff/modes/compare.go
```

### 6.4 Keep output compatibility

Be careful with JSON naming. `origin/main`'s compare result currently has snake_case fields in `modes.PixelDiffStats`:

```go
type PixelDiffStats struct {
    Threshold          int     `json:"threshold"`
    TotalPixels        int     `json:"total_pixels"`
    ChangedPixels      int     `json:"changed_pixels"`
    ChangedPercent     float64 `json:"changed_percent"`
    NormalizedWidth    int     `json:"normalized_width"`
    NormalizedHeight   int     `json:"normalized_height"`
    DiffComparisonPath string  `json:"diff_comparison_path"`
    DiffOnlyPath       string  `json:"diff_only_path"`
}
```

The branch `service.PixelDiffResult` uses camelCase:

```go
type PixelDiffResult struct {
    ChangedPercent     float64 `json:"changedPercent"`
    DiffComparisonPath string  `json:"diffComparisonPath,omitempty"`
    DiffOnlyPath       string  `json:"diffOnlyPath,omitempty"`
}
```

Do not accidentally change `require("diff").compareRegion` output shape. If `service.PixelDiffResult` uses camelCase internally, convert explicitly when returning `modes.PixelDiffStats`.

Pseudocode:

```go
func pixelDiffStatsFromService(result service.PixelDiffResult) PixelDiffStats {
    return PixelDiffStats{
        Threshold: result.Threshold,
        TotalPixels: result.TotalPixels,
        ChangedPixels: result.ChangedPixels,
        ChangedPercent: result.ChangedPercent,
        NormalizedWidth: result.NormalizedWidth,
        NormalizedHeight: result.NormalizedHeight,
        DiffComparisonPath: result.DiffComparisonPath,
        DiffOnlyPath: result.DiffOnlyPath,
    }
}
```

### 6.5 Tests

Run:

```bash
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes
```

Also run a smoke comparison through the CLI or built-in verb to ensure output artifacts still exist:

```bash
css-visual-diff verbs --repository examples/verbs examples review-sweep summary \
  --spec-file examples/specs/review-sweep.example.yaml \
  --out-dir /tmp/cssvd-review-sweep-smoke \
  --output json
```

If that specific command needs servers or fixtures, substitute the existing repo smoke command used by CI.

---

## 7. Recommended Port 3: Stable Artifact Path Reliability

### 7.1 Why this is a behavior to verify, not necessarily a file to port

The branch commit `94c8544 feat: return stable comparison artifact paths` improved the branch comparison object so scripts could reliably find:

```text
left_region.png
right_region.png
diff_only.png
diff_comparison.png
compare.json
compare.md
```

`origin/main` already returns many of these paths through `modes.CompareResult`:

```go
CompareSideResult.ElementScreenshot
PixelDiffStats.DiffComparisonPath
PixelDiffStats.DiffOnlyPath
CompareInputs.OutDir
```

Before porting any code, inspect the actual JSON from `require("diff").compareRegion`. If it already includes all paths Pyxis/operator scripts need, do nothing.

### 7.2 Acceptance criteria

A region comparison result should let an agent or human find these artifacts without guessing:

```json
{
  "inputs": { "out_dir": "/tmp/run" },
  "url1": { "element_screenshot": "/tmp/run/url1_region.png" },
  "url2": { "element_screenshot": "/tmp/run/url2_region.png" },
  "pixel_diff": {
    "diff_only_path": "/tmp/run/diff_only.png",
    "diff_comparison_path": "/tmp/run/diff_comparison.png"
  }
}
```

If current output fails this, update `modes.GenerateCompareResult` / artifact writing to fill paths consistently.

### 7.3 Do not port the proxy object just for paths

Do not introduce:

```javascript
comparison.artifacts.write(...)
comparison.artifacts.list()
```

just to solve path discoverability. Stable fields in the existing result object are enough.

---

## 8. Optional Internal Diff Utilities

The branch `service/selection_compare.go` includes useful small algorithms:

- deterministic key ordering for map diffs;
- include/exclude filters for style and attribute names;
- bounds delta calculation;
- text changed calculation.

These are good utilities, but the public `SelectionComparisonData` type is not aligned with mainline's current API.

If current code needs these operations, extract small functions rather than porting the whole comparison object model.

Possible internal helpers:

```go
func DiffBounds(left, right *Bounds) BoundsDiff
func DiffStringMaps(left, right map[string]string, include, exclude []string) []MapValueDiff
func DiffText(left, right string) TextDiff
```

Use them only if they simplify existing `modes/compare.go`, review-sweep adapters, or future snapshot diff work.

Do not add public JS APIs around them unless maintainers explicitly decide to add a convenience comparison layer.

---

## 9. APIs Not to Port

### 9.1 Do not port `cvd.compare.region`

It is convenient, but it is a bundled workflow API. It takes live locators and produces a rich proxy object. That shape belongs to the old branch's design, not the current `origin/main` primitive model.

It also duplicates the high-level operation already available as:

```javascript
require("diff").compareRegion({ ... })
```

The mainline approach should be:

- use `require("diff").compareRegion` for URL/selector pixel comparison;
- use `cvd.extract`, `cvd.snapshot`, `cvd.diff`, and `cvd.report` for fluent semantic checks;
- use JS userland helper functions if a project wants a convenience wrapper.

### 9.2 Do not port `cvd.compare.selections`

This API depends on the old branch's collected selection object. It creates a parallel world next to `extract` and `snapshot`.

Instead of:

```javascript
const left = await locator.collect({ inspect: "rich" })
const right = await locator.collect({ inspect: "rich" })
const comparison = await cvd.compare.selections(left, right)
```

mainline should prefer:

```javascript
const left = await cvd.extract(leftLocator, extractors)
const right = await cvd.extract(rightLocator, extractors)
const diff = cvd.diff(left, right)
```

If property-aware diffing is needed, add userland normalization/tolerance helpers or improve `cvd.diff` generically.

### 9.3 Do not port `locator.collect` / `cvd.collect.selection`

The collect API hides an opinionated profile system:

```text
minimal
rich
debug
```

That is ergonomic, but it is less explicit than extractors and probes. It also moves policy-like choices about "rich enough" into Go core.

Prefer JS-side helpers:

```javascript
function richExtractors(cvd, props, attrs) {
  return [
    cvd.extractors.exists(),
    cvd.extractors.visible(),
    cvd.extractors.text(),
    cvd.extractors.bounds(),
    cvd.extractors.computedStyle(props),
    cvd.extractors.attributes(attrs),
  ]
}
```

### 9.4 Do not port old native YAML config

Do not port:

```text
internal/cssvisualdiff/config/*
internal/cssvisualdiff/jsapi/config.go
cvd.config.load
cvd.loadConfig
```

Pyxis may still use YAML visual specs, but those are userland data loaded through verb field handling (`objectFromFile`). That is different from resurrecting the old Go-native `run --config` pipeline.

### 9.5 Do not port old modes and runner

Do not port:

```text
internal/cssvisualdiff/runner/runner.go
internal/cssvisualdiff/modes/capture.go
internal/cssvisualdiff/modes/inspect.go
internal/cssvisualdiff/modes/pixeldiff.go
internal/cssvisualdiff/modes/prepare.go
internal/cssvisualdiff/modes/stories.go
```

Those files belong to the architecture `origin/main` intentionally replaced.

---

## 10. Pyxis Migration Guidance

Pyxis currently lives at:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland
```

Its current `lib/compare-region.js` was written against the hair-booking convenience API. For `origin/main`, it should be adapted.

### 10.1 Current Pyxis shape

Current Pyxis code opens pages and calls:

```javascript
const comparison = await cvd.compare.region({
  name: target.page + '-' + section.name,
  left: leftPage.locator(section.original),
  right: rightPage.locator(section.react),
  threshold,
  inspect: 'rich',
  outDir,
  styleProps: DEFAULT_STYLE_PROPS,
  attributes: DEFAULT_ATTRIBUTES,
})

const written = await comparison.artifacts.write(outDir, ['json', 'markdown'])
catalog.record(comparison, target)
```

This will not work on plain `origin/main`.

### 10.2 Recommended Pyxis pixel comparison path

For pixel comparisons, Pyxis should use the existing native compare module:

```javascript
const diff = require('diff')

const result = diff.compareRegion({
  left: {
    url: target.prototypeUrl,
    selector: section.original,
    waitMs: target.waitMs,
  },
  right: {
    url: target.storybookUrl,
    selector: section.react,
    waitMs: target.waitMs,
  },
  viewport: target.viewport,
  output: {
    outDir,
    threshold,
    writeJson: true,
    writeMarkdown: true,
    writePngs: true,
  },
  computed: DEFAULT_STYLE_PROPS,
  attributes: DEFAULT_ATTRIBUTES,
})
```

Then Pyxis should adapt the returned object into its policy row shape.

Pseudocode:

```javascript
function rowFromCompareRegion(target, section, outDir, result) {
  const pixel = result.pixel_diff || {}
  return policies.withClassification({
    page: target.page,
    variant: target.variant,
    section: section.name,
    outDir,
    leftUrl: target.prototypeUrl,
    rightUrl: target.storybookUrl,
    leftSelector: section.original,
    rightSelector: section.react,
    changedPercent: pixel.changed_percent,
    changedPixels: pixel.changed_pixels,
    totalPixels: pixel.total_pixels,
    diffOnlyPath: pixel.diff_only_path,
    diffComparisonPath: pixel.diff_comparison_path,
    leftRegionPath: result.url1 && result.url1.element_screenshot,
    rightRegionPath: result.url2 && result.url2.element_screenshot,
    artifactJson: outDir + '/compare.json',
    artifactMarkdown: outDir + '/compare.md',
    source: 'require("diff").compareRegion',
  })
}
```

### 10.3 Recommended Pyxis semantic debugging path

For atom-to-page debugging, Pyxis should use current mainline primitives:

```javascript
const left = await cvd.extract(leftPage.locator(section.original), richExtractors(cvd, props, attrs))
const right = await cvd.extract(rightPage.locator(section.react), richExtractors(cvd, props, attrs))
const semanticDiff = cvd.diff(left, right)
await cvd.write.json(`${outDir}/semantic-diff.json`, semanticDiff)
await cvd.report(semanticDiff).writeMarkdown(`${outDir}/semantic-diff.md`)
```

This keeps pixel comparison and semantic comparison separate:

```text
Pixel compare: require("diff").compareRegion
Semantic compare: cvd.extract → cvd.diff → cvd.report
```

That split matches the mainline API better than a single `cvd.compare.region` object that tries to do everything.

### 10.4 Pyxis accepted differences bug to fix separately

While reviewing Pyxis, we noticed that top-level accepted differences in specs may not be merged into target sections. `userland/lib/registry.js` currently computes accepted differences from the individual target config, not from the whole spec.

The intended merge order should be:

```text
spec.acceptedDifferences[page][section]
target.acceptedDifferences[section]
section.acceptedDifferences
```

Pseudocode:

```javascript
function acceptedForSection(spec, target, section) {
  return [].concat(
    lookup(spec.acceptedDifferences, [target.page, section.name]) || [],
    lookup(target.acceptedDifferences, [section.name]) || [],
    section.acceptedDifferences || []
  )
}
```

This is a Pyxis-side issue and should not drive mainline API design.

---

## 11. Implementation Plan

### Phase 0: Start from clean `origin/main`

Do not continue from the conflicted worktree. Create a clean branch:

```bash
git fetch origin
git switch -c task/port-hair-booking-primitives origin/main
```

Use `git show` to copy files or hunks from the hair-booking branch:

```bash
git show bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/pixel.go > /tmp/pixel.go
```

### Phase 1: Port `locator.waitFor`

1. Copy/adapt `WaitForSelectorOptions` and wait service logic from the branch.
2. Add `waitFor` to `locatorHandle` methods in `jsapi/locator.go`.
3. Add docs and a small example.
4. Add service and JSAPI tests.

Acceptance criteria:

```javascript
await page.locator('#late').waitFor({ timeoutMs: 5000, visible: true })
```

works and returns a useful final status.

### Phase 2: Port internal pixel service

1. Add `service/pixel.go` and tests.
2. Refactor `modes/pixeldiff_util.go` to delegate.
3. Refactor `modes/compare.go` only where it simplifies code.
4. Preserve existing JSON output shape.

Acceptance criteria:

```bash
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes
```

passes, and a compare-region smoke still writes the same artifacts.

### Phase 3: Audit artifact path output

1. Run or inspect `require("diff").compareRegion` output.
2. Confirm it returns paths for JSON, Markdown, left screenshot, right screenshot, diff-only PNG, and diff-comparison PNG.
3. If any paths are missing, add fields to the existing output shape.

Acceptance criteria:

A JS caller can build an operator row without guessing filenames.

### Phase 4: Optional internal diff utilities

Only do this if needed by existing code.

1. Extract deterministic map diff or bounds-delta helpers.
2. Keep them in `service`, not `jsapi`.
3. Add unit tests.

Acceptance criteria:

The utilities reduce duplication or improve tests without adding new public APIs.

### Phase 5: Pyxis adaptation PR

In the Pyxis repo:

1. Replace `cvd.compare.region` calls in `userland/lib/compare-region.js` with `require("diff").compareRegion` for pixel comparisons.
2. Replace `catalog.record(comparison)` with Pyxis-side suite summary/catalog handling or `catalog.addResult` if the result fits.
3. Keep `cvd.extract`/`cvd.snapshot` for semantic snapshot flows.
4. Fix accepted-differences merging.
5. Run Pyxis smoke scripts.

Acceptance criteria:

Pyxis userland works against a clean `origin/main` binary without `cvd.compare.region`.

---

## 12. Test Strategy

### 12.1 Go unit tests

Run targeted tests first:

```bash
go test ./internal/cssvisualdiff/service
```

Then broader package tests:

```bash
go test ./internal/cssvisualdiff/modes ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli
```

Finally, if practical:

```bash
go test ./...
```

### 12.2 JSAPI smoke script for `locator.waitFor`

Create or adapt an example verb that:

1. starts from an HTML fixture where an element appears after a timeout;
2. calls `locator.waitFor`;
3. extracts text and bounds;
4. returns structured success.

Expected output:

```json
{
  "ok": true,
  "exists": true,
  "visible": true,
  "text": "Loaded"
}
```

### 12.3 Compare-region smoke for existing mainline API

Validate that this still works:

```javascript
require("diff").compareRegion({
  left: { url: leftUrl, selector: '#target' },
  right: { url: rightUrl, selector: '#target' },
  viewport: { width: 800, height: 600 },
  output: { outDir, threshold: 30, writeJson: true, writeMarkdown: true, writePngs: true },
})
```

Check that artifacts are written and returned paths are stable.

### 12.4 Pyxis smoke tests

After adapting Pyxis:

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland
./scripts/smoke-list-targets.sh
./scripts/smoke-inspect-section-archive.sh
./scripts/smoke-compare-section-archive.sh
./scripts/smoke-compare-page-archive.sh
./scripts/smoke-compare-spec-archive.sh
./scripts/smoke-snapshot-section-archive.sh
./scripts/smoke-diff-snapshots-archive.sh
```

If servers are required, start prototype and Storybook first as described in the Pyxis userland README.

---

## 13. Review Checklist for Colleagues

When reviewing a PR that ports code from hair-booking, check these points:

- [ ] Does the PR avoid `cvd.compare.region` and `cvd.compare.selections` unless maintainers explicitly approved a convenience layer?
- [ ] Does the PR avoid importing `internal/cssvisualdiff/config`?
- [ ] Does the PR avoid resurrecting old native mode files?
- [ ] Does `locator.waitFor` fit the same Promise/error model as existing locator methods?
- [ ] Does pixel service extraction preserve compare result JSON shape?
- [ ] Are artifact paths stable and visible in structured output?
- [ ] Are tests copied/adapted from the branch where useful?
- [ ] Is Pyxis compatibility handled in Pyxis userland rather than by bloating mainline core?

---

## 14. Alternatives Considered

### Alternative A: Merge the branch and resolve conflicts

Rejected. The branch contains legacy YAML config, old native modes, and public APIs that main appears to have moved away from. A full merge would regress architecture clarity.

### Alternative B: Port all branch JS APIs

Rejected by default. `cvd.compare.region`, `cvd.compare.selections`, and `collect` are convenient but duplicate or compete with the explicit mainline primitive model.

### Alternative C: Port only Pyxis compatibility shims

Rejected. Main should not absorb branch APIs merely because one userland was written against them. Pyxis should adapt to the current mainline API.

### Alternative D: Port no code

Rejected. `locator.waitFor` and `service/pixel.go` are genuinely useful, low-risk improvements that align with mainline design.

---

## 15. References

### Mainline files

```text
origin/main:internal/cssvisualdiff/jsapi/module.go
origin/main:internal/cssvisualdiff/jsapi/locator.go
origin/main:internal/cssvisualdiff/jsapi/probe.go
origin/main:internal/cssvisualdiff/jsapi/extract.go
origin/main:internal/cssvisualdiff/jsapi/diff.go
origin/main:internal/cssvisualdiff/jsapi/catalog.go
origin/main:internal/cssvisualdiff/service/extract.go
origin/main:internal/cssvisualdiff/modes/compare.go
origin/main:internal/cssvisualdiff/dsl/registrar.go
origin/main:internal/cssvisualdiff/doc/topics/javascript-api.md
origin/main:internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md
origin/main:examples/verbs/low-level-inspect.js
origin/main:examples/verbs/catalog-inspect-page.js
origin/main:examples/verbs/review-sweep.js
```

### Hair-booking files and commits

```text
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/jsapi/locator.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/dom.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/pixel.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/pixel_test.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/selection_compare.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/jsapi/compare.go
bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/jsapi/collect.go

a370857 feat: add js selector wait helpers
6ca2498 feat: extract pixel diff service primitives
94c8544 feat: return stable comparison artifact paths
29c8aca feat: add selection comparison service
88ddac5 feat: add canonical compare region js workflow
```

### Pyxis files

```text
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/README.md
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/compare-region.js
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/inspect.js
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/snapshot.js
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/registry.js
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/verbs/pyxis-pages.js
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/specs/app.pages.desktop.visual.yml
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/specs/app.components.visual.yml
```
