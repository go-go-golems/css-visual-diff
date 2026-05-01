---
Title: Investigation Diary
Ticket: overlay-screenshot-labels
Status: active
Topics:
    - frontend
    - capture
    - chromedp
    - cdp
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../2026-04-25--overlay-select-components/extension/content_scripts/modules/dom-overlay.js
      Note: Reference DOM overlay implementation from Chrome extension
ExternalSources: []
Summary: Chronological investigation diary for the overlay screenshot labels feature.
LastUpdated: 2026-04-28T08:30:00-04:00
WhatFor: Record investigation steps, what worked, what failed, and lessons learned.
WhenToUse: When resuming work on this ticket or reviewing past decisions.
---


# Investigation Diary

## Goal

This diary captures the investigation, design, and planning work for adding overlay labeling of components in screenshots to `css-visual-diff`. The feature will use the Chrome DevTools Protocol Overlay API to annotate captured webpages with bounding boxes, labels, and legends for known components or selectors.

---

## Step 1: Ticket Creation and Initial Exploration

Created the `overlay-screenshot-labels` ticket to track this feature. The user wants to export screenshots of captured webpages where known components or selectors are visually annotated — either with bounding boxes, markers, or a legend. This is useful for documentation, design reviews, and automated visual regression reports where stakeholders need to understand what each UI region represents.

The user pointed to an existing Chrome extension at `~/code/wesen/2026-04-25--overlay-select-components` as a reference for overlay ideas. I explored that extension's DOM overlay implementation, which uses injected CSS and absolutely positioned divs to draw hover boxes, selection boxes, and labels on top of the page. While that approach works for a content-script extension, `css-visual-diff` already uses `chromedp` (a Go CDP client), so we have access to the native Chrome DevTools Protocol Overlay domain — a much cleaner and more reliable approach.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to add overlay labeling of components in screenshots. I want to basically add functionality to export screenshots of a captured webpage for a list of known components or selectors, such that I can annotate what each component is and how, with potentially bounding box or maybe with a marker and a legend. I think we can use the overlay API of chrome, see ~/code/wesen/2026-04-25--overlay-select-components maybe for some ideas (it's a chrome extension). Keep a diary as you work. Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and detailed. Store in the ticket and then upload to remarkable."

**Assistant interpretation:** Create a docmgr ticket and write an exhaustive design/implementation guide for adding overlay annotation capabilities to `css-visual-diff`, using Chrome's CDP Overlay API rather than DOM injection, suitable for onboarding a new intern.

**Inferred user intent:** The user wants a production-quality annotation feature that can draw labeled bounding boxes on screenshots for documentation and comparison workflows, exposed both through Go APIs and the existing Goja JavaScript runtime so that script authors can use it programmatically.

### What I did
- Read the `ticket-research-docmgr-remarkable`, `diary`, and `docmgr` skills.
- Explored `css-visual-diff` project structure: `cmd/css-visual-diff/main.go`, `internal/cssvisualdiff/driver/chrome.go`, `internal/cssvisualdiff/modes/capture.go`, `internal/cssvisualdiff/jsapi/module.go`, `internal/cssvisualdiff/service/`.
- Explored the reference Chrome extension's overlay implementation (`dom-overlay.js`, `overlay.js`, `overlay.css`).
- Created the docmgr ticket `overlay-screenshot-labels` with docs.

### What worked
- The project has a clean layered architecture: `driver` (chromedp wrapper) → `service` (business logic) → `jsapi` (Goja bindings) → `modes` (CLI workflows).
- The extension's overlay code gives a clear mental model of what visual annotations look like: colored bounding boxes, name labels positioned above elements, and a persistent legend.
- `chromedp` supports CDP domain actions directly, so we can invoke `overlay.highlightNode`, `overlay.setShowGrid`, or use the `Overlay` domain programmatically.

### What didn't work
- The extension uses DOM-injected overlays, which is not the right approach for `css-visual-diff` because:
  - Injected DOM elements can interfere with page layout and screenshots.
  - The overlay divs themselves might be captured in screenshots in unpredictable ways.
  - CDP's native Overlay API draws on a compositor layer that does not affect DOM or layout.

