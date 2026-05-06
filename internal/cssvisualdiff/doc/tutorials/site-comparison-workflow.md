---
Title: Site Comparison Review Workflow
Slug: site-comparison-workflow
Short: Generate comparison artifacts from a YAML site spec, serve them in the interactive review site, and inspect screenshots, CSS differences, notes, and export output.
Topics:
- review
- site-comparison
- visual-regression
- javascript-verbs
- screenshots
Commands:
- verbs
- serve
Flags:
- repository
- specFile
- outDir
- data-dir
- port
- open
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

A site comparison run has two responsibilities. First it captures evidence: browser screenshots, cropped regions, pixel diffs, computed CSS, bounds, attributes, and a root `summary.json`. Then it serves that evidence in the interactive review site so a human can compare screenshots, inspect CSS differences, leave notes, drop pins, and export a handoff document.

The important design choice is that comparison and review are separate steps. The JavaScript verb in `examples/verbs/review-sweep.js` opens pages and writes files. The built-in `css-visual-diff serve` command reads those files and displays them. This keeps browser automation, artifact generation, and human review independent. You can regenerate evidence without changing the review site, and you can reopen a completed run without re-running Chromium.

## What the Workflow Produces

A complete run is a directory with one manifest and one artifact directory per page section. The review site does not infer this structure; it expects it. The `review-sweep` example verb writes the structure for you.

```text
/tmp/cssvd-review-site-smoke/
├── summary.json
└── smoke/
    └── artifacts/
        ├── app/
        │   ├── compare.json
        │   ├── compare.md
        │   ├── left_region.png
        │   ├── right_region.png
        │   ├── diff_only.png
        │   ├── diff_comparison.png
        │   ├── url1_full.png
        │   └── url2_full.png
        ├── hero/
        ├── card/
        └── cta/
```

The `summary.json` file drives the card list. Each row says which page and section it represents, how many pixels changed, how the result was classified, and where the screenshots live. Each `compare.json` file contains the detailed evidence for one section: source URLs, selectors, bounds, pixel statistics, computed CSS differences, and attribute differences. The PNG files are the visual evidence used by the view modes.

The key files are:

| File | Required | Purpose |
| --- | --- | --- |
| `summary.json` | Yes | The root manifest loaded by `/api/manifest`; it lists every review card. |
| `<page>/artifacts/<section>/compare.json` | Yes | Detailed comparison data loaded when a card expands. |
| `left_region.png` | Yes | Cropped screenshot from the left/prototype side. |
| `right_region.png` | Yes | Cropped screenshot from the right/implementation side. |
| `diff_only.png` | Yes | Image containing only changed pixels. |
| `diff_comparison.png` | Optional | Triptych/combined comparison image; useful for manual artifact browsing. |
| `compare.md` | Optional | Markdown summary for humans outside the review site. |
| `url1_full.png`, `url2_full.png` | Optional | Full-page screenshots for debugging capture context. |

## The Site Spec

The site spec is a YAML file consumed by the example JavaScript verb. It describes what to compare, not how the review site should render. The spec names pages, declares left and right URLs, defines section selectors, chooses viewport dimensions, and lists CSS properties and attributes to collect.

A minimal spec looks like this:

```yaml
name: cssvd-review-site-smoke
variant: desktop

viewport:
  width: 1000
  height: 760

defaults:
  waitMs: 250
  threshold: 30

policy:
  bands:
    - name: accepted
      maxChangedPercent: 0.5
    - name: review
      maxChangedPercent: 10
    - name: tune-required
      maxChangedPercent: 30
    - name: major-mismatch
      maxChangedPercent: 100

computed:
  - font-size
  - font-weight
  - line-height
  - color
  - background-color
  - border-radius
  - padding-top
  - padding-right
  - padding-bottom
  - padding-left

attributes:
  - id
  - class
  - aria-label

pages:
  smoke:
    leftUrl: http://127.0.0.1:18767/examples/pages/review-site-smoke-left.html
    rightUrl: http://127.0.0.1:18767/examples/pages/review-site-smoke-right.html
    sections:
      app:
        selector: "#app"
      hero:
        selector: ".hero"
      cta:
        selector: ".cta"
```

Each section normally uses the same selector on both sides. If your prototype and implementation use different DOM shapes, use `leftSelector` and `rightSelector`:

```yaml
pages:
  pricing:
    leftUrl: http://localhost:7070/pricing.html
    rightUrl: http://localhost:6006/iframe.html?id=pricing--desktop
    sections:
      cards:
        leftSelector: "#pricing-cards"
        rightSelector: "[data-component='PricingCards']"
```

The spec fields have these responsibilities:

| Field | Meaning |
| --- | --- |
| `name` | Human name for the run. The example verb does not require it for paths, but it makes specs easier to identify. |
| `variant` | Label copied into summary rows, often `desktop`, `tablet`, or `mobile`. |
| `viewport.width`, `viewport.height` | Browser viewport used for every comparison unless the verb is extended. |
| `defaults.waitMs` | Post-navigation wait before capturing. Increase this for pages with client-side rendering, animations, or slow fonts. |
| `defaults.threshold` | Per-channel pixel threshold passed to `diff.compareRegion`. Higher values ignore small pixel differences. |
| `policy.bands` | Classification thresholds applied to `changedPercent`. These labels drive the colored review cards. |
| `computed` | CSS properties extracted into `compare.json` for the CSS diff panel. |
| `attributes` | HTML attributes extracted for attribute comparisons. |
| `pages.<page>.leftUrl` | URL for the baseline/prototype side. |
| `pages.<page>.rightUrl` | URL for the candidate/implementation side. |
| `sections.<section>.selector` | CSS selector used on both sides. |
| `sections.<section>.leftSelector` | Optional left-side selector override. |
| `sections.<section>.rightSelector` | Optional right-side selector override. |
| `sections.<section>.leftWaitMs`, `rightWaitMs` | Optional wait overrides for one side of one section. |

Selectors should target stable visual regions. A section that is too broad produces noisy screenshots and large pixel percentages. A section that is too narrow can miss the visual problem. For a first run, choose a small set of regions such as `header`, `main`, `.hero`, `.card`, and the primary CTA. Add more sections after the review site confirms the pipeline is working.

## How the Example JavaScript Verb Works

The example generator lives at `examples/verbs/review-sweep.js`. It is a repository-scanned JavaScript verb, not a built-in Go command. You make it available by passing `--repository examples/verbs` to `css-visual-diff verbs`.

The verb exposes two commands:

```bash
css-visual-diff verbs --repository examples/verbs examples review-sweep from-spec \
  --specFile examples/specs/review-site-smoke.yaml \
  --outDir /tmp/cssvd-review-site-smoke

css-visual-diff verbs --repository examples/verbs examples review-sweep summary \
  --specFile examples/specs/review-site-smoke.yaml \
  --outDir /tmp/cssvd-review-site-smoke
```

`from-spec` reads the YAML spec, runs comparisons, writes all artifacts, and creates `summary.json`. `summary` walks an existing output directory and rebuilds `summary.json` from `compare.json` files. The second command is useful when you changed policy bands, copied artifacts, or lost the root manifest.

At the center of `from-spec` is a loop over pages and sections. For each section it calls the native diff module:

```js
const diff = require("diff")

const result = diff.compareRegion({
  left: {
    url: pageSpec.leftUrl,
    selector: leftSelector,
    waitMs: leftWait,
  },
  right: {
    url: pageSpec.rightUrl,
    selector: rightSelector,
    waitMs: rightWait,
  },
  viewport: {
    width: vpWidth,
    height: vpHeight,
  },
  output: {
    outDir: artifactDir,
    threshold: threshold,
    writeJson: true,
    writeMarkdown: writeMd,
    writePngs: true,
  },
  computed: computedProps,
  attributes: attrProps,
})
```

The native comparison writes section artifacts into `<outDir>/<page>/artifacts/<section>/`. The JS verb then converts the raw result into a row for `summary.json`: it copies pixel counts, computes the policy classification, records screenshot paths, extracts style changes, extracts attribute changes, and computes bounds deltas.

This split is useful when you adapt the workflow for a project. Keep the low-level comparison call stable, and customize the spec reader, policy bands, naming rules, or output rows in JavaScript. Project-specific concepts belong in the JS verb; the Go core stays focused on browser automation and artifact primitives.

## Running the Smoke Setup

The repository includes a deterministic setup for testing the review website with visible differences. Start the fixture server from the repository root:

```bash
python3 -m http.server 18767
```

In a second terminal, build a local binary and generate the review data:

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

Serve the review site:

```bash
/tmp/css-visual-diff-review-smoke serve \
  --data-dir /tmp/cssvd-review-site-smoke \
  --port 18098 \
  --open
```

Open `http://127.0.0.1:18098`. The page should show four cards: `smoke/app`, `smoke/hero`, `smoke/card`, and `smoke/cta`. The fixture intentionally changes spacing, color, typography, border radius, card layout, and button text, so the classifications should include `tune-required` and `major-mismatch`.

## Reading Screenshots in the Review Site

Screenshots are the main evidence. The review site provides four view modes because each mode answers a different question.

| View mode | Use it when you want to answer |
| --- | --- |
| Side-by-side | What changed in layout, content, color, or component shape? |
| Overlay | Do two regions align? Are text baselines, edges, or card positions shifted? |
| Slider | Where does the visual change begin and end across the region? |
| Diff only | Which pixels changed, independent of the original visual content? |

