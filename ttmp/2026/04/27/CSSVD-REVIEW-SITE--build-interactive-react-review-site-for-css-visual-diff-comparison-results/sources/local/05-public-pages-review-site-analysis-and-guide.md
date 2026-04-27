# Public Pages Visual Review Site: Analysis, User Guide, and Implementation Guide

## 1. Purpose

This document explains the small HTML review site we built during the Pyxis public-page visual tuning work. The site is not the final product. It is a working prototype of an operator feedback loop: run visual comparisons, bundle the resulting screenshots and metadata into a self-contained HTML page, let a human reviewer inspect each section, write notes, and copy those notes with enough technical context for a developer or agent to act on them.

The important idea is simple: visual-diff tools produce evidence, but evidence is only useful if a person can review it in a humane way. A raw folder of PNGs and JSON files is technically complete, but it is hard to scan, hard to annotate, and hard to hand off. The review site turns those artifacts into a queue of review cards. Each card shows the prototype, the React implementation, and the difference image side by side. It also exposes the `compare.json` metadata and a compact YAML summary that can be copied into an issue, a prompt, or a code review comment.

A future developer building this “for real” should treat the current implementation as a clear reference implementation, not as a finished architecture. The current version is a static HTML generator script. A production version should preserve the workflow and data model, but replace the hand-written single-file HTML with a maintainable application that can store feedback, track review state, compare multiple runs, and integrate with the existing `css-visual-diff` userland tools.

## 2. Quick glossary

- **Prototype** means the static reference page served from `prototype-design/standalone/public/*.html` on port `7070`.
- **React implementation** means the Storybook-rendered public page from `web/packages/pyxis-user-site` on port `6007`.
- **Visual target** means one page or page section declared in a visual-diff YAML spec.
- **Run** means one execution of `css-visual-diff` against a spec, producing PNGs, `compare.json`, and a summary JSON.
- **Review bundle** means a self-contained directory containing `index.html`, copied image artifacts, and copied `compare.json` files.
- **Review card** means one row in the generated HTML page, usually one page/section pair such as `about/content`.
- **Operator** means the human doing visual review and feedback, usually Manuel or another product/design reviewer.
- **Developer** means the person who will act on the notes and tune components.

## 3. The workflow in one picture

```text
┌─────────────────────────────┐
│ Prototype static HTML        │
│ localhost:7070               │
│ /standalone/public/*.html    │
└──────────────┬──────────────┘
               │
               │ compared by css-visual-diff
               ▼
┌─────────────────────────────┐       ┌─────────────────────────────┐
│ Visual spec YAML             │       │ React Storybook              │
│ public-pages.desktop.yml     │──────▶│ localhost:6007               │
│ selectors + story IDs        │       │ iframe.html?id=...           │
└──────────────┬──────────────┘       └─────────────────────────────┘
               │
               ▼
┌─────────────────────────────┐
│ css-visual-diff run output   │
│ /tmp/.../page/artifacts/...  │
│ diff_only.png                │
│ left_region.png              │
│ right_region.png             │
│ compare.json                 │
│ summary.json                 │
└──────────────┬──────────────┘
               │
               │ read by ticket script
               ▼
┌─────────────────────────────┐
│ Review bundle generator      │
│ 05-build-public-pages-review │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ Static HTML review site      │
│ /tmp/.../index.html          │
│ served on localhost:8097     │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ Human feedback               │
│ textarea notes + YAML copy   │
│ issue/prompt/code review     │
└─────────────────────────────┘
```

The current site lives in the last two boxes. It does not run comparisons itself. It consumes a completed `css-visual-diff` summary and the artifact paths referenced by that summary. This separation is deliberate. It keeps the review site focused on human review rather than screenshot capture, browser orchestration, or diff calculation.

## 4. Why this site exists

Visual regression work has two different jobs that are easy to confuse.

The first job is measurement. Measurement asks: how different are two renderings? What changed percentage did the image diff compute? What are the element bounds? Which computed styles differ? That job belongs to `css-visual-diff`, Playwright, and the visual spec files.

The second job is judgment. Judgment asks: does this difference matter? Is the React version better, worse, or acceptable? Should we tune a shared component, a page wrapper, or the spec selector? That job belongs to a human reviewer and the developer using their feedback.

The review site bridges these jobs. It keeps the mechanical evidence close to the human judgment. A reviewer can inspect `diff_only.png` first, then compare the React and prototype images, then type a note such as “About header is 6px too low; leave Shows as-is.” The clipboard button copies that note together with YAML context showing selectors, bounds, pixel percentages, and style differences. That makes the feedback actionable.

## 5. Current implementation inventory

### 5.1 Main generator script

File:

```text
ttmp/2026/04/27/PYXIS-APP-VISUAL-TUNING--pyxis-app-visual-tuning-topbar-dashboard-new-pages/scripts/05-build-public-pages-review.py
```

This Python script builds the review HTML page. It accepts two arguments:

```text
--summary-json  Path to a css-visual-diff suite summary JSON.
--output-dir    Directory where the self-contained review bundle should be written.
```

Example:

```bash
python3 ttmp/2026/04/27/PYXIS-APP-VISUAL-TUNING--pyxis-app-visual-tuning-topbar-dashboard-new-pages/scripts/05-build-public-pages-review.py \
  --summary-json /tmp/pyxis-public-pages-after-about-spacing.json \
  --output-dir /tmp/pyxis-public-pages-visual-review-about-spacing-20260427-172932
```

The script prints the generated `index.html` path. The latest served review bundle from this ticket was:

```text
/tmp/pyxis-public-pages-visual-review-about-spacing-20260427-172932/index.html
http://127.0.0.1:8097/index.html
```

