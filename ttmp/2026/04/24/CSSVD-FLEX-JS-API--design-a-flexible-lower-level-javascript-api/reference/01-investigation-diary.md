---
Title: Investigation diary
Ticket: CSSVD-FLEX-JS-API
Status: active
Topics:
    - visual-regression
    - browser-automation
    - tooling
    - chromedp
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md
      Note: Main design deliverable produced from this investigation.
ExternalSources: []
Summary: "Chronological diary for researching and documenting the flexible lower-level JavaScript API design."
LastUpdated: 2026-04-24T18:55:00-04:00
WhatFor: "Use this diary to understand what was read, why the design was shaped this way, what worked, what failed, and how to continue implementation."
WhenToUse: "When resuming the CSSVD-FLEX-JS-API ticket or reviewing the design guide."
---

# Diary

## Goal

This diary records the research and documentation work for designing a more flexible, lower-level JavaScript API for `css-visual-diff`. The goal was to create a new ticket and write an intern-facing guide that explains the current system, the gaps in the current high-level JS API, and a phased implementation plan for locators, extractors, snapshots, diffs, fluent builders, workflow composition, and YAML migration.

## Step 1: Create the ticket and gather evidence

I created a new ticket workspace for the lower-level JavaScript API design and then read the existing documentation and implementation files that define the current behavior. This step established the baseline: the current JS API is Promise-first and useful, but it is intentionally close to the high-level CLI/YAML concepts.

The important conclusion from this evidence pass was that the new API should be additive. It should not replace `page.preflight()` / `page.inspectAll()` immediately. Instead, it should add lower-level primitives beneath those methods and reuse the same Go service layer.

### Prompt Context

**User prompt (verbatim):** "Ok, read the documents, create a new ticket for a more flexible JS API, and reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticket, study the existing CSS visual diff documentation/source, write a very detailed design and implementation guide for a flexible lower-level JS API, store it in the ticket, and upload it to reMarkable.

**Inferred user intent:** The user wants the next API design to be concrete enough that an intern can implement it without needing to rediscover the existing architecture or design rationale.

### What I did

- Created ticket `CSSVD-FLEX-JS-API` under the repository-local `ttmp` root.
- Created a primary design document and an investigation diary using `docmgr`.
- Read current user-facing docs:
  - `README.md`
  - `docs/js-api.md`
  - `docs/js-verbs.md`
- Read the prior Goja/jsverbs design guide:
  - `ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/design/01-goja-javascript-api-analysis-design-and-implementation-guide.md`
- Read current source files for:
  - YAML schema and config loading,
  - native Goja module registration,
  - catalog/config adapters,
  - lazy JS verbs command discovery,
  - browser/preflight/style/prepare/inspect/catalog services,
  - integration tests proving the current API.

### Why

A design for a lower-level API has to be grounded in what already exists. The repository already has a working Goja/jsverbs layer and a reusable service layer, so the correct design is not greenfield. The guide needed to explain how to extend the current system safely.

### What worked

- `docmgr ticket create-ticket --root ./ttmp ...` created the ticket in the repository-local `ttmp` directory.
- The current docs clearly distinguish the high-level JS API from repository-scanned verbs.
- The current service layer already provides the right foundation for lower-level APIs: browser/page, prepare, preflight, style extraction, inspect artifacts, and catalog writing.

### What didn't work

- No command failure occurred during this step.
- The only known environmental caveat from the previous ticket remains relevant: `docmgr status` still resolves a parent workspace config by default, so I explicitly used `--root ./ttmp` when creating the ticket and docs.

### What I learned

- `docs/js-api.md` already documents Promise-first behavior, typed errors, preflight, inspect, catalog, YAML interop, and concurrency guidance.
- `docs/js-verbs.md` already documents the operator-facing repository-scanned CLI layer.
- `internal/cssvisualdiff/dsl/cvd_module.go` is the current adapter hot spot and should be split before adding many more methods.
- `internal/cssvisualdiff/service` is the correct place to add locator, extraction, snapshot, and diff logic.

### What was tricky to build

The tricky part was drawing a clear boundary between “current implemented API” and “future proposed API.” The design must not imply that locators, extractors, snapshots, or diffs already exist. It therefore names those as proposed layers and maps each one to concrete future files and phases.

### What warrants a second pair of eyes

- Whether `page.locator()` should mimic Playwright-style behavior or remain intentionally smaller.
- Whether `cvd.diff()` should start as a pure structural JSON diff or immediately include CSS-aware normalization.
- Where exactly to implement the per-page operation queue: `pageState` in the adapter or `service.PageService` in the service layer.

