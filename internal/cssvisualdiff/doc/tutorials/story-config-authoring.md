---
Title: "Author a Story css-visual-diff Config with Inspect"
Slug: story-config-authoring
Short: Iterate from one Storybook story to a correct co-located .css-visual-diff.yml file, then run single-file and directory comparisons.
Topics:
- visual-regression
- storybook
- config
- inspect
- co-located-configs
Commands:
- inspect
- screenshot
- css-md
- html
- run
Flags:
- config
- config-dir
- side
- section
- style
- selector
- output-file
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial shows the practical loop for turning one Storybook story into a reliable `*.css-visual-diff.yml` config. You start with a small co-located file, use `inspect` and the single-artifact commands until both sides render the intended element, then run the real comparison. After several stories have configs, use `run --config-dir` to scan the component tree and run them as a batch.

The important rule is: do not start by writing the perfect config. Start with a minimal config, inspect one side at a time, and only promote selectors into `sections[]` and `styles[]` after the artifacts prove they are correct.

## Directory Layout

Place each visual-diff config near the component or page story it describes. This keeps selector knowledge close to the source code and gives `run --config-dir` a predictable scan target.

```text
web/packages/pyxis-components/src/atoms/Button/
├── Button.tsx
├── Button.stories.tsx
├── button-primary.css-visual-diff.yml
├── button-loading.css-visual-diff.yml
└── button-disabled.css-visual-diff.yml
```

Use the suffixes that the scanner understands:

```text
*.css-visual-diff.yml
*.css-visual-diff.yaml
```

Do not name runnable story configs `.css-visual-diff.yml`. That exact name is reserved for a future project-level file and is skipped by directory scanning.

## Step 1: Start with One Story and One Small Config

Pick one story and one state. A focused config is easier to inspect than a giant file containing every variant.

```yaml
metadata:
  slug: atoms-button-primary
  title: Button primary visual parity

original:
  name: prototype
  url: http://localhost:7070/Pyxis%20Public%20Site.html
  wait_ms: 1000
  viewport: { width: 1200, height: 240 }
  root_selector: "#atom-capture-root"
  prepare:
    type: script
    wait_for: "window.React && window.ReactDOM && window.Btn"
    script: |
      document.body.innerHTML = '<div id="atom-capture-root"></div>';
      document.body.style.margin = '0';
      const root = document.getElementById('atom-capture-root');
      root.style.padding = '24px';
      ReactDOM.createRoot(root).render(
        React.createElement(Btn, { variant: 'primary', iconRight: 'chev' }, 'Get tickets')
      );
    after_wait_ms: 500

react:
  name: storybook
  url: http://localhost:6006/iframe.html?id=atoms-button--default&viewMode=story
  wait_ms: 1000
  viewport: { width: 1200, height: 240 }
  root_selector: "#storybook-root"

sections:
  - name: button
    selector_original: "#atom-capture-root button"
    selector_react: "#storybook-root button"

styles:
  - name: button
    selector_original: "#atom-capture-root button"
    selector_react: "#storybook-root button"
    include_bounds: true
    props:
      - height
      - padding
      - background-color
      - color
      - border
      - border-radius
      - font-size
      - font-weight
      - line-height
      - gap

output:
  dir: .css-visual-diff/out/atoms/Button/button-primary
  write_json: true
  write_markdown: true
  write_pngs: true
  write_prepared_html: true
  write_inspect_json: true
  validate_pngs: true

modes: [capture, cssdiff, pixeldiff, html-report]
```

This config still uses the current low-level schema. `sections[]` are screenshot regions. `styles[]` are computed-CSS probes. They may use the same selector for simple atoms, but they can diverge when a wrapper is the correct screenshot crop and a nested element carries the visual CSS.

## Step 2: Inspect Prepared HTML First

Prepared HTML answers whether the browser sees the DOM you expected after navigation, waits, and prepare hooks. Check both sides before looking at screenshots.

```bash
css-visual-diff html \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side original \
  --root \
  --output-file /tmp/button-primary-original-root.html

css-visual-diff html \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --root \
  --output-file /tmp/button-primary-react-root.html
```

Open the files and search for the expected element. If the element is absent on the original side, fix `original.url`, `wait_ms`, `prepare.wait_for`, or the prepare script. If the element is absent on the React side, fix the Storybook iframe URL, story ID, wait time, or `root_selector`.

Do not tune screenshot selectors until prepared HTML proves that the element exists.

## Step 3: Tune the Screenshot Region

A screenshot region should crop the visual thing you want to compare. For a button, the exact `button` element is often right. For a card or page section, a wrapper may be better.

```bash
css-visual-diff screenshot \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side original \
  --section button \
  --output-file /tmp/button-primary-original.png

css-visual-diff screenshot \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --section button \
  --output-file /tmp/button-primary-react.png
```

Review the PNGs. The crop should include the intended component and not much else. If the crop is wrong, try an ad-hoc selector first:

```bash
css-visual-diff screenshot \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --selector "#storybook-root button[data-variant='primary']" \
  --output-file /tmp/button-primary-react-candidate.png
```

When the ad-hoc selector produces the right crop, copy it back into the `sections[]` entry as `selector_original`, `selector_react`, or shared `selector`.

## Step 4: Tune the CSS Probe

The CSS probe should point at the element that owns the properties you care about. For a button, that is usually the `button`. For a composed widget, it may be a nested label, icon, input, or action element.

