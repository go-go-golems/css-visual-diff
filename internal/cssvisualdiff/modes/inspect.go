package modes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// InspectOptions describes a single-side inspection run against an existing
// css-visual-diff config. It intentionally uses the current low-level config
// schema: sections are screenshot regions and styles are computed-CSS probes.
type InspectOptions struct {
	Side string

	Root        bool
	Section     string
	Style       string
	Selector    string
	AllSections bool
	AllStyles   bool

	Props      []string
	Attributes []string

	OutDir     string
	Format     string
	OutputFile string
}

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

type InspectResult struct {
	OutputDir string                  `json:"output_dir,omitempty"`
	Results   []InspectArtifactResult `json:"results"`
}

func Inspect(ctx context.Context, cfg *config.Config, opts InspectOptions) (InspectResult, error) {
	if err := validateInspectOptions(opts); err != nil {
		return InspectResult{}, err
	}

	target, side, err := inspectTargetForSide(cfg, opts.Side)
	if err != nil {
		return InspectResult{}, err
	}

	requests, err := BuildInspectRequests(cfg, opts)
	if err != nil {
		return InspectResult{}, err
	}
	if opts.OutputFile != "" && len(requests) != 1 {
		return InspectResult{}, fmt.Errorf("--output-file supports exactly one inspect request")
	}

	format := normalizeInspectFormat(opts.Format, opts.OutputFile)
	outDir := opts.OutDir
	if outDir == "" && opts.OutputFile == "" {
		outDir = filepath.Join(cfg.Output.Dir, "inspect", side)
	}

	browser, err := driver.NewBrowser(ctx)
	if err != nil {
		return InspectResult{}, err
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return InspectResult{}, err
	}
	defer page.Close()

	if err := page.SetViewport(target.Viewport.Width, target.Viewport.Height); err != nil {
		return InspectResult{}, err
	}
	if err := page.Goto(target.URL); err != nil {
		return InspectResult{}, err
	}
	if target.WaitMS > 0 {
		page.Wait(time.Duration(target.WaitMS) * time.Millisecond)
	}
	if err := prepareTarget(page, target); err != nil {
		return InspectResult{}, err
	}

	result := InspectResult{OutputDir: outDir}
	for _, req := range requests {
		destDir := outDir
		if opts.OutputFile == "" && len(requests) > 1 {
			destDir = filepath.Join(outDir, sanitizeName(req.Name))
		}
		artifact, err := writeInspectArtifacts(page, target, side, req, format, destDir, opts.OutputFile)
		if err != nil {
			return result, err
		}
		result.Results = append(result.Results, artifact)
	}
	if opts.OutputFile == "" && len(result.Results) > 1 {
		if err := writeInspectIndex(outDir, result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func validateInspectOptions(opts InspectOptions) error {
	format := normalizeInspectFormat(opts.Format, opts.OutputFile)
	if _, err := canonicalInspectFormat(format); err != nil {
		return err
	}
	if opts.OutputFile != "" && format == InspectFormatBundle {
		return fmt.Errorf("--output-file requires a single-file --format, not bundle")
	}
	selected := 0
	for _, ok := range []bool{opts.Root, opts.Section != "", opts.Style != "", opts.Selector != "", opts.AllSections, opts.AllStyles} {
		if ok {
			selected++
		}
	}
	if selected != 1 {
		return fmt.Errorf("provide exactly one of --root, --section, --style, --selector, --all-sections, --all-styles")
	}
	return nil
}

func BuildInspectRequests(cfg *config.Config, opts InspectOptions) ([]InspectRequest, error) {
	if err := validateInspectOptions(opts); err != nil {
		return nil, err
	}
	_, side, err := inspectTargetForSide(cfg, opts.Side)
	if err != nil {
		return nil, err
	}
	props := opts.Props
	if len(props) == 0 {
		props = defaultInspectProps()
	}
	attrs := opts.Attributes
	if len(attrs) == 0 {
		attrs = defaultInspectAttributes()
	}

	switch {
	case opts.Root:
		target, _, _ := inspectTargetForSide(cfg, opts.Side)
		selector := rootSelectorForTarget(target)
		if selector == "" {
			return nil, fmt.Errorf("%s target has no root_selector", side)
		}
		return []InspectRequest{{Name: "root", Selector: selector, Props: props, Attributes: attrs, Source: "root"}}, nil
	case opts.Section != "":
		section, ok := findInspectSection(cfg.Sections, opts.Section)
		if !ok {
			return nil, fmt.Errorf("section %q not found; available sections: %s", opts.Section, inspectSectionNames(cfg.Sections))
		}
		selector := selectorForSection(section, side)
		if selector == "" {
			return nil, fmt.Errorf("section %q has no selector for side %s", opts.Section, side)
		}
		return []InspectRequest{{Name: section.Name, Selector: selector, Props: props, Attributes: attrs, Source: "section"}}, nil
	case opts.Style != "":
		style, ok := findInspectStyle(cfg.Styles, opts.Style)
		if !ok {
			return nil, fmt.Errorf("style %q not found; available styles: %s", opts.Style, inspectStyleNames(cfg.Styles))
		}
		styleProps := style.Props
		if len(opts.Props) > 0 {
			styleProps = opts.Props
		}
		if len(styleProps) == 0 {
			styleProps = defaultInspectProps()
		}
		styleAttrs := style.Attributes
		if len(opts.Attributes) > 0 {
			styleAttrs = opts.Attributes
		}
		if len(styleAttrs) == 0 {
			styleAttrs = defaultInspectAttributes()
		}
		selector := selectorForStyleSide(style, side)
		if selector == "" {
			return nil, fmt.Errorf("style %q has no selector for side %s", opts.Style, side)
		}
		return []InspectRequest{{Name: style.Name, Selector: selector, Props: styleProps, Attributes: styleAttrs, Source: "style"}}, nil
	case opts.Selector != "":
		return []InspectRequest{{Name: "selector", Selector: opts.Selector, Props: props, Attributes: attrs, Source: "flag"}}, nil
	case opts.AllSections:
		requests := make([]InspectRequest, 0, len(cfg.Sections))
		for _, section := range cfg.Sections {
			selector := selectorForSection(section, side)
			if selector == "" {
				continue
			}
			requests = append(requests, InspectRequest{Name: section.Name, Selector: selector, Props: props, Attributes: attrs, Source: "section"})
		}
		if len(requests) == 0 {
			return nil, fmt.Errorf("no sections with selectors for side %s", side)
		}
		return requests, nil
	case opts.AllStyles:
		requests := make([]InspectRequest, 0, len(cfg.Styles))
		for _, style := range cfg.Styles {
			selector := selectorForStyleSide(style, side)
			if selector == "" {
				continue
			}
			styleProps := style.Props
			if len(opts.Props) > 0 {
				styleProps = opts.Props
			}
			if len(styleProps) == 0 {
				styleProps = defaultInspectProps()
			}
			styleAttrs := style.Attributes
			if len(opts.Attributes) > 0 {
				styleAttrs = opts.Attributes
			}
			if len(styleAttrs) == 0 {
				styleAttrs = defaultInspectAttributes()
			}
			requests = append(requests, InspectRequest{Name: style.Name, Selector: selector, Props: styleProps, Attributes: styleAttrs, Source: "style"})
		}
		if len(requests) == 0 {
			return nil, fmt.Errorf("no styles with selectors for side %s", side)
		}
		return requests, nil
	default:
		return nil, fmt.Errorf("provide an inspect selector")
	}
}

func writeInspectArtifacts(page *driver.Page, target config.Target, side string, req InspectRequest, format, outDir, outputFile string) (InspectArtifactResult, error) {
	format, err := canonicalInspectFormat(format)
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
		RootSelector:   rootSelectorForTarget(target),
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
		return writeSingleInspectArtifact(page, req, metadata, format, outputFile)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return artifact, err
	}

	metadataPath := filepath.Join(outDir, "metadata.json")
	if err := writeJSON(metadataPath, metadata); err != nil {
		return artifact, err
	}

	if format == InspectFormatBundle || format == InspectFormatHTML {
		htmlPath := filepath.Join(outDir, "prepared.html")
		if err := writePreparedHTML(page, req.Selector, htmlPath); err != nil {
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
		style, err := evaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if format == InspectFormatBundle || format == InspectFormatCSSJSON {
			if err := writeJSON(filepath.Join(outDir, "computed-css.json"), style); err != nil {
				return artifact, err
			}
		}
		if format == InspectFormatBundle || format == InspectFormatCSSMarkdown {
			if err := writeInspectCSSMarkdown(filepath.Join(outDir, "computed-css.md"), req, style); err != nil {
				return artifact, err
			}
		}
	}
	if format == InspectFormatBundle || format == InspectFormatInspectJSON {
		inspectPath := filepath.Join(outDir, "inspect.json")
		if err := writeInspectJSON(page, req.Selector, inspectPath); err != nil {
			return artifact, err
		}
		artifact.InspectJSON = inspectPath
	}
	return artifact, nil
}

func writeSingleInspectArtifact(page *driver.Page, req InspectRequest, metadata InspectMetadata, format, path string) (InspectArtifactResult, error) {
	artifact := InspectArtifactResult{Metadata: metadata}
	switch format {
	case InspectFormatPNG:
		if err := page.Screenshot(req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.Screenshot = path
	case InspectFormatHTML:
		if err := writePreparedHTML(page, req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.HTML = path
	case InspectFormatCSSJSON:
		style, err := evaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if err := writeJSON(path, style); err != nil {
			return artifact, err
		}
	case InspectFormatCSSMarkdown:
		style, err := evaluateStyle(page, config.StyleSpec{Selector: req.Selector, Props: req.Props, Attributes: req.Attributes, IncludeBounds: true})
		if err != nil {
			return artifact, err
		}
		artifact.Style = &style
		if err := writeInspectCSSMarkdown(path, req, style); err != nil {
			return artifact, err
		}
	case InspectFormatInspectJSON:
		if err := writeInspectJSON(page, req.Selector, path); err != nil {
			return artifact, err
		}
		artifact.InspectJSON = path
	case InspectFormatMetadataJSON:
		if err := writeJSON(path, metadata); err != nil {
			return artifact, err
		}
	default:
		return artifact, fmt.Errorf("format %q does not support --output-file", format)
	}
	return artifact, nil
}

func writeInspectCSSMarkdown(path string, req InspectRequest, style StyleSnapshot) error {
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

func writeInspectIndex(outDir string, result InspectResult) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "index.json"), result); err != nil {
		return err
	}
	content := "# css-visual-diff Inspect Index\n\n"
	content += "| Name | Source | Selector | Screenshot | HTML | CSS JSON | CSS Markdown | Inspect JSON |\n"
	content += "| --- | --- | --- | --- | --- | --- | --- | --- |\n"
	for _, r := range result.Results {
		name := r.Metadata.Name
		base := sanitizeName(name)
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

func inspectTargetForSide(cfg *config.Config, side string) (config.Target, string, error) {
	switch strings.TrimSpace(strings.ToLower(side)) {
	case "original", "orig":
		return cfg.Original, "original", nil
	case "react", "implementation", "impl":
		return cfg.React, "react", nil
	default:
		return config.Target{}, "", fmt.Errorf("--side must be original or react")
	}
}

func selectorForStyleSide(style config.StyleSpec, side string) string {
	if side == "original" {
		return selectorForTarget(style.Selector, style.SelectorOriginal)
	}
	return selectorForTarget(style.Selector, style.SelectorReact)
}

func findInspectSection(sections []config.SectionSpec, name string) (config.SectionSpec, bool) {
	for _, section := range sections {
		if section.Name == name {
			return section, true
		}
	}
	return config.SectionSpec{}, false
}

func findInspectStyle(styles []config.StyleSpec, name string) (config.StyleSpec, bool) {
	for _, style := range styles {
		if style.Name == name {
			return style, true
		}
	}
	return config.StyleSpec{}, false
}

func inspectSectionNames(sections []config.SectionSpec) string {
	names := make([]string, 0, len(sections))
	for _, section := range sections {
		names = append(names, section.Name)
	}
	return strings.Join(names, ", ")
}

func inspectStyleNames(styles []config.StyleSpec) string {
	names := make([]string, 0, len(styles))
	for _, style := range styles {
		names = append(names, style.Name)
	}
	return strings.Join(names, ", ")
}

func normalizeInspectFormat(format, outputFile string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		if outputFile != "" {
			return InspectFormatPNG
		}
		return InspectFormatBundle
	}
	return format
}

func canonicalInspectFormat(format string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(format)) {
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

func defaultInspectProps() []string {
	return []string{"display", "width", "height", "margin", "padding", "font-family", "font-size", "font-weight", "line-height", "color", "background-color", "border", "border-radius", "gap"}
}

func defaultInspectAttributes() []string {
	return []string{"id", "class"}
}
