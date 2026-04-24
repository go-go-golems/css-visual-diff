# AGENT.md

## Build and test

- Run tests: `GOWORK=off go test ./...`
- Build the CLI: `GOWORK=off go build ./cmd/css-visual-diff`
- Run the CLI locally: `GOWORK=off go run ./cmd/css-visual-diff`

## Repository shape

- `cmd/css-visual-diff/` contains the CLI entrypoint.
- `internal/cssvisualdiff/` contains the imported comparison engine.
- `legacy/python-prototype/` preserves the earlier Python implementation and artifacts.
