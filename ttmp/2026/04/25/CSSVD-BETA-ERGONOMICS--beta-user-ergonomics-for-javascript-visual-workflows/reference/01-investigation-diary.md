---
Title: Investigation diary
Ticket: CSSVD-BETA-ERGONOMICS
Status: active
Topics:
  - tooling
  - frontend
  - visual-regression
  - browser-automation
DocType: reference
Intent: long-term
Owners: []
---

# Diary

## Goal

This diary records the design and implementation work for `CSSVD-BETA-ERGONOMICS`: a small, beta-user-focused follow-up to the flexible JavaScript API. The goal is to help real project users, especially Pyxis/Storybook users, without turning `css-visual-diff` into a large workflow framework.

The work deliberately prioritizes small primitives, stable artifact metadata, examples, and documentation over policy-heavy features such as tolerance engines, CSS normalization frameworks, or style preset registries.

## Step 1: Create the beta ergonomics ticket and scope the work

The first step was to translate the Pyxis maintainer follow-up request document into a local `css-visual-diff` ticket. The source document made it clear that the original major blocker — a JS-callable region/pixel comparison primitive — had already been solved by `cvd.compare.region(...)`, `cvd.compare.selections(...)`, and `cvd.image.diff(...)`.

The important design move was therefore to avoid reopening the whole API design. Instead, I scoped the new ticket around the high-value beta-user pain points that remained: selector readiness, artifact path metadata, a complete multi-section catalog example, and collection profile documentation.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to address the high priority items from /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md

Add analysis + design guide. We want to avoid adding too much complexity, but want to help our beta users."

**Assistant interpretation:** Create a focused ticket in this repo from the Pyxis follow-up request list, analyze the requested features, and write a design guide that keeps scope small.

**Inferred user intent:** Help beta users succeed with the new JS API by implementing practical ergonomics, while explicitly avoiding a large, complex workflow system.

**Commit (code):** `a370857` later created the ticket docs as part of the Phase 1 commit — "feat: add js selector wait helpers".

### What I did

- Read the Pyxis source document:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md
```

- Created the local ticket:

```bash
docmgr ticket create-ticket --root ./ttmp \
  --ticket CSSVD-BETA-ERGONOMICS \
  --title "Beta user ergonomics for JavaScript visual workflows" \
  --topics tooling,frontend,visual-regression,browser-automation
```

- Inspected the implementation points that would matter for the design:
  - `internal/cssvisualdiff/jsapi/locator.go`
  - `internal/cssvisualdiff/jsapi/module.go`
  - `internal/cssvisualdiff/service/dom.go`
  - `internal/cssvisualdiff/jsapi/compare.go`
  - `internal/cssvisualdiff/service/collection.go`
- Wrote:
  - `design/01-beta-user-ergonomics-for-js-visual-workflows.md`
  - `tasks.md`
  - `changelog.md`
  - this diary file.
- Related key implementation files with `docmgr doc relate`.
- Ran:

```bash
docmgr doctor --root ./ttmp --ticket CSSVD-BETA-ERGONOMICS --stale-after 30
```

### Why

The Pyxis request list contained both small, concrete needs and larger, policy-heavy ideas. Without scoping, the follow-up could easily turn into an overly ambitious feature set. The design guide draws a line: implement selector waits, stable artifact paths, examples, and docs now; defer bounds tolerances, CSS normalization, and style presets until beta usage clarifies the right semantics.

### What worked

- The source document was already clean and prioritized, which made it straightforward to extract P0/P1 items.
- The existing architecture had natural extension points:
  - locator readiness belongs in `service/dom.go` and `jsapi/locator.go`,
  - artifact write metadata belongs in `jsapi/compare.go`,
  - multi-section workflows belong in examples/userland JavaScript.
- `docmgr doctor` passed after removing one problematic external related-file entry.

### What didn't work

- I initially related the external Pyxis source file directly via `docmgr doc relate`. `docmgr doctor` reported it as missing because the relative path from this repo's ticket workspace crossed into another workspace in a way it did not validate:

```text
[WARNING] missing_related_file — related file not found: ../../../../../../../../../code/wesen/2026-04-23--pyxis/...
```

- I removed that relation and kept the external source path in the design guide and diary instead.

### What I learned

- Cross-workspace source documents are better recorded as prose references or external sources unless `docmgr` can validate the path cleanly.
- The right beta-user response is a small set of helpers and examples, not a new abstraction layer.

### What was tricky to build

The tricky part was not code; it was scope control. The source document included tempting P2/P3 features such as tolerances and CSS normalization. Those are reasonable future features, but they encode project policy. Implementing them too early would make the API feel heavier and more opinionated before beta users have shown enough patterns.

I handled this by explicitly adding a deferred section to the design guide and tasks. That makes the deferral intentional rather than accidental.

### What warrants a second pair of eyes

- Review whether the proposed `page.waitForSelector(...)` convenience is worth the additional surface area or whether `locator.waitFor(...)` alone would have been enough.
- Review whether the artifact path schema proposed in the design guide is sufficiently stable for downstream wrappers.

### What should be done in the future

- Implement the ticket in small commits at phase boundaries.
- Keep P2/P3 ideas deferred unless beta users repeatedly ask for the same policy behavior.

### Code review instructions

Start with:

```text
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/design/01-beta-user-ergonomics-for-js-visual-workflows.md
```

Then compare the proposed phases against the source request list. Validate ticket hygiene with:

```bash
docmgr doctor --root ./ttmp --ticket CSSVD-BETA-ERGONOMICS --stale-after 30
```

### Technical details

The ticket scope chosen from the source list was:

```text
Implement now:
- locator.waitFor / page.waitForSelector
- stable comparison.artifacts.write result map
- multi-section catalog example
- collection profile docs

