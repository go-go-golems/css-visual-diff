package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

type CaptureResult struct {
	Original   PageResult         `json:"original"`
	React      PageResult         `json:"react"`
	Coverage   CoverageSummary    `json:"coverage"`
	Validation []ValidationResult `json:"validation,omitempty"`
}

type PageResult struct {
	Name           string          `json:"name"`
	URL            string          `json:"url"`
	FullScreenshot string          `json:"full_screenshot"`
	PreparedHTML   string          `json:"prepared_html,omitempty"`
	InspectJSON    string          `json:"inspect_json,omitempty"`
	Sections       []SectionResult `json:"sections"`
}

type SectionResult struct {
	Name             string    `json:"name"`
	Selector         string    `json:"selector"`
	Exists           bool      `json:"exists"`
	Visible          bool      `json:"visible"`
	Bounds           *Bounds   `json:"bounds,omitempty"`
	TextStart        string    `json:"text_start,omitempty"`
	Screenshot       string    `json:"screenshot"`
	PNG              *PNGStats `json:"png,omitempty"`
	ValidationIssues []string  `json:"validation_issues,omitempty"`
}

type CoverageSummary struct {
	Total           int `json:"total"`
	OriginalMissing int `json:"original_missing"`
	ReactMissing    int `json:"react_missing"`
	OriginalHidden  int `json:"original_hidden"`
	ReactHidden     int `json:"react_hidden"`
}

type domCheck struct {
	Exists    bool    `json:"exists"`
	Visible   bool    `json:"visible"`
	Bounds    *Bounds `json:"bounds"`
	TextStart string  `json:"text_start"`
	Text      string  `json:"text"`
}

func RunCapture(ctx context.Context, cfg *config.Config) error {
	if !cfg.Output.WritePNGs && !cfg.Output.WriteJSON && !cfg.Output.WriteMarkdown {
		return nil
	}

	if err := os.MkdirAll(cfg.Output.Dir, 0o755); err != nil {
		return err
	}

	browser, err := driver.NewBrowser(ctx)
	if err != nil {
		return err
	}
	defer browser.Close()

	original, originalValidation, err := captureTarget(browser, cfg.Original, cfg.Sections, cfg.Output, "original")
	if err != nil {
		return err
	}

	react, reactValidation, err := captureTarget(browser, cfg.React, cfg.Sections, cfg.Output, "react")
	if err != nil {
		return err
	}

	result := CaptureResult{Original: original, React: react}
	result.Validation = append(result.Validation, originalValidation...)
	result.Validation = append(result.Validation, reactValidation...)
	result.Coverage = computeCoverage(original, react)

	if cfg.Output.WriteJSON {
		if err := writeJSON(filepath.Join(cfg.Output.Dir, "capture.json"), result); err != nil {
			return err
		}
	}

	if cfg.Output.WriteMarkdown {
		if err := writeMarkdown(filepath.Join(cfg.Output.Dir, "capture.md"), result); err != nil {
			return err
		}
	}

	return nil
}

