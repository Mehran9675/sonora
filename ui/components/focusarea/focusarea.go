// Package focusarea makes a whole text field clickable. giode's input
// only reacts to clicks on its text (80dp wide when empty), so a click
// elsewhere in the field would otherwise do nothing.
package focusarea

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/styles/properties"
)

// Area wraps an input. Create it once per input.
type Area struct {
	p *giode.PressableElement
}

// New returns an area of the given size (Width, Height, FlexGrow…)
// that focuses in when clicked anywhere.
func New(in *input.Input, st giode.Styles) *Area {
	st.Cursor = giode.CursorText
	content := giode.Box(giode.Styles{Width: -1, Height: heightOf(st)}, in)
	p := giode.Pressable(func(giode.PressState) giode.Element { return content }, st).
		Focusable(false).
		OnClick(func(giode.ClickEvent) { in.Focus() })
	return &Area{p: p}
}

// heightOf fills the area vertically when it has a height, and hugs
// the input otherwise.
func heightOf(st giode.Styles) properties.Height {
	if st.Height != 0 {
		return -1
	}
	return 0
}

// Element returns the area for layout.
func (a *Area) Element() giode.Element { return a.p }
