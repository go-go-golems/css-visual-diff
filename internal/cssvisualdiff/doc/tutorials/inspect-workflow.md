---
Title: Inspect One Side Before Comparing
Slug: inspect-workflow
Short: Debug a css-visual-diff YAML file by capturing one side, one selector, and one artifact at a time.
Topics:
- visual-regression
- inspect
- selectors
Commands:
- inspect
- screenshot
- css-md
- css-json
- html
Flags:
- config
- side
- style
- section
- selector
- output-file
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

The inspect workflow turns a full visual comparison into small checks that are easy to repeat. Instead of running every mode and then guessing why a selector failed, you load the same YAML config, choose one side, and ask for the exact artifact you need: a screenshot, prepared HTML, computed CSS, or structured DOM JSON.

## What You Will Validate

A visual diff is only useful after both sides render the elements you intended to compare. The inspect commands validate the setup in layers: first the prepared DOM, then the screenshot crop, then the computed CSS values for a specific element.

Use this order when a config is new or failing:

1. Capture prepared root HTML to confirm the prepare hook ran.
2. Capture a screenshot to confirm the selector crops the right element.
3. Capture CSS Markdown to confirm the CSS probe points at the exact element and properties.
4. Run the full comparison only after both sides inspect correctly.

## Step 1: Capture Prepared HTML

Prepared HTML shows what the browser sees after navigation, wait time, and the target's `prepare` hook. This is the fastest way to debug missing selectors because it answers whether the element exists at all.

```bash
css-visual-diff html \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side original \
  --root \
  --output-file /tmp/original-root.html
```

Open `/tmp/original-root.html` and search for the `data-comp`, `data-part`, class, or tag you expect. If the element is missing here, fix the URL, wait condition, or prepare script before tuning screenshot selectors.

## Step 2: Capture a Screenshot

A screenshot checks the visual crop. Use a named `--section` when you want the screenshot region, or a named `--style` when the CSS probe selector is also the exact element you want to crop.

```bash
css-visual-diff screenshot \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary.png
```

The command writes one PNG and prints the output path. The image should be tightly cropped around the selected element. If it contains too much surrounding layout, add or adjust a screenshot region in `sections[]`.

## Step 3: Capture Computed CSS

Computed CSS verifies that the CSS selector points at the element that carries the relevant visual styles. Use Markdown for manual review and JSON for scripts.

```bash
css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary-css.md
```

The Markdown output includes existence, bounds, and a property table. For a configured `styles[]` entry, the command uses that entry's `props` list unless you override it with `--props`.

## Step 4: Inspect Both Sides

A config is ready for comparison when original and React produce matching intended elements. Run the same command for both sides and compare the outputs.

```bash
css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side original \
  --style button-primary \
  --output-file /tmp/button-primary-original.md

css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary-react.md
```

If both files show the expected element and properties, move on to `run --modes capture,cssdiff,pixeldiff,html-report` for the full comparison.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `--side must be original or react` | The side flag is missing or misspelled. | Use `--side original` or `--side react`. |
| `provide exactly one of ...` | More than one selector source was provided, or none was provided. | Choose exactly one of `--root`, `--section`, `--style`, or `--selector`. |
| Selector screenshot fails | The selector does not match after prepare, or the element is not visible. | Run `html --root` and search the prepared HTML for the expected element. |
| CSS output exists but values are empty | The selector matches the wrong wrapper or a non-styled element. | Use a more specific `--style` selector or override with `--selector`. |
| Screenshot crop is too broad | The selector points at a layout wrapper instead of the visual widget. | Add a narrower `sections[]` selector or use the exact `styles[]` selector for atom-level crops. |

## See Also

- `css-visual-diff help config-selectors` — Explains screenshot regions and CSS probes in the YAML config.
- `css-visual-diff help artifact-commands` — Lists the single-artifact commands and their formats.
