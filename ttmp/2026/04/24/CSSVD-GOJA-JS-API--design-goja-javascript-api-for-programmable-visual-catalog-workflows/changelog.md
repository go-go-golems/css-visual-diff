# Changelog

## 2026-04-24

- Initial workspace created


## 2026-04-24 — Create Goja JavaScript API design guide

### Added
- Added `design/01-goja-javascript-api-analysis-design-and-implementation-guide.md`.

### Contents
- Explains the motivation for a Goja scripting API after the Pyxis catalog work.
- Maps current css-visual-diff architecture: CLI, config schema, driver, prepare hooks, inspect/artifact modes.
- Proposes Browser, Page, Probe, Preflight, Inspect, and Catalog JavaScript APIs.
- Includes TypeScript-style API references, pseudocode, diagrams, implementation phases, tests, documentation deliverables, risks, and acceptance criteria.
- Recommends keeping core logic in Go services and exposing JS through thin Goja adapters.

### Publishing
- Uploaded to reMarkable under `/ai/2026/04/24/CSSVD-GOJA-JS-API/`.


## 2026-04-24 — Research update for repository-scanned JS verbs

### Updated
- Expanded `design/01-goja-javascript-api-analysis-design-and-implementation-guide.md` after reading the current `internal/cssvisualdiff/dsl` prototype, the Discord helper-verbs design/diary, the loupedeck jsverbs implementation, and upstream `go-go-goja/pkg/jsverbs`.
- Reframed the implementation from greenfield Goja work to productizing the existing css-visual-diff jsverbs host.
- Added a concrete plan for scanning JavaScript scripts from embedded/filesystem repositories and exposing them as CLI verbs with Glazed flags under a lazy `css-visual-diff verbs` subtree.
- Added guidance for keeping the native module API (`require("css-visual-diff")`) separate from the CLI verb API (`__verb__(...)`).
- Added an implementation research diary at `reference/01-implementation-research-diary.md`.
- Updated `tasks.md` with repository-scanned verbs, service extraction, and artifact-cleanup follow-ups.

### Key findings
- `css-visual-diff` already has a working first-generation jsverbs host in `internal/cssvisualdiff/dsl`; implementation should extend it rather than start over.
- Loupedeck provides the closest implemented reference for lazy dynamic verbs, repository discovery, embedded builtins, duplicate detection, and custom runtime invocation.
- The current root-level generated script command injection should be treated as prototype UX; the target product shape should be `css-visual-diff verbs ...`.
- Generated `css-visual-diff-compare-*` PNG artifact directories under `internal/cssvisualdiff/dsl` should be cleaned up or explicitly classified as fixtures before more JS API work.


## 2026-04-24 — Incorporate maintainer clarifications

### Updated
- Clarified that the JS API must be Promise-first from the initial implementation, not a synchronous MVP with later migration.
- Clarified that `catalog()` should be backed by a Go-side catalog service/data model and only adapted into JavaScript.
- Expanded the error model with JS-visible error classes/codes for selector, prepare, browser/CDP, and artifact failures, with support for both thrown errors and collected batch failures.
- Added a maintainer-clarifications section answering what preflight is for, how `directReactGlobal` differs from selector/probe APIs, and what parallelization is realistic with CDP.
- Updated tasks to include Promise-first APIs, Go-side catalog service, async integration tests, and typed error handling.

### Decisions
- Promise support is required from day one.
- Preflight primarily checks selector/probe readiness before expensive extraction, especially missing selectors.
- `directReactGlobal` is a prepare/rendering mode, not a selector mode.
- Parallelism should be target/page-level with explicit concurrency limits; per-page CDP operations should be treated as mostly serialized.


## 2026-04-24 — Expand implementation task plan

### Updated
- Replaced the short TODO list in `tasks.md` with a phased implementation plan from ticket hygiene through lazy verbs, service extraction, Promise-first JS module, Go-side catalog service, built-in catalog verbs, and final docs/validation.
- Added explicit commit checkpoints at the end of major phases so implementation can proceed in reviewable intervals.

### Phase sequence
1. Stabilize existing dsl/jsverbs prototype.
2. Add lazy repository-scanned `css-visual-diff verbs` command tree.
3. Extract reusable browser/page/inspect/preflight services.
4. Implement Promise-first `require("css-visual-diff")` native module.
5. Implement Go-side catalog service and JS adapter.
6. Add built-in catalog verbs and examples.
7. Finish docs, migration notes, and validation.


## 2026-04-24 — Phase 1 dsl artifact cleanup

### Changed
- Updated `internal/cssvisualdiff/dsl/host_test.go` so embedded compare verb tests write PNG/output artifacts into `t.TempDir()` instead of the source package directory.
- Removed ignored generated `internal/cssvisualdiff/dsl/css-visual-diff-compare-*` directories from the working tree.
- Verified `.gitignore` ignores generated `css-visual-diff-compare-*` output directories.

### Validation
- Ran `gofmt -w internal/cssvisualdiff/dsl/host_test.go`.
- Ran `go test ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff`.


## 2026-04-24 — Phase 2 initial lazy verbs CLI implementation

### Changed
- Added `internal/cssvisualdiff/verbcli` with repository bootstrap, embedded built-in scripts repository, environment/CLI repository discovery, jsverbs scanning, duplicate verb path detection, and per-invocation runtime ownership.
- Exported small dsl helpers so the new verb CLI can reuse embedded scripts, shared sections, and runtime module registration without duplicating the existing dsl host.
- Replaced eager root-level generated script command registration in `cmd/css-visual-diff/main.go` with a lazy `css-visual-diff verbs` subtree.
- Preserved current built-in script verbs under `css-visual-diff verbs script compare ...`.
- Added tests for built-in command registration, duplicate verb path errors, and executing a filesystem repository text verb through the custom invoker.

