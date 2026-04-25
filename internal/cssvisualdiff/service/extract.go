package service

import (
	"fmt"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

type ExtractorKind string

const (
	ExtractorExists        ExtractorKind = "exists"
	ExtractorVisible       ExtractorKind = "visible"
	ExtractorText          ExtractorKind = "text"
	ExtractorBounds        ExtractorKind = "bounds"
	ExtractorComputedStyle ExtractorKind = "computedStyle"
	ExtractorAttributes    ExtractorKind = "attributes"
)

type ExtractorSpec struct {
	Kind       ExtractorKind `json:"kind"`
	Props      []string      `json:"props,omitempty"`
	Attributes []string      `json:"attributes,omitempty"`
	Text       TextOptions   `json:"text,omitempty"`
}

type ElementSnapshot struct {
	Selector   string            `json:"selector"`
	Exists     *bool             `json:"exists,omitempty"`
	Visible    *bool             `json:"visible,omitempty"`
	Text       string            `json:"text,omitempty"`
	Bounds     *Bounds           `json:"bounds,omitempty"`
	Computed   map[string]string `json:"computed,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func ExtractElement(page *driver.Page, locator LocatorSpec, extractors []ExtractorSpec) (ElementSnapshot, error) {
	snapshot := ElementSnapshot{Selector: locator.Selector}
	for _, extractor := range extractors {
		switch extractor.Kind {
		case ExtractorExists:
			status, err := LocatorStatus(page, locator)
			if err != nil {
				return ElementSnapshot{}, err
			}
			if status.Error != "" {
				return ElementSnapshot{}, fmt.Errorf("selector status for %q: %s", locator.Selector, status.Error)
			}
			exists := status.Exists
			snapshot.Exists = &exists
		case ExtractorVisible:
			status, err := LocatorStatus(page, locator)
			if err != nil {
				return ElementSnapshot{}, err
			}
			if status.Error != "" {
				return ElementSnapshot{}, fmt.Errorf("selector status for %q: %s", locator.Selector, status.Error)
			}
			visible := status.Visible
			snapshot.Visible = &visible
		case ExtractorText:
			text, err := LocatorText(page, locator, extractor.Text)
			if err != nil {
				return ElementSnapshot{}, err
			}
			snapshot.Text = text
		case ExtractorBounds:
			bounds, err := LocatorBounds(page, locator)
			if err != nil {
				return ElementSnapshot{}, err
			}
			snapshot.Bounds = bounds
		case ExtractorComputedStyle:
			computed, err := LocatorComputedStyle(page, locator, extractor.Props)
			if err != nil {
				return ElementSnapshot{}, err
			}
			snapshot.Computed = computed
		case ExtractorAttributes:
			attributes, err := LocatorAttributes(page, locator, extractor.Attributes)
			if err != nil {
				return ElementSnapshot{}, err
			}
			snapshot.Attributes = attributes
		default:
			return ElementSnapshot{}, fmt.Errorf("unsupported extractor kind %q", extractor.Kind)
		}
	}
	return snapshot, nil
}
