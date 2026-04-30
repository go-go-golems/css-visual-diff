package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
)

const DefaultPrepareWaitTimeout = 30 * time.Second
const DefaultPrepareAfterWait = 500 * time.Millisecond

func PrepareTarget(page *driver.Page, target PageTarget) error {
	prepare := target.Prepare
	if prepare == nil {
		return nil
	}

	prepareType := strings.TrimSpace(prepare.Type)
	if prepareType == "" || prepareType == "none" {
		return nil
	}

	if strings.TrimSpace(prepare.WaitFor) != "" {
		timeout := DefaultPrepareWaitTimeout
		if prepare.WaitForTimeoutMS > 0 {
			timeout = time.Duration(prepare.WaitForTimeoutMS) * time.Millisecond
		}
		if err := page.WaitForFunction(prepare.WaitFor, timeout); err != nil {
			return fmt.Errorf("prepare wait_for failed for target %q: %w", target.Name, err)
		}
	}

	var err error
	switch prepareType {
	case "script":
		err = RunScriptPrepare(page, prepare)
	case "direct-react-global":
		err = RunDirectReactGlobalPrepare(page, prepare)
	default:
		err = fmt.Errorf("unknown prepare type %q", prepare.Type)
	}
	if err != nil {
		return fmt.Errorf("prepare target %q: %w", target.Name, err)
	}

	afterWait := DefaultPrepareAfterWait
	if prepare.AfterWaitMS > 0 {
		afterWait = time.Duration(prepare.AfterWaitMS) * time.Millisecond
	}
	if afterWait > 0 {
		page.Wait(afterWait)
	}

	return nil
}

func RunScriptPrepare(page *driver.Page, prepare *PrepareSpec) error {
	script := prepare.Script
	if strings.TrimSpace(script) == "" && strings.TrimSpace(prepare.ScriptFile) != "" {
		data, err := os.ReadFile(prepare.ScriptFile)
		if err != nil {
			return fmt.Errorf("read prepare script_file %q: %w", prepare.ScriptFile, err)
		}
		script = string(data)
	}
	if strings.TrimSpace(script) == "" {
		return fmt.Errorf("script prepare requires script or script_file")
	}
	return page.Eval(script)
}

func RunDirectReactGlobalPrepare(page *driver.Page, prepare *PrepareSpec) error {
	script, err := BuildDirectReactGlobalScript(prepare)
	if err != nil {
		return err
	}
	var out DirectReactGlobalPrepareResult
	if err := page.Evaluate(script, &out); err != nil {
		return err
	}
	if !out.OK {
		return fmt.Errorf("direct-react-global prepare did not report ok")
	}
	return nil
}

type DirectReactGlobalPrepareResult struct {
	OK            bool               `json:"ok"`
	ComponentName string             `json:"component_name"`
	RootSelector  string             `json:"root_selector"`
	Props         map[string]any     `json:"props"`
	Bounds        map[string]float64 `json:"bounds"`
}

func BuildDirectReactGlobalScript(prepare *PrepareSpec) (string, error) {
	component := strings.TrimSpace(prepare.Component)
	if component == "" {
		return "", fmt.Errorf("direct-react-global prepare requires component")
	}
	rootSelector := strings.TrimSpace(prepare.RootSelector)
	if rootSelector == "" {
		return "", fmt.Errorf("direct-react-global prepare requires root_selector")
	}
	if prepare.Width <= 0 {
		return "", fmt.Errorf("direct-react-global prepare requires positive width")
	}

	props := prepare.Props
	if props == nil {
		props = map[string]any{}
	}
	background := strings.TrimSpace(prepare.Background)
	if background == "" {
		background = "#fff"
	}

	componentJSON, err := json.Marshal(component)
	if err != nil {
		return "", err
	}
	rootSelectorJSON, err := json.Marshal(rootSelector)
	if err != nil {
		return "", err
	}
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return "", err
	}
	backgroundJSON, err := json.Marshal(background)
	if err != nil {
		return "", err
	}

	script := fmt.Sprintf(`(() => {
  const componentName = %s;
  const props = %s;
  const rootSelector = %s;
  const width = %d;
  const minHeight = %d;
  const background = %s;

  if (!window.React) throw new Error('Missing React global');
  if (!window.ReactDOM) throw new Error('Missing ReactDOM global');

  const rootId = rootSelector.startsWith('#') ? rootSelector.slice(1) : 'capture-root';
  document.documentElement.style.margin = '0';
  document.documentElement.style.padding = '0';
  document.body.innerHTML = '<div id="' + rootId + '"></div>';
  document.body.style.margin = '0';
  document.body.style.padding = '0';
  document.body.style.background = background;
  document.body.style.overflow = 'visible';

  const root = document.getElementById(rootId);
  root.style.width = width + 'px';
  root.style.background = background;
  root.style.overflow = 'visible';
  if (minHeight > 0) root.style.minHeight = minHeight + 'px';

  const Component = window[componentName];
  if (!Component) throw new Error('Missing component global: ' + componentName);

  if (typeof window.ReactDOM.createRoot === 'function') {
    window.ReactDOM.createRoot(root).render(window.React.createElement(Component, props));
  } else if (typeof window.ReactDOM.render === 'function') {
    window.ReactDOM.render(window.React.createElement(Component, props), root);
  } else {
    throw new Error('ReactDOM has neither createRoot nor render');
  }

  const rect = root.getBoundingClientRect();
  return {
    ok: true,
    component_name: componentName,
    root_selector: '#' + rootId,
    props,
    bounds: { x: rect.x, y: rect.y, width: rect.width, height: rect.height }
  };
})()`, string(componentJSON), string(propsJSON), string(rootSelectorJSON), prepare.Width, prepare.MinHeight, string(backgroundJSON))

	return script, nil
}
