package studio

import (
	"fmt"
	"time"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/badge"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/progressbar"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"

	"github.com/mehran9675/sonora/ui/tr"
)

// voicePanel is the right-hand column: engine, voice, settings and the
// Generate button.
type voicePanel struct {
	p        *Page
	scroll   *scroll.Scroll
	engine   *picker.Trigger
	install  *giode.ButtonElement
	manage   *giode.ButtonElement
	generate *giode.ButtonElement
	eject    *giode.ButtonElement
	rt       *manager.Runtime // engine shown this frame
}

func newVoicePanel(p *Page) *voicePanel {
	v := &voicePanel{p: p}
	c := p.c
	v.engine = picker.NewTrigger(tr.Msg("Engine"), func() (string, string) {
		label, _ := statuspill.Describe(v.rt)
		return v.rt.Manifest().Name, joinNonEmpty(core.Parts(v.rt.Manifest().Vendor), tr.T(label))
	}, func() giode.Element { return badge.Engine(v.rt.Manifest(), 32) }, p.engines.Open)
	v.install = giode.Button("", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}).
		Icon(icons.Icon(icons.Download, 15, "")).OnClick(func() {
		rt := v.rt
		done := func(err error) {
			if err == nil {
				c.Success(tr.T("{name} is ready", "name", rt.Manifest().Name))
				c.Manager.LoadVoices(rt, false)
			} else {
				c.Error(err)
			}
		}
		// Engines with a model choice download the model the panel's
		// selector shows.
		if o := rt.Manifest().ModelOption(); o != nil {
			c.InstallModel(rt, c.Manager.Values(rt.Manifest()).String(o.ID), done)
			return
		}
		c.Install(rt, done)
	})
	tr.Bind(func(s string) { v.install.Label(s) }, "Install")
	v.manage = giode.Button("", giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border",
		FontSize: 13, Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}).
		Variant(giode.VariantOutlined).OnClick(func() { c.Go(core.PageEngines) })
	tr.Bind(func(s string) { v.manage.Label(s) }, "Open Engines")
	// footer sets Generate's text every frame.
	v.generate = giode.Button("", giode.Styles{}).OnClick(p.toggleGenerate)
	v.eject = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 12,
		Padding: giode.SymmetricInset(8, 4), BorderRadius: theme.RadiusSm}).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Eject, 13, "")).OnClick(func() {
		rt := v.rt
		if err := c.Manager.Eject(rt); err != nil {
			c.Error(err)
			return
		}
		c.Info(tr.T("{name} unloaded — memory freed", "name", rt.Manifest().Name))
	})
	tr.Bind(func(s string) { v.eject.Label(s) }, "Eject")
	v.scroll = giode.Scroll(v.content, theme.Scrolling(5))
	return v
}

func (v *voicePanel) element(rt *manager.Runtime) giode.Element {
	v.rt = rt
	card := theme.Card()
	card.Width, card.Height, card.Gap = 340, -1, 0
	card.Padding = giode.UniformInset(0)
	return giode.Stack(card,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, v.scroll),
		giode.Divider(giode.Styles{Color: "theme.border"}),
		v.footer(),
	)
}

// compactElement is the same panel for the phone layout: full width,
// no inner scrolling (the whole page scrolls).
func (v *voicePanel) compactElement(rt *manager.Runtime) giode.Element {
	v.rt = rt
	card := theme.Card()
	card.Width, card.Gap = -1, 0
	card.Padding = giode.UniformInset(0)
	return giode.Stack(card, v.content(), giode.Divider(giode.Styles{Color: "theme.border"}), v.footer())
}

func (v *voicePanel) content() giode.Element {
	rt := v.rt
	m := rt.Manifest()
	f := v.p.form(rt)
	els := []giode.Element{section(tr.T("Engine"), v.engineSection())}
	if callout := v.callout(); callout != nil {
		els = append(els, callout)
	}
	// The script card shows the hint on wide screens; a phone shows it
	// here.
	if giode.Compact() && m.Hint != "" && rt.State != manager.StateUnavailable {
		els = append(els, giode.Text(tr.T(m.Hint), giode.Styles{Color: "theme.muted", FontSize: 12, LineHeight: 17}))
	}
	ready := rt.State == manager.StateReady || rt.State == manager.StateWorking
	if ready {
		els = append(els, section(tr.T("Voice"), voiceChoice(f)))
		if m.Speed {
			els = append(els, giode.Stack(giode.Styles{Width: -1, Gap: 8},
				giode.Row(giode.Styles{Width: -1, Wrap: giode.WrapNoWrap},
					giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T("Speed"), theme.Overline)),
					giode.Text(fmt.Sprintf("%.2f×", f.speedValue()), giode.Styles{Color: "theme.text", FontSize: 12, FontWeight: giode.FontWeightBold})),
				syncSlider(f),
			))
		}
		for _, fl := range f.fields {
			els = append(els, fl.Element())
		}
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 22, Padding: giode.UniformInset(20)}, els...)
}

