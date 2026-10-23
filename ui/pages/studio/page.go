// Package studio is the main screen: write text, choose an engine and
// voice, generate, and listen.
package studio

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/emptystate"
	"github.com/mehran9675/sonora/ui/components/playerbar"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Page is the Studio. Create it once.
type Page struct {
	c       *core.Ctx
	editor  *editor
	engines *enginePicker
	forms   map[string]*engineForm
	panel   *voicePanel
	player  *playerbar.Bar
	// column scrolls the one-column layout used on narrow screens.
	column *scroll.Scroll
	// addEngine leads to the Engines page when no engine exists yet,
	// which happens on mobile until an API server is added.
	addEngine *giode.ButtonElement
}

// New builds the Studio and registers its hooks on c.
func New(c *core.Ctx) *Page {
	p := &Page{c: c, forms: map[string]*engineForm{}}
	p.editor = newEditor(c)
	p.engines = newEnginePicker(c, p.selectEngine)
	p.player = playerbar.New(c)
	p.panel = newVoicePanel(p)
	p.column = giode.Scroll(p.compactColumn, giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.15", Width: 4}})
	p.addEngine = giode.Button("Add an engine", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}).
		Icon(icons.Icon(icons.Plus, 15, "")).OnClick(func() { c.Go(core.PageEngines) })
	c.UseEngine = func(id string) {
		p.selectEngine(id)
		c.Go(core.PageStudio)
	}
	c.Reuse = func(e library.Entry) {
		p.editor.setText(e.Text)
		if c.Manager.Get(e.EngineID) != nil {
			p.selectEngine(e.EngineID)
			c.Settings.Update(func(s *settings.Settings) { s.Voices[e.EngineID] = e.Voice })
		}
		c.Open(c.Library.Path(e), e.EngineName, e.VoiceName, false)
		c.Go(core.PageStudio)
	}
	return p
}

// engine returns the selected engine, falling back to the first ready
// one, or nil when there are no engines at all.
func (p *Page) engine() *manager.Runtime {
	if rt := p.c.Manager.Get(p.c.Settings.Get().Engine); rt != nil {
		return rt
	}
	if ready := p.c.Manager.Ready(); len(ready) > 0 {
		return ready[0]
	}
	if all := p.c.Manager.Engines(); len(all) > 0 {
		return all[0]
	}
	return nil
}

func (p *Page) selectEngine(id string) {
	p.c.Settings.Update(func(s *settings.Settings) { s.Engine = id })
	if rt := p.c.Manager.Get(id); rt != nil {
		p.c.Manager.LoadVoices(rt, false)
	}
}

// form returns the cached settings form for rt.
func (p *Page) form(rt *manager.Runtime) *engineForm {
	id := rt.Manifest().ID
	f := p.forms[id]
	if f == nil || f.rt != rt {
		f = newEngineForm(p.c, rt)
		p.forms[id] = f
	}
	return f
}

// SaveDraft stores the editor text. It's safe from any goroutine, as
// OnBackground on Android requires.
func (p *Page) SaveDraft() {
	text := p.editor.savedText()
	p.c.Settings.Update(func(s *settings.Settings) { s.Draft = text })
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	rt := p.engine()
	if rt == nil {
		return emptystate.New(icons.Globe, "Connect a voice engine",
			"Add an API server, such as Kokoro-FastAPI on your computer or a cloud TTS service, and Sonora speaks through it.",
			p.addEngine)
	}
	p.c.Manager.LoadVoices(rt, false)
	if giode.Compact() {
		return giode.Box(giode.Styles{Width: -1, Height: -1}, p.column)
	}
	return giode.Row(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.UniformInset(24), Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Height: -1, Gap: 16},
			p.editor.element(rt, 0),
			p.player.Element(),
		),
		p.panel.element(rt),
	)
}

// compactColumn stacks the same parts in one scrolling column for
// narrow screens: script, voice settings with Generate, player.
func (p *Page) compactColumn() giode.Element {
	rt := p.engine()
	if rt == nil {
		return giode.Box(giode.Styles{})
	}
	g := core.Gutter()
	return giode.Stack(giode.Styles{Width: -1, Gap: 14, Padding: giode.UniformInset(g)},
		p.editor.element(rt, 240),
		p.panel.compactElement(rt),
		p.player.Element(),
	)
}
