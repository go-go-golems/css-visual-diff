---
Title: Opaque Go-backed JavaScript types review and builder API guide
Ticket: CSSVD-OPAQUE-JS-TYPES
Status: active
Topics:
    - javascript-api
    - goja
    - type-safety
    - code-review
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/verbs/low-level-inspect.js
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/doc/topics/javascript-api.md
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/builder_helpers.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/catalog.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/diff.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/extract.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/extractor.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/probe.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/proxy.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/snapshot.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/target.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/jsapi/unwrap.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/service/catalog_service.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/service/extract.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/service/runtime_types.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/service/snapshot.go
      Note: Evidence for JS-to-Go dictionary boundaries
    - Path: internal/cssvisualdiff/service/types.go
      Note: Evidence for JS-to-Go dictionary boundaries
ExternalSources: []
Summary: Code review and implementation guide for replacing loose JavaScript object dictionaries at Go API boundaries with opaque Go-backed handles/builders where stricter runtime feedback is valuable.
LastUpdated: 2026-05-01T09:55:00-04:00
WhatFor: Use this to onboard an intern to the css-visual-diff JavaScript/Go boundary and to plan the next tightening pass for opaque Go-backed types.
WhenToUse: When reviewing or implementing css-visual-diff goja APIs, builder APIs, catalog APIs, inspect/preflight APIs, or runtime type validation.
---




















# Opaque Go-backed JavaScript types review and builder API guide

## 1. Executive summary

`css-visual-diff` already has the core pattern the user asked for: the newer lower-level JavaScript API uses Go-backed `goja` Proxy values for `cvd.target(...)`, `cvd.probe(...)`, `cvd.extractors.*`, `page.locator(...)`, and page handles. Those values are opaque to JavaScript callers, are tracked in a Go-side registry, and are unwrapped by strict APIs such as `cvd.extract(...)` and `cvd.snapshot(...)` instead of trusting arbitrary JavaScript dictionaries. This is the right direction and should become the default for authoring objects and live handles.

The remaining loose surfaces are concentrated in legacy/high-level and serialization-oriented paths:

- `browser.page(url, options)` and `page.goto(url, options)` accept plain option dictionaries.
- `page.prepare(spec)`, `page.preflight(probes)`, `page.inspect(probe, options)`, and `page.inspectAll(probes, options)` accept `map[string]any` / `[]map[string]any` shapes.
- `cvd.catalog(options)` returns a Go-backed catalog object, but its methods still accept raw target/result/status dictionaries.
- `cvd.diff(before, after, options)` intentionally accepts arbitrary snapshot-like values; this is a structural comparison boundary and should stay broad, but its `options` argument can be tightened.
- `.build()` methods return plain serializable objects, which is useful for debugging/export, but callers should not need to round-trip through `.build()` for strict APIs.

Recommended direction: keep plain dictionaries only at true serialization boundaries, and introduce opaque Go-backed builders/handles at authoring and runtime-control boundaries. The next implementation pass should add builders for `PrepareSpec`, `InspectOptions`, `CatalogTarget`, `CatalogFailure`, and possibly `DiffOptions`; add adapter overloads that accept either existing raw objects or new builders where backwards compatibility matters; then gradually document strict forms as preferred.

The most important design principle for the intern is this:

> If a value represents a recipe, handle, policy, or operation that will later be interpreted by Go, prefer an opaque Go-backed type with fluent methods. If a value represents final data emitted to JSON, HTML, Markdown, or the review-site UI, a plain object is acceptable.

## 2. Problem statement and scope

The codebase embeds a JavaScript runtime with `goja` so users can script visual diff workflows. JavaScript is convenient for orchestration, but plain object dictionaries make it easy to send misspelled fields, malformed option shapes, or the wrong kind of object into Go. A raw object such as `{ selector: "#cta" }` does not tell Go whether the user intended a live locator, a reusable probe, a catalog target, a selector status, or a partial inspect request. That weakens runtime feedback.

Opaque Go-backed types fix this by making JavaScript values carry Go-owned identity and behavior. For example, `page.locator("#cta")` can return a Go-backed locator handle. Then `cvd.extract(locator, extractors)` can reject raw `{ selector: "#cta" }` and say: “expected a locator returned by `page.locator(...)`.” This gives immediate, in-context feedback and prevents accidental schema drift.

This review covers:

1. Where `map[string]any` / JavaScript dictionary objects currently cross into Go.
2. Where the code already uses opaque Go-backed Proxy types.
3. Which raw dictionary boundaries should stay as plain data.
4. Which boundaries should become typed builders or accept typed builders.
5. API improvements for existing opaque types.
6. A phased implementation guide suitable for a new intern.

This review does not propose rewriting the service layer. The service layer already has concrete Go types such as `service.PageTarget`, `service.ProbeSpec`, `service.ExtractorSpec`, `service.SnapshotProbeSpec`, `service.CatalogTargetRecord`, `service.DiffOptions`, and `service.PrepareSpec`. The main work is to expose better JavaScript-facing constructors/builders and to route strict APIs through Go-side unwrapping rather than repeated JSON map decoding.

## 3. System orientation for a new intern

### 3.1 High-level architecture

The important runtime path looks like this:

```text
JavaScript verb file
  |
  | require("css-visual-diff")
  v
go-go-goja runtime + goja VM
  |
  | native module registration
  v
internal/cssvisualdiff/jsapi
  |
  | decodes JS calls, owns Proxy handles/builders, schedules async work
  v
internal/cssvisualdiff/service
  |
  | browser, DOM, extract, snapshot, catalog, diff services
  v
Chrome / filesystem / JSON artifacts / review site
```

