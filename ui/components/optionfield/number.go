package optionfield

import (
	"math"
	"strconv"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/slider"

	"github.com/mehran9675/sonora/internal/engines"
)

// number is a slider over [Min, Max] snapped to Step.
type number struct {
	opt engines.Option
	st  Store
	s   *slider.Slider
}

func newNumber(opt engines.Option, st Store) *number {
	n := &number{opt: opt, st: st}
	n.s = giode.Input.Slider(giode.Styles{Color: "theme.primary", Background: "theme.text/0.12", BorderColor: "theme.primary"}).
		OnChange(func(v float32) { st.Set(opt.ID, n.fromUnit(v)) })
	return n
}

func (n *number) value() float64 {
	switch v := valueOr(n.st.Get(n.opt.ID), n.opt.Default).(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return n.opt.Min
}

func (n *number) fromUnit(u float32) float64 {
	v := n.opt.Min + float64(u)*(n.opt.Max-n.opt.Min)
	if n.opt.Step > 0 {
		v = math.Round(v/n.opt.Step) * n.opt.Step
	}
	return math.Round(v*1000) / 1000
}

func (n *number) display() string {
	v := n.value()
	if v == 0 && n.opt.Min == 0 && n.opt.Help != "" {
		return "Auto"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (n *number) Element() giode.Element {
	if !n.s.Dragging() && n.opt.Max > n.opt.Min {
		n.s.SetValue(float32((n.value() - n.opt.Min) / (n.opt.Max - n.opt.Min)))
	}
	return giode.Box(giode.Styles{Width: -1}, n.s)
}