Defer:
- bounds tolerance API
- CSS/style normalization hooks
- style presets
```

## Step 2: Implement selector wait helpers

The first implementation phase added the missing readiness primitive. Before this step, users could emulate selector waiting with `page.prepare({ type: "script", waitFor: "...", script: "void 0" })`, but that was too awkward for the common case of waiting for a Storybook/app selector to appear.

The new API keeps the solution small: `locator.waitFor(...)` is the primary method, and `page.waitForSelector(...)` is a thin convenience for users who expect a Playwright-like shape. Both call the same service-level polling helper.

### Prompt Context

**User prompt (verbatim):** "Alright, add tasks, and implement step by step, committing at appropriate intervals, keeping a detailed diary."

**Assistant interpretation:** Proceed with the ticket implementation in phases, update task/diary/changelog as work progresses, validate each phase, and commit at appropriate milestones.

**Inferred user intent:** Turn the design ticket into working code while preserving a clear implementation trail for review and continuation.

**Commit (code):** `a370857` — "feat: add js selector wait helpers"

### What I did

- Added service types in `internal/cssvisualdiff/service/dom.go`:
  - `WaitForSelectorOptions`
  - `WaitForSelectorResult`
- Added `service.WaitForLocator(page, locator, opts)` as a polling loop over existing `LocatorStatus`.
- Added `locator.waitFor(options?)` to `internal/cssvisualdiff/jsapi/locator.go`.
- Added `page.waitForSelector(selector, options?)` to `internal/cssvisualdiff/jsapi/module.go`.
- Extended `internal/cssvisualdiff/service/dom_test.go` with coverage for:
  - already-existing selectors,
  - delayed selectors,
  - `visible: true`,
  - timeout behavior,
  - invalid selectors.
- Extended `internal/cssvisualdiff/verbcli/command_test.go` so the repository-scanned locator smoke calls both:
  - `cta.waitFor({ timeoutMs: 1000 })`
  - `page.waitForSelector("#cta", { timeoutMs: 1000 })`
- Updated embedded docs:
  - `internal/cssvisualdiff/doc/topics/javascript-api.md`
  - `internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md`
- Ran:

```bash
gofmt -w internal/cssvisualdiff/service/dom.go \
  internal/cssvisualdiff/jsapi/locator.go \
  internal/cssvisualdiff/jsapi/module.go \
  internal/cssvisualdiff/service/dom_test.go \
  internal/cssvisualdiff/verbcli/command_test.go

