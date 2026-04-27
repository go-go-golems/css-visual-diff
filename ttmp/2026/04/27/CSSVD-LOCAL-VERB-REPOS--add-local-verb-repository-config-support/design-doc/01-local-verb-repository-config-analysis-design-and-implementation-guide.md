---
Title: Local Verb Repository Config Analysis, Design, and Implementation Guide
Ticket: CSSVD-LOCAL-VERB-REPOS
Status: active
Topics:
    - css-visual-diff
    - config
    - glazed
    - cli
    - javascript-api
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../glazed/pkg/config/plan_sources.go
      Note: Defines GitRootFile and WorkingDirFile source helpers
    - Path: ../../../../../../../go-minitrace/pkg/minitracecmd/repositories.go
      Note: Reference implementation for local repository config layers
    - Path: ../../../../../../../pinocchio/pkg/cmds/profilebootstrap/profile_selection.go
      Note: Reference for Pinocchio local overlay config plan
    - Path: README.md
      Note: User-facing quickstart that currently documents runtime --repository usage
    - Path: internal/cssvisualdiff/doc/topics/javascript-verbs.md
      Note: Embedded help topic that must document local repository config
    - Path: internal/cssvisualdiff/verbcli/bootstrap.go
      Note: Primary implementation target for repository discovery
    - Path: internal/cssvisualdiff/verbcli/command.go
      Note: Lazy verbs command builds generated Cobra commands after bootstrap discovery
ExternalSources: []
Summary: Design guide for adding .css-visual-diff.yml and .css-visual-diff.override.yml local verb repository discovery, modeled on go-minitrace and Pinocchio.
LastUpdated: 2026-04-27T17:52:24-04:00
WhatFor: Use this when implementing local project config discovery for css-visual-diff JavaScript verb repositories.
WhenToUse: Before changing internal/cssvisualdiff/verbcli bootstrap/config code or documenting local verb repository setup.
---


# Local Verb Repository Config Analysis, Design, and Implementation Guide

## Executive summary

`css-visual-diff` already has a repository-scanned JavaScript verb subsystem. Users can run built-in verbs such as `css-visual-diff verbs catalog inspect-page`, can add external verb folders with `--repository`, can set `CSS_VISUAL_DIFF_VERB_REPOSITORIES`, and can declare app-level repositories in the normal system/user Glazed config files. The missing piece is project-local discovery: a repo should be able to check in a small `.css-visual-diff.yml` file that points at its own verb repository, and an individual developer should be able to add `.css-visual-diff.override.yml` for local-only additions.

The implementation committed in `5fd1c68519662dafbccf1dc34cb05e90298eba32` is intentionally close to the recent `go-minitrace` change documented in Manuel's Obsidian note:

```text
/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/27/PROJ - go-minitrace - Local Query Repository Config.md
```

The equivalent `css-visual-diff` discovery chain should become:

```text
embedded built-in verb repository
  -> system/user app config repositories
  -> git-root .css-visual-diff.yml
  -> git-root .css-visual-diff.override.yml
  -> cwd .css-visual-diff.yml
  -> cwd .css-visual-diff.override.yml
  -> CSS_VISUAL_DIFF_VERB_REPOSITORIES
  -> verbs --repository / --verb-repository flags
```

The recommended user-facing local config shape is the shape the project already understands:

```yaml
verbs:
  repositories:
    - name: project
      path: ./verbs
    - name: shared-components
      path: ./tools/cssvd-verbs
      enabled: true
```

Relative `path` entries must continue to resolve relative to the config file that declared them, not relative to the shell's current directory. This is already true for app config file loading, so the main implementation task is not path normalization; it is adding repo/CWD config sources to the Glazed config plan and proving the behavior with root and nested-directory tests.

## Problem statement and scope

### The operator problem

Project-specific visual-diff workflows often need JavaScript verbs that live next to the project under test. Today, an operator has to remember one of these forms:

```bash
css-visual-diff verbs --repository ./examples/verbs examples catalog inspect-page ...
```

or:

```bash
export CSS_VISUAL_DIFF_VERB_REPOSITORIES=/absolute/path/to/project/verbs
css-visual-diff verbs examples catalog inspect-page ...
```

That is workable for one-off experiments, but it is brittle for repeated visual-review projects:

- commands in docs become longer and easier to mistype;
- nested working directories change the meaning of relative `--repository ./verbs` paths;
- every new shell needs the environment variable again;
- one developer's private verb folder should not have to be committed into shared config;
- automated scripts should not need custom environment setup when the repository itself can declare its command catalog.

### The implementation problem

`css-visual-diff` already has most of the pieces:

- a lazy `verbs` Cobra subtree that discovers command repositories only when `verbs` is invoked;
- a `Repository` model for embedded and filesystem-backed verb roots;
- config file parsing for `verbs.repositories`;
- relative path resolution from each config file's directory;
- dedupe by normalized repository root;
- tests around CLI repository parsing and app-config repository parsing.

The gap is that `loadConfigRepositories` only resolves system and user config files. It does not ask Glazed for git-root or current-working-directory config overlays.

### Scope for this ticket

This ticket started as a design and implementation guide, and the core implementation is now committed as `5fd1c68519662dafbccf1dc34cb05e90298eba32`. The implemented scope is:

