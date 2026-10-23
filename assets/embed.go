// Package assets embeds the fonts bundled with the app.
package assets

import "embed"

// Fonts holds Inter (SIL Open Font License, see fonts/OFL.txt).
//
//go:embed fonts/*.ttf
var Fonts embed.FS
