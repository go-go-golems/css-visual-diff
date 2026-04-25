---
Title: Investigation diary
Ticket: CSSVD-CODE-QUALITY-REVIEW
Status: active
Topics:
  - tooling
  - backend
  - frontend
  - visual-regression
DocType: reference
Intent: long-term
Owners: []
---

# Investigation diary

## 2026-04-25 — Initial lint and architecture/code-quality review

### User request

Run `make lint`, review the codebase for deprecated code, duplication, confusing parts, huge files/packages, and over-complex areas, then create a new code review ticket with a detailed analysis/design/implementation guide for a new intern and upload it to reMarkable.

### Commands run

```bash
make lint
```

Result: failed with 19 issues. Full output saved to `reference/01-make-lint-output.txt`.

```bash
docmgr ticket create-ticket --root ./ttmp \
  --ticket CSSVD-CODE-QUALITY-REVIEW \
  --title "Code quality review and cleanup plan for css-visual-diff" \
  --topics tooling,backend,frontend,visual-regression
```

Created local ticket workspace under `ttmp/2026/04/25/CSSVD-CODE-QUALITY-REVIEW--code-quality-review-and-cleanup-plan-for-css-visual-diff`.

```bash
rg --files -g '*.go' | wc -l
rg --files internal -g '*.go' | wc -l
rg --files -g '*.js' | wc -l
```

Observed 90 Go files, 87 internal Go files, and 6 JavaScript files.

```bash
python3 - <<'PY'
from pathlib import Path
rows=[]
for p in Path('.').rglob('*.go'):
    if any(part in {'.git','.bin'} for part in p.parts):
        continue
    n=sum(1 for _ in p.open())
    rows.append((n,str(p)))
for n,p in sorted(rows, reverse=True)[:30]:
    print(f'{n:5d} {p}')
print('TOTAL_GO_LOC', sum(n for n,_ in rows))
PY
```

Observed roughly 15,230 Go lines. Largest files include `verbcli/command_test.go`, `cmd/css-visual-diff/main.go`, `modes/matched_styles.go`, `jsapi/module.go`, `modes/inspect.go`, and `service/inspect.go`.

### Files inspected

- `cmd/css-visual-diff/main.go`
- `internal/cssvisualdiff/runner/runner.go`
- `internal/cssvisualdiff/modes/modes.go`
- `internal/cssvisualdiff/modes/inspect.go`
- `internal/cssvisualdiff/modes/prepare.go`
- `internal/cssvisualdiff/modes/pixeldiff_util.go`
- `internal/cssvisualdiff/modes/capture.go`
- `internal/cssvisualdiff/modes/matched_styles.go`
- `internal/cssvisualdiff/service/inspect.go`
- `internal/cssvisualdiff/service/pixel.go`
- `internal/cssvisualdiff/service/style.go`
- `internal/cssvisualdiff/service/browser.go`
- `internal/cssvisualdiff/service/catalog_service.go`
- `internal/cssvisualdiff/jsapi/module.go`
- `internal/cssvisualdiff/jsapi/proxy.go`
- `internal/cssvisualdiff/jsapi/compare.go`
- `internal/cssvisualdiff/jsapi/catalog.go`
- `internal/cssvisualdiff/verbcli/bootstrap.go`
- `internal/cssvisualdiff/dsl/host.go`

### Findings

- Lint failures are mostly cleanup debt after the service extraction.
- `modes/inspect.go` contains stale duplicate artifact writer functions that now have canonical exported equivalents in `service/inspect.go`.
- `modes/pixeldiff_util.go` and `modes/prepare.go` contain unused service wrapper functions.
- `cmd/css-visual-diff/main.go` is too large and should eventually be split into command packages.
- `jsapi/module.go` mixes module install, browser/page wrappers, error mapping, and page serialization; it should eventually be split after lint cleanup.
- The core architecture is sound: `service` should own reusable browser/DOM/image/comparison/catalog behavior, while `modes` and `jsapi` should adapt config/JS calls to services.

### Output document

Wrote:

```text
design/01-code-quality-review-cleanup-and-intern-architecture-guide.md
```

