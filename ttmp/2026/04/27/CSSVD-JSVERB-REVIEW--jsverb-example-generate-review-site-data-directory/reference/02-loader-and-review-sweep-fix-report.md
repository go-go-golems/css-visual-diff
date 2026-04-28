---
Title: Loader and Review Sweep Fix Report
Ticket: CSSVD-JSVERB-REVIEW
Status: active
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-goja/pkg/doc/16-nodejs-primitives.md
      Note: Referenced docs for fs/path/node primitives
    - Path: ../../../../../../../go-go-goja/pkg/doc/16-yaml-module.md
      Note: Referenced docs that required newer go-go-goja than css-visual-diff used
    - Path: examples/verbs/review-sweep.js
      Note: Report explains code fixes for module loading fallout and fs stat behavior
    - Path: go.mod
      Note: Report explains dependency skew root cause
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Loader and Review Sweep Fix Report

## Goal

Analyze why the `review-sweep` JSVerb could not load the expected go-go-goja default modules (`yaml`, `path`, etc.), explain what the previous implementation did, and document the fix.

## Executive Summary

The failure was not caused by the jsverbs source loader itself. The css-visual-diff binary was still pinned to `github.com/go-go-golems/go-go-goja v0.4.11`, which predates the newer default registry modules documented in the sibling `go-go-goja` repository. In that older version, only a smaller set of modules existed/registered (`fs`, `exec`, `timer`, etc.). Therefore `require("yaml")`, `require("path")`, `require("os")`, `require("crypto")`, `require("buffer")`, and `require("url")` failed with `Invalid module`.

Upgrading css-visual-diff to `go-go-goja v0.4.14` resolved the module availability problem. After the upgrade, the same module smoke test reports `OK` for `fs`, `yaml`, `path`, `os`, `time`, `timer`, `url`, `buffer`, and `exec`.

A second issue was found in the JS verb code: go-go-goja's `fs.statSync()` returns a plain object with boolean fields (`isDir`, `isFile`), not Node's method-based `Stats` object (`isDirectory()`). The summary verb was filtering every directory out, producing an empty `summary.json`. This was fixed by checking `stat.isDir === true`.

## What the Previous Implementation Did

The previous implementation created `examples/verbs/review-sweep.js` with:

- `examples review-sweep from-spec`
- `examples review-sweep summary`
- YAML-driven page/section spec support
- `diff.compareRegion()` orchestration
- `summary.json` rebuilding from existing artifacts
- Support for both known compare JSON shapes:
  - camelCase catalog/inspect format (`pixel.changedPercent`, `styles`, `left`, `right`)
  - snake_case `compareRegion` format (`pixel_diff.changed_percent`, `computed_diffs`, `url1`, `url2`)

This was directionally correct, but the first run exposed the module availability mismatch.

## Observed Failure

Command:

```bash
GOWORK=off ./dist/css-visual-diff verbs --repository examples/verbs \
  examples review-sweep summary \
  --specFile /tmp/pyxis-review-sweep.spec.yaml \
  --outDir /tmp/pyxis-public-pages-final-sweep \
  --output json
```

Initial error:

```text
Error: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
```

A targeted module smoke test showed:

```json
{
  "buffer": "Invalid module",
  "exec": "OK",
  "fs": "OK",
  "os": "Invalid module",
  "time": "Invalid module",
  "timer": "OK",
  "url": "Invalid module",
  "yaml": "Invalid module"
}
```

This proved that `fs` worked, but the newer docs' modules did not.

## Root Cause

The css-visual-diff `go.mod` was pinned to:

```go
github.com/go-go-golems/go-go-goja v0.4.11
```

The local sibling repo documented and implemented newer primitives in later commits/releases, including:

- `yaml`
- `path` / `node:path`
- `os` / `node:os`
- `crypto` / `node:crypto`
- goja_nodejs `Buffer`, `URL`, `URLSearchParams`, `util`

The older module cache for `v0.4.11` only had a smaller `modules/` set:

```text
common.go
database
exec
exports.go
fs
timer
typing.go
```

