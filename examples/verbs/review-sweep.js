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
    leftRegionPath: path.join(artifactDir, "left_region.png"),
    rightRegionPath: path.join(artifactDir, "right_region.png"),
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
 * compareRegion() writes url1_screenshot.png/url2_screenshot.png. The review
 * site spec and older pipelines use left_region.png/right_region.png. Copy
 * aliases so generated directories match the published review-site contract.
 */
function ensureReviewSiteArtifactAliases(artifactDir) {
  var fs = require("fs");
  var path = require("path");

  var aliases = [
    ["url1_screenshot.png", "left_region.png"],
    ["url2_screenshot.png", "right_region.png"],
  ];

  for (var i = 0; i < aliases.length; i++) {
    var src = path.join(artifactDir, aliases[i][0]);
    var dst = path.join(artifactDir, aliases[i][1]);
    if (fs.existsSync(src) && !fs.existsSync(dst)) {
      fs.copyFileSync(src, dst);
    }
  }
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

// ── Verb: from-spec ─────────────────────────────────────────────────────────

/**
 * Read a YAML spec, run diff.compareRegion() for each page/section,
 * write artifacts and summary.json to disk.
 */
async function fromSpec(spec, sweepOutput) {
  var fs = require("fs");
  var pathMod = require("path");
  var yaml = require("yaml");
  var diff = require("diff");

  // 1. Read and parse spec
  var specText = fs.readFileSync(spec.specFile, "utf8");
  var specObj = yaml.parse(specText);

  // 2. Validate
  var pageEntries = Object.entries(specObj.pages || {});
  if (pageEntries.length === 0) {
    throw new Error("Spec contains no pages");
  }

  var bands = resolveBands(specObj);
  var computedProps = specObj.computed || DEFAULT_COMPUTED;
  var attrProps = specObj.attributes || ["id", "class"];
  var waitMs = (specObj.defaults && specObj.defaults.waitMs) || 1000;
  var threshold = (specObj.defaults && specObj.defaults.threshold) || 30;
  var vpWidth = (specObj.viewport && specObj.viewport.width) || 920;
  var vpHeight = (specObj.viewport && specObj.viewport.height) || 1460;

  var outDir = sweepOutput.outDir;
  var writeMd = sweepOutput.writeMarkdown !== false;
  var failFast = sweepOutput.failFast === true;

  var rows = [];
  var errors = [];

  // 3. Run comparisons
  for (var pi = 0; pi < pageEntries.length; pi++) {
    var pageName = pageEntries[pi][0];
    var pageSpec = pageEntries[pi][1];
    var sectionEntries = Object.entries(pageSpec.sections || {});

    if (sectionEntries.length === 0) {
      console.warn("Page \"" + pageName + "\" has no sections, skipping");
      continue;
    }

    for (var si = 0; si < sectionEntries.length; si++) {
      var sectionName = sectionEntries[si][0];
      var sectionSpec = sectionEntries[si][1];
      var selector = sectionSpec.selector;

      if (!selector) {
        console.warn("Section \"" + pageName + "/" + sectionName + "\" has no selector, skipping");
        continue;
      }

      var leftSelector = sectionSpec.leftSelector || selector;
      var rightSelector = sectionSpec.rightSelector || selector;
      var leftWait = sectionSpec.leftWaitMs || pageSpec.leftWaitMs || waitMs;
      var rightWait = sectionSpec.rightWaitMs || pageSpec.rightWaitMs || waitMs;

      var artifactDir = pathMod.join(outDir, pageName, "artifacts", sectionName);
      fs.mkdirSync(artifactDir, { recursive: true });

      console.log("Comparing " + pageName + "/" + sectionName + "...");

      try {
        var result = diff.compareRegion({
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
        });

        ensureReviewSiteArtifactAliases(artifactDir);

        var row = buildRowFromCompareResult(pageName, sectionName, result, {
          defaults: { threshold: threshold },
          variant: specObj.variant,
          policy: { bands: bands },
        }, outDir);
        rows.push(row);
        console.log("  -> " + row.changedPercent.toFixed(2) + "% changed (" + row.classification + ")");
      } catch (err) {
        var errMsg = err && err.message ? err.message : String(err);
        console.error("  ERROR: " + errMsg);
        errors.push({ page: pageName, section: sectionName, error: errMsg });

        if (failFast) {
          throw err;
        }

        // Record a failure row so the reviewer can see it
        rows.push({
          page: pageName,
          section: sectionName,
          classification: "error",
          changedPercent: -1,
          changedPixels: 0,
          totalPixels: 0,
          threshold: threshold,
          variant: specObj.variant || "desktop",
          diffOnlyPath: "",
          diffComparisonPath: "",
          leftRegionPath: "",
          rightRegionPath: "",
          artifactJson: "",
          leftSelector: leftSelector,
          rightSelector: rightSelector,
          styleChangeCount: 0,
          attributeChangeCount: 0,
          styleDiffs: [],
          attributeDiffs: [],
          bounds: {},
          error: errMsg,
        });
      }
    }
  }

  // 4. Assemble and write summary
  var summary = buildSummary(rows);
  var summaryPath = pathMod.join(outDir, "summary.json");
  fs.mkdirSync(outDir, { recursive: true });
  fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));

  console.log("");
  console.log("Done: " + rows.length + " sections across " + summary.pageCount + " pages");
  if (errors.length > 0) {
    console.log("Errors: " + errors.length + " sections failed");
  }
  console.log("  max change: " + summary.maxChangedPercent.toFixed(2) + "%");
  console.log("  policy: " + (summary.policy.ok ? "PASS" : "FAIL") + " (" + summary.policy.worstClassification + ")");
  console.log("  summary: " + summaryPath);
  console.log("");
  console.log("Serve with: css-visual-diff serve --data-dir " + outDir + " --port 8098");

  return summary;
}

