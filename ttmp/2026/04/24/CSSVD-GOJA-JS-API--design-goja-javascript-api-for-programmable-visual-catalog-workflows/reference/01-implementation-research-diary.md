---
Title: Implementation research diary
Ticket: CSSVD-GOJA-JS-API
Status: active
Topics:
    - css-visual-diff
    - goja
    - javascript-api
    - jsverbs
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../design/01-goja-javascript-api-analysis-design-and-implementation-guide.md
      Note: Main implementation guide updated by this research pass.
    - Path: ../../../../../../../../internal/cssvisualdiff/dsl/host.go
      Note: Current embedded jsverbs host in css-visual-diff.
    - Path: ../../../../../../../../internal/cssvisualdiff/dsl/registrar.go
      Note: Current native Goja modules exposed to css-visual-diff scripts.
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/23/DISCORD-BOT-023--discord-helper-verbs-and-jsverbs-live-debugging-cli/reference/01-investigation-diary.md
      Note: Reference diary for the advanced Discord helper-verbs design.
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go
      Note: Implemented loupedeck repository-scanning pattern.
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go
      Note: Implemented loupedeck lazy verbs command and custom invoker pattern.
ExternalSources: []
Summary: Diary for the research pass that updated the css-visual-diff Goja/jsverbs design after reading nearby implementations and current repository code.
LastUpdated: 2026-04-24T00:00:00-04:00
WhatFor: Record how the implementation guide was reassessed and what evidence drove the updated recommendations.
WhenToUse: Read before continuing implementation of repository-scanned css-visual-diff JavaScript verbs or the public Goja API.
---

# Implementation research diary

## Goal

Reassess the existing Goja JavaScript API implementation guide for `css-visual-diff` after studying recent jsverbs work in the Discord bot repository, the implemented loupedeck jsverbs CLI cutover, upstream `go-go-goja`, and the current `css-visual-diff` code.

The key question was not only “what should the JS API look like?” but also “how should JavaScript files be scanned from repositories and exposed as CLI verbs with flags for visual comparison workflows?”

## What I read

### Current ticket and css-visual-diff code

- `ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/index.md`
- `tasks.md`
- `changelog.md`
- `design/01-goja-javascript-api-analysis-design-and-implementation-guide.md`
- `cmd/css-visual-diff/main.go`
- `internal/cssvisualdiff/dsl/host.go`
- `internal/cssvisualdiff/dsl/registrar.go`
- `internal/cssvisualdiff/dsl/codec.go`
- `internal/cssvisualdiff/dsl/sections.go`
- `internal/cssvisualdiff/dsl/scripts/compare.js`
- `internal/cssvisualdiff/modes/inspect.go`
- `internal/cssvisualdiff/modes/prepare.go`
- `internal/cssvisualdiff/driver/chrome.go`

### Related jsverbs work

- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/23/DISCORD-BOT-023--discord-helper-verbs-and-jsverbs-live-debugging-cli/reference/01-investigation-diary.md`
- `/home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/23/DISCORD-BOT-023--discord-helper-verbs-and-jsverbs-live-debugging-cli/design-doc/01-discord-helper-verbs-and-jsverbs-live-debugging-cli-design-and-implementation-guide.md`
- `/home/manuel/code/wesen/corporate-headquarters/loupedeck/ttmp/2026/04/18/LOUPE-JSVERBS-CLI--embed-jsverbs-as-first-class-loupedeck-cli-commands-and-tighten-js-scene-docs/reference/02-implementation-diary-for-jsverbs-cli-cutover.md`
- `/home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go`
- `/home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/command.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/model.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/doc/11-jsverbs-example-reference.md`

## Findings

### 1. css-visual-diff already has a Goja/jsverbs prototype

The original design guide reads partly like a greenfield plan, but the repository already has `internal/cssvisualdiff/dsl`:

- `NewHost()` scans embedded scripts with `jsverbs.ScanFS(embeddedScripts, "scripts")`.
- Shared sections are registered for targets, viewport, and output.
- `engine.NewBuilder()` wires go-go-goja runtime modules.
- `Commands()` exposes generated Glazed commands through `registry.CommandsWithInvoker(...)`.
- `cmd/css-visual-diff/main.go` eagerly adds those generated commands to the root command.

The guide needed to be updated from “add Goja” to “productize the existing Goja/jsverbs host”.

### 2. The nearby loupedeck implementation is the best concrete pattern

Loupedeck already implemented the exact shape needed here:

- lazy `verbs` subtree,
- embedded built-in repository,
- config/env/CLI repository discovery,
- `jsverbs.ScanFS` and `jsverbs.ScanDir`,
- duplicate verb path detection,
- generated Cobra commands backed by a custom invoker,
- runtime ownership kept in the host application rather than inside generic jsverbs.

The css-visual-diff guide should now explicitly recommend porting that product pattern instead of inventing a separate scanner/runner.

### 3. Discord bot design confirms the same repository/verb split

The Discord helper-verbs guide reached the same architectural conclusion from a different domain: keep the domain runtime separate from repository-scanned helper verbs, use jsverbs for command metadata and flag generation, and put dynamic commands under a clear `verbs` namespace.

For css-visual-diff, this means:

```text
YAML configs remain stable declarative plans.
Built-in root commands remain normal CLI workflows.
Repository-scanned JS verbs become dynamic workflow commands under `css-visual-diff verbs`.
The JS native module provides browser/page/catalog services to those verbs.
```

### 4. Promise support is no longer unknown

The original guide suggested starting synchronously if Promise support was not wired. Upstream go-go-goja docs now state that `registry.InvokeInRuntime(...)` waits for returned Promises from jsverb functions. After maintainer clarification, the guide now requires Promise-returning browser/page/catalog APIs from the first implementation rather than permitting a synchronous MVP.

### 5. The current prototype has source-tree artifact hygiene problems

A scan of `internal/cssvisualdiff/dsl` found many generated `css-visual-diff-compare-*` PNG artifact directories beside source files. The updated design now calls out cleanup: tests should use `t.TempDir()`, runtime defaults should write outside source packages, and generated artifact directories should be ignored or removed if not intentional fixtures.

## What I changed

Updated `design/01-goja-javascript-api-analysis-design-and-implementation-guide.md` by adding:

- new related-file references for the current css-visual-diff dsl implementation, loupedeck verbs code, and upstream go-go-goja jsverbs files,
- a research update explaining what already exists and what must change,
- a two-surface model: `require("css-visual-diff")` native API plus repository-scanned `__verb__` CLI scripts,
- a loupedeck-style repository scanning plan,
- a lazy `css-visual-diff verbs` subtree recommendation,
- a revised package layout with `dsl`, `verbcli`, and service packages,
- an example annotated catalog verb that exposes flags,
- expanded service extraction guidance,
- a list of assumptions to remove/de-emphasize,
- a hygiene section about generated artifacts under `internal/cssvisualdiff/dsl`,
- an updated phase plan focused on productizing the existing dsl prototype.

## What remains to do

- Convert the design update into implementation work:
  1. clean up generated artifacts under source packages,
  2. add `internal/cssvisualdiff/verbcli`,
  3. move dynamic commands under `css-visual-diff verbs`,
  4. extract inspect/preflight services,
  5. add `require("css-visual-diff")`,
  6. add real catalog verbs.
- Decide the exact external repository flag name: `--repository` vs `--verb-repository`.
- Decide whether embedded built-ins should remain under `internal/cssvisualdiff/dsl/scripts` or move to an `examples` embedded package.
- Validate current tests before code changes.

## Commands run during research

```bash
find ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows -maxdepth 3 -type f | sort
docmgr status --summary-only
docmgr ticket list
find /home/manuel/code/wesen/2026-04-20--js-discord-bot -iname '*diary*' -o -path '*/ttmp/*' -type f | head -200
find /home/manuel/code/wesen/corporate-headquarters/loupedeck -iname '*diary*' -o -path '*/ttmp/*' -type f | head -200
rg -n "jsverbs|verb|JSDoc|glazed|command" /home/manuel/code/wesen/2026-04-20--js-discord-bot --glob '!node_modules/**' --glob '!ttmp/**'
find cmd internal pkg -type f -name '*.go' | sort
rg -n "script compare|__verb__|dsl|goja|verbs|css-visual-diff script" README.md cmd internal pkg ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows -g '!**/*.png'
git status --short
find internal/cssvisualdiff/dsl -maxdepth 2 -type f | sort
```

## Follow-up: maintainer clarifications

The maintainer clarified four important design points after the first research update:

1. Promises are required from the beginning.
2. The catalog API should be implemented on the Go side, not just as JS helper code.
3. The error model can throw; define useful error types/classes.
4. The document should answer what preflight is for, what `directReactGlobal` does compared with selectors, and what can realistically be parallelized given CDP serialization.

I updated the main design guide accordingly:

- replaced the old synchronous-MVP fallback with a Promise-first requirement,
- expanded the catalog section to require `internal/cssvisualdiff/service/catalog_service.go` or equivalent,
- expanded the error section with `CvdError`, `SelectorError`, `PrepareError`, `BrowserError`, and `ArtifactError`,
- added a maintainer-clarifications section explaining:
  - preflight = selector/probe readiness validation before expensive extraction,
  - `directReactGlobal` = prepare mode that mounts a global React component into a controlled capture root,
  - parallelization should be coarse target/page-level or post-processing-level, not naive concurrent calls inside one CDP page session.

## Issues encountered

- `docmgr` in this working directory resolves to the parent workspace config rooted at `hair-booking/ttmp`, not the local `css-visual-diff/ttmp`. Because the user explicitly gave the local ticket path, I updated the local ticket files directly instead of using `docmgr` commands.
- A broad `find ... -exec sed ...` command printed binary PNG output from generated artifact directories under `internal/cssvisualdiff/dsl`; this accidentally confirmed the artifact hygiene issue but produced noisy terminal output.

## Implementation Step 1: stabilize dsl tests and remove generated source-tree artifacts

After creating the phased task plan, I started with Phase 1 because it is the safest prerequisite: the existing dsl tests were writing generated compare PNG directories into `internal/cssvisualdiff/dsl` whenever no `outDir` was supplied. Those directories are ignored by `.gitignore`, but they still polluted the working tree and made file inspection noisy.

### What I changed

- Updated `internal/cssvisualdiff/dsl/host_test.go` so both embedded compare command invocations pass explicit temporary output directories using `t.TempDir()`.
- Removed ignored `internal/cssvisualdiff/dsl/css-visual-diff-compare-*` directories from the working tree.
- Confirmed `.gitignore` ignores generated `css-visual-diff-compare-*` directories.
- Marked the relevant Phase 1 tasks complete in `tasks.md`.
- Added a changelog entry for the cleanup.

### Validation

```bash
gofmt -w internal/cssvisualdiff/dsl/host_test.go
go test ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff
```

Both package tests passed.

## Implementation Step 2: add initial lazy `verbs` CLI

I started Phase 2 by turning the current embedded script prototype into a lazy `css-visual-diff verbs` subtree. The goal of this step was to stop registering generated script commands at the root and create a product namespace that can later scan external repositories.

### What I changed

- Added `internal/cssvisualdiff/verbcli/bootstrap.go`:
  - built-in embedded repository from `internal/cssvisualdiff/dsl/scripts`,
  - environment repository discovery via `CSS_VISUAL_DIFF_VERB_REPOSITORIES`,
  - CLI repository discovery via `--repository` and `--verb-repository`,
  - path normalization and dedupe,
  - `jsverbs.ScanFS` / `ScanDir`,
  - `IncludePublicFunctions=false`,
  - duplicate full verb path detection.
- Added `internal/cssvisualdiff/verbcli/command.go`:
  - lazy Cobra `verbs` command,
  - generated Glazed commands from jsverbs metadata,
  - runtime invoker per command.
- Added `internal/cssvisualdiff/verbcli/runtime_factory.go`.
- Exported dsl helpers in `internal/cssvisualdiff/dsl/host.go` for embedded scripts, shared section registration, and runtime factory construction.
- Updated `cmd/css-visual-diff/main.go` to remove eager `dsl.NewHost().Commands()` root injection and add `verbcli.NewLazyCommand()`.
- Added `internal/cssvisualdiff/verbcli/command_test.go` for built-in command registration, duplicate detection, and filesystem verb execution.

### Validation

```bash
gofmt -w internal/cssvisualdiff/dsl/host.go internal/cssvisualdiff/verbcli/*.go cmd/css-visual-diff/main.go
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go run ./cmd/css-visual-diff --help
go run ./cmd/css-visual-diff verbs --help
go run ./cmd/css-visual-diff verbs script compare region --help
```

The tests passed and the help output now shows root-level `verbs` plus generated flags under `verbs script compare region`.

### Follow-up

App-config repository discovery remained outstanding immediately after this step. The initial implementation covered embedded, environment, and CLI repositories, which was enough to establish the lazy command shape and filesystem scanning path.

## Implementation Step 3: add app-config repository discovery

I completed the remaining repository-discovery gap in Phase 2 by adding app-config support. The implementation mirrors the loupedeck approach but uses the `css-visual-diff` app name.

### What I changed

- Added app-config structs to `internal/cssvisualdiff/verbcli/bootstrap.go`:
  - `appConfig`,
  - `verbsConfig`,
  - `repositorySpec`.
- Added `loadConfigRepositories(...)` using Glazed config plans for:
  - system app config,
  - XDG app config,
  - home app config.
- Added `loadRepositoriesFromConfigFile(...)` to parse:

```yaml
verbs:
  repositories:
    - name: local
      path: ./verbs
    - name: disabled
      path: ./disabled
      enabled: false
```

- Config repositories are loaded after the embedded built-in repository and before environment/CLI repositories.
- Added a unit test for relative config paths and disabled repositories.

### Validation

```bash
gofmt -w internal/cssvisualdiff/verbcli/bootstrap.go internal/cssvisualdiff/verbcli/command_test.go
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
```

Tests passed.

## Implementation Step 4: extract initial preflight/style services

I started Phase 3 with the smallest service extraction that is directly useful for the JS API: style evaluation and selector preflight. This avoids touching the whole inspect artifact writer in one large change while still creating Go-side service functions that the future Promise-first JS module can call.

### What I changed

- Added `internal/cssvisualdiff/service/types.go` with:
  - `Bounds`,
  - `StyleSnapshot`,
  - `ProbeSpec`,
  - `SelectorStatus`.
- Added `internal/cssvisualdiff/service/style.go` with `EvaluateStyle(...)`.
- Added `internal/cssvisualdiff/service/preflight.go` with `PreflightProbes(...)`.
  - It batches all probes into one page-side `document.querySelector` pass.
  - It reports existence, visibility, bounds, text prefix, and selector errors.
- Updated `internal/cssvisualdiff/modes/cssdiff.go` to alias `StyleSnapshot` / `Bounds` from the service package and route `evaluateStyle(...)` through `service.EvaluateStyle(...)`.
- Updated `internal/cssvisualdiff/modes/inspect.go` so `ensureInspectSelectorExists(...)` uses `service.PreflightProbes(...)`.
- Added `internal/cssvisualdiff/service/preflight_test.go` covering existing, missing, invalid, and hidden selectors.

### Validation

```bash
gofmt -w internal/cssvisualdiff/service internal/cssvisualdiff/modes/cssdiff.go internal/cssvisualdiff/modes/inspect.go
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Follow-up

This is only the initial Phase 3 extraction. Remaining Phase 3 work:

- extract prepare service,
- extract prepared-page inspect/artifact service,
- add no-reload-per-probe tests,
- keep CLI behavior unchanged while making service APIs available to JS adapters.

## Implementation Step 5: extract prepare service

I continued Phase 3 by moving prepare behavior into the service package. This makes `script` and `directReactGlobal` prepare available to the future JS page adapter without requiring it to import the CLI-oriented `modes` package.

### What I changed

- Added `internal/cssvisualdiff/service/prepare.go` with:
  - `PrepareTarget`,
  - `RunScriptPrepare`,
  - `RunDirectReactGlobalPrepare`,
  - `BuildDirectReactGlobalScript`,
  - `DirectReactGlobalPrepareResult`.
- Replaced `internal/cssvisualdiff/modes/prepare.go` with thin compatibility wrappers so existing modes and tests keep their old function names.

### Validation

```bash
gofmt -w internal/cssvisualdiff/service/prepare.go internal/cssvisualdiff/modes/prepare.go
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Follow-up

The next remaining Phase 3 item is extracting prepared-page inspect/artifact writing into a service API.

## Implementation Step 6: add browser/page service shell

I added a small browser/page service shell as another incremental Phase 3 step. The goal was to create the Go-side shape that the future JS `browser()` / `page()` adapter can wrap, without yet extracting the entire inspect artifact pipeline.

### What I changed

- Added `internal/cssvisualdiff/service/browser.go` with:
  - `BrowserService`,
  - `PageService`,
  - `NewBrowserService`,
  - `LoadAndPreparePage`.
- Updated `modes.Inspect(...)` to use `service.LoadAndPreparePage(...)` for viewport, navigation, wait, and prepare.

### Validation

```bash
gofmt -w internal/cssvisualdiff/service/browser.go internal/cssvisualdiff/modes/inspect.go
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Follow-up

The browser/page service is intentionally minimal. The next substantial extraction is still `InspectPreparedPage` / `InspectAll`, which will move artifact writing and inspect batching out of `modes.Inspect(...)`.

## Implementation Step 7: extract prepared-page inspect/artifact service

I completed the planned Phase 3 service extraction by moving inspect artifact writing and prepared-page batch inspection into the service package. This is the key seam needed by the future JavaScript `page.inspectAll(...)` adapter: it can operate on an already-loaded, already-prepared page without invoking the config-driven CLI mode.

### What I changed

- Added `internal/cssvisualdiff/service/inspect.go` with:
  - inspect format constants,
  - `InspectRequest`,
  - `InspectMetadata`,
  - `InspectArtifactResult`,
  - `InspectAllOptions`,
  - `InspectResult`,
  - `InspectPreparedPage`,
  - `WriteInspectArtifacts`,
  - `WriteSingleInspectArtifact`,
  - `WriteInspectIndex`,
  - shared `WritePreparedHTML`, `WriteInspectJSON`, `WriteJSON`, `RootSelectorForTarget`, and `SanitizeName` helpers.
- Updated `modes.Inspect(...)` to call `service.InspectPreparedPage(...)` after loading and preparing the page.
- Converted mode-level inspect types/constants to service aliases for compatibility.
- Added `internal/cssvisualdiff/service/inspect_test.go` with a no-reload-per-probe test. The test loads a target once, inspects two CSS probes, and asserts the HTTP root path was only requested once.

### Validation

```bash
gofmt -w internal/cssvisualdiff/service/inspect.go internal/cssvisualdiff/service/inspect_test.go internal/cssvisualdiff/modes/inspect.go
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Notes

Some older private helpers still exist in `modes` for compatibility with other mode code paths. The important architectural boundary is now in place: prepared-page inspect and artifact extraction can be called from `service` without going through `modes.Inspect(...)`.

## Implementation Step 8: add Promise-first `require("css-visual-diff")` MVP

I started Phase 4 by adding the first public native module shape. The goal was not to finish the full polished JS API, but to prove the important contract: repository-scanned async jsverbs can `require("css-visual-diff")`, await browser/page operations, preflight selectors, and write inspect artifacts through the Go service layer.

### What I changed

- Added `internal/cssvisualdiff/dsl/cvd_module.go`.
- Registered `require("css-visual-diff")` from the existing runtime registrar.
- Added a `promiseValue(...)` helper that:
  - creates a Goja Promise on the VM goroutine,
  - runs Go work in a goroutine,
  - resolves/rejects through `ctx.Owner.Post(...)` on the runtime owner thread.
- Added Promise-returning MVP methods:
  - `cvd.browser()`
  - `browser.newPage()`
  - `browser.page(url, options)`
  - `browser.close()`
  - `page.prepare(spec)`
  - `page.preflight(probes)`
  - `page.inspectAll(probes, options)`
  - `page.close()`
- Added a repository-scanned async verb integration test in `internal/cssvisualdiff/verbcli/command_test.go`.

### Validation

```bash
gofmt -w internal/cssvisualdiff/dsl/cvd_module.go internal/cssvisualdiff/dsl/registrar.go internal/cssvisualdiff/verbcli/command_test.go
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Caveats / follow-up

- This is an MVP. Some JS-visible result objects still expose Go exported field names (`Results`, `OutputDir`, `Exists`) rather than lowerCamel fields. The lowerCamel result codec task remains open.
- `page.goto` and single `page.inspect` are not implemented yet.
- JS-visible typed error classes are not implemented yet; current errors reject promises with Go errors.

## Implementation Step 9: complete Phase 4 JavaScript API surface

I completed the remaining Phase 4 API work on top of the Promise-first module MVP.

### What I changed

- Added `page.goto(url, options)`.
- Added single-probe `page.inspect(probe, options)`.
- Changed `page.preflight(...)` and `page.inspectAll(...)` JS-facing returns from Go exported field names to lowerCamel objects:
  - `Exists` became `exists`,
  - `TextStart` became `textStart`,
  - `OutputDir` became `outputDir`,
  - `Results` became `results`,
  - `InspectJSON` became `inspectJson`,
  - metadata fields such as `target_name`/`selector_source` became `targetName`/`selectorSource`.
- Added `CvdError`, `SelectorError`, `PrepareError`, `BrowserError`, and `ArtifactError` constructors to the module exports.
- Updated Promise rejection handling to construct classified JS errors on the runtime owner thread.
- Added basic validation for missing URLs and empty/selectorless inspect calls.
- Extended the verb integration tests to prove:
  - lowerCamel results are visible to JavaScript,
  - `page.goto(...)` works from a `browser.newPage()` flow,
  - missing selector inspect failures are caught as `SelectorError`,
  - typed errors inherit from `CvdError`.

### Validation

```bash
gofmt -w internal/cssvisualdiff/dsl/cvd_module.go internal/cssvisualdiff/verbcli/command_test.go
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Notes

The Phase 4 module is now complete enough for repository-scanned scripts to drive browser/page/preflight/inspect workflows with Promise-first semantics and JS-idiomatic result shapes. The next planned phase is the Go-side catalog service and a `cvd.catalog(...)` adapter.

## Implementation Step 10: preserve replay scripts for validations and smoke tests

After the Phase 4 binary-smoke discussion, I added ticket-local replay scripts under `scripts/` with numeric prefixes. These capture both the historical validation commands and the newer compiled-binary smoke checks so future reviewers can retrace what was run.

### Scripts added

- `001-phase1-dsl-temp-artifacts-test.sh`
- `002-phase2-lazy-verbs-cli-tests.sh`
- `003-phase3-service-extraction-tests.sh`
- `004-phase4-promise-module-tests.sh`
- `005-full-test-suite.sh`
- `006-binary-help-smoke.sh`
- `007-binary-js-api-success-smoke.sh`
- `008-binary-js-api-typed-error-smoke.sh`

### Validation

I re-ran the three compiled-binary scripts immediately after writing them:

```bash
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/006-binary-help-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/007-binary-js-api-success-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/008-binary-js-api-typed-error-smoke.sh
```

All three passed.

## Implementation Step 11: implement Go-side catalog service and `cvd.catalog(...)`

I completed Phase 5 by adding the catalog implementation on the Go side and exposing it through the Promise-first JS module. This keeps durable catalog behavior—manifest schema, path normalization, summary counts, and report writers—in Go rather than in ad hoc JavaScript helpers.

### What I changed

- Added `internal/cssvisualdiff/service/catalog_service.go` with:
  - `CatalogSchemaVersion`,
  - `CatalogOptions`,
  - `CatalogManifest`,
  - `CatalogTargetRecord`,
  - `CatalogPreflightRecord`,
  - `CatalogResultRecord`,
  - `CatalogFailureRecord`,
  - `CatalogSummary`,
  - `NewCatalog`,
  - `ArtifactDir`,
  - `AddTarget`,
  - `RecordPreflight`,
  - `AddResult`,
  - `AddFailure`,
  - `Summary`,
  - `WriteManifest`,
  - `WriteIndex`,
  - slug and relative-path normalization helpers.
- Added `internal/cssvisualdiff/dsl/catalog_adapter.go` with the JS adapter for `cvd.catalog(options)`.
- Added JS-facing catalog methods:
  - `artifactDir(slug)`,
  - `addTarget(target)`,
  - `recordPreflight(target, statuses)`,
  - `addResult(target, inspectResult)`,
  - `addFailure(target, error)`,
  - `summary()`,
  - `manifest()`,
  - `writeManifest()`,
  - `writeIndex()`.
- Added service tests and a repository-scanned JS verb integration test.
- Added `scripts/009-binary-catalog-smoke.sh` to replay a compiled-binary catalog smoke test.

### Validation

```bash
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/009-binary-catalog-smoke.sh
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

All tests passed.

### Notes

The initial catalog index is Markdown-only. The service data model is versioned and should be stable enough for built-in catalog verbs in the next phase. If we later add richer HTML reports, they should be additional writers using the same `CatalogManifest` model.

## Implementation Step 12: add built-in `catalog inspect-page` and example repository

I started Phase 6 by turning the lower-level JS API and catalog service into an operator-facing built-in command.

### What I changed

- Added `internal/cssvisualdiff/dsl/scripts/catalog.js`.
- The embedded built-in command is available as:

```bash
css-visual-diff verbs catalog inspect-page <url> <selector> <outDir> [flags]
```

- The implementation uses:
  - `cvd.catalog(...)`,
  - `cvd.browser()`,
  - `browser.page(...)`,
  - `page.preflight(...)`,
  - `page.inspect(...)`,
  - `catalog.recordPreflight(...)`,
  - `catalog.addResult(...)`,
  - `catalog.addFailure(...)`,
  - `catalog.writeManifest()`,
  - `catalog.writeIndex()`.
- Added authoring-mode selector-miss handling: by default, missing selectors are recorded as catalog failures and returned as structured rows without failing the command.
- Added CI-mode selector-miss handling: with `--failOnMissing`, the command writes manifest/index and exits non-zero.
- Added `examples/verbs/catalog-inspect-page.js` and `examples/verbs/README.md` as an external repository example with command examples.
- Added `scripts/010-binary-built-in-catalog-inspect-page-smoke.sh`.

### Validation

```bash
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/010-binary-built-in-catalog-inspect-page-smoke.sh
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

The binary smoke covered three cases:

1. successful `#cta` inspection that writes `computed-css.json`, `manifest.json`, and `index.md`,
2. authoring-mode missing selector (`failOnMissing=false`) that returns `ok=false` but exits zero,
3. CI-mode missing selector (`--failOnMissing`) that writes manifest/index and exits non-zero.

### Notes

`catalog inspect-config` / YAML interop remains the next Phase 6 item. The built-in `catalog inspect-page` command proves the core operator-facing catalog workflow independently of YAML.

## Implementation Step 13: add YAML config interop with `catalog inspect-config`

I completed the remaining Phase 6 item by adding YAML config interop for catalog workflows.

### What I changed

- Added `internal/cssvisualdiff/dsl/config_adapter.go`.
- Exposed `cvd.loadConfig(path)` from `require("css-visual-diff")` as a Promise-returning native helper backed by Go `config.Load`.
- Converted loaded configs to lowerCamel JavaScript objects:
  - `wait_ms` -> `waitMs`,
  - `root_selector` -> `rootSelector`,
  - `selector_original` -> `selectorOriginal`,
  - `selector_react` -> `selectorReact`,
  - prepare timing and script fields to lowerCamel.
- Added built-in command:

```bash
css-visual-diff verbs catalog inspect-config <configPath> <side> <outDir> [flags]
```

- `inspect-config` derives probes from `styles` first and falls back to `sections` when no styles exist.
- It records preflight statuses, writes inspect artifacts for selectors that exist, records failures for missing selectors, and writes the Go-side catalog manifest/index.
- Added `scripts/011-binary-built-in-catalog-inspect-config-smoke.sh`.

### Validation

```bash
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/011-binary-built-in-catalog-inspect-config-smoke.sh
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

All tests passed.

## Implementation Step 14: write final JS API / jsverbs docs and run final validation

I completed the documentation phase by adding user-facing docs for both layers of the new system.

### What I changed

- Added `docs/js-api.md` for the native module:
  - `cvd.browser`,
  - browser/page lifecycle,
  - `page.goto`, `page.prepare`, `page.preflight`, `page.inspect`, `page.inspectAll`,
  - artifact formats,
  - `cvd.catalog`,
  - `cvd.loadConfig`,
  - typed errors,
  - target/page-level concurrency guidance.
- Added `docs/js-verbs.md` for repository-scanned commands:
  - `css-visual-diff verbs ...`,
  - repository discovery sources,
  - `__package__`, `__section__`, `__verb__`,
  - generated fields/flags,
  - binding modes,
  - output modes,
  - built-in catalog command examples,
  - duplicate verb path errors,
  - migration from earlier root-level generated command injection.
- Updated `README.md` with a concise JavaScript verbs / programmable catalogs section and links to both docs.

### Final validation plan

I ran the final validation set after the docs update:

```bash
go test ./...
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/006-binary-help-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/007-binary-js-api-success-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/008-binary-js-api-typed-error-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/009-binary-catalog-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/010-binary-built-in-catalog-inspect-page-smoke.sh
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/011-binary-built-in-catalog-inspect-config-smoke.sh
```

## Implementation Step 15: store code review assessment and attempt reMarkable upload

I stored the senior-style “big brother” code review as a ticket document:

```text
review/01-big-brother-code-review-and-assessment.md
```

I also generated a local PDF copy under:

```text
review/pdf/01-big-brother-code-review-and-assessment.pdf
```

### Upload attempts

The dry run succeeded:

```bash
remarquee upload md --dry-run review/01-big-brother-code-review-and-assessment.md \
  --remote-dir /ai/2026/04/24/CSSVD-GOJA-JS-API \
  --name "CSSVD Goja JS API - Big Brother Code Review"
```

Actual upload failed with reMarkable/rmapi status 400. I tried:

```bash
remarquee upload md ...
remarquee cloud put ...
rmapi put ...
python3 /home/manuel/.local/bin/remarkable_upload.py ...
```

All upload/create attempts failed with variants of:

```text
failed to upload file [...pdf]: request failed with status 400
```

`remarquee status`, `remarquee cloud account --non-interactive`, and `remarquee cloud ls /ai/2026/04/24/CSSVD-GOJA-JS-API --long --non-interactive` succeeded, so read/auth status appears available, but cloud mutation/upload is currently failing.

## Implementation Step 16: fix GitHub Actions Chrome sandbox failure

The PR CI failed in the `run unit tests` job with Chromium aborting before tests could run browser-backed flows:

```text
FATAL:content/browser/zygote_host/zygote_host_impl_linux.cc:128] No usable sandbox!
If you want to live dangerously and need an immediate workaround, you can try using --no-sandbox.
```

This affected browser-backed tests such as:

- `TestPreflightProbesReportsExistingMissingAndInvalid`,
- `TestRepositoryVerbUsesPromiseFirstCVDModule`,
- `TestCVDModuleExposesLowerCamelGotoInspectAndTypedErrors`.

### Root cause

GitHub Actions `ubuntu-latest` can run Chrome/Chromium in an environment where the normal Linux sandbox is unavailable. Our `driver.NewBrowser` allocator used headless Chrome but did not pass `--no-sandbox`.

### Fix

I updated `internal/cssvisualdiff/driver/chrome.go` so allocator options are built by a helper:

- default options remain headless / no-first-run / no-default-browser-check,
- `chromedp.NoSandbox` is added when:
  - `CI=true`,
  - `GITHUB_ACTIONS=true`,
  - running as root,
  - or `CSS_VISUAL_DIFF_CHROME_NO_SANDBOX=true` is explicitly set.
- `CSS_VISUAL_DIFF_CHROME_NO_SANDBOX=false` explicitly disables the no-sandbox override, even in CI.

I added `internal/cssvisualdiff/driver/chrome_test.go` to cover the environment override behavior and boolean parsing.

### Validation

```bash
gofmt -w internal/cssvisualdiff/driver/chrome.go internal/cssvisualdiff/driver/chrome_test.go
go test ./internal/cssvisualdiff/driver ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
go test ./...
```

All passed locally. The important regression check is the `CI=true` targeted run, which exercises the same allocator branch GitHub Actions should take.

## Implementation Step 17: fix remaining dependency-scanning CI failures

After pushing the Chrome sandbox fix, the `golang-pipeline` test job passed, but dependency scanning still failed in two jobs.

### govulncheck failure

`govulncheck` reported standard-library vulnerabilities in Go `1.26.1`:

- `GO-2026-4947` in `crypto/x509`, fixed in `1.26.2`,
- `GO-2026-4946` in `crypto/x509`, fixed in `1.26.2`,
- `GO-2026-4870` in `crypto/tls`, fixed in `1.26.2`,
- `GO-2026-4866` in `crypto/x509`, fixed in `1.26.2`,
- `GO-2026-4865` in `html/template`, fixed in `1.26.2`.

Because the workflow uses `actions/setup-go` with `go-version-file: go.mod`, I bumped the module directive from:

```text
go 1.26.1
```

to:

```text
go 1.26.2
```

### gosec failure

`gosec` reported `G104` for ignored `exports.Set(...)` errors in `internal/cssvisualdiff/dsl/registrar.go`. I changed those calls to assign to `_`, matching the style already used elsewhere in the Goja adapter:

```go
_ = exports.Set("compareRegion", ...)
_ = exports.Set("agentBrief", ...)
_ = exports.Set("renderAgentBrief", ...)
```

### Validation

```bash
go test ./...
CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
govulncheck ./...
gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history -exclude-dir=ttmp ./...
```

Local results:

- `go test ./...`: passed,
- CI-mode targeted browser tests: passed,
- `govulncheck ./...`: `No vulnerabilities found`,
- `gosec ...`: `Issues: 0`.

## Implementation Step 18: address PR review comments

Codex left three inline review comments on PR #2. I addressed all three.

### 1. Reject `outputFile` with multiple inspect requests

Problem: `service.InspectPreparedPage` accepted multiple requests plus `OutputFile`, which meant every request wrote to the same file path and earlier artifacts could be overwritten.

Fix:

- Added pre-loop validation in `InspectPreparedPage`:

```go
if opts.OutputFile != "" && len(requests) != 1 {
    return result, fmt.Errorf("outputFile requires exactly one inspect request, got %d", len(requests))
}
```

- Added `TestInspectPreparedPageRejectsOutputFileWithMultipleRequests`.

### 2. Stop hijacking verb-level `--repository` flags

Problem: `repositoriesFromArgs` stripped `--repository` and `--verb-repository` anywhere in the arg list before the generated verb command could parse its own flags. That meant a user-defined verb flag named `repository` could never receive its value.

Fix:

- Changed bootstrap parsing to consume repository flags only from the leading bootstrap prefix.
- Parsing stops at the first non-bootstrap argument.
- `--` is honored and removed as a bootstrap delimiter.
- Added tests for both prefix-only parsing and `--` pass-through.

### 3. Include all records in `catalog.manifest()`

Problem: `catalog.manifest()` returned only metadata, targets, and summary, while `writeManifest()` persisted preflights, results, and failures too.

Fix:

- Added lowerCamel conversion for manifest `preflights`, `results`, and `failures`.
- Extended the repository-scanned catalog verb integration test to assert in-memory manifest record counts.

### Validation

```bash
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff
go test ./...
CI=true go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
govulncheck ./...
gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history -exclude-dir=ttmp ./...
```

All passed locally.
