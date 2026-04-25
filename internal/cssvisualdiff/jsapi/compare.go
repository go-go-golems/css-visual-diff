package jsapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

type selectionComparisonHandle struct {
	data service.SelectionComparisonData
}

func installCompareAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	compare := vm.NewObject()
	_ = compare.Set("selections", func(call goja.FunctionCall) goja.Value {
		left := mustUnwrapProxyBacking[collectedSelectionHandle](vm, defaultProxyRegistry, "css-visual-diff.compare.selections", call.Argument(0), "cvd.collectedSelection")
		right := mustUnwrapProxyBacking[collectedSelectionHandle](vm, defaultProxyRegistry, "css-visual-diff.compare.selections", call.Argument(1), "cvd.collectedSelection")
		rawOptions := exportOptionalObject(vm, "css-visual-diff.compare.selections", call.Argument(2))
		return promiseValue(ctx, vm, "css-visual-diff.compare.selections", func() (any, error) {
			opts, err := decodeInto[service.CompareSelectionOptions](rawOptions)
			if err != nil {
				return nil, err
			}
			return service.CompareSelections(left.data, right.data, opts)
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapSelectionComparison(ctx, vm, value.(service.SelectionComparisonData))
		})
	})
	_ = exports.Set("compare", compare)
}

func wrapSelectionComparison(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, data service.SelectionComparisonData) goja.Value {
	handle := &selectionComparisonHandle{data: data}
	return newProxyValue(vm, nil, ProxySpec{
		Owner: "cvd.selectionComparison",
		Methods: map[string]ProxyMethod{
			"summary":  handle.summary(vm),
			"toJSON":   handle.toJSON(vm),
			"left":     handle.left(vm),
			"right":    handle.right(vm),
			"artifact": handle.artifact(vm),
		},
		Properties: map[string]ProxyProperty{
			"pixel":      handle.pixel(vm),
			"bounds":     handle.bounds(vm),
			"styles":     handle.styles(vm),
			"attributes": handle.attributes(vm),
			"report":     handle.report(ctx, vm),
			"artifacts":  handle.artifacts(ctx, vm),
		},
		MethodOwners: map[string]MethodSpec{
			"collect": {Owner: "cvd.locator", Hint: "Collect browser facts from page locators before comparing selections."},
		},
	}, handle)
}

func (h *selectionComparisonHandle) summary(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		out := map[string]any{
			"schemaVersion":    h.data.SchemaVersion,
			"name":             h.data.Name,
			"left":             h.data.Left,
			"right":            h.data.Right,
			"boundsChanged":    h.data.Bounds.Changed,
			"textChanged":      h.data.Text.Changed,
			"styleChanges":     len(h.data.Styles),
			"attributeChanges": len(h.data.Attributes),
			"artifactCount":    len(h.data.Artifacts),
		}
		if h.data.Pixel != nil {
			out["pixel"] = h.data.Pixel
		}
		return vm.ToValue(lowerJSON(out))
	}
}

func (h *selectionComparisonHandle) toJSON(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value { return vm.ToValue(lowerJSON(h.data)) }
}

func (h *selectionComparisonHandle) left(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return vm.ToValue(lowerJSON(h.data.Left))
	}
}

func (h *selectionComparisonHandle) right(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		return vm.ToValue(lowerJSON(h.data.Right))
	}
}

func (h *selectionComparisonHandle) artifact(vm *goja.Runtime) ProxyMethod {
	return func(call goja.FunctionCall, receiver goja.Value) goja.Value {
		name := call.Argument(0).String()
		for _, artifact := range h.data.Artifacts {
			if artifact.Name == name {
				return vm.ToValue(lowerJSON(artifact))
			}
		}
		return goja.Null()
	}
}

func (h *selectionComparisonHandle) pixel(vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("summary", func() any {
			if h.data.Pixel == nil {
				return nil
			}
			return lowerJSON(h.data.Pixel)
		})
		return obj
	}
}

func (h *selectionComparisonHandle) bounds(vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("diff", func() any { return lowerJSON(h.data.Bounds) })
		return obj
	}
}

func (h *selectionComparisonHandle) styles(vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("diff", func(call goja.FunctionCall) any {
			props, err := stringListArg(vm, "css-visual-diff.selectionComparison.styles.diff", call.Argument(0))
			if err != nil {
				panic(typeMismatchError(vm, "css-visual-diff.selectionComparison.styles.diff", "array of CSS property names", call.Argument(0)))
			}
			return lowerJSON(filterMapDiffs(h.data.Styles, props))
		})
		return obj
	}
}

