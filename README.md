# css-visual-diff

`css-visual-diff` is a Go CLI for comparing rendered HTML/CSS across two browser targets.
It is being rebuilt from the proven `sbcap` comparison engine and now uses a standard
`go-go-golems` project layout.

## Current shape

The live tool now centers on:

- browser-driven capture via `chromedp`
- element-level compare workflows
- computed CSS diffs
- matched-style / cascade inspection
- pixel diff artifacts

The previous Python prototype has been preserved under:

- `legacy/python-prototype/`

## Development

```bash
GOWORK=off go mod tidy
GOWORK=off go test ./...
GOWORK=off go build ./cmd/css-visual-diff
```

## CLI

```bash
GOWORK=off go run ./cmd/css-visual-diff --help
GOWORK=off go run ./cmd/css-visual-diff compare --help
GOWORK=off go run ./cmd/css-visual-diff chromedp-probe --help
```
