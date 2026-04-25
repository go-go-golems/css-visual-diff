---
Title: Beta user ergonomics for JavaScript visual workflows
Ticket: CSSVD-BETA-ERGONOMICS
Status: active
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
Summary: ""
LastUpdated: 2026-04-25T15:34:23.245464353-04:00
WhatFor: ""
WhenToUse: ""
---



# Beta user ergonomics for JavaScript visual workflows

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

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
