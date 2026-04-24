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
