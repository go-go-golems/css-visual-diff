#!/usr/bin/env bash
set -euo pipefail

BIN="${CSSVD_BIN:-$(pwd)/css-visual-diff}"
if [[ ! -x "$BIN" ]]; then
  go build -o "$BIN" ./cmd/css-visual-diff
fi

for slug in javascript-api javascript-verbs pixel-accuracy-scripting-guide inspect-workflow config-selectors; do
  "$BIN" help "$slug" >/tmp/cssvd-help-$slug.txt
  test -s /tmp/cssvd-help-$slug.txt
  grep -qi "css-visual-diff\|JavaScript\|Pixel\|Inspect\|Config" /tmp/cssvd-help-$slug.txt
done

echo "OK: embedded help entries are available"
