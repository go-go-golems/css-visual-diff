---
Title: Interactive Visual Review Site
Slug: review-site
Short: Serve a self-contained React review site from css-visual-diff output for interactive visual comparison with local storage feedback.
Topics:
- review
- serve
- visual-regression
- react
- embed
Commands:
- serve
Flags:
- data-dir
- summary
- port
- host
- open
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

The visual review site is an interactive single-page application that turns the artifacts produced by css-visual-diff — screenshots, CSS diffs, and comparison metadata — into a human-friendly review interface. Instead of browsing a folder of PNGs and JSON files, the reviewer opens a single URL and sees every comparison as an interactive card they can annotate, score, and export for further work.

The site is built as a React application compiled by Vite, embedded into the Go binary at compile time, and served through the `css-visual-diff serve` command. No Node.js runtime is needed at serve time. The binary is fully self-contained.

## Why This Exists

Visual regression tools are good at measurement: they compute pixel differences, extract CSS property deltas, and classify sections by severity. But measurement is only half the job. The other half is human judgment: someone needs to look at the screenshots, decide whether each difference is acceptable, write notes about what to fix or accept, and pass those notes to a developer or coding agent.

The review site bridges measurement and judgment. It keeps the computed evidence — pixel percentages, CSS diffs, selector metadata — close to the human decision. Each comparison section gets its own card with images, metadata, a status dropdown, and a notes field. Everything the reviewer types is persisted in browser localStorage so it survives page reloads. When the review is done, the reviewer opens the export modal and copies a markdown + YAML block that contains all feedback in a format ready for an issue, a pull request comment, or an LLM prompt.

## Quick Start

Build the frontend and embed it into the binary, then serve a comparison run:

```bash
# Build the React SPA and copy into the Go embed directory
BUILD_WEB_LOCAL=1 go run ./cmd/build-web

# Compile the binary; the SPA is always embedded
go build -o dist/css-visual-diff ./cmd/css-visual-diff

# Serve a completed comparison run
css-visual-diff serve \
  --data-dir /tmp/my-comparison-run \
  --port 8097
```

Open `http://127.0.0.1:8097` in a browser. The site loads the summary manifest from the data directory and renders one review card per page/section comparison.

The data directory should be the output of a previous `css-visual-diff` run or a verb-based comparison suite. It must contain subdirectories like `<page>/artifacts/<section>/` with `compare.json`, `left_region.png`, `right_region.png`, and `diff_only.png` files. A `summary.json` at the root of the data directory provides the manifest that lists every row the reviewer should see.

## The Serve Command

The `serve` subcommand starts an HTTP server that provides three things: API endpoints for the manifest and per-section comparison data, artifact file serving for PNG screenshots and JSON files, and the embedded React SPA itself.

### Flags

| Flag | Default | Purpose |
| --- | --- | --- |
| `--data-dir` | (required) | Path to the css-visual-diff run output directory. |
| `--summary` | `<data-dir>/summary.json` | Explicit path to the summary JSON manifest. |
| `--port` | `8097` | HTTP server port. |
| `--host` | `127.0.0.1` | Bind address. |
| `--open` | `false` | Open the browser automatically after starting. |

### API Endpoints

The React SPA communicates with the Go server through three endpoints. You do not normally call these directly, but they are stable and documented here for scripting or integration.

**GET /api/manifest** returns the summary JSON. This is the entry point the SPA fetches on load. It contains an array of rows, each with `page`, `section`, `classification`, `changedPercent`, and paths to every artifact file. The SPA rewrites these absolute paths into relative `/artifacts/` URLs.

**GET /api/compare?page=NAME&section=NAME** returns the `compare.json` for a single page/section. The SPA loads this lazily when a reviewer expands a card. It contains the full comparison data: bounds, pixel counts, computed style differences, attribute changes, and source URLs for both prototype and React sides.

**GET /artifacts/{page}/{section}/{file}** serves a single artifact file from the data directory. The Go handler maps the three-part URL path back to the on-disk location `<data-dir>/<page>/artifacts/<section>/<file>`. This is how the SPA loads PNG screenshots and raw compare JSON.

### SPA Fallback

Any request that does not match `/api/` or `/artifacts/` falls through to the embedded React SPA. Unknown paths serve `index.html` so that client-side routing works. The SPA assets are always compiled into the Go binary from `internal/cssvisualdiff/review/embed/public/`.

## The Review Interface

When the page loads, the reviewer sees a header bar and a list of comparison cards. The header shows the total page and section counts, the worst classification, and classification tallies. Below the header, a toolbar provides view mode buttons and an "Add comment" toggle.

