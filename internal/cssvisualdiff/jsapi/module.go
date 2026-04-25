package jsapi

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

// Register installs the native require("css-visual-diff") module into a goja require registry.
func Register(ctx *engine.RuntimeModuleContext, reg *noderequire.Registry) {
	reg.RegisterNativeModule("css-visual-diff", func(vm *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		installCVDErrorClasses(vm, exports)
		installTargetAPI(vm, exports)
		installProbeAPI(vm, exports)
		installExtractorAPI(vm, exports)
		installExtractAPI(ctx, vm, exports)
		installSnapshotAPI(ctx, vm, exports)
		_ = exports.Set("catalog", func(raw map[string]any) (*goja.Object, error) {
			catalog, err := newCatalogFromJS(raw)
			if err != nil {
				return nil, err
			}
			return wrapCatalog(ctx, vm, catalog), nil
		})
		_ = exports.Set("loadConfig", func(path string) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.loadConfig", func() (any, error) {
				return config.Load(path)
			}, func(vm *goja.Runtime, value any) goja.Value {
				return vm.ToValue(lowerConfig(value.(*config.Config)))
			})
		})
		_ = exports.Set("browser", func(call goja.FunctionCall) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.browser", func() (any, error) {
				return service.NewBrowserService(ctx.Context)
			}, func(vm *goja.Runtime, value any) goja.Value {
				return wrapBrowser(ctx, vm, value.(*service.BrowserService))
			})
		})
	})
}

func installCVDErrorClasses(vm *goja.Runtime, exports *goja.Object) {
	value, err := vm.RunString(`(function(exports) {
class CvdError extends Error {
  constructor(message, code = "CVD_ERROR", details = undefined) {
    super(message);
    this.name = this.constructor.name;
    this.code = code;
    if (details !== undefined) this.details = details;
  }
}
class SelectorError extends CvdError {}
class PrepareError extends CvdError {}
class BrowserError extends CvdError {}
class ArtifactError extends CvdError {}
exports.CvdError = CvdError;
exports.SelectorError = SelectorError;
exports.PrepareError = PrepareError;
exports.BrowserError = BrowserError;
exports.ArtifactError = ArtifactError;
globalThis.__cssVisualDiffErrorClasses = { CvdError, SelectorError, PrepareError, BrowserError, ArtifactError };
})`)
	if err != nil {
		panic(vm.NewGoError(err))
	}
	fn, ok := goja.AssertFunction(value)
	if !ok {
		panic(vm.NewGoError(fmt.Errorf("css-visual-diff error class installer is not callable")))
	}
	if _, err := fn(goja.Undefined(), exports); err != nil {
		panic(err)
	}
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
				_ = reject(cvdErrorValue(vm, op, err))
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

func cvdErrorValue(vm *goja.Runtime, op string, err error) goja.Value {
	className, code := classifyCVDError(op, err)
	classesValue := vm.Get("__cssVisualDiffErrorClasses")
	if classesValue == nil || goja.IsUndefined(classesValue) || goja.IsNull(classesValue) {
		return vm.NewGoError(err)
	}
	ctor := classesValue.ToObject(vm).Get(className)
	if ctor == nil || goja.IsUndefined(ctor) || goja.IsNull(ctor) {
		return vm.NewGoError(err)
	}
	obj, newErr := vm.New(ctor, vm.ToValue(err.Error()), vm.ToValue(code), vm.ToValue(map[string]any{"operation": op}))
	if newErr != nil {
		return vm.NewGoError(err)
	}
	_ = obj.Set("operation", op)
	return obj
}

func classifyCVDError(op string, err error) (className string, code string) {
	message := ""
	if err != nil {
		message = strings.ToLower(err.Error())
	}
	switch {
	case strings.Contains(message, "selector") || strings.Contains(op, ".preflight"):
		return "SelectorError", "SELECTOR_ERROR"
	case strings.Contains(op, ".prepare") || strings.Contains(message, "prepare") || strings.Contains(message, "wait_for"):
		return "PrepareError", "PREPARE_ERROR"
	case strings.Contains(op, ".browser") || strings.Contains(op, ".goto") || strings.Contains(message, "navigate"):
		return "BrowserError", "BROWSER_ERROR"
	case strings.Contains(op, ".inspect") || strings.Contains(message, "artifact") || strings.Contains(message, "write"):
		return "ArtifactError", "ARTIFACT_ERROR"
	default:
		return "CvdError", "CVD_ERROR"
	}
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
			target, err := opts.toTarget(rawURL)
			if err != nil {
				page.Close()
				return nil, err
			}
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
	mu     sync.Mutex
	page   *service.PageService
	target config.Target
}

func newPageState(page *service.PageService, target config.Target) *pageState {
	return &pageState{page: page, target: target}
}

func (s *pageState) runExclusive(work func() (any, error)) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return work()
}

