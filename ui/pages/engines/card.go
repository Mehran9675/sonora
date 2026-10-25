package engines

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/ui/components/badge"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/logdialog"
	"github.com/mehran9675/sonora/ui/components/progressbar"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// card is one engine tile.
type card struct {
	c  *core.Ctx
	rt *manager.Runtime

	install   *giode.ButtonElement
	cancel    *giode.ButtonElement
	use       *giode.ButtonElement
	logBtn    *giode.ButtonElement
	remove    *giode.ButtonElement
	configure *giode.ButtonElement
	site      *giode.ButtonElement
	eject     *giode.ButtonElement
}

func newCard(c *core.Ctx, rt *manager.Runtime, cfg *configDialog, logs *logdialog.Dialog) *card {
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
				logs.Show(rt)
				return
			}
			c.Success(m.Name + " is installed")
		})
	})
	cd.cancel = giode.Button("Cancel", outline).Variant(giode.VariantOutlined).OnClick(func() { c.Manager.Cancel(rt) })
	cd.use = giode.Button("Use in Studio", primary).Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() {
		c.UseEngine(m.ID)
	})
	if m.Transcribes() {
		cd.use = giode.Button("Transcribe", primary).Icon(icons.Icon(icons.Captions, 15, "")).OnClick(func() {
			c.UseTranscriber(m.ID)
		})
	}
	cd.logBtn = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Terminal, 15, "")).
		OnClick(func() { logs.Show(rt) })
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
	cd.eject = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Eject, 15, "")).OnClick(func() {
		if err := c.Manager.Eject(rt); err != nil {
			c.Error(err)
			return
		}
		c.Info(m.Name + " unloaded from memory")
	})
	return cd
}

// Fixed heights for each part of a card, so every card is the same
// size whatever its state.
const (
	cardHeight   = 356
	descHeight   = 64 // three lines
	metaHeight   = 32 // two lines
	statusHeight = 38
)

func (cd *card) element() giode.Element {
	rt := cd.rt
	m := rt.Manifest()
	st := theme.Card()
	st.Width, st.Height, st.Gap, st.Overflow = 318, cardHeight, 12, giode.OverflowHidden
	if giode.Compact() {
		st.Width = -1 // one card per row on a phone
	}
	header := giode.Row(giode.Styles{Width: -1, Height: 48, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		badge.Engine(m, 44),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 4},
			giode.Text(m.Name, giode.Styles{Color: "theme.text", FontSize: 16, FontWeight: giode.FontWeightBold, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				statuspill.For(rt),
				giode.Text(m.Vendor, giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
			),
		),
	)
	var tags []giode.Element
	for _, t := range m.Tags {
		tags = append(tags, chip.Chip(t))
	}
	return giode.Stack(st,
		header,
		fixed(descHeight, giode.Text(m.Description, giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 19, MaxLines: 3, TextOverflow: giode.TextOverflowEllipsis})),
		giode.Row(giode.Styles{Width: -1, Height: 22, Gap: 6, Wrap: giode.WrapNoWrap, Overflow: giode.OverflowHidden}, tags...),
		fixed(metaHeight, giode.Text(metaLine(m), giode.Styles{Color: "theme.muted", FontSize: 12, LineHeight: 16, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})),
		fixed(statusHeight, cd.status()),
		giode.Box(giode.Styles{FlexGrow: 1}),
		cd.actions(),
	)
}

// status is the one-line state area: install progress, a problem, or
// whether the model is in memory.
func (cd *card) status() giode.Element {
	rt := cd.rt
	m := rt.Manifest()
	switch {
	case rt.State == manager.StateInstalling:
		return giode.Stack(giode.Styles{Width: -1, Gap: 8},
			progressbar.New(rt.Progress),
			giode.Text(rt.Status, giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
		)
	case rt.Err != "":
		return giode.Text(firstLine(rt.Err), giode.Styles{Color: "theme.danger", FontSize: 12, LineHeight: 16, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
	case rt.State == manager.StateUnavailable && m.Hint != "":
		return giode.Text(m.Hint, giode.Styles{Color: "theme.warning", FontSize: 12, LineHeight: 16, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
	case rt.Engine.Loaded():
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 99, Background: "theme.success"}),
			giode.Text("Model loaded in memory", giode.Styles{Color: "theme.secondary", FontSize: 12}))
	}
	return giode.Box(giode.Styles{})
}

// fixed reserves height dp for el and clips anything taller.
func fixed(height int, el giode.Element) giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: properties.Height(height), Overflow: giode.OverflowHidden}, el)
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
	if rt.Engine.Loaded() && !rt.Busy() {
		right = append(right, cd.eject)
	}
	if m.NeedsInstall() || len(rt.Log) > 0 {
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

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
