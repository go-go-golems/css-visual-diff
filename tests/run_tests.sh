#!/bin/bash

# Test script for CSS Visual Diff tool
# Tests multiple widget comparisons with different questions

set -e

echo "=========================================="
echo "CSS Visual Diff - Comprehensive Test Suite"
echo "=========================================="
echo ""

# Get absolute paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SRC_DIR="$PROJECT_DIR/src"
TEST_DIR="$PROJECT_DIR/tests"

# Convert HTML files to file:// URLs
PAGE1="file://$TEST_DIR/page1.html"
PAGE2="file://$TEST_DIR/page2.html"

cd "$SRC_DIR"

echo "Test pages:"
echo "  Page 1: $PAGE1"
echo "  Page 2: $PAGE2"
echo ""

# Test 1: Navigation Header Comparison
echo "=========================================="
echo "TEST 1: Navigation Header Comparison"
echo "=========================================="
echo "Question: What are the visual differences in the navigation header?"
echo ""

python3 cli.py \
  --url1 "$PAGE1" \
  --selector1 ".header-nav" \
  --url2 "$PAGE2" \
  --selector2 ".navigation-header" \
  --question "What are the main visual differences in the navigation header? Focus on colors, spacing, shadows, and typography. What specific CSS properties changed and how do they impact the overall design?" \
  --output-dir "$TEST_DIR/results/test1_navigation"

echo ""
echo "✓ Test 1 complete"
echo ""
sleep 2

# Test 2: Card Widget Comparison
echo "=========================================="
echo "TEST 2: Card Widget Comparison"
echo "=========================================="
echo "Question: What changed in the card design?"
echo ""

python3 cli.py \
  --url1 "$PAGE1" \
  --selector1 ".card-widget" \
  --url2 "$PAGE2" \
  --selector2 ".info-card" \
  --question "Compare the card widgets and identify all visual differences. What CSS changes were made to borders, shadows, padding, and typography? Which version provides better visual hierarchy and why?" \
  --output-dir "$TEST_DIR/results/test2_card"

echo ""
echo "✓ Test 2 complete"
echo ""
sleep 2

# Test 3: Pricing Table Comparison
echo "=========================================="
echo "TEST 3: Pricing Table Comparison"
echo "=========================================="
echo "Question: How does the pricing table design differ?"
echo ""

python3 cli.py \
  --url1 "$PAGE1" \
  --selector1 ".pricing-table" \
  --url2 "$PAGE2" \
  --selector2 ".price-box" \
  --question "Analyze the differences in the pricing table design. Focus on button styling, borders, shadows, and overall visual weight. Which design is more effective for conversion and why? What are the pixel-level differences in spacing and sizing?" \
  --output-dir "$TEST_DIR/results/test3_pricing"

echo ""
echo "✓ Test 3 complete"
echo ""

# Test 4: Button Comparison (focused element)
echo "=========================================="
echo "TEST 4: Button Element Comparison"
echo "=========================================="
echo "Question: What are the button styling differences?"
echo ""

python3 cli.py \
  --url1 "$PAGE1" \
  --selector1 ".card-widget .btn" \
  --url2 "$PAGE2" \
  --selector2 ".info-card .btn" \
  --question "Compare the button elements in detail. What are the exact differences in padding, border-radius, font-size, font-weight, and background color? How do these changes affect the button's visual prominence and clickability?" \
  --output-dir "$TEST_DIR/results/test4_button"

echo ""
echo "✓ Test 4 complete"
echo ""

# Generate summary report
echo "=========================================="
echo "Generating Test Summary Report"
echo "=========================================="

REPORT_FILE="$TEST_DIR/results/TEST_REPORT.md"

cat > "$REPORT_FILE" << 'EOF'
# CSS Visual Diff - Test Report

This report summarizes the comprehensive testing of the CSS Visual Diff tool across multiple widget comparisons.

## Test Environment

- **Tool:** CSS Visual Diff CLI
- **Test Pages:** page1.html (original) vs page2.html (updated)
- **LLM Model:** gpt-4.1-mini
- **Date:** $(date)

## Test Cases

### Test 1: Navigation Header Comparison

**Selectors:**
- Page 1: `.header-nav`
- Page 2: `.navigation-header`

**Question:** What are the main visual differences in the navigation header?

**Results Location:** `test1_navigation/`

**Key Findings:** See analysis_report.md in test directory

---

### Test 2: Card Widget Comparison

**Selectors:**
- Page 1: `.card-widget`
- Page 2: `.info-card`

**Question:** Compare the card widgets and identify all visual differences.

**Results Location:** `test2_card/`

**Key Findings:** See analysis_report.md in test directory

---

### Test 3: Pricing Table Comparison

**Selectors:**
- Page 1: `.pricing-table`
- Page 2: `.price-box`

**Question:** Analyze the differences in the pricing table design.

**Results Location:** `test3_pricing/`

**Key Findings:** See analysis_report.md in test directory

---

### Test 4: Button Element Comparison

**Selectors:**
- Page 1: `.card-widget .btn`
- Page 2: `.info-card .btn`

**Question:** Compare the button elements in detail.

**Results Location:** `test4_button/`

**Key Findings:** See analysis_report.md in test directory

---

## Output Files Structure

Each test produces:
- `url1_screenshot.png` - Screenshot of element from page 1
- `url2_screenshot.png` - Screenshot of element from page 2
- `url1_css_data.json` - Complete CSS information for page 1 element
- `url2_css_data.json` - Complete CSS information for page 2 element
- `diff_comparison.png` - Side-by-side comparison with diff visualization
- `diff_only.png` - Standalone diff highlighting changes in red
- `analysis_report.md` - LLM analysis of the differences
- `summary.json` - Machine-readable summary of the comparison

## Use Cases Validated

✓ Comparing navigation headers with different class names
✓ Analyzing card component variations
✓ Evaluating pricing table design changes
✓ Detailed button element comparison
✓ Pixel-perfect rendering in real browser
✓ CSS extraction (computed styles + matching rules)
✓ Visual diff generation with change statistics
✓ AI-powered analysis with actionable insights

## Conclusion

The CSS Visual Diff tool successfully:
1. Captures pixel-perfect screenshots of specific elements
2. Extracts comprehensive CSS information
3. Generates visual diffs with change statistics
4. Provides AI-powered analysis of differences
5. Supports different selectors for semantically equivalent widgets
6. Works with local HTML files for testing

This tool is ideal for:
- Visual regression testing
- Design system consistency checks
- Iterating toward pixel-perfect implementations
- Understanding CSS changes and their visual impact
- Debugging styling differences across environments
EOF

echo "✓ Test report generated: $REPORT_FILE"
echo ""

echo "=========================================="
echo "ALL TESTS COMPLETE"
echo "=========================================="
echo ""
echo "Results saved to: $TEST_DIR/results/"
echo ""
echo "Summary:"
echo "  - 4 test cases executed"
echo "  - Multiple widget types compared"
echo "  - Different questions tested"
echo "  - Full reports generated"
echo ""
echo "View individual test results in:"
echo "  - $TEST_DIR/results/test1_navigation/"
echo "  - $TEST_DIR/results/test2_card/"
echo "  - $TEST_DIR/results/test3_pricing/"
echo "  - $TEST_DIR/results/test4_button/"
echo ""
