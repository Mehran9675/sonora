// Package theme defines Sonora's look: palettes for dark and light
// mode, fonts, and the shared style tokens components build on.
package theme

import "github.com/mehran9675/giode"

// Dark is the default palette: deep ink surfaces with a violet-to-cyan
// accent.
var Dark = giode.Palette{
	Primary:    "#8b5cf6",
	Secondary:  "#a1a6b8",
	Background: "#0b0c12",
	Surface:    "#14161f",
	Text:       "#eceef6",
	Border:     "#252836",
	OnPrimary:  "#ffffff",
	Muted:      "#8a8fa6",
	Accent:     "#22d3ee",
	Success:    "#34d399",
	Warning:    "#fbbf24",
	Danger:     "#f87171",
	Info:       "#60a5fa",
}

// Light mirrors Dark on paper-white surfaces.
var Light = giode.Palette{
	Primary:    "#7c3aed",
	Secondary:  "#5b6075",
	Background: "#f4f5fa",
	Surface:    "#ffffff",
	Text:       "#13141c",
	Border:     "#e2e4ee",
	OnPrimary:  "#ffffff",
	Muted:      "#6b7088",
	Accent:     "#0891b2",
	Success:    "#059669",
	Warning:    "#d97706",
	Danger:     "#dc2626",
	Info:       "#2563eb",
}

// Apply installs both palettes and selects mode: "dark", "light", or
// "system" to follow the platform's appearance (dark where it can't be
// read).
func Apply(mode string) {
	giode.SetPalette(giode.ModeDark, Dark)
	giode.SetPalette(giode.ModeLight, Light)
	if mode == "system" {
		if dark, err := giode.SystemDarkMode(); err == nil && !dark {
			mode = "light"
		}
	}
	if mode == "light" {
		giode.SetMode(giode.ModeLight)
	} else {
		giode.SetMode(giode.ModeDark)
	}
}

// IsDark reports whether dark mode is active.
func IsDark() bool { return giode.CurrentMode() == giode.ModeDark }
