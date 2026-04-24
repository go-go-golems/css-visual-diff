package driver

import "testing"

func TestShouldDisableChromeSandboxEnvOverride(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("CSS_VISUAL_DIFF_CHROME_NO_SANDBOX", "false")
	if shouldDisableChromeSandbox() {
		t.Fatalf("explicit CSS_VISUAL_DIFF_CHROME_NO_SANDBOX=false should override CI defaults")
	}

	t.Setenv("CSS_VISUAL_DIFF_CHROME_NO_SANDBOX", "true")
	if !shouldDisableChromeSandbox() {
		t.Fatalf("CSS_VISUAL_DIFF_CHROME_NO_SANDBOX=true should disable sandbox")
	}
}

func TestEnvBool(t *testing.T) {
	for _, value := range []string{"1", "t", "true", "y", "yes", "on", " TRUE "} {
		if !envBool(value) {
			t.Fatalf("expected %q to be true", value)
		}
	}
	for _, value := range []string{"", "0", "false", "no", "off", "anything-else"} {
		if envBool(value) {
			t.Fatalf("expected %q to be false", value)
		}
	}
}
