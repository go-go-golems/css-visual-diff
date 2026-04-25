---
Title: Real-site compare-region validation
Ticket: CSSVD-JSAPI-PIXEL-WORKFLOWS
Status: active
Topics:
  - frontend
  - visual-regression
  - browser-automation
  - tooling
DocType: reference
Intent: validation
Owners: []
Summary: Live HTTPS validation runs for cvd.compare.region through the built-in script compare region verb.
LastUpdated: 2026-04-25T13:45:00-04:00
---

# Real-site validation: `cvd.compare.region` via built-in compare verb

Date: 2026-04-25

This directory records a real external-site validation run for the new public JavaScript comparison stack. The validation uses the built-in:

```bash
css-visual-diff verbs script compare region
```

That built-in now dogfoods the public `require("css-visual-diff")` API and calls `cvd.compare.region(...)` internally.

## Run 1 — equivalent real pages

Compared `https://example.com/` with `https://example.org/` using selector `body` on both sides.

Output directory:

```text
validation/real-site-example
```

Command:

```bash
go run ./cmd/css-visual-diff verbs script compare region \
  --leftUrl https://example.com/ \
  --rightUrl https://example.org/ \
  --leftSelector body \
  --rightSelector body \
  --width 1280 \
  --height 720 \
  --leftWaitMs 1000 \
  --rightWaitMs 1000 \
  --outDir ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/validation/real-site-example \
  --threshold 30 \
  --writeJson \
  --writeMarkdown \
  --writePngs \
  --output json
```

Observed result:

- Schema: `cssvd.selectionComparison.v1`
- Left URL: `https://example.com/`
- Right URL: `https://example.org/`
- Selector: `body` vs `body`
- Changed percent: `0`
- Changed pixels: `0`
- Total pixels: `100608`
- Bounds changed: `false`
- Text changed: `false`

This is a useful sanity check because `example.com` and `example.org` serve equivalent example-domain content in this browser context. The API successfully navigated to both real HTTPS sites, selected real DOM nodes, captured region screenshots, produced diff artifacts, wrote JSON/Markdown, and reported zero visual difference.

## Run 2 — visibly different real pages

Compared `https://example.com/` body against the IANA reserved-domains page main content.

Output directory:

```text
validation/real-site-different
```

Command:

```bash
go run ./cmd/css-visual-diff verbs script compare region \
  --leftUrl https://example.com/ \
  --rightUrl https://www.iana.org/domains/reserved \
  --leftSelector body \
  --rightSelector main \
  --width 1280 \
  --height 720 \
  --leftWaitMs 1500 \
  --rightWaitMs 1500 \
  --outDir ttmp/2026/04/25/CSSVD-JSAPI-PIXEL-WORKFLOWS--design-js-api-additions-for-pixel-comparison-and-workflow-orchestration/validation/real-site-different \
  --threshold 30 \
  --writeJson \
  --writeMarkdown \
  --writePngs \
  --output json
```

Observed result:

- Schema: `cssvd.selectionComparison.v1`
- Left URL: `https://example.com/`
- Right URL: `https://www.iana.org/domains/reserved`
- Selector: `body` vs `main`
- Changed percent: `10.295429500970773`
- Changed pixels: `98100`
- Total pixels: `952850`
- Normalized size: `850x1121`
- Bounds changed: `true`
- Text changed: `true`
- Style changes included:
  - `background-color`
  - `color`
  - `font-family`

This confirms the comparison path detects real visual/layout/content differences and emits useful structured evidence.

## Artifacts

Each run produced:

```text
left_region.png
right_region.png
diff_only.png
diff_comparison.png
compare.json
compare.md
command-output.json
```

The image artifacts are intentionally retained in this validation folder so future reviewers can inspect the actual rendered evidence.

## Conclusion

The real-site validation passed. The new JS API path works outside synthetic local fixtures:

```text
built-in verb -> require("css-visual-diff") -> cvd.compare.region -> collect -> screenshot -> pixel diff -> SelectionComparison JSON/Markdown/PNG artifacts
```

Caveat: this is still a small validation set. It does not cover login flows, cookie banners, lazy-loaded app shells, cross-origin iframe content, or sites with heavy animation. It does prove that the API works against live HTTPS pages and produces the expected artifact bundle.