```bash
css-visual-diff css-md \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side original \
  --style button \
  --output-file /tmp/button-primary-original-css.md

css-visual-diff css-md \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --style button \
  --output-file /tmp/button-primary-react-css.md
```

Compare the Markdown files. They show whether the selector exists, the element bounds, and the computed property values. If the values look empty or irrelevant, the selector probably points at a wrapper. Try a more specific probe:

```bash
css-visual-diff css-md \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --selector "#storybook-root button > span.label" \
  --props font-size,font-weight,line-height,color \
  --output-file /tmp/button-primary-label-css.md
```

When the probe is right, promote it into `styles[]`. It is normal for one screenshot `section` to have several `styles` probes when the story has multiple important sub-elements.

## Step 5: Inspect a Full Bundle When the Selector Looks Right

A bundle gives you all single-side artifacts together: metadata, screenshot, prepared HTML, computed CSS, and DOM inspection JSON.

```bash
css-visual-diff inspect \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --side react \
  --style button \
  --out /tmp/css-visual-diff-inspect/atoms/Button/button-primary/react
```

Use bundles when a one-file artifact is not enough to explain a mismatch. The metadata records which selector source was used, and `inspect.json` is useful when you need structured bounds, attributes, or DOM data.

## Step 6: Dry-Run the Config

Before launching browser-heavy comparison modes, validate that the config loads and the mode list resolves.

```bash
css-visual-diff run \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --dry-run
```

A dry run does not prove selectors match, because that is what the inspect steps were for. It proves that the YAML is valid, the command can load the file, and the requested modes are recognized.

## Step 7: Run the Proper Comparison for One Story

After both sides inspect correctly, run the real comparison for that one config.

```bash
css-visual-diff run \
  --config web/packages/pyxis-components/src/atoms/Button/button-primary.css-visual-diff.yml \
  --modes capture,cssdiff,pixeldiff,html-report
```

Review the configured output directory:

```text
.css-visual-diff/out/atoms/Button/button-primary/
├── capture.json
├── cssdiff.json
├── cssdiff.md
├── original.png
├── react.png
├── sections/
├── html-report/
└── ...
```

Use the HTML report when you want a single static place to inspect screenshots, diffs, and Markdown outputs.

## Step 8: Add More Story Configs in Subdirectories

Once one story works, repeat the same pattern for neighboring stories. Keep each file focused on one story or one coherent state group.

```text
web/packages/pyxis-components/src/atoms/
├── Button/
│   ├── button-primary.css-visual-diff.yml
│   ├── button-loading.css-visual-diff.yml
│   └── button-disabled.css-visual-diff.yml
├── Badge/
│   ├── badge-confirmed.css-visual-diff.yml
│   └── badge-cancelled.css-visual-diff.yml
└── Input/
    ├── input-default.css-visual-diff.yml
    └── input-error.css-visual-diff.yml
```

Give every config its own `metadata.slug` and `output.dir`. Stable output directories make it much easier to compare runs over time.

## Step 9: Run the Directory Scan

When the tree contains several co-located configs, use the same `run` verb with `--config-dir`.

Start with a dry run:

```bash
css-visual-diff run \
  --config-dir web/packages/pyxis-components/src/atoms \
  --dry-run \
  --output json
```

Then run the comparison modes:

```bash
css-visual-diff run \
  --config-dir web/packages/pyxis-components/src/atoms \
  --modes capture,cssdiff,pixeldiff,html-report
```

The scanner recursively finds `*.css-visual-diff.yml` and `*.css-visual-diff.yaml` files, sorts them for deterministic execution, and runs each config through the same execution path as `--config`.

## Working Loop Summary

Use this loop for each new story:

1. Create `story-name.css-visual-diff.yml` next to the story.
2. Add minimal `original`, `react`, one `section`, one `style`, `output`, and `modes`.
3. Run `html --root` for both sides.
4. Run `screenshot --section` for both sides.
5. Run `css-md --style` for both sides.
6. Promote good ad-hoc `--selector` values back into the YAML.
7. Run `run --config ... --dry-run`.
8. Run `run --config ... --modes capture,cssdiff,pixeldiff,html-report`.
9. After several files exist, run `run --config-dir ...`.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `no *.css-visual-diff.yml/.yaml configs found` | The files do not use the scanner suffix, or the root is too narrow. | Rename files to `name.css-visual-diff.yml` or point `--config-dir` at a higher directory. |
| The scan skips `.css-visual-diff.yml` | That exact name is reserved for project-level settings. | Use a runnable story name such as `button-primary.css-visual-diff.yml`. |
| Prepared HTML has no target element | URL, wait condition, Storybook ID, or prepare hook is wrong. | Fix the target setup before changing selectors. |
| Screenshot includes too much layout | The section selector points at a broad wrapper. | Try `screenshot --selector ...`; promote the better selector into `sections[]`. |
| CSS values look unrelated | The style selector points at a wrapper or child without the relevant visual styles. | Try a more specific `css-md --selector ... --props ...`; promote it into `styles[]`. |
| One side needs a different selector | Original and React DOM structures differ. | Use `selector_original` and `selector_react` instead of shared `selector`. |
| Batch run stops on the first failure | `run --config-dir` currently fails fast. | Run the failing config with `--config` and inspect it; add continue-on-error only when batch triage needs it. |

## See Also

- `css-visual-diff help inspect-workflow` — The shorter inspect-before-compare loop.
- `css-visual-diff help config-selectors` — Explains `sections[]`, `styles[]`, and `run --config-dir` scanning.
- `css-visual-diff help artifact-commands` — Lists the single-file artifact commands for selector tuning.
