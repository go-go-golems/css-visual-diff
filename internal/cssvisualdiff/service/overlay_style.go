package service

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

type LegendPosition string

const (
	LegendTopLeft     LegendPosition = "top-left"
	LegendTopRight    LegendPosition = "top-right"
	LegendBottomLeft  LegendPosition = "bottom-left"
	LegendBottomRight LegendPosition = "bottom-right"
)

type LabelPosition string

const (
	LabelAuto        LabelPosition = "auto"
	LabelAbove       LabelPosition = "above"
	LabelBelow       LabelPosition = "below"
	LabelInsideStart LabelPosition = "inside-start"
	LabelInsideEnd   LabelPosition = "inside-end"
)

type Insets struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type OverlayStyle struct {
	Label          LabelOverlayStyle  `json:"label"`
	Legend         LegendOverlayStyle `json:"legend"`
	TargetDefaults TargetOverlayStyle `json:"targetDefaults"`
}

type LabelOverlayStyle struct {
	FontFamily string `json:"fontFamily,omitempty"`
	FontSize   int    `json:"fontSize,omitempty"`
	Radius     int    `json:"radius,omitempty"`
	Padding    Insets `json:"padding"`
}

type LegendOverlayStyle struct {
	Position   LegendPosition `json:"position,omitempty"`
	Background *color.RGBA    `json:"background,omitempty"`
	Color      *color.RGBA    `json:"color,omitempty"`
}

type TargetOverlayStyle struct {
	BorderColor       *color.RGBA      `json:"borderColor,omitempty"`
	ContentBackground *color.RGBA      `json:"contentBackground,omitempty"`
	PaddingBackground *color.RGBA      `json:"paddingBackground,omitempty"`
	MarginBackground  *color.RGBA      `json:"marginBackground,omitempty"`
	BorderWidth       int              `json:"borderWidth,omitempty"`
	LabelColor        *color.RGBA      `json:"labelColor,omitempty"`
	Label             LabelTargetStyle `json:"label"`
}

type LabelTargetStyle struct {
	Background *color.RGBA   `json:"background,omitempty"`
	Color      *color.RGBA   `json:"color,omitempty"`
	Position   LabelPosition `json:"position,omitempty"`
}

func DefaultOverlayStyle() OverlayStyle {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 235}
	text := color.RGBA{R: 39, G: 34, B: 27, A: 255}
	return OverlayStyle{
		Label:          LabelOverlayStyle{FontFamily: "basicfont", FontSize: 13, Radius: 3, Padding: Insets{Top: 4, Right: 7, Bottom: 4, Left: 7}},
		Legend:         LegendOverlayStyle{Position: LegendBottomRight, Background: &white, Color: &text},
		TargetDefaults: TargetOverlayStyle{BorderWidth: 2, LabelColor: &white, Label: LabelTargetStyle{Position: LabelAuto}},
	}
}

func NormalizeOverlayStyle(style OverlayStyle) OverlayStyle {
	defaults := DefaultOverlayStyle()
	defaults = defaults.Merge(style)
	return defaults
}

func (s OverlayStyle) Merge(o OverlayStyle) OverlayStyle {
	if o.Label.FontFamily != "" {
		s.Label.FontFamily = o.Label.FontFamily
	}
	if o.Label.FontSize > 0 {
		s.Label.FontSize = o.Label.FontSize
	}
	if o.Label.Radius > 0 {
		s.Label.Radius = o.Label.Radius
	}
	if o.Label.Padding != (Insets{}) {
		s.Label.Padding = o.Label.Padding
	}
	if o.Legend.Position != "" {
		s.Legend.Position = o.Legend.Position
	}
	if o.Legend.Background != nil {
		s.Legend.Background = o.Legend.Background
	}
	if o.Legend.Color != nil {
		s.Legend.Color = o.Legend.Color
	}
	s.TargetDefaults = s.TargetDefaults.Merge(o.TargetDefaults)
	return s
}

func (s TargetOverlayStyle) Merge(o TargetOverlayStyle) TargetOverlayStyle {
	if o.BorderColor != nil {
		s.BorderColor = o.BorderColor
	}
	if o.ContentBackground != nil {
		s.ContentBackground = o.ContentBackground
	}
	if o.PaddingBackground != nil {
		s.PaddingBackground = o.PaddingBackground
	}
	if o.MarginBackground != nil {
		s.MarginBackground = o.MarginBackground
	}
	if o.BorderWidth > 0 {
		s.BorderWidth = o.BorderWidth
	}
	if o.LabelColor != nil {
		s.LabelColor = o.LabelColor
	}
	if o.Label.Background != nil {
		s.Label.Background = o.Label.Background
	}
	if o.Label.Color != nil {
		s.Label.Color = o.Label.Color
	}
	if o.Label.Position != "" {
		s.Label.Position = o.Label.Position
	}
	return s
}

