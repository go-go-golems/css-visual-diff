# css-visual-diff xgoja command-provider smoke

This example builds a generated xgoja binary that mounts `css-visual-diff.verbs`
as `css` and runs the built-in `script compare region` workflow through the
command provider.

The smoke compares two local HTML fixtures and asserts that JSON and Markdown
artifacts are written.

Run:

```bash
make smoke
```
