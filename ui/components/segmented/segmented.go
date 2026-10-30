// Package segmented is a row of mutually exclusive buttons.
package segmented

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Option is one segment.
type Option struct {
	Value string
	Label string // English (tr.Msg); shown translated
}

// Control is stateful; create it once per set of options.
type Control struct {
	opts     []Option
	selected func() string
	segs     []*giode.PressableElement
}

// New returns a control; selected returns the current value and
// onChange receives a new one.
func New(opts []Option, selected func() string, onChange func(string)) *Control {
	c := &Control{opts: opts, selected: selected}
	for _, o := range opts {
		o := o
		seg := giode.Pressable(func(s giode.PressState) giode.Element {
			on := c.selected() == o.Value
			bg, fg := "", "theme.muted"
			switch {
			case on:
				bg, fg = "theme.surface", "theme.text"
			case s.Hovered:
				fg = "theme.text"
			}
			st := giode.Styles{
				Width: -1, Justify: giode.JustifyCenter, Align: giode.AlignCenter,
				Padding: giode.SymmetricInset(10, 7), BorderRadius: theme.RadiusSm, Background: bg,
			}
			if on {
				st.BoxShadow = giode.BoxShadow{Y: 1, Blur: 4, Color: "#00000030"}
			}
			return giode.Row(st, giode.Text(tr.T(o.Label), giode.Styles{Color: fg, FontSize: 13, FontWeight: giode.FontWeightMedium,
				TextAlign: giode.TextAlignCenter, TextOverflow: giode.TextOverflowEllipsis}))
		}, giode.Styles{FlexGrow: 1}).OnClick(func(giode.ClickEvent) { onChange(o.Value) })
		tr.Bind(func(s string) { seg.Label(s) }, o.Label)
		c.segs = append(c.segs, seg)
	}
	return c
}

// Element returns the control for layout.
func (c *Control) Element() giode.Element {
	els := make([]giode.Element, len(c.segs))
	for i, s := range c.segs {
		els[i] = s
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 2, Padding: giode.UniformInset(3), Wrap: giode.WrapNoWrap,
		Background: "theme.text/0.06", BorderRadius: theme.RadiusSm + 2,
	}, els...)
}
