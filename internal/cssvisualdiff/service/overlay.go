package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/driver"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type OverlaySpec struct {
	Targets    []OverlayTarget `json:"targets"`
	Legend     bool            `json:"legend"`
	Screenshot string          `json:"screenshot"`
	Style      OverlayStyle    `json:"style"`
}

type OverlayTarget struct {
	Name     string             `json:"name"`
	Selector string             `json:"selector"`
	Label    string             `json:"label,omitempty"`
	Style    TargetOverlayStyle `json:"style"`
}

type OverlayResult struct {
	OutputPath string            `json:"outputPath"`
	Targets    []OverlayTarget   `json:"targets"`
	Colors     map[string]string `json:"colors"`
	Width      int               `json:"width"`
	Height     int               `json:"height"`
}

type documentBounds struct{ X, Y, Width, Height float64 }

type annotatedTarget struct {
	OverlayTarget
	Bounds documentBounds
	Style  TargetOverlayStyle
	Color  color.RGBA
}

var defaultPalette = []color.RGBA{
	{R: 0, G: 150, B: 255, A: 255},
	{R: 255, G: 99, B: 71, A: 255},
	{R: 50, G: 205, B: 50, A: 255},
	{R: 255, G: 191, B: 0, A: 255},
	{R: 186, G: 85, B: 211, A: 255},
	{R: 0, G: 206, B: 209, A: 255},
	{R: 255, G: 105, B: 180, A: 255},
	{R: 255, G: 140, B: 0, A: 255},
}

func OverlayScreenshot(page *driver.Page, spec OverlaySpec, outPath string) (*OverlayResult, error) {
	if page == nil {
		return nil, fmt.Errorf("page is required")
	}
	if outPath == "" {
		return nil, fmt.Errorf("output path is required")
	}
	if len(spec.Targets) == 0 {
		return nil, fmt.Errorf("at least one overlay target is required")
	}
	if spec.Screenshot == "" {
		spec.Screenshot = "fullPage"
	}
	if spec.Screenshot != "fullPage" {
		return nil, fmt.Errorf("unsupported overlay screenshot mode %q", spec.Screenshot)
	}
	spec.Style = NormalizeOverlayStyle(spec.Style)

	bounds, err := resolveDocumentBounds(page, spec.Targets)
	if err != nil {
		return nil, err
	}

	buf, err := page.FullScreenshotBytes()
	if err != nil {
		return nil, fmt.Errorf("capture full screenshot: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot image: %w", err)
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	annotated := make([]annotatedTarget, 0, len(spec.Targets))
	colors := map[string]string{}
	for i, target := range spec.Targets {
		if target.Name == "" {
			target.Name = SanitizeName(target.Selector)
		}
		style := ResolveTargetStyle(spec.Style, target.Style, defaultPalette[i%len(defaultPalette)])
		c := *style.BorderColor
		colors[target.Name] = RGBAHex(c)
		annotated = append(annotated, annotatedTarget{OverlayTarget: target, Bounds: bounds[i], Style: style, Color: c})
	}

	for _, target := range annotated {
		drawTargetBox(rgba, target)
	}
	for _, target := range annotated {
		drawTargetLabel(rgba, target, spec.Style.Label)
	}
	if spec.Legend {
		drawOverlayLegend(rgba, annotated, spec.Style)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("create output png: %w", err)
	}
	if err := png.Encode(f, rgba); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("encode output png: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close output png: %w", err)
	}

	return &OverlayResult{OutputPath: outPath, Targets: spec.Targets, Colors: colors, Width: rgba.Bounds().Dx(), Height: rgba.Bounds().Dy()}, nil
}

func resolveDocumentBounds(page *driver.Page, targets []OverlayTarget) ([]documentBounds, error) {
	payload, err := json.Marshal(targets)
	if err != nil {
		return nil, fmt.Errorf("marshal overlay targets: %w", err)
	}
	script := fmt.Sprintf(`(() => {
const targets = %s;
return targets.map((target) => {
  const selector = target.selector || "";
  if (!selector) return { error: "selector is empty" };
  let el;
  try { el = document.querySelector(selector); } catch (err) { return { error: String(err && err.message ? err.message : err) }; }
  if (!el) return { error: "selector not found" };
  const rect = el.getBoundingClientRect();
  return { x: rect.x + window.scrollX, y: rect.y + window.scrollY, width: rect.width, height: rect.height };
});
})()`, string(payload))
	var raw []struct {
		X, Y, Width, Height float64
		Error               string
	}
	if err := page.Evaluate(script, &raw); err != nil {
		return nil, fmt.Errorf("evaluate overlay bounds: %w", err)
	}
	if len(raw) != len(targets) {
		return nil, fmt.Errorf("overlay bounds length mismatch")
	}
	ret := make([]documentBounds, len(raw))
	for i, b := range raw {
		if b.Error != "" {
			return nil, fmt.Errorf("overlay target %q (%s): %s", targets[i].Name, targets[i].Selector, b.Error)
		}
		ret[i] = documentBounds{X: b.X, Y: b.Y, Width: b.Width, Height: b.Height}
	}
	return ret, nil
}

