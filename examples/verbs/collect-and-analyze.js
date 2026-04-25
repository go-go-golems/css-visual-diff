// Example collect-and-analyze css-visual-diff verb.
// Run with:
//   css-visual-diff verbs --repository examples/verbs examples compare collect-and-analyze \
//     http://127.0.0.1:8767/left.html http://127.0.0.1:8767/right.html '#cta' /tmp/cssvd-collect-analyze --output json

async function collectAndAnalyze(leftUrl, rightUrl, selector, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage;
  let rightPage;
  try {
    const viewport = cvd.viewport(values.width || 800, values.height || 600);
    leftPage = await browser.page(leftUrl, { viewport, waitMs: values.waitMs || 0, name: "left" });
    rightPage = await browser.page(rightUrl, { viewport, waitMs: values.waitMs || 0, name: "right" });

    const collectOptions = {
      inspect: values.inspect || "rich",
      styles: ["color", "background-color", "font-size", "font-weight", "line-height", "padding", "border-radius"],
      attributes: ["id", "class", "aria-label"],
    };
    const left = await leftPage.locator(selector).collect(collectOptions);
    const right = await cvd.collect.selection(rightPage.locator(values.rightSelector || selector), collectOptions);

    const comparison = await cvd.compare.selections(left, right, {
      name: values.name || "collect-and-analyze",
      threshold: values.threshold || 30,
      styleProps: collectOptions.styles,
      attributes: collectOptions.attributes,
    });

    const styleDiffs = comparison.styles.diff();
    const typographyDiffs = comparison.styles.diff(["font-size", "font-weight", "line-height"]);
    const classDiffs = comparison.attributes.diff(["class"]);
    await comparison.artifacts.write(outDir, ["json", "markdown"]);
    const reportPath = `${outDir}/compare.md`;
    const jsonPath = `${outDir}/compare.json`;

    return {
      ok: true,
      schemaVersion: comparison.toJSON().schemaVersion,
      leftExists: left.summary().exists,
      rightExists: right.summary().exists,
      boundsChanged: comparison.bounds.diff().changed,
      styleChangeCount: styleDiffs.length,
      typographyChangeCount: typographyDiffs.length,
      classChanged: classDiffs.length > 0,
      reportPath,
      jsonPath,
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}

__verb__("collectAndAnalyze", {
  parents: ["examples", "compare"],
  short: "Collect selector facts, compare them, and return custom analysis",
  fields: {
    leftUrl: { argument: true, required: true, help: "Left page URL" },
    rightUrl: { argument: true, required: true, help: "Right page URL" },
    selector: { argument: true, required: true, help: "CSS selector on the left page" },
    outDir: { argument: true, required: true, help: "Directory where analysis outputs are written" },
    values: { bind: "all" },
    rightSelector: { type: "string", help: "CSS selector on the right page; defaults to selector" },
    name: { type: "string", default: "collect-and-analyze", help: "Comparison name" },
    inspect: { type: "string", default: "rich", help: "Collection profile: minimal, rich, or debug" },
    threshold: { type: "int", default: 30, help: "Pixel threshold for future screenshot-backed comparisons" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation in ms" },
    width: { type: "int", default: 800, help: "Viewport width" },
    height: { type: "int", default: 600, help: "Viewport height" }
  }
});
