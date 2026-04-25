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
| Snapshot | What did a set of probes observe on this page? | `await cvd.snapshot(page, probes)` |
| Diff | What changed between two snapshots? | `cvd.diff(before, after)` |
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
    return await cvd.snapshot(page, [
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

const diff = cvd.diff(before, after, {
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

## 8. Writing Reports and Artifacts

A script becomes useful to a team when it leaves evidence behind. `cvd.write.json(...)` writes machine-readable data. `cvd.report(diff).writeMarkdown(...)` writes a human-readable explanation.

```js
async function compareAndWrite(beforeUrl, afterUrl, outDir) {
  const before = await snapshotHome(beforeUrl)
  const after = await snapshotHome(afterUrl)
  const diff = cvd.diff(before, after)

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

## 9. When to Use Inspect Artifacts Instead

Snapshots are compact and script-friendly. Inspect artifacts are evidence-rich. Use `page.inspect(...)` or `page.inspectAll(...)` when you need screenshots, prepared HTML, CSS JSON/Markdown, or DOM inspection JSON on disk.

| Use snapshots when... | Use inspect artifacts when... |
|---|---|
| You need fast data comparisons. | A human must review a screenshot or HTML file. |
| You want in-memory diffs. | You need a durable artifact bundle. |
| You are iterating on CSS selectors. | You are preparing CI evidence or a catalog. |
| You are writing custom assertions. | You want the standard bundle shape. |

A good workflow often uses both. During authoring, use locators and snapshots to quickly answer questions. Before review or CI, write inspect artifacts and a catalog so the evidence is durable.

## 10. A Full Authoring Loop

Here is a complete script that supports the kind of pixel-accuracy loop you use while building a page. It captures a baseline, captures the current implementation, diffs them, writes evidence, and returns a structured summary.

```js
async function compareCheckout(beforeUrl, afterUrl, outDir) {
  const cvd = require("css-visual-diff")

  async function capture(url) {
    const browser = await cvd.browser()
    try {
      const page = await browser.page(url, { viewport: cvd.viewport.desktop(), waitMs: 250 })
      return await cvd.snapshot(page, [
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
  const diff = cvd.diff(before, after)

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

## 11. Debugging the Feedback Loop

When a script fails, debug from the outside in. First prove the page loaded. Then prove the selector exists. Then prove the element is visible. Only then inspect CSS values or compare snapshots.

| Problem | Likely cause | What to do |
|---|---|---|
| `exists` is false | The selector does not match the prepared DOM. | Use `page.locator(selector).status()` and inspect prepared HTML with `css-visual-diff html`. |
| `visible` is false | The element exists but has zero bounds, `display:none`, or hidden visibility. | Extract bounds and computed `display` / `visibility`. |
| Text differs only in spacing | DOM text contains newlines or multiple spaces. | Use `text({ normalizeWhitespace: true, trim: true })`. |
| Diff changes bounds on every run | Layout is responsive or unstable. | Ignore expected paths or compare more stable CSS/text facts. |
| Script says a method belongs elsewhere | Locator/probe/extractor concepts were mixed. | Use locators for live page reads, probes for reusable recipes, extractors for what to read. |

## 12. Key Points

- Pixel accuracy is a loop, not a one-time screenshot. The loop is render, locate, extract, compare, report, and adjust.
- Locators are page-bound handles. They are best for exploration and direct questions about the current page.
- Probes are reusable recipes. They are best when you want repeatable checks across targets, runs, or CI jobs.
- Extractors make the question explicit. A script that asks for `text`, `bounds`, and `computedStyle(["color"])` is easier to review than a script that dumps everything.
- Snapshots are plain data. They are cheap to diff, write, store, and reason about.
- Inspect artifacts are durable evidence. Use them when a human needs screenshots, HTML, or full CSS/DOM files.

## See Also

- `css-visual-diff help javascript-api`
- `css-visual-diff help javascript-verbs`
- `css-visual-diff help inspect-workflow`
- `css-visual-diff help config-selectors`
