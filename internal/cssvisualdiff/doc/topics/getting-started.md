---
Title: "Getting Started with css-visual-diff"
Slug: getting-started
Short: Start here to discover the main workflows, example verbs, review site, overlay screenshots, and JavaScript API docs.
Topics:
- getting-started
- examples
- javascript
- review-site
- overlays
Commands:
- compare
- serve
- verbs
Flags:
- repository
- data-dir
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use this page as the map for the rest of the documentation. `css-visual-diff` has three common entry points:

1. compare one known region directly,
2. run an example JavaScript verb that generates a full review-site data directory,
3. write your own project-local JavaScript verbs once the workflow is clear.

## Discover the docs and examples

List all embedded help pages:

```bash
css-visual-diff help --list
```

Read this starter page again:

```bash
css-visual-diff help getting-started
```

Inspect the repository-scanned example commands:

```bash
css-visual-diff verbs --repository examples/verbs examples --help
css-visual-diff verbs --repository examples/verbs examples review-sweep --help
css-visual-diff verbs --repository examples/verbs examples overlay --help
```

The source examples live in:

```text
examples/verbs/
├── README.md                 # runnable example commands
├── low-level-inspect.js       # locator/extractor/snapshot example
├── overlay-examples.js        # annotated PNG and gallery exports
└── review-sweep.js            # YAML spec → review-site data directory
```

## Fastest end-to-end smoke run

Start a static file server from the repository root:

```bash
python3 -m http.server 18767
```

Generate comparison artifacts and a review-site manifest:

```bash
css-visual-diff verbs \
  --repository examples/verbs \
  examples review-sweep from-spec \
  --specFile examples/specs/review-site-smoke.yaml \
  --outDir /tmp/cssvd-review-site-smoke \
  --output json
```

Serve the interactive review site:

```bash
css-visual-diff serve \
  --data-dir /tmp/cssvd-review-site-smoke \
  --port 18098 \
  --open
```

This path demonstrates the intended larger workflow: JavaScript opens pages and writes evidence; `css-visual-diff serve` reviews completed evidence without rerunning Chromium.

## Direct one-region comparison

Use the built-in `compare` command when you already know the two URLs and selectors:

```bash
css-visual-diff compare \
  --url1 http://localhost:7070/prototype.html \
  --selector1 '[data-section="hero"]' \
  --url2 http://localhost:5173/ \
  --selector2 '[data-section="hero"]' \
  --viewport-w 1280 \
  --viewport-h 900 \
  --threshold 30 \
  --out /tmp/cssvd-hero-compare
```

This writes screenshots, `compare.json`, `compare.md`, and pixel diff images for one region.

## Annotated overlay screenshots

Overlay screenshots label page sections or components in context. Run the example fixture server first:

```bash
python3 -m http.server 18767
```

Then export one annotated PNG:

```bash
css-visual-diff verbs --repository examples/verbs \
  examples overlay annotated-png \
  http://127.0.0.1:18767/examples/pages/overlay-components.html \
  /tmp/cssvd-overlay \
  --contentAlphaPercent 10 \
  --output json
```

For API details, read:

```bash
css-visual-diff help javascript-api
```

and search for `cvd.overlaySpec`, `cvd.overlayTarget`, and `page.overlay`.

## What to read next

| Goal | Help page |
| --- | --- |
| Learn the JS browser API | `css-visual-diff help javascript-api` |
| Write project-local CLI workflows | `css-visual-diff help javascript-verbs` |
| Build pixel-perfect CSS feedback loops | `css-visual-diff help pixel-accuracy-scripting-guide` |
| Generate review-site data from a YAML spec | `css-visual-diff help site-comparison-workflow` |
| Use the interactive review website | `css-visual-diff help review-site` |
| Produce review-site data yourself | `css-visual-diff help review-site-data-spec` |
| Understand the review-sweep example | `css-visual-diff help js-verb-review-sweep` |

## Mental model

- Go owns stable browser automation, screenshots, CSS extraction, pixel diffing, overlay rendering, artifact writing, and serving the review UI.
- JavaScript verbs own project meaning: page lists, selectors, variants, waits, policies, accepted differences, and handoff formats.
- The review site reads completed artifacts. It does not run Chromium.

When in doubt, start with the smoke run above, inspect `examples/verbs/review-sweep.js`, then adapt the verb pattern to your project.
