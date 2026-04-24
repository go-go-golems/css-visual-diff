---
title: Big Brother Code Review and Assessment
slug: cssvd-goja-js-api-big-brother-review
created: 2026-04-24
status: final-review
project: css-visual-diff
tags:
  - css-visual-diff
  - goja
  - jsverbs
  - code-review
  - architecture
---

# Big Brother Assessment of the `css-visual-diff` Goja/jsverbs Work

## Executive summary

Your “little brother” did a surprisingly strong job for someone new to software engineering. The work is not just a feature patch; it is a multi-phase architectural extension:

- repository-scanned JavaScript verbs under `css-visual-diff verbs ...`,
- reusable Go services for browser/page/prepare/preflight/inspect/catalog behavior,
- Promise-first Goja native API via `require("css-visual-diff")`,
- typed JS-visible errors,
- Go-side catalog manifests/indexes,
- built-in catalog verbs,
- external examples,
- replay scripts and proper binary smoke tests,
- final user-facing docs.

The best part is that the work moved in the right architectural direction: the expensive browser/CSS/catalog logic was pulled into Go services, while JavaScript was used mostly as orchestration glue. That is the right long-term split.

The biggest weakness is that some of the adapter/orchestration code grew organically and now contains stringly typed heuristics, duplicated conversion logic, and a few edge-case hazards. This is normal for a first full-stack feature slice, but it should be cleaned up before the API becomes heavily depended on.

Overall: **good engineering instincts, good validation discipline after correction, but needs a cleanup pass around API contracts, error typing, operation serialization, and catalog edge cases.**

---

## 1. What was built

### 1.1 Lazy JavaScript verbs CLI

The CLI now exposes scanned JavaScript workflows under:

```bash
css-visual-diff verbs ...
```

Instead of injecting generated commands at root, the work correctly scoped them below a lazy `verbs` subtree. This matters because repository scanning can fail due to duplicate command paths or bad user scripts; those errors should not break normal root-level CLI usage.

Main files:

- `internal/cssvisualdiff/verbcli/bootstrap.go`
- `internal/cssvisualdiff/verbcli/command.go`
- `internal/cssvisualdiff/verbcli/runtime_factory.go`
- `cmd/css-visual-diff/main.go`

Repository discovery supports:

- embedded built-ins,
- app config,
- `CSS_VISUAL_DIFF_VERB_REPOSITORIES`,
- `--repository` / `--verb-repository`.

That is a strong design choice.

### 1.2 Service extraction

The implementation extracted reusable browser/inspection primitives into:

- `internal/cssvisualdiff/service/types.go`
- `internal/cssvisualdiff/service/style.go`
- `internal/cssvisualdiff/service/preflight.go`
- `internal/cssvisualdiff/service/prepare.go`
- `internal/cssvisualdiff/service/browser.go`
- `internal/cssvisualdiff/service/inspect.go`
- `internal/cssvisualdiff/service/catalog_service.go`

This was essential. Without it, the JavaScript API would have wrapped CLI modes directly, which would have made the system brittle and hard to test.

### 1.3 Promise-first JS API

The native module now exposes:

```js
const cvd = require("css-visual-diff")
```

With:

```js
await cvd.browser()
await browser.page(url, options)
await browser.newPage()
await browser.close()

await page.goto(url, options)
await page.prepare(spec)
await page.preflight(probes)
await page.inspect(probe, options)
await page.inspectAll(probes, options)
await page.close()

const catalog = cvd.catalog(options)
await cvd.loadConfig(path)
```

This honored the important constraint: **do not ship a synchronous MVP and migrate later.**

### 1.4 Catalog service

The catalog work introduced a Go-owned manifest/index layer:

- schema version,
- targets,
- preflights,
- results,
- failures,
- summaries,
- artifact directory normalization,
- `manifest.json`,
- `index.md`.

That was another important constraint: catalog durability belongs in Go, not only in JS helper code.

### 1.5 Built-in verbs and examples

Built-ins added:

```bash
css-visual-diff verbs catalog inspect-page ...
css-visual-diff verbs catalog inspect-config ...
```

External example repository added under:

```text
examples/verbs/
```

This is useful because it shows the intended operator-facing workflow, not just internal APIs.

### 1.6 Replay scripts and smoke validation

A strong late improvement was preserving replay scripts under the ticket:

```text
scripts/001-...
...
scripts/011-...
```

The binary smoke scripts are especially valuable:

