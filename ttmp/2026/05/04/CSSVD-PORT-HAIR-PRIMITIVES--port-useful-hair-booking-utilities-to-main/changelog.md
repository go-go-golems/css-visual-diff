# Changelog

## 2026-05-04

- Initial workspace created


## 2026-05-04

Created ticket, copied the previous branch port guide into design-doc, studied it, and added a phased implementation task plan.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/05/04/CSSVD-PORT-HAIR-PRIMITIVES--port-useful-hair-booking-utilities-to-main/design-doc/01-origin-main-port-guide-for-hair-booking-branch-primitives.md — Copied guide used as source for the task plan
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/05/04/CSSVD-PORT-HAIR-PRIMITIVES--port-useful-hair-booking-utilities-to-main/tasks.md — Phased task list generated from the guide


## 2026-05-04

Updated ticket overview, added implementation diary, related core files, and added missing vocabulary topics used by this ticket.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/05/04/CSSVD-PORT-HAIR-PRIMITIVES--port-useful-hair-booking-utilities-to-main/index.md — Ticket overview and related-file links
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/05/04/CSSVD-PORT-HAIR-PRIMITIVES--port-useful-hair-booking-utilities-to-main/reference/01-implementation-diary.md — Diary records initial planning work
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/vocabulary.yaml — Vocabulary entries for merge-planning


## 2026-05-04

Completed Phase 1: ported locator.waitFor into the service and JS API, added service and verbcli coverage, documented the API, updated low-level inspect example, and validated targeted packages.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/examples/verbs/low-level-inspect.js — Example now waits for the selector before extraction
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/doc/topics/javascript-api.md — Documents locator.waitFor options and return value
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/locator.go — Exposes locator.waitFor and lowers its result
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/dom.go — WaitForSelectorOptions
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/dom_test.go — Service tests for delayed
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command_test.go — Runtime JS API smoke coverage for locator.waitFor


## 2026-05-06

Added deterministic review-site smoke fixtures and playbook, built the CLI/review frontend, generated /tmp/cssvd-review-site-smoke, launched the review website on port 18098, and fixed ReviewCard artifact URLs so absolute artifact paths load through the serve endpoint.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/examples/pages/review-site-smoke-left.html — Baseline fixture page
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/examples/pages/review-site-smoke-right.html — Candidate fixture page with intentional diffs
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/examples/specs/review-site-smoke.yaml — Review-sweep smoke spec
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/review/embed/public/assets/index-DrF3HGaf.js — Rebuilt embedded review-site bundle
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/05/04/CSSVD-PORT-HAIR-PRIMITIVES--port-useful-hair-booking-utilities-to-main/playbook/01-review-site-smoke-test-playbook.md — Repeatable review-site smoke playbook
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/web/review-site/src/components/ReviewCard.tsx — Artifact path URL conversion fix


## 2026-05-06

Added a dedicated Glazed help page for the site-comparison review workflow, covering the YAML spec, review-sweep JS verb, generated artifacts, serving contract, screenshots, CSS diffs, smoke setup, troubleshooting, and related help pages.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/doc/tutorials/site-comparison-workflow.md — New site-comparison workflow help entry

