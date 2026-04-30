---
Title: JSVerb Review Site Generator — Analysis & Implementation Guide
Slug: jsverb-review-generator-analysis
Ticket: CSSVD-JSVERB-REVIEW
Status: draft
---

# JSVerb Review Site Generator — Analysis & Implementation Guide

## 1. Problem Statement

The css-visual-diff review site is a static SPA that reads a pre-built data directory containing a `summary.json` manifest and per-section `compare.json` + PNG artifacts. Currently, the only way to produce this data directory is through ad-hoc project-specific scripts (the Pyxis Python pipeline) or by manually running `css-visual-diff compare` per section and assembling the summary by hand.

We need a **general-purpose JSVerb example** that:

1. Reads a project spec (YAML file) declaring pages, sections, selectors, URLs, and policy bands.
2. Runs `diff.compareRegion()` for each declared section.
3. Writes all artifacts to disk in the exact directory layout the review site expects.
4. Assembles and writes a `summary.json` at the root.
5. Can be served immediately with `css-visual-diff serve --data-dir <outDir>`.

This verb demonstrates the full power of the css-visual-diff JavaScript API while solving a real need: any project can point the verb at a YAML spec and get an instant interactive review site.

## 2. Current State Analysis

### 2.1 What already exists

**Go runtime (VM setup)** — The verb CLI (`internal/cssvisualdiff/verbcli/`) constructs a go-go-goja runtime via `dsl.NewRuntimeFactory()` which calls:

```go
engine.NewBuilder(opts...).
    WithRequireOptions(noderequire.WithLoader(registry.RequireLoader())).
    WithModules(engine.DefaultRegistryModules()).
    WithRuntimeModuleRegistrars(newRuntimeRegistrar())
```

`DefaultRegistryModules()` registers **all** modules from the go-go-goja default registry. This includes:

| Module | require() name | Status |
|---|---|---|
| File I/O | `fs`, `node:fs` | ✅ Available (sync + async) |
| YAML | `yaml` | ✅ Available (parse/stringify/validate) |
| Path | `path`, `node:path` | ✅ Available |
| OS | `os`, `node:os` | ✅ Available |
| Crypto | `crypto`, `node:crypto` | ✅ Available |
| Events | `events`, `node:events` | ✅ Available |
| Exec | `exec` | ✅ Available |
| Timer | `timer` | ✅ Available |
| Time | `time` | ✅ Available |

**No changes needed to the Go runtime.** The VM already has `fs`, `yaml`, `path`, and everything the verb needs.

**CSS Visual Diff API** — `require("css-visual-diff")` provides:

- `cvd.browser()` → browser service with `page()`, `newPage()`
- `cvd.catalog({...})` → catalog builder with targets, results, manifests
- `cvd.loadConfig(path)` → loads a YAML config file into a JS object
- `cvd.write.json(path, data)` → writes JSON to disk
- `cvd.write.markdown(path, text)` → writes markdown to disk
- `cvd.diff(left, right, options)` → computes snapshot diffs between two JS values
- `cvd.extract(locator, extractors)` → extracts computed styles, bounds, attributes
- `cvd.snapshot(page, probes)` → structured snapshot API

**Diff module** — `require("diff")` provides:

- `diff.compareRegion({...})` → the core comparison primitive. Takes left/right URLs, selectors, viewport, output settings. Opens two browser pages, captures screenshots, computes pixel diff, extracts CSS/attributes/bounds, writes all artifacts (PNG + JSON + markdown). Returns a `CompareResult` object.

This is the key building block. The verb loops over sections, calls `compareRegion()` for each one, and collects the results.

### 2.2 What needs to be built

A single JS file: `examples/verbs/review-sweep.js`

The file declares:

