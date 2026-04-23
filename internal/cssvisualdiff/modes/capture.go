package modes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

type CaptureResult struct {
	Original PageResult      `json:"original"`
	React    PageResult      `json:"react"`
	Coverage CoverageSummary `json:"coverage"`
}

type PageResult struct {
	Name           string          `json:"name"`
	URL            string          `json:"url"`
	FullScreenshot string          `json:"full_screenshot"`
	Sections       []SectionResult `json:"sections"`
}

type SectionResult struct {
	Name       string `json:"name"`
	Selector   string `json:"selector"`
	Exists     bool   `json:"exists"`
	Visible    bool   `json:"visible"`
	Screenshot string `json:"screenshot"`
}

type CoverageSummary struct {
	Total           int `json:"total"`
	OriginalMissing int `json:"original_missing"`
	ReactMissing    int `json:"react_missing"`
	OriginalHidden  int `json:"original_hidden"`
	ReactHidden     int `json:"react_hidden"`
}

type domCheck struct {
	Exists  bool `json:"exists"`
	Visible bool `json:"visible"`
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

	original, err := captureTarget(browser, cfg.Original, cfg.Sections, cfg.Output.Dir, "original")
	if err != nil {
		return err
	}

	react, err := captureTarget(browser, cfg.React, cfg.Sections, cfg.Output.Dir, "react")
	if err != nil {
		return err
	}

	result := CaptureResult{Original: original, React: react}
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

func captureTarget(browser *driver.Browser, target config.Target, sections []config.SectionSpec, outDir, prefix string) (PageResult, error) {
	page, err := browser.NewPage()
	if err != nil {
		return PageResult{}, err
	}
	defer page.Close()

	if err := page.SetViewport(target.Viewport.Width, target.Viewport.Height); err != nil {
		return PageResult{}, err
	}

	if err := page.Goto(target.URL); err != nil {
		return PageResult{}, err
	}

	if target.WaitMS > 0 {
		page.Wait(time.Duration(target.WaitMS) * time.Millisecond)
	}
	if err := prepareTarget(page, target); err != nil {
		return PageResult{}, err
	}

	pageResult := PageResult{Name: target.Name, URL: target.URL}
	fullPath := filepath.Join(outDir, fmt.Sprintf("%s-full.png", prefix))
	if err := page.FullScreenshot(fullPath); err != nil {
		return PageResult{}, err
	}
	pageResult.FullScreenshot = fullPath

	for _, section := range sections {
		selector := section.Selector
		if selector == "" {
			if prefix == "original" {
				selector = section.SelectorOriginal
			} else if prefix == "react" {
				selector = section.SelectorReact
			}
		} else {
			if prefix == "original" && section.SelectorOriginal != "" {
				selector = section.SelectorOriginal
			}
			if prefix == "react" && section.SelectorReact != "" {
				selector = section.SelectorReact
			}
		}

		check := domCheck{}
		err := evaluateDOMCheck(page, selector, &check)
		if err != nil {
			return PageResult{}, err
		}
		result := SectionResult{
			Name:     section.Name,
			Selector: selector,
			Exists:   check.Exists,
			Visible:  check.Visible,
		}
		if check.Exists {
			screenshotPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.png", prefix, section.Name))
			if err := page.Screenshot(selector, screenshotPath); err != nil {
				return PageResult{}, err
			}
			result.Screenshot = screenshotPath
		}
		pageResult.Sections = append(pageResult.Sections, result)
	}

	return pageResult, nil
}

func writeJSON(path string, data any) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
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
	return os.WriteFile(path, []byte(content), 0o644)
}

func formatPageResult(label string, page PageResult) string {
	content := fmt.Sprintf("## %s\n\n", label)
	content += fmt.Sprintf("URL: %s\n\n", page.URL)
	content += fmt.Sprintf("Full screenshot: %s\n\n", page.FullScreenshot)
	content += "| Section | Exists | Visible | Screenshot |\n"
	content += "| --- | --- | --- | --- |\n"
	for _, s := range page.Sections {
		content += fmt.Sprintf("| %s | %t | %t | %s |\n", s.Name, s.Exists, s.Visible, s.Screenshot)
	}
	content += "\n"
	return content
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
	  if (!el) return { exists: false, visible: false };
	  const style = window.getComputedStyle(el);
	  const rect = el.getBoundingClientRect();
	  const visible = style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
	  return { exists: true, visible };
	})()`, selector)
	return page.Evaluate(script, out)
}
