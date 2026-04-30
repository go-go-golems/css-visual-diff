---
Title: 'JSVerb Example: Generate Review Site Data Directory'
Ticket: CSSVD-JSVERB-REVIEW
Status: active
Topics:
    - css-visual-diff
    - jsverb
    - review-site
    - yaml
    - goja
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../go-go-goja/pkg/doc/16-nodejs-primitives.md
      Note: fs/path/os module docs — file I/O primitives available in the goja VM
    - Path: ../../../../../../go-go-goja/pkg/doc/16-yaml-module.md
      Note: YAML module in go-go-goja — parse/stringify/validate for reading project specs
    - Path: examples/verbs/catalog-inspect-page.js
      Note: Example external verb — shows verb structure
    - Path: internal/cssvisualdiff/doc/topics/review-site-data-spec.md
      Note: Data format specification that the verb must produce — summary.json
    - Path: internal/cssvisualdiff/dsl/host.go
      Note: RuntimeFactory with DefaultRegistryModules() already includes fs
    - Path: internal/cssvisualdiff/dsl/registrar.go
      Note: Registers require('diff').compareRegion() — the core comparison primitive for the verb
    - Path: internal/cssvisualdiff/dsl/scripts/catalog.js
      Note: Builtin catalog verbs — reference for browser lifecycle and catalog APIs
    - Path: internal/cssvisualdiff/dsl/scripts/compare.js
      Note: Builtin compare.region verb — reference for how verbs call diff.compareRegion()
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-27T21:36:40.4591398-04:00
WhatFor: ""
WhenToUse: ""
---









# JSVerb Example: Generate Review Site Data Directory

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- css-visual-diff
- jsverb
- review-site
- yaml
- goja

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
