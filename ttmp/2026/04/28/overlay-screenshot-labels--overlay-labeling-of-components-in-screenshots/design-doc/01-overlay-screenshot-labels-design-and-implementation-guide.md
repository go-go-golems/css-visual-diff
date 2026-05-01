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
    - Path: internal/cssvisualdiff/driver/chrome.go
      Note: Page screenshot and CDP primitives; overlay methods will be added here
    - Path: internal/cssvisualdiff/service/overlay.go
      Note: New high-level overlay screenshot service; resolves selectors, applies highlights, and composites labels
    - Path: internal/cssvisualdiff/jsapi/module.go
      Note: Goja page proxy registration; overlay builder will be wired here
    - Path: internal/cssvisualdiff/jsapi/overlay.go
      Note: New Goja overlay builder implementation
    - Path: internal/cssvisualdiff/service/dom.go
      Note: LocatorBounds and selector resolution used by overlay label positioning
    - Path: internal/cssvisualdiff/verbcli/bootstrap.go
      Note: Current small app config only discovers JS verb repositories; it is not a visual-diff spec schema
ExternalSources: []
Summary: Comprehensive design and implementation guide for adding overlay labeling of components in screenshots to css-visual-diff, targeting a new intern audience.
LastUpdated: 2026-05-01T13:00:00-04:00
WhatFor: Provide architecture, API design, phased implementation plan, and risks for overlay screenshot annotation.
WhenToUse: When implementing or reviewing the overlay labeling feature.
---






# Overlay Screenshot Labels: Design and Implementation Guide

## Executive Summary

This document describes how to add **overlay labeling of components in screenshots** to `css-visual-diff`, a Go CLI tool that captures and compares rendered web pages. The current tool is intentionally **JavaScript-first**: the old native `run --config` YAML pipeline has been removed, and project-specific orchestration belongs in JavaScript verbs using `require("css-visual-diff")`. Today, the tool can take full-page and per-element screenshots, but those screenshots contain no visual indication of which UI component is which. The goal is to let JavaScript scripts declare typed overlay specs through fluent builders, avoiding untyped `map[string]any` payloads, and produce annotated PNGs where each component has a visible bounding box, text label, optional legend, and typed Go-side style customization for colors, labels, and legend appearance, while reserving `.css(...)` for real browser CSS only.

The recommended approach is a **hybrid pipeline**:

1. **Chrome DevTools Protocol (CDP) Overlay domain** draws colored bounding boxes around DOM nodes. This is native to Chrome, does not alter the page's DOM, and is guaranteed to match the rendered layout.
2. **Go image post-processing** composites text labels and a legend panel onto the captured screenshot. This gives us precise control over typography, placement, and avoids any risk of page CSS interfering with annotation visibility.

The feature will be exposed at three layers:

- **Driver layer** (`driver/chrome.go`): low-level CDP Overlay commands.
- **Service layer** (`service/overlay.go`): high-level orchestration — map selectors to nodes, apply highlights, capture screenshot, composite labels.
- **JavaScript API layer** (`jsapi/overlay.go`): Goja bindings exposing opaque, fluent overlay specs so script authors can write `await page.overlay(cvd.overlaySpec().target(cvd.overlayTarget("NavBar").selector("nav")).build()).screenshot("/tmp/out.png")`.

---

## Table of Contents

