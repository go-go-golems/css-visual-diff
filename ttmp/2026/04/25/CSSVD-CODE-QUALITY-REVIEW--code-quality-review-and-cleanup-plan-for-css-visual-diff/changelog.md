---
Title: Changelog
Ticket: CSSVD-CODE-QUALITY-REVIEW
Status: active
Topics:
  - tooling
  - backend
  - frontend
  - visual-regression
DocType: changelog
Intent: long-term
Owners: []
---

# Changelog

## 2026-04-25

- Created ticket `CSSVD-CODE-QUALITY-REVIEW` for a code quality and cleanup review of `css-visual-diff`.
- Ran `make lint`; saved the failing output with 19 issues to `reference/01-make-lint-output.txt`.
- Inventoried repository size: 90 Go files, 87 internal Go files, 6 JavaScript files, and about 15,230 Go lines.
- Inspected architecture and lint hotspots across CLI, runner, modes, service, JS API, DSL, and verbcli packages.
- Added `design/01-code-quality-review-cleanup-and-intern-architecture-guide.md`, a detailed architecture map and cleanup plan for a new intern.
- Added `reference/02-investigation-diary.md` with commands, findings, and next steps.
- Rewrote `tasks.md` into a phased cleanup plan.
- Uploaded the code quality review guide to reMarkable at `/ai/2026/04/25/CSSVD-CODE-QUALITY-REVIEW/cssvd-code-quality-review-guide`.
- Implemented Phase A lint cleanup: fixed PNG close error handling, removed stale mode-local inspect/pixel/prepare wrappers, applied staticcheck cleanups across JS comparison Markdown, HTML/catalog/diff renderers, style conversion, selector branching, and matched-style parameter naming. `make lint` now reports 0 issues and `go test ./... -count=1` passes.