1. Add local config file constants:
   - `.css-visual-diff.yml`
   - `.css-visual-diff.override.yml`
2. Extend the Glazed config plan in `internal/cssvisualdiff/verbcli/bootstrap.go` from system/user only to system/user/repo/cwd.
3. Keep the existing `verbs.repositories` document shape.
4. Keep the existing relative path behavior.
5. Add tests that prove:
   - git-root config is discovered;
   - git-root override config is discovered;
   - CWD config is discovered;
   - CWD override config is discovered;
   - relative repository paths are resolved from each declaring config file;
   - running from a nested directory still finds git-root config correctly.
6. Update user-facing docs for `css-visual-diff help javascript-verbs` and `README.md`.

Still out of scope after the first implementation:

- changing JavaScript verb metadata syntax;
- changing duplicate command-path handling;
- adding a `css-visual-diff config explain` command;
- implementing Pinocchio's full unified config document model;
- changing run-config discovery for `--config-dir`.

## Current-state architecture, with evidence

### Repository-scanned verbs are scoped under `css-visual-diff verbs`

The root executable wires `verbcli.NewLazyCommand()` into the Cobra tree. That happens in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/cmd/css-visual-diff/main.go:327-380
```

Key evidence:

- `main()` creates the root `css-visual-diff` command at `main.go:334-340`.
- It adds ordinary commands such as `run`, `inspect`, `compare`, and `llm-review` at `main.go:364-373`.
- It adds the lazy verb command at `main.go:374`.

The important design implication is that JavaScript verbs are not injected at process startup. They are resolved only when the user enters the `verbs` namespace.

### The lazy `verbs` command discovers repositories before building generated commands

The lazy command lives in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command.go:16-35
```

The flow is:

1. `NewLazyCommand()` creates a Cobra command with `DisableFlagParsing: true`.
2. In `RunE`, it calls `DiscoverBootstrap(args)`.
3. It receives a `Bootstrap` plus `remainingArgs`.
4. It builds the real command tree with `NewCommand(bootstrap)`.
5. It gives the generated command tree only `remainingArgs`.

This is why repository flags have to be parsed manually before normal Cobra parsing. If a filesystem repository is not discovered before `NewCommand`, the generated command simply does not exist.

### The generated verb tree scans repositories into `jsverbs.Registry` values

The core generated command path is in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command.go:42-80
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:276-320
```

The runtime sequence is:

```text
Bootstrap.Repositories
  -> ScanRepositories
  -> jsverbs.Registry per repository
  -> CollectDiscoveredVerbs
  -> buildCommands
  -> glazedcli.AddCommandsToRootCommand
```

Evidence:

- `newCommandWithInvokerFactory` calls `ScanRepositories` at `command.go:48`.
- It calls `CollectDiscoveredVerbs` at `command.go:52`.
- It turns verbs into Glazed commands at `command.go:56-60`.
- `ScanRepositories` uses `jsverbs.ScanFS` for embedded repositories and `jsverbs.ScanDir` for filesystem repositories at `bootstrap.go:286-290`.
- `CollectDiscoveredVerbs` rejects duplicate verb full paths at `bootstrap.go:302-320`.

### The repository model already supports source tracking

The `Repository` struct is defined in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:30-38
```

It records:

- `Name`: a display/debug name;
- `Source`: where the repository came from, such as `embedded`, `config`, `env`, or `cli`;
- `SourceRef`: the concrete source reference, such as a config file path or env var name;
- `RootDir`: filesystem repository root;
- `EmbeddedFS`, `Embedded`, and `EmbeddedAt`: embedded built-in repository details.

This source tracking is useful for tests, diagnostics, and future `config explain` style commands.

### Existing repository sources

The current source ordering is implemented in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:80-103
```

The current order is:

```text
builtinRepository()
  -> loadConfigRepositories(context.Background())
  -> repositoriesFromEnv(cwd)
  -> cliRepos
```

Evidence:

- `discoverBootstrap` initializes `repositories` with `builtinRepository()` at `bootstrap.go:81`.
- It loads config repositories at `bootstrap.go:84-90`.
- It loads environment repositories at `bootstrap.go:92-98`.
- It appends CLI repositories at `bootstrap.go:99-101`.

This order is part of today's behavior. Because `CollectDiscoveredVerbs` rejects duplicate command paths, this is not an override chain for verb command names. It is a discovery chain for repository roots.

### Config shape already exists

The existing config types are:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:50-62
```

They decode this YAML shape:

```yaml
verbs:
  repositories:
    - name: local
      path: ./verbs
      enabled: true
```

The important fields are:

- `verbs.repositories[].path`: required path to a directory containing JavaScript verb files;
- `verbs.repositories[].name`: optional display name;
- `verbs.repositories[].enabled`: optional boolean; `false` disables the entry.

### App config currently only uses system and user layers

