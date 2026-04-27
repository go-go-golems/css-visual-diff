package verbcli

import (
	"context"
	"os"
	"os/exec"
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

func TestLoadConfigRepositoriesDiscoversGitRootLocalConfigFromNestedDirectory(t *testing.T) {
	withIsolatedConfigEnvironment(t)
	root := initTempGitRepository(t)
	repoDir := filepath.Join(root, "verbs")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	writeFile(t, filepath.Join(root, LocalConfigFileName), `
verbs:
  repositories:
    - name: git-root-local
      path: ./verbs
`)
	nested := filepath.Join(root, "packages", "button")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	withChdir(t, nested)

	repos, err := loadConfigRepositories(context.Background())
	require.NoError(t, err)
	requireRepositoryRoot(t, repos, repoDir)
	requireRepositorySource(t, repos, repoDir, filepath.Join(root, LocalConfigFileName))
}

func TestLoadConfigRepositoriesDiscoversGitRootOverrideConfigFromNestedDirectory(t *testing.T) {
	withIsolatedConfigEnvironment(t)
	root := initTempGitRepository(t)
	repoDir := filepath.Join(root, "override-verbs")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	writeFile(t, filepath.Join(root, LocalOverrideConfigFileName), `
verbs:
  repositories:
    - name: git-root-override
      path: ./override-verbs
`)
	nested := filepath.Join(root, "ttmp", "ticket")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	withChdir(t, nested)

	repos, err := loadConfigRepositories(context.Background())
	require.NoError(t, err)
	requireRepositoryRoot(t, repos, repoDir)
	requireRepositorySource(t, repos, repoDir, filepath.Join(root, LocalOverrideConfigFileName))
}

func TestLoadConfigRepositoriesDiscoversWorkingDirLocalConfig(t *testing.T) {
	withIsolatedConfigEnvironment(t)
	root := initTempGitRepository(t)
	cwd := filepath.Join(root, "packages", "card")
	repoDir := filepath.Join(cwd, "verbs")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	writeFile(t, filepath.Join(cwd, LocalConfigFileName), `
verbs:
  repositories:
    - name: cwd-local
      path: ./verbs
`)
	withChdir(t, cwd)

	repos, err := loadConfigRepositories(context.Background())
	require.NoError(t, err)
	requireRepositoryRoot(t, repos, repoDir)
	requireRepositorySource(t, repos, repoDir, filepath.Join(cwd, LocalConfigFileName))
}

func TestLoadConfigRepositoriesDiscoversWorkingDirOverrideConfig(t *testing.T) {
	withIsolatedConfigEnvironment(t)
	root := initTempGitRepository(t)
	cwd := filepath.Join(root, "packages", "card")
	repoDir := filepath.Join(cwd, "private-verbs")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	writeFile(t, filepath.Join(cwd, LocalOverrideConfigFileName), `
verbs:
  repositories:
    - name: cwd-override
      path: ./private-verbs
`)
	withChdir(t, cwd)

	repos, err := loadConfigRepositories(context.Background())
	require.NoError(t, err)
	requireRepositoryRoot(t, repos, repoDir)
	requireRepositorySource(t, repos, repoDir, filepath.Join(cwd, LocalOverrideConfigFileName))
}

func TestDiscoverBootstrapDedupesLocalConfigRepositoriesByRoot(t *testing.T) {
	withIsolatedConfigEnvironment(t)
	root := initTempGitRepository(t)
	repoDir := filepath.Join(root, "verbs")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	writeFile(t, filepath.Join(root, LocalConfigFileName), `
verbs:
  repositories:
    - name: shared
      path: ./verbs
`)
	writeFile(t, filepath.Join(root, LocalOverrideConfigFileName), `
verbs:
  repositories:
    - name: duplicate
      path: ./verbs
`)
	withChdir(t, root)

	bootstrap, err := discoverBootstrap(root, nil)
	require.NoError(t, err)
	require.Equal(t, 1, countRepositoryRoot(bootstrap.Repositories, repoDir))
}

func initTempGitRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git init failed: %s", string(out))
	return root
}

func withChdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(old))
	})
}

func withIsolatedConfigEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(VerbRepositoriesEnvVar, "")
}

func requireRepositoryRoot(t *testing.T, repos []Repository, root string) {
	t.Helper()
	require.GreaterOrEqual(t, countRepositoryRoot(repos, root), 1, "expected repository root %s in %#v", root, repos)
}

func requireRepositorySource(t *testing.T, repos []Repository, root string, sourceRef string) {
	t.Helper()
	for _, repo := range repos {
		if repo.RootDir == root && repo.SourceRef == sourceRef {
			return
		}
	}
	require.Failf(t, "repository source not found", "expected root %s from %s in %#v", root, sourceRef, repos)
}

func countRepositoryRoot(repos []Repository, root string) int {
	count := 0
	for _, repo := range repos {
		if repo.RootDir == root {
			count++
		}
	}
	return count
}
