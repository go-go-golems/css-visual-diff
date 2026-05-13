# css-visual-diff external verb examples

This folder is a small repository-scanned verb example. It is intentionally kept outside the embedded built-ins so operators can see how their own project-local verb folders can be wired in with `--repository`.

## Inspect one page into a catalog

Start or choose a local page, then run:

```bash
css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-example \
  --slug cta \
  --artifacts css-json \
  --output json
```

The command writes:

- `/tmp/cssvd-example/manifest.json`
- `/tmp/cssvd-example/index.md`
- `/tmp/cssvd-example/artifacts/cta/computed-css.json`

## Lower-level locator/extractor/snapshot example

The `examples low-level inspect` command shows the newer script-native API. It uses `page.locator(...)`, `cvd.extract(...)`, `cvd.probe(...)`, `cvd.snapshot(...)`, and `cvd.write.json(...)` without writing the standard inspect artifact bundle.

```bash
css-visual-diff verbs --repository examples/verbs examples low-level inspect \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-low-level \
  --output json
```

The command writes:

- `/tmp/cssvd-low-level/element.json`
- `/tmp/cssvd-low-level/snapshot.json`

Use this example when you want a small programmable feedback loop while building UI components. Use the catalog example when you want durable manifests, indexes, and standard inspect artifacts.

## Overlay screenshot examples

The overlay examples use `examples/pages/overlay-components.html` as a deterministic local fixture.

Start a static server from the repository root:

```bash
python3 -m http.server 8767
```

Export one annotated full-page PNG:

```bash
css-visual-diff verbs --repository examples/verbs examples overlay annotated-png \
  http://127.0.0.1:8767/examples/pages/overlay-components.html \
  /tmp/cssvd-overlay-example \
  --output json
```

This writes:

- `/tmp/cssvd-overlay-example/full-page.organisms.annotated.png`

Export a small component-system gallery with extracted component screenshots, a full-page organism map, and a cropped Hero-parts overlay:

```bash
css-visual-diff verbs --repository examples/verbs examples overlay gallery \
  http://127.0.0.1:8767/examples/pages/overlay-components.html \
  /tmp/cssvd-overlay-gallery \
  --output json
```

This writes:

- `/tmp/cssvd-overlay-gallery/index.html`
- `/tmp/cssvd-overlay-gallery/components.json`
- `/tmp/cssvd-overlay-gallery/annotated/full-page.organisms.png`
- `/tmp/cssvd-overlay-gallery/annotated/hero.parts.crop.png`

Open the HTML file in a browser and inspect the PNGs to review label placement, crop bounds, and component grouping.

## Authoring mode: keep going on missing selectors

By default `failOnMissing=false`, so selector misses are recorded in the manifest and returned as a structured row instead of making the command fail:

```bash
css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
  http://127.0.0.1:8767/ '#missing' /tmp/cssvd-authoring \
  --slug missing \
  --output json
```

## CI mode: fail on missing selectors

For CI, pass `--failOnMissing` so selector misses still write the manifest/index but exit non-zero:

```bash
css-visual-diff verbs --repository examples/verbs examples catalog inspect-page \
  http://127.0.0.1:8767/ '#missing' /tmp/cssvd-ci \
  --slug missing \
  --failOnMissing \
  --output json
```
