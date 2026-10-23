package picker

import (
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Picker is a modal list. Create it once; it registers itself as an
// overlay.
type Picker struct {
	open     bool
	title    string
	items    func() []Item
	selected func() string
	onPick   func(Item)

	search *input.Input
	scroll *scroll.Scroll
	rows   *giode.KeyedCache[string, *giode.PressableElement]
	close  *giode.ButtonElement
	dialog *dialog.Dialog
}

// New returns a picker titled title listing items(); selected returns
// the current choice's ID; onPick runs when the user chooses.
func New(title string, items func() []Item, selected func() string, onPick func(Item)) *Picker {
	p := &Picker{title: title, items: items, selected: selected, onPick: onPick}
	field := theme.Field()
	field.Padding = giode.SymmetricInset(12, 10)
	p.search = giode.Input.Text("Search…", field)
	p.rows = giode.NewKeyedCache[string, *giode.PressableElement]()
	p.scroll = giode.Scroll(p.list, giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.2", Width: 6}})
	p.close = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(6), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 16, "")).OnClick(func() { p.open = false })
	p.dialog = giode.Dialog(&p.open, p.content, giode.Styles{
		Background: "theme.surface", BorderRadius: theme.RadiusLg, Padding: giode.UniformInset(20),
		BorderWidth: 1, BorderColor: "theme.border",
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000080"},
	})
	overlay.Add(p.dialog)
	return p
}

// Open shows the picker with an empty search.
func (p *Picker) Open() {
	p.search.SetText("")
	p.search.Focus()
	p.scroll.ScrollTo(0)
	p.open = true
}

func (p *Picker) content() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, MaxWidth: 460, Gap: 14},
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(p.title, theme.Heading)),
			p.close,
		),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 8, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, p.search),
		),
		giode.Box(giode.Styles{Width: -1, Height: 380}, p.scroll),
	)
}

func (p *Picker) list() giode.Element {
	q := strings.ToLower(strings.TrimSpace(p.search.Text()))
	var rows []giode.Element
	group := "\x00"
	for _, it := range p.items() {
		if !it.matches(q) {
			continue
		}
		if it.Group != group {
			group = it.Group
			if group != "" {
				rows = append(rows, giode.Box(giode.Styles{Padding: giode.Inset{Top: 10, Bottom: 4, Left: 10}},
					giode.Text(group, theme.Overline)))
			}
		}
		rows = append(rows, p.row(it))
	}
	if len(rows) == 0 {
		rows = append(rows, giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(24)},
			giode.Text("Nothing matches “"+q+"”", giode.Styles{Color: "theme.muted", TextAlign: giode.TextAlignCenter, Width: -1})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2, Padding: giode.Inset{Right: 10}}, rows...)
}

func (p *Picker) row(it Item) giode.Element {
	return p.rows.Get(it.ID+"\x00"+it.Label, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			sel := p.selected() == it.ID
			bg := ""
			switch {
			case sel:
				bg = "theme.primary/0.16"
			case s.Hovered:
				bg = "theme.text/0.05"
			}
			text := []giode.Element{giode.Text(it.Label, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium})}
			if it.Detail != "" {
				text = append(text, giode.Text(it.Detail, theme.Caption))
			}
			trail := giode.Box(giode.Styles{Width: 18})
			if sel {
				trail = giode.Box(giode.Styles{}, icons.Icon(icons.Check, 18, "theme.primary"))
			}
			return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 9), BorderRadius: theme.RadiusSm + 2, Background: bg},
				giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2}, text...),
				trail,
			)
		}).Label(it.Label).OnClick(func(giode.ClickEvent) {
			p.open = false
			p.onPick(it)
		})
	})
}
