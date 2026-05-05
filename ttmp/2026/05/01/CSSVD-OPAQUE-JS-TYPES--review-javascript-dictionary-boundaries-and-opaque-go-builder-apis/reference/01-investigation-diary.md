---
Title: Investigation diary
Ticket: CSSVD-OPAQUE-JS-TYPES
Status: active
Topics:
    - javascript-api
    - goja
    - type-safety
    - code-review
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/jsapi/catalog.go
      Note: Source file reviewed during the investigation.
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Source file reviewed during the investigation.
    - Path: internal/cssvisualdiff/jsapi/probe.go
      Note: Source file reviewed during the investigation.
    - Path: internal/cssvisualdiff/jsapi/proxy.go
      Note: Source file reviewed during the investigation.
    - Path: internal/cssvisualdiff/jsapi/unwrap.go
      Note: Source file reviewed during the investigation.
ExternalSources: []
Summary: Chronological investigation diary for the opaque Go-backed JavaScript types code review.
LastUpdated: 2026-05-01T09:55:00-04:00
WhatFor: Records commands, evidence, conclusions, and validation for the CSSVD-OPAQUE-JS-TYPES review ticket.
WhenToUse: Use when continuing the review or implementing the proposed builder/API tightening phases.
---






# Diary

## Goal

This diary records the investigation and documentation work for reviewing JavaScript dictionary boundaries in `css-visual-diff` and designing a tighter opaque Go-backed type/builder strategy.

## Step 1: Create ticket and map the JavaScript-to-Go boundary

I created a new docmgr ticket for the review, then scanned the repository for `map[string]any`, `Record<string, ...>`, Goja Proxy usage, builder APIs, and JavaScript examples. The goal was to find where JavaScript dictionaries are still used as API contracts and where the code already has opaque Go-backed types.

The scan showed that the newer lower-level JS API already has a strong Proxy-backed handle/builder system, while older page inspect/preflight/catalog APIs still accept raw dictionaries. That shaped the design guide around extending the existing pattern rather than replacing it.

### Prompt Context

**User prompt (verbatim):** "Create a new code review docmgr ticket. Start by analyzing the codebase for map[string]any / javascript dict objects being used to communicate with go, when we could use opaque go types with builder APIs to have tighter control over the type correctness and be able to give in context feedback at runtime. Also analyze the existing go opaque types to see if their apis could be improved, if they can have better builder apis.

 reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the     
 system needed to understand what it is, with prose paragraphs and bullet                                                              
 point sand pseudocode and diagrams and api references and file                                                                        
   references.                                                                                                                         
   It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a docmgr ticket, perform an evidence-backed code review of raw JS object/map boundaries versus opaque Go-backed builders, write a detailed intern-facing design/implementation guide, and upload the ticket docs to reMarkable.

**Inferred user intent:** Improve runtime type correctness and developer feedback in the embedded JavaScript API by moving more domain values from loose dictionaries to Go-owned handles/builders.

### What I did

- Ran `docmgr status --summary-only` to confirm the ticket root.
- Created ticket `CSSVD-OPAQUE-JS-TYPES`.
- Added a design document and investigation diary.
- Ran repository scans including:
  - `rg -n "map\[string\](any|interface\{\})|Record<string|goja\.DynamicObject|NewProxy|ProxyRegistry|newProxyValue|unwrap|ExportTo|Export\(|ToValue|Set\(" internal cmd pkg web examples ...`
  - `find internal -type f -name '*.go' | sort`
- Read the main `internal/cssvisualdiff/jsapi/*` files, relevant service types, examples, and public JavaScript API docs.

### Why

The user asked for a code-review/design ticket, not a code change. The design needed concrete file-backed evidence before recommending API changes.

### What worked

- The repository already has a strong Proxy-backed abstraction in `internal/cssvisualdiff/jsapi/proxy.go` and `unwrap.go`.
- Strict APIs `cvd.extract(...)` and `cvd.snapshot(...)` already demonstrate raw object rejection and typed unwrapping.
- Existing tests in `builders_test.go` and `proxy_test.go` document validation and wrong-parent errors.

### What didn't work

- The first broad `rg` produced a lot of historical ticket/documentation hits under `ttmp`, so I reran focused searches against `internal`, `cmd`, `pkg`, `web`, and `examples`.
- Existing unrelated working tree changes were present before this work:
  - modified older ticket files under `ttmp/2026/04/28/...`
  - untracked `.bin/`
  - untracked `.orig` files
  I did not touch those.

### What I learned

