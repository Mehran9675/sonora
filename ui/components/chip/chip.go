// Package chip draws small pill labels for tags and metadata.
package chip

import "github.com/mehran9675/giode"

// Chip is a neutral pill.
func Chip(text string) giode.Element {
	return giode.Box(giode.Styles{
		Background:   "theme.text/0.06",
		BorderRadius: 99,
		Padding:      giode.SymmetricInset(9, 3),
	}, giode.Text(text, giode.Styles{Color: "theme.secondary", FontSize: 11, FontWeight: giode.FontWeightMedium}))
}

// Accent is a pill tinted with a theme color token ("theme.success").
func Accent(text, color string) giode.Element {
	return giode.Box(giode.Styles{
		Background:   color + "/0.14",
		BorderRadius: 99,
		Padding:      giode.SymmetricInset(9, 3),
	}, giode.Text(text, giode.Styles{Color: color, FontSize: 11, FontWeight: giode.FontWeightBold}))
}