### Review Cards

Each card represents one page/section comparison. The collapsed card shows the page and section name, the computed classification (a colored pill), and the changed percentage. A dropdown lets the reviewer set a human status: unreviewed, accepted, needs-work, fixed, or wont-fix.

Classification is computed by css-visual-diff from the pixel-change percentage. Status is the reviewer's own verdict. They are independent. A card can be classified as "tune-required" but accepted by the reviewer if the visual difference is intentional or acceptable.

Clicking a card expands it to show the image comparison area, a general observation textarea, and artifact links.

### View Modes

The image comparison area supports four view modes, selected from the toolbar or by pressing keys 1 through 4.

**Side-by-side** (key `1`) shows the prototype screenshot on the left and the React screenshot on the right. Each image has a label strip showing "prototype" or "react" and the source URL.

**Overlay** (key `2`) stacks both images. The reviewer drags an opacity slider between A (prototype) and B (React). A "diff blend" toggle switches to CSS difference blend mode. Hold the `F` key to flash between the two images instantly. This is the fastest way to spot subtle alignment or color differences.

**Slider** (key `3`) uses CSS clip-path to show the prototype on the left of a draggable divider and React on the right. The reviewer drags the handle to sweep across the image.

**Diff only** (key `4`) shows only the `diff_only.png` artifact, which highlights the changed pixels. This is the best starting point for identifying where differences exist before examining the full screenshots.

### Zoom and Pan

All view modes support zoom and pan for inspecting pixel-level details. Scroll the mouse wheel to zoom in and out (0.25x to 8x). The zoom tracks toward the cursor position. Hold Shift and drag with the left mouse button to pan around the zoomed image. Double-click to reset zoom and pan to the default view. A small indicator in the bottom-left corner shows the current zoom percentage and pixel offset.

### Comment Pins

The reviewer can drop numbered annotation pins on any image. Click "Add comment" in the toolbar (or press `P`), then click on the image. Each pin has a type — issue, note, question, or praise — and a text field. Pins are visible as colored numbered circles on the image and listed in the sidebar Comments tab. The reviewer can change the type, edit the text, or delete a pin from the sidebar.

Pins are persisted in localStorage alongside the review status and notes.

### Sidebar

The right sidebar has three tabs. **Comments** lists all pins for the current card with inline editing. **CSS diff** shows every computed CSS property that differs between prototype and React, with the left and right values side by side. **Meta** shows bounds, pixel counts, selectors, and source URLs.

The CSS diff tab is especially useful for answering "why does this look different?" — it shows properties like `font-size`, `padding`, `background-color`, and `line-height` with exact values from both sides.

### Status and Notes

Each card has a status dropdown and a general observation textarea. Both are automatically persisted to localStorage. The status options are:

- **Unreviewed** — no human decision yet (default).
- **Accepted** — the difference is fine.
- **Needs work** — something should be fixed.
- **Fixed** — a change has been made after feedback.
- **Wont-fix** — acknowledged difference that will remain as-is.

The note field is free-form text for observations that apply to the whole section.

### Keyboard Shortcuts

The review site supports keyboard shortcuts for fast navigation without touching the mouse. All shortcuts are disabled when the cursor is inside a textarea, input, or dropdown.

| Key | Action |
| --- | --- |
| `j` | Move to the next card. |
| `k` | Move to the previous card. |
| `a` | Mark current card as accepted. |
| `n` | Mark current card as needs work. |
| `w` | Mark current card as won't fix. |
| `x` | Mark current card as fixed. |
| `1` | Switch to side-by-side view. |
| `2` | Switch to overlay view. |
| `3` | Switch to slider view. |
| `4` | Switch to diff-only view. |
| `e` | Open the export modal. |
| `p` | Enter comment pin mode. |

## The Export Modal

The "Send to LLM" button in the header (or pressing `E`) opens the export modal. This modal generates a markdown + YAML document containing all review feedback. The reviewer can choose to export all cards or only reviewed ones. The preview shows the full generated text.

The exported markdown includes, for each card:

- The page/section name, classification, and changed percentage.
- The reviewer's status decision.
- Any general observation notes.
- Computed style differences (property, left value, right value).
- A YAML code block with structured metadata: bounds, pixel counts, classification, and all review comments.

Clicking "Copy markdown" copies the full text to the clipboard. The reviewer can then paste it into a GitHub issue, a pull request comment, a chat message, or an LLM prompt.

## Local Storage Persistence

All reviewer feedback — status decisions, notes, and comment pins — is stored in the browser's localStorage under a key derived from the manifest content. This means:

- Feedback survives page reloads and browser restarts.
- Each comparison run gets its own storage namespace.
- Clearing browser data removes all stored feedback.
- Feedback is local to the browser and device. It is not synced to a server or shared with other users.

The storage key looks like `cssvd-review-run-<hash>`, where `<hash>` is derived from the page/section names in the manifest. This ensures that a different comparison run does not overwrite feedback from a previous run.

## Build Pipeline

The review site uses a two-stage build: first the React SPA is built by Vite, then the output is copied into a Go embed directory and compiled into the binary.

### Dagger Build (Recommended)

The `cmd/build-web` tool uses Dagger to build the frontend inside a `node:22` container. It creates a pnpm CacheVolume so that repeated builds reuse downloaded packages. This means Docker is required, but Node.js does not need to be installed on the host.

```bash
go run ./cmd/build-web
```

### Local Build (Fallback)

If Docker is unavailable, set `BUILD_WEB_LOCAL=1` to build using the locally installed pnpm:

```bash
BUILD_WEB_LOCAL=1 go run ./cmd/build-web
```

### Generating and Compiling

The Go embed directory is populated by the build tool. To regenerate it and compile:

```bash
# Build frontend and copy to embed directory
BUILD_WEB_LOCAL=1 go run ./cmd/build-web

# Compile with embedded SPA
go build -o dist/css-visual-diff ./cmd/css-visual-diff
```

The SPA is always embedded from `internal/cssvisualdiff/review/embed/public/`; there is no non-embedded filesystem fallback. If you change frontend code, rebuild the web assets and then rebuild the Go binary.

### Development Workflow

During frontend development, run the Vite dev server and the Go server separately:

```bash
# Terminal 1: Go server
go run ./cmd/css-visual-diff serve --data-dir /tmp/my-run --port 8097

# Terminal 2: Vite dev server with HMR
cd web/review-site && pnpm dev
```

The Vite dev server runs on port 5173 and proxies `/api` and `/artifacts` requests to the Go server on port 8097. Edit React components and see changes instantly with hot module replacement.

## Makefile Targets

The project Makefile provides these targets for common operations:

| Target | What it does |
| --- | --- |
| `build-web` | Build the React SPA using local pnpm and copy to embed directory. |
| `build-embed` | Build the frontend and then compile the Go binary with it embedded. |
| `dev-web` | Start the Vite dev server with hot reload. |
| `dev-serve` | Start the Go serve command with test data on port 8098. |

## Data Directory Layout

The serve command expects the data directory to follow the structure produced by css-visual-diff:

```text
/tmp/my-comparison-run/
  summary.json                         ← manifest listing all rows
  about/
    artifacts/
      content/
        compare.json
        left_region.png                ← prototype screenshot
        right_region.png               ← react screenshot
        diff_only.png                  ← changed-pixel highlight
        diff_comparison.png            ← side-by-side triptych
  shows/
    artifacts/
      content/
        ...
      header/
        ...
```

If the summary JSON is at a different location, use the `--summary` flag:

```bash
css-visual-diff serve \
  --data-dir /tmp/my-comparison-run \
  --summary /tmp/my-comparison-run.json
```

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| "No embedded SPA found" message | The embedded asset directory does not contain `index.html` when the binary was built. | Run `BUILD_WEB_LOCAL=1 go run ./cmd/build-web` then rebuild with `go build`. |
| Cards load but images show 404 | Artifact paths in the summary JSON do not match the data directory structure. | Ensure `--data-dir` points to the directory containing `<page>/artifacts/<section>/` subdirectories. |
| Images not loading | The Go server is not running, or the port is wrong. | Verify the server is running and the browser is accessing the correct port. |
| Export modal is empty | The summary JSON has no rows, or all rows have empty data. | Run css-visual-diff with `--summary` to produce a valid manifest. |
| Keyboard shortcuts do not work | Focus is inside a textarea, input, or dropdown. | Click outside the input field and try again. |
| Zoom is stuck at one level | The scroll event is being captured by a parent scrollable element. | Place the mouse directly over the image area and scroll. |
| localStorage notes disappeared | Browser data was cleared, or the manifest changed (different hash). | Re-open the same comparison run. Previous feedback is lost only if the browser data was cleared. |
| Dagger build fails | Docker is not running or not installed. | Start Docker, or use `BUILD_WEB_LOCAL=1` as a fallback. |

## See Also

- `css-visual-diff help inspect-workflow` — How to validate a config before running comparisons.
- `css-visual-diff help config-selectors` — How selectors map to screenshot regions.
- `css-visual-diff help javascript-verbs` — How verb scripts drive comparison suites that produce the data consumed by this review site.