The `jsapi` package is the membrane between dynamic JavaScript and typed Go. It should do three jobs:

1. Accept JavaScript calls and report JavaScript-friendly errors.
2. Convert or unwrap JavaScript values into service-layer Go types.
3. Return JavaScript-readable results, usually plain JSON-like values.

The service layer should remain mostly unaware of `goja`. It receives ordinary Go structs, performs browser/file/diff work, and returns ordinary Go structs.

### 3.2 Native module registration

The native module is registered in `internal/cssvisualdiff/jsapi/module.go`. `Register(...)` installs exports for error classes, target/probe/extractor APIs, extraction, snapshots, diffs, catalog, and browser creation (lines 16-41). This is the top-level surface of `require("css-visual-diff")`.

Important evidence:

- `Register(...)` installs builder/handle APIs via `installTargetAPI`, `installProbeAPI`, `installExtractorAPI`, `installExtractAPI`, `installSnapshotAPI`, and `installDiffAPI` (`module.go:20-26`).
- The same function still exposes `cvd.catalog` as a raw `map[string]any` options decoder (`module.go:27-33`).
- Browser/page methods mix object methods with raw map inputs such as `browser.page(rawURL string, rawOptions map[string]any)` (`module.go:149-171`).

### 3.3 Async browser work

Async operations use `promiseValue(...)` in `module.go:77-97`. It creates a JavaScript Promise, runs Go work in a goroutine, and posts resolution/rejection back to the runtime owner. Page operations are serialized by `pageState.runExclusive(...)` (`module.go:182-196`) to avoid unsafe concurrent operations against one browser page.

This matters for builder design:

- Fluent builders should stay synchronous because they only mutate Go-owned config.
- Browser/file operations should return Promises and go through `promiseValue(...)`.
- An API that looks like a builder method should not secretly perform browser work.

### 3.4 Service-layer contracts

Relevant service types are already concrete:

- `service.Viewport`, `service.PrepareSpec`, and `service.PageTarget` live in `internal/cssvisualdiff/service/runtime_types.go:3-32`.
- `service.ProbeSpec`, `service.StyleSnapshot`, `service.Bounds`, and `service.SelectorStatus` live in `internal/cssvisualdiff/service/types.go:3-35`.
- `service.ExtractorKind`, `service.ExtractorSpec`, and `service.ElementSnapshot` live in `internal/cssvisualdiff/service/extract.go:9-35`.
- `service.SnapshotProbeSpec`, `service.ProbeSnapshot`, and `service.PageSnapshot` live in `internal/cssvisualdiff/service/snapshot.go:5-23`.
- `service.CatalogTargetRecord` includes a flexible `Metadata map[string]any` field, which is a legitimate extension point (`catalog_service.go:40-48`).
- `service.DiffValues(before, after any, opts DiffOptions)` intentionally accepts arbitrary values for structural diffing (`diff.go:27-43`).

The task is not to invent new domain structs. Most needed structs already exist. The task is to decide when JavaScript should be allowed to create these structs through arbitrary object literals and when it should be guided through Go-backed API methods.

## 4. Current-state evidence: raw dictionaries at JavaScript-to-Go boundaries

### 4.1 Browser/page navigation options

`browser.page(url, options)` and `page.goto(url, options)` accept `rawOptions map[string]any`, then decode through `decodeInto[pageOptions]`:

- `browser.page` signature and decode: `module.go:149-152`.
- `page.goto` signature and decode: `module.go:204-208`.
- `pageOptions` has `Viewport`, `WaitMS`, and `Name` (`module.go:308-312`).
- Defaults and validation happen in `pageOptions.toTarget(...)` (`module.go:314-330`).

This is relatively low risk because the options are small and immediate. However, it still allows misspelled fields such as `{ waitMS: 500 }` instead of `{ waitMs: 500 }` to be silently ignored. A `cvd.pageOptions()` or `cvd.target(...)`-based path could provide better feedback.

### 4.2 Page prepare specs

`page.prepare(raw map[string]any)` decodes a raw dictionary through `decodePrepareSpec(raw)` and runs `service.PrepareTarget(...)` (`module.go:229-238`). The input struct includes many optional fields, including `Props map[string]any` (`module.go:342-360` and `runtime_types.go:9-23`).

This is a strong candidate for a builder because prepare specs have modes and mode-specific fields:

- `type: "script"` uses script/wait fields.
- `type: "directReactGlobal"` uses component/props/root/size/background fields.
- Misspelling `waitForTimeoutMs` or mixing direct-react fields with script mode is easy.

A builder can give mode-specific feedback such as: “`.component(...)` is only valid after `.directReactGlobal(...)`; for script prepare use `.script(...)`.”

### 4.3 Preflight and inspect probes

Legacy/high-level inspection methods accept raw probes:

- `page.preflight(raw []map[string]any)` (`module.go:241-255`).
- `page.inspect(rawProbe map[string]any, rawOptions map[string]any)` (`module.go:256-277`).
- `page.inspectAll(rawProbes []map[string]any, rawOptions map[string]any)` (`module.go:278-296`).
- Raw probes decode through `decodeProbes(...)` and `decodeInspectRequests(...)` (`module.go:349+ in source; see definitions around `probeInput`).

This is the main area where old and new APIs overlap. `cvd.probe(...)` already exists and is Go-backed. `cvd.snapshot(page, probes)` already requires probe builders. But `page.preflight(...)` and `page.inspectAll(...)` still expect arrays of plain probe dictionaries.

Recommended improvement: allow `page.preflight(...)`, `page.inspect(...)`, and `page.inspectAll(...)` to accept `cvd.probe(...)` builders directly. Keep raw objects temporarily for backward compatibility, but document builders as preferred and add strict variants if needed.

### 4.4 Catalog methods

The catalog object is Go-backed in the sense that `wrapCatalog(...)` closes over `*service.Catalog`, but its method inputs are mostly raw dictionaries:

- `catalog.addTarget(raw map[string]any)` (`catalog.go:37-43`).
- `catalog.recordPreflight(rawTarget map[string]any, rawStatuses []map[string]any)` (`catalog.go:44-55`).
- `catalog.addResult(rawTarget map[string]any, rawResult map[string]any)` (`catalog.go:56-67`).
- `catalog.addFailure(call)` decodes `call.Argument(0).Export()` as a target and reads error-ish fields from argument 1 (`catalog.go:68-75`, `catalog.go:230-240`).
- `catalogTargetInput` carries a flexible `Metadata map[string]any` (`catalog.go:96-104`).

Catalog APIs are a good target for builder improvements because they are workflow assembly APIs. Users often move the same target through `catalog.addTarget`, `page.preflight`, `page.inspectAll`, and `catalog.addResult`. A typed `CatalogTargetBuilder` or reuse of `TargetBuilder` can prevent schema mismatches.

However, catalog manifests are output data. `catalog.summary()` and `catalog.manifest()` returning plain lowerCamel objects is correct because those values are meant to be serialized and viewed.

### 4.5 Diff and report APIs

`cvd.diff(before, after, options)` intentionally accepts arbitrary `before` and `after` values by exporting them and calling `service.DiffValues(...)` (`diff.go:13-27`). The service normalizes arbitrary values through JSON marshal/unmarshal (`service/diff.go:58-67`) and recursively compares `map[string]any` and `[]any` (`service/diff.go:70-112`).

This should remain broad because a diff operation is not a domain-specific builder boundary. The inputs are final data. The only improvement is the `options` argument: `DiffOptions` currently has `IgnorePaths []string` (`service/diff.go:11-13`), so `cvd.diffOptions().ignorePaths([...])` could provide autocomplete-like guidance and field validation.

`cvd.report(raw map[string]any)` decodes a raw diff report (`diff.go:29-35`, `diff.go:71-83`). This is acceptable for compatibility with serialized diff JSON, but a future strict form could accept an opaque `DiffResult` handle if diffs become richer.

### 4.6 Generic codec pattern

`decodeInto[T]` JSON-marshals arbitrary values and JSON-unmarshals into a Go type (`jsapi/codec.go:5-15`). This has a useful property: it honors JSON tags and lowerCamel naming. It also has drawbacks:

- Unknown fields are ignored.
- Type mismatch errors happen away from the user’s exact method call.
- Mode-specific validation is not expressed.
- It encourages passing loosely shaped data across the boundary.

This codec should remain available for serialization boundaries and compatibility layers. New strict APIs should prefer direct `goja.Value` argument validation and Go-backed unwrapping.

## 5. Current-state evidence: opaque Go-backed types already exist

### 5.1 Proxy registry and unwrapping

The core implementation lives in `internal/cssvisualdiff/jsapi/proxy.go` and `unwrap.go`.

Key pieces:

- `proxyIDProperty` stores an internal `__cssVisualDiffProxyID` on proxy targets (`proxy.go:12`).
- `ProxyRegistry` maps ids to `{ Owner, Value }` Go bindings (`proxy.go:16-31`).
- `newProxyValue(...)` creates a `goja` Proxy and binds a Go backing value (`proxy.go:56-80`).
- The `Get` trap only returns known methods or throws a contextual unknown-method/wrong-parent error (`proxy.go:64-78`, `proxy.go:104-122`).
- `unwrapProxyBacking[T]` checks that a JavaScript value has a proxy id, that the registry owner matches, and that the Go backing type is expected (`unwrap.go:10-23`).
- `mustUnwrapProxyBacking[T]` converts failures into JS-visible type mismatch errors (`unwrap.go:41-47`).

This is exactly the pattern needed for strict runtime feedback. The key limitation is that `defaultProxyRegistry` is global (`proxy.go:14`). That works today, but a future cleanup should create per-module/per-runtime state so handles from one VM cannot accidentally be looked up in another long-lived process.

### 5.2 Error feedback for wrong methods

The Proxy machinery supports method owner hints:

- Unknown methods produce “available methods” and a closest-method suggestion (`proxy.go:104-113`, `proxy.go:141-151`).
- Wrong-parent calls produce messages like “`.computedStyle()` belongs to `cvd.locator`” plus an optional hint (`proxy.go:116-122`).
- Tests assert these messages (`proxy_test.go`, especially `TestProxyUnknownMethodError` and `TestProxyWrongParentError`).

This should be expanded and made easier to reuse. Every new builder should provide `MethodOwners` for common confusions.

### 5.3 Existing builders and handles

Existing Go-backed values:

1. **Target builder**
   - Created by `cvd.target(name)` (`target.go:12-16`).
   - Methods: `.url`, `.waitMs`, `.viewport`, `.root`, `.prepare`, `.build` (`target.go:27-43`).
   - It validates required strings, non-negative wait, and viewport shape (`target.go:46-87`, `builder_helpers.go:34-77`).
   - `.prepare(...)` still accepts a raw prepare object (`target.go:74-85`).

2. **Probe builder**
   - Created by `cvd.probe(name)` (`probe.go:14-18`).
   - Methods: `.selector`, `.required`, `.source`, `.text`, `.bounds`, `.styles`, `.attributes`, `.build` (`probe.go:21-40`).
   - It keeps both legacy/debug plain extractor maps and service-native `[]service.ExtractorSpec` (`probe.go:8-12`, `probe.go:64-101`).
   - `.build()` returns a plain object for debugging/serialization (`probe.go:106-121`).

3. **Extractor handles**
   - Created under `cvd.extractors.*` (`extractor.go:14-34`).
   - Each handle wraps `extractorHandle` and can convert directly to `service.ExtractorSpec` (`extractor.go:37-55`).
   - `.build()` returns a plain debug object (`extractor.go:41-44`, `extractor.go:57-66`).

4. **Locator handles**
   - Created by `page.locator(selector)` (`module.go:201-203`, `locator.go:16-35`).
   - Methods perform async page-bound reads: `.status`, `.exists`, `.visible`, `.text`, `.bounds`, `.computedStyle`, `.attributes` (`locator.go:20-28`, `locator.go:42-149`).
   - Wrong-parent hints explain that locators are live page-bound handles and probes are reusable recipes (`locator.go:29-34`).

5. **Page handles**
   - `wrapPage(...)` returns a plain object, but it stores a proxy id in the registry with owner `cvd.page` (`module.go:198-205`).
   - `cvd.snapshot(...)` unwraps `cvd.page` strictly (`snapshot.go:11-17`).

### 5.4 Strict lower-level APIs

Two APIs already reject raw objects:

- `cvd.extract(locator, extractors)` requires a `cvd.locator` and an array of `cvd.extractor` handles (`extract.go:11-27`, `extract.go:30-48`).
- `cvd.snapshot(page, probes)` requires a `cvd.page` and an array of `cvd.probe` builders (`snapshot.go:11-27`, `snapshot.go:30-55`).

This is the best current example of the desired style. It proves the architecture works: JavaScript can remain ergonomic while Go retains type identity and context-specific errors.

## 6. Gap analysis

### 6.1 Raw probe dictionaries remain in old page APIs

The largest correctness gap is that `page.preflight`, `page.inspect`, and `page.inspectAll` still use raw probe dictionaries while `cvd.probe` builders already exist. This creates two ways to define probes:

```js
// Loose legacy path
await page.inspectAll([
  { name: "cta", selector: "#cta", props: ["color"] }
], { outDir, artifacts: "bundle" })

// Strict newer path
await cvd.snapshot(page, [
  cvd.probe("cta").selector("#cta").styles(["color"])
])
```

The first path is still useful because it writes inspect artifacts. The API should not force users to choose between artifact writing and typed probes. The inspect path should accept typed probes too.

### 6.2 Prepare specs are still raw and mode-dependent

`PrepareSpec` is complex enough to deserve builders. The raw form is documented and useful, but it is easy for LLM-generated code or a new intern to mix fields incorrectly.

Suggested builders:

```js
const prep = cvd.prepare.script()
  .waitFor("window.React && window.ReactDOM", { timeoutMs: 5000 })
  .script("document.body.dataset.ready = 'true'")
  .afterWaitMs(250)

const prep = cvd.prepare.directReactGlobal("PPXDesktop")
  .props({ page: "shows" })
  .root("#capture-root")
  .size({ width: 920, minHeight: 1200 })
  .background("#fff")
```

Then:

```js
await page.prepare(prep)
cvd.target("shows").url(url).prepare(prep)
```

### 6.3 Catalog targets are raw even though target builders exist

Catalog target records and page targets are not identical, but they overlap strongly: name, URL, selector/root, viewport, description, metadata. Today catalog methods decode raw objects. A `cvd.catalogTarget(...)` builder or an extension of `cvd.target(...)` can reduce duplicate raw shapes.

Possible API:

```js
const target = cvd.catalogTarget("homepage")
  .url(url)
  .selector("#app")
  .viewport(cvd.viewport.desktop())
  .description("Homepage smoke target")
  .metadata("source", "storybook")

catalog.addTarget(target)
```

### 6.4 Builder APIs expose `.build()` as a plain object escape hatch

`.build()` is useful. It supports debugging, snapshotting generated config, and compatibility with old APIs. But strict APIs should unwrap builders directly, not require `.build()`.

Current good example: `cvd.snapshot(page, [cvd.probe(...)])` unwraps the probe builder directly (`snapshot.go:30-55`). Current weaker example: `targetBuilder.build()` returns a target dictionary, but there is no strict API that unwraps `cvd.target(...)` directly to load pages.

Recommendation: keep `.build()` but treat it as “export/debug,” not as the primary communication channel.

### 6.5 Registry is global

`defaultProxyRegistry` is package-global (`proxy.go:14`). This simplifies the current implementation but is not ideal for long-running processes, tests with multiple VMs, or future plugin isolation.

Recommended future state:

```go
type ModuleState struct {
    registry *ProxyRegistry
    ctx      *engine.RuntimeModuleContext
}

func Register(ctx *engine.RuntimeModuleContext, reg *require.Registry) {
    reg.RegisterNativeModule("css-visual-diff", func(vm *goja.Runtime, module *goja.Object) {
        state := &ModuleState{registry: NewProxyRegistry(), ctx: ctx}
        installTargetAPI(state, vm, exports)
        installProbeAPI(state, vm, exports)
        // ...
    })
}
```

Each wrapper then receives `state.registry` instead of passing `nil` and falling back to the global registry.

## 7. Proposed architecture

### 7.1 Boundary classification rule

Use this decision table:

| Boundary | Preferred shape | Why |
| --- | --- | --- |
| Live page/browser handle | Opaque Go-backed handle | Must not be spoofed by raw objects; owns browser resources. |
| Reusable recipe/config authored in JS | Opaque Go-backed builder | Gives method-level validation and wrong-parent feedback. |
| Mode-specific option groups | Opaque Go-backed builder or typed helper | Prevents invalid field combinations. |
| Final result data | Plain JSON-like object | Meant for serialization, review site, diffing, and human inspection. |
| User extension metadata | Plain `Record<string, unknown>` | Intended flexibility; validate only container position. |
| Generic diff input | Plain `any` | The whole purpose is structural comparison of data. |

### 7.2 Proposed package shape

```text
internal/cssvisualdiff/jsapi/
  state.go                 # ModuleState, per-VM registry, shared helpers
  proxy.go                 # generic Proxy construction and errors
  unwrap.go                # generic typed unwrapping
  builders.go              # shared builder conventions, seal/build/export helpers
  target.go                # PageTargetBuilder
  prepare.go               # PrepareBuilder family
  probe.go                 # ProbeBuilder, inspect conversion
  extractor.go             # ExtractorHandle
  inspect_options.go       # InspectOptionsBuilder
  catalog_builder.go       # CatalogTargetBuilder / CatalogFailureBuilder
  diff_options.go          # DiffOptionsBuilder
```

This keeps service code unchanged and concentrates JavaScript runtime behavior in `jsapi`.

### 7.3 Builder lifecycle

A builder should have three conceptual operations:

1. **Mutate** through fluent methods.
2. **Validate** either immediately or at `.build()`/strict unwrap time.
3. **Export** to service-native Go structs or plain JSON-like debug values.

Pseudocode:

```go
type PrepareBuilder struct {
    mode string
    spec service.PrepareSpec
}

func (b *PrepareBuilder) waitFor(vm *goja.Runtime) ProxyMethod {
    return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
        b.spec.WaitFor = requiredStringArg(vm, "cvd.prepare.waitFor", call.Argument(0))
        if timeout := optionalObjectField(call.Argument(1), "timeoutMs"); timeout != nil {
            b.spec.WaitForTimeoutMS = requiredPositiveIntArg(vm, "cvd.prepare.waitFor.timeoutMs", timeout)
        }
        return receiver
    }
}

