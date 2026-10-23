// Package badge draws an engine's avatar: a gradient tile with its
// initials.
package badge

import (
	"strings"
	"unicode"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/theme"
)

// Engine returns the avatar of m at size dp.
func Engine(m *engines.Manifest, size int) giode.Element {
	return giode.Box(giode.Styles{
		Width:              properties.Width(size),
		Height:             properties.Height(size),
		BorderRadius:       properties.BorderRadius(size * 3 / 10),
		BackgroundGradient: theme.Gradient(135, m.Accent),
		Align:              giode.AlignCenter,
		Justify:            giode.JustifyCenter,
		BoxShadow:          giode.BoxShadow{Y: 4, Blur: 12, Color: "#00000033"},
	}, giode.Text(Initials(m.Name), giode.Styles{
		Color:      "#ffffff",
		FontSize:   properties.FontSize(float32(size) * 0.36),
		FontWeight: giode.FontWeightExtraBold,
	}))
}

// Initials returns up to two letters standing for name.
func Initials(name string) string {
	var out []rune
	for _, w := range strings.Fields(name) {
		r := []rune(w)
		if len(r) > 0 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0])) {
			out = append(out, unicode.ToUpper(r[0]))
		}
		if len(out) == 2 {
			break
		}
	}
	if len(out) == 0 {
		return "?"
	}
	return string(out)
}
