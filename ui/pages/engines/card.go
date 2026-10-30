package engines

import (
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/ui/components/badge"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/logdialog"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/progressbar"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
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

	// Choosing a model to download, for engines with a model choice:
	// choosing shows the chooser in place of the card's details.
	choosing bool
	choice   string
	models   *picker.Picker
	model    *picker.Trigger
	download *giode.ButtonElement
	back     *giode.ButtonElement
	addModel *giode.ButtonElement
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

	cd.install = giode.Button("", primary).Icon(icons.Icon(icons.Download, 15, "")).OnClick(func() {
		if m.ModelOption() != nil {
			cd.choose()
			return
		}
		c.Install(rt, func(err error) {
			if err != nil {
				c.Error(err)
				logs.Show(rt)
				return
			}
			c.Success(tr.T("{engine} is installed", "engine", m.Name))
		})
	})
	tr.Bind(func(s string) { cd.install.Label(s) }, "Install")
	if o := m.ModelOption(); o != nil {
		cd.models = picker.New(tr.Msg("Choose a model"), func() []picker.Item {
			have := c.Manager.Downloaded(m.ID)
			var items []picker.Item
			for _, ch := range o.Choices {
				it := picker.Item{ID: ch.Value, Label: tr.T(ch.Label)}
				if have[ch.Value] {
					it.Detail = tr.T("Downloaded")
				}
				items = append(items, it)
			}
			return items
		}, func() string { return cd.choice }, func(it picker.Item) { cd.choice = it.ID })
		cd.model = picker.NewTrigger(tr.Msg("Model"), func() (string, string) {
			detail := ""
			if c.Manager.Downloaded(m.ID)[cd.choice] {
				detail = tr.T("Already downloaded")
			}
			for _, ch := range o.Choices {
				if ch.Value == cd.choice {
					return tr.T(ch.Label), detail
				}
			}
			return cd.choice, detail
		}, nil, cd.models.Open)
		cd.download = giode.Button("", primary).Icon(icons.Icon(icons.Download, 15, "")).OnClick(func() {
			cd.choosing = false
			c.InstallModel(rt, cd.choice, func(err error) {
				if err != nil {
					c.Error(err)
					logs.Show(rt)
					return
				}
				c.Success(tr.T("{engine} is ready with {model}", "engine", m.Name, "model", choiceLabel(o, cd.choice)))
			})
		})
		cd.back = giode.Button("", outline).Variant(giode.VariantOutlined).OnClick(func() { cd.choosing = false })
		cd.addModel = giode.Button("", giode.Styles{Color: "theme.primary", Background: "theme.primary", FontSize: 12,
			Padding: giode.SymmetricInset(8, 5), BorderRadius: theme.RadiusSm}).Variant(giode.VariantGhost).
			Icon(icons.Icon(icons.Plus, 13, "")).OnClick(cd.choose)
		tr.Bind(func(s string) { cd.download.Label(s) }, "Download")
		tr.Bind(func(s string) { cd.back.Label(s) }, "Cancel")
		tr.Bind(func(s string) { cd.addModel.Label(s) }, "Add model")
	}
	cd.cancel = giode.Button("", outline).Variant(giode.VariantOutlined).OnClick(func() { c.Manager.Cancel(rt) })
	tr.Bind(func(s string) { cd.cancel.Label(s) }, "Cancel")
	cd.use = giode.Button("", primary).Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() {
		c.UseEngine(m.ID)
	})
	use := tr.Msg("Use in Studio")
	if m.Transcribes() {
		cd.use = giode.Button("", primary).Icon(icons.Icon(icons.Captions, 15, "")).OnClick(func() {
			c.UseTranscriber(m.ID)
		})
		use = tr.Msg("Transcribe")
	}
	tr.Bind(func(s string) { cd.use.Label(s) }, use)
	cd.logBtn = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Terminal, 15, "")).
		OnClick(func() { logs.Show(rt) })
	cd.remove = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Trash, 15, "")).OnClick(func() {
		if m.Custom {
			c.Manager.RemoveCustom(m.ID)
			c.Info(tr.T("Removed {engine}", "engine", m.Name))
			return
		}
		c.Toasts.ShowAction(giode.ToastWarning, tr.T("Remove {engine} and its downloads?", "engine", m.Name), tr.T("Remove"), func() {
			if err := c.Manager.Uninstall(rt); err != nil {
				c.Error(err)
				return
			}
			c.Info(tr.T("{engine} removed", "engine", m.Name))
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
		c.Info(tr.T("{engine} unloaded from memory", "engine", m.Name))
	})
	return cd
}

// choose opens the model chooser, starting from the model the engine is
// set to use.
func (cd *card) choose() {
	m := cd.rt.Manifest()
	o := m.ModelOption()
	cd.choice = cd.c.Manager.Values(m).String(o.ID)
	// Adding a model: start from the first one that isn't here yet.
	if have := cd.c.Manager.Downloaded(m.ID); have[cd.choice] {
		for _, ch := range o.Choices {
			if !have[ch.Value] {
				cd.choice = ch.Value
				break
			}
		}
	}
	cd.choosing = true
}

// chooser asks which model to download.
func (cd *card) chooser() giode.Element {
	m := cd.rt.Manifest()
	msg := tr.T("You can add others later, or switch models in the Studio.")
	if m.Transcribes() {
		msg = tr.T("You can add others later, or switch models on the Transcribe page.")
	}
	cd.download.Disabled(cd.c.Manager.Downloaded(m.ID)[cd.choice])
	return giode.Stack(giode.Styles{Width: -1, Gap: 10},
		giode.Text(tr.T("Choose a model to download"), giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightBold}),
		giode.Text(msg, giode.Styles{Color: "theme.secondary", FontSize: 12, LineHeight: 17, MaxLines: 3, TextOverflow: giode.TextOverflowEllipsis}),
		cd.model.Element(),
	)
}

