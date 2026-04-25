package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

const CollectedSelectionSchemaVersion = "cssvd.collectedSelection.v1"

const (
	CollectInspectMinimal = "minimal"
	CollectInspectRich    = "rich"
	CollectInspectDebug   = "debug"
)

type CollectionErrorKind string

const (
	CollectionErrorInvalidSelector CollectionErrorKind = "invalidSelector"
	CollectionErrorBrowser         CollectionErrorKind = "browser"
	CollectionErrorArtifact        CollectionErrorKind = "artifact"
)

type CollectionError struct {
	Kind     CollectionErrorKind
	Selector string
	Op       string
	Err      error
}

func (e *CollectionError) Error() string {
	if e == nil {
		return "collection error"
	}
	parts := []string{"collect selection"}
	if e.Op != "" {
		parts = append(parts, e.Op)
	}
	if e.Selector != "" {
		parts = append(parts, fmt.Sprintf("for %q", e.Selector))
	}
	if e.Kind != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Kind))
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	return strings.Join(parts, ": ")
}

func (e *CollectionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type CollectOptions struct {
	Name          string      `json:"name,omitempty"`
	Inspect       string      `json:"inspect,omitempty"`
	Text          TextOptions `json:"text,omitempty"`
	IncludeHTML   bool        `json:"includeHtml,omitempty"`
	OuterHTML     bool        `json:"outerHtml,omitempty"`
	StyleProps    []string    `json:"styleProps,omitempty"`
	AllStyles     bool        `json:"allStyles,omitempty"`
	Attributes    []string    `json:"attributes,omitempty"`
	AllAttributes bool        `json:"allAttributes,omitempty"`
}

type ScreenshotDescriptor struct {
	Path   string `json:"path,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type CollectedSelectionData struct {
	SchemaVersion  string                `json:"schemaVersion"`
	Name           string                `json:"name,omitempty"`
	URL            string                `json:"url,omitempty"`
	Selector       string                `json:"selector"`
	Source         string                `json:"source,omitempty"`
	Status         SelectorStatus        `json:"status"`
	Exists         bool                  `json:"exists"`
	Visible        bool                  `json:"visible"`
	Bounds         *Bounds               `json:"bounds,omitempty"`
	Text           string                `json:"text,omitempty"`
	HTML           string                `json:"html,omitempty"`
	ComputedStyles map[string]string     `json:"computedStyles,omitempty"`
	Attributes     map[string]string     `json:"attributes,omitempty"`
	Screenshot     *ScreenshotDescriptor `json:"screenshot,omitempty"`
}

type SelectionData = CollectedSelectionData

func CollectSelection(page *driver.Page, locator LocatorSpec, opts CollectOptions) (CollectedSelectionData, error) {
	if strings.TrimSpace(locator.Selector) == "" {
		return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorInvalidSelector, Selector: locator.Selector, Op: "validate selector", Err: fmt.Errorf("selector is empty")}
	}
	opts = normalizeCollectOptions(locator, opts)

	status, err := LocatorStatus(page, locator)
	if err != nil {
		return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "status", Err: err}
	}
	if status.Error != "" {
		return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorInvalidSelector, Selector: locator.Selector, Op: "status", Err: errors.New(status.Error)}
	}

	out := CollectedSelectionData{
		SchemaVersion: CollectedSelectionSchemaVersion,
		Name:          firstNonEmpty(opts.Name, locator.Name),
		URL:           pageURL(page),
		Selector:      locator.Selector,
		Source:        locator.Source,
		Status:        status,
		Exists:        status.Exists,
		Visible:       status.Visible,
		Bounds:        status.Bounds,
	}

	if !status.Exists {
		return out, nil
	}

	if shouldCollectText(opts) {
		text, err := LocatorText(page, locator, opts.Text)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "text", Err: err}
		}
		out.Text = text
	}

	if opts.IncludeHTML {
		html, err := LocatorHTML(page, locator, opts.OuterHTML)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "html", Err: err}
		}
		out.HTML = html.HTML
	}

	if opts.AllStyles {
		styles, err := collectAllComputedStyles(page, locator)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "all computed styles", Err: err}
		}
		out.ComputedStyles = styles
	} else if len(opts.StyleProps) > 0 {
		styles, err := LocatorComputedStyle(page, locator, opts.StyleProps)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "computed styles", Err: err}
		}
		out.ComputedStyles = styles
	}

	if opts.AllAttributes {
		attrs, err := collectAllAttributes(page, locator)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "all attributes", Err: err}
		}
		out.Attributes = attrs
	} else if len(opts.Attributes) > 0 {
		attrs, err := LocatorAttributes(page, locator, opts.Attributes)
		if err != nil {
			return CollectedSelectionData{}, &CollectionError{Kind: CollectionErrorBrowser, Selector: locator.Selector, Op: "attributes", Err: err}
		}
		out.Attributes = attrs
	}

	return out, nil
}

func normalizeCollectOptions(locator LocatorSpec, opts CollectOptions) CollectOptions {
	if opts.Name == "" {
		opts.Name = locator.Name
	}
	inspect := strings.ToLower(strings.TrimSpace(opts.Inspect))
	if inspect == "" {
		inspect = CollectInspectRich
	}
	opts.Inspect = inspect
	switch inspect {
	case CollectInspectMinimal:
		// Status and bounds from preflight only.
	case CollectInspectDebug:
		if !opts.Text.NormalizeWhitespace && !opts.Text.Trim {
			opts.Text = TextOptions{NormalizeWhitespace: true, Trim: true}
		}
		opts.IncludeHTML = true
		opts.AllStyles = true
		opts.AllAttributes = true
	case CollectInspectRich:
		if !opts.Text.NormalizeWhitespace && !opts.Text.Trim {
			opts.Text = TextOptions{NormalizeWhitespace: true, Trim: true}
		}
		if len(opts.StyleProps) == 0 && !opts.AllStyles {
			opts.StyleProps = []string{"display", "position", "color", "background-color", "font-family", "font-size", "font-weight", "line-height", "margin", "padding", "border"}
		}
		if len(opts.Attributes) == 0 && !opts.AllAttributes {
			opts.Attributes = []string{"id", "class", "role", "aria-label", "data-testid"}
		}
	}
	return opts
}

func shouldCollectText(opts CollectOptions) bool {
	return opts.Inspect == CollectInspectRich || opts.Inspect == CollectInspectDebug || opts.Text.NormalizeWhitespace || opts.Text.Trim
}

func pageURL(page *driver.Page) string {
	var url string
	if err := page.Evaluate(`window.location.href`, &url); err != nil {
		return ""
	}
	return url
}

func collectAllAttributes(page *driver.Page, locator LocatorSpec) (map[string]string, error) {
	selectorJSON, err := json.Marshal(locator.Selector)
	if err != nil {
		return nil, fmt.Errorf("marshal selector: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  if (!selector) throw new Error("selector is empty");
	  let el;
	  try { el = document.querySelector(selector); } catch (err) { throw new Error(String(err && err.message ? err.message : err)); }
	  const attributes = {};
	  if (!el) return attributes;
	  for (const attr of Array.from(el.attributes || [])) attributes[attr.name] = attr.value == null ? "" : String(attr.value);
	  return attributes;
	})()`, string(selectorJSON))
	out := map[string]string{}
	if err := page.Evaluate(script, &out); err != nil {
		return nil, fmt.Errorf("evaluate all attributes for %q: %w", locator.Selector, err)
	}
	return sortedMap(out), nil
}

func collectAllComputedStyles(page *driver.Page, locator LocatorSpec) (map[string]string, error) {
	selectorJSON, err := json.Marshal(locator.Selector)
	if err != nil {
		return nil, fmt.Errorf("marshal selector: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  if (!selector) throw new Error("selector is empty");
	  let el;
	  try { el = document.querySelector(selector); } catch (err) { throw new Error(String(err && err.message ? err.message : err)); }
	  const computed = {};
	  if (!el) return computed;
	  const style = window.getComputedStyle(el);
	  for (let i = 0; i < style.length; i++) {
	    const name = style[i];
	    computed[name] = style.getPropertyValue(name) || "";
	  }
	  return computed;
	})()`, string(selectorJSON))
	out := map[string]string{}
	if err := page.Evaluate(script, &out); err != nil {
		return nil, fmt.Errorf("evaluate all computed styles for %q: %w", locator.Selector, err)
	}
	return sortedMap(out), nil
}

func sortedMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make(map[string]string, len(in))
	for _, key := range keys {
		out[key] = in[key]
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
