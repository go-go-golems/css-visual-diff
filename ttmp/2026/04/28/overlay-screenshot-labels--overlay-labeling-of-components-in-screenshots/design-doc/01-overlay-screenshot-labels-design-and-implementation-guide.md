---
Title: Overlay Screenshot Labels - Design and Implementation Guide
Ticket: overlay-screenshot-labels
Status: active
Topics:
    - frontend
    - capture
    - chromedp
    - cdp
    - goja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cssvisualdiff/config/config.go
      Note: Config schema; OverlaySpec and OverlayTarget will be added here
    - Path: internal/cssvisualdiff/driver/chrome.go
      Note: Page screenshot and CDP primitives; overlay methods will be added here
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Goja page proxy registration; overlay builder will be wired here
    - Path: internal/cssvisualdiff/modes/capture.go
      Note: Capture orchestration; overlay screenshot hook will be added here
    - Path: internal/cssvisualdiff/service/dom.go
      Note: LocatorBounds and selector resolution used by overlay label positioning
ExternalSources: []
Summary: Comprehensive design and implementation guide for adding overlay labeling of components in screenshots to css-visual-diff, targeting a new intern audience.
LastUpdated: 2026-04-28T08:30:00-04:00
WhatFor: Provide architecture, API design, phased implementation plan, and risks for overlay screenshot annotation.
WhenToUse: When implementing or reviewing the overlay labeling feature.
---






# Overlay Screenshot Labels: Design and Implementation Guide

## Executive Summary

This document describes how to add **overlay labeling of components in screenshots** to `css-visual-diff`, a Go CLI tool that captures and compares rendered web pages across two browser targets. Today, the tool can take full-page and per-element screenshots, but those screenshots contain no visual indication of which UI component is which. The goal is to let users — both via YAML config and via JavaScript scripts running inside the tool's Goja runtime — declare a list of named components (by CSS selector), and produce annotated screenshots where each component has a visible bounding box, a text label, and optionally a legend that maps colors to component names.

The recommended approach is a **hybrid pipeline**:

1. **Chrome DevTools Protocol (CDP) Overlay domain** draws colored bounding boxes around DOM nodes. This is native to Chrome, does not alter the page's DOM, and is guaranteed to match the rendered layout.
2. **Go image post-processing** composites text labels and a legend panel onto the captured screenshot. This gives us precise control over typography, placement, and avoids any risk of page CSS interfering with annotation visibility.

The feature will be exposed at three layers:

- **Driver layer** (`driver/chrome.go`): low-level CDP Overlay commands.
- **Service layer** (`service/overlay.go`): high-level orchestration — map selectors to nodes, apply highlights, capture screenshot, composite labels.
- **JavaScript API layer** (`jsapi/overlay.go`): Goja bindings so that script authors can write `await page.overlay([{name: "NavBar", selector: "nav"}]).screenshot("/tmp/out.png")`.

---

## Table of Contents

