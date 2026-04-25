package jsapi

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

type locatorHandle struct {
	page     *pageState
	selector string
}

func wrapLocator(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, state *pageState, selector string) goja.Value {
	locator := &locatorHandle{page: state, selector: selector}
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.locator",
		Methods: map[string]ProxyMethod{
			"status":        locator.status(ctx, vm),
			"exists":        locator.exists(ctx, vm),
			"visible":       locator.visible(ctx, vm),
			"waitFor":       locator.waitFor(ctx, vm),
			"text":          locator.text(ctx, vm),
			"bounds":        locator.bounds(ctx, vm),
			"computedStyle": locator.computedStyle(ctx, vm),
			"attributes":    locator.attributes(ctx, vm),
			"collect":       locator.collect(ctx, vm),
		},
		MethodOwners: map[string]MethodSpec{
			"selector": {Owner: "cvd.probe", Hint: "A locator already has a selector. For reusable inspection recipes, use cvd.probe(\"name\").selector(\"#selector\")."},
			"styles":   {Owner: "cvd.probe", Hint: "For direct style reads on a locator, use .computedStyle([\"color\"]) instead."},
			"required": {Owner: "cvd.probe", Hint: "Required/missing-selector policy belongs to reusable probes, not page-bound locators."},
			"build":    {Owner: "cvd.probe", Hint: "Locators are live page-bound handles and do not need .build()."},
			"diff":     {Owner: "cvd.selectionComparison", Hint: "Collect locators first with await locator.collect(), then compare with cvd.compare.selections(left, right), or use cvd.compare.region({ left, right })."},
		},
	}, locator)
}

func (l *locatorHandle) spec() service.LocatorSpec {
	return service.LocatorSpec{Selector: l.selector}
}

func (l *locatorHandle) status(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.locator.status", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				status, err := service.LocatorStatus(l.page.page.Page(), l.spec())
				if err != nil {
					return nil, err
				}
				return lowerSelectorStatus(status), nil
			})
		}, nil)
	}
}

func (l *locatorHandle) exists(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.locator.exists", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				status, err := service.LocatorStatus(l.page.page.Page(), l.spec())
				if err != nil {
					return nil, err
				}
				return status.Exists, nil
			})
		}, nil)
	}
}

func (l *locatorHandle) visible(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.locator.visible", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				status, err := service.LocatorStatus(l.page.page.Page(), l.spec())
				if err != nil {
					return nil, err
				}
				return status.Visible, nil
			})
		}, nil)
	}
}

func (l *locatorHandle) waitFor(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		rawOptions := exportOptionalObject(vm, "css-visual-diff.locator.waitFor", call.Argument(0))
		return promiseValue(ctx, vm, "css-visual-diff.locator.waitFor", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				opts, err := decodeInto[service.WaitForSelectorOptions](rawOptions)
				if err != nil {
					return nil, err
				}
				result, err := service.WaitForLocator(l.page.page.Page(), l.spec(), opts)
				if err != nil {
					return nil, err
				}
				return lowerJSON(result), nil
			})
		}, nil)
	}
}

func (l *locatorHandle) text(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		rawOptions := map[string]any{}
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) && !goja.IsNull(call.Argument(0)) {
			if exported, ok := call.Argument(0).Export().(map[string]any); ok {
				rawOptions = exported
			} else {
				panic(typeMismatchError(vm, "css-visual-diff.locator.text", "options object", call.Argument(0)))
			}
		}
		return promiseValue(ctx, vm, "css-visual-diff.locator.text", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				opts, err := decodeInto[service.TextOptions](rawOptions)
				if err != nil {
					return nil, err
				}
				return service.LocatorText(l.page.page.Page(), l.spec(), opts)
			})
		}, nil)
	}
}

func (l *locatorHandle) bounds(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.locator.bounds", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				bounds, err := service.LocatorBounds(l.page.page.Page(), l.spec())
				if err != nil {
					return nil, err
				}
				if bounds == nil {
					return nil, nil
				}
				return lowerBounds(*bounds), nil
			})
		}, nil)
	}
}

func (l *locatorHandle) computedStyle(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		props, err := stringListArg(vm, "css-visual-diff.locator.computedStyle", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.locator.computedStyle", "array of CSS property names", call.Argument(0)))
		}
		return promiseValue(ctx, vm, "css-visual-diff.locator.computedStyle", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				return service.LocatorComputedStyle(l.page.page.Page(), l.spec(), props)
			})
		}, nil)
	}
}

func (l *locatorHandle) attributes(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		attrs, err := stringListArg(vm, "css-visual-diff.locator.attributes", call.Argument(0))
		if err != nil {
			panic(typeMismatchError(vm, "css-visual-diff.locator.attributes", "array of attribute names", call.Argument(0)))
		}
		return promiseValue(ctx, vm, "css-visual-diff.locator.attributes", func() (any, error) {
			return l.page.runExclusive(func() (any, error) {
				return service.LocatorAttributes(l.page.page.Page(), l.spec(), attrs)
			})
		}, nil)
	}
}

func stringListArg(vm *goja.Runtime, operation string, value goja.Value) ([]string, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return []string{}, nil
	}
	exported := value.Export()
	items, ok := exported.([]any)
	if !ok {
		return nil, fmt.Errorf("%s: expected array", operation)
	}
	ret := make([]string, 0, len(items))
	for i, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s: expected string at index %d", operation, i)
		}
		ret = append(ret, text)
	}
	return ret, nil
}

func lowerSelectorStatus(status service.SelectorStatus) map[string]any {
	statuses := lowerSelectorStatuses([]service.SelectorStatus{status})
	if len(statuses) == 0 {
		return map[string]any{}
	}
	return statuses[0]
}
