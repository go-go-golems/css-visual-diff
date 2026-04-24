#!/usr/bin/env bash
# Full repository validation used after each major phase.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
go test ./...