The missing feature is visible in:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:134-158
```

The plan currently uses:

```go
glazedconfig.WithLayerOrder(glazedconfig.LayerSystem, glazedconfig.LayerUser)
```

and only adds:

```go
glazedconfig.SystemAppConfig("css-visual-diff")
glazedconfig.XDGAppConfig("css-visual-diff")
glazedconfig.HomeAppConfig("css-visual-diff")
```

That means the supported config files are effectively:

```text
/etc/css-visual-diff/config.yaml
$XDG_CONFIG_HOME/css-visual-diff/config.yaml
~/.css-visual-diff/config.yaml
```

There is no `.css-visual-diff.yml` git-root lookup and no current working directory lookup.

### Relative repository path resolution is already correct for config files

`loadRepositoriesFromConfigFile` resolves paths relative to the config file's directory:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go:160-185
```

Evidence:

- It sets `baseDir := filepath.Dir(path)` at `bootstrap.go:169`.
- It calls `normalizeFilesystemRepositoryPath(spec.Path, baseDir)` at `bootstrap.go:175`.
- `normalizeFilesystemRepositoryPath` joins non-absolute paths to `baseDir` at `bootstrap.go:262-264`.

This is the most important correctness property for local config. It ensures this file:

```text
/repo/.css-visual-diff.yml
```

with this body:

```yaml
verbs:
  repositories:
    - path: ./verbs
```

means:

```text
/repo/verbs
```

whether the command is run from `/repo`, `/repo/packages/button`, or `/repo/ttmp/...`.

### The docs already describe repositories, but not local files

The embedded help topic lives at:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/doc/topics/javascript-verbs.md
```

It currently states that the lazy `verbs` command discovers scripts from four sources:

1. embedded built-ins;
2. app config;
3. `CSS_VISUAL_DIFF_VERB_REPOSITORIES`;
4. CLI flags.

It also documents CLI flags, environment variables, and app config repositories. The content is accurate for current behavior, but it should be expanded once local config files are added.

The README has a short external repository example at:

```text
/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/README.md
```

The relevant section says external verb repositories can be supplied at runtime and shows:

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples catalog inspect-page ...
```

After this change, README should include the local config equivalent.

## Comparison target: go-minitrace local query repository config

The nearest implementation model is:

```text
/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitracecmd/repositories.go
```

Important pieces:

- local filename constants at `repositories.go:15-21`;
- Glazed system/user/repo/cwd plan at `repositories.go:36-53`;
- config-file-relative path normalization at `repositories.go:83-108` and `repositories.go:142-175`;
- final source-root construction at `repositories.go:177-210`.

The go-minitrace config chain is:

```text
/etc/go-minitrace/config.yaml
~/.go-minitrace/config.yaml
$XDG_CONFIG_HOME/go-minitrace/config.yaml
<git-root>/.go-minitrace.yml
<git-root>/.go-minitrace.override.yml
<cwd>/.go-minitrace.yml
<cwd>/.go-minitrace.override.yml
```

`css-visual-diff` copied the Glazed plan idea, not the exact data model. The data model is already different:

```yaml
# go-minitrace
queryRepositories:
  - ./query-commands

# css-visual-diff
verbs:
  repositories:
    - path: ./verbs
```

## Comparison target: Pinocchio local config overlays

Pinocchio's local config pattern is implemented in:

```text
/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/cmds/profilebootstrap/profile_selection.go:224-252
/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/configdoc/types.go:12-16
```

Important observations:

- Pinocchio's local filenames are `.pinocchio.yml` and `.pinocchio.override.yml`.
- Its Glazed plan includes `LayerSystem`, `LayerUser`, `LayerRepo`, `LayerCWD`, and `LayerExplicit`.
- It uses `GitRootFile(...)` and `WorkingDirFile(...)` for local layers.

Pinocchio's config document also uses an `app.repositories` field. Its merge behavior appends and dedupes repositories rather than replacing them:

```text
/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/configdoc/merge.go:20-22
/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/configdoc/merge.go:55-76
```

For the first `css-visual-diff` implementation, the safer move is to preserve the current css-visual-diff behavior: each resolved config file contributes zero or more repositories, and all enabled entries are appended with repository-root dedupe. A deeper Pinocchio-style unified config merge is optional future work.

## Glazed config plan API reference

The local config feature should use Glazed's existing plan API, not custom git-root discovery.

Relevant Glazed files:

```text
/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/config/plan.go
/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/config/plan_sources.go
```

Key API facts:

- Layers exist for `LayerSystem`, `LayerUser`, `LayerRepo`, `LayerCWD`, and `LayerExplicit` (`plan.go:13-19`).
- `NewPlan` constructs a plan (`plan.go:139-145`).
- `WithLayerOrder` controls source ordering (`plan.go:127-131`).
- `WithDedupePaths` dedupes repeated config paths (`plan.go:133-137`).
- `Plan.Resolve` returns `[]ResolvedConfigFile` plus a report (`plan.go:157-235`).
- `SystemAppConfig`, `XDGAppConfig`, and `HomeAppConfig` discover app config files (`plan_sources.go:18-84`).
- `WorkingDirFile(name)` discovers `<cwd>/<name>` (`plan_sources.go:104-126`).
- `GitRootFile(name)` discovers `<git-root>/<name>` (`plan_sources.go:128-150`).

This means the implementation can stay small: add constants, expand layer order, add four source specs.

## Proposed user-facing behavior

### Shared checked-in config

A repository can add:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./verbs
```

Then all developers can run:

```bash
css-visual-diff verbs project inspect-button --output json
```

from the repo root or any nested directory.

### Local developer override

A developer can add:

```yaml
# .css-visual-diff.override.yml
verbs:
  repositories:
    - name: manuel-scratch
      path: ./scratch/cssvd-verbs
