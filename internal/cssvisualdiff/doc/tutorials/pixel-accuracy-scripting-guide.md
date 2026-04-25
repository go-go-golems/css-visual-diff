---
Title: "Building Pixel-Accurate Websites with css-visual-diff Scripts"
Slug: pixel-accuracy-scripting-guide
Short: A textbook-style guide to using css-visual-diff JavaScript scripts as a practical feedback loop for pixel-accurate frontend work.
Topics:
- visual-regression
- javascript
- pixel-accuracy
- browser-automation
- frontend
Commands:
- verbs
- inspect
- screenshot
- css-md
- css-json
Flags:
- repository
- config
- side
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Pixel accuracy is not a single screenshot. It is a feedback loop. You render the page, look at the part that matters, measure the DOM and CSS that produced it, change the implementation, and run the same measurement again. `css-visual-diff` gives that loop a programmable shape: browser pages, locators, probes, extractors, snapshots, diffs, reports, and catalogs.

This guide teaches that shape from the ground up. The goal is not to memorize every method. The goal is to learn how to build small scripts that answer precise visual questions while you are building a website: Does the CTA exist? Is it visible? What text does it render? What color, spacing, font size, and bounds does the browser compute? Did those facts change after my CSS edit? When the answer is yes, can I write a report that explains the change?

## 1. The Mental Model

A visual comparison has three different layers. Keeping them separate makes scripts easier to write and easier to debug.

| Layer | Question | API shape |
|---|---|---|
| Page | What browser state am I inspecting? | `const page = await browser.page(url, options)` |
| Locator | Which element on this loaded page matters now? | `const cta = page.locator("#cta")` |
| Probe | What reusable recipe should I apply to a page? | `cvd.probe("cta").selector("#cta").text().styles([...])` |
| Extractor | What facts should I read from an element? | `cvd.extractors.text()`, `cvd.extractors.computedStyle([...])` |
| Snapshot | What did a set of probes observe on this page? | `await cvd.snapshot.page(page, probes)` |
| Diff | What changed between two snapshots? | `cvd.diff.structural(before, after)` |
| Report | How do I show the difference to a human or CI log? | `cvd.report(diff).markdown()` |

The distinction between a locator and a probe is the most important one. A locator is bound to one live page. It answers: “on this loaded page, find `#cta`.” A probe is a reusable recipe. It answers: “whenever I inspect a page, call this thing `cta`, use selector `#cta`, and extract text plus these CSS properties.”

When you are exploring, use locators. When you are building repeatable checks, use probes and snapshots.

## 2. The Smallest Useful Script

A script starts by opening a browser and page. Browser and page operations are asynchronous because they touch Chromium. Always close the browser in a `finally` block; otherwise an error in the middle of the script may leave a browser process running.

```js
async function inspectHome(url) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()

  try {
    const page = await browser.page(url, {
      viewport: { width: 1280, height: 720 },
      waitMs: 250,
      name: "home",
    })

    const cta = page.locator("#cta")
    return {
      exists: await cta.exists(),
      visible: await cta.visible(),
      text: await cta.text({ normalizeWhitespace: true, trim: true }),
    }
  } finally {
    await browser.close()
  }
}
```

This script does not compare images yet. That is intentional. Pixel accuracy starts with confidence that you are looking at the right element. A screenshot of the wrong element can be perfectly stable and completely useless.

For app and Storybook pages that render asynchronously, wait for the selector before reading or comparing it:

```js
await page.locator("[data-testid='primary-cta']").waitFor({
  timeoutMs: 30000,
  visible: true,
  afterWaitMs: 250,
})
```

`page.waitForSelector(selector, options)` is also available, but the locator form keeps the code centered on the selector you will inspect or compare next.

## 3. Turning a Function into a CLI Verb

A JavaScript file becomes a CLI command when it declares a verb. The metadata is static so `css-visual-diff` can scan the file without executing browser work.

```js
async function inspectHome(url) {
  // same function as above
}

__verb__("inspectHome", {
  parents: ["site"],
  short: "Inspect the homepage CTA",
  fields: {
    url: { argument: true, required: true, help: "Page URL" },
  },
})
```

Run it from a repository directory:

```bash
css-visual-diff verbs --repository ./verbs site inspect-home http://localhost:3000 --output json
```

The command name is normalized from `inspectHome` to `inspect-home`. The important design choice is that scripts live under `css-visual-diff verbs ...`, not at the root command. This keeps script discovery, duplicate command names, and repository-specific flags local to the programmable layer.

