// Package library is the Library screen: every clip generated so far
// and every recording made, searchable, playable and exportable.
package library

import (
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/scroll"

	lib "github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/ui/components/emptystate"
	"github.com/mehran9675/sonora/ui/components/pagehead"
	"github.com/mehran9675/sonora/ui/components/playerbar"
	"github.com/mehran9675/sonora/ui/components/segmented"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"

	"github.com/mehran9675/sonora/ui/tr"
)

// Page is the Library screen. Create it once.
type Page struct {
	c      *core.Ctx
	search *input.Input
	scroll *scroll.Scroll
	rows   *giode.KeyedCache[string, *row]
	player *playerbar.Bar
	studio *giode.ButtonElement

	// tab is "clips" or "recordings".
	tab        string
	tabs       *segmented.Control
	recordings *recordingList
	recRows    *giode.KeyedCache[string, *recRow]
}

// New builds the page.
func New(c *core.Ctx) *Page {
	p := &Page{c: c, rows: giode.NewKeyedCache[string, *row](), tab: "clips",
		recordings: &recordingList{c: c}, recRows: giode.NewKeyedCache[string, *recRow]()}
	p.tabs = segmented.New([]segmented.Option{{Value: "clips", Label: tr.Msg("Generated")}, {Value: "recordings", Label: tr.Msg("Recordings")}},
		func() string { return p.tab }, func(v string) { p.tab = v; p.recordings.refresh() })
	field := theme.Field()
	field.Width = 280
	p.search = giode.Input.Text("", field)
	tr.Bind(func(s string) { p.search.Placeholder(s) }, "Search…")
	p.scroll = giode.Scroll(p.list, theme.Scrolling(6))
	p.player = playerbar.New(c)
	p.studio = giode.Button("", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}).
		Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() { c.Go(core.PageStudio) })
	tr.Bind(func(s string) { p.studio.Label(s) }, "Open Studio")
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
	if p.tab == "recordings" {
		return p.recordingList()
	}
	entries := p.entries()
	if len(entries) == 0 {
		if len(p.c.Library.Entries()) == 0 {
			return emptystate.New(icons.Library, tr.T("No clips yet"),
				tr.T("Everything you generate in the Studio is saved here, ready to play again or export."), p.studio)
		}
		return emptystate.New(icons.Search, tr.T("No matches"), tr.T("Try a different word."))
	}
	els := make([]giode.Element, 0, len(entries))
	for _, e := range entries {
		e := e
		els = append(els, p.rows.Get(e.ID, func() *row { return newRow(p.c, e) }).element())
	}
	return column(els)
}

// recordingList lists the recordings folder.
func (p *Page) recordingList() giode.Element {
	q := strings.ToLower(strings.TrimSpace(p.search.Text()))
	all := p.recordings.list()
	var els []giode.Element
	for _, r := range all {
		if q != "" && !strings.Contains(strings.ToLower(r.Name), q) {
			continue
		}
		r := r
		key := r.Path + "|" + r.Modified.String()
		els = append(els, p.recRows.Get(key, func() *recRow { return newRecRow(p.c, p.recordings, r) }).element())
	}
	switch {
	case len(all) == 0:
		return emptystate.New(icons.Mic, tr.T("No recordings yet"),
			tr.T("Recordings you make to transcribe or to clone a voice are kept here, ready to play, transcribe or export."))
	case len(els) == 0:
		return emptystate.New(icons.Search, tr.T("No matches"), tr.T("Try a different word."))
	}
	return column(els)
}

// column stacks list rows.
func column(els []giode.Element) giode.Element {
	// Room for the scrollbar, which phones don't show.
	right := 12
	if giode.Compact() {
		right = 0
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.Inset{Right: right, Bottom: 12}}, els...)
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	// The search field spans the screen on a phone.
	field := theme.Field()
	field.Width = 280
	if giode.Compact() {
		field.Width = -1
	}
	p.search.Styles(field)
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 20, Padding: giode.UniformInset(core.Gutter())},
		pagehead.Head(giode.Compact(), tr.T("Library"), tr.T("Your generated audio and recordings, newest first."), p.search),
		giode.Box(giode.Styles{Width: -1, MaxWidth: 420}, p.tabs.Element()),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.scroll),
		p.player.Element(),
	)
}
