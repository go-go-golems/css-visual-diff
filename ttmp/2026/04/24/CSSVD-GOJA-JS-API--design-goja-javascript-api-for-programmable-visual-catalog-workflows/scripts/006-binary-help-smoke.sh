#!/usr/bin/env bash
# Proper-binary smoke: build the CLI and verify root/lazy verbs help surfaces.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
SMOKE_DIR="${SMOKE_DIR:-$(mktemp -d)}"
BIN="$SMOKE_DIR/css-visual-diff"
go build -o "$BIN" ./cmd/css-visual-diff
"$BIN" --help | tee "$SMOKE_DIR/root-help.txt"
"$BIN" verbs --help | tee "$SMOKE_DIR/verbs-help.txt"
grep -q "verbs" "$SMOKE_DIR/root-help.txt"
grep -q "script" "$SMOKE_DIR/verbs-help.txt"
printf 'binary help smoke success: %s\n' "$SMOKE_DIR"