1. A `__package__` annotation for the `examples.review` namespace.
2. A `__section__` for the project spec (YAML config path).
3. Two verbs:
   - `review-sweep from-spec` — reads a YAML spec, runs all comparisons, writes the full data directory.
   - `review-sweep summary` — walks an existing data directory and (re)builds `summary.json` from the `compare.json` files on disk. Useful for re-generating the summary after manual edits.

### 2.3 YAML spec format

The verb reads a YAML file that describes what to compare. This is a new format, designed to be simple and general-purpose.

```yaml
# review-sweep.spec.yaml
name: my-project-visual-sweep
variant: desktop

viewport:
  width: 920
  height: 1460

defaults:
  waitMs: 1000
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
  - font-family
  - font-size
  - font-weight
  - line-height
  - padding-top
  - padding-right
  - padding-bottom
  - padding-left
  - border-radius
  - color
  - background-color
  - box-shadow

attributes:
  - id
  - class

pages:
  about:
    leftUrl: http://localhost:7070/about.html
    rightUrl: http://localhost:6007/iframe.html?id=about--desktop&viewMode=story
    sections:
      content:
        selector: "[data-page='about']"
      header:
        selector: "header"
  pricing:
    leftUrl: http://localhost:7070/pricing.html
    rightUrl: http://localhost:6007/iframe.html?id=pricing--desktop&viewMode=story
    sections:
      cards:
        selector: "#pricing-cards"
      hero:
        selector: ".pricing-hero"
```

The spec is intentionally flat and explicit. Each page has `leftUrl`/`rightUrl` and a list of named sections with selectors. No indirection, no templating, no scripting — just declarative comparison targets.

### 2.4 Spec field reference

| Field | Level | Type | Default | Description |
|---|---|---|---|---|
| `name` | root | string | required | Project name, used as run directory prefix |
| `variant` | root | string | `"desktop"` | Variant label written into each row |
| `viewport.width` | root | int | `920` | Browser viewport width |
| `viewport.height` | root | int | `1460` | Browser viewport height |
| `defaults.waitMs` | root | int | `1000` | Wait after navigation (ms) |
| `defaults.threshold` | root | int | `30` | Pixel diff threshold (0-255) |
| `policy.bands[].name` | root | string | required | Classification name |
| `policy.bands[].maxChangedPercent` | root | float | required | Upper bound for this band |
| `computed` | root | string[] | (see below) | CSS properties to extract |
| `attributes` | root | string[] | `["id","class"]` | HTML attributes to extract |
| `pages.<name>.leftUrl` | page | string | required | Prototype URL |
| `pages.<name>.rightUrl` | page | string | required | React URL |
| `pages.<name>.leftWaitMs` | page | int | from defaults | Override wait for left page |
| `pages.<name>.rightWaitMs` | page | int | from defaults | Override wait for right page |
| `pages.<name>.sections.<name>.selector` | section | string | required | CSS selector for both sides |
| `pages.<name>.sections.<name>.leftSelector` | section | string | from selector | Override selector for left |
| `pages.<name>.sections.<name>.rightSelector` | section | string | from selector | Override selector for right |

Default `computed` properties when not specified:

```yaml
computed:
  - display
  - position
  - width
  - height
  - margin-top
  - margin-right
  - margin-bottom
  - margin-left
  - padding-top
  - padding-right
  - padding-bottom
  - padding-left
  - font-family
  - font-size
  - font-weight
  - line-height
  - color
  - background-color
  - background-image
  - border-radius
  - box-shadow
  - z-index
```

## 3. Verb Design

### 3.1 File location

```
examples/verbs/review-sweep.js
```

This is an external verb repository file. It is discovered via:

```bash
css-visual-diff verbs --repository examples/verbs examples review-sweep from-spec ...
```

### 3.2 Package declaration

```javascript
__package__({
  name: "review-sweep",
  parents: ["examples"],
  short: "Generate a review site data directory from a YAML spec"
});
```

### 3.3 Sections

