# Tasks

This ticket intentionally does **not** treat backward compatibility as a hard requirement. Implement the clean, coherent, JavaScript-first API described in the design docs. Do not preserve old ambiguous names merely because they already exist. If a short form remains, it must be an intentionally designed low-effort surface, not a historical alias.

Every implementation phase must include:

1. code changes,
2. Go/package tests,
3. an update to the embedded JavaScript API reference,
4. a real ticket-local smoke script under `scripts/`, and
5. diary/changelog notes with commands and outcomes.

Embedded JS API reference:

```text
internal/cssvisualdiff/doc/topics/javascript-api.md
```

Ticket smoke scripts directory:

```text
ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/scripts/
```

Use numeric prefixes so scripts can be replayed in order.

---

## Phase 0 — Ticket setup and design

- [x] Read Pyxis css-visual-diff maintainer feature request document.
- [x] Read Pyxis JS workflow exploration ticket index, design guide, diary, research notes, and tasks.
- [x] Create docmgr ticket `CSSVD-JSAPI-PIXEL-WORKFLOWS` under repository-local `./ttmp`.
- [x] Create source-analysis reference document.
- [x] Create main API design document.
- [x] Create JavaScript-centric collected-data/comparison-object design report.
- [x] Create full JavaScript API coherence and fluent primitive design report.
- [x] Update the design to explicitly avoid backward compatibility as a hard requirement.
- [x] Create investigation diary.
- [x] Relate important source files/docs with `docmgr doc relate`.
- [x] Upload design documents to reMarkable.
- [x] Run `docmgr doctor --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --stale-after 30`.

### Phase 0 docs/smoke

- [x] Document the design decision in `changelog.md`.
- [x] Record the design investigation in `reference/02-investigation-diary.md`.
- [x] Validate ticket hygiene with `docmgr doctor`.

---

## Phase 1 — Service model for collected selector data

Goal: introduce the Go-side primitive for “collected browser truth for one selector at one point in time.” This is the foundation for the JavaScript-first API.

### Implementation tasks

- [x] Add `internal/cssvisualdiff/service/collection.go`.
- [x] Define `SelectionData` / `CollectedSelectionData` with lowerable fields:
  - schema version,
  - name,
  - URL,
  - selector,
  - source,
  - status/existence/visibility,
  - bounds,
  - normalized text,
  - HTML if useful,
  - computed styles,
  - attributes,
  - optional screenshot descriptor.
- [x] Define `CollectOptions` with profiles:
  - `minimal`,
  - `rich`,
  - `debug`,
  - custom struct equivalent of JS `inspect` options.
- [x] Implement `CollectSelection(page *driver.Page, locator LocatorSpec, opts CollectOptions) (SelectionData, error)` using existing primitives from `service/dom.go`.
- [x] Add support for collecting all computed styles when `styles: "all"` or the Go equivalent is requested.
- [x] Add support for collecting all element attributes when `attributes: "all"` or the Go equivalent is requested.
- [x] Decide whether screenshot capture belongs in this phase or Phase 2; if included, store image/temp-path metadata without tying it to final artifact paths.
- [x] Ensure collection returns stable data and does not keep querying the browser after collection.
- [x] Add service-level errors that distinguish selector-not-found, invalid selector, browser failure, and artifact failure.

### Tests

- [x] Add `internal/cssvisualdiff/service/collection_test.go`.
- [x] Test `minimal` collection on a local HTML fixture.
- [x] Test `rich` collection includes bounds, text, computed styles, and attributes.
- [x] Test missing selector behavior.
- [x] Test invalid selector behavior.
- [x] Test all-styles/all-attributes behavior.
- [x] Test collected data is serializable through JSON.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with a collection model section:
  - define `CollectedSelection`,
  - explain live locator vs collected data,
  - document that collected data is immutable browser truth at collection time,
  - preview `locator.collect()` and `cvd.collect.selection(...)`.

### Real smoke script

- [x] Add `scripts/001-service-collection-smoke.sh`.
- [x] Script should run a focused Go test command, e.g. `go test ./internal/cssvisualdiff/service -run 'TestCollectSelection' -count=1`.
- [x] Script should fail on missing expected output or test failure.
- [x] Record smoke output in the diary.

