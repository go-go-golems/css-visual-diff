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