```javascript
__section__("spec", {
  title: "Spec",
  description: "YAML spec file declaring pages and sections to compare.",
  fields: {
    specFile: {
      type: "string",
      required: true,
      help: "Path to the YAML spec file"
    }
  }
});

__section__("sweepOutput", {
  title: "Output",
  description: "Output directory for the review site data.",
  fields: {
    outDir: {
      type: "string",
      required: true,
      help: "Output directory (will be created if needed)"
    },
    writeMarkdown: {
      type: "bool",
      default: true,
      help: "Write compare.md alongside compare.json"
    }
  }
});
```

### 3.4 Verb: from-spec

The main verb. Reads the YAML spec, runs comparisons, writes everything.

**Flow:**

```
1. Parse YAML spec file → spec object
2. Validate spec (at least one page, sections exist)
3. Create output directory structure
4. For each page/section:
   a. Call diff.compareRegion({
        left: { url, selector, waitMs },
        right: { url, selector, waitMs },
        viewport: { width, height },
        output: {
          outDir: "<outDir>/<page>/artifacts/<section>",
          threshold,
          writeJson: true,
          writeMarkdown,
          writePngs: true,
        },
        computed: spec.computed,
        attributes: spec.attributes,
      })
   b. Collect the CompareResult
   c. Classify using policy bands
   d. Build a SummaryRow from the result
5. Assemble SuiteSummary from all rows
6. Write summary.json to <outDir>/summary.json
7. Return summary statistics
```

**Key implementation detail:** `diff.compareRegion()` is synchronous (it blocks the runtime while opening browser pages, capturing screenshots, and computing diffs). This means the verb processes sections sequentially. For a spec with 13 sections, this takes approximately 2-5 minutes depending on page load times. This is acceptable because comparison is inherently browser-bound.

### 3.5 Verb: summary

A utility verb that rebuilds `summary.json` from existing artifacts without re-running comparisons. Useful when:

- The summary was lost or corrupted.
- The user changed policy bands and wants to re-classify.
- The user manually deleted some sections and wants an updated summary.

**Flow:**

```
1. Walk <outDir> for compare.json files (pattern: <page>/artifacts/<section>/compare.json)
2. For each compare.json:
   a. Read and parse the file
   b. Extract page, section, paths, pixel data, style/attribute diffs
   c. Classify using policy bands from spec (or hardcoded defaults)
3. Assemble SuiteSummary
4. Write summary.json
5. Return statistics
```

This verb uses `fs` directly to walk the directory tree and read JSON files. It does not open a browser.

## 4. Implementation Sketch

### 4.1 Classification function

```javascript
function classify(changedPercent, bands) {
  // bands is sorted by maxChangedPercent ascending
  for (const band of bands) {
    if (changedPercent <= band.maxChangedPercent) {
      return band.name;
    }
  }
  return bands[bands.length - 1].name;
}
```

### 4.2 Summary row builder

```javascript
function buildRow(pageName, sectionName, compareResult, spec) {
  const pixel = compareResult.pixel || {};
  const pct = pixel.changedPercent || 0;
  const classification = classify(pct, spec.policy.bands);
  
  const artifactDir = `${spec._outDir}/${pageName}/artifacts/${sectionName}`;
  
  const styleDiffs = (compareResult.styles || [])
    .filter(s => s.changed)
    .map(s => ({ property: s.name, left: s.left, right: s.right }));
  
  const attributeDiffs = (compareResult.attributes || [])
    .filter(a => a.changed)
    .map(a => ({ attribute: a.name, left: a.left || null, right: a.right || null }));

  return {
    page: pageName,
    section: sectionName,
    classification,
    changedPercent: pct,
    changedPixels: pixel.changedPixels || 0,
    totalPixels: pixel.totalPixels || 0,
    threshold: pixel.threshold || spec.defaults.threshold,
    variant: spec.variant || "desktop",
    diffOnlyPath: `${artifactDir}/diff_only.png`,
    diffComparisonPath: `${artifactDir}/diff_comparison.png`,
    leftRegionPath: `${artifactDir}/left_region.png`,
    rightRegionPath: `${artifactDir}/right_region.png`,
    artifactJson: `${artifactDir}/compare.json`,
    leftSelector: compareResult.left?.selector || "",
    rightSelector: compareResult.right?.selector || "",
    styleChangeCount: styleDiffs.length,
    attributeChangeCount: attributeDiffs.length,
    styleDiffs,
    attributeDiffs,
    bounds: compareResult.bounds || {},
    text: compareResult.text,
  };
}
```

