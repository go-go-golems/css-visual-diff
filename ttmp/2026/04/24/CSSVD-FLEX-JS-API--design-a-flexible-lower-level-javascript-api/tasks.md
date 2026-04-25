# Tasks

## Completed

- [x] Create ticket workspace for the flexible lower-level JavaScript API design.
- [x] Read current user-facing docs: `README.md`, `docs/js-api.md`, and `docs/js-verbs.md`.
- [x] Read the prior Goja/jsverbs implementation guide for historical context.
- [x] Read current implementation files for config schema, Goja module adapters, repository-scanned verbs, and reusable services.
- [x] Read the Discord bot Goja Proxy UI DSL references and implementation files.
- [x] Write the detailed intern-facing analysis/design/implementation guide.
- [x] Update the design guide with the `internal/cssvisualdiff/jsapi` subpackage plan.
- [x] Update the design guide with the Go-backed Proxy builder/handle model.
- [x] Remove workflow builder from the implementation plan; use ordinary JavaScript functions/loops for orchestration.
- [x] Update investigation diary and changelog.

## Validation / delivery

- [x] Run docmgr doctor for this ticket.
- [x] Generate a local PDF bundle for later transfer.
- [x] Regenerate updated PDF bundle after Proxy/jsapi design changes.
- [x] Upload the design guide bundle to reMarkable. _Initial upload: `/ai/2026/04/24/cssvd-flex-api`; updated Proxy/jsapi upload: `/ai/2026/04/24/cssvd-flex-api-updated`._
- [x] Verify the reMarkable remote listing if upload succeeds.

## Future implementation phases

- [ ] Phase 1: Move the native API out of `dsl/cvd_module.go` into `internal/cssvisualdiff/jsapi` without behavior changes.
- [ ] Phase 2: Add Proxy infrastructure and typed unwrapping helpers.
- [ ] Phase 3: Add per-page operation serialization before exposing many locator methods.
- [ ] Phase 4: Implement `service/dom.go` locator primitives.
- [ ] Phase 5: Expose `page.locator()` and async locator methods as Go-backed Proxy handles.
- [ ] Phase 6: Add Go-backed target/probe/extractor builders.
- [ ] Phase 7: Add strict `cvd.extract(locator, extractors)`.
- [ ] Phase 8: Add strict `cvd.snapshot(page, probes, options)`.
- [ ] Phase 9: Add `cvd.diff(...)`, `cvd.report(...)`, and `cvd.write.*` helpers.
