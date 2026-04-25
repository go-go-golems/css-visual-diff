---
Title: Investigation diary
Ticket: CSSVD-BETA-ERGONOMICS
Status: active
Topics:
  - tooling
  - frontend
  - visual-regression
  - browser-automation
DocType: reference
Intent: long-term
Owners: []
---

# Investigation diary

## 2026-04-25 — Ticket creation and scoped beta ergonomics design

### User request

Create a new ticket to address the high-priority items from the Pyxis maintainer follow-up request document:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md
```

The user emphasized that we want to avoid adding too much complexity but still help beta users.

### Source read

Read the full Pyxis follow-up request document. Main source priorities:

- P0: selector wait helper for JS scripts.
- P0: clarify artifact write/report schema.
- P1: multi-section comparison/catalog example.
- P1: document collection profiles and defaults.
- P2: bounds tolerances.
- P2: CSS/style normalization hooks.
- P3: style property presets.
- P3: no-op prepare documentation.

### Key decision

Scope this css-visual-diff ticket to the low-complexity, high-value beta ergonomics work:

1. Implement `locator.waitFor(...)` and optionally `page.waitForSelector(...)`.
2. Make `comparison.artifacts.write(...)` return a stable artifact path map.
3. Add a complete multi-section catalog example and smoke.
4. Clarify collection profile docs.

Explicitly defer tolerances, normalization hooks, and style presets. Those are useful ideas, but they risk adding policy/framework complexity before beta usage stabilizes.

### Code inspected

- `internal/cssvisualdiff/jsapi/locator.go`
  - Current locator methods: `status`, `exists`, `visible`, `text`, `bounds`, `computedStyle`, `attributes`, `collect`.
  - Natural home for `locator.waitFor`.
- `internal/cssvisualdiff/jsapi/module.go`
  - Current page methods and `pageState.runExclusive` serialization.
  - Possible home for a tiny `page.waitForSelector` convenience method.
- `internal/cssvisualdiff/service/dom.go`
  - Current DOM/locator service functions.
  - Natural home for `WaitForLocator` using existing `LocatorStatus`.
- `internal/cssvisualdiff/jsapi/compare.go`
  - Current `cvd.compare.region` implementation and `comparison.artifacts.write` implementation.
- `internal/cssvisualdiff/service/collection.go`
  - Collection profile constants and options; docs should reflect actual normalization behavior.

### Ticket created

```bash
docmgr ticket create-ticket --root ./ttmp \
  --ticket CSSVD-BETA-ERGONOMICS \
  --title "Beta user ergonomics for JavaScript visual workflows" \
  --topics tooling,frontend,visual-regression,browser-automation
```

### Documents written

- `design/01-beta-user-ergonomics-for-js-visual-workflows.md`
- `tasks.md`
- `changelog.md`
- `reference/01-investigation-diary.md`

### Next steps

1. Relate source files and source Pyxis doc to the ticket.
2. Run `docmgr doctor`.
3. Implement Phase 1 selector wait helper first.

## 2026-04-25 — Phase 1 selector wait helper

### What changed

- Added `WaitForSelectorOptions` and `WaitForSelectorResult` in `internal/cssvisualdiff/service/dom.go`.
- Added `service.WaitForLocator(page, locator, opts)`, implemented as a small polling loop over existing `LocatorStatus`.
- Added `locator.waitFor(options?)` to the Go-backed JS locator Proxy.
- Added `page.waitForSelector(selector, options?)` as a tiny convenience wrapper over the same service function.
- Updated `javascript-api` docs and the pixel-accuracy scripting guide with the readiness-wait pattern.
- Extended service DOM tests for:
  - already-existing selectors,
  - delayed selectors,
  - visible waits,
  - timeout errors,
  - invalid selectors.
- Extended the repository-scanned JS locator integration test to call both `locator.waitFor(...)` and `page.waitForSelector(...)`.

### Design notes

- The helper intentionally does not introduce a workflow builder. It only fills the missing primitive: “wait until this selector is ready”.
- `locator.waitFor(...)` is the preferred API because it keeps readiness tied to the same locator that will be inspected or compared.
- `page.waitForSelector(...)` exists for users coming from Playwright-style APIs and delegates to the same service behavior.
- Per-page serialization still applies because the JS wrappers call `pageState.runExclusive(...)`.

### Validation

```bash
go test ./... -count=1
make lint
```

Both passed. `make lint` reported `0 issues`.

### Issues encountered

- The first JS integration test attempted to read `waited.exists` from a Go struct exported directly to Goja. The exported object used Go field names, so `waited.exists` was undefined. I fixed the JS wrappers to lower `WaitForSelectorResult` through `lowerJSON(...)`, producing lowerCamel JSON fields.

## 2026-04-25 — Phase 2 stable artifact write result

### What changed

- Updated `internal/cssvisualdiff/jsapi/compare.go` so `comparison.artifacts.write(outDir, ["json", "markdown"])` returns stable keyed paths:
  - `json`,
  - `markdown`,
  - `leftRegion`,
  - `rightRegion`,
  - `diffOnly`,
  - `diffComparison`,
  - `written`,
  - `outDir`.
- The function still only writes requested JSON/Markdown files. It does not re-render PNGs.
- Known PNG paths are returned from `SelectionComparisonData.Artifacts`, pixel diff paths, and the standard `cvd.compare.region(...)` output filenames when those files exist in `outDir`.
- Extended the low-effort compare-region verb integration test to assert returned artifact paths and file existence.
- Updated `javascript-api` and `pixel-accuracy-scripting-guide` docs with the result shape and project-local CLI summary pattern.

### Validation

```bash
go test ./... -count=1
make lint
```

Both passed. `make lint` reported `0 issues`.

## 2026-04-25 — Phase 3 multi-section catalog example

### What changed

- Added `examples/verbs/compare-page-catalog.js`.
- The example loads left/right pages once, waits for each section selector with `locator.waitFor(...)`, compares `page` and `cta` sections, writes per-section JSON/Markdown/PNG artifacts, records comparisons into a catalog, and returns compact JSON.
- Updated `examples/verbs/README.md` with the command, expected artifacts, and why it demonstrates the project-local CLI pattern.
- Added `scripts/001-beta-multisection-example-smoke.sh` under this ticket.

### Validation

```bash
ttmp/2026/04/25/CSSVD-BETA-ERGONOMICS--beta-user-ergonomics-for-javascript-visual-workflows/scripts/001-beta-multisection-example-smoke.sh
```

Passed. The smoke validates:

- stdout JSON has two summaries,
- catalog comparison count is 2,
- each summary includes artifact paths from `comparison.artifacts.write(...)`,
- `manifest.json` and `index.md` exist,
- per-section PNG/JSON/Markdown artifacts exist for `page` and `cta`.

```bash
go test ./... -count=1
make lint
```

Both passed. `make lint` reported `0 issues`.