## 4. Reading What the Browser Actually Computed

CSS source files show intent. The browser shows the result. For pixel-accurate work, the result is what matters: computed color, actual font size, final bounds, rendered text, and real attributes after React, CSS, layout, and browser defaults have done their work.

```js
async function inspectButton(url) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  try {
    const page = await browser.page(url, { viewport: { width: 390, height: 844 } })
    const button = page.locator("[data-testid='primary-cta']")
    await button.waitFor({ timeoutMs: 30000, visible: true })

    const [status, text, bounds, styles, attrs] = await Promise.all([
      button.status(),
      button.text({ normalizeWhitespace: true, trim: true }),
      button.bounds(),
      button.computedStyle(["display", "height", "padding-left", "padding-right", "font-size", "font-weight", "color", "background-color", "border-radius"]),
      button.attributes(["class", "aria-label", "disabled"]),
    ])

    return { status, text, bounds, styles, attrs }
  } finally {
    await browser.close()
  }
}
```

`Promise.all` is safe here. Internally, operations on one page are serialized so Chromium is not asked to do conflicting page work at the same time. The JavaScript stays ergonomic while the Go side preserves page safety.

## 5. From Exploration to Repeatable Probes

Locators are good while exploring. Probes are better once you know what matters. A probe names the check, stores the selector, and declares which extractors should run.

```js
const ctaProbe = cvd.probe("primary-cta")
  .selector("[data-testid='primary-cta']")
  .required()
  .text()
  .bounds()
  .styles(["height", "font-size", "font-weight", "color", "background-color", "border-radius"])
  .attributes(["class", "aria-label"])
```

The probe is a Go-backed builder. If you call a method that belongs somewhere else, the API tells you. For example, `.computedStyle(...)` belongs to locators, while `.styles(...)` belongs to probes. This distinction helps both humans and LLM-written scripts recover from mistakes.

## 6. Snapshots: Capture the Facts You Care About

A snapshot applies a set of probes to one loaded page and returns plain serializable data. It does not write artifacts unless you explicitly choose another API that writes files. This makes snapshots ideal for quick authoring loops and CI assertions.

```js
async function snapshotHome(url) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  try {
    const page = await browser.page(url, { viewport: cvd.viewport.desktop() })
    return await cvd.snapshot.page(page, [
      cvd.probe("hero-title").selector("h1").text().styles(["font-size", "line-height", "color"]),
      cvd.probe("primary-cta").selector("[data-testid='primary-cta']").text().bounds().styles(["height", "background-color", "border-radius"]),
    ])
  } finally {
    await browser.close()
  }
}
```

The result has one entry per probe. Each entry contains the probe name, selector, and extracted snapshot data. Because the result is plain data, you can write it to JSON, compare it, store it in a catalog, or feed it into another script.

## 7. Diffing Two States

The simplest useful comparison is structural: take two snapshots and ask what changed. This works well for text, CSS properties, attributes, and bounds.

```js
const before = await snapshotHome("http://localhost:3000/before")
const after = await snapshotHome("http://localhost:3000/after")

const diff = cvd.diff.structural(before, after, {
  ignorePaths: [
    "results[1].snapshot.bounds.x",
    "results[1].snapshot.bounds.y",
  ],
})

if (!diff.equal) {
  console.log(cvd.report(diff).markdown())
}
```

Ignore paths are useful when a value is expected to move but the important visual contract is somewhere else. For example, a responsive layout may shift `x` and `y` while preserving height, typography, and color. Ignore paths let you state that policy directly.

## 8. Quick Pixel Region Comparisons

Structural diffs compare data. Pixel diffs compare rendered images. When you need the low-effort pixel-accuracy path, use `cvd.compare.region(...)`. It collects both locators with `inspect: "rich"` by default, captures region screenshots, computes a pixel diff, writes PNG artifacts, and returns a `SelectionComparison` handle.

