#!/usr/bin/env bash
# Replays Phase 1 validation: DSL tests should keep generated artifacts in t.TempDir().
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test ./internal/cssvisualdiff/dsl ./cmd/css-visual-diff
# The old leak pattern must not reappear in the repository tree.
if find internal/cssvisualdiff/dsl -maxdepth 1 -name 'css-visual-diff-compare-*' | grep -q .; then
  echo "unexpected generated css-visual-diff-compare-* artifact under internal/cssvisualdiff/dsl" >&2
  exit 1
fi
