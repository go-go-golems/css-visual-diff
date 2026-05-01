# Tasks

## Done

- [x] Recenter design around current JS-first architecture and remove deprecated core YAML config/capture-mode work.
- [x] Define opaque/fluent JS API direction for `cvd.overlayTarget`, `cvd.overlaySpec`, and `page.overlay(spec).screenshot(path)`.
- [x] Separate real browser CSS (`page.css(...)`) from typed Go-side overlay styling.
- [x] Implement `page.css(...)` / driver style injection helper.
- [x] Implement typed overlay style structs, color parsing, defaults, and merge/normalization logic.
- [x] Implement `service.OverlayScreenshot` with full-page screenshot capture, document-coordinate bounds, Go-drawn boxes, labels, and legend.
- [x] Implement Goja overlay builders and opaque spec unwrapping.
- [x] Add service and JS API tests.

## TODO

- [ ] Add richer docs/help page entries for the new APIs.
- [ ] Add end-to-end example verb/script files under `examples/`.
- [ ] Consider V2 crop support: `.cropTo(selector)` / `.cropToTarget(name)`.
- [ ] Consider removable style handles for `page.css(...)`.
- [ ] Consider CDP Overlay-domain integration later if native compositor highlights are still desired; current implementation draws boxes in Go for reliable multi-target full-page output.
