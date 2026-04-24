---
Title: Goja JavaScript API Analysis Design and Implementation Guide
Ticket: CSSVD-GOJA-JS-API
Status: active
Topics:
    - css-visual-diff
    - goja
    - javascript-api
    - visual-regression
    - catalog
    - automation
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/modes/inspect.go
      Note: Current single-side inspect implementation and artifact writer.
    - Path: internal/cssvisualdiff/modes/prepare.go
      Note: Current prepare hook implementation for script and direct-react-global.
    - Path: internal/cssvisualdiff/driver/chrome.go
      Note: Current chromedp browser/page abstraction.
    - Path: internal/cssvisualdiff/config/config.go
      Note: Current YAML schema for targets, sections, styles, output, and prepare.
    - Path: cmd/css-visual-diff/main.go
      Note: Current Cobra/Glazed CLI command wiring.
    - Path: internal/cssvisualdiff/dsl/host.go
      Note: Current embedded jsverbs host; scans only embedded scripts and registers generated commands at startup.
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: Current native Goja modules (`diff`, `report`) and the thin-adapter pattern already in this repository.
    - Path: internal/cssvisualdiff/dsl/scripts/compare.js
      Note: Current annotated `__verb__` script proving jsverbs command metadata already works for css-visual-diff.
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go
      Note: Reference implementation for repository discovery, embedded builtins, jsverbs scanning, require-loader setup, and duplicate verb detection.
    - Path: /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go
      Note: Reference implementation for lazy dynamic verbs command registration and custom runtime invoker wiring.
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/command.go
      Note: Upstream jsverbs command-description and pluggable invoker API.
    - Path: /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/model.go
      Note: Upstream jsverbs registry, scan options, verb model, sections, field metadata, and diagnostics.
ExternalSources: []
Summary: Intern-facing architecture and implementation guide for adding and productizing a Goja/jsverbs JavaScript API to css-visual-diff, including repository-scanned scripts exposed as CLI verbs.
LastUpdated: 2026-04-24T13:00:00-04:00
WhatFor: "Use this before implementing a Goja JavaScript scripting layer for css-visual-diff."
WhenToUse: "When designing or implementing programmable catalog workflows, batch inspection, selector preflight, or scriptable visual regression operations."
---

# Goja JavaScript API Analysis, Design, and Implementation Guide

This document designs a Goja-powered JavaScript API for `css-visual-diff`. It is written for a new intern who knows Go and JavaScript, but has not yet worked on this repository. The goal is not only to sketch an attractive API. The goal is to explain why the API should exist, how it fits into the existing Go code, what abstractions should be exposed to JavaScript, what should stay in Go, and how to implement the first useful version without turning the project into an unmaintainable scripting shell.

The immediate motivation comes from the Pyxis visual catalog work. We used `css-visual-diff` to extract prototype baselines from HTML files and compare them to Storybook/React output. The tool could do the job, but the workflow required many generated YAML files, shell loops, selector debugging, partial artifact inspection, and post-processing scripts. We reached the goal, but the path was slower and more manual than necessary. A scriptable API would let the operator describe a dynamic catalog workflow directly: start a browser once, load targets, prepare pages, preflight selectors, extract artifacts, write a manifest, and build a report.

A good JavaScript API should make the efficient workflow the natural workflow.

## 1. The problem this API solves

`css-visual-diff` currently works primarily through declarative YAML configs and CLI commands. A config says: load this original target, load this React target, inspect these sections and styles, write artifacts, and optionally run modes such as capture, css diff, pixel diff, or report generation. That is a good model for stable comparisons. It is easy to commit. It is easy to run in CI. It is easy to review as data.

But catalog authoring is not purely declarative. During Pyxis, we needed to do things like:

- generate many related targets from a small matrix of pages and variants,
- reuse repeated style-probe definitions,
- test selectors before paying for screenshots,
- skip missing selectors during authoring but fail on them during CI,
- run CSS-only passes before expensive PNG/HTML/inspect bundles,
- build a custom manifest and browsable index,
- record timing and failure information,
- reuse a browser across multiple targets,
- and branch based on what the DOM actually contains.

YAML can represent the final state, but it is awkward as a programming model. Shell scripts can orchestrate repeated CLI calls, but they do not know browser state. Go code can do everything, but each new workflow would require a new CLI mode or internal command.

A Goja scripting API gives us a middle layer:

```text
YAML config        → stable declarative comparisons
Go CLI modes       → batteries-included common workflows
Goja JS scripts    → programmable catalog and automation workflows
Go services        → reusable browser/artifact primitives
```

The API should not replace YAML. It should complement YAML. YAML is for stable definitions; JavaScript is for dynamic workflows.

## 2. Current system overview

Before designing the API, an intern should understand the current system.

```mermaid
flowchart TD
    CLI[cmd/css-visual-diff/main.go] --> Config[config.Load]
    Config --> Modes[internal/cssvisualdiff/modes]
    Modes --> Driver[internal/cssvisualdiff/driver]
    Driver --> Chrome[chromedp / headless Chrome]
    Modes --> Artifacts[PNG / CSS / HTML / JSON / Markdown]
```

The key files are:

| File | Current role |
|---|---|
| `cmd/css-visual-diff/main.go` | CLI commands, flags, config loading, mode dispatch. |
| `internal/cssvisualdiff/config/config.go` | YAML schema and validation. |
| `internal/cssvisualdiff/driver/chrome.go` | Thin wrapper around chromedp browser/page operations. |
| `internal/cssvisualdiff/modes/prepare.go` | Prepare hooks: `script` and `direct-react-global`. |
| `internal/cssvisualdiff/modes/inspect.go` | Single-side inspect command and bundle artifact writing. |
| `internal/cssvisualdiff/modes/capture.go` | Original-vs-react screenshot capture mode. |
| `internal/cssvisualdiff/modes/cssdiff.go` | Computed style extraction and CSS diffing. |
| `internal/cssvisualdiff/modes/pixeldiff.go` | Pixel-diff orchestration. |
| `internal/cssvisualdiff/modes/html_report.go` | Static HTML report generation. |

The current YAML schema has these major concepts:

```go
type Config struct {
    Metadata Metadata
    Original Target
    React    Target
    Sections []SectionSpec
    Styles   []StyleSpec
    Output   OutputSpec
    Modes    []string
}
```

A `Target` is a page to load:

```go
type Target struct {
    Name         string
    URL          string
    WaitMS       int
    Viewport     Viewport
    RootSelector string
    Prepare      *PrepareSpec
}
```

A `StyleSpec` is a selector used for computed CSS and style-probe artifacts:

```go
type StyleSpec struct {
    Name             string
    Selector         string
    SelectorOriginal string
    SelectorReact    string
    Props            []string
    IncludeBounds    bool
    Attributes       []string
    Report           []string
}
```

A `PrepareSpec` is a hook that modifies the page after navigation:

```go
type PrepareSpec struct {
    Type string // script or direct-react-global
    Script string
    ScriptFile string
    WaitFor string
    WaitForTimeoutMS int
    AfterWaitMS int
    Component string
    Props map[string]any
    RootSelector string
    Width int
    MinHeight int
    Background string
}
```

The Goja API should reuse these concepts but present them with JavaScript naming and workflow ergonomics.

## 3. Lessons from the Pyxis catalog work

The Pyxis work is the best concrete example of why this API should exist. We generated a prototype baseline catalog with:

| Category | Count |
|---|---:|
| Foundations/SystemPage configs | 1 |
| Public top-level page configs | 10 |
| Public widget/component configs | 18 |
| Total prototype configs | 29 |
| Total style probes | 165 |

