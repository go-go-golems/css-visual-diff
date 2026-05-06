---
Title: Port useful hair-booking utilities to main
Ticket: CSSVD-PORT-HAIR-PRIMITIVES
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - merge-planning
    - visual-diff
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/doc/tutorials/site-comparison-workflow.md
      Note: Dedicated site-comparison review workflow documentation
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: Registration point for existing require(diff).compareRegion API that should remain the high-level compare path
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: Destination for locator.waitFor jsapi method
    - Path: internal/cssvisualdiff/modes/compare.go
      Note: Existing compare-region result shape and artifact path behavior to preserve
    - Path: internal/cssvisualdiff/modes/pixeldiff_util.go
      Note: Existing pixel-diff utility to refactor through internal service
    - Path: internal/cssvisualdiff/service/dom.go
      Note: Destination for selector wait service support
ExternalSources: []
Summary: Plan and track a conservative port of useful hair-booking branch primitives into origin/main without reintroducing deprecated workflow APIs.
LastUpdated: 2026-05-04T17:08:00-04:00
WhatFor: Coordinate implementation tasks for locator.waitFor, internal pixel service extraction, artifact path auditing, and Pyxis adaptation.
WhenToUse: Before implementing or reviewing any hair-booking branch utility port into css-visual-diff main.
---



# Port useful hair-booking utilities to main

## Overview

This ticket tracks a conservative improvement pass for `css-visual-diff`: port only the branch utilities that strengthen the current origin/main JavaScript primitive model, and keep branch-specific workflow APIs out of core.

The accepted implementation candidates are `locator.waitFor(...)`, an internal pixel-diff service extraction, and stable compare-region artifact path behavior. Pyxis should be adapted in userland rather than by restoring `cvd.compare.region` or related branch APIs.

## Key Links

- [Origin/main port guide](./design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md)
- [Implementation diary](./reference/01-implementation-diary.md)
- [Task list](./tasks.md)
- **Related Files**: See frontmatter RelatedFiles field

## Status

Current status: **active**

## Topics

- css-visual-diff
- javascript-api
- merge-planning
- visual-diff

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
