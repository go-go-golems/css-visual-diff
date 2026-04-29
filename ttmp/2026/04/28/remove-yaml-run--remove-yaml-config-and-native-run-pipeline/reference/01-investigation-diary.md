---
Title: Investigation Diary
Ticket: remove-yaml-run
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - cli
    - config
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/dsl/scripts/catalog.js
      Note: Removed built-in inspect-config verb in Phase 3
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Removed cvd.loadConfig export in Phase 3
    - Path: internal/cssvisualdiff/modes/config_adapters.go
      Note: Temporary legacy config-to-service adapter layer introduced during Phase 1
    - Path: internal/cssvisualdiff/service/runtime_types.go
      Note: Final config-free runtime types introduced during Phase 1
    - Path: ttmp/2026/04/28/remove-yaml-run--remove-yaml-config-and-native-run-pipeline/design-doc/01-removing-yaml-config-and-native-run-pipeline-design-and-implementation-guide.md
      Note: Primary design guide produced during the investigation
ExternalSources: []
Summary: Chronological investigation diary for removing native YAML config and run pipeline complexity from css-visual-diff.
LastUpdated: 2026-04-28T10:45:00-04:00
WhatFor: Record the investigation, design decisions, and validation steps for the YAML/run removal ticket.
WhenToUse: Read before implementing or reviewing the remove-yaml-run work.
---






# Investigation Diary

## Goal

This diary records the analysis and design work for removing the native YAML config and `run --config` pipeline from `css-visual-diff`. The project direction is now JS-first: Pyxis is the only relevant consumer, Pyxis already uses JavaScript verbs and project-specific visual suite specs, and we do not need backwards compatibility for the old native config runner.

---

## Step 1: Create the Ticket and Confirm Scope

Created a separate docmgr ticket named `remove-yaml-run` for the complexity-reduction work. The central goal is to remove the native `config.Config`-driven workflow and move the tool toward a smaller core centered on browser automation, reusable services, and the Goja/JavaScript API.

The key scope distinction is important: this ticket targets the old native `css-visual-diff` YAML config shape and the `run --config` command path. It does **not** target Pyxis visual suite specs in `prototype-design/visual-diff/userland/specs/*.visual.yml`, because those are userland data loaded through JS verb fields (`objectFromFile`) and are the current Pyxis source of truth.

### Prompt Context

**User prompt (verbatim):** "Create a separate ticket to remove yaml config / run entirely, and all the complexity that is attached to it, since we are focusing on the JS scriptability as first feature now. We don't need backwards compatiblity, pyxis is the only relevant project rn.

 reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the     
 system needed to understand what it is, with prose paragraphs and bullet                                                              
 point sand pseudocode and diagrams and api references and file                                                                        
   references.                                                                                                                         
   It should be very clear and detailed. Store in the ticket and the nupload to remarkable.   

The goal is really to remove complexity from css-visual-diff by leveraging the flexibility of the JS approach."

**Assistant interpretation:** Create a new docmgr ticket and write a detailed intern-facing design/implementation guide for deleting the native YAML config runner and related API surface, while preserving and prioritizing JS scriptability for Pyxis.

**Inferred user intent:** The user wants to simplify `css-visual-diff` by removing an older declarative Go/YAML execution path that duplicates or constrains the more flexible JavaScript userland approach.

### What I did
- Created ticket `remove-yaml-run` with topics `css-visual-diff`, `javascript-api`, `cli`, and `config`.
- Searched the core repository for native YAML config and run-pipeline references.
- Searched the Pyxis visual-diff userland for real usage of native config/run paths.
- Wrote the primary design guide in `design-doc/01-removing-yaml-config-and-native-run-pipeline-design-and-implementation-guide.md`.

### Commands run

```bash
cd /home/manuel/code/wesen/corporate-headquarters/css-visual-diff && \
  docmgr ticket create-ticket --ticket remove-yaml-run \
  --title "Remove YAML Config and Native Run Pipeline" \
  --topics css-visual-diff,javascript-api,cli,config
```