__verb__("fromSpec", {
  short: "Run visual comparisons from a YAML spec and produce a review site data directory",
  fields: {
    spec: { bind: "spec" },
    sweepOutput: { bind: "sweepOutput" },
  },
});

// ── Verb: summary ───────────────────────────────────────────────────────────

/**
 * Walk an existing data directory, read compare.json files,
 * and rebuild summary.json. Useful when the summary is lost
 * or policy bands need to be re-applied.
 *
 * Handles two compare.json formats:
 *  1. Catalog/inspect pipeline: camelCase (pixel.changedPercent, styles, left/right)
 *  2. diff.compareRegion() output: snake_case (pixel_diff.changed_percent, computed_diffs, url1/url2)
 */
function summary(spec, sweepOutput) {
  var fs = require("fs");
  var pathMod = require("path");
  var yaml = require("yaml");

  var specText = fs.readFileSync(spec.specFile, "utf8");
  var specObj = yaml.parse(specText);

  var bands = resolveBands(specObj);
  var threshold = (specObj.defaults && specObj.defaults.threshold) || 30;

  var outDir = sweepOutput.outDir;
  var rows = [];

  // Walk for compare.json files
  var pageNames = fs.readdirSync(outDir).filter(function (name) {
    try {
      return fs.statSync(pathMod.join(outDir, name)).isDir === true;
    } catch (e) {
      return false;
    }
  });

  for (var pi = 0; pi < pageNames.length; pi++) {
    var pageName = pageNames[pi];
    var artifactsDir = pathMod.join(outDir, pageName, "artifacts");

    var sectionNames = [];
    try {
      sectionNames = fs.readdirSync(artifactsDir).filter(function (name) {
        try {
          return fs.statSync(pathMod.join(artifactsDir, name)).isDir === true;
        } catch (e) {
          return false;
        }
      });
    } catch (e) {
      continue;
    }

    for (var si = 0; si < sectionNames.length; si++) {
      var sectionName = sectionNames[si];
      var comparePath = pathMod.join(artifactsDir, sectionName, "compare.json");

      if (!fs.existsSync(comparePath)) {
        continue;
      }

      try {
        var data = JSON.parse(fs.readFileSync(comparePath, "utf8"));
        var artifactDir = pathMod.join(outDir, pageName, "artifacts", sectionName);

        var row = buildRowFromCompareJson(pageName, sectionName, data, bands, threshold, artifactDir, specObj);
        rows.push(row);
      } catch (err) {
        console.warn("Failed to read " + comparePath + ": " + (err && err.message ? err.message : String(err)));
      }
    }
  }

  var summaryObj = buildSummary(rows);
  var summaryPath = pathMod.join(outDir, "summary.json");
  fs.writeFileSync(summaryPath, JSON.stringify(summaryObj, null, 2));

  console.log("Rebuilt summary: " + rows.length + " rows -> " + summaryPath);
  console.log("  policy: " + (summaryObj.policy.ok ? "PASS" : "FAIL") + " (" + summaryObj.policy.worstClassification + ")");

  return summaryObj;
}

