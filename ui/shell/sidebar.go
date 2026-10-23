// Package shell is the window frame: title bar, sidebar navigation and
// the page area.
package shell

import (
	"fmt"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/components/navitem"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Sidebar is the left navigation column.
type Sidebar struct {
	c     *core.Ctx
	items []*navitem.Item
	theme *giode.PressableElement
}

// NewSidebar builds the navigation for c.
func NewSidebar(c *core.Ctx, onTheme func()) *Sidebar {
	s := &Sidebar{c: c}
	nav := func(icon, label string, p core.Page, badge func() string) {
		s.items = append(s.items, navitem.New(icon, label,
			func() bool { return c.Page() == p }, badge, func() { c.Go(p) }))
	}
	nav(icons.Waves, "Studio", core.PageStudio, nil)
	nav(icons.Library, "Library", core.PageLibrary, func() string {
		if n := len(c.Library.Entries()); n > 0 {
			return fmt.Sprint(n)
		}
		return ""
	})
	nav(icons.Chip, "Engines", core.PageEngines, func() string {
		return fmt.Sprintf("%d/%d", len(c.Manager.Ready()), len(c.Manager.Engines()))
	})
	nav(icons.Sliders, "Settings", core.PageSettings, nil)
	s.theme = giode.Pressable(func(ps giode.PressState) giode.Element {
		icon, label := icons.Sun, "Light mode"
		if !theme.IsDark() {
			icon, label = icons.Moon, "Dark mode"
		}
		fg := "theme.muted"
		bg := ""
		if ps.Hovered {
			fg, bg = "theme.text", "theme.text/0.05"
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 10), BorderRadius: theme.RadiusSm + 2, Background: bg},
			icons.Icon(icon, 18, fg), giode.Text(label, giode.Styles{Color: fg, FontSize: 13, FontWeight: giode.FontWeightMedium}))
	}).Label("Toggle theme").OnClick(func(giode.ClickEvent) { onTheme() })
	return s
}

// Element lays out the sidebar.
func (s *Sidebar) Element() giode.Element {
	els := []giode.Element{logo(), giode.Box(giode.Styles{Height: 18})}
	for _, it := range s.items {
		els = append(els, it.Element())
	}
	els = append(els,
		giode.Box(giode.Styles{FlexGrow: 1}),
		s.engineSummary(),
		s.theme,
	)
	return giode.Stack(giode.Styles{
		Width: 232, Height: -1, Gap: 4, Padding: giode.Inset{Top: 18, Left: 14, Right: 14, Bottom: 14},
		Background: "theme.surface", BorderWidth: 0,
	}, els...)
}

func logo() giode.Element {
	return giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Padding: giode.SymmetricInset(6, 0), Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{
			Width: 38, Height: 38, BorderRadius: 12, BackgroundGradient: theme.Brand(135),
			Align: giode.AlignCenter, Justify: giode.JustifyCenter,
			BoxShadow: giode.BoxShadow{Y: 6, Blur: 18, Color: "#8b5cf655"},
		}, icons.Icon(icons.Waves, 20, "#ffffff")),
		giode.Stack(giode.Styles{Gap: 1},
			giode.Text("Sonora", giode.Styles{Color: "theme.text", FontSize: 18, FontWeight: giode.FontWeightExtraBold}),
			giode.Text("Speech studio", theme.Caption),
		),
	)
}

func (s *Sidebar) engineSummary() giode.Element {
	ready := len(s.c.Manager.Ready())
	msg := fmt.Sprintf("%d engine%s ready", ready, plural(ready))
	if j := s.c.Manager.Job; j != nil {
		msg = "Generating with " + j.Runtime.Manifest().Name
	}
	for _, rt := range s.c.Manager.Engines() {
		if rt.Busy() && s.c.Manager.Job == nil {
			msg = "Installing " + rt.Manifest().Name
		}
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.UniformInset(12), Margin: giode.Inset{Bottom: 6},
		Background: "theme.text/0.04", BorderRadius: theme.RadiusMd},
		giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 99, Background: "theme.success"}),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(msg, giode.Styles{Color: "theme.secondary", FontSize: 12, MaxLines: 2})),
	)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
