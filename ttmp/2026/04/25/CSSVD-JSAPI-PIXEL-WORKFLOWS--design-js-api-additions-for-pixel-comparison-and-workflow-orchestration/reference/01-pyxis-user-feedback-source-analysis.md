---
Title: Pyxis user feedback source analysis
Ticket: CSSVD-JSAPI-PIXEL-WORKFLOWS
Status: active
Topics:
    - frontend
    - visual-regression
    - browser-automation
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
      Note: |-
        Primary Pyxis maintainer feedback document with prioritized css-visual-diff feature requests
        Primary downstream maintainer feature request source
    - Path: ../../../../../../../../../../code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/design/01-css-visual-diff-javascript-workflow-experiment-guide.md
      Note: |-
        Pyxis workflow experiment plan showing how css-visual-diff JS is used in practice
        Workflow experiment plan and migration context
    - Path: ../../../../../../../../../../code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/reference/01-exploration-diary.md
      Note: |-
        Chronological experiment diary with commands, failures, and evidence
        Experiment commands and evidence for JS compare-region smoke
    - Path: ../../../../../../../../../../code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/reference/02-copious-research-notes-for-technical-deep-dive.md
      Note: |-
        Expanded research notes and blog-post source material for Pyxis css-visual-diff JS workflow
        Expanded user-workflow research notes and narrative evidence
ExternalSources:
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
Summary: Source analysis of the first real user feedback for css-visual-diff JavaScript APIs, distilled into maintainer requirements and design tensions.
LastUpdated: 2026-04-25T09:10:00-04:00
WhatFor: Use this as the compact input brief before implementing css-visual-diff JS API additions for pixel comparison and workflow orchestration.
WhenToUse: Read this before the design document or before reviewing the implementation plan, especially when deciding whether a feature belongs in core or userland.
---


# Pyxis User Feedback Source Analysis

## Goal

This reference distills the first serious downstream user feedback for the `css-visual-diff` JavaScript API. The user is Pyxis, a real frontend project using `css-visual-diff` to compare standalone prototype pages/components against React/Storybook implementations.

The goal is to separate:

1. the actual user need,
2. the user's proposed API shapes,
3. the underlying platform gaps,
4. what should be implemented in `css-visual-diff` core,
5. what should remain in project/userland code.

## Source material read

Primary request document:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
```

Workflow exploration ticket:

```text
/home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
```

Key workflow documents:

```text
index.md
design/01-css-visual-diff-javascript-workflow-experiment-guide.md
reference/01-exploration-diary.md
reference/02-copious-research-notes-for-technical-deep-dive.md
tasks.md
```

## The real user story

Pyxis has already used `css-visual-diff` successfully through YAML configs. The YAML workflow gives reliable browser rendering, screenshots, pixel diffs, computed styles, and artifacts. The pain is no longer “can we compare two pages?” The pain is orchestration:

- repeated prototype and Storybook URLs,
- repeated selectors,
- repeated viewport and wait settings,
- manual summary/report copying,
- a need for authoring-mode versus CI-mode policy,
- a desire to build page registries and multi-section workflows in JavaScript.

The new JavaScript API helps with browser/page/locator/probe/snapshot/diff workflows, but it does not yet expose the pixel/region comparison primitive that powers the built-in `verbs script compare region` command and the YAML runner.

## Most important evidence

The Pyxis workflow experiment proved that the built-in JS `compare region` command can reproduce a known YAML section comparison:

```text
Archive content YAML diff:      7.1281%
Archive content JS region diff: 7.128146453089244%
```

The successful command shape was:

```bash
css-visual-diff verbs script compare region \
  --leftUrl http://localhost:7070/standalone/public/archive.html \
  --rightUrl 'http://localhost:6007/iframe.html?id=public-site-pages--archive-desktop&viewMode=story' \
  --leftSelector '#root > *' \
  --rightSelector "[data-page='archive']" \
  --width 920 \
  --height 1460 \
  --leftWaitMs 1000 \
  --rightWaitMs 1000 \
  --outDir ... \
  --writeJson \
  --writeMarkdown \
  --writePngs \
  --output json
