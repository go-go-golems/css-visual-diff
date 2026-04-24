#!/usr/bin/env bash
# Proper-binary smoke: repository-scanned verb uses cvd.catalog() to write manifest and index.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
SMOKE_DIR="${SMOKE_DIR:-$(mktemp -d)}"
BIN="$SMOKE_DIR/css-visual-diff"
REPO="$SMOKE_DIR/verbs"
OUT="$SMOKE_DIR/out"
mkdir -p "$REPO" "$OUT"
go build -o "$BIN" ./cmd/css-visual-diff
cat > "$REPO/catalog.js" <<'JS'
async function catalogSmoke(outDir) {
  const cvd = require("css-visual-diff");
  const catalog = cvd.catalog({ title: "Binary Catalog Smoke", outDir, artifactRoot: "../artifacts" });
  const target = { slug: "../Demo Target!", name: "Demo Target", url: "http://example.test", selector: "#root", viewport: { width: 320, height: 240 } };
  catalog.addTarget(target);
  catalog.recordPreflight(target, [{ name: "root", selector: "#root", exists: true, visible: true, textStart: "Ready" }]);
  catalog.addResult(target, { outputDir: catalog.artifactDir(target.slug), results: [{ metadata: { name: "root", selector: "#root", createdAt: "2026-04-24T00:00:00Z" }, style: { exists: true, computed: { color: "rgb(0, 0, 0)" } } }] });
  const manifestPath = await catalog.writeManifest();
  const indexPath = await catalog.writeIndex();
  const summary = catalog.summary();
  return { manifestPath, indexPath, targetCount: summary.targetCount, resultCount: summary.resultCount, artifactDir: catalog.artifactDir(target.slug) };
}
__verb__("catalogSmoke", {
  parents: ["custom"],
  fields: {
    outDir: { argument: true, required: true }
  }
});
JS
"$BIN" verbs --repository "$REPO" custom catalog-smoke "$OUT" --output json > "$SMOKE_DIR/result.json"
python3 - <<'PY' "$SMOKE_DIR/result.json" "$OUT"
import json, os, sys
rows=json.load(open(sys.argv[1]))
assert len(rows)==1, rows
row=rows[0]
assert row["manifestPath"] == os.path.join(sys.argv[2], "manifest.json"), row
assert row["indexPath"] == os.path.join(sys.argv[2], "index.md"), row
assert row["targetCount"] == 1, row
assert row["resultCount"] == 1, row
assert row["artifactDir"] == os.path.join(sys.argv[2], "artifacts", "demo-target"), row
manifest=open(os.path.join(sys.argv[2], "manifest.json")).read()
assert '"schema_version": "css-visual-diff.catalog.v1"' in manifest, manifest
assert '"slug": "demo-target"' in manifest, manifest
index=open(os.path.join(sys.argv[2], "index.md")).read()
assert "# Binary Catalog Smoke" in index, index
print("binary catalog smoke success:", json.dumps(row, sort_keys=True))
PY
printf 'smoke artifacts: %s\n' "$SMOKE_DIR"
