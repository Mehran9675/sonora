package studio

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/textfile"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// expandedMargin is the room, in dp, an expanded script card leaves for
// the tab bar, the title bar and page padding.
const expandedMargin = 150

// editor is the script card.
type editor struct {
	in     *input.Input
	open   *giode.ButtonElement
	sample *giode.ButtonElement
	clear  *giode.ButtonElement
	expand *giode.ButtonElement
	next   int
	// expanded grows the script card to the screen's height in the
	// phone layout, where it's otherwise a fixed size.
	expanded bool

	// draft mirrors the text for OnBackground, which Android calls on
	// its own thread, where the editor must not be touched.
	mu    sync.Mutex
	draft string
}

func newEditor(c *core.Ctx) *editor {
	e := &editor{}
	e.in = giode.Input.Text("Type or paste the text you want to hear…", giode.Styles{
		BorderColor: "transparent", Padding: giode.UniformInset(0), Color: "theme.text", FontSize: 17, Width: -1,
		Height: -1, // fill the card so a tap anywhere in it starts typing
	}).Multiline(true).Label("Script").OnChange(func(text string) {
		e.mu.Lock()
		e.draft = text
		e.mu.Unlock()
	})
	e.draft = c.Settings.Get().Draft
	e.in.SetText(e.draft)
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", FontSize: 13, Padding: giode.SymmetricInset(10, 6), BorderRadius: 8}
	e.sample = giode.Button("Try a sample", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Sparkles, 15, "")).OnClick(func() {
		e.in.ChangeText(samples[e.next%len(samples)])
		e.next++
	})
	e.clear = giode.Button("Clear", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Trash, 15, "")).OnClick(func() { e.in.ChangeText("") })
	e.expand = giode.Button("", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.Expand, 15, "")).OnClick(func() { e.expanded = !e.expanded })
	e.open = giode.Button("Open file", ghost).Variant(giode.VariantGhost).
		Icon(icons.Icon(icons.FileText, 15, "")).OnClick(func() { e.openFile(c) })
	return e
}

// openFile replaces the script with the text of a document the user
// picks: text, Markdown, subtitles, HTML, Word, OpenDocument, EPUB, PDF
// or RTF.
func (e *editor) openFile(c *core.Ctx) {
	var text, name string
	c.Background(func() error {
		f, err := c.PickFile(textfile.Extensions, "")
		if err != nil {
			return err
		}
		defer f.Close()
		name = f.Name
		text, err = textfile.Read(name, f)
		return err
	}, func(err error) {
		if err != nil {
			c.Error(err)
			return
		}
		e.in.ChangeText(text)
		words := len(strings.Fields(text))
		msg := fmt.Sprintf("Loaded %d words", words)
		if name != "" {
			msg += " from " + name
		}
		c.Success(msg)
	})
}

func (e *editor) text() string     { return e.in.Text() }
func (e *editor) setText(s string) { e.in.ChangeText(s) }

// savedText returns the text from any goroutine.
func (e *editor) savedText() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.draft
}

// element lays out the script card; height 0 fills the space available,
// a positive height fixes it (in the scrolling phone layout).
func (e *editor) element(rt *manager.Runtime, height int) giode.Element {
	text := e.in.Text()
	words := len(strings.Fields(text))
	secs := float64(words) / 2.6
	stats := fmt.Sprintf("%d characters · %d words · ≈ %s", len([]rune(text)), words, core.Clock(time.Duration(secs*float64(time.Second))))
	footer := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(stats, theme.Caption))}
	// On a phone the hint shows in the voice panel instead: beside the
	// stats it squeezes them, and the text field, to nothing.
	if hint := rt.Manifest().Hint; hint != "" && rt.State != manager.StateUnavailable && height == 0 {
		footer = append(footer, chip.Chip(hint))
	}
	card := theme.Card()
	card.FlexGrow, card.Width, card.Gap = 1, -1, 12
	header := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Script", theme.Heading)), e.open, e.sample, e.clear}
	if height > 0 {
		if e.expanded {
			// Fill the screen below the tab bar and page padding; Back
			// shrinks it again.
			giode.HandleBack(func() { e.expanded = false })
			height = max(height, int(float32(elements.WindowSize().Y)/giode.DisplayScale())-expandedMargin)
			e.expand.Icon(icons.Icon(icons.Collapse, 15, ""))
		} else {
			e.expand.Icon(icons.Icon(icons.Expand, 15, ""))
		}
		card.FlexGrow, card.Height = 0, properties.Height(height)
		header = append(header, e.expand)
	}
	card.Padding = giode.Inset{Top: 16, Left: 24, Right: 16, Bottom: 16}
	// A phone has room for the buttons' icons only.
	if giode.Compact() {
		card.Padding.Left = 18
		e.open.Label("")
		e.sample.Label("")
		e.clear.Label("")
	} else {
		e.open.Label("Open file")
		e.sample.Label("Try a sample")
		e.clear.Label("Clear")
	}
	return giode.Stack(card,
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Gap: 4}, header...),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Padding: giode.Inset{Right: 8}, Overflow: giode.OverflowHidden}, e.in),
		giode.Divider(giode.Styles{Color: "theme.border"}),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 8, Wrap: giode.WrapNoWrap}, footer...),
	)
}
