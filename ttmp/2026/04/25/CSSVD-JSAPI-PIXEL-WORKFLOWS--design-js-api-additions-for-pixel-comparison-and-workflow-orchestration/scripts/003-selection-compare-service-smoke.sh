#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[003] selection compare service smoke: focused Go tests"
go test ./internal/cssvisualdiff/service -run 'TestCompareSelections' -count=1

echo "[003] selection compare service smoke: JSON fixture/result generation"
tmp="$(mktemp -d "$PWD/.tmp-selection-compare-smoke-XXXXXX")"
trap 'rm -rf "$tmp"' EXIT
cat >"$tmp/selection_compare_fixture_test.go" <<'GO'
package smoke

import (
  "encoding/json"
  "os"
  "path/filepath"
  "testing"

  "github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

func TestSelectionCompareFixtureJSON(t *testing.T) {
  left := service.SelectionData{SchemaVersion: service.CollectedSelectionSchemaVersion, Name: "left", Selector: "#cta", Exists: true, Visible: true, Bounds: &service.Bounds{Width: 100, Height: 40}, Text: "Book now", ComputedStyles: map[string]string{"color":"black"}, Attributes: map[string]string{"class":"primary"}}
  right := service.SelectionData{SchemaVersion: service.CollectedSelectionSchemaVersion, Name: "right", Selector: "#cta", Exists: true, Visible: true, Bounds: &service.Bounds{Width: 110, Height: 40}, Text: "Book now", ComputedStyles: map[string]string{"color":"red"}, Attributes: map[string]string{"class":"secondary"}}
  comparison, err := service.CompareSelections(left, right, service.CompareSelectionOptions{Name: "fixture", StyleProps: []string{"color"}, Attributes: []string{"class"}})
  if err != nil { t.Fatal(err) }
  payload, err := json.MarshalIndent(comparison, "", "  ")
  if err != nil { t.Fatal(err) }
  if err := os.WriteFile(filepath.Join(os.Getenv("OUT_DIR"), "selection-comparison.json"), payload, 0644); err != nil { t.Fatal(err) }
}
GO
OUT_DIR="$tmp" go test "$tmp/selection_compare_fixture_test.go" -count=1
python3 - <<PY
import json, pathlib, sys
path = pathlib.Path('$tmp') / 'selection-comparison.json'
data = json.loads(path.read_text())
assert data['schemaVersion'] == 'cssvd.selectionComparison.v1', data
assert data['bounds']['changed'] is True, data
assert data['styles'][0]['name'] == 'color', data
assert data['attributes'][0]['name'] == 'class', data
print('[003] JSON fixture OK:', path)
PY

echo "[003] selection compare service smoke: help reference includes comparison model"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-selection-compare-help.txt
rg 'Selection comparison model|cssvd\.selectionComparison\.v1|cvd\.compare\.selections' /tmp/cssvd-jsapi-selection-compare-help.txt >/dev/null

echo "[003] selection compare service smoke: PASS"
