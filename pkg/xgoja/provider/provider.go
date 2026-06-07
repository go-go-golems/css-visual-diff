package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/dsl"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/jsapi"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/verbcli"
	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
)

const PackageID = "css-visual-diff"

type verbsCommandProviderConfig struct {
	Repositories []string `json:"repositories,omitempty"`
}

func Register(registry *providerapi.ProviderRegistry) error {
	return registry.Package(PackageID,
		moduleEntry("css-visual-diff", "CSS visual diff browser and artifact APIs.", jsapi.NewLoader),
		moduleEntry("diff", "Compatibility helper module for region comparison workflows.", dsl.NewDiffLoader),
		moduleEntry("report", "Compatibility helper module for rendering review briefs.", dsl.NewReportLoader),
		providerapi.CommandSetProvider{
			Name:         "verbs",
			DefaultMount: "css-diff",
			Description:  "Run css-visual-diff workflow verbs",
			NewCommandSet: newVerbsCommandSet,
		},
	)
}

func moduleEntry(name, description string, loader func() require.ModuleLoader) providerapi.Module {
	return providerapi.Module{
		Name:        name,
		DefaultAs:   name,
		Description: description,
		NewModuleFactory: func(providerapi.ModuleSetupContext) (require.ModuleLoader, error) {
			return loader(), nil
		},
	}
}

func newVerbsCommandSet(ctx providerapi.CommandSetContext) (*providerapi.CommandSet, error) {
	cfg := verbsCommandProviderConfig{}
	if len(ctx.Config) > 0 {
		if err := json.Unmarshal(ctx.Config, &cfg); err != nil {
			return nil, fmt.Errorf("decode css-visual-diff verbs command provider config: %w", err)
		}
	}
	args := make([]string, 0, len(cfg.Repositories)*2)
	for _, repo := range cfg.Repositories {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			continue
		}
		args = append(args, "--"+verbcli.VerbRepositoryFlag, repo)
	}
	bootstrap, _, err := verbcli.DiscoverBootstrap(args)
	if err != nil {
		return nil, fmt.Errorf("discover css-visual-diff verb repositories: %w", err)
	}
	commands, err := verbcli.NewCommandsWithInvokerFactory(bootstrap, xgojaInvokerFactory(ctx))
	if err != nil {
		return nil, fmt.Errorf("build css-visual-diff verb commands: %w", err)
	}
	return &providerapi.CommandSet{
		Commands: commands,
		ParserConfig: &glazedcli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
		},
	}, nil
}

func xgojaInvokerFactory(providerCtx providerapi.CommandSetContext) verbcli.InvokerFactory {
	return func(repo verbcli.ScannedRepository, _ *jsverbs.VerbSpec) jsverbs.VerbInvoker {
		return func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
			if registry == nil {
				return nil, fmt.Errorf("css-visual-diff jsverbs registry is nil")
			}
			if providerCtx.RuntimeFactory == nil {
				return nil, fmt.Errorf("css-visual-diff xgoja runtime factory is nil")
			}
			opts := []require.Option{require.WithLoader(registry.RequireLoader())}
			if !repo.Repository.Embedded && strings.TrimSpace(repo.Repository.RootDir) != "" {
				folders := []string{repo.Repository.RootDir, filepath.Join(repo.Repository.RootDir, "node_modules")}
				parent := filepath.Dir(repo.Repository.RootDir)
				if parent != repo.Repository.RootDir {
					folders = append(folders, parent, filepath.Join(parent, "node_modules"))
				}
				opts = append(opts, require.WithGlobalFolders(folders...))
			}
			rt, err := providerCtx.RuntimeFactory.NewRuntime(ctx, opts...)
			if err != nil {
				return nil, err
			}
			defer func() { _ = rt.Close(context.Background()) }()
			return registry.InvokeInRuntime(ctx, rt, verb, parsedValues)
		}
	}
}
