__package__({
  name: "compare",
  parents: ["script"],
  short: "Script-backed comparison workflows"
});

__section__("selectors", {
  title: "Selectors",
  description: "Selectors for left/right target regions.",
  fields: {
    leftSelector: {
      type: "string",
      required: true,
      help: "Selector on the left target"
    },
    rightSelector: {
      type: "string",
      help: "Selector on the right target (defaults to leftSelector)"
    }
  }
});

async function region(targets, viewport, output, selectors) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage;
  let rightPage;
  try {
    leftPage = await browser.page(targets.leftUrl, {
      viewport: { width: viewport.width, height: viewport.height },
      waitMs: targets.leftWaitMs,
      name: "left"
    });
    rightPage = await browser.page(targets.rightUrl, {
      viewport: { width: viewport.width, height: viewport.height },
      waitMs: targets.rightWaitMs,
      name: "right"
    });
    const comparison = await cvd.compare.region({
      name: "region",
      left: leftPage.locator(selectors.leftSelector),
      right: rightPage.locator(selectors.rightSelector || selectors.leftSelector),
      threshold: output.threshold || 30,
      inspect: "rich",
      outDir: output.outDir,
      styleProps: [
        "font-family",
        "font-size",
        "font-weight",
        "line-height",
        "padding-top",
        "padding-right",
        "padding-bottom",
        "padding-left",
        "border-radius",
        "color",
        "background-color",
        "box-shadow"
      ],
      attributes: ["id", "class"]
    });
    const json = comparison.toJSON();
    if (output.writeJson) await cvd.write.json(`${output.outDir}/compare.json`, json);
    if (output.writeMarkdown) await comparison.report.writeMarkdown(`${output.outDir}/compare.md`);
    return json;
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}

__verb__("region", {
  short: "Compare one region across two targets",
  fields: {
    targets: { bind: "targets" },
    viewport: { bind: "viewport" },
    output: { bind: "output" },
    selectors: { bind: "selectors" }
  }
});

async function brief(targets, viewport, output, selectors, question) {
  const result = await region(targets, viewport, output, selectors);
  const lines = [
    `# ${question || "Comparison brief"}`,
    "",
    `- Changed pixels: ${result.pixel ? result.pixel.changedPixels : 0}/${result.pixel ? result.pixel.totalPixels : 0} (${result.pixel ? result.pixel.changedPercent.toFixed(4) : "0.0000"}%)`,
    `- Bounds changed: ${result.bounds.changed}`,
    `- Text changed: ${result.text.changed}`,
    `- Style changes: ${result.styles ? result.styles.length : 0}`,
    `- Attribute changes: ${result.attributes ? result.attributes.length : 0}`,
  ];
  if (result.styles && result.styles.length) {
    lines.push("", "Style diffs:");
    result.styles.slice(0, 8).forEach((diff) => lines.push(`- ${diff.name}: ${diff.left} -> ${diff.right}`));
  }
  return lines.join("\n");
}

__verb__("brief", {
  short: "Render a concise text brief for one compared region",
  output: "text",
  fields: {
    targets: { bind: "targets" },
    viewport: { bind: "viewport" },
    output: { bind: "output" },
    selectors: { bind: "selectors" },
    question: {
      argument: true,
      help: "Question used as the brief heading"
    }
  }
});