### 4.3 Summary assembler

```javascript
function buildSummary(rows) {
  const classificationCounts = {};
  for (const row of rows) {
    classificationCounts[row.classification] = (classificationCounts[row.classification] || 0) + 1;
  }

  const pages = [...new Set(rows.map(r => r.page))];
  const maxPct = rows.length > 0 ? Math.max(...rows.map(r => r.changedPercent)) : 0;
  const worstRow = rows.reduce((worst, row) => {
    return row.changedPercent > worst.changedPercent ? row : worst;
  }, rows[0]);

  const policy = {
    ok: !rows.some(r => ["tune-required", "major-mismatch"].includes(r.classification)),
    worstClassification: worstRow ? worstRow.classification : "accepted",
    failureCount: rows.filter(r => ["tune-required", "major-mismatch"].includes(r.classification)).length,
  };

  return {
    classificationCounts,
    pageCount: pages.length,
    sectionCount: rows.length,
    maxChangedPercent: maxPct,
    policy,
    rows,
  };
}
```

### 4.4 Main verb: from-spec (pseudocode)

```javascript
async function fromSpec(spec, sweepOutput) {
  const fs = require("fs");
  const path = require("path");
  const yaml = require("yaml");
  const diff = require("diff");

  // 1. Read and parse spec
  const specText = fs.readFileSync(spec.specFile, "utf8");
  const specObj = yaml.parse(specText);

  // 2. Validate
  const pages = Object.entries(specObj.pages || {});
  if (pages.length === 0) {
    throw new Error("Spec contains no pages");
  }

  const bands = specObj.policy?.bands || [
    { name: "accepted", maxChangedPercent: 0.5 },
    { name: "review", maxChangedPercent: 10 },
    { name: "tune-required", maxChangedPercent: 30 },
    { name: "major-mismatch", maxChangedPercent: 100 },
  ];
  // Sort bands ascending by maxChangedPercent
  bands.sort((a, b) => a.maxChangedPercent - b.maxChangedPercent);

  const outDir = sweepOutput.outDir;
  const writeMd = sweepOutput.writeMarkdown !== false;
  const rows = [];

  // 3. Run comparisons
  for (const [pageName, pageSpec] of pages) {
    const sections = Object.entries(pageSpec.sections || {});
    if (sections.length === 0) {
      console.warn(`Page "${pageName}" has no sections, skipping`);
      continue;
    }

    for (const [sectionName, sectionSpec] of sections) {
      const selector = sectionSpec.selector;
      const leftSelector = sectionSpec.leftSelector || selector;
      const rightSelector = sectionSpec.rightSelector || selector;

      if (!selector) {
        console.warn(`Section "${pageName}/${sectionName}" has no selector, skipping`);
        continue;
      }

      const artifactDir = path.join(outDir, pageName, "artifacts", sectionName);
      fs.mkdirSync(artifactDir, { recursive: true });

      console.log(`Comparing ${pageName}/${sectionName}...`);

      const result = diff.compareRegion({
        left: {
          url: pageSpec.leftUrl,
          selector: leftSelector,
          waitMs: pageSpec.leftWaitMs || specObj.defaults?.waitMs || 1000,
        },
        right: {
          url: pageSpec.rightUrl,
          selector: rightSelector,
          waitMs: pageSpec.rightWaitMs || specObj.defaults?.waitMs || 1000,
        },
        viewport: {
          width: specObj.viewport?.width || 920,
          height: specObj.viewport?.height || 1460,
        },
        output: {
          outDir: artifactDir,
          threshold: specObj.defaults?.threshold || 30,
          writeJson: true,
          writeMarkdown: writeMd,
          writePngs: true,
        },
        computed: specObj.computed,
        attributes: specObj.attributes,
      });

      rows.push(buildRow(pageName, sectionName, result, { ...specObj, _outDir: outDir, policy: { bands } }));
      console.log(`  → ${result.pixel?.changedPercent?.toFixed(2) || 0}% changed (${rows[rows.length - 1].classification})`);
    }
  }

  // 4. Assemble and write summary
  const summary = buildSummary(rows);
  const summaryPath = path.join(outDir, "summary.json");
  fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));

  console.log(`\nDone: ${rows.length} sections across ${summary.pageCount} pages`);
  console.log(`  max change: ${summary.maxChangedPercent.toFixed(2)}%`);
  console.log(`  policy: ${summary.policy.ok ? "PASS" : "FAIL"} (${summary.policy.worstClassification})`);
  console.log(`  summary: ${summaryPath}`);
  console.log(`\nServe with: css-visual-diff serve --data-dir ${outDir} --port 8098`);

  return summary;
}
```

