// Package text draws tinyfont fonts. It is intentionally not compatible with
// Ebitengine text/v2. Coordinates refer to the text baseline.
package text

import (
	"github.com/sat0ken/ikaten"
	"image/color"
	"tinygo.org/x/tinyfont"
)

type display struct{ target *ikaten.Image }

func (d *display) Size() (int16, int16)              { w, h := d.target.Size(); return int16(w), int16(h) }
func (d *display) SetPixel(x, y int16, c color.RGBA) { d.target.SetRGBA(int(x), int(y), c) }
func (d *display) Display() error                    { return nil }

// Renderer keeps the tinyfont adapter outside the heap allocation hot path.
// Like Image, it is used from one goroutine.
type Renderer struct{ display display }

var defaultRenderer Renderer

func Draw(dst *ikaten.Image, str string, font tinyfont.Fonter, x, y int, clr color.Color) {
	defaultRenderer.Draw(dst, str, font, x, y, clr)
}
func (r *Renderer) Draw(dst *ikaten.Image, str string, font tinyfont.Fonter, x, y int, clr color.Color) {
	r.display.target = dst
	red, green, blue, alpha := clr.RGBA()
	c := color.RGBA{R: byte(red >> 8), G: byte(green >> 8), B: byte(blue >> 8), A: byte(alpha >> 8)}
	origin := x
	for _, char := range str {
		if char == '\n' {
			x = origin
			y += int(font.GetYAdvance())
			continue
		}
		if char == '\r' {
			x = origin
			continue
		}
		glyph := font.GetGlyph(char)
		glyph.Draw(&r.display, int16(x), int16(y), c)
		x += int(glyph.Info().XAdvance)
	}
	r.display.target = nil
}
