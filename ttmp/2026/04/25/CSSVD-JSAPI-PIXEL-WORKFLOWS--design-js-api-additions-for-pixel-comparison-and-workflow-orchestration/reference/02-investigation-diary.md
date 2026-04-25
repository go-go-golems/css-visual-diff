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
      Note: Phase 1 embedded JavaScript API reference update for the collected selection model (commit b13933a)
    - Path: internal/cssvisualdiff/service/collection.go
      Note: Phase 1 collected selector data service model (commit b13933a)
    - Path: internal/cssvisualdiff/service/collection_test.go
      Note: Phase 1 service tests for collection profiles
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/design/01-elegant-javascript-api-additions-for-pixel-comparison-workflows.md
      Note: Main API design proposal produced from this investigation
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/reference/01-pyxis-user-feedback-source-analysis.md
      Note: Source analysis of Pyxis feedback and workflow documents
    - Path: ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/001-service-collection-smoke.sh
      Note: Phase 1 replayable service collection smoke script
ExternalSources:
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
Summary: Chronological diary for designing css-visual-diff JS API additions from Pyxis user feedback.
LastUpdated: 2026-04-25T11:10:00-04:00
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
