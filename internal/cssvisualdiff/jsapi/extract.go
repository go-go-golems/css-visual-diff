package jsapi

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

func installExtractAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	_ = exports.Set("extract", func(call goja.FunctionCall) goja.Value {
		locator := mustUnwrapProxyBacking[locatorHandle](vm, defaultProxyRegistry, "css-visual-diff.extract", call.Argument(0), "cvd.locator")
		extractors, err := unwrapExtractorList(vm, call.Argument(1))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.extract", "array of cvd.extractors.* handles", call.Argument(1)))
		}
		return promiseValue(ctx, vm, "css-visual-diff.extract", func() (any, error) {
			return locator.page.runExclusive(func() (any, error) {
				snapshot, err := service.ExtractElement(locator.page.page.Page(), locator.spec(), extractors)
				if err != nil {
					return nil, err
				}
				return lowerElementSnapshot(snapshot), nil
			})
		}, nil)
	})
}

func unwrapExtractorList(vm *goja.Runtime, value goja.Value) ([]service.ExtractorSpec, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("extractors are required")
	}
	obj := value.ToObject(vm)
	length := int(obj.Get("length").ToInteger())
	if length < 0 {
		return nil, fmt.Errorf("invalid extractor array length")
	}
	extractors := make([]service.ExtractorSpec, 0, length)
	for i := 0; i < length; i++ {
		extractor, err := unwrapProxyBacking[extractorHandle](vm, defaultProxyRegistry, "css-visual-diff.extract", obj.Get(fmt.Sprintf("%d", i)), "cvd.extractor")
		if err != nil {
			return nil, err
		}
		extractors = append(extractors, extractor.toSpec())
	}
	return extractors, nil
}

func lowerElementSnapshot(snapshot service.ElementSnapshot) map[string]any {
	out := map[string]any{"selector": snapshot.Selector}
	if snapshot.Exists != nil {
		out["exists"] = *snapshot.Exists
	}
	if snapshot.Visible != nil {
		out["visible"] = *snapshot.Visible
	}
	if snapshot.Text != "" {
		out["text"] = snapshot.Text
	}
	if snapshot.Bounds != nil {
		out["bounds"] = lowerBounds(*snapshot.Bounds)
	}
	if snapshot.Computed != nil {
		out["computed"] = snapshot.Computed
	}
	if snapshot.Attributes != nil {
		out["attributes"] = snapshot.Attributes
	}
	return out
}
