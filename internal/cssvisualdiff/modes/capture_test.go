package modes

import (
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func TestSelectorForSection(t *testing.T) {
	section := config.SectionSpec{
		Selector:         ".shared",
		SelectorOriginal: "#original",
		SelectorReact:    "#react",
	}
	if got := selectorForSection(section, "original"); got != "#original" {
		t.Fatalf("expected original override, got %q", got)
	}
	if got := selectorForSection(section, "react"); got != "#react" {
		t.Fatalf("expected react override, got %q", got)
	}
	if got := selectorForSection(config.SectionSpec{Selector: ".shared"}, "original"); got != ".shared" {
		t.Fatalf("expected shared selector, got %q", got)
	}
}

func TestValidateTextExpectations(t *testing.T) {
	section := config.SectionSpec{
		ExpectTextOriginal: &config.TextExpectations{
			Includes: []string{"Prepared Header"},
			Excludes: []string{"DesignCanvas chrome"},
		},
	}
	if issues := validateTextExpectations(section, "original", "Prepared Header Prepared Footer"); len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
	issues := validateTextExpectations(section, "original", "DesignCanvas chrome")
	if len(issues) != 2 {
		t.Fatalf("expected missing include and forbidden text issues, got %v", issues)
	}
}

func TestRootSelectorForTarget(t *testing.T) {
	if got := rootSelectorForTarget(config.Target{}); got != "" {
		t.Fatalf("expected empty selector, got %q", got)
	}
	if got := rootSelectorForTarget(config.Target{RootSelector: "#app"}); got != "#app" {
		t.Fatalf("expected explicit root selector, got %q", got)
	}
	if got := rootSelectorForTarget(config.Target{Prepare: &config.PrepareSpec{RootSelector: "#capture-root"}}); got != "#capture-root" {
		t.Fatalf("expected prepare root selector, got %q", got)
	}
	if got := rootSelectorForTarget(config.Target{RootSelector: "#app", Prepare: &config.PrepareSpec{RootSelector: "#capture-root"}}); got != "#app" {
		t.Fatalf("expected explicit root selector to win, got %q", got)
	}
}
