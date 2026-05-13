---
Title: Review Site Data Specification
Slug: review-site-data-spec
Short: Specification for the summary JSON, compare.json, and artifact directory layout consumed by the review site, and how to produce them from css-visual-diff output.
Topics:
- review
- data-format
- summary
- compare
- artifacts
Commands:
- serve
- compare
- verbs
Flags:
- data-dir
- summary
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The review site consumes two kinds of data: a **summary manifest** that lists every comparison to review, and a tree of **per-section artifacts** (screenshots and metadata). This document specifies both formats exactly, explains the directory layout the serve command expects, and describes how to produce them from css-visual-diff output.

If you are building a script or verb pipeline that feeds the review site, this is the contract your output must satisfy. If you are just running `css-visual-diff serve`, this explains what goes inside the data directory you pass with `--data-dir`.

## Overview

The review site is a viewer. It does not run comparisons itself. It reads a pre-built summary and a set of artifact files produced by css-visual-diff's `compare` command or by user-defined verb scripts such as the `examples review-sweep from-spec` workflow. The separation is deliberate: capture and diff computation are expensive and browser-dependent, while review and annotation should be fast, local, and repeatable.

There are three layers of data:

```text
┌─────────────────────────────────────────────────┐
│ 1. Summary manifest (summary.json)              │
│    Lists every page/section row to review.      │
│    Contains classification, percentages,        │
│    and artifact paths.                          │
└──────────────────┬──────────────────────────────┘
                   │ references
                   ▼
┌─────────────────────────────────────────────────┐
│ 2. Per-section compare.json                     │
│    Full comparison data for one section:        │
│    bounds, pixels, styles, attributes, text,    │
│    source URLs.                                 │
└──────────────────┬──────────────────────────────┘
                   │ references
                   ▼
┌─────────────────────────────────────────────────┐
│ 3. Artifact PNG files                           │
│    Screenshots: left_region.png,                │
│    right_region.png, diff_only.png,             │
│    diff_comparison.png                          │
└─────────────────────────────────────────────────┘
```

The summary is loaded first when the page opens. Each compare.json is loaded lazily when a reviewer expands a card. The PNG files are loaded as `<img>` elements inside the view modes.

---

## Part 1: Directory Layout

The serve command expects a data directory with this structure:

```text
<data-dir>/
├── summary.json                              ← manifest (required)
│
├── <page>/                                   ← one folder per page
│   ├── artifacts/                            ← required subdirectory name
│   │   └── <section>/                        ← one folder per section
│   │       ├── compare.json                  ← full comparison metadata
│   │       ├── left_region.png               ← prototype screenshot (cropped)
│   │       ├── right_region.png              ← react screenshot (cropped)
│   │       ├── diff_only.png                 ← changed-pixel highlight
│   │       └── diff_comparison.png           ← side-by-side triptych (optional)
│   ├── manifest.json                         ← per-page catalog (optional)
│   └── 01-catalog-index.md                   ← per-page index (optional)
│
├── <page2>/
│   └── artifacts/
│       └── <section>/
│           └── ...
└── ...
```

### What is required

The serve command requires:

- A `summary.json` at the root of the data directory (or at the path given by `--summary`).
- For every row in the summary, the corresponding artifact directory must exist and contain at minimum `compare.json`, `left_region.png`, `right_region.png`, and `diff_only.png`.

### What is optional

- `diff_comparison.png` — the side-by-side triptych. The review site mostly uses the side-by-side/overlay/slider views from `left_region.png` and `right_region.png`, but keeping this file is useful for manual artifact browsing and compatibility.
- `compare.md` — a human-readable markdown version of the comparison. The review site does not use this file.
- Per-page `manifest.json` and `01-catalog-index.md` — these are produced by the catalog/inspect workflow and are not consumed by the review site.

### How paths are resolved

The summary JSON contains absolute paths in its `diffOnlyPath`, `leftRegionPath`, `rightRegionPath`, `artifactJson`, and related fields. The review site's React app rewrites these absolute paths into relative `/artifacts/` URLs using a pattern-matching rule:

```text
Absolute path:
  /tmp/my-run/shows/artifacts/content/diff_only.png

Extracted pattern:
  <page>/<section>/<file>
  shows/content/diff_only.png

Rewritten URL:
  /artifacts/shows/content/diff_only.png
```

