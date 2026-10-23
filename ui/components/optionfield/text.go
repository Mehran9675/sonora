package optionfield

import (
	"fmt"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/theme"
)

// text is a single-line input with optional suggestion chips.
type text struct {
	in   *input.Input
	sugg []*giode.PressableElement
}

func newText(opt engines.Option, st Store) *text {
	t := &text{}
	placeholder := opt.Placeholder
	if placeholder == "" && opt.Default != nil {
		placeholder = fmt.Sprint(opt.Default)
	}
	t.in = giode.Input.Text(placeholder, theme.Field()).Label(opt.Label).
		OnChange(func(s string) { st.Set(opt.ID, s) })
	if opt.Type == engines.OptionSecret {
		t.in.Password(true)
	}
	t.in.SetText(fmt.Sprint(valueOr(st.Get(opt.ID), valueOr(opt.Default, ""))))
	for _, s := range opt.Suggestions {
		s := s
		t.sugg = append(t.sugg, giode.Pressable(func(ps giode.PressState) giode.Element {
			bg := "theme.text/0.05"
			if ps.Hovered {
				bg = "theme.primary/0.18"
			}
			return giode.Box(giode.Styles{Background: bg, BorderRadius: 99, Padding: giode.SymmetricInset(9, 3)},
				giode.Text(s, giode.Styles{Color: "theme.secondary", FontSize: 11, FontFamily: theme.Mono}))
		}).Label(s).OnClick(func(giode.ClickEvent) { t.in.ChangeText(s) }))
	}
	return t
}

func (t *text) Element() giode.Element {
	if len(t.sugg) == 0 {
		return giode.Box(giode.Styles{Width: -1}, t.in)
	}
	chips := make([]giode.Element, len(t.sugg))
	for i, s := range t.sugg {
		chips[i] = s
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8},
		giode.Box(giode.Styles{Width: -1}, t.in),
		giode.Row(giode.Styles{Width: -1, Gap: 6}, chips...),
	)
}