---

## Phase 2 — Extract image/pixel diff service primitives

Goal: move image comparison out of the mode-shaped implementation and into reusable services.

### Implementation tasks

- [x] Identify private helpers in `internal/cssvisualdiff/modes/compare.go`:
  - PNG reading,
  - image padding/normalization,
  - threshold comparison,
  - diff-only image creation,
  - side-by-side comparison image creation,
  - PNG writing.
- [x] Add `internal/cssvisualdiff/service/pixel.go`.
- [x] Define `PixelDiffOptions` and `PixelDiffResult` with lowerCamel JS/JSON semantics.
- [x] Implement image-level functions:
  - `DiffImages(left image.Image, right image.Image, opts PixelDiffOptions)`,
  - `DiffPNGFiles(leftPath, rightPath string, opts PixelDiffOptions)`,
  - writer helpers for diff-only and comparison images.
- [x] Ensure image service functions do not depend on CLI mode types.
- [x] Update `modes/compare.go` to call the service functions where still needed.
- [x] Make writer helpers create parent directories.

### Tests

- [x] Add `internal/cssvisualdiff/service/pixel_test.go`.
- [x] Test identical images produce zero changed pixels.
- [x] Test one changed pixel above threshold.
- [x] Test threshold behavior.
- [x] Test different-size images normalize/pad consistently.
- [x] Test diff-only and comparison PNG files are written.
- [x] Test parent directories are created.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with the image/pixel primitive design:
  - clarify structural diffs vs pixel diffs,
  - introduce canonical namespace `cvd.image.diff(...)`,
  - document `changedPercent`, `changedPixels`, `totalPixels`, `normalizedWidth`, and `normalizedHeight`.

### Real smoke script

- [x] Add `scripts/002-pixel-service-smoke.sh`.
- [x] Script should run focused pixel tests and optionally generate two tiny PNG fixtures under a temp directory.
- [x] Script should verify diff PNG files exist and are non-empty.
- [x] Record smoke output in the diary.

---

## Phase 3 — Service-level selection comparison

Goal: compare two collected selection values without requiring JavaScript or CLI modes.

### Implementation tasks

- [x] Add `internal/cssvisualdiff/service/selection_compare.go`.
- [x] Define `SelectionComparisonData` with:
  - schema version,
  - name,
  - left/right collected data summaries,
  - pixel summary,
  - bounds diff,
  - text diff,
  - style diffs,
  - attribute diffs,
  - artifact descriptors.
- [x] Define `CompareSelectionOptions`:
  - threshold,
  - include/exclude style props,
  - normalization options if needed,
  - artifact planning options if needed.
- [x] Implement `CompareSelections(left SelectionData, right SelectionData, opts CompareSelectionOptions) (SelectionComparisonData, error)`.
- [x] Use the new pixel service when both selections include screenshot/image data.
- [x] Implement pure data diff helpers for bounds, text, styles, and attributes.
- [x] Keep comparison deterministic: sorted diff ordering, stable paths, stable schema.
- [x] Ensure comparison does not re-query the browser.

### Tests

- [x] Add `internal/cssvisualdiff/service/selection_compare_test.go`.
- [x] Test style diff filtering.
- [x] Test attribute diff filtering.
- [x] Test bounds diff output.
- [x] Test pixel diff integration with collected screenshot data.
- [x] Test deterministic ordering of diffs.
- [x] Test JSON serialization and schema version.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with `SelectionComparison` concepts:
  - compare collected selections,
  - distinguish comparison data from reports/artifacts,
  - show planned methods `styles.diff`, `bounds.diff`, `attributes.diff`, and `pixel.summary`.

### Real smoke script

- [x] Add `scripts/003-selection-compare-service-smoke.sh`.
- [x] Script should run focused service comparison tests.
- [x] Script should verify a JSON fixture/result can be produced from two simple collected selections.
- [x] Record smoke output in the diary.

---

## Phase 4 — JavaScript `CollectedSelection` handle

Goal: expose rich collected selector data to JavaScript as a Go-backed fluent/queryable object.