```bash
006-binary-help-smoke.sh
007-binary-js-api-success-smoke.sh
008-binary-js-api-typed-error-smoke.sh
009-binary-catalog-smoke.sh
010-binary-built-in-catalog-inspect-page-smoke.sh
011-binary-built-in-catalog-inspect-config-smoke.sh
```

This changed the validation story from “Go tests say it works” to “the compiled binary can actually run the new workflows.”

---

## 2. What was done well

### 2.1 The work was phased well

The implementation followed a sensible order:

1. clean up DSL test artifacts,
2. add lazy verbs CLI,
3. extract reusable services,
4. add Promise-first JS API,
5. add Go-side catalog service,
6. add built-in catalog verbs,
7. add docs and smoke scripts.

This is what I would expect from a more experienced engineer: reduce coupling before exposing APIs.

### 2.2 The `verbs` namespace decision was correct

Generated script commands are now behind:

```bash
css-visual-diff verbs ...
```

This avoids root command pollution and makes repository scan failures local to the dynamic-command subsystem.

That is a mature CLI design decision.

### 2.3 Promise-first was respected

The implementation did not take the easy shortcut of returning synchronous values from native module calls.

The helper in `internal/cssvisualdiff/dsl/cvd_module.go` makes native operations return Promises:

```go
func promiseValue(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, op string, work func() (any, error), wrap func(*goja.Runtime, any) goja.Value) goja.Value {
	promise, resolve, reject := vm.NewPromise()
	// ...
	go func() {
		value, err := work()
		_ = ctx.Owner.Post(ctx.Context, op, func(context.Context, *goja.Runtime) {
			if err != nil {
				_ = reject(cvdErrorValue(vm, op, err))
				return
			}
			// ...
		})
	}()
	return vm.ToValue(promise)
}
```

That is the right conceptual bridge: work happens off-thread, settlement happens on the Goja owner thread.

### 2.4 Service extraction created real reuse

The browser/page/preflight/inspect service split means:

- CLI modes can reuse logic,
- JS adapters can reuse logic,
- tests can exercise behavior below the CLI layer,
- future catalog/report commands are not locked into Cobra/Glazed internals.

The best example is `service.InspectPreparedPage(...)`, which lets `inspectAll` work against one already-loaded page instead of reloading per probe.

### 2.5 Binary smoke tests were added after the gap was identified

Initially, the validation was mostly `go test`, including integration-style tests. When asked whether the real binary had been smoked, the honest answer was “not enough.” Then the work was corrected by adding compiled-binary smoke scripts.

That is a good engineering habit: when a validation gap is discovered, convert it into a repeatable check.

### 2.6 Documentation caught up to the implementation

The final docs are not just API listings. They explain:

- Promise behavior,
- preflight,
- prepare modes,
- typed errors,
- catalog API,
- repository scanning,
- duplicate command paths,
- migration away from root-level generated commands.

That is important because this feature is a new mental model for users.

---

## 3. Where the work struggled

### 3.1 Validation came late, not first

The biggest process issue was that real compiled-binary smoke tests were not part of the initial validation loop. The work had meaningful Go tests, but the user had to ask:

> how much of this has been truly validated with smoke tests outside of unit tests by using the proper binary?

That question exposed the gap.

This is a common junior-engineer pattern: relying on package tests and forgetting that CLI parsing, generated command names, binary build, path resolution, shell quoting, and output formatting can fail even when Go tests pass.

Concrete example: the catalog smoke initially tried to call a generated command as camelCase:

```bash
custom catalogSmoke ...
```

But the generated CLI command was kebab-case:

```bash
custom catalog-smoke ...
```

The binary caught that. A Go-level direct invocation might not.

#### Lesson

For CLI features, “done” should include:

```bash
go build ./cmd/...
actual-binary command ...
actual-binary command --help
actual-binary command --output json
```

The replay scripts now encode this lesson.

### 3.2 The adapter layer grew large and stringly typed

The JS native adapter is functional, but it grew into a lot of hand-written conversion code.

Where to look:

- `internal/cssvisualdiff/dsl/cvd_module.go`
- `internal/cssvisualdiff/dsl/catalog_adapter.go`
- `internal/cssvisualdiff/dsl/config_adapter.go`

Example:

```go
func lowerInspectResult(result service.InspectResult) map[string]any {
	artifacts := make([]map[string]any, 0, len(result.Results))
	for _, artifact := range result.Results {
		artifacts = append(artifacts, lowerInspectArtifact(artifact))
	}
	return map[string]any{
		"outputDir": result.OutputDir,
		"results":   artifacts,
	}
}
```

And in the catalog adapter:

```go
func decodeCatalogInspectResult(raw map[string]any) (service.InspectResult, error) {
	input, err := decodeInto[catalogInspectResultInput](raw)
	if err != nil {
		return service.InspectResult{}, err
	}
	result := service.InspectResult{OutputDir: input.OutputDir}
	for _, artifact := range input.Results {
		result.Results = append(result.Results, service.InspectArtifactResult{
			Metadata:    decodeCatalogInspectMetadata(artifact.Metadata),
			Style:       decodeCatalogStyleSnapshot(artifact.Style),
			Screenshot:  artifact.Screenshot,
			HTML:        artifact.HTML,
			InspectJSON: artifact.InspectJSON,
		})
	}
	return result, nil
}
```

This is understandable for an MVP, but long-term it creates drift risk:

- Go service structs use snake_case JSON tags.
- JS API uses lowerCamel.
- The adapter manually maps both directions.
- Catalog adapter decodes JS-shaped inspect results back into Go-shaped inspect results.

#### Why it matters

Every new field must be added in multiple places. Missing one will silently drop data.

#### Cleanup sketch

Introduce explicit JS DTO structs near the adapter boundary:

```go
type JSInspectResult struct {
	OutputDir string              `json:"outputDir"`
	Results   []JSInspectArtifact `json:"results"`
}

func JSInspectResultFromService(result service.InspectResult) JSInspectResult
func (r JSInspectResult) ToService() service.InspectResult
```

Then centralize round-tripping:

```go
func toJSValue[T any](vm *goja.Runtime, value T) goja.Value {
	return vm.ToValue(value)
}

func decodeJS[T any](raw any) (T, error) {
	// one implementation
}
```

This would reduce the risk of field drift.

### 3.3 Error typing is useful but too heuristic

The implementation added typed JS errors, which is good. But classification is currently string/op based.

Where to look:

```text
internal/cssvisualdiff/dsl/cvd_module.go
```

Example:

```go
func classifyCVDError(op string, err error) (className string, code string) {
	message := ""
	if err != nil {
		message = strings.ToLower(err.Error())
	}
	switch {
	case strings.Contains(message, "selector") || strings.Contains(op, ".preflight"):
		return "SelectorError", "SELECTOR_ERROR"
	case strings.Contains(op, ".prepare") || strings.Contains(message, "prepare") || strings.Contains(message, "wait_for"):
		return "PrepareError", "PREPARE_ERROR"
	case strings.Contains(op, ".browser") || strings.Contains(op, ".goto") || strings.Contains(message, "navigate"):
		return "BrowserError", "BROWSER_ERROR"
	case strings.Contains(op, ".inspect") || strings.Contains(message, "artifact") || strings.Contains(message, "write"):
		return "ArtifactError", "ARTIFACT_ERROR"
	default:
		return "CvdError", "CVD_ERROR"
	}
}
```

#### Problem

This can misclassify errors:

- an artifact write error mentioning “selector” becomes `SelectorError`,
- an inspect selector miss can become `ArtifactError` depending on message,
- a browser error from inside inspect may be classified as artifact error.

#### Why it matters

Typed errors are part of the public JS API. Scripts will write:

```js
if (err instanceof cvd.SelectorError) { ... }
```

If classification is unstable, workflow logic becomes unreliable.

#### Cleanup sketch

Define typed Go errors in the service layer:

```go
type ErrorCode string

const (
	CodeSelector ErrorCode = "SELECTOR_ERROR"
	CodePrepare  ErrorCode = "PREPARE_ERROR"
	CodeBrowser  ErrorCode = "BROWSER_ERROR"
	CodeArtifact ErrorCode = "ARTIFACT_ERROR"
)

type CvdError struct {
	Code      ErrorCode
	Op        string
	Message   string
	Cause     error
	Details   map[string]any
}

func (e *CvdError) Error() string { return e.Message }
func (e *CvdError) Unwrap() error { return e.Cause }
```

Then wrap service errors intentionally:

```go
return nil, service.NewSelectorError("preflight", selector, err)
```

And map by type:

```go
var cvdErr *service.CvdError
if errors.As(err, &cvdErr) {
	return jsErrorFromCode(vm, cvdErr.Code, cvdErr)
}
```