The Go serve handler receives `/artifacts/shows/content/diff_only.png`, splits it into `page=shows`, `section=content`, `file=diff_only.png`, and serves the file from `<data-dir>/shows/artifacts/content/diff_only.png`.

This mapping works because css-visual-diff always produces artifacts under `<page>/artifacts/<section>/`. The `artifacts` directory name is inserted automatically by the handler.

### Path-matching rule in detail

The rewriting function in the React app (`src/utils/paths.ts`) uses this regular expression:

```typescript
// Match: /any/prefix/<page>/artifacts/<section>/<filename>
const match = absolutePath.match(
  /\/([^/]+)\/artifacts\/([^/]+)\/([^/]+(?:\.[a-z]+)?)$/
);
// match[1] = page name
// match[2] = section name
// match[3] = filename with extension
```

If the path does not match this pattern (for example, if the artifact directory has a different structure), the path is passed through unchanged. This will typically result in a 404 when the browser tries to load it. Keep your artifact directories in the standard layout.

---

## Part 2: Summary JSON Specification

The summary JSON is the manifest that the review site loads on startup. It tells the site how many cards to show and what data each card contains.

### File location

- Default: `<data-dir>/summary.json`
- Override: `css-visual-diff serve --summary /path/to/summary.json`

### Top-level shape

The summary JSON can be either a bare object or a single-element list wrapping that object. Both shapes are accepted for compatibility with older verb pipelines.

```typescript
type SummaryPayload = SuiteSummary | [SuiteSummary];
```

### SuiteSummary type

```typescript
interface SuiteSummary {
  /** Count of rows per classification label. */
  classificationCounts: Record<string, number>;

  /** Total number of distinct pages in this run. */
  pageCount: number;

  /** Total number of sections (may be greater than pageCount if pages have multiple sections). */
  sectionCount?: number;

  /** Highest changedPercent across all rows. */
  maxChangedPercent: number;

  /** Policy evaluation result. */
  policy: {
    /** Whether the run passes the policy. */
    ok: boolean;
    /** Worst classification produced by any row. */
    worstClassification: string;
    /** Number of rows that failed policy. */
    failureCount: number;
  };

  /** One entry per page/section comparison. */
  rows: SummaryRow[];
}
```

### SummaryRow type

Each row represents one page/section comparison that the reviewer should evaluate.

```typescript
interface SummaryRow {
  /** Page name, e.g. "about", "shows", "archive". */
  page: string;

  /** Section within the page, e.g. "content", "header", "page". */
  section: string;

  /**
   * Computed classification from the policy bands.
   * Common values: "accepted", "review", "tune-required", "major-mismatch".
   * Generator pipelines may also emit "error" rows for failed sections.
   */
  classification: string;

  /** Percentage of pixels that differ, 0–100. */
  changedPercent: number;

  /** Absolute count of changed pixels. */
  changedPixels: number;

  /** Total pixels in the normalized comparison area. */
  totalPixels?: number;

  /** Pixel diff threshold used for this comparison (0–255). */
  threshold?: number;

  /** Variant name, e.g. "desktop", "mobile". */
  variant?: string;

  /** Absolute path to the prototype screenshot. */
  leftRegionPath: string;

  /** Absolute path to the React screenshot. */
  rightRegionPath: string;

  /** Absolute path to the changed-pixel highlight image. */
  diffOnlyPath: string;

  /** Absolute path to the side-by-side triptych image. */
  diffComparisonPath?: string;

  /** Absolute path to the compare.json for this section. */
  artifactJson: string;

  /** Absolute path to the compare.md for this section (optional). */
  artifactMarkdown?: string;

  /** CSS selector used to crop the prototype screenshot. */
  leftSelector: string;

  /** CSS selector used to crop the React screenshot. */
  rightSelector: string;

  /** Number of CSS properties that differ. */
  styleChangeCount: number;

  /** Number of HTML attributes that differ. */
  attributeChangeCount: number;

  /** List of CSS properties that differ between prototype and React. */
  styleDiffs: StyleDiff[];

  /** List of HTML attributes that differ. */
  attributeDiffs: AttributeDiff[];

  /** Bounding box comparison between prototype and React crops. */
  bounds: BoundsComparison;

  /** Text content comparison (optional). */
  text?: TextComparison;
}
```

