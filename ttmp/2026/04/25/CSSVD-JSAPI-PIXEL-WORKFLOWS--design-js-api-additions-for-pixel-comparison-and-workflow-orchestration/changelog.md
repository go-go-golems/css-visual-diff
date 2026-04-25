# Changelog

## 2026-04-25

- Created ticket `CSSVD-JSAPI-PIXEL-WORKFLOWS` to track API additions prompted by Pyxis css-visual-diff JS workflow feedback.
- Read Pyxis maintainer feature requests and workflow exploration docs.
- Added source analysis document summarizing the real user need, proposed requests, and core-versus-userland classification.
- Added main design document proposing `cvd.compare.region(...)` as the public strict API instead of promoting internal `require("diff")` or adding a loose `cvd.comparePixels(...)` helper.
- Added investigation diary with commands, reasoning, assumptions, and next steps.
- Expanded tasks into phased implementation plan covering service extraction, JS API, built-in convergence, docs, smokes, and follow-up APIs.
- Related source files and Pyxis source documents to the focused docs.
- Ran `docmgr doctor --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --stale-after 30`; all checks passed after switching ticket topics to existing vocabulary values.
- Refined the proposal around a single Go-backed `cvd.comparison` result object with `summary()`, `toJSON()`, `bounds`, `styles`, `attributes`, `report`, and lazy `artifact(s)` query/write methods, while preserving plain serializable data for CLI output and JSON files.
- Replaced the earlier opt-in `evidence` idea with an `inspect` profile. Default `inspect: "rich"` should collect broad browser facts once, then let comparison-object methods and reports filter after the fact.
- Added `design/02-javascript-centric-collected-data-and-comparison-object-api.md`, a detailed intern-facing report that reframes the API around `CollectedSelection` objects and `SelectionComparison` objects, with `cvd.compare.region(...)` as a convenience wrapper over collect-then-compare.
- Added and refined `design/03-full-javascript-api-coherence-and-fluent-primitive-design.md`, an API-wide design pass. Updated it to treat backward compatibility as non-required and to prefer an opinionated low-effort surface plus composable primitive surface over historical aliases.
- Rewrote `tasks.md` into a detailed 12-phase implementation plan. Each phase now includes implementation tasks, tests, JavaScript API reference updates, and real ticket-local smoke scripts.
- Implemented Phase 1 collected selector data service in commit `b13933a`:
  - added `internal/cssvisualdiff/service/collection.go`,
  - added `internal/cssvisualdiff/service/collection_test.go`,
  - updated `internal/cssvisualdiff/doc/topics/javascript-api.md`,
  - added `scripts/001-service-collection-smoke.sh`,
  - marked Phase 1 tasks complete in `tasks.md`.
- Validation for Phase 1 passed with focused service tests, the ticket smoke script, embedded help rendering, and `go test ./... -count=1`.
- Implemented Phase 2 pixel diff service primitives in commit `6ca2498`:
  - added `internal/cssvisualdiff/service/pixel.go`,
  - added `internal/cssvisualdiff/service/pixel_test.go`,
  - routed compare and pixeldiff modes through the new service,
  - kept mode-local helper names as wrappers around service primitives,
  - updated `internal/cssvisualdiff/doc/topics/javascript-api.md` with structural-vs-image diff guidance and future `cvd.image.diff(...)`,
  - added `scripts/002-pixel-service-smoke.sh`,
  - marked Phase 2 tasks complete in `tasks.md`.
- Validation for Phase 2 passed with focused service tests, modes integration tests, the ticket smoke script, embedded help checks, and `go test ./... -count=1`.
- Implemented Phase 3 service-level selection comparison in commit `29c8aca`:
  - added `internal/cssvisualdiff/service/selection_compare.go`,
  - added `internal/cssvisualdiff/service/selection_compare_test.go`,
  - implemented `SelectionComparisonData` with schema version `cssvd.selectionComparison.v1`,
  - added bounds/text/style/attribute diff helpers,
  - integrated screenshot pixel comparison through the Phase 2 pixel service,
  - updated `internal/cssvisualdiff/doc/topics/javascript-api.md` with `SelectionComparison` concepts,
  - added `scripts/003-selection-compare-service-smoke.sh`,
  - marked Phase 3 tasks complete in `tasks.md`.
- Validation for Phase 3 passed with focused comparison tests, JSON fixture smoke generation, embedded help checks, and `go test ./... -count=1`.
- Implemented Phases 4 and 5 JavaScript collected/comparison handles in commit `5c76cd7`:
  - added `internal/cssvisualdiff/jsapi/collect.go`,
  - added `internal/cssvisualdiff/jsapi/compare.go`,
  - wired `locator.collect(...)` and `cvd.collect.selection(...)`,
  - wired `cvd.compare.selections(...)`,
  - added Go-backed `cvd.collectedSelection` and `cvd.selectionComparison` Proxy handles,
  - added Proxy property getters for namespaces like `comparison.styles.diff()`,
  - made Proxy handles safe for Promise resolution by returning `undefined` for `.then`,
  - added repository-scanned JS integration tests in `internal/cssvisualdiff/verbcli/command_test.go`,
  - updated `internal/cssvisualdiff/doc/topics/javascript-api.md`,
  - added `scripts/004-js-collected-selection-smoke.sh` and `scripts/005-js-selection-comparison-smoke.sh`,
  - marked Phases 4 and 5 complete in `tasks.md`.
- Validation for Phases 4 and 5 passed with focused verbcli integration tests, both ticket smoke scripts, embedded help checks, and `go test ./... -count=1`.
- Implemented Phases 6 through 8 in commit `88ddac5`:
  - added `cvd.compare.region({ left, right, ... })` as the opinionated low-effort collect/screenshot/compare API,
  - added canonical namespace wiring for `cvd.snapshot.page`, `cvd.diff.structural`, `cvd.image.diff`, `cvd.catalog.create`, and `cvd.config.load`,
  - updated examples, built-in catalog scripts, and tests to use canonical names,
  - rewrote built-in `script compare region` and `script compare brief` to dogfood public `require("css-visual-diff")` primitives,
  - removed built-in compare usage of internal `require("diff")` and `require("report")`,
  - updated embedded API/verbs/tutorial docs,
  - added `scripts/006-js-compare-region-smoke.sh`, `scripts/007-canonical-api-surface-smoke.sh`, and `scripts/008-built-in-compare-dogfood-smoke.sh`,
  - marked Phases 6, 7, and 8 complete in `tasks.md`.
- Validation for Phases 6 through 8 passed with focused package tests, all three new ticket smoke scripts, embedded help checks, and `go test ./... -count=1`.

## Key decision

The public API should center on a comparison object:

```js
const comparison = await cvd.compare.region({
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
})

await comparison.artifacts.write(outDir, ["diffComparison", "json", "markdown"])
return comparison.toJSON()
```

not:

```js
require("diff").compareRegion(...)
```

and not primarily:

```js
cvd.comparePixels({ left: { page, selector }, right: { page, selector } })
```

The design keeps one public module name, preserves strict Go-backed handles, and pushes project-specific registries/policies/reports into userland.