This would make error classes reliable.

### 3.4 Page operations are Promise-first but not explicitly serialized per page

The design notes correctly said CDP operations should mostly be serialized per page. The implementation returns Promises, but the page wrapper does not enforce per-page operation serialization.

Where to look:

```text
internal/cssvisualdiff/dsl/cvd_module.go
```

Example:

```go
_ = obj.Set("inspectAll", func(rawProbes []map[string]any, rawOptions map[string]any) goja.Value {
	return promiseValue(ctx, vm, "css-visual-diff.page.inspectAll", func() (any, error) {
		// ...
		result, err := service.InspectPreparedPage(state.page.Page(), state.target, "script", probes, opts)
		// ...
	}, nil)
})
```

#### Problem

A user can write:

```js
await Promise.all([
  page.preflight(probes),
  page.inspectAll(probes, { outDir }),
])
```

That schedules two goroutines operating against the same chromedp page. There is no per-page queue or mutex in the adapter.

#### Why it matters

Chromedp/CDP operations on one page are not really parallel. Interleaving could cause flaky behavior, especially with prepare/goto/inspect combinations.

#### Cleanup sketch

Add a per-page operation queue:

```go
type pageState struct {
	page   *service.PageService
	target config.Target
	mu     sync.Mutex
}

func (s *pageState) runExclusive(fn func() (any, error)) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}
```

Then:

```go
return promiseValue(ctx, vm, "css-visual-diff.page.inspectAll", func() (any, error) {
	return state.runExclusive(func() (any, error) {
		// inspect work
	})
}, nil)
```

This would make the documented concurrency guidance enforceable.

### 3.5 Promise settlement ignores scheduling failure

Where to look:

```text
internal/cssvisualdiff/dsl/cvd_module.go
```

Example:

```go
_ = ctx.Owner.Post(ctx.Context, op, func(context.Context, *goja.Runtime) {
	if err != nil {
		_ = reject(cvdErrorValue(vm, op, err))
		return
	}
	// ...
})
```

#### Problem

If `Owner.Post(...)` fails because the runtime is shutting down or context is canceled, the error is ignored. The Promise may remain pending forever from the JS perspective.

#### Why it matters

Pending Promises during shutdown are hard to debug. In CLI use this may be less visible because the runtime exits, but in longer-lived embedding it matters.

#### Cleanup sketch

At minimum, log or surface scheduling failure:

```go
if postErr := ctx.Owner.Post(...); postErr != nil {
	// Maybe log. Cannot safely reject off owner thread.
	log.Debug().Err(postErr).Str("op", op).Msg("failed to settle JS promise")
}
```

Better: use a shared promise helper in go-go-goja that owns this pattern consistently.

### 3.6 Built-in `catalog.js` contains too much workflow logic

The Go-side catalog service is good, but the built-in command logic in JS grew fairly large.

Where to look:

```text
internal/cssvisualdiff/dsl/scripts/catalog.js
```

Example:

```js
async function inspectConfig(configPath, side, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const cfg = await cvd.loadConfig(configPath);
  const targetConfig = side === "react" ? cfg.react : cfg.original;
  const probes = _probesFromConfig(cfg, side);
  // ...
  const statuses = await page.preflight(probes);
  catalog.recordPreflight(target, statuses);
  const readyProbes = [];
  const missing = [];
  for (let i = 0; i < probes.length; i++) {
    const status = statuses[i];
    if (status && status.exists && !status.error) readyProbes.push(probes[i]);
    else missing.push({ probe: probes[i], status });
  }
  // ...
}
```

#### Problem

This is a lot of business policy in an embedded JS file:

- selector choice,
- config styles-vs-sections priority,
- missing selector policy,
- failure recording,
- inspect result writing,
- catalog result wiring.

Some of that is okay for a script, but built-ins are product behavior. Product behavior is easier to test and maintain in Go services.

#### Why it matters

The more product behavior lives in JS built-ins, the harder it is to unit test directly and to reuse from future Go commands.

#### Cleanup sketch

Move “inspect config into catalog” into a Go service:

```go
type InspectConfigCatalogOptions struct {
	ConfigPath     string
	Side           string
	OutDir         string
	Artifacts      string
	FailOnMissing  bool
}

func InspectConfigCatalog(ctx context.Context, opts InspectConfigCatalogOptions) (CatalogRunSummary, error)
```

