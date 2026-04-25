#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "[001] service collection smoke: focused Go tests"
go test ./internal/cssvisualdiff/service -run 'TestCollectSelection' -count=1

echo "[001] service collection smoke: full service package compile/test"
go test ./internal/cssvisualdiff/service -count=1

echo "[001] service collection smoke: PASS"