// syncSlider keeps the speed slider in step with settings changed
// elsewhere (reusing a library entry), except while it's dragged.
func syncSlider(f *engineForm) giode.Element {
	if !f.speed.Dragging() {
		f.speed.SetValue(speedToSlider(f.speedValue()))
	}
	return giode.Box(giode.Styles{Width: -1}, f.speed)
}

// callout explains engines that can't generate yet.
func (v *voicePanel) callout() giode.Element {
	rt := v.rt
	m := rt.Manifest()
	var body []giode.Element
	switch rt.State {
	case manager.StateNotInstalled:
		msg := tr.T("{name} needs a one-time setup. Sonora downloads everything it needs automatically.", "name", m.Name)
		body = append(body, giode.Text(msg, giode.Styles{Color: "theme.text", FontSize: 13, LineHeight: 19}))
		if m.Size != "" {
			body = append(body, chip.Chip(m.Size))
		}
		if fl := v.p.form(rt).modelField(); fl != nil {
			body = append(body, fl.Element())
		}
		if rt.Err != "" {
			body = append(body, giode.Text(tr.T(firstLine(rt.Err)), giode.Styles{Color: "theme.danger", FontSize: 12, MaxLines: 3}))
		}
		body = append(body, giode.Row(giode.Styles{Gap: 8}, v.install, v.manage))
	case manager.StateInstalling:
		body = append(body,
			giode.Text(tr.T("Installing {name}…", "name", m.Name), giode.Styles{Color: "theme.text", FontSize: 13, FontWeight: giode.FontWeightBold}),
			progressbar.New(rt.Progress),
			giode.Text(tr.T(rt.Status), giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
		)
	case manager.StateUnavailable:
		body = append(body, giode.Text(tr.T(m.Hint), giode.Styles{Color: "theme.text", FontSize: 13, LineHeight: 19}))
	default:
		return nil
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(14),
		Background: "theme.primary/0.08", BorderRadius: theme.RadiusMd, BorderWidth: 1, BorderColor: "theme.primary/0.25"}, body...)
}

func (v *voicePanel) footer() giode.Element {
	rt := v.rt
	c := v.p.c
	busy := c.Manager.Job != nil
	ready := rt.State == manager.StateReady || rt.State == manager.StateWorking
	st := giode.Styles{Width: -1, Justify: giode.JustifyCenter, BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 15, BorderRadius: theme.RadiusMd,
		Padding: giode.SymmetricInset(16, 13), BoxShadow: giode.BoxShadow{Y: 8, Blur: 24, Color: "#8b5cf650"}}
	if busy {
		st.BackgroundGradient = nil
		st.Background = "theme.danger"
		st.BoxShadow = giode.BoxShadow{}
		v.generate.Label(tr.T("Stop")).Icon(icons.Solid(icons.Stop, 14, ""))
	} else {
		v.generate.Label(tr.T("Generate speech")).Icon(icons.Icon(icons.Sparkles, 18, ""))
	}
	v.generate.Styles(st).Disabled(!busy && (!ready || v.p.editor.text() == ""))

	var status giode.Element
	switch {
	case busy:
		job := c.Manager.Job
		line := giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: "theme.primary"}),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T(job.Runtime.Status),
				giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis})),
			giode.Text(core.Clock(time.Since(job.Started)), giode.Styles{Color: "theme.muted", FontSize: 12, FontFamily: theme.Mono}),
		)
		if job.Runtime.Progress >= 0 {
			status = giode.Stack(giode.Styles{Width: -1, Gap: 8}, progressbar.New(job.Runtime.Progress), line)
		} else {
			status = line
		}
	case rt.Err != "" && rt.State == manager.StateReady:
		status = giode.Text(tr.T(firstLine(rt.Err)), giode.Styles{Color: "theme.danger", FontSize: 12, MaxLines: 3})
	case !ready:
		status = giode.Text(tr.T("Install the engine to start generating."), theme.Caption)
	default:
		status = giode.Text(tr.T("Audio is saved to your Library automatically."), theme.Caption)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.UniformInset(20)},
		giode.Box(giode.Styles{Width: -1}, v.generate),
		status,
	)
}

// engineSection is the engine picker, plus an Eject control while the
// engine's model is held in memory.
func (v *voicePanel) engineSection() giode.Element {
	if !v.rt.Engine.Loaded() || v.rt.Busy() {
		return v.engine.Element()
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6},
		v.engine.Element(),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 4}},
			giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 99, Background: "theme.success"}),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T("Model loaded in memory"), theme.Caption)),
			v.eject,
		),
	)
}

func section(title string, body giode.Element) giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, giode.Text(title, theme.Overline), body)
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