Start with **Diff only** to locate the area of change. Move to **Side-by-side** to understand the visual meaning of that change. Use **Overlay** or **Slider** when the difference is spatial: shifted text, different padding, altered height, or different alignment.

The screenshot links at the bottom of each expanded card open the raw artifacts. Use these links when you need to inspect the exact PNG file outside the review UI or attach it to an issue.

## Reading CSS Differences

The CSS diff panel depends on the `computed` list in the spec. If a property is not listed, the example verb cannot show it as a style difference. Choose properties that explain the visual questions you expect to ask: font size, line height, margins, padding, color, background, radius, shadow, display, position, width, and height are good first choices.

CSS differences and pixel differences do not always have a one-to-one relationship. A changed `font-size` may produce many pixel differences. A changed `box-shadow` may produce a small pixel percentage but still matter visually. Use the CSS panel to explain likely causes; use the screenshots to decide visual importance.

## Serving Existing Runs

Once a data directory exists, serving it is independent of the original pages. The original URLs do not need to be live for review unless you want to re-run the comparison. The server only needs `summary.json`, `compare.json`, and artifact files on disk.

```bash
css-visual-diff serve \
  --data-dir /tmp/cssvd-review-site-smoke \
  --port 18098
```

The server provides:

| Endpoint | Purpose |
| --- | --- |
| `/api/manifest` | Returns the root `summary.json`. |
| `/api/compare?page=<page>&section=<section>` | Returns one section's `compare.json`. |
| `/artifacts/<page>/<section>/<file>` | Serves one PNG or JSON artifact from `<data-dir>/<page>/artifacts/<section>/<file>`. |

The React app converts absolute artifact paths from `summary.json` into `/artifacts/...` URLs before loading images. The Go server validates path segments so a browser request cannot escape the selected data directory.

## Adapting the Example to a Real Site

For a real site, copy the smoke spec or `examples/specs/review-sweep.example.yaml` and replace URLs and selectors:

```yaml
pages:
  homepage:
    leftUrl: https://production.example.com/
    rightUrl: http://localhost:5173/
    sections:
      header:
        selector: "header"
      hero:
        selector: ".hero"
      primary-action:
        selector: "[data-testid='primary-action']"
```

Run a small comparison first. One page and two or three sections are enough to validate URLs, selectors, waits, and artifact loading. Add more pages after the first data directory serves correctly.

If a selector fails, run a narrow inspection command or open the page in the browser and verify the DOM. A missing selector is usually caused by one of these conditions: the page has not finished rendering, the selector differs between prototype and implementation, the content is inside an iframe, or the target requires authentication.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `unknown flag: --spec-file` | The example verb uses camelCase field names from its JS metadata. | Use `--specFile` and `--outDir`. |
| The review site opens but shows no cards. | The server cannot read `summary.json`, or the data directory is wrong. | Check `css-visual-diff serve --data-dir ...` and verify `<data-dir>/summary.json` exists. |
| Cards show but images are broken. | Artifact paths do not match `<data-dir>/<page>/artifacts/<section>/<file>`, or the frontend bundle is stale. | Verify files exist, then rebuild the embedded review site with `make build-embed`; use `make build-web-local` if Docker/Dagger is unavailable. |
| `compare.json not found` when expanding a card. | A summary row references a page/section without a matching artifact directory. | Re-run `from-spec` or rebuild the summary with `examples review-sweep summary`. |
| A section records an error row. | The selector was missing, invalid, or not visible/capturable at comparison time. | Increase `waitMs`, set side-specific selectors, or inspect the page manually. |
| CSS diff is empty even though the screenshots differ. | The changed property is not listed in `computed`, or the difference is caused by layout/content rather than CSS properties in the list. | Add relevant properties to `computed` and re-run the comparison. |
| The same site compared to itself still has large diffs. | The page has animations, timestamps, randomized content, ads, lazy loading, or non-deterministic fonts. | Disable animations, wait longer, target smaller stable regions, or add project-specific prepare logic. |

## See Also

- `css-visual-diff help review-site` explains the interactive review UI, keyboard shortcuts, local storage, and export modal.
- `css-visual-diff help review-site-data-spec` defines the `summary.json`, `compare.json`, and artifact directory contracts in detail.
- `css-visual-diff help js-verb-review-sweep` explains the example verb that generates review-site data.
- `css-visual-diff help javascript-verbs` explains repository-scanned JavaScript verbs.
- `css-visual-diff help pixel-accuracy-scripting-guide` explains lower-level browser, locator, screenshot, and CSS inspection primitives.
