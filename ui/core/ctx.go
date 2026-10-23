// Package core holds the state shared by every page: the backend
// services, navigation, notifications and the now-playing clip.
package core

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/audio/playback"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
)

// Ctx is passed to every page and component that needs app services.
type Ctx struct {
	App      *giode.App
	Manager  *manager.Manager
	Settings *settings.Store
	Library  *library.Library
	Player   *playback.Player
	Toasts   *giode.ToasterElement
	Now      *NowPlaying

	page    Page
	compact bool
	// UseEngine selects an engine in the Studio (set by the Studio).
	UseEngine func(id string)
	// Reuse loads a library entry's text and settings into the Studio.
	Reuse func(e library.Entry)
}

// Page returns the visible page.
func (c *Ctx) Page() Page { return c.page }

// Go switches pages.
func (c *Ctx) Go(p Page) { c.page = p }

// Compact reports whether the window is narrow (a phone, or a small
// desktop window), so pages stack their parts in one column. The shell
// sets it each frame from the available width.
func (c *Ctx) Compact() bool { return c.compact }

// SetCompact is called by the shell before laying out a page.
func (c *Ctx) SetCompact(v bool) { c.compact = v }

// Gutter is the page padding: tighter on narrow screens.
func (c *Ctx) Gutter() int {
	if c.compact {
		return 16
	}
	return 28
}
