---
Title: Beta user ergonomics for JavaScript visual workflows
Ticket: CSSVD-BETA-ERGONOMICS
Status: active
Topics:
  - tooling
  - frontend
  - visual-regression
  - browser-automation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
  - Path: /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md
    Note: Source maintainer follow-up request list from Pyxis after validating cvd.compare.region.
  - Path: /home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/internal/cssvisualdiff/jsapi/locator.go
    Note: Current locator handle methods; natural home for locator.waitFor.
  - Path: /home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/internal/cssvisualdiff/jsapi/module.go
    Note: Current page methods; possible home for page.waitForSelector and page serialization behavior.
  - Path: /home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/internal/cssvisualdiff/service/dom.go
    Note: Current locator DOM service functions; natural home for service-level wait helper.
  - Path: /home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/internal/cssvisualdiff/jsapi/compare.go
    Note: Current comparison artifact writer and compare.region implementation.
  - Path: /home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/internal/cssvisualdiff/service/collection.go
    Note: Collection profile model and normalization behavior to document for beta users.
---

# Beta user ergonomics for JavaScript visual workflows

This document turns the Pyxis maintainer follow-up request list into a focused `css-visual-diff` implementation plan. The goal is not to add a large new framework. The flexible JavaScript API already solved the core blocker: project-local scripts can now call `cvd.compare.region(...)`, `cvd.compare.selections(...)`, and `cvd.image.diff(...)` directly.

The remaining high-priority work is about beta-user ergonomics: make the API easier to use correctly in real Storybook/app workflows, make artifacts easier to return from project-local verbs, and provide one complete multi-section catalog example that users can copy.

> [!summary]
> Implement only the small pieces that remove real beta-user friction:
>
> 1. Add `locator.waitFor(...)` and optionally `page.waitForSelector(...)` as thin readiness helpers.
> 2. Make `comparison.artifacts.write(...)` return a stable artifact path map.
> 3. Add one complete multi-section comparison + catalog example and docs.
> 4. Clarify collection profile defaults and `styleProps` / `attributes` semantics.
>
> Defer tolerances, CSS normalization hooks, and style presets until beta users have used the core workflow longer.

## 1. Source context

The source request document is:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md
```

It reports that Pyxis validated the new API successfully:

```text
YAML Archive content:                 7.1281%
Built-in script compare region smoke: 7.128146453089244%
New cvd.compare.region userland verb: 7.128146453089244%
Changed pixels:                       102172
```

That means the old “please expose a JS-callable pixel comparison primitive” request is satisfied. The new request list is about making the API easier and safer for project-local workflows.

## 2. Prioritization: what to implement now versus later

The source document includes P0 through P3 requests. For this ticket, we should implement P0 and the low-complexity P1 items. We should not implement the P2/P3 feature ideas yet.

| Source priority | Request | Decision | Rationale |
|---|---|---|---|
| P0 | Selector wait helper for JS scripts | Implement now | Directly fixes Storybook/app readiness races. Small API, high value. |
| P0 | Clarify artifact write/report schema | Implement now | Directly helps project-local verbs return paths without filename guessing. Small change. |
| P1 | Multi-section comparison/catalog example | Implement now | Documentation/example work; helps beta users copy the intended pattern. |
| P1 | Document collection profiles and defaults | Implement now | Documentation work; avoids incorrect assumptions about performance and diagnostics. |
| P2 | Tolerances for structural/bounds diffs | Defer | Policy semantics need more real usage. Userland can do this today. |
| P2 | CSS/style normalization hooks | Defer | Potentially complex and opinionated; userland can normalize today. |
| P3 | Built-in style property presets | Defer | Convenience only; easy to add later once vocabulary stabilizes. |
| P3 | Explicit no-op prepare docs | Fold into wait-helper docs | If wait helpers exist, no-op prepare becomes fallback documentation only. |

The guiding principle is:

> Prefer small ergonomic helpers and examples over introducing a second visual workflow abstraction.

## 3. Current API state relevant to this ticket

### 3.1 Locator methods

Current locator methods live in:

```text
internal/cssvisualdiff/jsapi/locator.go
```

The current methods are:

```js
locator.status()
locator.exists()
locator.visible()
locator.text(options?)
locator.bounds()
locator.computedStyle(props)
locator.attributes(names)
locator.collect(options?)
```

The locator is already the right object for selector readiness because it is page-bound and selector-bound.

### 3.2 Page methods

Current page methods are installed in:

```text
internal/cssvisualdiff/jsapi/module.go
```

Current relevant methods:

```js
page.locator(selector)
page.goto(url, options?)
page.prepare(spec)
page.preflight(probes)
page.inspect(probe, options?)
page.inspectAll(probes, options?)
page.close()
```

`page.prepare(...)` can already wait for arbitrary JS expressions through `waitFor`, but it is too verbose for the common “wait until this selector exists” case.

### 3.3 DOM service functions

Current locator services live in:

```text
internal/cssvisualdiff/service/dom.go
```

Current functions include:

```go
LocatorStatus(page, locator)
LocatorText(page, locator, opts)
LocatorHTML(page, locator, outer)
LocatorBounds(page, locator)
LocatorAttributes(page, locator, attrs)
LocatorComputedStyle(page, locator, props)
```

The new wait helper should be implemented here first, then exposed through JS.

### 3.4 Comparison artifact writing

Current comparison artifact writing lives in:

```text
internal/cssvisualdiff/jsapi/compare.go
```

`comparison.artifacts.write(outDir, ["json", "markdown"])` writes predictable files:

```text
compare.json
compare.md
```

It currently returns a generic result map with written file paths, but the source user feedback indicates the returned/write schema is not clear enough and does not present all artifact paths in the most convenient shape for compact wrapper output.

The desired beta-user shape is:

```js
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])

