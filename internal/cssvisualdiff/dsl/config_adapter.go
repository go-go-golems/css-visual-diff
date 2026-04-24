package dsl

import (
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

func lowerConfig(cfg *config.Config) map[string]any {
	if cfg == nil {
		return map[string]any{}
	}
	return map[string]any{
		"metadata": lowerConfigMetadata(cfg.Metadata),
		"original": lowerConfigTarget(cfg.Original),
		"react":    lowerConfigTarget(cfg.React),
		"sections": lowerConfigSections(cfg.Sections),
		"styles":   lowerConfigStyles(cfg.Styles),
		"output":   lowerConfigOutput(cfg.Output),
		"modes":    cfg.Modes,
	}
}

func lowerConfigMetadata(metadata config.Metadata) map[string]any {
	return map[string]any{
		"slug":        metadata.Slug,
		"title":       metadata.Title,
		"description": metadata.Description,
		"goal":        metadata.Goal,
	}
}

func lowerConfigTarget(target config.Target) map[string]any {
	ret := map[string]any{
		"name":         target.Name,
		"url":          target.URL,
		"waitMs":       target.WaitMS,
		"viewport":     lowerViewport(target.Viewport),
		"rootSelector": target.RootSelector,
	}
	if target.Prepare != nil {
		ret["prepare"] = lowerPrepareSpec(*target.Prepare)
	}
	return ret
}

func lowerPrepareSpec(prepare config.PrepareSpec) map[string]any {
	return map[string]any{
		"type":             prepare.Type,
		"script":           prepare.Script,
		"scriptFile":       prepare.ScriptFile,
		"waitFor":          prepare.WaitFor,
		"waitForTimeoutMs": prepare.WaitForTimeoutMS,
		"afterWaitMs":      prepare.AfterWaitMS,
		"component":        prepare.Component,
		"props":            prepare.Props,
		"rootSelector":     prepare.RootSelector,
		"width":            prepare.Width,
		"minHeight":        prepare.MinHeight,
		"background":       prepare.Background,
	}
}

func lowerConfigSections(sections []config.SectionSpec) []map[string]any {
	ret := make([]map[string]any, 0, len(sections))
	for _, section := range sections {
		ret = append(ret, map[string]any{
			"name":             section.Name,
			"selector":         section.Selector,
			"selectorOriginal": section.SelectorOriginal,
			"selectorReact":    section.SelectorReact,
			"ocrQuestion":      section.OCRQuestion,
		})
	}
	return ret
}

func lowerConfigStyles(styles []config.StyleSpec) []map[string]any {
	ret := make([]map[string]any, 0, len(styles))
	for _, style := range styles {
		ret = append(ret, map[string]any{
			"name":             style.Name,
			"selector":         style.Selector,
			"selectorOriginal": style.SelectorOriginal,
			"selectorReact":    style.SelectorReact,
			"props":            style.Props,
			"includeBounds":    style.IncludeBounds,
			"attributes":       style.Attributes,
			"report":           style.Report,
		})
	}
	return ret
}

func lowerConfigOutput(output config.OutputSpec) map[string]any {
	return map[string]any{
		"dir":               output.Dir,
		"writeJson":         output.WriteJSON,
		"writeMarkdown":     output.WriteMarkdown,
		"writePngs":         output.WritePNGs,
		"writePreparedHtml": output.WritePreparedHTML,
		"writeInspectJson":  output.WriteInspectJSON,
		"validatePngs":      output.ValidatePNGs,
	}
}