### 4.5 Utility verb: summary (pseudocode)

```javascript
async function summary(spec, sweepOutput) {
  const fs = require("fs");
  const path = require("path");
  const yaml = require("yaml");

  const specText = fs.readFileSync(spec.specFile, "utf8");
  const specObj = yaml.parse(specText);

  const bands = specObj.policy?.bands || [
    { name: "accepted", maxChangedPercent: 0.5 },
    { name: "review", maxChangedPercent: 10 },
    { name: "tune-required", maxChangedPercent: 30 },
    { name: "major-mismatch", maxChangedPercent: 100 },
  ];
  bands.sort((a, b) => a.maxChangedPercent - b.maxChangedPercent);

  const outDir = sweepOutput.outDir;
  const rows = [];

  // Walk for compare.json files
  const pages = fs.readdirSync(outDir).filter(name => {
    return fs.statSync(path.join(outDir, name)).isDirectory();
  });

  for (const pageName of pages) {
    const artifactsDir = path.join(outDir, pageName, "artifacts");
    if (!fs.existsSync(artifactsDir)) continue;

    const sections = fs.readdirSync(artifactsDir).filter(name => {
      return fs.statSync(path.join(artifactsDir, name)).isDirectory();
    });

    for (const sectionName of sections) {
      const comparePath = path.join(artifactsDir, sectionName, "compare.json");
      if (!fs.existsSync(comparePath)) continue;

      const data = JSON.parse(fs.readFileSync(comparePath, "utf8"));
      const artifactDir = path.join(outDir, pageName, "artifacts", sectionName);
      
      const pct = data.pixel?.changedPercent || 0;
      const classification = classify(pct, bands);

      const styleDiffs = (data.styles || [])
        .filter(s => s.changed)
        .map(s => ({ property: s.name, left: s.left, right: s.right }));
      const attributeDiffs = (data.attributes || [])
        .filter(a => a.changed)
        .map(a => ({ attribute: a.name, left: a.left || null, right: a.right || null }));

      rows.push({
        page: pageName,
        section: sectionName,
        classification,
        changedPercent: pct,
        changedPixels: data.pixel?.changedPixels || 0,
        totalPixels: data.pixel?.totalPixels || 0,
        threshold: data.pixel?.threshold || specObj.defaults?.threshold || 30,
        variant: specObj.variant || "desktop",
        diffOnlyPath: path.join(artifactDir, "diff_only.png"),
        diffComparisonPath: path.join(artifactDir, "diff_comparison.png"),
        leftRegionPath: path.join(artifactDir, "left_region.png"),
        rightRegionPath: path.join(artifactDir, "right_region.png"),
        artifactJson: comparePath,
        leftSelector: data.left?.selector || "",
        rightSelector: data.right?.selector || "",
        styleChangeCount: styleDiffs.length,
        attributeChangeCount: attributeDiffs.length,
        styleDiffs,
        attributeDiffs,
        bounds: data.bounds || {},
        text: data.text,
      });
    }
  }

  const summaryObj = buildSummary(rows);
  const summaryPath = path.join(outDir, "summary.json");
  fs.writeFileSync(summaryPath, JSON.stringify(summaryObj, null, 2));

  console.log(`Rebuilt summary: ${rows.length} rows → ${summaryPath}`);
  return summaryObj;
}
```