The codebase is mid-migration: lower-level script-native APIs are already typed/opaque, while artifact-oriented and catalog-oriented APIs still rely on raw object compatibility.

### What was tricky to build

The main challenge was separating intentionally plain result data from risky command/config dictionaries. `map[string]any` is not automatically bad: it is appropriate for final JSON-like outputs, diff inputs, and metadata extension fields. The design guide therefore uses a boundary classification rule rather than a blanket ban.

### What warrants a second pair of eyes

- The proposed per-module registry refactor should be reviewed carefully because it touches all Proxy creation/unwrapping paths.
- The recommendation to let legacy page inspect APIs accept both raw objects and probe builders needs compatibility tests before implementation.

### What should be done in the future

Implement the design phases in order: tests first, per-module state, prepare builders, probe-builder support in inspect/preflight, inspect option builders, catalog target builders, then probe/extractor API refinements.

### Code review instructions

- Start with `internal/cssvisualdiff/jsapi/proxy.go` and `unwrap.go` to understand the current opaque type mechanism.
- Then read `internal/cssvisualdiff/jsapi/module.go` for raw dictionary boundaries.
- Validate current behavior with `go test ./internal/cssvisualdiff/jsapi -count=1` before changing APIs.

### Technical details

Key evidence files:

- `internal/cssvisualdiff/jsapi/module.go`
- `internal/cssvisualdiff/jsapi/proxy.go`
- `internal/cssvisualdiff/jsapi/unwrap.go`
- `internal/cssvisualdiff/jsapi/target.go`
- `internal/cssvisualdiff/jsapi/probe.go`
- `internal/cssvisualdiff/jsapi/extractor.go`
- `internal/cssvisualdiff/jsapi/extract.go`
- `internal/cssvisualdiff/jsapi/snapshot.go`
- `internal/cssvisualdiff/jsapi/catalog.go`
- `internal/cssvisualdiff/service/runtime_types.go`
- `internal/cssvisualdiff/service/types.go`
- `internal/cssvisualdiff/service/extract.go`
- `internal/cssvisualdiff/service/snapshot.go`
- `internal/cssvisualdiff/service/catalog_service.go`

## Step 2: Write the intern-facing design guide

I wrote the primary design document as an onboarding-oriented code review and implementation guide. It explains the runtime architecture, maps raw dictionary boundaries, reviews the existing Proxy-backed handle/builder model, and proposes phased API improvements.

The guide intentionally includes prose, bullet lists, pseudocode, API sketches, diagrams, and file references so a new intern can understand both the current system and the recommended implementation sequence.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce the detailed deliverable inside the new ticket after code analysis.

**Inferred user intent:** Leave behind a practical design artifact that can guide implementation, review, and onboarding.

### What I did

- Replaced the generated design-doc template with a complete guide.
- Included sections for executive summary, problem statement, architecture orientation, evidence, gap analysis, proposed APIs, implementation phases, tests, risks, and intern checklists.
- Anchored key claims to files and line ranges observed during the investigation.

### Why

The user specifically requested a detailed analysis/design/implementation guide for a new intern, not just a short code review summary.

### What worked

The current source organization made the review straightforward because most JavaScript boundary code lives under `internal/cssvisualdiff/jsapi` and the service contracts are in `internal/cssvisualdiff/service`.

### What didn't work

N/A.

### What I learned

The best near-term implementation is additive: keep legacy raw objects for compatibility but allow/encourage opaque builders in those same APIs. That avoids a breaking migration while still improving type correctness for new scripts.

### What was tricky to build

The design needed to preserve legitimate dictionary use. For example, catalog metadata and generic diff inputs should remain flexible, while prepare specs and reusable probes benefit from opaque builders.

### What warrants a second pair of eyes

- Whether to extend `cvd.target(...)` with catalog metadata or introduce a separate `cvd.catalogTarget(...)` builder.
- Whether to add strict variants or compatibility overloads for page inspect APIs.

### What should be done in the future

After review, create implementation tickets for the individual phases rather than doing all API changes in one large PR.

### Code review instructions

Review the design doc for:

- Correctness of source-file references.
- Whether the boundary classification rule matches project goals.
- Whether the phased implementation plan is granular enough for an intern.

### Technical details

Primary document:

- `ttmp/2026/05/01/CSSVD-OPAQUE-JS-TYPES--review-javascript-dictionary-boundaries-and-opaque-go-builder-apis/design-doc/01-opaque-go-backed-javascript-types-review-and-builder-api-guide.md`
