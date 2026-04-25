// Example external verb repository for css-visual-diff.
// Run with:
//   css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
//     http://127.0.0.1:8767/ '#cta' /tmp/cssvd-example --output json

async function inspectPage(url, selector, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const catalog = cvd.catalog.create({
    title: values.title || "Example Visual Catalog",
    outDir,
    artifactRoot: "artifacts",
  });
  const target = {
    slug: values.slug || values.name || selector,
    name: values.name || values.slug || selector,
    url,
    selector,
    viewport: { width: values.width || 800, height: values.height || 600 },
  };
  const probe = {
    name: target.slug,
    selector,
    props: ["display", "color", "background-color", "font-size"],
    source: "examples.catalog.inspectPage",
  };

  catalog.addTarget(target);
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: target.viewport, waitMs: values.waitMs || 0, name: target.name });
    const preflight = await page.preflight([probe]);
    catalog.recordPreflight(target, preflight);
    if (!preflight[0].exists) {
      const err = new cvd.SelectorError(`selector did not match: ${selector}`, "SELECTOR_ERROR", { selector });
      err.operation = "examples.catalog.inspectPage";
      catalog.addFailure(target, err);
      await catalog.writeManifest();
      await catalog.writeIndex();
      if (values.failOnMissing) throw err;
      return { ok: false, slug: target.slug, code: err.code, manifestPath: `${outDir}/manifest.json` };
    }
    const artifactDir = catalog.artifactDir(target.slug);
    const artifact = await page.inspect(probe, { outDir: artifactDir, artifacts: values.artifacts || "css-json" });
    catalog.addResult(target, { outputDir: artifactDir, results: [artifact] });
    const manifestPath = await catalog.writeManifest();
    const indexPath = await catalog.writeIndex();
    return { ok: true, slug: target.slug, artifactDir, manifestPath, indexPath };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}

__verb__("inspectPage", {
  parents: ["examples", "catalog"],
  short: "Example external catalog inspect-page verb",
  fields: {
    url: { argument: true, required: true, help: "URL to inspect" },
    selector: { argument: true, required: true, help: "CSS selector to inspect" },
    outDir: { argument: true, required: true, help: "Output directory" },
    values: { bind: "all" },
    slug: { type: "string", help: "Catalog target slug" },
    name: { type: "string", help: "Catalog target name" },
    title: { type: "string", default: "Example Visual Catalog", help: "Catalog title" },
    artifacts: { type: "choice", choices: ["bundle", "css-json", "html", "png", "inspect-json"], default: "css-json" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation in ms" },
    width: { type: "int", default: 800, help: "Viewport width" },
    height: { type: "int", default: 600, help: "Viewport height" },
    failOnMissing: { type: "bool", default: false, help: "Fail with non-zero exit code when selector is missing" }
  }
});
