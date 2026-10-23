// Command makeicon draws Sonora's app icon: a violet-to-cyan rounded
// tile with white waveform bars, matching the logo in the sidebar.
//
//	go run ./tools/makeicon assets/icon
//
// writes icon.png (full bleed, for Windows and the window) and
// icon-macos.png (inset with a shadow, following macOS's icon grid).
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

const (
	size = 1024
	ss   = 4 // supersampling factor for smooth edges
)

var (
	from = color.NRGBA{0x8b, 0x5c, 0xf6, 0xff}
	to   = color.NRGBA{0x22, 0xd3, 0xee, 0xff}
	// bars are the waveform heights as fractions of the tile.
	bars = []float64{0.14, 0.30, 0.46, 0.26, 0.56, 0.36, 0.20}
)

func main() {
	out := "assets/icon"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	save(filepath.Join(out, "icon.png"), render(0, false))
	// macOS icons sit on an 824px body inside the 1024px canvas.
	save(filepath.Join(out, "icon-macos.png"), render(100, true))
}

// render draws the icon with the tile inset by margin pixels.
func render(margin int, shadow bool) image.Image {
	n := size * ss
	big := image.NewNRGBA(image.Rect(0, 0, n, n))
	m := float64(margin * ss)
	tile := float64(n) - 2*m
	radius := tile * 0.225
	inTile := func(x, y float64) bool { return inRoundRect(x-m, y-m, tile, tile, radius) }

	if shadow {
		// A soft drop shadow: darken pixels near and below the tile.
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				d := distToRoundRect(float64(x)-m, float64(y)-m-tile*0.02, tile, tile, radius)
				if d > 0 && d < tile*0.05 {
					a := 0.28 * (1 - d/(tile*0.05))
					big.SetNRGBA(x, y, color.NRGBA{0, 0, 0, uint8(255 * a * a)})
				}
			}
		}
	}

	barW := tile * 0.058
	gap := tile * 0.056
	total := float64(len(bars))*barW + float64(len(bars)-1)*gap
	left := m + (tile-total)/2
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			if !inTile(fx, fy) {
				continue
			}
			// 135° gradient: top-left violet to bottom-right cyan.
			t := ((fx - m) + (fy - m)) / (2 * tile)
			c := mix(from, to, t)
			// A gentle top highlight for depth.
			if h := 1 - (fy-m)/(tile*0.55); h > 0 {
				c = mix(c, color.NRGBA{255, 255, 255, 255}, 0.10*h)
			}
			for i, b := range bars {
				bx := left + float64(i)*(barW+gap)
				bh := tile * b
				by := m + (tile-bh)/2
				if inRoundRect(fx-bx, fy-by, barW, bh, barW/2) {
					c = color.NRGBA{255, 255, 255, 255}
				}
			}
			big.SetNRGBA(x, y, c)
		}
	}
	small := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(small, small.Bounds(), big, big.Bounds(), draw.Src, nil)
	return small
}

func inRoundRect(x, y, w, h, r float64) bool {
	return distToRoundRect(x, y, w, h, r) <= 0
}

// distToRoundRect is the signed distance from (x, y) to a rounded
// rectangle at the origin: negative inside.
func distToRoundRect(x, y, w, h, r float64) float64 {
	cx, cy := math.Abs(x-w/2)-(w/2-r), math.Abs(y-h/2)-(h/2-r)
	ox, oy := math.Max(cx, 0), math.Max(cy, 0)
	return math.Hypot(ox, oy) + math.Min(math.Max(cx, cy), 0) - r
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	t = math.Max(0, math.Min(1, t))
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.NRGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), l(a.A, b.A)}
}

func save(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
