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
	Name     string   `yaml:"name"`
	URL      string   `yaml:"url"`
	WaitMS   int      `yaml:"wait_ms"`
	Viewport Viewport `yaml:"viewport"`
}

type Viewport struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type SectionSpec struct {
	Name             string `yaml:"name"`
	Selector         string `yaml:"selector"`
	SelectorOriginal string `yaml:"selector_original"`
	SelectorReact    string `yaml:"selector_react"`
	OCRQuestion      string `yaml:"ocr_question"`
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
	Dir           string `yaml:"dir"`
	WriteJSON     bool   `yaml:"write_json"`
	WriteMarkdown bool   `yaml:"write_markdown"`
	WritePNGs     bool   `yaml:"write_pngs"`
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