```

Recommended `.gitignore` convention:

```gitignore
.css-visual-diff.override.yml
```

The override file name communicates local/private intent. The implementation should not force gitignore behavior; it should document it.

### CWD-specific override

If a monorepo has a root config and a package-local config, running from the package directory may discover both:

```text
/repo/.css-visual-diff.yml
/repo/packages/button/.css-visual-diff.yml
```

Given:

```yaml
# /repo/.css-visual-diff.yml
verbs:
  repositories:
    - name: shared
      path: ./verbs/shared
```

and:

```yaml
# /repo/packages/button/.css-visual-diff.yml
verbs:
  repositories:
    - name: button
      path: ./verbs
```

running from `/repo/packages/button` should load:

```text
/repo/verbs/shared
/repo/packages/button/verbs
```

because each relative path is anchored to the file that declared it.

## Proposed implementation

### Step 1: Add local file constants

In:

```text
internal/cssvisualdiff/verbcli/bootstrap.go
```

add constants beside the existing env and flag names:

```go
const (
    VerbRepositoriesEnvVar = "CSS_VISUAL_DIFF_VERB_REPOSITORIES"
    RepositoryFlag         = "repository"
    VerbRepositoryFlag     = "verb-repository"

    LocalConfigFileName         = ".css-visual-diff.yml"
    LocalOverrideConfigFileName = ".css-visual-diff.override.yml"
)
```

Naming notes:

- `LocalConfigFileName` is clearer than `LocalOverrideFileName` for the checked-in shared file.
- `LocalOverrideConfigFileName` clearly names the private override file.
- If the team wants exact parity with go-minitrace naming, use `LocalOverrideFileName` and `LocalProjectOverrideFileName`, but the actual filenames should be css-visual-diff specific.

### Step 2: Extend the Glazed config plan

Current code:

```go
plan := glazedconfig.NewPlan(
    glazedconfig.WithLayerOrder(glazedconfig.LayerSystem, glazedconfig.LayerUser),
    glazedconfig.WithDedupePaths(),
).Add(
    glazedconfig.SystemAppConfig("css-visual-diff").Named("system-app-config"),
    glazedconfig.XDGAppConfig("css-visual-diff").Named("xdg-app-config"),
    glazedconfig.HomeAppConfig("css-visual-diff").Named("home-app-config"),
)
```

Proposed code:

```go
plan := glazedconfig.NewPlan(
    glazedconfig.WithLayerOrder(
        glazedconfig.LayerSystem,
        glazedconfig.LayerUser,
        glazedconfig.LayerRepo,
        glazedconfig.LayerCWD,
    ),
    glazedconfig.WithDedupePaths(),
).Add(
    glazedconfig.SystemAppConfig("css-visual-diff").Named("system-app-config").Kind("app-config"),
    glazedconfig.HomeAppConfig("css-visual-diff").Named("home-app-config").Kind("app-config"),
    glazedconfig.XDGAppConfig("css-visual-diff").Named("xdg-app-config").Kind("app-config"),
    glazedconfig.GitRootFile(LocalConfigFileName).Named("git-root-local-css-visual-diff").Kind("verb-repository-overlay"),
    glazedconfig.GitRootFile(LocalOverrideConfigFileName).Named("git-root-local-css-visual-diff-override").Kind("verb-repository-overlay"),
    glazedconfig.WorkingDirFile(LocalConfigFileName).Named("cwd-local-css-visual-diff").Kind("verb-repository-overlay"),
    glazedconfig.WorkingDirFile(LocalOverrideConfigFileName).Named("cwd-local-css-visual-diff-override").Kind("verb-repository-overlay"),
)
```

Why include `.Kind(...)`?

- It is not required for loading.
- It makes the `PlanReport` more useful for debugging.
- It follows the go-minitrace and Pinocchio style.

Why order `HomeAppConfig` before `XDGAppConfig` here?

- go-minitrace uses system, home, XDG, then local.
- current css-visual-diff uses system, XDG, home.
- The implementation should choose deliberately and add a test if order matters.
- If compatibility matters more than parity, preserve current system, XDG, home ordering inside the user layer.

Recommended compatibility-preserving order:

```go
glazedconfig.SystemAppConfig("css-visual-diff").Named("system-app-config").Kind("app-config"),
glazedconfig.XDGAppConfig("css-visual-diff").Named("xdg-app-config").Kind("app-config"),
glazedconfig.HomeAppConfig("css-visual-diff").Named("home-app-config").Kind("app-config"),
...
```

Because `Plan.orderedSources` preserves source order within a layer, keeping XDG before home avoids an accidental behavior change.

### Step 3: Keep `loadRepositoriesFromConfigFile` mostly unchanged

This function already does what local config needs:

```go
baseDir := filepath.Dir(path)
normalized, err := normalizeFilesystemRepositoryPath(spec.Path, baseDir)
```

Do not change this behavior. It is what makes nested directory execution work.

Potential small improvements:

```go
if strings.TrimSpace(spec.Path) == "" {
    return nil, fmt.Errorf("config repository path in %s cannot be empty", path)
}
```

This is optional because `normalizeFilesystemRepositoryPath` already returns `repository path is empty`.

### Step 4: Preserve current dedupe behavior

`appendRepository` dedupes by repository identity:

```go
func repositoryIdentity(repo Repository) string {
    if repo.Embedded {
        return "embedded:" + repo.Name + ":" + repo.EmbeddedAt
    }
    return "path:" + filepath.Clean(repo.RootDir)
}
```

Keep this behavior. If the same repository path appears in both `.css-visual-diff.yml` and `.css-visual-diff.override.yml`, it should only be scanned once.

### Step 5: Decide and document command-path collision behavior

The current behavior rejects duplicate generated command paths:

```go
if prev, ok := seen[key]; ok {
    return nil, fmt.Errorf("duplicate jsverb path %q from %s and %s", ...)
}
```

This should remain unchanged for the first implementation. Local override files should override repository lists by adding local sources; they should not silently override commands with the same path. Silent command override would be surprising and could hide built-in workflows.

If future users want command overriding, it should be a separate design with explicit precedence rules and clear diagnostics.

## Implementation pseudocode

### Repository discovery flow

```text
User runs:
  css-visual-diff verbs [bootstrap flags] <generated command path> [generated command flags]

