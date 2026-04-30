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
- run
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

The review site is a viewer. It does not run comparisons itself. It reads a pre-built summary and a set of artifact files produced by css-visual-diff's `compare` command, `run` command, or user-defined verb scripts. The separation is deliberate: capture and diff computation are expensive and browser-dependent, while review and annotation should be fast, local, and repeatable.

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

- `diff_comparison.png` — the side-by-side triptych. The review site does not currently display it, but including it does no harm.
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

The summary JSON can be either a bare object or a single-element list wrapping that object. Both shapes are accepted because the verb pipeline that produces summaries has varied over time.

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
   * One of: "accepted", "review", "tune-required", "major-mismatch".
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
  /** Normalized comparison dimensions (after resizing to match). */
  normalizedWidth?: number;
  normalizedHeight?: number;
}

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

Each section has its own `compare.json` file at `<page>/artifacts/<section>/compare.json`. This file contains the full structured comparison data: bounding boxes, pixel counts, computed style differences, attribute changes, text content, and source URLs.

The review site loads compare.json lazily when the reviewer expands a card. The CSS diff sidebar tab and the Meta tab both read from this file.

### Top-level shape

```typescript
interface CompareData {
  /** Schema version identifier. */
  schemaVersion: string;  // "cssvd.selectionComparison.v1"

  /** Comparison name, usually "<page>-<section>". */
  name: string;

  /** Bounding box comparison between prototype and React crops. */
  bounds: BoundsComparison;

  /** Prototype side metadata. */
  left: CompareSide;

  /** React side metadata. */
  right: CompareSide;

  /** Pixel diff statistics. */
  pixel: PixelData;

  /** All computed CSS properties (both changed and unchanged). */
  styles: StyleChange[];

  /** All inspected HTML attributes (both changed and unchanged). */
  attributes: AttributeChange[];

  /** Text content comparison. */
  text?: TextComparison;

  /** List of artifact files produced for this section. */
  artifacts: ArtifactRef[];
}
```

### CompareSide type

Describes one side of the comparison (prototype or React).

```typescript
interface CompareSide {
  /** Name, e.g. "about-content". */
  name: string;

  /** CSS selector used to crop the element. */
  selector: string;

  /** Full URL of the page that was captured. */
  url: string;

  /** Bounding box of the cropped element on the page. */
  bounds: { height: number; width: number; x: number; y: number };

  /** Whether the element was found on the page. */
  exists: boolean;

  /** Whether the element was visible. */
  visible: boolean;
}
```

### PixelData type

Pixel diff statistics from the normalized image comparison.

```typescript
interface PixelData {
  /** Percentage of pixels that differ, 0–100. */
  changedPercent: number;

  /** Absolute count of changed pixels. */
  changedPixels: number;

  /** Total pixels in the normalized comparison area. */
  totalPixels: number;

  /** Width of the normalized comparison image. */
  normalizedWidth: number;

  /** Height of the normalized comparison image. */
  normalizedHeight: number;

  /** Pixel diff threshold (0–255) used during comparison. */
  threshold: number;

  /** Absolute path to diff_only.png. */
  diffOnlyPath: string;

  /** Absolute path to diff_comparison.png. */
  diffComparisonPath: string;
}
```

### StyleChange type

A single CSS property comparison. Note: the file contains all inspected properties, not just the changed ones. Filter on `changed === true` to get only the differences.

```typescript
interface StyleChange {
  /** Whether this property differs between prototype and React. */
  changed: boolean;

  /** CSS property name, e.g. "font-size". */
  name: string;

  /** Value on the prototype side. */
  left: string;

  /** Value on the React side. */
  right: string;
}
```

### AttributeChange type

A single HTML attribute comparison. Like styles, this includes unchanged attributes. Filter on `changed === true` for differences only.

```typescript
interface AttributeChange {
  /** Whether this attribute differs. */
  changed: boolean;

  /** HTML attribute name, e.g. "class". */
  name: string;

  /** Value on the prototype side. Null if absent. */
  left?: string | null;

  /** Value on the React side. Null if absent. */
  right?: string | null;
}
```

### ArtifactRef type

A reference to one artifact file produced for this section.

```typescript
interface ArtifactRef {
  /** File kind, e.g. "png". */
  kind: string;

  /** Logical artifact name, e.g. "diffOnly", "diffComparison". */
  name: string;

  /** Absolute path to the artifact file. */
  path: string;
}
```