### 5.2 Visual spec consumed upstream

File:

```text
prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml
```

This spec declares the public pages and sections that get compared. It is the source of truth for:

- The prototype URL path, such as `/standalone/public/about.html`.
- The Storybook story ID, such as `public-site-pages-about--desktop`.
- The viewport, wait time, threshold, and inspect mode.
- The page/section selectors used to crop each comparison.
- The policy bands used to classify rows as `accepted`, `review`, `tune-required`, or `major-mismatch`.

Important defaults:

```yaml
defaults:
  prototypeBase: http://localhost:7070
  storybookBase: http://localhost:6007
  viewport:
    width: 920
    height: 1460
  waitMs: 1000
  threshold: 30
  inspect: rich
  variant: desktop
```

Important policy bands:

```yaml
policy:
  bands:
    - name: accepted
      maxChangedPercent: 0.5
    - name: review
      maxChangedPercent: 10
    - name: tune-required
      maxChangedPercent: 30
    - name: major-mismatch
      maxChangedPercent: 100
```

### 5.3 Userland command layer

File:

```text
prototype-design/visual-diff/userland/verbs/pyxis-pages.js
```

This file defines the `pyxis pages ...` verbs used to run comparisons and inspect sections. The review site does not import this file directly, but the summary JSON that the review site consumes is usually produced through these commands.

Useful commands include:

```bash
# Run the whole public-pages spec and print a summary.
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --outDir /tmp/pyxis-public-pages-run \
  --summary \
  --output json \
  > /tmp/pyxis-public-pages-run.json

# Run only one page for faster iteration.
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page about \
  --outDir /tmp/pyxis-public-about-run \
  --summary \
  --output json \
  > /tmp/pyxis-public-about-run.json
```

### 5.4 Storybook page stories

Directory:

```text
web/packages/pyxis-user-site/src/pages/
```

The public pages were reorganized into page folders so they are easy to compare and easy to reason about:

```text
AboutPage/
  Page.tsx
  Page.css
  Page.stories.tsx
  index.ts
ArchivePage/
  Page.tsx
  Page.css
  Page.stories.tsx
  index.ts
BookPage/
  Page.tsx
  Page.css
  Page.stories.tsx
  index.ts
ShowsPage/
  Page.tsx
  Page.css
  Page.stories.tsx
  index.ts
ShowDetailPage/
  Page.tsx
  Page.css
  Page.stories.tsx
  index.ts
storybook.tsx
```

The review workflow depends on these stories because the visual spec compares prototype pages against Storybook iframe URLs. A typical story URL looks like this:

```text
http://localhost:6007/iframe.html?id=public-site-pages-about--desktop&viewMode=story
```

### 5.5 Review artifacts

A visual-diff run produces a nested artifact directory. A typical section contains:

```text
diff_only.png
left_region.png
right_region.png
diff_comparison.png
compare.json
```

The names matter:

- `diff_only.png` is the first image to inspect. It highlights changed pixels without the distraction of a wide triptych.
- `right_region.png` is the React implementation. It answers “what did we render?”
- `left_region.png` is the prototype. It answers “what were we matching?”
- `diff_comparison.png` is the side-by-side triptych. It is useful for broad context but too wide for first-pass review.
- `compare.json` is the structured evidence: selectors, bounds, pixel counts, changed styles, changed attributes, and source URLs.

The review site copies these files into its own output directory so the bundle is self-contained. This is important because `/tmp` visual-diff run directories are often cleaned up or overwritten.

## 6. The current review bundle structure

A generated review bundle looks like this:

```text
/tmp/pyxis-public-pages-visual-review-about-spacing-20260427-172932/
  index.html
  artifacts/
    about/
      content/
        diff_only.png
        left_region.png
        right_region.png
        compare.json
      page/
        diff_only.png
        left_region.png
        right_region.png
        compare.json
    shows/
      page/
        ...
      content/
        ...
      shows-list/
        ...
      header/
        ...
      mailing-list/
        ...
```

The static `index.html` uses relative links into `artifacts/`. That means the bundle can be served from any simple HTTP server:

```bash
cd /tmp/pyxis-public-pages-visual-review-about-spacing-20260427-172932
python3 -m http.server 8097 --bind 127.0.0.1
```

Then open:

```text
http://127.0.0.1:8097/index.html
```

The current workflow stores helper paths in `/tmp`:

```text
/tmp/pyxis-public-pages-review-http.pid
/tmp/pyxis-public-pages-review-http.log
/tmp/pyxis-public-pages-latest-review-path.txt
/tmp/pyxis-public-pages-latest-review-dir.txt
```

These files are operational conveniences. A production app should replace them with explicit run records.

## 7. How the generator works

The generator has five responsibilities.

### 7.1 Load rows from summary JSON

Function:

```python
def load_rows(path: Path) -> list[dict]:
    payload = json.loads(path.read_text())
    if isinstance(payload, list) and payload and "rows" in payload[0]:
        return payload[0]["rows"]
    if isinstance(payload, dict) and "rows" in payload:
        return payload["rows"]
    raise SystemExit(f"Unsupported summary JSON shape: {path}")
```

The `css-visual-diff --summary --output json` command may emit a list whose first item has `rows`, or a direct object with `rows`. The loader accepts both shapes because the command output changed during iteration.

A row is the review site’s main data record. Conceptually it looks like this:

```ts
type ReviewRow = {
  page: string;              // e.g. "about"
  section: string;           // e.g. "content"
  classification: string;    // e.g. "review"
  changedPercent: number;    // e.g. 7.6969
  diffOnlyPath: string;      // path to diff_only.png
};
```

