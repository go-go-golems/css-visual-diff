#!/usr/bin/env bash
# Replays Phase 4 validation through Go integration tests for repository-scanned async verbs.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
