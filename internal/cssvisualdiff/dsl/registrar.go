package dsl

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/modes"
	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/services"
	"github.com/go-go-golems/go-go-goja/engine"
)

type runtimeRegistrar struct{}

func newRuntimeRegistrar() engine.RuntimeModuleRegistrar {
	return runtimeRegistrar{}
}

func (runtimeRegistrar) ID() string {
	return "css-visual-diff-runtime-modules"
}

func (runtimeRegistrar) RegisterRuntimeModules(ctx *engine.RuntimeModuleContext, reg *noderequire.Registry) error {
	if ctx == nil {
		return fmt.Errorf("runtime module context is nil")
	}
	if reg == nil {
		return fmt.Errorf("require registry is nil")
	}

	reg.RegisterNativeModule("diff", func(vm *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		exports.Set("compareRegion", func(raw map[string]interface{}) (interface{}, error) {
			input, err := decodeInto[compareRegionInput](raw)
			if err != nil {
				return nil, err
			}
			settings := input.toCompareSettings()
			result, err := modes.GenerateCompareResult(ctx.Context, settings)
			if err != nil {
				return nil, err
			}
			if err := modes.WriteCompareArtifacts(result, settings.WriteJSON, settings.WriteMarkdown); err != nil {
				return nil, err
			}
			return toPlainValue(result)
		})
	})

	reg.RegisterNativeModule("report", func(vm *goja.Runtime, module *goja.Object) {
		exports := module.Get("exports").(*goja.Object)
		exports.Set("agentBrief", func(raw map[string]interface{}) (interface{}, error) {
			input, err := decodeInto[agentBriefInput](raw)
			if err != nil {
				return nil, err
			}
			brief := services.BuildAgentBrief(services.AgentBriefOptions{
				Question:   input.Question,
				Evidence:   input.Evidence,
				MaxBullets: input.MaxBullets,
			})
			return toPlainValue(brief)
		})
		exports.Set("renderAgentBrief", func(raw map[string]interface{}) (string, error) {
			input, err := decodeInto[agentBriefInput](raw)
			if err != nil {
				return "", err
			}
			brief := services.BuildAgentBrief(services.AgentBriefOptions{
				Question:   input.Question,
				Evidence:   input.Evidence,
				MaxBullets: input.MaxBullets,
			})
			return services.RenderAgentBriefText(brief), nil
		})
	})

	return nil
}

type compareRegionInput struct {
	Left       compareTargetInput `json:"left"`
	Right      compareTargetInput `json:"right"`
	Viewport   viewportInput      `json:"viewport"`
	Output     outputInput        `json:"output"`
	Computed   []string           `json:"computed"`
	Attributes []string           `json:"attributes"`
}

type compareTargetInput struct {
	URL      string `json:"url"`
	Selector string `json:"selector"`
	WaitMS   int    `json:"waitMs"`
}

type viewportInput struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type outputInput struct {
	OutDir        string `json:"outDir"`
	Threshold     int    `json:"threshold"`
	WriteJSON     bool   `json:"writeJson"`
	WriteMarkdown bool   `json:"writeMarkdown"`
	WritePNGs     bool   `json:"writePngs"`
}

type agentBriefInput struct {
	Question   string              `json:"question"`
	Evidence   modes.CompareResult `json:"evidence"`
	MaxBullets int                 `json:"maxBullets"`
}

func (input compareRegionInput) toCompareSettings() modes.CompareSettings {
	props := append([]string{}, input.Computed...)
	if len(props) == 0 {
		props = []string{
			"display",
			"position",
			"width",
			"height",
			"margin-top",
			"margin-right",
			"margin-bottom",
			"margin-left",
			"padding-top",
			"padding-right",
			"padding-bottom",
			"padding-left",
			"font-family",
			"font-size",
			"font-weight",
			"line-height",
			"color",
			"background-color",
			"background-image",
			"border-radius",
			"box-shadow",
			"z-index",
		}
	}

	attrs := append([]string{}, input.Attributes...)
	if len(attrs) == 0 {
		attrs = []string{"id", "class"}
	}

	viewportW := input.Viewport.Width
	if viewportW <= 0 {
		viewportW = 1280
	}
	viewportH := input.Viewport.Height
	if viewportH <= 0 {
		viewportH = 720
	}

	threshold := input.Output.Threshold
	if threshold == 0 {
		threshold = 30
	}

	leftSelector := strings.TrimSpace(input.Left.Selector)
	rightSelector := strings.TrimSpace(input.Right.Selector)
	if rightSelector == "" {
		rightSelector = leftSelector
	}

	return modes.CompareSettings{
		URL1:               input.Left.URL,
		Selector1:          leftSelector,
		WaitMS1:            input.Left.WaitMS,
		URL2:               input.Right.URL,
		Selector2:          rightSelector,
		WaitMS2:            input.Right.WaitMS,
		ViewportW:          viewportW,
		ViewportH:          viewportH,
		Props:              props,
		Attributes:         attrs,
		OutDir:             input.Output.OutDir,
		WriteJSON:          input.Output.WriteJSON,
		WriteMarkdown:      input.Output.WriteMarkdown,
		WritePNGs:          input.Output.WritePNGs,
		PixelDiffThreshold: threshold,
	}
}