The row does not directly store every artifact path. The generator derives sibling paths from `diffOnlyPath`.

### 7.2 Resolve artifact paths

Function:

```python
def artifact(row: dict, name: str) -> str:
    if name == "diff":
        return row.get("diffOnlyPath") or ""
    base = Path(row.get("diffOnlyPath") or "")
    artifacts = base.parent
    mapping = {
        "left": artifacts / "left_region.png",
        "right": artifacts / "right_region.png",
        "comparison": artifacts / "diff_comparison.png",
        "compare": artifacts / "compare.json",
    }
    return str(mapping[name])
```

This function encodes a convention: all section artifacts live beside `diff_only.png`. That convention comes from `css-visual-diff`. A production version should either keep this convention or replace it with a formal artifact manifest emitted by the comparison runner.

### 7.3 Copy artifacts into the bundle

Function:

```python
def bundle_artifact(path: str, output_dir: Path, page: str, section: str, name: str) -> str:
    source = Path(path)
    target_dir = output_dir / "artifacts" / page / section
    target = target_dir / f"{name}{source.suffix}"
    shutil.copy2(source, target)
    return target.relative_to(output_dir).as_posix()
```

This is the step that makes the review site portable. Early versions linked directly to `/tmp/...` artifact files. That made the page fragile. The current script copies every image and `compare.json` file into the bundle.

### 7.4 Convert compare metadata into YAML

Functions:

```python
def compare_yaml_summary(compare_path: str, row: dict) -> tuple[str, str]:
    raw = path.read_text()
    data = json.loads(raw)
    summary = {
        "page": row.get("page"),
        "section": row.get("section"),
        "classification": row.get("classification"),
        "changedPercent": row.get("changedPercent"),
        "bounds": data.get("bounds"),
        "pixel": data.get("pixel"),
        "left": data.get("left"),
        "right": data.get("right"),
        "changedStyles": [style for style in data.get("styles", []) if style.get("changed")][:20],
        "changedAttributes": [attr for attr in data.get("attributes", []) if attr.get("changed")][:20],
    }
    return raw, to_yaml(summary)
```

The YAML summary is designed for clipboard feedback. It contains enough context for a developer or coding agent to understand the comparison without opening the raw JSON.

The most useful fields are:

- `page` and `section`, which identify the card.
- `classification` and `changedPercent`, which describe severity.
- `bounds`, which often reveals spacing/layout problems.
- `pixel`, which records normalized image dimensions and changed pixel counts.
- `left` and `right`, which record selectors and URLs.
- `changedStyles`, which records computed CSS differences.
- `changedAttributes`, which records selector/attribute differences.

### 7.5 Render one HTML card per row

Each card contains:

- A header with `page / section`.
- A pill with classification and percent changed.
- Links to `left/prototype`, `right/react`, `diff_only`, and `compare.json`.
- A three-column image grid.
- A collapsible metadata panel.
- A notes textarea.
- A clipboard button.

Simplified HTML shape:

```html
<section class="card review" data-page="about" data-section="content">
  <header>
    <div><strong>about</strong> / content</div>
    <div class="pill">review · 7.697%</div>
  </header>

  <div class="links">
    <a href="artifacts/about/content/left_region.png">left/prototype</a>
    <a href="artifacts/about/content/right_region.png">right/react</a>
    <a href="artifacts/about/content/diff_only.png">diff_only</a>
    <a href="artifacts/about/content/compare.json">compare.json</a>
  </div>

  <div class="grid">
    <figure><figcaption>Prototype</figcaption><img src="..."></figure>
    <figure><figcaption>React</figcaption><img src="..."></figure>
    <figure><figcaption>Diff only</figcaption><img src="..."></figure>
  </div>

  <details class="compare-details">
    <summary>compare.json + YAML prompt context</summary>
    ...
  </details>

  <textarea placeholder="Reviewer notes for about/content"></textarea>
  <button type="button" onclick="copyCard(this)">To clipboard with YAML</button>
</section>
```

## 8. The current browser behavior

The page has intentionally small JavaScript. It does not use React, a database, or a build step. It defines three global functions:

```js
filterCards(cls)
copyCard(button)
copyAll()
```

### 8.1 Filtering

The toolbar buttons show all cards, only `tune-required`, or only `review` cards.

Pseudocode:

```text
function filterCards(classification):
    for each card in document.querySelectorAll('.card'):
        if classification is empty:
            show card
        else if card has class classification:
            show card
        else:
            hide card
```

### 8.2 Copying one card

The `copyCard` button copies the visible review note and YAML context for that card.

Pseudocode:

```text
function cardText(card):
    return join with newlines:
        page/section
        classification pill text
        "Notes: " + textarea value
        blank line
        "compare_yaml:"
        YAML metadata text

function copyCard(button):
    card = closest parent card
    navigator.clipboard.writeText(cardText(card))
```

Example clipboard output:

```text
about/content
review · 7.697%
Notes: Header and lead now look close enough. No further tuning needed.

compare_yaml:
page: about
section: content
classification: review
changedPercent: 7.696915186772811
bounds:
  changed: true
  left:
    height: 1088.71875
    width: 920
    x: 0
    y: 61
  right:
    height: 1136.46875
    width: 920
    x: 0
    y: 61
...
```

### 8.3 Copying all cards

`copyAll` concatenates all card outputs with separators. This is useful when Manuel reviews several sections and wants to pass the whole batch to a coding assistant or developer.

Pseudocode:

```text
function copyAll():
    texts = []
    for each card:
        texts.append(cardText(card))
    clipboard.write(join texts with "\n\n---\n\n")
```

## 9. How to use the current site

