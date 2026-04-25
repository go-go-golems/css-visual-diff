---
Title: Tasks
Ticket: CSSVD-CODE-QUALITY-REVIEW
Status: active
Topics:
  - tooling
  - backend
  - frontend
  - visual-regression
DocType: tasks
Intent: short-term
Owners: []
---

# Tasks

## Completed

- [x] Run `make lint` and save output.
- [x] Inventory Go/JS file counts and large files.
- [x] Inspect lint-reported files and representative architecture files.
- [x] Write detailed intern-oriented architecture and cleanup guide.
- [x] Record investigation diary.

## Recommended implementation tasks

### Phase A — Make lint green with minimal behavior changes

- [x] Fix `service/pixel.go` `Close()` error handling.
- [x] Remove named returns in `jsapi/module.go classifyCVDError`.
- [x] Rename `modes/matched_styles.go scanEnclosed` parameter `close` to `closeDelim`.
- [x] Replace `WriteString(fmt.Sprintf(...))` with `fmt.Fprintf` in `jsapi/compare.go` Markdown renderer.
- [x] Simplify `service/style.go` `StyleSnapshot` conversion or intentionally nolint with explanation.
- [x] Simplify `modes/capture.go selectorForSection` to satisfy staticcheck.
- [x] Delete stale duplicate inspect helpers from `modes/inspect.go`.
- [x] Delete unused wrappers from `modes/pixeldiff_util.go` and `modes/prepare.go`.
- [x] Run `go test ./... -count=1`.
- [x] Run `make lint`.

### Phase B — Clarify package boundaries

- [ ] Add package comments to `service` and `modes` explaining ownership.
- [ ] Keep reusable behavior in `service`; keep `modes` as config-to-service orchestration.
- [ ] Review mode tests for direct dependency on compatibility wrappers.

### Phase C — Split large files

- [ ] Split `cmd/css-visual-diff/main.go` into internal command-definition files.
- [ ] Split `jsapi/module.go` into `errors.go`, `browser.go`, `page.go`, and module installation.
- [ ] Split `verbcli/command_test.go` by behavior and extract shared fixtures.

### Phase D — Optional mode registry

- [ ] Replace `runner.Run` string switch with a small mode registry if mode count keeps growing.
- [ ] Use canonical mode lists for `full`, help/docs, and validation.
