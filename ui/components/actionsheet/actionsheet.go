// Package actionsheet is a dialog listing actions for one item - save
// a copy, share, export as… - opened from a single button, which keeps
// rows narrow enough for a phone. One sheet serves the whole app.
package actionsheet

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Action is one choice in the sheet.
type Action struct {
	Icon   string // path from package icons
	Label  string
	Detail string // optional second line
	Do     func()
}

// Sheet is the shared dialog. Create it once.
type Sheet struct {
	open    bool
	title   string
	actions []Action
	rows    []*giode.PressableElement
	close   *giode.ButtonElement
}

// New creates the sheet and registers it as an overlay.
func New() *Sheet {
	s := &Sheet{}
	s.close = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(6), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 16, "")).OnClick(func() { s.open = false })
	overlay.Add(giode.Dialog(&s.open, s.content, giode.Styles{
		Background: "theme.surface", BorderRadius: theme.RadiusLg, Padding: giode.UniformInset(20),
		BorderWidth: 1, BorderColor: "theme.border",
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000080"},
	}))
	return s
}

// Show opens the sheet titled title with actions.
func (s *Sheet) Show(title string, actions ...Action) {
	s.title, s.actions = title, actions
	s.rows = make([]*giode.PressableElement, len(actions))
	for i, a := range actions {
		a := a
		s.rows[i] = giode.Pressable(func(ps giode.PressState) giode.Element {
			bg := ""
			if ps.Hovered {
				bg = "theme.text/0.05"
			}
			text := []giode.Element{giode.Text(tr.T(a.Label), giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium})}
			if a.Detail != "" {
				text = append(text, giode.Text(tr.T(a.Detail), theme.Caption))
			}
			return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 12, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 11), BorderRadius: theme.RadiusSm + 2, Background: bg},
				giode.Box(giode.Styles{Width: 34, Height: 34, BorderRadius: 10, Background: "theme.primary/0.12",
					Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icons.Icon(a.Icon, 16, "theme.primary")),
				giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2}, text...),
			)
		}).Label(tr.T(a.Label)).OnClick(func(giode.ClickEvent) {
			s.open = false
			a.Do()
		})
	}
	s.open = true
}

func (s *Sheet) content() giode.Element {
	els := []giode.Element{giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Bottom: 4}},
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T(s.title), giode.Styles{Color: "theme.text", FontSize: 16, FontWeight: giode.FontWeightBold,
			MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})),
		s.close,
	)}
	for _, r := range s.rows {
		els = append(els, r)
	}
	return giode.Stack(giode.Styles{Width: -1, MaxWidth: 420, Gap: 2}, els...)
}
