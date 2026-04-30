# css-visual-diff

`css-visual-diff` is a programmable visual feedback tool for frontend work. It opens real browser pages, targets specific DOM regions, and produces the evidence you need to tune an implementation against a prototype: screenshots, cropped region images, pixel diffs, computed CSS, matched stylesheet values, structured JSON, Markdown summaries, and review-site datasets.

The tool is intentionally **JavaScript-first**. The old native `run --config` YAML pipeline has been removed. For project-scale workflows, write JavaScript verbs and let those scripts load whatever project data they need: YAML specs, JSON, generated registries, Storybook metadata, or ad-hoc lists of selectors.

The core idea is simple:

> Use Go and Chrome DevTools for reliable browser/artifact primitives. Use JavaScript for project-specific orchestration.

---

## What problems does it solve?

Visual frontend work is usually blocked by vague evidence. A screenshot says "it looks wrong", but not why. A CSS dump has the facts, but too many of them. A full visual regression suite can be useful in CI, but it is too broad for a human tuning one component.

`css-visual-diff` is designed for the middle of the loop:

1. choose the smallest meaningful target,
2. verify that selectors exist and point at the right elements,
3. compare only that target,
4. inspect compact JSON plus image artifacts,
5. change one CSS/token/component detail,
6. repeat,
7. run broader page or suite validation only after the local target is stable.

This makes visual feedback token-efficient for both humans and AI agents. Instead of pasting a full DOM, a full screenshot, and a wall of JSON into a prompt, you can hand over a few high-signal artifacts:

```text
summary.json                    compact classification + artifact paths
compare.json                    structured per-region evidence
left_region.png                 prototype crop
right_region.png                implementation crop
diff_only.png                   changed pixels only
computed CSS / style diffs      exact properties that differ
snapshot.json                   semantic DOM/style state, when needed
```

---

## Current shape

The live tool centers on:

- Chromium-driven capture through `chromedp` / Chrome DevTools Protocol,
- direct one-region comparisons with `css-visual-diff compare`,
- JavaScript verbs under `css-visual-diff verbs ...`,
- Goja-powered project scripts using `require("css-visual-diff")`,
- computed CSS and matched-style inspection,
- pixel diff artifacts,
- structured JSON and Markdown reports,
- review-site datasets served by `css-visual-diff serve`.

It does **not** provide the old native YAML manifest runner anymore. YAML is still useful as project data, but it is interpreted by your JavaScript userland, not by a built-in `run --config` command.

---

## Install / build

From a checkout:

```bash
go install ./cmd/css-visual-diff
```

Then verify the command:

```bash
css-visual-diff --help
css-visual-diff compare --help
css-visual-diff verbs --help
```

For repository development:

```bash
go test ./...
go build ./cmd/css-visual-diff
```

---

## The two ways to use it

### 1. Direct command: compare one region now

Use `compare` when you already know the two URLs and selectors.

```bash
css-visual-diff compare \
  --url1 http://localhost:7070/standalone/public/shows.html \
  --selector1 '[data-page="shows"]' \
  --url2 'http://localhost:6007/iframe.html?id=public-site-pages--shows-desktop&viewMode=story' \
  --selector2 '[data-page="shows"]' \
  --viewport-w 1280 \
  --viewport-h 1600 \
  --threshold 30 \
  --out /tmp/cssvd-shows-page
```

This writes a small artifact directory with JSON, Markdown, screenshots, and pixel diff images. It is the fastest way to answer:

> Are these two rendered regions visually close, and where do they differ?

### 2. JavaScript verbs: encode the workflow for a project

Use `verbs` when the work has project meaning: pages, sections, variants, policies, accepted differences, Storybook ports, archive directories, semantic snapshots, or CI gates.

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page shows \
  --section shows-list \
  --outDir /tmp/pyxis-shows-list \
  --summary \
  --output json
