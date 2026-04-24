package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidate_AllowsPerTargetSelectors(t *testing.T) {
	y := `
metadata:
  slug: demo
original:
  name: original
  url: http://example.com/original
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
react:
  name: react
  url: http://example.com/react
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
sections:
  - name: hero
    selector_original: "#hero"
    selector_react: ".Hero"
styles:
  - name: hero-title
    selector_original: "#hero h1"
    selector_react: ".Hero h1"
    props: [font-size]
output:
  dir: /tmp/out
  write_json: true
  write_markdown: true
  write_pngs: true
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(y), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validate ok, got %v", err)
	}
}

func TestValidate_AllowsDirectReactGlobalPrepare(t *testing.T) {
	y := `
metadata:
  slug: demo
original:
  name: original
  url: http://example.com/original
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
  prepare:
    type: direct-react-global
    component: PPXDesktop
    props: { page: shows }
    root_selector: "#capture-root"
    width: 920
react:
  name: react
  url: http://example.com/react
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output:
  dir: /tmp/out
  write_json: true
  write_markdown: true
  write_pngs: true
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(y), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validate ok, got %v", err)
	}
	if cfg.Original.Prepare == nil {
		t.Fatalf("expected original prepare spec")
	}
	if got := cfg.Original.Prepare.Props["page"]; got != "shows" {
		t.Fatalf("expected props.page shows, got %#v", got)
	}
}

func TestValidate_AllowsScriptPrepare(t *testing.T) {
	y := `
metadata:
  slug: demo
original:
  name: original
  url: http://example.com/original
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
  prepare:
    type: script
    wait_for: "window.ready"
    script: "document.body.innerHTML = '<main>prepared</main>'"
react:
  name: react
  url: http://example.com/react
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output:
  dir: /tmp/out
  write_json: true
  write_markdown: true
  write_pngs: true
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(y), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validate ok, got %v", err)
	}
}

func TestValidate_RejectsInvalidPrepareSpecs(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "direct react global without component",
			yaml: `
metadata: { slug: demo }
original:
  name: original
  url: http://example.com/original
  viewport: { width: 1280, height: 720 }
  prepare:
    type: direct-react-global
    root_selector: "#capture-root"
    width: 920
react:
  name: react
  url: http://example.com/react
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output: { dir: /tmp/out }
`,
		},
		{
			name: "direct react global without width",
			yaml: `
metadata: { slug: demo }
original:
  name: original
  url: http://example.com/original
  viewport: { width: 1280, height: 720 }
  prepare:
    type: direct-react-global
    component: PPXDesktop
    root_selector: "#capture-root"
react:
  name: react
  url: http://example.com/react
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output: { dir: /tmp/out }
`,
		},
		{
			name: "script without body",
			yaml: `
metadata: { slug: demo }
original:
  name: original
  url: http://example.com/original
  viewport: { width: 1280, height: 720 }
  prepare:
    type: script
react:
  name: react
  url: http://example.com/react
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output: { dir: /tmp/out }
`,
		},
		{
			name: "unknown type",
			yaml: `
metadata: { slug: demo }
original:
  name: original
  url: http://example.com/original
  viewport: { width: 1280, height: 720 }
  prepare:
    type: mystery
react:
  name: react
  url: http://example.com/react
  viewport: { width: 1280, height: 720 }
sections: []
styles: []
output: { dir: /tmp/out }
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			if err := yaml.Unmarshal([]byte(tt.yaml), &cfg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected validate error, got nil")
			}
		})
	}
}

func TestValidate_RequiresBothPerTargetSelectorsWhenSelectorOmitted(t *testing.T) {
	y := `
metadata:
  slug: demo
original:
  name: original
  url: http://example.com/original
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
react:
  name: react
  url: http://example.com/react
  wait_ms: 0
  viewport: { width: 1280, height: 720 }
sections:
  - name: hero
    selector_original: "#hero"
styles: []
output:
  dir: /tmp/out
  write_json: true
  write_markdown: true
  write_pngs: true
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(y), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validate error, got nil")
	}
}