Cobra lazy command:
  receives raw args because DisableFlagParsing = true

DiscoverBootstrap(args):
  cwd = os.Getwd()
  cliRepos, remainingArgs = repositoriesFromArgs(args, cwd)
  bootstrap = discoverBootstrap(cwd, cliRepos)
  return bootstrap, remainingArgs

discoverBootstrap(cwd, cliRepos):
  repositories = [builtinRepository()]
  seen = identity of builtin

  for repo in loadConfigRepositories(context.Background()):
    appendRepository(repositories, seen, repo)

  for repo in repositoriesFromEnv(cwd):
    appendRepository(repositories, seen, repo)

  for repo in cliRepos:
    appendRepository(repositories, seen, repo)

  return Bootstrap{Repositories: repositories}

NewCommand(bootstrap):
  scanned = ScanRepositories(bootstrap)
  verbs = CollectDiscoveredVerbs(scanned)
  commands = buildCommands(verbs)
  root = cobra.Command{Use: "verbs"}
  glazedcli.AddCommandsToRootCommand(root, commands)
  return root
```

### Config loading with local files

```text
loadConfigRepositories(ctx):
  plan = NewPlan(
    WithLayerOrder(system, user, repo, cwd),
    WithDedupePaths(),
  ).Add(
    /etc/css-visual-diff/config.yaml,
    $XDG_CONFIG_HOME/css-visual-diff/config.yaml,
    ~/.css-visual-diff/config.yaml,
    <git-root>/.css-visual-diff.yml,
    <git-root>/.css-visual-diff.override.yml,
    <cwd>/.css-visual-diff.yml,
    <cwd>/.css-visual-diff.override.yml,
  )

  files = plan.Resolve(ctx)

  ret = []Repository{}
  for file in files:
    repos = loadRepositoriesFromConfigFile(file.Path)
    ret += repos

  return ret
```

### Config-file-relative path normalization

```text
loadRepositoriesFromConfigFile(path):
  data = os.ReadFile(path)
  cfg = yaml.Unmarshal(data)
  baseDir = filepath.Dir(path)

  for spec in cfg.verbs.repositories:
    if spec.enabled is false:
      continue

    rootDir = normalizeFilesystemRepositoryPath(spec.path, baseDir)
    name = spec.name or filepath.Base(rootDir)
    append Repository{
      Name: name,
      Source: "config",
      SourceRef: path,
      RootDir: rootDir,
    }
```

### ASCII architecture diagram

```text
                ┌────────────────────────────┐
                │ css-visual-diff verbs ...  │
                └──────────────┬─────────────┘
                               │ raw args
                               ▼
                   ┌───────────────────────┐
                   │ verbcli.NewLazyCommand │
                   │ DisableFlagParsing     │
                   └──────────────┬────────┘
                                  │
                                  ▼
                       ┌──────────────────┐
                       │ DiscoverBootstrap │
                       └───────┬──────────┘
                               │
       ┌───────────────────────┼────────────────────────────┐
       │                       │                            │
       ▼                       ▼                            ▼
┌──────────────┐     ┌───────────────────┐        ┌────────────────┐
│ Glazed config │     │ Environment var    │        │ CLI bootstrap   │
│ plan          │     │ CSS_VISUAL_DIFF... │        │ --repository    │
└──────┬───────┘     └─────────┬─────────┘        └───────┬────────┘
       │                       │                          │
       ▼                       ▼                          ▼
┌───────────────────────────────────────────────────────────────────┐
│ Bootstrap.Repositories                                             │
│ builtin + config + env + cli, deduped by filesystem root           │
└──────────────────────────────┬────────────────────────────────────┘
                               ▼
                    ┌────────────────────┐
                    │ ScanRepositories    │
                    │ jsverbs.ScanFS/Dir  │
                    └──────────┬─────────┘
                               ▼
                    ┌────────────────────┐
                    │ CollectDiscovered   │
                    │ duplicate path check│
                    └──────────┬─────────┘
                               ▼
                    ┌────────────────────┐
                    │ Generated Cobra     │
                    │ verbs subtree       │
                    └────────────────────┘
