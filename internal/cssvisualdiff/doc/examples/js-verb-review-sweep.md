---
Title: JS verb review sweep example
Slug: js-verb-review-sweep
Short: Run a project-scale visual review sweep through a JavaScript verb.
Topics:
- javascript-verbs
- review-site
Commands:
- verbs
Flags: []
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Example
---

Use JavaScript verbs for project-scale comparison suites. The repository includes an external example that reads a project spec, compares declared regions, writes per-region artifacts, and emits a review-site `summary.json`.

```bash
css-visual-diff verbs --repository examples/verbs \
  examples review-sweep from-spec \
  --specFile examples/specs/review-sweep.example.yaml \
  --outDir /tmp/example-review

css-visual-diff serve --data-dir /tmp/example-review --port 8098
```

This is the replacement pattern for the removed native YAML runner: keep project data in userland, but orchestrate browser work and artifact generation from JavaScript.
