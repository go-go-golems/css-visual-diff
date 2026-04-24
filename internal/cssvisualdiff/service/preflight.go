package service

import (
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

func PreflightProbes(page *driver.Page, probes []ProbeSpec) ([]SelectorStatus, error) {
	probeJSON, err := json.Marshal(probes)
	if err != nil {
		return nil, fmt.Errorf("marshal preflight probes: %w", err)
	}
	script := fmt.Sprintf(`(() => {
	  const probes = %s;
	  return probes.map((probe) => {
	    const status = {
	      name: probe.name || "",
	      selector: probe.selector || "",
	      source: probe.source || "",
	      exists: false,
	      visible: false,
	      bounds: null,
	      text_start: "",
	      error: ""
	    };
	    if (!status.selector) {
	      status.error = "selector is empty";
	      return status;
	    }
	    let el;
	    try {
	      el = document.querySelector(status.selector);
	    } catch (err) {
	      status.error = String(err && err.message ? err.message : err);
	      return status;
	    }
	    if (!el) return status;
	    const rect = el.getBoundingClientRect();
	    const style = window.getComputedStyle(el);
	    status.exists = true;
	    status.bounds = { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	    status.visible = !!(rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none");
	    const text = (el.textContent || "").replace(/\s+/g, " ").trim();
	    status.text_start = text.slice(0, 160);
	    return status;
	  });
	})()`, string(probeJSON))

	var statuses []SelectorStatus
	if err := page.Evaluate(script, &statuses); err != nil {
		return nil, fmt.Errorf("evaluate selector preflight: %w", err)
	}
	return statuses, nil
}