### Supporting types

```typescript
interface StyleDiff {
  /** CSS property name, e.g. "font-size", "padding-top". */
  property: string;
  /** Value on the prototype side. */
  left: string;
  /** Value on the React side. */
  right: string;
}

interface AttributeDiff {
  /** HTML attribute name, e.g. "class", "data-page". */
  attribute: string;
  /** Value on the prototype side. Null if absent. */
  left: string | null;
  /** Value on the React side. Null if absent. */
  right: string | null;
}

interface BoundsComparison {
  /** Whether the bounds differ between prototype and React. */
  changed: boolean;
  /** Difference between right and left bounds. */
  delta: { height: number; width: number; x: number; y: number };
  /** Prototype crop bounds. */
  left: { height: number; width: number; x: number; y: number };
  /** React crop bounds. */
  right: { height: number; width: number; x: number; y: number };
  /** Optional normalized comparison dimensions (after resizing to match). */
  normalizedWidth?: number;
  normalizedHeight?: number;
}

For current snake_case `compare.json`, the raw bounds use `width`/`height`, and normalized image dimensions are stored in `pixel_diff.normalized_width` and `pixel_diff.normalized_height`. Summary rows may still use the camelCase `normalizedWidth`/`normalizedHeight` names if a generator chooses to copy those values into `bounds`.

interface TextComparison {
  /** Whether the text content differs. */
  changed: boolean;
  /** Full text content of the prototype element. */
  left: string;
  /** Full text content of the React element. */
  right: string;
}
```

### Required vs optional fields

The review site degrades gracefully when optional fields are missing. Here is the minimum viable row:

```json
{
  "page": "about",
  "section": "content",
  "classification": "review",
  "changedPercent": 7.25,
  "changedPixels": 62500,
  "diffOnlyPath": "/path/to/about/artifacts/content/diff_only.png",
  "leftRegionPath": "/path/to/about/artifacts/content/left_region.png",
  "rightRegionPath": "/path/to/about/artifacts/content/right_region.png",
  "artifactJson": "/path/to/about/artifacts/content/compare.json",
  "leftSelector": "[data-page='about']",
  "rightSelector": "[data-page='about']",
  "styleChangeCount": 0,
  "attributeChangeCount": 0,
  "styleDiffs": [],
  "attributeDiffs": [],
  "bounds": {
    "changed": false,
    "delta": { "height": 0, "width": 0, "x": 0, "y": 0 },
    "left": { "height": 800, "width": 920, "x": 0, "y": 61 },
    "right": { "height": 800, "width": 920, "x": 0, "y": 61 }
  }
}
```

The site will render a card with images and basic metadata. The CSS diff sidebar tab will show no style changes. The Meta tab will show whatever bounds data is available. All other fields enhance the display but are not strictly required.

### Example: minimal summary.json

```json
{
  "classificationCounts": { "review": 2 },
  "pageCount": 2,
  "maxChangedPercent": 9.1,
  "policy": { "ok": false, "worstClassification": "review", "failureCount": 0 },
  "rows": [
    {
      "page": "home",
      "section": "hero",
      "classification": "review",
      "changedPercent": 9.1,
      "changedPixels": 42000,
      "totalPixels": 460000,
      "diffOnlyPath": "/tmp/run/home/artifacts/hero/diff_only.png",
      "leftRegionPath": "/tmp/run/home/artifacts/hero/left_region.png",
      "rightRegionPath": "/tmp/run/home/artifacts/hero/right_region.png",
      "artifactJson": "/tmp/run/home/artifacts/hero/compare.json",
      "leftSelector": "#hero",
      "rightSelector": "#hero",
      "styleChangeCount": 3,
      "attributeChangeCount": 0,
      "styleDiffs": [
        { "property": "font-size", "left": "16px", "right": "14px" },
        { "property": "color", "left": "rgb(0,0,0)", "right": "rgb(26,26,24)" },
        { "property": "padding-top", "left": "40px", "right": "0px" }
      ],
      "attributeDiffs": [],
      "bounds": {
        "changed": true,
        "delta": { "height": 48, "width": 0, "x": 0, "y": 0 },
        "left": { "height": 800, "width": 920, "x": 0, "y": 61 },
        "right": { "height": 848, "width": 920, "x": 0, "y": 61 }
      }
    },
    {
      "page": "pricing",
      "section": "cards",
      "classification": "accepted",
      "changedPercent": 0.3,
      "changedPixels": 1400,
      "totalPixels": 460000,
      "diffOnlyPath": "/tmp/run/pricing/artifacts/cards/diff_only.png",
      "leftRegionPath": "/tmp/run/pricing/artifacts/cards/left_region.png",
      "rightRegionPath": "/tmp/run/pricing/artifacts/cards/right_region.png",
      "artifactJson": "/tmp/run/pricing/artifacts/cards/compare.json",
      "leftSelector": "#pricing-cards",
      "rightSelector": "#pricing-cards",
      "styleChangeCount": 0,
      "attributeChangeCount": 0,
      "styleDiffs": [],
      "attributeDiffs": [],
      "bounds": {
        "changed": false,
        "delta": { "height": 0, "width": 0, "x": 0, "y": 0 },
        "left": { "height": 600, "width": 920, "x": 0, "y": 100 },
        "right": { "height": 600, "width": 920, "x": 0, "y": 100 }
      }
    }
  ]
}
```

