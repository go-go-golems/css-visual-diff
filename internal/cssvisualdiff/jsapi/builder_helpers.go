package jsapi

import (
	"fmt"
	"math"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func requiredStringArg(vm *goja.Runtime, operation string, value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		panic(typeMismatchError(vm, operation, "non-empty string", value))
	}
	text, ok := value.Export().(string)
	if !ok {
		panic(typeMismatchError(vm, operation, "non-empty string", value))
	}
	text = strings.TrimSpace(text)
	if text == "" {
		panic(vm.NewTypeError("%s: expected non-empty string", operation))
	}
	return text
}

func optionalBoolArg(value goja.Value, fallback bool) bool {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return fallback
	}
	return value.ToBoolean()
}

func requiredNonNegativeIntArg(vm *goja.Runtime, operation string, value goja.Value) int {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		panic(typeMismatchError(vm, operation, "non-negative integer", value))
	}
	number := value.ToFloat()
	if math.Trunc(number) != number || number < 0 {
		panic(typeMismatchError(vm, operation, "non-negative integer", value))
	}
	return int(number)
}

func viewportFromCall(vm *goja.Runtime, operation string, args []goja.Value) config.Viewport {
	if len(args) == 1 && args[0] != nil && !goja.IsUndefined(args[0]) && !goja.IsNull(args[0]) {
		viewport, err := decodeInto[config.Viewport](args[0].Export())
		if err != nil {
			panic(typeMismatchError(vm, operation, "viewport object { width, height }", args[0]))
		}
		validateViewport(vm, operation, viewport)
		return viewport
	}
	if len(args) >= 2 {
		viewport := config.Viewport{Width: requiredPositiveIntArg(vm, operation+".width", args[0]), Height: requiredPositiveIntArg(vm, operation+".height", args[1])}
		validateViewport(vm, operation, viewport)
		return viewport
	}
	panic(vm.NewTypeError("%s: expected (width, height) or { width, height }", operation))
}

func requiredPositiveIntArg(vm *goja.Runtime, operation string, value goja.Value) int {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		panic(typeMismatchError(vm, operation, "positive integer", value))
	}
	number := value.ToFloat()
	if math.Trunc(number) != number || number <= 0 {
		panic(typeMismatchError(vm, operation, "positive integer", value))
	}
	return int(number)
}

func validateViewport(vm *goja.Runtime, operation string, viewport config.Viewport) {
	if viewport.Width <= 0 || viewport.Height <= 0 {
		panic(vm.NewTypeError("%s: expected positive width and height, got width=%d height=%d", operation, viewport.Width, viewport.Height))
	}
}

func decodeIntoValue(vm *goja.Runtime, value goja.Value, out any) error {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return fmt.Errorf("value is undefined or null")
	}
	return vm.ExportTo(value, out)
}
