package shell

import (
	"runtime"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
)

// Shell lays out the frame around the current page.
type Shell struct {
	c       *core.Ctx
	bar     giode.Element
	sidebar *Sidebar
	pages   map[core.Page]func() giode.Element
}

// New builds the shell; pages maps each page to its view function.
func New(c *core.Ctx, pages map[core.Page]func() giode.Element, onTheme func()) *Shell {
	bar := giode.TitleBar(c.App, "Sonora", giode.Styles{Background: "theme.surface", Color: "theme.muted"}).
		Icon(icons.Icon(icons.Waves, 14, "theme.primary"))
	if runtime.GOOS == "darwin" {
		bar.Variant(giode.TitleBarMac)
	}
	return &Shell{c: c, bar: bar, sidebar: NewSidebar(c, onTheme), pages: pages}
}

// Element lays out the whole window.
func (s *Shell) Element() giode.Element {
	page := s.pages[s.c.Page()]()
	layers := []giode.Element{
		giode.Stack(giode.Styles{Width: -1, Height: -1},
			s.bar,
			giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Wrap: giode.WrapNoWrap},
				s.sidebar.Element(),
				giode.Box(giode.Styles{Width: 1, Height: -1, Background: "theme.border"}),
				giode.Box(giode.Styles{FlexGrow: 1, Height: -1, Background: "theme.background"}, page),
			),
		),
	}
	// Toasts and modal layers cover the whole window, above the page.
	for _, el := range append([]giode.Element{s.c.Toasts}, overlay.All()...) {
		layers = append(layers, giode.Box(giode.Styles{
			Position: giode.PositionAbsolute, Top: 0, Left: 0, Width: -1, Height: -1, ZIndex: 10,
		}, el))
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Background: "theme.background"}, layers...)
}