### What I learned
- Chrome's CDP `Overlay` domain provides `highlightNode`, `hideHighlight`, `setShowGrid`, `setShowFlexOverlay`, `setShowScrollSnapOverlays`, and more.
- For custom labels and legends, we likely need a hybrid approach: CDP `Overlay.highlightNode` for the bounding box, plus a small injected script to render a legend panel, OR we can use CDP `Page.captureScreenshot` after using `Overlay` to draw highlights, then composite labels ourselves in Go using an image library.
- The `Overlay` domain can highlight nodes by `backendNodeId` or `nodeId`, which we can obtain from the `DOM` domain via `DOM.querySelector`.

### What was tricky to build
- Determining the right CDP API surface. `Overlay.highlightNode` shows a transient highlight with configurable `HighlightConfig` (showInfo, showExtensionLines, displayAsMaterial, contrastAlgorithm, contentColor, paddingColor, borderColor, marginColor, shapeColor, shapeMarginColor, gridHighlightConfig, flexContainerHighlightConfig, flexItemHighlightConfig, containerQueryContainerHighlightConfig, isSourceOrder). However, it does not natively support custom text labels.
- For persistent labeled annotations, the most reliable path is:
  1. Use `DOM.querySelector` to get `nodeId` for each component.
  2. Use `Overlay.highlightNode` with distinct colors per component to draw bounding boxes.
  3. Either inject a small legend DOM element and screenshot it, OR post-process the screenshot in Go to draw text labels.
- The user mentioned "marker and a legend" — a legend panel is likely desired. Post-processing in Go gives us full control over label placement without worrying about page CSS conflicts.

### What warrants a second pair of eyes
- Whether to use CDP `Overlay` alone or combine with Go image processing for labels.
- How the Goja API should expose overlay configuration (builder pattern vs. plain objects).

### What should be done in the future
- Prototype the CDP `Overlay.highlightNode` approach to verify color customization and screenshot capture behavior.
- Evaluate Go image libraries (`golang.org/x/image/draw`, `github.com/fogleman/gg`, `github.com/disintegration/imaging`) for compositing labels onto screenshots.

### Code review instructions
- Review the design doc for API consistency with existing `cvd.probe()`, `cvd.page.inspect()`, and `page.locator()` patterns.
- Verify that proposed file locations align with the existing `driver/` → `service/` → `jsapi/` layering.

### Technical details
- Key files explored:
  - `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/driver/chrome.go` — `Page.Screenshot`, `Page.Evaluate`, `chromedp.Run`.
  - `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/dom.go` — `LocatorBounds`, `LocatorStatus`.
  - `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/module.go` — `wrapPage`, proxy builder pattern.
  - `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/modes/capture.go` — screenshot orchestration for sections.
- Reference extension:
  - `/home/manuel/code/wesen/2026-04-25--overlay-select-components/extension/content_scripts/modules/dom-overlay.js` — DOM overlay approach.

---

## Step 2: Design Document Writing

Wrote the comprehensive design and implementation guide in `design-doc/01-overlay-screenshot-labels-design-and-implementation-guide.md`. The document covers:

- Executive summary and problem statement
- Current architecture with file references and line-anchored evidence
- Gap analysis against the desired feature
- Proposed solution using CDP Overlay API + Go image compositing
- Three annotation strategies compared (CDP-only, DOM-injected legend, Go post-processing)
- Recommended hybrid: CDP for bounding boxes, Go for labels and legend
- Detailed API design for Go service layer, driver extensions, Goja bindings
- Config schema extensions (`OverlaySpec`)
- Phased implementation plan with file-level guidance
- Testing strategy
- Risks and open questions

The document is written for a new intern and includes prose paragraphs, bullet points, pseudocode, ASCII diagrams, API references, and concrete file references.

### What I did
- Drafted the full design document.
- Related key source files to the ticket.
- Updated changelog.

### What worked
- The existing `service.InspectRequest`, `config.SectionSpec`, and `jsapi` builder patterns provide a natural place to hook in overlay configuration.

### What didn't work
- N/A

### What I learned
- The `Overlay` domain in CDP has a `highlightNode` command that accepts a `HighlightConfig` with RGBA color objects. We can cycle through a palette of colors for different components.
- `Page.captureScreenshot` captures the current viewport including any active CDP overlay highlights.
- For a legend, Go image processing is more reliable than DOM injection because it avoids z-index conflicts, page CSS resets, and responsive layout issues.

### What was tricky to build
- Balancing thoroughness with readability for an intern audience. The design doc includes both high-level concepts and concrete implementation details.

