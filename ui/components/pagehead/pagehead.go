// Package pagehead is the title row at the top of each page.
package pagehead

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/theme"
)

// Head returns a title, subtitle and trailing actions. On narrow
// screens (compact) the actions move below the title.
func Head(compact bool, title, subtitle string, actions ...giode.Element) giode.Element {
	titleStyle := theme.Title
	if compact {
		titleStyle.FontSize = 22
	}
	textStyle := giode.Styles{Gap: 4, FlexGrow: 1}
	if compact {
		// Stacked vertically, growing would push the page content away.
		textStyle.FlexGrow = 0
	}
	text := giode.Stack(textStyle,
		giode.Text(title, titleStyle),
		giode.Text(subtitle, theme.Subtle),
	)
	if compact {
		if len(actions) == 0 {
			return text
		}
		return giode.Stack(giode.Styles{Width: -1, Gap: 12}, text,
			giode.Row(giode.Styles{Width: -1, Gap: 8}, actions...))
	}
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap},
		append([]giode.Element{text}, actions...)...)
}
