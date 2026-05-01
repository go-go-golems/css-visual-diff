package jsapi

import (
	"testing"

	"github.com/dop251/goja"
)

func TestOverlayBuildersProduceOpaqueSpec(t *testing.T) {
	vm := goja.New()
	exports := vm.NewObject()
	installOverlayAPI(nil, vm, exports)
	_ = vm.Set("cvd", exports)
	value, err := vm.RunString(`
const spec = cvd.overlaySpec()
  .legend(true)
  .style({ legend: { position: "bottom-right" }, targetDefaults: { labelColor: "white" } })
  .target(cvd.overlayTarget("Hero").selector(".hero").borderColor("#ff6347").labelBackground("#ff6347"))
  .build();
spec;
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := overlaySpecFromValue(vm, value)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Targets) != 1 {
		t.Fatalf("expected one target, got %d", len(spec.Targets))
	}
	if spec.Targets[0].Name != "Hero" || spec.Targets[0].Selector != ".hero" {
		t.Fatalf("unexpected target: %#v", spec.Targets[0])
	}
	if spec.Targets[0].Style.BorderColor == nil || spec.Targets[0].Style.BorderColor.R != 255 {
		t.Fatalf("expected border color parsed, got %#v", spec.Targets[0].Style.BorderColor)
	}
}

func TestOverlayRejectsRawObject(t *testing.T) {
	vm := goja.New()
	_, err := overlaySpecFromValue(vm, vm.ToValue(map[string]any{"targets": []any{}}))
	if err == nil {
		t.Fatalf("expected raw object to be rejected")
	}
}