func captureTarget(browser *driver.Browser, target config.Target, sections []config.SectionSpec, output config.OutputSpec, prefix string) (PageResult, []ValidationResult, error) {
	page, err := browser.NewPage()
	if err != nil {
		return PageResult{}, nil, err
	}
	defer page.Close()

	if err := page.SetViewport(target.Viewport.Width, target.Viewport.Height); err != nil {
		return PageResult{}, nil, err
	}

	if err := page.Goto(target.URL); err != nil {
		return PageResult{}, nil, err
	}

	if target.WaitMS > 0 {
		page.Wait(time.Duration(target.WaitMS) * time.Millisecond)
	}
	if err := prepareTarget(page, target); err != nil {
		return PageResult{}, nil, err
	}

	var validation []ValidationResult
	pageResult := PageResult{Name: target.Name, URL: target.URL}
	rootSelector := rootSelectorForTarget(target)
	if output.WritePreparedHTML {
		preparedHTMLPath := filepath.Join(output.Dir, fmt.Sprintf("%s-prepared.html", prefix))
		if err := writePreparedHTML(page, rootSelector, preparedHTMLPath); err != nil {
			return PageResult{}, nil, err
		}
		pageResult.PreparedHTML = preparedHTMLPath
	}
	if output.WriteInspectJSON {
		inspectPath := filepath.Join(output.Dir, fmt.Sprintf("%s-inspect.json", prefix))
		if err := writeInspectJSON(page, rootSelector, inspectPath); err != nil {
			return PageResult{}, nil, err
		}
		pageResult.InspectJSON = inspectPath
	}

	fullPath := filepath.Join(output.Dir, fmt.Sprintf("%s-full.png", prefix))
	if rootSelector != "" {
		if err := page.Screenshot(rootSelector, fullPath); err != nil {
			return PageResult{}, nil, err
		}
	} else if err := page.FullScreenshot(fullPath); err != nil {
		return PageResult{}, nil, err
	}
	pageResult.FullScreenshot = fullPath

	for _, section := range sections {
		selector := selectorForSection(section, prefix)

		check := domCheck{}
		err := evaluateDOMCheck(page, selector, &check)
		if err != nil {
			return PageResult{}, nil, err
		}

		issues := validateTextExpectations(section, prefix, check.Text)
		result := SectionResult{
			Name:             section.Name,
			Selector:         selector,
			Exists:           check.Exists,
			Visible:          check.Visible,
			Bounds:           check.Bounds,
			TextStart:        check.TextStart,
			ValidationIssues: append([]string{}, issues...),
		}
		validationResult := ValidationResult{
			Target:  prefix,
			Section: section.Name,
			Status:  "ok",
			Issues:  append([]string{}, issues...),
			Bounds:  check.Bounds,
		}

		if check.Exists {
			screenshotPath := filepath.Join(output.Dir, fmt.Sprintf("%s-%s.png", prefix, section.Name))
			if err := page.Screenshot(selector, screenshotPath); err != nil {
				return PageResult{}, nil, err
			}
			result.Screenshot = screenshotPath

			if output.ValidatePNGs {
				stats, err := analyzePNG(screenshotPath, 20)
				if err != nil {
					issues = append(issues, fmt.Sprintf("failed to analyze PNG: %v", err))
				} else {
					pngExpectations := pngExpectationsForSection(section, prefix)
					pngIssues := validatePNGStats(stats, pngExpectations)
					issues = append(issues, pngIssues...)
					result.PNG = &stats
					validationResult.PNG = &stats
				}
				result.ValidationIssues = append([]string{}, issues...)
				validationResult.Issues = append([]string{}, issues...)
			}
		}

		if len(validationResult.Issues) > 0 {
			validationResult.Status = "failed"
		}
		if len(validationResult.Issues) > 0 || output.ValidatePNGs || textExpectationsForSection(section, prefix) != nil {
			validation = append(validation, validationResult)
		}
		pageResult.Sections = append(pageResult.Sections, result)
	}

	return pageResult, validation, nil
}

func writeJSON(path string, data any) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func writePreparedHTML(page *driver.Page, selector, path string) error {
	selectorJSON, _ := json.Marshal(selector)
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  const el = selector ? document.querySelector(selector) : document.documentElement;
	  return el ? el.outerHTML : document.documentElement.outerHTML;
	})()`, string(selectorJSON))
	var html string
	if err := page.Evaluate(script, &html); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(html), 0o644)
}

func writeInspectJSON(page *driver.Page, selector, path string) error {
	selectorJSON, _ := json.Marshal(selector)
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  const root = selector ? document.querySelector(selector) : document.documentElement;
	  const inspect = (el, depth = 0) => {
	    if (!el || depth > 8) return null;
	    const rect = el.getBoundingClientRect();
	    const style = window.getComputedStyle(el);
	    const attrs = {};
	    for (const attr of el.attributes || []) attrs[attr.name] = attr.value;
	    return {
	      tag: el.tagName ? el.tagName.toLowerCase() : '',
	      id: el.id || '',
	      class_name: el.className || '',
	      attributes: attrs,
	      text: (el.children && el.children.length === 0 ? el.textContent : '').trim().slice(0, 200),
	      bounds: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
	      computed: {
	        display: style.display,
	        position: style.position,
	        boxSizing: style.boxSizing,
	        width: style.width,
	        height: style.height,
	        margin: style.margin,
	        padding: style.padding,
	        color: style.color,
	        backgroundColor: style.backgroundColor,
	        fontFamily: style.fontFamily,
	        fontSize: style.fontSize,
	        fontWeight: style.fontWeight,
	        lineHeight: style.lineHeight,
	        border: style.border,
	        borderRadius: style.borderRadius,
	        gap: style.gap,
	        gridTemplateColumns: style.gridTemplateColumns,
	        flexDirection: style.flexDirection,
	        alignItems: style.alignItems,
	        justifyContent: style.justifyContent
	      },
	      children: Array.from(el.children || []).map((child) => inspect(child, depth + 1)).filter(Boolean)
	    };
	  };
	  return inspect(root);
	})()`, string(selectorJSON))
	var out any
	if err := page.Evaluate(script, &out); err != nil {
		return err
	}
	return writeJSON(path, out)
}

