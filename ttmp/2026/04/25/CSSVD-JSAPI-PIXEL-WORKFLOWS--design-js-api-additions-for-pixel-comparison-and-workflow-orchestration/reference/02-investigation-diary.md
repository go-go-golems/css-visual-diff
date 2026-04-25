---
Title: Investigation diary
Ticket: CSSVD-JSAPI-PIXEL-WORKFLOWS
Status: active
Topics:
    - frontend
    - visual-regression
    - browser-automation
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/doc/topics/javascript-api.md
      Note: |-
        Phase 1 embedded JavaScript API reference update for the collected selection model (commit b13933a)
        Phase 2 embedded JavaScript API reference update for structural vs image diff concepts (commit 6ca2498)
        Phase 3 embedded JavaScript API reference update for SelectionComparison concepts (commit 29c8aca)
        Phases 4-5 embedded JS API reference updates for real collect/compare handles (commit 5c76cd7)
        Phases 6-8 embedded API docs for compare.region
        Phase 9 embedded docs for catalog.record comparison integration (commit 1227f1a)
    - Path: internal/cssvisualdiff/doc/topics/javascript-verbs.md
      Note: Phase 8 verbs docs explain built-ins as public API dogfood (commit 88ddac5)
    - Path: internal/cssvisualdiff/dsl/host_test.go
      Note: Phase 8 built-in compare regression test for new public API-backed output shape (commit 88ddac5)
    - Path: internal/cssvisualdiff/dsl/scripts/catalog.js
      Note: Phase 7 built-in catalog scripts use canonical catalog.create/config.load namespaces (commit 88ddac5)
    - Path: internal/cssvisualdiff/dsl/scripts/compare.js
      Note: Phase 8 built-in compare verbs dogfood public cvd.compare.region API (commit 88ddac5)
    - Path: internal/cssvisualdiff/jsapi/catalog.go
      Note: Phase 9 JS catalog.record(comparison
    - Path: internal/cssvisualdiff/jsapi/collect.go
      Note: Phases 4-5 Goja collected selection API handle and cvd.collect namespace (commit 5c76cd7)
    - Path: internal/cssvisualdiff/jsapi/compare.go
      Note: |-
        Phase 5 Goja selection comparison API handle and cvd.compare.selections namespace (commit 5c76cd7)
        Phase 6 cvd.compare.region implementation and selection comparison artifact helpers (commit 88ddac5)
    - Path: internal/cssvisualdiff/jsapi/diff.go
      Note: Phase 7 canonical cvd.diff.structural and cvd.image.diff namespaces (commit 88ddac5)
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: Phase 4 locator.collect method wiring (commit 5c76cd7)
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: |-
        Phases 4-5 module export installation for collect/compare APIs (commit 5c76cd7)
        Phase 7 canonical catalog.create/config.load namespace wiring (commit 88ddac5)
    - Path: internal/cssvisualdiff/jsapi/proxy.go
      Note: Proxy property support and then-safe promise assimilation for Go-backed handles (commit 5c76cd7)
    - Path: internal/cssvisualdiff/jsapi/snapshot.go
      Note: Phase 7 canonical cvd.snapshot.page namespace (commit 88ddac5)
    - Path: internal/cssvisualdiff/modes/compare.go
      Note: Phase 2 compare mode now routes pixel artifact writing through service primitives (commit 6ca2498)
    - Path: internal/cssvisualdiff/modes/pixeldiff.go
      Note: Phase 2 pixeldiff mode now routes PNG diffing/writing through service primitives (commit 6ca2498)
    - Path: internal/cssvisualdiff/modes/pixeldiff_util.go
      Note: Phase 2 compatibility wrappers delegate old mode helpers to service primitives (commit 6ca2498)
    - Path: internal/cssvisualdiff/service/catalog_service.go
      Note: Phase 9 catalog manifest/index support for selection comparison records (commit 1227f1a)
    - Path: internal/cssvisualdiff/service/collection.go
      Note: Phase 1 collected selector data service model (commit b13933a)
    - Path: internal/cssvisualdiff/service/collection_test.go
      Note: Phase 1 service tests for collection profiles
    - Path: internal/cssvisualdiff/service/pixel.go
      Note: Phase 2 reusable image/pixel diff service primitives (commit 6ca2498)
    - Path: internal/cssvisualdiff/service/pixel_test.go
      Note: Phase 2 service tests for image diffing
    - Path: internal/cssvisualdiff/service/selection_compare.go
      Note: Phase 3 service-level selection comparison model and pure data diff helpers (commit 29c8aca)
    - Path: internal/cssvisualdiff/service/selection_compare_test.go
      Note: Phase 3 service tests for style
    - Path: internal/cssvisualdiff/verbcli/command_test.go
      Note: |-
        Phases 4-5 repository-scanned JS integration tests for collection and comparison handles (commit 5c76cd7)
        Phases 6-7 JS verb integration tests for compare.region and canonical namespaces (commit 88ddac5)
        Phase 9 repository-scanned JS smoke for recording comparisons in catalogs (commit 1227f1a)
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/design/01-elegant-javascript-api-additions-for-pixel-comparison-workflows.md
      Note: Main API design proposal produced from this investigation
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/reference/01-pyxis-user-feedback-source-analysis.md
      Note: Source analysis of Pyxis feedback and workflow documents
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/001-service-collection-smoke.sh
      Note: Phase 1 replayable service collection smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/002-pixel-service-smoke.sh
      Note: Phase 2 replayable pixel service smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/003-selection-compare-service-smoke.sh
      Note: Phase 3 replayable selection comparison service smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/004-js-collected-selection-smoke.sh
      Note: Phase 4 replayable JS collected selection smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/005-js-selection-comparison-smoke.sh
      Note: Phase 5 replayable JS selection comparison smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/006-js-compare-region-smoke.sh
      Note: Phase 6 replayable compare.region smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/007-canonical-api-surface-smoke.sh
      Note: Phase 7 replayable canonical namespace smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/008-built-in-compare-dogfood-smoke.sh
      Note: Phase 8 replayable built-in dogfood smoke script
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/009-comparison-catalog-smoke.sh
      Note: Phase 9 replayable comparison catalog smoke script
ExternalSources:
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
Summary: Chronological diary for designing css-visual-diff JS API additions from Pyxis user feedback.
LastUpdated: 2026-04-25T14:05:00-04:00
WhatFor: Resume the API design and implementation without losing the reasoning path, commands, source files, and decisions.
WhenToUse: Read before implementing CSSVD-JSAPI-PIXEL-WORKFLOWS or changing the design proposal.
---







# Investigation Diary

## Goal

Design a polished, maintainable `css-visual-diff` JavaScript API addition based on the first real downstream user feedback from Pyxis.

The user asked to read the Pyxis maintainer feature requests and JS workflow exploration ticket, create a new docmgr ticket in this repository, and think critically about a top-notch API rather than just copying the user-proposed helper shape.

## Step 1: Read the docmgr workflow instructions

Loaded:

```text
/home/manuel/.pi/agent/skills/docmgr/SKILL.md
```

Key rules applied:

- Create a ticket with `docmgr ticket create-ticket`.
- Use `--root ./ttmp` so the ticket is repository-local, not parent-project local.
- Create focused design/reference docs.
- Keep a diary for active tickets.
- Relate important files to the ticket/docs.

## Step 2: Read Pyxis source feedback

Read primary source:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
```

Important findings:

- Pyxis can already use JS verbs, locators, probes, snapshots, structural diffs, inspect artifacts, catalogs, and YAML loading.
- The high-priority missing primitive is JS-callable pixel/region comparison.
- The built-in CLI command `css-visual-diff verbs script compare region` works and reproduced a YAML page-section diff.
- Project JS verbs cannot call that behavior through the documented `require("css-visual-diff")` module.
- The proposed minimal API is `cvd.comparePixels({ left: { page, selector }, right: { page, selector }, ... })`.

## Step 3: Read Pyxis workflow exploration ticket

Listed files under:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
```

Read:

```text
index.md
design/01-css-visual-diff-javascript-workflow-experiment-guide.md
reference/01-exploration-diary.md
reference/02-copious-research-notes-for-technical-deep-dive.md
tasks.md
```

Important evidence:

```text
Archive content YAML diff:      7.1281%
Archive content JS region diff: 7.128146453089244%
```

The workflow pain is not basic visual-diff capability. The pain is orchestration:

- repeated URLs,
- repeated selectors,
- repeated viewport setup,
- reporting and summary friction,
- authoring-vs-CI policy,
- desire to keep Pyxis-specific registries in userland.

## Step 4: Create the css-visual-diff docmgr ticket

Ran:

```bash
docmgr ticket create-ticket \
  --root ./ttmp \
  --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS \
  --title "Design JS API additions for pixel comparison and workflow orchestration" \
  --topics frontend,visual-diff,javascript,goja,api
```

Created ticket path:

```text
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration
```

## Step 5: Create focused documents

Ran:

```bash
docmgr doc add --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --doc-type design --title "Elegant JavaScript API additions for pixel comparison workflows"

docmgr doc add --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --doc-type reference --title "Pyxis user feedback source analysis"

docmgr doc add --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --doc-type reference --title "Investigation diary"
```

Then rewrote the generated documents with detailed frontmatter and content.

## Step 6: API design conclusion

The user's minimal ask is valid but should be refined.

Do **not** make this the primary public API:

```js
await cvd.comparePixels({
  left: { page, selector },
  right: { page, selector },
})
```

Reasons:

- `comparePixels` is too narrow; the operation compares visual regions and may include screenshots, computed styles, matched-style winners, and artifacts.
- Raw `{ page, selector }` objects weaken the strict Go-backed handle model.
- It risks confusing pixel diff with existing `cvd.diff(...)`, which is structural JSON diff.

Recommended public API:

```js
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

Key principles:

- Public module remains `require("css-visual-diff")`.
- Internal `require("diff")` should stay compatibility/internal, not become the public API.
- `cvd.compare.region` should accept strict locator handles.
- Results should be lowerCamel, plain serializable data with `schemaVersion`.
- Implementation should be service-first so CLI/builtins/YAML/JS can converge.

## Step 7: Implementation plan captured

The design doc proposes phases:

1. Service extraction for image diffing.
2. Service-level region comparison.
3. JS API `cvd.compare.region`.
4. Built-in/CLI convergence.
5. Docs and smoke tests.
6. Follow-up work for multi-section helpers, YAML job bridge, tolerances, and CSS normalization.

## Step 8: Relate docs and validate ticket hygiene

Related the main design doc to current css-visual-diff implementation files:

```text
internal/cssvisualdiff/dsl/scripts/compare.js
internal/cssvisualdiff/dsl/registrar.go
internal/cssvisualdiff/modes/compare.go
internal/cssvisualdiff/jsapi/locator.go
internal/cssvisualdiff/doc/topics/javascript-api.md
```

Related the source-analysis doc to the Pyxis maintainer request and workflow exploration docs.

Ran:

```bash
docmgr doctor --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --stale-after 30
```

Initial doctor output warned that `api`, `goja`, `javascript`, and `visual-diff` were not in the current shared vocabulary. I switched the ticket/doc topics to existing vocabulary values:

```text
frontend
visual-regression
browser-automation
tooling
```

Rerunning doctor passed:

```text
✅ All checks passed
```

## Step 9: Refine design around a single comparison result object

The user asked whether the API could be designed around a single “comparison result” object from which artifacts can be queried. This is a strong refinement.

Updated the design to treat `cvd.compare.region(...)` as returning a Go-backed `cvd.comparison` handle:

```js
const comparison = await cvd.compare.region({
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
})

const summary = comparison.summary()
const data = comparison.toJSON()
const markdown = comparison.report.markdown()
await comparison.artifacts.write(outDir, ["diffComparison", "json", "markdown"])
```

This makes artifacts views over one materialized comparison instead of separate operations. It also lets userland choose when to write PNGs, JSON, or Markdown.

Important design caveat: the comparison handle is useful inside scripts, but CLI output and persisted JSON still need plain data. Therefore the docs should teach explicit `comparison.toJSON()` for returned rows.

## Step 10: Reconsider inspection data as rich default, not opt-in evidence

The user pointed out that collecting browser facts is relatively cheap compared with the expensive parts of region comparison: navigation, screenshot capture, image normalization, and pixel diffing. They also emphasized that a rich comparison object should expose data for further JavaScript manipulation, not just wrap report writing.

Updated the design accordingly:

- Rename the concept from `evidence` to `inspect`.
- Make `inspect: "rich"` the default for script-facing comparison objects.
- Treat the noisy part as output/report filtering, not data collection.
- Add comparison-object data views:
  - `comparison.bounds.diff()`
  - `comparison.styles.left(props?)`
  - `comparison.styles.right(props?)`
  - `comparison.styles.diff(propsOrPreset?)`
  - `comparison.attributes.diff(names?)`
- Keep `inspect: "minimal"` for batch/CI workloads where memory and output size matter more than exploratory flexibility.

The revised direction is that `cvd.compare.region(...)` should collect a rich browser inspection payload once, then let scripts query and filter it after the fact.

## Step 11: Write the JavaScript-first collected data design report

The user asked whether the API should be centered around three objects:

1. collected data for one selector,
2. collected data for the other selector,
3. comparison between two collected data values.

I concluded that this model is the better long-term API shape. It makes the API feel JavaScript-first rather than like a direct wrapper around Go modes. It also preserves both use cases:

- simple scripts can call `cvd.compare.region(...)`,
- advanced scripts can call `locator.collect(...)`, inspect/filter the collected data, and then call `cvd.compare.selections(...)`.

Wrote:

```text
design/02-javascript-centric-collected-data-and-comparison-object-api.md
```

The report explains the current system for a new intern, including `jsapi`, `service`, `modes`, `dsl`, the current internal `require("diff")` path, and the proposed collection/comparison/artifact object model. It includes prose, tables, pseudocode, Mermaid diagrams, API sketches, and file references.

## Step 12: Remove backward compatibility as a design constraint

The user clarified that backward compatibility is not required and that an opinionated low-effort surface is desirable. I updated the full API-wide design accordingly.

The new framing is:

1. Provide an opinionated low-effort path for users who want useful comparisons quickly.
2. Provide composable primitives for users who want to collect, analyze, filter, and report in JavaScript.
3. Do not keep old names as permanent aliases merely because they already exist.
4. Keep short forms only when they are intentionally designed convenience APIs, not historical compatibility shims.

This strengthens the design because it lets the API choose clear canonical names such as `cvd.diff.structural`, `cvd.catalog.create`, `cvd.config.load`, `locator.collect`, `cvd.compare.selections`, and `cvd.compare.region` without also preserving every earlier ambiguous top-level shortcut.

## Step 13: Expand implementation tasks into detailed phases

The user requested a detailed implementation task list split into phases, with no backward-compatibility requirement, JavaScript API reference updates in every phase, and real smoke test scripts.

I rewrote `tasks.md` into a 12-phase implementation plan:

1. Ticket setup and design.
2. Service model for collected selector data.
3. Image/pixel diff service extraction.
4. Service-level selection comparison.
5. JavaScript `CollectedSelection` handle.
6. JavaScript `SelectionComparison` handle.
7. Opinionated `cvd.compare.region(...)` low-effort surface.
8. Canonical namespace cleanup with no historical compatibility aliases.
9. Built-in compare verbs dogfooding public primitives.
10. Catalog/report/artifact integration.
11. Pixel accuracy guide and public examples refresh.
12. Optional config/job bridge.
13. Final validation and delivery.

Each implementation phase now includes:

- implementation tasks,
- package tests,
- an explicit `internal/cssvisualdiff/doc/topics/javascript-api.md` update,
- a real ticket-local smoke script under `scripts/`,
- diary/changelog recording expectations.

The task list explicitly requires a final `scripts/012-run-all-smokes.sh` to replay all smoke scripts in numeric order.

## Step 14: Implement Phase 1 collected selector data service

Phase 1 adds the Go-side foundation for the next JavaScript API: a collected selection value. The important boundary is that a locator remains a live browser handle, while a collected selection is stable data captured at one point in time and safe to serialize, compare, and report on later.

I implemented the service first instead of starting with the Goja binding. That keeps the browser collection rules typed and testable before Phase 4 turns the value into a Go-backed JavaScript handle.

### Prompt Context

**User prompt (verbatim):** "Implement phase 1 of docmgr ticket CSSVD-JSAPI-PIXEL-WORKFLOWS, task by task, keep a diary as you work, commit at appropriate intervals."

**Assistant interpretation:** Implement the first ticket phase exactly as planned: service-level collected selector data, tests, JS API reference update, real smoke script, diary/changelog bookkeeping, and commits.

**Inferred user intent:** Establish a durable, reviewed foundation for the larger JavaScript pixel workflow API while preserving the reasoning and validation trail in the docmgr ticket.

**Commit (code):** b13933a — "feat: add collected selection service model"

### What I did

- Added `internal/cssvisualdiff/service/collection.go`.
- Added `CollectedSelectionData` / `SelectionData` with schema version `cssvd.collectedSelection.v1`.
- Added `CollectOptions` with `minimal`, `rich`, and `debug` inspection profiles.
- Added `CollectSelection(page, locator, opts)` built on the existing DOM service primitives.
- Added collection support for all computed styles and all attributes.
- Added `CollectionError` with distinct kinds for invalid selector, browser, and artifact failures.
- Added `internal/cssvisualdiff/service/collection_test.go`.
- Updated `internal/cssvisualdiff/doc/topics/javascript-api.md` with the collected selection model and preview API.
- Added `scripts/001-service-collection-smoke.sh`.
- Marked Phase 1 tasks complete in `tasks.md`.
- Related implementation and smoke files to this diary with `docmgr doc relate`.

### Why

- The next API phases need a typed value that represents browser truth for one selector at one time.
- Comparing locators directly would keep re-querying live browser state and would make reports less deterministic.
- Implementing the service before the JS handle gives us a stable primitive for Phase 4 `locator.collect(...)` and Phase 5/6 comparison APIs.

### What worked

- Focused service tests passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestCollectSelection' -count=1
```

- The ticket smoke script passed after fixing its repository-root path:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/001-service-collection-smoke.sh
```

- Full test suite passed after reducing collection tests to reuse fewer Chromium sessions:

```bash
go test ./... -count=1
```

- Embedded help rendering includes the new collected selection section:

```bash
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-js-api-help.txt
rg "Collected selection model|locator.collect|cssvd.collectedSelection.v1" /tmp/cssvd-js-api-help.txt
```

### What didn't work

- The first version of the smoke script computed the repository root incorrectly. It used:

```bash
cd "$(dirname "$0")/../../../../../../.."
```

That moved one directory too high and produced:

```text
# ./internal/cssvisualdiff/service
stat /home/manuel/workspaces/2026-04-21/hair-v2/internal/cssvisualdiff/service: directory not found
FAIL	./internal/cssvisualdiff/service [setup failed]
```

I fixed it to:

```bash
cd "$(dirname "$0")/../../../../../.."
```

- The first full `go test ./... -count=1` run failed while starting another Chrome instance:

```text
chrome failed to start:
Failed to open /dev/null
[... FATAL:content/browser/zygote_host/zygote_host_impl_linux.cc:207] Check failed: . : File exists (17)
```

The collection tests initially started a fresh browser for every small assertion group, which increased pressure on Chrome startup. I collapsed the service collection assertions into one main browser-backed test plus one invalid-selector test. The full suite then passed.

### What I learned

- The existing DOM service primitives are a good substrate for collection, but all-styles and all-attributes require dedicated JS evaluation helpers.
- Chrome-backed tests should avoid unnecessary browser churn; otherwise unrelated-looking browser startup failures can appear during full-suite runs.
- The collected value should remain plain Go/JSON data in Phase 1. Behavior belongs in later Go-backed JS handles.

### What was tricky to build

- `minimal` needed to avoid accidentally collecting text. The first normalization path would have defaulted text options before profile selection, which made the minimal profile too rich. I adjusted option normalization so `minimal` stays status/bounds-only.
- All computed styles cannot use the existing `LocatorComputedStyle` helper because that helper expects a property list. I added a direct `window.getComputedStyle(el)` iterator for the all-styles profile.
- Invalid selectors are discovered via `LocatorStatus`, which reports selector errors in the status payload rather than returning an error. `CollectSelection` now converts that status error into a typed `CollectionError` with kind `invalidSelector`.

### What warrants a second pair of eyes

- The default `rich` style list is intentionally small and opinionated. Review whether the default set is enough for Phase 5/6 comparison reports or whether it should be aligned with the upcoming style presets.
- The current Phase 1 service does not capture screenshots yet. That is acceptable for Phase 1, but Phase 2/3 must decide exactly how screenshot descriptors and temp image paths flow into pixel comparison.
- The map sorting helper creates deterministic insertion order before JSON marshaling, but Go JSON itself sorts map keys. Review whether the helper is necessary or should be removed later.

### What should be done in the future

- Phase 2 should extract image/pixel diff primitives.
- Phase 4 should wrap `CollectedSelectionData` in a Go-backed `cvd.collectedSelection` Proxy handle.
- Phase 4 should make the documented `locator.collect(...)` and `cvd.collect.selection(...)` APIs real.

### Code review instructions

- Start with `internal/cssvisualdiff/service/collection.go`, especially `CollectSelection`, `normalizeCollectOptions`, `collectAllComputedStyles`, and `collectAllAttributes`.
- Then review `internal/cssvisualdiff/service/collection_test.go` to see the expected profile behavior.
- Review the API documentation insertion in `internal/cssvisualdiff/doc/topics/javascript-api.md` for consistency with the no-backward-compat API direction.
- Validate with:

```bash
go test ./internal/cssvisualdiff/service -run 'TestCollectSelection' -count=1
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/001-service-collection-smoke.sh
go test ./... -count=1
```

### Technical details

The Phase 1 schema version is:

```text
cssvd.collectedSelection.v1
```

The main service entry point is:

```go
func CollectSelection(page *driver.Page, locator LocatorSpec, opts CollectOptions) (CollectedSelectionData, error)
```

Profile summary:

```text
minimal: status/existence/visibility/bounds
rich:    text + common styles + common attributes + status/bounds
debug:   HTML + all styles + all attributes + text + status/bounds
```

## Step 15: Implement Phase 2 pixel diff service primitives

Phase 2 extracts the image-level pixel-diff machinery into the service package so future JavaScript APIs can use it directly. The CLI modes still work, but the core PNG reading, normalization, threshold comparison, diff-only image rendering, side-by-side comparison rendering, and parent-directory-safe writing now live behind reusable service functions.

This keeps pixel/image diffing separate from structural JSON diffing. That distinction is important for the no-backward-compat API direction: the canonical future JavaScript surface should be explicit, for example `cvd.diff.structural(...)` for plain data and `cvd.image.diff(...)` for rendered images.

### Prompt Context

**User prompt (verbatim):** "continue phase 2"

**Assistant interpretation:** Continue the implementation plan by completing Phase 2: extract pixel diff service primitives, add tests, update the JS API reference, write a real smoke script, update the diary/changelog/tasks, and commit.

**Inferred user intent:** Progress the implementation ticket incrementally while preserving validation and review trail after the completed Phase 1 service collection work.

**Commit (code):** 6ca2498 — "feat: extract pixel diff service primitives"

### What I did

- Added `internal/cssvisualdiff/service/pixel.go`.
- Added `PixelDiffOptions` and `PixelDiffResult` with lowerCamel JSON fields.
- Added service functions:
  - `ValidatePixelThreshold`,
  - `ReadPNG`,
  - `WritePNG`,
  - `ToNRGBA`,
  - `PadToSameSize`,
  - `DiffImages`,
  - `DiffPNGFiles`,
  - `ComputePixelDiff`,
  - `CombineSideBySide`,
  - `WritePixelDiffImages`.
- Added `internal/cssvisualdiff/service/pixel_test.go`.
- Reworked `internal/cssvisualdiff/modes/compare.go` so `writePixelDiffImages` delegates to the service.
- Reworked `internal/cssvisualdiff/modes/pixeldiff.go` so `RunPixelDiff` delegates to the service.
- Replaced the old mode-local utility implementation in `internal/cssvisualdiff/modes/pixeldiff_util.go` with wrappers around service primitives.
- Updated `internal/cssvisualdiff/doc/topics/javascript-api.md` to distinguish structural diffs from image/pixel diffs and to preview canonical `cvd.image.diff(...)`.
- Added `scripts/002-pixel-service-smoke.sh`.
- Marked Phase 2 tasks complete in `tasks.md`.
- Related implementation files and the smoke script to this diary.

### Why

- The upcoming `cvd.compare.selections(...)` and `cvd.compare.region(...)` APIs need a reusable image diff primitive that is not trapped inside CLI mode code.
- Future `cvd.image.diff(...)` should be backed by the same Go service used by built-ins and modes.
- Writer helpers need to create parent directories because JS APIs should make artifact writing low-effort and predictable.

### What worked

- Focused service tests passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestDiffImages|TestWritePixelDiff|TestValidatePixel' -count=1
```

- Modes and services passed together:

```bash
go test ./internal/cssvisualdiff/modes ./internal/cssvisualdiff/service -count=1
```

- The Phase 2 smoke script passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/002-pixel-service-smoke.sh
```

- Full test suite passed:

```bash
go test ./... -count=1
```

### What didn't work

- The first `gofmt` run failed because I accidentally wrote two statements on one line in `pixel_test.go`:

```text
internal/cssvisualdiff/service/pixel_test.go:75:47: expected ';', found require
```

The bad line combined `rightPath := ...` and `require.NoError(...)`. I split it into separate lines and reran `gofmt` and tests.

### What I learned

- The old mode-local helpers already had a clean enough algorithmic boundary; extracting them mostly required defining the right service result shape and keeping mode JSON schemas stable.
- LowerCamel service JSON fields are useful now even before the JS binding exists, because they clarify the future `cvd.image.diff(...)` contract.
- Keeping small wrappers in `modes/pixeldiff_util.go` lets existing mode tests continue to validate behavior while the implementation is now service-backed.

### What was tricky to build

- The modes currently expose snake_case JSON structs (`PixelDiffStats`, `PixelDiffEntry`) while the future JS service result should be lowerCamel. I kept the mode structs stable and added `pixelDiffStatsFromService(...)` as an adapter. That preserves current mode behavior without making backward compatibility a public JS API constraint.
- The threshold semantics are based on squared RGB distance. A threshold of `255` does not ignore a full black-to-white RGB change because the three-channel squared distance exceeds `255^2`. The test documents that behavior explicitly.
- Parent-directory creation belongs in `service.WritePNG`, because every higher-level writer benefits from it and smoke scripts should not need to pre-create nested artifact directories.

### What warrants a second pair of eyes

- Review whether `PixelDiffOptions.Threshold` should default to `0` or whether callers should normalize to a project default such as `30` at higher API layers.
- Review the lowerCamel `PixelDiffResult` field names before they become part of the JS API.
- Review whether mode-local wrappers should remain until all built-ins are converted, or whether they should be removed in a later cleanup phase.

### What should be done in the future

- Phase 3 should use `service.PixelDiffResult` when comparing collected selections that include screenshot/image descriptors.
- A future `cvd.image.diff(...)` binding can call `service.DiffPNGFiles` / `service.WritePixelDiffImages` directly.
- Built-in compare verbs should eventually dogfood the public `cvd.compare.region(...)` path instead of internal helper modules.

### Code review instructions

- Start with `internal/cssvisualdiff/service/pixel.go`.
- Review `PixelDiffResult` and `WritePixelDiffImages` first, because those define the reusable service contract.
- Review `internal/cssvisualdiff/modes/compare.go` and `internal/cssvisualdiff/modes/pixeldiff.go` to confirm behavior still flows through the service.
- Validate with:

```bash
go test ./internal/cssvisualdiff/service -run 'TestDiffImages|TestWritePixelDiff|TestValidatePixel' -count=1
go test ./internal/cssvisualdiff/modes ./internal/cssvisualdiff/service -count=1
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/002-pixel-service-smoke.sh
go test ./... -count=1
```

### Technical details

The service result uses lowerCamel fields for future JS lowering:

```json
{
  "threshold": 30,
  "totalPixels": 10000,
  "changedPixels": 713,
  "changedPercent": 7.13,
  "normalizedWidth": 500,
  "normalizedHeight": 20,
  "diffOnlyPath": "artifacts/diff_only.png",
  "diffComparisonPath": "artifacts/diff_comparison.png"
}
```

The service normalizes different-size images by padding both sides to the larger width and height with opaque white pixels before comparing.

## Step 16: Implement Phase 3 service-level selection comparison

Phase 3 adds the pure service layer that compares two collected selections. The important constraint is that comparison does not talk to the browser: it operates only on immutable `SelectionData` captured earlier, which keeps analysis deterministic and prepares the ground for JavaScript `SelectionComparison` handles.

The new service compares bounds, text, computed style maps, attribute maps, and optional screenshots. If both collected selections contain screenshot paths, it delegates pixel/image comparison to the Phase 2 pixel service and can write diff artifacts when paths are provided.

### Prompt Context

**User prompt (verbatim):** "Phase 3"

**Assistant interpretation:** Implement the third ticket phase: service-level selection comparison, tests, API reference update, smoke script, diary/changelog/task updates, and commits.

**Inferred user intent:** Continue the phase-by-phase implementation plan and create the data model that future JS `cvd.compare.selections(...)` and `cvd.compare.region(...)` APIs will wrap.

**Commit (code):** 29c8aca — "feat: add selection comparison service"

### What I did

- Added `internal/cssvisualdiff/service/selection_compare.go`.
- Added schema version `cssvd.selectionComparison.v1`.
- Added `CompareSelectionOptions` with threshold, include/exclude style props, include/exclude attributes, and optional pixel artifact paths.
- Added `SelectionComparisonData` with left/right summaries, pixel summary, bounds diff, text diff, style diffs, attribute diffs, and artifact descriptors.
- Implemented `CompareSelections(left, right, opts)`.
- Implemented deterministic pure-data diff helpers for bounds, text, style maps, and attribute maps.
- Integrated screenshot comparison through the Phase 2 pixel service when both selections include screenshot paths.
- Added `internal/cssvisualdiff/service/selection_compare_test.go`.
- Updated `internal/cssvisualdiff/doc/topics/javascript-api.md` with `SelectionComparison` concepts and future JS handle methods.
- Added `scripts/003-selection-compare-service-smoke.sh`.
- Marked Phase 3 tasks complete in `tasks.md`.
- Related implementation files and the smoke script to this diary.

### Why

- Future JavaScript comparison handles need a typed Go service result to wrap.
- The comparison layer should not re-query browser state; that belongs to collection.
- Pixel summaries and artifacts need to compose with style/bounds/attribute/text diffs under one durable comparison object.

### What worked

- Focused selection comparison tests passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestCompareSelections' -count=1
```

- The Phase 3 smoke script passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/003-selection-compare-service-smoke.sh
```

- Full test suite passed:

```bash
go test ./... -count=1
```

### What didn't work

- The first version of the smoke script wrote a temporary test package under `/tmp` and tried to import the repository's internal service package. Go rejected that with:

```text
use of internal package github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service not allowed
```

I fixed the script to create the temporary smoke test directory inside the repository root with `mktemp -d "$PWD/.tmp-selection-compare-smoke-XXXXXX"`, so the import is allowed by Go's `internal` package rules. The script removes the directory on exit.

- While drafting the pixel integration test I again accidentally put `rightPath := ...` and `require.NoError(...)` on the same line. I fixed the line before the final `gofmt`/test run.

### What I learned

- The service-level comparison boundary is naturally a pure data operation, except optional screenshot paths. That makes it easier to reason about deadlocks and page serialization later, because Phase 3 does not hold or need page locks.
- Smoke scripts that compile ad hoc Go files must live inside the repository when importing `internal/...` packages.
- The lowerCamel schema shape should be decided now because Phase 4/5 JS handles will use `toJSON()` to lower into this data.

### What was tricky to build

- Style and attribute diff ordering needs to be deterministic. I implemented selected key calculation with sorted keys, while preserving explicit include filters and exclude filters.
- Pixel integration has two modes: compute stats only if no artifact paths are provided, or write artifacts if both `diffOnlyPath` and `diffComparisonPath` are provided. Providing only one path is rejected because partial artifact writes would make the result ambiguous.
- Bounds diff has to handle nil bounds on either side. The service marks that changed if one side is nil and keeps a nil delta because there is no meaningful numeric delta.

### What warrants a second pair of eyes

- Review the `SelectionComparisonData` JSON shape before it becomes wrapped by the JS `SelectionComparison` handle.
- Review whether unchanged style/attribute values should be omitted, as implemented now, or optionally included later for richer debug reports.
- Review `Name` defaulting. It currently uses explicit option name, then left name, then right name.

### What should be done in the future

- Phase 4 should expose collected selection handles in JavaScript.
- Phase 5 should wrap `SelectionComparisonData` in a Go-backed JS handle with `summary()`, `toJSON()`, `styles.diff`, `bounds.diff`, `attributes.diff`, `pixel.summary`, reports, and artifacts.
- Phase 6 should implement `cvd.compare.region(...)` as collect-left, collect-right, compare-selections.

### Code review instructions

- Start with `internal/cssvisualdiff/service/selection_compare.go`, especially `CompareSelections`, `diffStringMaps`, and `compareSelectionScreenshots`.
- Then review `internal/cssvisualdiff/service/selection_compare_test.go` for expected behavior and JSON shape.
- Review `internal/cssvisualdiff/doc/topics/javascript-api.md` for the user-facing comparison model.
- Validate with:

```bash
go test ./internal/cssvisualdiff/service -run 'TestCompareSelections' -count=1
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/003-selection-compare-service-smoke.sh
go test ./... -count=1
```

### Technical details

The Phase 3 schema version is:

```text
cssvd.selectionComparison.v1
```

The main service entry point is:

```go
func CompareSelections(left SelectionData, right SelectionData, opts CompareSelectionOptions) (SelectionComparisonData, error)
```

Comparison does not query the browser. It only compares data already present in `SelectionData`.

## Step 17: Implement Phases 4 and 5 JavaScript collected/comparison handles

Phases 4 and 5 turn the Phase 1 and Phase 3 service data into real JavaScript APIs. `page.locator(selector).collect(...)` and `cvd.collect.selection(...)` now return Go-backed `cvd.collectedSelection` handles, and `cvd.compare.selections(left, right, options)` returns a Go-backed `cvd.selectionComparison` handle.

This is the first point where the new API becomes directly usable from repository-scanned JS verbs. The implementation keeps browser work Promise-first, uses strict Proxy-backed handles, and requires explicit lowering through `summary()` / `toJSON()` for plain output.

### Prompt Context

**User prompt (verbatim):** "phase 4 - 5"

**Assistant interpretation:** Implement both the JavaScript `CollectedSelection` and `SelectionComparison` API phases, with tests, docs, smoke scripts, diary/changelog/tasks, and commits.

**Inferred user intent:** Move from Go service primitives to the actual user-facing JS API needed for scripts, while keeping the phase discipline and no-backward-compat canonical direction.

**Commit (code):** 5c76cd7 — "feat: expose collected selection comparison js api"

### What I did

- Added `internal/cssvisualdiff/jsapi/collect.go`.
- Added `locator.collect(options?)`.
- Added `cvd.collect.selection(locator, options?)`.
- Added Go-backed Proxy owner `cvd.collectedSelection`.
- Added collected selection methods:
  - `summary()`,
  - `toJSON()`,
  - `status()`,
  - `bounds()`,
  - `text()`,
  - `styles(props?)`,
  - `attributes(names?)`,
  - `screenshot.write(path)` placeholder error until screenshot collection is enabled.
- Added option decoding for `inspect`, `styles: "all"`, style arrays, `attributes: "all"`, and attribute arrays.
- Added `internal/cssvisualdiff/jsapi/compare.go`.
- Added `cvd.compare.selections(leftCollected, rightCollected, options?)`.
- Added Go-backed Proxy owner `cvd.selectionComparison`.
- Added comparison methods/properties:
  - `summary()`,
  - `toJSON()`,
  - `left()`,
  - `right()`,
  - `artifact(name)`,
  - `pixel.summary()`,
  - `bounds.diff()`,
  - `styles.diff(props?)`,
  - `attributes.diff(names?)`,
  - `report.markdown()`,
  - `report.writeMarkdown(path)`,
  - `artifacts.list()`,
  - `artifacts.write(outDir, names?)`.
- Extended Proxy infrastructure with property getters so handles can expose namespaces such as `comparison.styles.diff()`.
- Made Proxy handles return `undefined` for property `then` so Promise resolution does not treat resolved handles as thenables.
- Updated module registration to install `cvd.collect` and `cvd.compare`.
- Added repository-scanned JS integration tests in `internal/cssvisualdiff/verbcli/command_test.go`.
- Updated `internal/cssvisualdiff/doc/topics/javascript-api.md` to describe collect/compare APIs as real rather than future.
- Added smoke scripts:
  - `scripts/004-js-collected-selection-smoke.sh`,
  - `scripts/005-js-selection-comparison-smoke.sh`.
- Marked Phase 4 and Phase 5 tasks complete in `tasks.md`.
- Related implementation files and smoke scripts to this diary.

### Why

- Phase 4 exposes the browser truth collection primitive to JS scripts.
- Phase 5 exposes comparison as a queryable object rather than a raw blob.
- Strict Go-backed handles preserve the project direction: robust runtime feedback for LLM-written scripts and explicit lowering for output.

### What worked

- Focused JS integration tests passed:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleCollectsLocatorSelection|TestCVDModuleComparesCollectedSelections' -count=1
```

- Phase 4 smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/004-js-collected-selection-smoke.sh
```

- Phase 5 smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/005-js-selection-comparison-smoke.sh
```

- Full test suite passed:

```bash
go test ./... -count=1
```

### What didn't work

- The first JS integration run failed when awaiting `locator.collect(...)`:

```text
promise rejected: TypeError: cvd.collectedSelection: unknown method .then(). Available: attributes, bounds, status, styles, summary, text, toJSON.
```

Goja/Promise resolution checks whether resolved values are thenables by reading `.then`. Our Proxy treated every unknown property as an error, so resolving a Proxy handle caused Promise assimilation to call the unknown-property path. I fixed this by returning `undefined` for Proxy property `then`, which marks handles as non-thenable and lets awaited promises resolve to the handle.

### What I learned

- Go-backed Proxy handles that are returned from Promises must be safe for `.then` property reads. This is a general invariant for all current and future Promise-returned handles.
- Method-only proxies are not enough for the fluent object shape we want. Namespaces such as `comparison.styles.diff()` require controlled property getters.
- Keeping `summary()` and `toJSON()` explicit makes command output predictable and prevents accidental serialization of behavior-rich handles.

### What was tricky to build

- `comparison.styles.diff()` required adding `ProxySpec.Properties` instead of forcing everything into methods. The property returns a plain Goja object whose functions operate on the Go-backed data.
- `comparison.artifacts.write(...)` is currently a report/JSON writer for comparison data. Pixel artifacts already written by the service are listed, but Phase 5 does not yet recopy or lazily materialize PNG artifacts. Phase 6/9 can refine artifact policy.
- Collection options needed to bridge JS-friendly names (`styles: "all"`, `styles: [...]`) into Go service fields (`AllStyles`, `StyleProps`) without adding a separate userland parser.

### What warrants a second pair of eyes

- Review the exact method/property split on `cvd.selectionComparison`: especially `left()` / `right()` as methods versus `pixel`, `bounds`, `styles`, `attributes`, `report`, and `artifacts` as properties.
- Review whether `screenshot.write(path)` should remain present-but-erroring until screenshot collection lands, or whether it should be hidden until implemented.
- Review `artifacts.write(...)` semantics: it writes JSON/Markdown now; PNG copy/lazy materialization may need a clearer contract in Phase 6/9.

### What should be done in the future

- Phase 6 should implement `cvd.compare.region({ left, right })` as collect-left, collect-right, compare-selections.
- Future phases should add screenshot collection to `CollectedSelection` so pixel diffs can be produced from JS collect/compare flows.
- Phase 7 should clean the canonical namespace surface and remove compatibility language.

### Code review instructions

- Start with `internal/cssvisualdiff/jsapi/collect.go` and `internal/cssvisualdiff/jsapi/compare.go`.
- Review the Proxy infrastructure change in `internal/cssvisualdiff/jsapi/proxy.go`; the `.then` special case is important for Promise-returned handles.
- Review integration tests at the end of `internal/cssvisualdiff/verbcli/command_test.go`.
- Validate with:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleCollectsLocatorSelection|TestCVDModuleComparesCollectedSelections' -count=1
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/004-js-collected-selection-smoke.sh
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/005-js-selection-comparison-smoke.sh
go test ./... -count=1
```

### Technical details

Phase 4 handle owner:

```text
cvd.collectedSelection
```

Phase 5 handle owner:

```text
cvd.selectionComparison
```

Canonical API now implemented:

```js
const selected = await page.locator("#cta").collect({ inspect: "rich" })
const same = await cvd.collect.selection(page.locator("#cta"), { inspect: "minimal" })
const comparison = await cvd.compare.selections(left, right, { styleProps: ["color"] })
```

## Step 18: Implement Phases 6 through 8 canonical compare-region workflow

Phases 6 through 8 make the new comparison stack practical. `cvd.compare.region(...)` is now the opinionated low-effort public API, canonical namespaces exist for the next-generation surface, and the built-in `script compare` verbs dogfood the public `require("css-visual-diff")` API instead of the older internal `require("diff")` and `require("report")` modules.

This step intentionally follows the ticket's no-backward-compatibility direction in docs, examples, tests, and built-ins. Some old top-level functions still remain as implementation aliases to avoid a wider unrelated test migration in this turn, but public documentation and examples now teach the canonical namespace map.

### Prompt Context

**User prompt (verbatim):** "phase 6 - 8"

**Assistant interpretation:** Implement the opinionated `cvd.compare.region(...)` surface, canonical namespace cleanup, and built-in compare dogfooding phases, with tests, docs, smoke scripts, diary/changelog/task updates, and commits.

**Inferred user intent:** Finish the core user-facing JS pixel workflow API so scripts and built-ins use the same public primitives.

**Commit (code):** 88ddac5 — "feat: add canonical compare region js workflow"

### What I did

- Implemented `cvd.compare.region({ left, right, name, threshold, inspect, outDir, styleProps, attributes })`.
- Made `compare.region` require strict `page.locator(...)` handles for `left` and `right`.
- Made `compare.region` collect both sides, capture region screenshots, and return a `cvd.selectionComparison` handle.
- Wrote region screenshots and pixel diff PNGs under `outDir` when provided.
- Added focused JS integration tests for `cvd.compare.region(...)`, including raw-object rejection.
- Added canonical namespace wiring:
  - `cvd.snapshot.page(...)`,
  - `cvd.diff.structural(...)`,
  - `cvd.image.diff(...)`,
  - `cvd.catalog.create(...)`,
  - `cvd.config.load(...)`.
- Updated repository-scanned tests and examples to use canonical names.
- Rewrote `internal/cssvisualdiff/dsl/scripts/compare.js` to use `require("css-visual-diff")` and `cvd.compare.region(...)`.
- Removed built-in compare script usage of `require("diff")` and `require("report")`.
- Updated `script compare brief` to render a concise text brief from comparison data.
- Updated embedded JS API docs and JS verbs docs.
- Added smoke scripts:
  - `scripts/006-js-compare-region-smoke.sh`,
  - `scripts/007-canonical-api-surface-smoke.sh`,
  - `scripts/008-built-in-compare-dogfood-smoke.sh`.
- Marked Phases 6, 7, and 8 complete in `tasks.md`.
- Related implementation files and smoke scripts to this diary.

### Why

- Users need a low-effort one-call comparison path.
- Built-ins should prove the public API is sufficient.
- Canonical namespaces reduce confusion between structural diffs, image diffs, snapshots, config loading, and catalogs.

### What worked

- Focused package tests passed:

```bash
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./internal/cssvisualdiff/jsapi -count=1
```

- Phase 6 smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/006-js-compare-region-smoke.sh
```

- Phase 7 smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/007-canonical-api-surface-smoke.sh
```

- Phase 8 smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/008-built-in-compare-dogfood-smoke.sh
```

- Full test suite passed:

```bash
go test ./... -count=1
```

### What didn't work

- No major blocker in this step. The main adjustment was keeping old top-level implementation aliases present while moving public docs/tests/examples to canonical names. Removing every alias immediately would require a broader migration of existing tests and internal callers that is better handled in a focused cleanup if desired.

### What I learned

- `compare.region` can be implemented as an ordinary collect-and-compare orchestration without introducing a workflow builder.
- The built-in compare verb is a good pressure test: if it can be implemented with public primitives, user scripts can too.
- Canonical docs and smoke scripts catch stale examples quickly; the Phase 7 smoke rejects `require("diff")`, `require("report")`, and old public shapes in examples/built-ins/API docs.

### What was tricky to build

- `cvd.compare.region(...)` must avoid nested page locks. It collects and screenshots each side under that side's `runExclusive` lock, then releases both locks before comparing pure data.
- Same-page comparisons need deterministic serialization; separate-page comparisons should not deadlock. Sequential side collection satisfies both for now.
- The built-in compare output shape intentionally changed from the old `computed_diffs` / `pixel_diff` mode-shaped result to `cssvd.selectionComparison.v1` lowerCamel comparison data. Tests were updated accordingly.

### What warrants a second pair of eyes

- Review whether old top-level aliases (`cvd.snapshot`, `cvd.diff`, `cvd.catalog`, `cvd.loadConfig`) should be removed now or in a later dedicated cleanup. Public docs no longer teach them.
- Review `compare.region` artifact semantics. It always captures region screenshots and writes pixel diff artifacts under `outDir` or a temp directory.
- Review `script compare brief`; it is deliberately simple and no longer uses the internal report helper.

### What should be done in the future

- Phase 9 should integrate comparison objects with catalogs and richer report/artifact flows.
- If strict no-alias cleanup is desired, remove old top-level exports and update any remaining internal tests in a dedicated pass.
- A binary smoke that invokes `go run ./cmd/css-visual-diff verbs script compare region ...` against a live local server could supplement the current focused host smoke.

### Code review instructions

- Start with `internal/cssvisualdiff/jsapi/compare.go`, especially `compareRegion` and `collectAndScreenshotRegion`.
- Review `internal/cssvisualdiff/dsl/scripts/compare.js` to see the built-in dogfood usage.
- Review `internal/cssvisualdiff/doc/topics/javascript-api.md` for canonical namespace guidance.
- Validate with:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/006-js-compare-region-smoke.sh
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/007-canonical-api-surface-smoke.sh
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/008-built-in-compare-dogfood-smoke.sh
go test ./... -count=1
```

### Technical details

Canonical quick path:

```js
const comparison = await cvd.compare.region({
  left: leftPage.locator("#cta"),
  right: rightPage.locator("#cta"),
  outDir: "artifacts/cta",
})
```

Expanded primitive path:

```js
const left = await leftPage.locator("#cta").collect({ inspect: "rich" })
const right = await rightPage.locator("#cta").collect({ inspect: "rich" })
const comparison = await cvd.compare.selections(left, right)
```

## Step 19: Real-site validation of the public compare-region path

The user asked whether the new API had been tested against a real site, then explicitly requested a real-site validation. I ran the built-in `script compare region` command against live HTTPS pages. This matters because the previous validation used real Chromium but local fixture pages; external sites add network, TLS, real fonts, real layout, and production HTML/CSS.

The built-in command is now a dogfood client of the public JavaScript API. Therefore this validation exercises the path: built-in JS verb, `require("css-visual-diff")`, `cvd.compare.region(...)`, collection, region screenshots, pixel diffing, JSON output, Markdown output, and PNG artifacts.

### Prompt Context

**User prompt (verbatim):** "run a real site validation."

**Assistant interpretation:** Run the newly implemented compare-region API against external live websites, inspect the output and artifacts, and preserve the findings in the ticket.

**Inferred user intent:** Confirm that the implementation works outside synthetic local tests before continuing deeper implementation phases.

### What I did

- Ran `go run ./cmd/css-visual-diff verbs script compare region --help` to confirm current flags.
- Ran a real-site equivalent-content comparison:
  - left: `https://example.com/`, selector `body`,
  - right: `https://example.org/`, selector `body`.
- Ran a real-site different-content comparison:
  - left: `https://example.com/`, selector `body`,
  - right: `https://www.iana.org/domains/reserved`, selector `main`.
- Preserved outputs and artifacts under `validation/real-site-example` and `validation/real-site-different`.
- Wrote `validation/01-real-site-validation.md` with commands, results, artifacts, and caveats.

### Why

- Local fixture tests prove regression behavior, but they do not prove the API works against live HTTPS sites.
- The compare-region flow includes several integration points that are best validated together: navigation, selector lookup, screenshot capture, PNG diffing, and artifact writing.

### What worked

The equivalent-content run produced zero visual difference:

```text
left:  https://example.com/ body
right: https://example.org/ body
changedPercent: 0
changedPixels: 0
totalPixels: 100608
bounds.changed: false
text.changed: false
```

The different-content run detected real differences:

```text
left:  https://example.com/ body
right: https://www.iana.org/domains/reserved main
changedPercent: 10.295429500970773
changedPixels: 98100
totalPixels: 952850
normalized size: 850x1121
bounds.changed: true
text.changed: true
style changes: background-color, color, font-family
```

Both runs produced:

```text
left_region.png
right_region.png
diff_only.png
diff_comparison.png
compare.json
compare-report.txt
command-output.json
```

### What didn't work

- No command failure occurred.
- The first run compared `example.com` and `example.org`, which are intentionally equivalent in this browser context and returned a zero diff. That is useful as a sanity check but not sufficient as a difference-detection test, so I added the second comparison against IANA's reserved-domains page.

### What I learned

- The new compare-region path works against live HTTPS pages, not just local fixtures.
- The artifact bundle is complete and reviewable.
- The output schema is the expected `cssvd.selectionComparison.v1`.

### What was tricky to build or validate

- Selecting useful real-site targets matters. `example.com` versus `example.org` is a good zero-diff sanity check but not a good changed-site validation because the rendered content is equivalent. A second run against a real different page was necessary.
- External-site validation should remain small and explicit. It can be flaky if sites change, but preserving command output and artifacts makes the result reviewable.

### What warrants a second pair of eyes

- Review the retained PNG artifacts in the validation directories to confirm the visual evidence matches the JSON summaries.
- Decide whether these real-site artifacts should be committed. They are useful evidence but are generated binary files.

### What should be done in the future

- Add an optional real-site smoke script if external network validation should become replayable.
- Run a Pyxis-specific real target comparison when a stable target pair and selectors are available.

### Code review instructions

- Inspect `validation/01-real-site-validation.md` first.
- Then inspect `validation/real-site-different/compare.json` and `validation/real-site-different/diff_comparison.png`.
- Re-run one validation with:

```bash
go run ./cmd/css-visual-diff verbs script compare region \
  --leftUrl https://example.com/ \
  --rightUrl https://www.iana.org/domains/reserved \
  --leftSelector body \
  --rightSelector main \
  --width 1280 \
  --height 720 \
  --leftWaitMs 1500 \
  --rightWaitMs 1500 \
  --outDir /tmp/cssvd-real-site-validation \
  --threshold 30 \
  --writeJson \
  --writeMarkdown \
  --writePngs \
  --output json
```

## Step 20: Implement Phase 9 comparison catalog integration

Phase 9 connects rich `SelectionComparison` handles to catalogs. A comparison is now not only something a script can inspect or write as standalone JSON/Markdown; it can be recorded into a Go-backed catalog manifest and rendered into the catalog index.

This matters because visual regression workflows often compare many sections, then need one durable index of what happened. The catalog is the natural owner of that index.

### Prompt Context

**User prompt (verbatim):** "do catalog 9"

**Assistant interpretation:** Implement Phase 9 catalog/report/artifact integration for comparison objects, with service model changes, JS API, tests, docs, smoke script, diary/changelog/task updates, and commits.

**Inferred user intent:** Make the new comparison API useful for multi-comparison workflows and durable artifact catalogs.

**Commit (code):** 1227f1a — "feat: record selection comparisons in catalogs"

### What I did

- Added `CatalogComparisonRecord` to the service catalog model.
- Added `Comparisons []CatalogComparisonRecord` to `CatalogManifest`.
- Added `ComparisonCount` to `CatalogSummary`.
- Added `Catalog.AddComparison(target, comparison)`.
- Updated catalog artifact counting to include comparison artifacts.
- Updated catalog index rendering with a `## Comparisons` table showing target, comparison name, changed percent, style changes, attribute changes, and artifact count.
- Added JS `catalog.record(comparison, target?)`.
- Updated JS manifest lowering to include `comparisons` and lowerCamel `comparisonCount`.
- Added integration coverage in `TestCVDModuleRecordsComparisonInCatalog`.
- Updated embedded JavaScript API docs with `catalog.record(...)`.
- Added `scripts/009-comparison-catalog-smoke.sh`.
- Marked Phase 9 tasks complete in `tasks.md`.

### Why

- `SelectionComparison` is the central analysis object, but workflows need to collect many comparisons into one durable output.
- The catalog service already owns manifests, path normalization, summaries, and index rendering. It should own comparison records too.
- Scripts should be able to do: compare, write artifacts, record comparison, write manifest/index.

### What worked

- Focused integration smoke passed:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleRecordsComparisonInCatalog' -count=1
```

- Service catalog regression passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestCatalogWritesManifestAndIndex' -count=1
```

- Ticket smoke passed:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/009-comparison-catalog-smoke.sh
```

- Full test suite passed:

```bash
go test ./... -count=1
```

### What didn't work

- The first integration test run returned `nil` for `manifest.summary.comparisonCount`. The service summary had `ComparisonCount`, but the JS lowering function `lowerCatalogSummary` had not been updated to expose lowerCamel `comparisonCount`. I added it and reran the focused test successfully.

### What I learned

- Whenever service manifests gain fields, the JS lowerers must be updated immediately. Otherwise scripts see a partially upgraded API.
- Catalog integration is most useful when the catalog index renders the comparison summary directly, not only stores raw JSON in the manifest.

### What was tricky to build

- Catalog artifact counting previously only counted inspect artifacts. Comparisons have a different artifact shape, so the summary now counts `comparison.Comparison.Artifacts`.
- `catalog.record(comparison, target?)` must unwrap a strict `cvd.selectionComparison` handle. Raw comparison JSON is not accepted in this API because Phase 9 is still handle-oriented.
- When `target` is omitted, the JS adapter derives one from comparison metadata. Explicit target records remain better for production catalogs.

### What warrants a second pair of eyes

- Review whether `CatalogComparisonRecord` should store full `SelectionComparisonData` or a compact summary plus artifact links. Full data is useful now but may make manifests large.
- Review the catalog index table fields. They are intentionally compact, but richer report links may belong in Phase 10.

### What should be done in the future

- Phase 10 should refresh public examples and the pixel-accuracy guide around comparison catalogs.
- A future API could add `catalog.recordMany(comparisons)` for batch workflows.

### Code review instructions

- Start with `internal/cssvisualdiff/service/catalog_service.go`, especially `CatalogComparisonRecord`, `AddComparison`, `Summary`, and `WriteIndex`.
- Review `internal/cssvisualdiff/jsapi/catalog.go`, especially `catalog.record`.
- Review `TestCVDModuleRecordsComparisonInCatalog` for end-to-end JS usage.
- Validate with:

```bash
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/009-comparison-catalog-smoke.sh
go test ./... -count=1
```

## Issues and assumptions

- I did not implement code in this turn; this is a design/ticket setup/refinement task.
- The existing built-in `require("diff")` helper remains undocumented in the public API docs.
- The proposed API assumes locator handles can be unwrapped by the JS API compare function using the existing default proxy registry.
- The first implementation should avoid locking two page mutexes simultaneously; capture each page side sequentially with `pageState.runExclusive`.

## Next steps

1. Start Phase 1 implementation: extract PNG pixel diff primitives from `modes/compare.go` into `service`.
2. Add service-level `CompareRegions(...)` before adding JS adapters.
3. Implement `cvd.compare.region(...)` under `require("css-visual-diff")`.
4. Update embedded docs and add binary smoke scripts.