### 9.1 Start the required servers

The review workflow assumes the prototype server and Storybook server are already running.

```text
Prototype static server:         http://localhost:7070
pyxis-user-site Storybook:       http://localhost:6007
Review site static server:       http://localhost:8097
```

If Storybook produces strange missing-CSS or missing-export behavior, restart it or touch the relevant CSS/TSX file before trusting screenshots. This happened during the tuning work and can produce misleading screenshots.

### 9.2 Run a visual comparison

For a full sweep:

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis

rm -rf /tmp/pyxis-public-pages-run
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --outDir /tmp/pyxis-public-pages-run \
  --summary \
  --output json \
  > /tmp/pyxis-public-pages-run.json
```

For one page:

```bash
rm -rf /tmp/pyxis-public-about-run
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --page about \
  --outDir /tmp/pyxis-public-about-run \
  --summary \
  --output json \
  > /tmp/pyxis-public-about-run.json
```

### 9.3 Generate the review bundle

```bash
OUT=/tmp/pyxis-public-pages-visual-review-$(date +%Y%m%d-%H%M%S)
python3 ttmp/2026/04/27/PYXIS-APP-VISUAL-TUNING--pyxis-app-visual-tuning-topbar-dashboard-new-pages/scripts/05-build-public-pages-review.py \
  --summary-json /tmp/pyxis-public-pages-run.json \
  --output-dir "$OUT"

echo "$OUT"
```

### 9.4 Serve the review bundle

```bash
PORT=8097
cd "$OUT"
python3 -m http.server "$PORT" --bind 127.0.0.1
```

Open:

```text
http://127.0.0.1:8097/index.html
```

### 9.5 Review cards in the correct order

Use this inspection order:

1. Inspect `diff_only.png` first. It shows where the meaningful visual changes are.
2. Inspect `right/react` next. This shows what the implementation currently renders.
3. Inspect `left/prototype` after that. This shows what the implementation is trying to match.
4. Open `compare.json + YAML prompt context` when the visual difference looks like spacing, bounds, typography, or selector mismatch.
5. Write a note in the textarea.
6. Click `To clipboard with YAML`.
7. Paste the result into a ticket, PR review, chat message, or implementation prompt.

### 9.6 What good feedback looks like

Good feedback tells the developer what to do and where to look. It does not merely say “wrong.”

Better:

```text
about/content
review · 7.697%
Notes: Accept. Header title/lead are close enough after the wrapper padding fix. Do not tune further unless page copy changes.
```

Also good:

```text
book/content
review · 5.982%
Notes: Form spacing is acceptable. The remaining diff is mostly font antialiasing and expected form-control rendering.
```

Less useful:

```text
shows/page
Notes: bad
```

If a section is visually acceptable despite a high number, say so explicitly. During this ticket, Shows remained numerically `tune-required`, but the user accepted it visually. That human decision is part of the record and should be preserved.

## 10. Data model for a production version

The current static site has an implicit data model. A real app should make it explicit.

### 10.1 Core entities

```ts
type VisualRun = {
  id: string;
  name: string;
  createdAt: string;
  createdBy?: string;
  gitSha?: string;
  branch?: string;
  specPath: string;
  summaryJsonPath?: string;
  prototypeBase: string;
  storybookBase: string;
  viewport: { width: number; height: number };
  threshold: number;
};

type ReviewTarget = {
  id: string;
  runId: string;
  page: string;
  section: string;
  variant: string;
  prototypeUrl: string;
  reactUrl: string;
  prototypeSelector: string;
  reactSelector: string;
  priority?: string;
};

type ComparisonResult = {
  id: string;
  runId: string;
  targetId: string;
  classification: 'accepted' | 'review' | 'tune-required' | 'major-mismatch';
  changedPercent: number;
  changedPixels: number;
  totalPixels: number;
  normalizedWidth: number;
  normalizedHeight: number;
  bounds: BoundsComparison;
  changedStyles: StyleChange[];
  changedAttributes: AttributeChange[];
  artifacts: ArtifactSet;
};

type ArtifactSet = {
  diffOnlyUrl: string;
  leftRegionUrl: string;
  rightRegionUrl: string;
  diffComparisonUrl?: string;
  compareJsonUrl: string;
};

type ReviewNote = {
  id: string;
  resultId: string;
  author: string;
  createdAt: string;
  status: 'unreviewed' | 'accepted' | 'needs-work' | 'fixed' | 'wont-fix';
  note: string;
  copiedYaml?: string;
};
```

### 10.2 Why these entities matter

A run groups a set of comparisons. Without a run entity, it is hard to answer “what did the reviewer see?” two days later. A target records the selector-level intent. Without a target entity, selector bugs are hard to diagnose. A comparison result records the measured outcome. Without that, feedback is only subjective. A review note records human judgment. Without that, the team keeps revisiting already-accepted differences.

### 10.3 Suggested storage choices

A full app could use any of these storage approaches:

- **Filesystem-only JSON** is simplest. Store each run as a directory with `run.json`, `results.json`, `notes.json`, and copied artifacts. This is close to the current script and easy to review in Git.
- **SQLite** is better if multiple people will review and filter many runs. Store metadata and notes in tables, keep image artifacts on disk or object storage.
- **Postgres + object storage** is best if the tool becomes a team service with authentication, remote artifact upload, and long-term history.

For an intern building the first real version, SQLite plus filesystem artifacts is the best compromise.

Suggested filesystem layout:

```text
visual-review-data/
  runs/
    2026-04-27T17-29-32Z-about-spacing/
      run.json
      summary.json
      results.json
      notes.json
      artifacts/
        about/content/diff_only.png
        about/content/left_region.png
        about/content/right_region.png
        about/content/compare.json
