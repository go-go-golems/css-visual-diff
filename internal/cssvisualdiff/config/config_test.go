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
