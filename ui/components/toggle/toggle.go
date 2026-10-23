// Package toggle is an on/off switch.
package toggle

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"
)

// Toggle is a stateful switch with a label. Create it once.
type Toggle struct {
	p  *giode.PressableElement
	on bool
	fn func(bool)
}

// New returns a switch labelled label.
func New(label string, on bool, onChange func(bool)) *Toggle {
	t := &Toggle{on: on, fn: onChange}
	t.p = giode.Pressable(func(s giode.PressState) giode.Element {
		track := "theme.text/0.14"
		if t.on {
			track = "theme.primary"
		} else if s.Hovered {
			track = "theme.text/0.2"
		}
		knobLeft := 3
		if t.on {
			knobLeft = 19
		}
		sw := giode.Box(giode.Styles{Width: 38, Height: 22, BorderRadius: 99, Background: track},
			giode.Box(giode.Styles{
				Position: giode.PositionAbsolute, Top: 3, Left: properties.Left(knobLeft),
				Width: 16, Height: 16, BorderRadius: 99, Background: "#ffffff",
				BoxShadow: giode.BoxShadow{Y: 1, Blur: 3, Color: "#00000040"},
			}),
		)
		return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			sw, giode.Text(label, giode.Styles{Color: "theme.text", FontSize: 13}))
	}).Label(label).OnClick(func(giode.ClickEvent) {
		t.on = !t.on
		if t.fn != nil {
			t.fn(t.on)
		}
	})
	return t
}

// Set changes the state without calling onChange.
func (t *Toggle) Set(on bool) { t.on = on }

// On reports the state.
func (t *Toggle) On() bool { return t.on }

// Element returns the switch for layout.
func (t *Toggle) Element() giode.Element { return t.p }
