package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

func TestInspectCommandNoArgsShowsHelp(t *testing.T) {
	cmd := newInspectCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected inspect with no args to show help without error, got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Usage:") || !strings.Contains(got, "--config") || !strings.Contains(got, "--side") {
		t.Fatalf("expected help output with usage/config/side, got:\n%s", got)
	}
}

func TestInspectArtifactCommandNoArgsShowsHelp(t *testing.T) {
	cmd := newInspectArtifactCommand("screenshot", "Capture one inspected selector as a PNG file", "png")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected screenshot with no args to show help without error, got %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Usage:") || !strings.Contains(got, "--output-file") {
		t.Fatalf("expected help output with usage/output-file, got:\n%s", got)
	}
}

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

func TestRunCommandIncludesAIReviewProfileFlags(t *testing.T) {
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
	for _, name := range []string{"profile", "profile-registries", "profile-config-file"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("expected flag %q", name)
		}
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file %s: %v", path, err)
	}
}
