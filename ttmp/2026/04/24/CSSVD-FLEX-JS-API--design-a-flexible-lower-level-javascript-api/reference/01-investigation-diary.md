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

## Step 7: Expand ticket tasks into granular implementation phases

Before continuing code work, I paused to make the ticket task list much more granular. The previous task list only had broad implementation phases, which was not detailed enough for task-by-task execution or for deciding when to commit.

The ticket now has a phase-by-phase checklist from the no-behavior `jsapi` refactor through Proxy infrastructure, page serialization, DOM services, locators, builders, extraction, snapshots, diffs, docs, smoke scripts, and final delivery. This gives us a concrete progress tracker for the rest of CSSVD-FLEX-JS-API.

### Prompt Context

**User prompt (verbatim):** "Let's work on CSSVD-FLEX-JS-API, task by task, keep a detailed diary (read the current diary so far), commit at appropriate intervals."

**Follow-up prompt (verbatim):** "add detailed tasks split in phases to the ticket, so that we have a more granular wya of keeping track where we are."

**Assistant interpretation:** Read the existing diary, then make the ticket task list detailed enough to track implementation phase by phase before continuing code.

**Inferred user intent:** The user wants a disciplined implementation process with explicit checkpoints, diary updates, validation, and commits.

### What I did

- Read the diary skill, docmgr skill, git commit instructions, and go-go-goja module authoring skill.
- Read the current CSSVD-FLEX-JS-API diary and tasks file.
- Checked branch and working tree state.
- Committed the previously untracked ticket/design documents before implementation work:
  - `17240de docs: add flexible js api implementation ticket`
- Rewrote `tasks.md` into detailed phased checklists:
  - Phase 0: ticket/design/baseline bookkeeping.
  - Phase 1: no-behavior `internal/cssvisualdiff/jsapi` package refactor.
  - Phase 2: Proxy infrastructure and typed unwrapping.
  - Phase 3: per-page operation serialization.
  - Phase 4: service DOM locator primitives.
  - Phase 5: `page.locator()` Proxy handle and locator methods.
  - Phase 6: target/probe/extractor builders.
  - Phase 7: strict `cvd.extract(locator, extractors)`.
  - Phase 8: strict `cvd.snapshot(page, probes, options)`.
  - Phase 9: diff/report/write primitives.
  - Phase 10: docs/examples/smoke scripts/delivery.

### Why

The implementation will touch multiple subsystems. A granular task list prevents losing track of intermediate safety steps such as baseline tests, no-behavior refactor validation, Proxy error tests, per-page serialization tests, and documentation updates.

### What worked

- The existing diary had enough context to resume safely.
- The ticket docs were committed cleanly before implementation changes.
- The task list now separates mechanical refactoring from behavior changes.

### What didn't work

No command failed during this planning step.

### What I learned

Phase 1 needs to be treated as a no-behavior refactor and validated before adding Proxy behavior. The initial code move also makes it clear that `decodeInto(...)` currently lives in `dsl/codec.go`, so Phase 1 must either move or duplicate the codec helper into `jsapi` while keeping `toPlainValue(...)` available to the existing `diff` and `report` modules.

### What was tricky to build

The tricky part is that I had already started the mechanical `git mv` for Phase 1 before the follow-up request arrived. I left those implementation file moves uncommitted and used this step to improve the ticket tracking first. This keeps the task-detail commit separate from the still-in-progress code refactor.

### What warrants a second pair of eyes

- The Phase 1 checklist should be reviewed to ensure it really is behavior-preserving.
- The Phase 2 Proxy checklist should be reviewed before implementation because it will define the long-term ergonomics and error model.

### What should be done in the future

Resume Phase 1 from the current working tree, finish the `jsapi` package move, run targeted tests, update the diary/changelog, and commit the no-behavior refactor separately.

### Code review instructions

Review:

```text
ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api/tasks.md
```

Validation:

```bash
docmgr doctor --root ./ttmp --ticket CSSVD-FLEX-JS-API --stale-after 30
```

### Technical details

Commands included:

```bash
git status --short --branch
git add ttmp/2026/04/24/CSSVD-FLEX-JS-API--design-a-flexible-lower-level-javascript-api
git commit -m "docs: add flexible js api implementation ticket"
git rev-parse HEAD
```

## Step 8: Complete Phase 1 and Phase 2 implementation

I completed the first two implementation phases in one focused pass. Phase 1 was the no-behavior refactor that moves the native `require("css-visual-diff")` API out of the broad `dsl` package and into a dedicated `internal/cssvisualdiff/jsapi` package. Phase 2 added the first reusable Goja Proxy infrastructure and typed unwrapping helpers for future lower-level handles/builders.