### What warrants a second pair of eyes
- The proposed `OverlayLabelConfig` schema and whether it should live in `config.Config` or be passed ad-hoc in scripts.

### What should be done in the future
- Run `docmgr doctor` and upload to reMarkable.

### Code review instructions
- Read the design doc in `design-doc/01-overlay-screenshot-labels-design-and-implementation-guide.md`.
- Check that all referenced files exist and line numbers are accurate.

---

## Step 3: Validation and Upload

Ran `docmgr doctor` to validate the ticket, added missing vocabulary entries, and uploaded the document bundle to reMarkable.

### What I did
- `docmgr doctor --ticket overlay-screenshot-labels --stale-after 30`
- Added vocabulary entries for `cdp`, `overlay`, `screenshot` topics.
- Ran `remarquee upload bundle` with dry-run first, then actual upload.
- Verified upload with `remarquee cloud ls`.

### What worked
- Dry-run caught no issues.
- Upload succeeded and the bundle appeared in `/ai/2026/04/28/overlay-screenshot-labels`.

### What didn't work
- N/A

### What I learned
- The `remarquee` CLI works reliably for bundles with `--toc-depth 2`.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Begin Phase 1 implementation when the ticket is prioritized.

### Code review instructions
- N/A

---

## Step 4: Baseline JS Overlay API Implementation and Crop Planning

I checkpointed the implemented JS-first overlay API before starting crop work so the crop changes can be reviewed separately. The baseline implementation provides typed Goja builders, typed Go-side overlay styling, real browser CSS injection through `page.css(...)`, and Go-rendered full-page annotated screenshots.

I then added a detailed crop-support implementation guide to the design document. The plan keeps crop intentionally small for V1: `.cropTo(selector)` plus `.cropPadding(...)`, with target filtering, coordinate translation, and focused organism/component outputs.

### Prompt Context

**User prompt (verbatim):** "ok, add a detailed guide / implementation plan to add crop support to the ticket and upload to remarkable, then add tasks to the ticket, and implement the tasks one by one, committing at appropriate intervals, and keeping a diary."

**Assistant interpretation:** Add a crop-support plan to the overlay ticket, publish it to reMarkable, add concrete tasks, then implement those tasks incrementally with commits and diary updates.

**Inferred user intent:** The user wants crop support handled as a documented, reviewable feature rather than an ad-hoc code change, with intermediate commits and enough diary context to resume or audit the work later.

**Commit (code):** 88c45cb51955185316485c32482d6c734fc2c558 — "Implement JS overlay screenshot API"

### What I did
- Committed the baseline overlay API implementation before crop work.
- Fixed pre-commit lint failures before the baseline commit:
  - checked `f.Close()` in `OverlayScreenshot`,
  - added the missing `LegendBottomRight` switch case,
  - renamed a local `close` variable in RGB parsing to avoid shadowing the predeclared identifier.
- Added a new "Crop Support: Detailed Implementation Guide" section to the design document.
- Rewrote `tasks.md` with explicit crop tasks.

### Why
- Crop changes build directly on the baseline overlay pipeline, so committing the baseline first gives reviewers a clean boundary.
- The crop feature touches geometry, image clipping, JS builder decoding, and tests; documenting the intended pipeline reduces implementation ambiguity.

### What worked
- The baseline commit passed the repository pre-commit hook after lint fixes.
- The crop plan maps cleanly onto the existing Go-rendered overlay architecture: crop can be implemented by cropping the full screenshot, translating document-coordinate bounds, and drawing only intersecting targets.

### What didn't work
- The first attempt to commit the baseline failed in the pre-commit hook with:
  - `internal/cssvisualdiff/service/overlay.go:127:15: Error return value of f.Close is not checked (errcheck)`
  - `internal/cssvisualdiff/service/overlay.go:263:2: missing cases in switch of type service.LegendPosition: service.LegendBottomRight (exhaustive)`
  - `internal/cssvisualdiff/service/overlay_style.go:230:2: variable close has same name as predeclared identifier (predeclared)`
- I fixed those issues and reran the commit successfully.

### What I learned
- The repository pre-commit hook runs both lint and tests, so commits are a useful validation checkpoint.
- The current Go-rendered overlay design makes crop simpler than a CDP overlay implementation would: all annotations are already drawn onto an in-memory image.

