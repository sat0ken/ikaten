package pixtext

import (
	"image/color"
	"testing"

	ebiten "github.com/sat0ken/ikaten"
)

func TestGlyphs(t *testing.T) {
	seen := map[byte]bool{}
	for _, g := range glyphs {
		if seen[g.ch] {
			t.Errorf("%q twice", g.ch)
		}
		seen[g.ch] = true
		ink := false
		for _, row := range g.rows {
			if len(row) != Width {
				t.Errorf("%q: row %q is not %d wide", g.ch, row, Width)
			}
			for i := range len(row) {
				switch row[i] {
				case '#':
					ink = true
				case '.':
				default:
					t.Errorf("%q: stray %q", g.ch, row[i])
				}
			}
		}
		if !ink {
			t.Errorf("%q has no ink", g.ch)
		}
	}
	for _, ch := range "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ:!?.-" {
		if glyphIndex[ch] == 0 {
			t.Errorf("no glyph for %q", ch)
		}
	}
}

func TestLine(t *testing.T) {
	var l Line
	check := func(line *Line, want string) {
		t.Helper()
		if got := string(line.b[:line.n]); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
	check(l.Reset().Str("Score ").Num(0), "SCORE 0")
	check(l.Reset().Num(-42), "-42")
	check(l.Reset().Pad(7, 2).Str(":").Pad(5, 2), "07:05")
	check(l.Reset().Pad(1234, 2), "1234")
	check(l.Reset().Str("x").Num(3), "x3")
	if w := l.Reset().Str("AB").Width(2); w != 22 {
		t.Errorf("width %d, want 22", w)
	}
	if w := l.Reset().Width(1); w != 0 {
		t.Errorf("empty width %d", w)
	}
}

func TestDraw(t *testing.T) {
	dst := ebiten.NewImage(40, 20)
	red := color.RGBA{255, 0, 0, 255}
	var l Line
	l.Reset().Str("1").Shade(color.RGBA{0, 0, 255, 255}).Draw(dst, 0, 0, 2, red)
	// The 1's foot is a bar from x=1 to 3, on row 6: at scale 2, pixels 2..7
	// on rows 12 and 13. The shadow is two pixels lower and to the right.
	if got := dst.RGBAAt(4, 12); got != red {
		t.Errorf("ink %v", got)
	}
	if got := dst.RGBAAt(8, 14); got.B < 200 || got.R > 0 {
		t.Errorf("shadow %v", got)
	}
	if got := dst.RGBAAt(0, 0); got.A != 0 {
		t.Errorf("blank %v", got)
	}
	l.Reset().Str("10").Right(dst, 40, 0, 1, red)
	if got := dst.RGBAAt(39, 1); got != red { // the 0's right side
		t.Errorf("right aligned %v", got)
	}
	// More colors than fonts: each line still comes out in its own.
	for i := range 2 * len(fonts) {
		c := color.RGBA{uint8(i * 20), 255, 0, 255}
		l.Reset().Str("8").Draw(dst, 0, 0, 1, c)
		if got := dst.RGBAAt(1, 0); int(got.R)-int(c.R) > 7 || int(c.R)-int(got.R) > 7 || got.G < 250 {
			t.Fatalf("color %d: %v", i, got)
		}
	}
	// A see-through color is drawn tinted.
	l.Reset().Str("8").Draw(dst, 0, 0, 1, color.RGBA{0, 0, 200, 200})
	if got := dst.RGBAAt(1, 0); got.B < 150 {
		t.Errorf("tinted %v", got)
	}
	allocs := testing.AllocsPerRun(10, func() {
		l.Reset().Str("SCORE ").Num(12345).Center(dst, 20, 0, 1, red)
	})
	if allocs != 0 {
		t.Errorf("%v allocations", allocs)
	}
}