// easy to return from project-local verbs
return {
  changedPercent: comparison.pixel.summary().changedPercent,
  compareJson: written.json,
  compareMarkdown: written.markdown,
  diffComparison: written.diffComparison,
}
```

## 4. Design principle: small helpers, not a workflow framework

The source request asks for beta-user ergonomics. The wrong response would be to add a full workflow builder or policy engine:

```js
cvd.workflow().section(...).wait(...).compare(...).catalog(...).run()
```

Do not do this.

JavaScript is already the workflow language. Users can write loops, arrays, helper functions, and project-specific policy. The core should provide missing primitives and clear examples.

The target API should stay small:

```js
await page.locator(selector).waitFor({ timeoutMs: 30000, visible: true })
const comparison = await cvd.compare.region({ left, right, outDir })
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])
catalog.record(comparison, target)
```

That is enough for beta users to build robust project-local CLIs.

## 5. Feature 1: `locator.waitFor(...)`

### Problem

Real apps often load the root document before the meaningful comparison selector exists. This is especially common for:

- Storybook iframe stories,
- MSW-backed stories,
- RTK Query and async data loading,
- route-level lazy loading,
- app shells where the root appears before the target section.

Current workaround:

```js
await page.prepare({
  type: "script",
  waitFor: 'document.querySelector("[data-page=\\"archive\\"]")',
  waitForTimeoutMs: 30000,
  script: "void 0",
  afterWaitMs: 500,
})
```

This is too much ceremony for a common case.

### Proposed public API

Primary API:

```js
await page.locator('[data-page="archive"]').waitFor({
  timeoutMs: 30000,
  visible: true,
  afterWaitMs: 500,
})
```

Options:

```ts
type WaitForOptions = {
  timeoutMs?: number      // default 5000
  pollIntervalMs?: number // default 100
  visible?: boolean       // default false; true means require visible status
  afterWaitMs?: number    // default 0; sleep after condition becomes true
}
```

Return value:

```js
{
  selector: "[data-page='archive']",
  exists: true,
  visible: true,
  bounds: { x, y, width, height },
  elapsedMs: 742,
}
```

Failure:

```text
SelectorError: css-visual-diff.locator.waitFor: selector "[data-page='archive']" did not become visible within 30000ms
```

### Optional secondary API

Add only if it is trivial after `locator.waitFor`:

```js
await page.waitForSelector(selector, options)
```

This should delegate to `page.locator(selector).waitFor(options)` conceptually. If there is concern about API surface area, skip `page.waitForSelector` and document only `locator.waitFor`.

### Service-level design

Add to `internal/cssvisualdiff/service/dom.go`:

```go
type WaitForSelectorOptions struct {
    TimeoutMS      int  `json:"timeoutMs,omitempty"`
    PollIntervalMS int  `json:"pollIntervalMs,omitempty"`
    Visible        bool `json:"visible,omitempty"`
    AfterWaitMS    int  `json:"afterWaitMs,omitempty"`
}

type WaitForSelectorResult struct {
    Selector  string          `json:"selector"`
    Exists    bool            `json:"exists"`
    Visible   bool            `json:"visible"`
    Bounds    *Bounds         `json:"bounds,omitempty"`
    TextStart string          `json:"textStart,omitempty"`
    ElapsedMS int             `json:"elapsedMs"`
}