func (b *PrepareBuilder) toSpec() (service.PrepareSpec, error) {
    if b.mode == "" { return service.PrepareSpec{}, errors.New("prepare mode is required") }
    if b.mode == "direct-react-global" && b.spec.Component == "" {
        return service.PrepareSpec{}, errors.New("component is required")
    }
    return b.spec, nil
}
```

### 7.4 Strict API unwrapping with compatibility fallbacks

For legacy APIs, use a two-stage strategy:

```go
func decodeProbeArgument(vm *goja.Runtime, value goja.Value) (service.ProbeSpec, bool, error) {
    if b, err := unwrapProxyBacking[probeBuilder](vm, registry, "page.inspect", value, "cvd.probe"); err == nil {
        spec, err := b.toInspectRequest()
        return spec, true, err
    }

    // compatibility fallback for old scripts
    var raw map[string]any
    if err := decodeIntoValue(vm, value, &raw); err != nil {
        return service.ProbeSpec{}, false, typeError
    }
    spec, err := decodeRawProbe(raw)
    return spec, false, err
}
```

In strict-only APIs, skip the fallback:

```go
func mustProbeBuilder(vm *goja.Runtime, value goja.Value) *probeBuilder {
    return mustUnwrapProxyBacking[probeBuilder](vm, state.registry, "css-visual-diff.snapshot", value, "cvd.probe")
}
```

### 7.5 Error message standard

Every strict boundary should answer four questions:

1. What operation failed?
2. What type/value was expected?
3. What did the user pass?
4. What should the user do instead?

Example messages:

```text
page.inspectAll: expected each probe to be cvd.probe("name").selector("...") or a legacy probe object; got number at probes[0].

