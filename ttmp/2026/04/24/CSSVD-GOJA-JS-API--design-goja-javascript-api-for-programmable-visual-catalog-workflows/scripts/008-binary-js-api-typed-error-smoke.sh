#!/usr/bin/env bash
# Proper-binary smoke: Promise rejection is a typed JS SelectorError/CvdError.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
SMOKE_DIR="${SMOKE_DIR:-$(mktemp -d)}"
BIN="$SMOKE_DIR/css-visual-diff"
REPO="$SMOKE_DIR/verbs"
WEB="$SMOKE_DIR/web"
OUT="$SMOKE_DIR/out"
mkdir -p "$REPO" "$WEB" "$OUT"
go build -o "$BIN" ./cmd/css-visual-diff
cat > "$WEB/index.html" <<'HTML'
<html><body><main id="app">Ready</main></body></html>
HTML
cat > "$REPO/missing.js" <<'JS'
async function missing(url, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.newPage();
    const nav = await page.goto(url, { viewport: { width: 320, height: 240 }, name: "typed-error-smoke" });
    await page.inspect({ name: "missing", selector: "#missing" }, { outDir, artifacts: "html" });
    return { ok: true, url: nav.url };
  } catch (err) {
    return {
      ok: false,
      name: err.name,
      code: err.code,
      operation: err.operation,
      isSelector: err instanceof cvd.SelectorError,
      isCvd: err instanceof cvd.CvdError
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("missing", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true },
    outDir: { argument: true, required: true }
  }
});
JS
python3 -m http.server 8766 --bind 127.0.0.1 --directory "$WEB" >"$SMOKE_DIR/server.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
sleep 1
URL="http://127.0.0.1:8766/"
"$BIN" verbs --repository "$REPO" custom missing "$URL" "$OUT" --output json > "$SMOKE_DIR/result.json"
python3 - <<'PY' "$SMOKE_DIR/result.json"
import json, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["ok"] is False, row
assert row["name"] == "SelectorError", row
assert row["code"] == "SELECTOR_ERROR", row
assert row["operation"] == "css-visual-diff.page.inspect", row
assert row["isSelector"] is True, row
assert row["isCvd"] is True, row
print("binary typed-error smoke success:", json.dumps(row, sort_keys=True))
PY
printf 'smoke artifacts: %s\n' "$SMOKE_DIR"