func wrapPage(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, state *pageState) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set(proxyIDProperty, defaultProxyRegistry.bind("cvd.page", state))
	_ = obj.Set("locator", func(selector string) goja.Value {
		return wrapLocator(ctx, vm, state, selector)
	})
	_ = obj.Set("goto", func(rawURL string, rawOptions map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.goto", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				opts, err := decodeInto[pageOptions](rawOptions)
				if err != nil {
					return nil, err
				}
				target, err := opts.toTarget(rawURL)
				if err != nil {
					return nil, err
				}
				if err := state.page.Page().SetViewport(target.Viewport.Width, target.Viewport.Height); err != nil {
					return nil, err
				}
				if err := state.page.Page().Goto(target.URL); err != nil {
					return nil, err
				}
				if target.WaitMS > 0 {
					state.page.Page().Wait(time.Duration(target.WaitMS) * time.Millisecond)
				}
				state.target = target
				return targetSummary(target), nil
			})
		}, nil)
	})
	_ = obj.Set("prepare", func(raw map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.prepare", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				prepare, err := decodePrepareSpec(raw)
				if err != nil {
					return nil, err
				}
				state.target.Prepare = &prepare
				return nil, service.PrepareTarget(state.page.Page(), state.target)
			})
		}, nil)
	})
	_ = obj.Set("preflight", func(raw []map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.preflight", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				probes, err := decodeProbes(raw)
				if err != nil {
					return nil, err
				}
				statuses, err := service.PreflightProbes(state.page.Page(), probes)
				if err != nil {
					return nil, err
				}
				return lowerSelectorStatuses(statuses), nil
			})
		}, nil)
	})
	_ = obj.Set("inspect", func(rawProbe map[string]any, rawOptions map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.inspect", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				requests, err := decodeInspectRequests([]map[string]any{rawProbe})
				if err != nil {
					return nil, err
				}
				opts, err := decodeInspectOptions(rawOptions)
				if err != nil {
					return nil, err
				}
				result, err := service.InspectPreparedPage(state.page.Page(), state.target, "script", requests, opts)
				if err != nil {
					return nil, err
				}
				if len(result.Results) == 0 {
					return nil, fmt.Errorf("inspect produced no results")
				}
				return lowerInspectArtifact(result.Results[0]), nil
			})
		}, nil)
	})
	_ = obj.Set("inspectAll", func(rawProbes []map[string]any, rawOptions map[string]any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.inspectAll", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				probes, err := decodeInspectRequests(rawProbes)
				if err != nil {
					return nil, err
				}
				opts, err := decodeInspectOptions(rawOptions)
				if err != nil {
					return nil, err
				}
				result, err := service.InspectPreparedPage(state.page.Page(), state.target, "script", probes, opts)
				if err != nil {
					return nil, err
				}
				return lowerInspectResult(result), nil
			})
		}, nil)
	})
	_ = obj.Set("close", func() goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.close", func() (any, error) {
			return state.runExclusive(func() (any, error) {
				state.page.Close()
				return nil, nil
			})
		}, nil)
	})
	return obj
}

type pageOptions struct {
	Viewport config.Viewport `json:"viewport"`
	WaitMS   int             `json:"waitMs"`
	Name     string          `json:"name"`
}

func (o pageOptions) toTarget(url string) (config.Target, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return config.Target{}, fmt.Errorf("url is required")
	}
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
	return config.Target{Name: name, URL: url, WaitMS: o.WaitMS, Viewport: viewport}, nil
}

func targetSummary(target config.Target) map[string]any {
	return map[string]any{
		"name":     target.Name,
		"url":      target.URL,
		"waitMs":   target.WaitMS,
		"viewport": lowerViewport(target.Viewport),
	}
}

type inspectAllOptionsInput struct {
	OutDir     string `json:"outDir"`
	Format     string `json:"format"`
	Artifacts  string `json:"artifacts"`
	OutputFile string `json:"outputFile"`
}