go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli -count=1
go test ./... -count=1
make lint
```

- Updated ticket tasks, changelog, and this diary.
- Committed the phase.

### Why

Real app pages often load their root document before the meaningful comparison region is ready. This is especially common for Storybook iframes, MSW-backed stories, RTK Query, lazy-loaded routes, and prototype shells. A direct selector wait helper removes the need for users to remember the `page.prepare(..., script: "void 0")` workaround.

### What worked

- Reusing `LocatorStatus` kept the service implementation small and avoided a second selector-status implementation.
- The page-level and locator-level JS methods both fit into the existing `pageState.runExclusive(...)` serialization model.
- Service tests were able to use an `httptest.Server` page with `setTimeout(...)` to verify delayed selector appearance and visibility changes.
- Full validation passed:

```bash
go test ./... -count=1
make lint
```

### What didn't work

The first JS integration attempt returned the Go struct directly from the JS wrapper. In JavaScript, this produced Go field names rather than lowerCamel JSON field names. The test tried to read:

```js
waited.exists
```

but the value was undefined, and the Go test failed with:

```text
Error:       Not equal:
             expected: bool(true)
             actual  : <nil>(<nil>)
Test:        TestCVDModuleExposesLocatorMethods
```

I fixed this by lowering `WaitForSelectorResult` through `lowerJSON(...)` before returning it from both JS wrappers.

### What I learned

- Returning Go structs directly to Goja is not the same as returning the JSON-shaped public API. For public JS objects, use the same lowerCamel lowering path used elsewhere in the JS API.
- Small helpers still need integration tests through `verbcli`, because service tests alone would not catch JS field casing issues.

### What was tricky to build

The subtle issue was field casing at the Go/JS boundary. `WaitForSelectorResult` has Go fields like `Exists`, but the public JS API expects JSON/lowerCamel fields like `exists`. The symptom was a nil field in a JS verb output row, not a compile error.

The fix was to treat wait results like other public data handles: lower via JSON before returning to JS. That preserved the documented shape and made the test pass.

### What warrants a second pair of eyes

- Timeout semantics: `timeoutMs: 0` currently means “use default 5000ms”, not “no timeout”. Review whether that matches user expectations.
- Polling uses `page.Wait(...)`, which maps to Chromedp sleep inside the page context. This is consistent with existing code, but review whether a context-aware sleep would be preferable later.
- `page.waitForSelector(...)` adds API surface. Confirm it is worth keeping alongside `locator.waitFor(...)`.

### What should be done in the future

- Consider documenting timeout defaults in examples more visibly.
- If beta users need wait conditions beyond existence/visibility, add them cautiously; do not turn this into a general workflow scheduler.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/service/dom.go
internal/cssvisualdiff/jsapi/locator.go
internal/cssvisualdiff/jsapi/module.go
```

Then review tests:

```text
internal/cssvisualdiff/service/dom_test.go
internal/cssvisualdiff/verbcli/command_test.go
```

Validate with:

```bash
go test ./... -count=1
make lint
```

### Technical details

Public API:

```js
await page.locator('[data-page="archive"]').waitFor({
  timeoutMs: 30000,
  pollIntervalMs: 100,
  visible: true,
  afterWaitMs: 500,
})

await page.waitForSelector('[data-page="archive"]', {
  timeoutMs: 30000,
  visible: true,
})
```

Returned shape:

```js
{
  selector: '[data-page="archive"]',
  exists: true,
  visible: true,
  bounds: { x: 0, y: 0, width: 920, height: 1400 },
  textStart: "Archive",
  elapsedMs: 742
}
```

## Step 3: Return stable artifact paths from `comparison.artifacts.write(...)`

The second implementation phase addressed a concrete beta-user integration problem: project-local commands need to return compact JSON rows with artifact paths. Before this change, users could predict `compare.json` and `compare.md`, but the API did not clearly return a stable path map for JSON, Markdown, region PNGs, and diff PNGs.

The change keeps the method behavior small. `artifacts.write(...)` still writes only the requested JSON/Markdown files; it does not re-render PNGs. It now also returns all known artifact paths in a convenient keyed object.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementing the beta ergonomics phases and commit at a natural boundary.

**Inferred user intent:** Make each high-priority beta-user improvement real and validated before moving to the next.

**Commit (code):** `94c8544` — "feat: return stable comparison artifact paths"

### What I did

- Updated `internal/cssvisualdiff/jsapi/compare.go`:
  - `writeComparisonArtifacts(...)` now returns stable keys:
    - `outDir`
    - `json`
    - `markdown`
    - `leftRegion`
    - `rightRegion`
    - `diffOnly`
    - `diffComparison`
    - `written`
  - JSON/Markdown are written only when requested.
  - PNG paths are discovered from comparison artifacts, pixel diff fields, and standard `cvd.compare.region(...)` filenames when present.