func drawTargetBox(img *image.RGBA, target annotatedTarget) {
	x := int(math.Round(target.Bounds.X))
	y := int(math.Round(target.Bounds.Y))
	w := int(math.Round(target.Bounds.Width))
	h := int(math.Round(target.Bounds.Height))
	if w <= 0 || h <= 0 {
		return
	}
	fill := target.Style.ContentBackground
	if fill != nil && fill.A > 0 {
		drawRect(img, image.Rect(x, y, x+w, y+h), *fill)
	}
	border := target.Color
	bw := target.Style.BorderWidth
	if bw <= 0 {
		bw = 2
	}
	for i := 0; i < bw; i++ {
		drawRectOutline(img, image.Rect(x-i, y-i, x+w+i, y+h+i), border)
	}
}

func drawTargetLabel(img *image.RGBA, target annotatedTarget, global LabelOverlayStyle) {
	label := target.Label
	if label == "" {
		label = target.Name
	}
	face := basicfont.Face7x13
	pad := global.Padding
	if pad == (Insets{}) {
		pad = Insets{Top: 4, Right: 7, Bottom: 4, Left: 7}
	}
	textW := font.MeasureString(face, label).Ceil()
	textH := face.Metrics().Height.Ceil()
	w := textW + pad.Left + pad.Right
	h := textH + pad.Top + pad.Bottom
	x := int(math.Round(target.Bounds.X))
	y := int(math.Round(target.Bounds.Y)) - h - 2
	if target.Style.Label.Position == LabelBelow {
		y = int(math.Round(target.Bounds.Y+target.Bounds.Height)) + 2
	}
	if target.Style.Label.Position == LabelInsideStart || y < 0 {
		y = int(math.Round(target.Bounds.Y)) + 2
	}
	if target.Style.Label.Position == LabelInsideEnd {
		y = int(math.Round(target.Bounds.Y+target.Bounds.Height)) - h - 2
	}
	if x+w > img.Bounds().Max.X {
		x = img.Bounds().Max.X - w - 2
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if y+h > img.Bounds().Max.Y {
		y = img.Bounds().Max.Y - h
	}
	bg := target.Color
	if target.Style.Label.Background != nil {
		bg = *target.Style.Label.Background
	}
	fg := color.RGBA{255, 255, 255, 255}
	if target.Style.Label.Color != nil {
		fg = *target.Style.Label.Color
	}
	drawRect(img, image.Rect(x, y, x+w, y+h), bg)
	d := font.Drawer{Dst: img, Src: image.NewUniform(fg), Face: face, Dot: fixed.P(x+pad.Left, y+pad.Top+face.Ascent)}
	d.DrawString(label)
}

func drawOverlayLegend(img *image.RGBA, targets []annotatedTarget, style OverlayStyle) {
	if len(targets) == 0 {
		return
	}
	face := basicfont.Face7x13
	pad := 10
	rowH := 20
	swatch := 11
	maxText := 0
	for _, t := range targets {
		if w := font.MeasureString(face, t.Name).Ceil(); w > maxText {
			maxText = w
		}
	}
	w := pad*2 + swatch + 8 + maxText
	h := pad*2 + rowH*len(targets)
	bounds := img.Bounds()
	x := bounds.Max.X - w - 12
	y := bounds.Max.Y - h - 12
	switch style.Legend.Position {
	case LegendTopLeft:
		x, y = 12, 12
	case LegendTopRight:
		x, y = bounds.Max.X-w-12, 12
	case LegendBottomLeft:
		x, y = 12, bounds.Max.Y-h-12
	case LegendBottomRight, "":
		x, y = bounds.Max.X-w-12, bounds.Max.Y-h-12
	}
	bg := color.RGBA{255, 255, 255, 235}
	if style.Legend.Background != nil {
		bg = *style.Legend.Background
	}
	fg := color.RGBA{39, 34, 27, 255}
	if style.Legend.Color != nil {
		fg = *style.Legend.Color
	}
	drawRect(img, image.Rect(x, y, x+w, y+h), bg)
	for i, t := range targets {
		ry := y + pad + i*rowH + 4
		drawRect(img, image.Rect(x+pad, ry, x+pad+swatch, ry+swatch), t.Color)
		d := font.Drawer{Dst: img, Src: image.NewUniform(fg), Face: face, Dot: fixed.P(x+pad+swatch+8, ry+face.Ascent)}
		d.DrawString(t.Name)
	}
}

func drawRect(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	r = r.Intersect(img.Bounds())
	if r.Empty() {
		return
	}
	draw.Draw(img, r, image.NewUniform(c), image.Point{}, draw.Over)
}

func drawRectOutline(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	drawRect(img, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), c)
	drawRect(img, image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), c)
	drawRect(img, image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), c)
	drawRect(img, image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), c)
}
