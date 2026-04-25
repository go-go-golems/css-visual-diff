// Example lower-level css-visual-diff verb.
// Run with:
//   css-visual-diff verbs --repository examples/verbs examples low-level inspect \
//     http://127.0.0.1:8767/ '#cta' /tmp/cssvd-low-level --output json

async function inspect(url, selector, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, {
      viewport: cvd.viewport(values.width || 800, values.height || 600),
      waitMs: values.waitMs || 0,
      name: values.name || "low-level-inspect",
    });

    const locator = page.locator(selector);
    const element = await cvd.extract(locator, [
      cvd.extractors.exists(),
      cvd.extractors.visible(),
      cvd.extractors.text(),
      cvd.extractors.bounds(),
      cvd.extractors.computedStyle(["display", "color", "background-color", "font-size", "font-weight", "line-height"]),
      cvd.extractors.attributes(["id", "class", "aria-label"]),
    ]);

    const snapshot = await cvd.snapshot(page, [
      cvd.probe(values.name || "target")
        .selector(selector)
        .required()
        .text()
        .bounds()
        .styles(["display", "color", "background-color", "font-size", "font-weight", "line-height"])
        .attributes(["id", "class", "aria-label"]),
    ]);

    const elementPath = `${outDir}/element.json`;
    const snapshotPath = `${outDir}/snapshot.json`;
    await cvd.write.json(elementPath, element);
    await cvd.write.json(snapshotPath, snapshot);

    return {
      ok: !!element.exists,
      selector,
      visible: !!element.visible,
      text: element.text || "",
      color: element.computed ? element.computed.color : "",
      elementPath,
      snapshotPath,
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}

__verb__("inspect", {
  parents: ["examples", "low-level"],
  short: "Inspect one selector with lower-level locator/extractor/snapshot APIs",
  fields: {
    url: { argument: true, required: true, help: "URL to inspect" },
    selector: { argument: true, required: true, help: "CSS selector to inspect" },
    outDir: { argument: true, required: true, help: "Directory where JSON outputs are written" },
    values: { bind: "all" },
    name: { type: "string", default: "target", help: "Name for the generated probe" },
    waitMs: { type: "int", default: 0, help: "Wait after navigation in ms" },
    width: { type: "int", default: 800, help: "Viewport width" },
    height: { type: "int", default: 600, help: "Viewport height" }
  }
});