- Extended `internal/cssvisualdiff/verbcli/command_test.go` to assert the new returned paths and file existence.
- Updated docs:
  - `internal/cssvisualdiff/doc/topics/javascript-api.md`
  - `internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md`
- Ran:

```bash
gofmt -w internal/cssvisualdiff/jsapi/compare.go internal/cssvisualdiff/verbcli/command_test.go
go test ./... -count=1
make lint
```

- Updated ticket tasks, changelog, and this diary.
- Committed the phase.

### Why

Beta users often wrap `css-visual-diff` in project-local commands that return small JSON rows for CI, PR comments, or coding agents. Those rows need stable artifact paths:

```json
{
  "changedPercent": 7.1281,
  "compareJson": ".../compare.json",
  "compareMarkdown": ".../compare.md",
  "diffComparison": ".../diff_comparison.png"
}
```

Guessing filenames works, but it is brittle and undocumented. Returning the paths directly makes wrappers simpler and safer.

### What worked

- The implementation did not require a new artifact model.
- Existing `SelectionComparisonData.Artifacts` and `PixelDiffResult` fields already contained most of the needed information.
- Falling back to standard `outDir` filenames worked for region screenshots, because `cvd.compare.region(...)` writes predictable names.
- Full validation passed:

```bash
go test ./... -count=1
make lint
```

### What didn't work

No test or lint failure occurred in the final implementation pass for this step. The main risk was semantic: deciding whether `artifacts.write(...)` should also create or copy PNG artifacts. I kept it intentionally narrow: it writes JSON/Markdown and reports known PNG paths.

### What I learned

- A stable result map can solve most beta-user integration issues without adding new artifact-writing options.
- The distinction between “write requested reports” and “return known artifacts” should be explicit in docs to avoid users expecting PNG re-rendering.

### What was tricky to build

The tricky part was avoiding overreach. The source request included `diffComparison` in the requested names list, but the current system already writes diff PNGs during `compare.region(...)`. If `artifacts.write(...)` also tried to write PNGs, it would duplicate responsibility and complicate the method.

The chosen solution separates responsibilities: comparison creates PNGs; artifact writing creates JSON/Markdown and returns known paths.

### What warrants a second pair of eyes

- Review whether returned paths should be absolute or preserve the caller's `outDir` form. The current implementation returns paths based on the passed `outDir`, matching existing behavior.
- Review whether `leftRegion` and `rightRegion` should eventually become explicit `SelectionArtifact` entries rather than discovered by filename convention.

### What should be done in the future

- If future comparison APIs write different image names, update the known-path discovery or add explicit artifact entries from the source of truth.
- Consider adding a public artifact kind constants table in docs if more artifact kinds appear.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/jsapi/compare.go
```

Review:

```text
writeComparisonArtifacts
addKnownComparisonArtifactPaths
addPathIfExists
```

Then review integration assertions in:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Validate with:

```bash
go test ./... -count=1
make lint
```

### Technical details

Example JS result:

```js
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])

return {
  compareJson: written.json,
  compareMarkdown: written.markdown,
  diffComparison: written.diffComparison,
  leftRegion: written.leftRegion,
  rightRegion: written.rightRegion,
}
```

Returned shape:

```js
{
  outDir: "artifacts/cta",
  json: "artifacts/cta/compare.json",
  markdown: "artifacts/cta/compare.md",
  leftRegion: "artifacts/cta/left_region.png",
  rightRegion: "artifacts/cta/right_region.png",
  diffOnly: "artifacts/cta/diff_only.png",
  diffComparison: "artifacts/cta/diff_comparison.png",
  written: ["artifacts/cta/compare.json", "artifacts/cta/compare.md"]
}
```

## Step 4: Add a multi-section catalog example and smoke

The third implementation phase made the recommended beta workflow copyable. A single-region example is useful for learning the API, but real beta users need page-level validation: load two pages once, wait for each section, compare multiple regions, write per-section artifacts, and produce one catalog.

This step adds that example as a repository-scanned verb and backs it with a ticket-local smoke script. It demonstrates that the new wait helper and stable artifact write result work together in the intended project-local CLI pattern.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue to the next phase, adding user-facing examples and validating them with a smoke script.

**Inferred user intent:** Give beta users a complete copy/paste workflow rather than only API reference snippets.

**Commit (code/docs):** `a28c5df` — "docs: add beta multi-section catalog example"

### What I did

- Added:

```text
examples/verbs/compare-page-catalog.js
```

- The example command:
  - creates a browser,
  - loads left and right pages once,
  - compares two sections (`page` and `cta`),
  - calls `locator.waitFor(...)` before each comparison,
  - calls `cvd.compare.region(...)`,
  - writes `compare.json` and `compare.md`,
  - records each comparison into a catalog,
  - writes `manifest.json` and `index.md`,
  - returns compact JSON with per-section artifact paths.
- Updated:

```text
examples/verbs/README.md
```

- Added a ticket smoke script:

```text
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

