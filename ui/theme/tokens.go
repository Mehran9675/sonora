package theme

import "github.com/mehran9675/giode"

// Brand gradient stops.
const (
	BrandFrom = "#8b5cf6"
	BrandTo   = "#22d3ee"
)

// Corner radii.
const (
	RadiusSm = 8
	RadiusMd = 12
	RadiusLg = 16
)

// Brand returns the signature violet-to-cyan gradient.
func Brand(angle float32) *giode.LinearGradient {
	return giode.Linear(angle, BrandFrom, BrandTo)
}

// Gradient builds a gradient from an engine's accent colors.
func Gradient(angle float32, colors []string) *giode.LinearGradient {
	if len(colors) < 2 {
		return Brand(angle)
	}
	return giode.Linear(angle, colors[0], colors[1])
}

// Card is the panel style used across pages.
func Card() giode.Styles {
	return giode.Styles{
		Background:    "theme.surface",
		BorderWidth:   1,
		BorderColor:   "theme.border",
		BorderRadius:  RadiusLg,
		Padding:       giode.UniformInset(20),
		FlexDirection: giode.FlexDirectionColumn,
		Gap:           14,
	}
}

// Text styles.
var (
	Title    = giode.Styles{Color: "theme.text", FontSize: 26, FontWeight: giode.FontWeightExtraBold}
	Heading  = giode.Styles{Color: "theme.text", FontSize: 15, FontWeight: giode.FontWeightBold}
	Body     = giode.Styles{Color: "theme.text", FontSize: 14}
	Subtle   = giode.Styles{Color: "theme.muted", FontSize: 13}
	Caption  = giode.Styles{Color: "theme.muted", FontSize: 12}
	Overline = giode.Styles{Color: "theme.muted", FontSize: 11, FontWeight: giode.FontWeightBold,
		TextTransform: giode.TextTransformUppercase}
	MonoText = giode.Styles{Color: "theme.secondary", FontSize: 12, FontFamily: Mono}
)

// Field is the style of text inputs.
func Field() giode.Styles {
	return giode.Styles{
		Background:   "theme.text/0.03",
		BorderColor:  "theme.border",
		BorderWidth:  1,
		BorderRadius: RadiusSm + 2,
		Padding:      giode.SymmetricInset(12, 10),
		Color:        "theme.text",
		FontSize:     14,
	}
}

// Scrolling returns the style of a scroll area whose scrollbar is width
// dp wide. Phones (the "compact" layout) show none, as their own apps
// don't; the content still scrolls by touch.
func Scrolling(width int) giode.Styles {
	return giode.Styles{ScrollBar: giode.ScrollBar{
		ThumbColor: "theme.text/0.15", Width: width, HiddenOn: "compact",
	}}
}
