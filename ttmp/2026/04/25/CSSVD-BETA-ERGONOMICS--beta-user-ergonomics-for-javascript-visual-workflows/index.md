---
Title: Beta user ergonomics for JavaScript visual workflows
Ticket: CSSVD-BETA-ERGONOMICS
Status: complete
Topics:
    - tooling
    - frontend
    - visual-regression
    - browser-automation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/verbs/README.md
      Note: Examples README should be updated with multi-section catalog workflow.
    - Path: internal/cssvisualdiff/jsapi/compare.go
      Note: Comparison artifact writer and cvd.compare.region implementation; target for stable artifact write result.
    - Path: internal/cssvisualdiff/jsapi/locator.go
      Note: Current locator methods; natural implementation point for locator.waitFor.
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Current page wrapper and serialization; possible home for page.waitForSelector.
    - Path: internal/cssvisualdiff/service/collection.go
      Note: Collection profile constants/options; target for profile docs.
    - Path: internal/cssvisualdiff/service/dom.go
      Note: DOM locator services; natural home for service-level wait helper.
ExternalSources: []
Summary: "Completed low-complexity beta-user ergonomics for JavaScript visual workflows: selector waits, stable artifact path results, a multi-section catalog example, and collection profile documentation."
LastUpdated: 2026-04-25T16:05:00-04:00
WhatFor: "Use this ticket to understand the beta-user ergonomics follow-up after the flexible css-visual-diff JavaScript API work."
WhenToUse: "Reference when helping Pyxis or other beta users write project-local visual validation verbs without adding a workflow framework."
---



# Beta user ergonomics for JavaScript visual workflows

## Overview

This ticket closes the small, high-value beta-user ergonomics follow-up for the JavaScript visual workflow API. It intentionally implemented practical primitives and examples rather than a workflow builder: selector readiness waits, stable artifact path returns, a copyable multi-section catalog example, and clearer collection profile documentation.

The scoped implementation is complete. Bounds tolerance APIs, CSS normalization hooks, and built-in style presets remain explicitly deferred until beta usage provides stronger evidence for the right policy shape.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

## Topics

- tooling
- frontend
- visual-regression
- browser-automation

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
