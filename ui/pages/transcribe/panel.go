package transcribe

import (
	"os"
	"path/filepath"
	"time"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/audio/decode"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/badge"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/optionfield"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/progressbar"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/components/voicerecorder"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"

	"github.com/mehran9675/sonora/ui/tr"
)

// panel is the right-hand column: engine, audio, language, settings
// and the Transcribe button.
type panel struct {
	p        *Page
	scroll   *scroll.Scroll
	engines  *picker.Picker
	engine   *picker.Trigger
	install  *giode.ButtonElement
	manage   *giode.ButtonElement
	pick     *giode.ButtonElement
	clear    *giode.ButtonElement
	play     *giode.PressableElement
	recorder *voicerecorder.Recorder
	run      *giode.ButtonElement
	eject    *giode.ButtonElement
	forms    map[string]*form
	rt       *manager.Runtime // engine shown this frame
}

// form holds one engine's language, translation and option fields.
type form struct {
	rt        *manager.Runtime
	language  *optionfield.Field
	translate *optionfield.Field
	fields    []*optionfield.Field
}

func newPanel(p *Page) *panel {
	v := &panel{p: p, forms: map[string]*form{}}
	c := p.c
	v.engines = picker.New(tr.Msg("Choose a transcription engine"), func() []picker.Item {
		var ready, other []picker.Item
		for _, rt := range c.Manager.EnginesFor(engines.TaskTranscribe) {
			m := rt.Manifest()
			label, _ := statuspill.Describe(rt)
			it := picker.Item{ID: m.ID, Label: m.Name, Detail: joinNonEmpty(m.Vendor, label)}
			if rt.State == manager.StateReady || rt.State == manager.StateWorking {
				it.Group = tr.T("Ready to use")
				ready = append(ready, it)
			} else {
				it.Group = tr.T("Available to install")
				other = append(other, it)
			}
		}
		return append(ready, other...)
	}, func() string { return c.Settings.Get().TranscribeEngine }, func(it picker.Item) { p.selectEngine(it.ID) })
	v.engine = picker.NewTrigger(tr.Msg("Engine"), func() (string, string) {
		label, _ := statuspill.Describe(v.rt)
		return v.rt.Manifest().Name, joinNonEmpty(v.rt.Manifest().Vendor, label)
	}, func() giode.Element { return badge.Engine(v.rt.Manifest(), 32) }, v.engines.Open)

	primary := giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border",
		FontSize: 13, Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}
	v.install = giode.Button("", primary).Icon(icons.Icon(icons.Download, 15, "")).OnClick(func() {
		rt := v.rt
		done := func(err error) {
			if err != nil {
				c.Error(err)
				return
			}
			c.Success(tr.T("{name} is ready", "name", rt.Manifest().Name))
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
	v.manage = giode.Button("", outline).Variant(giode.VariantOutlined).
		OnClick(func() { c.ShowEngines(engines.TaskTranscribe) })
	tr.Bind(func(s string) { v.manage.Label(s) }, "Open Engines")
	v.pick = giode.Button("", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Upload, 15, "")).OnClick(v.pickFile)
	tr.Bind(func(s string) { v.pick.Label(s) }, "Choose file…")
	v.clear = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(5), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 14, "")).OnClick(p.dropSource)
	v.recorder = voicerecorder.NewWith(c, voicerecorder.Options{
		Min: time.Second, Name: "speech",
		Hint: tr.Msg("Speak normally; press Stop when you're done."), Ready: tr.Msg("Press Stop when you're done."),
	}, func(path string) {
		p.setSource(path, tr.T("Recording {time}", "time", time.Now().Format("Jan 2, 15:04")), false)
	})
	v.play = giode.Pressable(func(s giode.PressState) giode.Element {
		icon := icons.Solid(icons.Play, 12, "theme.primary")
		if c.IsPlaying(p.source) {
			icon = icons.Solid(icons.Pause, 12, "theme.primary")
		}
		bg := "theme.primary/0.14"
		if s.Hovered {
			bg = "theme.primary/0.26"
		}
		return giode.Box(giode.Styles{Width: 28, Height: 28, BorderRadius: 99, Background: bg,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icon)
	}).OnClick(func(giode.ClickEvent) {
		if filepath.Ext(p.source) != ".wav" {
			c.Info(tr.T("Only WAV files play here; the transcript will play once it's done"))
			return
		}
		c.PlayPause(p.source, p.sourceName, tr.T("Audio to transcribe"))
	})
	tr.Bind(func(s string) { v.play.Label(s) }, "Play audio")
	v.run = giode.Button("", giode.Styles{}).OnClick(p.start)
	v.eject = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 12,
		Padding: giode.SymmetricInset(8, 4), BorderRadius: theme.RadiusSm}).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Eject, 13, "")).OnClick(func() {
		if err := c.Manager.Eject(v.rt); err != nil {
			c.Error(err)
			return
		}
		c.Info(tr.T("{name} unloaded — memory freed", "name", v.rt.Manifest().Name))
	})
	tr.Bind(func(s string) { v.eject.Label(s) }, "Eject")
	v.scroll = giode.Scroll(v.content, theme.Scrolling(5))
	return v
}

