#!/usr/bin/env bash
# Proper-binary smoke: built-in `verbs catalog inspect-config` loads a css-visual-diff YAML config and writes a catalog.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
SMOKE_DIR="${SMOKE_DIR:-$(mktemp -d)}"
BIN="$SMOKE_DIR/css-visual-diff"
WEB="$SMOKE_DIR/web"
OUT="$SMOKE_DIR/out"
CONFIG="$SMOKE_DIR/config.yaml"
mkdir -p "$WEB" "$OUT"
go build -o "$BIN" ./cmd/css-visual-diff
cat > "$WEB/index.html" <<'HTML'
<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book</button></body></html>
HTML
python3 -m http.server 8768 --bind 127.0.0.1 --directory "$WEB" >"$SMOKE_DIR/server.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
sleep 1
URL="http://127.0.0.1:8768/"
cat > "$CONFIG" <<YAML
metadata:
  slug: config-smoke
  title: Config Smoke
original:
  name: original
  url: $URL
  wait_ms: 0
  viewport:
    width: 400
    height: 300
react:
  name: react
  url: $URL
  wait_ms: 0
  viewport:
    width: 400
    height: 300
styles:
  - name: cta
    selector: "#cta"
    props: [color]
output:
  dir: $SMOKE_DIR/config-output
YAML
"$BIN" verbs catalog inspect-config "$CONFIG" original "$OUT" --artifacts css-json --output json > "$SMOKE_DIR/result.json"
python3 - <<'PY' "$SMOKE_DIR/result.json" "$OUT"
import json, os, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["ok"] is True, row
assert row["slug"] == "config-smoke-original", row
assert row["side"] == "original", row
assert row["inspectedCount"] == 1, row
assert row["missingCount"] == 0, row
assert row["resultCount"] == 1, row
assert os.path.exists(os.path.join(sys.argv[2], "manifest.json")), row
assert os.path.exists(os.path.join(sys.argv[2], "index.md")), row
assert os.path.exists(os.path.join(sys.argv[2], "artifacts", "config-smoke-original", "computed-css.json")), row
print("built-in catalog inspect-config smoke:", json.dumps(row, sort_keys=True))
PY
printf 'smoke artifacts: %s\n' "$SMOKE_DIR"