The current public API behavior should be unchanged. Existing `css-visual-diff` JS scripts still load the same module name, and the repository-scanned verb tests still exercise the same browser/page/catalog behavior. The new Proxy helpers are currently infrastructure-only; they are covered by unit tests but are not yet wired into the existing browser/page/catalog wrappers.

### Prompt Context

**User prompt (verbatim):** "continue, do phase 1 and phase 2 in their entirety."

**Assistant interpretation:** Finish both the no-behavior `jsapi` package refactor and the initial Proxy/unwrap infrastructure, validate them, update ticket bookkeeping, and prepare for a commit.

**Inferred user intent:** The user wants the implementation to progress through the granular ticket phases, not just update docs.

### What I did

- Moved native JS API files into `internal/cssvisualdiff/jsapi`:
  - `dsl/cvd_module.go` -> `jsapi/module.go`
  - `dsl/catalog_adapter.go` -> `jsapi/catalog.go`
  - `dsl/config_adapter.go` -> `jsapi/config.go`
- Changed the moved files to `package jsapi`.
- Renamed `registerCVDModule(...)` to exported `Register(...)`.
- Added `internal/cssvisualdiff/jsapi/codec.go` with a local `decodeInto(...)` helper.
- Kept `dsl/codec.go` in place so existing `diff` and `report` native modules can still use `decodeInto(...)` and `toPlainValue(...)`.
- Updated `internal/cssvisualdiff/dsl/registrar.go` to import `jsapi` and call `jsapi.Register(ctx, reg)`.
- Added `internal/cssvisualdiff/jsapi/proxy.go` with:
  - `ProxyRegistry`,
  - `ProxySpec`,
  - `MethodSpec`,
  - `ProxyMethod`,
  - `newProxyValue(...)`,
  - unknown-method errors,
  - wrong-parent errors,
  - type-mismatch errors,
  - simple did-you-mean suggestions.
- Added `internal/cssvisualdiff/jsapi/unwrap.go` with:
  - `unwrapProxyBacking[T](...)`,
  - `unwrapProxyBinding(...)`,
  - `mustUnwrapProxyBacking[T](...)`.
- Added `internal/cssvisualdiff/jsapi/proxy_test.go` covering:
  - unknown method errors,
  - wrong-parent method errors,
  - successful backing unwrap,
  - raw object rejection,
  - wrong-owner rejection.
- Marked Phase 1 and Phase 2 tasks complete and set the next active phase to Phase 3.

### Why

The native API had grown into a large `dsl/cvd_module.go` file even though `dsl` should own jsverbs host/scanning/runtime wiring. Moving the native `require("css-visual-diff")` adapter code into `jsapi` creates a cleaner boundary before adding lower-level handles, locators, builders, snapshots, and diffs.

The Proxy infrastructure is needed before adding strict lower-level APIs because those APIs should be able to distinguish Go-backed handles/builders from raw JS objects and provide LLM-friendly feedback when users call the wrong method on the wrong object.

### What worked

The no-behavior refactor compiled cleanly after adding the local `jsapi` codec helper and rewiring the registrar. Existing targeted tests and the full test suite passed.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/jsapi -count=1

go test ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff ./internal/cssvisualdiff/service -count=1

go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

The first targeted test run failed as expected after moving files but before adding the codec/helper rewiring:

```text
# github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/jsapi
internal/cssvisualdiff/jsapi/catalog.go:21:16: undefined: decodeInto
internal/cssvisualdiff/jsapi/catalog.go:108:16: undefined: decodeInto
internal/cssvisualdiff/jsapi/catalog.go:135:17: undefined: decodeInto
internal/cssvisualdiff/jsapi/catalog.go:190:16: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:152:17: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:196:17: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:328:15: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:358:16: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:402:17: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:414:17: undefined: decodeInto
internal/cssvisualdiff/jsapi/module.go:414:17: too many errors
FAIL	github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/jsapi [build failed]
# github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl [github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl.test]
internal/cssvisualdiff/dsl/registrar.go:51:2: undefined: registerCVDModule
FAIL	github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl [build failed]
```

The fix was to add `internal/cssvisualdiff/jsapi/codec.go` and replace the old registrar call with `jsapi.Register(ctx, reg)`.

### What I learned

The file move is clean if `dsl/codec.go` remains in `dsl` for the existing `diff` and `report` native modules while `jsapi/codec.go` gets its own `decodeInto(...)`. This intentionally duplicates a tiny JSON codec helper for now rather than creating another shared internal package before it is necessary.

Goja's `Runtime.NewProxy(...)` API is sufficient for the future builder/handle design. A hidden internal proxy id exposed through the `Get` trap gives strict APIs a reliable way to unwrap only values created by the `jsapi` package.

### What was tricky to build

The trickiest part was designing Proxy infrastructure that is useful now without prematurely converting browser/page/catalog wrappers. I kept Phase 2 infrastructure-only and added tests using a small fake `cvd.probe` Proxy. That proves the mechanics and error style without changing existing public API behavior.

