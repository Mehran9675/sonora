// Package progressbar draws one slim progress bar: filled to a known
// fraction, or a sliding stripe while the amount of work is unknown.
package progressbar

import (
	"image"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/ui/theme"
)

// Bar is stateless; build it every frame.
type Bar struct {
	// Value is the fraction done in [0, 1], or negative when unknown.
	Value float32
}

// New returns a bar showing v (negative for indeterminate).
func New(v float32) *Bar { return &Bar{Value: v} }

// Layout implements giode.Element.
func (b *Bar) Layout(gtx layout.Context) layout.Dimensions {
	h := gtx.Dp(unit.Dp(6))
	w := gtx.Constraints.Max.X
	size := image.Pt(w, h)
	r := h / 2
	track := properties.CalcColor("theme.text/0.08")
	paint.FillShape(gtx.Ops, track, clip.UniformRRect(image.Rect(0, 0, w, h), r).Op(gtx.Ops))

	var lo, hi int
	if b.Value >= 0 {
		lo, hi = 0, int(float32(w)*min(1, b.Value))
	} else {
		// A stripe a third of the width, gliding across every 1.4 s.
		t := math.Mod(float64(gtx.Now.UnixMilli())/1400, 1)
		stripe := float64(w) / 3
		start := -stripe + t*(float64(w)+stripe)
		lo, hi = max(0, int(start)), min(w, int(start+stripe))
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(16 * time.Millisecond)})
	}
	if hi > lo {
		defer clip.UniformRRect(image.Rect(0, 0, w, h), r).Push(gtx.Ops).Pop()
		from, to := properties.CalcColor(theme.BrandFrom), properties.CalcColor(theme.BrandTo)
		paint.LinearGradientOp{Stop1: f32.Pt(float32(lo), 0), Stop2: f32.Pt(float32(max(hi, lo+1)), 0), Color1: from, Color2: to}.Add(gtx.Ops)
		defer clip.UniformRRect(image.Rect(lo, 0, hi, h), r).Push(gtx.Ops).Pop()
		paint.PaintOp{}.Add(gtx.Ops)
	}
	return layout.Dimensions{Size: size}
}