---

## Part 3: Compare JSON Specification

Each section has its own `compare.json` file at `<page>/artifacts/<section>/compare.json`. The review site loads this file lazily when a card expands. Current css-visual-diff comparison output uses the Go/native snake_case shape produced by `css-visual-diff compare` and `require("diff").compareRegion(...)`. The frontend also has adapters for an older camelCase shape, but new generators should prefer the snake_case format below.

### Current snake_case shape

```typescript
interface CompareResult {
  inputs: CompareInputs;
  url1: CompareSideResult;
  url2: CompareSideResult;
  computed_diffs: StyleDiff[];
  winner_diffs?: WinnerDiff[];
  pixel_diff: PixelDiffStats;
}

interface CompareInputs {
  url1: string;
  selector1: string;
  wait_ms1: number;
  url2: string;
  selector2: string;
  wait_ms2: number;
  viewport_w: number;
  viewport_h: number;
  props: string[];
  attrs: string[];
  out_dir: string;
}

interface CompareSideResult {
  url: string;
  selector: string;
  full_screenshot: string;      // url1_full.png or url2_full.png
  element_screenshot: string;   // url1_screenshot.png or url2_screenshot.png
  computed: {
    exists: boolean;
    visible: boolean;
    bounds?: { x: number; y: number; width: number; height: number };
    computed: Record<string, string>;
    attributes: Record<string, string | null>;
  };
  matched?: unknown;
}

interface StyleDiff {
  property: string;
  left: string;
  right: string;
  changed: boolean;
}

interface PixelDiffStats {
  threshold: number;
  total_pixels: number;
  changed_pixels: number;
  changed_percent: number;
  normalized_width: number;
  normalized_height: number;
  diff_comparison_path: string;
  diff_only_path: string;
}
```

The review-site frontend reads the snake_case format through `web/review-site/src/utils/compareData.ts`. It derives changed styles from `computed_diffs`, attributes from `url1.computed.attributes` vs `url2.computed.attributes`, bounds from `url1.computed.bounds` and `url2.computed.bounds`, source URLs from `url1.url` and `url2.url`, and pixel stats from `pixel_diff`.

### Current example

