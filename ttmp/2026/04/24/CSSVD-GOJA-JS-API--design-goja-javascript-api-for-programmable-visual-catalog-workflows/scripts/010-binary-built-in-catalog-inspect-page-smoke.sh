#!/usr/bin/env bash
# Proper-binary smoke: built-in `verbs catalog inspect-page` success, authoring-mode missing selector, and CI-mode missing selector.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
SMOKE_DIR="${SMOKE_DIR:-$(mktemp -d)}"
BIN="$SMOKE_DIR/css-visual-diff"
WEB="$SMOKE_DIR/web"
OUT_OK="$SMOKE_DIR/out-ok"
OUT_MISSING="$SMOKE_DIR/out-missing"
OUT_CI="$SMOKE_DIR/out-ci"
mkdir -p "$WEB" "$OUT_OK" "$OUT_MISSING" "$OUT_CI"
go build -o "$BIN" ./cmd/css-visual-diff
cat > "$WEB/index.html" <<'HTML'
<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book</button></body></html>
HTML
python3 -m http.server 8767 --bind 127.0.0.1 --directory "$WEB" >"$SMOKE_DIR/server.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
sleep 1
URL="http://127.0.0.1:8767/"
"$BIN" verbs catalog inspect-page "$URL" "#cta" "$OUT_OK" --slug cta --name CTA --artifacts css-json --output json > "$SMOKE_DIR/success.json"
python3 - <<'PY' "$SMOKE_DIR/success.json" "$OUT_OK"
import json, os, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["ok"] is True, row
assert row["exists"] is True, row
assert row["slug"] == "cta", row
assert row["artifactDir"] == os.path.join(sys.argv[2], "artifacts", "cta"), row
assert os.path.exists(os.path.join(sys.argv[2], "manifest.json")), row
assert os.path.exists(os.path.join(sys.argv[2], "index.md")), row
assert os.path.exists(os.path.join(sys.argv[2], "artifacts", "cta", "computed-css.json")), row
print("built-in catalog inspect-page success smoke:", json.dumps(row, sort_keys=True))
PY
"$BIN" verbs catalog inspect-page "$URL" "#missing" "$OUT_MISSING" --slug missing --artifacts css-json --output json > "$SMOKE_DIR/missing.json"
python3 - <<'PY' "$SMOKE_DIR/missing.json" "$OUT_MISSING"
import json, os, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["ok"] is False, row
assert row["exists"] is False, row
assert row["code"] == "SELECTOR_ERROR", row
assert os.path.exists(os.path.join(sys.argv[2], "manifest.json")), row
assert os.path.exists(os.path.join(sys.argv[2], "index.md")), row
print("built-in catalog inspect-page authoring missing-selector smoke:", json.dumps(row, sort_keys=True))
PY
set +e
"$BIN" verbs catalog inspect-page "$URL" "#missing" "$OUT_CI" --slug missing-ci --failOnMissing --artifacts css-json --output json >"$SMOKE_DIR/ci.json" 2>"$SMOKE_DIR/ci.err"
CI_RC=$?
set -e
if [ "$CI_RC" -eq 0 ]; then
  echo "expected --failOnMissing to exit non-zero" >&2
  cat "$SMOKE_DIR/ci.json" >&2
  exit 1
fi
test -f "$OUT_CI/manifest.json"
test -f "$OUT_CI/index.md"
printf 'built-in catalog inspect-page CI missing-selector smoke: non-zero rc=%s\n' "$CI_RC"
printf 'smoke artifacts: %s\n' "$SMOKE_DIR"
