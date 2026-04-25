package service

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
)

const SelectionComparisonSchemaVersion = "cssvd.selectionComparison.v1"

type CompareSelectionOptions struct {
	Name               string   `json:"name,omitempty"`
	Threshold          int      `json:"threshold,omitempty"`
	StyleProps         []string `json:"styleProps,omitempty"`
	ExcludeStyleProps  []string `json:"excludeStyleProps,omitempty"`
	Attributes         []string `json:"attributes,omitempty"`
	ExcludeAttributes  []string `json:"excludeAttributes,omitempty"`
	DiffOnlyPath       string   `json:"diffOnlyPath,omitempty"`
	DiffComparisonPath string   `json:"diffComparisonPath,omitempty"`
}

type SelectionSummary struct {
	Name     string  `json:"name,omitempty"`
	URL      string  `json:"url,omitempty"`
	Selector string  `json:"selector"`
	Source   string  `json:"source,omitempty"`
	Exists   bool    `json:"exists"`
	Visible  bool    `json:"visible"`
	Bounds   *Bounds `json:"bounds,omitempty"`
}

type BoundsDiff struct {
	Changed bool         `json:"changed"`
	Left    *Bounds      `json:"left,omitempty"`
	Right   *Bounds      `json:"right,omitempty"`
	Delta   *BoundsDelta `json:"delta,omitempty"`
}

type BoundsDelta struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type TextDiff struct {
	Changed bool   `json:"changed"`
	Left    string `json:"left,omitempty"`
	Right   string `json:"right,omitempty"`
}

type MapValueDiff struct {
	Name    string `json:"name"`
	Left    string `json:"left,omitempty"`
	Right   string `json:"right,omitempty"`
	Changed bool   `json:"changed"`
}

type SelectionArtifact struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

type SelectionComparisonData struct {
	SchemaVersion string              `json:"schemaVersion"`
	Name          string              `json:"name,omitempty"`
	Left          SelectionSummary    `json:"left"`
	Right         SelectionSummary    `json:"right"`
	Pixel         *PixelDiffResult    `json:"pixel,omitempty"`
	Bounds        BoundsDiff          `json:"bounds"`
	Text          TextDiff            `json:"text"`
	Styles        []MapValueDiff      `json:"styles,omitempty"`
	Attributes    []MapValueDiff      `json:"attributes,omitempty"`
	Artifacts     []SelectionArtifact `json:"artifacts,omitempty"`
}

func CompareSelections(left SelectionData, right SelectionData, opts CompareSelectionOptions) (SelectionComparisonData, error) {
	if err := ValidatePixelThreshold(opts.Threshold); err != nil {
		return SelectionComparisonData{}, err
	}
	out := SelectionComparisonData{
		SchemaVersion: SelectionComparisonSchemaVersion,
		Name:          firstNonEmpty(opts.Name, left.Name, right.Name),
		Left:          summarizeSelection(left),
		Right:         summarizeSelection(right),
		Bounds:        diffBounds(left.Bounds, right.Bounds),
		Text:          diffText(left.Text, right.Text),
		Styles:        diffStringMaps(left.ComputedStyles, right.ComputedStyles, opts.StyleProps, opts.ExcludeStyleProps),
		Attributes:    diffStringMaps(left.Attributes, right.Attributes, opts.Attributes, opts.ExcludeAttributes),
	}

	if left.Screenshot != nil && right.Screenshot != nil && left.Screenshot.Path != "" && right.Screenshot.Path != "" {
		pixel, err := compareSelectionScreenshots(left.Screenshot.Path, right.Screenshot.Path, opts)
		if err != nil {
			return SelectionComparisonData{}, err
		}
		out.Pixel = &pixel
		out.Artifacts = append(out.Artifacts, artifactsFromPixel(pixel)...)
	}
	return out, nil
}

func summarizeSelection(in SelectionData) SelectionSummary {
	return SelectionSummary{
		Name:     in.Name,
		URL:      in.URL,
		Selector: in.Selector,
		Source:   in.Source,
		Exists:   in.Exists,
		Visible:  in.Visible,
		Bounds:   in.Bounds,
	}
}

func diffBounds(left, right *Bounds) BoundsDiff {
	out := BoundsDiff{Left: left, Right: right}
	if left == nil || right == nil {
		out.Changed = left != right
		return out
	}
	delta := BoundsDelta{
		X:      right.X - left.X,
		Y:      right.Y - left.Y,
		Width:  right.Width - left.Width,
		Height: right.Height - left.Height,
	}
	out.Delta = &delta
	out.Changed = !floatEqual(delta.X, 0) || !floatEqual(delta.Y, 0) || !floatEqual(delta.Width, 0) || !floatEqual(delta.Height, 0)
	return out
}

func diffText(left, right string) TextDiff {
	return TextDiff{Changed: left != right, Left: left, Right: right}
}

func diffStringMaps(left, right map[string]string, include []string, exclude []string) []MapValueDiff {
	keys := selectedMapKeys(left, right, include, exclude)
	out := make([]MapValueDiff, 0, len(keys))
	for _, key := range keys {
		l := left[key]
		r := right[key]
		if l == r {
			continue
		}
		out = append(out, MapValueDiff{Name: key, Left: l, Right: r, Changed: true})
	}
	return out
}

func selectedMapKeys(left, right map[string]string, include []string, exclude []string) []string {
	excluded := stringSet(exclude)
	keys := map[string]bool{}
	if len(include) > 0 {
		for _, key := range include {
			if key != "" && !excluded[key] {
				keys[key] = true
			}
		}
	} else {
		for key := range left {
			if !excluded[key] {
				keys[key] = true
			}
		}
		for key := range right {
			if !excluded[key] {
				keys[key] = true
			}
		}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	return ordered
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		if value != "" {
			out[value] = true
		}
	}
	return out
}

func compareSelectionScreenshots(leftPath, rightPath string, opts CompareSelectionOptions) (PixelDiffResult, error) {
	if opts.DiffOnlyPath == "" && opts.DiffComparisonPath == "" {
		result, _, _, _, err := DiffPNGFiles(leftPath, rightPath, PixelDiffOptions{Threshold: opts.Threshold})
		return result, err
	}
	if opts.DiffOnlyPath == "" || opts.DiffComparisonPath == "" {
		return PixelDiffResult{}, fmt.Errorf("both diffOnlyPath and diffComparisonPath are required when writing selection comparison pixel artifacts")
	}
	return WritePixelDiffImages(leftPath, rightPath, opts.DiffComparisonPath, opts.DiffOnlyPath, PixelDiffOptions{Threshold: opts.Threshold})
}

func artifactsFromPixel(pixel PixelDiffResult) []SelectionArtifact {
	artifacts := []SelectionArtifact{}
	if pixel.DiffOnlyPath != "" {
		artifacts = append(artifacts, SelectionArtifact{Name: "diffOnly", Path: filepath.Clean(pixel.DiffOnlyPath), Kind: "png"})
	}
	if pixel.DiffComparisonPath != "" {
		artifacts = append(artifacts, SelectionArtifact{Name: "diffComparison", Path: filepath.Clean(pixel.DiffComparisonPath), Kind: "png"})
	}
	return artifacts
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}