cvd.target.prepare: expected cvd.prepare.* builder or prepare object; got cvd.probe. Prepare steps belong to targets/pages, not probes.

catalog.addTarget: expected cvd.catalogTarget() or cvd.target() builder; got plain object with no url/name/slug. Use cvd.catalogTarget("slug").url(...).
```

## 8. API design recommendations

### 8.1 Improve `cvd.target(...)`

Current methods are good but incomplete for strict workflow composition. Add:

```ts
interface TargetBuilder {
  url(url: string): this
  waitMs(ms: number): this
  viewport(width: number, height: number): this
  viewport(v: Viewport | ViewportBuilder): this
  root(selector: string): this
  prepare(spec: PrepareBuilder | PrepareObject): this
  description(text: string): this              // useful when exporting to catalog
  metadata(key: string, value: unknown): this  // optional catalog bridge
  metadata(values: Record<string, unknown>): this
  build(): PageTargetObject
  buildCatalogTarget?(): CatalogTargetObject
}
```

Intern implementation notes:

- Keep `url`, `waitMs`, `viewport`, and `root` as-is.
- Change `prepare(...)` to first try unwrapping `PrepareBuilder`, then fall back to raw object decoding.
- Consider separate `cvd.catalogTarget(...)` if mixing catalog fields into `cvd.target(...)` makes the concept too broad.

### 8.2 Improve `cvd.probe(...)`

Current probe methods support selector, required, source, text, bounds, styles, and attributes. Add small ergonomic methods:

```ts
interface ProbeBuilder {
  selector(selector: string): this
  required(value?: boolean): this
  optional(): this
  source(source: string): this
  text(options?: TextOptions): this
  exists(): this
  visible(): this
  bounds(): this
  styles(props: string[]): this
  attributes(names: string[]): this
  extract(extractor: ExtractorHandle): this
  extract(extractors: ExtractorHandle[]): this
  build(): ProbeObject
}
```

Why:

- `.optional()` is clearer than `.required(false)` for non-fatal snapshot probes.
- `.exists()` and `.visible()` align probe recipes with extractor handles.
- `.text(options)` allows callers to choose trimming/normalization instead of hardcoding normalized+trimmed text.
- `.extract(...)` lets new extractor types be added without adding one method per extractor to the probe builder.

Implementation caution: `probeBuilder` currently stores both `extractors []map[string]any` and `specs []service.ExtractorSpec` (`probe.go:8-12`). That duplicate state is useful for `.build()` and service calls, but it can drift. Prefer making `specs` the source of truth and generating plain `.build()` output from `specs`.

### 8.3 Improve `cvd.extractors.*`

Current extractor handles are functions that return immutable-ish handles. Add method-level options where useful:

```ts
cvd.extractors.text({ trim: true, normalizeWhitespace: true })
cvd.extractors.computedStyle(["color", "font-size"])
cvd.extractors.attributes(["id", "class"])
```

Intern implementation notes:

- `extractorHandle` already converts to `service.ExtractorSpec` (`extractor.go:53-55`). Extend the backing struct rather than returning raw maps.
- Validate extractor kind against `service.ExtractorKind` constants, not arbitrary strings.

### 8.4 Add `cvd.prepare.*` builders

Proposed API:

```ts
cvd.prepare.script()
  .waitFor(expression: string, options?: { timeoutMs?: number })
  .script(source: string)
  .scriptFile(path: string)
  .afterWaitMs(ms: number)