```

The repository path points at a folder of JavaScript files that register commands with `__verb__`. Those commands can use the native module:

```js
const cvd = require("css-visual-diff")
const browser = await cvd.browser()
const page = await browser.page(url, {
  viewport: cvd.viewport(1280, 1600),
  waitMs: 500,
})

const result = await page.inspectAll([
  {
    name: "shows-list",
    selector: ".pyxis-show-grid",
    props: ["display", "grid-template-columns", "gap", "color"],
    attributes: ["class", "data-pyxis-component"],
  },
], {
  outDir: "/tmp/pyxis-inspect/shows-list",
  artifacts: "css-json",
})
```

The JavaScript layer is where project-specific decisions belong. The Go core should stay focused on browser actions and artifacts.

---

## Mental model: token-efficient visual feedback

A good `css-visual-diff` workflow does not start with "run everything". It starts with the smallest question that can change your next edit.

| Question | Command shape | Evidence to read |
| --- | --- | --- |
| Does this selector exist on both sides? | `inspect-spec` / `inspect-section` | compact JSON with `exists`, `visible`, bounds, selected CSS |
| Is this one region visually close? | `compare-spec --page ... --section ... --summary` | changed percent + `diff_only.png`, `left_region.png`, `right_region.png` |
| Is this atom or layout? | inspect nested elements with style presets | bounds, display/grid/flex, typography, spacing |
| Did my CSS/token fix help? | rerun the same narrow comparison | changed percent and same artifact paths |
| Is the whole page now acceptable? | compare the page after section-level fixes | page summary and per-section rows |
| Can CI gate this? | `compare-all --mode ci` | policy result + failures list |
| What changed semantically? | `snapshot-section` + `diff-snapshots` | snapshot JSON and Markdown diff |

The goal is to move from high-volume evidence to high-signal evidence. For example, the Pyxis operator guide recommends comparing `shows-list` rather than the whole Shows page when tuning poster layout. The useful output is not the full suite object; it is a compact row and three images:

```text
/tmp/pyxis-shows-list/shows/artifacts/shows-list/diff_only.png
/tmp/pyxis-shows-list/shows/artifacts/shows-list/right_region.png
/tmp/pyxis-shows-list/shows/artifacts/shows-list/left_region.png
```

Those artifacts are small enough to inspect manually, upload to a reviewer, or pass into an AI prompt with a short instruction such as:

> Compare the prototype crop and implementation crop. Focus on poster card spacing, image treatment, and typography. Use `diff_only.png` to locate the largest changes.

---

## Pyxis-style workflow examples

The Pyxis project uses `css-visual-diff` through `prototype-design/visual-diff/userland`. That directory contains:

```text
prototype-design/visual-diff/userland/
├── lib/                    # reusable JS modules
├── specs/                  # project-specific visual suite specs
├── verbs/                  # registered css-visual-diff commands
└── scripts/                # smoke scripts and repeatable operator flows
```

The specs are YAML, but they are **Pyxis specs**, not native `css-visual-diff` manifests. They are loaded and interpreted by Pyxis JavaScript.

### Start with discovery

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages list-targets \
  --output json
```

Use this before guessing page names or section names.

### Inspect selectors before comparing

If a comparison looks strange, first ask whether the selectors are valid and visible:

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages inspect-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page shows \
  --section shows-list \
  --elements '&,.pyxis-show-grid,.pyxis-show-tile,.pyxis-poster' \
  --stylePreset layout \
  --summary \
  --output json
```

Useful Pyxis style presets include:

- `layout` — bounds, display, grid/flex, width/height questions,
- `typography` — font family, size, line height, weight, color,
- `surface` — backgrounds, borders, radius, shadows,
- `spacing` — margin and padding,
- `pageShell` — page-level shell/container properties.

### Compare one section during tuning

```bash
OUT=/tmp/pyxis-user-shows-list-tune
rm -rf "$OUT"

