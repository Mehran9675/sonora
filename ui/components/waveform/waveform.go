// Package waveform draws a clip as rounded bars, colors the played
// part with the brand gradient and seeks on click or drag. While audio
// is being generated it shows an animated placeholder instead.
package waveform

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/ui/theme"
)

// Waveform is stateful (it owns pointer input). Create it once.
type Waveform struct {
	// Peaks returns the clip's loudness values in [0, 1].
	Peaks func() []float32
	// Progress returns the playback position in [0, 1].
	Progress func() float32
	// Busy reports that audio is being generated.
	Busy func() bool
	// Playing reports that audio is playing (frames are requested).
	Playing func() bool
	// OnSeek is called with the clicked position in [0, 1].
	OnSeek func(f float32)
	// Height in dp.
	Height int

	start time.Time
}

// Layout implements giode.Element.
func (w *Waveform) Layout(gtx layout.Context) layout.Dimensions {
	h := gtx.Dp(56)
	if w.Height > 0 {
		h = gtx.Dp(unit.Dp(w.Height))
	}
	size := image.Pt(gtx.Constraints.Max.X, h)
	peaks := w.Peaks()
	busy := w.Busy != nil && w.Busy()

	w.handleInput(gtx, size, len(peaks) > 0 && !busy)

	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	if len(peaks) > 0 && !busy {
		event.Op(gtx.Ops, w)
		pointer.CursorPointer.Add(gtx.Ops)
	}

	bar, gap := gtx.Dp(3), gtx.Dp(2)
	n := max(1, (size.X+gap)/(bar+gap))
	minH := gtx.Dp(3)
	played := w.Progress()
	from := properties.CalcColor(theme.BrandFrom)
	to := properties.CalcColor(theme.BrandTo)
	rest := properties.CalcColor("theme.text/0.16")

	now := gtx.Now
	if busy && w.start.IsZero() {
		w.start = now
	}
	if !busy {
		w.start = time.Time{}
	}
	t := now.Sub(w.start).Seconds()

	for i := 0; i < n; i++ {
		x := i * (bar + gap)
		frac := float32(x) / float32(max(1, size.X))
		var v float32
		var col color.NRGBA
		switch {
		case busy:
			// A travelling ripple while the model works.
			phase := float64(i)*0.32 - t*5
			v = 0.25 + 0.55*float32(math.Pow(math.Sin(phase)*0.5+0.5, 3))
			col = mix(from, to, frac)
			col.A = uint8(120 + 135*v)
		case len(peaks) == 0:
			v = 0.06
			col = rest
		default:
			v = bucket(peaks, i, n)
			if frac < played {
				col = mix(from, to, frac)
			} else {
				col = rest
			}
		}
		bh := max(minH, int(v*float32(h)))
		y := (h - bh) / 2
		r := clip.UniformRRect(image.Rect(x, y, x+bar, y+bh), bar/2)
		paint.FillShape(gtx.Ops, col, r.Op(gtx.Ops))
	}

	if busy || (w.Playing != nil && w.Playing()) {
		gtx.Execute(op.InvalidateCmd{At: now.Add(33 * time.Millisecond)})
	}
	return layout.Dimensions{Size: size}
}

func (w *Waveform) handleInput(gtx layout.Context, size image.Point, enabled bool) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: w, Kinds: pointer.Press | pointer.Drag})
		if !ok {
			return
		}
		e, ok := ev.(pointer.Event)
		if !ok || !enabled || w.OnSeek == nil {
			continue
		}
		w.OnSeek(max(0, min(1, e.Position.X/float32(max(1, size.X)))))
		gtx.Execute(op.InvalidateCmd{})
	}
}

// bucket returns the loudest peak among those falling in bar i of n.
func bucket(peaks []float32, i, n int) float32 {
	lo := i * len(peaks) / n
	hi := max(lo+1, (i+1)*len(peaks)/n)
	var v float32
	for _, p := range peaks[lo:min(hi, len(peaks))] {
		v = max(v, p)
	}
	return 0.08 + 0.92*v
}

func mix(a, b color.NRGBA, t float32) color.NRGBA {
	l := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }
	return color.NRGBA{R: l(a.R, b.R), G: l(a.G, b.G), B: l(a.B, b.B), A: l(a.A, b.A)}
}
