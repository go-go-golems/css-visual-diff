// Review sweep verb for css-visual-diff.
// Reads a YAML spec, runs visual comparisons, and produces a data directory
// that can be served directly with `css-visual-diff serve`.
//
// Usage:
//   css-visual-diff verbs --repository examples/verbs \
//     examples review-sweep from-spec \
//     --spec-file my-project.spec.yaml \
//     --out-dir /tmp/my-project-review
//
//   css-visual-diff verbs --repository examples/verbs \
//     examples review-sweep summary \
//     --spec-file my-project.spec.yaml \
//     --out-dir /tmp/my-project-review

__package__({
  name: "review-sweep",
  parents: ["examples"],
  short: "Generate a review site data directory from a YAML spec",
});

// ── Sections ────────────────────────────────────────────────────────────────

__section__("spec", {
  title: "Spec",
  description: "YAML spec file declaring pages and sections to compare.",
  fields: {
    specFile: {
      type: "string",
      required: true,
      help: "Path to the YAML spec file",
    },
  },
});

__section__("sweepOutput", {
  title: "Output",
  description: "Output directory for the review site data.",
  fields: {
    outDir: {
      type: "string",
      required: true,
      help: "Output directory (will be created if needed)",
    },
    writeMarkdown: {
      type: "bool",
      default: true,
      help: "Write compare.md alongside compare.json",
    },
    failFast: {
      type: "bool",
      default: false,
      help: "Abort on first error instead of recording failures and continuing",
    },
  },
});

// ── Helpers ─────────────────────────────────────────────────────────────────

/**
 * Classify a changed percentage into a policy band.
 * Bands must be sorted ascending by maxChangedPercent.
 */
function classify(changedPercent, bands) {
  for (const band of bands) {
    if (changedPercent <= band.maxChangedPercent) {
      return band.name;
    }
  }
  return bands[bands.length - 1].name;
}

/**
 * Resolve policy bands from a spec object, with sensible defaults.
 * Returns bands sorted ascending by maxChangedPercent.
 */
function resolveBands(spec) {
  const raw = (spec.policy && spec.policy.bands) || [
    { name: "accepted", maxChangedPercent: 0.5 },
    { name: "review", maxChangedPercent: 10 },
    { name: "tune-required", maxChangedPercent: 30 },
    { name: "major-mismatch", maxChangedPercent: 100 },
  ];
  return raw.slice().sort((a, b) => a.maxChangedPercent - b.maxChangedPercent);
}

/**
 * Default CSS properties to extract when spec doesn't override.
 */
const DEFAULT_COMPUTED = [
  "display",
  "position",
  "width",
  "height",
  "margin-top",
  "margin-right",
  "margin-bottom",
  "margin-left",
  "padding-top",
  "padding-right",
  "padding-bottom",
  "padding-left",
  "font-family",
  "font-size",
  "font-weight",
  "line-height",
  "color",
  "background-color",
  "background-image",
  "border-radius",
  "box-shadow",
  "z-index",
];

/**
 * Build a SummaryRow from a diff.compareRegion() result.
 *
 * The JS return value from compareRegion uses snake_case JSON tags:
 *   pixel_diff.changed_percent, computed_diffs, url1, url2, etc.
 */
