"""
Browser automation module for capturing element screenshots and CSS information.
"""

import json
from pathlib import Path
from typing import Dict, Any, Tuple
from playwright.sync_api import sync_playwright, Page, ElementHandle


def capture_element(
    url: str,
    selector: str,
    output_dir: Path,
    prefix: str
) -> Dict[str, Any]:
    """
    Capture screenshot and CSS information for an element.
    
    Args:
        url: The URL to navigate to
        selector: CSS selector for the element
        output_dir: Directory to save outputs
        prefix: Prefix for output files (e.g., 'url1', 'url2')
    
    Returns:
        Dictionary containing paths and metadata
    """
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        
        # Navigate to URL
        page.goto(url, wait_until='networkidle')
        
        # Wait for element to be visible
        element = page.wait_for_selector(selector, state='visible', timeout=30000)
        
        if not element:
            raise Exception(f"Element not found: {selector}")
        
        # Capture screenshot
        screenshot_path = output_dir / f"{prefix}_screenshot.png"
        element.screenshot(path=str(screenshot_path))
        
        # Extract computed styles
        computed_styles = page.evaluate("""
            (selector) => {
                const element = document.querySelector(selector);
                if (!element) return null;
                
                const computed = window.getComputedStyle(element);
                const styles = {};
                
                // Get all computed style properties
                for (let i = 0; i < computed.length; i++) {
                    const prop = computed[i];
                    styles[prop] = computed.getPropertyValue(prop);
                }
                
                return styles;
            }
        """, selector)
        
        # Extract matching CSS rules
        matching_rules = page.evaluate("""
            (selector) => {
                const element = document.querySelector(selector);
                if (!element) return [];
                
                const rules = [];
                const sheets = Array.from(document.styleSheets);
                
                for (const sheet of sheets) {
                    try {
                        const cssRules = Array.from(sheet.cssRules || []);
                        
                        for (const rule of cssRules) {
                            if (rule.selectorText && element.matches(rule.selectorText)) {
                                rules.push({
                                    selector: rule.selectorText,
                                    cssText: rule.style.cssText,
                                    href: sheet.href || 'inline'
                                });
                            }
                        }
                    } catch (e) {
                        // Skip stylesheets that can't be accessed (CORS)
                        continue;
                    }
                }
                
                return rules;
            }
        """, selector)
        
        # Get element dimensions and position
        bounding_box = element.bounding_box()
        
        # Get element HTML
        element_html = page.evaluate("""
            (selector) => {
                const element = document.querySelector(selector);
                return element ? element.outerHTML : null;
            }
        """, selector)
        
        browser.close()
        
        # Save CSS data
        css_data = {
            'url': url,
            'selector': selector,
            'computed_styles': computed_styles,
            'matching_rules': matching_rules,
            'bounding_box': bounding_box,
            'element_html': element_html
        }
        
        css_path = output_dir / f"{prefix}_css_data.json"
        with open(css_path, 'w') as f:
            json.dump(css_data, f, indent=2)
        
        return {
            'screenshot_path': str(screenshot_path),
            'css_path': str(css_path),
            'css_data': css_data
        }


def capture_both_elements(
    url1: str,
    selector1: str,
    url2: str,
    selector2: str,
    output_dir: Path
) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    """
    Capture screenshots and CSS information for both elements.
    
    Returns:
        Tuple of (element1_data, element2_data)
    """
    output_dir.mkdir(parents=True, exist_ok=True)
    
    print(f"Capturing element from {url1}...")
    element1_data = capture_element(url1, selector1, output_dir, "url1")
    
    print(f"Capturing element from {url2}...")
    element2_data = capture_element(url2, selector2, output_dir, "url2")
    
    return element1_data, element2_data
