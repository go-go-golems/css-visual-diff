---
Title: Artifact Commands for Selector Tuning
Slug: artifact-commands
Short: Use screenshot, css-md, css-json, html, and inspect-json to produce one artifact at a time while tuning selectors.
Topics:
- inspect
- artifacts
- selectors
Commands:
- screenshot
- css-md
- css-json
- html
- inspect-json
Flags:
- config
- side
- output-file
- section
- style
- selector
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

Artifact commands are thin shortcuts around `inspect --format ... --output-file ...`. They exist for fast loops where you want one file and do not want a full artifact directory.

## Command Summary

Each command loads the config, prepares one side, resolves one selector, and writes exactly one output file.

| Command | Output | Use when |
| --- | --- | --- |
| `screenshot` | PNG image | You are tuning the visual crop or checking rendered appearance. |
| `css-md` | Markdown table | You are manually reviewing computed CSS values. |
| `css-json` | JSON object | You are scripting around computed CSS values. |
| `html` | Prepared HTML | You are debugging prepare output or missing selectors. |
| `inspect-json` | DOM/tree JSON | You need structured DOM, bounds, and computed summaries. |

## Screenshot Example

A screenshot command writes one PNG for the selected root, section, style, or ad-hoc selector.

```bash
css-visual-diff screenshot \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --style button-primary \
  --output-file /tmp/button-primary.png
```

Use `--section` when the screenshot crop is a named region:

```bash
css-visual-diff screenshot \
  --config page.css-visual-diff.yml \
  --side original \
  --section nav \
  --output-file /tmp/nav-original.png
```

## CSS Markdown Example

`css-md` writes a human-readable table of configured computed properties. For `--style`, the property list comes from the YAML entry unless you pass `--props`.

```bash
css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side original \
  --style badge-confirmed \
  --output-file /tmp/badge-confirmed-original.md
```

Override properties during exploration:

```bash
css-visual-diff css-md \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side react \
  --selector "[data-comp='button-primary'] button" \
  --props height,padding,border-radius,background-color,color \
  --output-file /tmp/button-probe.md
```

## Prepared HTML Example

`html --root` is the first command to run when a selector is missing. It tells you what the page looked like after prepare ran.

```bash
css-visual-diff html \
  --config examples/pyxis-atoms-prototype-vs-storybook.yaml \
  --side original \
  --root \
  --output-file /tmp/original-root.html
```

Search the output for the selector you expected. If the element is missing from prepared HTML, the problem is the URL, wait condition, or prepare script rather than the screenshot command.

## Equivalent Inspect Commands

The artifact commands are convenience wrappers. These two commands are equivalent:

```bash
css-visual-diff screenshot --config file.yml --side react --style button --output-file /tmp/button.png

css-visual-diff inspect --config file.yml --side react --style button --format png --output-file /tmp/button.png
```

Use `inspect` without `--output-file` when you want a bundle containing screenshot, prepared HTML, CSS JSON, CSS Markdown, inspect JSON, and metadata.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Artifact command asks for `--output-file` | Single-artifact commands always write one explicit file. | Add `--output-file /tmp/name.ext`, or use `inspect --out DIR` for bundle output. |
| `--all-styles` is unavailable | Artifact commands produce one file and do not support batch selectors. | Use `inspect --all-styles --out DIR`. |
| CSS Markdown has unexpected values | The selector matches a wrapper or different child than intended. | Re-run with `--selector` and a more specific CSS selector. |
| HTML output is too small | You used `--style` or `--selector`, so HTML is scoped to that element. | Use `--root` to capture the prepared root HTML. |

## See Also

- `css-visual-diff help inspect-workflow` — Step-by-step workflow for validating a config before comparison.
- `css-visual-diff help config-selectors` — Selector model for screenshot regions and CSS probes.
