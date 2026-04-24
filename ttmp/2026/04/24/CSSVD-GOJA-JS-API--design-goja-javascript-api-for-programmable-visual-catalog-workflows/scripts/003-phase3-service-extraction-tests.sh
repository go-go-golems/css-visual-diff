#!/usr/bin/env bash
# Replays Phase 3 validation: service extraction for style/preflight/prepare/browser/inspect.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./cmd/css-visual-diff
