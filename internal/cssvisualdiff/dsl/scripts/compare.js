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

function region(targets, viewport, output, selectors) {
  return require("diff").compareRegion({
    left: {
      url: targets.leftUrl,
      selector: selectors.leftSelector,
      waitMs: targets.leftWaitMs,
    },
    right: {
      url: targets.rightUrl,
      selector: selectors.rightSelector || selectors.leftSelector,
      waitMs: targets.rightWaitMs,
    },
    viewport: {
      width: viewport.width,
      height: viewport.height,
    },
    output,
    computed: [
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

function brief(targets, viewport, output, selectors, question) {
  const result = region(targets, viewport, output, selectors);
  return require("report").renderAgentBrief({
    question: question,
    evidence: result,
    maxBullets: 8,
  });
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
