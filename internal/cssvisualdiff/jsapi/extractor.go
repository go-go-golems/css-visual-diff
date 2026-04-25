package jsapi

import (
	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

type extractorHandle struct {
	kind       string
	props      []string
	attributes []string
}

func installExtractorAPI(vm *goja.Runtime, exports *goja.Object) {
	extractors := vm.NewObject()
	_ = extractors.Set("exists", func() goja.Value { return newExtractorHandle(vm, extractorHandle{kind: "exists"}) })
	_ = extractors.Set("visible", func() goja.Value { return newExtractorHandle(vm, extractorHandle{kind: "visible"}) })
	_ = extractors.Set("text", func() goja.Value { return newExtractorHandle(vm, extractorHandle{kind: "text"}) })
	_ = extractors.Set("bounds", func() goja.Value { return newExtractorHandle(vm, extractorHandle{kind: "bounds"}) })
	_ = extractors.Set("computedStyle", func(call goja.FunctionCall) goja.Value {
		props, err := stringListArg(vm, "cvd.extractors.computedStyle", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "cvd.extractors.computedStyle", "array of CSS property names", call.Argument(0)))
		}
		return newExtractorHandle(vm, extractorHandle{kind: "computedStyle", props: props})
	})
	_ = extractors.Set("attributes", func(call goja.FunctionCall) goja.Value {
		attrs, err := stringListArg(vm, "cvd.extractors.attributes", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "cvd.extractors.attributes", "array of attribute names", call.Argument(0)))
		}
		return newExtractorHandle(vm, extractorHandle{kind: "attributes", attributes: attrs})
	})
	_ = exports.Set("extractors", extractors)
}

func newExtractorHandle(vm *goja.Runtime, extractor extractorHandle) goja.Value {
	backing := &extractor
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.extractor",
		Methods: map[string]ProxyMethod{
			"build": func(call goja.FunctionCall, receiver goja.Value) goja.Value {
				return vm.ToValue(backing.toPlain())
			},
		},
		MethodOwners: map[string]MethodSpec{
			"selector": {Owner: "cvd.probe", Hint: "Extractors describe what to read. Selectors belong to probes or locators."},
			"styles":   {Owner: "cvd.probe", Hint: "For a standalone extractor use cvd.extractors.computedStyle([...])."},
		},
	}, backing)
}

func (e extractorHandle) toSpec() service.ExtractorSpec {
	return service.ExtractorSpec{Kind: service.ExtractorKind(e.kind), Props: e.props, Attributes: e.attributes, Text: service.TextOptions{NormalizeWhitespace: true, Trim: true}}
}

func (e extractorHandle) toPlain() map[string]any {
	out := map[string]any{"kind": e.kind}
	if len(e.props) > 0 {
		out["props"] = e.props
	}
	if len(e.attributes) > 0 {
		out["attributes"] = e.attributes
	}
	return out
}
