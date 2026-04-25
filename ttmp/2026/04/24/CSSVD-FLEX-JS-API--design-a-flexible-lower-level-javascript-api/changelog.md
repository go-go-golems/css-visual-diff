# Changelog

## 2026-04-24

- Initial workspace created.
- Created the primary design document: `design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md`.
- Documented the current architecture from `README.md`, `docs/js-api.md`, `docs/js-verbs.md`, `internal/cssvisualdiff/config`, `internal/cssvisualdiff/dsl`, `internal/cssvisualdiff/verbcli`, and `internal/cssvisualdiff/service`.
- Proposed a layered lower-level JavaScript API: page locators, extractors, snapshots, fluent builders, diffs, reporting, workflows, and catalog integration.
- Added a phased implementation plan suitable for an intern to follow.
- Updated `tasks.md` and `reference/01-investigation-diary.md`.
- Ran `docmgr doctor --root ./ttmp --ticket CSSVD-FLEX-JS-API --stale-after 30`; after normalizing ticket topics to known vocabulary, all checks passed.
- Generated local PDF bundle at `pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide.pdf`.
- Attempted reMarkable upload, but cloud create/upload returned `request failed with status 400`; upload was skipped for now at user request.
- Retried reMarkable upload with a short local PDF filename (`/tmp/cssvd-flex-api.pdf`); upload succeeded to `/ai/2026/04/24/cssvd-flex-api` and the remote listing was verified.

## 2026-04-24 — Proxy-backed JS API architecture update

- Updated the design guide to move the native `require("css-visual-diff")` implementation out of `internal/cssvisualdiff/dsl/cvd_module.go` and into a dedicated `internal/cssvisualdiff/jsapi` subpackage.
- Added a preparatory no-behavior refactor phase for moving promise/error/browser/page/catalog/config adapters into `jsapi`.
- Reworked the lower-level API design to use Go-backed Goja Proxy wrappers for live handles and DSL builders instead of raw JS object builders.
- Added method-owner, wrong-parent, unknown-method, and type-mismatch error guidance based on the Discord bot UI DSL Proxy pattern.
- Clarified that final result data should remain plain JSON-serializable data, while handles/builders/specs should be Go-backed Proxy objects.
- Clarified compatibility boundaries: existing high-level methods may continue accepting raw probe objects, while new strict lower-level APIs should require Go-backed locators/probes/extractors and reject raw objects with helpful errors.
- Removed workflow builder from the plan. Future orchestration should use ordinary JavaScript functions, loops, and branches.
- Regenerated updated local PDF bundle at `pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide-updated.pdf`.
- Uploaded updated PDF to reMarkable as `/ai/2026/04/24/cssvd-flex-api-updated` and verified the remote listing.
