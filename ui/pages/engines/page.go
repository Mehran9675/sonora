// Package engines is the Engines screen: every engine as a card with
// one-click install, logs and configuration, plus adding new ones.
package engines

import (
	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/logdialog"
	"github.com/mehran9675/sonora/ui/components/pagehead"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Page is the Engines screen. Create it once.
type Page struct {
	c      *core.Ctx
	scroll *scroll.Scroll
	cards  map[*manager.Runtime]*card
	add    *giode.ButtonElement
	adder  *addDialog
	config *configDialog
	logs   *logdialog.Dialog
}

// New builds the page.
func New(c *core.Ctx) *Page {
	p := &Page{c: c, cards: map[*manager.Runtime]*card{}}
	p.adder = newAddDialog(c)
	p.config = newConfigDialog(c)
	p.logs = logdialog.New()
	p.add = giode.Button("Add engine", giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(14, 9), BorderRadius: theme.RadiusSm + 2}).
		Icon(icons.Icon(icons.Plus, 16, "")).OnClick(p.adder.show)
	p.scroll = giode.Scroll(p.content, giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.15", Width: 6}})
	return p
}

func (p *Page) card(rt *manager.Runtime) *card {
	cd := p.cards[rt]
	if cd == nil {
		cd = newCard(p.c, rt, p.config, p.logs)
		p.cards[rt] = cd
	}
	return cd
}

func (p *Page) content() giode.Element {
	var ready, rest []giode.Element
	for _, rt := range p.c.Manager.Engines() {
		el := p.card(rt).element()
		if rt.State == manager.StateReady || rt.State == manager.StateWorking {
			ready = append(ready, el)
		} else {
			rest = append(rest, el)
		}
	}
	els := []giode.Element{
		pagehead.Head(giode.Compact(), "Engines", "Install text-to-speech models with one click, or connect your own.", p.add),
	}
	if len(ready) > 0 {
		els = append(els, group("Installed", ready))
	}
	if len(rest) > 0 {
		els = append(els, group("Available", rest))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 26, Padding: giode.UniformInset(core.Gutter())}, els...)
}

func group(title string, cards []giode.Element) giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Gap: 14},
		giode.Text(title, theme.Overline),
		giode.Row(giode.Styles{Width: -1, Gap: 16}, cards...),
	)
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.scroll)
}