### What should be done in the future

- Run `docmgr doctor` after all documents are written.
- Upload the resulting document bundle to reMarkable.
- When implementation begins, start with adapter refactoring and page-operation serialization before adding locator methods.

### Code review instructions

- Start with the design guide:
  - `ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md`
- Check the references section for all files read.
- Validate that proposed phases preserve existing high-level JS API compatibility.

### Technical details

Commands used:

```bash
docmgr status --summary-only

docmgr ticket create-ticket \
  --root ./ttmp \
  --ticket CSSVD-FLEX-JS-API \
  --title "Design a flexible lower-level JavaScript API" \
  --topics css-visual-diff,javascript,goja,api-design,visual-diff

docmgr doc add \
  --root ./ttmp \
  --ticket CSSVD-FLEX-JS-API \
  --doc-type design-doc \
  --title "Flexible JavaScript API analysis design and implementation guide"

docmgr doc add \
  --root ./ttmp \
  --ticket CSSVD-FLEX-JS-API \
  --doc-type reference \
  --title "Investigation diary"
```

## Step 2: Write the primary design guide

I wrote the primary analysis/design/implementation guide. The document explains the current system, defines vocabulary, maps current YAML concepts to proposed JavaScript concepts, gives detailed API sketches, proposes implementation files, includes pseudocode and Mermaid diagrams, and lays out phased implementation and validation plans.

The design intentionally proposes a layered API rather than one large new abstraction. The recommended implementation order is: refactor adapters, serialize per-page operations, implement DOM locator services, expose `page.locator()`, add extractor builders and `cvd.extract()`, add builders, add snapshots, add diffs/reports, then add workflow builders last.

### Prompt Context

**User prompt (verbatim):** same as Step 1.

**Assistant interpretation:** Produce the requested detailed intern-facing design guide and store it in the ticket.

**Inferred user intent:** Make the next implementation phase easy to delegate and review.

### What I did

- Wrote `design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md`.
- Updated `index.md` with links and ticket overview.
- Updated `tasks.md` with completed documentation work, delivery tasks, and future implementation phases.
- Updated `changelog.md` with the documentation milestone.
- Updated this diary.

### Why

A lower-level JS API will touch many parts of the system: Goja runtime adapters, services, selectors, DOM extraction, artifact writing, catalog data, and CLI verbs. The guide gives an intern a map before implementation begins.

### What worked

- The existing codebase already has strong seams for the design:
  - `service/browser.go` for browser/page ownership,
  - `service/preflight.go` for selector status,
  - `service/style.go` for computed CSS,
  - `service/inspect.go` for artifact writing,
  - `service/catalog_service.go` for Go-backed manifests and indexes,
  - `dsl/cvd_module.go` for Promise-first Goja exposure.
- The docs already contain enough current API references to ground the proposal.

### What didn't work

- No command failure occurred while writing the documents.
- Upload is not yet completed at this diary step.

### What I learned

The lower-level API should be implemented as an additive second layer, not as a replacement. The safest first public feature is `page.locator()` plus a handful of async locator methods. Workflow builders should come much later.

### What was tricky to build

The design needed to be detailed without overcommitting to too much implementation in one PR. I addressed this by separating:

- API reference sketches,
- service architecture,
- Goja adapter pseudocode,
- phased implementation plan,
- acceptance criteria.

### What warrants a second pair of eyes

- The proposed `cvd.extract()` result shape.
- The proposed `cvd.snapshot()` vs current `page.inspectAll()` relationship.
- The proposed file split under `internal/cssvisualdiff/dsl`.

### What should be done in the future

- Convert the API reference sections into `docs/js-low-level-api.md` once implementation starts.
- Add smoke scripts under this ticket during implementation.
- Consider generating TypeScript declaration files from the documented API once stable.

### Code review instructions

Read the design in this order:

1. Executive summary.
2. Section 3, current system.
3. Section 6, proposed API layers.
4. Section 9, implementation architecture.
5. Section 12, implementation phases.
6. Section 19, file-by-file implementation guide.

### Technical details

Main written file:

```text
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md
```

## Step 3: Validate docs and skip reMarkable upload for now

I validated the ticket documentation and prepared the PDF bundle for delivery. The reMarkable upload path hit the same cloud mutation/upload failure pattern seen earlier, so the upload was intentionally skipped after the user asked to skip it for now.

The useful result is that the ticket and local PDF are ready. The remaining blocker is only external reMarkable upload behavior, not the design documentation itself.

### Prompt Context

**User prompt (verbatim):** "ok skip the upload for now"

**Assistant interpretation:** Stop retrying reMarkable upload and leave the generated local ticket/PDF artifacts in place.

