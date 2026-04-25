package jsapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

type selectionComparisonHandle struct {
	data service.SelectionComparisonData
}

func installCompareAPI(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, exports *goja.Object) {
	compare := vm.NewObject()
	_ = compare.Set("region", func(call goja.FunctionCall) goja.Value {
		opts := decodeCompareRegionOptions(vm, call.Argument(0))
		return promiseValue(ctx, vm, "css-visual-diff.compare.region", func() (any, error) {
			return compareRegion(opts)
		}, func(vm *goja.Runtime, value any) goja.Value {
			return wrapSelectionComparison(ctx, vm, value.(service.SelectionComparisonData))
		})
	})
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

type compareRegionOptions struct {
	Left       *locatorHandle
	Right      *locatorHandle
	Name       string
	Threshold  int
	Inspect    string
	OutDir     string
	StyleProps []string
	Attributes []string
}

func decodeCompareRegionOptions(vm *goja.Runtime, value goja.Value) compareRegionOptions {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		panic(typeMismatchError(vm, "css-visual-diff.compare.region", "options object with left and right locators", value))
	}
	obj := value.ToObject(vm)
	left := mustUnwrapProxyBacking[locatorHandle](vm, defaultProxyRegistry, "css-visual-diff.compare.region", obj.Get("left"), "cvd.locator")
	right := mustUnwrapProxyBacking[locatorHandle](vm, defaultProxyRegistry, "css-visual-diff.compare.region", obj.Get("right"), "cvd.locator")
	raw := exportOptionalObject(vm, "css-visual-diff.compare.region", value)
	threshold := 30
	if rawThreshold, ok := raw["threshold"]; ok {
		threshold = intNumber(rawThreshold)
	}
	inspect := service.CollectInspectRich
	if rawInspect, ok := raw["inspect"].(string); ok && rawInspect != "" {
		inspect = rawInspect
	}
	styleProps, _ := stringSliceFromAny(raw["styleProps"])
	if len(styleProps) == 0 {
		styleProps, _ = stringSliceFromAny(raw["styles"])
	}
	attributes, _ := stringSliceFromAny(raw["attributes"])
	return compareRegionOptions{
		Left:       left,
		Right:      right,
		Name:       stringFromAny(raw["name"]),
		Threshold:  threshold,
		Inspect:    inspect,
		OutDir:     stringFromAny(raw["outDir"]),
		StyleProps: styleProps,
		Attributes: attributes,
	}
}

func compareRegion(opts compareRegionOptions) (service.SelectionComparisonData, error) {
	if opts.Left == nil || opts.Right == nil {
		return service.SelectionComparisonData{}, fmt.Errorf("css-visual-diff.compare.region requires left and right page.locator(selector) handles")
	}
	if opts.Inspect == "" {
		opts.Inspect = service.CollectInspectRich
	}
	outDir := opts.OutDir
	if outDir == "" {
		outDir = filepath.Join(os.TempDir(), fmt.Sprintf("cssvd-compare-region-%d", time.Now().UnixNano()))
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return service.SelectionComparisonData{}, err
	}
	leftShot := filepath.Join(outDir, "left_region.png")
	rightShot := filepath.Join(outDir, "right_region.png")
	left, err := collectAndScreenshotRegion(opts.Left, opts.Name, opts.Inspect, opts.StyleProps, opts.Attributes, leftShot)
	if err != nil {
		return service.SelectionComparisonData{}, err
	}
	right, err := collectAndScreenshotRegion(opts.Right, opts.Name, opts.Inspect, opts.StyleProps, opts.Attributes, rightShot)
	if err != nil {
		return service.SelectionComparisonData{}, err
	}
	return service.CompareSelections(left, right, service.CompareSelectionOptions{
		Name:               opts.Name,
		Threshold:          opts.Threshold,
		StyleProps:         opts.StyleProps,
		Attributes:         opts.Attributes,
		DiffOnlyPath:       filepath.Join(outDir, "diff_only.png"),
		DiffComparisonPath: filepath.Join(outDir, "diff_comparison.png"),
	})
}

