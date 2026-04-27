---
Title: "JavaScript Verbs and Repository-Scanned Workflows"
Slug: javascript-verbs
Short: Turn annotated JavaScript files into typed css-visual-diff CLI commands under the verbs namespace.
Topics:
- javascript
- verbs
- cli
- visual-regression
Commands:
- verbs
Flags:
- repository
- verb-repository
- output
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

`css-visual-diff verbs` turns annotated JavaScript files into CLI commands. It is intended for programmable visual catalog workflows where YAML alone is too static.

## Command namespace

Repository-scanned commands live under:

```bash
css-visual-diff verbs ...
```

They are not injected at the root of the CLI. This keeps scan errors, duplicate command paths, and script-specific flags local to the `verbs` subtree.

Built-ins include:

```bash
css-visual-diff verbs script compare region ...
css-visual-diff verbs script compare brief ...
css-visual-diff verbs catalog inspect-page ...
css-visual-diff verbs catalog inspect-config ...
```

## Repository sources

The lazy `verbs` command discovers scripts from these sources:

1. embedded built-ins,
2. system/user app config,
3. git-root local config (`.css-visual-diff.yml`, `.css-visual-diff.override.yml`),
4. current-working-directory local config (`.css-visual-diff.yml`, `.css-visual-diff.override.yml`),
5. `CSS_VISUAL_DIFF_VERB_REPOSITORIES`,
6. CLI flags.

### CLI repository

```bash
css-visual-diff verbs --repository ./examples/verbs examples catalog inspect-page \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-example \
  --output json
```

`--verb-repository` is an alias for `--repository`.

### Environment repository

```bash
export CSS_VISUAL_DIFF_VERB_REPOSITORIES="/path/to/project/verbs:/path/to/other/verbs"
css-visual-diff verbs custom command ...
```

Use the platform path-list separator (`:` on Linux/macOS).

### Project-local repository config

For repeatable project workflows, put a local config file at the git root:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./verbs
```

Then run the generated command without a bootstrap flag:

```bash
css-visual-diff verbs project command ...
```

Relative paths are resolved relative to the config file that declares them. This means `path: ./verbs` in `/repo/.css-visual-diff.yml` resolves to `/repo/verbs`, even when the command is run from `/repo/packages/button`.

Use `.css-visual-diff.override.yml` for private local repositories that should not be committed. Add it to `.gitignore` when it contains machine-specific or experimental paths.

Current-working-directory config files are also discovered. This lets a monorepo combine a git-root `.css-visual-diff.yml` with a package-local `.css-visual-diff.yml` or `.css-visual-diff.override.yml` when the command is run from that package directory.

### App config repositories

System and user app config can also declare repositories:

```yaml
verbs:
  repositories:
    - name: local
      path: ./verbs
    - name: disabled
      path: ./disabled
      enabled: false
```

Relative paths are resolved relative to the config file.

## Minimal verb file

```js
function hello(name) {
  return `hello ${name}`
}

__verb__("hello", {
  parents: ["custom"],
  output: "text",
  fields: {
    name: { argument: true, required: true }
  }
})
```

Run:

```bash
css-visual-diff verbs --repository ./verbs custom hello Manuel
```

## Sentinels

Supported metadata sentinels:

- `__package__({...})`
- `__section__("slug", {...})`
- `__verb__("functionName", {...})`
- ``doc`...` ``

These are statically scanned to generate commands. At runtime they are installed as no-op helpers so the script can be required safely.

Metadata must be static literal JavaScript: object literals, arrays, strings, numbers, booleans, and null. Do not use computed values, spreads, function calls, or template substitutions inside metadata.

## Package metadata

`__package__` controls default command grouping.

```js
__package__({
  name: "catalog",
  parents: ["examples"],
  short: "Example catalog workflows"
})
```

## Verb metadata

```js
__verb__("inspectPage", {
  parents: ["examples", "catalog"],
  short: "Inspect one page into a catalog",
  output: "structured",
  fields: {
    url: { argument: true, required: true, help: "URL to inspect" },
    selector: { argument: true, required: true, help: "CSS selector" },
    outDir: { argument: true, required: true, help: "Output directory" },
    values: { bind: "all" },
    failOnMissing: { type: "bool", default: false },
  }
})
```

Command names are normalized for CLI usage. A function named `inspectPage` becomes command `inspect-page`.

## Fields and generated flags

Common field keys:

- `type`
- `help`
- `short`
- `default`
- `choices`
- `required`
- `argument` / `arg`
- `bind`
- `section`

Supported field types:

- `string`
- `bool` / `boolean`
- `int` / `integer`
- `float` / `number`
- `stringList` / `list` / `[]string`
- `choice`
- `choiceList`

Argument fields become positional arguments. Other fields become flags.

Example:

```js
fields: {
  url: { argument: true, required: true },
  artifacts: {
    type: "choice",
    choices: ["bundle", "css-json", "html"],
    default: "css-json",
  },
  failOnMissing: { type: "bool", default: false },
}
```

Generates usage like:

```text
verbs examples catalog inspect-page <url> [flags]
  --artifacts css-json
  --failOnMissing