### Implementation tasks

- [x] Add `internal/cssvisualdiff/jsapi/collect.go`.
- [x] Add `locator.collect(options?)` to `internal/cssvisualdiff/jsapi/locator.go`.
- [x] Add namespace function `cvd.collect.selection(locator, options?)`.
- [x] Add Go-backed Proxy owner `cvd.collectedSelection`.
- [x] Implement methods:
  - `summary()` → compact plain summary,
  - `toJSON(options?)` → plain serializable data,
  - `status()` → selector status,
  - `bounds()` → bounds or null,
  - `text(options?)` → text view,
  - `styles(propsOrPreset?)` → filtered style map,
  - `attributes(names?)` → filtered attribute map,
  - `screenshot.write(path)` if screenshot data exists.
- [x] Decode `inspect` options:
  - `"minimal"`,
  - `"rich"` as default,
  - `"debug"`,
  - object form with `styles`, `attributes`, `matchedStyles`, `bounds`, `text`, `screenshots`.
- [x] Add helpful wrong-parent errors:
  - `.diff()` on collected selection → use `cvd.compare.selections(left, right)`,
  - `.selector()` on collected selection → collected selections are immutable; create a new locator.
- [x] Ensure collection uses `pageState.runExclusive`.
- [x] Ensure `toJSON()` is explicit and examples return plain data from verbs.

### Tests

- [x] Add `internal/cssvisualdiff/jsapi/collect_test.go`.
- [x] Test `await page.locator('#id').collect()` returns a `cvd.collectedSelection` handle.
- [x] Test `summary()`, `toJSON()`, `bounds()`, `styles()`, and `attributes()`.
- [x] Test `inspect: "minimal"` and `inspect: "rich"`.
- [x] Test raw-object rejection where a collected-selection handle is required.
- [x] Test wrong-parent feedback messages.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with the final `locator.collect()` / `cvd.collect.selection(...)` API.
- [x] Add examples showing:
  - quick collection,
  - filtering style properties after collection,
  - returning `collected.toJSON()` from a verb.

### Real smoke script

- [x] Add `scripts/004-js-collected-selection-smoke.sh`.
- [x] Script should build/run a tiny repository-scanned JS verb that:
  - opens a local HTML page,
  - collects one selector with `inspect: "rich"`,
  - returns `summary()` and filtered styles,
  - writes optional JSON output.
- [x] Script should assert expected JSON keys with `jq` or a small Python snippet.
- [x] Record smoke output in the diary.

---

## Phase 5 — JavaScript `SelectionComparison` handle

Goal: expose comparison between two collected selections as a rich Go-backed object for JavaScript analysis.

### Implementation tasks

- [x] Add `internal/cssvisualdiff/jsapi/compare.go` if not already present.
- [x] Install `cvd.compare.selections(leftCollected, rightCollected, options?)`.
- [x] Add Go-backed Proxy owner `cvd.selectionComparison`.
- [x] Implement methods/namespaces:
  - `summary()`,
  - `toJSON(options?)`,
  - `left()` / `right()` if useful,
  - `pixel.summary()`,
  - `bounds.diff()`,
  - `styles.diff(propsOrPreset?)`,
  - `attributes.diff(names?)`,
  - `report.markdown(options?)`,
  - `report.writeMarkdown(path, options?)`,
  - `artifact(name)`,
  - `artifacts.list()`,
  - `artifacts.write(outDir, names?)`.
- [x] Ensure all returned diffs are plain JavaScript arrays/maps that users can filter/map/reduce.
- [x] Add `cvd.styles.presets` minimal set if needed by docs/tests:
  - `typography`,
  - `spacing`,
  - `layout`,
  - `surface`,
  - `interaction`.
- [x] Add tailored type errors:
  - if user passes a locator, suggest `await locator.collect()` or `cvd.compare.region(...)`,
  - if user passes raw JSON, suggest `cvd.compare.data(...)` only if such a function exists; otherwise reject.

### Tests