```js
async function compareCTA(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  let leftPage, rightPage

  try {
    leftPage = await browser.page(leftUrl, { viewport: cvd.viewport.desktop(), waitMs: 250 })
    rightPage = await browser.page(rightUrl, { viewport: cvd.viewport.desktop(), waitMs: 250 })

    const comparison = await cvd.compare.region({
      name: "primary-cta",
      left: leftPage.locator("[data-testid='primary-cta']"),
      right: rightPage.locator("[data-testid='primary-cta']"),
      outDir,
      threshold: 30,
      styleProps: ["font-size", "line-height", "color", "background-color", "border-radius"],
      attributes: ["class", "aria-label"],
    })

    await comparison.artifacts.write(outDir, ["json", "markdown"])
    return comparison.summary()
  } finally {
    if (leftPage) await leftPage.close()
    if (rightPage) await rightPage.close()
    await browser.close()
  }
}
```

The returned object is behavior-rich. Use `comparison.summary()` for compact command output, `comparison.toJSON()` for full machine-readable data, `comparison.report.markdown()` for a human explanation, and `comparison.artifacts.write(...)` for durable JSON/Markdown output.

`cvd.compare.region(...)` uses the collection profile system. The default `inspect: "rich"` is intentionally diagnostic: it collects normalized text, a focused set of computed styles, a focused set of attributes, and selector status/bounds before doing the pixel comparison. Use `inspect: "minimal"` when a high-volume CI pass only needs status/bounds/pixels. Use `inspect: "debug"` for one-off failures where you want inner HTML, all computed styles, and all attributes in the collected data.

`styleProps` and `attributes` are collection filters. Passing them does not merely hide fields later; it narrows what browser facts are collected. For example, `styleProps: ["font-size", "line-height"]` means only those computed CSS properties are captured for style comparison. If you omit `styleProps` in `rich` mode, the default set is `display`, `position`, `color`, `background-color`, `font-family`, `font-size`, `font-weight`, `line-height`, `margin`, `padding`, and `border`. If you omit `attributes` in `rich` mode, the default set is `id`, `class`, `role`, `aria-label`, and `data-testid`.

`comparison.artifacts.write(...)` also returns a path map that is useful in project-local CLI summaries:

```js
const written = await comparison.artifacts.write(outDir, ["json", "markdown"])
return {
  changedPercent: comparison.pixel.summary().changedPercent,
  compareJson: written.json,
  compareMarkdown: written.markdown,
  diffComparison: written.diffComparison,
}
```

The files written under `outDir` are the evidence a reviewer needs:

- `left_region.png` shows the left rendered region.
- `right_region.png` shows the right rendered region.
- `diff_only.png` highlights changed pixels.
- `diff_comparison.png` places left, right, and diff side by side.
- `compare.json` and `compare.md` are written when you call `comparison.artifacts.write(...)`.

## 9. Collect First, Then Analyze in JavaScript

The quick path is best when you want the standard answer. The primitive path is better when you want policy. For example, a typography-only check can collect rich data once, then filter the comparison in JavaScript.

```js
const left = await leftPage.locator("[data-testid='primary-cta']").collect({
  inspect: "rich",
  styleProps: ["font-size", "font-weight", "line-height", "color"],
  attributes: ["class"],
})

const right = await cvd.collect.selection(rightPage.locator("[data-testid='primary-cta']"), {
  inspect: "rich",
  styleProps: ["font-size", "font-weight", "line-height", "color"],
  attributes: ["class"],
})

const comparison = await cvd.compare.selections(left, right, {
  styleProps: ["font-size", "font-weight", "line-height", "color"],
  attributes: ["class"],
})

const typography = comparison.styles.diff(["font-size", "font-weight", "line-height"])
const classChanges = comparison.attributes.diff(["class"])

return {
  ok: typography.length === 0,
  typography,
  classChanges,
  bounds: comparison.bounds.diff(),
}
```

This pattern is the heart of the new API. Browser work happens once. Analysis happens in ordinary JavaScript. Reports and catalogs are then views over the same comparison data.

## 10. Writing Reports and Artifacts

A script becomes useful to a team when it leaves evidence behind. `cvd.write.json(...)` writes machine-readable data. `cvd.report(diff).writeMarkdown(...)` writes a human-readable explanation.

```js
async function compareAndWrite(beforeUrl, afterUrl, outDir) {
  const before = await snapshotHome(beforeUrl)
  const after = await snapshotHome(afterUrl)
  const diff = cvd.diff.structural(before, after)

  await cvd.write.json(`${outDir}/before.json`, before)
  await cvd.write.json(`${outDir}/after.json`, after)
  await cvd.write.json(`${outDir}/diff.json`, diff)
  await cvd.report(diff).writeMarkdown(`${outDir}/diff.md`)

  return {
    equal: diff.equal,
    changeCount: diff.changeCount,
    report: `${outDir}/diff.md`,
  }
}
```

