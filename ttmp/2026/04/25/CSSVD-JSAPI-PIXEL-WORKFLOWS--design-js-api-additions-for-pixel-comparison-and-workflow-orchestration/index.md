---
Title: Design JS API additions for pixel comparison and workflow orchestration
Ticket: CSSVD-JSAPI-PIXEL-WORKFLOWS
Status: active
Topics:
    - frontend
    - visual-regression
    - browser-automation
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/dsl/scripts/compare.js
      Note: Built-in compare-region JS verb that demonstrates the current internal helper workflow
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: Native module registration for require("diff").compareRegion and require("report")
    - Path: internal/cssvisualdiff/modes/compare.go
      Note: Current compare-region implementation to extract/reuse from services
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: Existing strict locator handle API proposed as cvd.compare.region input
    - Path: internal/cssvisualdiff/doc/topics/javascript-api.md
      Note: Public API reference to update when compare APIs are implemented
ExternalSources:
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-LIB--implement-pyxis-css-visual-diff-javascript-userland-library/design/02-css-visual-diff-maintainer-feature-requests.md
    - /home/manuel/code/wesen/2026-04-23--pyxis/ttmp/2026/04/25/PYXIS-CSSVD-JS-WORKFLOW--explore-css-visual-diff-javascript-scripting-workflow/
Summary: Ticket for designing and implementing public css-visual-diff JavaScript API additions around strict region pixel comparison, stable result schemas, and future workflow orchestration based on Pyxis user feedback.
LastUpdated: 2026-04-25T09:25:00-04:00
WhatFor: Track design and implementation work that turns internal compare-region machinery into an elegant public JS API.
WhenToUse: Use when implementing cvd.compare.region, extracting pixel compare services, updating JS API docs, or evaluating follow-up workflow helpers.
---

# Design JS API additions for pixel comparison and workflow orchestration

## Overview

Pyxis is the first real downstream user of the `css-visual-diff` JavaScript API. Their feedback shows that the current API is useful for browser/page/locator/probe/snapshot/structural-diff workflows, but lacks one key primitive: a documented JavaScript-callable region pixel comparison API equivalent to the existing built-in `verbs script compare region` command.

This ticket designs and will track implementation of a polished API addition:

```js
await cvd.compare.region({
  name: "archive-content",
  left: leftPage.locator("#root > *"),
  right: rightPage.locator("[data-page='archive']"),
  threshold: 30,
  outDir,
  artifacts: ["screenshots", "diff", "json", "markdown"],
})
```

The design intentionally prefers strict page-bound locator handles and a `cvd.compare.region(...)` namespace over a loose `cvd.comparePixels({ page, selector })` helper.

## Current conclusion

Implement `cvd.compare.region(...)` under the public module:

```js
const cvd = require("css-visual-diff")
```

Do not promote the internal helper module:

```js
require("diff").compareRegion(...)
```

as public API. Instead, extract shared Go services so the internal helper, built-in verb, CLI compare mode, future YAML job bridge, and public JS API can converge on one implementation.

## Key Links

- [Main design proposal](./design/01-elegant-javascript-api-additions-for-pixel-comparison-workflows.md)
- [JavaScript-centric collected data and comparison object API](./design/02-javascript-centric-collected-data-and-comparison-object-api.md)
- [Pyxis user feedback source analysis](./reference/01-pyxis-user-feedback-source-analysis.md)
- [Investigation diary](./reference/02-investigation-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Source findings

Pyxis proved that the built-in JS compare-region command can reproduce a known YAML comparison:

```text
Archive content YAML diff:      7.1281%
Archive content JS region diff: 7.128146453089244%
```

The missing piece is not a full project workflow framework. It is a core primitive that lets project scripts compare two page-bound visual regions without shelling out.

## Status

Current status: **active design complete; implementation not started**.

Completed:

- Read Pyxis maintainer feature requests.
- Read Pyxis JS workflow exploration ticket.
- Created this docmgr ticket.
- Wrote source analysis.
- Wrote main API design proposal.
- Wrote JavaScript-centric collected-data/comparison-object design report.
- Wrote investigation diary.

Next:

- Relate files through `docmgr doc relate`.
- Run `docmgr doctor`.
- Start Phase 1 service extraction for image diffing.

## Topics

- frontend
- visual-regression
- browser-automation
- tooling

## Structure

- `design/` — API design and implementation plan.
- `reference/` — source analysis and chronological diary.
- `scripts/` — future smoke scripts and validation helpers.
- `various/` — temporary research outputs if needed.
- `archive/` — deprecated/reference-only artifacts.