Another subtle point is that `mustUnwrapProxyBacking(...)` currently panics with a Goja `TypeError` via `typeMismatchError(...)`. Future integration may want to route these through the existing CSSVD typed error class hierarchy, but for Phase 2 the important behavior is that wrong raw objects fail immediately with an actionable operation-specific message.

### What warrants a second pair of eyes

- Whether the proxy id registry should eventually delete bindings when handles close, or whether runtime-lifetime bindings are acceptable.
- Whether the hidden `__cssVisualDiffProxyID` property should become a symbol-like internal key later.
- Whether `typeMismatchError(...)` should construct `CvdError` subclasses instead of `TypeError` once strict APIs are wired into the exported module.
- Whether the simple Levenshtein did-you-mean threshold is too aggressive or too conservative.

### What should be done in the future

Phase 3 should add page operation serialization. After that, Phase 4 can add pure Go DOM locator services and Phase 5 can expose `page.locator()` using the new Proxy infrastructure.

### Code review instructions

Start review with the package boundary:

```text
internal/cssvisualdiff/dsl/registrar.go
internal/cssvisualdiff/jsapi/module.go
internal/cssvisualdiff/jsapi/catalog.go
internal/cssvisualdiff/jsapi/config.go
```

Then review the new Proxy infrastructure:

```text
internal/cssvisualdiff/jsapi/proxy.go
internal/cssvisualdiff/jsapi/unwrap.go
internal/cssvisualdiff/jsapi/proxy_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1
go test ./... -count=1
```

### Technical details

Important implementation details:

- `jsapi.Register(ctx, reg)` preserves the public module name `require("css-visual-diff")`.
- `dsl` still registers the legacy `diff` and `report` native modules.
- `ProxyRegistry` stores Go backing values keyed by an internal int64 id.
- `newProxyValue(...)` returns a Goja Proxy whose `Get` trap:
  - returns own methods,
  - returns a stable `toString`,
  - returns the hidden proxy id,
  - throws wrong-parent errors for known methods owned by another object,
  - throws unknown-method errors with available methods and did-you-mean suggestions otherwise.

## Step 9: Complete Phase 3 page operation serialization

I completed Phase 3 by adding a per-page mutex to the JavaScript API page state and routing every existing page operation through that guard. This keeps concurrent JavaScript calls such as `Promise.all([page.preflight(...), page.preflight(...)])` safe on a single page while still allowing independent pages to proceed through their own page-level locks.

This phase does not add new public API methods. It hardens the current Promise-first browser/page API so future locator and extraction methods can safely share the same page handle without racing CDP operations or mutating page target state concurrently.

### Prompt Context

**User prompt (verbatim):** "phase 3"

**Assistant interpretation:** Implement Phase 3 from the ticket: add per-page operation serialization, validate it, update docs/diary/tasks, and commit.

**Inferred user intent:** Prepare the current page API for lower-level locator/extractor methods by making same-page concurrent calls safe.

### What I did

- Added a `sync.Mutex` to `pageState` in `internal/cssvisualdiff/jsapi/module.go`.
- Added `pageState.runExclusive(...)`.
- Wrapped all current page operations in `runExclusive(...)`:
  - `page.goto(...)`
  - `page.prepare(...)`
  - `page.preflight(...)`
  - `page.inspect(...)`
  - `page.inspectAll(...)`
  - `page.close()`
- Added repository-scanned JS verb tests in `internal/cssvisualdiff/verbcli/command_test.go`:
  - `TestCVDModuleSerializesSamePagePromiseAll`
  - `TestCVDModuleAllowsConcurrentOperationsOnSeparatePages`
- Marked Phase 3 tasks complete and set the active phase to Phase 4.

### Why

Chromedp/CDP operations on a single page are not a good fit for unguarded concurrent use. The existing API returns Promises, so JavaScript users can naturally write `Promise.all(...)`. Rather than forbid that pattern, the page handle should serialize operations internally and preserve safety.

This is especially important before adding `page.locator()` and `locator.*` methods because users are likely to query several locators at once.

### What worked

The per-page mutex was a small focused change. Same-page concurrent calls now complete safely in a repository-scanned JS verb test. Separate page handles also complete concurrently from the JavaScript point of view because each page has its own `pageState` and lock.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModule(SerializesSamePagePromiseAll|AllowsConcurrentOperationsOnSeparatePages)' -count=1

go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

My first version of the same-page concurrency test used:

```js
page.inspect({ name: "a", selector: "#a" }, { artifacts: "json" })
```

That failed because `"json"` is not a supported inspect artifact format:

```text
promise rejected: ArtifactError: unsupported inspect format "json"
```

I changed the test to use three concurrent `page.preflight(...)` calls instead. That still exercises concurrent same-page CDP script evaluation without introducing artifact-format concerns.

