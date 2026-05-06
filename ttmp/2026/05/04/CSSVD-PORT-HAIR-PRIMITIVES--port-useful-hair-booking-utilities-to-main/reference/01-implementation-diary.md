---
Title: Implementation Diary
Ticket: CSSVD-PORT-HAIR-PRIMITIVES
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - merge-planning
    - visual-diff
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological notes for porting useful hair-booking primitives into origin/main.
LastUpdated: 2026-05-04T17:03:04.015345641-04:00
WhatFor: Track what was copied, studied, planned, implemented, tested, and deferred while porting useful hair-booking primitives.
WhenToUse: Before resuming implementation work on this ticket or reviewing task status.
---

# Implementation Diary

## Goal

Maintain a chronological record for porting only the useful primitives from the divergent hair-booking branch into `origin/main`, while preserving the current smaller css-visual-diff JavaScript API model.

## Context

The source guide was copied from the previous branch workspace into this ticket:

- Source: `/home/manuel/workspaces/2026-04-21/hair-v2/css-visual-diff/ttmp/2026/05/04/CSSVD-MAIN-PORT-GUIDE--port-useful-hair-booking-primitives-to-origin-main/design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md`
- Local copy: `../design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md`

The guide recommends porting `locator.waitFor(...)`, extracting internal pixel service utilities, auditing artifact path stability, and migrating Pyxis userland away from branch-only APIs. It explicitly recommends not porting `cvd.compare.region`, `cvd.compare.selections`, `locator.collect`, `cvd.collect.selection`, old YAML config, or old runner/native modes.

## Diary

### 2026-05-04 — Ticket setup and phase planning

- Created ticket `CSSVD-PORT-HAIR-PRIMITIVES` with topics `css-visual-diff`, `javascript-api`, `merge-planning`, and `visual-diff`.
- Copied the existing origin-main port guide with `cp` into `design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md`.
- Updated the copied document's `Ticket` frontmatter to this ticket so `docmgr doc list` discovers it under the new workspace.
- Studied the guide's decision matrix and implementation plan.
- Added a phased task list covering setup guardrails, `locator.waitFor`, internal pixel service extraction, artifact path auditing, optional semantic diff utilities, Pyxis migration, and validation.

### 2026-05-04 — Phase 1 implementation: `locator.waitFor`

- Read the hair-booking branch implementations from `bookmark/2026-05-01/hair-booking:internal/cssvisualdiff/service/dom.go`, `jsapi/locator.go`, and `service/dom_test.go`.
- Ported the selector wait primitive into `internal/cssvisualdiff/service/dom.go` with `WaitForSelectorOptions`, `WaitForSelectorResult`, and `WaitForLocator`.
- Chose origin/main semantics from the guide: `visible` defaults to required visibility, and callers can pass `{ visible: false }` to wait only for DOM presence.
- Added `locator.waitFor(options?)` to `internal/cssvisualdiff/jsapi/locator.go`, using the same Promise, `runExclusive`, and CVD error model as existing locator methods.
- Added service tests for immediate selectors, delayed selectors, hidden selectors with default visibility, hidden selectors with `Visible: false`, missing selector timeouts, and invalid selector errors.
- Extended the runtime module smoke in `internal/cssvisualdiff/verbcli/command_test.go` to call `locator.waitFor` from JavaScript.
- Updated `internal/cssvisualdiff/doc/topics/javascript-api.md` and `examples/verbs/low-level-inspect.js` so authors see the new wait primitive before extraction.
- Validation passed: `go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli`.

## Quick Reference

Implementation should proceed in this order:

1. Clean `origin/main` setup and branch-source audit.
2. Port `locator.waitFor` as a locator primitive.
3. Port `service/pixel.go` as an internal service and preserve existing output JSON.
4. Audit compare-region artifact paths and fill gaps without adding branch comparison proxies.
5. Optionally extract small deterministic semantic diff helpers only if they simplify existing code.
6. Adapt Pyxis userland to `require("diff").compareRegion` and current `cvd.extract`/`cvd.diff` primitives.
7. Run targeted Go tests, compare-region smoke tests, and Pyxis smoke scripts.

## Related

- [Origin/main port guide](../design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md)
- [Task list](../tasks.md)
