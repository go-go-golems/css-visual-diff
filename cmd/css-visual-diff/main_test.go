package main

import "testing"

func TestNewLLMReviewCommandIncludesProfileFlags(t *testing.T) {
	cmd := newLLMReviewCommand()
	for _, name := range []string{"profile", "profile-registries", "config-file", "question", "print-inference-settings"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("expected flag %q", name)
		}
	}
}
