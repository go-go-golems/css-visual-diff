// Overlay screenshot examples.
//
// Serve the fixture page from the repository root with:
//   python3 -m http.server 8767
//
// Then run, for example:
//   css-visual-diff verbs --repository examples/verbs examples overlay annotated-png \
//     http://127.0.0.1:8767/examples/pages/overlay-components.html /tmp/cssvd-overlay --output json

const fs = require("fs");
const path = require("path");
const cvd = require("css-visual-diff");

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function targetMapToSpec(map, options) {
  options = options || {};
  const builder = cvd.overlaySpec()
    .legend(options.legend !== false)
    .screenshot("fullPage")
    .style(options.style || {
      label: { fontSize: 13, radius: 3, padding: [4, 7] },
      legend: { position: "bottom-right", background: "rgba(255,255,255,0.92)", color: "#27221b" },
      targetDefaults: { borderWidth: 2, labelColor: "white" },
    });

  if (options.cropTo) builder.cropTo(options.cropTo);
  if (options.cropPadding !== undefined) builder.cropPadding(options.cropPadding);

  const styleByName = options.styleByName || {};
  for (const [name, selector] of Object.entries(map)) {
    const target = cvd.overlayTarget(name).selector(selector);
    if (styleByName[name]) target.style(styleByName[name]);
    builder.target(target);
  }
  return builder.build();
}

const organismSelectors = {
  Header: ".site-header",
  Hero: ".hero",
  "Feature Grid": ".features",
  Newsletter: ".newsletter",
  Footer: ".site-footer",
};

const organismStyles = {
  Header: { borderColor: "#0096ff", label: { background: "#0096ff" } },
  Hero: { borderColor: "#ff6347", label: { background: "#ff6347" } },
  "Feature Grid": { borderColor: "#32cd32", label: { background: "#32cd32" } },
  Newsletter: { borderColor: "#ba55d3", label: { background: "#ba55d3" } },
  Footer: { borderColor: "#ff8c00", label: { background: "#ff8c00" } },
};

const heroPartSelectors = {
  Hero: ".hero",
  Eyebrow: ".hero .eyebrow",
  Headline: ".hero h1",
  Copy: ".hero .lede",
  Actions: ".hero-actions",
  Media: ".hero-media",
};

const componentProbes = [
  cvd.probe("Logo").selector(".site-logo").styles(["display", "font-size", "font-weight", "color"]).attributes(["class"]).build(),
  cvd.probe("Primary Button").selector(".button-primary").styles(["display", "background-color", "color", "border-radius"]).attributes(["class"]).build(),
  cvd.probe("Secondary Button").selector(".button-secondary").styles(["display", "background-color", "color", "border-radius"]).attributes(["class"]).build(),
  cvd.probe("Feature Card").selector(".feature-card:first-child").styles(["display", "padding", "background-color", "border-radius"]).attributes(["class"]).build(),
  cvd.probe("Newsletter Form").selector(".newsletter-form").styles(["display", "gap", "max-width"]).attributes(["class"]).build(),
];

async function withPage(url, values, fn) {
  values = values || {};
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, {
      viewport: cvd.viewport(values.width || 1280, values.height || 1400),
      waitMs: values.waitMs || 250,
      name: values.name || "overlay-example",
    });
    await page.css(`html { scroll-behavior: auto !important; }`);
    return await fn(page);
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}

async function annotatedPng(url, outDir, values) {
  ensureDir(outDir);
  return withPage(url, values, async (page) => {
    const outPath = path.join(outDir, "full-page.organisms.annotated.png");
    const result = await page
      .overlay(targetMapToSpec(organismSelectors, { styleByName: organismStyles }))
      .screenshot(outPath);
    return { ok: true, kind: "annotated-png", outPath: result.outputPath, width: result.width, height: result.height, targetCount: result.targets.length, colors: result.colors };
  });
}