func ResolveTargetStyle(global OverlayStyle, target TargetOverlayStyle, palette color.RGBA) TargetOverlayStyle {
	style := global.TargetDefaults.Merge(target)
	if style.BorderColor == nil {
		style.BorderColor = &palette
	}
	if style.ContentBackground == nil {
		c := WithAlpha(*style.BorderColor, 0.10)
		style.ContentBackground = &c
	}
	if style.Label.Background == nil {
		style.Label.Background = style.BorderColor
	}
	if style.Label.Color == nil {
		if style.LabelColor != nil {
			style.Label.Color = style.LabelColor
		} else {
			c := color.RGBA{255, 255, 255, 255}
			style.Label.Color = &c
		}
	}
	if style.Label.Position == "" {
		style.Label.Position = LabelAuto
	}
	if style.BorderWidth <= 0 {
		style.BorderWidth = 2
	}
	return style
}

func WithAlpha(c color.RGBA, a float64) color.RGBA {
	if a < 0 {
		a = 0
	}
	if a > 1 {
		a = 1
	}
	c.A = uint8(math.Round(a * 255))
	return c
}

func ParseColor(s string) (color.RGBA, error) {
	raw := strings.TrimSpace(strings.ToLower(s))
	switch raw {
	case "white":
		return color.RGBA{255, 255, 255, 255}, nil
	case "black":
		return color.RGBA{0, 0, 0, 255}, nil
	case "red":
		return color.RGBA{255, 0, 0, 255}, nil
	case "blue":
		return color.RGBA{0, 0, 255, 255}, nil
	case "green":
		return color.RGBA{0, 128, 0, 255}, nil
	case "transparent":
		return color.RGBA{0, 0, 0, 0}, nil
	}
	if strings.HasPrefix(raw, "#") {
		return parseHexColor(raw)
	}
	if strings.HasPrefix(raw, "rgb(") || strings.HasPrefix(raw, "rgba(") {
		return parseRGBColor(raw)
	}
	return color.RGBA{}, fmt.Errorf("unsupported color %q", s)
}

func parseHexColor(s string) (color.RGBA, error) {
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 && len(h) != 8 {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid hex color %q", s)
	}
	if len(h) == 6 {
		return color.RGBA{R: uint8((v >> 16) & 0xff), G: uint8((v >> 8) & 0xff), B: uint8(v & 0xff), A: 255}, nil
	}
	return color.RGBA{R: uint8((v >> 24) & 0xff), G: uint8((v >> 16) & 0xff), B: uint8((v >> 8) & 0xff), A: uint8(v & 0xff)}, nil
}

func parseRGBColor(s string) (color.RGBA, error) {
	open := strings.IndexByte(s, '(')
	closeIdx := strings.LastIndexByte(s, ')')
	if open < 0 || closeIdx <= open {
		return color.RGBA{}, fmt.Errorf("invalid rgb color %q", s)
	}
	parts := strings.Split(s[open+1:closeIdx], ",")
	if len(parts) != 3 && len(parts) != 4 {
		return color.RGBA{}, fmt.Errorf("invalid rgb color %q", s)
	}
	vals := [3]uint8{}
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 || n > 255 {
			return color.RGBA{}, fmt.Errorf("invalid rgb color %q", s)
		}
		vals[i] = uint8(n)
	}
	a := uint8(255)
	if len(parts) == 4 {
		f, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil || f < 0 || f > 1 {
			return color.RGBA{}, fmt.Errorf("invalid alpha in %q", s)
		}
		a = uint8(math.Round(f * 255))
	}
	return color.RGBA{R: vals[0], G: vals[1], B: vals[2], A: a}, nil
}

func ParseLegendPosition(s string) (LegendPosition, error) {
	p := LegendPosition(strings.TrimSpace(s))
	switch p {
	case LegendTopLeft, LegendTopRight, LegendBottomLeft, LegendBottomRight:
		return p, nil
	}
	return "", fmt.Errorf("unsupported legend position %q", s)
}

func ParseLabelPosition(s string) (LabelPosition, error) {
	p := LabelPosition(strings.TrimSpace(s))
	switch p {
	case LabelAuto, LabelAbove, LabelBelow, LabelInsideStart, LabelInsideEnd:
		return p, nil
	}
	return "", fmt.Errorf("unsupported label position %q", s)
}

func RGBAHex(c color.RGBA) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }
