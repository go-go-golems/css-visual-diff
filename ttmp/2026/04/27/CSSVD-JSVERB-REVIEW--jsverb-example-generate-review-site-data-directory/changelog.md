# Changelog

## 2026-04-27

- Initial workspace created


## 2026-04-27

Created ticket and wrote 777-line analysis/implementation guide covering: current VM module availability (all modules already available, no Go changes needed), YAML spec format, verb design (from-spec + summary), implementation sketches with pseudocode, task breakdown, open questions.


## 2026-04-27

Implemented review-sweep JSVerb and fixed loader/module issue. Added verbs and example spec across commits fd50847, eff6f31, 2620bb2, df8e67f. Fixed dependency skew by upgrading go-go-goja to v0.4.14 and corrected fs.statSync/isDir handling in commit 73591a6. Wrote loader fix report and updated diary.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/examples/verbs/review-sweep.js — Main implementation and fixes
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/go.mod — go-go-goja dependency upgrade
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/27/CSSVD-JSVERB-REVIEW--jsverb-example-generate-review-site-data-directory/reference/02-loader-and-review-sweep-fix-report.md — Root-cause report


## 2026-04-27

Updated review-site-data-spec help topic to reference the new examples review-sweep verbs and example YAML spec.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/doc/topics/review-site-data-spec.md — Documents new review-sweep workflow


## 2026-04-27

Fixed review-site CSS diff crash when compare.json is in diff.compareRegion snake_case format. Added compareData normalization helpers and used them in StylesTab, MetaTab, and export generation (commit ab161b7).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/web/review-site/src/components/MetaTab.tsx — Meta tab uses normalized bounds/sources
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/web/review-site/src/components/StylesTab.tsx — CSS diff tab no longer assumes styles[] exists
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/web/review-site/src/utils/compareData.ts — Normalization helpers for catalog/inspect and compareRegion JSON variants
- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/web/review-site/src/utils/export.ts — Export uses normalized compare data

