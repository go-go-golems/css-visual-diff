---
Title: Changelog
Ticket: CSSVD-BETA-ERGONOMICS
Status: active
Topics:
  - tooling
  - frontend
  - visual-regression
  - browser-automation
DocType: changelog
Intent: long-term
Owners: []
---

# Changelog

## 2026-04-25

- Created ticket `CSSVD-BETA-ERGONOMICS` from Pyxis follow-up maintainer requests after `cvd.compare.region(...)` validated the original JS-callable pixel comparison need.
- Read the Pyxis source request document `03-clean-css-visual-diff-maintainer-follow-up-requests-after-flexible-js-api.md`.
- Inspected current JS API, locator/page wrappers, DOM services, comparison artifact writer, and collection profile code to ground the design in existing implementation points.
- Added `design/01-beta-user-ergonomics-for-js-visual-workflows.md`, a scoped analysis/design guide focused on beta-user ergonomics without adding unnecessary workflow complexity.
- Rewrote `tasks.md` into four implementation phases:
  1. selector wait helper,
  2. stable artifact write result,
  3. multi-section catalog example,
  4. collection profile docs.
- Explicitly deferred bounds tolerances, CSS normalization hooks, and style presets as higher-complexity ideas to revisit after more beta usage.
- Implemented Phase 1 selector wait helper: added service-level `WaitForLocator`, JS `locator.waitFor(...)`, JS `page.waitForSelector(...)`, service tests for existing/delayed/visible/timeout/invalid selectors, verbcli integration coverage, and embedded docs for readiness waits.
- Implemented Phase 2 stable artifact write result: `comparison.artifacts.write(...)` now returns keyed `json`, `markdown`, `leftRegion`, `rightRegion`, `diffOnly`, `diffComparison`, and `written` paths; updated verbcli coverage and embedded docs.
- Implemented Phase 3 multi-section catalog example: added `examples/verbs/compare-page-catalog.js`, documented it in `examples/verbs/README.md`, added `scripts/001-beta-multisection-example-smoke.sh`, and validated manifest/index/per-section artifacts/stdout artifact paths.
- Completed Phase 4 collection profile documentation: verified `normalizeCollectOptions`, documented `minimal`/`rich`/`debug`, clarified `styleProps` and `attributes` as collection filters, updated `javascript-api` and `pixel-accuracy-scripting-guide`, rendered embedded help, reran the beta smoke, `go test ./... -count=1`, and `make lint`.
