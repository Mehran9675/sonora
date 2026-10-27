// Package core holds the state shared by every page: the backend
// services, navigation, notifications and the now-playing clip.
package core

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/audio/playback"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/components/actionsheet"
)

// Ctx is passed to every page and component that needs app services.
type Ctx struct {
	App      *giode.App
	Manager  *manager.Manager
	Settings *settings.Store
	Library  *library.Library
	// Transcripts holds every transcription.
	Transcripts *transcripts.Store
	Player      *playback.Player
	Toasts      *giode.ToasterElement
	// Sheet lists actions for an item (export, share).
	Sheet *actionsheet.Sheet
	Now   *NowPlaying

	page Page
	// awake counts StayAwake holders; release ends the KeepAlive.
	awake   int
	release func()
	// UseEngine selects an engine in the Studio (set by the Studio).
	UseEngine func(id string)
	// Reuse loads a library entry's text and settings into the Studio.
	Reuse func(e library.Entry)
	// UseTranscriber selects a speech-to-text engine on the Transcribe
	// page (set by that page).
	UseTranscriber func(id string)
	// Speak puts text into the Studio's script (set by the Studio).
	Speak func(text string)
	// Transcribe opens an audio file on the Transcribe page (set by that
	// page).
	Transcribe func(path, title string)
	// ShowEngines opens the Engines page on the voice or transcription
	// tab (set by that page).
	ShowEngines func(task engines.Task)
}

// Page returns the visible page.
func (c *Ctx) Page() Page { return c.page }

// Go switches pages.
func (c *Ctx) Go(p Page) { c.page = p }

// Gutter is the page padding: tighter in the compact layout.
func Gutter() int {
	if giode.Compact() {
		return 16
	}
	return 28
}
