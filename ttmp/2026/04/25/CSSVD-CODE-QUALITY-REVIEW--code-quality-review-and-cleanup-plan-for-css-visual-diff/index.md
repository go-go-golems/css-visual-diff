---
Title: Code quality review and cleanup plan for css-visual-diff
Ticket: CSSVD-CODE-QUALITY-REVIEW
Status: active
Topics:
    - tooling
    - backend
    - frontend
    - visual-regression
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/css-visual-diff/main.go
      Note: Top-level CLI assembly and largest production file; report recommends splitting command definitions.
    - Path: internal/cssvisualdiff/jsapi/compare.go
      Note: SelectionComparison JS handle and Markdown artifact renderer lint findings.
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: JS module/browser/page/error wrapper hotspot; report recommends later file split.
    - Path: internal/cssvisualdiff/modes/inspect.go
      Note: Contains stale duplicate inspect artifact helpers after service extraction.
    - Path: internal/cssvisualdiff/runner/runner.go
      Note: Config-mode orchestration and switch dispatch; report discusses optional mode registry.
    - Path: internal/cssvisualdiff/service/catalog_service.go
      Note: Go-backed catalog manifest/index service for comparison records.
    - Path: internal/cssvisualdiff/service/inspect.go
      Note: Canonical inspect artifact service implementation.
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-25T14:32:33.92067141-04:00
WhatFor: ""
WhenToUse: ""
---


# Code quality review and cleanup plan for css-visual-diff

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- tooling
- backend
- frontend
- visual-regression

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
