package modes

import (
	"strings"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func TestBuildDirectReactGlobalScript(t *testing.T) {
	script, err := buildDirectReactGlobalScript(&config.PrepareSpec{
		Component:    "PPXDesktop",
		Props:        map[string]any{"page": "shows"},
		RootSelector: "#capture-root",
		Width:        920,
		MinHeight:    1400,
		Background:   "#fff",
	})
	if err != nil {
		t.Fatalf("buildDirectReactGlobalScript: %v", err)
	}

	for _, want := range []string{
		`const componentName = "PPXDesktop"`,
		`"page":"shows"`,
		`const rootSelector = "#capture-root"`,
		`const width = 920`,
		`const minHeight = 1400`,
		`window.ReactDOM.createRoot`,
		`window.React.createElement(Component, props)`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("expected generated script to contain %q\nscript:\n%s", want, script)
		}
	}
}

func TestBuildDirectReactGlobalScript_ValidatesRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		prepare config.PrepareSpec
	}{
		{
			name: "missing component",
			prepare: config.PrepareSpec{
				RootSelector: "#capture-root",
				Width:        920,
			},
		},
		{
			name: "missing root selector",
			prepare: config.PrepareSpec{
				Component: "PPXDesktop",
				Width:     920,
			},
		},
		{
			name: "missing width",
			prepare: config.PrepareSpec{
				Component:    "PPXDesktop",
				RootSelector: "#capture-root",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildDirectReactGlobalScript(&tt.prepare); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
