// Command makeicon draws Sonora's app icon: an S for Sonora made of
// three parallel sound-wave strands, the top bowl turned one way and the
// bottom the other - speech going out and coming in - glowing violet to
// cyan on a deep navy tile.
//
//	go run ./tools/makeicon assets/icon
//
// writes icon.png (full bleed, for Windows, Linux and the window),
// icon-macos.png (inset with a shadow, on macOS's icon grid) and the
// layers of Android's adaptive icon: android-foreground.png (the S on a
// transparent 108dp canvas, inside the middle 66dp) and
// android-monochrome.png (its shape, tinted for themed icons). The
// background layer is the navy color set in giode.toml.
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
	ss   = 3 // supersampling factor for smooth edges
)

var (
	violet = color.NRGBA{0x8b, 0x5c, 0xf6, 0xff}
	cyan   = color.NRGBA{0x22, 0xd3, 0xee, 0xff}
	white  = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	// The tile: deep navy, a little lighter at the top.
	navyTop    = color.NRGBA{0x1e, 0x1b, 0x4b, 0xff}
	navyBottom = color.NRGBA{0x0b, 0x10, 0x26, 0xff}
)

func main() {
	out := "assets/icon"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	save(filepath.Join(out, "icon.png"), render(tile{margin: 0}))
	// macOS icons sit on an 824px body inside the 1024px canvas.
	save(filepath.Join(out, "icon-macos.png"), render(tile{margin: 100, shadow: true}))
	// Android shows the middle 66 of the layer's 108dp: the S keeps to
	// it, and the launcher draws the navy background and the shape.
	save(filepath.Join(out, "android-foreground.png"), render(tile{none: true, scale: 66.0 / 108}))
	save(filepath.Join(out, "android-monochrome.png"), render(tile{none: true, scale: 66.0 / 108, mono: true}))
}

// tile says how to draw: the tile inset by margin pixels (with a drop
// shadow), or no tile at all with the S scaled by scale, in color or as
// a white shape.
type tile struct {
	margin int
	shadow bool
	none   bool
	scale  float64
	mono   bool
}

// strand is one stroke of the S: two arcs of concentric ellipses, the
// top bowl and the bottom one, joined in the middle.
type strand struct {
	top, bottom arc
	width       float64
}

// arc is part of an ellipse centred at (cx, cy) with radii rx, ry, from
// angle a0 to a1 (radians, clockwise from 3 o'clock, y pointing down).
type arc struct {
	cx, cy, rx, ry, a0, a1 float64
}

// strands lays out the S in the unit square.
func strands() []strand {
	const (
		n     = 3
		width = 0.052 // stroke width
		step  = 0.083 // from one strand to the next
		inner = 0.050 // radius of the innermost strand
		sx    = 1.32  // the bowls are wider than tall
	)
	outer := inner + step*(n-1)
	// Strand i of the top bowl meets strand n-1-i of the bottom one in
	// the middle, so the bowls' centres are inner+outer apart.
	gap := inner + outer
	total := gap + 2*outer
	top := 0.5 - total/2 + outer
	bottom := top + gap
	var out []strand
	for i := 0; i < n; i++ {
		rt := inner + step*float64(i)
		rb := inner + step*float64(n-1-i)
		out = append(out, strand{
			// The top bowl runs from its foot in the middle, round the
			// left and over the top, to the upper right.
			top: arc{0.5, top, rt * sx, rt, deg(90), deg(330)},
			// The bottom bowl runs from its head in the middle, round the
			// right and under the bottom, to the lower left.
			bottom: arc{0.5, bottom, rb * sx, rb, deg(-90), deg(150)},
			width:  width,
		})
	}
	return out
}

func deg(d float64) float64 { return d * math.Pi / 180 }

