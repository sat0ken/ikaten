//go:build ignore

// gen halves the flappy images from Ebitengine's examples/resources/images/flappy.
// Usage: go run gen.go <source directory>, then run go generate in games/flappy.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func half(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx()/2, b.Dy()/2))
	for y := 0; y < b.Dy()/2; y++ {
		for x := 0; x < b.Dx()/2; x++ {
			var r, g, bl, a uint32
			for _, o := range [4][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
				c := color.NRGBAModel.Convert(src.At(b.Min.X+2*x+o[0], b.Min.Y+2*y+o[1])).(color.NRGBA)
				w := uint32(c.A)
				r, g, bl, a = r+uint32(c.R)*w, g+uint32(c.G)*w, bl+uint32(c.B)*w, a+w
			}
			if a > 0 {
				dst.SetNRGBA(x, y, color.NRGBA{uint8(r / a), uint8(g / a), uint8(bl / a), uint8(a / 4)})
			}
		}
	}
	return dst
}

func main() {
	for _, name := range []string{"gopher.png", "tiles.png"} {
		in, err := os.Open(filepath.Join(os.Args[1], name))
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(in)
		in.Close()
		if err != nil {
			log.Fatal(err)
		}
		out, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, half(img)); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
}