## 5. Module Availability — No Go Changes Needed

The css-visual-diff verb runtime already includes all required modules:

| Need | Module | How it's available |
|---|---|---|
| Read YAML spec | `require("yaml")` | `yaml` module registered via `DefaultRegistryModules()` |
| Read/write JSON | `require("fs")` | `fs` module registered via `DefaultRegistryModules()` |
| Join paths | `require("path")` | `path` module registered via `DefaultRegistryModules()` |
| Create directories | `require("fs").mkdirSync(..., {recursive: true})` | `fs` supports recursive mkdir |
| Run comparisons | `require("diff").compareRegion({...})` | Registered in `registrar.go` as runtime module |
| Write summary | `require("fs").writeFileSync(...)` | `fs` module |
| Generate UUID for run | `require("crypto").randomUUID()` | `crypto` is a default data-only module |

The Go runtime factory at `dsl/host.go:NewRuntimeFactory()` calls `engine.DefaultRegistryModules()` which runs `modules.EnableAll(reg)`. This registers every module that has called `modules.Register()` via `init()`, including `fs`, `yaml`, `os`, `exec`, `crypto`, `path`, `events`, `time`, `timer`, and `database`.

**No changes to Go code are required.** The verb is purely a JavaScript file in `examples/verbs/`.

## 6. compareRegion() Return Value

The `diff.compareRegion()` function returns a `CompareResult` object. Understanding its shape is critical for building summary rows from it.

The Go function `toPlainValue(result)` converts the Go `modes.CompareResult` struct into a plain JS object. Based on the Go struct definition in `modes.CompareSettings` and the conversion logic, the return value has this shape:

```typescript
interface CompareResult {
  schemaVersion: string;
  name: string;
  left: {
    name: string;
    selector: string;
    url: string;
    bounds: { height: number; width: number; x: number; y: number };
    exists: boolean;
    visible: boolean;
  };
  right: {
    name: string;
    selector: string;
    url: string;
    bounds: { height: number; width: number; x: number; y: number };
    exists: boolean;
    visible: boolean;
  };
  pixel: {
    changedPercent: number;
    changedPixels: number;
    totalPixels: number;
    normalizedWidth: number;
    normalizedHeight: number;
    threshold: number;
    diffOnlyPath: string;
    diffComparisonPath: string;
  };
  bounds: {
    changed: boolean;
    delta: { height: number; width: number; x: number; y: number };
    left: { height: number; width: number; x: number; y: number };
    right: { height: number; width: number; x: number; y: number };
  };
  styles: {
    changed: boolean;
    name: string;
    left: string;
    right: string;
  }[];
  attributes: {
    changed: boolean;
    name: string;
    left?: string | null;
    right?: string | null;
  }[];
  text: {
    changed: boolean;
    left: string;
    right: string;
  };
  artifacts: {
    kind: string;
    name: string;
    path: string;
  }[];
}
```

The `from-spec` verb maps each `CompareResult` into a `SummaryRow` using the `buildRow()` function described in section 4.2.

## 7. Directory Output

After running the verb, the output directory has this exact structure:

```
<outDir>/
├── summary.json                           ← assembled summary manifest
├── about/
│   └── artifacts/
│       ├── content/
│       │   ├── compare.json               ← from diff.compareRegion()
│       │   ├── compare.md                 ← (if writeMarkdown: true)
│       │   ├── left_region.png
│       │   ├── right_region.png
│       │   ├── diff_only.png
│       │   └── diff_comparison.png
│       └── header/
│           └── ...
├── pricing/
│   └── artifacts/
│       └── cards/
│           └── ...
└── ...
```

This is precisely the layout expected by `css-visual-diff serve --data-dir <outDir>`. The `summary.json` at the root is loaded by the `/api/manifest` endpoint. Each `compare.json` is loaded by `/api/compare?page=X&section=Y`. Each PNG is served by `/artifacts/X/Y/Z.png`.

## 8. Usage Examples

### Running a full sweep from a spec

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep from-spec \
  --spec-file my-project.spec.yaml \
  --out-dir /tmp/my-project-review
```

### Rebuilding summary after editing a compare.json

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep summary \
  --spec-file my-project.spec.yaml \
  --out-dir /tmp/my-project-review
```

### Serving the result

```bash
css-visual-diff serve --data-dir /tmp/my-project-review --port 8098
```

### With a CI check

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep from-spec \
  --spec-file my-project.spec.yaml \
  --out-dir /tmp/my-project-review \
  --output json | jq '.policy.ok'
```

## 9. Task Breakdown

| # | Task | Depends on | Estimate |
|---|---|---|---|
| 1 | Create `examples/verbs/review-sweep.js` with package, sections, `from-spec` verb | — | 2h |
| 2 | Implement `classify()` and `buildRow()` helpers | 1 | 30m |
| 3 | Implement `buildSummary()` helper | 2 | 30m |
| 4 | Implement `from-spec` main flow with `diff.compareRegion()` loop | 1-3 | 1h |
| 5 | Implement `summary` verb (directory walk, JSON read, summary rebuild) | 2-3 | 1h |
| 6 | Write example spec YAML file (`examples/specs/review-sweep.example.yaml`) | 1 | 30m |
| 7 | Test with Pyxis data or a local test server | 1-6 | 1h |
| 8 | Update `review-site-data-spec` help topic to reference the verb | 7 | 15m |

Total estimate: ~6-7 hours.

## 10. Open Questions

1. **Error handling strategy for missing pages:** If `diff.compareRegion()` fails for one section (e.g., selector not found, page timeout), should the verb skip that section and continue, or abort? Recommendation: skip and record a failure row with `classification: "error"` and `changedPercent: -1`. This lets the reviewer see which sections failed in the review site.

2. **Parallel comparisons:** `diff.compareRegion()` is synchronous in the goja runtime. For large specs (20+ sections), sequential processing could take 10+ minutes. A future optimization would be to process independent pages in parallel using multiple runtime instances, but this requires Go-level changes to the verb invocation pipeline. Not in scope for this first implementation.

3. **Config reuse:** The YAML spec format is similar to (but simpler than) the css-visual-diff `run` command's YAML config. Should we support loading `run`-style configs too, or keep the spec format separate? Recommendation: keep them separate for now. The `run` config has many more features (modes, capture plans) that are irrelevant for the review sweep. The review spec is intentionally minimal.

4. **Left/right wait override per section:** The spec supports per-page wait overrides but not per-section. This is intentional — section-level waits would complicate the spec for marginal benefit. If needed, users can split a page into multiple entries with different wait times.

## 11. See Also

- `css-visual-diff help review-site-data-spec` — full data format specification
- `css-visual-diff help review-site` — review site user guide
- `css-visual-diff help javascript-verbs` — verb API reference (if it exists)
- `css-visual-diff help compare` — compare command reference
- `css-visual-diff help run` — run command with YAML configs
- go-go-goja `yaml` module docs — YAML parse/stringify API
- go-go-goja `nodejs-primitives` docs — fs/path/crypto APIs
