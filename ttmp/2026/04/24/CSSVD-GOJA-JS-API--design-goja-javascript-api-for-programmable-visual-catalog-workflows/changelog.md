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


## 2026-04-24 — Retroactive replay scripts and binary smoke scripts

### Added
- Added `scripts/001-phase1-dsl-temp-artifacts-test.sh` through `scripts/008-binary-js-api-typed-error-smoke.sh` under this ticket.
- The scripts replay the main validation steps used during Phases 1–4, including targeted Go test suites, the full `go test ./...`, compiled-binary help smoke, compiled-binary JS API success smoke, and compiled-binary typed-error smoke.

### Validation
- Ran the compiled-binary smoke scripts:
  - `scripts/006-binary-help-smoke.sh`
  - `scripts/007-binary-js-api-success-smoke.sh`
  - `scripts/008-binary-js-api-typed-error-smoke.sh`


## 2026-04-24 — Phase 5 Go-side catalog service and JS adapter

### Added
- Added `internal/cssvisualdiff/service/catalog_service.go` with a versioned catalog manifest, target/preflight/result/failure records, summary calculation, slug/path normalization, `artifactDir` derivation, `manifest.json` writer, and Markdown index writer.
- Added `internal/cssvisualdiff/service/catalog_service_test.go` covering manifest JSON, Markdown index output, path normalization, slug normalization, summary counts, and failure records.
- Added `internal/cssvisualdiff/dsl/catalog_adapter.go` and exported `cvd.catalog(options)` from `require("css-visual-diff")`.
- Implemented JS-facing catalog methods: `artifactDir`, `addTarget`, `recordPreflight`, `addResult`, `addFailure`, `summary`, `manifest`, `writeManifest`, and `writeIndex`.
- Added a repository-scanned JS integration test that writes a catalog manifest and index from a verb.
- Added `scripts/009-binary-catalog-smoke.sh` for compiled-binary validation of `cvd.catalog(...)`.

### Validation
- Ran `scripts/009-binary-catalog-smoke.sh`.
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.


## 2026-04-24 — Phase 6 partial: built-in catalog inspect-page verb and examples

### Added
- Added embedded built-in `verbs catalog inspect-page` in `internal/cssvisualdiff/dsl/scripts/catalog.js`.
- The verb uses the Promise-first `require("css-visual-diff")` API and Go-side catalog service to inspect one URL/selector, write artifacts, record preflight, write `manifest.json`, and write `index.md`.
- Added authoring-mode behavior (`failOnMissing=false`) that records selector misses and returns a structured row without failing.
- Added CI-mode behavior (`--failOnMissing`) that writes manifest/index and then exits non-zero on selector misses.
- Added `examples/verbs/catalog-inspect-page.js` as an external repository-scanned verb example.
- Added `examples/verbs/README.md` with success, authoring-mode, and CI-mode command examples.
- Added `scripts/010-binary-built-in-catalog-inspect-page-smoke.sh` to validate the built-in verb with the compiled binary against a local HTTP fixture.

### Validation
- Ran `scripts/010-binary-built-in-catalog-inspect-page-smoke.sh`.
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.

### Remaining Phase 6 work
- Add the YAML/config interop verb (`catalog inspect-config`) or equivalent `cvd.loadConfig(...)` helpers.


## 2026-04-24 — Phase 6 completion: YAML config interop catalog verb

### Added
- Added `cvd.loadConfig(path)` as a Promise-returning native module helper backed by Go `config.Load`.
- Added lowerCamel config conversion for metadata, targets, prepare specs, sections, styles, output, and modes.
- Added embedded built-in `verbs catalog inspect-config` to inspect the `original` or `react` side of a css-visual-diff YAML config into a catalog.
- Added `scripts/011-binary-built-in-catalog-inspect-config-smoke.sh` to validate `inspect-config` through the compiled binary.

### Validation
- Ran `scripts/011-binary-built-in-catalog-inspect-config-smoke.sh`.
- Ran `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.


## 2026-04-24 — Phase 7 docs and final validation

### Added
- Added `docs/js-api.md` documenting `require("css-visual-diff")`, browser/page methods, preflight, prepare modes, inspect artifacts, catalog API, YAML config loading, typed JS errors, and concurrency guidance.
- Added `docs/js-verbs.md` documenting repository-scanned verbs, `__verb__` metadata, generated flags, binding modes, output modes, built-in catalog commands, duplicate path errors, and migration to the `verbs` namespace.
- Updated `README.md` with `css-visual-diff verbs ...` examples, built-in catalog command examples, external repository usage, and the root-command migration note.

### Validation
- Ran final Go tests and binary smoke scripts for help, JS API success, typed errors, catalog service, built-in `inspect-page`, and built-in `inspect-config`.


## 2026-04-24 — Big Brother assessment document

### Added
- Added `review/01-big-brother-code-review-and-assessment.md` with a senior-style assessment of the Goja/jsverbs implementation work.
- Generated a local PDF copy under `review/pdf/` for manual transfer/reference.

### reMarkable upload status
- Dry-run upload via `remarquee upload md --dry-run` succeeded.
- Actual cloud upload failed repeatedly with rmapi/reMarkable cloud `request failed with status 400`, including via `remarquee upload md`, `remarquee cloud put`, direct `rmapi put`, and the legacy uploader script.
- Existing remote folder `/ai/2026/04/24/CSSVD-GOJA-JS-API` is readable, but create/upload operations currently fail.


## 2026-04-24 — CI fix: Chrome sandbox in GitHub Actions

### Fixed
- Fixed GitHub Actions unit test failures where Chromium crashed with `No usable sandbox!` on `ubuntu-latest`.
- `driver.NewBrowser` now includes `chromedp.NoSandbox` automatically when running under `CI=true`, `GITHUB_ACTIONS=true`, or as root.
- Added `CSS_VISUAL_DIFF_CHROME_NO_SANDBOX` as an explicit override for local/CI environments.
- Added driver tests for the sandbox environment override parsing.

### Validation
- Ran `go test ./internal/cssvisualdiff/driver ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `go test ./...`.


## 2026-04-24 — CI fix follow-up: govulncheck and gosec

### Fixed
- Bumped the Go toolchain directive from `go 1.26.1` to `go 1.26.2` so GitHub Actions uses the standard-library release containing fixes for the reported `crypto/x509`, `crypto/tls`, and `html/template` vulnerabilities.
- Handled `goja.Object.Set` return values in the DSL runtime registrar to satisfy gosec `G104`.

### Validation
- Ran `go test ./...`.
- Ran `CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `govulncheck ./...` (`No vulnerabilities found`).
- Ran `gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history -exclude-dir=ttmp ./...` (`Issues: 0`).


## 2026-04-24 — PR review feedback fixes

### Fixed
- `service.InspectPreparedPage` now rejects `OutputFile` unless exactly one inspect request is provided, preventing multi-probe `inspectAll(..., { outputFile })` calls from overwriting the same file repeatedly.
- `verbcli.repositoriesFromArgs` now only consumes `--repository` / `--verb-repository` flags from the bootstrap prefix and honors `--`, so verb-level `--repository` flags are passed through to generated commands instead of being hijacked by bootstrap discovery.
- `catalog.manifest()` now includes lowerCamel `preflights`, `results`, and `failures`, matching the records persisted by `writeManifest()`.

### Validation
- Ran `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff`.
- Ran `go test ./...`.
- Ran `CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- Ran `govulncheck ./...`.
- Ran `gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history -exclude-dir=ttmp ./...`.