func writeMarkdown(path string, result CaptureResult) error {
	content := "# css-visual-diff Capture Report\n\n"
	content += "## Coverage Summary\n\n"
	content += fmt.Sprintf("- Total selectors: %d\n", result.Coverage.Total)
	content += fmt.Sprintf("- Original missing: %d\n", result.Coverage.OriginalMissing)
	content += fmt.Sprintf("- React missing: %d\n", result.Coverage.ReactMissing)
	content += fmt.Sprintf("- Original hidden: %d\n", result.Coverage.OriginalHidden)
	content += fmt.Sprintf("- React hidden: %d\n\n", result.Coverage.ReactHidden)
	content += formatPageResult("Original", result.Original)
	content += "\n"
	content += formatPageResult("React", result.React)
	if len(result.Validation) > 0 {
		content += "\n## Validation\n\n"
		content += "| Target | Section | Status | Issues |\n"
		content += "| --- | --- | --- | --- |\n"
		for _, v := range result.Validation {
			content += fmt.Sprintf("| %s | %s | %s | %s |\n", v.Target, v.Section, v.Status, strings.Join(v.Issues, "; "))
		}
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func formatPageResult(label string, page PageResult) string {
	content := fmt.Sprintf("## %s\n\n", label)
	content += fmt.Sprintf("URL: %s\n\n", page.URL)
	content += fmt.Sprintf("Full screenshot: %s\n\n", page.FullScreenshot)
	if page.PreparedHTML != "" {
		content += fmt.Sprintf("Prepared HTML: %s\n\n", page.PreparedHTML)
	}
	if page.InspectJSON != "" {
		content += fmt.Sprintf("Inspect JSON: %s\n\n", page.InspectJSON)
	}
	content += "| Section | Exists | Visible | Screenshot | Validation |\n"
	content += "| --- | --- | --- | --- | --- |\n"
	for _, s := range page.Sections {
		content += fmt.Sprintf("| %s | %t | %t | %s | %s |\n", s.Name, s.Exists, s.Visible, s.Screenshot, strings.Join(s.ValidationIssues, "; "))
	}
	content += "\n"
	return content
}

func rootSelectorForTarget(target config.Target) string {
	return service.RootSelectorForTarget(toServicePageTarget(target))
}

func selectorForSection(section config.SectionSpec, prefix string) string {
	selector := section.Selector
	switch prefix {
	case "original":
		if selector == "" || section.SelectorOriginal != "" {
			selector = section.SelectorOriginal
		}
	case "react":
		if selector == "" || section.SelectorReact != "" {
			selector = section.SelectorReact
		}
	}
	return selector
}

func textExpectationsForSection(section config.SectionSpec, prefix string) *config.TextExpectations {
	if prefix == "original" && section.ExpectTextOriginal != nil {
		return section.ExpectTextOriginal
	}
	if prefix == "react" && section.ExpectTextReact != nil {
		return section.ExpectTextReact
	}
	return section.ExpectText
}

func pngExpectationsForSection(section config.SectionSpec, prefix string) *config.PNGExpectations {
	if prefix == "original" && section.ExpectPNGOriginal != nil {
		return section.ExpectPNGOriginal
	}
	if prefix == "react" && section.ExpectPNGReact != nil {
		return section.ExpectPNGReact
	}
	return section.ExpectPNG
}

func validateTextExpectations(section config.SectionSpec, prefix, text string) []string {
	expectations := textExpectationsForSection(section, prefix)
	if expectations == nil {
		return nil
	}
	var issues []string
	for _, expected := range expectations.Includes {
		if !strings.Contains(text, expected) {
			issues = append(issues, fmt.Sprintf("missing expected text %q", expected))
		}
	}
	for _, forbidden := range expectations.Excludes {
		if strings.Contains(text, forbidden) {
			issues = append(issues, fmt.Sprintf("found forbidden text %q", forbidden))
		}
	}
	return issues
}

func computeCoverage(original PageResult, react PageResult) CoverageSummary {
	summary := CoverageSummary{Total: len(original.Sections)}
	for i, section := range original.Sections {
		if !section.Exists {
			summary.OriginalMissing++
		}
		if section.Exists && !section.Visible {
			summary.OriginalHidden++
		}
		if i < len(react.Sections) {
			if !react.Sections[i].Exists {
				summary.ReactMissing++
			}
			if react.Sections[i].Exists && !react.Sections[i].Visible {
				summary.ReactHidden++
			}
		}
	}
	return summary
}

func evaluateDOMCheck(page *driver.Page, selector string, out *domCheck) error {
	script := fmt.Sprintf(`(() => {
	  const el = document.querySelector(%q);
	  if (!el) return { exists: false, visible: false, bounds: null, text_start: "", text: "" };
	  const style = window.getComputedStyle(el);
	  const rect = el.getBoundingClientRect();
	  const visible = style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
	  const text = (el.textContent || '').replace(/\s+/g, ' ').trim();
	  return {
	    exists: true,
	    visible,
	    bounds: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
	    text_start: text.slice(0, 240),
	    text: text.slice(0, 4000)
	  };
	})()`, selector)
	return page.Evaluate(script, out)
}
