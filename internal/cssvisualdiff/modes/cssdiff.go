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

type StyleSnapshot struct {
	Exists     bool              `json:"exists"`
	Computed   map[string]string `json:"computed"`
	Bounds     *Bounds           `json:"bounds,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type StyleDiff struct {
	Property string `json:"property"`
	Original string `json:"original"`
	React    string `json:"react"`
}

type styleEvalResult struct {
	Exists     bool              `json:"exists"`
	Computed   map[string]string `json:"computed"`
	Bounds     *Bounds           `json:"bounds"`
	Attributes map[string]string `json:"attributes"`
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

	if err := reactPage.Goto(cfg.React.URL); err != nil {
		return err
	}
	if cfg.React.WaitMS > 0 {
		reactPage.Wait(time.Duration(cfg.React.WaitMS) * time.Millisecond)
	}

	result := CSSDiffResult{}
	for _, style := range cfg.Styles {
		origSelector := selectorForTarget(style.Selector, style.SelectorOriginal)
		reactSelector := selectorForTarget(style.Selector, style.SelectorReact)

		origSpec := style
		origSpec.Selector = origSelector
		reactSpec := style
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

func evaluateStyle(page *driver.Page, spec config.StyleSpec) (StyleSnapshot, error) {
	propsJSON, _ := json.Marshal(spec.Props)
	attrsJSON, _ := json.Marshal(spec.Attributes)
	script := fmt.Sprintf(`(() => {
	  const props = %s;
	  const attrs = %s;
	  const el = document.querySelector(%q);
	  if (!el) return { exists: false, computed: {}, bounds: null, attributes: {} };
	  const style = window.getComputedStyle(el);
	  const computed = {};
	  props.forEach((p) => {
	    computed[p] = style.getPropertyValue(p) || style[p] || "";
	  });
	  let bounds = null;
	  if (%t) {
	    const rect = el.getBoundingClientRect();
	    bounds = { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	  }
	  const attributes = {};
	  attrs.forEach((a) => {
	    attributes[a] = el.getAttribute(a);
	  });
	  return { exists: true, computed, bounds, attributes };
	})()`, string(propsJSON), string(attrsJSON), spec.Selector, spec.IncludeBounds)

	out := styleEvalResult{}
	if err := page.Evaluate(script, &out); err != nil {
		return StyleSnapshot{}, err
	}

	return StyleSnapshot{
		Exists:     out.Exists,
		Computed:   out.Computed,
		Bounds:     out.Bounds,
		Attributes: out.Attributes,
	}, nil
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