### Example: full compare.json

```json
{
  "schemaVersion": "cssvd.selectionComparison.v1",
  "name": "shows-content",
  "bounds": {
    "changed": true,
    "delta": { "height": 110.75, "width": 0, "x": 0, "y": 0 },
    "left":  { "height": 1739.11, "width": 920, "x": 0, "y": 61 },
    "right": { "height": 1849.86, "width": 920, "x": 0, "y": 61 }
  },
  "left": {
    "name": "shows-content",
    "selector": "[data-page='shows']",
    "url": "http://localhost:7070/standalone/public/shows.html",
    "bounds": { "height": 1739.11, "width": 920, "x": 0, "y": 61 },
    "exists": true,
    "visible": true
  },
  "right": {
    "name": "shows-content",
    "selector": "[data-page='shows']",
    "url": "http://localhost:6007/iframe.html?id=public-site-pages-shows--desktop&viewMode=story",
    "bounds": { "height": 1849.86, "width": 920, "x": 0, "y": 61 },
    "exists": true,
    "visible": true
  },
  "pixel": {
    "changedPercent": 11.605,
    "changedPixels": 197517,
    "totalPixels": 1702000,
    "normalizedWidth": 920,
    "normalizedHeight": 1850,
    "threshold": 30,
    "diffOnlyPath": "/tmp/run/shows/artifacts/content/diff_only.png",
    "diffComparisonPath": "/tmp/run/shows/artifacts/content/diff_comparison.png"
  },
  "styles": [
    { "changed": true,  "name": "background-color", "left": "rgba(0,0,0,0)",  "right": "rgb(255,255,255)" },
    { "changed": true,  "name": "font-size",        "left": "16px",          "right": "14px" },
    { "changed": false, "name": "display",           "left": "block",         "right": "block" }
  ],
  "attributes": [
    { "changed": true, "name": "class", "right": "pyxis-public-page pyxis-shows-page" }
  ],
  "text": {
    "changed": false,
    "left": "Providence, RIUpcoming shows...",
    "right": "Providence, RIUpcoming shows..."
  },
  "artifacts": [
    { "kind": "png", "name": "diffOnly",        "path": "/tmp/run/shows/artifacts/content/diff_only.png" },
    { "kind": "png", "name": "diffComparison",   "path": "/tmp/run/shows/artifacts/content/diff_comparison.png" }
  ]
}
```

---

## Part 4: Artifact PNG Files

Each section directory must contain at minimum these four PNG files. The images represent the visual evidence of the comparison.

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

Both `left_region.png` and `right_region.png` are cropped to the selector bounds on their respective pages. Before pixel comparison, css-visual-diff normalizes both images to the same dimensions (the larger of the two). The normalized dimensions are recorded in `compare.json` under `pixel.normalizedWidth` and `pixel.normalizedHeight`.

The `diff_only.png` image has the same normalized dimensions. It shows the same viewport area as the screenshots, but only the changed pixels are colored (typically red or magenta). Unchanged pixels are black or transparent.

### How screenshots are produced

Screenshots are captured by css-visual-diff using chromedp (Go Chrome DevTools Protocol). The browser navigates to the target URL, waits for the specified wait time, then captures a full-page or viewport screenshot. The relevant region is then cropped using the selector's bounding box.

```bash
# Single comparison produces all four PNG files:
css-visual-diff compare \
  --url1 http://localhost:7070/page.html \
  --selector1 "#content" \
  --url2 http://localhost:6007/iframe.html?id=page--desktop \
  --selector2 "#content" \
  --out /tmp/my-run/about/content
```

This creates:

```text
/tmp/my-run/about/content/
  compare.json
  compare.md
  left_region.png
  right_region.png
  diff_only.png
  diff_comparison.png
```

---

## Part 5: How to Generate Review Data

The Pyxis verb scripts are project-specific: they know about Pyxis page names, Storybook URLs, and selector conventions. This section explains how to build equivalent pipelines for any project using the general-purpose css-visual-diff commands.

There are three approaches, from simplest to most flexible.

### Approach A: Use `css-visual-diff compare` directly

This is the simplest approach and works when you have exactly two URLs to compare. Run `css-visual-diff compare` once per page/section, then build a summary JSON from the outputs.

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
  left_region.png
  right_region.png
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
import json, os, glob