Then the JS built-in becomes thin:

```js
async function inspectConfig(configPath, side, outDir, values) {
  return await cvd.catalogInspectConfig({ configPath, side, outDir, ...values })
}
```

This keeps JavaScript useful for custom orchestration but makes built-in behavior robust.

### 3.7 Duplicate failure recording in built-in catalog verbs

There is a subtle correctness issue in the built-in JS workflows.

Where to look:

```text
internal/cssvisualdiff/dsl/scripts/catalog.js
```

Example from `inspectPage`:

```js
if (!status || !status.exists || status.error) {
  const err = new cvd.SelectorError(message, "SELECTOR_ERROR", { selector: values.selector, target });
  err.operation = "css-visual-diff.verbs.catalog.inspectPage";
  catalog.addFailure(target, err);
  const manifestPath = await catalog.writeManifest();
  const indexPath = await catalog.writeIndex();
  if (failOnMissing) throw err;
  return _catalogFailureRow(target, status, manifestPath, indexPath, message);
}
```

Then later:

```js
} catch (err) {
  catalog.addFailure(target, err);
  const manifestPath = await catalog.writeManifest();
  const indexPath = await catalog.writeIndex();
  if (failOnMissing) throw err;
  return _catalogFailureRow(...);
}
```

#### Problem

When `failOnMissing` is true, the code records the failure, throws, catches the same error, records it again, writes again, and rethrows.

`inspectConfig` has a similar pattern for `failOnMissing`.

#### Why it matters

CI-mode manifests can contain duplicate failures. This may inflate summary counts and make reports noisy.

#### Cleanup sketch

Use a helper that records once:

```js
function markRecorded(err) {
  err.__catalogRecorded = true
  return err
}

function recordFailureOnce(catalog, target, err) {
  if (!err.__catalogRecorded) {
    catalog.addFailure(target, err)
    err.__catalogRecorded = true
  }
}
```

Or better, structure control flow without throwing inside the same try block after recording.

### 3.8 Catalog Markdown rendering needs escaping

Where to look:

```text
internal/cssvisualdiff/service/catalog_service.go
```

Example:

```go
b.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` |\n", target.Slug, target.Name, target.URL, target.Selector))
```

And:

```go
b.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s |\n", failure.Target.Slug, failure.Code, failure.Operation, failure.Message))
```

#### Problem

Markdown table cells are written raw. Values containing `|`, newlines, backticks, or HTML-ish content can break the table.

#### Why it matters

Catalog indexes are meant to be operator-facing reports. Bad escaping makes them unreliable for real-world URLs, selectors, failure messages, or target names.

#### Cleanup sketch

Add Markdown escaping helpers:

```go
func mdCell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

