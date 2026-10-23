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
// the right.
func setting(label, help string, control giode.Element) giode.Element {
	return giode.Row(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(label, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium}),
			giode.Text(help, theme.Caption),
		),
		giode.Box(giode.Styles{Width: 320}, control),
	)
}
