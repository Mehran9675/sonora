package studio

import (
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
)

// toggleGenerate starts a generation, or stops the running one.
func (p *Page) toggleGenerate() {
	c := p.c
	if j := c.Manager.Job; j != nil {
		c.Manager.Cancel(j.Runtime)
		return
	}
	rt := p.engine()
	f := p.form(rt)
	text := p.editor.text()
	c.Settings.Update(func(s *settings.Settings) { s.Draft = text })
	in := manager.Input{
		Text:          text,
		Voice:         f.voice(),
		Speed:         f.speedValue(),
		Reference:     f.reference,
		ReferenceText: f.transcript.Text(),
	}
	c.Manager.Generate(rt, in, func(e library.Entry, err error) {
		if err != nil {
			c.Error(err)
			return
		}
		c.Open(c.Library.Path(e), e.VoiceName, e.EngineName, true)
		// Engines that report voices once loaded have them now.
		if len(rt.Voices) == 0 {
			rt.VoicesLoaded = false
		}
	})
}