**Inferred user intent:** Avoid spending more time on the known reMarkable cloud upload failure and finish the documentation handoff.

### What I did

- Ran `docmgr doctor --root ./ttmp --ticket CSSVD-FLEX-JS-API --stale-after 30`.
- Fixed unknown-topic warnings by changing ticket topics to known vocabulary values.
- Re-ran doctor successfully.
- Ran `remarquee status` and `remarquee cloud account --non-interactive` successfully.
- Ran a dry-run bundle upload successfully.
- Tried actual bundle upload to `/ai/2026/04/24/CSSVD-FLEX-JS-API`; it failed creating the remote ticket directory with HTTP 400.
- Tried actual bundle upload directly to `/ai/2026/04/24`; it failed uploading the PDF with HTTP 400.
- Generated a local PDF manually with pandoc at `pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide.pdf`.
- Tried `remarquee cloud put` for that PDF; it also failed with HTTP 400.
- Updated tasks and changelog to record the skipped upload.

### Why

The design deliverable should not be blocked by a known cloud upload mutation issue when local docs and the generated PDF are already available.

### What worked

- `docmgr doctor` passes.
- `remarquee status` and account lookup work.
- `remarquee upload bundle --dry-run` works.
- Local PDF generation works.

### What didn't work

Actual cloud upload failed:

```text
Error: failed to create remote directory "CSSVD-FLEX-JS-API" under "24": request failed with status 400
```

Then direct upload to the existing date folder failed:

```text
Error: failed to upload file [/tmp/remarquee-upload-bundle-.../CSSVD-FLEX-JS-API - Flexible JavaScript API Design Guide.pdf]: request failed with status 400
```

Finally, direct `cloud put` of the locally generated PDF failed:

```text
Error: failed to upload file [.../CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide.pdf]: request failed with status 400
```

### What I learned

The reMarkable issue is not specific to bundle conversion or PDF generation. The account can read/list, but cloud mutation/upload operations are failing with HTTP 400.

### What was tricky to build

The generated markdown bundle initially failed direct pandoc conversion because concatenated documents included multiple YAML frontmatter blocks. I generated the local PDF by stripping frontmatter before invoking pandoc.

### What warrants a second pair of eyes

The reMarkable upload tooling or account/cloud API state needs separate investigation if upload becomes important again.

### What should be done in the future

Retry upload later or transfer the local PDF manually:

```text
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide.pdf
```

### Code review instructions

No code changed. Review the ticket docs and generated PDF.

### Technical details

Validation command:

```bash
docmgr doctor --root ./ttmp --ticket CSSVD-FLEX-JS-API --stale-after 30
```

Successful result:

```text
## Doctor Report (1 findings)

### CSSVD-FLEX-JS-API

- ✅ All checks passed
```

## Step 4: Retry reMarkable upload successfully with a short PDF filename

The user asked to try the upload again. I retried using the already-generated local PDF but copied it to a short temporary filename first. This succeeded, which suggests the earlier HTTP 400 may have been related to remote folder creation, long names, spaces, or metadata generated by the bundle upload path rather than the PDF content itself.

The document is now available in the reMarkable cloud date folder as `cssvd-flex-api`.

### Prompt Context

**User prompt (verbatim):** "try upload again"

**Assistant interpretation:** Retry the reMarkable upload now instead of leaving it skipped.

**Inferred user intent:** See whether the previous cloud upload failure was transient or avoidable with a different upload path.

### What I did

- Checked `remarquee status`.
- Checked the cloud account with `remarquee cloud account --non-interactive`.
- Listed `/ai/2026/04/24`.
- Copied the local generated PDF to `/tmp/cssvd-flex-api.pdf`.
- Uploaded the short-named PDF with `remarquee cloud put /tmp/cssvd-flex-api.pdf /ai/2026/04/24 --non-interactive`.
- Verified the remote listing.
- Updated `tasks.md` and `changelog.md`.

### Why

The previous upload attempts failed with HTTP 400 using folder creation and longer bundle/document names. Retrying with a short filename was a low-risk way to test whether upload itself could succeed.

### What worked

Upload succeeded:

```text
OK: uploaded /tmp/cssvd-flex-api.pdf -> /ai/2026/04/24
```

Remote listing verified:

```text
[f]	02-postmortem-and-next-developer-handoff
[f]	cssvd-flex-api
[d]	CSSVD-GOJA-JS-API
```

### What didn't work

No new failure occurred in this step. Earlier failures remain recorded in Step 3.

### What I learned