// render draws the icon.
func render(t tile) image.Image {
	n := size * ss
	big := image.NewNRGBA(image.Rect(0, 0, n, n))
	m := float64(t.margin * ss)
	side := float64(n) - 2*m
	radius := side * 0.225
	// The S's unit square: the tile, or the middle of an empty canvas.
	art, ax := side, m
	if t.none {
		art = float64(n) * t.scale
		ax = (float64(n) - art) / 2
	}
	ss := strands()
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			var c color.NRGBA
			if !t.none {
				d := distToRoundRect(fx-m, fy-m, side, side, radius)
				if t.shadow && d > 0 {
					// A soft drop shadow below the tile.
					ds := distToRoundRect(fx-m, fy-m-side*0.02, side, side, radius)
					if ds > 0 && ds < side*0.05 {
						a := 0.30 * (1 - ds/(side*0.05))
						big.SetNRGBA(x, y, color.NRGBA{0, 0, 0, uint8(255 * a * a)})
					}
					continue
				}
				if d > 0 {
					continue
				}
				c = mix(navyTop, navyBottom, (fy-m)/side)
			}
			// The S, in the art square's unit coordinates.
			u, v := (fx-ax)/art, (fy-ax)/art
			dist, along := math.Inf(1), 0.0
			for _, s := range ss {
				if d, a := s.dist(u, v); d < dist {
					dist, along = d, a
				}
			}
			px := 1 / art // one pixel in unit coordinates, for antialiasing
			cover := clamp(0.5-dist/px, 0, 1)
			ink := mix(violet, cyan, along)
			// A highlight down the middle of each stroke.
			if dist < 0 {
				ink = mix(ink, white, 0.22*clamp(1+dist/(ss[0].width/2), 0, 1))
			}
			switch {
			case t.mono:
				c = color.NRGBA{255, 255, 255, uint8(255 * cover)}
			case t.none:
				c = ink
				c.A = uint8(255 * cover)
			default:
				// A faint glow around the strokes, then the strokes.
				if glow := clamp(1-dist/0.06, 0, 1); glow > 0 && dist > 0 {
					c = mix(c, ink, 0.22*glow*glow)
				}
				c = mix(c, ink, cover)
			}
			big.SetNRGBA(x, y, c)
		}
	}
	small := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(small, small.Bounds(), big, big.Bounds(), draw.Src, nil)
	return small
}

// dist is the signed distance from (u, v) to the strand's edge
// (negative inside), and how far along the S the nearest point is, 0
// at its top end to 1 at its bottom end.
func (s strand) dist(u, v float64) (float64, float64) {
	dt, at := s.top.dist(u, v)
	db, ab := s.bottom.dist(u, v)
	if dt < db {
		// The top bowl is drawn from the middle backwards: its far end
		// is the S's start.
		return dt - s.width/2, 0.5 * (1 - at)
	}
	return db - s.width/2, 0.5 + 0.5*ab
}

// dist is the distance from (u, v) to the arc's centre line and the
// fraction of the arc where the nearest point lies. In the ellipse's
// squashed (circular) coordinates the radial gap is |ρ - ry|; dividing
// by the gradient's length turns it into the distance across the
// ellipse, exact enough near the stroke.
func (a arc) dist(u, v float64) (float64, float64) {
	k := a.rx / a.ry
	x, y := (u-a.cx)/k, v-a.cy
	t := math.Atan2(y, x)
	for t < a.a0 {
		t += 2 * math.Pi
	}
	for t > a.a0+2*math.Pi {
		t -= 2 * math.Pi
	}
	if t <= a.a1 {
		grad := math.Hypot(math.Cos(t)/k, math.Sin(t))
		return math.Abs(math.Hypot(x, y)-a.ry) / grad, (t - a.a0) / (a.a1 - a.a0)
	}
	// Past an end: the distance to the nearer end point (a round cap).
	end := func(t float64) float64 {
		return math.Hypot(u-(a.cx+a.rx*math.Cos(t)), v-(a.cy+a.ry*math.Sin(t)))
	}
	if d0, d1 := end(a.a0), end(a.a1); d0 < d1 {
		return d0, 0
	} else {
		return d1, 1
	}
}

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

// distToRoundRect is the signed distance from (x, y) to a rounded
// rectangle at the origin: negative inside.
func distToRoundRect(x, y, w, h, r float64) float64 {
	cx, cy := math.Abs(x-w/2)-(w/2-r), math.Abs(y-h/2)-(h/2-r)
	ox, oy := math.Max(cx, 0), math.Max(cy, 0)
	return math.Hypot(ox, oy) + math.Min(math.Max(cx, cy), 0) - r
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	t = clamp(t, 0, 1)
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