```

## 11. Proposed production architecture

A production version should split responsibilities into three layers.

```text
┌─────────────────────────────────────────┐
│ Web UI                                  │
│ React app for review cards, filters,    │
│ notes, run comparison, keyboard nav     │
└───────────────────┬─────────────────────┘
                    │ HTTP/JSON
                    ▼
┌─────────────────────────────────────────┐
│ Review API                              │
│ Runs, results, artifacts, notes, status │
└───────────────────┬─────────────────────┘
                    │ filesystem / SQLite
                    ▼
┌─────────────────────────────────────────┐
│ Artifact store                          │
│ copied PNGs, compare.json, summaries    │
└─────────────────────────────────────────┘
```

The screenshot capture and diff computation can remain outside the app at first. The app can begin by importing a completed `css-visual-diff` summary, exactly like the current script does. Later, the app can add a “Run comparison” button that shells out to `css-visual-diff` or calls a worker service.

### 11.1 Frontend responsibilities

The frontend should provide:

- A run picker.
- Summary counts by classification.
- Filters by page, section, status, classification, and reviewer.
- A card view like the current static HTML.
- A larger side-by-side image viewer.
- A diff-first review mode.
- Persistent notes.
- Status changes such as `accepted`, `needs-work`, `fixed`, and `wont-fix`.
- Clipboard export with YAML or Markdown.
- Keyboard shortcuts for fast review.

### 11.2 Backend responsibilities

The backend should provide:

- Import a summary JSON and copy artifacts into a stable run directory.
- List runs.
- List results for a run.
- Serve artifacts.
- Save and retrieve review notes.
- Save and retrieve status decisions.
- Export selected notes as Markdown/YAML.
- Optionally execute `css-visual-diff` jobs.

### 11.3 Worker responsibilities, optional

A worker is optional for version one. If added, it should:

- Start comparison runs.
- Stream progress.
- Capture logs.
- Detect missing servers.
- Mark runs as failed, completed, or partially completed.

## 12. Suggested API design

A minimal API could look like this.

### 12.1 Import a run

```http
POST /api/runs/import
Content-Type: application/json

{
  "name": "public pages after about spacing",
  "summaryJsonPath": "/tmp/pyxis-public-pages-after-about-spacing.json",
  "copyArtifacts": true,
  "gitSha": "46e138d"
}
```

Response:

```json
{
  "runId": "run_20260427_172932_about_spacing",
  "resultCount": 13,
  "artifactRoot": "visual-review-data/runs/run_20260427_172932_about_spacing/artifacts"
}
```

### 12.2 List runs

```http
GET /api/runs
```

Response:

```json
[
  {
    "id": "run_20260427_172932_about_spacing",
    "name": "public pages after about spacing",
    "createdAt": "2026-04-27T17:29:32Z",
    "gitSha": "46e138d",
    "counts": {
      "accepted": 1,
      "review": 9,
      "tune-required": 3
    }
  }
]
```

### 12.3 List results for a run

```http
GET /api/runs/{runId}/results?classification=review&page=about
```

Response:

```json
[
  {
    "id": "result_about_content",
    "page": "about",
    "section": "content",
    "classification": "review",
    "changedPercent": 7.6969,
    "status": "accepted",
    "artifacts": {
      "diffOnlyUrl": "/api/artifacts/run_.../about/content/diff_only.png",
      "leftRegionUrl": "/api/artifacts/run_.../about/content/left_region.png",
      "rightRegionUrl": "/api/artifacts/run_.../about/content/right_region.png",
      "compareJsonUrl": "/api/artifacts/run_.../about/content/compare.json"
    }
  }
]
```

### 12.4 Save a review note

```http
PUT /api/results/{resultId}/review
Content-Type: application/json

{
  "status": "accepted",
  "note": "Header and lead are close enough after wrapper padding fix. Do not tune further.",
  "author": "manuel"
}
```

Response:

```json
{
  "ok": true,
  "review": {
    "status": "accepted",
    "note": "Header and lead are close enough after wrapper padding fix. Do not tune further.",
    "updatedAt": "2026-04-27T17:35:00Z"
  }
}
```

### 12.5 Export notes

```http
GET /api/runs/{runId}/export?format=markdown&includeYaml=true
```

Response:

```markdown
# Visual Review Notes: public pages after about spacing

## about/content — accepted — 7.697%

Header and lead are close enough after wrapper padding fix. Do not tune further.