```

## Binding modes

### Positional/default binding

A parameter with the same name as a field receives that value:

```js
function inspect(url, selector, outDir) {}
```

### `bind: "all"`

Passes one object containing all resolved field values:

```js
function inspectPage(url, selector, outDir, values) {
  if (values.failOnMissing) { /* ... */ }
}

__verb__("inspectPage", {
  fields: {
    url: { argument: true, required: true },
    selector: { argument: true, required: true },
    outDir: { argument: true, required: true },
    values: { bind: "all" },
    failOnMissing: { type: "bool", default: false },
  }
})
```

### `bind: "context"`

Passes invocation context:

```js
function debug(ctx) {
  return ctx
}

__verb__("debug", {
  fields: { ctx: { bind: "context" } }
})
```

Context contains verb name, function name, module path, source file, root directory, raw values, and section values.

### Section binding

Sections group related flags.

```js
__section__("target", {
  fields: {
    url: { type: "string", required: true },
    waitMs: { type: "int", default: 0 },
  }
})

function run(target) {
  return target.url
}

__verb__("run", {
  fields: { target: { bind: "target" } }
})
```

## Output modes

Default output is structured/Glazed output. Return objects or arrays to get rows that can be formatted as table/json/yaml/csv.

```bash
css-visual-diff verbs ... --output json
```

For plain text commands:

```js
__verb__("brief", {
  output: "text",
  fields: { /* ... */ }
})
```

Common output aliases include:

- `structured`
- `glaze`
- `table`
- `text`
- `raw`
- `writer`
- `plain`

Prefer structured output for catalog workflows and write large artifacts to files.

## Built-in catalog commands

### Inspect one page

```bash
css-visual-diff verbs catalog inspect-page \
  http://127.0.0.1:8767/ '#cta' /tmp/cssvd-page \
  --slug cta \
  --name CTA \
  --artifacts css-json \
  --output json
```

Authoring mode: selector misses are recorded and returned as structured rows without failing:

```bash
css-visual-diff verbs catalog inspect-page \
  http://127.0.0.1:8767/ '#missing' /tmp/cssvd-authoring \
  --slug missing \
  --output json
```

CI mode: selector misses write manifest/index and exit non-zero:

```bash
css-visual-diff verbs catalog inspect-page \
  http://127.0.0.1:8767/ '#missing' /tmp/cssvd-ci \
  --slug missing \
  --failOnMissing \
  --output json
```

### Inspect a YAML config

```bash
css-visual-diff verbs catalog inspect-config \
  ./page.css-visual-diff.yml original /tmp/cssvd-config \
  --artifacts css-json \
  --output json
```

`inspect-config` loads the Go YAML config, picks `original` or `react`, derives probes from `styles` first and `sections` second, then writes a catalog manifest/index.

## Duplicate command paths

If two repositories define the same generated command path, command construction fails with a clear error such as:

```text
duplicate jsverb path "script compare region" from builtin:scripts/compare.js and /repo/verbs/compare.js
```

Resolve by changing `parents`, function name, or `__verb__({ name/command })` metadata.

## Migration note

Earlier prototypes injected generated JavaScript commands at the root command. The supported shape is now:

```bash
css-visual-diff verbs ...
```

This avoids surprising root command conflicts and makes script repository problems local to the `verbs` subtree.

## Replay and smoke scripts

The ticket for the Goja/jsverbs work includes reproducible scripts under:

```text
ttmp/2026/04/24/CSSVD-GOJA-JS-API--design-goja-javascript-api-for-programmable-visual-catalog-workflows/scripts/
```

Useful binary smokes:

```bash
scripts/006-binary-help-smoke.sh
scripts/007-binary-js-api-success-smoke.sh
scripts/008-binary-js-api-typed-error-smoke.sh
scripts/009-binary-catalog-smoke.sh
scripts/010-binary-built-in-catalog-inspect-page-smoke.sh
scripts/011-binary-built-in-catalog-inspect-config-smoke.sh
```
