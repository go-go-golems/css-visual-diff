# CSS Visual Diff

A CLI tool for pixel-perfect visual comparison of web page elements with AI-powered analysis.

## Overview

CSS Visual Diff captures screenshots of specific elements from two different URLs, generates visual diffs, extracts comprehensive CSS information, and uses a visual LLM to analyze the differences. Perfect for iterative design work, visual regression testing, and understanding CSS changes.

## Features

- **Pixel-Perfect Rendering** - Uses real Chromium browser via Playwright for accurate rendering
- **Flexible Selectors** - Compare elements with different CSS selectors across pages
- **Visual Diff Generation** - Side-by-side comparison with pixel-level difference highlighting
- **CSS Extraction** - Captures both computed styles and matching CSS rules
- **AI Analysis** - GPT-4 vision model provides detailed insights about differences
- **Comprehensive Output** - Screenshots, diffs, CSS data, and analysis reports

## Installation

### Prerequisites

- Python 3.11+
- pip3

### Setup

```bash
# Install Python dependencies
sudo pip3 install playwright pillow openai click

# Install Playwright browsers
playwright install chromium
```

## Usage

### Basic Command

```bash
python3 src/cli.py \
  --url1 https://example.com/page1 \
  --selector1 ".header-nav" \
  --url2 https://example.com/page2 \
  --selector2 ".navigation-header" \
  --question "What are the main visual differences and what CSS changes caused them?"
```

### Options

| Option | Required | Description |
|--------|----------|-------------|
| `--url1` | Yes | First URL to compare |
| `--selector1` | Yes | CSS selector for element in first URL |
| `--url2` | Yes | Second URL to compare |
| `--selector2` | Yes | CSS selector for element in second URL |
| `--question` | Yes | Question to ask the LLM about the comparison |
| `--output-dir` | No | Output directory (default: `./output_TIMESTAMP`) |
| `--threshold` | No | Pixel difference threshold 0-255 (default: 30) |
| `--model` | No | LLM model to use (default: `gpt-4.1-mini`) |
| `--no-analysis` | No | Skip LLM analysis (only capture and diff) |

### Examples

#### Compare Navigation Headers

```bash
python3 src/cli.py \
  --url1 file:///path/to/page1.html \
  --selector1 ".header-nav" \
  --url2 file:///path/to/page2.html \
  --selector2 ".navigation-header" \
  --question "What are the visual differences in the navigation header? Focus on colors, spacing, and typography."
```

#### Compare Buttons

```bash
python3 src/cli.py \
  --url1 https://staging.example.com \
  --selector1 ".btn-primary" \
  --url2 https://production.example.com \
  --selector2 ".button-primary" \
  --question "What are the exact differences in button styling? Compare padding, colors, and border-radius."
```

#### Compare Pricing Tables

```bash
python3 src/cli.py \
  --url1 https://example.com/pricing-old \
  --selector1 ".pricing-card" \
  --url2 https://example.com/pricing-new \
  --selector2 ".price-box" \
  --question "Which pricing table design is more effective for conversion and why?"
```

## Output Files

Each comparison generates:

| File | Description |
|------|-------------|
| `url1_screenshot.png` | Screenshot of element from first URL |
| `url2_screenshot.png` | Screenshot of element from second URL |
| `url1_css_data.json` | Complete CSS information for first element |
| `url2_css_data.json` | Complete CSS information for second element |
| `diff_comparison.png` | Side-by-side comparison with diff visualization |
| `diff_only.png` | Standalone diff highlighting changes in red |
| `analysis_report.md` | LLM analysis of the differences |
| `summary.json` | Machine-readable summary of the comparison |

## CSS Data Structure

The CSS data JSON files contain:

```json
{
  "url": "https://example.com",
  "selector": ".header-nav",
  "computed_styles": {
    "background": "linear-gradient(...)",
    "padding": "20px 30px",
    "font-size": "24px",
    ...
  },
  "matching_rules": [
    {
      "selector": ".header-nav",
      "cssText": "background: ...; padding: ...;",
      "href": "https://example.com/styles.css"
    }
  ],
  "bounding_box": {
    "x": 20,
    "y": 20,
    "width": 1240,
    "height": 68
  },
  "element_html": "<nav class=\"header-nav\">...</nav>"
}
```

## Testing

Run the comprehensive test suite:

```bash
cd tests
./run_tests.sh
```

This will:
1. Compare navigation headers across test pages
2. Compare card widgets with different selectors
3. Compare pricing tables
4. Compare button elements
5. Generate a comprehensive test report

Test results are saved to `tests/results/` with subdirectories for each test case.

## Use Cases

### Visual Regression Testing

Compare elements across different versions or environments to detect unintended changes:

```bash
python3 src/cli.py \
  --url1 https://staging.myapp.com \
  --selector1 ".checkout-form" \
  --url2 https://production.myapp.com \
  --selector2 ".checkout-form" \
  --question "Are there any visual differences in the checkout form?"
```

### Design System Consistency

Verify components match design specifications:

```bash
python3 src/cli.py \
  --url1 https://design-system.myapp.com/button \
  --selector1 ".btn-example" \
  --url2 https://myapp.com/dashboard \
  --selector2 ".dashboard-button" \
  --question "Does the dashboard button match the design system specifications?"
```

### Iterating to Pixel-Perfect

Compare your implementation against a reference design:

```bash
python3 src/cli.py \
  --url1 file:///path/to/reference.html \
  --selector1 ".hero-section" \
  --url2 file:///path/to/implementation.html \
  --selector2 ".hero" \
  --question "What CSS changes are needed to match the reference design exactly?"
```

### A/B Testing Analysis

Understand visual differences between variants:

```bash
python3 src/cli.py \
  --url1 https://myapp.com/variant-a \
  --selector1 ".cta-button" \
  --url2 https://myapp.com/variant-b \
  --selector2 ".cta-button" \
  --question "Which button design is more visually prominent and likely to convert better?"
```

## Architecture

The tool consists of four main modules:

1. **browser_capture.py** - Browser automation for element capture and CSS extraction
2. **image_diff.py** - Visual diff generation with pixel-level comparison
3. **llm_analysis.py** - AI-powered analysis using OpenAI's vision API
4. **cli.py** - Command-line interface and orchestration

## Requirements

- Python 3.11+
- Playwright (with Chromium)
- Pillow (PIL)
- OpenAI Python SDK
- Click

## Environment Variables

Set `OPENAI_API_KEY` for LLM analysis:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Limitations

- Requires internet connection for LLM analysis
- LLM analysis incurs API costs (typically 2000-4000 tokens per comparison)
- Screenshots are limited to visible viewport (use `--viewport-size` if needed)
- CORS restrictions may prevent CSS extraction from external stylesheets

## License

MIT

## Contributing

Contributions welcome! Please open an issue or pull request.

## Support

For issues or questions, please open a GitHub issue.
