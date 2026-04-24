package modes

import (
	"strings"
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func testInspectConfig() *config.Config {
	return &config.Config{
		Original: config.Target{
			Name:         "original",
			URL:          "http://example.com/original",
			Viewport:     config.Viewport{Width: 1200, Height: 800},
			RootSelector: "#original-root",
		},
		React: config.Target{
			Name:         "react",
			URL:          "http://example.com/react",
			Viewport:     config.Viewport{Width: 1200, Height: 800},
			RootSelector: "#react-root",
		},
		Sections: []config.SectionSpec{
			{Name: "hero", SelectorOriginal: "#original-hero", SelectorReact: "#react-hero"},
			{Name: "shared", Selector: "[data-shared]"},
		},
		Styles: []config.StyleSpec{
			{Name: "button", SelectorOriginal: "#original-hero button", SelectorReact: "#react-hero button", Props: []string{"color", "border-radius"}, Attributes: []string{"class"}},
			{Name: "title", Selector: "h1", Props: []string{"font-size"}},
		},
	}
}

func TestBuildInspectRequests_StyleUsesSideSelectorAndConfiguredProps(t *testing.T) {
	cfg := testInspectConfig()
	reqs, err := BuildInspectRequests(cfg, InspectOptions{Side: "react", Style: "button"})
	if err != nil {
		t.Fatalf("BuildInspectRequests: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected one request, got %d", len(reqs))
	}
	if got, want := reqs[0].Selector, "#react-hero button"; got != want {
		t.Fatalf("selector = %q, want %q", got, want)
	}
	if got, want := strings.Join(reqs[0].Props, ","), "color,border-radius"; got != want {
		t.Fatalf("props = %q, want %q", got, want)
	}
	if got, want := strings.Join(reqs[0].Attributes, ","), "class"; got != want {
		t.Fatalf("attrs = %q, want %q", got, want)
	}
}

func TestBuildInspectRequests_SectionUsesDefaultProps(t *testing.T) {
	cfg := testInspectConfig()
	reqs, err := BuildInspectRequests(cfg, InspectOptions{Side: "original", Section: "hero"})
	if err != nil {
		t.Fatalf("BuildInspectRequests: %v", err)
	}
	if got, want := reqs[0].Selector, "#original-hero"; got != want {
		t.Fatalf("selector = %q, want %q", got, want)
	}
	if len(reqs[0].Props) == 0 {
		t.Fatalf("expected default props")
	}
}

func TestBuildInspectRequests_AllStyles(t *testing.T) {
	cfg := testInspectConfig()
	reqs, err := BuildInspectRequests(cfg, InspectOptions{Side: "react", AllStyles: true})
	if err != nil {
		t.Fatalf("BuildInspectRequests: %v", err)
	}
	if len(reqs) != 2 {
		t.Fatalf("expected two requests, got %d", len(reqs))
	}
	if got, want := reqs[0].Selector, "#react-hero button"; got != want {
		t.Fatalf("first selector = %q, want %q", got, want)
	}
	if got, want := reqs[1].Selector, "h1"; got != want {
		t.Fatalf("second selector = %q, want %q", got, want)
	}
}

func TestBuildInspectRequests_RequiresExactlyOneSelectorSource(t *testing.T) {
	cfg := testInspectConfig()
	if _, err := BuildInspectRequests(cfg, InspectOptions{Side: "react"}); err == nil {
		t.Fatalf("expected error without selector source")
	}
	if _, err := BuildInspectRequests(cfg, InspectOptions{Side: "react", Root: true, Style: "button"}); err == nil {
		t.Fatalf("expected error with multiple selector sources")
	}
}

func TestValidateInspectOptions_OutputFileRejectsBundle(t *testing.T) {
	err := validateInspectOptions(InspectOptions{Side: "react", Style: "button", Format: InspectFormatBundle, OutputFile: "button.png"})
	if err == nil {
		t.Fatalf("expected bundle output-file error")
	}
}

func TestCanonicalInspectFormatAliases(t *testing.T) {
	got, err := canonicalInspectFormat("css")
	if err != nil {
		t.Fatalf("canonicalInspectFormat: %v", err)
	}
	if got != InspectFormatCSSMarkdown {
		t.Fatalf("format = %q, want %q", got, InspectFormatCSSMarkdown)
	}
	got, err = canonicalInspectFormat("screenshot")
	if err != nil {
		t.Fatalf("canonicalInspectFormat: %v", err)
	}
	if got != InspectFormatPNG {
		t.Fatalf("format = %q, want %q", got, InspectFormatPNG)
	}
}

func TestInspectFormatRequiresExistingSelector(t *testing.T) {
	for _, format := range []string{InspectFormatBundle, InspectFormatPNG, InspectFormatHTML, InspectFormatInspectJSON} {
		if !inspectFormatRequiresExistingSelector(format) {
			t.Fatalf("format %q should require an existing selector", format)
		}
	}
	for _, format := range []string{InspectFormatCSSJSON, InspectFormatCSSMarkdown, InspectFormatMetadataJSON} {
		if inspectFormatRequiresExistingSelector(format) {
			t.Fatalf("format %q should not require an existing selector preflight", format)
		}
	}
}
