# Tasks

## Tracking rules

- Work task-by-task and keep this file current before moving to the next phase.
- Keep implementation commits focused. Prefer one commit per phase, or one commit per risky subphase if a phase is large.
- Update `reference/01-investigation-diary.md` after each meaningful implementation step, including failed commands and validation output.
- Keep `changelog.md` aligned with completed milestones and commit hashes.
- Do not mark a phase complete until its validation commands pass.
- Current active implementation phase: **Complete — implementation, embedded docs, examples, and smokes delivered**.

## Phase 0 — Ticket, design, and baseline bookkeeping

- [x] Create ticket workspace for the flexible lower-level JavaScript API design.
- [x] Read current user-facing docs: `README.md`, `docs/js-api.md`, and `docs/js-verbs.md`.
- [x] Read the prior Goja/jsverbs implementation guide for historical context.
- [x] Read current implementation files for config schema, Goja module adapters, repository-scanned verbs, and reusable services.
- [x] Read the Discord bot Goja Proxy UI DSL references and implementation files.
- [x] Write the detailed intern-facing analysis/design/implementation guide.
- [x] Update the design guide with the `internal/cssvisualdiff/jsapi` subpackage plan.
- [x] Update the design guide with the Go-backed Proxy builder/handle model.
- [x] Remove workflow builder from the implementation plan; use ordinary JavaScript functions/loops for orchestration.
- [x] Generate local PDF bundles.
- [x] Upload the updated design guide bundle to reMarkable as `/ai/2026/04/24/cssvd-flex-api-updated`.
- [x] Commit ticket/design docs: `17240de docs: add flexible js api implementation ticket`.

## Phase 1 — Move native API into `internal/cssvisualdiff/jsapi` with no behavior changes

Goal: split the native `require("css-visual-diff")` module implementation out of `internal/cssvisualdiff/dsl` while preserving all existing JS API behavior.

### Phase 1.1 — Orientation and safety checks

- [x] Read the current diary before implementation.
- [x] Confirm current branch and git status before editing.
- [x] Commit existing ticket docs before changing code.
- [x] Run a baseline targeted test set before completing the phase:
  - `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`
  - `go test ./internal/cssvisualdiff/service`

### Phase 1.2 — Mechanical package move

- [x] Create `internal/cssvisualdiff/jsapi/`.
- [x] Move `internal/cssvisualdiff/dsl/cvd_module.go` to `internal/cssvisualdiff/jsapi/module.go`.
- [x] Move `internal/cssvisualdiff/dsl/catalog_adapter.go` to `internal/cssvisualdiff/jsapi/catalog.go`.
- [x] Move `internal/cssvisualdiff/dsl/config_adapter.go` to `internal/cssvisualdiff/jsapi/config.go`.
- [x] Change moved files from `package dsl` to `package jsapi`.
- [x] Rename `registerCVDModule(...)` to exported `Register(...)`.
- [x] Add `internal/cssvisualdiff/jsapi/codec.go` or otherwise provide local `decodeInto(...)` helpers needed by moved adapters.
- [x] Keep `dsl`-owned helpers such as `toPlainValue(...)` available for existing `diff` / `report` modules.

### Phase 1.3 — Rewire runtime registration

- [x] Update `internal/cssvisualdiff/dsl/registrar.go` to import `internal/cssvisualdiff/jsapi`.
- [x] Replace `registerCVDModule(ctx, reg)` with `jsapi.Register(ctx, reg)`.
- [x] Confirm there is no import cycle between `dsl` and `jsapi`.
- [x] Confirm `jsapi` imports only service/config/engine/goja dependencies, not `dsl`.

### Phase 1.4 — Compilation and formatting

- [x] Run `gofmt -w` on touched Go files.
- [x] Run targeted compile/tests:
  - `go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl`
  - `go test ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`
- [x] Fix any package-private symbol or import errors from the move.

### Phase 1.5 — Behavior-preservation validation

- [x] Run existing JS module integration tests:
  - `go test ./internal/cssvisualdiff/verbcli -run 'TestLazyCommand.*(CVD|Catalog|Error|Inspect)' -count=1`
- [x] Run service tests that existing adapters depend on:
  - `go test ./internal/cssvisualdiff/service -count=1`
- [x] Run broader validation:
  - `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1`
- [x] Run full suite if targeted validation is clean:
  - `go test ./... -count=1`

### Phase 1.6 — Documentation and commit

- [x] Update diary with what moved, validation output, and any failures.
- [x] Update changelog with Phase 1 result.
- [x] Mark Phase 1 tasks complete.
- [x] Commit Phase 1 code and docs together if they describe the same completed implementation milestone.

