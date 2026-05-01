# Changelog

## 2026-04-28

- Initial workspace created


## 2026-04-28

Created ticket and wrote comprehensive design/implementation guide for overlay screenshot labels

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/css-visual-diff/ttmp/2026/04/28/overlay-screenshot-labels--overlay-labeling-of-components-in-screenshots/design-doc/01-overlay-screenshot-labels-design-and-implementation-guide.md — Primary design document

## 2026-05-01

Updated the overlay design guide to match the current JavaScript-first architecture.

### Changes

- Removed stale implementation guidance for the deleted native `run --config` YAML pipeline.
- Reframed overlay targets as JavaScript API/userland data rather than Go-owned config schema.
- Added script sketches for name-to-selector maps, userland YAML loading, annotated full-page PNG export, component-system extraction, HTML galleries, annotated organisms, and annotated full screens.
- Updated phased implementation plan to prioritize JS API examples and docs instead of config/capture-mode integration.

## 2026-05-01 v2

Refined the JavaScript API design for safer typed inputs and style customization.

### Changes

- Replaced raw overlay input arrays/maps with opaque `OverlaySpec` and `OverlayTargetSpec` values produced by fluent builders.
- Added `cvd.overlayTarget(name)` and `cvd.overlaySpec()` builder sketches.
- Added target-level and spec-level `.css(...)` styling for label, legend, and CDP highlight appearance.
- Updated script examples to build typed specs before calling `page.overlay(spec).screenshot(path)`.
- Updated service pseudocode to accept `OverlaySpec`, parse constrained overlay CSS, merge per-target styles, and render labels/legend from resolved styles.

## 2026-05-01 v3

Separated real browser CSS from Go-side overlay styling.

### Changes

- Reserved `.css(...)` for real browser CSS injection through Chromium/page APIs.
- Removed the custom CSS-like overlay styling DSL and CSS parser recommendation.
- Replaced overlay styling with typed `.style({...})` objects plus fluent methods such as `.borderColor(...)`, `.labelBackground(...)`, and `.labelPosition(...)`.
- Updated examples to use typed style objects for labels, legends, target defaults, and per-target colors.
- Updated service model and pseudocode to normalize typed styles instead of parsing overlay CSS.

## 2026-05-01 implementation

Implemented the first working overlay screenshot API.

### Changes

- Added `driver.Page.FullScreenshotBytes` for screenshot pipelines that need in-memory image processing.
- Added `driver.Page.AddStyleTag` and JS `page.css(cssText)` for real browser CSS injection.
- Added typed overlay style structs, color parsing, defaults, and merge logic in `internal/cssvisualdiff/service/overlay_style.go`.
- Added `service.OverlayScreenshot` in `internal/cssvisualdiff/service/overlay.go`, drawing bounding boxes, labels, and legends in Go after full-page screenshot capture.
- Added opaque Goja builders `cvd.overlayTarget(name)` and `cvd.overlaySpec()` plus `page.overlay(spec).screenshot(path)`.
- Added tests for service screenshot output and JS builder/spec opacity.

### Validation

- `go test ./...` passes.

## 2026-05-01 crop support

Implemented crop support for overlay screenshots.

### Changes

- Added `OverlayCrop` to the service model with selector/target fields and typed padding.
- Added crop geometry helpers for document-bounds to image-rect conversion, padding expansion, overlap checks, and coordinate translation.
- Wired crop into `OverlayScreenshot`: cropped outputs now draw only intersecting targets and report only those targets/colors.
- Added service tests for crop selector output dimensions, padding behavior, target filtering, and missing crop selector errors.
- Added JS builder methods `.cropTo(selector)` and `.cropPadding(value)`.
- Added JS builder tests for crop selector/padding and invalid padding arrays.

### Commits

- `ea31089` — Add overlay crop geometry
- `fdc97fc` — Test overlay crop rendering
- `b541949` — Add JS overlay crop builders

### Validation

- Pre-commit lint and `GOWORK=off go test ./...` passed for all crop code commits.

## 2026-05-01 example scripts and fixture validation

Added runnable overlay example verbs and validated them against a local fixture page.

### Changes

- Added `examples/pages/overlay-components.html` as a deterministic static page with header, hero, feature grid, newsletter, and footer organisms.
- Added `examples/verbs/overlay-examples.js` with:
  - `examples overlay annotated-png` for one full-page annotated organism map.
  - `examples overlay gallery` for component extraction, full-page organism map, cropped Hero-parts overlay, JSON, and an HTML gallery.
- Updated `examples/verbs/README.md` with commands and expected output files.

### Validation

- Served the fixture with `python3 -m http.server 19876 --directory "$PWD"`.
- Ran `GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples overlay annotated-png http://127.0.0.1:19876/examples/pages/overlay-components.html /tmp/cssvd-overlay-example`.
- Ran `GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs examples overlay gallery http://127.0.0.1:19876/examples/pages/overlay-components.html /tmp/cssvd-overlay-gallery`.
- Verified generated PNG dimensions:
  - `/tmp/cssvd-overlay-example/full-page.organisms.annotated.png` — `1280x1527`.
  - `/tmp/cssvd-overlay-gallery/annotated/full-page.organisms.png` — `1280x1527`.
  - `/tmp/cssvd-overlay-gallery/annotated/hero.parts.crop.png` — `1280x602`.
- `go test ./...` passes.

### Review notes

- The examples functionally work and produce the expected artifacts.
- Visual review suggests label/overlay aesthetics need user feedback: the cropped Hero parts are correctly labeled, but dense nested boxes can overlap and may obscure content.

