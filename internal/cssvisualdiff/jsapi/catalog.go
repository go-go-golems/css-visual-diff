package jsapi

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
	"github.com/go-go-golems/go-go-goja/engine"
)

type catalogOptionsInput struct {
	Title        string `json:"title"`
	OutDir       string `json:"outDir"`
	ArtifactRoot string `json:"artifactRoot"`
	IndexName    string `json:"indexName"`
}

func newCatalogFromJS(raw map[string]any) (*service.Catalog, error) {
	input, err := decodeInto[catalogOptionsInput](raw)
	if err != nil {
		return nil, err
	}
	return service.NewCatalog(service.CatalogOptions{
		Title:        input.Title,
		OutDir:       input.OutDir,
		ArtifactRoot: input.ArtifactRoot,
		IndexName:    input.IndexName,
	})
}

func wrapCatalog(ctx *engine.RuntimeModuleContext, vm *goja.Runtime, catalog *service.Catalog) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("artifactDir", func(slug string) string {
		return catalog.ArtifactDir(slug)
	})
	_ = obj.Set("addTarget", func(raw map[string]any) (map[string]any, error) {
		target, err := decodeCatalogTarget(raw)
		if err != nil {
			return nil, err
		}
		return lowerCatalogTarget(catalog.AddTarget(target)), nil
	})
	_ = obj.Set("recordPreflight", func(rawTarget map[string]any, rawStatuses []map[string]any) (map[string]any, error) {
		target, err := decodeCatalogTarget(rawTarget)
		if err != nil {
			return nil, err
		}
		statuses, err := decodeCatalogSelectorStatuses(rawStatuses)
		if err != nil {
			return nil, err
		}
		record := catalog.RecordPreflight(target, statuses)
		return lowerCatalogPreflightRecord(record), nil
	})
	_ = obj.Set("addResult", func(rawTarget map[string]any, rawResult map[string]any) (map[string]any, error) {
		target, err := decodeCatalogTarget(rawTarget)
		if err != nil {
			return nil, err
		}
		result, err := decodeCatalogInspectResult(rawResult)
		if err != nil {
			return nil, err
		}
		record := catalog.AddResult(target, result)
		return lowerCatalogResultRecord(record), nil
	})
	_ = obj.Set("addFailure", func(call goja.FunctionCall) goja.Value {
		target, err := decodeCatalogTarget(call.Argument(0).Export())
		if err != nil {
			panic(vm.NewGoError(err))
		}
		failure := catalogFailureFromValue(vm, call.Argument(1))
		record := catalog.AddFailure(target, failure)
		return vm.ToValue(lowerCatalogFailureRecord(record))
	})
	_ = obj.Set("summary", func() map[string]any {
		return lowerCatalogSummary(catalog.Summary())
	})
	_ = obj.Set("manifest", func() map[string]any {
		return lowerCatalogManifest(catalog.Manifest())
	})
	_ = obj.Set("writeManifest", func() goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.catalog.writeManifest", func() (any, error) {
			return catalog.WriteManifest()
		}, nil)
	})
	_ = obj.Set("writeIndex", func() goja.Value {
		return promiseValue(ctx, vm, "css-visual-diff.catalog.writeIndex", func() (any, error) {
			return catalog.WriteIndex()
		}, nil)
	})
	return obj
}

type catalogTargetInput struct {
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	URL         string           `json:"url"`
	Selector    string           `json:"selector"`
	Viewport    service.Viewport `json:"viewport"`
	Description string           `json:"description"`
	Metadata    map[string]any   `json:"metadata"`
}

func decodeCatalogTarget(raw any) (service.CatalogTargetRecord, error) {
	input, err := decodeInto[catalogTargetInput](raw)
	if err != nil {
		return service.CatalogTargetRecord{}, err
	}
	return service.CatalogTargetRecord{
		Slug:        input.Slug,
		Name:        input.Name,
		URL:         input.URL,
		Selector:    input.Selector,
		Viewport:    input.Viewport,
		Description: input.Description,
		Metadata:    input.Metadata,
	}, nil
}

