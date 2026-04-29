package modes

import (
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/config"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/service"
)

// toServiceViewport is a temporary adapter for legacy config-driven modes.
// The long-term JS-first runtime uses service.Viewport directly.
func toServiceViewport(v config.Viewport) service.Viewport {
	return service.Viewport{Width: v.Width, Height: v.Height}
}

// toServicePrepareSpec is a temporary adapter for legacy config-driven modes.
// It should disappear when the native YAML run pipeline is removed.
func toServicePrepareSpec(prepare *config.PrepareSpec) *service.PrepareSpec {
	if prepare == nil {
		return nil
	}
	return &service.PrepareSpec{
		Type:             prepare.Type,
		Script:           prepare.Script,
		ScriptFile:       prepare.ScriptFile,
		WaitFor:          prepare.WaitFor,
		WaitForTimeoutMS: prepare.WaitForTimeoutMS,
		AfterWaitMS:      prepare.AfterWaitMS,
		Component:        prepare.Component,
		Props:            prepare.Props,
		RootSelector:     prepare.RootSelector,
		Width:            prepare.Width,
		MinHeight:        prepare.MinHeight,
		Background:       prepare.Background,
	}
}

// toServicePageTarget is a temporary adapter for legacy config-driven modes.
// New code should construct service.PageTarget directly.
func toServicePageTarget(target config.Target) service.PageTarget {
	return service.PageTarget{
		Name:         target.Name,
		URL:          target.URL,
		WaitMS:       target.WaitMS,
		Viewport:     toServiceViewport(target.Viewport),
		RootSelector: target.RootSelector,
		Prepare:      toServicePrepareSpec(target.Prepare),
	}
}

// toServiceStyleEvalSpec is a temporary adapter for legacy config-driven modes.
// New code should construct service.StyleEvalSpec directly.
func toServiceStyleEvalSpec(spec config.StyleSpec) service.StyleEvalSpec {
	return service.StyleEvalSpec{
		Selector:      spec.Selector,
		Props:         spec.Props,
		Attributes:    spec.Attributes,
		IncludeBounds: spec.IncludeBounds,
		Report:        spec.Report,
	}
}