function buildRowFromCompareResult(pageName, sectionName, result, spec, outDir) {
  const path = require("path");

  const pd = result.pixel_diff || {};
  const pct = pd.changed_percent || 0;
  const bands = resolveBands(spec);
  const classification = classify(pct, bands);

  const artifactDir = path.join(outDir, pageName, "artifacts", sectionName);

  // Style diffs come from computed_diffs
  const computedDiffs = result.computed_diffs || [];
  const styleDiffs = computedDiffs
    .filter(function (d) { return d.changed; })
    .map(function (d) {
      return { property: d.property, left: d.left || "", right: d.right || "" };
    });

  // Winner diffs contain matched stylesheet values
  const winnerDiffs = result.winner_diffs || [];

  // Attribute diffs — extract from computed snapshots
  const leftAttrs = (result.url1 && result.url1.computed && result.url1.computed.attributes) || {};
  const rightAttrs = (result.url2 && result.url2.computed && result.url2.computed.attributes) || {};
  const attrKeys = Object.keys(Object.assign({}, leftAttrs, rightAttrs));
  const attributeDiffs = attrKeys
    .filter(function (k) { return leftAttrs[k] !== rightAttrs[k]; })
    .map(function (k) {
      return { attribute: k, left: leftAttrs[k] || null, right: rightAttrs[k] || null };
    });

  // Bounds from computed snapshots
  const leftBounds = (result.url1 && result.url1.computed && result.url1.computed.bounds) || null;
  const rightBounds = (result.url2 && result.url2.computed && result.url2.computed.bounds) || null;
  const boundsChanged = leftBounds && rightBounds &&
    (leftBounds.height !== rightBounds.height || leftBounds.width !== rightBounds.width);

  var boundsObj = {};
  if (leftBounds && rightBounds) {
    boundsObj = {
      changed: !!boundsChanged,
      delta: {
        height: rightBounds.height - leftBounds.height,
        width: rightBounds.width - leftBounds.width,
        x: (rightBounds.x || 0) - (leftBounds.x || 0),
        y: (rightBounds.y || 0) - (leftBounds.y || 0),
      },
      left: leftBounds,
      right: rightBounds,
    };
  }

  return {
    page: pageName,
    section: sectionName,
    classification: classification,
    changedPercent: pct,
    changedPixels: pd.changed_pixels || 0,
    totalPixels: pd.total_pixels || 0,
    threshold: pd.threshold || spec.defaults.threshold || 30,
    variant: spec.variant || "desktop",
    diffOnlyPath: path.join(artifactDir, "diff_only.png"),
    diffComparisonPath: path.join(artifactDir, "diff_comparison.png"),
    leftRegionPath: path.join(artifactDir, "url1_screenshot.png"),
    rightRegionPath: path.join(artifactDir, "url2_screenshot.png"),
    artifactJson: path.join(artifactDir, "compare.json"),
    leftSelector: (result.inputs && result.inputs.selector1) || "",
    rightSelector: (result.inputs && result.inputs.selector2) || "",
    styleChangeCount: styleDiffs.length,
    attributeChangeCount: attributeDiffs.length,
    styleDiffs: styleDiffs,
    attributeDiffs: attributeDiffs,
    bounds: boundsObj,
  };
}

/**
 * Assemble a SuiteSummary from collected rows.
 */
function buildSummary(rows) {
  var classificationCounts = {};
  for (var i = 0; i < rows.length; i++) {
    var cls = rows[i].classification;
    classificationCounts[cls] = (classificationCounts[cls] || 0) + 1;
  }

  var pages = {};
  for (var j = 0; j < rows.length; j++) {
    pages[rows[j].page] = true;
  }
  var pageCount = Object.keys(pages).length;

  var maxPct = 0;
  var worstClassification = "accepted";
  var failureCount = 0;
  for (var k = 0; k < rows.length; k++) {
    if (rows[k].changedPercent > maxPct) {
      maxPct = rows[k].changedPercent;
    }
    if (rows[k].classification === "tune-required" || rows[k].classification === "major-mismatch") {
      failureCount++;
    }
    if (rows[k].classification === "error") {
      failureCount++;
    }
  }
  if (rows.length > 0) {
    worstClassification = rows.reduce(function (worst, row) {
      var order = { accepted: 0, review: 1, "tune-required": 2, "major-mismatch": 3, error: 4 };
      return (order[row.classification] || 0) > (order[worst.classification] || 0) ? row : worst;
    }).classification;
  }

  return {
    classificationCounts: classificationCounts,
    pageCount: pageCount,
    sectionCount: rows.length,
    maxChangedPercent: maxPct,
    policy: {
      ok: failureCount === 0,
      worstClassification: worstClassification,
      failureCount: failureCount,
    },
    rows: rows,
  };
}