// pickFile asks for an audio (or video) file. Desktop pickers return
// the file itself; Android's hands over a stream, copied first.
func (v *panel) pickFile() {
	c := v.p.c
	var path, name string
	var temp bool
	c.Background(func() error {
		f, err := c.PickFile(append([]string{"audio/*", "video/*"}, decode.Extensions...), tr.Msg("Record instead."))
		if err != nil {
			return err
		}
		defer f.Close()
		// Desktop pickers hand over the file itself: read it in place.
		if file, ok := f.ReadCloser.(*os.File); ok && filepath.IsAbs(file.Name()) {
			path, name = file.Name(), f.Name
			return nil
		}
		name = f.Name
		if name == "" {
			name = tr.T("Audio from {time}", "time", time.Now().Format("Jan 2, 15:04"))
		}
		path, err = core.SavePicked(f, core.PickedDir(), "audio")
		temp = true
		return err
	}, func(err error) {
		if err != nil {
			c.Error(err)
			return
		}
		v.p.setSource(path, name, temp)
	})
}

// form returns the cached fields for rt.
func (v *panel) form(rt *manager.Runtime) *form {
	m := rt.Manifest()
	f := v.forms[m.ID]
	if f != nil && f.rt == rt {
		return f
	}
	c := v.p.c
	f = &form{rt: rt}
	global := optionfield.Store{
		Get: func(id string) any {
			s := c.Settings.Get()
			if id == "translate" {
				return s.Translate
			}
			return language(m, s.Language)
		},
		Set: func(id string, val any) {
			c.Settings.Update(func(s *settings.Settings) {
				switch id {
				case "translate":
					s.Translate, _ = val.(bool)
				default:
					s.Language, _ = val.(string)
				}
			})
		},
	}
	var choices []engines.Choice
	for _, l := range engines.Languages {
		if language(m, l.Value) == l.Value {
			choices = append(choices, l)
		}
	}
	if len(choices) > 1 {
		f.language = optionfield.New(engines.Option{ID: "language", Label: tr.Msg("Spoken language"), Type: engines.OptionSelect,
			Choices: choices, Help: tr.Msg("Detecting works well; naming the language helps short or noisy recordings.")}, global)
	}
	if m.Translate {
		f.translate = optionfield.New(engines.Option{ID: "translate", Label: tr.Msg("Translate into English"), Type: engines.OptionBool,
			Help: tr.Msg("Write the text in English, whatever language is spoken.")}, global)
	}
	store := c.OptionStore(m, nil)
	for _, o := range m.Options {
		if o.Scope != engines.ScopeSetup {
			f.fields = append(f.fields, optionfield.New(o, store))
		}
	}
	v.forms[m.ID] = f
	return f
}

func (v *panel) element(rt *manager.Runtime) giode.Element {
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

// compactElement is the panel for the phone layout, without inner
// scrolling.
func (v *panel) compactElement(rt *manager.Runtime) giode.Element {
	v.rt = rt
	card := theme.Card()
	card.Width, card.Gap = -1, 0
	card.Padding = giode.UniformInset(0)
	return giode.Stack(card, v.content(), giode.Divider(giode.Styles{Color: "theme.border"}), v.footer())
}

func (v *panel) content() giode.Element {
	rt := v.rt
	els := []giode.Element{section(tr.T("Engine"), v.engineSection())}
	if callout := v.callout(); callout != nil {
		els = append(els, callout)
	}
	els = append(els, section(tr.T("Audio"), v.sourceSection()))
	if rt.State == manager.StateReady || rt.State == manager.StateWorking {
		f := v.form(rt)
		if f.language != nil {
			els = append(els, f.language.Element())
		}
		if f.translate != nil {
			els = append(els, f.translate.Element())
		}
		for _, fl := range f.fields {
			els = append(els, fl.Element())
		}
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 22, Padding: giode.UniformInset(20)}, els...)
}