### What I learned

The current `inspect` API expects the existing inspect format names such as `bundle`, `css-json`, or `html`; raw `json` is not a valid shortcut. For Phase 3, preflight is the cleaner concurrency smoke because it avoids filesystem artifact output and focuses on page-operation safety.

### What was tricky to build

The key subtlety is the target state on `pageState`. `goto` mutates `state.target`, and `prepare` mutates `state.target.Prepare`, while `inspect` and `inspectAll` read `state.target`. The lock needs to cover both CDP calls and target state access; otherwise a future concurrent `goto` plus `inspect` could read a half-updated target.

The lock is intentionally per page rather than per browser. That keeps operations on one page serialized while preserving concurrency between independent pages.

### What warrants a second pair of eyes

- Whether `page.close()` should set a closed flag and reject later operations with a clearer error.
- Whether long-running operations should support cancellation or timeouts at this layer.
- Whether Phase 5 locator methods should reuse this same `runExclusive(...)` helper or introduce batch operations to reduce queuing overhead.

### What should be done in the future

Phase 4 should add pure Go DOM locator primitives in `internal/cssvisualdiff/service/dom.go`. Phase 5 can then expose those primitives through `page.locator()` and the Proxy infrastructure.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/jsapi/module.go
```

Review `pageState`, `runExclusive(...)`, and the wrappers in `wrapPage(...)`.

Then review tests:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModule(SerializesSamePagePromiseAll|AllowsConcurrentOperationsOnSeparatePages)' -count=1
go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1
go test ./... -count=1
```

### Technical details

`pageState` now owns the synchronization primitive:

```go
type pageState struct {
    mu     sync.Mutex
    page   *service.PageService
    target config.Target
}

func (s *pageState) runExclusive(work func() (any, error)) (any, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    return work()
}
```

Each Promise work closure calls `state.runExclusive(...)`, so the work still runs asynchronously from JavaScript's perspective but only one operation touches the underlying page at a time.

## Step 10: Complete Phase 4 service DOM locator primitives

I completed Phase 4 by adding pure Go service-layer DOM locator primitives. These functions are intentionally independent of Goja and provide the reusable browser-page operations that the future `page.locator()` JavaScript API will wrap.

The new service file covers selector status, text, HTML, bounds, attributes, and computed style. Missing selectors return empty/structured results where that is natural, while invalid selectors return errors for direct locator operations. `LocatorStatus` reuses the existing preflight implementation so status behavior stays aligned with the current high-level API.

### Prompt Context

**User prompt (verbatim):** "phase 4"

**Assistant interpretation:** Implement Phase 4 from the ticket: add pure Go DOM locator primitives, test them, update docs/diary/tasks, and commit.

**Inferred user intent:** Build the Go service foundation required before exposing `page.locator()` in the JS API.

### What I did

- Added `internal/cssvisualdiff/service/dom.go`.
- Defined:
  - `LocatorSpec`
  - `TextOptions`
  - `ElementHTML`
- Implemented:
  - `LocatorStatus(...)`
  - `LocatorText(...)`
  - `LocatorHTML(...)`
  - `LocatorBounds(...)`
  - `LocatorAttributes(...)`
  - `LocatorComputedStyle(...)`
- Added `internal/cssvisualdiff/service/dom_test.go`.
- Added tests for:
  - existing visible selector,
  - missing selector,
  - hidden selector,
  - invalid selector,
  - text normalization,
  - inner/outer HTML,
  - bounds,
  - attributes including missing attributes,
  - computed styles.
- Marked Phase 4 tasks complete and set the active phase to Phase 5.

### Why

The lower-level JavaScript API should wrap reusable Go services rather than putting DOM logic directly in Goja adapters. This keeps service behavior testable without JavaScript runtime concerns and gives future CLI modes a shared implementation path.

### What worked

`LocatorStatus(...)` could reuse `PreflightProbes(...)`, preserving existing selector status behavior. `LocatorComputedStyle(...)` could reuse `EvaluateStyle(...)`, preserving current computed-style extraction semantics.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestLocatorDOMPrimitives' -count=1

go test ./internal/cssvisualdiff/service -count=1

go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

The first `TestLocatorDOMPrimitives` run failed because the test HTML contained a literal backslash-n sequence instead of an actual newline:

```text
Error: "Book\\n now" does not contain "Book now"
```

I fixed the fixture to use a real newline inside the button text so the existing whitespace normalization behavior matched the assertion.

### What I learned

The current preflight text normalization collapses real whitespace, but not a literal `\\n` sequence. That is correct browser behavior, and the test fixture was wrong.

For missing selectors, empty direct results are more useful than errors:

- text -> `""`
- HTML -> `{ exists: false, html: "" }`
- bounds -> `nil`
- attributes -> empty map
- computed style -> empty map