func (h *selectionComparisonHandle) attributes(vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("diff", func(call goja.FunctionCall) any {
			names, err := stringListArg(vm, "css-visual-diff.selectionComparison.attributes.diff", call.Argument(0))
			if err != nil {
				panic(typeMismatchError(vm, "css-visual-diff.selectionComparison.attributes.diff", "array of attribute names", call.Argument(0)))
			}
			return lowerJSON(filterMapDiffs(h.data.Attributes, names))
		})
		return obj
	}
}

func (h *selectionComparisonHandle) report(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("markdown", func() string { return renderSelectionComparisonMarkdown(h.data) })
		_ = obj.Set("writeMarkdown", func(path string) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.selectionComparison.report.writeMarkdown", func() (any, error) {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					return nil, err
				}
				return path, os.WriteFile(path, []byte(renderSelectionComparisonMarkdown(h.data)), 0o644)
			}, nil)
		})
		return obj
	}
}

func (h *selectionComparisonHandle) artifacts(ctx *engine.RuntimeModuleContext, vm *goja.Runtime) ProxyProperty {
	return func(receiver goja.Value) goja.Value {
		obj := vm.NewObject()
		_ = obj.Set("list", func() any { return lowerJSON(h.data.Artifacts) })
		_ = obj.Set("write", func(outDir string, rawNames goja.Value) goja.Value {
			return promiseValue(ctx, vm, "css-visual-diff.selectionComparison.artifacts.write", func() (any, error) {
				names, err := stringListArg(vm, "css-visual-diff.selectionComparison.artifacts.write", rawNames)
				if err != nil {
					return nil, err
				}
				return writeComparisonArtifacts(outDir, h.data, names)
			}, nil)
		})
		return obj
	}
}

func filterMapDiffs(diffs []service.MapValueDiff, names []string) []service.MapValueDiff {
	if len(names) == 0 {
		return diffs
	}
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
	}
	out := make([]service.MapValueDiff, 0, len(diffs))
	for _, diff := range diffs {
		if wanted[diff.Name] {
			out = append(out, diff)
		}
	}
	return out
}

func renderSelectionComparisonMarkdown(data service.SelectionComparisonData) string {
	var b strings.Builder
	b.WriteString("# Selection Comparison\n\n")
	if data.Name != "" {
		b.WriteString(fmt.Sprintf("- Name: %s\n", data.Name))
	}
	b.WriteString(fmt.Sprintf("- Left selector: `%s`\n", data.Left.Selector))
	b.WriteString(fmt.Sprintf("- Right selector: `%s`\n", data.Right.Selector))
	if data.Pixel != nil {
		b.WriteString(fmt.Sprintf("- Changed pixels: %d/%d (%.4f%%)\n", data.Pixel.ChangedPixels, data.Pixel.TotalPixels, data.Pixel.ChangedPercent))
	}
	b.WriteString(fmt.Sprintf("- Bounds changed: %t\n", data.Bounds.Changed))
	b.WriteString(fmt.Sprintf("- Text changed: %t\n", data.Text.Changed))
	b.WriteString(fmt.Sprintf("- Style changes: %d\n", len(data.Styles)))
	b.WriteString(fmt.Sprintf("- Attribute changes: %d\n\n", len(data.Attributes)))
	if len(data.Styles) > 0 {
		b.WriteString("## Style diffs\n\n| Property | Left | Right |\n| --- | --- | --- |\n")
		for _, diff := range data.Styles {
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", diff.Name, diff.Left, diff.Right))
		}
		b.WriteString("\n")
	}
	if len(data.Attributes) > 0 {
		b.WriteString("## Attribute diffs\n\n| Attribute | Left | Right |\n| --- | --- | --- |\n")
		for _, diff := range data.Attributes {
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", diff.Name, diff.Left, diff.Right))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func writeComparisonArtifacts(outDir string, data service.SelectionComparisonData, names []string) (map[string]any, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
	}
	if len(wanted) == 0 || wanted["json"] {
		bytes, err := jsonMarshalIndent(lowerJSON(data))
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(outDir, "compare.json"), bytes, 0o644); err != nil {
			return nil, err
		}
	}
	if len(wanted) == 0 || wanted["markdown"] {
		if err := os.WriteFile(filepath.Join(outDir, "compare.md"), []byte(renderSelectionComparisonMarkdown(data)), 0o644); err != nil {
			return nil, err
		}
	}
	written := []string{}
	if len(wanted) == 0 || wanted["json"] {
		written = append(written, filepath.Join(outDir, "compare.json"))
	}
	if len(wanted) == 0 || wanted["markdown"] {
		written = append(written, filepath.Join(outDir, "compare.md"))
	}
	return map[string]any{"outDir": outDir, "written": written}, nil
}