The document contains:

- architecture map,
- runtime flow for YAML runs,
- runtime flow for JS verbs,
- package map and line-count hotspots,
- lint issue analysis,
- duplicated/deprecated code findings,
- concrete cleanup sketches,
- pseudocode,
- diagrams,
- intern reading path,
- lint fix checklist.

### Next steps

1. Validate ticket with `docmgr doctor`.
2. Upload the design document to reMarkable.
3. If requested, implement Phase A lint cleanup in a separate code change.

## 2026-04-25 — reMarkable upload

### Commands run

```bash
remarquee status
remarquee upload md --dry-run \
  ttmp/2026/04/25/CSSVD-CODE-QUALITY-REVIEW--code-quality-review-and-cleanup-plan-for-css-visual-diff/design/01-code-quality-review-cleanup-and-intern-architecture-guide.md \
  --remote-dir /ai/2026/04/25/CSSVD-CODE-QUALITY-REVIEW \
  --name cssvd-code-quality-review-guide

remarquee upload md \
  ttmp/2026/04/25/CSSVD-CODE-QUALITY-REVIEW--code-quality-review-and-cleanup-plan-for-css-visual-diff/design/01-code-quality-review-cleanup-and-intern-architecture-guide.md \
  --remote-dir /ai/2026/04/25/CSSVD-CODE-QUALITY-REVIEW \
  --name cssvd-code-quality-review-guide

remarquee cloud ls /ai/2026/04/25/CSSVD-CODE-QUALITY-REVIEW --long --non-interactive
```

### Result

Uploaded the report to reMarkable at:

```text
/ai/2026/04/25/CSSVD-CODE-QUALITY-REVIEW/cssvd-code-quality-review-guide
```

## 2026-04-25 — Phase A lint cleanup implementation

### What changed

- Fixed `internal/cssvisualdiff/service/pixel.go` so `ReadPNG` and `WritePNG` check file close errors without named return values.
- Changed `internal/cssvisualdiff/jsapi/module.go classifyCVDError` to return unnamed `(string, string)` values.
- Replaced `WriteString(fmt.Sprintf(...))` patterns with `fmt.Fprintf(...)` in:
  - `internal/cssvisualdiff/jsapi/compare.go`,
  - `internal/cssvisualdiff/modes/html_report.go`,
  - `internal/cssvisualdiff/service/catalog_service.go`,
  - `internal/cssvisualdiff/service/diff.go`.
- Simplified `internal/cssvisualdiff/modes/capture.go selectorForSection` with a side switch.
- Simplified `internal/cssvisualdiff/service/style.go` by converting `styleEvalResult` to `StyleSnapshot` directly.
- Renamed `internal/cssvisualdiff/modes/matched_styles.go scanEnclosed` parameters from `open, close` to `openDelim, closeDelim`.
- Removed stale duplicate inspect artifact helpers from `internal/cssvisualdiff/modes/inspect.go`; kept a small test-facing wrapper for `inspectFormatRequiresExistingSelector` delegating to `service`.
- Removed unused service wrappers from `internal/cssvisualdiff/modes/pixeldiff_util.go` and `internal/cssvisualdiff/modes/prepare.go`.

### Validation

```bash
go test ./... -count=1
```

Passed.

```bash
make lint
```

Passed with `0 issues`.

### Issues encountered

- After deleting the stale inspect helpers, `modes/inspect_test.go` still referenced `inspectFormatRequiresExistingSelector`. I restored a tiny delegating wrapper to keep existing tests focused on mode behavior while leaving the real implementation in `service`.
- A broad replacement in `html_report.go` initially used `fmt.Fprintf(b, ...)` where `b` was a value `strings.Builder`; this failed typechecking because only `*strings.Builder` implements `io.Writer`. I corrected those top-level calls to `fmt.Fprintf(&b, ...)` while leaving helper functions that receive `*strings.Builder` as `fmt.Fprintf(b, ...)`.
- Once the original 19 lint issues were fixed, additional `QF1012` issues surfaced because golangci-lint's `max-same-issues` had hidden them. I fixed those too until `make lint` reported `0 issues`.
