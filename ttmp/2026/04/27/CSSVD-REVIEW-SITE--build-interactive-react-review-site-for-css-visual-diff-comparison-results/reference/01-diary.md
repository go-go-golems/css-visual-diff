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

---

## Step 3: Full implementation — React app, Go embed, Dagger pipeline, serve command

### Prompt Context

**User prompt (verbatim):** "build all of it. Use the dagger build pipeline (see skill)."

**Assistant interpretation:** Implement the entire review site: scaffold React app, create all components, wire RTK store, build Go internal/web package, create Dagger build pipeline, create serve command, test end-to-end.

**Inferred user intent:** Get a working end-to-end system that can be served from a single Go binary.

### What I did

**React app (web/review-site/):**
- Scaffolded Vite+React+TypeScript+Tailwind+RTK in web/review-site/
- Created TypeScript types: summary.ts, compare.ts, review.ts
- Created RTK store with 4 slices: cardsSlice, viewSlice, reviewSlice, commentsSlice
- Created localStorage sync middleware that persists review state on every action
- Built utility functions: paths.ts (artifact URL rewriting), export.ts (markdown+yaml), storage.ts
- Built 12 components: App, Header, CardList, ReviewCard, ViewModeSideBySide, ViewModeOverlay, ViewModeSlider, ViewModeDiff, CommentPin, Sidebar, CommentsTab, StylesTab, MetaTab
- Configured Vite with dev proxy to Go server

**Go backend:**
- Created internal/cssvisualdiff/review/ package with embed.go, embed_none.go, static.go, generate.go
- Created cmd/build-web/ Dagger pipeline with local pnpm fallback and CacheVolume
- Created cmd/css-visual-diff/serve.go with serve subcommand
- Added 3 API endpoints: GET /api/manifest, GET /api/compare, GET /artifacts/{path...}
- Added SPA fallback serving for embedded React app
- Added Makefile targets: build-web, build-embed, dev-web, dev-serve

### Why

The Dagger pipeline ensures reproducible builds. The embed approach means a single binary distribution. The serve command is the only entry point a user needs.

### What worked

- `pnpm build` produces a ~270KB JS bundle + 21KB CSS — compact.
- `BUILD_WEB_LOCAL=1 go run ./cmd/build-web` builds the frontend and copies to embed/public.
- `go build -tags embed` produces a 73MB binary that contains the full SPA.
- All three API endpoints return correct data when tested against real Pyxis comparison data.
- Artifact serving correctly maps `/artifacts/shows/content/diff_only.png` → `data-dir/shows/artifacts/content/diff_only.png`.

### What didn't work

- Port 8097 was already in use from a previous test. Used 8098 instead.
- Initial artifact handler used naive path joining that didn't insert the `artifacts/` subdirectory. Fixed by splitting the URL path and inserting `artifacts` between page and section.

### What was tricky to build

- The artifact path mapping: the React app expects `/artifacts/page/section/file` but the on-disk structure is `page/artifacts/section/file`. The serve handler needs to insert `artifacts` between parts[0] and parts[1].
- TypeScript strict mode caught several unused imports and implicit `any` types that needed cleanup.
- The `localStorageSync` middleware needs to combine data from both `review` and `comments` slices since comments are stored as part of the card review but managed in a separate slice.

### What warrants a second pair of eyes

- The artifact path mapping in serve.go — it assumes the pattern `page/section/file` always holds. Could break if css-visual-diff changes its directory structure.
- The localStorage sync middleware accesses state from two different slices — this is technically fragile if the slice names change.

### What should be done in the future

- Add the ExportModal component (currently just shows an alert).
- Add keyboard shortcuts for fast review (j/k navigation, a/n/f status shortcuts).
- Add a `--summary` flag that accepts the summary JSON path directly instead of requiring it in the data dir.
- Add zoom/pan for images.
- Add synchronized scroll between prototype and React images.
- Add run comparison (diff between two runs).
- Test with Dagger (not just local pnpm).
- Add integration tests.

### Code review instructions

- Start with `cmd/css-visual-diff/serve.go` — the serve command and HTTP routing.
- Then `internal/cssvisualdiff/review/` — the embed/static/generate package.
- Then `cmd/build-web/main.go` — the Dagger pipeline.
- Then `web/review-site/src/store/slices/cardsSlice.ts` — the main data flow.
- Then `web/review-site/src/components/ReviewCard.tsx` — the core UI component.
- Validate: `make build-embed && ./dist/css-visual-diff serve --data-dir /tmp/pyxis-public-pages-final-sweep --port 8098`

### Technical details

**Build commands:**
```bash
# Build frontend locally
BUILD_WEB_LOCAL=1 GOWORK=off go run ./cmd/build-web

# Build Go binary with embedded SPA
GOWORK=off go build -tags embed -o dist/css-visual-diff ./cmd/css-visual-diff

# Run
./dist/css-visual-diff serve --data-dir /tmp/pyxis-public-pages-final-sweep --port 8098
```

**Commit:** Two commits:
1. feat: add interactive React review site with Go embed and Dagger build
2. fix: address lint errors in serve, build-web, and review packages