### What was tricky to build
- The baseline implementation had a subtle screenshot decoding issue: `chromedp.FullScreenshot` with quality can return JPEG bytes, so the service now uses `image.Decode` with JPEG registered instead of assuming PNG input. The final output is still encoded as PNG.
- The crop plan has to distinguish annotation scope from image extent: selective selectors reduce labels, but only crop reduces the output image.

### What warrants a second pair of eyes
- The current baseline uses Go-drawn boxes instead of CDP Overlay domain highlights. That is intentional for reliability, but reviewers should confirm it satisfies the visual requirements.
- Crop coordinate assumptions rely on screenshot pixels matching document CSS pixels at device scale factor `1`.

### What should be done in the future
- Implement the crop tasks one by one.
- Add `.cropToTarget(name)` after `.cropTo(selector)` is stable.

### Code review instructions
- Start with `internal/cssvisualdiff/service/overlay.go` and `internal/cssvisualdiff/jsapi/overlay.go` for the baseline API.
- Review the new crop guide in `design-doc/01-overlay-screenshot-labels-design-and-implementation-guide.md` before reviewing crop implementation commits.
- Validate with `go test ./...` and rely on the pre-commit hook for lint + full tests.

### Technical details
- Baseline commit: `88c45cb51955185316485c32482d6c734fc2c558`.
- Crop V1 API planned: `.cropTo(selector)` and `.cropPadding(value)`.
- Crop V1 behavior planned: draw only targets intersecting the crop rect, translate target bounds into crop-local coordinates, and include only drawn targets in legend/result colors.

---

## Step 5: Crop Service Geometry and Rendering

I implemented the service-side crop model and wired it into the existing Go-rendered overlay pipeline. Crop now resolves a crop selector in document coordinates, expands it by typed padding, clamps it to the screenshot image, copies that region into a new image, translates target bounds into crop-local coordinates, and draws only targets that intersect the crop.

This turned crop from a design-only task into working image behavior. The implementation keeps full-page output as the default when `spec.Crop` is nil, so existing overlay screenshots remain unchanged.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Implement the documented crop tasks incrementally and commit each reviewable chunk.

**Inferred user intent:** The user wants crop support to be implemented with traceable steps, validation, and commits.

**Commit (code):** ea310891a0a77ffd3b518e959013d1ae18202c1f — "Add overlay crop geometry"

### What I did
- Added `OverlayCrop` to `OverlaySpec`.
- Added crop helpers in `internal/cssvisualdiff/service/overlay.go`:
  - `documentBoundsToRect`,
  - `expandRect`,
  - `translateBounds`,
  - `rectsOverlap`,
  - `resolveCropRect`,
  - `cropOverlayImage`.
- Wired crop into `OverlayScreenshot` before drawing boxes, labels, and legend.
- Made `OverlayResult.Targets` and `OverlayResult.Colors` reflect only targets drawn in the cropped output.

### Why
- Component-system organism PNGs need focused output dimensions, not full-page screenshots with selective labels.
- The service already had full-page screenshots and document-coordinate bounds, so crop belongs naturally in the service pipeline before drawing.

### What worked
- The Go-rendered overlay architecture made crop straightforward: crop the screenshot image, translate bounds, then draw as before.
- Pre-commit lint and full tests passed after removing unused helpers.

### What didn't work
- The first crop geometry commit attempt failed lint because some planned helpers were not yet used:
  - `documentBoundsToRect is unused`,
  - `rectToDocumentBounds is unused`,
  - `expandRect is unused`,
  - `translateBounds is unused`,
  - `rectsOverlap is unused`,
  - `uniformInsets is unused`.
- I fixed this by wiring the crop logic into `OverlayScreenshot` in the same commit and removing helpers that were still unused.

### What I learned
- In this repo, pre-commit lint prevents placeholder helpers from being committed, so implementation commits need to keep helpers and call sites together.
- Crop filtering should happen before colors/result targets are computed; otherwise legends and returned metadata include off-crop targets.

### What was tricky to build
- The important invariant is coordinate space: target bounds start in document coordinates, but drawing after crop happens in crop-local image coordinates. Translating by `cropRect.Min` after filtering keeps drawing logic unchanged.
- The result metadata also has to be filtered so userland sees what is actually visible in the output PNG.

### What warrants a second pair of eyes
- Confirm that clamping/translation behaves correctly on very tall pages and near image edges.
- Confirm that returning an error when a crop contains no targets is the preferred policy.