async function componentGallery(url, outDir, values) {
  ensureDir(outDir);
  ensureDir(path.join(outDir, "annotated"));
  return withPage(url, values, async (page) => {
    const inspect = await page.inspectAll(componentProbes, {
      outDir: path.join(outDir, "components"),
      artifacts: "bundle",
    });

    const fullMap = await page
      .overlay(targetMapToSpec(organismSelectors, { styleByName: organismStyles }))
      .screenshot(path.join(outDir, "annotated", "full-page.organisms.png"));

    const heroParts = await page
      .overlay(targetMapToSpec(heroPartSelectors, {
        legend: false,
        cropTo: ".hero",
        cropPadding: 24,
        styleByName: {
          Hero: { borderColor: "#27221b", label: { background: "#27221b" } },
          Headline: { borderColor: "#ff6347", label: { background: "#ff6347" } },
          Actions: { borderColor: "#0096ff", label: { background: "#0096ff" } },
          Media: { borderColor: "#32cd32", label: { background: "#32cd32" } },
        },
      }))
      .screenshot(path.join(outDir, "annotated", "hero.parts.crop.png"));

    const model = {
      url,
      generatedAt: new Date().toISOString(),
      components: inspect.results,
      annotated: [
        { name: "Full page organism map", path: fullMap.outputPath, width: fullMap.width, height: fullMap.height },
        { name: "Hero parts crop", path: heroParts.outputPath, width: heroParts.width, height: heroParts.height },
      ],
    };
    fs.writeFileSync(path.join(outDir, "components.json"), JSON.stringify(model, null, 2));
    writeGalleryHtml(outDir, model);
    return { ok: true, kind: "component-gallery", outDir, html: path.join(outDir, "index.html"), components: model.components.length, annotated: model.annotated };
  });
}

function rel(from, to) {
  return path.relative(from, to).replace(/\\/g, "/");
}

function escapeHtml(value) {
  return String(value).replace(/[&<>\"]/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;" }[ch]));
}

function writeGalleryHtml(outDir, model) {
  const annotated = model.annotated.map((item) => `
    <section class="panel">
      <h2>${escapeHtml(item.name)}</h2>
      <p>${item.width}×${item.height}</p>
      <img src="${escapeHtml(rel(outDir, item.path))}" alt="${escapeHtml(item.name)}" />
    </section>`).join("\n");

  const components = model.components.map((item) => {
    const meta = item.metadata || {};
    const screenshot = item.screenshot ? `<img src="${escapeHtml(rel(outDir, item.screenshot))}" alt="${escapeHtml(meta.name || "component")}" />` : "";
    return `<article class="card"><h3>${escapeHtml(meta.name || "component")}</h3><p><code>${escapeHtml(meta.selector || "")}</code></p>${screenshot}</article>`;
  }).join("\n");

  fs.writeFileSync(path.join(outDir, "index.html"), `<!doctype html>
<html><head><meta charset="utf-8"><title>Overlay component gallery</title>
<style>
body{font-family:system-ui,sans-serif;margin:32px;background:#f7f3ea;color:#27221b}img{max-width:100%;height:auto;border:1px solid #ddd3c2;background:white}.panel,.card{background:white;border:1px solid #ddd3c2;border-radius:8px;padding:16px;margin:0 0 20px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:16px}code{background:#f7f3ea;padding:2px 5px;border-radius:4px}
</style></head><body>
<h1>Overlay component gallery</h1>
<p>Source: <code>${escapeHtml(model.url)}</code></p>
<h2>Annotated screenshots</h2>${annotated}
<h2>Extracted components</h2><div class="grid">${components}</div>
</body></html>`);
}

__verb__("annotatedPng", {
  parents: ["examples", "overlay"],
  short: "Export one full-page annotated overlay PNG",
  fields: {
    url: { argument: true, required: true, help: "URL to annotate" },
    outDir: { argument: true, required: true, help: "Output directory" },
    values: { bind: "all" },
    width: { type: "int", default: 1280, help: "Viewport width" },
    height: { type: "int", default: 1400, help: "Viewport height" },
    waitMs: { type: "int", default: 250, help: "Wait after navigation in ms" },
  },
});

async function gallery(url, outDir, values) {
  return componentGallery(url, outDir, values);
}

__verb__("gallery", {
  parents: ["examples", "overlay"],
  short: "Export component screenshots, annotated PNGs, JSON, and an HTML gallery",
  fields: {
    url: { argument: true, required: true, help: "URL to document" },
    outDir: { argument: true, required: true, help: "Output directory" },
    values: { bind: "all" },
    width: { type: "int", default: 1280, help: "Viewport width" },
    height: { type: "int", default: 1400, help: "Viewport height" },
    waitMs: { type: "int", default: 250, help: "Wait after navigation in ms" },
  },
});