A short filename and direct `cloud put` to an existing date folder can avoid the previous upload failure path. The failed operation may have involved remote directory creation or document naming rather than basic account/auth/upload capability.

### What was tricky to build

The upload target is not the originally requested ticket subfolder because creating `/ai/2026/04/24/CSSVD-FLEX-JS-API` previously failed. The successful upload is in the existing date folder:

```text
/ai/2026/04/24/cssvd-flex-api
```

### What warrants a second pair of eyes

If future organization matters, investigate why creating a new reMarkable folder under `/ai/2026/04/24` returned HTTP 400 while direct file upload to the existing date folder succeeded.

### What should be done in the future

Optionally rename/move the document on-device or retry folder creation later. For now, the PDF is uploaded and verified.

### Code review instructions

No code changed. Review the ticket docs and uploaded PDF.

### Technical details

Commands:

```bash
remarquee status
remarquee cloud account --non-interactive
remarquee cloud ls /ai/2026/04/24 --long --non-interactive
cp ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide.pdf /tmp/cssvd-flex-api.pdf
remarquee cloud put /tmp/cssvd-flex-api.pdf /ai/2026/04/24 --non-interactive
remarquee cloud ls /ai/2026/04/24 --long --non-interactive
```

## Step 5: Update design for jsapi subpackage, Go-backed Proxy objects, and no workflow builder

After reviewing the design direction, the user clarified two important architectural preferences. First, the native JS API should use Go-backed structs wrapped in Goja Proxy objects rather than plain JavaScript objects, following the Discord bot UI DSL pattern. Second, there is no need for a workflow builder; ordinary JavaScript functions and loops should remain the orchestration mechanism.

I updated the design guide to reflect those decisions. The most important changes are that `internal/cssvisualdiff/dsl/cvd_module.go` should not simply grow new files in the same package; instead, the native `require("css-visual-diff")` API should move to a dedicated `internal/cssvisualdiff/jsapi` package. The `dsl` package should stay focused on jsverbs host/runtime/scanning responsibilities.

### Prompt Context

**User prompt (verbatim):** "Update the design implementation guide with all these things (refactoring cvd_module.go, etc...)"

**Follow-up prompt (verbatim):** "no need for workflow builder."

**Assistant interpretation:** Update the existing ticket design guide with the subpackage/refactor plan, Goja Proxy-backed structs/builders, method-specific validation/error behavior, and remove the workflow-builder phase.

**Inferred user intent:** The user wants the design guide to reflect the refined implementation direction before anyone starts coding.

### What I did

- Updated the primary design guide frontmatter and summary.
- Replaced the “plain objects plus optional fluent builders” principle with a strict Go-backed Proxy builder/handle principle.
- Added a dedicated `internal/cssvisualdiff/jsapi` package layout.
- Added Proxy infrastructure guidance: `proxy.go`, `unwrap.go`, method-owner maps, wrong-parent errors, unknown-method errors, and type-mismatch errors.
- Rewrote the Goja adapter section around Proxy-backed `BrowserHandle`, `PageHandle`, `LocatorHandle`, `TargetBuilder`, `ProbeBuilder`, `ExtractorHandle`, and `CatalogHandle`.
- Rewrote the implementation phases so Phase 1 moves the native API out of `dsl/cvd_module.go`, Phase 2 adds Proxy infrastructure, and there is no workflow-builder phase.
- Updated the review checklist, acceptance criteria, file-by-file implementation guide, references, tasks, and changelog.

### Why

The Proxy-backed approach gives better control over validation and much better feedback for LLM-written JavaScript. It can distinguish a typo, an unknown method, a method that belongs to a different object type, and a raw-object type mismatch. This is exactly what the Discord bot UI DSL demonstrated.

### What worked

- The design now has a clearer package boundary:
  - `dsl` = jsverbs host/runtime/scanning.
  - `jsapi` = native `require("css-visual-diff")` module, Goja Proxy wrappers, Promise bridge, typed errors, and adapters.
  - `service` = Go implementation with no Goja dependency.
- The workflow-builder references were removed from the implementation plan and replaced with explicit JavaScript control flow.

### What didn't work

No command failed during this documentation update.

### What I learned

The lower-level API should not optimize for arbitrary JavaScript flexibility. It should optimize for controlled fluent authoring with precise host-side errors. Raw object compatibility can remain in the existing high-level API, but new strict APIs should prefer typed Proxy-backed values.

### What was tricky to build

The tricky part was preserving compatibility for current scripts while making the new lower-level API stricter. The design now handles this by separating compatibility methods (`page.preflight`, `page.inspectAll`) from strict new methods (`cvd.extract`, `cvd.snapshot`).

