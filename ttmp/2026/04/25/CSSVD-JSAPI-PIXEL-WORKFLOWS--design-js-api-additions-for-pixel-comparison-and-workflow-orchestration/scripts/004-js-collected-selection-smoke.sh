#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[004] js collected selection smoke: focused Go integration test"
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleCollectsLocatorSelection' -count=1

echo "[004] js collected selection smoke: help reference includes collection API"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-collect-help.txt
rg 'locator\.collect|cvd\.collect\.selection|cssvd\.collectedSelection\.v1' /tmp/cssvd-jsapi-collect-help.txt >/dev/null

echo "[004] js collected selection smoke: PASS"