type catalogSelectorStatusInput struct {
	Name      string          `json:"name"`
	Selector  string          `json:"selector"`
	Source    string          `json:"source"`
	Exists    bool            `json:"exists"`
	Visible   bool            `json:"visible"`
	Bounds    *service.Bounds `json:"bounds"`
	TextStart string          `json:"textStart"`
	Error     string          `json:"error"`
}

func decodeCatalogSelectorStatuses(raw []map[string]any) ([]service.SelectorStatus, error) {
	inputs, err := decodeInto[[]catalogSelectorStatusInput](raw)
	if err != nil {
		return nil, err
	}
	ret := make([]service.SelectorStatus, 0, len(inputs))
	for _, input := range inputs {
		ret = append(ret, service.SelectorStatus{
			Name:      input.Name,
			Selector:  input.Selector,
			Source:    input.Source,
			Exists:    input.Exists,
			Visible:   input.Visible,
			Bounds:    input.Bounds,
			TextStart: input.TextStart,
			Error:     input.Error,
		})
	}
	return ret, nil
}

type catalogInspectResultInput struct {
	OutputDir string                        `json:"outputDir"`
	Results   []catalogInspectArtifactInput `json:"results"`
}

type catalogInspectArtifactInput struct {
	Metadata    catalogInspectMetadataInput `json:"metadata"`
	Style       *catalogStyleSnapshotInput  `json:"style"`
	Screenshot  string                      `json:"screenshot"`
	HTML        string                      `json:"html"`
	InspectJSON string                      `json:"inspectJson"`
}

type catalogInspectMetadataInput struct {
	Side           string           `json:"side"`
	TargetName     string           `json:"targetName"`
	URL            string           `json:"url"`
	Viewport       service.Viewport `json:"viewport"`
	Name           string           `json:"name"`
	Selector       string           `json:"selector"`
	SelectorSource string           `json:"selectorSource"`
	RootSelector   string           `json:"rootSelector"`
	PrepareType    string           `json:"prepareType"`
	Format         string           `json:"format"`
	CreatedAt      string           `json:"createdAt"`
}

type catalogStyleSnapshotInput struct {
	Exists     bool              `json:"exists"`
	Computed   map[string]string `json:"computed"`
	Bounds     *service.Bounds   `json:"bounds"`
	Attributes map[string]string `json:"attributes"`
}

func decodeCatalogInspectResult(raw map[string]any) (service.InspectResult, error) {
	input, err := decodeInto[catalogInspectResultInput](raw)
	if err != nil {
		return service.InspectResult{}, err
	}
	result := service.InspectResult{OutputDir: input.OutputDir}
	for _, artifact := range input.Results {
		result.Results = append(result.Results, service.InspectArtifactResult{
			Metadata:    decodeCatalogInspectMetadata(artifact.Metadata),
			Style:       decodeCatalogStyleSnapshot(artifact.Style),
			Screenshot:  artifact.Screenshot,
			HTML:        artifact.HTML,
			InspectJSON: artifact.InspectJSON,
		})
	}
	return result, nil
}

func decodeCatalogInspectMetadata(input catalogInspectMetadataInput) service.InspectMetadata {
	createdAt, _ := time.Parse(time.RFC3339Nano, input.CreatedAt)
	return service.InspectMetadata{
		Side:           input.Side,
		TargetName:     input.TargetName,
		URL:            input.URL,
		Viewport:       input.Viewport,
		Name:           input.Name,
		Selector:       input.Selector,
		SelectorSource: input.SelectorSource,
		RootSelector:   input.RootSelector,
		PrepareType:    input.PrepareType,
		Format:         input.Format,
		CreatedAt:      createdAt,
	}
}

func decodeCatalogStyleSnapshot(input *catalogStyleSnapshotInput) *service.StyleSnapshot {
	if input == nil {
		return nil
	}
	return &service.StyleSnapshot{Exists: input.Exists, Computed: input.Computed, Bounds: input.Bounds, Attributes: input.Attributes}
}

