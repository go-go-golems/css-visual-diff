package jsapi

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

type collectedSelectionHandle struct {
	data service.SelectionData
}

func installCollectAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	collect := vm.NewObject()
	_ = collect.Set("selection", func(call goja.FunctionCall) goja.Value {
		locator := mustUnwrapProxyBacking[locatorHandle](vm, defaultProxyRegistry, "css-visual-diff.collect.selection", call.Argument(0), "cvd.locator")
		rawOptions := exportOptionalObject(vm, "css-visual-diff.collect.selection", call.Argument(1))
		return promiseValue(ctx, vm, "css-visual-diff.collect.selection", func() (any, error) {
			return collectFromLocator(locator, rawOptions)
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapCollectedSelection(ctx, vm, value.(service.SelectionData))
		})
	})
	_ = exports.Set("collect", collect)
}

func (l *locatorHandle) collect(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		rawOptions := exportOptionalObject(vm, "css-visual-diff.locator.collect", call.Argument(0))
		return promiseValue(ctx, vm, "css-visual-diff.locator.collect", func() (any, error) {
			return collectFromLocator(l, rawOptions)
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapCollectedSelection(ctx, vm, value.(service.SelectionData))
		})
	}
}

func collectFromLocator(locator *locatorHandle, rawOptions map[string]any) (any, error) {
	return locator.page.runExclusive(func() (any, error) {
		opts, err := decodeCollectOptions(rawOptions)
		if err != nil {
			return nil, err
		}
		return service.CollectSelection(locator.page.page.Page(), locator.spec(), opts)
	})
}

func decodeCollectOptions(raw map[string]any) (service.CollectOptions, error) {
	if raw == nil {
		return service.CollectOptions{}, nil
	}
	if inspect, ok := raw["inspect"].(string); ok {
		switch inspect {
		case service.CollectInspectMinimal, service.CollectInspectRich, service.CollectInspectDebug, "":
		default:
			return service.CollectOptions{}, fmt.Errorf("collect inspect must be one of minimal, rich, debug; got %q", inspect)
		}
	}
	if value, ok := raw["styles"]; ok {
		switch typed := value.(type) {
		case string:
			if typed == "all" {
				raw["allStyles"] = true
			} else {
				return service.CollectOptions{}, fmt.Errorf("collect styles string must be \"all\", got %q", typed)
			}
		case []any:
			raw["styleProps"] = typed
		}
	}
	if value, ok := raw["attributes"]; ok {
		switch typed := value.(type) {
		case string:
			if typed == "all" {
				raw["allAttributes"] = true
			} else {
				return service.CollectOptions{}, fmt.Errorf("collect attributes string must be \"all\", got %q", typed)
			}
		case []any:
			raw["attributes"] = typed
		}
	}
	return decodeInto[service.CollectOptions](raw)
}

func wrapCollectedSelection(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, data service.SelectionData) goja.Value {
	handle := &collectedSelectionHandle{data: data}
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.collectedSelection",
		Methods: map[string]ProxyMethod{
			"summary":    handle.summary(vm),
			"toJSON":     handle.toJSON(vm),
			"status":     handle.status(vm),
			"bounds":     handle.bounds(vm),
			"text":       handle.text(vm),
			"styles":     handle.styles(vm),
			"attributes": handle.attributes(vm),
		},
		Properties: map[string]ProxyProperty{
			"screenshot": handle.screenshot(ctx, vm),
		},
		MethodOwners: map[string]MethodSpec{
			"diff":     {Owner: "cvd.selectionComparison", Hint: "Compare collected selections with cvd.compare.selections(left, right)."},
			"selector": {Owner: "cvd.locator", Hint: "Collected selections are immutable data. Create a new page.locator(selector) for a live selector handle."},
		},
	}, handle)
}

func (h *collectedSelectionHandle) summary(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return vm.ToValue(map[string]any{
			"schemaVersion": h.data.SchemaVersion,
			"name":          h.data.Name,
			"url":           h.data.URL,
			"selector":      h.data.Selector,
			"source":        h.data.Source,
			"exists":        h.data.Exists,
			"visible":       h.data.Visible,
			"bounds":        lowerBoundsPtr(h.data.Bounds),
			"text":          h.data.Text,
		})
	}
}

func (h *collectedSelectionHandle) toJSON(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value { return vm.ToValue(lowerJSON(h.data)) }
}

func (h *collectedSelectionHandle) status(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return vm.ToValue(lowerSelectorStatus(h.data.Status))
	}
}

func (h *collectedSelectionHandle) bounds(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return vm.ToValue(lowerBoundsPtr(h.data.Bounds))
	}
}

func (h *collectedSelectionHandle) text(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value { return vm.ToValue(h.data.Text) }
}

func (h *collectedSelectionHandle) styles(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		props, err := stringListArg(vm, "css-visual-diff.collectedSelection.styles", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.collectedSelection.styles", "array of CSS property names", call.Argument(0)))
		}
		return vm.ToValue(filterStringMap(h.data.ComputedStyles, props))
	}
}

func (h *collectedSelectionHandle) attributes(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		names, err := stringListArg(vm, "css-visual-diff.collectedSelection.attributes", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.collectedSelection.attributes", "array of attribute names", call.Argument(0)))
		}
		return vm.ToValue(filterStringMap(h.data.Attributes, names))
	}
}

func (h *collectedSelectionHandle) screenshot(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("write", func(path string) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.collectedSelection.screenshot.write", func() (any, error) {
				if h.data.Screenshot == nil || h.data.Screenshot.Path == "" {
					return nil, fmt.Errorf("collected selection has no screenshot data")
				}
				// Phase 4 collection only carries descriptors. Screenshot capture/writing is completed by later artifact phases.
				return nil, fmt.Errorf("collected selection screenshot write is not available until screenshot collection is enabled")
			}, nil)
		})
		return obj
	}
}

func exportOptionalObject(vm *goja.Runtime, operation string, value goja.Value) map[string]any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return map[string]any{}
	}
	exported, ok := value.Export().(map[string]any)
	if !ok {
		panic(typeMismatchError(vm, operation, "options object", value))
	}
	return exported
}

func lowerJSON(value any) any {
	bytes, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var out any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return value
	}
	return out
}

func lowerBoundsPtr(bounds *service.Bounds) any {
	if bounds == nil {
		return nil
	}
	return lowerBounds(*bounds)
}

func filterStringMap(values map[string]string, keys []string) map[string]string {
	if len(keys) == 0 {
		out := make(map[string]string, len(values))
		for k, v := range values {
			out[k] = v
		}
		return out
	}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := values[key]; ok {
			out[key] = value
		}
	}
	return out
}