```

The key gap: project-specific JS verbs cannot call this behavior through the documented public module. The command is implemented via the undocumented native helper module:

```js
require("diff").compareRegion(...)
```

while the documented module is:

```js
const cvd = require("css-visual-diff")
```

## Prioritized requests from Pyxis

| Priority | User request | Maintainer interpretation |
| --- | --- | --- |
| P0 | JS-callable pixel/region comparison API | Required. This is the missing primitive that unlocks userland orchestration. |
| P1 | Config/job bridge for YAML runner parity | Important, but should be staged after pixel compare service extraction. |
| P1 | Stable compare artifact/result schema | Required if pixel compare is public. Schema versioning should be part of P0/P1. |
| P2 | Multi-section compare helper | Useful, but only after region compare exists. Likely a thin helper or userland first. |
| P2 | Tolerances and normalization hooks | Good future API work for structural/style diffs, independent of pixel compare. |
| P2 | Accepted-difference annotations | Report-layer concept; likely userland first with hooks in core later. |
| P3 | Style property presets | Cheap convenience if kept small and documented. |
| P3 | Storybook URL helper | Userland for now. Too project/ecosystem-specific for core unless it generalizes. |
| P3 | Safer output ergonomics | Should be implemented broadly: create parent directories and document shell redirection footgun. |

## Critical analysis of the proposed `cvd.comparePixels(...)`

The user's minimal ask is:

```js
await cvd.comparePixels({
  left: { page: leftPage, selector: '#root > *' },
  right: { page: rightPage, selector: "[data-page='archive']" },
  threshold: 30,
  outDir,
  writeJson: true,
  writeMarkdown: true,
  writePngs: true,
})
```

This is directionally right but not ideal as the final API.

### What is good about it

- It uses already-open pages, avoiding subprocess and duplicated browser setup.
- It keeps orchestration in JavaScript.
- It exposes the one primitive userland cannot currently recreate cleanly.
- It maps to existing internals and known output artifacts.

### What should be improved

- The name `comparePixels` is too narrow. The existing compare-region operation also captures screenshots, computed styles, matched-style winners, and artifact paths.
- Raw `{ page, selector }` objects weaken the strict handle model recently introduced in the API.
- `writeJson`, `writeMarkdown`, `writePngs` are okay but should be normalized into a coherent artifact/output option model.
- Result keys should be lowerCamel and schema-versioned in JavaScript.
- The API should distinguish region/element comparison from lower-level image diffing.

## The core maintainer problem to solve

The API should expose **pixel comparison as a composable browser primitive**, not as a CLI-command wrapper.

The elegant primitive is not “run the built-in verb from JS.” It is:

1. capture one visual region from a page-bound selector,
2. capture another visual region,
3. normalize images to comparable dimensions,
4. compute pixel differences with a threshold,
5. optionally collect semantic evidence,
6. optionally write artifacts,
7. return a stable plain JS object.

Once that primitive exists, Pyxis can build its own registry, policy, reports, and multi-section runner in userland.

## Core versus userland classification

### Should be core now

- `cvd.compare.region(...)` or equivalent.
- Stable `PixelCompareResult` schema.
- Parent-directory creation for JSON/Markdown/image artifacts.
- A documented migration path away from internal `require("diff")`.
- Service extraction so CLI, built-in verbs, YAML runner, and JS API use one implementation.

### Should be core soon, after P0

- `cvd.compare.sections(...)` if userland implementations converge on one shape.
- `cvd.job.fromConfig(...)` or `cvd.runConfig(...)` bridge for YAML runner parity.
- Numeric tolerances for `cvd.diff(...)`.
- CSS normalization helpers for common representation differences.

### Should remain userland for now

- Pyxis page registry.
- Pyxis policy bands.
- Pyxis accepted-difference lists.
- Pyxis report templates.
- Storybook URL helper unless multiple downstream users request it.
- Token-specific CSS normalization.

## Design tensions to preserve

### Strict handles versus convenient objects

The new lower-level API intentionally uses Go-backed handles/builders so invalid method calls can produce useful errors. `cvd.compare.region(...)` should follow that direction.

Prefer:

```js
await cvd.compare.region({
  left: leftPage.locator('#root > *'),
  right: rightPage.locator("[data-page='archive']"),
})
```

over raw objects:

```js
await cvd.comparePixels({
  left: { page: leftPage, selector: '#root > *' },
  right: { page: rightPage, selector: "[data-page='archive']" },
})
```

A convenience helper can convert page+selector into a region handle later, but the public strict operation should make page-bound identity explicit.

### `cvd.diff` versus pixel diff

`cvd.diff(...)` currently means structural JSON diff. Do not overload it to mean image diff. Use a namespace:

```js
cvd.compare.region(...)
cvd.image.diff(...)
```

or:

```js
cvd.pixel.diff(...)
cvd.pixel.compareRegion(...)
```

### Existing `require("diff")` helper

`require("diff")` should not be promoted as the main public API. It exists as a builtin-script compatibility/internal helper. The public shape should be under `require("css-visual-diff")`.

### One operation versus layered primitives

The user wants one operation because they are blocked. The maintainable design needs layers:

- capture region screenshot,
- pixel diff images,
- compare regions convenience,
- compare sections helper.

Implement the convenience first if needed, but design it as a composition of reusable services.

## Copy/paste examples for the design doc

Recommended first-class shape:

```js
const cvd = require("css-visual-diff")

