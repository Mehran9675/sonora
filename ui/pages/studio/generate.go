package studio

import (
	"errors"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"

	"github.com/mehran9675/sonora/ui/tr"
)

// toggleGenerate starts a generation, or stops the running one.
func (p *Page) toggleGenerate() {
	c := p.c
	if j := c.Manager.Job; j != nil {
		c.Manager.Cancel(j.Runtime)
		return
	}
	rt := p.engine()
	if rt == nil {
		return
	}
	f := p.form(rt)
	reference := ""
	if f.cloneMode && rt.Manifest().Clone != engines.CloneNone {
		if f.reference == "" {
			c.Error(errors.New(tr.T("Choose a recording of the voice to clone")))
			return
		}
		if rt.Manifest().Transcript == engines.CloneRequired && strings.TrimSpace(f.transcript.Text()) == "" {
			c.Error(errors.New(tr.T("{name} needs the transcript of the recording", "name", rt.Manifest().Name)))
			return
		}
		reference = f.reference
	}
	text := p.editor.text()
	c.Settings.Update(func(s *settings.Settings) { s.Draft = text })
	in := manager.Input{
		Text:          text,
		Voice:         f.voice(),
		Speed:         f.speedValue(),
		Reference:     reference,
		ReferenceText: f.transcript.Text(),
	}
	awake := c.StayAwake(tr.T("Generating speech"))
	c.Manager.Generate(rt, in, func(e library.Entry, err error) {
		awake()
		if err != nil {
			c.Error(err)
			return
		}
		c.NotifyIfAway(tr.T("Speech is ready"), firstWords(e.Text))
		c.Open(c.Library.Path(e), e.VoiceName, e.EngineName, true)
		// Engines that report voices once loaded have them now.
		if len(rt.Voices) == 0 {
			rt.VoicesLoaded = false
		}
	})
}

// firstWords shortens text for a notification.
func firstWords(text string) string {
	if r := []rune(strings.TrimSpace(text)); len(r) > 80 {
		return string(r[:80]) + "…"
	}
	return strings.TrimSpace(text)
}
