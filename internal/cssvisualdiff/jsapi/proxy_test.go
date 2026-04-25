package jsapi

import (
	"strings"
	"testing"

	"github.com/dop251/goja"
)

type testProbeBuilder struct {
	selector string
}

func newTestProbe(vm *goja.Runtime, registry *ProxyRegistry) goja.Value {
	probe := &testProbeBuilder{}
	return newProxyValue(vm, registry, ProxySpec{
		Owner: "cvd.probe",
		Methods: map[string]ProxyMethod{
			"selector": func(call goja.FunctionCall, receiver goja.Value) goja.Value {
				probe.selector = call.Argument(0).String()
				return receiver
			},
			"styles": func(call goja.FunctionCall, receiver goja.Value) goja.Value {
				return receiver
			},
		},
		MethodOwners: map[string]MethodSpec{
			"computedStyle": {Owner: "cvd.locator", Hint: "For probe style capture, use .styles([\"color\"]) instead."},
		},
	}, probe)
}

func TestProxyUnknownMethodError(t *testing.T) {
	vm := goja.New()
	registry := NewProxyRegistry()
	if err := vm.Set("probe", newTestProbe(vm, registry)); err != nil {
		t.Fatalf("set probe: %v", err)
	}

	_, err := vm.RunString(`probe.style(["color"])`)
	if err == nil {
		t.Fatalf("expected unknown method error")
	}
	message := err.Error()
	for _, want := range []string{"cvd.probe", "unknown method .style()", "Available:", "styles", "Did you mean .styles()?"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in error %q", want, message)
		}
	}
}

func TestProxyWrongParentError(t *testing.T) {
	vm := goja.New()
	registry := NewProxyRegistry()
	if err := vm.Set("probe", newTestProbe(vm, registry)); err != nil {
		t.Fatalf("set probe: %v", err)
	}

	_, err := vm.RunString(`probe.computedStyle(["color"])`)
	if err == nil {
		t.Fatalf("expected wrong-parent error")
	}
	message := err.Error()
	for _, want := range []string{"cvd.probe", ".computedStyle() is not available here", "belongs to cvd.locator", "use .styles"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in error %q", want, message)
		}
	}
}

func TestProxyUnwrapBacking(t *testing.T) {
	vm := goja.New()
	registry := NewProxyRegistry()
	probeValue := newTestProbe(vm, registry)
	if err := vm.Set("probe", probeValue); err != nil {
		t.Fatalf("set probe: %v", err)
	}
	if _, err := vm.RunString(`probe.selector("#cta")`); err != nil {
		t.Fatalf("run script: %v", err)
	}

	probe, err := unwrapProxyBacking[testProbeBuilder](vm, registry, "test.unwrap", probeValue, "cvd.probe")
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if probe.selector != "#cta" {
		t.Fatalf("expected selector to be updated, got %q", probe.selector)
	}
}

func TestProxyUnwrapRejectsRawObject(t *testing.T) {
	vm := goja.New()
	registry := NewProxyRegistry()
	raw := vm.ToValue(map[string]any{"selector": "#cta"})

	_, err := unwrapProxyBacking[testProbeBuilder](vm, registry, "cvd.snapshot", raw, "cvd.probe")
	if err == nil {
		t.Fatalf("expected raw object rejection")
	}
	if !strings.Contains(err.Error(), "cvd.snapshot: expected cvd.probe") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProxyMustUnwrapRejectsWrongOwner(t *testing.T) {
	vm := goja.New()
	registry := NewProxyRegistry()
	locatorValue := newProxyValue(vm, registry, ProxySpec{
		Owner: "cvd.locator",
		Methods: map[string]ProxyMethod{
			"text": func(call goja.FunctionCall, receiver goja.Value) goja.Value { return vm.ToValue("text") },
		},
	}, &struct{}{})

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic")
		}
		if !strings.Contains(recovered.(goja.Value).String(), "cvd.snapshot: expected cvd.probe") {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()
	_ = mustUnwrapProxyBacking[testProbeBuilder](vm, registry, "cvd.snapshot", locatorValue, "cvd.probe")
}