func decodeInspectOptions(raw map[string]any) (service.InspectAllOptions, error) {
	opts, err := decodeInto[inspectAllOptionsInput](raw)
	if err != nil {
		return service.InspectAllOptions{}, err
	}
	format := opts.Format
	if format == "" {
		format = opts.Artifacts
	}
	if format == "" {
		format = service.InspectFormatBundle
	}
	return service.InspectAllOptions{OutDir: opts.OutDir, Format: format, OutputFile: opts.OutputFile}, nil
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
	Name          string   `json:"name"`
	Selector      string   `json:"selector"`
	Props         []string `json:"props"`
	Attrs         []string `json:"attrs"`
	Attributes    []string `json:"attributes"`
	Source        string   `json:"source"`
	Required      bool     `json:"required"`
	OutputFile    string   `json:"outputFile"`
	OutputFileOld string   `json:"output_file"`
}

func (p probeInput) attributes() []string {
	if len(p.Attributes) > 0 {
		return p.Attributes
	}
	return p.Attrs
}

func decodeProbes(raw []map[string]any) ([]service.ProbeSpec, error) {
	inputs, err := decodeInto[[]probeInput](raw)
	if err != nil {
		return nil, err
	}
	ret := make([]service.ProbeSpec, 0, len(inputs))
	for _, input := range inputs {
		ret = append(ret, service.ProbeSpec{Name: input.Name, Selector: input.Selector, Props: input.Props, Attributes: input.attributes(), Source: input.Source, Required: input.Required})
	}
	return ret, nil
}

func decodeInspectRequests(raw []map[string]any) ([]service.InspectRequest, error) {
	inputs, err := decodeInto[[]probeInput](raw)
	if err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one inspect request is required")
	}
	ret := make([]service.InspectRequest, 0, len(inputs))
	for _, input := range inputs {
		selector := strings.TrimSpace(input.Selector)
		if selector == "" {
			return nil, fmt.Errorf("selector is required for inspect request")
		}
		name := input.Name
		if name == "" {
			name = service.SanitizeName(selector)
		}
		ret = append(ret, service.InspectRequest{Name: name, Selector: selector, Props: input.Props, Attributes: input.attributes(), Source: input.Source})
	}
	return ret, nil
}

func lowerSelectorStatuses(statuses []service.SelectorStatus) []map[string]any {
	ret := make([]map[string]any, 0, len(statuses))
	for _, status := range statuses {
		entry := map[string]any{
			"name":      status.Name,
			"selector":  status.Selector,
			"source":    status.Source,
			"exists":    status.Exists,
			"visible":   status.Visible,
			"textStart": status.TextStart,
			"error":     status.Error,
		}
		if status.Bounds != nil {
			entry["bounds"] = lowerBounds(*status.Bounds)
		}
		ret = append(ret, entry)
	}
	return ret
}

func lowerInspectResult(result service.InspectResult) map[string]any {
	artifacts := make([]map[string]any, 0, len(result.Results))
	for _, artifact := range result.Results {
		artifacts = append(artifacts, lowerInspectArtifact(artifact))
	}
	return map[string]any{
		"outputDir": result.OutputDir,
		"results":   artifacts,
	}
}

func lowerInspectArtifact(artifact service.InspectArtifactResult) map[string]any {
	ret := map[string]any{
		"metadata":    lowerInspectMetadata(artifact.Metadata),
		"screenshot":  artifact.Screenshot,
		"html":        artifact.HTML,
		"inspectJson": artifact.InspectJSON,
	}
	if artifact.Style != nil {
		ret["style"] = lowerStyleSnapshot(*artifact.Style)
	}
	return ret
}

func lowerInspectMetadata(metadata service.InspectMetadata) map[string]any {
	return map[string]any{
		"side":           metadata.Side,
		"targetName":     metadata.TargetName,
		"url":            metadata.URL,
		"viewport":       lowerViewport(metadata.Viewport),
		"name":           metadata.Name,
		"selector":       metadata.Selector,
		"selectorSource": metadata.SelectorSource,
		"rootSelector":   metadata.RootSelector,
		"prepareType":    metadata.PrepareType,
		"format":         metadata.Format,
		"createdAt":      metadata.CreatedAt.Format(time.RFC3339Nano),
	}
}

func lowerStyleSnapshot(style service.StyleSnapshot) map[string]any {
	ret := map[string]any{
		"exists":     style.Exists,
		"computed":   style.Computed,
		"attributes": style.Attributes,
	}
	if style.Bounds != nil {
		ret["bounds"] = lowerBounds(*style.Bounds)
	}
	return ret
}

func lowerBounds(bounds service.Bounds) map[string]any {
	return map[string]any{
		"x":      bounds.X,
		"y":      bounds.Y,
		"width":  bounds.Width,
		"height": bounds.Height,
	}
}

func lowerViewport(viewport config.Viewport) map[string]any {
	return map[string]any{
		"width":  viewport.Width,
		"height": viewport.Height,
	}
}
