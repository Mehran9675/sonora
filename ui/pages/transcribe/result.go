package transcribe

import (
	"fmt"
	"strings"
	"time"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// result shows the selected transcript, editable, with its exports.
type result struct {
	p     *Page
	entry *transcripts.Entry
	text  *input.Input
	save  *time.Timer // debounces saving edits

	play, copy, export, speak *giode.ButtonElement
}

func newResult(p *Page) *result {
	r := &result{p: p}
	c := p.c
	r.text = giode.Input.Text("The transcript appears here.", giode.Styles{
		BorderColor: "transparent", Padding: giode.UniformInset(0), Color: "theme.text", FontSize: 16, LineHeight: 24,
		Width: -1, Height: -1,
	}).Multiline(true).Label("Transcript").OnChange(r.edited)
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 13, Padding: giode.SymmetricInset(10, 6), BorderRadius: 8}
	button := func(label, icon string, fn func()) *giode.ButtonElement {
		return giode.Button(label, ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icon, 15, "")).OnClick(func() {
			if r.entry != nil {
				fn()
			}
		})
	}
	r.play = button("Play", icons.Play, func() {
		c.PlayPause(c.Transcripts.AudioPath(*r.entry), r.entry.Title, "Transcribed audio")
	})
	r.copy = button("Copy", icons.Copy, func() {
		c.App.SetClipboard(r.text.Text())
		c.Success("Copied")
	})
	r.export = button("Export", icons.Share, func() { c.ExportTranscript(*r.entry, r.text.Text()) })
	r.speak = button("Speak", icons.Speaker, func() { c.Speak(r.text.Text()) })
	return r
}

// show selects e.
func (r *result) show(e transcripts.Entry) {
	r.flush()
	r.entry = &e
	r.text.SetText(e.Text)
}

// clear deselects the transcript (after it's deleted).
func (r *result) clear() {
	if r.save != nil {
		r.save.Stop()
	}
	r.entry = nil
	r.text.SetText("")
}

// edited saves the user's changes shortly after they stop typing.
func (r *result) edited(text string) {
	if r.entry == nil {
		return
	}
	r.entry.Text = text
	if r.save != nil {
		r.save.Stop()
	}
	id := r.entry.ID
	r.save = time.AfterFunc(600*time.Millisecond, func() { r.p.c.Transcripts.SetText(id, text) })
}

// flush saves pending edits now.
func (r *result) flush() {
	if r.save != nil && r.save.Stop() && r.entry != nil {
		r.p.c.Transcripts.SetText(r.entry.ID, r.entry.Text)
	}
}

// element lays out the card; height 0 fills the space available.
func (r *result) element(height int) giode.Element {
	card := theme.Card()
	card.Width, card.Gap = -1, 12
	card.Padding = giode.Inset{Top: 16, Left: 24, Right: 16, Bottom: 16}
	if giode.Compact() {
		card.Padding.Left = 18
	}
	textHeight := 360
	if height > 0 {
		textHeight = height - 120
	}
	if r.entry == nil {
		return giode.Stack(card,
			giode.Text("Transcript", theme.Heading),
			giode.Text("Choose an audio file or record, then press Transcribe. The text appears here, "+
				"ready to edit, copy, export or share as text or subtitles, or turn back into speech in the Studio.",
				giode.Styles{Color: "theme.muted", FontSize: 13, LineHeight: 19}),
		)
	}
	e := r.entry
	meta := []giode.Element{chip.Chip(e.EngineName)}
	if l := languageName(e.Language); l != "" {
		label := l
		if e.Translated {
			label += " → English"
		}
		meta = append(meta, chip.Chip(label))
	}
	words := len(strings.Fields(r.text.Text()))
	meta = append(meta, giode.Text(fmt.Sprintf("%s  ·  %d words  ·  %s", core.Clock(e.Duration), words, core.Ago(e.Created)), theme.Caption))

	// A phone has room for the buttons' icons only.
	for _, b := range []struct {
		btn   *giode.ButtonElement
		label string
	}{{r.play, "Play"}, {r.copy, "Copy"}, {r.export, "Export"}, {r.speak, "Speak"}} {
		if giode.Compact() {
			b.label = ""
		}
		b.btn.Label(b.label)
	}
	return giode.Stack(card,
		giode.Stack(giode.Styles{Width: -1, Gap: 6},
			giode.Text(e.Title, giode.Styles{Color: "theme.text", FontSize: 15, FontWeight: giode.FontWeightBold, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter}, meta...),
		),
		giode.Box(giode.Styles{Width: -1, Height: properties.Height(textHeight), Padding: giode.Inset{Right: 8},
			Overflow: giode.OverflowHidden}, r.text),
		giode.Divider(giode.Styles{Color: "theme.border"}),
		giode.Row(giode.Styles{Width: -1, Gap: 4, Align: giode.AlignCenter}, r.play, r.copy, r.export, r.speak),
	)
}

// languageName names a language code ("de" → "German").
func languageName(code string) string {
	for _, l := range engines.Languages {
		if l.Value != "" && l.Value == strings.ToLower(code) {
			return l.Label
		}
	}
	return strings.ToUpper(code)
}
