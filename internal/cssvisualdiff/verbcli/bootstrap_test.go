package verbcli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepositoriesFromArgsOnlyConsumesBootstrapPrefix(t *testing.T) {
	cwd := t.TempDir()
	repoDir := filepath.Join(cwd, "verbs")
	writeFile(t, filepath.Join(repoDir, "hello.js"), `function hello() { return "ok"; }`)

	repos, remaining, err := repositoriesFromArgs([]string{
		"--repository", repoDir,
		"custom", "run",
		"--repository", "verb-level-value",
	}, cwd)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	require.Equal(t, repoDir, repos[0].RootDir)
	require.Equal(t, []string{"custom", "run", "--repository", "verb-level-value"}, remaining)
}

func TestRepositoriesFromArgsHonorsDoubleDash(t *testing.T) {
	cwd := t.TempDir()
	repoDir := filepath.Join(cwd, "verbs")
	writeFile(t, filepath.Join(repoDir, "hello.js"), `function hello() { return "ok"; }`)

	repos, remaining, err := repositoriesFromArgs([]string{
		"--repository", repoDir,
		"--",
		"custom", "run", "--repository", "verb-level-value",
	}, cwd)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	require.Equal(t, []string{"custom", "run", "--repository", "verb-level-value"}, remaining)
}