func mdCode(s string) string {
	s = strings.ReplaceAll(s, "`", "\\`")
	return "`" + s + "`"
}
```

Then use those helpers in every Markdown table row.

### 3.9 Catalog target slug collisions are silently collapsed

Where to look:

```text
internal/cssvisualdiff/service/catalog_service.go
```

Example:

```go
func (c *Catalog) AddTarget(target CatalogTargetRecord) CatalogTargetRecord {
	target = NormalizeCatalogTarget(target)
	if !c.hasTarget(target.Slug) {
		c.manifest.Targets = append(c.manifest.Targets, target)
	}
	c.touch()
	return target
}
```

#### Problem

If two different targets normalize to the same slug, the second one is silently ignored in the target list.

Example:

```text
"CTA Primary" -> cta-primary
"cta primary" -> cta-primary
```

#### Why it matters

Catalogs can accidentally merge distinct targets. That can corrupt reports.

#### Cleanup sketch

Track targets by slug and compare identity:

```go
func (c *Catalog) AddTarget(target CatalogTargetRecord) (CatalogTargetRecord, error) {
	target = NormalizeCatalogTarget(target)
	existing, ok := c.targetsBySlug[target.Slug]
	if ok && !sameTarget(existing, target) {
		return target, fmt.Errorf("catalog target slug collision %q", target.Slug)
	}
	// add normally
}
```

Or generate unique suffixes. For CI/reporting, explicit error is probably safer.

### 3.10 Config adapter is intentionally partial, but should say so in code

Where to look:

```text
internal/cssvisualdiff/dsl/config_adapter.go
```

Example:

```go
func lowerConfigSections(sections []config.SectionSpec) []map[string]any {
	ret := make([]map[string]any, 0, len(sections))
	for _, section := range sections {
		ret = append(ret, map[string]any{
			"name":             section.Name,
			"selector":         section.Selector,
			"selectorOriginal": section.SelectorOriginal,
			"selectorReact":    section.SelectorReact,
			"ocrQuestion":      section.OCRQuestion,
		})
	}
	return ret
}
```

#### Problem

The config loader drops many config fields from the JS-facing object:

- text expectations,
- PNG expectations,
- expected result metadata,
- related files,
- some output semantics.

This is probably fine for `inspect-config`, but the public name `loadConfig` implies a general loader.

#### Why it matters

Users may expect round-trippable config data and be surprised that fields disappear.

#### Cleanup sketch

Either document `loadConfig` as a view optimized for inspect workflows, or provide fuller mapping.

Better:

```go
func lowerConfig(cfg *config.Config) (JSConfig, error)
```

With explicit DTOs that include all fields. Then tests can assert expected fields.

---

## 4. What we learned

### 4.1 The service boundary was the right prerequisite

The early decision to extract style/preflight/prepare/browser/inspect services before building the JS API prevented a mess.

If the JS API had wrapped `modes.Inspect(...)` directly, it would have inherited CLI config assumptions and been harder to use programmatically.

Lesson: **extract reusable domain services before adding a second interface.**

### 4.2 Promise-first matters, but Promise-first is not the same as concurrency-safe

The API now returns Promises, which is good. But Promise-first APIs invite users to do:

```js
await Promise.all([...])
```

So the implementation must decide what is safe to parallelize.

Lesson: **async API design must include serialization/parallelism rules.**

The docs say page operations should be serialized, but the implementation should eventually enforce that.

### 4.3 Binary smoke tests catch a different class of bugs

The generated command name issue (`catalogSmoke` vs `catalog-smoke`) is exactly the kind of thing unit tests often miss.

Lesson: **for CLI features, binary smoke scripts are not optional.**

The final replay scripts are one of the best outcomes of the whole exercise.

### 4.4 Typed public APIs need typed internal errors

The JS API exposes typed errors. That is good. But the Go side still classifies many errors with string heuristics.

Lesson: **if the public API has typed errors, the service layer should have typed errors too.**

### 4.5 Documentation should be written while the API is still fresh

The final docs are strong because they were written while implementation details and validation scripts were still recent.

Lesson: **docs are not a final polish step; they are part of stabilizing the public API.**

---

## 5. What I would praise in code review

If I were reviewing this as the senior engineer, I would explicitly praise:

1. **Correct scoping under `verbs`** — good CLI hygiene and avoids root command surprises.
2. **Service extraction before JS exposure** — good architectural sequencing and reduces long-term coupling.
3. **Promise-first native module** — correct public contract and avoids future breaking migration.
4. **Go-side catalog service** — correct ownership of schema, summary, path normalization, writers.
5. **Typed JS errors** — good UX for script authors and enables authoring/CI policies.
6. **Replay scripts** — very good habit and makes validation auditable.
7. **Final docs** — explain concepts, not only flags.

---

## 6. What I would ask to improve before calling it mature

### Highest priority

#### 1. Add per-page operation serialization

Prevent accidental same-page concurrent chromedp operations.

Suggested location:

```text
internal/cssvisualdiff/dsl/cvd_module.go
```

Add a mutex to `pageState`.

#### 2. Replace string-based error classification with typed Go errors

Suggested locations:

```text
internal/cssvisualdiff/service/errors.go
internal/cssvisualdiff/dsl/cvd_module.go
```

Typed service errors should drive JS error classes.

#### 3. Fix duplicate failure recording in built-in catalog scripts

Suggested location:

```text
internal/cssvisualdiff/dsl/scripts/catalog.js
```

Avoid adding the same failure before throwing and again in catch.

#### 4. Add Markdown escaping to catalog index rendering

Suggested location:

```text
internal/cssvisualdiff/service/catalog_service.go
```

Add tests for `|`, newlines, and backticks in names/selectors/messages.

### Medium priority

#### 5. Introduce JS DTO structs

Suggested locations:

```text
internal/cssvisualdiff/dsl/js_types.go
internal/cssvisualdiff/dsl/catalog_adapter.go
internal/cssvisualdiff/dsl/config_adapter.go
```

Reduce manual map-building drift.

#### 6. Handle catalog slug collisions explicitly

Suggested location:

```text
internal/cssvisualdiff/service/catalog_service.go
```

Either error or generate unique slugs.

#### 7. Expand config adapter coverage or rename it as a view

If `cvd.loadConfig` is public, users will expect complete config data.

### Lower priority

#### 8. Move built-in workflow policy from JS into Go service

Especially `inspect-config`, which is product behavior.

#### 9. Add richer catalog reports

The Markdown index is a good MVP. HTML report/index could come later.

#### 10. Add target-level worker-limit helpers

The docs mention target/page-level concurrency. Eventually the API could provide a helper for safe bounded parallel catalog runs.

---

## 7. Process assessment

### What the junior engineer did well

- Broke the work into phases.
- Committed at reviewable intervals.
- Responded to validation feedback.
- Preserved a diary.
- Added replay scripts after being prompted.
- Used tests and binary smokes.
- Wrote real docs at the end.
- Avoided overloading the CLI root.
- Kept catalog state in Go.

This is a strong learning curve.

### What they struggled with

- Initial validation did not include compiled-binary smoke tests.
- Some APIs were implemented first and refined later after failures.
- Adapter code became large and repetitive.
- Error typing was implemented at the JS boundary rather than from typed service errors upward.
- Some edge cases only became visible after smoke scripts:
  - generated kebab-case command names,
  - single-probe artifact directory behavior,
  - binary CLI flag placement.

These are normal struggles. The important thing is that each issue became a test, script, or doc.

---

## 8. Production readiness assessment

### Ready enough for

- internal use,
- exploratory catalog workflows,
- scripted inspection,
- operator-driven smoke/catalog generation,
- building more examples,
- dogfooding on real projects.

### Not yet ready for

- a fully stable public API promise,
- high-concurrency catalog runs,
- untrusted script execution,
- strict long-term manifest compatibility,
- complex multi-target reports without cleanup.

### Confidence by area

| Area | Confidence | Notes |
| --- | ---: | --- |
| Lazy verbs CLI | High | Good tests and binary help smoke |
| Repository scanning | Medium-high | CLI/env/config covered in tests; more binary smokes useful |
| Browser/page API happy path | High | Go tests + binary smokes |
| Typed errors | Medium | Works, but classification is heuristic |
| Catalog manifest/index | Medium-high | Good MVP; needs escaping/collision handling |
| Built-in catalog verbs | Medium | Useful, but business logic in JS should be tightened |
| Docs | High | Solid coverage |
| Concurrency behavior | Medium-low | Documented, not enforced |

---

## 9. Recommended next cleanup PRs

### PR 1: Error model hardening

- Add `service.CvdError`.
- Use `errors.As`.
- Replace `classifyCVDError(...)` heuristics.
- Add tests for exact error classes.

### PR 2: Page operation serialization

- Add `sync.Mutex` to `pageState`.
- Serialize `goto`, `prepare`, `preflight`, `inspect`, `inspectAll`, `close`.
- Add a test that `Promise.all` does not interleave same-page operations unsafely.

### PR 3: Catalog robustness

- Escape Markdown index cells.
- Detect slug collisions.
- Add tests for both.
- Fix duplicate failure recording in built-in JS.

### PR 4: Adapter DTO cleanup

- Introduce JS-facing DTO structs.
- Reduce `map[string]any` conversions.
- Add field-coverage tests for `loadConfig`.

### PR 5: Built-in workflow service extraction

- Move `inspect-page` and `inspect-config` workflow policy into Go service functions.
- Keep JS built-ins as thin wrappers.

---

## 10. Final big-brother verdict

This was a good piece of work. It shows strong instincts:

- isolate dynamic commands,
- extract services,
- make async APIs async from the start,
- keep durable catalog data in Go,
- write tests,
- add binary smokes,
- document the feature.

The main thing the junior engineer needs to learn is that **working feature slices are not the same as stable API layers**. The feature works, and it is validated. But now that it has a public shape, the next step is hardening:

- typed errors from the service layer,
- less manual conversion code,
- safer same-page operation ordering,
- stronger catalog edge-case handling,
- thinner built-in JS workflows.

If this were my younger sibling’s work, I would say:

> You built the right thing in the right direction. You moved too fast in a few adapter and edge-case areas, but you corrected validation gaps when challenged. The architecture is promising. Now do the senior-engineer cleanup pass: make the implicit contracts explicit, type the errors, serialize the dangerous operations, and harden the catalog outputs.

That is a very respectable outcome.