```json
{
  "inputs": {
    "url1": "http://localhost:7070/page.html",
    "selector1": "#content",
    "wait_ms1": 250,
    "url2": "http://localhost:5173/",
    "selector2": "#content",
    "wait_ms2": 250,
    "viewport_w": 1280,
    "viewport_h": 720,
    "props": ["font-size", "color", "padding-top"],
    "attrs": ["id", "class"],
    "out_dir": "/tmp/run/home/artifacts/content"
  },
  "url1": {
    "url": "http://localhost:7070/page.html",
    "selector": "#content",
    "full_screenshot": "/tmp/run/home/artifacts/content/url1_full.png",
    "element_screenshot": "/tmp/run/home/artifacts/content/url1_screenshot.png",
    "computed": {
      "exists": true,
      "visible": true,
      "bounds": { "x": 0, "y": 61, "width": 920, "height": 800 },
      "computed": { "font-size": "16px", "color": "rgb(0, 0, 0)" },
      "attributes": { "id": "content", "class": "hero" }
    }
  },
  "url2": {
    "url": "http://localhost:5173/",
    "selector": "#content",
    "full_screenshot": "/tmp/run/home/artifacts/content/url2_full.png",
    "element_screenshot": "/tmp/run/home/artifacts/content/url2_screenshot.png",
    "computed": {
      "exists": true,
      "visible": true,
      "bounds": { "x": 0, "y": 61, "width": 920, "height": 848 },
      "computed": { "font-size": "14px", "color": "rgb(26, 26, 24)" },
      "attributes": { "id": "content", "class": "hero compact" }
    }
  },
  "computed_diffs": [
    { "property": "font-size", "left": "16px", "right": "14px", "changed": true },
    { "property": "color", "left": "rgb(0, 0, 0)", "right": "rgb(26, 26, 24)", "changed": true }
  ],
  "winner_diffs": [],
  "pixel_diff": {
    "threshold": 30,
    "total_pixels": 780160,
    "changed_pixels": 62500,
    "changed_percent": 8.011,
    "normalized_width": 920,
    "normalized_height": 848,
    "diff_comparison_path": "/tmp/run/home/artifacts/content/diff_comparison.png",
    "diff_only_path": "/tmp/run/home/artifacts/content/diff_only.png"
  }
}
```

### Legacy camelCase shape

Older prototypes and some adapter code refer to a camelCase shape with fields such as `left`, `right`, `pixel`, `styles`, and `attributes`. The current frontend still tolerates enough of that shape for the sidebar adapters, but it is no longer the primary format emitted by `css-visual-diff compare` or `examples/verbs/review-sweep from-spec`. New producers should emit the snake_case format or at least ensure the `summary.json` row contains all fields needed for the card list and image URLs.

---

## Part 4: Artifact PNG Files

Each review-site section directory should contain the review-site aliases below. The images represent the visual evidence of the comparison. Note that direct `css-visual-diff compare` writes element screenshots as `url1_screenshot.png` and `url2_screenshot.png`; the `examples review-sweep from-spec` verb copies those files to the review-site aliases `left_region.png` and `right_region.png`. If you produce review-site data yourself, either write the aliases directly or copy them as part of summary generation.

### Required files

| File | Description |
| --- | --- |
| `left_region.png` | Screenshot of the prototype element, cropped to the selector bounds. |
| `right_region.png` | Screenshot of the React element, cropped to the selector bounds. |
| `diff_only.png` | Image showing only the pixels that differ, highlighted against a neutral background. |

### Optional files

| File | Description |
| --- | --- |
| `diff_comparison.png` | Side-by-side triptych: left, diff, right. Useful for broad context. |

### Image dimensions

Both `left_region.png` and `right_region.png` are cropped to the selector bounds on their respective pages. Before pixel comparison, css-visual-diff normalizes both images to the same dimensions (the larger of the two). The normalized dimensions are recorded in current `compare.json` files under `pixel_diff.normalized_width` and `pixel_diff.normalized_height`.

The `diff_only.png` image has the same normalized dimensions. It shows the same viewport area as the screenshots, but only the changed pixels are colored (typically red or magenta). Unchanged pixels are black or transparent.

### How screenshots are produced

Screenshots are captured by css-visual-diff using chromedp (Go Chrome DevTools Protocol). The browser navigates to the target URL, waits for the specified wait time, then captures a full-page or viewport screenshot. The relevant region is then cropped using the selector's bounding box.

```bash
# Direct comparison produces compare.json, compare.md, url1/url2 screenshots, and diff PNGs:
css-visual-diff compare \
  --url1 http://localhost:7070/page.html \
  --selector1 "#content" \
  --url2 http://localhost:6007/iframe.html?id=page--desktop \
  --selector2 "#content" \
  --out /tmp/my-run/about/artifacts/content
```

This creates:

```text
/tmp/my-run/about/artifacts/content/
  compare.json
  compare.md
  url1_screenshot.png              # copy/alias to left_region.png for the review site
  url2_screenshot.png              # copy/alias to right_region.png for the review site
  diff_only.png
  diff_comparison.png
```

---

## Part 5: How to Generate Review Data

This section explains how to build review-site data pipelines for any project using the general-purpose css-visual-diff commands and JavaScript verbs.

There are two common approaches, from simplest to most flexible.

### Approach A: Use `css-visual-diff compare` directly