For invalid selectors, direct operations should return errors because the selector itself is malformed.

### What was tricky to build

The tricky part was deciding how much behavior to duplicate versus reuse. `LocatorStatus(...)` reuses `PreflightProbes(...)` and `LocatorComputedStyle(...)` reuses `EvaluateStyle(...)`. The direct text/html/bounds/attributes functions use small focused page scripts because the existing service layer did not expose those exact data shapes yet.

Another subtle point is attribute handling. DOM `getAttribute(...)` can return `null`, but the Go API returns `map[string]string`, so the script normalizes missing attributes to an empty string to avoid JSON unmarshalling surprises.

### What warrants a second pair of eyes

- Whether `LocatorText` defaults should eventually normalize/trim by default in the JS API, even though the service options are explicit.
- Whether `LocatorComputedStyle` should avoid the small `configStyleSpec(...)` adapter helper and instead have `EvaluateStyle` accept a service-native spec later.
- Whether missing selectors should be configurable as strict errors for some future extraction modes.

### What should be done in the future

Phase 5 should expose these service functions through a Go-backed `LocatorHandle` Proxy returned by `page.locator(selector)`.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/service/dom.go
```

Then review tests:

```text
internal/cssvisualdiff/service/dom_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/service -run 'TestLocatorDOMPrimitives' -count=1
go test ./internal/cssvisualdiff/service -count=1
go test ./... -count=1
```

### Technical details

The service functions are intentionally Goja-free and operate on `*driver.Page`. The future JS adapter should only decode inputs, call these functions through the per-page serialization guard, and lower results into JavaScript values.

## Step 11: Complete Phase 5 `page.locator()` JavaScript API

I completed Phase 5 by exposing the new service DOM primitives through a Go-backed Proxy locator handle. `page.locator(selector)` is synchronous and returns a page-bound handle; methods that touch Chromium return Promises and run through the existing per-page serialization guard from Phase 3.

This is the first visible lower-level JS API milestone. It lets scripts query element status, existence, visibility, text, bounds, computed styles, and attributes without writing inspect artifacts or using YAML-style probe objects.

### Prompt Context

**User prompt (verbatim):** "phase 5"

**Assistant interpretation:** Implement Phase 5 from the ticket: expose `page.locator(selector)` and async locator methods as Go-backed Proxy handles, validate behavior, update docs/diary/tasks, and commit.

**Inferred user intent:** Make the first lower-level JS API available on top of the Phase 4 DOM service primitives.

### What I did

- Added `internal/cssvisualdiff/jsapi/locator.go`.
- Added `locatorHandle` with references to the owning `pageState` and selector.
- Added `wrapLocator(...)` using the Phase 2 `newProxyValue(...)` helper.
- Added synchronous `page.locator(selector)` in `wrapPage(...)`.
- Implemented Promise-returning locator methods:
  - `locator.status()`
  - `locator.exists()`
  - `locator.visible()`
  - `locator.text(options)`
  - `locator.bounds()`
  - `locator.computedStyle(props)`
  - `locator.attributes(names)`
- Routed all locator browser work through `pageState.runExclusive(...)`.
- Added wrong-parent Proxy guidance for common probe methods accidentally called on locators:
  - `.selector(...)`
  - `.styles(...)`
  - `.required(...)`
  - `.build()`
- Added repository-scanned JS verb tests:
  - `TestCVDModuleExposesLocatorMethods`
  - `TestCVDModuleLocatorWrongParentError`
- Marked Phase 5 tasks complete and set the active phase to Phase 6.

### Why

The lower-level API should let JavaScript inspect the current page directly without forcing users through high-level artifact-writing `inspect` calls. A Go-backed Proxy handle gives us controlled method access, useful wrong-parent errors, and future typed unwrapping for strict APIs such as `cvd.extract(locator, extractors)`.

### What worked

The Phase 4 service functions mapped cleanly into locator methods. The Phase 3 page operation lock also made it straightforward to safely run locator methods under `Promise.all(...)`.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModule(ExposesLocatorMethods|LocatorWrongParentError)' -count=1

go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

No test failure occurred during Phase 5 after implementation. The main thing to watch is that locator Proxy errors are currently `TypeError`s, not `CvdError` subclasses.

### What I learned

The Proxy infrastructure is now doing useful public API work. The wrong-parent error for `page.locator("#cta").styles(["color"])` produces an actionable message that points users to `.computedStyle(...)` on locators and explains that `.styles(...)` belongs to future probe builders.

### What was tricky to build

The tricky part was argument handling for methods like `computedStyle(props)` and `attributes(names)`. These methods should reject non-array arguments immediately and clearly because future LLM-generated code may pass strings or object literals. I added a small `stringListArg(...)` helper that validates array inputs before starting async browser work.

Another subtle point is that `page.locator(selector)` itself is synchronous and does not validate selector syntax immediately. Syntax errors surface when a method queries the DOM. This matches the intended handle semantics: constructing a locator is cheap and page-bound, while evaluating it is async.

### What warrants a second pair of eyes

- Whether `locator.text()` should default to normalized/trimmed text in the JS API. It currently follows explicit `TextOptions`, so callers use `{ normalizeWhitespace: true, trim: true }`.
- Whether locator handles should include a `.toString()` or `.selector` read-only property in the public API later.
- Whether Proxy registries should become module-scoped so future strict APIs can unwrap locator handles created by `page.locator(...)`. The current locator uses `newProxyValue` with a local registry because no strict cross-call unwrapping exists yet.

### What should be done in the future

Phase 6 should add Go-backed target, probe, viewport, and extractor builders. Once extractor handles exist, Phase 7 can implement strict `cvd.extract(locator, extractors)` and should revisit the locator Proxy registry/unwrapping story.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/jsapi/locator.go
internal/cssvisualdiff/jsapi/module.go
```

