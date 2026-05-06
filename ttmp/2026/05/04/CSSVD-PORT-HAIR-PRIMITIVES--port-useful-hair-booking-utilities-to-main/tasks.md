# Tasks

## TODO

- [ ] [Phase 0: setup] Start from a clean origin/main branch; do not continue the conflicted hair-booking merge.
- [ ] [Phase 0: source audit] Use git show against bookmark/2026-05-01/hair-booking for only the accepted primitive files/hunks.
- [ ] [Phase 0: guardrails] Record APIs that must not be ported: cvd.compare.region, cvd.compare.selections, locator.collect, cvd.collect.selection, old YAML config, and old runner/modes.
- [x] [Phase 1: locator.waitFor] Port/adapt WaitForSelectorOptions and wait-loop service logic into internal/cssvisualdiff/service/dom.go.
- [x] [Phase 1: locator.waitFor] Expose locator.waitFor(options) in internal/cssvisualdiff/jsapi/locator.go using the existing promise/runExclusive/error model.
- [x] [Phase 1: locator.waitFor] Add tests for immediate existence, delayed appearance, hidden elements with visible true/false, invalid selector errors, and timeout behavior.
- [x] [Phase 1: locator.waitFor] Update JavaScript API docs and add or adapt a small example showing waitFor before extraction.
- [ ] [Phase 2: pixel service] Copy/adapt internal/cssvisualdiff/service/pixel.go and pixel_test.go from the branch as an internal service, not a public JS API.
- [ ] [Phase 2: pixel service] Refactor modes/pixeldiff_util.go and only the necessary parts of modes/compare.go to delegate to the new pixel service.
- [ ] [Phase 2: pixel service] Preserve origin/main compare-region JSON shape, especially snake_case PixelDiffStats fields and existing artifact names.
- [ ] [Phase 2: pixel service] Run go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes and capture any regressions in the ticket changelog.
- [ ] [Phase 3: artifact paths] Inspect actual require("diff").compareRegion JSON output and verify paths for left/right screenshots, diff_only.png, diff_comparison.png, compare.json, and compare.md are discoverable.
- [ ] [Phase 3: artifact paths] If any artifact paths are missing, add stable fields to the existing result shape without introducing comparison.artifacts proxies.
- [ ] [Phase 3: artifact paths] Add a CLI or JS smoke that writes PNG/JSON/Markdown artifacts and asserts returned paths point to existing files.
- [ ] [Phase 4: optional internal utilities] Evaluate selection_compare.go algorithms for deterministic map diffing, bounds deltas, and text changes; port only small internal helpers that simplify existing code.
- [ ] [Phase 4: optional internal utilities] Keep semantic diff helper work out of public jsapi unless maintainers explicitly approve a convenience layer.
- [ ] [Phase 5: Pyxis migration] Update Pyxis userland/lib/compare-region.js to use require("diff").compareRegion for pixel comparisons against origin/main.
- [ ] [Phase 5: Pyxis migration] Replace hair-booking comparison/categorization assumptions with Pyxis-side row/catalog summary construction or existing catalog.addResult where appropriate.
- [ ] [Phase 5: Pyxis migration] Fix Pyxis acceptedDifferences merge order across spec, target, and section scopes as a separate Pyxis-side bug.
- [ ] [Phase 5: validation] Run targeted Go tests, compare-region smoke, and Pyxis userland smoke scripts; document commands and results in the ticket.
