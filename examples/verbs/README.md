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
