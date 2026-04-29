package modes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

type CSSDiffResult struct {
	Styles []StyleResult `json:"styles"`
}

type StyleResult struct {
	Name             string        `json:"name"`
	Selector         string        `json:"selector,omitempty"`
	OriginalSelector string        `json:"original_selector,omitempty"`
	ReactSelector    string        `json:"react_selector,omitempty"`
	Original         StyleSnapshot `json:"original"`
	React            StyleSnapshot `json:"react"`
	Diffs            []StyleDiff   `json:"diffs"`
}

type StyleSnapshot = service.StyleSnapshot

type Bounds = service.Bounds

type StyleDiff struct {
	Property string `json:"property"`
	Original string `json:"original"`
	React    string `json:"react"`
}

func CSSDiff(ctx context.Context, cfg *config.Config) error {
	if !cfg.Output.WriteJSON && !cfg.Output.WriteMarkdown {
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

	originalPage, err := browser.NewPage()
	if err != nil {
		return err
	}
	defer originalPage.Close()

	reactPage, err := browser.NewPage()
	if err != nil {
		return err
	}
	defer reactPage.Close()

	if err := originalPage.SetViewport(cfg.Original.Viewport.Width, cfg.Original.Viewport.Height); err != nil {
		return err
	}
	if err := reactPage.SetViewport(cfg.React.Viewport.Width, cfg.React.Viewport.Height); err != nil {
		return err
	}

	if err := originalPage.Goto(cfg.Original.URL); err != nil {
		return err
	}
	if cfg.Original.WaitMS > 0 {
		originalPage.Wait(time.Duration(cfg.Original.WaitMS) * time.Millisecond)
	}
	if err := prepareTarget(originalPage, cfg.Original); err != nil {
		return err
	}

	if err := reactPage.Goto(cfg.React.URL); err != nil {
		return err
	}
	if cfg.React.WaitMS > 0 {
		reactPage.Wait(time.Duration(cfg.React.WaitMS) * time.Millisecond)
	}
	if err := prepareTarget(reactPage, cfg.React); err != nil {
		return err
	}

	result := CSSDiffResult{}
	for _, style := range cfg.Styles {
		origSelector := selectorForTarget(style.Selector, style.SelectorOriginal)
		reactSelector := selectorForTarget(style.Selector, style.SelectorReact)

		origSpec := toServiceStyleEvalSpec(style)
		origSpec.Selector = origSelector
		reactSpec := toServiceStyleEvalSpec(style)
		reactSpec.Selector = reactSelector

		origSnap, err := evaluateStyle(originalPage, origSpec)
		if err != nil {
			return err
		}
		reactSnap, err := evaluateStyle(reactPage, reactSpec)
		if err != nil {
			return err
		}
		diffs := buildDiffs(style.Props, origSnap, reactSnap)
		result.Styles = append(result.Styles, StyleResult{
			Name:             style.Name,
			Selector:         style.Selector,
			OriginalSelector: origSelector,
			ReactSelector:    reactSelector,
			Original:         origSnap,
			React:            reactSnap,
			Diffs:            diffs,
		})
	}

	if cfg.Output.WriteJSON {
		if err := writeJSON(filepath.Join(cfg.Output.Dir, "cssdiff.json"), result); err != nil {
			return err
		}
	}
	if cfg.Output.WriteMarkdown {
		if err := writeCSSMarkdown(filepath.Join(cfg.Output.Dir, "cssdiff.md"), result); err != nil {
			return err
		}
	}

	return nil
}

func evaluateStyle(page *driver.Page, spec service.StyleEvalSpec) (StyleSnapshot, error) {
	return service.EvaluateStyle(page, spec)
}

func buildDiffs(props []string, orig StyleSnapshot, react StyleSnapshot) []StyleDiff {
	var diffs []StyleDiff
	for _, prop := range props {
		origVal := strings.TrimSpace(orig.Computed[prop])
		reactVal := strings.TrimSpace(react.Computed[prop])
		if origVal != reactVal {
			diffs = append(diffs, StyleDiff{Property: prop, Original: origVal, React: reactVal})
		}
	}
	return diffs
}

func writeCSSMarkdown(path string, result CSSDiffResult) error {
	content := "# css-visual-diff CSS Diff Report\n\n"
	for _, s := range result.Styles {
		content += fmt.Sprintf("## %s\n\n", s.Name)
		if s.OriginalSelector != "" && s.ReactSelector != "" && s.OriginalSelector != s.ReactSelector {
			content += fmt.Sprintf("Selector (original): `%s`\n\n", s.OriginalSelector)
			content += fmt.Sprintf("Selector (react): `%s`\n\n", s.ReactSelector)
		} else if s.Selector != "" {
			content += fmt.Sprintf("Selector: `%s`\n\n", s.Selector)
		} else if s.OriginalSelector != "" {
			content += fmt.Sprintf("Selector: `%s`\n\n", s.OriginalSelector)
		}
		if !s.Original.Exists && !s.React.Exists {
			content += "Both original and react are missing this selector.\n\n"
			continue
		}
		content += "| Property | Original | React |\n"
		content += "| --- | --- | --- |\n"
		for _, diff := range s.Diffs {
			content += fmt.Sprintf("| %s | %s | %s |\n", diff.Property, diff.Original, diff.React)
		}
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func selectorForTarget(defaultSelector, override string) string {
	if override != "" {
		return override
	}
	return defaultSelector
}
