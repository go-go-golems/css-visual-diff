#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
echo "[006] compare.region smoke: focused integration tests"
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleCompareRegion' -count=1
echo "[006] compare.region smoke: help reference"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-region-help.txt
rg 'cvd\.compare\.region|leftPage\.locator|diff_comparison\.png' /tmp/cssvd-jsapi-region-help.txt >/dev/null
echo "[006] compare.region smoke: PASS"