- [x] Add `internal/cssvisualdiff/jsapi/compare_test.go`.
- [x] Test comparing two collected selections.
- [x] Test `summary()` and `toJSON()` lowerCamel shape.
- [x] Test `styles.diff()` with no filter and with presets.
- [x] Test `bounds.diff()` and `attributes.diff()`.
- [x] Test report markdown generation.
- [x] Test artifact listing/writing with parent directory creation.
- [x] Test helpful type mismatch messages.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with `cvd.compare.selections(...)` and the `SelectionComparison` handle.
- [x] Include examples for:
  - custom JavaScript filtering,
  - policy classification,
  - writing selective artifacts,
  - returning `comparison.summary()` or `comparison.toJSON()`.

### Real smoke script

- [x] Add `scripts/005-js-selection-comparison-smoke.sh`.
- [x] Script should run a repository-scanned JS verb that:
  - opens two local pages,
  - collects one selector from each,
  - compares collected selections,
  - returns pixel/style/bounds summaries,
  - writes selected artifacts.
- [x] Script should assert `diffComparison`, `compare.json`, and `compare.md` exist when requested.
- [x] Record smoke output in the diary.

---

## Phase 6 — Opinionated `cvd.compare.region(...)` low-effort surface

Goal: provide the simple path as a deliberately opinionated collect-then-compare helper, not as a Go-mode wrapper.

### Implementation tasks

- [x] Implement `cvd.compare.region({ left, right, name?, threshold?, inspect?, outDir? })` as:
  1. collect left locator,
  2. collect right locator,
  3. compare selections,
  4. return `cvd.selectionComparison` handle.
- [x] Require strict locator handles for `left` and `right`.
- [x] Use `inspect: "rich"` by default.
- [x] Use a sensible default threshold.
- [x] Do not accept loose `{ page, selector }` objects as the primary API.
- [x] Provide tailored error messages that tell users to call `page.locator(selector)`.
- [x] Ensure comparing two locators on different pages does not deadlock; collect each side under its page's `runExclusive` without nested locks.
- [x] Ensure comparing two locators on the same page is serialized and deterministic.

### Tests