// sourceSection picks or records the audio, or shows the chosen file.
func (v *panel) sourceSection() giode.Element {
	p := v.p
	switch {
	case v.recorder.Recording():
		return v.recorder.Element()
	case p.source == "":
		return giode.Stack(giode.Styles{Width: -1, Gap: 10},
			giode.Text(tr.T("Pick a recording or video, or record right here."), giode.Styles{Color: "theme.muted", FontSize: 12, LineHeight: 17}),
			giode.Row(giode.Styles{Width: -1, Gap: 8}, v.pick, v.recorder.Button()),
		)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(10, 8), Background: "theme.primary/0.1", BorderRadius: theme.RadiusSm + 2},
		v.play,
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(p.sourceName,
			giode.Styles{Color: "theme.text", FontSize: 13, TextOverflow: giode.TextOverflowEllipsis})),
		v.clear,
	)
}

// callout explains engines that can't transcribe yet.
func (v *panel) callout() giode.Element {
	rt := v.rt
	m := rt.Manifest()
	var body []giode.Element
	switch rt.State {
	case manager.StateNotInstalled:
		body = append(body, giode.Text(tr.T("{name} needs a one-time setup. Sonora downloads everything it needs automatically.", "name", m.Name),
			giode.Styles{Color: "theme.text", FontSize: 13, LineHeight: 19}))
		if m.Size != "" {
			body = append(body, chip.Chip(m.Size))
		}
		if o := m.ModelOption(); o != nil {
			for _, fl := range v.form(rt).fields {
				if fl.ID() == o.ID {
					body = append(body, fl.Element())
				}
			}
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

func (v *panel) footer() giode.Element {
	rt := v.rt
	c := v.p.c
	job := c.Manager.TranscribeJob
	ready := rt.State == manager.StateReady || rt.State == manager.StateWorking
	st := giode.Styles{Width: -1, Justify: giode.JustifyCenter, BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 15,
		BorderRadius: theme.RadiusMd, Padding: giode.SymmetricInset(16, 13), BoxShadow: giode.BoxShadow{Y: 8, Blur: 24, Color: "#8b5cf650"}}
	if job != nil {
		st.BackgroundGradient, st.Background, st.BoxShadow = nil, "theme.danger", giode.BoxShadow{}
		v.run.Label(tr.T("Stop")).Icon(icons.Solid(icons.Stop, 14, ""))
	} else {
		v.run.Label(tr.T("Transcribe")).Icon(icons.Icon(icons.Captions, 18, ""))
	}
	v.run.Styles(st).Disabled(job == nil && (!ready || v.p.source == "" || v.recorder.Recording()))

	var status giode.Element
	switch {
	case job != nil:
		line := giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: "theme.primary"}),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T(job.Runtime.Status),
				giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis})),
			giode.Text(core.Clock(time.Since(job.Started)), giode.Styles{Color: "theme.muted", FontSize: 12, FontFamily: theme.Mono}),
		)
		status = line
		if job.Runtime.Progress >= 0 {
			status = giode.Stack(giode.Styles{Width: -1, Gap: 8}, progressbar.New(job.Runtime.Progress), line)
		}
	case rt.Err != "" && rt.State == manager.StateReady:
		status = giode.Text(tr.T(firstLine(rt.Err)), giode.Styles{Color: "theme.danger", FontSize: 12, MaxLines: 3})
	case !ready:
		status = giode.Text(tr.T("Install the engine to start transcribing."), theme.Caption)
	default:
		status = giode.Text(tr.T("Transcripts are kept below, with their audio."), theme.Caption)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.UniformInset(20)},
		giode.Box(giode.Styles{Width: -1}, v.run),
		status,
	)
}

// engineSection is the engine picker, plus Eject while a model is in
// memory.
func (v *panel) engineSection() giode.Element {
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

func joinNonEmpty(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + " · " + b
}
