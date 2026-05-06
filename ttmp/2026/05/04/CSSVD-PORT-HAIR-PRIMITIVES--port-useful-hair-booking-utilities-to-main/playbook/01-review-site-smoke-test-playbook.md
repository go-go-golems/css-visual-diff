---
Title: Review Site Smoke Test Playbook
Ticket: CSSVD-PORT-HAIR-PRIMITIVES
Status: active
Topics:
    - css-visual-diff
    - javascript-api
    - merge-planning
    - visual-diff
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../../../tmp/cssvd-review-site-smoke/summary.json
      Note: Generated smoke-test review-site summary
    - Path: examples/pages/review-site-smoke-left.html
      Note: Left fixture page for deterministic review-site smoke data
    - Path: examples/pages/review-site-smoke-right.html
      Note: Right fixture page with intentional visual/CSS differences
    - Path: examples/specs/review-site-smoke.yaml
      Note: Review-sweep spec that drives the smoke test
    - Path: internal/cssvisualdiff/review/embed/public/index.html
      Note: Embedded review-site frontend rebuilt after path fix
    - Path: web/review-site/src/components/ReviewCard.tsx
      Note: Review card now routes absolute artifact paths through served artifact URLs
ExternalSources: []
Summary: How to build css-visual-diff, generate deterministic comparison data, and serve the review website.
LastUpdated: 2026-05-06T13:37:54.800157647-04:00
WhatFor: Provide a repeatable local smoke test for the interactive comparison/review website.
WhenToUse: After changing review-site data generation, the embedded React site, artifact path handling, or compare-region output.
---


# Review Site Smoke Test Playbook

## Purpose

Build a local `css-visual-diff` binary, serve deterministic fixture pages, generate review-site comparison data, and launch the interactive comparison website so a human can inspect side-by-side, overlay, slider, diff-only, CSS diffs, notes, pins, and LLM export.

## Environment Assumptions

- Run from the repository root.
- Go toolchain is available.
- Chromium dependencies used by `css-visual-diff` are available on the machine.
- Ports `18767` and `18098` are free.
- The embedded review-site frontend exists under `internal/cssvisualdiff/review/embed/public`. If it does not, run `make build-web` first.

## Test Fixtures

- Left page: `examples/pages/review-site-smoke-left.html`
- Right page: `examples/pages/review-site-smoke-right.html`
- Spec: `examples/specs/review-site-smoke.yaml`
- Generated data: `/tmp/cssvd-review-site-smoke`
- Local binary: `/tmp/css-visual-diff-review-smoke`

## Commands

Terminal 1 — serve repo files for the fixture URLs:

```bash
python3 -m http.server 18767
```

Terminal 2 — build the CLI and generate review data:

```bash
GOWORK=off go build -o /tmp/css-visual-diff-review-smoke ./cmd/css-visual-diff

rm -rf /tmp/cssvd-review-site-smoke
/tmp/css-visual-diff-review-smoke verbs \
  --repository examples/verbs \
  examples review-sweep from-spec \
  --specFile examples/specs/review-site-smoke.yaml \
  --outDir /tmp/cssvd-review-site-smoke \
  --output json
```

Terminal 2 — serve the review website:

```bash
/tmp/css-visual-diff-review-smoke serve \
  --data-dir /tmp/cssvd-review-site-smoke \
  --port 18098 \
  --open
```

Open manually if `--open` is unavailable:

```text
http://127.0.0.1:18098
```

## Exit Criteria

- `/tmp/cssvd-review-site-smoke/summary.json` exists.
- Per-section artifacts exist under `/tmp/cssvd-review-site-smoke/smoke/artifacts/{app,hero,card,cta}/`.
- The review website loads cards for `smoke/app`, `smoke/hero`, `smoke/card`, and `smoke/cta`.
- Expanding a card shows images in side-by-side, overlay, slider, and diff-only modes.
- CSS diff and metadata panels load without console errors.
- Setting a status, adding a note, and dropping a pin persist across a reload.
- `Send to LLM` opens a markdown/YAML export modal.

## Notes

The fixture pages intentionally differ in spacing, color, radius, typography, shadows, and button text. The run should therefore produce visible diffs and non-empty CSS changes. This is a smoke test for the review website, not a pass/fail regression threshold.