The current shell/YAML workflow looks roughly like this:

```bash
scripts/11-generate-prototype-baseline-configs.mjs
scripts/06-run-prototype-baseline-sample.sh
scripts/07-run-prototype-baseline-full.sh
scripts/12-build-prototype-baseline-index.mjs
```

That worked, but it had several friction points.

### 3.1 Generated YAML became a substitute for a real programming API

We wrote JavaScript to generate YAML, then shell to run the YAML, then JavaScript to build an index from outputs. The generator had to encode domain knowledge such as page names, variants, selectors, probe lists, output paths, and artifact policies.

A Goja API could express that directly:

```js
for (const page of ["shows", "detail", "archive", "book", "about"]) {
  targets.push(publicPageTarget(page, "desktop"))
  targets.push(publicPageTarget(page, "mobile"))
}
```

Instead of generating YAML and then asking another command to interpret it.

### 3.2 Missing selectors were too expensive to discover

Before the recent fix, a missing selector could hang inside `chromedp.Screenshot` until the shell timeout killed the process. We patched `inspect` to preflight selector existence for selector-backed artifact formats. That was necessary, but the workflow would be even better if selector status were first-class data in a script:

```js
const preflight = await page.preflight(probes)
if (preflight.missing().length) {
  console.log(preflight.markdown())
}
```

### 3.3 The operator needed timing and control

When extraction felt slow, we had to inspect Go source and partial output directories to answer basic questions:

- Does `--all-styles` reload per style?
- Is the time spent in page load, prepare, screenshot, or CSS extraction?
- Did we hang on a selector or just process a large screenshot?

A scriptable API should expose timing per phase.

### 3.4 The correct workflow is multi-pass

During authoring, the efficient workflow is:

1. Generate or define targets.
2. Load and prepare each target.
3. Preflight selectors.
4. Run CSS-only or metadata-only checks.
5. Inspect a representative sample of PNGs.
6. Run full bundles.
7. Build index/report.

YAML can represent a target, but it does not naturally represent this multi-pass workflow.

## 4. Design principles for the Goja API

The API should follow these principles.

### Principle 1: Keep domain logic in Go services

The JavaScript module should not reimplement screenshotting, CSS extraction, report generation, or browser orchestration. Those should live in Go service packages. The Goja layer should adapt those services into a JavaScript-friendly interface.

Bad architecture:

```text
Goja Loader contains business logic, artifact layout, chromedp calls, and config parsing.
```

Good architecture:

```text
Go services implement browser and artifact operations.
Goja adapters decode JS options and call services.
JS scripts orchestrate workflows.
```

### Principle 2: Use JavaScript naming conventions

Go structs use names like `RootSelector`, `WaitMS`, and `MinHeight`. JavaScript should use lower camel case:

```js
{
  rootSelector: "#capture-root",
  waitMs: 1000,
  minHeight: 430
}
```

The adapter should translate between JS and Go shapes.

### Principle 3: Make selector status explicit

The API should have a `preflight()` method and structured selector status objects. Missing selectors are common during authoring and should not be exceptional unless the script chooses strict mode.

### Principle 4: Make artifacts declarative at the call site

The script should not manually call `screenshot`, `computedStyle`, `inspectDom`, and `writeJson` for every probe unless it wants low-level control. A high-level `inspectAll()` call should write the standard bundle.

### Principle 5: Support both strict and exploratory modes

During authoring:

```js
await page.inspectAll(probes, { failOnMissing: false })
```

During CI:

```js
await page.inspectAll(probes, { failOnMissing: true })
```

Same API, different policy.

## 5. Proposed JavaScript module

The module should be loaded as:

```js
const cvd = require("css-visual-diff")
```

Top-level exports:

```ts
type CVD = {
  browser(options?: BrowserOptions): Promise<Browser>
  loadConfig(path: string): Config
  writeConfig(path: string, config: Config): void
  targetFromConfig(config: Config, side: "original" | "react"): TargetSpec
  probesFromConfig(config: Config, options?: ProbeOptions): ProbeSpec[]
  catalog(options: CatalogOptions): Catalog
  ensureServer(options: StaticServerOptions): Promise<ServerHandle>
  glob(pattern: string): string[]
  mkdir(path: string): void
  readJson(path: string): any
  writeJson(path: string, value: any): void
  writeMarkdown(path: string, markdown: string): void
  parallel<T, R>(items: T[], options: ParallelOptions, fn: (item: T) => Promise<R>): Promise<R[]>
  timer(name?: string): Timer
  log: Logger
}
```

The first implementation does not need every export. The minimum viable API is:

```ts
browser()
page.prepare()
page.preflight()
page.inspectAll()
catalog()
```

Those five pieces would already make dynamic catalog scripts much easier.

## 6. Browser and Page API

The core object model should mirror the work operators do.

```mermaid
classDiagram
    class Browser {
      +page(url, options) Page
      +newPage(options) Page
      +close()
    }
    class Page {
      +goto(url)
      +setViewport(width, height)
      +prepare(spec)
      +selector(selector) SelectorStatus
      +preflight(probes) PreflightResult
      +inspect(probe, options) InspectResult
      +inspectAll(probes, options) InspectAllResult
      +close()
    }
    class Catalog {
      +artifactDir(slug) string
      +addResult(target, result)
      +addFailure(target, error)
      +writeManifest()
      +writeIndex()
    }
    Browser --> Page
    Page --> Catalog
```

Suggested API:

```ts
interface Browser {
  newPage(options?: PageOptions): Promise<Page>
  page(url: string, options?: PageOptions): Promise<Page>
  close(): Promise<void>
}

interface Page {
  goto(url: string, options?: GotoOptions): Promise<void>
  setViewport(width: number, height: number): Promise<void>
  wait(ms: number): Promise<void>
  waitFor(expr: string, options?: WaitOptions): Promise<void>
  eval<T = any>(js: string): Promise<T>
  prepare(spec: PrepareSpec): Promise<PrepareResult>
  selector(selector: string): Promise<SelectorStatus>
  preflight(probes: ProbeSpec[]): Promise<PreflightResult>
  screenshot(selector: string, path: string, options?: ScreenshotOptions): Promise<ScreenshotResult>
  computedStyle(selector: string, props?: string[], attrs?: string[]): Promise<StyleResult>
  inspectDom(selector: string, options?: InspectDomOptions): Promise<InspectDomResult>
  preparedHtml(selector?: string): Promise<string>
  inspect(probe: ProbeSpec, options: InspectOptions): Promise<InspectResult>
  inspectAll(probes: ProbeSpec[], options: InspectAllOptions): Promise<InspectAllResult>
  close(): Promise<void>
}
```

Example:

```js
const browser = await cvd.browser({ headless: true })
const page = await browser.page("http://localhost:7070/standalone/public/shows.html", {
  viewport: { width: 920, height: 1660 }
})

await page.prepare({
  type: "script",
  waitFor: "document.querySelector('#root > *:first-child')",
  script: `document.body.style.margin = '0'`,
  afterWaitMs: 250
})

const result = await page.inspectAll(probes, {
  outDir: "various/prototype-baseline/artifacts/public/shows",
  artifacts: "bundle"
})

await page.close()
await browser.close()
```

## 7. Prepare API

The JS API should support the current prepare types using JS-friendly names.