Then review tests:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModule(ExposesLocatorMethods|LocatorWrongParentError)' -count=1
go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1
go test ./... -count=1
```

### Technical details

Example supported JS:

```js
const cta = page.locator("#cta")
const [status, text, styles, attrs] = await Promise.all([
  cta.status(),
  cta.text({ normalizeWhitespace: true, trim: true }),
  cta.computedStyle(["color", "display"]),
  cta.attributes(["id", "class"]),
])
```

All methods that touch Chromium run through:

```go
l.page.runExclusive(func() (any, error) { ... })
```

## Step 12: Complete Phase 6 Go-backed target/probe/extractor builders

I completed Phase 6 by adding the first Go-backed fluent builders for targets, probes, viewports, and extractors. These builders are synchronous authoring objects backed by Go structs and wrapped in Goja Proxies, matching the design decision that lower-level DSL objects should provide controlled methods and actionable errors rather than behaving like arbitrary raw JavaScript objects.

This phase still does not add strict extraction or snapshot execution. It creates the values that Phase 7 and Phase 8 will consume. `.build()` returns plain serializable values for debugging, tests, and YAML/config interop, while the Proxy-backed handles remain the authoring-time API.

### Prompt Context

**User prompt (verbatim):** "phase 6"

**Assistant interpretation:** Implement Phase 6 from the ticket: add Go-backed target/probe/extractor builders, tests, validation, diary/tasks/changelog updates, and commit.

**Inferred user intent:** Add the JS-native builder layer that will eventually replace many YAML object literals and feed strict extract/snapshot APIs.

### What I did

- Added `internal/cssvisualdiff/jsapi/builder_helpers.go` with shared validation helpers.
- Added `internal/cssvisualdiff/jsapi/target.go`:
  - `cvd.target(name)`
  - `.url(...)`
  - `.waitMs(...)`
  - `.viewport(width, height)` / `.viewport({ width, height })`
  - `.root(...)`
  - `.prepare(...)`
  - `.build()`
  - `cvd.viewport(width, height)` plus `desktop`, `tablet`, and `mobile` helpers.
- Added `internal/cssvisualdiff/jsapi/probe.go`:
  - `cvd.probe(name)`
  - `.selector(...)`
  - `.required(...)`
  - `.source(...)`
  - `.text()`
  - `.bounds()`
  - `.styles(...)`
  - `.attributes(...)`
  - `.build()`
- Added `internal/cssvisualdiff/jsapi/extractor.go`:
  - `cvd.extractors.exists()`
  - `cvd.extractors.visible()`
  - `cvd.extractors.text()`
  - `cvd.extractors.bounds()`
  - `cvd.extractors.computedStyle(props)`
  - `cvd.extractors.attributes(names)`
- Updated module registration to install target, probe, and extractor APIs.
- Added `internal/cssvisualdiff/jsapi/builders_test.go` for builder chaining, validation errors, and wrong-parent errors.
- Added `TestCVDModuleExposesTargetProbeAndExtractorBuilders` to the repository-scanned JS verb tests.
- Marked Phase 6 tasks complete and set the active phase to Phase 7.

### Why

Targets, probes, and extractors are the JS-native authoring layer for the future strict APIs. They let scripts express page targets, reusable inspection recipes, and extraction plans without YAML and without unvalidated raw object literals.

### What worked

The Phase 2 Proxy infrastructure made it straightforward to add fluent chainable objects. The builders can validate immediately, return the receiver for chaining, and emit wrong-parent errors when users call methods that belong to another object type.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/jsapi -run 'Test(TargetProbeAndExtractorBuilders|BuilderValidationErrors|BuilderWrongParentErrors)' -count=1

go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleExposesTargetProbeAndExtractorBuilders' -count=1

go test ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

The first builder unit test failed because Goja exported arrays from Go-backed maps as `[]string`, while the test expected `[]any`:

```text
Error: Not equal:
expected: []interface {}([]interface {}{"color", "font-size"})
actual  : []string([]string{"color", "font-size"})
```

I fixed the test to assert `[]string` for Go-backed builder output.

The repository-scanned builder smoke initially failed when passing `cvd.viewport.mobile()` into `.viewport(...)`:

```text
promise rejected: TypeError: cvd.target.viewport: expected positive width and height, got width=0 height=0
```

The cause was that `vm.ExportTo` did not decode the lowerCamel map returned by `lowerViewport(...)` into `config.Viewport` as expected. I changed `viewportFromCall(...)` to route object decoding through the JSON-based `decodeInto[config.Viewport](args[0].Export())`, matching the rest of the JS adapter behavior.

### What I learned

Go-backed builder `.build()` results may export Go slices as typed slices, not generic JavaScript `[]any`, in Go unit tests. Repository-scanned JS integration tests remain important because they exercise the real JavaScript-facing shape.

The existing JSON codec helper is a safer path for lowerCamel object decoding than direct `ExportTo(...)` when target structs do not have JSON tags.

### What was tricky to build

The main design tension is that Phase 6 needs extractor builders before the service-layer `ExtractorSpec` exists. I kept extractor handles lightweight and JSAPI-local for now, with `.build()` producing plain serializable values. Phase 7 should formalize the extraction service structs and update extractor unwrapping as needed.

Another tricky point is registry lifetime. The builders use `newProxyValue(...)` with an internal registry just like locators. That is enough for fluent method trapping and `.build()` now. Phase 7 strict unwrapping should introduce module-level state/registry if `cvd.extract(...)` needs to unwrap builder/locator handles across API calls.

### What warrants a second pair of eyes

- Whether `cvd.viewport.desktop/tablet/mobile` should be considered stable public names or examples only.
- Whether `.required()` should default to `true` when called with no argument. It currently does.
- Whether extractor kinds should use `computedStyle` or a shorter `style`/`styles` name in the eventual service schema.
- Whether Phase 7 should convert local extractor handles to service-native `ExtractorSpec` values instead of retaining JSAPI-local maps.

### What should be done in the future

Phase 7 should define service extraction specs/results, introduce strict unwrapping of locator and extractor Proxy handles, and implement `cvd.extract(locator, extractors)`.

### Code review instructions

Start with:

```text
internal/cssvisualdiff/jsapi/target.go
internal/cssvisualdiff/jsapi/probe.go
internal/cssvisualdiff/jsapi/extractor.go
internal/cssvisualdiff/jsapi/builder_helpers.go
```

Then review registration and tests:

```text
internal/cssvisualdiff/jsapi/module.go
internal/cssvisualdiff/jsapi/builders_test.go
internal/cssvisualdiff/verbcli/command_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/jsapi -run 'Test(TargetProbeAndExtractorBuilders|BuilderValidationErrors|BuilderWrongParentErrors)' -count=1
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleExposesTargetProbeAndExtractorBuilders' -count=1
go test ./... -count=1
```

### Technical details

Example supported JS:

```js
const target = cvd.target("booking")
  .url("http://example.test/booking")
  .viewport(cvd.viewport.mobile())
  .waitMs(25)
  .root("#app")

