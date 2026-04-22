package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	geppettobootstrap "github.com/go-go-golems/geppetto/pkg/cli/bootstrap"
	geppettoengine "github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	profilebootstrap "github.com/go-go-golems/pinocchio/pkg/cmds/profilebootstrap"
)

type BootstrapOptions struct {
	ConfigFile        string
	Profile           string
	ProfileRegistries []string
}

type BootstrapResult struct {
	Parsed   *values.Values
	Resolved *profilebootstrap.ResolvedCLIEngineSettings
}

func ResolveEngineSettings(ctx context.Context, opts BootstrapOptions) (*BootstrapResult, error) {
	parsed, err := profilebootstrap.NewCLISelectionValues(profilebootstrap.CLISelectionInput{
		ConfigFile:        strings.TrimSpace(opts.ConfigFile),
		Profile:           strings.TrimSpace(opts.Profile),
		ProfileRegistries: append([]string(nil), opts.ProfileRegistries...),
	})
	if err != nil {
		return nil, err
	}

	resolved, err := profilebootstrap.ResolveCLIEngineSettings(ctx, parsed)
	if err != nil {
		return nil, err
	}

	return &BootstrapResult{
		Parsed:   parsed,
		Resolved: resolved,
	}, nil
}

func (r *BootstrapResult) Close() {
	if r == nil || r.Resolved == nil || r.Resolved.Close == nil {
		return
	}
	r.Resolved.Close()
}

func (r *BootstrapResult) BuildEngine() (geppettoengine.Engine, error) {
	if r == nil || r.Resolved == nil {
		return nil, fmt.Errorf("bootstrap result is nil")
	}
	return profilebootstrap.NewEngineFromResolvedCLIEngineSettings(r.Resolved)
}

func SelectedProfile(r *BootstrapResult) string {
	if r == nil || r.Resolved == nil || r.Resolved.ProfileSelection == nil {
		return ""
	}
	return strings.TrimSpace(r.Resolved.ProfileSelection.Profile)
}

func SelectedModel(r *BootstrapResult) string {
	if r == nil || r.Resolved == nil || r.Resolved.FinalInferenceSettings == nil || r.Resolved.FinalInferenceSettings.Chat == nil || r.Resolved.FinalInferenceSettings.Chat.Engine == nil {
		return ""
	}
	return strings.TrimSpace(*r.Resolved.FinalInferenceSettings.Chat.Engine)
}

func SelectedAPIType(r *BootstrapResult) string {
	if r == nil || r.Resolved == nil || r.Resolved.FinalInferenceSettings == nil || r.Resolved.FinalInferenceSettings.Chat == nil || r.Resolved.FinalInferenceSettings.Chat.ApiType == nil {
		return ""
	}
	return strings.TrimSpace(string(*r.Resolved.FinalInferenceSettings.Chat.ApiType))
}

func SelectedRegistry(r *BootstrapResult) string {
	if r == nil || r.Resolved == nil || r.Resolved.ResolvedEngineProfile == nil {
		return ""
	}
	return strings.TrimSpace(r.Resolved.ResolvedEngineProfile.RegistrySlug.String())
}

func WriteInferenceSettingsDebug(w io.Writer, r *BootstrapResult) error {
	if r == nil || r.Resolved == nil {
		return fmt.Errorf("bootstrap result is nil")
	}
	_, err := geppettobootstrap.HandleInferenceDebugOutput(
		w,
		profilebootstrap.BootstrapConfig(),
		r.Parsed,
		geppettobootstrap.InferenceDebugSettings{PrintInferenceSettings: true},
		r.Resolved,
		geppettobootstrap.InferenceDebugOutputOptions{
			CommandBase: r.Resolved.BaseInferenceSettings,
		},
	)
	return err
}
