// Package studio is the main screen: write text, choose an engine and
// voice, generate, and listen.
package studio

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/playerbar"
	"github.com/mehran9675/sonora/ui/core"
)

// Page is the Studio. Create it once.
type Page struct {
	c       *core.Ctx
	editor  *editor
	engines *enginePicker
	forms   map[string]*engineForm
	panel   *voicePanel
	player  *playerbar.Bar
}

// New builds the Studio and registers its hooks on c.
func New(c *core.Ctx) *Page {
	p := &Page{c: c, forms: map[string]*engineForm{}}
	p.editor = newEditor(c)
	p.engines = newEnginePicker(c, p.selectEngine)
	p.player = playerbar.New(c)
	p.panel = newVoicePanel(p)
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
// one.
func (p *Page) engine() *manager.Runtime {
	if rt := p.c.Manager.Get(p.c.Settings.Get().Engine); rt != nil {
		return rt
	}
	if ready := p.c.Manager.Ready(); len(ready) > 0 {
		return ready[0]
	}
	return p.c.Manager.Engines()[0]
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

// SaveDraft stores the editor text (called on shutdown).
func (p *Page) SaveDraft() {
	text := p.editor.text()
	p.c.Settings.Update(func(s *settings.Settings) { s.Draft = text })
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	rt := p.engine()
	p.c.Manager.LoadVoices(rt, false)
	return giode.Row(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.UniformInset(24), Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Height: -1, Gap: 16},
			p.editor.element(rt),
			p.player.Element(),
		),
		p.panel.element(rt),
	)
}
