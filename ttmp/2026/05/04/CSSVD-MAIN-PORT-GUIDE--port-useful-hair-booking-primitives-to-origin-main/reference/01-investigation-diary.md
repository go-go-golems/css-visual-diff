---
Title: Investigation diary
Ticket: CSSVD-MAIN-PORT-GUIDE
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - merge-planning
    - pyxis
    - visual-diff
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/04/CSSVD-MAIN-PORT-GUIDE--port-useful-hair-booking-primitives-to-origin-main/design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md
      Note: Primary handoff guide created from this investigation.
ExternalSources: []
Summary: Chronological record of the second-pass port-guide investigation and delivery.
LastUpdated: 2026-05-04T17:30:00-04:00
WhatFor: Record how the port guidance was derived and what evidence shaped the final recommendation.
WhenToUse: When continuing the ticket, reviewing the handoff guide, or checking why cvd.compare.region was excluded from the port list.
---

# Diary

## Goal

This diary records the creation of the `CSSVD-MAIN-PORT-GUIDE` ticket and the detailed handoff document for colleagues working on `origin/main`. The main decision captured here is that only clean primitives should be ported from `bookmark/2026-05-01/hair-booking`; bespoke helper APIs such as `cvd.compare.region` should stay out of main unless maintainers intentionally reintroduce a convenience layer.

## Step 1: Create the ticket and write the port guide

The user asked for a new document that can be handed to colleagues on `origin/main`. I created a docmgr ticket, added a design document and diary, then wrote a detailed intern-facing implementation guide. The guide explains the mainline JavaScript API, the hair-booking branch additions, the revised port decision, and the implementation steps for accepted primitives.

The key correction from earlier analysis is that Pyxis using `cvd.compare.region` is not by itself a reason to port that API. If `origin/main` deliberately moved toward explicit primitives, Pyxis should adapt to `require("diff").compareRegion`, `cvd.extract`, `cvd.snapshot`, `cvd.diff`, and `cvd.report` rather than forcing `cvd.compare.region` back into main.

### Prompt Context

**User prompt (verbatim):** "write a new document that i can give our colleagues working on origin/main that gives a detailed breadown of the things to port and how.
Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a ticket-backed handoff document for origin/main maintainers and interns, with technical port guidance, architecture context, pseudocode, diagrams, file references, and reMarkable upload.

**Inferred user intent:** The user wants the corrected narrower port strategy documented in a form colleagues can use to implement and review the work without re-reading the whole chat.

**Commit (code):** N/A — documentation-only work in the ticket workspace.

### What I did

- Ran `docmgr status --summary-only` to confirm the docmgr root and existing ticket state.
- Created ticket `CSSVD-MAIN-PORT-GUIDE` with `docmgr ticket create-ticket`.
- Added a design document and this diary using `docmgr doc add`.
- Re-read the ticket-research and diary skill references for required structure.
- Wrote the primary design document at:
  - `ttmp/2026/05/04/CSSVD-MAIN-PORT-GUIDE--port-useful-hair-booking-primitives-to-origin-main/design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md`
- Wrote this diary at:
  - `ttmp/2026/05/04/CSSVD-MAIN-PORT-GUIDE--port-useful-hair-booking-primitives-to-origin-main/reference/01-investigation-diary.md`

### Why

The previous second-pass document over-emphasized saving `cvd.compare.region` because Pyxis userland currently uses it. The user's follow-up clarified that `origin/main` may have intentionally removed that helper to keep the JS API clean and fluent. The new guide reflects that corrected design center.

### What worked

- `docmgr ticket create-ticket` created the expected ticket workspace.
- `docmgr doc add` created the expected design and diary documents with frontmatter.
- The existing evidence from the repository was sufficient to write a file-backed guide:
  - `origin/main` docs show no `cvd.compare.region` export.
  - `origin/main` has `require("diff").compareRegion` as the high-level URL/selector compare operation.
  - hair-booking contains `locator.waitFor` and `service/pixel.go`, which are cleanly separable from the bespoke comparison APIs.

### What didn't work

- The current repository worktree remains in a half-broken merge state from the earlier branch investigation. I did not attempt to repair or continue that merge because the guide explicitly recommends starting from a clean `origin/main` branch.

### What I learned

- The strongest candidates to port are small primitives, not the most visible Pyxis-facing API.
- `cvd.compare.region` is best understood as a convenience molecule. It may be useful, but it does not match the mainline API philosophy unless maintainers choose to add a convenience layer.
- `locator.waitFor` is the clearest port candidate because it lives at the right abstraction level: a method on a live locator.

### What was tricky to build

The tricky part was separating "currently useful to Pyxis" from "belongs in origin/main." Pyxis was written against the branch API, so at first glance `cvd.compare.region` looks mandatory. Once compared to `origin/main`'s documented API, the better conclusion is that Pyxis should migrate to mainline primitives and `require("diff").compareRegion` rather than shaping mainline around a branch-only convenience API.

### What warrants a second pair of eyes

- Confirm with the maintainers who removed `cvd.compare.region` that the omission was intentional and not an accidental casualty of the remove-YAML/refactor work.
- Check whether `require("diff").compareRegion` returns all stable artifact paths Pyxis needs; if not, port only the path reliability behavior.
- Review whether `catalog.addResult` is sufficient for comparison artifacts or whether a generic catalog artifact record is needed.

### What should be done in the future

- Implement the guide in small PRs: `locator.waitFor`, `service/pixel.go`, artifact path audit, then Pyxis adaptation.
- Fix Pyxis accepted-differences merging separately in `userland/lib/registry.js`.

### Code review instructions

Start review with:

1. `internal/cssvisualdiff/jsapi/module.go` to confirm no bespoke compare API is registered.
2. `internal/cssvisualdiff/jsapi/locator.go` for `waitFor` wiring.
3. `internal/cssvisualdiff/service/dom.go` for wait semantics.
4. `internal/cssvisualdiff/service/pixel.go` for internal pixel extraction.
5. `internal/cssvisualdiff/modes/compare.go` to ensure output compatibility is preserved.

Validation commands for future implementation:

```bash
go test ./internal/cssvisualdiff/service ./internal/cssvisualdiff/modes ./internal/cssvisualdiff/jsapi ./internal/cssvisualdiff/verbcli
```

Pyxis validation after userland migration:

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis/prototype-design/visual-diff/userland
./scripts/smoke-list-targets.sh
./scripts/smoke-inspect-section-archive.sh
./scripts/smoke-compare-section-archive.sh
./scripts/smoke-compare-page-archive.sh
./scripts/smoke-compare-spec-archive.sh
./scripts/smoke-snapshot-section-archive.sh
./scripts/smoke-diff-snapshots-archive.sh
```

### Technical details

The final recommendation in the design doc is:

```text
Port:
  - locator.waitFor(...)
  - service/pixel.go and tests
  - artifact path reliability if main has gaps
  - small internal semantic diff helpers only if needed

Do not port:
  - cvd.compare.region
  - cvd.compare.selections
  - locator.collect / cvd.collect.selection
  - service/collection.go wholesale
  - catalog.record(comparison) as-is
  - old YAML config/native modes/runner
```