### What warrants a second pair of eyes

- The Proxy registry/unwrapping design should be reviewed carefully before implementation.
- The exact package migration path from `dsl/catalog_adapter.go` and `dsl/config_adapter.go` into `jsapi` should be done as a no-behavior refactor first.
- The method-owner map should allow methods that intentionally exist on multiple object types, such as `close` or `attributes`.

### What should be done in the future

- Regenerate and upload the updated PDF to reMarkable.
- When coding begins, implement Phase 1 as a mechanical refactor with no behavior changes.
- Add Proxy behavior tests before implementing many user-facing methods.

### Code review instructions

Review the updated sections in the design guide:

1. Section 5: Design principles.
2. Section 6: Proposed API layers.
3. Section 9: Architecture for implementation.
4. Section 10: Goja adapter design with Proxy-backed handles.
5. Section 12: Implementation phases.
6. Section 19: File-by-file implementation guide.

### Technical details

Key design files updated:

```text
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/design-doc/01-flexible-javascript-api-analysis-design-and-implementation-guide.md
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/tasks.md
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/changelog.md
```

## Step 6: Regenerate and upload the updated design guide to reMarkable

After updating the design guide for the `jsapi` subpackage, Go-backed Proxy object model, and removal of the workflow builder phase, I regenerated the PDF bundle and uploaded the updated version to reMarkable.

I used a short remote filename again because the earlier successful upload showed that short names avoid the HTTP 400 failure path. To avoid overwriting or deleting the previous upload, I uploaded the new PDF as `cssvd-flex-api-updated`.

### Prompt Context

**User prompt (verbatim):** "upload document to remarkable when done"

**Assistant interpretation:** After finishing the design guide updates, regenerate the document bundle and upload the updated version to reMarkable.

**Inferred user intent:** Make the latest design, not the older pre-Proxy version, available on the reMarkable device/cloud.

### What I did

- Regenerated a bundled PDF from the updated design guide, diary, tasks, and changelog.
- Stripped YAML frontmatter while building the bundle so pandoc would not treat multiple frontmatter blocks as invalid metadata.
- Saved the updated local PDF at:
  - `pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide-updated.pdf`
- Copied it to a short temporary upload filename:
  - `/tmp/cssvd-flex-api-updated.pdf`
- Uploaded it with `remarquee cloud put`.
- Verified the remote listing.
- Updated tasks and changelog.

### Why

The reMarkable copy needed to reflect the latest design decisions: `internal/cssvisualdiff/jsapi`, Proxy-backed handles/builders, strict lower-level APIs, typed unwrapping, and no workflow builder.

### What worked

Upload succeeded:

```text
OK: uploaded /tmp/cssvd-flex-api-updated.pdf -> /ai/2026/04/24
```

Remote listing verified:

```text
[f]	02-postmortem-and-next-developer-handoff
[f]	cssvd-flex-api
[f]	cssvd-flex-api-updated
[d]	CSSVD-GOJA-JS-API
[d]	GCB-009
[d]	PYXIS-COMPONENT-VISUAL-PARITY
```

### What didn't work

Pandoc emitted a non-fatal font warning:

```text
[WARNING] Missing character: There is no ✅ (U+2705) (U+2705) in font [lmmono10-regular]:!
```

The PDF was still generated and uploaded successfully.

### What I learned

The reliable upload path is still direct `remarquee cloud put` with a short PDF filename into an existing folder.

### What was tricky to build

The generated ticket docs include YAML frontmatter and Unicode symbols. The frontmatter must be stripped for a concatenated bundle. The Unicode warning is cosmetic but may be worth avoiding in future PDF-bound docs.

### What warrants a second pair of eyes

If annotations on uploaded docs become important, avoid overwriting existing documents with `--force`. This upload intentionally used a new filename instead.

### What should be done in the future

Use `/ai/2026/04/24/cssvd-flex-api-updated` as the current reMarkable copy of the design.

### Code review instructions

No code changed. Review the updated design guide and PDF bundle.

### Technical details

Commands included:

```bash
pandoc "$TMP_MD" \
  -o "$TICKET_DIR/pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide-updated.pdf" \
  --toc --toc-depth=2 --pdf-engine=xelatex
cp "$TICKET_DIR/pdf/CSSVD-FLEX-JS-API-flexible-javascript-api-design-guide-updated.pdf" /tmp/cssvd-flex-api-updated.pdf
remarquee cloud put /tmp/cssvd-flex-api-updated.pdf /ai/2026/04/24 --non-interactive
remarquee cloud ls /ai/2026/04/24 --long --non-interactive
```