```ts
type PrepareSpec = ScriptPrepare | DirectReactGlobalPrepare | NoPrepare

type ScriptPrepare = {
  type: "script"
  waitFor?: string
  waitForTimeoutMs?: number
  script?: string
  scriptFile?: string
  afterWaitMs?: number
}

type DirectReactGlobalPrepare = {
  type: "directReactGlobal"
  waitFor?: string
  waitForTimeoutMs?: number
  component: string
  props?: any
  rootSelector: string
  width: number
  minHeight?: number
  background?: string
  afterWaitMs?: number
}
```

A public page prepare:

```js
await page.prepare({
  type: "script",
  waitFor: "document.querySelector('#root > *:first-child') && (!document.fonts || document.fonts.status === 'loaded')",
  script: `
    const root = document.querySelector('#root')
    if (root) {
      root.style.minHeight = '0px'
      root.style.height = 'auto'
    }
    document.body.style.margin = '0'
  `,
  afterWaitMs: 250
})
```

A direct prototype fixture prepare:

```js
await page.prepare({
  type: "directReactGlobal",
  waitFor: "window.React && window.ReactDOM && window.PPXCatalogShowTile",
  component: "PPXCatalogShowTile",
  rootSelector: "#capture-root",
  props: { index: 0, compact: false, width: 270 },
  width: 270,
  minHeight: 430,
  background: "#fff",
  afterWaitMs: 500
})
```

The first implementation should convert these specs to the existing Go `config.PrepareSpec` and call the existing `prepareTarget` service logic, not duplicate prepare code in the JS adapter.

## 8. Probe API

A probe is the script-facing form of a style/section selector.

```ts
type ProbeSpec = {
  name: string
  selector: string
  props?: string[]
  attrs?: string[]
  includeBounds?: boolean
  screenshot?: boolean | ScreenshotOptions
  css?: boolean
  html?: boolean | "self" | "root"
  inspectJson?: boolean
  required?: boolean
  kind?: "root" | "section" | "style" | "component" | "page"
}
```

Example:

```js
const showGridProbe = {
  name: "show-grid",
  selector: "#root > div > main > :nth-child(3)",
  props: ["display", "width", "height", "gap", "grid-template-columns", "margin", "padding"],
  screenshot: true,
  css: true,
  inspectJson: true,
  required: true,
  kind: "section"
}
```

Probe factories are where JavaScript becomes more elegant than YAML:

```js
function cardProbe(name, index) {
  return {
    name,
    selector: `#capture-root > div > div:nth-of-type(2) > div:nth-child(${index})`,
    props: ["display", "width", "height", "padding", "background-color", "border", "border-radius", "box-shadow"],
    screenshot: true,
    css: true,
    inspectJson: true,
    kind: "section"
  }
}

const foundationProbes = [
  cardProbe("color-card", 1),
  cardProbe("typography-card", 2),
  cardProbe("badges-tags-card", 3),
  cardProbe("buttons-card", 4),
  cardProbe("form-fields-card", 5),
]
```

This is clearer and less error-prone than writing or generating repeated YAML blocks.

## 9. Selector preflight API

Selector preflight should be one of the main selling points of the API.

```ts
type SelectorStatus = {
  name?: string
  selector: string
  exists: boolean
  visible: boolean
  bounds?: Bounds
  textStart?: string
  error?: string
}

type PreflightResult = {
  statuses: SelectorStatus[]
  ok(): ProbeSpec[]
  missing(): SelectorStatus[]
  hidden(): SelectorStatus[]
  assertAll(): void
  markdown(): string
}
```

Usage:

```js
const preflight = await page.preflight(probes)

if (preflight.missing().length) {
  cvd.log.warn(preflight.markdown())
}

const validProbes = preflight.ok()
```

The output should be readable:

```text
| Status | Name | Selector | Bounds |
|---|---|---|---|
| ok | nav | #root > div > header | 920×60 |
| ok | page-header | #root > div > main > :first-child | 856×101 |
| missing | first-show-tile | #root > div > main > :nth-child(2) > div:first-child | — |
```

This would have made the Pyxis selector problems immediately obvious.

## 10. Inspect and artifact API

The API should expose both low-level artifact functions and high-level bundle functions.

Low-level:

```js
await page.screenshot(selector, "out.png")
const css = await page.computedStyle(selector, ["display", "gap"])
const dom = await page.inspectDom(selector)
const html = await page.preparedHtml(selector)
```

High-level:

```js
await page.inspect(probe, {
  outDir: "various/prototype-baseline/sample/show-grid",
  artifacts: "bundle",
  preparedHtml: "per-probe",
  failOnMissing: true
})
```

Batch:

```js
await page.inspectAll(probes, {
  outDir: "various/prototype-baseline/sample/prototype-public-shows",
  artifacts: ["png", "css.md", "css.json", "inspect.json", "metadata.json"],
  preparedHtml: "root-once",
  failOnMissing: false
})
```

Suggested types:

```ts
type ArtifactKind = "png" | "css.md" | "css.json" | "inspect.json" | "metadata.json" | "html"

type InspectOptions = {
  outDir: string
  artifacts?: ArtifactKind[] | "bundle" | "css-only" | "png-only"
  preparedHtml?: "per-probe" | "root-once" | false
  rootSelector?: string
  failOnMissing?: boolean
  overwrite?: boolean
}

type InspectResult = {
  name: string
  selector: string
  exists: boolean
  screenshot?: string
  cssJson?: string
  cssMarkdown?: string
  html?: string
  inspectJson?: string
  metadata?: string
  timing: Timing
  error?: string
}

type InspectAllResult = {
  outputDir: string
  results: InspectResult[]
  ok: InspectResult[]
  failed: InspectResult[]
  timing: Timing
}
```

## 11. Catalog API

The catalog API should manage manifest and report generation, and it should be implemented on the Go side as a real service rather than as JavaScript-only helper code. JavaScript scripts should call `cvd.catalog(...)`, but the object behind it should be backed by Go structs and writers so manifests, indexes, path normalization, timestamps, schema versions, and future report formats stay consistent across CLI modes and JS verbs.

Recommended split:

```text
internal/cssvisualdiff/service/catalog_service.go   # owns manifest/index data model and writers
internal/cssvisualdiff/dsl/catalog_adapter.go       # adapts Go catalog service to goja values/promises
```

The JS layer may provide small ergonomic wrappers, but the durable catalog model belongs in Go.

```ts
type CatalogOptions = {
  title: string
  outDir: string
  artifactRoot?: string
  indexName?: string
}

interface Catalog {
  artifactDir(slug: string): string
  addTarget(target: TargetSpec): void
  addResult(target: TargetSpec, result: InspectAllResult): void
  addFailure(target: TargetSpec, error: any): void
  recordPreflight(target: TargetSpec, preflight: PreflightResult): void
  summary(): CatalogSummary
  writeManifest(): Promise<void>
  writeIndex(options?: IndexOptions): Promise<void>
}
```

Example:

```js
const catalog = cvd.catalog({
  title: "Pyxis Prototype Baseline Catalog",
  outDir: `${ticket}/various/prototype-baseline`,
  artifactRoot: "artifacts"
})

const outDir = catalog.artifactDir("prototype-public-shows")
const result = await page.inspectAll(probes, { outDir, artifacts: "bundle" })
catalog.addResult(target, result)
await catalog.writeManifest()
await catalog.writeIndex()
```

This keeps report logic out of ad hoc scripts.

## 12. YAML interop

The JS API should load existing YAML configs. We do not want two worlds that cannot talk to each other.

```js
const cfg = cvd.loadConfig("sources/prototype-configs/prototype-public-shows.css-visual-diff.yml")
const target = cvd.targetFromConfig(cfg, "original")
const probes = cvd.probesFromConfig(cfg, { source: "styles" })