func catalogFailureFromValue(vm *goja.Runtime, value goja.Value) service.CatalogFailureRecord {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return service.CatalogFailureRecord{Message: "unknown failure"}
	}
	obj := value.ToObject(vm)
	return service.CatalogFailureRecord{
		Name:      valueString(obj.Get("name")),
		Code:      valueString(obj.Get("code")),
		Operation: valueString(obj.Get("operation")),
		Message:   valueString(obj.Get("message")),
	}
}

func lowerCatalogManifest(manifest service.CatalogManifest) map[string]any {
	return map[string]any{
		"schemaVersion": manifest.SchemaVersion,
		"title":         manifest.Title,
		"outDir":        manifest.OutDir,
		"artifactRoot":  manifest.ArtifactRoot,
		"createdAt":     manifest.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":     manifest.UpdatedAt.Format(time.RFC3339Nano),
		"targets":       lowerCatalogTargets(manifest.Targets),
		"preflights":    lowerCatalogPreflightRecords(manifest.Preflights),
		"results":       lowerCatalogResultRecords(manifest.Results),
		"failures":      lowerCatalogFailureRecords(manifest.Failures),
		"summary":       lowerCatalogSummary(manifest.Summary),
	}
}

func lowerCatalogTargets(targets []service.CatalogTargetRecord) []map[string]any {
	ret := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		ret = append(ret, lowerCatalogTarget(target))
	}
	return ret
}

func lowerCatalogTarget(target service.CatalogTargetRecord) map[string]any {
	return map[string]any{
		"slug":        target.Slug,
		"name":        target.Name,
		"url":         target.URL,
		"selector":    target.Selector,
		"viewport":    lowerViewport(target.Viewport),
		"description": target.Description,
		"metadata":    target.Metadata,
	}
}

func lowerCatalogPreflightRecords(records []service.CatalogPreflightRecord) []map[string]any {
	ret := make([]map[string]any, 0, len(records))
	for _, record := range records {
		ret = append(ret, lowerCatalogPreflightRecord(record))
	}
	return ret
}

func lowerCatalogPreflightRecord(record service.CatalogPreflightRecord) map[string]any {
	return map[string]any{
		"target":     lowerCatalogTarget(record.Target),
		"statuses":   lowerSelectorStatuses(record.Statuses),
		"recordedAt": record.RecordedAt.Format(time.RFC3339Nano),
	}
}

func lowerCatalogResultRecords(records []service.CatalogResultRecord) []map[string]any {
	ret := make([]map[string]any, 0, len(records))
	for _, record := range records {
		ret = append(ret, lowerCatalogResultRecord(record))
	}
	return ret
}

func lowerCatalogResultRecord(record service.CatalogResultRecord) map[string]any {
	return map[string]any{
		"target":     lowerCatalogTarget(record.Target),
		"result":     lowerInspectResult(record.Result),
		"recordedAt": record.RecordedAt.Format(time.RFC3339Nano),
	}
}

func lowerCatalogFailureRecords(records []service.CatalogFailureRecord) []map[string]any {
	ret := make([]map[string]any, 0, len(records))
	for _, record := range records {
		ret = append(ret, lowerCatalogFailureRecord(record))
	}
	return ret
}

func lowerCatalogFailureRecord(record service.CatalogFailureRecord) map[string]any {
	return map[string]any{
		"target":     lowerCatalogTarget(record.Target),
		"name":       record.Name,
		"code":       record.Code,
		"operation":  record.Operation,
		"message":    record.Message,
		"recordedAt": record.RecordedAt.Format(time.RFC3339Nano),
	}
}

func lowerCatalogSummary(summary service.CatalogSummary) map[string]any {
	return map[string]any{
		"targetCount":    summary.TargetCount,
		"preflightCount": summary.PreflightCount,
		"resultCount":    summary.ResultCount,
		"failureCount":   summary.FailureCount,
		"artifactCount":  summary.ArtifactCount,
	}
}

func valueString(value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	if exported := value.Export(); exported != nil {
		return fmt.Sprint(exported)
	}
	return value.String()
}