- [x] Add JS API tests for `cvd.compare.region(...)`.
- [x] Test successful one-call comparison.
- [x] Test same-page comparison.
- [x] Test separate-page comparison.
- [x] Test raw object rejection.
- [x] Test default rich inspection can be queried after the comparison.
- [x] Test artifact writing from the returned comparison handle.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` to present `cvd.compare.region(...)` as the first low-effort comparison API.
- [x] Explain that it is equivalent to collect-left, collect-right, compare-selections.
- [x] Include one short example and one expanded equivalent primitive example.

### Real smoke script

- [x] Add `scripts/006-js-compare-region-smoke.sh`.
- [x] Script should run a repository-scanned JS verb using only `cvd.compare.region(...)`.
- [x] Script should serve two local fixture pages with a known visual/style difference.
- [x] Script should assert JSON output includes `changedPercent`, `styles`, and `bounds` summaries.
- [x] Script should assert selected artifacts exist.
- [x] Record smoke output in the diary.

---

## Phase 7 — Canonical namespace cleanup, no backward-compat aliases

Goal: make the public JS API coherent and explicit. Since backward compatibility is not required, do not preserve old ambiguous names unless they are intentionally selected as low-effort APIs.

### Implementation tasks

- [x] Decide final canonical namespaces and update module exports accordingly:
  - `cvd.collect.selection`,
  - `cvd.compare.selections`,
  - `cvd.compare.region`,
  - `cvd.image.diff`,
  - `cvd.diff.structural`,
  - `cvd.snapshot.page`,
  - `cvd.catalog.create`,
  - `cvd.config.load`,
  - future `cvd.job.fromConfig`.
- [x] Remove or hide public exposure of ambiguous old top-level names if they are not intentionally retained.
- [x] Keep `require("diff")` and `require("report")` internal only, or remove them once built-in scripts no longer need them.
- [x] Update error messages to reference canonical names only.
- [x] Update examples under `examples/verbs/` to use canonical names.

### Tests

- [x] Update jsapi tests to use canonical names.
- [x] Update verbcli tests/examples to use canonical names.
- [x] Add tests that internal helper modules are not documented or not user-facing, depending on final implementation.

### JavaScript API reference update

- [x] Rewrite `internal/cssvisualdiff/doc/topics/javascript-api.md` around the canonical namespace map.
- [x] Remove compatibility-alias language.
- [x] Add a “quick path vs primitive path” section:
  - quick: `cvd.compare.region`,
  - primitive: `locator.collect` + `cvd.compare.selections`.

### Real smoke script

- [x] Add `scripts/007-canonical-api-surface-smoke.sh`.
- [x] Script should run `css-visual-diff help javascript-api` and assert canonical names appear.
- [x] Script should run a JS verb using canonical names only.
- [x] Script should fail if examples still use internal `require("diff")` or `require("report")` public patterns.
- [x] Record smoke output in the diary.

---

## Phase 8 — Rework built-in compare verbs to dogfood public primitives

Goal: make built-in JS verbs prove the public API is sufficient.

### Implementation tasks

- [x] Rewrite `internal/cssvisualdiff/dsl/scripts/compare.js` to use `require("css-visual-diff")` public primitives.
- [x] Implement `script compare region` using `cvd.compare.region(...)`.
- [x] Implement `script compare brief` using the returned comparison object's report/summary APIs.
- [x] Remove direct use of internal `require("diff").compareRegion` from built-in scripts.
- [x] Decide whether internal `require("diff")` remains only for legacy private tests or is removed.
- [x] Ensure generated command fields still map cleanly to the new public API concepts.

### Tests

- [x] Update `internal/cssvisualdiff/dsl/host_test.go` for new output shape if intentionally changed.
- [x] Update `internal/cssvisualdiff/verbcli/command_test.go` for canonical public API usage.
- [x] Add regression tests for `verbs script compare region` artifact output.
- [x] Add regression tests for `verbs script compare brief` text output.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-verbs.md` to explain that built-in compare verbs are ordinary examples of the public JS API.
- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` to link from `cvd.compare.region` to the built-in verb example.

### Real smoke script

- [x] Add `scripts/008-built-in-compare-dogfood-smoke.sh`.
- [x] Script should run `css-visual-diff verbs script compare region` against two local pages.
- [x] Script should verify artifacts and output fields.
- [x] Script should run `css-visual-diff verbs script compare brief` and assert meaningful text output.
- [x] Record smoke output in the diary.

---

## Phase 9 — Catalog, reports, and artifact integration

Goal: make rich comparison objects useful in catalogs and reports, not only standalone scripts.

### Implementation tasks

- [x] Redesign catalog public API as `cvd.catalog.create(options)` if selected as canonical.
- [x] Add `catalog.record(comparison)` or equivalent for `cvd.selectionComparison` handles.
- [x] Ensure catalog manifests can include comparison summaries and artifact paths.
- [x] Add report rendering for comparison objects:
  - summary,
  - pixel stats,
  - bounds diffs,
  - filtered style diffs,
  - artifact links.
- [x] Ensure `comparison.artifacts.write(...)` and `catalog.artifactDir(...)` work together.
- [x] Ensure all file writers create parent directories.

### Tests

- [x] Add jsapi catalog/comparison integration tests.
- [x] Add service catalog tests for comparison records if catalog service changes.
- [x] Test manifest JSON contains comparison records.
- [x] Test index Markdown links to comparison artifacts.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with catalog + comparison integration.
- [x] Add an example that compares multiple sections and records them into a catalog.

### Real smoke script

- [x] Add `scripts/009-comparison-catalog-smoke.sh`.
- [x] Script should run a JS verb that:
  - compares two or more sections,
  - writes artifacts,
  - records comparisons in a catalog,
  - writes manifest and index.
- [x] Script should assert manifest/index/artifact files exist and contain expected entries.
- [x] Record smoke output in the diary.

---

## Phase 10 — Pixel accuracy guide and public examples refresh

Goal: update user-facing docs/examples so the new API is easy to learn.

### Implementation tasks

- [x] Update `examples/verbs/low-level-inspect.js` or replace it with canonical examples.
- [x] Add `examples/verbs/compare-region.js` using `cvd.compare.region`.
- [x] Add `examples/verbs/collect-and-analyze.js` using `locator.collect` and `cvd.compare.selections`.
- [x] Update `examples/verbs/README.md`.
- [x] Update README snippets if they mention older API names.

### Tests

- [x] Add/adjust binary smoke tests for examples.
- [x] Ensure examples run without external project dependencies.
- [x] Ensure examples use canonical API names only.

### JavaScript API reference update

- [x] Update `internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md` to teach:
  - quick comparison path,
  - collect-then-compare path,
  - rich inspection filtering,
  - selective artifact writing,
  - returning plain summaries from verbs.
- [x] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` examples to match the final API.