async function compareArchive(outDir) {
  const browser = await cvd.browser()
  try {
    const leftPage = await browser.page("http://localhost:7070/standalone/public/archive.html", {
      viewport: cvd.viewport(920, 1460),
      waitMs: 1000,
      name: "archive-prototype",
    })
    const rightPage = await browser.page("http://localhost:6007/iframe.html?id=public-site-pages--archive-desktop&viewMode=story", {
      viewport: cvd.viewport(920, 1460),
      waitMs: 1000,
      name: "archive-storybook",
    })

    return await cvd.compare.region({
      name: "archive-content",
      left: leftPage.locator("#root > *"),
      right: rightPage.locator("[data-page='archive']"),
      threshold: 30,
      outDir,
      artifacts: ["screenshots", "diff", "json", "markdown"],
      evidence: {
        computed: ["font-family", "font-size", "line-height", "color", "background-color"],
        attributes: ["id", "class"],
        matched: true,
      },
    })
  } finally {
    await browser.close()
  }
}
```

Optional lower-level image shape:

```js
const diff = await cvd.image.diff({
  left: "out/archive-left.png",
  right: "out/archive-right.png",
  threshold: 30,
  outDir: "out/archive-diff",
  artifacts: ["diff", "comparison"],
})
```

Optional multi-section userland shape:

```js
const sections = [
  ["page", "#root", "[data-story-frame='pyxis-page-shell']"],
  ["content", "#root > *", "[data-page='archive']"],
]

const results = []
for (const [name, leftSelector, rightSelector] of sections) {
  results.push(await cvd.compare.region({
    name,
    left: leftPage.locator(leftSelector),
    right: rightPage.locator(rightSelector),
    outDir: `${outDir}/${name}`,
    threshold: 30,
    artifacts: ["screenshots", "diff", "json", "markdown"],
  }))
}
```

## Usage examples

Use this document when reviewing proposed implementation changes. If a proposed API only wraps the CLI, it has not solved the core problem. If a proposed API requires users to pass raw arbitrary page-like objects everywhere, it has drifted away from the strict handle design. If a proposed API returns unversioned snake_case Go structs directly, it is not yet a stable JavaScript API.
