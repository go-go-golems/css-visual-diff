package service

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

type styleEvalResult struct {
	Exists     bool              `json:"exists"`
	Computed   map[string]string `json:"computed"`
	Bounds     *Bounds           `json:"bounds"`
	Attributes map[string]string `json:"attributes"`
}

func EvaluateStyle(page *driver.Page, spec config.StyleSpec) (StyleSnapshot, error) {
	props := spec.Props
	if props == nil {
		props = []string{}
	}
	attrs := spec.Attributes
	if attrs == nil {
		attrs = []string{}
	}
	propsJSON, _ := json.Marshal(props)
	attrsJSON, _ := json.Marshal(attrs)
	script := fmt.Sprintf(`(() => {
	  const props = %s;
	  const attrs = %s;
	  const el = document.querySelector(%q);
	  if (!el) return { exists: false, computed: {}, bounds: null, attributes: {} };
	  const style = window.getComputedStyle(el);
	  const computed = {};
	  props.forEach((p) => {
	    computed[p] = style.getPropertyValue(p) || style[p] || "";
	  });
	  let bounds = null;
	  if (%t) {
	    const rect = el.getBoundingClientRect();
	    bounds = { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	  }
	  const attributes = {};
	  attrs.forEach((a) => {
	    attributes[a] = el.getAttribute(a);
	  });
	  return { exists: true, computed, bounds, attributes };
	})()`, string(propsJSON), string(attrsJSON), spec.Selector, spec.IncludeBounds)

	out := styleEvalResult{}
	if err := page.Evaluate(script, &out); err != nil {
		return StyleSnapshot{}, err
	}

	return StyleSnapshot(out), nil
}