```

## File-level implementation plan

### File: `internal/cssvisualdiff/verbcli/bootstrap.go`

Primary implementation file.

Changes:

1. Add local config filename constants.
2. Extend `loadConfigRepositories` plan layer order.
3. Add `GitRootFile` and `WorkingDirFile` source specs.
4. Keep existing config parsing and path normalization.

Suggested patch shape:

```diff
 const (
     VerbRepositoriesEnvVar = "CSS_VISUAL_DIFF_VERB_REPOSITORIES"
     RepositoryFlag         = "repository"
     VerbRepositoryFlag     = "verb-repository"
+
+    LocalConfigFileName         = ".css-visual-diff.yml"
+    LocalOverrideConfigFileName = ".css-visual-diff.override.yml"
 )
```

```diff
 plan := glazedconfig.NewPlan(
-    glazedconfig.WithLayerOrder(glazedconfig.LayerSystem, glazedconfig.LayerUser),
+    glazedconfig.WithLayerOrder(
+        glazedconfig.LayerSystem,
+        glazedconfig.LayerUser,
+        glazedconfig.LayerRepo,
+        glazedconfig.LayerCWD,
+    ),
     glazedconfig.WithDedupePaths(),
 ).Add(
-    glazedconfig.SystemAppConfig("css-visual-diff").Named("system-app-config"),
-    glazedconfig.XDGAppConfig("css-visual-diff").Named("xdg-app-config"),
-    glazedconfig.HomeAppConfig("css-visual-diff").Named("home-app-config"),
+    glazedconfig.SystemAppConfig("css-visual-diff").Named("system-app-config").Kind("app-config"),
+    glazedconfig.XDGAppConfig("css-visual-diff").Named("xdg-app-config").Kind("app-config"),
+    glazedconfig.HomeAppConfig("css-visual-diff").Named("home-app-config").Kind("app-config"),
+    glazedconfig.GitRootFile(LocalConfigFileName).Named("git-root-local-css-visual-diff").Kind("verb-repository-overlay"),
+    glazedconfig.GitRootFile(LocalOverrideConfigFileName).Named("git-root-local-css-visual-diff-override").Kind("verb-repository-overlay"),
+    glazedconfig.WorkingDirFile(LocalConfigFileName).Named("cwd-local-css-visual-diff").Kind("verb-repository-overlay"),
+    glazedconfig.WorkingDirFile(LocalOverrideConfigFileName).Named("cwd-local-css-visual-diff-override").Kind("verb-repository-overlay"),
 )
```

### File: `internal/cssvisualdiff/verbcli/bootstrap_test.go`

Add focused unit tests around discovery.

Current file only tests CLI prefix parsing. Add tests for local config discovery. Because `glazed/pkg/config` uses package-level function variables for `getwdFunc` and `gitRootFunc` internally, direct unit tests may be easiest through a temporary git repository and `os.Chdir`. Do not rely on the developer's real repo.

Recommended tests:

#### `TestLoadConfigRepositoriesDiscoversGitRootLocalConfig`

Setup:

1. Create temp dir.
2. Run `git init` inside it, or create a helper that shells out to git.
3. Create `verbs/hello.js` directory/file.
4. Write `.css-visual-diff.yml` with `path: ./verbs`.
5. `chdir` into a nested directory.
6. Call `loadConfigRepositories(context.Background())`.
7. Assert repository `RootDir == <temp>/verbs` and `SourceRef == <temp>/.css-visual-diff.yml`.

Pseudocode:

```go
func TestLoadConfigRepositoriesDiscoversGitRootLocalConfig(t *testing.T) {
    root := t.TempDir()
    runGitInit(t, root)
    repoDir := filepath.Join(root, "verbs")
    require.NoError(t, os.MkdirAll(repoDir, 0o755))
    writeFile(t, filepath.Join(root, LocalConfigFileName), `
verbs:
  repositories:
    - name: local
      path: ./verbs
`)
    nested := filepath.Join(root, "packages", "button")
    require.NoError(t, os.MkdirAll(nested, 0o755))
    withChdir(t, nested)

    repos, err := loadConfigRepositories(context.Background())
    require.NoError(t, err)
    require.Contains(t, repositoryRootDirs(repos), repoDir)
}
```

#### `TestLoadConfigRepositoriesDiscoversGitRootOverrideConfig`

Same setup, but write `.css-visual-diff.override.yml`.

#### `TestLoadConfigRepositoriesDiscoversWorkingDirConfig`

Setup:

1. Create temp root and nested dir.
2. Create nested `verbs` dir.
3. Write nested `.css-visual-diff.yml`.
4. `chdir` into nested dir.
5. Assert root dir is nested-relative.

#### `TestLoadConfigRepositoriesDiscoversWorkingDirOverrideConfig`

Same as above for `.css-visual-diff.override.yml`.

#### `TestLocalConfigRepositoryDedupeByRoot`

Setup:

1. Write both `.css-visual-diff.yml` and `.css-visual-diff.override.yml` pointing at the same `./verbs` dir.
2. Assert returned repos contain one entry for that root after `discoverBootstrap`, because `loadConfigRepositories` itself appends files but `discoverBootstrap` dedupes repositories.

Why use `discoverBootstrap` for this test?

- `appendRepository` dedupe happens in `discoverBootstrap`, not `loadConfigRepositories`.
- Testing the public-ish bootstrap behavior matches real runtime behavior.

### File: `internal/cssvisualdiff/verbcli/command_test.go`

Add one integration-style generated command test.

Goal: prove that local config is not merely loaded, but affects generated command lookup.

Pseudocode:

```go
func TestLazyCommandRunsVerbFromGitRootLocalConfig(t *testing.T) {
    root := t.TempDir()
    runGitInit(t, root)
    writeFile(t, filepath.Join(root, "verbs", "hello.js"), `
function hello(name) { return "hello " + name; }
__verb__("hello", {
  parents: ["local"],
  output: "text",
  fields: { name: { argument: true, required: true } }
});
`)
    writeFile(t, filepath.Join(root, LocalConfigFileName), `
verbs:
  repositories:
    - name: local
      path: ./verbs
`)

    nested := filepath.Join(root, "subdir")
    os.MkdirAll(nested, 0o755)
    withChdir(t, nested)

    cmd := NewLazyCommand()
    var out bytes.Buffer
    cmd.SetOut(&out)
    cmd.SetErr(&out)
    cmd.SetArgs([]string{"local", "hello", "Manuel"})

    require.NoError(t, cmd.Execute())
    require.Contains(t, out.String(), "hello Manuel")
}
```

This test protects the exact user story: no `--repository`, no environment variable, nested directory works.

### File: `internal/cssvisualdiff/doc/topics/javascript-verbs.md`

Update the `Repository sources` section.

Proposed text:

```markdown
The lazy `verbs` command discovers scripts from these sources:

1. embedded built-ins,
2. system/user app config,
3. git-root local config (`.css-visual-diff.yml`, `.css-visual-diff.override.yml`),
4. current-working-directory local config (`.css-visual-diff.yml`, `.css-visual-diff.override.yml`),
5. `CSS_VISUAL_DIFF_VERB_REPOSITORIES`,
6. CLI flags.
```

Add a local config subsection:

```markdown
### Project-local repository config

Add this file at the git root:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./verbs
```

Relative paths are resolved relative to the config file. A developer-local override can live in `.css-visual-diff.override.yml`; add that file to `.gitignore` if it should stay private.
```

### File: `README.md`

Update the short external repository section near the existing `--repository examples/verbs` example.

Proposed addition:

```markdown
For repeated project workflows, check in a local repository config:

```yaml
# .css-visual-diff.yml
verbs:
  repositories:
    - name: project
      path: ./verbs
```

Then run the generated verb without a bootstrap flag:

```bash
GOWORK=off go run ./cmd/css-visual-diff verbs project inspect-page ...
```

Use `.css-visual-diff.override.yml` for private local repositories and add it to `.gitignore`.
```

### Optional file: `.gitignore`

Consider adding:

```gitignore
.css-visual-diff.override.yml
```

Do not ignore `.css-visual-diff.yml`; it is intended to be shared.

## Test strategy

### Unit tests

Run targeted tests first:

```bash
GOWORK=off go test ./internal/cssvisualdiff/verbcli -run 'Test(LoadConfigRepositories|LazyCommand|RepositoriesFromArgs)' -count=1
```

Then run all verbcli tests:

```bash
GOWORK=off go test ./internal/cssvisualdiff/verbcli -count=1
```

Expected coverage:

- config plan discovers local files;
- `enabled: false` still works;
- relative paths use config file directory;
- nested current directory does not break git-root discovery;
- CLI bootstrap parsing still only consumes prefix flags.

### Full test suite

Run:

```bash
GOWORK=off go test ./... -count=1
```

This is the required pre-commit validation.

### Manual smoke test

From a temporary git repository:

```bash
tmp=$(mktemp -d)
cd "$tmp"
git init
mkdir -p verbs
cat > .css-visual-diff.yml <<'YAML'
verbs:
  repositories:
    - name: smoke
      path: ./verbs
YAML
cat > verbs/hello.js <<'JS'
function hello(name) { return `hello ${name}` }
__verb__("hello", {
  parents: ["smoke"],
  output: "text",
  fields: { name: { argument: true, required: true } }
})
JS
mkdir -p nested/deeper
cd nested/deeper
GOWORK=off go run /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/cmd/css-visual-diff verbs smoke hello Manuel
```

Expected output includes:

```text
hello Manuel
```

### Negative smoke test

Disable the repository:

```yaml
verbs:
  repositories:
    - name: smoke
      path: ./verbs
      enabled: false
```

Expected behavior:

```text
unknown command "smoke" for "verbs"
```

or equivalent Cobra unknown command output.

## Risks and tricky details

### Risk: Tests might read the developer's real config files

`loadConfigRepositories` discovers system and user config files. Unit tests that call it may unintentionally pick up real `~/.css-visual-diff/config.yaml` files.

Mitigations:

- Prefer tests that assert the presence of expected temp roots rather than exact repository slice length.
- If possible, add a helper that accepts a plan or resolver for deterministic tests. That is a larger refactor but can make tests cleaner.
- Use temp git repos and `t.Setenv` for environment variables.

### Risk: Git root discovery depends on the `git` executable

