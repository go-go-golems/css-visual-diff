package modes

import (
	"strings"

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
