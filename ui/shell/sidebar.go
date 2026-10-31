// Package shell is the window frame: title bar, sidebar navigation and
// the page area.
package shell

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"sync"

	"github.com/mehran9675/giode"
	"golang.org/x/image/draw"

	"github.com/mehran9675/sonora/assets"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/navitem"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
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
	nav(icons.Waves, tr.Msg("Studio"), core.PageStudio, nil)
	nav(icons.Captions, tr.Msg("Transcribe"), core.PageTranscribe, func() string {
		if n := len(c.Transcripts.Entries()); n > 0 {
			return fmt.Sprint(n)
		}
		return ""
	})
	nav(icons.Library, tr.Msg("Library"), core.PageLibrary, func() string {
		if n := len(c.Library.Entries()); n > 0 {
			return fmt.Sprint(n)
		}
		return ""
	})
	nav(icons.Chip, tr.Msg("Engines"), core.PageEngines, func() string {
		return fmt.Sprintf("%d/%d", len(c.Manager.Ready()), len(c.Manager.Engines()))
	})
	nav(icons.Sliders, tr.Msg("Settings"), core.PageSettings, nil)
	s.theme = giode.Pressable(func(ps giode.PressState) giode.Element {
		icon, label := icons.Sun, tr.T("Light mode")
		if !theme.IsDark() {
			icon, label = icons.Moon, tr.T("Dark mode")
		}
		fg := "theme.muted"
		bg := ""
		if ps.Hovered {
			fg, bg = "theme.text", "theme.text/0.05"
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 10), BorderRadius: theme.RadiusSm + 2, Background: bg},
			icons.Icon(icon, 18, fg), giode.Text(label, giode.Styles{Color: fg, FontSize: 13, FontWeight: giode.FontWeightMedium}))
	}).OnClick(func(giode.ClickEvent) { onTheme() })
	tr.Bind(func(l string) { s.theme.Label(l) }, "Toggle theme")
	return s
}

// Tabs lays out the same navigation as a bottom bar for narrow screens.
func (s *Sidebar) Tabs() giode.Element {
	tabs := make([]giode.Element, len(s.items))
	for i, it := range s.items {
		tabs[i] = it.Tab()
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 4, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(8, 6),
		Background: "theme.surface"}, tabs...)
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

// logoImage is the app icon, decoded once and shrunk for the sidebar so
// it stays crisp.
var logoImage = sync.OnceValue(func() image.Image {
	src, err := png.Decode(bytes.NewReader(assets.Icon))
	if err != nil {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	dst := image.NewNRGBA(image.Rect(0, 0, 128, 128))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
})

// logo is the app icon and name at the top of the sidebar.
func logo() giode.Element {
	return giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Padding: giode.SymmetricInset(6, 0), Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Width: 38, Height: 38, BorderRadius: 10,
			BoxShadow: giode.BoxShadow{Y: 6, Blur: 18, Color: "#8b5cf644"}},
			giode.Image(logoImage(), giode.Styles{Width: 38, Height: 38, Fit: giode.FitContain})),
		giode.Stack(giode.Styles{Gap: 1},
			giode.Text("Sonora", giode.Styles{Color: "theme.text", FontSize: 18, FontWeight: giode.FontWeightExtraBold}),
			giode.Text(tr.T("Speech studio"), theme.Caption),
		),
	)
}

func (s *Sidebar) engineSummary() giode.Element {
	ready := len(s.c.Manager.Ready())
	msg := tr.N("{n} engines ready", ready)
	m := s.c.Manager
	for _, rt := range m.Engines() {
		if rt.Busy() && rt.State == manager.StateInstalling {
			msg = tr.T("Installing {name}", "name", rt.Manifest().Name)
		}
	}
	if j := m.TranscribeJob; j != nil {
		msg = tr.T("Transcribing with {name}", "name", j.Runtime.Manifest().Name)
	}
	if j := m.Job; j != nil {
		msg = tr.T("Generating with {name}", "name", j.Runtime.Manifest().Name)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.UniformInset(12), Margin: giode.Inset{Bottom: 6},
		Background: "theme.text/0.04", BorderRadius: theme.RadiusMd},
		giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 99, Background: "theme.success"}),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(msg, giode.Styles{Color: "theme.secondary", FontSize: 12, MaxLines: 2})),
	)
}