### What should be done in the future
- Add public `.cropToTarget(name)` if scripts frequently crop to an existing target.

### Code review instructions
- Review `OverlayScreenshot`, `resolveCropRect`, and `cropOverlayImage` in `internal/cssvisualdiff/service/overlay.go`.
- Validate with `go test ./internal/cssvisualdiff/service` or `go test ./...`.

### Technical details
- Crop default remains nil/full-page.
- V1 crop supports selector-based crop through `OverlayCrop.Selector`; `OverlayCrop.Target` exists in the model for follow-up use.

---

## Step 6: Crop Tests and JavaScript Builder API

I added tests for the service crop behavior and exposed crop through the fluent JS overlay builder. Scripts can now write `.cropTo(".hero")` and `.cropPadding(...)` on `cvd.overlaySpec()` before calling `page.overlay(spec).screenshot(path)`.

This completes the planned V1 crop API: focused cropped images are available from JavaScript, padding accepts scalar/two-value/four-value forms, and invalid padding is rejected before service execution.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Finish the crop task list by adding test coverage and JS-facing builder methods.

**Inferred user intent:** The user wants crop usable from scripts, not only available as an internal Go service field.

**Commit (code):** fdc97fc594f9d48c2e8d8fd6d8d72a7fddac1247 — "Test overlay crop rendering"

**Commit (code):** b5419498e58c74310f4eb07ccbaf2a1c6eb5d0b4 — "Add JS overlay crop builders"

### What I did
- Added a service test that crops `.hero` with 10px padding and verifies a `240x140` output for a `220x120` element.
- Added service assertions that an outside target is omitted from result targets/colors.
- Added a missing crop selector error test.
- Added `.cropTo(selector)` and `.cropPadding(value)` to the overlay spec builder.
- Added JS tests for four-value padding and invalid three-value padding.

### Why
- Crop is geometry-heavy, so tests need to verify dimensions and filtering, not only file existence.
- JS builder tests protect the opaque builder contract and catch accidental acceptance of malformed crop padding.

### What worked
- The existing `decodeInsets` helper was reusable for `.cropPadding(...)`.
- The service test could use deterministic CSS dimensions to assert exact crop output dimensions.
- Pre-commit lint and full tests passed on both commits.

### What didn't work
- N/A after the service wiring commit; tests and JS builder commits passed pre-commit on the first try.

### What I learned
- The output result object is a good place to expose crop filtering behavior: tests can assert `result.Targets` and `result.Colors` without image OCR.
- Keeping crop padding decoding in the builder gives fast script-level feedback for invalid API usage.

### What was tricky to build
- The service test needs a viewport and element dimensions that avoid crop clamping, otherwise expected width/height depends on page edges. I used a `.hero` at `(40,40)` with `220x120` size and `10px` padding so the expected crop is exactly `240x140`.

### What warrants a second pair of eyes
- Review whether `.cropPadding([vertical, horizontal])` and `.cropPadding([top, right, bottom, left])` should accept only integer values or also floats rounded to integers.
- Review whether missing crop selectors should be a hard error or produce a full-page fallback. Current behavior is a hard error.

### What should be done in the future
- Add a real example verb that creates organism crops for a component gallery.
- Add docs/help entries for `.cropTo` and `.cropPadding`.

### Code review instructions
- Review `internal/cssvisualdiff/service/overlay_test.go` for crop behavior expectations.
- Review `internal/cssvisualdiff/jsapi/overlay.go` and `internal/cssvisualdiff/jsapi/overlay_test.go` for builder behavior.
- Validate with `go test ./...`.

### Technical details
- `.cropPadding(24)` maps to all sides.
- `.cropPadding([16, 24])` maps to vertical/horizontal.
- `.cropPadding([8, 16, 24, 32])` maps to top/right/bottom/left.
- `.cropPadding([1, 2, 3])` is rejected.

---

## Step 7: Example Scripts and Fixture Validation

I added runnable example verbs and a deterministic local HTML fixture so the overlay API can be exercised outside unit tests. The examples cover both a single annotated full-page PNG and a component-system gallery with individual inspected components, a full-page organism map, and a cropped Hero-parts overlay.

I then served the fixture page locally and ran both example verbs with `go run` against the real CLI/Goja/browser stack. The commands generated the expected image, JSON, and HTML artifacts under `/tmp`.