1. [What `css-visual-diff` Is and How It Works](#what-css-visual-diff-is-and-how-it-works)
2. [Problem Statement](#problem-statement)
3. [Current-State Architecture](#current-state-architecture)
4. [Gap Analysis](#gap-analysis)
5. [Proposed Solution](#proposed-solution)
6. [Annotation Strategies Compared](#annotation-strategies-compared)
7. [API Design](#api-design)
8. [JavaScript Data Model and Script Examples](#javascript-data-model-and-script-examples)
9. [Phased Implementation Plan](#phased-implementation-plan)
10. [Testing and Validation Strategy](#testing-and-validation-strategy)
11. [Risks, Alternatives, and Open Questions](#risks-alternatives-and-open-questions)
12. [References](#references)

---

## What `css-visual-diff` Is and How It Works

`css-visual-diff` is a command-line tool written in Go. Its purpose is to load two versions of a web page — typically an "original" implementation and a "React" refactor — and produce artifacts that help developers verify visual parity: screenshots, computed-style diffs, pixel diffs, and AI-generated reviews.

### The high-level flow

The current tool has two live usage paths:

1. **Direct commands** such as `css-visual-diff compare`, `llm-review`, `chromedp-probe`, and `serve`. These are narrow, built-in commands for one-off work or serving generated review datasets.
2. **JavaScript verbs** invoked through `css-visual-diff verbs ...`. These scripts are scanned from verb repositories and use `require("css-visual-diff")` to launch browsers, open pages, inspect selectors, compare regions, write artifacts, and load any project-specific data they need.

A typical JS-first workflow performs the following steps:

1. A JS verb loads project data if needed (YAML specs, JSON registries, Storybook metadata, or inline selector maps).
2. The script calls `const cvd = require("css-visual-diff")` and creates a browser/page.
3. The page is prepared by script code: viewport, waits, navigation, application-specific clicks or JS evaluation.
4. The script calls native page APIs such as `preflight`, `inspectAll`, `snapshot`, and — after this feature — `overlay(...).screenshot(path)`.
5. The script writes project-shaped outputs and returns structured rows for Glazed/CLI formatting.

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

**Why this matters for overlay labels:** any new feature must be added bottom-up. First extend the `driver` so it can speak new CDP commands. Then add `service` functions that orchestrate those commands for a useful workflow. Then expose that workflow through the `jsapi` so scripts can use it. Do **not** add a new native manifest/config format for overlays; if a project wants YAML, its JavaScript verb should load that YAML as userland data and convert it into overlay targets.

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

This must work for **script-driven runs**. Goja scripts should be able to call an overlay API dynamically, including after inspecting the page to discover selectors. If a project stores selector definitions in YAML, JSON, or a component registry, that parsing happens in JavaScript userland before calling the native overlay primitive.

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
| `internal/cssvisualdiff/verbcli/bootstrap.go` | ~330 | Small app config loader for discovering JS verb repositories from `.css-visual-diff.yml`; not a visual spec schema |
| `cmd/css-visual-diff/main.go` | ~380 | Cobra root command; registers direct commands and `verbs`, but no old `run --config` pipeline |

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
_ = obj.Set("someAsyncPageMethod", func(value goja.Value) goja.Value {
    return promiseValue(ctx, vm, "css-visual-diff.page.someAsyncPageMethod", func() (any, error) {
        return state.runExclusive(func() (any, error) {
            spec, err := decodeTypedOrOpaqueSpec(value)
            if err != nil {
                return nil, err
            }
            // ... business logic ...
            return lowerResult(result), nil
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
| Declare overlay targets from JavaScript | Scripts can build probes but not typed overlay specs | Need opaque `OverlaySpec` / `OverlayTargetSpec` builders instead of `[]map[string]any` |
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

// OverlaySpec is the typed, validated service input produced by the JS builders.
type OverlaySpec struct {
    Targets    []OverlayTarget `json:"targets"`
    Legend     bool            `json:"legend"`
    Screenshot string          `json:"screenshot"` // "fullPage" for V1
    Style      OverlayStyle    `json:"style"`
}

// OverlayTarget identifies one component to annotate.
type OverlayTarget struct {
    Name     string             `json:"name"`
    Selector string             `json:"selector"`
    Label    string             `json:"label,omitempty"`
    Style    TargetOverlayStyle `json:"style"`
}

// OverlayStyle controls Go-side label/legend rendering and target defaults.
type OverlayStyle struct {
    Label          LabelOverlayStyle  `json:"label"`
    Legend         LegendOverlayStyle `json:"legend"`
    TargetDefaults TargetOverlayStyle `json:"target_defaults"`
}

// TargetOverlayStyle controls one target's CDP highlight and label appearance.
type TargetOverlayStyle struct {
    BorderColor       *color.RGBA `json:"border_color,omitempty"`
    ContentBackground *color.RGBA `json:"content_background,omitempty"`
    PaddingBackground *color.RGBA `json:"padding_background,omitempty"`
    MarginBackground  *color.RGBA `json:"margin_background,omitempty"`
    BorderWidth       int         `json:"border_width,omitempty"`
    LabelColor        *color.RGBA `json:"label_color,omitempty"`
    Label             LabelTargetStyle `json:"label"`
}

// LabelOverlayStyle and LegendOverlayStyle are intentionally typed structs, not CSS.
type LabelOverlayStyle struct { /* font family, font size, radius, padding */ }
type LegendOverlayStyle struct { /* position, background, text color */ }
type LabelTargetStyle struct { /* background, text color, position */ }

// OverlayResult holds the annotated image and metadata.
type OverlayResult struct {
    Image      image.Image       `json:"-"`
    OutputPath string            `json:"output_path"`
    Targets    []OverlayTarget   `json:"targets"`
    Colors     map[string]string `json:"colors"` // hex color per target name
}

// OverlayScreenshot captures a screenshot with annotated overlays.
// Steps:
//   1. Normalize typed spec-level + target-level overlay styles.
//   2. Resolve each selector to a NodeID.
//   3. Apply a distinct HighlightConfig per target.
//   4. Capture full-page screenshot.
//   5. Decode PNG to image.Image.
//   6. Query bounding boxes for each target.
//   7. Draw text labels + legend.
//   8. Hide highlights.
//   9. Encode and write final PNG.
func OverlayScreenshot(page *driver.Page, spec OverlaySpec, outPath string) (*OverlayResult, error)
```

### JavaScript API (Goja)

Create a new file `internal/cssvisualdiff/jsapi/overlay.go`. Expose **opaque overlay spec objects** and fluent builders. The important design constraint is: do **not** make the page method accept `[]map[string]any` or arbitrary arrays of plain objects. That pattern is too loose, makes validation inconsistent, and leaks Go implementation details into user scripts.

Also keep a sharp semantic boundary:

- `.css(...)` means **real browser CSS** that is injected/evaluated in Chromium.
- Overlay label/legend/highlight appearance is **Go-side renderer style**, configured with typed `.style(...)` objects and fluent style methods.

This avoids a confusing fake CSS language where users expect variables, media queries, selectors, `calc()`, DevTools visibility, or normal cascade behavior, but the Go renderer only supports a tiny subset.

#### Browser CSS API

Add or document a page-level browser CSS helper separately from overlays:

```js
const injected = await page.css(`
  html { scroll-behavior: auto !important; }
  body { background: white !important; }
`)

// Optional future API if we return handles:
// await injected.remove()
```

This is ordinary browser CSS. It is useful for stabilizing screenshots, disabling animations, forcing print-like backgrounds, or applying project-specific debug outlines. It should not control Go-rendered overlay labels or legends.

#### Opaque overlay builders

Scripts construct branded/native-backed overlay values:

- `cvd.overlayTarget(name)` returns an `OverlayTargetBuilder`.
- `cvd.overlaySpec()` returns an `OverlaySpecBuilder`.
- `.build()` returns an opaque `OverlayTargetSpec` or `OverlaySpec` value.
- `page.overlay(spec)` accepts only the opaque `OverlaySpec` value, or a builder that can be finalized internally.

The Goja wrapper should therefore look conceptually like this:

```go
// In module installation:
_ = exports.Set("overlayTarget", func(name string) goja.Value {
    return wrapOverlayTargetBuilder(ctx, vm, name)
})
_ = exports.Set("overlaySpec", func() goja.Value {
    return wrapOverlaySpecBuilder(ctx, vm)
})

// In wrapPage:
_ = obj.Set("overlay", func(value goja.Value) goja.Value {
    spec, err := decodeOpaqueOverlaySpec(value)
    if err != nil {
        return rejectPromise(vm, err)
    }
    return wrapOverlayScreenshotBuilder(ctx, vm, state, spec)
})
```

Recommended V1: use Go struct-backed builders because they are easy to validate and hard for userland to accidentally forge.

#### Fluent builder API with typed Go-side style

```js
const spec = cvd.overlaySpec()
  .target(
    cvd.overlayTarget("NavBar")
      .selector("nav.top")
      .label("Navigation")
      .borderColor("#0096ff")
      .labelBackground("rgba(0, 150, 255, 0.92)")
      .labelColor("white"),
  )
  .target(
    cvd.overlayTarget("Hero")
      .selector(".hero-section")
      .style({
        borderColor: "#ff6347",
        contentBackground: "rgba(255, 99, 71, 0.12)",
        label: {
          background: "#ff6347",
          color: "white",
          position: "inside-start",
        },
      }),
  )
  .legend(true)
  .screenshot("fullPage")
  .style({
    label: {
      fontFamily: "Inter, system-ui, sans-serif",
      fontSize: 13,
      radius: 4,
      padding: [4, 7],
    },
    legend: {
      position: "bottom-right",
      background: "rgba(255, 255, 255, 0.92)",
      color: "#27221b",
    },
    targetDefaults: {
      borderWidth: 2,
      labelColor: "white",
    },
  })
  .build()

const result = await page.overlay(spec).screenshot("/tmp/annotated.png")
```

Notes:

- `.selector(cssSelector)` is required for every target.
- `.label(text)` is optional; default label text is the target name.
- `.style({...})` may appear on both the target builder and the spec builder.
- Common target style properties should also have fluent convenience methods such as `.borderColor(...)`, `.labelBackground(...)`, `.labelColor(...)`, and `.labelPosition(...)`.
- Target-level style overrides spec-level defaults.
- `.build()` validates required fields and returns an opaque value, not a plain object.
- `page.overlay(spec)` returns a screenshot builder with `.screenshot(path)`.

The result should remain plain JSON because it is output data, not input spec:

```js
{
  outputPath: "/tmp/annotated.png",
  colors: { NavBar: "#0096ff", Hero: "#ff6347" },
  targets: [
    { name: "NavBar", selector: "nav.top", label: "Navigation", color: "#0096ff" }
  ]
}
```

#### Typed style schema

The Go-side style object is a structured API, not CSS:

```ts
type OverlayStyle = {
  label?: {
    fontFamily?: string
    fontSize?: number
    radius?: number
    padding?: number | [number, number] | [number, number, number, number]
  }
  legend?: {
    position?: "top-left" | "top-right" | "bottom-left" | "bottom-right"
    background?: Color
    color?: Color
  }
  targetDefaults?: TargetOverlayStyle
}

type TargetOverlayStyle = {
  borderColor?: Color
  contentBackground?: Color
  paddingBackground?: Color
  marginBackground?: Color
  borderWidth?: number
  labelColor?: Color
  label?: {
    background?: Color
    color?: Color
    position?: "auto" | "above" | "below" | "inside-start" | "inside-end"
  }
}
```

Color values can still use familiar scalar strings (`#ff6347`, `rgb(...)`, `rgba(...)`, `white`, `transparent`), but they are parsed as individual values, not as CSS declarations.

#### Why no custom overlay CSS parser?

A constrained CSS-looking DSL creates false expectations. Users would reasonably expect CSS variables, browser selectors, media queries, shorthands, `calc()`, and DevTools visibility. Supporting only a subset would be surprising, while supporting real CSS would not map cleanly to CDP highlight configs and Go image compositing.

Therefore:

- use `page.css(...)` for real browser CSS injection;
- use `overlaySpec.style(...)` / `overlayTarget.style(...)` for Go-side rendering;
- parse only scalar values such as colors, enum positions, and padding arrays;
- do not parse CSS syntax in the overlay service.

#### Programmatic discovery without raw maps

Scripts can still start from userland maps, YAML, or DOM discovery; the conversion point should be a builder function, not passing raw objects into Go:

```js
function targetFromEntry([name, selector], style = {}) {
  return cvd.overlayTarget(name).selector(selector).style(style).build()
}

function specFromMap(map, styleByName = {}) {
  const builder = cvd.overlaySpec().legend(true)
  for (const entry of Object.entries(map)) {
    const [name] = entry
    builder.target(targetFromEntry(entry, styleByName[name] ?? {}))
  }
  return builder.build()
}

const spec = specFromMap(
  { NavBar: "nav.top", Hero: ".hero" },
  { Hero: { borderColor: "#ff6347", label: { background: "#ff6347" } } },
)
const result = await page.overlay(spec).screenshot("/tmp/out.png")
```

### Userland YAML / JSON Integration

There is intentionally no core visual-diff config schema for overlays. The old native YAML `run --config` pipeline has been removed. If a project wants to store overlay targets in YAML, JSON, Storybook metadata, or generated component registries, a JavaScript verb should load that file and convert it into opaque overlay specs with `cvd.overlayTarget(...)` and `cvd.overlaySpec(...)`.

For example, userland may read a map like `{ NavBar: "nav.top", Hero: ".hero" }`, convert each entry into an `OverlayTargetSpec` builder, attach typed style objects, and pass a finalized `OverlaySpec` to `page.overlay(spec)`. This keeps the Go core focused on browser/artifact primitives rather than project-specific schema design.

---

## JavaScript Data Model and Script Examples

### Opaque spec model

The primary JS API should use opaque builder-produced values rather than plain target objects.

```js
const spec = cvd.overlaySpec()
  .target(cvd.overlayTarget("NavBar").selector("nav.top"))
  .target(cvd.overlayTarget("Hero").selector(".hero-section"))
  .target(cvd.overlayTarget("Primary CTA").selector("button.primary"))
  .build()

await page.overlay(spec).screenshot("/tmp/landing-annotated.png")
```

For convenience, userland can write helpers that convert maps into specs while still keeping the native API typed:

```js
function overlaySpecFromMap(map, styleByName = {}) {
  const builder = cvd.overlaySpec()
  for (const [name, selector] of Object.entries(map)) {
    const target = cvd.overlayTarget(name).selector(selector)
    if (styleByName[name]) target.style(styleByName[name])
    builder.target(target)
  }
  return builder.build()
}
```

### Full-page annotated PNG script sketch

This example takes a map of human-readable names to selectors, opens a page, verifies selectors with `preflight`, builds an opaque overlay spec, and exports an annotated full-page PNG.

```js
// examples/scripts/annotate-full-page.js
const cvd = require("css-visual-diff")

function overlaySpecFromMap(map) {
  const spec = cvd.overlaySpec()
    .legend(true)
    .screenshot("fullPage")
    .style({
      label: { fontSize: 13, radius: 4, padding: [4, 7] },
      legend: { position: "bottom-right", background: "rgba(255, 255, 255, 0.92)" },
      targetDefaults: { borderWidth: 2, labelColor: "white" },
    })

  for (const [name, selector] of Object.entries(map)) {
    spec.target(cvd.overlayTarget(name).selector(selector))
  }
  return spec.build()
}

async function main() {
  const url = "http://localhost:3000/landing"
  const outPath = "/tmp/landing-annotated.png"

  const components = {
    NavBar: "nav.top",
    Hero: ".hero-section",
    "Feature Grid": "#features",
    "Primary CTA": "button.primary",
    Footer: "footer",
  }

  const browser = await cvd.browser()
  try {
    const page = await browser.page(url, {
      viewport: cvd.viewport(1440, 1600),
      waitMs: 500,
      name: "landing-page",
    })

    // Optional real browser CSS, separate from Go-side overlay style.
    await page.css(`html { scroll-behavior: auto !important; }`)

    const probeTargets = Object.entries(components).map(([name, selector]) => ({ name, selector }))
    const statuses = await page.preflight(probeTargets)
    const missing = statuses.filter((s) => !s.exists)
    if (missing.length > 0) {
      throw new cvd.SelectorError(
        `Missing overlay selectors: ${missing.map((s) => `${s.name}=${s.selector}`).join(", ")}`,
      )
    }

    const result = await page.overlay(overlaySpecFromMap(components)).screenshot(outPath)
    console.log(JSON.stringify({ ok: true, outputPath: result.outputPath, colors: result.colors }, null, 2))
  } finally {
    await browser.close()
  }
}

main()
```

### Loading YAML as userland data

YAML is still fine for project-owned specs; it is just not interpreted by the Go core. A verb can load YAML and map it to fluent builders and typed style objects:

```js
const fs = require("fs")
const yaml = require("yaml")
const cvd = require("css-visual-diff")

function overlaySpecFromYaml(path) {
  const specData = yaml.parse(fs.readFileSync(path, "utf8"))
  const builder = cvd.overlaySpec().legend(specData.legend ?? true)
  if (specData.style) builder.style(specData.style)

  for (const [name, entry] of Object.entries(specData.components ?? {})) {
    const selector = typeof entry === "string" ? entry : entry.selector
    const target = cvd.overlayTarget(name).selector(selector)
    if (typeof entry === "object" && entry.label) target.label(entry.label)
    if (typeof entry === "object" && entry.style) target.style(entry.style)
    builder.target(target)
  }
  return builder.build()
}
```

Example userland YAML:

```yaml
style:
  label:
    fontSize: 13
    radius: 4
  legend:
    position: bottom-right
    background: rgba(255, 255, 255, 0.92)
  targetDefaults:
    borderWidth: 2
    labelColor: white
components:
  NavBar:
    selector: nav.top
    label: Navigation
    style:
      borderColor: '#0096ff'
      label:
        background: '#0096ff'
  Hero:
    selector: .hero-section
    style:
      borderColor: '#ff6347'
      label:
        background: '#ff6347'
  Feature Grid: '#features'
  Primary CTA: button.primary
  Footer: footer
```

### JavaScript Example Cookbook

This section sketches the scripts we want to make possible once the overlay API exists. The examples are intentionally userland-oriented: they keep project meaning in JavaScript while relying on the Go core for browser, screenshot, selector, CSS extraction, and overlay primitives.

#### Example 1: annotated full-page PNG from a name-to-selector map

```js
// examples/scripts/overlay-annotated-png.js
const cvd = require("css-visual-diff")

const componentSelectors = {
  Header: "header.site-header",
  Navigation: "nav.primary-nav",
  Hero: "section.hero",
  "Hero CTA": "section.hero a.button-primary",
  "Feature Cards": "section.features",
  Footer: "footer.site-footer",
}

const perTargetStyle = {
  Header: { borderColor: "#0096ff", label: { background: "#0096ff" } },
  Hero: { borderColor: "#ff6347", label: { background: "#ff6347" } },
  "Feature Cards": { borderColor: "#32cd32", label: { background: "#32cd32" } },
}

function overlaySpecFromMap(map) {
  const builder = cvd.overlaySpec()
    .legend(true)
    .screenshot("fullPage")
    .style({ legend: { position: "bottom-right" }, label: { fontSize: 13 } })

  for (const [name, selector] of Object.entries(map)) {
    const target = cvd.overlayTarget(name).selector(selector)
    if (perTargetStyle[name]) target.style(perTargetStyle[name])
    builder.target(target)
  }
  return builder.build()
}

async function main() {
  const browser = await cvd.browser()
  try {
    const page = await browser.page("http://localhost:3000/", {
      name: "home",
      viewport: cvd.viewport(1440, 1800),
      waitMs: 500,
    })

    const result = await page.overlay(overlaySpecFromMap(componentSelectors))
      .screenshot("/tmp/cssvd/home.annotated.png")
    console.log(JSON.stringify(result, null, 2))
  } finally {
    await browser.close()
  }
}

main()
```

#### Example 2: extract a component-system inventory with individual PNGs

This script treats a rendered page as evidence for a component system. It extracts a list of known atoms, molecules, and organisms; writes one cropped PNG per component; captures selected CSS properties; and emits a `components.json` manifest that a later report page can consume.

```js
// examples/scripts/component-system-extract.js
const fs = require("fs")
const path = require("path")
const cvd = require("css-visual-diff")

const outDir = "/tmp/cssvd/component-system"
const components = [
  cvd.probe("Button / Primary").selector(".button-primary").props(["display", "font-size", "background-color"]).build(),
  cvd.probe("Button / Secondary").selector(".button-secondary").props(["display", "font-size", "background-color"]).build(),
  cvd.probe("Logo").selector(".site-logo").build(),
  cvd.probe("Nav Item").selector(".primary-nav li:first-child").build(),
  cvd.probe("Feature Card").selector(".feature-card:first-child").build(),
  cvd.probe("Hero").selector("section.hero").build(),
  cvd.probe("Feature Grid").selector("section.features").build(),
]

async function main() {
  fs.mkdirSync(outDir, { recursive: true })
  const browser = await cvd.browser()
  try {
    const page = await browser.page("http://localhost:3000/", { viewport: cvd.viewport(1440, 1800), waitMs: 500 })
    const inspected = await page.inspectAll(components, { outDir, artifacts: "screenshot-css-json" })
    fs.writeFileSync(path.join(outDir, "components.json"), JSON.stringify({ components: inspected.results }, null, 2))
    console.log(JSON.stringify({ outDir, count: inspected.results.length }, null, 2))
  } finally {
    await browser.close()
  }
}

main()
```

#### Example 3: component-system HTML gallery with individual and annotated PNGs

This script combines component extraction with overlays. It writes cropped PNGs for individual components, annotated PNGs for organisms and full screens, `components.json`, and a static `index.html` gallery.

```js
// examples/scripts/component-system-gallery.js
const fs = require("fs")
const path = require("path")
const cvd = require("css-visual-diff")

const outDir = "/tmp/cssvd/component-gallery"
const url = "http://localhost:3000/"
const system = {
  atoms: { Logo: ".site-logo", "Primary Button": ".button-primary", "Secondary Button": ".button-secondary" },
  molecules: { "Navigation Item": ".primary-nav li:first-child", "Feature Card": ".feature-card:first-child" },
  organisms: { Header: "header.site-header", Hero: "section.hero", "Feature Grid": "section.features", Footer: "footer.site-footer" },
  screens: { "Home Page": "body" },
}

function probesFromSystem(system) {
  return Object.entries(system).flatMap(([level, map]) =>
    Object.entries(map).map(([name, selector]) => cvd.probe(name).selector(selector).attribute("class").build()),
  )
}

function overlaySpecFromMap(map, style = {}) {
  const builder = cvd.overlaySpec().legend(true).screenshot("fullPage").style(style)
  for (const [name, selector] of Object.entries(map)) builder.target(cvd.overlayTarget(name).selector(selector))
  return builder.build()
}

function writeHtmlReport(outDir, model) {
  const esc = (s) => String(s).replace(/[&<>\"]/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;" }[ch]))
  const cards = model.components.map((component) => `<article class="card"><h3>${esc(component.name)}</h3><p><code>${esc(component.selector)}</code></p>${component.image ? `<img src="${esc(path.relative(outDir, component.image))}">` : ""}</article>`).join("\n")
  const annotated = model.annotated.map((item) => `<section class="annotated"><h2>${esc(item.name)}</h2><img src="${esc(path.relative(outDir, item.path))}"></section>`).join("\n")
  fs.writeFileSync(path.join(outDir, "index.html"), `<!doctype html><html><head><meta charset="utf-8"><title>Component System</title><style>body{font-family:system-ui;margin:32px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:18px}.card,.annotated{border:1px solid #ddd;padding:14px}img{max-width:100%}</style></head><body><h1>Component System</h1><div class="grid">${cards}</div>${annotated}</body></html>`)
}

async function main() {
  fs.mkdirSync(outDir, { recursive: true })
  const browser = await cvd.browser()
  try {
    const page = await browser.page(url, { viewport: cvd.viewport(1440, 1800), waitMs: 500 })
    const inspected = await page.inspectAll(probesFromSystem(system), { outDir: path.join(outDir, "components"), artifacts: "screenshot-css-json" })
    const fullScreenPath = path.join(outDir, "annotated", "home.organisms.annotated.png")
    fs.mkdirSync(path.dirname(fullScreenPath), { recursive: true })
    await page.overlay(overlaySpecFromMap(system.organisms, {
      legend: { position: "bottom-right" },
      targetDefaults: { labelColor: "white", borderWidth: 2 },
    })).screenshot(fullScreenPath)
    const model = { url, components: inspected.results, annotated: [{ name: "Home Page / Organism Map", path: fullScreenPath }] }
    fs.writeFileSync(path.join(outDir, "components.json"), JSON.stringify(model, null, 2))
    writeHtmlReport(outDir, model)
    console.log(JSON.stringify({ ok: true, html: path.join(outDir, "index.html") }, null, 2))
  } finally {
    await browser.close()
  }
}

main()
```

#### Example 4: annotated organisms with nested child labels

```js
// examples/scripts/organism-overlays.js
const path = require("path")
const fs = require("fs")
const cvd = require("css-visual-diff")
const outDir = "/tmp/cssvd/organism-overlays"
const organisms = {
  Hero: { root: "section.hero", parts: { Eyebrow: "section.hero .eyebrow", Headline: "section.hero h1", Copy: "section.hero .lede", CTA: "section.hero .button-primary", Media: "section.hero .hero-media" } },
  Header: { root: "header.site-header", parts: { Logo: "header.site-header .site-logo", Navigation: "header.site-header nav.primary-nav", Actions: "header.site-header .header-actions" } },
}

function organismSpec(name, organism) {
  const builder = cvd.overlaySpec()
    .legend(true)
    .screenshot("fullPage")
    .style({ legend: { position: "bottom-right" }, targetDefaults: { labelColor: "white" } })
    .target(cvd.overlayTarget(name).selector(organism.root).style({ borderColor: "#27221b", label: { background: "#27221b" } }))
  for (const [partName, selector] of Object.entries(organism.parts)) builder.target(cvd.overlayTarget(partName).selector(selector))
  return builder.build()
}

async function main() {
  fs.mkdirSync(outDir, { recursive: true })
  const browser = await cvd.browser()
  try {
    const page = await browser.page("http://localhost:3000/", { viewport: cvd.viewport(1440, 1800), waitMs: 500 })
    for (const [organismName, organism] of Object.entries(organisms)) {
      const slug = organismName.toLowerCase().replace(/[^a-z0-9]+/g, "-")
      const result = await page.overlay(organismSpec(organismName, organism)).screenshot(path.join(outDir, `${slug}.parts.annotated.png`))
      console.log(`${organismName}: ${result.outputPath}`)
    }
  } finally {
    await browser.close()
  }
}

main()
```

A useful V2/V3 enhancement is `.cropTo(selector)` or `.cropToTarget(name)` on the overlay spec, so an organism-level PNG contains only the organism bounds plus padding.

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
1. Define `OverlaySpec`, `OverlayTarget`, `OverlayResult`, and `OverlayScreenshot`.
2. Implement the pipeline: normalize typed styles → resolve → highlight → screenshot → decode → label → legend → encode → write.
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
1. Implement `cvd.overlayTarget(name)` and `cvd.overlaySpec()` fluent builders.
2. Ensure `.build()` returns opaque/branded `OverlayTargetSpec` and `OverlaySpec` values rather than plain JS objects.
3. Implement `page.overlay(spec)` so it only accepts an opaque `OverlaySpec` (or an unbuilt `OverlaySpecBuilder` that is finalized internally), never `[]map[string]any`.
4. Add typed style decoding/normalization for spec-level and target-level overlay styling; do not parse custom overlay CSS.
5. Follow the `promiseValue` + `runExclusive` pattern exactly.
6. Return a plain result object: `{ outputPath: string, colors: Record<string, string>, targets: [...] }`.

**Validation (manual script test):**
Create `examples/scripts/test-overlay.js`:

```js
const cvd = require("css-visual-diff")
async function main() {
  const browser = await cvd.browser()
  const page = await browser.page("https://example.com")
  const spec = cvd.overlaySpec()
    .target(cvd.overlayTarget("Heading").selector("h1").style({ borderColor: "#0096ff", label: { background: "#0096ff" } }))
    .legend(true)
    .build()
  const result = await page.overlay(spec).screenshot("/tmp/test-overlay.png")
  console.log(JSON.stringify(result, null, 2))
  await browser.close()
}
main()
```

Run:
```bash
GOWORK=off go run ./cmd/css-visual-diff verbs script run examples/scripts/test-overlay.js
```

### Phase 4: JavaScript Verb Examples and Documentation

**Files to create/modify:**
- `examples/scripts/annotate-full-page.js` or an example verb under `examples/verbs/`
- `internal/cssvisualdiff/doc/topics/javascript-api.md`
- `README.md`

**Tasks:**
1. Write a self-contained JavaScript example that takes a name-to-selector map, preflights selectors, and exports a full-page annotated PNG.
2. Write a component-system extraction example that writes individual component screenshots, CSS/JSON metadata, and `components.json`.
3. Write a static HTML gallery example that lists extracted atoms/molecules/organisms/screens and links to annotated organism/full-screen PNGs.
4. Write a YAML/JSON loader example in userland and convert it to fluent overlay builders.
5. Document that `.css-visual-diff.yml` only discovers verb repositories and is not where overlay targets belong.
6. Update the JavaScript API docs with `page.css(...)` for real browser CSS, `cvd.overlayTarget`, `cvd.overlaySpec`, typed `.style(...)` overlay styling, and `page.overlay(spec).screenshot(path)`.

**Validation:**
```bash
GOWORK=off go run ./cmd/css-visual-diff verbs --repository examples/verbs --help
# Then run the concrete example verb or script once it exists.
```

---

## Testing and Validation Strategy

### Unit tests

- **Driver tests:** mock or launch a real Chrome instance to verify that `HighlightNode` + `FullScreenshot` produces an image different from a non-highlighted screenshot. Use image hash comparison (perceptual or simple average color shift).
- **Service tests:** use a static HTML fixture served via `httptest.Server`. Capture an overlay screenshot and assert that the output file exists and is larger than the non-annotated version (labels add pixels).
- **JS builder/decoder tests:** in `jsapi` tests, add valid and invalid cases for opaque `OverlayTargetSpec` / `OverlaySpec` values, missing selectors, invalid typed style values, color parsing failures, and attempts to pass raw arrays/maps to `page.overlay`.

### Integration tests

- Run the JavaScript verb/script example manually and visually inspect the PNG.
- If a userland YAML example is added, verify the JS verb loads it and produces the same target array as the inline-map example.

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

1. **Should the legend be optional?** Yes — expose `legend: boolean` in the JS options object, defaulting to `true`.
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
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/internal/cssvisualdiff/verbcli/bootstrap.go` | App config discovery for JS verb repositories. Important because `.css-visual-diff.yml` is only for finding scripts, not for overlay target schema. |
| `/home/manuel/code/wesen/corporate-headquarters/css-visual-diff/cmd/css-visual-diff/main.go` | Root CLI command registration. Confirms the live entry points are direct commands plus `verbs`, not the removed `run --config` pipeline. |
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
func OverlayScreenshot(page *driver.Page, spec OverlaySpec, outPath string) (*OverlayResult, error) {
    // 1. Normalize typed Go-side overlay styles.
    spec, err := NormalizeOverlaySpec(spec)
    if err != nil {
        return nil, err
    }

    // 2. Resolve selectors, merge target styles, and assign colors.
    type annotatedTarget struct {
        OverlayTarget
        nodeID cdp.NodeID
        color  color.RGBA
        style  TargetOverlayStyle
    }
    var annotated []annotatedTarget
    for i, t := range spec.Targets {
        nodeID, err := page.ResolveNodeID(t.Selector)
        if err != nil {
            return nil, fmt.Errorf("resolve %q: %w", t.Selector, err)
        }
        style := ResolveTargetStyle(spec.Style, t.Style, t, defaultPalette[i % len(defaultPalette)])
        if style.BorderColor == nil {
            style.BorderColor = ptr(defaultPalette[i % len(defaultPalette)])
        }
        annotated = append(annotated, annotatedTarget{
            OverlayTarget: t,
            nodeID:        nodeID,
            color:         *style.BorderColor,
            style:         style,
        })
    }

    // 3. Apply CDP highlights.
    for _, at := range annotated {
        cfg := highlightConfigFromStyle(at.style)
        if err := page.HighlightNode(at.Selector, cfg); err != nil {
            return nil, err
        }
    }

    // 4. Capture screenshot
    var buf []byte
    if err := chromedp.Run(page.Context(), chromedp.FullScreenshot(&buf, 90)); err != nil {
        _ = page.HideHighlight()
        return nil, err
    }

    // 5. Decode PNG
    img, err := png.Decode(bytes.NewReader(buf))
    if err != nil {
        _ = page.HideHighlight()
        return nil, err
    }
    rgba := image.NewRGBA(img.Bounds())
    draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

    // 6. Query bounds for label placement
    for i := range annotated {
        bounds, err := service.LocatorBounds(page, service.LocatorSpec{Selector: annotated[i].Selector})
        if err != nil {
            continue // skip label if element disappeared
        }
        annotated[i].bounds = bounds
    }

    // 7. Draw labels
    for _, at := range annotated {
        if at.bounds == nil {
            continue
        }
        label := at.Label
        if label == "" {
            label = at.Name
        }
        drawLabel(rgba, *at.bounds, label, at.style)
    }

    // 8. Draw legend
    if spec.Legend {
        drawLegend(rgba, annotated, spec.Style.Legend)
    }

    // 9. Hide highlights
    _ = page.HideHighlight()

    // 10. Encode and write
    f, err := os.Create(outPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    if err := png.Encode(f, rgba); err != nil {
        return nil, err
    }

    // 11. Build result
    colors := make(map[string]string)
    for _, at := range annotated {
        colors[at.Name] = rgbaToHex(at.color)
    }
    return &OverlayResult{
        Image:      rgba,
        OutputPath: outPath,
        Targets:    spec.Targets,
        Colors:     colors,
    }, nil
}
```

---

*Document version: 2026-04-28*
*Ticket: overlay-screenshot-labels*
