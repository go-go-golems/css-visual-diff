package jsapi

import (
	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

type probeBuilder struct {
	spec       service.ProbeSpec
	extractors []map[string]any
	specs      []service.ExtractorSpec
}

func installProbeAPI(vm *goja.Runtime, exports *goja.Object) {
	_ = exports.Set("probe", func(call goja.FunctionCall) goja.Value {
		name := requiredStringArg(vm, "css-visual-diff.probe", call.Argument(0))
		return newProbeBuilder(vm, name)
	})
}

func newProbeBuilder(vm *goja.Runtime, name string) goja.Value {
	builder := &probeBuilder{spec: service.ProbeSpec{Name: name}}
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.probe",
		Methods: map[string]ProxyMethod{
			"selector":   builder.selector(vm),
			"required":   builder.required(vm),
			"source":     builder.source(vm),
			"text":       builder.text(vm),
			"bounds":     builder.bounds(vm),
			"styles":     builder.styles(vm),
			"attributes": builder.attributes(vm),
			"build":      builder.build(vm),
		},
		MethodOwners: map[string]MethodSpec{
			"computedStyle": {Owner: "cvd.locator", Hint: "For probe style capture, use .styles([\"color\"]) instead."},
			"exists":        {Owner: "cvd.locator", Hint: "Probes are recipes. To query a loaded page directly, use page.locator(selector).exists()."},
			"visible":       {Owner: "cvd.locator", Hint: "Probes are recipes. To query a loaded page directly, use page.locator(selector).visible()."},
		},
	}, builder)
}

func (b *probeBuilder) selector(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.spec.Selector = requiredStringArg(vm, "cvd.probe.selector", call.Argument(0))
		return receiver
	}
}

func (b *probeBuilder) required(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.spec.Required = optionalBoolArg(call.Argument(0), true)
		return receiver
	}
}

func (b *probeBuilder) source(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.spec.Source = requiredStringArg(vm, "cvd.probe.source", call.Argument(0))
		return receiver
	}
}

func (b *probeBuilder) text(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.extractors = append(b.extractors, map[string]any{"kind": "text"})
		b.specs = append(b.specs, service.ExtractorSpec{Kind: service.ExtractorText, Text: service.TextOptions{NormalizeWhitespace: true, Trim: true}})
		return receiver
	}
}

func (b *probeBuilder) bounds(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.extractors = append(b.extractors, map[string]any{"kind": "bounds"})
		b.specs = append(b.specs, service.ExtractorSpec{Kind: service.ExtractorBounds})
		return receiver
	}
}

func (b *probeBuilder) styles(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		props, err := stringListArg(vm, "cvd.probe.styles", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "cvd.probe.styles", "array of CSS property names", call.Argument(0)))
		}
		b.spec.Props = append([]string{}, props...)
		b.extractors = append(b.extractors, map[string]any{"kind": "computedStyle", "props": props})
		b.specs = append(b.specs, service.ExtractorSpec{Kind: service.ExtractorComputedStyle, Props: props})
		return receiver
	}
}

func (b *probeBuilder) attributes(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		attrs, err := stringListArg(vm, "cvd.probe.attributes", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "cvd.probe.attributes", "array of attribute names", call.Argument(0)))
		}
		b.spec.Attributes = append([]string{}, attrs...)
		b.extractors = append(b.extractors, map[string]any{"kind": "attributes", "attributes": attrs})
		b.specs = append(b.specs, service.ExtractorSpec{Kind: service.ExtractorAttributes, Attributes: attrs})
		return receiver
	}
}

func (b *probeBuilder) build(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		if b.spec.Selector == "" {
			panic(vm.NewTypeError("cvd.probe.build: selector is required. Use .selector(\"#selector\")."))
		}
		out := map[string]any{
			"name":       b.spec.Name,
			"selector":   b.spec.Selector,
			"props":      b.spec.Props,
			"attributes": b.spec.Attributes,
			"source":     b.spec.Source,
			"required":   b.spec.Required,
			"extractors": b.extractors,
		}
		return vm.ToValue(out)
	}
}
