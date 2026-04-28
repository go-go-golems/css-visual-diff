---
Title: Diary
Ticket: CSSVD-JSVERB-REVIEW
Status: active
Topics:
    - css-visual-diff
    - jsverb
    - review-site
    - yaml
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/specs/review-sweep.example.yaml
      Note: Example YAML spec for the new verb
    - Path: examples/verbs/review-sweep.js
      Note: Implemented review-sweep verbs and fixed fs.statSync/isDir + artifact aliases
    - Path: go.mod
      Note: Upgraded go-go-goja to v0.4.14 for yaml/path/default registry modules
    - Path: go.sum
      Note: Updated checksums for go-go-goja v0.4.14 and base64dec
ExternalSources: []
Summary: Implementation diary for review-sweep JSVerb and loader/module fixes.
LastUpdated: 2026-04-27T22:25:00-04:00
WhatFor: Track implementation decisions, failures, validation commands, and review notes for CSSVD-JSVERB-REVIEW.
WhenToUse: Read before resuming work on the review-sweep JSVerb or go-go-goja module-loading issues.
---


# Diary

## Goal

Capture the implementation of the `review-sweep` JSVerb example, including the module loader investigation, the `go-go-goja` dependency upgrade, and validation against Pyxis artifacts.

## Step 1: Scaffold and Implement review-sweep Verbs

The first implementation pass created `examples/verbs/review-sweep.js` as an external JSVerb example. It defines the `examples review-sweep` namespace, accepts a YAML spec file plus output directory, and provides two verbs: `from-spec` for running browser comparisons and `summary` for rebuilding `summary.json` from existing artifacts.

The implementation followed the ticket guide: read YAML, run `diff.compareRegion()` for each page/section, map comparison output to the review-site `SummaryRow` shape, and write a root `summary.json` that `css-visual-diff serve` can consume.

### Prompt Context

**User prompt (verbatim):** "Implement it, keep appropriate commits at appropriate intervals, keep a diary."

**Assistant interpretation:** Implement the planned JSVerb example from the ticket, commit focused increments, and maintain an implementation diary.

**Inferred user intent:** Turn the design document into working code while preserving a clear audit trail.

**Commit (code):** `fd50847` — "feat(review-sweep): scaffold verb with package, sections, helpers"

**Commit (code):** `eff6f31` — "feat(review-sweep): implement from-spec verb"

**Commit (code):** `2620bb2` — "feat(review-sweep): implement summary verb"

**Commit (code):** `df8e67f` — "docs: add example review-sweep spec YAML"

### What I did

- Added `examples/verbs/review-sweep.js`.
- Added package/section declarations for `examples review-sweep`.
- Added helpers: `classify()`, `resolveBands()`, `buildRowFromCompareResult()`, `buildSummary()`.
- Implemented `fromSpec()` to read a YAML spec and call `diff.compareRegion()`.
- Implemented `summary()` to walk existing `compare.json` artifacts and rebuild `summary.json`.
- Added `examples/specs/review-sweep.example.yaml` as a user-copyable template.

### Why

- The review site needs a general data generator, not only one-off Pyxis scripts.
- JSVerb is the right place because it has access to css-visual-diff comparison primitives and the go-go-goja file/YAML APIs.

### What worked

- Verb discovery worked immediately:
  - `examples review-sweep from-spec`
  - `examples review-sweep summary`
- The command schemas rendered in `--help` with `--specFile`, `--outDir`, `--writeMarkdown`, and `--failFast`.

### What didn't work

- Running the `summary` verb initially failed with:

```text
Error: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
```

- This happened before any directory walking or summary generation could run.

### What I learned

- The implementation assumed the current go-go-goja docs matched the css-visual-diff dependency version. That assumption was false; css-visual-diff was pinned to an older module version.

### What was tricky to build

- There are two compare JSON shapes to support:
  - catalog/inspect camelCase format (`pixel.changedPercent`, `styles`, `attributes`, `left`, `right`)
  - `diff.compareRegion()` snake_case format (`pixel_diff.changed_percent`, `computed_diffs`, `url1`, `url2`)
- The summary verb therefore has to normalize both.

### What warrants a second pair of eyes

- The row mapping in `buildRowFromCompareResult()` should be reviewed against actual `modes.CompareResult` JSON fields.
- The initial `from-spec` browser path still needs a reliable end-to-end browser smoke test.

### What should be done in the future

- Add an automated test harness for JSVerb module loading and summary rebuilding.

### Code review instructions

- Start with `examples/verbs/review-sweep.js`.
- Validate discovery with:

```bash
GOWORK=off ./dist/css-visual-diff verbs --repository examples/verbs examples review-sweep --help
```

