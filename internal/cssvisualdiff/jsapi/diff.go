package jsapi

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

func installDiffAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	structuralDiff := func(call goja.FunctionCall) goja.Value {
		var opts service.DiffOptions
		if len(call.Arguments) > 2 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			var err error
			opts, err = decodeInto[service.DiffOptions](call.Argument(2).Export())
			if err != nil {
				panic(typeMismatchError(vm, "css-visual-diff.diff", "options object", call.Argument(2)))
			}
		}
		diff, err := service.DiffValues(call.Argument(0).Export(), call.Argument(1).Export(), opts)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return vm.ToValue(lowerSnapshotDiff(diff))
	}
	_ = exports.Set("diff", structuralDiff)
	if diffValue := exports.Get("diff"); diffValue != nil {
		_ = diffValue.ToObject(vm).Set("structural", structuralDiff)
	}

	image := vm.NewObject()
	_ = image.Set("diff", func(call goja.FunctionCall) goja.Value {
		raw := exportOptionalObject(vm, "css-visual-diff.image.diff", call.Argument(0))
		return promiseValue(ctx, vm, "css-visual-diff.image.diff", func() (any, error) {
			left := stringFromAny(raw["left"])
			right := stringFromAny(raw["right"])
			threshold := intNumber(raw["threshold"])
			diffOnlyPath := stringFromAny(raw["diffOnlyPath"])
			diffComparisonPath := stringFromAny(raw["diffComparisonPath"])
			if diffOnlyPath != "" || diffComparisonPath != "" {
				return service.WritePixelDiffImages(left, right, diffComparisonPath, diffOnlyPath, service.PixelDiffOptions{Threshold: threshold})
			}
			result, _, _, _, err := service.DiffPNGFiles(left, right, service.PixelDiffOptions{Threshold: threshold})
			return result, err
		}, func(vm *goja.Runtime, value any) goja.Value { return vm.ToValue(lowerJSON(value)) })
	})
	_ = exports.Set("image", image)

	_ = exports.Set("report", func(raw map[string]any) *goja.Object {
		diff, err := decodeSnapshotDiffReport(raw)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return wrapDiffReport(ctx, vm, diff)
	})

	write := vm.NewObject()
	_ = write.Set("json", func(path string, value any) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.write.json", func() (any, error) {
			bytes, err := jsonMarshalIndent(value)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return nil, err
			}
			return path, os.WriteFile(path, bytes, 0o644)
		}, nil)
	})
	_ = write.Set("markdown", func(path string, markdown string) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.write.markdown", func() (any, error) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return nil, err
			}
			return path, os.WriteFile(path, []byte(markdown), 0o644)
		}, nil)
	})
	_ = exports.Set("write", write)
}

func wrapDiffReport(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, diff service.SnapshotDiff) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("markdown", func() string { return service.RenderDiffMarkdown(diff) })
	_ = obj.Set("writeMarkdown", func(path string) goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.report.writeMarkdown", func() (any, error) {
			markdown := service.RenderDiffMarkdown(diff)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return nil, err
			}
			return path, os.WriteFile(path, []byte(markdown), 0o644)
		}, nil)
	})
	return obj
}

func jsonMarshalIndent(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}

func decodeSnapshotDiffReport(raw map[string]any) (service.SnapshotDiff, error) {
	if _, hasSnake := raw["change_count"]; !hasSnake {
		if changeCount, hasCamel := raw["changeCount"]; hasCamel {
			clone := make(map[string]any, len(raw)+1)
			for k, v := range raw {
				clone[k] = v
			}
			clone["change_count"] = changeCount
			raw = clone
		}
	}
	return decodeInto[service.SnapshotDiff](raw)
}

func lowerSnapshotDiff(diff service.SnapshotDiff) map[string]any {
	changes := make([]map[string]any, 0, len(diff.Changes))
	for _, change := range diff.Changes {
		changes = append(changes, map[string]any{"path": change.Path, "before": change.Before, "after": change.After})
	}
	return map[string]any{"equal": diff.Equal, "changeCount": diff.ChangeCount, "changes": changes}
}
