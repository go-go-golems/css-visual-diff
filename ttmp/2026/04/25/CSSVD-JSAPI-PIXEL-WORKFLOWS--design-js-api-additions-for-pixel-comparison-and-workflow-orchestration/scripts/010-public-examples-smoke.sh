#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[010] public examples smoke: create fixture site"
WORKDIR="$(mktemp -d)"
OUTROOT="$(mktemp -d)"
cleanup() {
  kill "${SERVER_PID:-}" 2>/dev/null || true
  rm -rf "$WORKDIR" "$OUTROOT"
}
trap cleanup EXIT

cat >"$WORKDIR/index.html" <<'HTML'
<html><body><button id="cta" class="primary" style="color: rgb(0, 0, 0); background: rgb(240, 240, 240); font-size: 16px; padding: 8px; border-radius: 4px">Book now</button></body></html>
HTML
cat >"$WORKDIR/left.html" <<'HTML'
<html><body><button id="cta" class="primary" style="color: rgb(0, 0, 0); background: rgb(240, 240, 240); font-size: 16px; padding: 8px; border-radius: 4px">Book now</button></body></html>
HTML
cat >"$WORKDIR/right.html" <<'HTML'
<html><body><button id="cta" class="secondary" style="color: rgb(255, 0, 0); background: rgb(250, 250, 250); font-size: 18px; padding: 12px; border-radius: 8px">Book now</button></body></html>
HTML
python3 -m http.server 8769 --directory "$WORKDIR" >/tmp/cssvd-public-examples-http.log 2>&1 &
SERVER_PID=$!
sleep 1

BASE=http://127.0.0.1:8769

echo "[010] public examples smoke: examples compare region"
REGION_OUT="$OUTROOT/compare-region"
go run ./cmd/css-visual-diff verbs --repository examples/verbs examples compare region \
  "$BASE/left.html" "$BASE/right.html" '#cta' "$REGION_OUT" --output json >/tmp/cssvd-example-compare-region.json
python3 - <<PY
import json, pathlib
rows=json.loads(pathlib.Path('/tmp/cssvd-example-compare-region.json').read_text())
row=rows[0]
assert row['ok'] is True, row
assert row['schemaVersion'] == 'cssvd.selectionComparison.v1', row
assert row['changedPercent'] > 0, row
PY
test -s "$REGION_OUT/diff_comparison.png"
test -s "$REGION_OUT/compare.json"
test -s "$REGION_OUT/compare.md"

echo "[010] public examples smoke: examples compare collect-and-analyze"
ANALYZE_OUT="$OUTROOT/collect-analyze"
go run ./cmd/css-visual-diff verbs --repository examples/verbs examples compare collect-and-analyze \
  "$BASE/left.html" "$BASE/right.html" '#cta' "$ANALYZE_OUT" --output json >/tmp/cssvd-example-collect-analyze.json
python3 - <<PY
import json, pathlib
rows=json.loads(pathlib.Path('/tmp/cssvd-example-collect-analyze.json').read_text())
row=rows[0]
assert row['ok'] is True, row
assert row['leftExists'] is True and row['rightExists'] is True, row
assert row['styleChangeCount'] > 0, row
PY
test -s "$ANALYZE_OUT/compare.json"
test -s "$ANALYZE_OUT/compare.md"

echo "[010] public examples smoke: examples low-level inspect"
LOW_OUT="$OUTROOT/low-level"
go run ./cmd/css-visual-diff verbs --repository examples/verbs examples low-level inspect \
  "$BASE/index.html" '#cta' "$LOW_OUT" --output json >/tmp/cssvd-example-low-level.json
python3 - <<PY
import json, pathlib
rows=json.loads(pathlib.Path('/tmp/cssvd-example-low-level.json').read_text())
row=rows[0]
assert row['ok'] is True, row
assert row['text'] == 'Book now', row
PY
test -s "$LOW_OUT/element.json"
test -s "$LOW_OUT/snapshot.json"

echo "[010] public examples smoke: examples catalog inspect-page"
CAT_OUT="$OUTROOT/catalog"
go run ./cmd/css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
  "$BASE/index.html" '#cta' "$CAT_OUT" --slug cta --artifacts css-json --output json >/tmp/cssvd-example-catalog.json
test -s "$CAT_OUT/manifest.json"
test -s "$CAT_OUT/index.md"
test -s "$CAT_OUT/artifacts/cta/computed-css.json"

echo "[010] public examples smoke: canonical examples/docs scan"
! rg 'require\("diff"\)|require\("report"\)' examples
! rg 'cvd\.snapshot\(|cvd\.diff\(|cvd\.catalog\(|cvd\.loadConfig' examples internal/cssvisualdiff/doc/topics/javascript-api.md internal/cssvisualdiff/doc/tutorials/pixel-accuracy-scripting-guide.md

echo "[010] public examples smoke: PASS"