// downloadedModels names the engine's downloaded models, or "" when it
// has no model choice.
func (cd *card) downloadedModels() string {
	o := cd.rt.Manifest().ModelOption()
	if o == nil {
		return ""
	}
	have := cd.c.Manager.Downloaded(cd.rt.Manifest().ID)
	var names []string
	for _, ch := range o.Choices {
		if have[ch.Value] {
			names = append(names, shortLabel(tr.T(ch.Label)))
		}
	}
	if len(names) == 0 {
		return ""
	}
	return tr.T("Downloaded: {models}", "models", strings.Join(names, ", "))
}

// allDownloaded reports whether every model of the engine is here.
func (cd *card) allDownloaded() bool {
	o := cd.rt.Manifest().ModelOption()
	return o != nil && len(cd.c.Manager.Downloaded(cd.rt.Manifest().ID)) >= len(o.Choices)
}

// choiceLabel returns the translated label of a model choice.
func choiceLabel(o *engines.Option, value string) string {
	for _, ch := range o.Choices {
		if ch.Value == value {
			return shortLabel(tr.T(ch.Label))
		}
	}
	return value
}

// shortLabel drops the size from a model label: "Tiny · 39M · 110 MB"
// reads "Tiny".
func shortLabel(label string) string {
	name, _, _ := strings.Cut(label, " · ")
	return name
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
				giode.Text(core.Parts(m.Vendor), giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
			),
		),
	)
	var tags []giode.Element
	for _, t := range m.Tags {
		tags = append(tags, chip.Chip(tr.T(t)))
	}
	if cd.choosing && cd.rt.State != manager.StateInstalling {
		// The chooser takes the place of the description, tags, details
		// and status.
		return giode.Stack(st,
			header,
			fixed(descHeight+22+metaHeight+statusHeight+3*12, cd.chooser()),
			giode.Box(giode.Styles{FlexGrow: 1}),
			giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap},
				cd.download, cd.back),
		)
	}
	return giode.Stack(st,
		header,
		fixed(descHeight, giode.Text(tr.T(m.Description), giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 19, MaxLines: 3, TextOverflow: giode.TextOverflowEllipsis})),
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
			giode.Text(tr.T(rt.Status), giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
		)
	case rt.Err != "":
		return giode.Text(tr.T(firstLine(rt.Err)), giode.Styles{Color: "theme.danger", FontSize: 12, LineHeight: 16, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
	case rt.State == manager.StateUnavailable && m.Hint != "":
		return giode.Text(tr.T(m.Hint), giode.Styles{Color: "theme.warning", FontSize: 12, LineHeight: 16, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
	case rt.Engine.Loaded():
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 99, Background: "theme.success"}),
			giode.Text(tr.T("Model loaded in memory"), giode.Styles{Color: "theme.secondary", FontSize: 12}))
	case cd.downloadedModels() != "":
		text := giode.Text(cd.downloadedModels(), giode.Styles{Color: "theme.secondary", FontSize: 12, LineHeight: 16,
			MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
		if rt.State != manager.StateReady || cd.allDownloaded() {
			return text
		}
		// Download another model.
		return giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, text), cd.addModel)
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
