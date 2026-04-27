---
title: Diary
doc_type: reference
intent: long-term
status: active
topics:
  - css-visual-diff
  - react
  - review
  - visual-regression
  - glazed
created: "2026-04-27"
...

# Diary

## Goal

Build an interactive React review site for css-visual-diff comparison results. The site takes the output of css-visual-diff (screenshots, compare.json, summary JSON) and presents them as an interactive comparison page with local storage for reviewer feedback. The app will be a static SPA bundled into the css-visual-diff Go binary via go:embed and served through a glazed `serve` verb.

---

## Step 1: Ticket creation and source material gathering

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to build a comparison serving website into css-visual-diff, as a react rtk/redux vite app (see skill) as a static webpage that can be used with a simple static `serve` verb or so. It takes the output of css-visual-diff with screenshots and CSS diff and then presents them as an interactive comparison page with local storage to give feedback. The page has already been built as a mock in ~/Downloads/diff-review.jsx . Import it into the ticket with docmgr import file --file XXX --ticket YYY and read it."

**Assistant interpretation:** Create a docmgr ticket, import the mock JSX and reference docs, understand the full data model from css-visual-diff output, then plan and implement a React+RTK+Vite review site.

**Inferred user intent:** Replace the current static HTML generator (Python script producing a single index.html) with a proper React SPA that can be served from the Go binary and provides interactive review with persistent feedback.

### What I did

- Created docmgr ticket `CSSVD-REVIEW-SITE` with topics: css-visual-diff, react, review, visual-regression, glazed.
- Imported `~/Downloads/diff-review.jsx` (the React mock) into ticket sources.
- Imported `05-public-pages-review-site-analysis-and-guide.md` (the Pyxis reference doc with full architecture analysis) into ticket sources.
- Created this diary document.
- Created design doc `01-design-react-review-site-architecture-and-implementation-guide.md`.
- Added 9 tasks covering diary, design doc, scaffolding, component porting, Go embed, data loading, local storage, and reMarkable upload.
- Read and analyzed:
  - The diff-review.jsx mock (~700 lines of React with inline mock data, comment pins, overlay/slider/diff view modes, sidebar with CSS diff tab, export modal).
  - The reference guide (~1500 lines covering the current static HTML generator, data model, proposed API, component decomposition).
  - The css-visual-diff Go CLI structure (main.go with run/inspect/compare/llm-review/verbs commands).
  - The verbcli system (JS verb repositories scanned and loaded via go-go-goja).
  - A real compare.json artifact (full schema with bounds, pixel, styles, attributes, text, artifacts).
  - A real summary JSON (suite-level with rows, classificationCounts, policy).

### Why

Before writing a single line of code, I needed to understand the complete data flow: what css-visual-diff produces (compare.json per section, summary.json per suite), how the current static generator works, what the mock JSX envisions, and how the Go binary serves things. This avoids building the wrong thing.

### What I learned

- **The mock JSX uses inline mock data**, not real screenshots. The actual app needs to load compare.json files and reference real PNG artifacts (diff_only.png, left_region.png, right_region.png, diff_comparison.png).
- **The summary JSON is the entry point.** Each row has page, section, classification, changedPercent, and paths to all artifacts. The review site should load this first, then lazily load compare.json per card.
- **The Go binary already has a `verbs` subsystem** for JS-based workflow commands. The `serve` command could be either a native Go subcommand or a verb. Since it just needs to serve static files from an embedded FS, a native Go command is simpler.
- **The current static HTML generator** copies artifacts into a self-contained bundle. The React version will do the same but at build time: Vite builds the SPA, Go embeds it, and at runtime the `serve` command copies or symlinks artifacts next to the embedded SPA.
- **The mock has rich interaction patterns** that the static HTML never had: overlay mode with opacity slider, slider mode with clip-path, comment pins with types (issue/note/question/praise), sidebar tabs for CSS diff and metadata, and a markdown export modal. These are all worth preserving.

### What should be done in the future

- Consider adding WebSocket support for live re-loading when new comparison results arrive.
- Consider a comparison-between-runs feature (showing delta of deltas).
- Consider server-side persistence (SQLite) as an alternative to local storage only.

---

## Step 2: Writing the design and implementation guide

### What I did

- Writing the detailed design/implementation guide document that explains every part of the system to a new intern.
- Covers: project overview, data model, component architecture, Go integration, build pipeline, local storage schema, and step-by-step implementation instructions.

### What was tricky to build

- Balancing depth vs. brevity: the document needs to be detailed enough for an intern but not so long it becomes unreadable. Using diagrams, pseudocode, and structured sections helps.
- Mapping from the mock JSX's inline mock data to the real data loading strategy (summary.json → per-card compare.json lazy loading).
- Deciding on the Go embed strategy: embed the built SPA assets into the binary, serve them from a `serve` subcommand that also proxies artifact files from a user-specified directory.
