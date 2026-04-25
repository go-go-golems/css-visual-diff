#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
echo "[007] canonical API smoke: help contains canonical names"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-canonical-help.txt
for pattern in 'cvd\.compare\.region' 'cvd\.compare\.selections' 'cvd\.snapshot\.page' 'cvd\.diff\.structural' 'cvd\.image\.diff' 'cvd\.catalog\.create' 'cvd\.config\.load'; do
  rg "$pattern" /tmp/cssvd-jsapi-canonical-help.txt >/dev/null
done
echo "[007] canonical API smoke: examples/builtins avoid internal helpers and old public shapes"
! rg 'require\("diff"\)|require\("report"\)' internal/cssvisualdiff/dsl/scripts examples
! rg 'cvd\.snapshot\(|cvd\.diff\(|cvd\.catalog\(|cvd\.loadConfig' internal/cssvisualdiff/dsl/scripts examples internal/cssvisualdiff/doc/topics/javascript-api.md
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleDiffReportAndWritePrimitives|TestCVDModuleSnapshotsPageWithProbeBuilders|TestCVDModuleCompareRegionLowEffortAPI' -count=1
echo "[007] canonical API smoke: PASS"
