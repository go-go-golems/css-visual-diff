package dsl

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

func registerCVDModule(ctx *engine.RuntimeModuleContext, reg *noderequire.Registry) {
	reg.RegisterNativeModule("css-visual-diff", func(vm *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		_ = exports.Set("browser", func(call goja.FunctionCall) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.browser", func() (any, error) {
				return service.NewBrowserService(ctx.Context)
			}, func(vm *goja.Runtime, value any) goja.Value {
				return wrapBrowser(ctx, vm, value.(*service.BrowserService))
			})
		})
	})
}

func promiseValue(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, op string, work func() (any, error), wrap func(*goja.Runtime, any) goja.Value) goja.Value {
	promise, resolve, reject := vm.NewPromise()
	if ctx == nil || ctx.Owner == nil {
		panic(vm.NewGoError(fmt.Errorf("%s requires runtime owner", op)))
	}
	go func() {
		value, err := work()
		_ = ctx.Owner.Post(ctx.Context, op, func(context.Context, *goja.Runtime) {
			if err != nil {
				_ = reject(vm.NewGoError(err))
				return
			}
			if wrap != nil {
				_ = resolve(wrap(vm, value))
				return
			}
			_ = resolve(vm.ToValue(value))
		})
	}()
	return vm.ToValue(promise)
}

func wrapBrowser(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, browser *service.BrowserService) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("newPage", func(call goja.FunctionCall) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.browser.newPage", func() (any, error) {
			page, err := browser.NewPage()
			if err != nil {
				return nil, err
			}
			return newPageState(page, config.Target{}), nil
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapPage(ctx, vm, value.(*pageState))
		})
	})
	_ = obj.Set("page", func(rawURL string, rawOptions map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.browser.page", func() (any, error) {
			opts, err := decodeInto[pageOptions](rawOptions)
			if err != nil {
				return nil, err
			}
			page, err := browser.NewPage()
			if err != nil {
				return nil, err
			}
			target := opts.toTarget(rawURL)
			if err := service.LoadAndPreparePage(page.Page(), target); err != nil {
				page.Close()
				return nil, err
			}
			return newPageState(page, target), nil
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapPage(ctx, vm, value.(*pageState))
		})
	})
	_ = obj.Set("close", func() goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.browser.close", func() (any, error) {
			browser.Close()
			return nil, nil
		}, nil)
	})
	return obj
}

type pageState struct {
	page   *service.PageService
	target config.Target
}

func newPageState(page *service.PageService, target config.Target) *pageState {
	return &pageState{page: page, target: target}
}

func wrapPage(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, state *pageState) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("prepare", func(raw map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.prepare", func() (any, error) {
			prepare, err := decodePrepareSpec(raw)
			if err != nil {
				return nil, err
			}
			state.target.Prepare = &prepare
			return nil, service.PrepareTarget(state.page.Page(), state.target)
		}, nil)
	})
	_ = obj.Set("preflight", func(raw []map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.preflight", func() (any, error) {
			probes, err := decodeProbes(raw)
			if err != nil {
				return nil, err
			}
			return service.PreflightProbes(state.page.Page(), probes)
		}, nil)
	})
	_ = obj.Set("inspectAll", func(rawProbes []map[string]any, rawOptions map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.inspectAll", func() (any, error) {
			probes, err := decodeInspectRequests(rawProbes)
			if err != nil {
				return nil, err
			}
			opts, err := decodeInto[inspectAllOptionsInput](rawOptions)
			if err != nil {
				return nil, err
			}
			format := opts.Format
			if format == "" {
				format = opts.Artifacts
			}
			if format == "" {
				format = service.InspectFormatBundle
			}
			return service.InspectPreparedPage(state.page.Page(), state.target, "script", probes, service.InspectAllOptions{OutDir: opts.OutDir, Format: format})
		}, nil)
	})
	_ = obj.Set("close", func() goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.close", func() (any, error) {
			state.page.Close()
			return nil, nil
		}, nil)
	})
	return obj
}

type pageOptions struct {
	Viewport config.Viewport `json:"viewport"`
	WaitMS   int             `json:"waitMs"`
	Name     string          `json:"name"`
}

func (o pageOptions) toTarget(url string) config.Target {
	viewport := o.Viewport
	if viewport.Width <= 0 {
		viewport.Width = 1280
	}
	if viewport.Height <= 0 {
		viewport.Height = 720
	}
	name := o.Name
	if name == "" {
		name = "script"
	}
	return config.Target{Name: name, URL: url, WaitMS: o.WaitMS, Viewport: viewport}
}

type inspectAllOptionsInput struct {
	OutDir    string `json:"outDir"`
	Format    string `json:"format"`
	Artifacts string `json:"artifacts"`
}

type prepareSpecInput struct {
	Type             string         `json:"type"`
	Script           string         `json:"script"`
	ScriptFile       string         `json:"scriptFile"`
	WaitFor          string         `json:"waitFor"`
	WaitForTimeoutMS int            `json:"waitForTimeoutMs"`
	AfterWaitMS      int            `json:"afterWaitMs"`
	Component        string         `json:"component"`
	Props            map[string]any `json:"props"`
	RootSelector     string         `json:"rootSelector"`
	Width            int            `json:"width"`
	MinHeight        int            `json:"minHeight"`
	Background       string         `json:"background"`
}

func decodePrepareSpec(raw map[string]any) (config.PrepareSpec, error) {
	input, err := decodeInto[prepareSpecInput](raw)
	if err != nil {
		return config.PrepareSpec{}, err
	}
	prepareType := input.Type
	if prepareType == "directReactGlobal" {
		prepareType = "direct-react-global"
	}
	return config.PrepareSpec{
		Type:             prepareType,
		Script:           input.Script,
		ScriptFile:       input.ScriptFile,
		WaitFor:          input.WaitFor,
		WaitForTimeoutMS: input.WaitForTimeoutMS,
		AfterWaitMS:      input.AfterWaitMS,
		Component:        input.Component,
		Props:            input.Props,
		RootSelector:     input.RootSelector,
		Width:            input.Width,
		MinHeight:        input.MinHeight,
		Background:       input.Background,
	}, nil
}

type probeInput struct {
	Name       string   `json:"name"`
	Selector   string   `json:"selector"`
	Props      []string `json:"props"`
	Attributes []string `json:"attrs"`
	Source     string   `json:"source"`
	Required   bool     `json:"required"`
}

func decodeProbes(raw []map[string]any) ([]service.ProbeSpec, error) {
	inputs, err := decodeInto[[]probeInput](raw)
	if err != nil {
		return nil, err
	}
	ret := make([]service.ProbeSpec, 0, len(inputs))
	for _, input := range inputs {
		ret = append(ret, service.ProbeSpec{Name: input.Name, Selector: input.Selector, Props: input.Props, Attributes: input.Attributes, Source: input.Source, Required: input.Required})
	}
	return ret, nil
}

func decodeInspectRequests(raw []map[string]any) ([]service.InspectRequest, error) {
	inputs, err := decodeInto[[]probeInput](raw)
	if err != nil {
		return nil, err
	}
	ret := make([]service.InspectRequest, 0, len(inputs))
	for _, input := range inputs {
		name := input.Name
		if name == "" {
			name = service.SanitizeName(input.Selector)
		}
		ret = append(ret, service.InspectRequest{Name: name, Selector: input.Selector, Props: input.Props, Attributes: input.Attributes, Source: input.Source})
	}
	return ret, nil
}
