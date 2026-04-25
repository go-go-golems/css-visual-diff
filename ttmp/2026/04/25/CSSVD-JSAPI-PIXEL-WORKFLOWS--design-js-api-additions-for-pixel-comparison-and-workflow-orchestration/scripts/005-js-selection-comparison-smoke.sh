#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[005] js selection comparison smoke: focused Go integration test"
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleComparesCollectedSelections' -count=1

echo "[005] js selection comparison smoke: help reference includes comparison API"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-compare-help.txt
rg 'cvd\.compare\.selections|cssvd\.selectionComparison\.v1|comparison\.styles\.diff' /tmp/cssvd-jsapi-compare-help.txt >/dev/null

echo "[005] js selection comparison smoke: PASS"