For CI, return structured rows and let Glazed format them as JSON, YAML, table, or CSV. For humans, write Markdown reports and screenshots/catalogs when needed.

## 11. When to Use Inspect Artifacts Instead

Snapshots are compact and script-friendly. Inspect artifacts are evidence-rich. Use `page.inspect(...)` or `page.inspectAll(...)` when you need screenshots, prepared HTML, CSS JSON/Markdown, or DOM inspection JSON on disk.

| Use snapshots when... | Use inspect artifacts when... |
|---|---|
| You need fast data comparisons. | A human must review a screenshot or HTML file. |
| You want in-memory diffs. | You need a durable artifact bundle. |
| You are iterating on CSS selectors. | You are preparing CI evidence or a catalog. |
| You are writing custom assertions. | You want the standard bundle shape. |

A good workflow often uses both. During authoring, use locators and snapshots to quickly answer questions. Before review or CI, write inspect artifacts and a catalog so the evidence is durable.

## 12. A Full Authoring Loop

Here is a complete script that supports the kind of pixel-accuracy loop you use while building a page. It captures a baseline, captures the current implementation, diffs them, writes evidence, and returns a structured summary.

```js
async function compareCheckout(beforeUrl, afterUrl, outDir) {
  const cvd = require("css-visual-diff")

  async function capture(url) {
    const browser = await cvd.browser()
    try {
      const page = await browser.page(url, { viewport: cvd.viewport.desktop(), waitMs: 250 })
      return await cvd.snapshot.page(page, [
        cvd.probe("checkout-title").selector("h1").text().styles(["font-size", "font-weight", "line-height", "color"]),
        cvd.probe("checkout-card").selector("[data-testid='checkout-card']").bounds().styles(["background-color", "border-radius", "box-shadow", "padding"]),
        cvd.probe("pay-button").selector("[data-testid='pay-button']").text().bounds().styles(["height", "font-size", "background-color", "color", "border-radius"]),
      ])
    } finally {
      await browser.close()
    }
  }

  const before = await capture(beforeUrl)
  const after = await capture(afterUrl)
  const diff = cvd.diff.structural(before, after)

  await cvd.write.json(`${outDir}/before.json`, before)
  await cvd.write.json(`${outDir}/after.json`, after)
  await cvd.write.json(`${outDir}/diff.json`, diff)
  await cvd.report(diff).writeMarkdown(`${outDir}/diff.md`)

  return {
    ok: diff.equal,
    changeCount: diff.changeCount,
    report: `${outDir}/diff.md`,
  }
}

__verb__("compareCheckout", {
  parents: ["site"],
  short: "Compare checkout page visual facts",
  fields: {
    beforeUrl: { argument: true, required: true },
    afterUrl: { argument: true, required: true },
    outDir: { argument: true, required: true },
  },
})
```

Run it:

```bash
css-visual-diff verbs --repository ./verbs site compare-checkout \
  http://localhost:3000/baseline/checkout \
  http://localhost:3000/checkout \
  /tmp/cssvd-checkout \
  --output json
```

## 13. Debugging the Feedback Loop

When a script fails, debug from the outside in. First prove the page loaded. Then prove the selector exists. Then prove the element is visible. Only then inspect CSS values or compare snapshots.

| Problem | Likely cause | What to do |
|---|---|---|
| `exists` is false | The selector does not match the prepared DOM. | Use `page.locator(selector).status()` and inspect prepared HTML with `css-visual-diff html`. |
| `visible` is false | The element exists but has zero bounds, `display:none`, or hidden visibility. | Extract bounds and computed `display` / `visibility`. |
| Text differs only in spacing | DOM text contains newlines or multiple spaces. | Use `text({ normalizeWhitespace: true, trim: true })`. |
| Diff changes bounds on every run | Layout is responsive or unstable. | Ignore expected paths or compare more stable CSS/text facts. |
| Script says a method belongs elsewhere | Locator/probe/extractor concepts were mixed. | Use locators for live page reads, probes for reusable recipes, extractors for what to read. |

## 14. Package the Loop as a Project CLI

A one-off script is useful while exploring. A reusable project command is what turns visual validation into a daily development loop.

Use a repository layout like:

```text
visual-verbs/
  shared.js
  homepage.js
  checkout.js
```

Each page-specific file can expose a command through `__verb__`:

```js
async function validateHomepage(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff")
  const browser = await cvd.browser()
  const catalog = cvd.catalog.create({
    title: "Homepage visual validation",
    outDir,
    artifactRoot: "artifacts",
  })

  const sections = [
    { name: "hero", selector: "[data-section='hero']" },
    { name: "primary-cta", selector: "[data-testid='primary-cta']" },
    { name: "footer", selector: "footer" },
  ]

  const summaries = []
  let leftPage, rightPage
  try {
    leftPage = await browser.page(leftUrl, { viewport: cvd.viewport.desktop(), waitMs: 500 })
    rightPage = await browser.page(rightUrl, { viewport: cvd.viewport.desktop(), waitMs: 500 })

    for (const section of sections) {
      const artifactDir = catalog.artifactDir(section.name)
      const comparison = await cvd.compare.region({
        name: section.name,
        left: leftPage.locator(section.selector),
        right: rightPage.locator(section.selector),
        outDir: artifactDir,
      })
      await comparison.artifacts.write(artifactDir, ["json", "markdown"])
      catalog.record(comparison, { slug: section.name, selector: section.selector })
      summaries.push(comparison.summary())
    }

    return {
      ok: summaries.every(s => !s.pixel || s.pixel.changedPercent < 2.0),
      manifestPath: await catalog.writeManifest(),
      indexPath: await catalog.writeIndex(),
      summaries,
    }
  } finally {
    if (leftPage) await leftPage.close()
    if (rightPage) await rightPage.close()
    await browser.close()
  }
}

__verb__("validateHomepage", {
  parents: ["site"],
  short: "Validate homepage visual implementation against a reference URL",
  fields: {
    leftUrl: { argument: true, required: true },
    rightUrl: { argument: true, required: true },
    outDir: { argument: true, required: true },
  },
})
```

Run it directly:

```bash
css-visual-diff verbs --repository ./visual-verbs site validate-homepage \
  http://localhost:4100 \
  http://localhost:4200 \
  ./artifacts/visual/homepage/latest \
  --output json
```

Or wrap it in project-native commands:

```json
{
  "scripts": {
    "visual:homepage": "css-visual-diff verbs --repository ./visual-verbs site validate-homepage",
    "visual:homepage:local": "npm run visual:homepage -- http://localhost:4100 http://localhost:4200 ./artifacts/visual/homepage/latest --output json"
  }
}
```

The command has two products:

```text
stdout JSON:   ok/fail summary, changed percentages, report paths
artifact dir:  index.md, manifest.json, screenshots, pixel diffs, compare.json, compare.md
```

This split is important. Automation should read the compact JSON result. Humans should open the catalog index and image artifacts. CI can run the same command, upload the artifact directory, and fail only when the JSON policy says the difference is a blocker.

A mature project usually has both fast and full commands:

```text
visual:quick   one page, one viewport, high-value selectors; local authoring
visual:full    many pages/viewports/sections; CI or pre-release validation
```

The loop becomes:

```text
edit CSS/component
run visual command
open index.md and diff_comparison.png
inspect JSON/Markdown evidence
fix implementation
run again
```

For coding agents, this packaged command is a concrete feedback loop: after changing UI code, run the visual command, inspect structured differences, and only claim success when the command passes or remaining differences are documented.

## 15. Key Points

- Pixel accuracy is a loop, not a one-time screenshot. The loop is render, locate, extract, compare, report, and adjust.
- Locators are page-bound handles. They are best for exploration and direct questions about the current page.
- Probes are reusable recipes. They are best when you want repeatable checks across targets, runs, or CI jobs.
- Extractors make the question explicit. A script that asks for `text`, `bounds`, and `computedStyle(["color"])` is easier to review than a script that dumps everything.
- Snapshots are plain data. They are cheap to diff, write, store, and reason about.
- `cvd.compare.region(...)` is the quickest way to get pixel artifacts plus structured selector comparison data.
- Catalogs turn many section comparisons into one review packet with `manifest.json`, `index.md`, and per-section artifacts.
- Repository-local verbs package visual validation as reusable CLI commands for local authoring, CI, and coding-agent loops.
- Inspect artifacts are durable evidence. Use them when a human needs screenshots, HTML, or full CSS/DOM files.

## See Also

- `css-visual-diff help javascript-api`
- `css-visual-diff help javascript-verbs`
- `css-visual-diff help inspect-workflow`
- `css-visual-diff help config-selectors`
