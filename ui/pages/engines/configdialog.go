package engines

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/optionfield"
	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/theme"
)

// configDialog edits a custom engine's setup options in place.
type configDialog struct {
	c      *core.Ctx
	open   bool
	rt     *manager.Runtime
	fields []*optionfield.Field
	done   *giode.ButtonElement
}

func newConfigDialog(c *core.Ctx) *configDialog {
	d := &configDialog{c: c}
	d.done = giode.Button("Done", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(18, 9), BorderRadius: theme.RadiusSm + 2}).OnClick(func() {
		d.open = false
		if d.rt != nil {
			c.Manager.LoadVoices(d.rt, true)
		}
	})
	overlay.Add(giode.Dialog(&d.open, d.content, dialogStyle()))
	return d
}

func (d *configDialog) show(rt *manager.Runtime) {
	d.rt = rt
	m := rt.Manifest()
	st := optionfield.Store{
		Get: func(id string) any { return d.c.Settings.Values(m.ID)[id] },
		Set: func(id string, v any) { d.c.Settings.SetValue(m.ID, id, v) },
	}
	d.fields = nil
	for _, o := range m.Options {
		if o.Scope == engines.ScopeSetup {
			d.fields = append(d.fields, optionfield.New(o, st))
		}
	}
	d.open = true
}

func (d *configDialog) content() giode.Element {
	if d.rt == nil {
		return giode.Box(giode.Styles{})
	}
	els := []giode.Element{giode.Text(d.rt.Manifest().Name, giode.Styles{Color: "theme.text", FontSize: 18, FontWeight: giode.FontWeightBold})}
	for _, f := range d.fields {
		els = append(els, f.Element())
	}
	els = append(els, giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, d.done))
	return giode.Stack(giode.Styles{Width: -1, MaxWidth: 520, Gap: 16}, els...)
}
