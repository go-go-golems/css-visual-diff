#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[002] pixel service smoke: focused Go tests"
go test ./internal/cssvisualdiff/service -run 'TestDiffImages|TestWritePixelDiff|TestValidatePixel' -count=1

echo "[002] pixel service smoke: modes integration tests"
go test ./internal/cssvisualdiff/modes -run 'TestComputePixelDiff|TestRunPixelDiff' -count=1

echo "[002] pixel service smoke: help reference includes pixel diff model"
go run ./cmd/css-visual-diff help javascript-api >/tmp/cssvd-jsapi-pixel-help.txt
rg 'cvd\.image\.diff|changedPercent|normalizedWidth' /tmp/cssvd-jsapi-pixel-help.txt >/dev/null

echo "[002] pixel service smoke: PASS"