func WaitForLocator(page *driver.Page, locator LocatorSpec, opts WaitForSelectorOptions) (WaitForSelectorResult, error) {
    opts = normalizeWaitOptions(opts)
    deadline := time.Now().Add(time.Duration(opts.TimeoutMS) * time.Millisecond)
    started := time.Now()

    for {
        status, err := LocatorStatus(page, locator)
        if err != nil {
            return WaitForSelectorResult{}, err
        }
        if status.Error != "" {
            return WaitForSelectorResult{}, fmt.Errorf("selector %q: %s", locator.Selector, status.Error)
        }
        ready := status.Exists && (!opts.Visible || status.Visible)
        if ready {
            if opts.AfterWaitMS > 0 {
                page.Wait(time.Duration(opts.AfterWaitMS) * time.Millisecond)
            }
            return waitResultFromStatus(status, time.Since(started)), nil
        }
        if time.Now().After(deadline) {
            return waitResultFromStatus(status, time.Since(started)), fmt.Errorf("selector %q did not become %s within %dms", locator.Selector, waitCondition(opts), opts.TimeoutMS)
        }
        page.Wait(time.Duration(opts.PollIntervalMS) * time.Millisecond)
    }
}
```

The service should use existing `LocatorStatus` rather than inventing a second DOM status implementation.

### JS adapter design

In `internal/cssvisualdiff/jsapi/locator.go`, add method:

```go
"waitFor": locator.waitFor(ctx, vm),
```

Implementation sketch:

```go
func (l *locatorHandle) waitFor(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
    return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
        raw := exportOptionalObject(vm, "css-visual-diff.locator.waitFor", call.Argument(0))
        return promiseValue(ctx, vm, "css-visual-diff.locator.waitFor", func() (any, error) {
            return l.page.runExclusive(func() (any, error) {
                opts, err := decodeInto[service.WaitForSelectorOptions](raw)
                if err != nil { return nil, err }
                result, err := service.WaitForLocator(l.page.page.Page(), l.spec(), opts)
                if err != nil { return nil, err }
                return lowerWaitForSelectorResult(result), nil
            })
        }, nil)
    }
}
```

This preserves existing per-page serialization. A script can call `Promise.all` across pages, while each individual page remains safe.

### Tests

Service tests:

```text
internal/cssvisualdiff/service/dom_test.go
```

Add tests for:

- selector already exists,
- selector appears after a short timeout via `setTimeout`,
- visible option waits for visibility,
- timeout produces a helpful error,
- invalid selector returns a selector-like error.

JS integration tests:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Add a small repository-scanned verb that:

```js
const waited = await page.locator('#delayed').waitFor({ timeoutMs: 2000 })
return { exists: waited.exists, selector: waited.selector }
```

### Documentation

Update:

- `internal/cssvisualdiff/doc/topics/javascript-api.md`
- `internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md`
- `examples/verbs/README.md` if the example uses it.

Document the fallback no-op prepare pattern only as a fallback:

```js
await page.prepare({ type: "script", waitFor: "...", script: "void 0" })
```

## 6. Feature 2: stable artifact write result

### Problem

Beta users need compact stdout JSON from project-local verbs:

```json
{
  "section": "content",
  "changedPercent": 7.1281,
  "compareJson": ".../compare.json",
  "compareMarkdown": ".../compare.md",
  "diffComparison": ".../diff_comparison.png"
}
```

They can guess `compare.json` and `compare.md`, but they should not have to.

### Proposed public API

Keep the current call:

```js
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])
```

Clarify and stabilize the returned shape:

```js
{
  json: ".../compare.json",              // present if json was requested/written
  markdown: ".../compare.md",           // present if markdown was requested/written
  leftRegion: ".../left_region.png",    // present if known from comparison artifacts
  rightRegion: ".../right_region.png",  // present if known
  diffOnly: ".../diff_only.png",        // present if known
  diffComparison: ".../diff_comparison.png", // present if known
  written: [".../compare.json", ".../compare.md"] // optional compatibility/list view
}
```

Important: do not rewrite PNGs in `artifacts.write`. Region PNGs and diff PNGs are produced by `cvd.compare.region(...)` / `service.CompareSelections`. `artifacts.write` should only write requested JSON/Markdown and return all known artifact paths.

### Implementation sketch

Current implementation is in:

```text
internal/cssvisualdiff/jsapi/compare.go
```

Extend `writeComparisonArtifacts` to return a stable map:

```go
func writeComparisonArtifacts(outDir string, data service.SelectionComparisonData, names []string) (map[string]any, error) {
    result := map[string]any{}
    written := []string{}

    if wants("json") {
        path := filepath.Join(outDir, "compare.json")
        write JSON
        result["json"] = path
        written = append(written, path)
    }

    if wants("markdown") {
        path := filepath.Join(outDir, "compare.md")
        write Markdown
        result["markdown"] = path
        written = append(written, path)
    }

    for _, artifact := range data.Artifacts {
        switch artifact.Kind {
        case "leftRegion":
            result["leftRegion"] = artifact.Path
        case "rightRegion":
            result["rightRegion"] = artifact.Path
        case "diffOnly":
            result["diffOnly"] = artifact.Path
        case "diffComparison":
            result["diffComparison"] = artifact.Path
        }
    }

    if data.Pixel != nil {
        if data.Pixel.DiffOnlyPath != "" { result["diffOnly"] = data.Pixel.DiffOnlyPath }
        if data.Pixel.DiffComparisonPath != "" { result["diffComparison"] = data.Pixel.DiffComparisonPath }
    }

    result["written"] = written
    return result, nil
}
```

If artifact kind strings differ, use the actual constants/values from `service.SelectionArtifact`. Do not add a new artifact model just for this.

### Tests

Add/extend tests in:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Expected JS:

```js
const comparison = await cvd.compare.region({ left, right, outDir })
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])
return {
  json: !!written.json,
  markdown: !!written.markdown,
  diffComparison: !!written.diffComparison,
  writtenCount: written.written.length,
}
```

Assertions:

- `written.json` ends with `compare.json`,
- `written.markdown` ends with `compare.md`,
- `written.diffComparison` ends with `diff_comparison.png`,
- files exist.

### Documentation

Update `javascript-api` to explicitly list the return shape.

## 7. Feature 3: complete multi-section comparison + catalog example

### Problem

Beta users need page-level workflows, not only one selector. They need a complete recommended pattern for:

- loading left and right pages once,
- waiting for each section,
- comparing multiple sections,
- writing per-section artifacts,
- recording comparisons into a catalog,
- writing `manifest.json` and `index.md`,
- returning compact stdout JSON.

### Proposed example file

Add:

```text
examples/verbs/compare-page-catalog.js
```

Suggested public command:

```bash
css-visual-diff verbs --repository examples/verbs examples compare page-catalog \
  http://localhost:3000/original \
  http://localhost:3000/react \
  /tmp/cssvd-page-catalog \
  --output json
