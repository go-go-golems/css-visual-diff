# css-visual-diff external verb examples

This folder is a small repository-scanned verb example. It is intentionally kept outside the embedded built-ins so operators can see how their own project-local verb folders can be wired in with `--repository`.

All examples use the canonical JavaScript API names:

- `cvd.compare.region(...)` for the quick region comparison path.
- `page.locator(...).collect(...)` and `cvd.compare.selections(...)` for the primitive collect-and-analyze path.
- `cvd.snapshot.page(...)` for probe snapshots.
- `cvd.catalog.create(...)` for catalog workflows.
- `cvd.config.load(...)` when loading YAML config files.

## Quick region comparison

Use `examples compare region` when you want the low-effort visual answer: load two pages, compare one selector, and write screenshots, pixel diff PNGs, JSON, and Markdown.

```bash
css-visual-diff verbs --repository examples/verbs examples compare region \
  http://127.0.0.1:8767/left.html \
  http://127.0.0.1:8767/right.html \
  '#cta' \
  /tmp/cssvd-compare-region \
  --output json
```

The command writes:

- `/tmp/cssvd-compare-region/left_region.png`
- `/tmp/cssvd-compare-region/right_region.png`
- `/tmp/cssvd-compare-region/diff_only.png`
- `/tmp/cssvd-compare-region/diff_comparison.png`
- `/tmp/cssvd-compare-region/compare.json`
- `/tmp/cssvd-compare-region/compare.md`

## Collect and analyze

Use `examples compare collect-and-analyze` when you want JavaScript policy logic over collected data. This example collects selector facts from both pages, compares them with `cvd.compare.selections(...)`, filters typography and class diffs in JavaScript, and writes JSON/Markdown evidence.

```bash
css-visual-diff verbs --repository examples/verbs examples compare collect-and-analyze \
  http://127.0.0.1:8767/left.html \
  http://127.0.0.1:8767/right.html \
  '#cta' \
  /tmp/cssvd-collect-analyze \
  --output json
```

The command writes:

- `/tmp/cssvd-collect-analyze/compare.json`
- `/tmp/cssvd-collect-analyze/compare.md`

## Multi-section page comparison catalog

Use `examples compare page-catalog` when you want the reusable project pattern: load two pages once, wait for each important selector, compare multiple sections, write per-section artifacts, record each comparison in a catalog, and return compact JSON for CI or agents.

```bash
css-visual-diff verbs --repository examples/verbs examples compare page-catalog \
  http://127.0.0.1:8767/left.html \
  http://127.0.0.1:8767/right.html \
  /tmp/cssvd-page-catalog \
  --output json
```

The command writes:

- `/tmp/cssvd-page-catalog/manifest.json`
- `/tmp/cssvd-page-catalog/index.md`
- `/tmp/cssvd-page-catalog/artifacts/page/diff_comparison.png`
- `/tmp/cssvd-page-catalog/artifacts/page/compare.json`
- `/tmp/cssvd-page-catalog/artifacts/cta/diff_comparison.png`
- `/tmp/cssvd-page-catalog/artifacts/cta/compare.json`

It also demonstrates `locator.waitFor(...)` and the stable artifact path map returned by `comparison.artifacts.write(...)`.

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

The `examples low-level inspect` command shows the lower-level script-native API. It uses `page.locator(...)`, `cvd.extract(...)`, `cvd.probe(...)`, `cvd.snapshot.page(...)`, and `cvd.write.json(...)` without writing the standard inspect artifact bundle.

```bash
css-visual-diff verbs --repository examples/verbs examples low-level inspect \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-low-level \
  --output json
```

The command writes:

- `/tmp/cssvd-low-level/element.json`
- `/tmp/cssvd-low-level/snapshot.json`

Use the compare examples while iterating on pixel-perfect UI feedback. Use the catalog example when you want durable manifests, indexes, and standard inspect artifacts.

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
