package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

const (
	InspectFormatBundle       = "bundle"
	InspectFormatPNG          = "png"
	InspectFormatHTML         = "html"
	InspectFormatCSSJSON      = "css-json"
	InspectFormatCSSMarkdown  = "css-md"
	InspectFormatInspectJSON  = "inspect-json"
	InspectFormatMetadataJSON = "metadata-json"
)

type InspectRequest struct {
	Name       string   `json:"name"`
	Selector   string   `json:"selector"`
	Props      []string `json:"props,omitempty"`
	Attributes []string `json:"attributes,omitempty"`
	Source     string   `json:"source"`
}

type InspectMetadata struct {
	Side           string          `json:"side"`
	TargetName     string          `json:"target_name"`
	URL            string          `json:"url"`
	Viewport       config.Viewport `json:"viewport"`
	Name           string          `json:"name"`
	Selector       string          `json:"selector"`
	SelectorSource string          `json:"selector_source"`
	RootSelector   string          `json:"root_selector,omitempty"`
	PrepareType    string          `json:"prepare_type,omitempty"`
	Format         string          `json:"format"`
	CreatedAt      time.Time       `json:"created_at"`
}

type InspectArtifactResult struct {
	Metadata    InspectMetadata `json:"metadata"`
	Style       *StyleSnapshot  `json:"style,omitempty"`
	Screenshot  string          `json:"screenshot,omitempty"`
	HTML        string          `json:"html,omitempty"`
	InspectJSON string          `json:"inspect_json,omitempty"`
}

type InspectAllOptions struct {
	OutDir     string
	Format     string
	OutputFile string
}

type InspectResult struct {
	OutputDir string                  `json:"output_dir,omitempty"`
	Results   []InspectArtifactResult `json:"results"`
}