```yaml
page: about
section: content
classification: review
changedPercent: 7.6969
...
```
```

## 13. UI design guide for the real app

### 13.1 Preserve the card mental model

The card model works because it matches the operator’s task. One card equals one decision. The production app should keep this shape even if it improves the styling.

Each card should answer five questions quickly:

1. What am I reviewing? Page and section.
2. How severe is it? Classification and changed percent.
3. Where is the visual difference? Diff-only image.
4. What does React look like? Right image.
5. What should happen next? Reviewer note and status.

### 13.2 Use diff-first image ordering

The current grid shows Prototype, React, Diff. The written workflow says inspect Diff first. The production app should consider making Diff the first or largest panel.

Recommended layout:

```text
┌────────────────────────────────────────────┐
│ page / section        review · 7.697%      │
├────────────────────────────────────────────┤
│ Diff only                                  │
│ [large image]                              │
├──────────────────────┬─────────────────────┤
│ Prototype             │ React               │
│ [left image]          │ [right image]       │
├────────────────────────────────────────────┤
│ Notes, status, copy/export                 │
└────────────────────────────────────────────┘
```

For wide screens, the app can offer multiple layouts:

- **Diff-first** for review.
- **Side-by-side** for detailed comparison.
- **Triptych** for broad context.
- **Metadata** for debugging selector/style issues.

### 13.3 Use status, not only classification

Classification is computed by the tool. Status is decided by the reviewer. They are different.

A row may be `tune-required` but `accepted`. That happened with Shows. A row may be `review` but `needs-work` if the remaining difference is visually important. The production app must store both values.

Recommended status values:

- `unreviewed`: no human decision yet.
- `accepted`: reviewer says this is fine.
- `needs-work`: reviewer wants a change.
- `fixed`: developer says a change was made after feedback.
- `wont-fix`: acknowledged difference that should remain.

### 13.4 Add keyboard shortcuts

Visual review benefits from speed. Suggested shortcuts:

| Shortcut | Action |
|---|---|
| `j` | next card |
| `k` | previous card |
| `a` | mark accepted |
| `n` | mark needs-work |
| `f` | mark fixed |
| `c` | copy current card note with YAML |
| `d` | focus diff image |
| `r` | focus React image |
| `p` | focus prototype image |
| `/` | focus filter/search |

### 13.5 Add zoom and pan

The current static page relies on browser zoom and scrollable figures. A real app should provide:

- Fit-to-width.
- Actual-size view.
- Synchronized pan between prototype and React images.
- Optional overlay mode.
- Toggle to highlight changed pixels.

### 13.6 Make metadata useful, not noisy

Raw `compare.json` is useful for debugging, but most reviewers do not need to see it first. The app should show a compact metadata summary by default:

```text
Bounds changed:
  left:  y=61 h=1088.72
  right: y=61 h=1136.47
  delta: h=47.75

Changed styles:
  color: left rgb(31, 30, 28), right rgb(26, 26, 24)
  font-size: left 16px, right 14px
```

Then provide raw JSON behind an advanced disclosure.

## 14. Import algorithm for the production app

The current generator’s import algorithm can be translated almost directly.

```text
function importRun(summaryJsonPath, runName):
    summary = parseJson(summaryJsonPath)
    rows = normalizeRows(summary)

    run = createRun({
        name: runName,
        createdAt: now(),
        summaryJsonPath,
        gitSha: currentGitShaIfAvailable(),
    })

    for row in rows:
        target = createOrFindTarget({
            runId: run.id,
            page: row.page,
            section: row.section,
            variant: row.variant or 'desktop',
        })

        artifactDir = dirname(row.diffOnlyPath)
        copiedArtifacts = copyArtifacts({
            diffOnly: artifactDir + '/diff_only.png',
            leftRegion: artifactDir + '/left_region.png',
            rightRegion: artifactDir + '/right_region.png',
            diffComparison: artifactDir + '/diff_comparison.png',
            compareJson: artifactDir + '/compare.json',
        }, destination = run.artifactRoot / row.page / row.section)

        compare = parseJson(copiedArtifacts.compareJson)

        createComparisonResult({
            runId: run.id,
            targetId: target.id,
            classification: row.classification,
            changedPercent: row.changedPercent,
            changedPixels: compare.pixel.changedPixels,
            totalPixels: compare.pixel.totalPixels,
            bounds: compare.bounds,
            changedStyles: changed(compare.styles),
            changedAttributes: changed(compare.attributes),
            artifacts: copiedArtifacts,
        })

    return run
```

The important failure mode is missing artifacts. The importer should not silently produce broken cards. It should record missing files and show them in the UI.

Suggested missing-artifact behavior:

```ts
type ArtifactImportStatus =
  | { ok: true; copiedPath: string }
  | { ok: false; sourcePath: string; error: 'missing' | 'unreadable' | 'unsupported' };
```

## 15. Handling accepted differences

The visual spec currently has an `acceptedDifferences: {}` section. It is empty. The review app should provide a way to record human-accepted differences and optionally write them back into a spec or sidecar file.

There are two kinds of accepted differences:

1. **Run-local accepted differences.** These apply only to one review run. Example: “Shows page is fine for this pass.”
2. **Spec-level accepted differences.** These are stable and intentional. Example: “Browser-native form controls differ between prototype and React; ignore minor antialiasing.”

A production app should not automatically change the visual spec. It should propose changes and require explicit confirmation.

Suggested sidecar file:

```yaml
acceptedDifferences:
  public-pages.desktop:
    shows:
      page:
        status: accepted
        reason: "User accepted remaining poster/list deltas during 2026-04-27 review."
        acceptedBy: manuel
        acceptedAt: 2026-04-27T17:29:00Z
        maxObservedChangedPercent: 11.620
```

## 16. Error handling and known pitfalls

### 16.1 Stale Storybook/Vite transforms

During this ticket, Storybook occasionally served stale or incomplete module transforms. Symptoms included:

- CSS files appearing empty or unstyled in screenshots.
- Runtime export errors such as a module not providing an export that existed in the source file.
- Pages rendering in a broken stacked layout even though CSS media queries were not active.

Mitigation:

- Restart Storybook.
- Touch the relevant CSS/TSX files.
- Re-run a small focused comparison before trusting a full sweep.

A production app should detect common failures by checking page console errors and screenshot dimensions. A run with module export errors should be marked suspect.

### 16.2 Selector mismatch

A high diff can mean the implementation is visually wrong. It can also mean the selector is wrong. One example from this ticket was the Shows header target: comparing the wrapper instead of the actual `PublicPageHeader` component inflated the header diff. The fix was to narrow the React selector:

```yaml
header:
  original: "[data-section='shows-header']"
  react: "[data-section='shows-header'] [data-pyxis-component='public-page-header']"
