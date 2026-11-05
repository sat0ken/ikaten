//go:build ignore

// gen draws the large uboat images: the sea and sky strips, the periscope's
// mask and the sonar plot. They go to flash as assets because drawing them at
// start took about 100 KiB of RAM.
// Usage: go run gen.go, then run go generate in games/uboat.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

// These match the constants in game.go and art.go; TestAssetSizes checks the
// sizes that follow from them.
const (
	screenWidth  = 120
	screenHeight = 120
	horizon      = 58
	scopeX       = 60.0
	scopeY       = 60.0
	scopeRadius  = 54.0
	sonarRadius  = 52
	stripWidth   = 256
)

// noise is a small hash so that the strips look the same on every run.
func noise(x, y int) uint32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ h>>13) * 1274126177
	return h ^ h>>16
}

// waves draws wave crests that thin out toward the horizon.
func waves() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, stripWidth, screenHeight-horizon))
	crest, light := color.NRGBA{70, 104, 140, 255}, color.NRGBA{110, 146, 180, 255}
	for y := range screenHeight - horizon {
		length := 2 + y/6
		gap := 12 + 40/(y+1)
		for x := 0; x < stripWidth; {
			x += gap/2 + int(noise(x, y)%uint32(gap))
			c := crest
			if noise(y, x)%4 == 0 {
				c = light
			}
			for i := range length {
				img.SetNRGBA((x+i)%stripWidth, y, c)
			}
			x += length
		}
	}
	return img
}

func clouds() image.Image {
	const height = 22
	img := image.NewNRGBA(image.Rect(0, 0, stripWidth, height))
	for _, c := range [...][3]int{{20, 8, 18}, {70, 12, 26}, {130, 6, 14}, {190, 10, 22}, {236, 7, 16}} {
		for y := range height {
			for x := range c[2] * 2 {
				dx, dy := float64(x-c[2])/float64(c[2]), float64(y-c[1])/5
				if dx*dx+dy*dy < 1 && noise(x+c[0], y)%5 != 0 {
					img.SetNRGBA((c[0]+x)%stripWidth, y, color.NRGBA{232, 236, 242, 255})
				}
			}
		}
	}
	return img
}

// mask is black outside a circle with a dark rim, like a periscope eyepiece.
func mask() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	for y := range screenHeight {
		for x := range screenWidth {
			d := math.Hypot(float64(x)+.5-scopeX, float64(y)+.5-scopeY)
			switch {
			case d > scopeRadius+1.5:
				img.SetNRGBA(x, y, color.NRGBA{0, 0, 0, 255})
			case d > scopeRadius:
				img.SetNRGBA(x, y, color.NRGBA{30, 34, 40, 255})
			}
		}
	}
	return img
}

// sonar draws the plot's rim, its half-range ring and a dotted cross.
func sonar() image.Image {
	const size = 2*sonarRadius + 1
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	ring, dim := color.NRGBA{60, 200, 110, 255}, color.NRGBA{24, 80, 48, 255}
	for y := range size {
		for x := range size {
			dx, dy := float64(x-sonarRadius), float64(y-sonarRadius)
			d := math.Hypot(dx, dy)
			switch {
			case math.Abs(d-sonarRadius) < 0.6:
				img.SetNRGBA(x, y, ring)
			case math.Abs(d-sonarRadius/2) < 0.5, d < sonarRadius && (x == sonarRadius || y == sonarRadius) && (x+y)%2 == 0:
				img.SetNRGBA(x, y, dim)
			}
		}
	}
	return img
}

func main() {
	for name, img := range map[string]image.Image{"waves.png": waves(), "clouds.png": clouds(), "mask.png": mask(), "sonar.png": sonar()} {
		out, err := os.Create(name)
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
}
