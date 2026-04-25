package modes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
	return service.CanonicalInspectFormat(strings.TrimSpace(strings.ToLower(format)))
}

func inspectFormatRequiresExistingSelector(format string) bool {
	return service.InspectFormatRequiresExistingSelector(format)
}

func defaultInspectProps() []string {
	return []string{"display", "width", "height", "margin", "padding", "font-family", "font-size", "font-weight", "line-height", "color", "background-color", "border", "border-radius", "gap"}
}

func defaultInspectAttributes() []string {
	return []string{"id", "class"}
}
