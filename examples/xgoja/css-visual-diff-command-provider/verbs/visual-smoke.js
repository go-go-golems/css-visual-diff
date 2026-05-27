__package__({
  name: "smoke",
  short: "Generated xgoja visual-diff smoke verbs"
});

__section__("artifact", {
  title: "Artifact",
  fields: {
    outDir: { type: "string", default: "./dist/artifacts", help: "Artifact output directory" },
    question: { type: "string", default: "What changed in the pricing card?", help: "Brief heading" }
  }
});

function compareWidget(artifact) {
  const fs = require("fs");
  const report = require("report");
  fs.mkdirSync(artifact.outDir, { recursive: true });

  const evidence = {
    left: {
      url: "fixture:left",
      selector: ".pricing-card",
      box: { x: 24, y: 32, width: 320, height: 180 },
      computed: {
        "font-size": "16px",
        "background-color": "rgb(255, 255, 255)",
        "border-radius": "12px"
      }
    },
    right: {
      url: "fixture:right",
      selector: ".pricing-card",
      box: { x: 24, y: 32, width: 320, height: 184 },
      computed: {
        "font-size": "18px",
        "background-color": "rgb(247, 250, 255)",
        "border-radius": "16px"
      }
    },
    diff: {
      pixelMismatch: 1842,
      mismatchRatio: 0.031,
      changedAttributes: ["class"],
      changedStyles: ["font-size", "background-color", "border-radius"]
    }
  };

  const brief = report.renderAgentBrief({
    question: artifact.question,
    evidence,
    maxBullets: 6,
  });
  const markdownPath = `${artifact.outDir.replace(/\/$/, "")}/agent-brief.md`;
  const jsonPath = `${artifact.outDir.replace(/\/$/, "")}/evidence.json`;
  fs.writeFileSync(markdownPath, brief);
  fs.writeFileSync(jsonPath, JSON.stringify(evidence, null, 2));
  return { markdown_path: markdownPath, json_path: jsonPath, changed_styles: evidence.diff.changedStyles };
}

__verb__("compareWidget", {
  name: "compare-widget",
  short: "Render a css-visual-diff agent brief from synthetic fixture evidence",
  sections: ["artifact"],
  fields: {
    artifact: { bind: "artifact" }
  }
});
