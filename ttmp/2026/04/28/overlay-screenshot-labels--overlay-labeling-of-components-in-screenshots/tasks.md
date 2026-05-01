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
- [x] Commit baseline overlay API implementation (`88c45cb`).
- [x] Add crop support implementation guide to design doc and upload v4 to reMarkable.
- [x] Crop Task 1: Add `OverlayCrop` service model and crop geometry helpers (`ea31089`).
- [x] Crop Task 2: Wire crop logic into `OverlayScreenshot` before drawing annotations (`ea31089`).
- [x] Crop Task 3: Add service tests for crop selector, padding, and filtering (`fdc97fc`).
- [x] Crop Task 4: Add JS builder methods `.cropTo(...)` and `.cropPadding(...)` (`b541949`).
- [x] Crop Task 5: Add JS builder tests for crop methods and invalid padding (`b541949`).
- [x] Crop Task 6: Run full validation and commit crop implementation (pre-commit lint + `go test ./...` passed for each crop code commit).

## Future TODO

- [ ] Add richer docs/help page entries for the new APIs.
- [ ] Add end-to-end example verb/script files under `examples/`.
- [ ] Consider V2 crop support: `.cropToTarget(name)` public JS builder method (service model already has a `Target` field).
- [ ] Consider removable style handles for `page.css(...)`.
- [ ] Consider CDP Overlay-domain integration later if native compositor highlights are still desired; current implementation draws boxes in Go for reliable multi-target full-page output.
