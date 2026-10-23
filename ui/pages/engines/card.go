package engines

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/ui/components/badge"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/logview"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// card is one engine tile.
type card struct {
	c       *core.Ctx
	rt      *manager.Runtime
	showLog bool

	install   *giode.ButtonElement
	cancel    *giode.ButtonElement
	use       *giode.ButtonElement
	logBtn    *giode.ButtonElement
	remove    *giode.ButtonElement
	configure *giode.ButtonElement
	site      *giode.ButtonElement
	log       *logview.View
}

func newCard(c *core.Ctx, rt *manager.Runtime, cfg *configDialog) *card {
	cd := &card{c: c, rt: rt}
	m := rt.Manifest()
	primary := giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(12, 8), BorderRadius: theme.RadiusSm}
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 13,
		Padding: giode.SymmetricInset(10, 8), BorderRadius: theme.RadiusSm}

	cd.install = giode.Button("Install", primary).Icon(icons.Icon(icons.Download, 15, "")).OnClick(func() {
		c.Manager.Install(rt, func(err error) {
			if err != nil {
				c.Error(err)
				cd.showLog = true
				return
			}
			c.Success(m.Name + " is installed")
		})
	})
	cd.cancel = giode.Button("Cancel", outline).Variant(giode.VariantOutlined).OnClick(func() { c.Manager.Cancel(rt) })
	cd.use = giode.Button("Use in Studio", primary).Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() {
		c.UseEngine(m.ID)
	})
	cd.logBtn = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Terminal, 15, "")).
		OnClick(func() { cd.showLog = !cd.showLog })
	cd.remove = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Trash, 15, "")).OnClick(func() {
		if m.Custom {
			c.Manager.RemoveCustom(m.ID)
			c.Info("Removed " + m.Name)
			return
		}
		c.Toasts.ShowAction(giode.ToastWarning, "Remove "+m.Name+" and its downloads?", "Remove", func() {
			if err := c.Manager.Uninstall(rt); err != nil {
				c.Error(err)
				return
			}
			c.Info(m.Name + " removed")
		})
	})
	cd.configure = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Sliders, 15, "")).
		OnClick(func() { cfg.show(rt) })
	cd.site = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.External, 15, "")).OnClick(func() {
		c.Error(paths.Open(m.Homepage))
	})
	cd.log = logview.New(func() []string { return rt.Log })
	return cd
}

func (cd *card) element() giode.Element {
	rt := cd.rt
	m := rt.Manifest()
	st := theme.Card()
	st.Width, st.Gap = 318, 14
	header := giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		badge.Engine(m, 44),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(m.Name, giode.Styles{Color: "theme.text", FontSize: 16, FontWeight: giode.FontWeightBold}),
			giode.Text(m.Vendor, theme.Caption),
		),
		statuspill.For(rt),
	)
	els := []giode.Element{header}
	if m.Description != "" {
		els = append(els, giode.Text(m.Description, giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 19, MaxLines: 4}))
	}
	var tags []giode.Element
	for _, t := range m.Tags {
		tags = append(tags, chip.Chip(t))
	}
	if len(tags) > 0 {
		els = append(els, giode.Row(giode.Styles{Width: -1, Gap: 6}, tags...))
	}
	if meta := metaLine(m); meta != "" {
		els = append(els, giode.Text(meta, theme.Caption))
	}
	if rt.State == manager.StateInstalling {
		els = append(els, progress(rt))
	}
	if rt.State == manager.StateUnavailable && m.Hint != "" {
		els = append(els, giode.Text(m.Hint, giode.Styles{Color: "theme.warning", FontSize: 12}))
	}
	if rt.Err != "" && rt.State != manager.StateInstalling {
		els = append(els, giode.Text(firstLine(rt.Err), giode.Styles{Color: "theme.danger", FontSize: 12, MaxLines: 3}))
	}
	if cd.showLog {
		els = append(els, cd.log.Element(180))
	}
	els = append(els, giode.Box(giode.Styles{Height: 2}), cd.actions())
	return giode.Stack(st, els...)
}

func (cd *card) actions() giode.Element {
	rt := cd.rt
	m := rt.Manifest()
	var left, right []giode.Element
	switch rt.State {
	case manager.StateNotInstalled:
		left = append(left, cd.install)
	case manager.StateInstalling:
		left = append(left, cd.cancel)
	case manager.StateReady, manager.StateWorking:
		left = append(left, cd.use)
	}
	if m.Kind == engines.KindPython || len(rt.Log) > 0 {
		right = append(right, cd.logBtn)
	}
	if m.Homepage != "" {
		right = append(right, cd.site)
	}
	if m.Custom {
		right = append(right, cd.configure)
	}
	installed := m.NeedsInstall() && (rt.State == manager.StateReady || rt.Err != "")
	if (installed || m.Custom) && !rt.Busy() {
		right = append(right, cd.remove)
	}
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap},
		append(append(left, giode.Box(giode.Styles{FlexGrow: 1})), right...)...)
}

func metaLine(m *engines.Manifest) string {
	parts := []string{}
	if m.Size != "" {
		parts = append(parts, m.Size)
	}
	if m.License != "" {
		parts = append(parts, m.License)
	}
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "  ·  "
		}
		out += p
	}
	return out
}

func progress(rt *manager.Runtime) giode.Element {
	bar := giode.Element(giode.Box(giode.Styles{Width: -1, Height: 6, BorderRadius: 99, Background: "theme.text/0.08"},
		giode.Box(giode.Styles{WidthPct: 35, Height: 6, BorderRadius: 99, BackgroundGradient: theme.Brand(90)})))
	if rt.Progress >= 0 {
		bar = giode.Progress(rt.Progress, giode.Styles{Width: -1, Height: 6, Background: "theme.text/0.08", Color: "theme.primary", BorderRadius: 99})
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8},
		bar,
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: "theme.primary"}),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(rt.Status, giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis})),
		),
	)
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