const probe = cvd.probe("cta")
  .selector("#cta")
  .required()
  .text()
  .bounds()
  .styles(["color"])
  .attributes(["id"])

const extractor = cvd.extractors.attributes(["id", "class"])
```

## Step 13: Complete Phase 7 strict `cvd.extract(locator, extractors)`

I completed Phase 7 by adding service-level extraction specs/results and exposing strict `cvd.extract(locator, extractors)` in the JavaScript API. This is the first strict API that requires Go-backed Proxy values across API calls: the locator must come from `page.locator(...)`, and extractors must come from `cvd.extractors.*(...)`.

This phase turns the lower-level API into a useful composition surface. Users can now query a page-bound element and choose exactly which facts to extract without creating inspect artifacts, while raw object inputs are rejected with explicit type guidance.

### Prompt Context

**User prompt (verbatim):** "phase 7"

**Assistant interpretation:** Implement Phase 7 from the ticket: define service extraction specs/results, implement extraction from locator plus extractors, expose strict `cvd.extract(...)`, validate raw-object rejection, update docs/diary/tasks, and commit.

**Inferred user intent:** Make the new locator and extractor builders work together as a strict lower-level extraction API.

### What I did

- Added `internal/cssvisualdiff/service/extract.go`.
- Defined:
  - `ExtractorKind`
  - `ExtractorSpec`
  - `ElementSnapshot`
- Implemented `service.ExtractElement(...)` over a `LocatorSpec` and list of `ExtractorSpec` values.
- Added `internal/cssvisualdiff/service/extract_test.go` covering:
  - multiple extractors on one element,
  - missing selector behavior,
  - invalid selector error behavior.
- Updated `internal/cssvisualdiff/jsapi/proxy.go` so `newProxyValue(..., nil, ...)` uses a package-level default registry instead of creating isolated one-off registries.
- Updated `extractorHandle` to convert to `service.ExtractorSpec`.
- Added `internal/cssvisualdiff/jsapi/extract.go` with:
  - `installExtractAPI(...)`
  - `unwrapExtractorList(...)`
  - `lowerElementSnapshot(...)`
- Registered `cvd.extract(...)` from the native module.
- Added repository-scanned JS verb tests:
  - `TestCVDModuleExtractsFromLocatorWithExtractorHandles`
  - `TestCVDModuleExtractRejectsRawObjects`
- Marked Phase 7 tasks complete and set the active phase to Phase 8.

### Why

The design requires lower-level APIs to be strict and Go-backed, not raw-object permissive. `cvd.extract(...)` is the first place where that matters because it receives values created elsewhere: locators from pages and extractor handles from `cvd.extractors`. The Proxy registry now enables that strict unwrapping.

### What worked

The Phase 4 DOM service functions made `ExtractElement(...)` straightforward. The Phase 2 Proxy registry could be reused with one change: proxies created with a nil registry now bind to a shared default registry, so strict APIs can unwrap them later.

Validation commands that passed:

```bash
go test ./internal/cssvisualdiff/service -run 'TestExtractElement' -count=1