const browser = await cvd.browser()
const page = await browser.page(target.url, { viewport: target.viewport })
await page.prepare(target.prepare)
await page.inspectAll(probes, { outDir: "/tmp/debug", artifacts: "bundle" })
```

This gives us a migration path:

1. Existing YAML configs continue working.
2. JS scripts can consume YAML for dynamic workflows.
3. Eventually JS scripts can generate YAML if we still want committed declarations.

## 13. Example: Pyxis catalog in the proposed API

This is a realistic script for the Pyxis workflow.

```js
const cvd = require("css-visual-diff")

const repo = "/home/manuel/code/wesen/2026-04-23--pyxis"
const ticket = `${repo}/ttmp/2026/04/23/PYXIS-STORYBOOK-CATALOG--build-storybook-screenshot-and-css-catalog-for-atoms-molecules-and-public-components`
const baseUrl = "http://localhost:7070"

const cardProps = ["display", "width", "height", "padding", "background-color", "border", "border-radius", "box-shadow"]

function foundationCard(name, index) {
  return {
    name,
    selector: `#capture-root > div > div:nth-of-type(2) > div:nth-child(${index})`,
    props: cardProps,
    screenshot: true,
    css: true,
    inspectJson: true,
  }
}

function foundationsTarget() {
  return {
    slug: "prototype-foundations-system",
    kind: "foundations",
    url: `${baseUrl}/Pyxis%20Full%20App.html`,
    viewport: { width: 1440, height: 2870 },
    prepare: {
      type: "directReactGlobal",
      waitFor: "window.React && window.ReactDOM && window.SystemPage",
      component: "SystemPage",
      rootSelector: "#capture-root",
      props: {},
      width: 1240,
      minHeight: 2650,
      background: "#F3F1EB",
      afterWaitMs: 500,
    },
    probes: [
      { name: "full-system-page", selector: "#capture-root", props: ["display", "width", "height", "padding", "font-family", "background-color"] },
      foundationCard("color-card", 1),
      foundationCard("typography-card", 2),
      foundationCard("badges-tags-card", 3),
      foundationCard("buttons-card", 4),
      foundationCard("form-fields-card", 5),
      foundationCard("stats-card", 6),
      foundationCard("icons-card", 7),
      foundationCard("navigation-card", 9),
      foundationCard("empty-state-card", 11),
    ]
  }
}

async function main() {
  await cvd.ensureServer({
    dir: `${repo}/prototype-design`,
    port: 7070,
    probe: `${baseUrl}/Pyxis%20Public%20Site.html`
  })

  const targets = [
    foundationsTarget(),
    ...publicPageTargets(),
    ...publicWidgetTargets(),
  ]

  const browser = await cvd.browser()
  const catalog = cvd.catalog({
    title: "Pyxis Prototype Baseline Catalog",
    outDir: `${ticket}/various/prototype-baseline`
  })

  for (const target of targets) {
    const timer = cvd.timer(target.slug)
    const page = await browser.page(target.url, { viewport: target.viewport })

    await page.prepare(target.prepare)

    const preflight = await page.preflight(target.probes)
    catalog.recordPreflight(target, preflight)

    if (preflight.missing().length) {
      cvd.log.warn(preflight.markdown())
    }

    const result = await page.inspectAll(preflight.ok(), {
      outDir: catalog.artifactDir(target.slug),
      artifacts: "bundle",
      preparedHtml: "root-once",
      failOnMissing: false
    })

    catalog.addResult(target, result)
    cvd.log.info(`${target.slug}: ${timer.stop()}`)
    await page.close()
  }

  await catalog.writeManifest()
  await catalog.writeIndex()
  await browser.close()
}

main()
```

Notice what is missing from this script: no shell loops, no generated YAML as an intermediate format, no guessing whether selectors exist, and no separate index builder.

## 14. Error model

The API should support two policies.

Exploratory authoring:

```js
const result = await page.inspectAll(probes, { failOnMissing: false })
for (const failure of result.failed) {
  console.log(failure.name, failure.error)
}
```

Strict CI:

```js
await page.inspectAll(probes, { failOnMissing: true })
```

Missing selector errors should be structured, and the API may either return them in exploratory result objects or throw them in strict paths. Define JS-visible error types early so scripts can distinguish authoring mistakes from browser/process failures.

```ts
class CvdError extends Error {
  code: string
  cause?: unknown
  details?: any
}

class SelectorError extends CvdError {
  code: "selector_missing" | "selector_hidden" | "selector_invalid"
  name: string
  selector: string
  source: "style" | "section" | "flag" | "root" | "probe"
  url: string
  targetName: string
  hint?: string
}

class PrepareError extends CvdError {
  code: "prepare_failed" | "prepare_timeout" | "prepare_invalid"
  targetName: string
  prepareType: string
}

class BrowserError extends CvdError {
  code: "browser_start_failed" | "navigation_failed" | "cdp_failed" | "timeout"
}

class ArtifactError extends CvdError {
  code: "artifact_write_failed" | "artifact_exists" | "artifact_invalid"
  path?: string
  artifact?: string
}
```

Implementation note: Go adapters should throw with `vm.NewGoError(err)` or a helper that constructs one of the JS error classes and attaches `code`/`details`. Do not throw anonymous strings. For batch APIs, use options such as `failOnMissing` and `failFast` to decide whether structured failures are collected or thrown.

Example message:

```text
style "first-show-tile" selector did not match: #root > div > main > :nth-child(2) > div:first-child
Hint: run page.preparedHtml('#root') and inspect the children under main.
```

The important principle is that ordinary authoring mistakes should produce precise errors, not browser timeouts.

## 15. Implementation architecture

The Go implementation should be split into service packages and Goja adapters.

Recommended structure:

```text
internal/cssvisualdiff/service/
  browser_service.go
  inspect_service.go
  prepare_service.go
  preflight_service.go
  catalog_service.go
  artifact_service.go

internal/cssvisualdiff/js/
  module.go
  browser_adapter.go
  page_adapter.go
  catalog_adapter.go
  codecs.go
  promises.go
```

The service layer should not import Goja. It should define ordinary Go types and methods:

```go
type BrowserService interface {
    NewPage(ctx context.Context, opts PageOptions) (*PageService, error)
    Close() error
}

type PageService interface {
    Goto(ctx context.Context, url string) error
    SetViewport(ctx context.Context, viewport Viewport) error
    Prepare(ctx context.Context, spec PrepareSpec) error
    Preflight(ctx context.Context, probes []ProbeSpec) ([]SelectorStatus, error)
    InspectAll(ctx context.Context, probes []ProbeSpec, opts InspectOptions) (InspectAllResult, error)
    Close() error
}
```

The Goja adapter should only:

- decode JavaScript objects into Go option structs,
- call service methods,
- convert Go results into JavaScript objects,
- expose module functions through `require("css-visual-diff")`,
- provide clear thrown errors or result objects.

A simplified adapter shape:

```go
type Module struct{}

func (m *Module) Name() string { return "css-visual-diff" }

