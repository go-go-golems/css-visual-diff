#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
echo "[008] built-in compare dogfood smoke: focused host tests"
go test ./internal/cssvisualdiff/dsl -run 'TestEmbeddedCompareCommandsExecute' -count=1
echo "[008] built-in compare dogfood smoke: source uses public API"
rg 'cvd\.compare\.region' internal/cssvisualdiff/dsl/scripts/compare.js >/dev/null
! rg 'require\("diff"\)|require\("report"\)' internal/cssvisualdiff/dsl/scripts/compare.js
echo "[008] built-in compare dogfood smoke: PASS"
