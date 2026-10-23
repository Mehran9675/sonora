// Package emptystate is the placeholder shown when a list is empty.
package emptystate

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// New returns a centered icon, title and hint, plus optional actions.
func New(icon, title, hint string, actions ...giode.Element) giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Padding: giode.UniformInset(40)},
		giode.Box(giode.Styles{
			Width: 64, Height: 64, BorderRadius: 20, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
			Background: "theme.primary/0.12",
		}, icons.Icon(icon, 28, "theme.primary")),
		giode.Box(giode.Styles{Height: 4}),
		giode.Text(title, theme.Heading),
		giode.Text(hint, giode.Styles{Color: "theme.muted", FontSize: 13, TextAlign: giode.TextAlignCenter, MaxWidth: 380}),
		giode.Row(giode.Styles{Gap: 8, Margin: giode.Inset{Top: 8}}, actions...),
	)
}
