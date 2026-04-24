package verbcli

import (
	"context"
	"fmt"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/spf13/cobra"
)

type InvokerFactory func(repo ScannedRepository, verb *jsverbs.VerbSpec) jsverbs.VerbInvoker

func NewLazyCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "verbs",
		Short:              "Run annotated css-visual-diff workflow verbs",
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bootstrap, remainingArgs, err := DiscoverBootstrap(args)
			if err != nil {
				return err
			}
			resolvedCmd, err := NewCommand(bootstrap)
			if err != nil {
				return err
			}
			adoptHelpAndOutput(cmd, resolvedCmd)
			resolvedCmd.SetArgs(remainingArgs)
			return resolvedCmd.ExecuteContext(cmd.Context())
		},
	}
}

func NewCommand(bootstrap Bootstrap) (*cobra.Command, error) {
	return newCommandWithInvokerFactory(bootstrap, runtimeInvokerFactory)
}

func newCommandWithInvokerFactory(bootstrap Bootstrap, invokers InvokerFactory) (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "verbs",
		Short: "Run annotated css-visual-diff workflow verbs",
	}

	repositories, err := ScanRepositories(bootstrap)
	if err != nil {
		return nil, err
	}
	discovered, err := CollectDiscoveredVerbs(repositories)
	if err != nil {
		return nil, err
	}
	commands, err := buildCommands(discovered, invokers)
	if err != nil {
		return nil, err
	}
	if err := glazedcli.AddCommandsToRootCommand(root, commands, nil, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{
		MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares,
	})); err != nil {
		return nil, err
	}
	return root, nil
}

func buildCommands(discovered []DiscoveredVerb, invokers InvokerFactory) ([]cmds.Command, error) {
	commands := make([]cmds.Command, 0, len(discovered))
	for _, discoveredVerb := range discovered {
		repo := discoveredVerb.Repository
		verb := discoveredVerb.Verb
		cmd, err := repo.Registry.CommandForVerbWithInvoker(verb, invokers(repo, verb))
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

func runtimeInvokerFactory(repo ScannedRepository, _ *jsverbs.VerbSpec) jsverbs.VerbInvoker {
	return func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
		factory, err := newRuntimeFactory(repo)
		if err != nil {
			return nil, err
		}
		rt, err := factory.NewRuntime(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rt.Close(context.Background()) }()

		return registry.InvokeInRuntime(ctx, rt, verb, parsedValues)
	}
}

func adoptHelpAndOutput(source *cobra.Command, target *cobra.Command) {
	if source == nil || target == nil {
		return
	}
	target.SetOut(source.OutOrStdout())
	target.SetErr(source.ErrOrStderr())
	root := source.Root()
	if root == nil {
		return
	}
	target.SetHelpFunc(root.HelpFunc())
	if usageFunc := root.UsageFunc(); usageFunc != nil {
		target.SetUsageFunc(usageFunc)
	}
	target.SetHelpTemplate(root.HelpTemplate())
	target.SetUsageTemplate(root.UsageTemplate())
}

func describeRepository(repo ScannedRepository) string {
	if repo.Repository.Embedded {
		return fmt.Sprintf("%s:%s", repo.Repository.Name, repo.Repository.EmbeddedAt)
	}
	return repo.Repository.RootDir
}
