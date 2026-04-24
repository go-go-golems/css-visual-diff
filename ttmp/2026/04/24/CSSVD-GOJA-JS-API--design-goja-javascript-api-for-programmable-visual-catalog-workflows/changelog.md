# Changelog

## 2026-04-24

- Initial workspace created


## 2026-04-24 — Create Goja JavaScript API design guide

### Added
- Added `design/01-goja-javascript-api-analysis-design-and-implementation-guide.md`.

### Contents
- Explains the motivation for a Goja scripting API after the Pyxis catalog work.
- Maps current css-visual-diff architecture: CLI, config schema, driver, prepare hooks, inspect/artifact modes.
- Proposes Browser, Page, Probe, Preflight, Inspect, and Catalog JavaScript APIs.
- Includes TypeScript-style API references, pseudocode, diagrams, implementation phases, tests, documentation deliverables, risks, and acceptance criteria.
- Recommends keeping core logic in Go services and exposing JS through thin Goja adapters.

### Publishing
- Uploaded to reMarkable under `/ai/2026/04/24/CSSVD-GOJA-JS-API/`.