def build_summary(run_dir):
    rows = []
    for compare_path in sorted(glob.glob(f"{run_dir}/*/artifacts/*/compare.json")):
        with open(compare_path) as f:
            data = json.load(f)

        parts = compare_path.replace(f"{run_dir}/", "").split("/")
        page = parts[0]
        section = parts[2]  # page/artifacts/section/compare.json

        artifact_dir = os.path.dirname(compare_path)

        row = {
            "page": page,
            "section": section,
            "classification": classify(data["pixel"]["changedPercent"]),
            "changedPercent": data["pixel"]["changedPercent"],
            "changedPixels": data["pixel"]["changedPixels"],
            "totalPixels": data["pixel"]["totalPixels"],
            "variant": "desktop",
            "diffOnlyPath": os.path.join(artifact_dir, "diff_only.png"),
            "diffComparisonPath": os.path.join(artifact_dir, "diff_comparison.png"),
            "leftRegionPath": os.path.join(artifact_dir, "left_region.png"),
            "rightRegionPath": os.path.join(artifact_dir, "right_region.png"),
            "artifactJson": compare_path,
            "leftSelector": data["left"]["selector"],
            "rightSelector": data["right"]["selector"],
            "styleChangeCount": sum(1 for s in data["styles"] if s["changed"]),
            "attributeChangeCount": sum(1 for a in data["attributes"] if a["changed"]),
            "styleDiffs": [
                {"property": s["name"], "left": s["left"], "right": s["right"]}
                for s in data["styles"] if s["changed"]
            ],
            "attributeDiffs": [
                {"attribute": a["name"], "left": a.get("left"), "right": a.get("right")}
                for a in data["attributes"] if a["changed"]
            ],
            "bounds": data["bounds"],
        }
        rows.append(row)

    classification_counts = {}
    for row in rows:
        cls = row["classification"]
        classification_counts[cls] = classification_counts.get(cls, 0) + 1

    pages = sorted(set(r["page"] for r in rows))
    max_pct = max(r["changedPercent"] for r in rows) if rows else 0

    summary = {
        "classificationCounts": classification_counts,
        "pageCount": len(pages),
        "maxChangedPercent": max_pct,
        "policy": {
            "ok": all(r["classification"] in ("accepted", "review") for r in rows),
            "worstClassification": max(
                rows, key=lambda r: severity(r["classification"])
            )["classification"] if rows else "accepted",
            "failureCount": sum(
                1 for r in rows if r["classification"] in ("tune-required", "major-mismatch")
            ),
        },
        "rows": rows,
    }

    with open(os.path.join(run_dir, "summary.json"), "w") as f:
        json.dump(summary, f, indent=2)

    print(f"Wrote {len(rows)} rows to {run_dir}/summary.json")

def classify(pct):
    if pct <= 0.5:  return "accepted"
    if pct <= 10:   return "review"
    if pct <= 30:   return "tune-required"
    return "major-mismatch"

def severity(cls):
    return {"accepted": 0, "review": 1, "tune-required": 2, "major-mismatch": 3}.get(cls, 0)

if __name__ == "__main__":
    import sys
    build_summary(sys.argv[1])
```

#### Step 3: Serve

```bash
css-visual-diff serve --data-dir /tmp/review-run --port 8097
```

### Approach B: Use verb scripts for custom pipelines

The `verbs` subsystem lets you write JavaScript verb scripts that orchestrate comparison, catalog, and summary generation in a single command. This is the preferred approach for project-scale suites and is what the Pyxis project uses.

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

---

## Part 6: The build-summary Utility

The `css-visual-diff` project includes a helper script at `ttmp/.../scripts/build-summary.py` in the CSSVD-REVIEW-SITE ticket workspace. This script implements Approach A's Step 2: it walks a run directory, reads every `compare.json`, and writes a `summary.json`.

This script is a reference implementation, not a production tool. For real projects, you should adapt it or write your own. The key design decisions in the script are:

1. **Classification bands are hardcoded** to match the Pyxis policy (0.5%, 10%, 30%). Adjust these thresholds for your own policy.
2. **Paths are written as absolute paths** pointing into the run directory. The review site's React app rewrites them to relative URLs.
3. **The `variant` field defaults to "desktop"** if not present in compare.json. If you compare at multiple viewports, add a `variant` field to your verb output or project spec.

To adapt this script for your project:

- Change the classification thresholds in `classify()`.
- Add or remove fields from the row to match what your compare.json files contain.
- Adjust the path layout if your artifacts are not under `<page>/artifacts/<section>/`.