```

The review app should make selectors visible. It should not hide this information behind raw JSON only.

### 16.3 High percentage does not always mean high priority

Shows stayed above 10% numerically, but the user accepted it visually. The app must support human judgment overriding numerical priority. The number starts the conversation; it does not end it.

### 16.4 Linked artifacts can disappear

Do not link to arbitrary `/tmp` artifacts from a long-lived review page. Always copy artifacts into the bundle or run storage. This is one of the most important lessons from the prototype.

### 16.5 Clipboard API requires a browser context

The static page uses `navigator.clipboard.writeText`. This usually works when served over `localhost`. It may fail when opening the file directly as `file://`. Serve the bundle over HTTP.

## 17. Security considerations

The current generator escapes most dynamic HTML values using Python’s `html.escape`, but a production version should treat all imported summary content and review notes as untrusted.

Important rules:

- Escape page names, section names, paths, and metadata before rendering HTML.
- Do not use unsanitized `innerHTML` for review notes.
- Restrict artifact serving to known run directories.
- Do not allow arbitrary file reads through artifact URLs.
- If the app shells out to `css-visual-diff`, validate arguments and avoid passing user-controlled strings to a shell.
- If multiple users review, add authentication before storing notes.

The current script is safe enough for local trusted use. It is not a multi-user web application.

## 18. Implementation plan for the intern

### Phase 1: Rebuild the current static workflow as a small app

Goal: Match the existing behavior with cleaner code.

Tasks:

- Create a `visual-review` package or app.
- Implement an importer that reads the summary JSON and copies artifacts.
- Render a run page with review cards.
- Include diff/prototype/React images.
- Include metadata summary and raw JSON disclosure.
- Include textarea notes in local component state.
- Include clipboard export.

Definition of done:

- Given `/tmp/pyxis-public-pages-after-about-spacing.json`, the app can import it and show all cards.
- The UI can copy one card with note + YAML.
- The UI can copy all cards with notes + YAML.

### Phase 2: Persist notes and status

Goal: Make human feedback durable.

Tasks:

- Add SQLite or filesystem-backed persistence.
- Add status controls.
- Save notes automatically or through a Save button.
- Show unreviewed/accepted/needs-work counts.
- Add filters by status and classification.

Definition of done:

- Reloading the page preserves notes and statuses.
- A reviewer can mark Shows as accepted even if its classification is `tune-required`.

### Phase 3: Add run history

Goal: Make the tool useful across multiple tuning iterations.

Tasks:

- Add a runs index page.
- Store git SHA, branch, spec path, creation time, and command/log metadata.
- Allow comparing the same target across two runs.
- Show improved/regressed changed-percent deltas.

Definition of done:

- A developer can answer “did About improve after commit 46e138d?” inside the app.

### Phase 4: Add comparison execution

Goal: Let the app start visual-diff runs.

Tasks:

- Add a backend job endpoint.
- Validate required servers are reachable.
- Run `css-visual-diff` as a child process.
- Stream logs.
- Import artifacts automatically when the job completes.

Definition of done:

- A user can click “Run public pages desktop spec” and get a new review run.

### Phase 5: Add collaboration polish

Goal: Make it suitable for team use.

Tasks:

- Add authentication if needed.
- Add reviewer names.
- Add PR/issue export.
- Add comments or threaded notes.
- Add shareable deep links to specific cards.
- Add artifact retention policy.

Definition of done:

- A reviewer can send a developer a link to `run/about/content` with persisted feedback.

## 19. Suggested technology choices

Because this repo already uses React, Storybook, Vite, and TypeScript, the natural implementation stack is:

- **Frontend:** React + Vite + TypeScript.
- **Backend:** Node/Express, Fastify, or a Go service. Choose Go if integrating with existing Go tooling is important; choose Node if the developer wants easiest integration with frontend package scripts.
- **Storage:** SQLite for metadata and notes; filesystem for images and JSON artifacts.
- **Image serving:** Static file middleware rooted at the run artifact directory.
- **Job execution:** Child process wrapper around `css-visual-diff` for later phases.

A first version can also be a pure React app that imports a prebuilt JSON manifest. That is close to the current static generator and reduces backend work, but it will not persist notes unless using local storage or a file API.

## 20. Component decomposition for a React implementation

Recommended components:

```text
VisualReviewApp
  RunPicker
  RunSummaryBar
  FilterToolbar
  ReviewCardList
    ReviewCard
      CardHeader
      ArtifactLinks
      ImageComparisonPanel
        DiffImage
        PrototypeImage
        ReactImage
      MetadataPanel
        BoundsSummary
        PixelSummary
        StyleChangesTable
        RawCompareJson
      ReviewControls
        StatusSelect
        NotesTextarea
        ClipboardButton
```

### 20.1 Component responsibilities

| Component | Responsibility |
|---|---|
| `VisualReviewApp` | Owns selected run, filters, and global keyboard shortcuts. |
| `RunPicker` | Lists historical runs and lets the user choose one. |
| `RunSummaryBar` | Shows counts by classification and status. |
| `FilterToolbar` | Filters cards by page, section, classification, status, and search text. |
| `ReviewCardList` | Sorts and renders cards. |
| `ReviewCard` | Owns per-target layout and review controls. |
| `ImageComparisonPanel` | Shows diff/prototype/React images and layout toggles. |
| `MetadataPanel` | Shows useful compare metadata without forcing raw JSON reading. |
| `ReviewControls` | Saves notes, status, and clipboard output. |

### 20.2 Clipboard utility pseudocode

