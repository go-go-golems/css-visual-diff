package jsapi

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

func installSnapshotAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	snapshotPage := func(call goja.FunctionCall) goja.Value {
		page := mustUnwrapProxyBacking[pageState](vm, defaultProxyRegistry, "css-visual-diff.snapshot.page", call.Argument(0), "cvd.page")
		probes, err := unwrapSnapshotProbes(vm, call.Argument(1))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.snapshot", "array of cvd.probe() builders", call.Argument(1)))
		}
		return promiseValue(ctx, vm, "css-visual-diff.snapshot.page", func() (any, error) {
			return page.runExclusive(func() (any, error) {
				snapshot, err := service.SnapshotPage(page.page.Page(), probes)
				if err != nil {
					return nil, err
				}
				return lowerPageSnapshot(snapshot), nil
			})
		}, nil)
	}
	_ = exports.Set("snapshot", snapshotPage)
	if snapshotValue := exports.Get("snapshot"); snapshotValue != nil {
		_ = snapshotValue.ToObject(vm).Set("page", snapshotPage)
	}
}

func unwrapSnapshotProbes(vm *goja.Runtime, value goja.Value) ([]service.SnapshotProbeSpec, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("probes are required")
	}
	obj := value.ToObject(vm)
	length := int(obj.Get("length").ToInteger())
	if length < 0 {
		return nil, fmt.Errorf("invalid probe array length")
	}
	probes := make([]service.SnapshotProbeSpec, 0, length)
	for i := 0; i < length; i++ {
		probe, err := unwrapProxyBacking[probeBuilder](vm, defaultProxyRegistry, "css-visual-diff.snapshot", obj.Get(fmt.Sprintf("%d", i)), "cvd.probe")
		if err != nil {
			return nil, err
		}
		if probe.spec.Selector == "" {
			return nil, fmt.Errorf("probe %q selector is required", probe.spec.Name)
		}
		extractors := append([]service.ExtractorSpec{}, probe.specs...)
		if len(extractors) == 0 {
			extractors = []service.ExtractorSpec{{Kind: service.ExtractorExists}}
		}
		probes = append(probes, service.SnapshotProbeSpec{Name: probe.spec.Name, Selector: probe.spec.Selector, Source: probe.spec.Source, Required: probe.spec.Required, Extractors: extractors})
	}
	return probes, nil
}

func lowerPageSnapshot(snapshot service.PageSnapshot) map[string]any {
	results := make([]map[string]any, 0, len(snapshot.Results))
	for _, result := range snapshot.Results {
		entry := map[string]any{
			"name":     result.Name,
			"selector": result.Selector,
			"snapshot": lowerElementSnapshot(result.Snapshot),
		}
		if result.Source != "" {
			entry["source"] = result.Source
		}
		if result.Error != "" {
			entry["error"] = result.Error
		}
		results = append(results, entry)
	}
	return map[string]any{"results": results}
}