### Real smoke script

- [x] Add `scripts/010-public-examples-smoke.sh`.
- [x] Script should run every public example intended to be executable.
- [x] Script should validate expected output/artifacts.
- [x] Record smoke output in the diary.

---

## Phase 11 — Optional config/job bridge after primitive API lands

Goal: only after the new primitive API is stable, design and implement YAML/job interop if still needed.

### Implementation tasks

- [ ] Decide canonical namespace: likely `cvd.config.load(path)` and `cvd.job.fromConfig(path)`.
- [ ] Add `JobHandle` backed by Go services/runner logic, not JS reimplementation of YAML behavior.
- [ ] Support job operations if in scope:
  - `job.plan()`,
  - `job.preflight(options)`,
  - `job.run(options)`,
  - `job.collect(options)` if it can produce collected selections/comparisons.
- [ ] Ensure job outputs can be converted to `CollectedSelection`, `SelectionComparison`, snapshots, or catalog records where appropriate.

### Tests

- [ ] Add config/job service tests.
- [ ] Add jsapi job tests with a tiny YAML config.
- [ ] Test error messages for unsupported modes/options.

### JavaScript API reference update

- [ ] Update `internal/cssvisualdiff/doc/topics/javascript-api.md` with config/job bridge only after implementation is real.
- [ ] Explain when to use JS primitives versus YAML job bridge.

### Real smoke script

- [ ] Add `scripts/011-config-job-bridge-smoke.sh` if Phase 11 is implemented.
- [ ] Script should run a tiny YAML config through the JS job API and validate output.
- [ ] Record smoke output in the diary.

---

## Phase 12 — Final validation and delivery

Goal: prove the whole API works as a coherent surface.

### Implementation tasks

- [ ] Run `gofmt` on changed Go files.
- [ ] Run `go test ./... -count=1`.
- [ ] Run all ticket smoke scripts in numeric order.
- [ ] Run embedded help rendering:
  - `go run ./cmd/css-visual-diff help javascript-api`,
  - `go run ./cmd/css-visual-diff help javascript-verbs`,
  - `go run ./cmd/css-visual-diff help pixel-accuracy-scripting-guide`.
- [ ] Run `docmgr doctor --root ./ttmp --ticket CSSVD-JSAPI-PIXEL-WORKFLOWS --stale-after 30`.
- [ ] Update changelog with implementation summary.
- [ ] Update investigation diary with commands, failures, fixes, and final validation output.
- [ ] Upload final implementation/design bundle to reMarkable.
- [ ] Prepare PR summary explaining:
  - no-backward-compat canonical surface,
  - low-effort API path,
  - primitive API path,
  - robust Go-backed handles and tailored errors,
  - smoke scripts and validation.

### JavaScript API reference final check

- [ ] Ensure `internal/cssvisualdiff/doc/topics/javascript-api.md` contains the final canonical API and no stale compatibility guidance.
- [ ] Ensure examples in docs match executable examples.
- [ ] Ensure every public object has a brief reference entry:
  - Browser,
  - Page,
  - Locator,
  - CollectedSelection,
  - SelectionComparison,
  - Probe,
  - Snapshot,
  - StructuralDiff,
  - ImageDiff,
  - Catalog,
  - Config/Job if implemented.

### Final smoke script

- [ ] Add `scripts/012-run-all-smokes.sh`.
- [ ] Script should run scripts `001` through `011` where present.
- [ ] Script should fail fast and print the failing script name.
- [ ] Record final smoke output in the diary.
