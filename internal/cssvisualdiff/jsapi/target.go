package jsapi

import (
	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

type targetBuilder struct {
	target service.PageTarget
}

func installTargetAPI(vm *goja.Runtime, exports *goja.Object) {
	_ = exports.Set("target", func(call goja.FunctionCall) goja.Value {
		name := requiredStringArg(vm, "css-visual-diff.target", call.Argument(0))
		return newTargetBuilder(vm, name)
	})

	viewportFn := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(lowerViewport(viewportFromCall(vm, "css-visual-diff.viewport", call.Arguments)))
	}).ToObject(vm)
	_ = viewportFn.Set("desktop", func() map[string]any { return lowerViewport(service.Viewport{Width: 1280, Height: 720}) })
	_ = viewportFn.Set("tablet", func() map[string]any { return lowerViewport(service.Viewport{Width: 1024, Height: 768}) })
	_ = viewportFn.Set("mobile", func() map[string]any { return lowerViewport(service.Viewport{Width: 390, Height: 844}) })
	_ = exports.Set("viewport", viewportFn)
}

func newTargetBuilder(vm *goja.Runtime, name string) goja.Value {
	builder := &targetBuilder{target: service.PageTarget{Name: name}}
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.target",
		Methods: map[string]ProxyMethod{
			"url":      builder.url(vm),
			"waitMs":   builder.waitMs(vm),
			"viewport": builder.viewport(vm),
			"root":     builder.root(vm),
			"prepare":  builder.prepare(vm),
			"build":    builder.build(vm),
		},
		MethodOwners: map[string]MethodSpec{
			"selector": {Owner: "cvd.probe", Hint: "Targets describe pages. Element selectors belong to cvd.probe(...).selector(...)."},
			"styles":   {Owner: "cvd.probe", Hint: "Style extraction belongs to probes or locators, not targets."},
		},
	}, builder)
}

func (b *targetBuilder) url(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.URL = requiredStringArg(vm, "cvd.target.url", call.Argument(0))
		return receiver
	}
}

func (b *targetBuilder) waitMs(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.WaitMS = requiredNonNegativeIntArg(vm, "cvd.target.waitMs", call.Argument(0))
		return receiver
	}
}

func (b *targetBuilder) viewport(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.Viewport = viewportFromCall(vm, "cvd.target.viewport", call.Arguments)
		return receiver
	}
}

func (b *targetBuilder) root(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.RootSelector = requiredStringArg(vm, "cvd.target.root", call.Argument(0))
		return receiver
	}
}

func (b *targetBuilder) prepare(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		var raw map[string]any
		if err := decodeIntoValue(vm, call.Argument(0), &raw); err != nil {
			panic(typeMismatchError(vm, "cvd.target.prepare", "prepare object", call.Argument(0)))
		}
		prepare, err := decodePrepareSpec(raw)
		if err != nil {
			panic(vm.NewTypeError("cvd.target.prepare: %s", err.Error()))
		}
		b.target.Prepare = &prepare
		return receiver
	}
}

func (b *targetBuilder) build(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		if b.target.URL == "" {
			panic(vm.NewTypeError("cvd.target.build: url is required. Use .url(\"https://example.test\")."))
		}
		if b.target.Viewport.Width <= 0 {
			b.target.Viewport.Width = 1280
		}
		if b.target.Viewport.Height <= 0 {
			b.target.Viewport.Height = 720
		}
		return vm.ToValue(lowerPageTarget(b.target))
	}
}

func lowerPageTarget(target service.PageTarget) map[string]any {
	ret := map[string]any{
		"name":         target.Name,
		"url":          target.URL,
		"waitMs":       target.WaitMS,
		"viewport":     lowerViewport(target.Viewport),
		"rootSelector": target.RootSelector,
	}
	if target.Prepare != nil {
		ret["prepare"] = lowerServicePrepareSpec(*target.Prepare)
	}
	return ret
}

func lowerServicePrepareSpec(prepare service.PrepareSpec) map[string]any {
	return map[string]any{
		"type":             prepare.Type,
		"script":           prepare.Script,
		"scriptFile":       prepare.ScriptFile,
		"waitFor":          prepare.WaitFor,
		"waitForTimeoutMs": prepare.WaitForTimeoutMS,
		"afterWaitMs":      prepare.AfterWaitMS,
		"component":        prepare.Component,
		"props":            prepare.Props,
		"rootSelector":     prepare.RootSelector,
		"width":            prepare.Width,
		"minHeight":        prepare.MinHeight,
		"background":       prepare.Background,
	}
}