cvd.prepare.directReactGlobal(component: string)
  .props(values: Record<string, unknown>)
  .root(selector: string)
  .size(options: { width?: number, minHeight?: number })
  .background(color: string)
  .waitFor(expression: string, options?: { timeoutMs?: number })
  .afterWaitMs(ms: number)
```

Recommended wrong-parent hints:

- Calling `.selector(...)` on prepare: “Selectors for extraction belong to `cvd.probe(...).selector(...)` or `page.locator(...)`; prepare roots use `.root(...)`.”
- Calling `.props(...)` on script prepare: “Props only apply to `cvd.prepare.directReactGlobal(...)`.”
- Calling `.script(...)` on direct-react prepare: “Script source applies to `cvd.prepare.script()`; direct-react prepare renders a named global component.”

### 8.5 Add inspect/preflight option builders

Proposed API:

```ts
const options = cvd.inspectOptions()
  .outDir("/tmp/cssvd/catalog/artifacts/homepage")
  .artifacts("bundle")
  .outputFile("inspect.json")

await page.inspectAll(probes, options)
```

This wraps `service.InspectAllOptions`, whose raw JS input currently accepts `outDir`, `format`, `artifacts`, and `outputFile` (`module.go:342-360`). It should validate artifact format names early and normalize `format`/`artifacts` in one place.

### 8.6 Add catalog target builders

Proposed API:

```ts
const target = cvd.catalogTarget("homepage")
  .name("Homepage")
  .url("http://localhost:3000")
  .selector("#root")
  .viewport(cvd.viewport.desktop())
  .description("Main homepage shell")
  .metadata({ source: "storybook", owner: "frontend" })

catalog.addTarget(target)
```

Acceptable overloads:

- `catalog.addTarget(cvd.catalogTarget(...))` — preferred.
- `catalog.addTarget(cvd.target(...))` — possible if target has enough catalog fields.
- `catalog.addTarget({ ... })` — compatibility fallback, with deprecation warning later if desired.

Keep `metadata` flexible. It is an intentional `map[string]any` extension field (`catalog_service.go:40-48`). The builder should validate that metadata is an object, but it should not over-constrain user-defined metadata keys.

### 8.7 Add diff options builder only if diff grows

Current diff options are tiny (`ignorePaths`). A builder is optional:

```ts
cvd.diff(before, after, cvd.diffOptions().ignorePaths(["results[0].snapshot.bounds.x"]))
```

This becomes more valuable if numeric tolerances, CSS normalization, ignore-key patterns, or path globs are added.

## 9. Detailed implementation guide

### Phase 1: inventory and tests for existing raw boundaries

Goal: create a safety net before changing API shapes.

Tasks:

1. Add unit tests that document current raw object compatibility for:
   - `page.preflight([{...}])`
   - `page.inspect({...}, {...})`
   - `page.inspectAll([{...}], {...})`
   - `catalog.addTarget({...})`
   - `catalog.addResult({...}, {...})`
2. Add tests that strict APIs reject raw objects:
   - `cvd.extract({ selector: "#x" }, [...])`
   - `cvd.snapshot(page, [{ selector: "#x" }])`
3. Add tests for user-facing error strings.

Suggested test file names:

```text
internal/cssvisualdiff/jsapi/legacy_object_compat_test.go
internal/cssvisualdiff/jsapi/strict_boundary_test.go
```

### Phase 2: introduce per-module state

Goal: eliminate the global registry as the default dependency before adding more opaque types.

Pseudocode:

```go
type moduleState struct {
    ctx      *engine.RuntimeModuleContext
    registry *ProxyRegistry
}