- Ran:

```bash
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
go test ./... -count=1
make lint
```

- Updated ticket tasks, changelog, and this diary.
- Committed the phase.

### Why

The source Pyxis request explicitly asked for a complete multi-section comparison/catalog example. This is the workflow most beta users will need after their first single-region comparison works.

The example also prevents docs from drifting into pseudocode. A smoke-tested example becomes a usability test.

### What worked

- The example successfully exercised the new `locator.waitFor(...)` API.
- The stable artifact map from `comparison.artifacts.write(...)` made the returned JSON straightforward.
- The smoke verified both human-review artifacts and automation-facing stdout JSON.
- Validation passed:

```bash
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
go test ./... -count=1
make lint
```

### What didn't work

No implementation failure occurred in this step. Pre-commit skipped lint/test for this docs/example-only commit because no staged Go files matched the hook's filters, but I had already run `go test ./... -count=1` and `make lint` manually before committing.

### What I learned

- The project-local CLI pattern is clearer when expressed as a full example than as prose in the README alone.
- Smoke scripts for examples should validate both files and returned JSON; otherwise it is easy for examples to write artifacts but return unhelpful output.

### What was tricky to build

The main design choice was keeping the example generic enough for `examples/verbs` while still resembling the Pyxis page/section workflow. I used `#page` and `#cta` defaults so the smoke could be simple, but added flags for right-side selectors to show how real projects can compare prototype selectors against implementation selectors.

Another subtlety was artifact placement: the example uses `catalog.artifactDir(section.name)` so the catalog and artifacts share one predictable directory structure.

### What warrants a second pair of eyes

- Review whether the example should include only one section by default or keep both `page` and `cta`. Two sections better demonstrates catalogs, but single-section examples are simpler.
- Review whether `visible: true` in the waits is too strict for some users. It is appropriate for visual comparison, but docs should clarify that `visible: false` is available when presence is enough.

### What should be done in the future

- Consider adding this smoke to a future all-smokes runner.
- If examples grow further, split `examples/verbs/README.md` into sections or add a small examples index.

### Code review instructions

Start with:

```text
examples/verbs/compare-page-catalog.js
```

Then review:

```text
examples/verbs/README.md
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

Validate with:

```bash
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

### Technical details

Command:

```bash
css-visual-diff verbs --repository examples/verbs examples compare page-catalog \
  http://127.0.0.1:8767/left.html \
  http://127.0.0.1:8767/right.html \
  /tmp/cssvd-page-catalog \
  --output json
```

Expected output directory:

```text
/tmp/cssvd-page-catalog/
  manifest.json
  index.md
  artifacts/
    page/
      left_region.png
      right_region.png
      diff_only.png
      diff_comparison.png
      compare.json
      compare.md
    cta/
      left_region.png
      right_region.png
      diff_only.png
      diff_comparison.png
      compare.json
      compare.md
```

## Step 5: Backfill diary and begin collection profile documentation

After the first three implementation phases, the diary existed but did not follow the stricter implementation diary format. This step backfills the narrative with prompt context, exact failures, review instructions, and future follow-ups so a reviewer or future agent can understand how the ticket evolved.

At the same time, Phase 4 collection-profile documentation has begun in `javascript-api.md`, but it is not yet complete or committed. The documentation now describes `minimal`, `rich`, and `debug` profiles and records the current `rich` default style and attribute sets from `service/collection.go`.

### Prompt Context

**User prompt (verbatim):** "read the diary skill and properly write up the diary. Backfill for previous stuff too."

**Assistant interpretation:** Load and follow the diary skill, then rewrite the ticket diary so it contains structured, reviewable implementation history for the work already done.

**Inferred user intent:** Ensure the ticket is maintainable and reviewable, not just implemented. The user wants the diary to capture the process, mistakes, validation, and continuation points.

**Commit (code):** Not yet committed for this diary rewrite / Phase 4 doc start.

