---
Title: Design a flexible lower-level JavaScript API
Ticket: CSSVD-FLEX-JS-API
Status: active
Topics:
    - visual-regression
    - browser-automation
    - tooling
    - chromedp
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: docs/js-api.md
      Note: Current high-level JS API reference.
    - Path: docs/js-verbs.md
      Note: Current repository-scanned JS verbs reference.
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Native css-visual-diff module implementation after Phase 1 refactor; future lower-level adapters should extend this package.
    - Path: internal/cssvisualdiff/service
      Note: Existing Go service boundary that future locator/extractor/snapshot/diff services should extend.
ExternalSources: []
Summary: "Ticket workspace for designing a lower-level, more flexible JavaScript API for css-visual-diff."
LastUpdated: 2026-04-24T20:45:00-04:00
WhatFor: "Use this ticket to plan and implement a JS-native API that can replace many YAML concepts with composable JavaScript objects and functions."
WhenToUse: "When working on page locators, extractor pipelines, snapshot/diff APIs, JS target/probe builders, or YAML migration helpers."
---

# Design a flexible lower-level JavaScript API

## Overview

This ticket designs the next JavaScript API layer for `css-visual-diff`. The current `require("css-visual-diff")` API is Promise-first and useful, but it is close to the high-level CLI/YAML workflow. The new design proposes lower-level primitives: `page.locator()`, locator methods, extractor builders, in-memory snapshots, diffs, reports, fluent target/probe builders, and eventual workflow composition.

The main deliverable is an intern-facing analysis/design/implementation guide with codebase orientation, API sketches, pseudocode, diagrams, implementation phases, validation plans, and file references.

## Key Links

- [Primary design guide](./design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md)
- [Investigation diary](./reference/01-investigation-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Status

Current status: **active**.

The design document is drafted. Validation and reMarkable delivery are tracked in `tasks.md`. Implementation is starting with the no-behavior `internal/cssvisualdiff/jsapi` package refactor.

## Topics

- css-visual-diff
- javascript
- goja
- api-design
- visual-diff

## Structure

- `design-doc/` — Architecture and implementation guides.
- `reference/` — Investigation diary and reusable references.
- `playbooks/` — Future command sequences and test procedures.
- `scripts/` — Future smoke/replay scripts.
- `various/` — Working notes and research.
- `archive/` — Deprecated or reference-only artifacts.