func (m *Module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    exports.Set("browser", func(call goja.FunctionCall) goja.Value {
        opts := decodeBrowserOptions(vm, call.Argument(0))
        browser, err := service.NewBrowser(opts)
        if err != nil {
            panic(vm.ToValue(err.Error()))
        }
        return wrapBrowser(vm, browser)
    })

    exports.Set("catalog", func(call goja.FunctionCall) goja.Value {
        opts := decodeCatalogOptions(vm, call.Argument(0))
        catalog := service.NewCatalog(opts)
        return wrapCatalog(vm, catalog)
    })
}
```

The adapter should not contain artifact writing logic. Artifact writing belongs in `artifact_service.go` or the existing modes refactored into reusable services.

## 16. Refactoring needed before implementation

The current `modes` package mixes reusable operations with mode orchestration. That is normal for a CLI-first tool, but a scripting API needs reusable service functions.

For example, `inspect.go` currently has useful pieces:

- request building,
- page navigation,
- prepare call,
- artifact writing,
- index writing.

A service extraction could produce:

```go
func BuildInspectRequests(cfg *config.Config, opts InspectOptions) ([]InspectRequest, error)
func InspectPreparedPage(ctx context.Context, page *driver.Page, target config.Target, requests []InspectRequest, opts ArtifactOptions) (InspectAllResult, error)
func WriteInspectArtifact(ctx context.Context, page *driver.Page, req InspectRequest, opts ArtifactOptions) (InspectResult, error)
func PreflightSelectors(ctx context.Context, page *driver.Page, requests []InspectRequest) ([]SelectorStatus, error)
```

That refactor would benefit both CLI and Goja. The CLI would call the service. The JS adapter would call the same service.

## 17. Minimal implementation plan

The first implementation should be small and useful. Do not implement everything in this design at once.

### Phase 1: service extraction

- Extract selector status/preflight logic from `inspect.go` into a reusable service function.
- Extract inspect artifact writing into a reusable function that accepts a prepared page.
- Keep CLI behavior unchanged.
- Add service-level tests.

### Phase 2: Goja module skeleton

- Add a native module package, for example `internal/cssvisualdiff/js`.
- Expose `require("css-visual-diff")` with a small set of functions.
- Add a runtime integration test that can load the module.

### Phase 3: browser/page wrappers

Expose:

```js
const browser = cvd.browser()
const page = browser.page(url, { viewport })
await page.prepare(spec)
await page.preflight(probes)
await page.inspectAll(probes, options)
```

Promises should be part of the API from the first implementation. Upstream `go-go-goja` already supports waiting for Promises returned by jsverb functions through `registry.InvokeInRuntime(...)`; native modules that create Promises must settle them through the runtime owner thread. Even if the underlying chromedp operation blocks in Go, the JavaScript contract should be Promise-based so scripts do not need a later migration.

Implementation rule: every browser/page/catalog operation that can touch I/O, CDP, files, or timers returns a Promise in JavaScript. Pure object builders can stay synchronous.

### Phase 4: catalog helper

Add:

```js
const catalog = cvd.catalog({ title, outDir })
catalog.artifactDir(slug)
catalog.addResult(target, result)
catalog.writeManifest()
catalog.writeIndex()
```

### Phase 5: Pyxis-like example script

Add an example under:

```text
examples/scripts/pyxis-prototype-catalog.js
```

The example should be runnable against local prototype HTML and should demonstrate:

- a direct React global target,
- a standalone page target,
- preflight,
- `inspectAll`,
- catalog index writing.

## 18. Testing strategy

Tests should exist at three levels.

### Service tests

Pure Go tests should validate selector preflight, artifact option decoding, and result structures without Goja.

```go
func TestPreflightSelectorsReportsMissing(t *testing.T) { ... }
func TestInspectAllSkipsMissingWhenConfigured(t *testing.T) { ... }
```

### Goja module loading tests

A runtime test should verify module loading:

```go
func TestGojaModuleLoads(t *testing.T) {
    vm := newRuntimeWithModules()
    _, err := vm.RunString(`
      const cvd = require("css-visual-diff")
      if (!cvd.browser) throw new Error("missing browser")
    `)
    require.NoError(t, err)
}
```

### Integration tests with a tiny HTTP page

Use a local test page:

```html
<div id="root"><button class="primary">Click</button></div>
```

Then run JS:

```js
const browser = cvd.browser()
const page = browser.page(testUrl, { viewport: { width: 400, height: 300 } })
const preflight = page.preflight([{ name: "button", selector: "button.primary" }])
if (preflight.missing().length) throw new Error("button missing")
page.inspectAll(preflight.ok(), { outDir, artifacts: "css-only" })
```

Assert that output files exist and contain expected values.

## 19. Documentation deliverables

The implementation should include documentation in the repository, not only in this ticket.

Recommended docs:

```text
docs/js-api.md
docs/js-api-catalog-workflows.md
examples/scripts/README.md
examples/scripts/pyxis-prototype-catalog.js
```

`docs/js-api.md` should include:

- module loading,
- Browser/Page/Catalog object references,
- prepare specs,
- probe specs,
- inspect options,
- error model.

`docs/js-api-catalog-workflows.md` should include:

- selector preflight workflow,
- CSS-only pass workflow,
- full artifact pass workflow,
- batch/concurrency guidance,
- when to use YAML vs JS.

The documentation should explicitly answer:

> Does `inspectAll` reload for every probe?

The answer should be no. A page is loaded/prepared once per target unless the script chooses otherwise.

## 20. Risks and guardrails

### Risk: turning JS into an untyped second implementation

Guardrail: keep core logic in Go services. JS should orchestrate, not implement browser internals.

### Risk: too much API at once

Guardrail: implement the minimum useful set first: `browser`, `page.prepare`, `page.preflight`, `page.inspectAll`, and `catalog`.

### Risk: ambiguous async behavior

Guardrail: Promise support is required from the beginning. Every I/O, CDP, browser, file, and catalog-write operation should return a real Promise, and native module code must settle those promises on the go-go-goja runtime owner thread. Do not ship a synchronous API shape that would need to be broken later.

### Risk: scripts become unreproducible

Guardrail: scripts should write manifests recording input URLs, selectors, viewports, git hashes if available, and tool version.

### Risk: JS API and YAML drift apart

Guardrail: maintain YAML interop helpers and shared Go structs for decode/encode where possible.

## 21. Acceptance criteria

A first successful implementation should satisfy these criteria:

1. A Goja runtime can `require("css-visual-diff")`.
2. A JS script can open a browser and page.
3. A JS script can prepare a page using `script` and `directReactGlobal` style specs.
4. A JS script can preflight a list of probes and receive structured missing/ok statuses.
5. A JS script can write at least CSS-only artifacts for a list of probes.
6. A JS script can write full bundle artifacts for a list of probes.
7. A JS script can write a catalog manifest and basic index.
8. Existing CLI behavior and tests continue to pass.
9. Documentation includes one complete runnable example.
10. Missing selectors never look like timeouts; they are structured status or clear errors.

## 22. Research update: what must change after studying the current implementation and nearby jsverbs systems

The original version of this document was directionally correct about the desired **Browser / Page / Probe / Catalog** abstraction, but it under-described an important implementation reality: `css-visual-diff` already has a first-generation Goja/jsverbs surface. The implementation plan should therefore stop sounding like a greenfield Goja feature and instead describe a **productization and expansion** of the existing `internal/cssvisualdiff/dsl` package.

The research pass looked at three concrete systems:

1. the current `css-visual-diff` codebase,
2. the newer `js-discord-bot` helper-verb design/diary,
3. the implemented loupedeck jsverbs CLI cutover and upstream `go-go-goja/pkg/jsverbs` APIs.

The main conclusion is:

```text
Do not build a separate ad hoc script runner.
Promote the existing embedded dsl/jsverbs host into a repository-scanned, lazy `verbs` command tree,
and add the Browser/Page/Catalog native module surface behind those verbs.
```

### 22.1 What already exists and should be preserved

The current repository already has these working pieces:

| Existing file | What it proves | Keep / change |
|---|---|---|
| `internal/cssvisualdiff/dsl/host.go` | `jsverbs.ScanFS(...)`, shared sections, `engine.NewBuilder()`, caller-owned runtime invoker, and `registry.InvokeInRuntime(...)` already work. | Keep the runtime/invoker idea, but generalize scanning beyond embedded scripts. |
| `internal/cssvisualdiff/dsl/registrar.go` | Native modules can expose Go services to JS. Current modules are `diff` and `report`. | Keep the thin native-module adapter pattern, but add a real `css-visual-diff` module. |
| `internal/cssvisualdiff/dsl/scripts/compare.js` | Annotated `__verb__` scripts already become CLI commands with flags and shared sections. | Keep as a starter built-in verb, but move user-facing verbs under a `verbs` subtree. |
| `cmd/css-visual-diff/main.go` | The root command currently builds the script host eagerly and adds generated script commands directly to the root command. | Replace eager root-level injection with lazy dynamic command registration. |

This means the implementation plan should remove any wording that suggests we first need to prove Goja, `require()`, or `__verb__` scanning in this repository. That proof already exists. The new work is to make it scalable, discoverable, repository-backed, and powerful enough for visual catalog workflows.

### 22.2 What should be updated in the design

#### Update 1: distinguish two JavaScript surfaces

The document should describe two related but different JavaScript surfaces:

1. **Native module API**: `require("css-visual-diff")`, used by scripts to open browsers, prepare pages, preflight probes, inspect artifacts, and write catalogs.
2. **CLI verb API**: annotated `__verb__(...)` scripts scanned from embedded and filesystem repositories and exposed as `css-visual-diff verbs ...` commands with Glazed/Cobra flags.

The old design focuses almost entirely on the native module API. The missing half is the CLI product surface that lets users place workflow scripts in repositories and invoke them without writing Go or touching the main command tree.

Recommended mental model:

```text
JavaScript script file
  declares __verb__("catalog-page", { fields: ... })
  calls require("css-visual-diff")
        ↓