### What I did

- Read the diary skill:

```text
/home/manuel/.pi/agent/skills/diary/SKILL.md
```

- Rewrote this diary using the required structure:
  - `Goal`,
  - numbered steps,
  - prompt context,
  - what changed,
  - why,
  - what worked,
  - what did not work,
  - what was tricky,
  - review instructions,
  - technical details.
- Backfilled steps for:
  - ticket creation and scoped design,
  - selector wait helpers,
  - stable artifact path result,
  - multi-section catalog example.
- Began Phase 4 docs in:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
```

by adding collection profile guidance based on the actual `normalizeCollectOptions` behavior in:

```text
internal/cssvisualdiff/service/collection.go
```

### Why

The user explicitly asked for the diary to be brought up to the project standard. This matters because the ticket spans multiple commits and contains both product/API decisions and implementation details. Without a good diary, future continuation would require reconstructing decisions from git diffs and chat context.

### What worked

- The commit hashes were available from git history:
  - `a370857` for selector wait helpers,
  - `94c8544` for stable artifact paths,
  - `a28c5df` for the multi-section example.
- The earlier terse diary already contained enough facts to backfill exact failures and validation commands.
- `service/collection.go` clearly shows the actual `rich` defaults, so docs can avoid guessing.

### What didn't work

This step is in progress. Phase 4 documentation has started but has not yet been fully validated or committed.

Current uncommitted file at the time of this diary rewrite:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
```

### What I learned

- The diary skill expects a much more structured narrative than a simple chronological log.
- Backfilling is easiest while commit hashes and failures are still fresh.
- Collection profile docs should quote actual behavior from `normalizeCollectOptions`, not general intentions.

### What was tricky to build

The tricky part was avoiding inventing history while backfilling. I only recorded failures that actually happened in the session: the Go struct/lowerCamel issue in Step 2 and the pre-commit skip note in Step 4. For Step 3, where no concrete failure occurred, I recorded the semantic tradeoff instead.

### What warrants a second pair of eyes

- Review the backfilled diary for accuracy against the commits.
- Review the new collection profile docs before committing Phase 4, especially the wording around `styleProps` and `attributes` narrowing collection and comparison.

### What should be done in the future

- Finish Phase 4 documentation in both `javascript-api` and `pixel-accuracy-scripting-guide`.
- Run embedded help rendering.
- Run `go test ./... -count=1`, `make lint`, the beta smoke script, and `docmgr doctor`.
- Commit Phase 4 docs and diary updates as a docs-focused commit.

### Code review instructions

Review this diary first, then inspect the uncommitted documentation diff:

```bash
git diff -- internal/cssvisualdiff/doc/topics/javascript-api.md \
  ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/reference/01-investigation-diary.md
```

### Technical details

Current `rich` defaults from `normalizeCollectOptions`:

```js
styles = [
  "display", "position", "color", "background-color",
  "font-family", "font-size", "font-weight", "line-height",
  "margin", "padding", "border"
]

attributes = ["id", "class", "role", "aria-label", "data-testid"]
```

## Step 6: Finish collection profile documentation

Phase 4 completed the beta ergonomics documentation pass for collection profiles. The implementation already had useful profiles, but the docs did not explain enough of the actual behavior for beta users to choose between them confidently.

The main clarification is that `styleProps` and `attributes` are collection filters. They change what facts are captured from Chromium, not merely what a later report displays. That distinction matters for performance, JSON size, and post-hoc debugging.

### Prompt Context

**User prompt (verbatim):** "Continue with phase 4 in docmgr ticket CSSVD-BETA-ERGONOMICS, keeping a diary (see skill), committing ata ppropriate intervals."

**Assistant interpretation:** Complete the Phase 4 collection profile documentation work in the beta ergonomics ticket, keep the diary current, validate, and commit at the phase boundary.

**Inferred user intent:** Finish the low-complexity beta-user docs polish so users know when to use `minimal`, `rich`, and `debug`, and so future reviewers can see exactly what was verified.

**Commit (docs):** This entry was committed in the docs-only commit `docs: clarify collection profiles for beta users` immediately after validation.

### What I did

- Verified the actual implementation in:

```text
internal/cssvisualdiff/service/collection.go
```

especially `normalizeCollectOptions(...)` and `shouldCollectText(...)`.

