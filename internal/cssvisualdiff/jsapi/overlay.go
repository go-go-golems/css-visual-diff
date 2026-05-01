package jsapi

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

const (
	overlayTargetBuilderOwner = "cvd.overlayTarget"
	overlayTargetSpecOwner    = "cvd.overlayTargetSpec"
	overlaySpecBuilderOwner   = "cvd.overlaySpec"
	overlaySpecOwner          = "cvd.overlaySpecValue"
)

type overlayTargetBuilder struct{ target service.OverlayTarget }
type overlayTargetSpec struct{ target service.OverlayTarget }
type overlaySpecBuilder struct{ spec service.OverlaySpec }
type overlaySpecValue struct{ spec service.OverlaySpec }

func installOverlayAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	_ = exports.Set("overlayTarget", func(call goja.FunctionCall) goja.Value {
		name := requiredStringArg(vm, "cvd.overlayTarget", call.Argument(0))
		return newOverlayTargetBuilder(vm, service.OverlayTarget{Name: name})
	})
	_ = exports.Set("overlaySpec", func(call goja.FunctionCall) goja.Value {
		return newOverlaySpecBuilder(vm, service.OverlaySpec{Legend: true, Screenshot: "fullPage"})
	})
}

func newOverlayTargetBuilder(vm *goja.Runtime, target service.OverlayTarget) goja.Value {
	b := &overlayTargetBuilder{target: target}
	return newProxyValue(vm, nil, ProxySpec{Owner: overlayTargetBuilderOwner, Methods: map[string]ProxyMethod{
		"selector": b.selector(vm), "label": b.label(vm), "style": b.style(vm),
		"borderColor": b.borderColor(vm), "contentBackground": b.contentBackground(vm),
		"labelBackground": b.labelBackground(vm), "labelColor": b.labelColor(vm), "labelPosition": b.labelPosition(vm),
		"build": b.build(vm),
	}}, b)
}

func (b *overlayTargetBuilder) selector(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.Selector = requiredStringArg(vm, "cvd.overlayTarget.selector", call.Argument(0))
		return receiver
	}
}
func (b *overlayTargetBuilder) label(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.target.Label = requiredStringArg(vm, "cvd.overlayTarget.label", call.Argument(0))
		return receiver
	}
}
func (b *overlayTargetBuilder) style(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		st, err := decodeTargetStyleValue(vm, "cvd.overlayTarget.style", call.Argument(0))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		b.target.Style = b.target.Style.Merge(st)
		return receiver
	}
}
func (b *overlayTargetBuilder) borderColor(vm *goja.Runtime) ProxyMethod {
	return targetColorMethod(vm, "cvd.overlayTarget.borderColor", func(c color.RGBA) { b.target.Style.BorderColor = &c })
}
func (b *overlayTargetBuilder) contentBackground(vm *goja.Runtime) ProxyMethod {
	return targetColorMethod(vm, "cvd.overlayTarget.contentBackground", func(c color.RGBA) { b.target.Style.ContentBackground = &c })
}
func (b *overlayTargetBuilder) labelBackground(vm *goja.Runtime) ProxyMethod {
	return targetColorMethod(vm, "cvd.overlayTarget.labelBackground", func(c color.RGBA) { b.target.Style.Label.Background = &c })
}
func (b *overlayTargetBuilder) labelColor(vm *goja.Runtime) ProxyMethod {
	return targetColorMethod(vm, "cvd.overlayTarget.labelColor", func(c color.RGBA) { b.target.Style.Label.Color = &c })
}
func (b *overlayTargetBuilder) labelPosition(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		p, err := service.ParseLabelPosition(requiredStringArg(vm, "cvd.overlayTarget.labelPosition", call.Argument(0)))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		b.target.Style.Label.Position = p
		return receiver
	}
}
func (b *overlayTargetBuilder) build(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		if strings.TrimSpace(b.target.Selector) == "" {
			panic(vm.NewTypeError("cvd.overlayTarget.build: selector is required"))
		}
		spec := &overlayTargetSpec{target: b.target}
		return newProxyValue(vm, nil, ProxySpec{Owner: overlayTargetSpecOwner, Methods: map[string]ProxyMethod{}}, spec)
	}
}

func targetColorMethod(vm *goja.Runtime, op string, set func(color.RGBA)) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		c, err := service.ParseColor(requiredStringArg(vm, op, call.Argument(0)))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		set(c)
		return receiver
	}
}

func newOverlaySpecBuilder(vm *goja.Runtime, spec service.OverlaySpec) goja.Value {
	b := &overlaySpecBuilder{spec: spec}
	return newProxyValue(vm, nil, ProxySpec{Owner: overlaySpecBuilderOwner, Methods: map[string]ProxyMethod{
		"target": b.target(vm), "targets": b.targets(vm), "legend": b.legend(vm), "screenshot": b.screenshot(vm), "style": b.style(vm),
		"cropTo": b.cropTo(vm), "cropPadding": b.cropPadding(vm), "build": b.build(vm),
	}}, b)
}