/**
 * Build a row from a compare.json on disk.
 * Handles both camelCase (catalog/inspect) and snake_case (compareRegion) formats.
 */
function buildRowFromCompareJson(pageName, sectionName, data, bands, defaultThreshold, artifactDir, spec) {
  var fs = require("fs");
  var pathMod = require("path");

  function artifactPath(preferred, fallback) {
    var preferredPath = pathMod.join(artifactDir, preferred);
    if (fs.existsSync(preferredPath) || !fallback) {
      return preferredPath;
    }
    return pathMod.join(artifactDir, fallback);
  }

  // Detect format
  var isSnakeCase = !!data.pixel_diff;

  var pct, changedPixels, totalPixels, thresh;
  var leftSelector, rightSelector;
  var styleDiffs, attributeDiffs, boundsObj;

  if (isSnakeCase) {
    // Format from diff.compareRegion() JSON on disk
    var pd = data.pixel_diff || {};
    pct = pd.changed_percent || 0;
    changedPixels = pd.changed_pixels || 0;
    totalPixels = pd.total_pixels || 0;
    thresh = pd.threshold || defaultThreshold;
    leftSelector = (data.inputs && data.inputs.selector1) || "";
    rightSelector = (data.inputs && data.inputs.selector2) || "";

    var cDiffs = data.computed_diffs || [];
    styleDiffs = cDiffs
      .filter(function (d) { return d.changed; })
      .map(function (d) { return { property: d.property, left: d.left || "", right: d.right || "" }; });

    // Attributes from computed snapshots
    var lAttrs = (data.url1 && data.url1.computed && data.url1.computed.attributes) || {};
    var rAttrs = (data.url2 && data.url2.computed && data.url2.computed.attributes) || {};
    var aKeys = Object.keys(Object.assign({}, lAttrs, rAttrs));
    attributeDiffs = aKeys
      .filter(function (k) { return lAttrs[k] !== rAttrs[k]; })
      .map(function (k) { return { attribute: k, left: lAttrs[k] || null, right: rAttrs[k] || null }; });

    boundsObj = {};
  } else {
    // Format from catalog/inspect pipeline (camelCase)
    var pixel = data.pixel || {};
    pct = pixel.changedPercent || 0;
    changedPixels = pixel.changedPixels || 0;
    totalPixels = pixel.totalPixels || 0;
    thresh = pixel.threshold || defaultThreshold;
    leftSelector = (data.left && data.left.selector) || "";
    rightSelector = (data.right && data.right.selector) || "";

    var styles = data.styles || [];
    styleDiffs = styles
      .filter(function (s) { return s.changed; })
      .map(function (s) { return { property: s.name, left: s.left, right: s.right }; });

    var attrs = data.attributes || [];
    attributeDiffs = attrs
      .filter(function (a) { return a.changed; })
      .map(function (a) { return { attribute: a.name, left: a.left || null, right: a.right || null }; });

    boundsObj = data.bounds || {};
  }

  var classification = classify(pct, bands);

  return {
    page: pageName,
    section: sectionName,
    classification: classification,
    changedPercent: pct,
    changedPixels: changedPixels,
    totalPixels: totalPixels,
    threshold: thresh,
    variant: (spec && spec.variant) || "desktop",
    diffOnlyPath: pathMod.join(artifactDir, "diff_only.png"),
    diffComparisonPath: pathMod.join(artifactDir, "diff_comparison.png"),
    leftRegionPath: isSnakeCase ? artifactPath("left_region.png", "url1_screenshot.png") : pathMod.join(artifactDir, "left_region.png"),
    rightRegionPath: isSnakeCase ? artifactPath("right_region.png", "url2_screenshot.png") : pathMod.join(artifactDir, "right_region.png"),
    artifactJson: pathMod.join(artifactDir, "compare.json"),
    leftSelector: leftSelector,
    rightSelector: rightSelector,
    styleChangeCount: styleDiffs.length,
    attributeChangeCount: attributeDiffs.length,
    styleDiffs: styleDiffs,
    attributeDiffs: attributeDiffs,
    bounds: boundsObj,
    text: data.text || undefined,
  };
}

__verb__("summary", {
  short: "Rebuild summary.json from existing compare.json artifacts on disk",
  fields: {
    spec: { bind: "spec" },
    sweepOutput: { bind: "sweepOutput" },
  },
});