func InspectPreparedPage(page *driver.Page, target config.Target, side string, requests []InspectRequest, opts InspectAllOptions) (InspectResult, error) {
	result := InspectResult{OutputDir: opts.OutDir}
	if opts.OutputFile != "" && len(requests) != 1 {
		return result, fmt.Errorf("outputFile requires exactly one inspect request, got %d", len(requests))
	}
	for _, req := range requests {
		destDir := opts.OutDir
		if opts.OutputFile == "" && len(requests) > 1 {
			destDir = filepath.Join(opts.OutDir, SanitizeName(req.Name))
		}
		artifact, err := WriteInspectArtifacts(page, target, side, req, opts.Format, destDir, opts.OutputFile)
		if err != nil {
			return result, err
		}
		result.Results = append(result.Results, artifact)
	}
	if opts.OutputFile == "" && len(result.Results) > 1 {
		if err := WriteInspectIndex(opts.OutDir, result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func WriteInspectArtifacts(page *driver.Page, target config.Target, side string, req InspectRequest, format, outDir, outputFile string) (InspectArtifactResult, error) {
	format, err := CanonicalInspectFormat(format)
	if err != nil {
		return InspectArtifactResult{}, err
	}
	metadata := InspectMetadata{
		Side:           side,
		TargetName:     target.Name,
		URL:            target.URL,
		Viewport:       target.Viewport,
		Name:           req.Name,
		Selector:       req.Selector,
		SelectorSource: req.Source,
		RootSelector:   RootSelectorForTarget(target),
		Format:         format,
		CreatedAt:      time.Now().UTC(),
	}
	if target.Prepare != nil {
		metadata.PrepareType = target.Prepare.Type
	}
	artifact := InspectArtifactResult{Metadata: metadata}

	if outputFile != "" {
		if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil && filepath.Dir(outputFile) != "." {
			return artifact, err
		}
		return WriteSingleInspectArtifact(page, req, metadata, format, outputFile)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return artifact, err
	}

	metadataPath := filepath.Join(outDir, "metadata.json")
	if err := WriteJSON(metadataPath, metadata); err != nil {
		return artifact, err
	}

	if InspectFormatRequiresExistingSelector(format) {
		if err := EnsureInspectSelectorExists(page, req); err != nil {
			return artifact, err
		}
	}

	if format == InspectFormatBundle || format == InspectFormatHTML {
		htmlPath := filepath.Join(outDir, "prepared.html")
		if err := WritePreparedHTML(page, req.Selector, htmlPath); err != nil {
			return artifact, err
		}
		artifact.HTML = htmlPath
	}
	if format == InspectFormatBundle || format == InspectFormatPNG {
		pngPath := filepath.Join(outDir, "screenshot.png")
		if err := page.Screenshot(req.Selector, pngPath); err != nil {
			return artifact, err
		}
		artifact.Screenshot = pngPath
	}
	if format == InspectFormatBundle || format == InspectFormatCSSJSON || format == InspectFormatCSSMarkdown {
		style, err := EvaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if format == InspectFormatBundle || format == InspectFormatCSSJSON {
			if err := WriteJSON(filepath.Join(outDir, "computed-css.json"), style); err != nil {
				return artifact, err
			}
		}
		if format == InspectFormatBundle || format == InspectFormatCSSMarkdown {
			if err := WriteInspectCSSMarkdown(filepath.Join(outDir, "computed-css.md"), req, style); err != nil {
				return artifact, err
			}
		}
	}
	if format == InspectFormatBundle || format == InspectFormatInspectJSON {
		inspectPath := filepath.Join(outDir, "inspect.json")
		if err := WriteInspectJSON(page, req.Selector, inspectPath); err != nil {
			return artifact, err
		}
		artifact.InspectJSON = inspectPath
	}
	return artifact, nil
}

func InspectFormatRequiresExistingSelector(format string) bool {
	switch format {
	case InspectFormatBundle, InspectFormatPNG, InspectFormatHTML, InspectFormatInspectJSON:
		return true
	default:
		return false
	}
}

func EnsureInspectSelectorExists(page *driver.Page, req InspectRequest) error {
	statuses, err := PreflightProbes(page, []ProbeSpec{{Name: req.Name, Selector: req.Selector, Source: req.Source}})
	if err != nil {
		return fmt.Errorf("preflight selector %q for %s %q: %w", req.Selector, req.Source, req.Name, err)
	}
	if len(statuses) != 1 {
		return fmt.Errorf("preflight selector %q for %s %q returned %d statuses", req.Selector, req.Source, req.Name, len(statuses))
	}
	status := statuses[0]
	if status.Error != "" {
		return fmt.Errorf("preflight selector %q for %s %q: %s", req.Selector, req.Source, req.Name, status.Error)
	}
	if !status.Exists {
		return fmt.Errorf("%s %q selector did not match: %s", req.Source, req.Name, req.Selector)
	}
	return nil
}

func WriteSingleInspectArtifact(page *driver.Page, req InspectRequest, metadata InspectMetadata, format, path string) (InspectArtifactResult, error) {
	artifact := InspectArtifactResult{Metadata: metadata}
	if InspectFormatRequiresExistingSelector(format) {
		if err := EnsureInspectSelectorExists(page, req); err != nil {
			return artifact, err
		}
	}
	switch format {
	case InspectFormatPNG:
		if err := page.Screenshot(req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.Screenshot = path
	case InspectFormatHTML:
		if err := WritePreparedHTML(page, req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.HTML = path
	case InspectFormatCSSJSON:
		style, err := EvaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if err := WriteJSON(path, style); err != nil {
			return artifact, err
		}
	case InspectFormatCSSMarkdown:
		style, err := EvaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if err := WriteInspectCSSMarkdown(path, req, style); err != nil {
			return artifact, err
		}
	case InspectFormatInspectJSON:
		if err := WriteInspectJSON(page, req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.InspectJSON = path
	case InspectFormatMetadataJSON:
		if err := WriteJSON(path, metadata); err != nil {
			return artifact, err
		}
	default:
		return artifact, fmt.Errorf("format %q does not support --output-file", format)
	}
	return artifact, nil
}

func WriteInspectCSSMarkdown(path string, req InspectRequest, style StyleSnapshot) error {
	content := fmt.Sprintf("# css-visual-diff CSS Inspect: %s\n\n", req.Name)
	content += fmt.Sprintf("Selector: `%s`\n\n", req.Selector)
	content += fmt.Sprintf("Exists: `%t`\n\n", style.Exists)
	if style.Bounds != nil {
		content += fmt.Sprintf("Bounds: x=%.2f y=%.2f width=%.2f height=%.2f\n\n", style.Bounds.X, style.Bounds.Y, style.Bounds.Width, style.Bounds.Height)
	}
	content += "| Property | Value |\n| --- | --- |\n"
	for _, prop := range req.Props {
		content += fmt.Sprintf("| %s | %s |\n", prop, style.Computed[prop])
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func WriteInspectIndex(outDir string, result InspectResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(outDir, "index.json"), result); err != nil {
		return err
	}
	content := "# css-visual-diff Inspect Index\n\n"
	content += "| Name | Source | Selector | Screenshot | HTML | CSS JSON | CSS Markdown | Inspect JSON |\n"
	content += "| --- | --- | --- | --- | --- | --- | --- | --- |\n"
	for _, r := range result.Results {
		name := r.Metadata.Name
		base := SanitizeName(name)
		content += fmt.Sprintf("| %s | %s | `%s` | %s | %s | %s | %s | %s |\n",
			name,
			r.Metadata.SelectorSource,
			r.Metadata.Selector,
			filepath.Join(base, "screenshot.png"),
			filepath.Join(base, "prepared.html"),
			filepath.Join(base, "computed-css.json"),
			filepath.Join(base, "computed-css.md"),
			filepath.Join(base, "inspect.json"),
		)
	}
	return os.WriteFile(filepath.Join(outDir, "index.md"), []byte(content), 0o644)
}

func WriteJSON(path string, data any) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func WritePreparedHTML(page *driver.Page, selector, path string) error {
	selectorJSON, _ := json.Marshal(selector)
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  const el = selector ? document.querySelector(selector) : document.documentElement;
	  return el ? el.outerHTML : document.documentElement.outerHTML;
	})()`, string(selectorJSON))
	var html string
	if err := page.Evaluate(script, &html); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(html), 0o644)
}

func WriteInspectJSON(page *driver.Page, selector, path string) error {
	selectorJSON, _ := json.Marshal(selector)
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  const root = selector ? document.querySelector(selector) : document.documentElement;
	  const inspect = (el, depth = 0) => {
	    if (!el || depth > 8) return null;
	    const rect = el.getBoundingClientRect();
	    const style = window.getComputedStyle(el);
	    const attrs = {};
	    for (const attr of el.attributes || []) attrs[attr.name] = attr.value;
	    return {
	      tag: el.tagName ? el.tagName.toLowerCase() : '',
	      id: el.id || '',
	      class_name: el.className || '',
	      attributes: attrs,
	      text: (el.children && el.children.length === 0 ? el.textContent : '').trim().slice(0, 200),
	      bounds: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
	      computed: {
	        display: style.display,
	        position: style.position,
	        boxSizing: style.boxSizing,
	        width: style.width,
	        height: style.height,
	        margin: style.margin,
	        padding: style.padding,
	        color: style.color,
	        backgroundColor: style.backgroundColor,
	        fontFamily: style.fontFamily,
	        fontSize: style.fontSize,
	        fontWeight: style.fontWeight,
	        lineHeight: style.lineHeight,
	        border: style.border,
	        borderRadius: style.borderRadius,
	        gap: style.gap,
	        gridTemplateColumns: style.gridTemplateColumns,
	        flexDirection: style.flexDirection,
	        alignItems: style.alignItems,
	        justifyContent: style.justifyContent
	      },
	      children: Array.from(el.children || []).map((child) => inspect(child, depth + 1)).filter(Boolean)
	    };
	  };
	  return inspect(root);
	})()`, string(selectorJSON))
	var out any
	if err := page.Evaluate(script, &out); err != nil {
		return err
	}
	return WriteJSON(path, out)
}

func RootSelectorForTarget(target config.Target) string {
	if target.RootSelector != "" {
		return target.RootSelector
	}
	if target.Prepare != nil && target.Prepare.RootSelector != "" {
		return target.Prepare.RootSelector
	}
	return ""
}

func SanitizeName(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '-' || r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

func CanonicalInspectFormat(format string) (string, error) {
	switch format {
	case "", InspectFormatBundle:
		return InspectFormatBundle, nil
	case InspectFormatPNG, "screenshot":
		return InspectFormatPNG, nil
	case InspectFormatHTML:
		return InspectFormatHTML, nil
	case InspectFormatCSSJSON:
		return InspectFormatCSSJSON, nil
	case InspectFormatCSSMarkdown, "css", "css-markdown":
		return InspectFormatCSSMarkdown, nil
	case InspectFormatInspectJSON:
		return InspectFormatInspectJSON, nil
	case InspectFormatMetadataJSON:
		return InspectFormatMetadataJSON, nil
	default:
		return "", fmt.Errorf("unsupported inspect format %q", format)
	}
}
