package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Slug               string         `yaml:"slug"`
	Title              string         `yaml:"title"`
	Description        string         `yaml:"description"`
	Goal               string         `yaml:"goal"`
	ExpectedResult     ExpectedResult `yaml:"expected_result"`
	PotentialQuestions []string       `yaml:"potential_questions"`
	RelatedFiles       []RelatedFile  `yaml:"related_files"`
}

type ExpectedResult struct {
	Format      []string `yaml:"format"`
	Description string   `yaml:"description"`
}

type RelatedFile struct {
	Path   string `yaml:"path"`
	Reason string `yaml:"reason"`
}

type Target struct {
	Name         string       `yaml:"name"`
	URL          string       `yaml:"url"`
	WaitMS       int          `yaml:"wait_ms"`
	Viewport     Viewport     `yaml:"viewport"`
	RootSelector string       `yaml:"root_selector"`
	Prepare      *PrepareSpec `yaml:"prepare"`
}

type PrepareSpec struct {
	Type string `yaml:"type"`

	Script     string `yaml:"script"`
	ScriptFile string `yaml:"script_file"`

	WaitFor          string `yaml:"wait_for"`
	WaitForTimeoutMS int    `yaml:"wait_for_timeout_ms"`
	AfterWaitMS      int    `yaml:"after_wait_ms"`

	Component    string         `yaml:"component"`
	Props        map[string]any `yaml:"props"`
	RootSelector string         `yaml:"root_selector"`
	Width        int            `yaml:"width"`
	MinHeight    int            `yaml:"min_height"`
	Background   string         `yaml:"background"`
}

type Viewport struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type SectionSpec struct {
	Name               string            `yaml:"name"`
	Selector           string            `yaml:"selector"`
	SelectorOriginal   string            `yaml:"selector_original"`
	SelectorReact      string            `yaml:"selector_react"`
	OCRQuestion        string            `yaml:"ocr_question"`
	ExpectText         *TextExpectations `yaml:"expect_text"`
	ExpectTextOriginal *TextExpectations `yaml:"expect_text_original"`
	ExpectTextReact    *TextExpectations `yaml:"expect_text_react"`
	ExpectPNG          *PNGExpectations  `yaml:"expect_png"`
	ExpectPNGOriginal  *PNGExpectations  `yaml:"expect_png_original"`
	ExpectPNGReact     *PNGExpectations  `yaml:"expect_png_react"`
}

type TextExpectations struct {
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
}

type PNGExpectations struct {
	Width              int               `yaml:"width"`
	Height             int               `yaml:"height"`
	MinWidth           int               `yaml:"min_width"`
	MinHeight          int               `yaml:"min_height"`
	MaxWidth           int               `yaml:"max_width"`
	MaxHeight          int               `yaml:"max_height"`
	TopStripNear       *ColorExpectation `yaml:"top_strip_near"`
	TopStripNotNear    *ColorExpectation `yaml:"top_strip_not_near"`
	BottomStripNear    *ColorExpectation `yaml:"bottom_strip_near"`
	BottomStripNotNear *ColorExpectation `yaml:"bottom_strip_not_near"`
}

type ColorExpectation struct {
	RGB       []int `yaml:"rgb"`
	Tolerance int   `yaml:"tolerance"`
}

type StyleSpec struct {
	Name             string   `yaml:"name"`
	Selector         string   `yaml:"selector"`
	SelectorOriginal string   `yaml:"selector_original"`
	SelectorReact    string   `yaml:"selector_react"`
	Props            []string `yaml:"props"`
	IncludeBounds    bool     `yaml:"include_bounds"`
	Attributes       []string `yaml:"attributes"`
	Report           []string `yaml:"report"`
}

type OutputSpec struct {
	Dir               string `yaml:"dir"`
	WriteJSON         bool   `yaml:"write_json"`
	WriteMarkdown     bool   `yaml:"write_markdown"`
	WritePNGs         bool   `yaml:"write_pngs"`
	WritePreparedHTML bool   `yaml:"write_prepared_html"`
	WriteInspectJSON  bool   `yaml:"write_inspect_json"`
	ValidatePNGs      bool   `yaml:"validate_pngs"`
}

