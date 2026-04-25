#!/usr/bin/env bash
set -euo pipefail

BIN="${CSSVD_BIN:-$(pwd)/css-visual-diff}"
if [[ ! -x "$BIN" ]]; then
  go build -o "$BIN" ./cmd/css-visual-diff
fi

WORKDIR="$(mktemp -d)"
OUTDIR="$WORKDIR/out"
mkdir -p "$OUTDIR"
HTML="$WORKDIR/index.html"
cat > "$HTML" <<'HTML'
<!doctype html>
<html>
  <body>
    <button id="cta" class="primary" style="color: rgb(255, 0, 0); background-color: rgb(255, 255, 255); font-size: 16px; font-weight: 700; line-height: 20px;">Book now</button>
  </body>
</html>
HTML
python3 -m http.server 8767 --directory "$WORKDIR" >/tmp/cssvd-low-level-http.log 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true; rm -rf "$WORKDIR"' EXIT
sleep 1

"$BIN" verbs --repository examples/verbs examples low-level inspect \
  http://127.0.0.1:8767/ '#cta' "$OUTDIR" \
  --output json >/tmp/cssvd-low-level-smoke.json

grep -q '"ok": true' /tmp/cssvd-low-level-smoke.json
test -s "$OUTDIR/element.json"
test -s "$OUTDIR/snapshot.json"
grep -q 'Book now' "$OUTDIR/element.json"

echo "OK: lower-level external verb smoke passed"
