// Package levelmeter shows live input loudness as a row of bars that
// scroll from right to left, like a recording app's meter.
package levelmeter

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/mehran9675/giode/styles/properties"
)

// Meter is stateless apart from its data source.
type Meter struct {
	// Levels returns recent loudness values in [0, 1], oldest first.
	Levels func() []float32
	// Color is the bar color token.
	Color string
}

// Layout implements giode.Element. It redraws continuously while laid
// out, so only show it while recording.
func (m *Meter) Layout(gtx layout.Context) layout.Dimensions {
	h := gtx.Dp(unit.Dp(36))
	w := gtx.Constraints.Max.X
	bar, gap := gtx.Dp(3), gtx.Dp(2)
	n := max(1, (w+gap)/(bar+gap))
	levels := m.Levels()
	col := properties.CalcColor(m.Color)
	quiet := properties.CalcColor("theme.text/0.14")
	for i := 0; i < n; i++ {
		// Newest level at the right edge.
		j := len(levels) - n + i
		v := float32(0)
		if j >= 0 {
			v = levels[j]
		}
		bh := max(gtx.Dp(3), int(v*float32(h)))
		x := i * (bar + gap)
		y := (h - bh) / 2
		c := quiet
		if v > 0.05 {
			c = col
		}
		paint.FillShape(gtx.Ops, c, clip.UniformRRect(image.Rect(x, y, x+bar, y+bh), bar/2).Op(gtx.Ops))
	}
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(33 * time.Millisecond)})
	return layout.Dimensions{Size: image.Pt(w, h)}
}