### Prompt Context

**User prompt (verbatim):** "add example scripts and then actually test this against a test webpage (i don'tknow if you already have test pages), to verify that it all works, and to ask for my feedback."

**Assistant interpretation:** Add practical example scripts, create or reuse a test webpage, run the examples end-to-end, inspect the outputs, and ask the user to judge visual quality.

**Inferred user intent:** The user wants proof that the API works in a realistic workflow, not just unit tests, and wants to review the visual output before further polish.

### What I did
- Added `examples/pages/overlay-components.html` as a static fixture page.
- Added `examples/verbs/overlay-examples.js` with `annotated-png` and `gallery` verbs.
- Updated `examples/verbs/README.md` with commands and output descriptions.
- Served the repo with `python3 -m http.server 19876 --directory "$PWD"`.
- Ran the example verbs against `http://127.0.0.1:19876/examples/pages/overlay-components.html`.
- Used Python/PIL to verify generated PNG dimensions.
- Ran `go test ./...`.

### Why
- Examples make the new API discoverable and give operators a copy-paste workflow.
- A static fixture page provides stable selectors and dimensions for manual and automated validation.
- End-to-end CLI runs verify integration among verb scanning, Goja builders, browser navigation, overlay rendering, crop, artifact writing, and HTML output.

### What worked
- `examples overlay annotated-png` generated `/tmp/cssvd-overlay-example/full-page.organisms.annotated.png` with five organism labels.
- `examples overlay gallery` generated `/tmp/cssvd-overlay-gallery/index.html`, `components.json`, extracted component artifacts, full-page organism map, and `annotated/hero.parts.crop.png`.
- PNG dimensions matched expectations:
  - full-page outputs: `1280x1527`,
  - Hero crop: `1280x602`.
- `go test ./...` passed after adding examples.

### What didn't work
- Initial attempts to serve on ports `8767` and `9876` hit existing servers or wrong directories, causing fixture `404` responses and selector-not-found errors.
- The first gallery run failed with `unsupported inspect format "screenshot-css-json"`; the correct existing inspect format for full bundles is `bundle`.

### What I learned
- The example README should instruct users to serve from the repository root or use `--directory "$PWD"`; otherwise the fixture path can 404.
- Existing inspect formats are constrained to `bundle`, `png`, `html`, `css-json`, `css-md`, `inspect-json`, and `metadata-json`.
- End-to-end examples are valuable because they catch API assumptions that unit tests do not, such as verb naming and inspect artifact format strings.

### What was tricky to build
- JS verb names are exposed in kebab-case on the CLI (`annotatedPng` appears as `annotated-png`), so the README must show the CLI-facing name.
- The gallery example needs to bridge several output conventions: inspect artifacts are generated by the core service, overlay artifacts are explicit output paths, and the HTML page needs relative links.

### What warrants a second pair of eyes
- Visual aesthetics: labels and nested boxes work, but the Hero crop has dense labels/boxes and may obscure content. User feedback is needed on label placement, fill opacity, and whether boxes should default to outline-only.
- The gallery HTML is intentionally simple; it may need styling polish before being promoted as a polished example.

### What should be done in the future
- Add screenshot thumbnails or links from generated result rows to make CLI output easier to use.
- Consider defaulting content fill to transparent/outline-only if users find the overlays too visually heavy.
- Add a help doc page for overlay examples.

### Code review instructions
- Review `examples/verbs/overlay-examples.js` for API clarity and copy-paste suitability.
- Review `examples/pages/overlay-components.html` as the fixture source.
- Run:
  - `python3 -m http.server 19876 --directory "$PWD"`
  - `GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples overlay annotated-png http://127.0.0.1:19876/examples/pages/overlay-components.html /tmp/cssvd-overlay-example`
  - `GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples overlay gallery http://127.0.0.1:19876/examples/pages/overlay-components.html /tmp/cssvd-overlay-gallery`

### Technical details
- Generated artifacts inspected:
  - `/tmp/cssvd-overlay-example/full-page.organisms.annotated.png`
  - `/tmp/cssvd-overlay-gallery/annotated/full-page.organisms.png`
  - `/tmp/cssvd-overlay-gallery/annotated/hero.parts.crop.png`
  - `/tmp/cssvd-overlay-gallery/index.html`
- Validation command: `go test ./...`.
