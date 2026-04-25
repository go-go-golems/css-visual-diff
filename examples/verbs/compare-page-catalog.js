// Example multi-section page comparison with a catalog.
// Run with:
//   css-visual-diff verbs --repository examples/verbs examples compare page-catalog \
//     http://127.0.0.1:8767/left.html http://127.0.0.1:8767/right.html /tmp/cssvd-page-catalog --output json

async function pageCatalog(leftUrl, rightUrl, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  const catalog = cvd.catalog.create({
    title: values.title || "Example page comparison",
    outDir,
    artifactRoot: "artifacts",
  });

  const sections = [
    { name: "page", leftSelector: values.pageSelector || "#page", rightSelector: values.rightPageSelector || values.pageSelector || "#page" },
    { name: "cta", leftSelector: values.ctaSelector || "#cta", rightSelector: values.rightCtaSelector || values.ctaSelector || "#cta" },
  ];

  let leftPage;
  let rightPage;
  try {
    const viewport = cvd.viewport(values.width || 920, values.height || 1460);
    leftPage = await browser.page(leftUrl, { viewport, waitMs: values.waitMs || 0, name: "left" });
    rightPage = await browser.page(rightUrl, { viewport, waitMs: values.waitMs || 0, name: "right" });

    const summaries = [];
    for (const section of sections) {
      await leftPage.locator(section.leftSelector).waitFor({ timeoutMs: values.timeoutMs || 30000, visible: true });
      await rightPage.locator(section.rightSelector).waitFor({ timeoutMs: values.timeoutMs || 30000, visible: true });

      const artifactDir = catalog.artifactDir(section.name);
      const comparison = await cvd.compare.region({
        name: section.name,
        left: leftPage.locator(section.leftSelector),
        right: rightPage.locator(section.rightSelector),
        outDir: artifactDir,
        threshold: values.threshold || 30,
        inspect: values.inspect || "rich",
        styleProps: ["display", "color", "background-color", "font-size", "font-weight", "line-height", "padding", "border-radius"],
        attributes: ["id", "class", "aria-label"],
      });

      const written = await comparison.artifacts.write(artifactDir, ["json", "markdown"]);
      catalog.record(comparison, {
        slug: section.name,
        name: section.name,
        url: leftUrl,
        selector: section.leftSelector,
      });

      summaries.push({
        section: section.name,
        changedPercent: comparison.pixel.summary() ? comparison.pixel.summary().changedPercent : 0,
        boundsChanged: comparison.bounds.diff().changed,
        styleChanges: comparison.styles.diff().length,
        attributeChanges: comparison.attributes.diff().length,
        artifacts: written,
      });
    }

    return {
      ok: true,
      summaries,
      manifestPath: await catalog.writeManifest(),
      indexPath: await catalog.writeIndex(),
      catalog: catalog.summary(),
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}

__verb__("pageCatalog", {
  parents: ["examples", "compare"],
  short: "Compare two pages section-by-section and write a catalog",
  fields: {
    leftUrl: { argument: true, required: true, help: "Left/reference page URL" },
    rightUrl: { argument: true, required: true, help: "Right/implementation page URL" },
    outDir: { argument: true, required: true, help: "Catalog output directory" },
    values: { bind: "all" },
    pageSelector: { type: "string", default: "#page", help: "Page/root selector on the left page" },
    rightPageSelector: { type: "string", help: "Page/root selector on the right page" },
    ctaSelector: { type: "string", default: "#cta", help: "CTA selector on the left page" },
    rightCtaSelector: { type: "string", help: "CTA selector on the right page" },
    inspect: { type: "string", default: "rich", help: "Collection profile: minimal, rich, or debug" },
    threshold: { type: "int", default: 30, help: "Pixel threshold" },
    timeoutMs: { type: "int", default: 30000, help: "Selector wait timeout" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation in ms" },
    width: { type: "int", default: 920, help: "Viewport width" },
    height: { type: "int", default: 1460, help: "Viewport height" }
  }
});
