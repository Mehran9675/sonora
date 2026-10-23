// Package library is the Library screen: every clip generated so far,
// searchable, playable and exportable.
package library

import (
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/scroll"

	lib "github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/ui/components/emptystate"
	"github.com/mehran9675/sonora/ui/components/focusarea"
	"github.com/mehran9675/sonora/ui/components/pagehead"
	"github.com/mehran9675/sonora/ui/components/playerbar"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Page is the Library screen. Create it once.
type Page struct {
	c      *core.Ctx
	search *input.Input
	find   *focusarea.Area
	scroll *scroll.Scroll
	rows   *giode.KeyedCache[string, *row]
	player *playerbar.Bar
	studio *giode.ButtonElement
}

// New builds the page.
func New(c *core.Ctx) *Page {
	p := &Page{c: c, rows: giode.NewKeyedCache[string, *row]()}
	p.search = giode.Input.Text("Search your clips…", theme.Field())
	p.find = focusarea.New(p.search, giode.Styles{Width: 280})
	p.scroll = giode.Scroll(p.list, giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.15", Width: 6}})
	p.player = playerbar.New(c)
	p.studio = giode.Button("Open Studio", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}).
		Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() { c.Go(core.PageStudio) })
	return p
}

func (p *Page) entries() []lib.Entry {
	q := strings.ToLower(strings.TrimSpace(p.search.Text()))
	var out []lib.Entry
	for _, e := range p.c.Library.Entries() {
		if q == "" || strings.Contains(strings.ToLower(e.Text+" "+e.EngineName+" "+e.VoiceName), q) {
			out = append(out, e)
		}
	}
	return out
}

func (p *Page) list() giode.Element {
	entries := p.entries()
	if len(entries) == 0 {
		if len(p.c.Library.Entries()) == 0 {
			return emptystate.New(icons.Library, "No clips yet",
				"Everything you generate in the Studio is saved here, ready to play again or export.", p.studio)
		}
		return emptystate.New(icons.Search, "No matches", "Try a different word.")
	}
	els := make([]giode.Element, 0, len(entries))
	for _, e := range entries {
		e := e
		els = append(els, p.rows.Get(e.ID, func() *row { return newRow(p.c, e) }).element())
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.Inset{Right: 12, Bottom: 12}}, els...)
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.UniformInset(28)},
		pagehead.Head("Library", "Your generated audio, newest first.", p.find.Element()),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.scroll),
		p.player.Element(),
	)
}