```ts
function buildClipboardText(result: ComparisonResult, note: ReviewNote): string {
  return [
    `${result.page}/${result.section}`,
    `${result.classification} · ${result.changedPercent.toFixed(3)}%`,
    `Status: ${note.status}`,
    `Notes: ${note.note}`,
    '',
    'compare_yaml:',
    toYaml({
      page: result.page,
      section: result.section,
      classification: result.classification,
      changedPercent: result.changedPercent,
      bounds: result.bounds,
      pixel: {
        changedPixels: result.changedPixels,
        totalPixels: result.totalPixels,
        normalizedWidth: result.normalizedWidth,
        normalizedHeight: result.normalizedHeight,
      },
      artifacts: result.artifacts,
      changedStyles: result.changedStyles.slice(0, 20),
      changedAttributes: result.changedAttributes.slice(0, 20),
    }),
  ].join('\n');
}
```

## 21. Testing strategy

The production app should be tested at three levels.

### 21.1 Importer tests

Importer tests should use small fixture summary JSON files and fake artifact directories.

Test cases:

- It accepts a list-shaped summary with `payload[0].rows`.
- It accepts an object-shaped summary with `payload.rows`.
- It copies all expected artifacts.
- It records missing artifacts as errors.
- It parses `compare.json` and extracts changed styles.
- It preserves page, section, classification, and changed percent.

### 21.2 UI tests

UI tests should use a fixture run.

Test cases:

- It renders one card per result.
- It filters by classification.
- It filters by status.
- It displays diff/prototype/React images.
- It expands metadata.
- It saves notes.
- It copies expected clipboard text.

### 21.3 End-to-end tests

End-to-end tests should import a real or realistic run and serve the UI.

Test cases:

- Import a run from a known summary JSON.
- Open the run in the browser.
- Mark one result accepted.
- Reload the page.
- Verify the status and note persist.
- Export notes and verify Markdown/YAML includes the decision.

## 22. Operational runbook

### 22.1 Create a new review run

```bash
cd /home/manuel/code/wesen/2026-04-23--pyxis

RUN_NAME=public-pages-$(date +%Y%m%d-%H%M%S)
ARTIFACT_DIR=/tmp/$RUN_NAME
SUMMARY_JSON=/tmp/$RUN_NAME.json

rm -rf "$ARTIFACT_DIR"
css-visual-diff verbs \
  --repository prototype-design/visual-diff/userland \
  pyxis pages compare-spec \
  prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml \
  --outDir "$ARTIFACT_DIR" \
  --summary \
  --output json \
  > "$SUMMARY_JSON"

REVIEW_DIR=/tmp/$RUN_NAME-review
python3 ttmp/2026/04/27/PYXIS-APP-VISUAL-TUNING--pyxis-app-visual-tuning-topbar-dashboard-new-pages/scripts/05-build-public-pages-review.py \
  --summary-json "$SUMMARY_JSON" \
  --output-dir "$REVIEW_DIR"

cd "$REVIEW_DIR"
python3 -m http.server 8097 --bind 127.0.0.1
```

### 22.2 Review and produce feedback

1. Open `http://127.0.0.1:8097/index.html`.
2. Click `Tune-required` first, unless a human has already accepted those rows.
3. Inspect `diff_only` first.
4. Write notes in each relevant textarea.
5. Click `To clipboard with YAML` for individual feedback, or `All notes to clipboard` for a batch.
6. Paste feedback into the implementation ticket or handoff document.

### 22.3 Validate after changes

After implementing changes, run:

```bash
cd web/packages/pyxis-user-site
pnpm exec tsc --noEmit
pnpm exec vite build
pnpm exec storybook build
```

Then rerun the visual sweep and regenerate the review site.

## 23. What should not be copied into the real app

The prototype served its purpose, but some parts should not become long-term architecture.

Do not copy these as-is:

- Inline JavaScript event handlers such as `onclick="copyCard(this)"`.
- One giant Python f-string that emits the entire app.
- No persistence for notes.
- No explicit run database.
- No missing-artifact diagnostics in the UI.
- No authentication or multi-user support.
- No typed data model.

Do copy these ideas:

- Self-contained artifact bundles.
- One card per page/section result.
- Diff-first review workflow.
- Human status separate from computed classification.
- Clipboard output that includes YAML context.
- Raw compare metadata available behind a disclosure.
- Stable links to prototype, React, diff, and compare JSON artifacts.

## 24. Handoff summary for the next developer

The current review site is a static, local-first tool that converts `css-visual-diff` output into an operator-friendly HTML page. It exists because a human reviewer needs to look at visual evidence, write feedback, and pass that feedback to someone who can change the code. The key design lesson is that the review site should not replace visual-diff measurement; it should make visual-diff evidence reviewable, annotatable, and shareable.

The first production milestone should be modest: import a summary JSON, copy artifacts, render review cards, and persist notes/status. Once that works, add run history and comparison execution. The app will become useful quickly if it preserves the current workflow while solving the current prototype’s biggest limitation: feedback disappears when the page reloads.

The next developer should start by reading these files:

```text
ttmp/2026/04/27/PYXIS-APP-VISUAL-TUNING--pyxis-app-visual-tuning-topbar-dashboard-new-pages/scripts/05-build-public-pages-review.py
prototype-design/visual-diff/userland/specs/public-pages.desktop.visual.yml
prototype-design/visual-diff/userland/verbs/pyxis-pages.js
web/packages/pyxis-user-site/src/pages/storybook.tsx
web/packages/pyxis-user-site/src/pages/*Page/Page.stories.tsx
```

Then they should generate one fresh review bundle and use it manually. The quickest way to understand the system is to experience the loop: run comparisons, open the review page, inspect a card, type a note, copy the YAML, and imagine where that note should live if this were a real collaborative app.