func collectAndScreenshotRegion(locator *locatorHandle, name, inspect string, styleProps, attributes []string, screenshotPath string) (service.SelectionData, error) {
	value, err := locator.page.runExclusive(func() (any, error) {
		opts := service.CollectOptions{Name: name, Inspect: inspect, StyleProps: styleProps, Attributes: attributes}
		selected, err := service.CollectSelection(locator.page.page.Page(), locator.spec(), opts)
		if err != nil {
			return nil, err
		}
		if selected.Exists && selected.Visible {
			if err := os.MkdirAll(filepath.Dir(screenshotPath), 0o755); err != nil {
				return nil, err
			}
			if err := locator.page.page.Page().Screenshot(locator.selector, screenshotPath); err != nil {
				return nil, err
			}
			selected.Screenshot = &service.ScreenshotDescriptor{Path: screenshotPath}
		}
		return selected, nil
	})
	if err != nil {
		return service.SelectionData{}, err
	}
	return value.(service.SelectionData), nil
}

func intNumber(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	default:
		return 0
	}
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func stringSliceFromAny(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, false
		}
		out = append(out, text)
	}
	return out, true
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
		fmt.Fprintf(&b, "- Name: %s\n", data.Name)
	}
	fmt.Fprintf(&b, "- Left selector: `%s`\n", data.Left.Selector)
	fmt.Fprintf(&b, "- Right selector: `%s`\n", data.Right.Selector)
	if data.Pixel != nil {
		fmt.Fprintf(&b, "- Changed pixels: %d/%d (%.4f%%)\n", data.Pixel.ChangedPixels, data.Pixel.TotalPixels, data.Pixel.ChangedPercent)
	}
	fmt.Fprintf(&b, "- Bounds changed: %t\n", data.Bounds.Changed)
	fmt.Fprintf(&b, "- Text changed: %t\n", data.Text.Changed)
	fmt.Fprintf(&b, "- Style changes: %d\n", len(data.Styles))
	fmt.Fprintf(&b, "- Attribute changes: %d\n\n", len(data.Attributes))
	if len(data.Styles) > 0 {
		b.WriteString("## Style diffs\n\n| Property | Left | Right |\n| --- | --- | --- |\n")
		for _, diff := range data.Styles {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", diff.Name, diff.Left, diff.Right)
		}
		b.WriteString("\n")
	}
	if len(data.Attributes) > 0 {
		b.WriteString("## Attribute diffs\n\n| Attribute | Left | Right |\n| --- | --- | --- |\n")
		for _, diff := range data.Attributes {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", diff.Name, diff.Left, diff.Right)
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
	wants := func(name string) bool { return len(wanted) == 0 || wanted[name] }

	result := map[string]any{"outDir": outDir}
	written := []string{}
	if wants("json") {
		path := filepath.Join(outDir, "compare.json")
		bytes, err := jsonMarshalIndent(lowerJSON(data))
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, bytes, 0o644); err != nil {
			return nil, err
		}
		result["json"] = path
		written = append(written, path)
	}
	if wants("markdown") {
		path := filepath.Join(outDir, "compare.md")
		if err := os.WriteFile(path, []byte(renderSelectionComparisonMarkdown(data)), 0o644); err != nil {
			return nil, err
		}
		result["markdown"] = path
		written = append(written, path)
	}
	addKnownComparisonArtifactPaths(result, outDir, data)
	result["written"] = written
	return result, nil
}

func addKnownComparisonArtifactPaths(result map[string]any, outDir string, data service.SelectionComparisonData) {
	for _, artifact := range data.Artifacts {
		switch artifact.Name {
		case "diffOnly", "diffComparison", "leftRegion", "rightRegion":
			if artifact.Path != "" {
				result[artifact.Name] = artifact.Path
			}
		}
	}
	if data.Pixel != nil {
		if data.Pixel.DiffOnlyPath != "" {
			result["diffOnly"] = data.Pixel.DiffOnlyPath
		}
		if data.Pixel.DiffComparisonPath != "" {
			result["diffComparison"] = data.Pixel.DiffComparisonPath
		}
	}
	addPathIfExists(result, "leftRegion", filepath.Join(outDir, "left_region.png"))
	addPathIfExists(result, "rightRegion", filepath.Join(outDir, "right_region.png"))
	addPathIfExists(result, "diffOnly", filepath.Join(outDir, "diff_only.png"))
	addPathIfExists(result, "diffComparison", filepath.Join(outDir, "diff_comparison.png"))
}

func addPathIfExists(result map[string]any, key, path string) {
	if _, ok := result[key]; ok {
		return
	}
	if _, err := os.Stat(path); err == nil {
		result[key] = path
	}
}
