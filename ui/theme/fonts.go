package theme

import (
	"log"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/sonora/assets"
)

// Mono is the typeface used for logs and paths.
const Mono = "Go Mono"

// LoadFonts registers Inter for the interface and Go Mono for logs.
// Inter is registered first, so it is the default typeface.
func LoadFonts() {
	faces := []struct {
		file   string
		weight font.Weight
		style  font.Style
	}{
		{"Inter-Regular.ttf", font.Normal, font.Regular},
		{"Inter-Medium.ttf", font.Medium, font.Regular},
		// giode's "bold" maps to SemiBold, which reads cleaner at UI
		// sizes; "extrabold" gets the real Bold.
		{"Inter-SemiBold.ttf", font.Bold, font.Regular},
		{"Inter-Bold.ttf", font.ExtraBold, font.Regular},
		{"Inter-Italic.ttf", font.Normal, font.Italic},
	}
	var coll []font.FontFace
	for _, f := range faces {
		data, err := assets.Fonts.ReadFile("fonts/" + f.file)
		if err != nil {
			log.Printf("font %s: %v", f.file, err)
			continue
		}
		face, err := opentype.Parse(data)
		if err != nil {
			log.Printf("font %s: %v", f.file, err)
			continue
		}
		coll = append(coll, font.FontFace{
			Font: font.Font{Typeface: "Inter", Weight: f.weight, Style: f.style},
			Face: face,
		})
	}
	for _, ff := range gofont.Collection() {
		if ff.Font.Typeface == Mono {
			coll = append(coll, ff)
		}
	}
	giode.RegisterFonts(coll)
}