1. [What `css-visual-diff` Is and How It Works](#what-css-visual-diff-is-and-how-it-works)
2. [Problem Statement](#problem-statement)
3. [Current-State Architecture](#current-state-architecture)
4. [Gap Analysis](#gap-analysis)
5. [Proposed Solution](#proposed-solution)
6. [Annotation Strategies Compared](#annotation-strategies-compared)
7. [API Design](#api-design)
8. [Data Models and Config Schema](#data-models-and-config-schema)
9. [Phased Implementation Plan](#phased-implementation-plan)
10. [Testing and Validation Strategy](#testing-and-validation-strategy)
11. [Risks, Alternatives, and Open Questions](#risks-alternatives-and-open-questions)
12. [References](#references)

---

## What `css-visual-diff` Is and How It Works

`css-visual-diff` is a command-line tool written in Go. Its purpose is to load two versions of a web page — typically an "original" implementation and a "React" refactor — and produce artifacts that help developers verify visual parity: screenshots, computed-style diffs, pixel diffs, and AI-generated reviews.

### The high-level flow

When you run `css-visual-diff run --config plan.yml`, the tool performs the following steps:

1. **Parse a YAML config** that declares two `Target` objects (`original` and `react`), a list of `SectionSpec` objects (named regions to inspect), and an `OutputSpec` (what files to write).
2. **Launch a headless Chrome browser** via `chromedp`, a Go library that speaks the Chrome DevTools Protocol (CDP). See `internal/cssvisualdiff/driver/chrome.go`.
3. **Navigate each target** to its URL, set the viewport, optionally run a prepare script (e.g., open a modal), and wait for stability.
4. **Capture screenshots** — both full-page and per-section — by evaluating CDP commands such as `Page.captureScreenshot`.
5. **Run additional modes** if requested: `cssdiff` (compute-style comparison), `pixeldiff` (image diff), `ai-review` (LLM-based assessment), and `html-report` (generate a summary page).

### Key architectural layers

Think of the codebase as four concentric layers, each depending only on the layers beneath it:

```
┌─────────────────────────────────────────┐
│  CLI / Modes                            │  cmd/css-visual-diff/main.go
│  (capture, compare, report generation)  │  internal/cssvisualdiff/modes/
├─────────────────────────────────────────┤
│  JavaScript API (Goja runtime)          │  internal/cssvisualdiff/jsapi/
│  (scriptable probes, inspect, diff)     │
├─────────────────────────────────────────┤
│  Service Layer                          │  internal/cssvisualdiff/service/
│  (business logic: snapshot, extract)    │
├─────────────────────────────────────────┤
│  Driver Layer                           │  internal/cssvisualdiff/driver/
│  (chromedp wrapper: browser, page)      │
└─────────────────────────────────────────┘
```

**Why this matters for overlay labels:** any new feature must be added bottom-up. First extend the `driver` so it can speak new CDP commands. Then add `service` functions that orchestrate those commands for a useful workflow. Then expose that workflow through the `jsapi` so scripts can use it. Finally, wire it into CLI `modes` if it should be available from YAML configs.

### How screenshots work today

Open `internal/cssvisualdiff/driver/chrome.go`. The `Page` struct wraps a `chromedp` context. It exposes two screenshot methods:

```go
func (p *Page) FullScreenshot(path string) error
func (p *Page) Screenshot(selector, path string) error
```

Both work by running a `chromedp` action that ultimately calls CDP's `Page.captureScreenshot`:

```go
var buf []byte
if err := chromedp.Run(p.ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
    return err
}
return os.WriteFile(path, buf, 0o644)
```

The `chromedp.FullScreenshot` helper scrolls the full page, stitches the image, and returns a PNG. The `chromedp.Screenshot` variant clips to a single element's bounding box (found via `DOM.querySelector` + `DOM.getBoxModel` under the hood).

This is the exact point where we will hook in our overlay drawing: **after the page is prepared and before the screenshot is captured**, we will ask Chrome to highlight the nodes we care about.

### How JavaScript scripts interact with the tool

`css-visual-diff` embeds a Goja JavaScript engine so that users can write reusable workflow scripts. The native module `css-visual-diff` is registered in `internal/cssvisualdiff/jsapi/module.go`. A script can write:

```js
const cvd = require("css-visual-diff")
const browser = await cvd.browser()
const page = await browser.page("https://example.com")
const statuses = await page.preflight([{ name: "cta", selector: "#cta" }])
```

The `page` object returned here is a Goja proxy wrapping a Go `*pageState` struct. Every method on that proxy (`goto`, `preflight`, `inspect`, `inspectAll`, `close`) is implemented in Go and marshals data across the JS/Go boundary. We will add an `overlay` method to this proxy.

---

## Problem Statement

Today, when `css-visual-diff` produces a screenshot of a section — say, `#hero-banner` — the output is a raw PNG of the rendered pixels. There is no indication on the image that this region is called "hero-banner", nor are any nested components labeled. In design reviews and documentation, stakeholders must mentally map selectors to visual regions. This is tedious and error-prone.

The user wants to:

- Provide a list of **named components** (each with a human-readable name and a CSS selector).
- Produce a screenshot where each matched component is visually distinguished — e.g., with a colored bounding box.
- Add a **text label** near each box showing the component name.
- Optionally include a **legend** (a color key) so that a single glance at the screenshot explains every annotation.

This must work for:

- **YAML-driven runs** (the existing `capture` mode should support an `overlay` field in the config).
- **Script-driven runs** (Goja scripts should be able to call an overlay API dynamically, after inspecting the page to discover selectors).

---

## Current-State Architecture

To understand where the new code fits, we need to map the existing subsystems that touch screenshots, selectors, and the JavaScript runtime.

### File map with responsibilities

| File | Lines | Responsibility |
|------|-------|----------------|
| `internal/cssvisualdiff/driver/chrome.go` | 176 | Browser lifecycle, viewport, navigation, screenshot, evaluate |
| `internal/cssvisualdiff/service/dom.go` | 174 | Selector → element queries: `LocatorStatus`, `LocatorBounds`, `LocatorText`, `LocatorHTML` |
| `internal/cssvisualdiff/service/extract.go` | ~150 | `ExtractElement` orchestrator: runs a list of extractors (exists, visible, text, bounds, computedStyle, attributes) |
| `internal/cssvisualdiff/service/snapshot.go` | ~50 | `SnapshotPage`: runs probes across a page, returns `PageSnapshot` |
| `internal/cssvisualdiff/service/inspect.go` | 415 | `InspectPreparedPage`: captures screenshots, HTML, JSON per selector; writes artifacts |
| `internal/cssvisualdiff/jsapi/module.go` | 552 | Goja module registration, `wrapPage`, `wrapBrowser`, promise helper |
| `internal/cssvisualdiff/jsapi/probe.go` | ~120 | Probe builder API for scripts: `cvd.probe("name").selector("#id").text().build()` |
| `internal/cssvisualdiff/jsapi/snapshot.go` | ~80 | `page.snapshot([...probes])` binding |
| `internal/cssvisualdiff/config/config.go` | 258 | YAML config structs: `Config`, `Target`, `SectionSpec`, `StyleSpec`, `OutputSpec` |
| `internal/cssvisualdiff/modes/capture.go` | 435 | `RunCapture`: orchestrates full capture workflow for original + react targets |

### How selectors resolve to elements today

In `service/dom.go`, the `LocatorBounds` function is representative:

```go
func LocatorBounds(page *driver.Page, locator LocatorSpec) (*Bounds, error) {
    selectorJSON, err := json.Marshal(locator.Selector)
    // ...
    script := fmt.Sprintf(`(() => {
      const selector = %s;
      let el = document.querySelector(selector);
      const rect = el.getBoundingClientRect();
      return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
    })()`, string(selectorJSON))
    var bounds *Bounds
    if err := page.Evaluate(script, &bounds); err != nil {
        return nil, err
    }
    return bounds, nil
}
```

This pattern — marshal a selector into a JS snippet, evaluate it via `chromedp.Evaluate`, and unmarshal the result into a Go struct — is used everywhere. Our overlay feature will follow the same pattern, but we will also need to obtain **CDP node IDs** so that the `Overlay` domain can highlight elements natively.

### The JavaScript API pattern

Open `internal/cssvisualdiff/jsapi/module.go` and look at `wrapPage`. Every method on the page proxy follows this template:

```go
_ = obj.Set("preflight", func(raw []map[string]any) goja.Value {
    return promiseValue(ctx, vm, "css-visual-diff.page.preflight", func() (any, error) {
        return state.runExclusive(func() (any, error) {
            probes, err := decodeProbes(raw)
            // ... business logic ...
            return lowerSelectorStatuses(statuses), nil
        })
    }, nil)
})
```

Key conventions:

- **Exclusive access**: `state.runExclusive` locks the page so that concurrent JS calls do not interleave CDP commands.
- **Promises**: all async work returns a Goja Promise. The `promiseValue` helper spawns a goroutine, performs the work, and resolves/rejects on the Goja event loop via `ctx.Owner.Post`.
- **Lowering**: Go structs are converted to plain JS objects via `lower*` functions (e.g., `lowerBounds`, `lowerViewport`).

Our overlay API must follow these exact conventions.

---

## Gap Analysis

| Desired Capability | Current State | Gap |
|--------------------|---------------|-----|
| Draw colored bounding boxes around elements | Not supported | Need CDP Overlay domain integration in driver |
| Capture screenshots with overlays visible | Screenshots are raw page pixels | Need to trigger highlights before `Page.captureScreenshot` |
| Composite text labels onto screenshots | No image processing | Need Go image manipulation (or DOM-injected labels) |
| Generate a legend mapping colors to names | Not supported | Need legend compositing logic |
| Declare overlay targets in YAML config | Config has `sections`, `styles` | Need `OverlaySpec` added to config schema |
| Scriptable overlay API | `page` proxy has no overlay methods | Need `page.overlay()` or similar in `jsapi` |
| Dynamic selector discovery + overlay | Scripts can inspect but not annotate | Need overlay builder in `jsapi` |

---

## Proposed Solution

### Overview

We introduce an **Overlay Service** that sits between the driver and the JavaScript API. The service accepts a list of `OverlayTarget` structs (name + selector), performs the following pipeline, and returns the path to an annotated screenshot:

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  OverlayTarget  │────▶│  Resolve Node IDs│────▶│  Apply Highlights│
│  (name, selector)│     │  (DOM.querySelector)│   │  (Overlay.highlightNode)│
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                          │
                                                          ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Composite Legend│◀────│  Draw Labels      │◀────│  Capture Screenshot  │
│  + Write PNG     │     │  (Go image/draw)  │     │  (Page.captureScreenshot)│
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

### Why CDP Overlay + Go compositing?

We evaluated three strategies (see next section). The winner is:

- **CDP `Overlay.highlightNode`** for bounding boxes because it draws on Chrome's compositor overlay layer. It is pixel-perfect, immune to page z-index or CSS `pointer-events`, and disappears cleanly.
- **Go `image/draw` + `golang.org/x/image/font`** for text labels and legend because it gives us complete control over positioning, contrast, and readability without fighting the page's own styles.

### The pipeline in detail

#### Step 1: Resolve selectors to CDP Node IDs

CDP's `Overlay` domain can highlight a DOM node, but it needs either a `nodeId` or a `backendNodeId`. We obtain these via the `DOM` domain:

```go
// Pseudocode — will become driver method
func (p *Page) ResolveNodeID(selector string) (cdp.NodeID, error) {
    var nodeID cdp.NodeID
    err := chromedp.Run(p.ctx,
        chromedp.QueryAfter(selector, func(ctx context.Context, execCtx cdp.Executor, nodes ...*cdp.Node) error {
            if len(nodes) == 0 {
                return fmt.Errorf("selector not found: %s", selector)
            }
            nodeID = nodes[0].NodeID
            return nil
        }, chromedp.ByQuery),
    )
    return nodeID, err
}
```

> **For interns:** `chromedp.QueryAfter` is a chromedp action that runs a callback after resolving nodes matching a selector. `cdp.NodeID` is an integer identifier that Chrome assigns to each DOM node for the lifetime of a CDP session. We need this ID because `Overlay.highlightNode` accepts it as a parameter.

#### Step 2: Apply highlights

The CDP `Overlay` domain defines `HighlightConfig`, a struct that controls the appearance of the highlight:

```go
// From github.com/chromedp/cdproto/overlay
import "github.com/chromedp/cdproto/overlay"

cfg := &overlay.HighlightConfig{
    ShowInfo:               true,
    ShowExtensionLines:     false,
    ContentColor:           &cdp.RGBA{R: 0, G: 150, B: 255, A: 0.3},
    PaddingColor:           &cdp.RGBA{R: 0, G: 150, B: 255, A: 0.2},
    BorderColor:            &cdp.RGBA{R: 0, G: 100, B: 200, A: 0.8},
    MarginColor:            &cdp.RGBA{R: 0, G: 100, B: 200, A: 0.1},
}
```

To highlight a node:

```go
err := chromedp.Run(p.ctx, overlay.HighlightNode(nodeID).WithHighlightConfig(cfg))
```

For multiple components, we assign each a distinct color from a pre-defined palette. The palette ensures sufficient contrast and avoids collisions for small numbers of components.

#### Step 3: Capture screenshot

After all highlights are applied, we capture the screenshot exactly as before:

```go
var buf []byte
err := chromedp.Run(p.ctx, chromedp.FullScreenshot(&buf, 90))
```

The highlights will be visible in the captured image because Chrome renders them on the compositor overlay before encoding the PNG.

#### Step 4: Composite labels and legend in Go

We decode the PNG into an `image.Image`, then use Go's standard `image/draw` package and a font library (e.g., `golang.org/x/image/font/opentype`) to draw:

1. **Per-component labels**: a small rounded rectangle + text placed just above the bounding box. We know the bounding box coordinates from Step 1 (or we can re-query them).
2. **Legend panel**: a box in a corner of the image listing all components with their color swatch and name.

```go
// Pseudocode for label drawing
func drawLabel(img *image.RGBA, bounds Bounds, name string, color color.RGBA) {
    // Draw bounding box outline (optional, if CDP highlight is not enough)
    // Draw label background
    // Draw label text
}
```

> **Note:** if Chrome's `Overlay.highlightNode` already draws a visible bounding box, we may skip drawing the box in Go and only draw the text label. This avoids double-rendering and keeps the image clean.

#### Step 5: Clean up highlights

After the screenshot is captured, we remove highlights so they do not pollute subsequent operations:

```go
_ = chromedp.Run(p.ctx, overlay.HideHighlight())
```

---

## Annotation Strategies Compared

### Strategy A: CDP Overlay only

Use `Overlay.highlightNode` with `ShowInfo: true`. Chrome draws a tooltip-like info box showing the node's tag name, id, classes, and dimensions.

- **Pros:** zero post-processing, no Go image dependencies.
- **Cons:** `ShowInfo` displays technical DOM info, not custom human-readable names. The info box styling is fixed by Chrome and may be hard to read in screenshots.

### Strategy B: DOM-injected labels

Inject a script into the page that creates absolutely positioned `div` elements for labels and a legend panel, similar to the reference Chrome extension.

- **Pros:** simple to implement with `page.Evaluate`; labels are part of the page so they appear in the screenshot automatically.
- **Cons:** injected elements can be affected by page CSS (z-index, `pointer-events`, transforms, `overflow: hidden` on ancestors). They may also shift layout if not carefully implemented with `position: fixed` and a high z-index. Cleaning them up requires another script injection.

### Strategy C: CDP Overlay + Go image compositing (recommended)

Use CDP Overlay for the bounding box (accurate, native, clean) and Go `image/draw` for the labels and legend (fully controlled, immune to page CSS).

- **Pros:** best of both worlds. Bounding boxes are pixel-perfect. Labels are readable and styled consistently. No DOM pollution.
- **Cons:** requires a Go image manipulation dependency. Label positioning must account for edges of the screenshot (labels near the top need to be drawn inside the image bounds, not above).

**Decision:** Use Strategy C.

---

## API Design

### Go Driver Extensions

Add two methods to `driver.Page` in `internal/cssvisualdiff/driver/chrome.go`:

```go
// HighlightNode applies a CDP overlay highlight to the node matched by selector.
// If no node matches, returns an error.
func (p *Page) HighlightNode(selector string, cfg *overlay.HighlightConfig) error

// HideHighlight removes all active CDP overlay highlights.
func (p *Page) HideHighlight() error

// ResolveNodeID returns the CDP NodeID for the first element matching selector.
func (p *Page) ResolveNodeID(selector string) (cdp.NodeID, error)
```

### Go Service Layer

Create a new file `internal/cssvisualdiff/service/overlay.go`:

```go
package service

import (
    "image"
    "image/color"

    "github.com/chromedp/cdproto/overlay"
    "github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

// OverlayTarget identifies one component to annotate.
type OverlayTarget struct {
    Name     string `json:"name"`
    Selector string `json:"selector"`
}

// OverlayResult holds the annotated image and metadata.
type OverlayResult struct {
    Image      image.Image       `json:"-"`
    OutputPath string            `json:"output_path"`
    Targets    []OverlayTarget   `json:"targets"`
    Colors     map[string]string `json:"colors"` // hex color per target name
}

// OverlayScreenshot captures a screenshot with annotated overlays.
// Steps:
//   1. Resolve each selector to a NodeID.
//   2. Apply a distinct HighlightConfig per target.
//   3. Capture full-page screenshot.
//   4. Decode PNG to image.Image.
//   5. Query bounding boxes for each target.
//   6. Draw text labels + legend.
//   7. Hide highlights.
//   8. Encode and write final PNG.
func OverlayScreenshot(page *driver.Page, targets []OverlayTarget, outPath string) (*OverlayResult, error)
```

### JavaScript API (Goja)

Create a new file `internal/cssvisualdiff/jsapi/overlay.go`. Expose an overlay builder and a convenience method on the page proxy.

```go
// In jsapi/module.go, inside wrapPage:
_ = obj.Set("overlay", func(raw []map[string]any) goja.Value {
    return wrapOverlayBuilder(ctx, vm, state, raw)
})
```

The builder API for scripts:

```js
// Builder pattern
const overlay = page.overlay([
  { name: "NavBar", selector: "nav.top" },
  { name: "Hero", selector: ".hero-section" },
  { name: "CTA", selector: "button.primary" },
])
const result = await overlay.screenshot("/tmp/annotated.png")
// result = { outputPath: "/tmp/annotated.png", colors: { NavBar: "#0096ff", ... } }
```

Or, for programmatic discovery:

```js
const statuses = await page.preflight([...])
const targets = statuses
  .filter(s => s.exists)
  .map(s => ({ name: s.name, selector: s.selector }))
const result = await page.overlay(targets).screenshot("/tmp/out.png")
```

The Goja wrapper will:

1. Decode the raw `[]map[string]any` into `[]service.OverlayTarget`.
2. Call `service.OverlayScreenshot` inside `state.runExclusive`.
3. Return a Promise that resolves to an object with `outputPath` and `colors`.

### CLI / Config Integration

Extend `config.Config` with an optional overlay field:

```go
// In config/config.go, add to Config:
Overlay *OverlaySpec `yaml:"overlay,omitempty"`

// And define:
type OverlaySpec struct {
    Enabled bool              `yaml:"enabled"`
    Targets []OverlayTarget   `yaml:"targets"`
    // Future: LabelPosition, LegendPosition, Palette, etc.
}

type OverlayTarget struct {
    Name     string `yaml:"name"`
    Selector string `yaml:"selector"`
}
```

Then, in `modes/capture.go`, after capturing the full screenshot but before iterating sections, optionally run the overlay pipeline and write an additional annotated screenshot:

```go
if cfg.Overlay != nil && cfg.Overlay.Enabled {
    annotatedPath := filepath.Join(output.Dir, fmt.Sprintf("%s-annotated.png", prefix))
    _, err := service.OverlayScreenshot(page.Page(), cfg.Overlay.Targets, annotatedPath)
    if err != nil {
        log.Warn().Err(err).Msg("overlay screenshot failed")
    }
}
```

---

## Data Models and Config Schema

### YAML Example

```yaml
metadata:
  slug: landing-page-overlay
  title: Landing Page with Overlay Labels

original:
  url: http://localhost:3000/original
  viewport:
    width: 1280
    height: 720

react:
  url: http://localhost:3000/react
  viewport:
    width: 1280
    height: 720

sections:
  - name: hero
    selector: ".hero"
  - name: features
    selector: "#features"

# NEW: overlay annotation
overlay:
  enabled: true
  targets:
    - name: NavBar
      selector: "nav.top"
    - name: Hero
      selector: ".hero-section"
    - name: CTA
      selector: "button.primary"

output:
  dir: ./out
  write_pngs: true
  write_json: true
```

### Color Palette

Define a default palette in `service/overlay.go`:

```go
var defaultPalette = []color.RGBA{
    {R: 0, G: 150, B: 255, A: 200},   // Blue
    {R: 255, G: 99, B: 71, A: 200},   // Tomato
    {R: 50, G: 205, B: 50, A: 200},   // LimeGreen
    {R: 255, G: 215, B: 0, A: 200},   // Gold
    {R: 186, G: 85, B: 211, A: 200},  // MediumOrchid
    {R: 0, G: 206, B: 209, A: 200},   // DarkTurquoise
    {R: 255, G: 105, B: 180, A: 200}, // HotPink
    {R: 255, G: 140, B: 0, A: 200},   // DarkOrange
}
```

Each target receives a color by index (`i % len(palette)`). The same color is used for the CDP highlight and the Go-drawn label.

---

## Phased Implementation Plan

### Phase 1: Driver Foundation

**Files to modify:**
- `internal/cssvisualdiff/driver/chrome.go`

**Tasks:**
1. Add `ResolveNodeID(selector string) (cdp.NodeID, error)`.
2. Add `HighlightNode(selector string, cfg *overlay.HighlightConfig) error`.
3. Add `HideHighlight() error`.
4. Add unit-style integration test in `driver/chrome_test.go` (if it exists; else create `driver_test.go`) that launches a headless browser, loads a data URI HTML page with a div, highlights it, captures a screenshot, and asserts the image is non-empty.

**Validation:**
```bash
GOWORK=off go test ./internal/cssvisualdiff/driver/... -v -run TestHighlightNode
```

### Phase 2: Service Layer

**Files to create:**
- `internal/cssvisualdiff/service/overlay.go`
- `internal/cssvisualdiff/service/overlay_test.go`

**Tasks:**
1. Define `OverlayTarget`, `OverlayResult`, and `OverlayScreenshot`.
2. Implement the pipeline: resolve → highlight → screenshot → decode → label → encode → write.
3. For image manipulation, add a dependency such as `golang.org/x/image` and `github.com/golang/freetype` (or `github.com/fogleman/gg` for higher-level drawing). **Decision needed:** `gg` is more ergonomic for rectangles and text; `x/image/draw` + `freetype` is lighter. For an intern-friendly codebase, `gg` is recommended.
4. Implement `drawLabel` and `drawLegend` helpers.
5. Handle edge cases: labels that would draw above the image top edge should be drawn inside the image instead; legend should not obscure important content (default to bottom-right corner with a semi-transparent background).

**Validation:**
```bash
GOWORK=off go test ./internal/cssvisualdiff/service/... -v -run TestOverlayScreenshot
```

### Phase 3: JavaScript API

**Files to create/modify:**
- `internal/cssvisualdiff/jsapi/overlay.go` (new)
- `internal/cssvisualdiff/jsapi/module.go` (add `overlay` to `wrapPage`)

**Tasks:**
1. Implement `wrapOverlayBuilder` that returns a Goja object with `.screenshot(path)`.
2. Decode `[]map[string]any` into `[]service.OverlayTarget` with validation (name and selector required).
3. Follow the `promiseValue` + `runExclusive` pattern exactly.
4. Return a plain object: `{ outputPath: string, colors: Record<string, string> }`.

**Validation (manual script test):**
Create `examples/scripts/test-overlay.js`:

```js
const cvd = require("css-visual-diff")
async function main() {
  const browser = await cvd.browser()
  const page = await browser.page("https://example.com")
  const result = await page.overlay([
    { name: "Heading", selector: "h1" },
  ]).screenshot("/tmp/test-overlay.png")
  console.log(JSON.stringify(result, null, 2))
  await browser.close()
}
main()
```

Run:
```bash
GOWORK=off go run ./cmd/css-visual-diff verbs script run examples/scripts/test-overlay.js
```

### Phase 4: Config Schema and CLI Integration

**Files to modify:**
- `internal/cssvisualdiff/config/config.go`
- `internal/cssvisualdiff/modes/capture.go`

**Tasks:**
1. Add `OverlaySpec` and `OverlayTarget` to config structs.
2. Add validation: if `overlay.enabled` is true, each target must have a non-empty `name` and `selector`.
3. In `captureTarget` (in `modes/capture.go`), after the full screenshot is captured, check `cfg.Overlay` and call `service.OverlayScreenshot` to produce an `*-annotated.png`.
4. Include the annotated screenshot path in the JSON/Markdown output.

**Validation:**
```bash
GOWORK=off go run ./cmd/css-visual-diff run --config examples/overlay-test.yaml
```

### Phase 5: Documentation and Examples

**Files to create:**
- `examples/overlay-example.yaml`
- `examples/scripts/overlay-dynamic.js`

**Tasks:**
1. Write a self-contained YAML example that demonstrates overlay annotation.
2. Write a JavaScript verb example that discovers selectors dynamically and overlays them.
3. Update `README.md` with a section on overlay mode.

---

## Testing and Validation Strategy

### Unit tests

- **Driver tests:** mock or launch a real Chrome instance to verify that `HighlightNode` + `FullScreenshot` produces an image different from a non-highlighted screenshot. Use image hash comparison (perceptual or simple average color shift).
- **Service tests:** use a static HTML fixture served via `httptest.Server`. Capture an overlay screenshot and assert that the output file exists and is larger than the non-annotated version (labels add pixels).
- **Config tests:** in `config/config_test.go`, add test cases for valid and invalid `OverlaySpec` blocks.

### Integration tests

- Add an overlay target to one of the existing example configs and verify the CLI run produces `original-annotated.png` and `react-annotated.png`.
- Run the JavaScript verb example manually and visually inspect the PNG.

### Visual regression guard

Because this feature produces images, automated pixel-perfect assertions are brittle. Instead:

- Keep a "golden" annotated screenshot in `testdata/` and compare using perceptual hashing (e.g., `github.com/vitali-fedulov/images`).
- Or, at minimum, assert that the image dimensions match the viewport and that the average color is shifted by overlay highlights (statistical smoke test).

---

## Risks, Alternatives, and Open Questions

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| CDP Overlay highlights do not appear in `Page.captureScreenshot` on some Chrome versions | Low | High | Validate on CI Chrome version; fallback to DOM-injected boxes if CDP fails |
| Go image library adds heavy dependency | Low | Medium | Use `golang.org/x/image` (already common) or keep `gg` optional |
| Labels overlap or go off-screen | Medium | Medium | Implement bounding-box clamping; place labels inside image when near edges |
| Performance degradation for many targets | Medium | Low | Batch CDP calls; keep palette small; legend scales vertically |

### Alternatives considered

- **DOM-injected labels only:** rejected due to CSS interference risk.
- **CDP `Overlay.setShowGrid` / `setShowFlexOverlay`:** these are for debugging CSS layout, not for arbitrary component labeling. Not applicable.
- **Third-party screenshot SaaS:** rejected because the tool must work offline and in CI.

### Open questions

1. **Should the legend be optional?** Yes — add `OverlaySpec.Legend bool` (default true).
2. **Should label position be configurable?** For the first version, always place labels above the bounding box, clamped to image bounds. Future: `LabelPosition: "above" | "below" | "auto"`.
3. **Should overlay work with per-section screenshots?** For V1, only full-page annotated screenshots. Per-section overlays can be added later by cropping the annotated full page or by running the pipeline on a clipped viewport.
4. **Font choice:** use a built-in font (Go's `basicfont` or embed a small TTF) to avoid system dependency. **Decision:** embed `golang.org/x/image/font/gofont/goregular` for consistency across OSes.
5. **Should scripts have raw CDP access?** The user suggested "tools to walk the DOM themselves (or at least interact with cdp, maybe through scripting, to do so)". This is out of scope for the overlay feature but is a natural follow-up: expose `page.cdp(action)` that accepts a CDP command descriptor. Not sketched here per user request.

---

## References

### Files referenced in this document

| Absolute Path | Why it matters |
|---------------|----------------|
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/driver/chrome.go` | Browser and page lifecycle, screenshot primitives, viewport control. The overlay driver methods (`HighlightNode`, `HideHighlight`) will be added here. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/dom.go` | Selector resolution, bounding-box queries. `OverlayScreenshot` will reuse `LocatorBounds` for label positioning. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/extract.go` | Extractor orchestration pattern. Overlay targets are conceptually similar to extractors. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/service/inspect.go` | Artifact writing and metadata patterns. Overlay results should follow the same artifact conventions. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/module.go` | Goja module registration and `wrapPage`. The `overlay` builder will be registered here. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/jsapi/probe.go` | Builder pattern for probes. The overlay builder should mirror this API style. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/config/config.go` | Config schema. `OverlaySpec` and `OverlayTarget` will be added here. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/modes/capture.go` | Capture orchestration. Overlay screenshot production will be hooked into `captureTarget`. |
| `/home/manuel/code/wesen/2026-04-25--overlay-select-components/extension/content_scripts/modules/dom-overlay.js` | Reference DOM overlay implementation. Demonstrates visual design of boxes and labels (used for inspiration, not implementation). |

### External API references

- **Chrome DevTools Protocol — Overlay domain:** https://chromedevtools.github.io/devtools-protocol/tot/Overlay/
  - `Overlay.highlightNode`
  - `Overlay.hideHighlight`
  - `Overlay.HighlightConfig`
- **Chrome DevTools Protocol — DOM domain:** https://chromedevtools.github.io/devtools-protocol/tot/DOM/
  - `DOM.querySelector`
  - `DOM.getBoxModel`
- **chromedp Go package:** https://pkg.go.dev/github.com/chromedp/chromedp
  - `chromedp.QueryAfter`
  - `chromedp.FullScreenshot`
  - `chromedp.Run`
- **chromedp/cdproto/overlay:** https://pkg.go.dev/github.com/chromedp/cdproto/overlay
  - `overlay.HighlightConfig`
  - `overlay.HighlightNode`
- **Goja JavaScript engine:** https://pkg.go.dev/github.com/dop251/goja
  - `Runtime.NewPromise`
  - `Runtime.ToValue`
- **Go image/draw:** https://pkg.go.dev/image/draw
- **Go x/image/font:** https://pkg.go.dev/golang.org/x/image/font
- **fogleman/gg (optional):** https://pkg.go.dev/github.com/fogleman/gg
  - Higher-level 2D drawing: `gg.NewContext`, `gg.DrawRectangle`, `gg.LoadFontFace`, `gg.DrawStringAnchored`

---

## Appendix: Complete Pseudocode for `OverlayScreenshot`

This appendix ties together all concepts into one readable pseudocode function. It is not copy-paste Go, but it maps closely to the intended implementation.

```go
func OverlayScreenshot(page *driver.Page, targets []OverlayTarget, outPath string) (*OverlayResult, error) {
    // 1. Resolve selectors and assign colors
    type annotatedTarget struct {
        OverlayTarget
        nodeID cdp.NodeID
        color  color.RGBA
    }
    var annotated []annotatedTarget
    for i, t := range targets {
        nodeID, err := page.ResolveNodeID(t.Selector)
        if err != nil {
            return nil, fmt.Errorf("resolve %q: %w", t.Selector, err)
        }
        annotated = append(annotated, annotatedTarget{
            OverlayTarget: t,
            nodeID:        nodeID,
            color:         defaultPalette[i % len(defaultPalette)],
        })
    }

    // 2. Apply CDP highlights
    for _, at := range annotated {
        cfg := &overlay.HighlightConfig{
            ShowInfo:      false, // we draw our own labels
            ContentColor:  rgbaToCDP(at.color, 0.3),
            PaddingColor:  rgbaToCDP(at.color, 0.2),
            BorderColor:   rgbaToCDP(at.color, 0.8),
            MarginColor:   rgbaToCDP(at.color, 0.1),
        }
        if err := page.HighlightNode(at.Selector, cfg); err != nil {
            return nil, err
        }
    }

    // 3. Capture screenshot
    var buf []byte
    if err := chromedp.Run(page.Context(), chromedp.FullScreenshot(&buf, 90)); err != nil {
        _ = page.HideHighlight()
        return nil, err
    }

    // 4. Decode PNG
    img, err := png.Decode(bytes.NewReader(buf))
    if err != nil {
        _ = page.HideHighlight()
        return nil, err
    }
    rgba := image.NewRGBA(img.Bounds())
    draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

    // 5. Query bounds for label placement
    for i := range annotated {
        bounds, err := service.LocatorBounds(page, service.LocatorSpec{Selector: annotated[i].Selector})
        if err != nil {
            continue // skip label if element disappeared
        }
        annotated[i].bounds = bounds
    }

    // 6. Draw labels
    for _, at := range annotated {
        if at.bounds == nil {
            continue
        }
        drawLabel(rgba, *at.bounds, at.Name, at.color)
    }

    // 7. Draw legend
    drawLegend(rgba, annotated)

    // 8. Hide highlights
    _ = page.HideHighlight()

    // 9. Encode and write
    f, err := os.Create(outPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    if err := png.Encode(f, rgba); err != nil {
        return nil, err
    }

    // 10. Build result
    colors := make(map[string]string)
    for _, at := range annotated {
        colors[at.Name] = rgbaToHex(at.color)
    }
    return &OverlayResult{
        Image:      rgba,
        OutputPath: outPath,
        Targets:    targets,
        Colors:     colors,
    }, nil
}
```

---

*Document version: 2026-04-28*
*Ticket: overlay-screenshot-labels*