css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page shows \
  --section shows-list \
  --outDir "$OUT" \
  --summary \
  --output json \
  > "$OUT-summary.json"
```

Then inspect only what matters:

```bash
jq '.[0].rows[] | {
  section,
  classification,
  changedPercent,
  diffOnlyPath,
  leftRegionPath,
  rightRegionPath
}' "$OUT-summary.json"
```

### Use aliases for repeated workflows

Pyxis defines convenience verbs for common operations. For the user-site Shows page:

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-user-shows-section \
  shows-list \
  --outDir /tmp/pyxis-user-shows-list-tune \
  --output json
```

For public component targets:

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-public-component \
  show-tile-redroom \
  --outDir /tmp/pyxis-public-component-show-tile-redroom \
  --output json
```

This is useful when a page-level diff points at a specific atom or component. Compare the component first, fix it, then return to the page-level section.

### Compare app components with explicit specs

The default Pyxis registry is public-page oriented. For app components and app pages, pass the spec explicitly:

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/app.components.visual.yml \
  --page app-topbar-dashboard \
  --section component \
  --outDir /tmp/pyxis-topbar-dashboard-tune \
  --summary \
  --output json
```

This is the workflow used for atom-to-page debugging. A page diff may show an 8% pixel change, but the real fix might be a token mismatch in an atom such as `Button`, `Avatar`, or `PyxisLogo`.

### Capture semantic snapshots

Pixel diffs answer "where did pixels change?" Snapshots answer "what semantic DOM/style facts changed?"

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages snapshot-section archive content \
  --outDir /tmp/pyxis-snapshot-before \
  --stylePreset pageShell \
  --output json

css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages snapshot-section archive content \
  --outDir /tmp/pyxis-snapshot-after \
  --stylePreset pageShell \
  --output json

css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages diff-snapshots \
  /tmp/pyxis-snapshot-before/snapshot.json \
  /tmp/pyxis-snapshot-after/snapshot.json \
  --outDir /tmp/pyxis-snapshot-diff \
  --output json
```

Use this when a screenshot changed and you need to know whether the change is layout, typography, text, attributes, or component structure.

### Gate with CI policy only after local evidence is stable

```bash
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-all \
  --page archive \
  --outDir /tmp/pyxis-archive-ci \
  --threshold 30 \
  --inspect minimal \
  --mode ci \
  --maxChangedPercent 10 \
  --output json
```

CI mode is broad validation. Do not use it as the first command in a tuning loop.

---

## Writing an ad-hoc script for the problem at hand

The best workflows are often small scripts created for one ticket. They keep the terminal output small and write durable artifacts for review.

Create a local verb repository:

```text
visual-tools/
└── ticket-verbs.js
```

Example: inspect one selector and write only the facts needed for a CSS fix.

```js
async function inspectTokenSurface(url, selector, outDir, values) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()

  try {
    const page = await browser.page(url, {
      viewport: cvd.viewport(values.width || 1280, values.height || 900),
      waitMs: values.waitMs || 300,
      name: values.name || "target",
    })

    const locator = page.locator(selector)
    const element = await cvd.extract(locator, [
      cvd.extractors.exists(),
      cvd.extractors.visible(),
      cvd.extractors.text(),
      cvd.extractors.bounds(),
      cvd.extractors.computedStyle([
        "display",
        "width",
        "height",
        "font-family",
        "font-size",
        "font-weight",
        "line-height",
        "color",
        "background-color",
        "border-radius",
        "box-shadow",
      ]),
      cvd.extractors.attributes(["class", "data-pyxis-component", "data-pyxis-part"]),
    ])

    await cvd.write.json(`${outDir}/element.json`, element)

    return {
      ok: !!element.exists && !!element.visible,
      selector,
      text: element.text || "",
      bounds: element.bounds,
      color: element.computed && element.computed.color,
      background: element.computed && element.computed["background-color"],
      out: `${outDir}/element.json`,
    }
  } finally {
    await browser.close()
  }
}

