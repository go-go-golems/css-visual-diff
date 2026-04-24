# Tasks

## Phase 0 — Research, planning, and ticket hygiene

- [x] Create ticket workspace for the Goja JavaScript API design.
- [x] Write intern-facing analysis/design/implementation guide.
- [x] Upload design guide to reMarkable.
- [x] Research current css-visual-diff `internal/cssvisualdiff/dsl` implementation and nearby jsverbs systems.
- [x] Update implementation guide with repository-scanned CLI verb architecture.
- [x] Add implementation research diary.
- [x] Incorporate maintainer clarifications: Promise-first API, Go-side catalog service, typed errors, preflight/directReactGlobal/parallelization answers.
- [ ] Review updated repository-scanned verbs plan with css-visual-diff maintainer.

## Phase 1 — Stabilize the existing dsl/jsverbs prototype

Goal: make the current embedded-script prototype clean and safe to build on before changing command UX.

- [x] Remove ignored generated `internal/cssvisualdiff/dsl/css-visual-diff-compare-*` artifact directories from the working tree.
- [x] Confirm `.gitignore` continues to ignore generated `css-visual-diff-compare-*` output directories.
- [x] Ensure dsl tests use temp output directories and do not write under source packages.
- [x] Run `go test ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff`.
- [x] Commit Phase 1 cleanup separately from later CLI restructuring.

## Phase 2 — Add lazy repository-scanned `css-visual-diff verbs` command tree

Goal: make JS verbs a first-class CLI namespace and stop injecting generated script commands at the root.

- [x] Add `internal/cssvisualdiff/verbcli/bootstrap.go` with `Bootstrap`, `Repository`, embedded repository, path normalization, env/CLI repository discovery, and duplicate repository dedupe.
- [x] Add app-config repository discovery.
- [x] Add `jsverbs.ScanFS`/`ScanDir` repository scanning with `IncludePublicFunctions=false`.
- [x] Add duplicate full verb path detection with useful source paths.
- [x] Add `internal/cssvisualdiff/verbcli/command.go` with lazy Cobra command registration modeled after loupedeck.
- [x] Add `internal/cssvisualdiff/verbcli/invoker.go` or equivalent custom invoker that creates a css-visual-diff-owned runtime per invocation.
- [x] Preserve embedded built-in verbs under `css-visual-diff verbs ...`.
- [x] Remove eager `dsl.NewHost().Commands()` root-level injection from `cmd/css-visual-diff/main.go`.
- [x] Add tests for built-in verbs help, filesystem repository scanning, duplicate path errors, and generated command execution.
- [x] Run `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- [x] Commit initial Phase 2 lazy verbs CLI.

## Phase 3 — Extract reusable browser/page/inspect/preflight services

Goal: make browser operations callable from both existing CLI modes and JS modules without duplicating chromedp/artifact logic.

- [ ] Add service types for `BrowserService`, `PageService`, `ProbeSpec`, `SelectorStatus`, `InspectAllOptions`, and `InspectAllResult`.
- [ ] Extract prepare logic so `script` and `directReactGlobal` can be called against an already-created page.
- [x] Extract style evaluation service shared by CSS diff and inspect paths.
- [x] Extract batched preflight service that checks selectors in one page evaluation where possible.
- [ ] Extract `InspectPreparedPage` / `InspectAll` service from `modes.Inspect` artifact writers.
- [x] Keep existing selector-preflight behavior unchanged by routing inspect selector checks through the extracted preflight service.
- [x] Add tests for missing selectors and hidden/zero-bounds selectors in preflight service.
- [x] Run `go test ./internal/cssvisualdiff/modes ./internal/cssvisualdiff/service ./cmd/css-visual-diff`.
- [x] Commit Phase 3 initial preflight/style service extraction.

## Phase 4 — Implement Promise-first `require("css-visual-diff")` native module

Goal: expose the Go services to repository-scanned scripts through a stable async JS API.

- [ ] Add native module registration for `require("css-visual-diff")`.
- [ ] Implement Promise-returning `cvd.browser(options?)`.
- [ ] Implement Promise-returning `browser.page(url, options?)`, `browser.newPage(options?)`, and `browser.close()`.
- [ ] Implement Promise-returning `page.goto`, `page.prepare`, `page.preflight`, `page.inspect`, `page.inspectAll`, and `page.close`.
- [ ] Implement lowerCamel JS option/result codecs with validation errors.
- [ ] Implement JS-visible error classes/codes: `CvdError`, `SelectorError`, `PrepareError`, `BrowserError`, `ArtifactError`.
- [ ] Ensure native promises settle through the go-go-goja runtime owner thread.
- [ ] Add runtime integration tests for `require("css-visual-diff")` and one async verb using a tiny HTTP page.
- [ ] Run `go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff`.
- [ ] Commit Phase 4 Promise-first JS module.

## Phase 5 — Implement Go-side catalog service and JS adapter

Goal: keep catalog manifests and indexes durable, typed, and usable from both JS and Go CLI code.

- [ ] Add `internal/cssvisualdiff/service/catalog_service.go` with manifest schema/version, target records, preflight records, result records, failure records, summary, slug/path normalization, and writers.
- [ ] Add `cvd.catalog(options)` adapter backed by the Go catalog service.
- [ ] Implement `catalog.artifactDir`, `addTarget`, `recordPreflight`, `addResult`, `addFailure`, `summary`, `writeManifest`, and `writeIndex`.
- [ ] Add tests for manifest JSON, markdown/HTML index rendering if implemented, path normalization, and failures.
- [ ] Add one JS integration test that writes a catalog manifest and index from a verb.
- [ ] Commit Phase 5 catalog service.

## Phase 6 — Add built-in catalog verbs and examples

Goal: prove the full workflow as operator-facing commands with flags.

- [ ] Add built-in `catalog inspect-page` verb.
- [ ] Add built-in `catalog inspect-config` or YAML interop verb.
- [ ] Add built-in `compare region` / `compare brief` verbs under the new `verbs` namespace, replacing root-level generated command usage.
- [ ] Add external example repository under `examples/verbs/` or documented fixture demonstrating repository scanning.
- [ ] Add command examples for authoring mode (`failOnMissing=false`) and CI mode (`failOnMissing=true`).
- [ ] Run end-to-end smoke test against a local HTTP fixture.
- [ ] Commit Phase 6 built-in verbs/examples.

## Phase 7 — Documentation, migration notes, and final validation

Goal: make the feature understandable and safe to use.

- [ ] Add `docs/js-api.md` for `require("css-visual-diff")`.
- [ ] Add `docs/js-verbs.md` for `__verb__`, repositories, generated flags, output modes, and duplicate command paths.
- [ ] Update README with `css-visual-diff verbs ...` examples and migration note for prior root-level generated script commands.
- [ ] Document preflight, `directReactGlobal`, Promise behavior, error types, and target-level concurrency guidance.
- [ ] Run `go test ./...`.
- [ ] Run relevant CLI smoke tests.
- [ ] Update ticket changelog and diary.
- [ ] Commit final docs/validation.
