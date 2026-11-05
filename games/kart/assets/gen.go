//go:build ignore

// gen paints track.png, the kart circuit seen from above, from package course.
// Usage: go run gen.go, then run go generate in games/kart.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/sat0ken/ikaten/games/kart/course"
)

var (
	asphalt   = color.RGBA{96, 96, 108, 255}
	curbRed   = color.RGBA{220, 40, 40, 255}
	curbWhite = color.RGBA{235, 235, 235, 255}
	grassA    = color.RGBA{64, 172, 72, 255}
	grassB    = color.RGBA{52, 150, 62, 255}
	wall      = color.RGBA{70, 70, 80, 255}
	padYellow = color.RGBA{255, 210, 40, 255}
	padOrange = color.RGBA{255, 120, 20, 255}
)

// noise varies the asphalt a little so that motion is visible.
func noise(x, y int) int {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ h>>13) * 1274126177
	return int(h>>24)%9 - 4
}

func shade(c color.RGBA, d int) color.RGBA {
	f := func(v uint8) uint8 { return uint8(min(max(int(v)+d, 0), 255)) }
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

// wrapped returns the distance from a to b along a loop of the given length.
func wrapped(a, b, length float32) float32 {
	d := float32(math.Mod(float64(a-b), float64(length)))
	if d > length/2 {
		d -= length
	} else if d < -length/2 {
		d += length
	}
	return d
}

func texel(x, y int) color.RGBA {
	const edge = course.Wall - 2
	if x < edge || y < edge || x >= course.Size-edge || y >= course.Size-edge {
		return wall
	}
	w := course.Locate(float32(x)+.5, float32(y)+.5, 0, -1)
	switch {
	case w.Dist < course.HalfWidth:
		// The start line is a checkered band across the road at s = 0.
		if d := wrapped(w.S, 0, course.Length); d > -2 && d < 2 {
			if (int(math.Floor(float64(w.Offset)))/2+int(math.Floor(float64(d+2)))/2)%2 == 0 {
				return curbWhite
			}
			return color.RGBA{30, 30, 30, 255}
		}
		for _, p := range course.Pads {
			along := wrapped(w.S, p.S, course.Length)
			across := w.Offset - p.Offset
			if along > -course.PadLength/2 && along < course.PadLength/2 && across > -course.PadHalfWidth && across < course.PadHalfWidth {
				// Chevrons point in the direction of travel.
				if int(math.Floor(float64(along-float32(math.Abs(float64(across)))*0.8)))%4 < 2 {
					return padOrange
				}
				return padYellow
			}
		}
		return shade(asphalt, noise(x, y))
	case w.Dist < course.HalfWidth+course.CurbWidth:
		if int(w.S/3)%2 == 0 {
			return curbRed
		}
		return curbWhite
	}
	if (x/8+y/8)%2 == 0 {
		return grassA
	}
	return grassB
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, course.Size, course.Size))
	for y := range course.Size {
		for x := range course.Size {
			img.SetRGBA(x, y, texel(x, y))
		}
	}
	out, err := os.Create("track.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(out, img); err != nil {
		log.Fatal(err)
	}
	if err := out.Close(); err != nil {
		log.Fatal(err)
	}
}
