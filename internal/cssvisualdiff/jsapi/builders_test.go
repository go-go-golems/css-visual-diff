package jsapi

import (
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestTargetProbeAndExtractorBuilders(t *testing.T) {
	vm := goja.New()
	exports := vm.NewObject()
	installTargetAPI(vm, exports)
	installProbeAPI(vm, exports)
	installExtractorAPI(vm, exports)
	require.NoError(t, vm.Set("cvd", exports))

	value, err := vm.RunString(`(() => {
  const target = cvd.target("booking")
    .url("http://example.test/booking")
    .viewport(1440, 900)
    .waitMs(250)
    .root("#app")
    .build();
  const viewport = cvd.viewport.desktop();
  const probe = cvd.probe("cta")
    .selector("#cta")
    .required()
    .source("react")
    .text()
    .bounds()
    .styles(["color", "font-size"])
    .attributes(["id", "class"])
    .build();
  const extractor = cvd.extractors.computedStyle(["color"]).build();
  return { target, viewport, probe, extractor };
})()`)
	require.NoError(t, err)
	out := value.Export().(map[string]any)

	target := out["target"].(map[string]any)
	require.Equal(t, "booking", target["name"])
	require.Equal(t, "http://example.test/booking", target["url"])
	require.EqualValues(t, 250, target["waitMs"])
	require.Equal(t, "#app", target["rootSelector"])
	require.EqualValues(t, 1440, target["viewport"].(map[string]any)["width"])
	require.EqualValues(t, 900, target["viewport"].(map[string]any)["height"])

	viewport := out["viewport"].(map[string]any)
	require.EqualValues(t, 1280, viewport["width"])
	require.EqualValues(t, 720, viewport["height"])

	probe := out["probe"].(map[string]any)
	require.Equal(t, "cta", probe["name"])
	require.Equal(t, "#cta", probe["selector"])
	require.Equal(t, "react", probe["source"])
	require.Equal(t, true, probe["required"])
	require.Equal(t, []string{"color", "font-size"}, probe["props"])
	require.Equal(t, []string{"id", "class"}, probe["attributes"])
	require.Len(t, probe["extractors"], 4)

	extractor := out["extractor"].(map[string]any)
	require.Equal(t, "computedStyle", extractor["kind"])
	require.Equal(t, []string{"color"}, extractor["props"])
}

func TestBuilderValidationErrors(t *testing.T) {
	vm := goja.New()
	exports := vm.NewObject()
	installTargetAPI(vm, exports)
	installProbeAPI(vm, exports)
	installExtractorAPI(vm, exports)
	require.NoError(t, vm.Set("cvd", exports))

	cases := []struct {
		name    string
		script  string
		message string
	}{
		{name: "target name", script: `cvd.target("")`, message: "css-visual-diff.target: expected non-empty string"},
		{name: "target viewport", script: `cvd.target("x").viewport("wide")`, message: "cvd.target.viewport"},
		{name: "target build url", script: `cvd.target("x").build()`, message: "cvd.target.build: url is required"},
		{name: "probe selector", script: `cvd.probe("x").build()`, message: "cvd.probe.build: selector is required"},
		{name: "probe styles", script: `cvd.probe("x").selector("#x").styles("color")`, message: "cvd.probe.styles"},
		{name: "extractor computedStyle", script: `cvd.extractors.computedStyle("color")`, message: "cvd.extractors.computedStyle"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := vm.RunString(tc.script)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.message)
		})
	}
}

func TestBuilderWrongParentErrors(t *testing.T) {
	vm := goja.New()
	exports := vm.NewObject()
	installTargetAPI(vm, exports)
	installProbeAPI(vm, exports)
	installExtractorAPI(vm, exports)
	require.NoError(t, vm.Set("cvd", exports))

	_, err := vm.RunString(`cvd.probe("cta").computedStyle(["color"])`)
	require.Error(t, err)
	message := err.Error()
	for _, want := range []string{"cvd.probe", ".computedStyle() is not available here", "belongs to cvd.locator", "use .styles"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in error %q", want, message)
		}
	}
}