go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleExtract' -count=1

go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/dsl ./internal/cssvisualdiff/verbcli ./cmd/css-visual-diff -count=1

go test ./... -count=1
```

### What didn't work

No failing test occurred after the first implementation pass for Phase 7. The main design issue discovered during implementation was that previous locator/extractor proxies were created with isolated registries when callers passed `nil`. That would make cross-call strict unwrapping impossible. I fixed this by introducing a package-level default registry.

### What I learned

Strict APIs need a shared identity boundary. A local per-proxy registry is fine for method trapping, but it is not enough when a top-level function such as `cvd.extract(...)` must prove that a value came from `page.locator(...)` or `cvd.extractors.*(...)`.

The current `ElementSnapshot` intentionally omits absent fields from the JS result unless an extractor was requested and produced data. Boolean fields use pointers internally so `exists: false` and omitted `exists` can be distinguished.

### What was tricky to build

The tricky part was preserving the Go-backed Proxy requirement without over-engineering module state. I used a package-level default registry as the smallest step that makes strict unwrapping work. A future cleanup may replace it with per-runtime module state, but this phase keeps the implementation simple and covered by integration tests.

Another subtle point is invalid selectors. `LocatorStatus(...)` returns a status with an error string rather than a Go error. `ExtractElement(...)` promotes that status error into a Go error for strict `exists`/`visible` extraction so malformed selectors reject the Promise.

### What warrants a second pair of eyes

- Whether the package-level default Proxy registry is acceptable or should become per-runtime state before this API is considered stable.
- Whether `ElementSnapshot` should always include requested fields even when empty, especially for `text: ""` and missing selectors.
- Whether extractor default text options should normalize/trim text; currently extractor text defaults to normalized/trimmed through `extractorHandle.toSpec()`.

### What should be done in the future

Phase 8 should implement strict `cvd.snapshot(page, probes, options)` and likely reuse `service.ExtractElement(...)` for each probe. If per-runtime state is introduced, Phase 8 may also be a good time to clean up the registry shape.

### Code review instructions

Start with service extraction:

```text
internal/cssvisualdiff/service/extract.go
internal/cssvisualdiff/service/extract_test.go
```

Then review JS strict API plumbing:

```text
internal/cssvisualdiff/jsapi/proxy.go
internal/cssvisualdiff/jsapi/extractor.go
internal/cssvisualdiff/jsapi/extract.go
internal/cssvisualdiff/jsapi/module.go
```

Then review integration tests:

```text
internal/cssvisualdiff/verbcli/command_test.go
```

Validation commands:

```bash
go test ./internal/cssvisualdiff/service -run 'TestExtractElement' -count=1
go test ./internal/cssvisualdiff/verbcli -run 'TestCVDModuleExtract' -count=1
go test ./... -count=1
```

### Technical details

Example supported JS:

```js
const snapshot = await cvd.extract(page.locator("#cta"), [
  cvd.extractors.exists(),
  cvd.extractors.visible(),
  cvd.extractors.text(),
  cvd.extractors.bounds(),
  cvd.extractors.computedStyle(["color"]),
  cvd.extractors.attributes(["id", "class"]),
])
```

Raw object locator inputs are rejected:

```js
await cvd.extract({ selector: "#cta" }, [cvd.extractors.text()])
```

with an error containing:

```text
css-visual-diff.extract: expected cvd.locator
```
