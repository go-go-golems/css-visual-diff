package verbcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	glazerunner "github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/stretchr/testify/require"
)

func TestNewCommandIncludesBuiltinVerbs(t *testing.T) {
	cmd, err := NewCommand(Bootstrap{Repositories: []Repository{builtinRepository()}})
	require.NoError(t, err)

	found, _, err := cmd.Find([]string{"script", "compare", "region"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "region", found.Name())
}

func TestDuplicateVerbPathsReturnError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "duplicate.js"), `
function region() { return { ok: true }; }
__verb__("region", {
  parents: ["script", "compare"],
  fields: {}
});
`)

	_, err := NewCommand(Bootstrap{Repositories: []Repository{
		builtinRepository(),
		{Name: "custom", Source: "test", RootDir: dir},
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), `duplicate jsverb path "script compare region"`)
}

func TestFilesystemRepositoryVerbExecutes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hello.js"), `
function hello(name) { return "hello " + name; }
__verb__("hello", {
  parents: ["custom"],
  output: "text",
  fields: {
    name: { argument: true, required: true }
  }
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {"name": "Manuel"},
	}))
	require.NoError(t, err)

	writerCommand, ok := commands[0].(cmds.WriterCommand)
	require.True(t, ok)
	var out bytes.Buffer
	require.NoError(t, writerCommand.RunIntoWriter(context.Background(), parsedValues, &out))
	require.Contains(t, out.String(), "hello Manuel")
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
