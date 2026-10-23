// Package pagehead is the title row at the top of each page.
package pagehead

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/theme"
)

// Head returns a title, subtitle and trailing actions.
func Head(title, subtitle string, actions ...giode.Element) giode.Element {
	text := giode.Stack(giode.Styles{Gap: 4, FlexGrow: 1},
		giode.Text(title, theme.Title),
		giode.Text(subtitle, theme.Subtle),
	)
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap},
		append([]giode.Element{text}, actions...)...)
}
