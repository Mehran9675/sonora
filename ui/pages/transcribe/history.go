package transcribe

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// history lists earlier transcripts; a click opens one.
type history struct {
	p    *Page
	rows *giode.KeyedCache[string, *row]
}

func newHistory(p *Page) *history {
	return &history{p: p, rows: giode.NewKeyedCache[string, *row]()}
}

func (h *history) element() giode.Element {
	entries := h.p.c.Transcripts.Entries()
	if len(entries) == 0 {
		return giode.Box(giode.Styles{})
	}
	els := []giode.Element{giode.Text("Earlier transcripts", theme.Overline)}
	for _, e := range entries {
		e := e
		els = append(els, h.rows.Get(e.ID, func() *row { return newRow(h.p, e) }).element())
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.Inset{Top: 8, Bottom: 12}}, els...)
}

// row is one past transcript.
type row struct {
	p      *Page
	id     string
	open   *giode.PressableElement
	export *giode.ButtonElement
	remove *giode.ButtonElement
}

func newRow(p *Page, e transcripts.Entry) *row {
	r := &row{p: p, id: e.ID}
	c := p.c
	r.open = giode.Pressable(func(s giode.PressState) giode.Element {
		// Read the stored entry each frame: its text may have been edited.
		cur, _ := c.Transcripts.Get(r.id)
		selected := p.result.entry != nil && p.result.entry.ID == r.id
		bg := ""
		switch {
		case selected:
			bg = "theme.primary/0.10"
		case s.Hovered:
			bg = "theme.text/0.04"
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(10, 8), BorderRadius: theme.RadiusSm + 2, Background: bg},
			giode.Box(giode.Styles{Width: 34, Height: 34, BorderRadius: 10, Background: "theme.primary/0.12",
				Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icons.Icon(icons.Captions, 16, "theme.primary")),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(cur.Title, giode.Styles{Color: "theme.text", FontSize: 13, FontWeight: giode.FontWeightMedium, TextOverflow: giode.TextOverflowEllipsis}),
				giode.Text(cur.Text, giode.Styles{Color: "theme.muted", FontSize: 12, MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
			),
			giode.Text(core.Clock(cur.Duration)+"  ·  "+core.Ago(cur.Created), theme.Caption),
		)
	}).Label("Open transcript").OnClick(func(giode.ClickEvent) {
		if cur, ok := c.Transcripts.Get(r.id); ok {
			p.result.show(cur)
		}
	})
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(8), BorderRadius: 8}
	r.export = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Share, 15, "")).OnClick(func() {
		if cur, ok := c.Transcripts.Get(r.id); ok {
			c.ExportTranscript(cur, cur.Text)
		}
	})
	r.remove = giode.Button("", ghost).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Trash, 15, "")).OnClick(func() {
		giode.AfterLayout(func() {
			cur, _ := c.Transcripts.Get(r.id)
			if path := c.Transcripts.AudioPath(cur); c.Player.Loaded() == path {
				c.Player.Unload()
				c.Now = nil
			}
			if p.result.entry != nil && p.result.entry.ID == r.id {
				p.result.clear()
			}
			c.Transcripts.Remove(r.id)
		})
	})
	return r
}

func (r *row) element() giode.Element {
	return giode.Row(giode.Styles{Width: -1, Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, r.open), r.export, r.remove)
}
