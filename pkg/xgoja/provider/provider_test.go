package provider

import (
	"context"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
)

func TestRegisterProvider(t *testing.T) {
	registry := providerapi.NewProviderRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	for _, name := range []string{"css-visual-diff", "diff", "report"} {
		mod, ok := registry.ResolveModule(PackageID, name)
		if !ok {
			t.Fatalf("missing module %s.%s", PackageID, name)
		}
		if mod.DefaultAs != name {
			t.Fatalf("default alias for %s = %q, want %q", name, mod.DefaultAs, name)
		}
	}
	provider, ok := registry.ResolveCommandSetProvider(PackageID, "verbs")
	if !ok {
		t.Fatalf("missing command provider %s.verbs", PackageID)
	}
	if provider.DefaultMount != "css-diff" {
		t.Fatalf("default mount = %q, want css-diff", provider.DefaultMount)
	}
}

func TestCSSVisualDiffLoaderInstallsExports(t *testing.T) {
	mod := resolveModule(t, "css-visual-diff")
	exports := loadModule(t, mod)
	for _, name := range []string{"target", "viewport", "probe", "extractors", "catalog", "browser"} {
		if value := exports.Get(name); value == nil || goja.IsUndefined(value) {
			t.Fatalf("missing export %q", name)
		}
	}
}

func TestVerbsCommandProviderBuildsBuiltinCommands(t *testing.T) {
	provider := resolveCommandProvider(t, "verbs")
	set, err := provider.NewCommandSet(providerapi.CommandSetContext{
		Context:   context.Background(),
		PackageID: PackageID,
		Name:      "verbs",
		Mount:     "css",
	})
	if err != nil {
		t.Fatalf("create command set: %v", err)
	}
	if set == nil || len(set.Commands) == 0 {
		t.Fatalf("expected built-in css-visual-diff verb commands")
	}
}

func resolveModule(t *testing.T, name string) providerapi.Module {
	t.Helper()
	registry := providerapi.NewProviderRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	mod, ok := registry.ResolveModule(PackageID, name)
	if !ok {
		t.Fatalf("missing module %s.%s", PackageID, name)
	}
	return mod
}

func resolveCommandProvider(t *testing.T, name string) providerapi.CommandSetProvider {
	t.Helper()
	registry := providerapi.NewProviderRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	provider, ok := registry.ResolveCommandSetProvider(PackageID, name)
	if !ok {
		t.Fatalf("missing command provider %s.%s", PackageID, name)
	}
	return provider
}

func loadModule(t *testing.T, mod providerapi.Module) *goja.Object {
	t.Helper()
	loader, err := mod.NewModuleFactory(providerapi.ModuleSetupContext{Name: mod.Name, As: mod.DefaultAs})
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}
	vm := goja.New()
	moduleObj := vm.NewObject()
	exports := vm.NewObject()
	if err := moduleObj.Set("exports", exports); err != nil {
		t.Fatalf("set exports: %v", err)
	}
	loader(vm, moduleObj)
	return exports
}
