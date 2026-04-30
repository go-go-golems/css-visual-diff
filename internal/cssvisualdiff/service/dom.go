package service

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

// LocatorSpec identifies one element on a loaded page.
type LocatorSpec struct {
	Name     string `json:"name,omitempty"`
	Selector string `json:"selector"`
	Source   string `json:"source,omitempty"`
}

type TextOptions struct {
	NormalizeWhitespace bool `json:"normalizeWhitespace,omitempty"`
	Trim                bool `json:"trim,omitempty"`
}

type ElementHTML struct {
	Exists bool   `json:"exists"`
	HTML   string `json:"html"`
}

func LocatorStatus(page *driver.Page, locator LocatorSpec) (SelectorStatus, error) {
	statuses, err := PreflightProbes(page, []ProbeSpec{{Name: locator.Name, Selector: locator.Selector, Source: locator.Source}})
	if err != nil {
		return SelectorStatus{}, err
	}
	if len(statuses) == 0 {
		return SelectorStatus{Name: locator.Name, Selector: locator.Selector, Source: locator.Source, Error: "locator status produced no result"}, nil
	}
	return statuses[0], nil
}

func LocatorText(page *driver.Page, locator LocatorSpec, opts TextOptions) (string, error) {
	payload, err := json.Marshal(struct {
		Locator LocatorSpec `json:"locator"`
		Options TextOptions `json:"options"`
	}{Locator: locator, Options: opts})
	if err != nil {
		return "", fmt.Errorf("marshal locator text request: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const request = %s;
	  const selector = request.locator.selector || "";
	  if (!selector) throw new Error("selector is empty");
	  let el;
	  try {
	    el = document.querySelector(selector);
	  } catch (err) {
	    throw new Error(String(err && err.message ? err.message : err));
	  }
	  if (!el) return "";
	  let text = el.textContent || "";
	  if (request.options.normalizeWhitespace) text = text.replace(/\s+/g, " ");
	  if (request.options.trim) text = text.trim();
	  return text;
	})()`, string(payload))
	var text string
	if err := page.Evaluate(script, &text); err != nil {
		return "", fmt.Errorf("evaluate locator text for %q: %w", locator.Selector, err)
	}
	return text, nil
}

func LocatorHTML(page *driver.Page, locator LocatorSpec, outer bool) (ElementHTML, error) {
	payload, err := json.Marshal(struct {
		Locator LocatorSpec `json:"locator"`
		Outer   bool        `json:"outer"`
	}{Locator: locator, Outer: outer})
	if err != nil {
		return ElementHTML{}, fmt.Errorf("marshal locator html request: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const request = %s;
	  const selector = request.locator.selector || "";
	  if (!selector) throw new Error("selector is empty");
	  let el;
	  try {
	    el = document.querySelector(selector);
	  } catch (err) {
	    throw new Error(String(err && err.message ? err.message : err));
	  }
	  if (!el) return { exists: false, html: "" };
	  return { exists: true, html: request.outer ? el.outerHTML : el.innerHTML };
	})()`, string(payload))
	var out ElementHTML
	if err := page.Evaluate(script, &out); err != nil {
		return ElementHTML{}, fmt.Errorf("evaluate locator html for %q: %w", locator.Selector, err)
	}
	return out, nil
}

func LocatorBounds(page *driver.Page, locator LocatorSpec) (*Bounds, error) {
	selectorJSON, err := json.Marshal(locator.Selector)
	if err != nil {
		return nil, fmt.Errorf("marshal locator selector: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const selector = %s;
	  if (!selector) throw new Error("selector is empty");
	  let el;
	  try {
	    el = document.querySelector(selector);
	  } catch (err) {
	    throw new Error(String(err && err.message ? err.message : err));
	  }
	  if (!el) return null;
	  const rect = el.getBoundingClientRect();
	  return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	})()`, string(selectorJSON))
	var bounds *Bounds
	if err := page.Evaluate(script, &bounds); err != nil {
		return nil, fmt.Errorf("evaluate locator bounds for %q: %w", locator.Selector, err)
	}
	return bounds, nil
}

func LocatorAttributes(page *driver.Page, locator LocatorSpec, attrs []string) (map[string]string, error) {
	if attrs == nil {
		attrs = []string{}
	}
	payload, err := json.Marshal(struct {
		Selector string   `json:"selector"`
		Attrs    []string `json:"attrs"`
	}{Selector: locator.Selector, Attrs: attrs})
	if err != nil {
		return nil, fmt.Errorf("marshal locator attributes request: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const request = %s;
	  if (!request.selector) throw new Error("selector is empty");
	  let el;
	  try {
	    el = document.querySelector(request.selector);
	  } catch (err) {
	    throw new Error(String(err && err.message ? err.message : err));
	  }
	  const attributes = {};
	  if (!el) return attributes;
	  request.attrs.forEach((name) => {
	    const value = el.getAttribute(name);
	    attributes[name] = value == null ? "" : String(value);
	  });
	  return attributes;
	})()`, string(payload))
	out := map[string]string{}
	if err := page.Evaluate(script, &out); err != nil {
		return nil, fmt.Errorf("evaluate locator attributes for %q: %w", locator.Selector, err)
	}
	return out, nil
}

func LocatorComputedStyle(page *driver.Page, locator LocatorSpec, props []string) (map[string]string, error) {
	if props == nil {
		props = []string{}
	}
	snapshot, err := EvaluateStyle(page, configStyleSpec(locator.Selector, props, nil, false))
	if err != nil {
		return nil, fmt.Errorf("evaluate locator computed style for %q: %w", locator.Selector, err)
	}
	if !snapshot.Exists {
		return map[string]string{}, nil
	}
	return snapshot.Computed, nil
}

func configStyleSpec(selector string, props []string, attrs []string, includeBounds bool) StyleEvalSpec {
	return StyleEvalSpec{Selector: selector, Props: props, Attributes: attrs, IncludeBounds: includeBounds}
}
