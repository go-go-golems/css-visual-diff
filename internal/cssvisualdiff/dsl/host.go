package dsl

import (
	"context"
	"io/fs"

	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazedvalues "github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

type Host struct {
	registry *jsverbs.Registry
	factory  *engine.Factory
}

func NewHost() (*Host, error) {
	registry, err := jsverbs.ScanFS(embeddedScripts, "scripts")
	if err != nil {
		return nil, err
	}
	if err := RegisterSharedSections(registry); err != nil {
		return nil, err
	}

	factory, err := NewRuntimeFactory(registry)
	if err != nil {
		return nil, err
	}

	return &Host{registry: registry, factory: factory}, nil
}

func EmbeddedScriptsFS() (fs.FS, string) {
	return embeddedScripts, "scripts"
}

func RegisterSharedSections(registry *jsverbs.Registry) error {
	return registerSharedSections(registry)
}

func NewRuntimeFactory(registry *jsverbs.Registry, opts ...engine.Option) (*engine.Factory, error) {
	builder := engine.NewBuilder(opts...).
		WithRequireOptions(noderequire.WithLoader(registry.RequireLoader())).
		WithModules(engine.DefaultRegistryModule("fs")).
		WithRuntimeModuleRegistrars(newRuntimeRegistrar())
	return builder.Build()
}

func (h *Host) Commands() ([]cmds.Command, error) {
	return h.registry.CommandsWithInvoker(h.invoke)
}

func (h *Host) invoke(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *glazedvalues.Values) (interface{}, error) {
	rt, err := h.factory.NewRuntime(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	return registry.InvokeInRuntime(ctx, rt, verb, parsedValues)
}
