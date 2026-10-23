// Package settings is the Settings screen.
package settings

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/theme"
)

// section is a titled settings card.
func section(title, subtitle string, rows ...giode.Element) giode.Element {
	st := theme.Card()
	st.Width, st.Gap = -1, 18
	head := giode.Stack(giode.Styles{Gap: 4},
		giode.Text(title, theme.Heading),
		giode.Text(subtitle, theme.Caption),
	)
	return giode.Stack(st, append([]giode.Element{head}, rows...)...)
}

// setting is one labelled row: label and help on the left, control on
// the right - or the control below the label on narrow screens.
func (p *Page) setting(label, help string, control giode.Element) giode.Element {
	textStyle := giode.Styles{FlexGrow: 1, Gap: 3}
	if giode.Compact() {
		textStyle.FlexGrow = 0 // stacked: must not grow vertically
	}
	text := giode.Stack(textStyle,
		giode.Text(label, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium}),
		giode.Text(help, theme.Caption),
	)
	if giode.Compact() {
		return giode.Stack(giode.Styles{Width: -1, Gap: 10}, text, giode.Box(giode.Styles{Width: -1}, control))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		text, giode.Box(giode.Styles{Width: 320}, control))
}