This is the simplest approach and works when you have exactly two URLs to compare. Run `css-visual-diff compare` once per page/section, copy `url1_screenshot.png`/`url2_screenshot.png` to the review-site aliases, then build a summary JSON from the outputs.

#### Step 1: Run comparisons

For each page and section you want to compare, run:

```bash
css-visual-diff compare \
  --url1 http://localhost:7070/about.html \
  --selector1 "[data-page='about']" \
  --url2 http://localhost:6007/iframe.html?id=about--desktop \
  --selector2 "[data-page='about']" \
  --out /tmp/review-run/about/artifacts/content
```

This produces the artifact directory for one section:

```text
/tmp/review-run/about/artifacts/content/
  compare.json
  compare.md
  url1_screenshot.png
  url2_screenshot.png
  left_region.png                  # alias for review site
  right_region.png                 # alias for review site
  diff_only.png
  diff_comparison.png
```

Repeat for every page/section. Place each output at `<page>/artifacts/<section>/` inside a common run directory.

#### Step 2: Build the summary JSON

Write a small script (Python, bash, or Node) that:

1. Walks the run directory looking for `compare.json` files.
2. Reads each `compare.json` to extract page, section, classification, changedPercent, and artifact paths.
3. Assembles a `SuiteSummary` object with all rows.
4. Writes `summary.json` at the root of the run directory.

Here is a Python sketch:

```python
import glob, json, os, shutil

def build_summary(run_dir):
    rows = []
    for compare_path in sorted(glob.glob(f"{run_dir}/*/artifacts/*/compare.json")):
        with open(compare_path) as f:
            data = json.load(f)

        parts = compare_path.replace(f"{run_dir}/", "").split("/")
        page = parts[0]
        section = parts[2]  # page/artifacts/section/compare.json
        artifact_dir = os.path.dirname(compare_path)

        # Direct compare writes url1_screenshot/url2_screenshot. The review site
        # expects left_region/right_region URLs in summary.json.
        alias(os.path.join(artifact_dir, "url1_screenshot.png"), os.path.join(artifact_dir, "left_region.png"))
        alias(os.path.join(artifact_dir, "url2_screenshot.png"), os.path.join(artifact_dir, "right_region.png"))

        pixel = data.get("pixel_diff", {})
        url1 = data.get("url1", {})
        url2 = data.get("url2", {})
        left_computed = (url1.get("computed") or {})
        right_computed = (url2.get("computed") or {})
        left_bounds = left_computed.get("bounds") or {}
        right_bounds = right_computed.get("bounds") or {}
        left_attrs = left_computed.get("attributes") or {}
        right_attrs = right_computed.get("attributes") or {}

        style_diffs = [
            {"property": s.get("property", ""), "left": s.get("left", ""), "right": s.get("right", "")}
            for s in data.get("computed_diffs", []) if s.get("changed")
        ]
        attribute_diffs = [
            {"attribute": name, "left": left_attrs.get(name), "right": right_attrs.get(name)}
            for name in sorted(set(left_attrs) | set(right_attrs))
            if left_attrs.get(name) != right_attrs.get(name)
        ]
        pct = pixel.get("changed_percent", 0)

        row = {
            "page": page,
            "section": section,
            "classification": classify(pct),
            "changedPercent": pct,
            "changedPixels": pixel.get("changed_pixels", 0),
            "totalPixels": pixel.get("total_pixels", 0),
            "threshold": pixel.get("threshold", 30),
            "variant": "desktop",
            "diffOnlyPath": os.path.join(artifact_dir, "diff_only.png"),
            "diffComparisonPath": os.path.join(artifact_dir, "diff_comparison.png"),
            "leftRegionPath": os.path.join(artifact_dir, "left_region.png"),
            "rightRegionPath": os.path.join(artifact_dir, "right_region.png"),
            "artifactJson": compare_path,
            "leftSelector": url1.get("selector", ""),
            "rightSelector": url2.get("selector", ""),
            "styleChangeCount": len(style_diffs),
            "attributeChangeCount": len(attribute_diffs),
            "styleDiffs": style_diffs,
            "attributeDiffs": attribute_diffs,
            "bounds": bounds(left_bounds, right_bounds),
        }
        rows.append(row)

    classification_counts = {}
    for row in rows:
        cls = row["classification"]
        classification_counts[cls] = classification_counts.get(cls, 0) + 1

    pages = sorted(set(r["page"] for r in rows))
    max_pct = max((r["changedPercent"] for r in rows), default=0)

    summary = {
        "classificationCounts": classification_counts,
        "pageCount": len(pages),
        "sectionCount": len(rows),
        "maxChangedPercent": max_pct,
        "policy": {
            "ok": all(r["classification"] in ("accepted", "review") for r in rows),
            "worstClassification": max(rows, key=lambda r: severity(r["classification"]))["classification"] if rows else "accepted",
            "failureCount": sum(1 for r in rows if r["classification"] in ("tune-required", "major-mismatch", "error")),
        },
        "rows": rows,
    }

    with open(os.path.join(run_dir, "summary.json"), "w") as f:
        json.dump(summary, f, indent=2)

    print(f"Wrote {len(rows)} rows to {run_dir}/summary.json")

def alias(src, dst):
    if os.path.exists(src) and not os.path.exists(dst):
        shutil.copyfile(src, dst)

def bounds(left, right):
    if not left or not right:
        return {}
    return {
        "changed": any(left.get(k, 0) != right.get(k, 0) for k in ("x", "y", "width", "height")),
        "delta": {k: right.get(k, 0) - left.get(k, 0) for k in ("x", "y", "width", "height")},
        "left": left,
        "right": right,
    }

def classify(pct):
    if pct <= 0.5:  return "accepted"
    if pct <= 10:   return "review"
    if pct <= 30:   return "tune-required"
    return "major-mismatch"

def severity(cls):
    return {"accepted": 0, "review": 1, "tune-required": 2, "major-mismatch": 3, "error": 4}.get(cls, 0)

if __name__ == "__main__":
    import sys
    build_summary(sys.argv[1])
```