Glazed's `GitRootFile` uses `git rev-parse --show-toplevel`. Tests that rely on it need git available. In this development environment that is acceptable, but CI should already have git because the repository is checked out by git.

### Risk: CWD local file may duplicate git-root local file

When the current working directory is the git root, both `GitRootFile(".css-visual-diff.yml")` and `WorkingDirFile(".css-visual-diff.yml")` discover the same file. `WithDedupePaths` handles this at the config-file level, so the file should only be loaded once.

### Risk: Repository command-path collisions are not override semantics

The phrase "local override" can be misunderstood. In this design, `.css-visual-diff.override.yml` is an override file for local repository discovery, not a mechanism for replacing built-in verb commands. If two repositories declare the same command path, the existing duplicate-path error remains.

This should be documented clearly.

### Risk: Existing app config ordering may matter

The current source order inside the user layer is XDG before home. If implementation copies go-minitrace exactly, home may come before XDG. Avoid accidental changes unless there is a reason.

### Risk: `~` and environment expansion differ from go-minitrace

`css-visual-diff` currently expands `~/` in `normalizeFilesystemRepositoryPath` and stats the directory immediately. It does not explicitly call `os.ExpandEnv` for `$HOME/foo` config paths. go-minitrace preserves `$` through config normalization and expands environment variables later in `SourceRootsFromPaths`.

Do not change this casually. If environment variable expansion inside config paths is desired, design it separately and add tests.

## Alternatives considered

### Alternative A: Keep using only `--repository`

This is the status quo. It is simple but does not solve repeatability or nested-directory ergonomics.

Rejected because the requested workflow explicitly asks for local repository support like minitrace.

### Alternative B: Use only environment variables

`CSS_VISUAL_DIFF_VERB_REPOSITORIES` already exists. It is good for ad hoc local sessions and CI overrides, but it is invisible in the repository and easy to forget.

Rejected as the primary solution because project-local command catalogs should be declarative and documented by the repo itself.

### Alternative C: Add a new top-level config shape

Example:

```yaml
verbRepositories:
  - ./verbs
```

This would be shorter, but it duplicates an existing supported shape. `css-visual-diff` already documents `verbs.repositories`, so local files should use the same schema.

Rejected for now to avoid migration and user confusion.

### Alternative D: Adopt Pinocchio's full unified document model

Pinocchio supports `app`, `profile`, and `profiles` blocks with presence-aware merge semantics. That is powerful, but too broad for this feature.

Rejected for the first implementation because `css-visual-diff` only needs verb repository discovery.

### Alternative E: Allow local override files to replace repository lists

This would mirror go-minitrace's current replacement semantics for `queryRepositories`. However, css-visual-diff already appends repositories from all config files, and Pinocchio appends/dedupes `app.repositories`.

Rejected for now because append/dedupe better matches the existing css-visual-diff implementation and Pinocchio's repository-list behavior.

## Implementation checklist for a new intern

1. Read this guide fully.
2. Open `internal/cssvisualdiff/verbcli/bootstrap.go`.
3. Add local filename constants.
4. Extend `loadConfigRepositories` with repo and cwd Glazed config layers.
5. Keep `loadRepositoriesFromConfigFile` behavior unchanged.
6. Add tests in `bootstrap_test.go` for local file discovery.
7. Add at least one end-to-end lazy command test in `command_test.go`.
8. Update `internal/cssvisualdiff/doc/topics/javascript-verbs.md`.
9. Update `README.md`.
10. Optionally add `.css-visual-diff.override.yml` to `.gitignore`.
11. Run targeted tests.
12. Run `GOWORK=off go test ./... -count=1`.
13. Perform the manual temp-git-repo smoke test from a nested directory.
14. Commit only the code, tests, and docs relevant to this ticket.

## Review guide

Start review in this order:

1. `internal/cssvisualdiff/verbcli/bootstrap.go`
   - Confirm source order is intentional.
   - Confirm local filenames are correct.
   - Confirm relative paths still use config file directory.
2. `internal/cssvisualdiff/verbcli/bootstrap_test.go`
   - Confirm tests isolate temp repos and do not depend on Manuel's checkout.
   - Confirm nested-directory behavior is tested.
3. `internal/cssvisualdiff/verbcli/command_test.go`
   - Confirm generated command is available without `--repository`.
4. `internal/cssvisualdiff/doc/topics/javascript-verbs.md`
   - Confirm docs explain `.css-visual-diff.yml` and `.css-visual-diff.override.yml`.
5. `README.md`
   - Confirm the quickstart example is short and copy/pasteable.

## References

Primary css-visual-diff files:

- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go`
- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command.go`
- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap_test.go`
- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/command_test.go`
- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/doc/topics/javascript-verbs.md`
- `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/README.md`

Reference implementation files:

- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/minitracecmd/repositories.go`
- `/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/query/commands.go`
- `/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/cmds/profilebootstrap/profile_selection.go`
- `/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/configdoc/types.go`
- `/home/manuel/code/wesen/corporate-headquarters/pinocchio/pkg/configdoc/merge.go`
- `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/config/plan.go`
- `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/config/plan_sources.go`

Project note:

- `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/27/PROJ - go-minitrace - Local Query Repository Config.md`