func (b *overlaySpecBuilder) target(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		t, err := overlayTargetFromValue(vm, call.Argument(0))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		b.spec.Targets = append(b.spec.Targets, t)
		return receiver
	}
}
func (b *overlaySpecBuilder) targets(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		arr := call.Argument(0).ToObject(vm)
		l := int(arr.Get("length").ToInteger())
		for i := 0; i < l; i++ {
			t, err := overlayTargetFromValue(vm, arr.Get(fmt.Sprintf("%d", i)))
			if err != nil {
				panic(vm.NewTypeError(err.Error()))
			}
			b.spec.Targets = append(b.spec.Targets, t)
		}
		return receiver
	}
}
func (b *overlaySpecBuilder) legend(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.spec.Legend = optionalBoolArg(call.Argument(0), true)
		return receiver
	}
}
func (b *overlaySpecBuilder) screenshot(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		b.spec.Screenshot = requiredStringArg(vm, "cvd.overlaySpec.screenshot", call.Argument(0))
		return receiver
	}
}
func (b *overlaySpecBuilder) style(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		st, err := decodeOverlayStyleValue(vm, "cvd.overlaySpec.style", call.Argument(0))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		b.spec.Style = b.spec.Style.Merge(st)
		return receiver
	}
}
func (b *overlaySpecBuilder) cropTo(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		if b.spec.Crop == nil {
			b.spec.Crop = &service.OverlayCrop{}
		}
		b.spec.Crop.Selector = requiredStringArg(vm, "cvd.overlaySpec.cropTo", call.Argument(0))
		b.spec.Crop.Target = ""
		return receiver
	}
}
func (b *overlaySpecBuilder) cropPadding(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		padding, ok := decodeInsets(call.Argument(0).Export())
		if !ok {
			panic(typeMismatchError(vm, "cvd.overlaySpec.cropPadding", "number, [vertical, horizontal], or [top, right, bottom, left]", call.Argument(0)))
		}
		if b.spec.Crop == nil {
			b.spec.Crop = &service.OverlayCrop{}
		}
		b.spec.Crop.Padding = padding
		return receiver
	}
}
func (b *overlaySpecBuilder) build(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		if len(b.spec.Targets) == 0 {
			panic(vm.NewTypeError("cvd.overlaySpec.build: at least one target is required"))
		}
		val := &overlaySpecValue{spec: b.spec}
		return newProxyValue(vm, nil, ProxySpec{Owner: overlaySpecOwner, Methods: map[string]ProxyMethod{}}, val)
	}
}

func overlayTargetFromValue(vm *goja.Runtime, value goja.Value) (service.OverlayTarget, error) {
	if b, err := unwrapProxyBacking[overlayTargetBuilder](vm, defaultProxyRegistry, "cvd.overlaySpec.target", value, overlayTargetBuilderOwner); err == nil {
		if strings.TrimSpace(b.target.Selector) == "" {
			return service.OverlayTarget{}, fmt.Errorf("cvd.overlaySpec.target: selector is required")
		}
		return b.target, nil
	}
	if s, err := unwrapProxyBacking[overlayTargetSpec](vm, defaultProxyRegistry, "cvd.overlaySpec.target", value, overlayTargetSpecOwner); err == nil {
		return s.target, nil
	}
	return service.OverlayTarget{}, fmt.Errorf("cvd.overlaySpec.target: expected cvd.overlayTarget builder/spec, got %s", valueKind(value))
}

func overlaySpecFromValue(vm *goja.Runtime, value goja.Value) (service.OverlaySpec, error) {
	if b, err := unwrapProxyBacking[overlaySpecBuilder](vm, defaultProxyRegistry, "page.overlay", value, overlaySpecBuilderOwner); err == nil {
		return b.spec, nil
	}
	if s, err := unwrapProxyBacking[overlaySpecValue](vm, defaultProxyRegistry, "page.overlay", value, overlaySpecOwner); err == nil {
		return s.spec, nil
	}
	return service.OverlaySpec{}, fmt.Errorf("page.overlay: expected cvd.overlaySpec builder/spec, got %s", valueKind(value))
}

func wrapOverlayScreenshotBuilder(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, state *pageState, spec service.OverlaySpec) goja.Value {
	obj := vm.NewObject()
	_ = obj.Set("screenshot", func(path string) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.page.overlay.screenshot", func() (any, error) {
			return state.runExclusive(func() (any, error) { return service.OverlayScreenshot(state.page.Page(), spec, path) })
		}, func(vm *goja.Runtime, value any) goja.Value {
			return vm.ToValue(lowerOverlayResult(value.(*service.OverlayResult)))
		})
	})
	return obj
}

