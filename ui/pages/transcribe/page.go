// Package transcribe is the speech-to-text screen: pick or record
// audio, choose an engine and language, transcribe, then edit, copy,
// export or speak the text.
package transcribe

import (
	"os"
	"path/filepath"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/components/emptystate"
	"github.com/mehran9675/sonora/ui/components/playerbar"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Page is the Transcribe screen. Create it once.
type Page struct {
	c       *core.Ctx
	panel   *panel
	result  *result
	history *history
	player  *playerbar.Bar
	// left scrolls the result and history in the wide layout; column
	// scrolls everything in the phone layout.
	left, column *scroll.Scroll
	addEngine    *giode.ButtonElement

	// source is the audio waiting to be transcribed; temp marks a copy
	// Sonora made (a recording, or a file from the Android picker) to
	// delete once it's transcribed.
	source, sourceName string
	temp               bool
}

// New builds the page and registers its hooks on c.
func New(c *core.Ctx) *Page {
	p := &Page{c: c}
	p.panel = newPanel(p)
	p.result = newResult(p)
	p.history = newHistory(p)
	p.player = playerbar.New(c)
	bar := giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.15", Width: 5}}
	p.left = giode.Scroll(p.leftColumn, bar)
	p.column = giode.Scroll(p.compactColumn, bar)
	p.addEngine = giode.Button("Open Engines", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}).
		Icon(icons.Icon(icons.Chip, 15, "")).OnClick(func() { c.ShowEngines(engines.TaskTranscribe) })
	c.UseTranscriber = func(id string) {
		p.selectEngine(id)
		c.Go(core.PageTranscribe)
	}
	if es := c.Transcripts.Entries(); len(es) > 0 {
		p.result.show(es[0])
	}
	// Recordings and picked files are only kept until they're
	// transcribed; clear what an earlier session left behind.
	olds, _ := filepath.Glob(filepath.Join(core.RecordingsDir(), "*"))
	for _, f := range olds {
		os.Remove(f)
	}
	return p
}

// engine returns the selected speech-to-text engine, falling back to
// the first ready one, or nil when there are none.
func (p *Page) engine() *manager.Runtime {
	m := p.c.Manager
	if rt := m.Get(p.c.Settings.Get().TranscribeEngine); rt != nil && rt.Manifest().Transcribes() {
		return rt
	}
	if ready := m.ReadyFor(engines.TaskTranscribe); len(ready) > 0 {
		return ready[0]
	}
	if all := m.EnginesFor(engines.TaskTranscribe); len(all) > 0 {
		return all[0]
	}
	return nil
}

func (p *Page) selectEngine(id string) {
	p.c.Settings.Update(func(s *settings.Settings) { s.TranscribeEngine = id })
}

// setSource replaces the audio waiting to be transcribed.
func (p *Page) setSource(path, name string, temp bool) {
	p.dropSource()
	p.source, p.sourceName, p.temp = path, name, temp
}

// dropSource forgets the waiting audio, deleting Sonora's own copy.
func (p *Page) dropSource() {
	if p.temp && p.source != "" {
		os.Remove(p.source)
	}
	p.source, p.sourceName, p.temp = "", "", false
}

// start transcribes the waiting audio, or stops the running job.
func (p *Page) start() {
	c := p.c
	if j := c.Manager.TranscribeJob; j != nil {
		c.Manager.Cancel(j.Runtime)
		return
	}
	rt := p.engine()
	if rt == nil || p.source == "" {
		return
	}
	s := c.Settings.Get()
	in := manager.TranscribeInput{Source: p.source, Title: p.sourceName,
		Language: language(rt.Manifest(), s.Language), Translate: s.Translate}
	c.Manager.Transcribe(rt, in, func(e transcripts.Entry, err error) {
		if err != nil {
			c.Error(err)
			return
		}
		p.dropSource()
		p.result.show(e)
	})
}

// language returns want if the engine understands it, else "" (detect).
func language(m *engines.Manifest, want string) string {
	if len(m.Languages) == 0 {
		return want
	}
	for _, l := range m.Languages {
		if l == want {
			return want
		}
	}
	return ""
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	rt := p.engine()
	if rt == nil {
		return emptystate.New(icons.Captions, "Add a speech-to-text engine",
			"Install Whisper or another transcription engine from the Engines page, or connect a transcription API.",
			p.addEngine)
	}
	if giode.Compact() {
		return giode.Box(giode.Styles{Width: -1, Height: -1}, p.column)
	}
	return giode.Row(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.UniformInset(24), Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Height: -1, Gap: 16},
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.left),
			p.player.Element(),
		),
		p.panel.element(rt),
	)
}

func (p *Page) leftColumn() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Gap: 16, Padding: giode.Inset{Right: 10}},
		p.result.element(0),
		p.history.element(),
	)
}

// compactColumn stacks everything for narrow screens: settings with
// the Transcribe button, the result, the player and past transcripts.
func (p *Page) compactColumn() giode.Element {
	rt := p.engine()
	if rt == nil {
		return giode.Box(giode.Styles{})
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 14, Padding: giode.UniformInset(core.Gutter())},
		p.panel.compactElement(rt),
		p.result.element(320),
		p.player.Element(),
		p.history.element(),
	)
}