### Validation
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go run ./cmd/css-visual-diff --help` to verify root help lists `verbs` without eager generated commands.
- Ran `go run ./cmd/css-visual-diff verbs --help` and `go run ./cmd/css-visual-diff verbs script compare region --help` to verify lazy command generation and generated flags.

### Follow-up
- App-config repository discovery was implemented in the next Phase 2 follow-up commit.


## 2026-04-24 — Phase 2 app-config repository discovery

### Changed
- Added css-visual-diff app-config repository discovery to `internal/cssvisualdiff/verbcli/bootstrap.go`.
- Config files can declare `verbs.repositories` with `name`, `path`, and optional `enabled: false`.
- Config repositories are loaded after embedded built-ins and before environment/CLI repositories.
- Added a unit test for relative config repository paths and disabled entries.

### Validation
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.


## 2026-04-24 — Phase 3 initial service extraction

### Changed
- Added `internal/cssvisualdiff/service` with shared `Bounds`, `StyleSnapshot`, `ProbeSpec`, and `SelectorStatus` types.
- Extracted style evaluation into `service.EvaluateStyle(...)` and kept `modes.evaluateStyle(...)` as a compatibility wrapper.
- Added batched selector preflight service `service.PreflightProbes(...)` that evaluates all probes in one page-side JavaScript pass.
- Updated `modes.ensureInspectSelectorExists(...)` to use the shared preflight service instead of a style-evaluation-only check.
- Added service tests for existing, missing, invalid, and hidden selectors.

### Validation
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff`.
- Ran `go test ./...`.

### Follow-up
- Full Phase 3 still needs prepare-service extraction plus `InspectPreparedPage` / `InspectAll` service extraction.


## 2026-04-24 — Phase 3 prepare service extraction

### Changed
- Added `internal/cssvisualdiff/service/prepare.go` with exported `PrepareTarget`, `RunScriptPrepare`, `RunDirectReactGlobalPrepare`, and `BuildDirectReactGlobalScript`.
- Replaced `internal/cssvisualdiff/modes/prepare.go` implementation with compatibility wrappers around the new service package.
- Preserved existing prepare tests in `modes` while making prepare available to future JS adapters.

### Validation
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff`.
- Ran `go test ./...`.


## 2026-04-24 — Phase 3 browser/page service shell

### Changed
- Added `internal/cssvisualdiff/service/browser.go` with initial `BrowserService`, `PageService`, and `LoadAndPreparePage` helpers.
- Routed `modes.Inspect(...)` page setup through `service.LoadAndPreparePage(...)` while preserving CLI behavior.

### Validation
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff`.
- Ran `go test ./...`.


## 2026-04-24 — Phase 3 inspect artifact service extraction

### Changed
- Added `internal/cssvisualdiff/service/inspect.go` with inspect format constants, inspect request/result metadata types, `InspectPreparedPage`, `WriteInspectArtifacts`, single-artifact writers, inspect index writer, and shared HTML/inspect JSON helpers.
- Updated `modes.Inspect(...)` to delegate prepared-page artifact extraction to `service.InspectPreparedPage(...)`.
- Kept existing mode-level inspect types as aliases to service types for compatibility.
- Added a service test proving multiple probes are inspected on one prepared page without reloading the target URL per probe.

### Validation
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff`.
- Ran `go test ./...`.

### Phase status
- Phase 3 service extraction is functionally complete for the planned Browser/Page, prepare, preflight, style, and inspect artifact service layers.


## 2026-04-24 — Phase 4 Promise-first css-visual-diff module MVP

### Changed
- Added native module registration for `require("css-visual-diff")`.
- Added Promise-returning MVP methods:
  - `cvd.browser()`
  - `browser.newPage()`
  - `browser.page(url, options)`
  - `browser.close()`
  - `page.prepare(spec)`
  - `page.preflight(probes)`
  - `page.inspectAll(probes, options)`
  - `page.close()`
- Added runtime-owner promise settlement helper so native module operations resolve/reject through the go-go-goja runtime owner thread.
- Added a repository-scanned async verb integration test that uses `require("css-visual-diff")`, opens a browser/page, preflights `#cta`, writes CSS inspect artifacts, and returns a structured row.

### Validation
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.

### Follow-up
- JS result objects currently expose Go exported field names in some paths (`Results`, `OutputDir`, `Exists`); lowerCamel result codecs remain to be implemented.
- `page.goto`, single `page.inspect`, and JS-visible typed error classes remain pending.


## 2026-04-24 — Phase 4 completion: page navigation, lowerCamel results, typed JS errors

### Changed
- Completed the Promise-first page API with `page.goto(...)` and single-probe `page.inspect(...)`.
- Converted JS-visible preflight and inspect results to lowerCamel fields such as `exists`, `textStart`, `outputDir`, `results`, `inspectJson`, `targetName`, `selectorSource`, and `createdAt`.
- Added lowerCamel option decoding for `outputFile`, `waitMs`, `rootSelector`, prepare timing fields, and both `attrs`/`attributes` probe spellings.
- Added JS-visible error constructors exported by `require("css-visual-diff")`:
  - `CvdError`
  - `SelectorError`
  - `PrepareError`
  - `BrowserError`
  - `ArtifactError`
- Promise rejections now use classified JS error objects with `name`, `code`, `operation`, and `details` fields.
- Added basic validation errors for missing page URLs and empty inspect request lists/selectors.
- Extended verb integration tests to assert lowerCamel results, `page.goto(...)`, `page.inspect(...)`, and `err instanceof cvd.SelectorError`/`cvd.CvdError`.

### Validation
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.