func lowerOverlayResult(r *service.OverlayResult) map[string]any {
	return map[string]any{"outputPath": r.OutputPath, "colors": r.Colors, "width": r.Width, "height": r.Height, "targets": r.Targets}
}

func decodeOverlayStyleValue(vm *goja.Runtime, op string, value goja.Value) (service.OverlayStyle, error) {
	m, ok := value.Export().(map[string]any)
	if !ok {
		return service.OverlayStyle{}, fmt.Errorf("%s: expected style object", op)
	}
	var st service.OverlayStyle
	if v, ok := m["label"].(map[string]any); ok {
		st.Label = decodeLabelOverlayStyle(v)
	}
	if v, ok := m["legend"].(map[string]any); ok {
		l, err := decodeLegendStyle(v)
		if err != nil {
			return st, err
		}
		st.Legend = l
	}
	if v, ok := m["targetDefaults"].(map[string]any); ok {
		t, err := decodeTargetStyleMap(v)
		if err != nil {
			return st, err
		}
		st.TargetDefaults = t
	}
	return st, nil
}

func decodeTargetStyleValue(vm *goja.Runtime, op string, value goja.Value) (service.TargetOverlayStyle, error) {
	m, ok := value.Export().(map[string]any)
	if !ok {
		return service.TargetOverlayStyle{}, fmt.Errorf("%s: expected target style object", op)
	}
	return decodeTargetStyleMap(m)
}

func decodeLabelOverlayStyle(m map[string]any) service.LabelOverlayStyle {
	var s service.LabelOverlayStyle
	if v, ok := m["fontFamily"].(string); ok {
		s.FontFamily = v
	}
	if v, ok := number(m["fontSize"]); ok {
		s.FontSize = int(v)
	}
	if v, ok := number(m["radius"]); ok {
		s.Radius = int(v)
	}
	if p, ok := decodeInsets(m["padding"]); ok {
		s.Padding = p
	}
	return s
}
func decodeLegendStyle(m map[string]any) (service.LegendOverlayStyle, error) {
	var s service.LegendOverlayStyle
	if v, ok := m["position"].(string); ok {
		p, err := service.ParseLegendPosition(v)
		if err != nil {
			return s, err
		}
		s.Position = p
	}
	if v, ok := m["background"].(string); ok {
		c, err := service.ParseColor(v)
		if err != nil {
			return s, err
		}
		s.Background = &c
	}
	if v, ok := m["color"].(string); ok {
		c, err := service.ParseColor(v)
		if err != nil {
			return s, err
		}
		s.Color = &c
	}
	return s, nil
}
func decodeTargetStyleMap(m map[string]any) (service.TargetOverlayStyle, error) {
	var s service.TargetOverlayStyle
	for key, set := range map[string]func(color.RGBA){"borderColor": func(c color.RGBA) { s.BorderColor = &c }, "contentBackground": func(c color.RGBA) { s.ContentBackground = &c }, "paddingBackground": func(c color.RGBA) { s.PaddingBackground = &c }, "marginBackground": func(c color.RGBA) { s.MarginBackground = &c }, "labelColor": func(c color.RGBA) { s.LabelColor = &c }} {
		if v, ok := m[key].(string); ok {
			c, err := service.ParseColor(v)
			if err != nil {
				return s, err
			}
			set(c)
		}
	}
	if v, ok := number(m["borderWidth"]); ok {
		s.BorderWidth = int(v)
	}
	if lm, ok := m["label"].(map[string]any); ok {
		if v, ok := lm["background"].(string); ok {
			c, err := service.ParseColor(v)
			if err != nil {
				return s, err
			}
			s.Label.Background = &c
		}
		if v, ok := lm["color"].(string); ok {
			c, err := service.ParseColor(v)
			if err != nil {
				return s, err
			}
			s.Label.Color = &c
		}
		if v, ok := lm["position"].(string); ok {
			p, err := service.ParseLabelPosition(v)
			if err != nil {
				return s, err
			}
			s.Label.Position = p
		}
	}
	return s, nil
}
func number(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}
func decodeInsets(v any) (service.Insets, bool) {
	if n, ok := number(v); ok {
		i := int(n)
		return service.Insets{Top: i, Right: i, Bottom: i, Left: i}, true
	}
	arr, ok := v.([]any)
	if !ok {
		return service.Insets{}, false
	}
	nums := []int{}
	for _, x := range arr {
		n, ok := number(x)
		if !ok {
			return service.Insets{}, false
		}
		nums = append(nums, int(n))
	}
	switch len(nums) {
	case 2:
		return service.Insets{Top: nums[0], Bottom: nums[0], Right: nums[1], Left: nums[1]}, true
	case 4:
		return service.Insets{Top: nums[0], Right: nums[1], Bottom: nums[2], Left: nums[3]}, true
	default:
		return service.Insets{}, false
	}
}
