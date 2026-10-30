// Package statuspill shows an engine's state as a colored dot and
// label.
package statuspill

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/tr"
)

// For returns the pill for rt.
func For(rt *manager.Runtime) giode.Element {
	label, color := Describe(rt)
	return Pill(label, color)
}

// Describe returns the translated label and theme color for rt's state.
func Describe(rt *manager.Runtime) (string, string) {
	switch rt.State {
	case manager.StateReady:
		return tr.T("Ready"), "theme.success"
	case manager.StateWorking:
		return tr.T("Working"), "theme.accent"
	case manager.StateInstalling:
		return tr.T("Installing"), "theme.primary"
	case manager.StateUnavailable:
		return tr.T("Unavailable"), "theme.warning"
	}
	if rt.Err != "" {
		return tr.T("Install failed"), "theme.danger"
	}
	return tr.T("Not installed"), "theme.muted"
}

// Pill is a dot plus label in color.
func Pill(label, color string) giode.Element {
	return giode.Row(giode.Styles{
		Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Background: color + "/0.12", BorderRadius: 99, Padding: giode.SymmetricInset(10, 4),
	},
		giode.Box(giode.Styles{Width: 7, Height: 7, BorderRadius: 99, Background: color}),
		giode.Text(label, giode.Styles{Color: color, FontSize: 12, FontWeight: giode.FontWeightBold}),
	)
}
