package library

import (
	"strings"

	"github.com/mehran9675/giode"

	lib "github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"

	"github.com/mehran9675/sonora/ui/tr"
)

// row is one clip in the list.
type row struct {
	c      *core.Ctx
	e      lib.Entry
	play   *giode.PressableElement
	reuse  *giode.ButtonElement
	share  *giode.ButtonElement
	remove *giode.ButtonElement
}

func newRow(c *core.Ctx, e lib.Entry) *row {
	r := &row{c: c, e: e}
	path := c.Library.Path(e)
	r.play = playButton(c, path, e.VoiceName, e.EngineName)
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(8), BorderRadius: 8}
	r.reuse = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Reuse, 16, "")).
		OnClick(func() { c.Reuse(e) })
	r.share = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Share, 16, "")).
		OnClick(func() { c.ExportFile(path, e.VoiceName+" - "+firstWords(e.Text, 6)) })
	r.remove = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Trash, 16, "")).OnClick(func() {
		giode.AfterLayout(func() {
			if c.Player.Loaded() == path {
				c.Player.Unload()
				c.Now = nil
			}
			c.Library.Remove(e.ID)
		})
	})
	return r
}

// actions are the row's buttons: reuse, export (save a copy, share) and
// delete.
func (r *row) actions() giode.Element {
	return giode.Row(giode.Styles{Gap: 2, Wrap: giode.WrapNoWrap}, r.reuse, r.share, r.remove)
}

func (r *row) element() giode.Element {
	e := r.e
	card := theme.Card()
	card.FlexDirection, card.Align, card.Gap, card.Width = giode.FlexDirectionRow, giode.AlignCenter, 16, -1
	card.Padding = giode.SymmetricInset(16, 14)
	card.Wrap = giode.WrapNoWrap
	return giode.Row(card,
		r.play,
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 8},
			// Two lines are always reserved so every card has the same height.
			giode.Box(giode.Styles{Width: -1, Height: 40, Overflow: giode.OverflowHidden},
				giode.Text(e.Text, giode.Styles{Color: "theme.text", FontSize: 14, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis, LineHeight: 20})),
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
				chip.Chip(e.EngineName),
				chip.Chip(e.VoiceName),
				giode.Text(core.Clock(e.Duration)+"  ·  "+core.Ago(e.Created), theme.Caption),
			),
		),
		r.actions(),
	)
}

// playButton plays and pauses the audio at path.
func playButton(c *core.Ctx, path, title, subtitle string) *giode.PressableElement {
	b := giode.Pressable(func(s giode.PressState) giode.Element {
		playing := c.IsPlaying(path)
		icon := icons.Solid(icons.Play, 14, "theme.primary")
		bg := "theme.primary/0.14"
		if playing {
			icon = icons.Solid(icons.Pause, 14, "#ffffff")
			bg = "theme.primary"
		} else if s.Hovered {
			bg = "theme.primary/0.24"
		}
		return giode.Box(giode.Styles{Width: 40, Height: 40, BorderRadius: 99, Background: bg,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icon)
	}).OnClick(func(giode.ClickEvent) { c.PlayPause(path, title, subtitle) })
	tr.Bind(func(s string) { b.Label(s) }, "Play")
	return b
}

// firstWords returns up to n words of text, for a file name.
func firstWords(text string, n int) string {
	w := strings.Fields(text)
	if len(w) > n {
		w = w[:n]
	}
	return strings.Join(w, " ")
}