#### Step 3: Serve

```bash
css-visual-diff serve --data-dir /tmp/review-run --port 8097
```

### Approach B: Use verb scripts for custom pipelines

The `verbs` subsystem lets you write JavaScript verb scripts that orchestrate comparison, catalog, and summary generation in a single command. This is the preferred approach for project-scale suites.

A verb script has access to the `css-visual-diff` JavaScript API, which provides browser automation, catalog management, screenshot capture, and structured output. The script can load project-specific data files, run comparisons for many pages, collect results, and emit a summary JSON in the exact format the review site expects.

#### Example verb structure

Create a verb repository directory:

```text
my-verbs/
  verbs.js         ← verb definitions
  summary.js       ← summary builder
```

The verb script uses the `css-visual-diff` JS API to:

1. Open a browser.
2. Navigate to each prototype page and capture a screenshot.
3. Navigate to each React page and capture a screenshot.
4. Compare the two screenshots pixel by pixel.
5. Extract computed CSS for both sides.
6. Write per-section `compare.json` files.
7. Write a top-level `summary.json`.

This approach gives you full control over the comparison pipeline, including custom prepare scripts, dynamic selectors, and multi-viewport comparisons. See `css-visual-diff help javascript-verbs` for the full verb API documentation.

#### Built-in example: review-sweep

The repository includes a complete external verb example at `examples/verbs/review-sweep.js`. It reads a small project spec, runs comparisons, writes artifacts in the review-site directory layout, and emits `summary.json`.

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep from-spec \
  --specFile examples/specs/review-sweep.example.yaml \
  --outDir /tmp/example-review

css-visual-diff serve --data-dir /tmp/example-review --port 8098
```

If you already have artifacts on disk and only need to rebuild `summary.json`, use:

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep summary \
  --specFile examples/specs/review-sweep.example.yaml \
  --outDir /tmp/example-review
```

The example demonstrates `require("yaml")`, `require("fs")`, `require("path")`, and `require("diff").compareRegion(...)` inside the go-go-goja VM.

### Choosing an approach

| Approach | Best for | Pros | Cons |
| --- | --- | --- | --- |
| A: `compare` | Quick one-off reviews, 1–5 sections | Simple, no YAML needed | Manual, tedious for many sections |
| B: Verb scripts | Complex or project-scale pipelines | Full flexibility, programmatic, can load project specs | More code to write and maintain |

Both approaches produce the same directory structure and JSON formats. The review site does not care how the data was produced; it only cares that `summary.json` and the artifact directories exist and follow the spec.
