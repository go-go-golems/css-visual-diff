#!/usr/bin/env bash
# Proper-binary smoke: repository-scanned async verb uses require("css-visual-diff") for page/preflight/inspect.
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
<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book</button></body></html>
HTML
cat > "$REPO/inspect.js" <<'JS'
async function inspect(url, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  try {
    const page = await browser.page(url, { viewport: { width: 400, height: 300 } });
    const probes = [{ name: "cta", selector: "#cta", props: ["color"] }];
    const statuses = await page.preflight(probes);
    const single = await page.inspect(probes[0], { outDir: outDir + "/single", artifacts: "css-json" });
    const result = await page.inspectAll(probes, { outDir: outDir + "/all", artifacts: "css-json" });
    await page.close();
    return {
      exists: statuses[0].exists,
      textStart: statuses[0].textStart,
      singleName: single.metadata.name,
      singleColor: single.style.computed.color,
      count: result.results.length,
      outputDir: result.outputDir
    };
  } finally {
    await browser.close();
  }
}
__verb__("inspect", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true },
    outDir: { argument: true, required: true }
  }
});
JS
python3 -m http.server 8765 --bind 127.0.0.1 --directory "$WEB" >"$SMOKE_DIR/server.log" 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null || true' EXIT
sleep 1
URL="http://127.0.0.1:8765/"
"$BIN" verbs --repository "$REPO" custom inspect "$URL" "$OUT" --output json > "$SMOKE_DIR/result.json"
python3 - <<'PY' "$SMOKE_DIR/result.json" "$OUT"
import json, os, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["exists"] is True, row
assert row["textStart"] == "Book", row
assert row["singleName"] == "cta", row
assert row["singleColor"] == "rgb(255, 0, 0)", row
assert row["count"] == 1, row
assert row["outputDir"] == os.path.join(sys.argv[2], "all"), row
assert os.path.exists(os.path.join(sys.argv[2], "single", "computed-css.json"))
assert os.path.exists(os.path.join(sys.argv[2], "all", "computed-css.json"))
print("binary js api success smoke success:", json.dumps(row, sort_keys=True))
PY
printf 'smoke artifacts: %s\n' "$SMOKE_DIR"