So the runtime was doing what the old dependency could do. The docs being referenced were from the newer sibling checkout, not from the css-visual-diff dependency version.

## Fixes Applied

### 1. Upgrade go-go-goja

Changed:

```diff
-github.com/go-go-golems/go-go-goja v0.4.11
+github.com/go-go-golems/go-go-goja v0.4.14
```

This pulled in the default registry modules needed by the verb.

### 2. Add missing indirect dependency

The newer goja_nodejs Buffer path required:

```go
github.com/dop251/base64dec v0.0.0-20231022112746-c6c9f9a96217 // indirect
```

This was added by `go get github.com/dop251/goja_nodejs/buffer@v0.0.0-20260212111938-1f56ff5bcf14` after the first build reported a missing go.sum entry.

### 3. Fix fs.statSync directory checks

The previous JS used Node-style methods:

```javascript
fs.statSync(path).isDirectory()
```

The go-go-goja fs module returns plain fields:

```json
{
  "isDir": true,
  "isFile": false,
  "mode": 2147484141,
  "name": "about",
  "size": 4096,
  "modTime": "..."
}
```

Fixed to:

```javascript
fs.statSync(path).isDir === true
```

### 4. Add review-site artifact aliases

`diff.compareRegion()` writes cropped screenshots as:

```text
url1_screenshot.png
url2_screenshot.png
```

The published review-site data contract uses:

```text
left_region.png
right_region.png
```

The `from-spec` verb now copies aliases after each successful comparison:

```javascript
url1_screenshot.png -> left_region.png
url2_screenshot.png -> right_region.png
```

The summary rebuilding helper also tolerates both names for snake_case `compareRegion` artifacts.

## Validation

### Module smoke test

After upgrade and rebuild:

```json
{
  "buffer": "OK",
  "exec": "OK",
  "fs": "OK",
  "os": "OK",
  "time": "OK",
  "timer": "OK",
  "url": "OK",
  "yaml": "OK"
}
```

### Pyxis summary rebuild

Command:

```bash
GOWORK=off ./dist/css-visual-diff verbs --repository examples/verbs \
  examples review-sweep summary \
  --specFile /tmp/pyxis-review-sweep.spec.yaml \
  --outDir /tmp/pyxis-public-pages-final-sweep \
  --output json \
  --fields pageCount,sectionCount,maxChangedPercent,classificationCounts,policy
```

Result:

```json
{
  "classificationCounts": {
    "review": 7,
    "tune-required": 6
  },
  "maxChangedPercent": 11.604994124559342,
  "pageCount": 5,
  "policy": {
    "failureCount": 6,
    "ok": false,
    "worstClassification": "tune-required"
  },
  "sectionCount": 13
}
```

This matches the known Pyxis run: 13 rows, 5 pages, 7 review, 6 tune-required.

## Remaining Notes

A smoke test of `from-spec` against tiny local HTML pages reached `Comparing smoke/hero...` but exceeded the 180-second shell timeout in this environment. That appears to be a browser/driver hang rather than a JS module-loading issue; it should be revisited separately if `from-spec` hangs in a normal browser-capable environment.

The non-browser `summary` path is fully validated against real Pyxis artifacts.

## Commits

- `73591a6` — `fix(review-sweep): enable yaml/path modules and correct fs stat handling`

## Review Instructions

Start with:

- `go.mod` / `go.sum`: confirm the `go-go-goja` upgrade to `v0.4.14` is acceptable.
- `examples/verbs/review-sweep.js`: review `ensureReviewSiteArtifactAliases()`, stat handling, and both JSON-format paths in `buildRowFromCompareJson()`.

Validate with:

```bash
GOWORK=off go build -o dist/css-visual-diff ./cmd/css-visual-diff

GOWORK=off ./dist/css-visual-diff verbs --repository examples/verbs \
  examples review-sweep summary \
  --specFile /tmp/pyxis-review-sweep.spec.yaml \
  --outDir /tmp/pyxis-public-pages-final-sweep \
  --output json \
  --fields pageCount,sectionCount,maxChangedPercent,classificationCounts,policy
```
