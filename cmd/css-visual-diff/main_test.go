package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

func TestNewLLMReviewCommandIncludesProfileFlags(t *testing.T) {
	cmd := newLLMReviewCommand()
	for _, name := range []string{"profile", "profile-registries", "config-file", "question", "print-inference-settings"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("expected flag %q", name)
		}
	}
}

func TestRunCommandDryRunDecodesConfigFlag(t *testing.T) {
	runCmd, err := NewRunCommand()
	if err != nil {
		t.Fatal(err)
	}
	cmd, err := cli.BuildCobraCommand(runCmd, cli.WithParserConfig(cli.CobraParserConfig{
		ShortHelpSections: []string{schema.DefaultSlug},
		MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
	}))
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	writeTestFile(t, configPath, `
metadata:
  slug: dry-run-test
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
sections: []
styles: []
output:
  dir: ./out
  write_json: true
  write_markdown: true
  write_pngs: true
modes: [capture]
`)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--dry-run", "--modes", "capture", "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected dry-run command to decode --config and succeed, got %v\noutput:\n%s", err, out.String())
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file %s: %v", path, err)
	}
}
