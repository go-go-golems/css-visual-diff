package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
)

const CatalogSchemaVersion = "css-visual-diff.catalog.v1"

type CatalogOptions struct {
	Title        string
	OutDir       string
	ArtifactRoot string
	IndexName    string
}

type Catalog struct {
	opts     CatalogOptions
	manifest CatalogManifest
}

type CatalogManifest struct {
	SchemaVersion string                    `json:"schema_version"`
	Title         string                    `json:"title"`
	OutDir        string                    `json:"out_dir"`
	ArtifactRoot  string                    `json:"artifact_root"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	Targets       []CatalogTargetRecord     `json:"targets"`
	Preflights    []CatalogPreflightRecord  `json:"preflights,omitempty"`
	Results       []CatalogResultRecord     `json:"results,omitempty"`
	Comparisons   []CatalogComparisonRecord `json:"comparisons,omitempty"`
	Failures      []CatalogFailureRecord    `json:"failures,omitempty"`
	Summary       CatalogSummary            `json:"summary"`
}

type CatalogTargetRecord struct {
	Slug        string          `json:"slug"`
	Name        string          `json:"name,omitempty"`
	URL         string          `json:"url,omitempty"`
	Selector    string          `json:"selector,omitempty"`
	Viewport    config.Viewport `json:"viewport,omitempty"`
	Description string          `json:"description,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

type CatalogPreflightRecord struct {
	Target     CatalogTargetRecord `json:"target"`
	Statuses   []SelectorStatus    `json:"statuses"`
	RecordedAt time.Time           `json:"recorded_at"`
}

type CatalogResultRecord struct {
	Target     CatalogTargetRecord `json:"target"`
	Result     InspectResult       `json:"result"`
	RecordedAt time.Time           `json:"recorded_at"`
}

type CatalogComparisonRecord struct {
	Target     CatalogTargetRecord     `json:"target"`
	Comparison SelectionComparisonData `json:"comparison"`
	RecordedAt time.Time               `json:"recorded_at"`
}

type CatalogFailureRecord struct {
	Target     CatalogTargetRecord `json:"target"`
	Name       string              `json:"name,omitempty"`
	Code       string              `json:"code,omitempty"`
	Operation  string              `json:"operation,omitempty"`
	Message    string              `json:"message"`
	RecordedAt time.Time           `json:"recorded_at"`
}

type CatalogSummary struct {
	TargetCount     int `json:"target_count"`
	PreflightCount  int `json:"preflight_count"`
	ResultCount     int `json:"result_count"`
	ComparisonCount int `json:"comparison_count"`
	FailureCount    int `json:"failure_count"`
	ArtifactCount   int `json:"artifact_count"`
}

func NewCatalog(opts CatalogOptions) (*Catalog, error) {
	if strings.TrimSpace(opts.OutDir) == "" {
		return nil, fmt.Errorf("catalog outDir is required")
	}
	if strings.TrimSpace(opts.Title) == "" {
		opts.Title = "css-visual-diff Catalog"
	}
	if strings.TrimSpace(opts.ArtifactRoot) == "" {
		opts.ArtifactRoot = "artifacts"
	}
	if strings.TrimSpace(opts.IndexName) == "" {
		opts.IndexName = "index.md"
	}
	opts.ArtifactRoot = CleanCatalogRelativePath(opts.ArtifactRoot)
	if opts.ArtifactRoot == "" || opts.ArtifactRoot == "." {
		opts.ArtifactRoot = "artifacts"
	}
	opts.IndexName = CleanCatalogRelativePath(opts.IndexName)
	if opts.IndexName == "" || opts.IndexName == "." {
		opts.IndexName = "index.md"
	}
	now := time.Now().UTC()
	c := &Catalog{opts: opts}
	c.manifest = CatalogManifest{
		SchemaVersion: CatalogSchemaVersion,
		Title:         opts.Title,
		OutDir:        opts.OutDir,
		ArtifactRoot:  opts.ArtifactRoot,
		CreatedAt:     now,
		UpdatedAt:     now,
		Targets:       []CatalogTargetRecord{},
	}
	return c, nil
}

func (c *Catalog) Options() CatalogOptions {
	if c == nil {
		return CatalogOptions{}
	}
	return c.opts
}

func (c *Catalog) ArtifactDir(slug string) string {
	if c == nil {
		return ""
	}
	return filepath.Join(c.opts.OutDir, c.opts.ArtifactRoot, SanitizeCatalogSlug(slug))
}

func (c *Catalog) AddTarget(target CatalogTargetRecord) CatalogTargetRecord {
	if c == nil {
		return CatalogTargetRecord{}
	}
	target = NormalizeCatalogTarget(target)
	if !c.hasTarget(target.Slug) {
		c.manifest.Targets = append(c.manifest.Targets, target)
	}
	c.touch()
	return target
}

func (c *Catalog) RecordPreflight(target CatalogTargetRecord, statuses []SelectorStatus) CatalogPreflightRecord {
	record := CatalogPreflightRecord{Target: c.AddTarget(target), Statuses: statuses, RecordedAt: time.Now().UTC()}
	c.manifest.Preflights = append(c.manifest.Preflights, record)
	c.touch()
	return record
}

func (c *Catalog) AddResult(target CatalogTargetRecord, result InspectResult) CatalogResultRecord {
	record := CatalogResultRecord{Target: c.AddTarget(target), Result: result, RecordedAt: time.Now().UTC()}
	c.manifest.Results = append(c.manifest.Results, record)
	c.touch()
	return record
}

func (c *Catalog) AddComparison(target CatalogTargetRecord, comparison SelectionComparisonData) CatalogComparisonRecord {
	record := CatalogComparisonRecord{Target: c.AddTarget(target), Comparison: comparison, RecordedAt: time.Now().UTC()}
	c.manifest.Comparisons = append(c.manifest.Comparisons, record)
	c.touch()
	return record
}

func (c *Catalog) AddFailure(target CatalogTargetRecord, failure CatalogFailureRecord) CatalogFailureRecord {
	failure.Target = c.AddTarget(target)
	if strings.TrimSpace(failure.Message) == "" {
		failure.Message = "unknown failure"
	}
	failure.RecordedAt = time.Now().UTC()
	c.manifest.Failures = append(c.manifest.Failures, failure)
	c.touch()
	return failure
}

func (c *Catalog) Summary() CatalogSummary {
	if c == nil {
		return CatalogSummary{}
	}
	artifactCount := 0
	for _, result := range c.manifest.Results {
		for _, artifact := range result.Result.Results {
			if artifact.Screenshot != "" {
				artifactCount++
			}
			if artifact.HTML != "" {
				artifactCount++
			}
			if artifact.InspectJSON != "" {
				artifactCount++
			}
			if artifact.Style != nil {
				artifactCount++
			}
		}
	}
	for _, comparison := range c.manifest.Comparisons {
		artifactCount += len(comparison.Comparison.Artifacts)
	}
	return CatalogSummary{
		TargetCount:     len(c.manifest.Targets),
		PreflightCount:  len(c.manifest.Preflights),
		ResultCount:     len(c.manifest.Results),
		ComparisonCount: len(c.manifest.Comparisons),
		FailureCount:    len(c.manifest.Failures),
		ArtifactCount:   artifactCount,
	}
}

func (c *Catalog) Manifest() CatalogManifest {
	manifest := c.manifest
	manifest.Summary = c.Summary()
	return manifest
}

func (c *Catalog) WriteManifest() (string, error) {
	if c == nil {
		return "", fmt.Errorf("catalog is nil")
	}
	if err := os.MkdirAll(c.opts.OutDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(c.opts.OutDir, "manifest.json")
	if err := WriteJSON(path, c.Manifest()); err != nil {
		return "", err
	}
	return path, nil
}

func (c *Catalog) WriteIndex() (string, error) {
	if c == nil {
		return "", fmt.Errorf("catalog is nil")
	}
	if err := os.MkdirAll(c.opts.OutDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(c.opts.OutDir, c.opts.IndexName)
	manifest := c.Manifest()
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(manifest.Title)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Schema: `%s`\n\n", manifest.SchemaVersion)
	b.WriteString("## Summary\n\n")
	b.WriteString("| Targets | Preflights | Results | Comparisons | Failures | Artifacts |\n")
	b.WriteString("| ---: | ---: | ---: | ---: | ---: | ---: |\n")
	fmt.Fprintf(&b, "| %d | %d | %d | %d | %d | %d |\n\n", manifest.Summary.TargetCount, manifest.Summary.PreflightCount, manifest.Summary.ResultCount, manifest.Summary.ComparisonCount, manifest.Summary.FailureCount, manifest.Summary.ArtifactCount)
	if len(manifest.Targets) > 0 {
		b.WriteString("## Targets\n\n")
		b.WriteString("| Slug | Name | URL | Selector |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, target := range manifest.Targets {
			fmt.Fprintf(&b, "| %s | %s | %s | `%s` |\n", target.Slug, target.Name, target.URL, target.Selector)
		}
		b.WriteString("\n")
	}
	if len(manifest.Results) > 0 {
		b.WriteString("## Results\n\n")
		b.WriteString("| Target | Output Dir | Items |\n")
		b.WriteString("| --- | --- | ---: |\n")
		for _, result := range manifest.Results {
			fmt.Fprintf(&b, "| %s | `%s` | %d |\n", result.Target.Slug, result.Result.OutputDir, len(result.Result.Results))
		}
		b.WriteString("\n")
	}
	if len(manifest.Comparisons) > 0 {
		b.WriteString("## Comparisons\n\n")
		b.WriteString("| Target | Name | Changed % | Style changes | Attribute changes | Artifacts |\n")
		b.WriteString("| --- | --- | ---: | ---: | ---: | ---: |\n")
		for _, comparison := range manifest.Comparisons {
			changed := 0.0
			if comparison.Comparison.Pixel != nil {
				changed = comparison.Comparison.Pixel.ChangedPercent
			}
			fmt.Fprintf(&b, "| %s | %s | %.4f%% | %d | %d | %d |\n", comparison.Target.Slug, comparison.Comparison.Name, changed, len(comparison.Comparison.Styles), len(comparison.Comparison.Attributes), len(comparison.Comparison.Artifacts))
		}
		b.WriteString("\n")
	}
	if len(manifest.Failures) > 0 {
		b.WriteString("## Failures\n\n")
		b.WriteString("| Target | Code | Operation | Message |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, failure := range manifest.Failures {
			fmt.Fprintf(&b, "| %s | %s | `%s` | %s |\n", failure.Target.Slug, failure.Code, failure.Operation, failure.Message)
		}
		b.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func NormalizeCatalogTarget(target CatalogTargetRecord) CatalogTargetRecord {
	target.Slug = SanitizeCatalogSlug(target.Slug)
	if target.Slug == "" {
		target.Slug = SanitizeCatalogSlug(target.Name)
	}
	if target.Slug == "" {
		target.Slug = SanitizeCatalogSlug(target.URL)
	}
	if target.Slug == "" {
		target.Slug = "target"
	}
	return target
}

func SanitizeCatalogSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	var b strings.Builder
	lastSep := false
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastSep = false
		case r == '-' || r == '_':
			if !lastSep {
				b.WriteRune(r)
				lastSep = true
			}
		default:
			if !lastSep {
				b.WriteRune('-')
				lastSep = true
			}
		}
	}
	return strings.Trim(b.String(), "-_")
}

func CleanCatalogRelativePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return strings.Join(cleaned, string(filepath.Separator))
}

func (c *Catalog) hasTarget(slug string) bool {
	for _, target := range c.manifest.Targets {
		if target.Slug == slug {
			return true
		}
	}
	return false
}

func (c *Catalog) touch() {
	if c == nil {
		return
	}
	c.manifest.UpdatedAt = time.Now().UTC()
	c.manifest.Summary = c.Summary()
}

func CloneJSON[T any](value T) (T, error) {
	var out T
	b, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}