## Phase 2 — Add Proxy infrastructure and typed unwrapping helpers

Goal: establish the Go-backed object model before adding new public lower-level APIs.

### Phase 2.1 — Proxy helper design

- [x] Add `internal/cssvisualdiff/jsapi/proxy.go`.
- [x] Define a `ProxyRegistry` or equivalent identity mechanism for mapping Goja Proxy values back to Go structs.
- [x] Define shared method metadata: owner name, available methods, and optional wrong-parent hints.
- [x] Add helper for constructing Proxy-backed objects with a Go backing value.

### Phase 2.2 — Error helpers

- [x] Add unknown-method error helper with available method list and optional did-you-mean suggestion.
- [x] Add wrong-parent error helper for methods that exist on another object type.
- [x] Add type-mismatch error helper for strict APIs expecting Go-backed handles/builders.
- [x] Ensure errors are surfaced as JS-visible typed `CvdError` or a more specific subclass where appropriate.

### Phase 2.3 — Typed unwrapping

- [x] Add `internal/cssvisualdiff/jsapi/unwrap.go`.
- [x] Add unwrap helpers for browser/page/catalog handles if converted in this phase.
- [x] Add planned unwrap helpers or placeholders for locator/probe/extractor handles.
- [x] Ensure unwrap errors include operation names and migration hints.

### Phase 2.4 — Tests and validation

- [x] Add Goja/runtime tests for unknown method errors.
- [x] Add Goja/runtime tests for wrong-parent errors.
- [x] Add Goja/runtime tests for raw-object type mismatch errors.
- [x] Run `go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli -count=1`.
- [x] Update diary/changelog and commit Phase 2.

## Phase 3 — Add per-page operation serialization

Goal: make lower-level locator APIs safe when users call several page operations concurrently.

- [x] Add a per-page mutex or queue to the page handle/state.
- [x] Wrap `goto`, `prepare`, `preflight`, `inspect`, `inspectAll`, and `close` in the per-page operation guard.
- [x] Ensure operations on different pages can still run concurrently.
- [x] Add a test using `Promise.all` on one page that proves calls complete without races/panics/hangs.
- [x] Add a test using two pages to ensure page-level isolation.
- [x] Run targeted and full tests.
- [x] Update diary/changelog and commit Phase 3.

## Phase 4 — Implement service DOM locator primitives

Goal: add Go service functions for locator status/text/html/bounds/attributes/style without importing Goja.

- [x] Add `internal/cssvisualdiff/service/dom.go`.
- [x] Define `LocatorSpec`, `TextOptions`, `ElementHTML`, and related option/result structs if needed.
- [x] Implement `LocatorStatus` reusing selector readiness logic from `PreflightProbes` where possible.
- [x] Implement `LocatorText`.
- [x] Implement `LocatorHTML`.
- [x] Implement `LocatorBounds`.
- [x] Implement `LocatorAttributes`.
- [x] Implement `LocatorComputedStyle` reusing `EvaluateStyle` where possible.
- [x] Add `internal/cssvisualdiff/service/dom_test.go` with existing/missing/hidden/invalid selector cases.
- [x] Run `go test ./internal/cssvisualdiff/service -count=1`.
- [x] Update diary/changelog and commit Phase 4.

## Phase 5 — Expose `page.locator()` and locator methods as Go-backed Proxy handles

Goal: ship the first visible lower-level JS API.

- [x] Add `internal/cssvisualdiff/jsapi/locator.go`.
- [x] Add `page.locator(selector)` as a synchronous method returning a Go-backed Proxy handle.
- [x] Add Promise-returning `locator.status()`.
- [x] Add Promise-returning `locator.exists()`.
- [x] Add Promise-returning `locator.visible()`.
- [x] Add Promise-returning `locator.text(options)`.
- [x] Add Promise-returning `locator.bounds()`.
- [x] Add Promise-returning `locator.computedStyle(props)`.
- [x] Add Promise-returning `locator.attributes(names)`.
- [x] Add wrong-parent guidance for probe/extractor methods accidentally called on locators.
- [x] Add repository-scanned JS verb tests for locator methods.
- [x] Update docs/examples if public API is considered ready.
- [x] Update diary/changelog and commit Phase 5.

## Phase 6 — Add Go-backed target/probe/extractor builders

Goal: make YAML replacement ergonomic and strongly validated.

### Phase 6.1 — Target and viewport builders

- [x] Add `internal/cssvisualdiff/jsapi/target.go`.
- [x] Implement `cvd.target(name)` as a Go-backed `TargetBuilder` Proxy.
- [x] Implement `cvd.viewport(width, height)` and named viewport helpers if included in the first cut.
- [x] Add builder validation for URL, viewport dimensions, wait time, root selector, and prepare settings.

