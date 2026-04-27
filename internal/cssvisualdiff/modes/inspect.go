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
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

const (
	InspectFormatBundle       = service.InspectFormatBundle
	InspectFormatPNG          = service.InspectFormatPNG
	InspectFormatHTML         = service.InspectFormatHTML
	InspectFormatCSSJSON      = service.InspectFormatCSSJSON
	InspectFormatCSSMarkdown  = service.InspectFormatCSSMarkdown
	InspectFormatInspectJSON  = service.InspectFormatInspectJSON
	InspectFormatMetadataJSON = service.InspectFormatMetadataJSON
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

type InspectRequest = service.InspectRequest

type InspectMetadata = service.InspectMetadata

type InspectArtifactResult = service.InspectArtifactResult

type InspectResult = service.InspectResult

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

	if err := service.LoadAndPreparePage(page, target); err != nil {
		return InspectResult{}, err
	}

	return service.InspectPreparedPage(page, target, side, requests, service.InspectAllOptions{
		OutDir:     outDir,
		Format:     format,
		OutputFile: opts.OutputFile,
	})
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

//nolint:unused // Kept for the legacy inspect artifact path while the CLI uses service-backed inspection.
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

	if inspectFormatRequiresExistingSelector(format) {
		if err := ensureInspectSelectorExists(page, req); err != nil {
			return artifact, err
		}
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

func inspectFormatRequiresExistingSelector(format string) bool {
	switch format {
	case InspectFormatBundle, InspectFormatPNG, InspectFormatHTML, InspectFormatInspectJSON:
		return true
	default:
		return false
	}
}

//nolint:unused // Used by the legacy inspect artifact helpers retained for compatibility.
func ensureInspectSelectorExists(page *driver.Page, req InspectRequest) error {
	statuses, err := service.PreflightProbes(page, []service.ProbeSpec{{
		Name:     req.Name,
		Selector: req.Selector,
		Source:   req.Source,
	}})
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

//nolint:unused // Used by the legacy inspect artifact helpers retained for compatibility.
func writeSingleInspectArtifact(page *driver.Page, req InspectRequest, metadata InspectMetadata, format, path string) (InspectArtifactResult, error) {
	artifact := InspectArtifactResult{Metadata: metadata}
	if inspectFormatRequiresExistingSelector(format) {
		if err := ensureInspectSelectorExists(page, req); err != nil {
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

//nolint:unused // Used by the legacy inspect artifact helpers retained for compatibility.
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

//nolint:unused // Used by the legacy inspect artifact helpers retained for compatibility.
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
