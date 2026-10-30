// Package app assembles Sonora: services, pages and the window.
package app

import (
	"cmp"
	"image"
	"log"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/assets"
	"github.com/mehran9675/sonora/internal/audio/playback"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/components/actionsheet"
	"github.com/mehran9675/sonora/ui/core"
	enginespage "github.com/mehran9675/sonora/ui/pages/engines"
	librarypage "github.com/mehran9675/sonora/ui/pages/library"
	settingspage "github.com/mehran9675/sonora/ui/pages/settings"
	"github.com/mehran9675/sonora/ui/pages/studio"
	"github.com/mehran9675/sonora/ui/pages/transcribe"
	"github.com/mehran9675/sonora/ui/shell"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

// Run opens the window and blocks until it closes.
func Run() error {
	theme.LoadFonts()
	store := settings.Open()
	theme.Apply(store.Get().Theme)

	c := &core.Ctx{
		Settings:    store,
		Library:     library.Open(paths.Sub("library")),
		Transcripts: transcripts.Open(paths.Sub("transcripts")),
		Player:      playback.New(),
		Toasts:      giode.Toaster().Placement(giode.ToastBottomRight),
		Sheet:       actionsheet.New(),
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
		// The chosen language, or the system's when Sonora has it.
		Locale:   cmp.Or(store.Get().UILanguage, tr.System()),
		Logger:   log.Default(),
		LogLevel: giode.LogLevelWarning,
		// Android may end a backgrounded app without warning, so the
		// draft is saved whenever the app leaves the screen.
		OnBackground: func() {
			core.SetAway(true)
			if studioPage != nil {
				studioPage.SaveDraft()
			}
		},
		OnForeground: func() { core.SetAway(false) },
		// With the System theme, follow the platform's appearance as it
		// changes.
		OnSystemThemeChange: func(bool) {
			if store.Get().Theme == "system" {
				theme.Apply("system")
				giode.InvalidateCaches()
			}
		},
		OnShutdown: func() {
			if studioPage != nil {
				studioPage.SaveDraft()
			}
			c.Manager.Shutdown()
		},
	}
	if core.Mobile {
		// The OS owns the window and runs one instance.
		cfg.Frameless, cfg.Center = false, false
	} else {
		cfg.SingleInstanceLock = &giode.SingleInstanceLock{UniqueID: "com.sonora.app"}
	}
	c.App = giode.New(cfg)
	c.Manager = manager.New(c.App.Post, store, c.Library, c.Transcripts)

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
	transcribePage := transcribe.New(c)
	libraryPage := librarypage.New(c)
	enginesPage := enginespage.New(c)
	settingsPage := settingspage.New(c, applyTheme)
	frame := shell.New(c, map[core.Page]func() giode.Element{
		core.PageStudio:     studioPage.Element,
		core.PageTranscribe: transcribePage.Element,
		core.PageLibrary:    libraryPage.Element,
		core.PageEngines:    enginesPage.Element,
		core.PageSettings:   settingsPage.Element,
	}, toggleTheme)

	return c.App.Run(frame.Element)
}
