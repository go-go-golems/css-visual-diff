---
Title: Removing YAML Config and Native Run Pipeline - Design and Implementation Guide
Ticket: remove-yaml-run
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - cli
    - config
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../2026-04-23--pyxis/prototype-design/visual-diff/userland/README.md
      Note: Documents that Pyxis JS userland and visual specs are canonical and native run configs are retired
    - Path: ../../../../../../../../2026-04-23--pyxis/prototype-design/visual-diff/userland/verbs/pyxis-pages.js
      Note: Shows JS-first Pyxis verbs using objectFromFile specs rather than native config
    - Path: cmd/css-visual-diff/main.go
      Note: Contains the native run command
    - Path: internal/cssvisualdiff/config/config.go
      Note: Defines native YAML config schema and loader to delete after moving runtime types
    - Path: internal/cssvisualdiff/dsl/scripts/catalog.js
      Note: Contains built-in catalog inspectConfig verb to delete
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Exposes cvd.loadConfig today and must be migrated to config-free runtime types
    - Path: internal/cssvisualdiff/runner/runner.go
      Note: Fixed Go mode runner over config.Config to delete with run command
ExternalSources: []
Summary: Detailed plan to remove native YAML config/run complexity from css-visual-diff and make JS scriptability the primary workflow.
LastUpdated: 2026-04-28T09:15:00-04:00
WhatFor: Guide an intern through understanding and implementing the removal of the native YAML config runner and related APIs.
WhenToUse: Use when simplifying css-visual-diff around the Goja/JavaScript API and Pyxis userland workflows.
---








# Removing YAML Config and Native Run Pipeline: Design and Implementation Guide

## Executive Summary

`css-visual-diff` currently has two competing orchestration models:

1. **Native Go/YAML runner** — users write a `*.css-visual-diff.yml`-style config with `original`, `react`, `sections`, `styles`, `output`, and `modes`, then invoke `css-visual-diff run --config ...`. Go code parses the config and runs a fixed mode pipeline (`capture`, `cssdiff`, `matched-styles`, `pixeldiff`, `ai-review`, `html-report`).
2. **JavaScript-first runtime** — users write JavaScript verbs/scripts against the Goja module `require("css-visual-diff")`, and orchestrate browser pages, locators, probes, inspect artifacts, catalogs, comparisons, policies, and project-specific specs themselves.

The project direction is now explicitly JS-first. Pyxis is the only relevant current consumer, and Pyxis already documents that native `css-visual-diff run` configs are retired. The goal of this ticket is therefore to **delete the native YAML config/run pipeline and its compatibility bridge**, not to preserve backwards compatibility.

The desired end state is a smaller core:

```text
css-visual-diff binary
  ├── verbs ...                    # primary user workflow
  ├── JS module: css-visual-diff    # browser/page/locator/inspect/diff/catalog primitives
  ├── service layer                 # reusable Go primitives used by JS
  └── driver layer                  # chromedp wrapper
```

And to remove the old path:

```text
DELETE / DEPRECATE TO ZERO
  ├── css-visual-diff run --config / --config-dir
  ├── internal/cssvisualdiff/config.Load / Config / SectionSpec / StyleSpec / OutputSpec
  ├── internal/cssvisualdiff/runner
  ├── config-driven modes that only exist for run
  ├── cvd.loadConfig(path)
  └── verbs catalog inspect-config
```

This is a complexity-reduction project. It should make future features, including overlay screenshot labeling, easier to design because they can be exposed once as JS primitives instead of twice as YAML schema plus JS API.

---

## Definitions

### Native YAML config

In this document, **native YAML config** means the old `css-visual-diff` config shape implemented by `internal/cssvisualdiff/config/config.go`:

```yaml
metadata:
  slug: example
original:
  url: http://localhost:7070/page.html
react:
  url: http://localhost:6007/iframe.html?id=page--story
sections:
  - name: hero
    selector_original: '[data-section="hero"]'
    selector_react: '[data-pyxis-component="hero"]'
styles:
  - name: button
    selector: '[data-component="button"]'
    props: [display, color, background-color]
output:
  dir: ./out
modes: [capture, cssdiff]
```

This is the thing we remove.

### Pyxis visual suite spec

