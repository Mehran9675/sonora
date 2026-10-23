// Package app assembles Sonora: services, pages and the window.
package app

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/assets"
	"github.com/mehran9675/sonora/internal/audio/playback"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/core"
	enginespage "github.com/mehran9675/sonora/ui/pages/engines"
	librarypage "github.com/mehran9675/sonora/ui/pages/library"
	settingspage "github.com/mehran9675/sonora/ui/pages/settings"
	"github.com/mehran9675/sonora/ui/pages/studio"
	"github.com/mehran9675/sonora/ui/shell"
	"github.com/mehran9675/sonora/ui/theme"
)

// Run opens the window and blocks until it closes.
func Run() error {
	theme.LoadFonts()
	store := settings.Open()
	theme.Apply(store.Get().Theme)

	c := &core.Ctx{
		Settings: store,
		Library:  library.Open(paths.Sub("library")),
		Player:   playback.New(),
		Toasts:   giode.Toaster().Placement(giode.ToastBottomRight),
	}
	var studioPage *studio.Page
	cfg := giode.Config{
		Title:      "Sonora",
		Icon:       assets.Icon,
		Size:       image.Pt(1280, 820),
		MinWidth:   360, // narrow windows get the phone layout
		MinHeight:  560,
		Resizable:  true,
		Frameless:  true,
		Center:     true,
		Background: "theme.background",
		Logger:     log.Default(),
		LogLevel:   giode.LogLevelWarning,
		OnShutdown: func() {
			if studioPage != nil {
				studioPage.SaveDraft()
			}
			c.Manager.Shutdown()
		},
	}
	// SONORA_WINDOW_SIZE=400x840 starts with that size, to try the
	// phone layout on a desktop.
	if w, h, ok := parseSize(os.Getenv("SONORA_WINDOW_SIZE")); ok {
		cfg.Size = image.Pt(w, h)
	}
	if core.Mobile {
		// The OS owns the window and runs one instance.
		cfg.Frameless, cfg.Center = false, false
	} else {
		cfg.SingleInstanceLock = &giode.SingleInstanceLock{UniqueID: "com.sonora.app"}
	}
	c.App = giode.New(cfg)
	c.Manager = manager.New(c.App.Post, store, c.Library)

	applyTheme := func(mode string) {
		store.Update(func(s *settings.Settings) { s.Theme = mode })
		theme.Apply(mode)
		giode.InvalidateCaches()
	}
	toggleTheme := func() {
		if theme.IsDark() {
			applyTheme("light")
		} else {
			applyTheme("dark")
		}
	}

	studioPage = studio.New(c)
	libraryPage := librarypage.New(c)
	enginesPage := enginespage.New(c)
	settingsPage := settingspage.New(c, applyTheme)
	frame := shell.New(c, map[core.Page]func() giode.Element{
		core.PageStudio:   studioPage.Element,
		core.PageLibrary:  libraryPage.Element,
		core.PageEngines:  enginesPage.Element,
		core.PageSettings: settingsPage.Element,
	}, toggleTheme)

	return c.App.Run(frame.Element)
}

// parseSize reads "WIDTHxHEIGHT".
func parseSize(v string) (w, h int, ok bool) {
	if _, err := fmt.Sscanf(v, "%dx%d", &w, &h); err != nil || w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}