```bash
cd /home/manuel/code/wesen/corporate-headquarters/css-visual-diff && \
  rg -n "config\.Load|RunCommand|NewRunCommand|runner\.Run|NormalizeModes|cvd\.loadConfig|loadConfig|inspectConfig|inspect-config|type Config|type Target|type PrepareSpec|type SectionSpec|type StyleSpec|type OutputSpec" \
  cmd internal examples README.md -S
```

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff && \
  rg -n 'native .css-visual-diff run. configs are retired|Future additions should prefer|objectFromFile|compareSpec|inspectSpec|targetsFromSpec|defaultSpec|refresh-spec-mirrors' \
  userland/README.md userland/specs/README.md userland/verbs/pyxis-pages.js userland/lib/registry.js userland/scripts/refresh-spec-mirrors.py -S
```

### What worked
- The repository evidence is clear: native YAML config is concentrated in `internal/cssvisualdiff/config`, the `run` command, the `runner`, config-driven modes, `cvd.loadConfig`, and the built-in `catalog inspect-config` verb.
- Pyxis already documents that native `css-visual-diff run` configs are retired and should not be maintained as a parallel path.

### What didn't work
- One early ripgrep command used unescaped shell backticks in the search expression. The shell tried to run `css-visual-diff run` as command substitution, producing `Error: --config or --config-dir is required`. I reran the search with single quotes and a safer regex.

### What I learned
- The old config package is not just YAML parsing. Its structs (`Target`, `Viewport`, `PrepareSpec`) have leaked into reusable service and JS layers. Removing native YAML cleanly requires first extracting a small runtime type package, then deleting the YAML-specific parts.
- The biggest simplification opportunity is not just deleting `config.Load`; it is removing the entire `run` mode matrix (`capture`, `cssdiff`, `matched-styles`, `pixeldiff`, `ai-review`, `html-report`) as a privileged Go orchestrator and relying on JavaScript verbs for orchestration.

### What was tricky to build
- The term "YAML config" can refer to several things:
  - old native `css-visual-diff` visual configs,
  - Pyxis visual suite specs,
  - `.css-visual-diff.yml` verb repository discovery config.
- The guide therefore defines the deletion target precisely: remove the native Go/YAML runner and compatibility bridge (`cvd.loadConfig` / `inspect-config`), but preserve generic JS-first ways to load project data (`objectFromFile`) because Pyxis depends on that pattern.

### What warrants a second pair of eyes
- Whether to remove `.css-visual-diff.yml` verb repository discovery in the same implementation or defer it. The guide recommends deferring unless the owner explicitly wants to force all repositories through `--repository` / environment variables.

### What should be done in the future
- Implement the phased removal plan.
- Run the Pyxis smoke scripts after deletion to verify that JS userland remains intact.

### Code review instructions
- Start with `cmd/css-visual-diff/main.go` to verify that `run` and `inspect --config` are gone.
- Then inspect `internal/cssvisualdiff/config`, `internal/cssvisualdiff/runner`, and `internal/cssvisualdiff/jsapi/module.go` for stale config dependencies.
- Validate with `GOWORK=off go test ./...` and Pyxis `userland/scripts/smoke-*.sh` scripts.

---

## Step 2: Validation and reMarkable Delivery

Validated the ticket documentation with `docmgr doctor`, then uploaded the design guide and diary as a reMarkable bundle. The upload was verified by listing the remote destination.

### What I did
- Related the key source files and Pyxis userland files to the design document.
- Added implementation tasks and a changelog entry.
- Ran `docmgr doctor --ticket remove-yaml-run --stale-after 30`.
- Ran `remarquee upload bundle --dry-run`, then the real upload.
- Verified the remote upload path.

### What worked
- `docmgr doctor` passed after the design and diary frontmatter were kept valid.
- reMarkable upload completed successfully.

### What didn't work
- N/A

### What I learned
- The existing docmgr vocabulary already contained the topics used by this ticket, so no new vocabulary entries were needed.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Begin Phase 1 of the implementation plan when ready.

### Code review instructions
- N/A

---

## Step 3: Phase 1 Runtime Types and Adapter Cleanup

Implemented Phase 1 by making the config-free service runtime types the source of truth for browser/page operations, then made the legacy config-driven mode code delegate into those final service types through a single temporary adapter file. This cleaned up the earlier loop where config-to-service conversions were scattered across old mode files.

The important architectural decision was to work from the desired final code backward: `service` and `jsapi` now speak `service.Viewport`, `service.PrepareSpec`, `service.PageTarget`, and `service.StyleEvalSpec`; legacy mode code still accepts old `config.*` values, but only converts them at the boundary in `modes/config_adapters.go`. That means Phase 4 can delete the native mode layer and its adapters together.

### Prompt Context

**User prompt (verbatim):** "clean it all up, start actually with delegating the final code, and then working backwards, that should make it easier, no?"

**Assistant interpretation:** Stop chasing compile errors through old mode files. Establish the final config-free service API first, then make old config-driven code delegate to it through a small adapter layer.

**Inferred user intent:** The user wants the refactor to reduce complexity, not spread temporary conversion logic across the codebase.

**Commit (code):** 194dca427da821290f85527e876e755407fe4886 — "Refactor runtime page types out of YAML config"

### What I did
- Added `internal/cssvisualdiff/service/runtime_types.go` with config-free `Viewport`, `PrepareSpec`, and `PageTarget`.
- Added `service.StyleEvalSpec` so style extraction no longer depends on `config.StyleSpec`.
- Updated service APIs (`PrepareTarget`, `LoadAndPreparePage`, `InspectPreparedPage`, `WriteInspectArtifacts`, `EvaluateStyle`) to accept service runtime types.
- Updated Goja JS API page state, target builder, viewport helpers, catalog decoding, and prepare decoding to use service runtime types.
- Added `internal/cssvisualdiff/modes/config_adapters.go` as the single temporary adapter layer from old `config.*` structs into new `service.*` structs.
- Removed duplicated adapter helpers from individual mode files.
- Updated service tests to use service runtime types directly.

### Why
- Phase 1's purpose is to break the dependency from reusable runtime layers (`service` and `jsapi`) to the native YAML config package.
- Old config-driven modes still exist until Phase 4, so they need a temporary bridge. Centralizing that bridge keeps the codebase understandable and makes deletion easy later.

### What worked
- `GOWORK=off go test ./...` passed.
- The pre-commit hook also ran `golangci-lint` and `go test ./...`; both passed.
- `rg -n "internal/cssvisualdiff/config" internal/cssvisualdiff/service internal/cssvisualdiff/jsapi --glob '*.go'` now only shows JS compatibility code for `cvd.loadConfig`, which is scheduled for Phase 3.
- Temporary config adapters are now centralized in `internal/cssvisualdiff/modes/config_adapters.go`.

### What didn't work
- The initial migration changed shared service signatures before centralizing adapters, which caused a compile cascade through config-driven modes.
- There were duplicate helper definitions (`toServicePageTarget`, `toServicePrepareSpec`) before cleanup.
- Service tests initially still used `config.Target` and `config.Viewport`, which no longer matched service API signatures.

### What I learned
- The best refactor shape is final API first, compatibility wrapper second. In this codebase that means `service`/`jsapi` should not import native config types, while old `modes` may temporarily adapt them.
- `StyleEvalSpec` needed to include `Report` temporarily because matched-style legacy code uses that field to decide whether to include box-model bounds.

### What was tricky to build
- The old `config` package held both obsolete YAML schema and reusable runtime concepts. Pulling out `Viewport`, `PrepareSpec`, and `PageTarget` exposed additional hidden coupling through style extraction and catalog metadata.
- `jsapi/config.go` still exists for `cvd.loadConfig` until Phase 3. It now contains a small conversion from `config.Viewport` to `service.Viewport`; this is intentionally temporary and should be deleted with `cvd.loadConfig` rather than polished further.

### What warrants a second pair of eyes
- Confirm that `service.StyleEvalSpec.Report` is acceptable as a temporary bridge for matched styles, or whether Phase 2 should separate matched-style inputs from computed-style inputs before Phase 4 deletion.
- Confirm that `modes/config_adapters.go` is treated as temporary and removed with config-driven modes.

### What should be done in the future
- Phase 2: remove remaining non-compat `config` imports from JS/service-facing code where possible, while preserving `cvd.loadConfig` only until Phase 3.
- Phase 3: delete `cvd.loadConfig`, `jsapi/config.go`, and built-in `catalog inspect-config`.

### Code review instructions
- Start with `internal/cssvisualdiff/service/runtime_types.go` and `internal/cssvisualdiff/service/style.go` to see the final runtime types.
- Then inspect `internal/cssvisualdiff/modes/config_adapters.go` to see the only intentional old-config bridge.
- Finally check `internal/cssvisualdiff/jsapi/module.go`, `target.go`, `catalog.go`, and `builder_helpers.go` for service runtime type usage.
- Validate with `GOWORK=off go test ./...`.

---

## Step 4: Remove Native YAML JS Compatibility Bridge

Removed the old JavaScript compatibility bridge for native YAML configs. This completed the practical cleanup from Phase 2 and Phase 3: reusable service and JS API layers no longer import `internal/cssvisualdiff/config`, and the built-in `catalog inspect-config` verb is gone.

This step deliberately did **not** remove generic YAML/object loading for JS verbs. Pyxis still uses `objectFromFile` for project-specific visual specs, which is the desired JS-first model. What disappeared is only the old native `css-visual-diff` config schema bridge (`cvd.loadConfig`) and the catalog command built on top of it.

### Prompt Context

**User prompt (verbatim):** (see Step 3)

**Assistant interpretation:** Continue removing YAML/run complexity after cleaning up the Phase 1 runtime type refactor.

**Inferred user intent:** The user wants momentum toward the simplified JS-first architecture and removal of old compatibility paths.

**Commit (code):** 3d954892506fe50207eb704565424c9ac4a99c63 — "Remove YAML config JS compatibility bridge"

### What I did
- Removed `cvd.loadConfig(path)` from `internal/cssvisualdiff/jsapi/module.go`.
- Deleted `internal/cssvisualdiff/jsapi/config.go`, which only lowered native YAML config structs into JS objects.
- Removed `_selectorForSide`, `_probesFromConfig`, `inspectConfig`, and the `inspectConfig` verb registration from `internal/cssvisualdiff/dsl/scripts/catalog.js`.
- Updated `internal/cssvisualdiff/verbcli/command_test.go` so built-in verb tests no longer expect `catalog inspect-config`.
- Removed `inspect-config` and `loadConfig` references from README and JavaScript API/verbs docs.
- Checked Phase 2 and Phase 3 tasks in docmgr.

### Why
- `cvd.loadConfig` preserved the old native YAML schema as a JS API surface. Keeping it would force the config package and lowering code to remain alive.
- `catalog inspect-config` depended entirely on that bridge and duplicated JS-first inspection workflows.
- Removing both keeps scripts focused on explicit project specs and direct JS primitives instead of old Go config objects.

### What worked
- `rg -n "loadConfig|inspectConfig|inspect-config|lowerConfig" internal README.md cmd examples -S` now only finds unrelated `loadConfigRepositories` references for verb repository discovery.
- `rg -n "internal/cssvisualdiff/config" internal/cssvisualdiff/service internal/cssvisualdiff/jsapi --glob '*.go'` returns no service/jsapi config imports.
- `GOWORK=off go test ./...` passed.
- The pre-commit hook ran `golangci-lint` and `go test ./...`; both passed.

### What didn't work
- N/A

### What I learned
- Removing the compatibility bridge was a clean deletion once Phase 1 made the JS runtime independent of `config.Target`, `config.Viewport`, and `config.PrepareSpec`.
- The remaining `loadConfigRepositories` name in `verbcli/bootstrap.go` is unrelated to native visual configs; it loads app config for JS verb repositories and should not be removed in this phase.

### What was tricky to build
- The name `loadConfig` appears both in the old JS API (`cvd.loadConfig`) and in verb repository discovery (`loadConfigRepositories`). The latter is still part of JS scriptability and should remain unless a separate ticket decides to remove `.css-visual-diff.yml` repository overlays.

### What warrants a second pair of eyes
- Check the docs diff to ensure we did not leave stale user instructions for `catalog inspect-config` or `cvd.loadConfig`.

### What should be done in the future
- Phase 4: remove the native `run` command, `runner` package, config-driven modes, and the temporary `modes/config_adapters.go` layer.

### Code review instructions
- Review `internal/cssvisualdiff/jsapi/module.go` and confirm there is no `loadConfig` export.
- Review `internal/cssvisualdiff/dsl/scripts/catalog.js` and confirm only JS-first catalog verbs remain.
- Run `GOWORK=off go test ./...`.
