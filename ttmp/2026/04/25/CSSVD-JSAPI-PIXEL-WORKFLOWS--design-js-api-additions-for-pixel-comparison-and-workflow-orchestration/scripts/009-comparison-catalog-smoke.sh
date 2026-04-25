#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
echo "[009] comparison catalog smoke: focused integration test"
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleRecordsComparisonInCatalog' -count=1
echo "[009] comparison catalog smoke: service catalog test"
go test ./internal/cssvisualdiff/service -run 'TestCatalogWritesManifestAndIndex' -count=1
echo "[009] comparison catalog smoke: help reference includes catalog.record"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-catalog-help.txt
rg 'catalog\.record|manifest\.comparisons|## Comparisons' /tmp/cssvd-jsapi-catalog-help.txt >/dev/null
echo "[009] comparison catalog smoke: PASS"
