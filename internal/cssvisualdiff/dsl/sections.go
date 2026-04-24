package dsl

import "github.com/go-go-golems/go-go-goja/pkg/jsverbs"

func registerSharedSections(registry *jsverbs.Registry) error {
	return registry.AddSharedSections(targetsSection(), viewportSection(), outputSection())
}

func targetsSection() *jsverbs.SectionSpec {
	return &jsverbs.SectionSpec{
		Slug:        "targets",
		Title:       "Targets",
		Description: "Left/right target URLs and wait settings for script-backed css-visual-diff verbs.",
		Fields: map[string]*jsverbs.FieldSpec{
			"leftUrl": {
				Type:     "string",
				Required: true,
				Help:     "Left target URL",
			},
			"rightUrl": {
				Type:     "string",
				Required: true,
				Help:     "Right target URL",
			},
			"leftWaitMs": {
				Type:    "int",
				Default: 0,
				Help:    "Wait after left navigation in milliseconds",
			},
			"rightWaitMs": {
				Type:    "int",
				Default: 0,
				Help:    "Wait after right navigation in milliseconds",
			},
		},
	}
}

func viewportSection() *jsverbs.SectionSpec {
	return &jsverbs.SectionSpec{
		Slug:        "viewport",
		Title:       "Viewport",
		Description: "Shared viewport settings for script-backed css-visual-diff verbs.",
		Fields: map[string]*jsverbs.FieldSpec{
			"width": {
				Type:    "int",
				Default: 1280,
				Help:    "Viewport width",
			},
			"height": {
				Type:    "int",
				Default: 720,
				Help:    "Viewport height",
			},
		},
	}
}

func outputSection() *jsverbs.SectionSpec {
	return &jsverbs.SectionSpec{
		Slug:        "output",
		Title:       "Output",
		Description: "Artifact and reporting settings for script-backed css-visual-diff verbs.",
		Fields: map[string]*jsverbs.FieldSpec{
			"outDir": {
				Type: "string",
				Help: "Output directory for artifacts (default: generated timestamped directory)",
			},
			"threshold": {
				Type:    "int",
				Default: 30,
				Help:    "Pixel diff threshold (0-255)",
			},
			"writeJson": {
				Type:    "bool",
				Default: false,
				Help:    "Write compare.json when supported by the host verb",
			},
			"writeMarkdown": {
				Type:    "bool",
				Default: false,
				Help:    "Write compare.md when supported by the host verb",
			},
			"writePngs": {
				Type:    "bool",
				Default: true,
				Help:    "Write screenshots and diff PNGs",
			},
		},
	}
}
