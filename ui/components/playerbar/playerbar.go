// Package playerbar is the now-playing strip: play/pause, the
// waveform, time and export actions.
package playerbar

import (
	"time"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/components/waveform"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Bar is stateful; create one per page that shows it.
type Bar struct {
	c     *core.Ctx
	play  *giode.PressableElement
	wave  *waveform.Waveform
	save  *giode.ButtonElement
	share *giode.ButtonElement
}

// New returns the bar for c's player.
func New(c *core.Ctx) *Bar {
	b := &Bar{c: c}
	b.play = giode.Pressable(func(s giode.PressState) giode.Element {
		icon := icons.Solid(icons.Play, 18, "#ffffff")
		if c.Player.Playing() {
			icon = icons.Solid(icons.Pause, 18, "#ffffff")
		}
		st := giode.Styles{Width: 46, Height: 46, BorderRadius: 99, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
			BackgroundGradient: theme.Brand(135), BoxShadow: giode.BoxShadow{Y: 6, Blur: 16, Color: "#8b5cf655"}}
		if c.Now == nil {
			st.BackgroundGradient, st.Background, st.BoxShadow = nil, "theme.text/0.1", giode.BoxShadow{}
		} else if s.Hovered {
			st.Transform = giode.Transform{Scale: 1.06}
		}
		return giode.Box(st, icon)
	}).OnClick(func(giode.ClickEvent) {
		if c.Now != nil {
			c.Player.Toggle()
		}
	})
	tr.Bind(func(s string) { b.play.Label(s) }, "Play")
	b.wave = &waveform.Waveform{
		Peaks: func() []float32 {
			if c.Now == nil {
				return nil
			}
			return c.Now.Peaks
		},
		Progress: c.Player.Progress,
		Busy:     func() bool { return c.Manager.Job != nil },
		Playing:  c.Player.Playing,
		OnSeek: func(f float32) {
			c.Player.Seek(f)
			if !c.Player.Playing() {
				c.Player.Play()
			}
		},
		Height: 52,
	}
	icon := giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(9), BorderRadius: 10}
	b.save = giode.Button("", icon).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Download, 18, "")).OnClick(func() {
		if c.Now != nil {
			c.SaveClip(c.Now.Path, c.Now.Title)
		}
	})
	b.share = giode.Button("", icon).Variant(giode.VariantGhost).Icon(icons.Icon(core.ShareIcon(), 18, "")).OnClick(func() {
		if c.Now != nil {
			c.ShareClip(c.Now.Path)
		}
	})
	return b
}

// Element lays out the bar.
func (b *Bar) Element() giode.Element {
	c := b.c
	title, sub, clock := tr.T("Nothing to play yet"), tr.T("Generated audio appears here"), "0:00"
	if c.Manager.Job != nil {
		title, sub = tr.T("Generating…"), c.Manager.Job.Runtime.Manifest().Name
	} else if c.Now != nil {
		title, sub = c.Now.Title, c.Now.Subtitle
		pos := float32(c.Now.Duration) * c.Player.Progress()
		clock = core.Clock(time.Duration(pos)) + " / " + core.Clock(c.Now.Duration)
	}
	card := theme.Card()
	card.FlexDirection, card.Align, card.Gap, card.Width = giode.FlexDirectionRow, giode.AlignCenter, 18, -1
	card.Padding = giode.SymmetricInset(18, 14)
	card.Wrap = giode.WrapNoWrap
	row := []giode.Element{b.play}
	if giode.Compact() {
		// Narrow: give the waveform the room the title would take.
		card.Gap, card.Padding = 12, giode.SymmetricInset(14, 12)
	} else {
		row = append(row, giode.Stack(giode.Styles{Width: 150, Gap: 3},
			giode.Text(title, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightBold, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(sub, giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}),
		))
	}
	row = append(row,
		giode.Box(giode.Styles{FlexGrow: 1}, b.wave),
		giode.Text(clock, giode.Styles{Color: "theme.secondary", FontSize: 12, FontFamily: theme.Mono}),
	)
	if c.Now != nil {
		row = append(row, b.save, b.share)
	}
	return giode.Row(card, row...)
}