A **Pyxis visual suite spec** is a project-specific userland YAML file like:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml
```

It uses `schemaVersion: pyxis.visual-suite.v1`, `defaults`, `targets`, `sections`, and project fields like policy metadata. Pyxis JS verbs load this through `objectFromFile` or generated CommonJS mirrors.

This is **not** the same as native `css-visual-diff` YAML config. Pyxis specs remain valid because they are JS userland data, not Go-native config.

### JS scriptability

**JS scriptability** means the Goja runtime and verb system:

- `css-visual-diff verbs ...`
- built-in and external JS verb repositories
- `require("css-visual-diff")`
- `cvd.browser()`
- `browser.page(url, options)`
- `page.locator(selector)`
- `page.inspect(...)`
- `page.inspectAll(...)`
- `cvd.compare.region(...)`
- `cvd.catalog(...)`
- `cvd.diff(...)`
- generic verb field loaders such as `objectFromFile`

This becomes the primary product surface.

---

## Problem Statement

The native YAML runner duplicates orchestration responsibility that now belongs in JS scripts. It creates several maintenance problems:

1. **Two ways to describe workflows.** A user can encode comparisons in native YAML or in JS userland specs. This creates ambiguity about where new features should be added.
2. **Rigid Go schema.** Every new feature requires adding fields to Go structs, YAML validation, lowerCamel JS interop, docs, tests, and config-driven mode support.
3. **Config types leaked into reusable layers.** Structs from `internal/cssvisualdiff/config` are used by `service`, `jsapi`, and `modes`, even when those layers are not loading YAML.
4. **More code paths to test.** Browser prepare, capture, inspect, CSS extraction, pixel diff, AI review, and HTML report exist as config-driven Go modes in addition to JS workflows.
5. **Pyxis does not need it.** Pyxis explicitly uses the JS userland and project-specific visual suite specs as the canonical workflow.

The goal is to remove this complexity without weakening JS-first functionality. After this change, adding a feature should generally mean:

1. Add or update a Go service primitive if needed.
2. Expose it through the Goja JS API.
3. Use it from Pyxis or another userland JS verb.

It should **not** mean adding new native YAML fields or new `run` modes.

---

## Evidence-Based Current State

### Native `run` command exists in the root CLI

`cmd/css-visual-diff/main.go` defines `RunCommand`, `RunSettings`, and `NewRunCommand`. The command accepts:

- `--config`
- `--config-dir`
- `--modes`
- `--dry-run`
- `--pixeldiff-threshold`
- AI profile flags

Evidence from `rg`:

```text
cmd/css-visual-diff/main.go:36:type RunCommand struct {
cmd/css-visual-diff/main.go:51:func NewRunCommand() (*RunCommand, error) {
cmd/css-visual-diff/main.go:153: cfg, err := config.Load(configPath)
cmd/css-visual-diff/main.go:163: modesList, err := runner.NormalizeModes(modesRaw)
cmd/css-visual-diff/main.go:184: result, err := runner.Run(ctx, cfg, modesList, settings.DryRun, runOptions)
cmd/css-visual-diff/main.go:328: runCmd, err := NewRunCommand()
```

The high-level current `run` flow is:

```text
CLI args
  ↓
resolveRunConfigPaths(settings)
  ↓
for each configPath:
  config.Load(configPath)
    ↓
  runner.NormalizeModes(settings.Modes or cfg.Modes)
    ↓
  runner.Run(ctx, cfg, modesList, dryRun, runOptions)
    ↓
  modes.Capture / PixelDiff / CSSDiff / MatchedStyles / AIReview / HTMLReport
```

### Config package carries both YAML schema and shared runtime types

`internal/cssvisualdiff/config/config.go` defines the old schema:

```text
internal/cssvisualdiff/config/config.go:34:type Target struct {
internal/cssvisualdiff/config/config.go:43:type PrepareSpec struct {
internal/cssvisualdiff/config/config.go:66:type SectionSpec struct {
internal/cssvisualdiff/config/config.go:103:type StyleSpec struct {
internal/cssvisualdiff/config/config.go:114:type OutputSpec struct {
internal/cssvisualdiff/config/config.go:124:type Config struct {
```

This matters because not all of these are equally obsolete:

- `Config`, `SectionSpec`, `StyleSpec`, `OutputSpec`, and YAML `Load/Validate` belong to the old native runner and should disappear.
- `Viewport`, `Target`, and `PrepareSpec` describe runtime browser/page behavior and are still useful to JS APIs, but should move to a neutral runtime/service package so they no longer imply YAML config support.

### Runner is entirely config-driven

`internal/cssvisualdiff/runner/runner.go` defines fixed mode orchestration:

```go
var defaultModes = []string{"capture", "cssdiff"}

func Run(ctx context.Context, cfg *config.Config, modesList []string, dryRun bool, options RunOptions) (RunResult, error) {
    for _, mode := range modesList {
        switch mode {
        case "capture":
            err = modes.Capture(ctx, cfg)
        case "pixeldiff":
            err = modes.PixelDiff(ctx, cfg, threshold)
        case "cssdiff":
            err = modes.CSSDiff(ctx, cfg)
        case "matched-styles":
            err = modes.MatchedStyles(ctx, cfg)
        case "story-discovery":
            err = modes.StoryDiscovery(ctx, cfg)
        case "ai-review":
            err = modes.AIReview(ctx, cfg)
        case "html-report":
            err = modes.HTMLReport(ctx, cfg)
        }
    }
}
```

Once `run --config` is removed, this package has no obvious reason to exist.

### Compatibility bridge exists in JS API

The JS API currently exposes `cvd.loadConfig(path)`:

```text
internal/cssvisualdiff/jsapi/module.go:35: _ = exports.Set("loadConfig", func(path string) goja.Value {
internal/cssvisualdiff/jsapi/module.go:37:     return config.Load(path)
```

The lowering code lives in:

```text
internal/cssvisualdiff/jsapi/config.go
```

This is specifically for native YAML interop. It loads the old Go schema and returns a lowerCamel JS object. Under a JS-first architecture, scripts should load their own project data with `objectFromFile`, CommonJS modules, or future generic file helpers instead of using a native `css-visual-diff` config schema.

### Built-in `catalog inspect-config` verb exists only for old YAML configs

`internal/cssvisualdiff/dsl/scripts/catalog.js` defines `inspectConfig`:

```text
internal/cssvisualdiff/dsl/scripts/catalog.js:62:async function inspectConfig(configPath, side, outDir, values) {
internal/cssvisualdiff/dsl/scripts/catalog.js:65:  const cfg = await cvd.loadConfig(configPath);
internal/cssvisualdiff/dsl/scripts/catalog.js:157:__verb__("inspectConfig", {
```

This command converts `styles` or `sections` from the old Go config into probes and catalogs. It is a compatibility bridge and should be deleted with `cvd.loadConfig`.

### Pyxis does not need native config/run

Pyxis userland says:

```text
userland/README.md:5: ... The JS userland and specs/*.visual.yml files are the canonical Pyxis workflow; native css-visual-diff run configs are retired and should not be maintained as a parallel path.
```

Pyxis specs README also says:

```text
userland/specs/README.md:78: Future additions should prefer project-specific fields such as policy, accepted differences, semantic snapshot presets, and role metadata over native css-visual-diff run config shape.
```

Pyxis commands use JS verbs and generic object loading:

```text
userland/verbs/pyxis-pages.js:315: spec: { argument: true, type: 'objectFromFile', required: true, help: 'JSON/YAML visual spec with pages and sections' }
userland/verbs/pyxis-pages.js:392: spec: { argument: true, type: 'objectFromFile', required: true, help: 'JSON/YAML visual spec with pages and sections' }
```

This proves that the relevant project already follows the desired architecture.

---

## Desired End State

The future architecture should look like this:

```text
                       ┌───────────────────────────────┐
                       │ Pyxis visual suite YAML specs │
                       │ userland/specs/*.visual.yml   │
                       └───────────────┬───────────────┘
                                       │ objectFromFile / generated JS mirror
                                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│ JS USERLAND                                                         │
│ userland/verbs/pyxis-pages.js                                       │
│ userland/lib/registry.js                                            │
│ userland/lib/compare-region.js                                      │
│ userland/lib/inspect.js                                             │
│ userland/lib/snapshot.js                                            │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ require("css-visual-diff")
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│ GOJA JS API                                                         │
│ cvd.browser, page.locator, page.inspectAll, cvd.compare, cvd.catalog│
└───────────────────────────────┬─────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│ GO SERVICE LAYER                                                    │
│ browser/page loading, prepare, inspect artifacts, diff/catalog       │
└───────────────────────────────┬─────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│ DRIVER LAYER                                                        │
│ chromedp browser/page/evaluate/screenshot                           │
└─────────────────────────────────────────────────────────────────────┘
```

Removed from the architecture:

```text
native css-visual-diff YAML config
  ├── config.Load / Validate
  ├── css-visual-diff run --config
  ├── runner.NormalizeModes / runner.Run
  ├── modes.Capture(cfg), modes.CSSDiff(cfg), modes.MatchedStyles(cfg), ...
  ├── cvd.loadConfig(path)
  └── verbs catalog inspect-config
```

---

## Scope: What to Remove, Keep, or Move

### Remove

| Item | Why remove |
|---|---|
| `css-visual-diff run` command | Main native YAML entrypoint; no longer needed. |
| `--config`, `--config-dir`, `--modes`, `--dry-run` runner flags | Only support native YAML pipeline. |
| `internal/cssvisualdiff/config.Load` and validation | Old schema loader. |
| `config.Config`, `SectionSpec`, `StyleSpec`, `OutputSpec`, `Metadata` | Native config schema. |
| `internal/cssvisualdiff/runner` | Only orchestrates config-driven modes. |
| `cvd.loadConfig(path)` | Compatibility bridge for old schema. |
| `internal/cssvisualdiff/jsapi/config.go` | Only lowers old config shape into JS. |
| Built-in `verbs catalog inspect-config` | Uses `cvd.loadConfig`; old-schema bridge. |
| README/docs/tutorial content for native YAML configs | Avoid documenting removed workflows. |
| Old examples under `examples/*.yaml` that are native configs | Prevent users from copying dead patterns. |

### Move or rename

| Current item | Proposed destination | Why |
|---|---|---|
| `config.Viewport` | `service.Viewport` or `runtime.Viewport` | Generic runtime concept, not YAML-specific. |
| `config.Target` | `service.PageTarget` or `runtime.PageTarget` | Used by browser/page loading and metadata. |
| `config.PrepareSpec` | `service.PrepareSpec` or `runtime.PrepareSpec` | JS `page.prepare()` still needs it. |
| helper `RootSelectorForTarget` if present | service/runtime helper | Useful for inspect metadata. |
| `configStyleSpec` helper | service-local request type | Avoid importing config just to evaluate style. |

### Keep

| Item | Why keep |
|---|---|
| `css-visual-diff verbs ...` | Primary product workflow. |
| Verb repository support via `--repository` | Needed for Pyxis userland. |
| Generic `objectFromFile` verb field type | Pyxis uses it to load visual suite YAML specs. |
| Pyxis `*.visual.yml` specs | Project source of truth; not native config. |
| `cvd.browser`, `browser.page`, `page.locator`, `page.inspect`, `page.inspectAll` | Core JS primitives. |
| `cvd.catalog`, `cvd.diff`, `cvd.write` | Useful JS-first building blocks. |
| Direct compare workflow if still used by JS (`cvd.compare.region`) | JS userland depends on region comparison. |

### Decide separately

| Item | Recommendation |
|---|---|
| `.css-visual-diff.yml` verb repository discovery | Defer. It is YAML, but it supports JS scriptability rather than native visual configs. Since Pyxis can pass `--repository`, it can be removed later if desired, but it is not required for this ticket. |
| Standalone direct CLI commands like `compare` or `chromedp-probe` | Keep unless they import config or duplicate JS flows badly. This ticket is specifically about native YAML config/run. |

---

## Proposed Refactor Strategy

The implementation should be done in layers so that the codebase keeps compiling between phases.

### Phase 1: Introduce config-free runtime types

Create a small package or service-local types file for runtime page concepts.

Recommended new file:

```text
internal/cssvisualdiff/service/runtime_types.go
```

Potential content:

```go
package service

// Viewport is the browser viewport used when loading a page.
type Viewport struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

// PrepareSpec describes optional page preparation after navigation.
type PrepareSpec struct {
    Type string `json:"type"`

    Script     string `json:"script"`
    ScriptFile string `json:"scriptFile"`

    WaitFor          string `json:"waitFor"`
    WaitForTimeoutMS int    `json:"waitForTimeoutMs"`
    AfterWaitMS      int    `json:"afterWaitMs"`

    Component    string         `json:"component"`
    Props        map[string]any `json:"props"`
    RootSelector string         `json:"rootSelector"`
    Width        int            `json:"width"`
    MinHeight    int            `json:"minHeight"`
    Background   string         `json:"background"`
}

// PageTarget is what the browser service needs to load and prepare a page.
type PageTarget struct {
    Name         string       `json:"name"`
    URL          string       `json:"url"`
    WaitMS       int          `json:"waitMs"`
    Viewport     Viewport     `json:"viewport"`
    RootSelector string       `json:"rootSelector,omitempty"`
    Prepare      *PrepareSpec `json:"prepare,omitempty"`
}
```

Then update service functions:

```go
// Before
func PrepareTarget(page *driver.Page, target config.Target) error
func RunScriptPrepare(page *driver.Page, prepare *config.PrepareSpec) error
func BuildDirectReactGlobalScript(prepare *config.PrepareSpec) (string, error)
func LoadAndPreparePage(page *driver.Page, target config.Target) error
func InspectPreparedPage(page *driver.Page, target config.Target, side string, ...)

// After
func PrepareTarget(page *driver.Page, target service.PageTarget) error
func RunScriptPrepare(page *driver.Page, prepare *service.PrepareSpec) error
func BuildDirectReactGlobalScript(prepare *service.PrepareSpec) (string, error)
func LoadAndPreparePage(page *driver.Page, target service.PageTarget) error
func InspectPreparedPage(page *driver.Page, target service.PageTarget, side string, ...)
```

This phase is important because it decouples reusable JS-first services from the soon-to-be-deleted YAML package.

### Phase 2: Update JS API to use runtime types directly

Files to modify:

```text
internal/cssvisualdiff/jsapi/module.go
internal/cssvisualdiff/jsapi/target.go
internal/cssvisualdiff/jsapi/catalog.go
internal/cssvisualdiff/jsapi/builder_helpers.go
```

The current `pageOptions.toTarget` returns `config.Target`. Change it to return `service.PageTarget`.

Pseudocode:

```go
type pageOptions struct {
    Viewport service.Viewport `json:"viewport"`
    WaitMS   int              `json:"waitMs"`
    Name     string           `json:"name"`
}

func (o pageOptions) toTarget(url string) (service.PageTarget, error) {
    url = strings.TrimSpace(url)
    if url == "" {
        return service.PageTarget{}, fmt.Errorf("url is required")
    }
    viewport := o.Viewport
    if viewport.Width <= 0 {
        viewport.Width = 1280
    }
    if viewport.Height <= 0 {
        viewport.Height = 720
    }
    name := o.Name
    if name == "" {
        name = "script"
    }
    return service.PageTarget{Name: name, URL: url, WaitMS: o.WaitMS, Viewport: viewport}, nil
}
```

Change `decodePrepareSpec` to return `service.PrepareSpec`, not `config.PrepareSpec`.

Remove `cvd.loadConfig(path)` from `install Register` in `jsapi/module.go`.

Delete `internal/cssvisualdiff/jsapi/config.go` after no code references `lowerConfig`.

### Phase 3: Remove built-in old-config verb

File:

```text
internal/cssvisualdiff/dsl/scripts/catalog.js
```

Remove:

- `_selectorForSide`
- `_probesFromConfig`
- `inspectConfig`
- `__verb__("inspectConfig", ...)`

Keep:

- `inspectPage`
- catalog helpers unrelated to native config

Then update docs/tests:

- Remove `inspect-config` from `README.md`.
- Remove `inspect-config` from `internal/cssvisualdiff/doc/topics/javascript-api.md`.
- Remove `inspect-config` from `internal/cssvisualdiff/doc/topics/javascript-verbs.md`.
- Update `internal/cssvisualdiff/verbcli/command_test.go` where it expects `catalog inspect-config` to exist.

### Phase 4: Delete native run command

File:

```text
cmd/css-visual-diff/main.go
```

Remove:

- `RunCommand`
- `RunSettings`
- `NewRunCommand`
- `RunIntoGlazeProcessor`
- `resolveRunConfigPaths`
- `runOneConfig`
- `discoverRunConfigFiles`
- `shouldSkipConfigScanDir`
- `isRunConfigFileName`
- `containsMode`
- coverage/story-discovery row emission functions only used by `run`
- registration of `runCmd` in the root command

Also remove tests:

```text
cmd/css-visual-diff/main_test.go
```

Delete or update tests that only validate `run` behavior:

- `TestDiscoverRunConfigFilesFindsCoLocatedConfigs`
- `TestResolveRunConfigPathsScansConfigDir`
- `TestRunCommandDryRunDecodesConfigFlag`
- `TestRunCommandIncludesAIReviewProfileFlags`
- helper `writeRunConfig`

If the file also tests other commands, edit it; otherwise delete it.

### Phase 5: Delete runner and config-driven modes

Delete:

```text
internal/cssvisualdiff/runner/runner.go
```

Then examine `internal/cssvisualdiff/modes/`. The goal is to remove config-driven mode wrappers, not necessarily every reusable utility.

Likely delete or heavily trim:

```text
internal/cssvisualdiff/modes/capture.go          # config-driven capture report
internal/cssvisualdiff/modes/cssdiff.go          # config-driven style diff over Config.Styles
internal/cssvisualdiff/modes/matched_styles.go   # config-driven matched-style analysis
internal/cssvisualdiff/modes/pixeldiff.go        # config-driven pixel diff over capture outputs
internal/cssvisualdiff/modes/html_report.go      # config-driven report over capture/diff outputs
internal/cssvisualdiff/modes/ai_review.go        # config-driven AI review over capture outputs
internal/cssvisualdiff/modes/stories.go          # config-driven story discovery if only run-mode based
internal/cssvisualdiff/modes/modes.go            # wrappers around Config
```

Be careful with utility functions/types that are reused by direct compare or JS APIs. For example:

- `modes/compare.go` provides direct region comparison functionality and may be used by JS compare APIs.
- Pixel diff helpers might be useful to direct compare.
- Image analysis helpers might be used by compare artifacts.

Recommended rule:

```text
If a file's public API takes *config.Config, delete it or rewrite it.
If a helper is used by JS-first compare/inspect/catalog APIs, move it to service or keep it config-free.
```

### Phase 6: Delete native config package

After all imports of `internal/cssvisualdiff/config` are gone, delete:

```text
internal/cssvisualdiff/config/config.go
internal/cssvisualdiff/config/config_test.go
```

Use this verification command:

```bash
rg -n "internal/cssvisualdiff/config|config\." internal cmd --glob '*.go'
```

The result should not include imports of the old package. It may include unrelated local variable names named `config`, but those should not refer to `internal/cssvisualdiff/config`.

### Phase 7: Clean examples and docs

Remove or rewrite:

```text
examples/pyxis-atoms-prototype-vs-storybook.yaml
examples/pyxis-prototype-only.yaml
examples/pyxis-prototype-vs-app.yaml
examples/pyxis-public-shows.yaml
examples/pyxis-storybook-shows-desktop.yaml
```

These should not be maintained as native config examples. If useful, replace with JS examples:

```text
examples/scripts/compare-region.js
examples/scripts/inspect-page.js
examples/scripts/catalog-page.js
examples/userland-specs/example.visual.yml
```

Update docs:

- `README.md`
- `internal/cssvisualdiff/doc/topics/javascript-api.md`
- `internal/cssvisualdiff/doc/topics/javascript-verbs.md`
- `internal/cssvisualdiff/doc/topics/config-selectors.md` — likely delete.
- `internal/cssvisualdiff/doc/tutorials/story-config-authoring.md` — delete or rewrite as JS/spec authoring.
- `internal/cssvisualdiff/doc/tutorials/inspect-workflow.md` — remove config-specific steps.

### Phase 8: Validate Pyxis remains functional

From Pyxis visual-diff userland:

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff

css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages list-targets \
  --output json

css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page archive \
  --section content \
  --outDir /tmp/pyxis-archive-content \
  --summary \
  --output json
```

If local servers are needed, start the prototype and Storybook servers according to Pyxis `userland/README.md`.

---

## Implementation Checklist

### Mechanical removal checklist

Use this exact order during implementation:

1. Add `service.Viewport`, `service.PrepareSpec`, `service.PageTarget`.
2. Replace config imports in service files:
   - `service/prepare.go`
   - `service/browser.go`
   - `service/inspect.go`
   - `service/catalog_service.go`
   - `service/style.go`
   - `service/dom.go`
3. Replace config imports in JS API files:
   - `jsapi/module.go`
   - `jsapi/target.go`
   - `jsapi/catalog.go`
   - `jsapi/builder_helpers.go`
4. Remove `cvd.loadConfig` and `jsapi/config.go`.
5. Remove `catalog inspect-config` from `dsl/scripts/catalog.js`.
6. Remove tests/docs that mention `inspect-config`.
7. Remove `run` command from `cmd/css-visual-diff/main.go`.
8. Remove run-command tests.
9. Remove `internal/cssvisualdiff/runner`.
10. Remove or rewrite config-driven modes.
11. Delete `internal/cssvisualdiff/config`.
12. Delete old native YAML examples.
13. Run `go test ./...`.
14. Run Pyxis userland smoke tests.

### Verification commands

```bash
cd /home/manuel/code/wesen/corporate-headquarters/css-visual-diff

# There should be no old native config package imports.
rg -n "internal/cssvisualdiff/config" --glob '*.go'

# There should be no loadConfig bridge.
rg -n "loadConfig|inspectConfig|inspect-config|config\.Load" cmd internal README.md examples -S

# There should be no run command implementation.
rg -n "RunCommand|NewRunCommand|runner\.Run|--config-dir|css-visual-diff YAML config" cmd internal README.md -S

# Build and test.
GOWORK=off go test ./...
GOWORK=off go build ./cmd/css-visual-diff
```

Expected result after successful implementation:

- `rg` searches return no old native config/run references, except possibly historical ticket docs under `ttmp/` if searched globally.
- `go test ./...` passes.
- `css-visual-diff verbs --help` still works.
- Pyxis userland verbs still work.

---

## Pseudocode: Before and After

### Before: native run command

```go
func runOneConfig(ctx context.Context, settings RunSettings, path string) error {
    cfg := config.Load(path)
    modes := runner.NormalizeModes(settings.Modes or cfg.Modes)
    return runner.Run(ctx, cfg, modes, settings.DryRun, options)
}
```

### After: no native runner

```go
func rootCommand() *cobra.Command {
    root := &cobra.Command{Use: "css-visual-diff"}

    // Keep JS-first verb system.
    root.AddCommand(newVerbsCommand())

    // Optional: keep direct low-level utilities if useful.
    root.AddCommand(newCompareCommand())
    root.AddCommand(newChromedpProbeCommand())

    // Do not register run.
    // Do not register config-driven inspect.
    return root
}
```

### Before: JS native config bridge

```js
const cfg = await cvd.loadConfig("page.css-visual-diff.yml")
const probes = cfg.styles.map(style => ({
  name: style.name,
  selector: style.selectorReact || style.selector,
  props: style.props,
}))
```

### After: project JS decides its own spec shape

```js
// Loaded by verb field type: objectFromFile
async function compareSpec(spec, values) {
  const targets = lib.registry.targetsFromSpec(spec)
  for (const target of targets) {
    await lib.compareRegion.comparePage(target, values)
  }
}
```

Or, for ad-hoc scripts:

```js
const cvd = require("css-visual-diff")

async function main() {
  const browser = await cvd.browser()
  const page = await browser.page("http://localhost:7070/archive.html", {
    viewport: { width: 920, height: 1460 },
    waitMs: 1000,
  })

  const result = await page.inspectAll([
    { name: "content", selector: "[data-page='archive']", props: ["display", "gap"] },
  ], {
    outDir: "/tmp/archive-inspect",
    artifacts: "bundle",
  })

  await browser.close()
  return result
}
```

---

## Testing Strategy

### Go tests

Run all tests:

```bash
GOWORK=off go test ./...
```

Expect to delete or rewrite tests tied to removed behavior:

- run config discovery tests
- run command dry-run tests
- config loader tests
- inspect-config verb discovery tests
- config-driven mode tests

Keep and update tests for:

- `driver`
- `service` browser/page/inspect/catalog primitives
- `jsapi` browser/page/locator/probe/snapshot/diff/catalog APIs
- `verbcli` repository discovery and JS verb execution
- direct compare if retained

### Compile-time checks

Use import search as a hard gate:

```bash
rg -n "internal/cssvisualdiff/config" --glob '*.go'
```

This should be empty after the old package is deleted.

### CLI smoke tests

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs --help
GOWORK=off go run ./cmd/css-visual-diff verbs catalog inspect-page --help
```

`inspect-config` should no longer exist:

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs catalog inspect-config --help
# expected: unknown command or not found
```

### Pyxis smoke tests

From the Pyxis visual-diff directory:

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff
./userland/scripts/smoke-list-targets.sh
```

If servers are running, also run one compare smoke:

```bash
./userland/scripts/smoke-compare-spec-archive.sh
```

These validate that removing native configs did not break the JS-first project that matters.

---

## Risks and Mitigations

### Risk: deleting too much reusable code

Some config-driven files contain utility functions that might still be useful for JS-first compare or artifact generation.

**Mitigation:** do not delete solely by filename. Delete by dependency direction:

```text
Delete APIs that take *config.Config.
Keep or move helpers that operate on explicit URLs/selectors/files.
```

### Risk: confusing Pyxis YAML specs with native configs

Pyxis still uses YAML files, but they are not native `css-visual-diff` configs.

**Mitigation:** docs must say:

```text
Removed: native css-visual-diff YAML config and run command.
Kept: project-specific YAML/JSON data loaded by JS verbs via objectFromFile.
```

### Risk: breaking verb repository discovery

`.css-visual-diff.yml` can configure verb repositories. It is YAML, but it supports JS scriptability.

**Mitigation:** leave it alone for this ticket unless explicitly scoped in a follow-up. The user asked to remove YAML config/run complexity; the highest-value target is the native visual config runner.

### Risk: docs still mention removed workflows

Old docs/examples can mislead future engineers.

**Mitigation:** include docs/examples cleanup in the definition of done. Search for `run --config`, `loadConfig`, and `inspect-config` after code deletion.

---

## Alternatives Considered

### Alternative A: Keep native YAML but freeze it

This would avoid immediate deletion work but preserve the complexity. New features would still have to decide whether to support YAML config, JS API, or both. This conflicts with the explicit goal of simplifying around JS.

Rejected.

### Alternative B: Keep `cvd.loadConfig` only

This would remove `run --config` but keep old config loading for scripts. It still keeps the old schema alive and requires maintaining `config.Config` and `jsapi/config.go`.

Rejected, because the user said no backwards compatibility is required.

### Alternative C: Convert native configs into JS scripts automatically

A converter could read old YAML and generate JS. This would help migrations, but Pyxis is already migrated and no backwards compatibility is needed.

Rejected as unnecessary complexity.

### Alternative D: Remove all YAML support everywhere

This would include Pyxis `objectFromFile` YAML specs and `.css-visual-diff.yml` verb repo config. That would harm the JS-first workflow because project-specific specs are useful as data.

Rejected for this ticket. The goal is not "ban YAML as a file format"; the goal is "remove the native Go/YAML runner and schema complexity."

---

## Intern Implementation Notes

If you are implementing this as your first larger cleanup in this repository, follow these rules:

1. **Keep the code compiling after each phase.** Do not delete the config package first; move runtime types first.
2. **Use ripgrep constantly.** After each phase, search for the deleted symbol names.
3. **Prefer explicit JS primitives over Go orchestration.** If you feel tempted to add a new Go mode, ask whether it should instead be a JS verb.
4. **Do not preserve backwards compatibility shims.** The ticket explicitly says no backwards compatibility is needed.
5. **Do preserve Pyxis userland.** Pyxis is the relevant consumer. Its JS verbs and visual suite specs should keep working.
6. **When in doubt, ask whether the code knows about `config.Config`.** If yes, it probably belongs to the removal path.

---

## File Reference Map

### Core files to remove or change

| File | Action | Reason |
|---|---|---|
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/cmd/css-visual-diff/main.go` | Remove `run` command and config-driven inspect path. | Main CLI entrypoint for old workflow. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/config/config.go` | Delete after moving runtime types. | Native YAML schema and loader. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/config/config_test.go` | Delete. | Tests old YAML schema. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/runner/runner.go` | Delete. | Fixed mode runner over `*config.Config`. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/module.go` | Remove `cvd.loadConfig`; migrate target types. | JS API should not expose native config. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/config.go` | Delete. | Lowering code only for old config. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/dsl/scripts/catalog.js` | Remove `inspectConfig` verb. | Compatibility bridge for old config. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/prepare.go` | Migrate to config-free runtime types. | Prepare remains useful for JS pages. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/browser.go` | Migrate to `service.PageTarget`. | Browser loading remains useful. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/inspect.go` | Migrate metadata from `config.Viewport` to service runtime type. | Inspect remains useful for JS. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/catalog_service.go` | Migrate catalog target viewport type. | Catalog remains useful for JS. |

### Pyxis files that justify the decision

| File | Evidence |
|---|---|
| `/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/README.md` | States JS userland and `specs/*.visual.yml` are canonical; native `run` configs are retired. |
| `/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/specs/README.md` | Says future additions should prefer project-specific fields over native config shape. |
| `/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/verbs/pyxis-pages.js` | Defines `inspectSpec` and `compareSpec` using `objectFromFile`. |
| `/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/lib/registry.js` | Converts project specs into runtime target records. |
| `/home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland/scripts/refresh-spec-mirrors.py` | Keeps YAML specs and CommonJS mirrors in sync for JS runtime use. |

---

## Definition of Done

This ticket is complete when:

1. `css-visual-diff run` no longer exists.
2. Native YAML config loading no longer exists.
3. `cvd.loadConfig` no longer exists.
4. `verbs catalog inspect-config` no longer exists.
5. No Go package imports `internal/cssvisualdiff/config`.
6. Docs and examples no longer tell users to write native `css-visual-diff` YAML configs.
7. JS-first workflows still pass:
   - built-in verb help works,
   - JS API tests pass,
   - Pyxis `list-targets` smoke passes,
   - Pyxis `compare-spec` smoke passes when required servers are available.
8. Future feature docs, including overlay screenshot labels, can target the JS API first without mentioning native YAML support.

---

*Document version: 2026-04-28*  
*Ticket: remove-yaml-run*
