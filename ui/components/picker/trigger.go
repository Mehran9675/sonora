package picker

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Trigger is the field that opens a picker, showing the current choice.
type Trigger struct {
	p *giode.PressableElement
}

// NewTrigger returns a field named name (for accessibility) showing label() (and an optional leading
// element from lead()) that calls onClick.
func NewTrigger(name string, label func() (title, detail string), lead func() giode.Element, onClick func()) *Trigger {
	t := &Trigger{}
	t.p = giode.Pressable(func(s giode.PressState) giode.Element {
		title, detail := label()
		border := "theme.border"
		if s.Hovered {
			border = "theme.text/0.25"
		}
		text := []giode.Element{giode.Text(title, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium,
			TextOverflow: giode.TextOverflowEllipsis})}
		if detail != "" {
			text = append(text, giode.Text(detail, giode.Styles{Color: "theme.muted", FontSize: 12, TextOverflow: giode.TextOverflowEllipsis}))
		}
		row := []giode.Element{}
		if lead != nil {
			if el := lead(); el != nil {
				row = append(row, el)
			}
		}
		row = append(row,
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2}, text...),
			icons.Icon(icons.ChevronDown, 16, "theme.muted"),
		)
		return giode.Row(giode.Styles{
			Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 9), BorderRadius: theme.RadiusSm + 2,
			Background: "theme.text/0.03", BorderWidth: 1, BorderColor: border,
		}, row...)
	}).OnClick(func(giode.ClickEvent) { onClick() })
	tr.Bind(func(s string) { t.p.Label(s) }, name)
	return t
}

// Element returns the field for layout.
func (t *Trigger) Element() giode.Element { return t.p }
