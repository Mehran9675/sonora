package icons

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"
)

// Icon strokes path at size dp in color (empty follows the parent's
// text color).
func Icon(path string, size int, color string) giode.Element {
	return giode.IconSVGStroke(path, 2, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}

// Solid fills path instead of stroking it (used for the play glyph).
func Solid(path string, size int, color string) giode.Element {
	return giode.IconSVG(path, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}
