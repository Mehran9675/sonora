package shell

import (
	"runtime"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
)

// compactBelow is the width, in dp, up to which the app uses the
// one-column phone layout with a bottom tab bar. Inside it,
// giode.Compact() reports true, which is how pages and components
// adapt.
const compactBelow = 760

// Shell lays out the frame around the current page.
type Shell struct {
	c       *core.Ctx
	bar     giode.Element // nil on mobile, where the OS draws no window frame
	sidebar *Sidebar
	pages   map[core.Page]func() giode.Element
}

// New builds the shell; pages maps each page to its view function.
func New(c *core.Ctx, pages map[core.Page]func() giode.Element, onTheme func()) *Shell {
	s := &Shell{c: c, sidebar: NewSidebar(c, onTheme), pages: pages}
	if !core.Mobile {
		bar := giode.TitleBar(c.App, "Sonora", giode.Styles{Background: "theme.surface", Color: "theme.muted"}).
			Icon(icons.Icon(icons.Waves, 14, "theme.primary"))
		if runtime.GOOS == "darwin" {
			bar.Variant(giode.TitleBarMac)
		}
		s.bar = bar
	}
	return s
}

// Element lays out the whole window, choosing the layout by width.
func (s *Shell) Element() giode.Element {
	return giode.Responsive(
		giode.Breakpoint{MaxWidth: compactBelow, Name: "compact", View: s.compact},
		giode.Breakpoint{View: s.wide},
	)
}

// wide is the desktop layout: sidebar beside the page.
func (s *Shell) wide() giode.Element {
	return s.frame(giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Wrap: giode.WrapNoWrap},
		s.sidebar.Element(),
		giode.Box(giode.Styles{Width: 1, Height: -1, Background: "theme.border"}),
		giode.Box(giode.Styles{FlexGrow: 1, Height: -1, Background: "theme.background"}, s.page()),
	))
}

// compact is the phone layout: the page above a bottom tab bar.
func (s *Shell) compact() giode.Element {
	return s.frame(giode.Stack(giode.Styles{Width: -1, FlexGrow: 1},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Background: "theme.background"}, s.page()),
		giode.Box(giode.Styles{Width: -1, Height: 1, Background: "theme.border"}),
		s.sidebar.Tabs(),
	))
}

// page builds the current page. Back on any page but the Studio returns
// to the Studio; it's registered first, so a page or an open dialog
// that handles Back itself takes precedence.
func (s *Shell) page() giode.Element {
	if s.c.Page() != core.PageStudio {
		giode.HandleBack(func() { s.c.Go(core.PageStudio) })
	}
	return s.pages[s.c.Page()]()
}

// frame adds the title bar and the layers that cover the whole window.
func (s *Shell) frame(body giode.Element) giode.Element {
	main := []giode.Element{}
	if s.bar != nil {
		main = append(main, s.bar)
	}
	main = append(main, body)
	layers := []giode.Element{giode.Stack(giode.Styles{Width: -1, Height: -1}, main...)}
	// Toasts and modal layers cover the whole window, above the page.
	for _, el := range append([]giode.Element{s.c.Toasts}, overlay.All()...) {
		layers = append(layers, giode.Box(giode.Styles{
			Position: giode.PositionAbsolute, Top: 0, Left: 0, Width: -1, Height: -1, ZIndex: 10,
		}, el))
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Background: "theme.background"}, layers...)
}