type Config struct {
	Metadata Metadata      `yaml:"metadata"`
	Original Target        `yaml:"original"`
	React    Target        `yaml:"react"`
	Sections []SectionSpec `yaml:"sections"`
	Styles   []StyleSpec   `yaml:"styles"`
	Output   OutputSpec    `yaml:"output"`
	Modes    []string      `yaml:"modes"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.normalizeOutput(path)
	return &cfg, nil
}

func (c *Config) Validate() error {
	var errs []string
	if strings.TrimSpace(c.Metadata.Slug) == "" {
		errs = append(errs, "metadata.slug is required")
	}
	if strings.TrimSpace(c.Original.URL) == "" {
		errs = append(errs, "original.url is required")
	} else if err := validateURL(c.Original.URL); err != nil {
		errs = append(errs, fmt.Sprintf("original.url invalid: %v", err))
	}
	if strings.TrimSpace(c.React.URL) == "" {
		errs = append(errs, "react.url is required")
	} else if err := validateURL(c.React.URL); err != nil {
		errs = append(errs, fmt.Sprintf("react.url invalid: %v", err))
	}
	if strings.TrimSpace(c.Output.Dir) == "" {
		errs = append(errs, "output.dir is required")
	}
	errs = append(errs, validatePrepare("original", c.Original.Prepare)...)
	errs = append(errs, validatePrepare("react", c.React.Prepare)...)
	for i, s := range c.Sections {
		if strings.TrimSpace(s.Name) == "" {
			errs = append(errs, fmt.Sprintf("sections[%d] must include name", i))
			continue
		}
		if strings.TrimSpace(s.Selector) == "" && strings.TrimSpace(s.SelectorOriginal) == "" && strings.TrimSpace(s.SelectorReact) == "" {
			errs = append(errs, fmt.Sprintf("sections[%d] must include selector or selector_original/selector_react", i))
			continue
		}
		if strings.TrimSpace(s.Selector) == "" {
			if strings.TrimSpace(s.SelectorOriginal) == "" || strings.TrimSpace(s.SelectorReact) == "" {
				errs = append(errs, fmt.Sprintf("sections[%d] must include both selector_original and selector_react when selector is omitted", i))
			}
		}
	}
	for i, s := range c.Styles {
		if strings.TrimSpace(s.Name) == "" {
			errs = append(errs, fmt.Sprintf("styles[%d] must include name", i))
			continue
		}
		if strings.TrimSpace(s.Selector) == "" && strings.TrimSpace(s.SelectorOriginal) == "" && strings.TrimSpace(s.SelectorReact) == "" {
			errs = append(errs, fmt.Sprintf("styles[%d] must include selector or selector_original/selector_react", i))
			continue
		}
		if strings.TrimSpace(s.Selector) == "" {
			if strings.TrimSpace(s.SelectorOriginal) == "" || strings.TrimSpace(s.SelectorReact) == "" {
				errs = append(errs, fmt.Sprintf("styles[%d] must include both selector_original and selector_react when selector is omitted", i))
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (c *Config) normalizeOutput(configPath string) {
	if c.Output.Dir == "" {
		return
	}
	if filepath.IsAbs(c.Output.Dir) {
		return
	}
	base := filepath.Dir(configPath)
	c.Output.Dir = filepath.Join(base, c.Output.Dir)
}

func validatePrepare(label string, prepare *PrepareSpec) []string {
	if prepare == nil {
		return nil
	}

	prepareType := strings.TrimSpace(prepare.Type)
	if prepareType == "" || prepareType == "none" {
		return nil
	}

	var errs []string
	switch prepareType {
	case "script":
		if strings.TrimSpace(prepare.Script) == "" && strings.TrimSpace(prepare.ScriptFile) == "" {
			errs = append(errs, fmt.Sprintf("%s.prepare.script requires script or script_file", label))
		}
	case "direct-react-global":
		if strings.TrimSpace(prepare.Component) == "" {
			errs = append(errs, fmt.Sprintf("%s.prepare.direct-react-global requires component", label))
		}
		if strings.TrimSpace(prepare.RootSelector) == "" {
			errs = append(errs, fmt.Sprintf("%s.prepare.direct-react-global requires root_selector", label))
		}
		if prepare.Width <= 0 {
			errs = append(errs, fmt.Sprintf("%s.prepare.direct-react-global requires positive width", label))
		}
	default:
		errs = append(errs, fmt.Sprintf("%s.prepare.type %q is not supported", label, prepare.Type))
	}
	return errs
}

func validateURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("missing scheme or host")
	}
	return nil
}
