---
Title: Design Goja JavaScript API for programmable visual catalog workflows
Ticket: CSSVD-GOJA-JS-API
Status: active
Topics:
    - css-visual-diff
    - goja
    - javascript-api
    - visual-regression
    - catalog
    - automation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: design/01-goja-javascript-api-analysis-design-and-implementation-guide.md
      Note: Primary design guide, now updated with repository-scanned jsverbs CLI architecture.
    - Path: reference/01-implementation-research-diary.md
      Note: Research diary for the update pass and evidence from Discord/loupedeck/go-go-goja.
ExternalSources: []
Summary: "Design and implementation workspace for the css-visual-diff Goja/jsverbs API, including repository-scanned JavaScript verbs exposed as CLI commands."
LastUpdated: 2026-04-24T13:00:00-04:00
WhatFor: ""
WhenToUse: ""
---

# Design Goja JavaScript API for programmable visual catalog workflows

Document workspace for CSSVD-GOJA-JS-API.

Current focus: productize the existing `internal/cssvisualdiff/dsl` Goja/jsverbs prototype into a repository-scanned `css-visual-diff verbs ...` command surface, while adding a coherent `require("css-visual-diff")` module for programmable browser/page/preflight/inspect/catalog workflows.
