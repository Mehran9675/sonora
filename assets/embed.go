// Package assets embeds the fonts and icon bundled with the app.
package assets

import (
	"embed"
	_ "embed"
)

// Fonts holds Inter (SIL Open Font License, see fonts/OFL.txt).
//
//go:embed fonts/*.ttf
var Fonts embed.FS

// Icon is the app icon as a 1024×1024 PNG; tools/makeicon draws it.
//
//go:embed icon/icon.png
var Icon []byte
