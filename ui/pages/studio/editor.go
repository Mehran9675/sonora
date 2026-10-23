package studio

import (
	"fmt"
	"strings"
	"time"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/focusarea"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// editor is the script card.
type editor struct {
	in     *input.Input
	area   *focusarea.Area
	sample *giode.ButtonElement
	clear  *giode.ButtonElement
	next   int
}

func newEditor(c *core.Ctx) *editor {
	e := &editor{}
	e.in = giode.Input.Text("Type or paste the text you want to hear…", giode.Styles{
		BorderColor: "transparent", Padding: giode.UniformInset(0), Color: "theme.text", FontSize: 17, Width: -1,
	}).Multiline(true).Label("Script")
	e.in.SetText(c.Settings.Get().Draft)
	e.area = focusarea.New(e.in, giode.Styles{Width: -1, Height: -1, FlexGrow: 1, Padding: giode.Inset{Right: 8}, Overflow: giode.OverflowHidden})
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 13, Padding: giode.SymmetricInset(10, 6), BorderRadius: 8}
	e.sample = giode.Button("Try a sample", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() {
		e.in.ChangeText(samples[e.next%len(samples)])
		e.next++
	})
	e.clear = giode.Button("Clear", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Trash, 15, "")).OnClick(func() { e.in.ChangeText("") })
	return e
}

func (e *editor) text() string     { return e.in.Text() }
func (e *editor) setText(s string) { e.in.ChangeText(s) }

func (e *editor) element(rt *manager.Runtime) giode.Element {
	text := e.in.Text()
	words := len(strings.Fields(text))
	secs := float64(words) / 2.6
	stats := fmt.Sprintf("%d characters · %d words · ≈ %s", len([]rune(text)), words, core.Clock(time.Duration(secs*float64(time.Second))))
	footer := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(stats, theme.Caption))}
	if hint := rt.Manifest().Hint; hint != "" && rt.State != manager.StateUnavailable {
		footer = append(footer, chip.Chip(hint))
	}
	card := theme.Card()
	card.FlexGrow, card.Width, card.Gap = 1, -1, 12
	card.Padding = giode.Inset{Top: 16, Left: 24, Right: 16, Bottom: 16}
	return giode.Stack(card,
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Gap: 4},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Script", theme.Heading)),
			e.sample, e.clear,
		),
		e.area.Element(),
		giode.Divider(giode.Styles{Color: "theme.border"}),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 8, Wrap: giode.WrapNoWrap}, footer...),
	)
}
