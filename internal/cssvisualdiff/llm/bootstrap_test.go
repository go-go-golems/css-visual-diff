package llm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveEngineSettingsLoadsSelectedPinocchioProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("PINOCCHIO_PROFILE", "")
	t.Setenv("PINOCCHIO_PROFILE_REGISTRIES", "")

	registryPath := filepath.Join(tmp, "profiles.yaml")
	require.NoError(t, os.WriteFile(registryPath, []byte(`slug: local-css-visual-diff
profiles:
  openai-fast:
    slug: openai-fast
    inference_settings:
      chat:
        api_type: openai
        engine: gpt-4.1-mini
  claude-fast:
    slug: claude-fast
    inference_settings:
      chat:
        api_type: claude
        engine: claude-sonnet-4-20250514
`), 0o644))

	first, err := ResolveEngineSettings(context.Background(), BootstrapOptions{
		Profile:           "openai-fast",
		ProfileRegistries: []string{registryPath},
	})
	require.NoError(t, err)
	defer first.Close()

	require.Equal(t, "openai-fast", SelectedProfile(first))
	require.Equal(t, "gpt-4.1-mini", SelectedModel(first))
	require.Equal(t, "openai", SelectedAPIType(first))
	require.Equal(t, "local-css-visual-diff", SelectedRegistry(first))
	require.NotNil(t, first.Resolved)
	require.NotNil(t, first.Resolved.BaseInferenceSettings)
	require.NotNil(t, first.Resolved.FinalInferenceSettings)
	require.NotNil(t, first.Resolved.ResolvedEngineProfile)

	second, err := ResolveEngineSettings(context.Background(), BootstrapOptions{
		Profile:           "claude-fast",
		ProfileRegistries: []string{registryPath},
	})
	require.NoError(t, err)
	defer second.Close()

	require.Equal(t, "claude-fast", SelectedProfile(second))
	require.Equal(t, "claude-sonnet-4-20250514", SelectedModel(second))
	require.Equal(t, "claude", SelectedAPIType(second))
}