jsverbs scanner
  extracts command metadata and flags
        ↓
css-visual-diff verbs catalog page --url ... --out-dir ...
  creates a css-visual-diff-owned Goja runtime
  registers native modules
  invokes the selected JS function
```

#### Update 2: make repository scanning a first-class phase

The design currently says to add example scripts under `examples/scripts`, but it does not specify how the CLI discovers user scripts. That should be expanded into a full repository model borrowed from loupedeck and the Discord helper-verbs design.

Recommended command shape:

```text
css-visual-diff verbs --help
css-visual-diff verbs list
css-visual-diff verbs catalog inspect-page --url http://localhost:7070 --selector '#root' --out-dir ./artifacts
css-visual-diff verbs catalog pyxis-baseline --repository /path/to/project --base-url http://localhost:7070 --out-dir ./various/prototype-baseline
```

Recommended source layout:

```text
css-visual-diff repo:
  internal/cssvisualdiff/dsl/scripts/        # embedded built-in verbs, current location

user/project repo:
  css-visual-diff/verbs/                     # read-only / normal workflow verbs
    catalog/
      inspect-page.js
      build-baseline.js
      compare-storybook.js
  css-visual-diff/verbs-rw/                  # optional mutating verbs if ever needed
```

Recommended repository sources, in precedence order:

1. embedded built-in repository,
2. app config repositories,
3. environment repositories, for example `CSS_VISUAL_DIFF_VERB_REPOSITORIES`,
4. repeated CLI `--repository` or `--verb-repository` flags.

Use the loupedeck implementation as the concrete pattern: normalize paths, dedupe repositories, scan embedded and filesystem repositories with `jsverbs.ScanFS` / `jsverbs.ScanDir`, reject duplicate full verb paths, and set `require()` global folders so scripts can use relative helpers.

#### Update 3: use a lazy `verbs` subtree, not eager root command injection

The current `main.go` eagerly builds `dsl.NewHost()`, calls `host.Commands()`, and injects generated commands directly into the root command. That was good for a first prototype, but it has three product problems:

- repository scanning cost is paid even when the user only asks for `run`, `inspect`, or `--help`,
- generated script commands are mixed with built-in commands at the top level,
- filesystem scan errors can break unrelated CLI usage.

The updated design should require a lazy subtree:

```go
func NewLazyVerbsCommand() *cobra.Command {
    return &cobra.Command{
        Use:                "verbs",
        Short:              "Run annotated css-visual-diff workflow verbs",
        DisableFlagParsing: true,
        Args:               cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            bootstrap, err := verbcli.DiscoverBootstrapFromCommand(cmd)
            if err != nil { return err }
            resolved, err := verbcli.NewCommand(bootstrap)
            if err != nil { return err }
            adoptHelpAndOutput(cmd, resolved)
            resolved.SetArgs(args)
            return resolved.ExecuteContext(cmd.Context())
        },
    }
}
```

This mirrors the loupedeck and Discord helper-verbs direction and makes dynamic verbs feel like a real product namespace.

#### Update 4: replace the greenfield `internal/cssvisualdiff/js` package name with a clearer split

The current document proposes:

```text
internal/cssvisualdiff/js/
```

After reviewing the current code, that is too vague and conflicts with the existing `dsl` package. A better split is:

```text
internal/cssvisualdiff/dsl/              # reusable JS runtime/module host code
  codec.go
  registrar.go                          # grows require("css-visual-diff") or delegates to module package
  sections.go
  scripts/compare.js                    # embedded built-ins

internal/cssvisualdiff/verbcli/          # CLI product layer for repository-scanned verbs
  bootstrap.go                          # repositories, config/env/CLI discovery, embedded repo
  command.go                            # lazy/dynamic Cobra command tree
  invoker.go                            # custom jsverbs invoker and output handling
  runtime_factory.go                    # creates go-go-goja runtime per invocation

internal/cssvisualdiff/service/          # Go service layer with no goja dependency
  browser_service.go
  inspect_service.go
  prepare_service.go
  preflight_service.go
  catalog_service.go
```

If the team prefers keeping all JS code under `dsl`, then `verbcli` can be `internal/cssvisualdiff/dsl/verbcli`, but the service/adapters/CLI boundaries should remain explicit.

#### Update 5: describe jsverbs fields as the flag API for catalog scripts

The design should show how catalog workflow scripts expose flags. Example:

```js
__package__({
  name: "catalog",
  parents: ["catalog"],
  short: "Programmable visual catalog workflows"
});