### Phase 6.2 — Probe builders

- [x] Add `internal/cssvisualdiff/jsapi/probe.go`.
- [x] Implement `cvd.probe(name)` as a Go-backed `ProbeBuilder` Proxy.
- [x] Add `.selector(...)`, `.required(...)`, `.source(...)`, `.text()`, `.bounds()`, `.styles(...)`, `.attributes(...)`.
- [x] Add helpful errors for `.style(...)` vs `.styles(...)` and other common mistakes.

### Phase 6.3 — Extractor builders

- [x] Add `internal/cssvisualdiff/jsapi/extractor.go`.
- [x] Implement `cvd.extractors.exists()`.
- [x] Implement `cvd.extractors.visible()`.
- [x] Implement `cvd.extractors.text()`.
- [x] Implement `cvd.extractors.bounds()`.
- [x] Implement `cvd.extractors.computedStyle(props)`.
- [x] Implement `cvd.extractors.attributes(names)`.

### Phase 6.4 — Validation

- [x] Add tests for builder chaining.
- [x] Add tests that strict APIs can unwrap builders without requiring `.build()`.
- [x] Add tests for invalid builder arguments.
- [x] Update diary/changelog and commit Phase 6.

## Phase 7 — Add strict `cvd.extract(locator, extractors)`

Goal: composable extraction with typed Go-backed inputs.

- [x] Add `internal/cssvisualdiff/service/extract.go`.
- [x] Define `ExtractorSpec`, `ElementSnapshot`, extraction options, and error/status result shapes.
- [x] Implement extraction from a single locator using multiple extractors.
- [x] Expose `cvd.extract(locator, extractors, options)` in `jsapi`.
- [x] Require `LocatorHandle` and `ExtractorHandle` Proxy values; reject raw JS objects with helpful migration hints.
- [x] Add tests for multiple facts from one locator.
- [x] Add tests for missing selector behavior.
- [x] Add tests for invalid selector typed errors.
- [x] Update diary/changelog and commit Phase 7.

## Phase 8 — Add strict `cvd.snapshot(page, probes, options)`

Goal: inspect many Go-backed probe builders into an in-memory structured result without necessarily writing standard inspect artifacts.

- [x] Define `PageSnapshot` service result shape.
- [x] Implement snapshot orchestration over a page and multiple probe builders.
- [x] Expose `cvd.snapshot(page, probes, options)`.
- [x] Require a `PageHandle` Proxy and Go-backed `ProbeBuilder` values.
- [x] Return plain JSON-serializable snapshot data.
- [x] Add optional artifact writing only when explicitly requested.
- [x] Add tests for strict raw-object rejection.
- [x] Add tests for snapshot data shape stability.
- [x] Update diary/changelog and commit Phase 8.

## Phase 9 — Add diff, report, and write primitives

Goal: compare snapshots without YAML.

- [x] Add `internal/cssvisualdiff/service/diff.go`.
- [x] Implement deterministic structural JSON diff.
- [x] Add initial CSS-aware normalization hooks if low risk; otherwise document as follow-up.
- [x] Expose `cvd.diff(before, after, options)`.
- [x] Expose `cvd.report(diff)` with Markdown rendering.
- [x] Expose `cvd.write.json(path, value)`.
- [x] Expose `cvd.write.markdown(path, markdown)` or equivalent.
- [x] Add tests for equal/changed/ignored/tolerance cases.
- [x] Update diary/changelog and commit Phase 9.

## Phase 10 — Public docs, examples, smoke scripts, and delivery

Goal: make the new lower-level API usable and reviewable by operators and future implementers.

- [x] Add embedded Glazed help coverage for the lower-level API instead of a standalone `docs/js-low-level-api.md`.
- [x] Move `docs/js-api.md` into embedded help as `internal/cssvisualdiff/doc/topics/javascript-api.md` and update it with locator/extract/snapshot/diff APIs.
- [x] Add `examples/verbs/low-level-inspect.js`.
- [x] Add ticket smoke scripts under `scripts/` with numeric prefixes.
- [x] Add and run a compiled-binary smoke for the external lower-level verb.
- [x] Run `go test ./... -count=1`.
- [x] Run current Phase 10 binary smokes for embedded help and lower-level external verb behavior.
- [x] Run `docmgr doctor --root ./ttmp --ticket CSSVD-FLEX-JS-API --stale-after 30`.
- [x] Regenerate and upload the updated implementation/user-guide PDF to reMarkable as `/ai/2026/04/24/cssvd-flex-api-implementation`.
- [x] Update diary/changelog and commit final docs.
