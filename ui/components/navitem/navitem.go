// Package navitem is a navigation entry: a row in the sidebar, or a
// tab in the bottom bar on narrow screens.
package navitem

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Item is a clickable sidebar row. Create it once.
type Item struct {
	p       *giode.PressableElement
	compact bool
}

// New returns an item showing icon and label; active reports whether
// it is the current page; badge (optional) returns a count to show.
// label is English (tr.Msg); it shows in the current language.
func New(icon, label string, active func() bool, badge func() string, onClick func()) *Item {
	it := &Item{}
	it.p = giode.Pressable(func(s giode.PressState) giode.Element {
		on := active()
		bg, fg := "", "theme.muted"
		switch {
		case on:
			bg, fg = "theme.primary/0.16", "theme.text"
		case s.Hovered:
			bg, fg = "theme.text/0.05", "theme.text"
		}
		iconColor := fg
		if on {
			iconColor = "theme.primary"
		}
		if it.compact {
			// A bottom-bar tab: icon over a short label, no badge.
			return giode.Stack(giode.Styles{
				Width: -1, Gap: 4, Align: giode.AlignCenter, Padding: giode.SymmetricInset(4, 8),
				BorderRadius: theme.RadiusSm + 2, Background: bg,
			}, icons.Icon(icon, 20, iconColor), giode.Text(tr.T(label), giode.Styles{Color: fg, FontSize: 11, FontWeight: giode.FontWeightMedium}))
		}
		row := []giode.Element{
			icons.Icon(icon, 18, iconColor),
			giode.Box(giode.Styles{FlexGrow: 1},
				giode.Text(tr.T(label), giode.Styles{Color: fg, FontSize: 14, FontWeight: giode.FontWeightMedium})),
		}
		if badge != nil {
			if b := badge(); b != "" {
				row = append(row, giode.Box(giode.Styles{
					Background: "theme.text/0.08", BorderRadius: 99, Padding: giode.SymmetricInset(7, 1),
				}, giode.Text(b, giode.Styles{Color: "theme.muted", FontSize: 11, FontWeight: giode.FontWeightBold})))
			}
		}
		return giode.Row(giode.Styles{
			Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 10), BorderRadius: theme.RadiusSm + 2, Background: bg,
		}, row...)
	}).OnClick(func(giode.ClickEvent) { onClick() })
	tr.Bind(func(s string) { it.p.Label(s) }, label)
	return it
}

// Element returns the item as a sidebar row.
func (it *Item) Element() giode.Element {
	it.compact = false
	return it.p
}

// Tab returns the item as an equal-width bottom-bar tab.
func (it *Item) Tab() giode.Element {
	it.compact = true
	return giode.Box(giode.Styles{FlexGrow: 1}, it.p)
}
