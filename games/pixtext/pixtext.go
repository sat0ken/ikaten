// Package pixtext draws lines of a small 5 by 7 pixel font for the games made
// for a 240 by 240 screen. A Line is built in a fixed buffer, so that drawing
// a score makes no strings and allocates nothing.
package pixtext

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// The font's size in pixels at scale 1. Letters are Advance apart.
const (
	Width   = 5
	Height  = 7
	Advance = Width + 1
)

// glyphIndex maps a letter to its place in glyphs, plus one; 0 is none.
var glyphIndex [128]uint8

// A font is the glyphs drawn in one color, each cut out of one image.
type font struct {
	img   *ebiten.Image
	glyph [len(glyphs)]*ebiten.Image
	c     color.RGBA
	used  int
}

// white is drawn tinted for colors that are not opaque. The others come
// from fonts, which keep the colors last used: copying a glyph of the
// right color is much faster than shading a white one pixel by pixel.
var (
	white    font
	fonts    [6]font
	fontTick int
)

func init() {
	for i := range glyphs {
		glyphIndex[glyphs[i].ch] = uint8(i + 1)
	}
	newFont(&white)
	white.paint(color.RGBA{255, 255, 255, 255})
	for i := range fonts {
		newFont(&fonts[i])
	}
}

func newFont(f *font) {
	f.img = ebiten.NewImage(len(glyphs)*Advance, Height)
	// SubImage allocates, so the glyphs are cut once here.
	for i := range glyphs {
		r := image.Rect(i*Advance, 0, i*Advance+Width, Height)
		f.glyph[i] = f.img.SubImage(r).(*ebiten.Image)
	}
}

// paint draws the glyphs anew in c.
func (f *font) paint(c color.RGBA) {
	f.c = c
	f.img.Clear()
	var ink color.Color = c // made once, not for every pixel
	for i := range glyphs {
		for y, row := range glyphs[i].rows {
			for x := range len(row) {
				if row[x] == '#' {
					f.img.Set(i*Advance+x, y, ink)
				}
			}
		}
	}
}

// fontFor returns the glyphs in c, painting over the font used longest ago
// if no font has c.
func fontFor(c color.RGBA) *font {
	fontTick++
	oldest := &fonts[0]
	for i := range fonts {
		f := &fonts[i]
		if f.used > 0 && f.c == c {
			f.used = fontTick
			return f
		}
		if f.used < oldest.used {
			oldest = f
		}
	}
	oldest.paint(c)
	oldest.used = fontTick
	return oldest
}

// Line is a line of text to draw. The zero value is empty.
type Line struct {
	b      [48]byte
	n      int
	shadow color.RGBA
}

// Reset empties the line and drops its shadow.
func (l *Line) Reset() *Line {
	l.n = 0
	l.shadow = color.RGBA{}
	return l
}

// Str adds s. Small letters are drawn as capitals, apart from x.
func (l *Line) Str(s string) *Line {
	for i := range len(s) {
		l.add(s[i])
	}
	return l
}

func (l *Line) add(ch byte) {
	if ch >= 'a' && ch <= 'z' && ch != 'x' {
		ch -= 'a' - 'A'
	}
	if l.n < len(l.b) {
		l.b[l.n] = ch
		l.n++
	}
}

// Num adds n in decimal.
func (l *Line) Num(n int) *Line { return l.Pad(n, 1) }

// Pad adds n in decimal with zeros in front to at least digits digits.
func (l *Line) Pad(n, digits int) *Line {
	if n < 0 {
		l.add('-')
		n = -n
	}
	var d [20]byte
	i := len(d)
	for n > 0 || len(d)-i < digits {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
	}
	for ; i < len(d); i++ {
		l.add(d[i])
	}
	return l
}

// Shade gives the letters a shadow of color c, one pixel at the scale down
// and to the right, so that they read over a busy picture.
func (l *Line) Shade(c color.RGBA) *Line {
	l.shadow = c
	return l
}

// Len is the number of letters.
func (l *Line) Len() int { return l.n }

// Width is the line's width in pixels at scale.
func (l *Line) Width(scale int) int {
	if l.n == 0 {
		return 0
	}
	return (l.n*Advance - 1) * scale
}

// Draw draws the line with its top left at x, y.
func (l *Line) Draw(dst *ebiten.Image, x, y, scale int, c color.RGBA) {
	if l.shadow.A != 0 {
		l.draw(dst, x+scale, y+scale, scale, l.shadow)
	}
	l.draw(dst, x, y, scale, c)
}

// Right draws the line with its right edge at x.
func (l *Line) Right(dst *ebiten.Image, x, y, scale int, c color.RGBA) {
	l.Draw(dst, x-l.Width(scale), y, scale, c)
}

// Center draws the line centered on x.
func (l *Line) Center(dst *ebiten.Image, x, y, scale int, c color.RGBA) {
	l.Draw(dst, x-l.Width(scale)/2, y, scale, c)
}

func (l *Line) draw(dst *ebiten.Image, x, y, scale int, c color.RGBA) {
	f, tint := &white, c.A != 255
	if !tint {
		f = fontFor(c)
	}
	var op ebiten.DrawImageOptions
	for i := range l.n {
		ch := l.b[i]
		if ch >= 128 || glyphIndex[ch] == 0 {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x+i*Advance*scale), float64(y))
		if tint {
			op.ColorScale.Reset()
			op.ColorScale.Scale(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, float32(c.A)/255)
		}
		dst.DrawImage(f.glyph[glyphIndex[ch]-1], &op)
	}
}