__section__("target", {
  title: "Target",
  fields: {
    url: { type: "string", required: true, help: "Page URL to inspect" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation" },
    width: { type: "int", default: 1280, help: "Viewport width" },
    height: { type: "int", default: 720, help: "Viewport height" }
  }
});

__section__("output", {
  title: "Output",
  fields: {
    outDir: { type: "string", required: true, help: "Artifact directory" },
    artifacts: { type: "choice", choices: ["css-only", "png-only", "bundle"], default: "bundle" },
    failOnMissing: { type: "bool", default: false }
  }
});

function inspectPage(target, output, selector) {
  const cvd = require("css-visual-diff");
  const browser = cvd.browser();
  const page = browser.page(target.url, {
    viewport: { width: target.width, height: target.height },
    waitMs: target.waitMs,
  });

  const probes = [{ name: "selected", selector }];
  const preflight = page.preflight(probes);
  if (output.failOnMissing) preflight.assertAll();

  const result = page.inspectAll(preflight.ok(), {
    outDir: output.outDir,
    artifacts: output.artifacts,
    failOnMissing: output.failOnMissing,
  });

  browser.close();
  return result.summary || result;
}

__verb__("inspect-page", {
  short: "Inspect one page selector and write css-visual-diff artifacts",
  fields: {
    target: { bind: "target" },
    output: { bind: "output" },
    selector: { argument: true, required: true, help: "CSS selector to inspect" }
  }
});
```

That scanned script should become an operator command like:

```bash
css-visual-diff verbs catalog inspect-page \
  --url http://localhost:7070/standalone/public/shows.html \
  --width 920 \
  --height 1660 \
  --out-dir various/prototype-baseline/artifacts/public-shows \
  '#root > div > main'
```

This is the missing bridge between the JS API design and “exposed as CLI verbs with flags that can easily be invoked”.

### 22.3 What should be expanded

#### Expand service extraction around inspect, preflight, and prepared pages

The current `modes.Inspect(...)` owns the entire operation: load config, create browser, create page, navigate, prepare, build requests, write artifacts, write index. For scripts, that is too coarse. Extract the reusable pieces so CLI modes and JS verbs share one implementation.

Add service functions with these contracts:

```go
type BrowserService struct { ... }
type PageService struct { ... }

type ProbeSpec struct {
    Name       string
    Selector   string
    Props      []string
    Attributes []string
    Source     string
    Required   bool
}

type SelectorStatus struct {
    Name     string
    Selector string
    Exists   bool
    Visible  bool
    Bounds   *Bounds
    Error    string
}

type InspectAllOptions struct {
    OutDir        string
    Artifacts     []string
    PreparedHTML  string
    FailOnMissing bool
    Overwrite     bool
}

func LoadAndPreparePage(ctx context.Context, browser *driver.Browser, target config.Target) (*driver.Page, error)
func PreparePage(ctx context.Context, page *driver.Page, target config.Target, prepare *config.PrepareSpec) error
func PreflightProbes(ctx context.Context, page *driver.Page, probes []ProbeSpec) ([]SelectorStatus, error)
func InspectPreparedPage(ctx context.Context, page *driver.Page, target config.Target, probes []ProbeSpec, opts InspectAllOptions) (InspectAllResult, error)
```

The design should explicitly say that the existing `modes.Inspect(...)` becomes a thin caller of these services rather than staying the only place artifact writing works.

#### Expand native module scope from `diff`/`report` to `css-visual-diff`

Keep `require("diff")` and `require("report")` as compatibility or low-level modules if useful, but add a coherent public module:

```js
const cvd = require("css-visual-diff");
```

Minimum exports:

- `browser(options?)`,
- `loadConfig(path)`,
- `targetFromConfig(config, side)`,
- `probesFromConfig(config, options?)`,
- `catalog(options)`,
- `mkdir/readJson/writeJson/writeMarkdown` if scripts need convenience IO.

This keeps workflow scripts readable and avoids forcing users to know the internal `diff` and `report` module split.

#### Expand output handling for generated commands

The implementation should decide how a script return value maps to CLI output. Reuse upstream jsverbs behavior:

- `output: "glaze"`: objects become rows, arrays become many rows, primitives become `{ value: ... }`,
- `output: "text"`: strings and bytes are written directly, other values fall back to pretty JSON.

For catalog verbs, default to structured rows for summaries and write large artifacts to files. Do not print huge manifests or binary-ish data to stdout.

### 22.4 What should be removed or de-emphasized

#### Remove the assumption that Promise support is unavailable

The earlier plan says “If the project does not yet have Promise support wired into Goja, start with synchronous calls.” That is now rejected. Upstream `go-go-goja` documents jsverbs Promise handling: if a verb returns a `Promise`, `InvokeInRuntime(...)` waits for fulfillment/rejection through the runtime owner. The updated design should require Promise-returning JS methods from the first implementation.

Better wording:

> Browser/page/catalog operations return Promises from day one. The Go side may block internally on chromedp or filesystem work, but the JavaScript contract is asynchronous and any native-module Promise must settle through the go-go-goja runtime owner thread.

#### Remove root-level generated command injection as the target product shape

The current implementation adds generated script commands directly to the root. The design should call that a prototype only. The product shape should be:

```text
css-visual-diff verbs ...
```

Keep the built-in `run`, `inspect`, `compare`, `llm-review`, and single-artifact commands as stable root commands. Keep dynamic user scripts under `verbs`.

#### Remove broad generic helpers from the MVP

The top-level export list in the original design includes `parallel`, `glob`, server helpers, timers, and file helpers. Those are useful, but they are not necessary for the first implementation. The revised MVP should be stricter:

1. repository-scanned jsverbs command tree,
2. `require("css-visual-diff")`,
3. browser/page prepare/preflight/inspectAll,
4. catalog manifest/index,
5. YAML interop helpers.

Add generic convenience helpers only after real catalog scripts prove they need them.

### 22.5 What should be improved in the current implementation before adding more API

#### Improvement 1: stop writing generated artifacts under `internal/cssvisualdiff/dsl`

A repository scan found many `css-visual-diff-compare-*` PNG artifact directories under `internal/cssvisualdiff/dsl`. That is a code hygiene problem: test/manual outputs should not live beside runtime source and embedded scripts.

The implementation should:

- default script output directories to `os.MkdirTemp(...)` or a configurable output root,
- add ignore rules for generated `css-visual-diff-compare-*` directories,
- move/remove existing generated artifacts from the source package if they are not intentional fixtures,
- make tests use `t.TempDir()`.

#### Improvement 2: make scan errors local to the `verbs` subtree

With a lazy `verbs` command, malformed user scripts should break `css-visual-diff verbs ...`, not `css-visual-diff inspect ...` or `css-visual-diff run ...`. This is especially important once filesystem repositories are supported.

#### Improvement 3: add duplicate command path detection

Repository-scanned verbs need deterministic collision handling. If two scripts declare the same full path, fail with an error that includes both sources:

```text
duplicate jsverb path "catalog inspect-page" from builtin:scripts/inspect-page.js and /repo/css-visual-diff/verbs/catalog/inspect-page.js
```

Do not silently pick one.

#### Improvement 4: expose shared sections intentionally

The existing `targets`, `viewport`, and `output` sections are useful. For visual catalog work, add or revise sections such as:

- `target`: URL, wait, viewport, root selector,
- `prepare`: prepare type, wait expression, after wait,
- `artifacts`: bundle/css-only/png-only, fail-on-missing, overwrite,
- `catalog`: title, out dir, artifact root, index name.

Shared sections are part of the flag UX; they should be documented and stable.

## 23. Updated implementation plan after research

The implementation plan should be reordered. The next useful milestone is not “add a Goja module skeleton”; it is “turn the existing skeleton into a repository-scanned verbs product”.

### Phase A: clean up and stabilize the existing dsl prototype

- Keep `internal/cssvisualdiff/dsl/host.go` tests passing.
- Move generated artifact directories out of `internal/cssvisualdiff/dsl` if they are not intentional fixtures.
- Rename or document current built-in verbs so users understand `script compare region` and `script compare brief`.
- Confirm `go test ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff` is green.

### Phase B: add `internal/cssvisualdiff/verbcli`

Implement the loupedeck-style dynamic command layer:

- `Bootstrap` and `Repository` structs,
- embedded built-in repository,
- config/env/CLI repository discovery,
- `jsverbs.ScanFS` and `jsverbs.ScanDir`,
- duplicate full-path detection,
- lazy `css-visual-diff verbs` Cobra command,
- generated command tests and help-output tests.

### Phase C: move dynamic script commands under `verbs`

- Stop injecting `host.Commands()` directly into the root command.
- Add `rootCmd.AddCommand(verbcli.NewLazyCommand(...))`.
- Preserve built-in root commands unchanged.
- Document migration from any existing root-level script commands to `css-visual-diff verbs script compare region ...`.

### Phase D: extract inspect/preflight services

- Refactor `modes.Inspect(...)` so browser/page setup, prepare, preflight, and artifact writing can be called from both CLI and JS.
- Add service tests for missing selectors and CSS-only vs bundle behavior.
- Make `inspectAll` operate on a prepared page without reloading for each probe.

### Phase E: add the coherent `require("css-visual-diff")` module

- Keep current `diff` and `report` modules if needed.
- Add public Browser/Page/Catalog wrappers.
- Convert JS lowerCamel options to Go structs.
- Use `vm.NewGoError(err)` for thrown errors; avoid `panic(vm.ToValue(err.Error()))` when wrapping Go errors.
- Add runtime integration tests that `require("css-visual-diff")` and execute a tiny HTTP-page catalog script.

### Phase F: add real catalog verbs

Add built-in verbs that demonstrate the final product shape:

```text
css-visual-diff verbs catalog inspect-page
css-visual-diff verbs catalog inspect-config
css-visual-diff verbs catalog build-manifest
css-visual-diff verbs compare region
css-visual-diff verbs compare brief
```

Then add an external example repository that simulates the Pyxis workflow and proves that repository-scanned scripts can be invoked with flags.

### Phase G: documentation and acceptance

Repository docs should include:

- `docs/js-api.md` for the native module,
- `docs/js-verbs.md` for script annotations, repository discovery, and flag generation,
- `examples/verbs/README.md`,
- a runnable catalog script with command examples.

Acceptance should now include:

1. `css-visual-diff verbs --help` does not scan arbitrary user repositories until the subtree is invoked.
2. Built-in verbs are always available.
3. A filesystem repository can define a `__verb__` script and expose it as a CLI command with flags.
4. Duplicate verb paths fail clearly.
5. A verb can call `require("css-visual-diff")` and run preflight plus `inspectAll` on a local test page.
6. Existing `run`, `inspect`, `compare`, and root help behavior remain unchanged.
7. No generated artifacts are written under source packages by default.

## 24. Maintainer clarifications and answers

These notes capture follow-up design decisions and clarify terminology that can otherwise be confusing.

### 24.1 Promises are required from day one

The API should not start synchronous and migrate later. Use Promises immediately for:

- `cvd.browser(...)`,
- `browser.page(...)` / `browser.newPage(...)`,
- `page.goto(...)`,
- `page.prepare(...)`,
- `page.preflight(...)`,
- `page.inspect(...)` / `page.inspectAll(...)`,
- `catalog.writeManifest(...)`,
- `catalog.writeIndex(...)`.

The Go implementation can internally block on chromedp/file operations, but the goja adapter should expose Promise-returning methods and settle them through the runtime owner pattern. This keeps scripts idiomatic and avoids an API break.

### 24.2 Catalog belongs on the Go side

The catalog API should not be just a JS convenience object. It should be backed by Go services that own:

- manifest schema and schema version,
- path and slug normalization,
- artifact root handling,
- target/result/failure/preflight records,
- summary calculation,
- markdown/HTML index rendering,
- JSON writing,
- future compatibility with non-JS CLI modes.

JavaScript should orchestrate catalog construction; Go should implement the durable data model and writers.

### 24.3 What is preflight for?

Preflight is primarily a selector validation pass. It answers: “If I run expensive artifact extraction now, which probes will definitely fail because their selectors are missing, hidden, or malformed?”

It should check at least:

- selector matches an element,
- optional visibility / non-zero bounds,
- optional root containment,
- basic bounds/text metadata for diagnostics.

Preflight is useful because `chromedp.Screenshot(selector, ...)` and other selector-backed operations can otherwise fail late, fail opaquely, or appear as timeouts. It lets catalog scripts choose policy:

```js
const preflight = await page.preflight(probes)

// authoring mode: skip missing selectors and write a report
const result = await page.inspectAll(preflight.ok(), { failOnMissing: false })

// CI mode: throw a structured SelectorError
preflight.assertAll()
```

So yes: the most important purpose is checking for missing selectors, but the richer goal is structured probe readiness before screenshots/CSS/DOM artifacts.

### 24.4 What is `directReactGlobal` for compared to selectors?

Selectors identify elements that already exist in a loaded page. They answer: “Which DOM node should I inspect or screenshot?”

`directReactGlobal` is a prepare mode. It creates the DOM to inspect by mounting a React component exposed on `window` into a controlled root. It answers: “Before I inspect anything, how do I render this component fixture into the page?”

In other words:

```text
prepare/directReactGlobal -> render or reshape the page
selector/probe            -> choose elements inside the prepared page
inspect/preflight         -> validate/extract artifacts from those elements
```

Typical use cases:

- Prototype HTML exports that expose `window.React`, `window.ReactDOM`, and `window.SomeComponent` but do not have a Storybook route for every component state.
- Component fixtures where the script wants to render one component with explicit props, width, min height, and background.
- Visual catalog baselines where a component should be isolated from the surrounding app shell.

A normal `script` prepare mutates an already loaded page. `directReactGlobal` is more specialized: it replaces/sets up a capture root and renders a named global React component into it.

### 24.5 What can actually be parallelized?

Do not assume that operations inside one chromedp page can run in parallel. A single page/target CDP session is effectively serialized for most useful actions: navigate, evaluate, screenshot, get box model, and write artifacts must be coordinated. Within one prepared page, `inspectAll` should usually run probe extraction in a deterministic sequence, or at most use very small internal parallelism for non-CDP file/format work after data is captured.

Useful parallelism is mostly at coarser boundaries:

1. **Across independent targets/pages**: multiple browser pages or browser contexts can process different catalog targets concurrently, with a configurable worker limit.
2. **Across independent browser instances**: useful for isolation, but heavier.
3. **Across CPU/file post-processing**: writing JSON/Markdown, rendering indexes, diffing already-captured images, and building summaries can overlap after CDP data is collected.
4. **Across preflight batches only if implemented as one page evaluation**: rather than many concurrent CDP calls, evaluate all selector checks in one JS expression inside the page.

Recommended API shape:

```js
await cvd.mapTargets(targets, { concurrency: 2 }, async (target) => {
  const page = await browser.page(target.url, { viewport: target.viewport })
  await page.prepare(target.prepare)
  const preflight = await page.preflight(target.probes) // internally one batched evaluate if possible
  const result = await page.inspectAll(preflight.ok(), { outDir: catalog.artifactDir(target.slug) })
  await page.close()
  return result
})
```

Avoid exposing a generic `parallel()` helper as a central primitive in the MVP. Instead expose target-level concurrency with explicit limits, because browser/CDP resources are the bottleneck and scripts need guardrails.

## 25. Final recommendation

The right abstraction is not “JavaScript that runs YAML.” The right abstraction is a programmable workbench:

```text
Browser owns pages.
Page owns navigation, prepare, selector preflight, and artifacts.
Probe describes what to inspect.
Catalog records outputs and writes reports.
YAML remains the stable declaration format.
JavaScript becomes the dynamic orchestration format.
```

If we implement that shape, `css-visual-diff` becomes much more effective for large catalog work. The operator can express the actual workflow instead of translating it into YAML generation plus shell loops. Selector mistakes become immediate. Browser reuse becomes possible. Timing becomes visible. Reports become part of the script instead of an afterthought.

That is the difference between a CLI that can extract artifacts and a tool that helps people build reliable visual systems.
