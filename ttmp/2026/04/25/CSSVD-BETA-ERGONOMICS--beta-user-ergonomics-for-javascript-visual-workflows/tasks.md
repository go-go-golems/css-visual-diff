---
Title: Tasks
Ticket: CSSVD-BETA-ERGONOMICS
Status: complete
Topics:
  - tooling
  - frontend
  - visual-regression
  - browser-automation
DocType: tasks
Intent: short-term
Owners: []
---

# Tasks

## Phase 1 — Selector wait helper

- [x] Add `WaitForSelectorOptions` and `WaitForSelectorResult` to `internal/cssvisualdiff/service/dom.go`.
- [x] Add `service.WaitForLocator(page, locator, opts)` using existing `LocatorStatus` polling.
- [x] Add `locator.waitFor(options?)` to `internal/cssvisualdiff/jsapi/locator.go`.
- [x] Optionally add `page.waitForSelector(selector, options?)` if it can delegate cleanly and remain tiny.
- [x] Add service tests for existing selector, delayed selector, visible wait, timeout, and invalid selector.
- [x] Add JS integration test through repository-scanned verbs.
- [x] Update embedded JS API docs.

## Phase 2 — Stable artifact write result

- [x] Update `comparison.artifacts.write(outDir, names)` result shape with stable keys:
  - `json`,
  - `markdown`,
  - `leftRegion`,
  - `rightRegion`,
  - `diffOnly`,
  - `diffComparison`,
  - `written`.
- [x] Ensure JSON/Markdown files are written only when requested.
- [x] Return known PNG paths without rewriting PNGs.
- [x] Add tests for returned paths and file existence.
- [x] Update docs with the result schema.

## Phase 3 — Multi-section catalog example

- [x] Add `examples/verbs/compare-page-catalog.js`.
- [x] Update `examples/verbs/README.md`.
- [x] Add ticket smoke `scripts/001-beta-multisection-example-smoke.sh`.
- [x] Smoke should validate `manifest.json`, `index.md`, per-section artifacts, and stdout artifact paths.

## Phase 4 — Collection profile docs

- [x] Verify actual `normalizeCollectOptions` behavior in `service/collection.go`.
- [x] Document `minimal`, `rich`, and `debug` profile recommendations.
- [x] Clarify `styleProps` and `attributes` semantics.
- [x] Add notes to `javascript-api` and `pixel-accuracy-scripting-guide`.

## Deferred explicitly

These are intentionally not part of this closed ticket. They should only be reopened if beta usage shows repeated demand for shared policy rather than project-local JavaScript checks.

- Bounds tolerance API is deferred until beta usage clarifies policy needs.
- CSS/style normalization hooks are deferred because defaults are too opinionated.
- Built-in style property presets are deferred until usage vocabulary stabilizes.

## Validation

- [x] `go test ./... -count=1`.
- [x] `make lint`.
- [x] Ticket smoke script passes.
- [x] Embedded help renders for `javascript-api` and `pixel-accuracy-scripting-guide`.
- [x] `docmgr doctor --root ./ttmp --ticket CSSVD-BETA-ERGONOMICS --stale-after 30` passes.