- Updated:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md
```

- Documented the collection profiles:
  - `minimal` collects status/bounds by default.
  - `rich` is the default diagnostic profile and collects normalized text, focused computed styles, focused attributes, status, and bounds.
  - `debug` enables normalized text, inner HTML by default, all computed styles, and all attributes.
- Clarified that `outerHtml: true` is needed when a debug/custom collection should include the wrapper element instead of inner HTML.
- Clarified that `styleProps` and `attributes` narrow browser collection as well as comparison/reporting.
- Fixed the tutorial example to use `styleProps`, matching the actual JS API option name, instead of `styles`.
- Rendered embedded help:

```bash
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-help-jsapi-beta.txt
go run ./cmd/css-visual-diff help pixel-accuracy-scripting-guide >/tmp/cssvd-help-pixel-beta.txt
wc -l /tmp/cssvd-help-jsapi-beta.txt /tmp/cssvd-help-pixel-beta.txt
```

Output:

```text
  1002 /tmp/cssvd-help-jsapi-beta.txt
   656 /tmp/cssvd-help-pixel-beta.txt
  1658 total
```

- Reran the beta example smoke:

```bash
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

It passed.

- Ran full validation:

```bash
go test ./... -count=1
make lint
```

Both passed. `make lint` reported `0 issues`.

### Why

The collection profiles are one of the main usability levers in the JS API. Beta users need to understand the tradeoff:

- `minimal` is cheap and CI-friendly but has limited diagnostic value.
- `rich` is the default because it captures enough facts for most failures.
- `debug` is intentionally heavy and best for one-off investigations.

Without this explanation, users may either over-collect in routine CI or under-collect and lose the facts needed to debug a visual mismatch.

### What worked

- `normalizeCollectOptions(...)` made the real behavior easy to verify.
- The docs could describe actual defaults rather than aspirational API behavior.
- Help rendering succeeded for both embedded docs.
- The multi-section smoke still passed after docs changes, confirming the documented workflow remains executable.
- Full Go tests and lint passed.

### What didn't work

No command failed in this phase. One documentation issue was found during review: the tutorial's primitive collection example used `styles` as an option name, but the actual collection option is `styleProps`. I corrected the example.

### What I learned

- `debug` sets `IncludeHTML = true` but does not set `OuterHTML = true`, so the default HTML captured is inner HTML. This needed to be stated precisely.
- `rich` fills default `styleProps` and `attributes` only when those lists are empty and `allStyles`/`allAttributes` are not enabled.
- Documentation for filters should be explicit that collection-time filters affect what data exists later.

### What was tricky to build

The tricky part was describing defaults without oversimplifying them. For example, saying “debug includes HTML” is true but incomplete: it includes inner HTML unless `outerHtml: true` is also set. Saying “rich includes styles” is also incomplete: it includes a specific focused default set unless the user passes `styleProps`.

I resolved this by reading the implementation and documenting the exact default style and attribute arrays.

### What warrants a second pair of eyes

- Review whether the docs overpromise `minimal` by saying custom lower-level option fields can request extra facts. The Go option model supports it, but examples should still steer ordinary users toward `rich` or `debug`.
- Review whether the rich default style list is the right long-term list for beta users. The docs now reflect the current implementation, not necessarily a permanent policy.

### What should be done in the future

- If the default `rich` style/attribute sets change, update both docs in the same commit.
- If beta users repeatedly ask for named style presets, revisit the deferred style-preset task with real usage examples.

### Code review instructions

Start with the implementation being documented:

```text
internal/cssvisualdiff/service/collection.go
```

Then review the docs:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md
```

Validate with:

```bash
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-help-jsapi-beta.txt
go run ./cmd/css-visual-diff help pixel-accuracy-scripting-guide >/tmp/cssvd-help-pixel-beta.txt
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
go test ./... -count=1
make lint
docmgr doctor --root ./ttmp --ticket CSSVD-BETA-ERGONOMICS --stale-after 30
```

### Technical details

Actual `rich` default styles from `normalizeCollectOptions(...)`:

```js
[
  "display", "position", "color", "background-color",
  "font-family", "font-size", "font-weight", "line-height",
  "margin", "padding", "border"
]
```

Actual `rich` default attributes:

```js
["id", "class", "role", "aria-label", "data-testid"]
```

Profile behavior summary:

```text
minimal: status + bounds by default
rich:    normalized text + focused styles + focused attributes + status/bounds
debug:   normalized text + inner HTML + all styles + all attributes + status/bounds
```
