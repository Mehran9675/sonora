// Package logdialog shows an engine's full log in a modal dialog.
package logdialog

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/logview"
	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Dialog is created once and registers itself as an overlay.
type Dialog struct {
	open  bool
	rt    *manager.Runtime
	view  *logview.View
	close *giode.ButtonElement
}

// New returns a closed dialog.
func New() *Dialog {
	d := &Dialog{}
	d.close = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(6), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 16, "")).OnClick(func() { d.open = false })
	overlay.Add(giode.Dialog(&d.open, d.content, giode.Styles{
		Background: "theme.surface", BorderRadius: theme.RadiusLg, Padding: giode.UniformInset(20),
		BorderWidth: 1, BorderColor: "theme.border", BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000080"},
	}))
	return d
}

// Show opens the log of rt; it keeps updating while open.
func (d *Dialog) Show(rt *manager.Runtime) {
	d.rt = rt
	d.view = logview.New(func() []string { return rt.Log })
	d.open = true
}

func (d *Dialog) content() giode.Element {
	if d.rt == nil {
		return giode.Box(giode.Styles{})
	}
	body := giode.Element(d.view.Element(420))
	if len(d.rt.Log) == 0 {
		body = giode.Box(giode.Styles{Width: -1, Height: 120, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Text("Nothing logged yet. Install or use the engine to see its output here.", theme.Subtle))
	}
	return giode.Stack(giode.Styles{Width: 720, Gap: 14},
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(d.rt.Manifest().Name+" · log", theme.Heading)),
			d.close,
		),
		body,
	)
}
