#!/usr/bin/env bash
# Replays Phase 2 validation: lazy `css-visual-diff verbs` repository-scanned commands.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff
