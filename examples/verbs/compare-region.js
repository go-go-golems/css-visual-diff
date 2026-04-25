// Example compare-region css-visual-diff verb.
// Run with:
//   css-visual-diff verbs --repository examples/verbs examples compare region \
//     http://127.0.0.1:8767/left.html http://127.0.0.1:8767/right.html '#cta' /tmp/cssvd-compare-region --output json

async function region(leftUrl, rightUrl, selector, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage;
  let rightPage;
  try {
    const viewport = cvd.viewport(values.width || 800, values.height || 600);
    leftPage = await browser.page(leftUrl, { viewport, waitMs: values.waitMs || 0, name: "left" });
    rightPage = await browser.page(rightUrl, { viewport, waitMs: values.waitMs || 0, name: "right" });

    const comparison = await cvd.compare.region({
      name: values.name || "region",
      left: leftPage.locator(selector),
      right: rightPage.locator(values.rightSelector || selector),
      outDir,
      threshold: values.threshold || 30,
      inspect: values.inspect || "rich",
      styleProps: ["display", "color", "background-color", "font-size", "font-weight", "line-height", "padding", "border-radius"],
      attributes: ["id", "class", "aria-label"],
    });

    await comparison.artifacts.write(outDir, ["json", "markdown"]);
    const summary = comparison.summary();
    return {
      ok: true,
      schemaVersion: comparison.toJSON().schemaVersion,
      name: summary.name,
      changedPercent: summary.pixel ? summary.pixel.changedPercent : 0,
      changedPixels: summary.pixel ? summary.pixel.changedPixels : 0,
      boundsChanged: summary.boundsChanged,
      styleChanges: summary.styleChanges,
      attributeChanges: summary.attributeChanges,
      artifacts: comparison.artifacts.list(),
      outDir,
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}

__verb__("region", {
  parents: ["examples", "compare"],
  short: "Compare one selector across two pages using cvd.compare.region",
  fields: {
    leftUrl: { argument: true, required: true, help: "Left page URL" },
    rightUrl: { argument: true, required: true, help: "Right page URL" },
    selector: { argument: true, required: true, help: "CSS selector on the left page" },
    outDir: { argument: true, required: true, help: "Directory where artifacts are written" },
    values: { bind: "all" },
    rightSelector: { type: "string", help: "CSS selector on the right page; defaults to selector" },
    name: { type: "string", default: "region", help: "Comparison name" },
    inspect: { type: "string", default: "rich", help: "Collection profile: minimal, rich, or debug" },
    threshold: { type: "int", default: 30, help: "Pixel threshold" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation in ms" },
    width: { type: "int", default: 800, help: "Viewport width" },
    height: { type: "int", default: 600, help: "Viewport height" }
  }
});