```

Example shape:

```js
async function comparePageCatalog(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  const catalog = cvd.catalog.create({
    title: "Page comparison",
    outDir,
    artifactRoot: "artifacts",
  })

  const sections = [
    { name: "page", leftSelector: "#root", rightSelector: "#root" },
    { name: "content", leftSelector: "#root > *", rightSelector: "[data-page='archive']" },
  ]

  let leftPage, rightPage
  try {
    leftPage = await browser.page(leftUrl, { viewport: { width: 920, height: 1460 }, waitMs: 1000 })
    rightPage = await browser.page(rightUrl, { viewport: { width: 920, height: 1460 }, waitMs: 1000 })

    const summaries = []
    for (const section of sections) {
      await leftPage.locator(section.leftSelector).waitFor({ timeoutMs: 30000 })
      await rightPage.locator(section.rightSelector).waitFor({ timeoutMs: 30000 })

      const artifactDir = catalog.artifactDir(section.name)
      const comparison = await cvd.compare.region({
        name: section.name,
        left: leftPage.locator(section.leftSelector),
        right: rightPage.locator(section.rightSelector),
        outDir: artifactDir,
        threshold: 30,
        inspect: "rich",
      })

      const written = await comparison.artifacts.write(artifactDir, ["json", "markdown"])
      catalog.record(comparison, {
        slug: section.name,
        name: section.name,
        url: leftUrl,
        selector: section.leftSelector,
      })

      summaries.push({
        section: section.name,
        pixel: comparison.pixel.summary(),
        artifacts: written,
      })
    }

    return {
      summaries,
      manifestPath: await catalog.writeManifest(),
      indexPath: await catalog.writeIndex(),
      catalog: catalog.summary(),
    }
  } finally {
    if (leftPage) await leftPage.close()
    if (rightPage) await rightPage.close()
    await browser.close()
  }
}
```

If `locator.waitFor` is not implemented in the same PR, the example should not be added yet or should use the fallback `page.prepare` pattern. Prefer implementing wait first, then adding this example.

### Smoke script

Add a ticket-local smoke:

```text
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

The smoke should:

1. Start a tiny local HTTP server with two pages.
2. Run the example verb.
3. Assert `manifest.json` exists.
4. Assert `index.md` exists.
5. Assert at least two section artifact directories exist.
6. Assert the stdout JSON includes artifact paths from `artifacts.write`.

## 8. Feature 4: collection profile documentation

### Problem

Docs list `minimal`, `rich`, and `debug`, but beta users need recommended defaults and semantics.

### Proposed docs

Add to `javascript-api` and `pixel-accuracy-scripting-guide`:

| Profile | Use when | Avoid when |
|---|---|---|
| `minimal` | CI only needs existence/bounds/pixels. | You need text/style diagnosis. |
| `rich` | Default authoring/review mode. | Very large suites where style extraction is too expensive. |
| `debug` | Deep one-off diagnosis with HTML/all styles/all attrs. | Routine CI/page-suite runs. |

Clarify:

- `cvd.compare.region(...)` defaults to `inspect: "rich"`.
- `styleProps` affects which style properties are collected/compared for targeted comparison workflows.
- `attributes` affects which attributes are collected/compared.
- If users want broad post-hoc diagnosis, use `inspect: "rich"`; if users want a lean CI pass, choose `minimal` and collect more only after failure.

Before documenting exact rich defaults, verify `normalizeCollectOptions` in `service/collection.go` and document the actual behavior, not a guessed list.

## 9. Explicitly deferred ideas

### 9.1 Bounds tolerances

Do not implement yet.

Userland can do this today:

```js
function within(delta, tolerance) {
  return Math.abs(delta || 0) <= tolerance
}

const bounds = comparison.bounds.diff()
const ok =
  within(bounds.delta.x, 1) &&
  within(bounds.delta.y, 4) &&
  within(bounds.delta.width, 2) &&
  within(bounds.delta.height, 8)
```

Reason to defer: tolerance semantics belong to project policy. We should collect a few beta examples before baking them into reports.

### 9.2 CSS normalization hooks

Do not implement yet.

Userland can normalize selected style maps before policy checks. Official normalization quickly becomes opinionated: colors, zero units, font stacks, line-height, resolved values, and prototype-vs-React differences all have edge cases.

Reason to defer: high complexity, unclear defaults.

### 9.3 Style property presets

Do not implement yet.

Users can define:

```js
const typography = ["font-family", "font-size", "font-weight", "line-height", "letter-spacing", "color"]
```

Reason to defer: convenience only; presets should reflect real usage vocabulary.

## 10. Implementation phases

### Phase 1 — Selector wait helper

- Add `WaitForSelectorOptions` and `WaitForSelectorResult` to `service/dom.go`.
- Add `WaitForLocator(page, locator, opts)` service function.
- Add `locator.waitFor(options?)` to `jsapi/locator.go`.
- Optionally add `page.waitForSelector(selector, options?)` in `jsapi/module.go` if it remains tiny.
- Add service and JS integration tests.
- Update docs.

### Phase 2 — Artifact write result shape

- Update `writeComparisonArtifacts` to return stable keyed paths.
- Include known PNG artifacts from `SelectionComparisonData.Artifacts` and pixel paths.
- Add tests for returned paths and file existence.
- Update docs.

### Phase 3 — Example and smoke

- Add `examples/verbs/compare-page-catalog.js`.
- Update `examples/verbs/README.md`.
- Add ticket smoke script.
- Add any necessary command tests if not already covered.

### Phase 4 — Profile docs

- Verify actual `normalizeCollectOptions` behavior.
- Update embedded docs with decision table and semantics.
- Add notes to example README.

## 11. Acceptance criteria

This ticket is done when:

- `locator.waitFor(...)` works in repository-scanned JS verbs.
- Selector wait timeout errors are clear enough for beta users and coding agents.
- `comparison.artifacts.write(...)` returns stable `json`, `markdown`, and known image artifact paths.
- A complete multi-section catalog example exists and is smoked.
- Collection profiles are documented with recommended defaults and caveats.
- P2/P3 ideas are explicitly deferred in docs/ticket notes, not silently mixed into the implementation.
- Validation passes:

```bash
go test ./... -count=1
make lint
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
css-visual-diff help javascript-api
css-visual-diff help pixel-accuracy-scripting-guide
```

## 12. Review guidance

Reviewers should check for scope control. This ticket should not introduce:

- workflow builders,
- global project policy engines,
- CSS normalization frameworks,
- bounds tolerance report semantics,
- large new config schemas.

The right patch should feel small: one wait primitive, one better artifact result, one complete example, and clearer docs.