### Technical details

- `diff.compareRegion()` writes Go's `modes.CompareResult` JSON shape, not the Pyxis camelCase shape.

## Step 2: Diagnose and Fix Default Module Loading

The second pass took over from the `Invalid module` failure. The first hypothesis was that the jsverbs source loader prevented fallback to native default registry modules. That was not the root cause. The real problem was dependency skew: css-visual-diff used `go-go-goja v0.4.11`, while the docs and sibling checkout referenced newer modules only present in `v0.4.14`.

After upgrading go-go-goja, the module smoke test passed for `yaml`, `path`, `os`, `time`, `timer`, `url`, `buffer`, `exec`, and `fs`. Then the summary verb exposed a second issue: go-go-goja `fs.statSync()` returns plain boolean fields (`isDir`, `isFile`), not Node-style methods (`isDirectory()`). Fixing that made the Pyxis summary rebuild produce the expected 13 rows.

### Prompt Context

**User prompt (verbatim):** "We're taking over for your little brother who has been struggling with the Loader to get the default registry modules (the read-only ones). 

Analyze what they did, and then write a report document in the current ticket, and fix things."

**Assistant interpretation:** Investigate the module loader/default registry problem, document the findings in the ticket, and apply the necessary fixes.

**Inferred user intent:** Stop chasing the wrong loader theory, identify the actual module availability issue, and get the JSVerb working.

**Commit (code):** `73591a6` — "fix(review-sweep): enable yaml/path modules and correct fs stat handling"

### What I did

- Confirmed current git state and recent review-sweep commits.
- Reproduced module failures using a temporary debug verb.
- Compared css-visual-diff's `go.mod` dependency version with the sibling `go-go-goja` docs/repo.
- Found css-visual-diff pinned to `github.com/go-go-golems/go-go-goja v0.4.11`.
- Upgraded to `v0.4.14` with:

```bash
GONOSUMDB=github.com/go-go-golems/* GOPRIVATE=github.com/go-go-golems/* \
  go get github.com/go-go-golems/go-go-goja@ebaf324
```

- Added the missing indirect `base64dec` go.sum entry by running:

```bash
go get github.com/dop251/goja_nodejs/buffer@v0.0.0-20260212111938-1f56ff5bcf14
```

- Fixed `fs.statSync(...).isDirectory()` calls to `fs.statSync(...).isDir === true`.
- Added `ensureReviewSiteArtifactAliases()` so `from-spec` copies `url1_screenshot.png` / `url2_screenshot.png` to `left_region.png` / `right_region.png`.
- Wrote `reference/02-loader-and-review-sweep-fix-report.md`.

### Why

- The JSVerb requires `yaml` and `path` to implement the spec-driven workflow cleanly.
- The review-site data contract names cropped images `left_region.png` and `right_region.png`, while `compareRegion()` writes `url1_screenshot.png` and `url2_screenshot.png`.

### What worked

- Module smoke test after upgrade:

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

- Pyxis summary rebuild validated successfully:

```json
{
  "classificationCounts": { "review": 7, "tune-required": 6 },
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

### What didn't work

- A tiny local `from-spec` browser smoke test hung after printing:

```text
Comparing smoke/hero...
```

- The shell command timed out after 180 seconds. This looks like a browser/driver hang rather than a module-loading failure and needs separate follow-up.

### What I learned

- The go-go-goja docs in the sibling repo describe newer primitives, but css-visual-diff only gets them after upgrading its module dependency.
- The go-go-goja `fs` module is Node-like but not an exact clone; `statSync()` returns data fields instead of method-bearing `Stats` objects.

### What was tricky to build

- The error message `Invalid module` initially looked like a loader fallthrough problem, but a module matrix showed a pattern that matched dependency contents instead.
- `fs` worked while `yaml` and `path` failed because `v0.4.11` contained `fs`, `exec`, and `timer`, but not the newer module packages.

### What warrants a second pair of eyes

- Confirm `go-go-goja v0.4.14` is the intended dependency bump for css-visual-diff.
- Review whether `from-spec` should invoke the newer selection-comparison pipeline instead of `diff.compareRegion()` if browser hangs persist.

### What should be done in the future

- Add a small JSVerb regression test that requires `yaml`, `path`, and `fs` inside the css-visual-diff verb runtime.
- Investigate the `from-spec` browser timeout separately.

### Code review instructions

- Review `go.mod` and `go.sum` first.
- Then review `examples/verbs/review-sweep.js`, focusing on:
  - `ensureReviewSiteArtifactAliases()`
  - `summary()` directory walking
  - `buildRowFromCompareJson()` format normalization

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

### Technical details

- Report document: `reference/02-loader-and-review-sweep-fix-report.md`
- Main code file: `examples/verbs/review-sweep.js`
- Dependency change: `github.com/go-go-golems/go-go-goja v0.4.11 => v0.4.14`

## Step 3: Fix Review Site CSS Diff Crash for compareRegion JSON

The served React app crashed when the reviewer clicked the CSS diff sidebar tab. The stack trace was minified, but the error was clear: `e.styles is undefined`. The crash happened because the UI assumed every `/api/compare` response used the catalog/inspect `compare.json` shape with `styles[]` and `attributes[]`.

The new JSVerb path can also produce `diff.compareRegion()` JSON, whose shape uses `computed_diffs`, `pixel_diff`, and `url1`/`url2`. I added a small normalization layer so the UI can render either format without special-casing inside each component.

### Prompt Context

**User prompt (verbatim):** "for the serve react app, when I click on css diff http://127.0.0.1:8097/ I get:  [HTTP/1.1 200 OK 0ms]

Uncaught TypeError: can't access property \"filter\", e.styles is undefined
    xi http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    Do http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    vc http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Fc http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Mu http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    ku http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Ou http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    gu http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    ld http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    id http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    od http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    ud http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    ud http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    rd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    hu http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Jo http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Ko http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    notify http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    ee http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    notify http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    s http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    c http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    t http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    f http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    f http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    s <anonymous code>:3
    dispatch http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    kr http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    qt http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    dispatch <anonymous code>:6
    onClick http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    yd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    wd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    _n http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    wd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    up http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    cp http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    Cd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    bd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Sd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    Sd http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:8
    createRoot http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:9
    <anonymous> http://127.0.0.1:8097/assets/index-Cr2p6vhA.js:18


Probably we should run this in vite to get better sourcemap?"

**Assistant interpretation:** Reproduce and fix the CSS diff tab crash; Vite sourcemaps would help, but the error is already specific enough to patch the bad assumption.

**Inferred user intent:** Make the review app robust for generated review-sweep data, not only the original Pyxis/camelCase data.

**Commit (code):** `ab161b7` — "fix(review-site): normalize compare.json variants for sidebar tabs"

### What I did

- Added `web/review-site/src/utils/compareData.ts`.
- Added helpers for normalized changed styles, changed attributes, bounds, source URLs, and pixel stats.
- Updated `StylesTab` to use `changedStyles()` / `changedAttributes()` instead of `compareData.styles.filter(...)`.
- Updated `MetaTab` to use normalized bounds and source URLs.
- Updated export generation to use the same normalization helpers.
- Rebuilt the web app and embedded binary.
- Restarted the server on `8097` and clicked the CSS diff tab with Playwright.

### Why

- `/api/compare` can now return at least two valid compare JSON variants.
- The UI should normalize those variants at the edge rather than assuming one exact backend shape.

### What worked

Validation commands:

```bash
BUILD_WEB_LOCAL=1 GOWORK=off go run ./cmd/build-web
GOWORK=off go build -tags embed -o dist/css-visual-diff ./cmd/css-visual-diff
```

Playwright validation on `http://127.0.0.1:8097/`:

- Loaded the review app.
- Clicked `CSS diff`.
- Observed the computed style diff list render.
- Checked browser console errors: `0`.

### What didn't work

- The first browser reload still served the old `index-Cr2p6vhA.js` because I had not rebuilt the Go binary after exporting the new web assets. Rebuilding with `-tags embed` fixed the embedded asset hash.

### What I learned

- The build pipeline has two distinct steps: export web assets, then rebuild the Go binary with `-tags embed`. Doing only the first leaves `dist/css-visual-diff` serving old assets.

### What was tricky to build

- The compare data type currently models only the catalog/inspect shape. The normalizer has to use loose records for `computed_diffs`, `pixel_diff`, and `url1`/`url2` while preserving strict component-level types.

### What warrants a second pair of eyes

- `compareData.ts` should be reviewed as the compatibility boundary. If more compare formats appear, they should be normalized there, not inside React components.

### What should be done in the future

- Consider changing `/api/compare` to normalize server-side so the React app receives one canonical shape.
- Add a small fixture test with both compare JSON variants.

### Code review instructions

- Start with `web/review-site/src/utils/compareData.ts`.
- Then inspect `StylesTab`, `MetaTab`, and `utils/export.ts` for use of the helpers.

### Technical details

- Catalog/inspect format uses `styles[]`, `attributes[]`, `bounds`, `left`, `right`, `pixel`.
- `diff.compareRegion()` format uses `computed_diffs`, `pixel_diff`, `url1`, `url2`.