__verb__("inspect-token-surface", {
  parents: ["ticket"],
  short: "Inspect one element with token/surface CSS properties",
  fields: {
    url: { argument: true, required: true },
    selector: { argument: true, required: true },
    outDir: { argument: true, required: true },
    values: { bind: "all" },
    width: { type: "int", default: 1280 },
    height: { type: "int", default: 900 },
    waitMs: { type: "int", default: 300 },
    name: { type: "string", default: "target" },
  },
})
```

Run it:

```bash
css-visual-diff verbs \
  --repository visual-tools \
  ticket inspect-token-surface \
  'http://localhost:6008/iframe.html?id=shell-apptopbar--dashboard' \
  '[data-pyxis-component="app-topbar"]' \
  /tmp/topbar-token-surface \
  --output json
```

This is the intended extension model: write the script that answers the current visual question, keep its output compact, and let the artifact files carry the detailed evidence.

---

## Review site workflow

When a script or suite writes review-site data, serve it locally:

```bash
css-visual-diff serve \
  --data-dir /tmp/pyxis-review-run \
  --port 8097
```

The data directory should contain `summary.json` plus per-page/per-section artifact directories:

```text
/tmp/pyxis-review-run/
├── summary.json
└── shows/
    └── artifacts/
        └── shows-list/
            ├── compare.json
            ├── left_region.png
            ├── right_region.png
            ├── diff_only.png
            └── diff_comparison.png
```

The review server is for local/operator review. It serves only paths under `--data-dir` and validates page/section/artifact path segments before reading files.

---

## Built-in and example verbs

Inspect a page into a catalog:

```bash
css-visual-diff verbs catalog inspect-page \
  http://127.0.0.1:8767/ \
  '#cta' \
  /tmp/cssvd-page \
  --slug cta \
  --artifacts css-json \
  --output json
```

Use the example repository for script patterns:

```bash
css-visual-diff verbs \
  --repository examples/verbs \
  examples low-level inspect \
  http://127.0.0.1:8767/ \
  '#cta' \
  /tmp/cssvd-low-level \
  --output json
```

Run the example review sweep:

```bash
css-visual-diff verbs \
  --repository examples/verbs \
  examples review-sweep from-spec \
  --specFile examples/specs/review-sweep.example.yaml \
  --outDir /tmp/example-review

css-visual-diff serve --data-dir /tmp/example-review --port 8098
```

---

## Project-local verb repositories

For repeatable project workflows, check in a small repository config:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./visual-diff/userland
```

Relative paths are resolved from the config file that declares them. Use `.css-visual-diff.override.yml` for private local repositories and keep it gitignored.

You can also pass repositories explicitly:

```bash
css-visual-diff verbs --repository prototype-design/visual-diff/userland --help
```

---

## Embedded help

The CLI includes Glazed help pages:

```bash
css-visual-diff help javascript-api
css-visual-diff help javascript-verbs
css-visual-diff help pixel-accuracy-scripting-guide
css-visual-diff help review-site-data-spec
```

Use these when authoring scripts or debugging output formats.

---

## Working rules

- Prefer the narrowest target that answers the current question.
- Inspect selectors before trusting a visual diff.
- Read image artifacts before reading full JSON.
- Use `--summary` for operator loops and full JSON only for debugging internals.
- Keep project meaning in JavaScript userland: specs, registries, policy bands, accepted differences, report shape.
- Add core features as browser/service/JS API primitives, not as a new native manifest format.
- Use direct commands for one-off work and JS verbs for repeatable workflows.
- Treat stale docs as stale APIs: if a command no longer exists, remove examples that teach it.

---

## Historical note

Earlier versions had a native YAML `run --config` pipeline with Go-owned manifests and mode dispatch. That path was deliberately removed so the project could focus on a smaller, more flexible core. Project-specific YAML is still welcome, but it should be loaded by JavaScript verbs as userland data.
