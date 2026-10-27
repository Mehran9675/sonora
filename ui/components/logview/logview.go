// Package logview shows an engine's log in a monospaced, scrollable
// panel that follows new lines.
package logview

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/ui/theme"
)

// View is a stateful log panel. Create it once per log.
type View struct {
	scroll *scroll.Scroll
	lines  func() []string
	seen   int
}

// New returns a panel showing lines().
func New(lines func() []string) *View {
	v := &View{lines: lines}
	v.scroll = giode.Scroll(func() giode.Element {
		ls := v.lines()
		rows := make([]giode.Element, 0, len(ls))
		for _, l := range ls {
			rows = append(rows, giode.Text(l, theme.MonoText))
		}
		return giode.Stack(giode.Styles{Width: -1, Gap: 2, Padding: giode.UniformInset(12)}, rows...)
	}, theme.Scrolling(6))
	return v
}

// Element returns the panel at height dp.
func (v *View) Element(height int) giode.Element {
	if n := len(v.lines()); n != v.seen {
		v.seen = n
		v.scroll.ScrollTo(1e9) // follow the tail
	}
	return giode.Box(giode.Styles{
		Width: -1, Height: properties.Height(height), Background: "theme.background", BorderRadius: theme.RadiusSm + 2,
		BorderWidth: 1, BorderColor: "theme.border", Overflow: giode.OverflowHidden,
	}, v.scroll)
}
