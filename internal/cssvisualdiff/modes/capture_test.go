package modes

import (
	"testing"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

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