func newModuleState(ctx *engine.RuntimeModuleContext) *moduleState {
    return &moduleState{ctx: ctx, registry: NewProxyRegistry()}
}
```

Then change install functions:

```go
installTargetAPI(state, vm, exports)
installProbeAPI(state, vm, exports)
installExtractorAPI(state, vm, exports)
installExtractAPI(state, vm, exports)
```

The new state should flow into `newProxyValue(...)` and `unwrapProxyBacking(...)`. This is mostly plumbing but important because each new builder increases registry usage.

### Phase 3: add `PrepareBuilder`

Goal: replace mode-heavy prepare raw objects with in-context APIs.

Files:

- Add `internal/cssvisualdiff/jsapi/prepare.go`.
- Update `target.go` `prepare(...)` method.
- Update `module.go` `page.prepare(...)` method.
- Add `internal/cssvisualdiff/jsapi/prepare_test.go`.

Pseudocode for `page.prepare(...)`:

```go
_ = obj.Set("prepare", func(call goja.FunctionCall) goja.Value {
    spec, err := decodePrepareArgument(vm, state.registry, call.Argument(0))
    if err != nil { panic(cvdTypeError(vm, err)) }
    return promiseValue(... service.PrepareTarget(... spec) ...)
})
```

`decodePrepareArgument` should:

1. Try unwrap `cvd.prepare` builder.
2. Fall back to raw object decoding for compatibility.
3. Validate mode-specific required fields.

### Phase 4: allow probe builders in preflight/inspect APIs

Goal: unify old artifact-writing APIs with new typed probe authoring.

Files:

- Update `module.go` page methods around `preflight`, `inspect`, and `inspectAll`.
- Add helper functions in `probe.go`:
  - `probeBuilder.toProbeSpec()`
  - `probeBuilder.toInspectRequest()`
  - `unwrapProbeListOrLegacy(...)`
- Add tests.

API after this phase:

```js
const probes = [
  cvd.probe("cta").selector("#cta").styles(["color"]).attributes(["class"])
]

await page.preflight(probes)
await page.inspectAll(probes, cvd.inspectOptions().outDir(outDir).artifacts("bundle"))
```

Compatibility:

```js
await page.inspectAll([{ name: "cta", selector: "#cta", props: ["color"] }], { outDir })
```

should continue working unless the project explicitly chooses a breaking strict mode.

### Phase 5: add inspect options builder

Goal: make artifact output options discoverable and validated.

Files:

- Add `internal/cssvisualdiff/jsapi/inspect_options.go`.
- Update `decodeInspectOptions(...)` call sites.
- Add tests for valid formats and typo errors.

Builder API:

```js
cvd.inspectOptions()
  .outDir("/tmp/out")
  .artifacts("bundle")
  .outputFile("inspect.json")
