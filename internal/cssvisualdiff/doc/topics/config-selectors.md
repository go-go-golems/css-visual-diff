---
Title: "Config Selectors: Regions and CSS Probes"
Slug: config-selectors
Short: Understand how css-visual-diff uses sections as screenshot regions and styles as computed-CSS probes.
Topics:
- config
- selectors
- visual-regression
Commands:
- inspect
- run
- screenshot
- css-md
Flags:
- section
- style
- selector
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

A css-visual-diff config can hold many named selectors. The current schema separates screenshot regions from CSS probes because screenshots and computed-style inspection need different selector precision.

## Screenshot Regions

Screenshot regions live under `sections[]`. A section defines the visual crop used by capture, pixel diff, and report views. It is usually a wrapper or page area rather than the exact element that owns every CSS property.

```yaml
sections:
  - name: button-primary
    selector_original: "[data-comp='button-primary']"
    selector_react: "[data-comp='button-primary']"
```

Use a section when you care about the whole visible widget, including layout, whitespace, labels, icons, and nested content.

## CSS Probes

CSS probes live under `styles[]`. A style entry defines the exact element whose computed properties should be captured. It is often more specific than the screenshot region.

```yaml
styles:
  - name: button-primary
    selector_original: "[data-comp='button-primary'] button"
    selector_react: "[data-comp='button-primary'] button"
    include_bounds: true
    props: [height, padding, background-color, color, border-radius]
```

Use a style when you want to inspect property values such as `padding`, `font-size`, `border-radius`, `color`, or `gap`.

## Why They Are Separate

The best screenshot selector is not always the best CSS selector. A widget wrapper may be the right screenshot crop, while the nested `button`, `input`, `svg`, or `h1` is the element that carries the important computed CSS.

A practical pattern is:

```yaml
sections:
  - name: card
    selector: "[data-card='artist']"

styles:
  - name: card-title
    selector: "[data-card='artist'] h2"
    props: [font-family, font-size, font-weight, color]
  - name: card-action
    selector: "[data-card='artist'] button"
    props: [height, padding, background-color, border-radius]
```

## Inspecting Selectors

Use `screenshot` for the crop and `css-md` for the computed CSS values.

```bash
css-visual-diff screenshot --config card.css-visual-diff.yml --side react --section card --output-file /tmp/card.png
css-visual-diff css-md --config card.css-visual-diff.yml --side react --style card-action --output-file /tmp/card-action.md
```

For ad-hoc tuning, bypass named entries with `--selector` and `--props`:

```bash
css-visual-diff css-md \
  --config card.css-visual-diff.yml \
  --side react \
  --selector "[data-card='artist'] button" \
  --props height,padding,background-color,border-radius \
  --output-file /tmp/probe.md
```

## Future Region Format

The planned authoring format can merge both ideas into a `regions[]` list where each region has one screenshot selector and nested CSS probes. The current implementation keeps `sections[]` and `styles[]` for compatibility and treats them as the low-level execution schema.

```yaml
regions:
  - name: button-primary
    screenshot:
      selector: "[data-comp='button-primary']"
    css:
      - name: button
        selector: "[data-comp='button-primary'] button"
        props: [height, padding, background-color, border-radius]
```

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Pixel diff looks wrong but CSS values match | The screenshot region includes surrounding layout differences. | Narrow the `sections[]` selector or add a focused section. |
| CSS values do not explain the visual issue | The style selector points at a wrapper instead of the styled child. | Add a nested `styles[]` probe for the exact element. |
| Same selector works on one side only | Original and React DOM structures differ. | Use `selector_original` and `selector_react` instead of shared `selector`. |
| Too many selectors in one file | A giant config is hard to maintain. | Co-locate smaller `*.css-visual-diff.yml` files near components. |

## See Also

- `css-visual-diff help inspect-workflow` — Shows the recommended inspect-before-compare loop.
- `css-visual-diff help artifact-commands` — Explains single-file artifact commands for selector tuning.
