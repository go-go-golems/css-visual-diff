#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[001] beta multisection example smoke: create fixture site"
WORKDIR="$(mktemp -d)"
OUTROOT="$(mktemp -d)"
cleanup() {
  kill "${SERVER_PID:-}" 2>/dev/null || true
  rm -rf "$WORKDIR" "$OUTROOT"
}
trap cleanup EXIT

cat >"$WORKDIR/left.html" <<'HTML'
<html><body>
<main id="page" style="width: 420px; padding: 16px; background: rgb(255, 255, 255); color: rgb(0, 0, 0)">
  <h1 style="font-size: 28px; line-height: 32px">Reference page</h1>
  <button id="cta" class="primary" style="color: rgb(0, 0, 0); background: rgb(240, 240, 240); font-size: 16px; padding: 8px; border-radius: 4px">Book now</button>
</main>
</body></html>
HTML
cat >"$WORKDIR/right.html" <<'HTML'
<html><body>
<main id="page" style="width: 420px; padding: 20px; background: rgb(255, 255, 255); color: rgb(0, 0, 0)">
  <h1 style="font-size: 30px; line-height: 34px">Reference page</h1>
  <button id="cta" class="secondary" style="color: rgb(255, 0, 0); background: rgb(250, 250, 250); font-size: 18px; padding: 12px; border-radius: 8px">Book now</button>
</main>
</body></html>
HTML
python3 -m http.server 8771 --directory "$WORKDIR" >/tmp/cssvd-beta-ergonomics-http.log 2>&1 &
SERVER_PID=$!
sleep 1

BASE=http://127.0.0.1:8771
OUT="$OUTROOT/page-catalog"

echo "[001] beta multisection example smoke: run examples compare page-catalog"
go run ./cmd/css-visual-diff verbs --repository examples/verbs examples compare page-catalog \
  "$BASE/left.html" "$BASE/right.html" "$OUT" --output json >/tmp/cssvd-beta-page-catalog.json

python3 - <<PY
import json, pathlib
rows=json.loads(pathlib.Path('/tmp/cssvd-beta-page-catalog.json').read_text())
assert len(rows) == 1, rows
row=rows[0]
assert row['ok'] is True, row
assert row['catalog']['comparisonCount'] == 2, row
assert len(row['summaries']) == 2, row
for summary in row['summaries']:
    artifacts=summary['artifacts']
    assert artifacts['json'].endswith('/compare.json'), artifacts
    assert artifacts['markdown'].endswith('/compare.md'), artifacts
    assert artifacts['diffComparison'].endswith('/diff_comparison.png'), artifacts
    assert artifacts['leftRegion'].endswith('/left_region.png'), artifacts
    assert artifacts['rightRegion'].endswith('/right_region.png'), artifacts
PY

test -s "$OUT/manifest.json"
test -s "$OUT/index.md"
for section in page cta; do
  test -s "$OUT/artifacts/$section/left_region.png"
  test -s "$OUT/artifacts/$section/right_region.png"
  test -s "$OUT/artifacts/$section/diff_only.png"
  test -s "$OUT/artifacts/$section/diff_comparison.png"
  test -s "$OUT/artifacts/$section/compare.json"
  test -s "$OUT/artifacts/$section/compare.md"
done

echo "[001] beta multisection example smoke: PASS"