```

### Phase 6: add catalog target builder

Goal: reduce raw target dictionaries in catalog workflows.

Files:

- Add `internal/cssvisualdiff/jsapi/catalog_target.go` or extend `catalog.go`.
- Update `catalog.addTarget`, `recordPreflight`, `addResult`, and `addFailure` to accept builders.
- Add tests.

Pseudocode:

```go
func decodeCatalogTargetValue(vm *goja.Runtime, value goja.Value) (service.CatalogTargetRecord, error) {
    if b, err := unwrapProxyBacking[catalogTargetBuilder](...); err == nil {
        return b.toRecord()
    }
    if b, err := unwrapProxyBacking[targetBuilder](...); err == nil {
        return catalogRecordFromPageTarget(b.target)
    }
    return decodeCatalogTarget(value.Export())
}
```

### Phase 7: improve probe/extractor builders

Goal: reduce duplicated state and add missing fluent methods.

Tasks:

1. Make `[]service.ExtractorSpec` the source of truth inside `probeBuilder`.
2. Generate `.build().extractors` from `specs`.
3. Add `.exists()`, `.visible()`, `.optional()`, `.text(options)`, and `.extract(...)`.
4. Update documentation and examples.

### Phase 8: documentation and migration examples

Goal: make the preferred typed path obvious.

Files:

- Update `internal/cssvisualdiff/doc/topics/javascript-api.md`.
- Update `examples/verbs/low-level-inspect.js`.
- Add one catalog workflow example using typed builders.

Documentation should state:

- Raw objects are supported at legacy compatibility and serialization boundaries.
- New scripts should prefer Go-backed builders/handles.
- Strict APIs reject raw objects intentionally.

## 10. Testing and validation strategy

### 10.1 Unit tests

Add tests for each builder:

```go
func TestPrepareBuilderScript(t *testing.T)
func TestPrepareBuilderDirectReactGlobal(t *testing.T)
func TestPrepareBuilderWrongParentErrors(t *testing.T)
func TestInspectOptionsBuilderValidation(t *testing.T)
func TestCatalogTargetBuilder(t *testing.T)
```

Use the existing pattern in `builders_test.go`, which creates a `goja` VM, installs exports, runs JavaScript, and inspects exported values.

### 10.2 Boundary tests

For every method that accepts either builder or raw object, test both paths:

```go
func TestPageInspectAllAcceptsProbeBuilders(t *testing.T)
func TestPageInspectAllStillAcceptsLegacyObjects(t *testing.T)
func TestCatalogAddTargetAcceptsCatalogTargetBuilder(t *testing.T)
func TestCatalogAddTargetStillAcceptsLegacyObject(t *testing.T)
```

### 10.3 Error tests

Every wrong-parent hint should have a test. Examples:

```js
cvd.prepare.script().props({})
cvd.catalogTarget("x").styles(["color"])
cvd.probe("x").url("http://example.test")
page.locator("#x").build()
```

Assertions should check substrings, not exact full error strings, so helpful wording can evolve.

### 10.4 Integration smoke tests

Run existing relevant tests:

```bash
go test ./internal/cssvisualdiff/jsapi -count=1
go test ./internal/cssvisualdiff/verbcli -run TestCVDModuleExposes -count=1
go test ./internal/cssvisualdiff/service -count=1
```

Then run a low-level verb against a local page if browser dependencies are available:

```bash
css-visual-diff verbs --repository examples/verbs examples low-level inspect \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-low-level --output json
```

## 11. Risks and tradeoffs

### Risk: too many builders make simple scripts verbose

Mitigation: keep small literal options accepted where the operation is immediate and low risk. Do not force builders for final result objects or simple one-off data. Provide helpers like `cvd.viewport.desktop()` and `cvd.prepare.script()` to keep common code short.

### Risk: backwards compatibility

Mitigation: add builder support as an overload before removing raw object support. If strict-only behavior is desired later, add new names such as `page.inspectStrict(...)` or a runtime option before breaking existing scripts.

### Risk: registry lifetime and leaks

The current registry never deletes bindings. For short-lived script processes this is acceptable. For long-lived runtimes, add per-module state first and consider cleanup on page/browser close or use finalizer-like lifecycle hooks. Avoid relying on hidden JavaScript ids across VMs.

### Risk: confusing `.build()` semantics

Mitigation: document `.build()` as an export/debug method. Strict APIs should accept builder handles directly. `.build()` should produce stable JSON-compatible data, not a new opaque type.

### Risk: metadata needs flexibility

Mitigation: keep metadata as `Record<string, unknown>` / `map[string]any`. The builder should validate where metadata can appear, not every key inside it.

## 12. Concrete review findings

### Finding 1: strict lower-level APIs already implement the desired pattern

`cvd.extract` and `cvd.snapshot` unwrap Proxy-backed handles/builders and reject raw objects. This is the model to copy.

Files:

- `internal/cssvisualdiff/jsapi/extract.go:11-48`
- `internal/cssvisualdiff/jsapi/snapshot.go:11-55`
- `internal/cssvisualdiff/jsapi/proxy.go:56-80`
- `internal/cssvisualdiff/jsapi/unwrap.go:10-47`

### Finding 2: legacy inspect/preflight APIs should accept probe builders

The code already has `cvd.probe`, but artifact-producing page APIs still accept only raw probe dictionaries.

Files:

- `internal/cssvisualdiff/jsapi/module.go:241-296`
- `internal/cssvisualdiff/jsapi/probe.go:21-121`

Priority: high.

### Finding 3: prepare specs deserve opaque builders

Prepare specs are mode-dependent and currently raw.

Files:

- `internal/cssvisualdiff/jsapi/module.go:229-238`
- `internal/cssvisualdiff/jsapi/target.go:74-85`
- `internal/cssvisualdiff/service/runtime_types.go:9-23`

Priority: high.

### Finding 4: catalog method inputs should gain builders

Catalog itself is Go-backed, but method inputs are dictionary-shaped.

Files:

- `internal/cssvisualdiff/jsapi/catalog.go:32-93`
- `internal/cssvisualdiff/jsapi/catalog.go:96-120`
- `internal/cssvisualdiff/service/catalog_service.go:40-48`

Priority: medium-high.

### Finding 5: keep plain output data plain

Lowering functions that return plain maps are appropriate for JSON/result boundaries:

- `lowerElementSnapshot` (`extract.go:50-70`)
- `lowerPageSnapshot` (`snapshot.go:57-73`)
- `lowerSnapshotDiff` (`diff.go:85-90`)
- catalog manifest/summary lowering (`catalog.go:243+`)

Priority: do not change unless adding typed result handles for a specific feature.

## 13. Intern checklist

Before implementing a new JS API, answer:

- Is this value a live handle, builder/recipe, option policy, or final data?
- Does Go need to know this value came from our API rather than an arbitrary object?
- Can the user benefit from method-level validation?
- Are there common wrong-parent mistakes we can explain?
- Should raw object compatibility remain temporarily?
- What service-layer struct should this builder produce?
- What tests prove raw objects are rejected or accepted as intended?

When adding a builder:

1. Define the Go backing struct.
2. Add a constructor export.
3. Wrap it with `newProxyValue(...)`.
4. Add fluent methods that validate arguments immediately.
5. Add `toSpec()` / `toRecord()` for service-native conversion.
6. Add `.build()` only for debug/serialization output.
7. Add unwrapping helpers for strict APIs.
8. Add wrong-parent hints.
9. Add tests for success, validation errors, wrong-parent errors, and raw object rejection/compatibility.
10. Update docs and examples.

## 14. References

- `internal/cssvisualdiff/jsapi/module.go` — native module registration, browser/page methods, raw inspect/preflight/prepare boundaries.
- `internal/cssvisualdiff/jsapi/proxy.go` — Go-backed Proxy registry and method error machinery.
- `internal/cssvisualdiff/jsapi/unwrap.go` — typed unwrapping from JavaScript values to Go backing structs.
- `internal/cssvisualdiff/jsapi/target.go` — existing target builder and raw prepare handoff.
- `internal/cssvisualdiff/jsapi/probe.go` — existing probe builder and current duplicate extractor state.
- `internal/cssvisualdiff/jsapi/extractor.go` — existing extractor handles.
- `internal/cssvisualdiff/jsapi/extract.go` — strict `cvd.extract` implementation.
- `internal/cssvisualdiff/jsapi/snapshot.go` — strict `cvd.snapshot` implementation.
- `internal/cssvisualdiff/jsapi/catalog.go` — Go-backed catalog object with raw method inputs.
- `internal/cssvisualdiff/jsapi/diff.go` — intentionally broad structural diff API.
- `internal/cssvisualdiff/service/runtime_types.go` — `Viewport`, `PrepareSpec`, `PageTarget`.
- `internal/cssvisualdiff/service/types.go` — `ProbeSpec`, `SelectorStatus`, style/result primitives.
- `internal/cssvisualdiff/service/extract.go` — extractor and element snapshot service contracts.
- `internal/cssvisualdiff/service/snapshot.go` — page snapshot service contracts.
- `internal/cssvisualdiff/service/catalog_service.go` — catalog records and manifest contracts.
- `internal/cssvisualdiff/doc/topics/javascript-api.md` — public JavaScript API documentation.
- `examples/verbs/low-level-inspect.js` — example using the newer locator/extractor/probe APIs.
