//go:build ignore

// gen draws the large uboat images: the sea and sky strips and the sonar
// plot. They go to flash as assets because drawing them at start took about
// 100 KiB of RAM.
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
	screenWidth  = 240
	screenHeight = 240
	horizon      = 116
	sonarRadius  = 104
	stripWidth   = 512
)

// noise is a small hash so that the strips look the same on every run.
func noise(x, y int) uint32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ h>>13) * 1274126177
	return h ^ h>>16
}

// The strips are drawn twice as wide, so they are stored at half width:
// flash reads are slow on the badge and this halves them. Each texel is two
// pixels across and one down.
const texels = stripWidth / 2

// waves draws wave crests that thin out toward the horizon: a light crest
// with a dark trough under it, longer and further apart nearer the viewer.
func waves() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, texels, screenHeight-horizon))
	trough := color.NRGBA{30, 56, 90, 255}
	crest, light, foam := color.NRGBA{70, 104, 140, 255}, color.NRGBA{110, 146, 180, 255}, color.NRGBA{170, 196, 220, 255}
	for y := 0; y < screenHeight-horizon; y++ {
		near := y * y / (screenHeight - horizon) // grows faster than y: perspective
		length := 1 + near/10
		gap := 8 + 60/(y+2)
		for x := 0; x < texels; {
			x += gap/2 + int(noise(x, y)%uint32(gap))
			c := crest
			switch noise(y, x) % 6 {
			case 0:
				c = light
			case 1:
				if y > 40 {
					c = foam
				}
			}
			for i := range length {
				img.SetNRGBA((x+i)%texels, y, c)
				if y+1 < screenHeight-horizon && i > 0 && i < length-1 && y > 8 {
					img.SetNRGBA((x+i)%texels, y+1, trough)
				}
			}
			x += length + 1
		}
	}
	return img
}

// clouds draws puffy clouds from overlapping ellipses, white on top and grey
// along their flat bottoms.
func clouds() image.Image {
	const height = 44
	img := image.NewNRGBA(image.Rect(0, 0, texels, height))
	white, shade, dark := color.NRGBA{236, 240, 246, 255}, color.NRGBA{200, 210, 224, 255}, color.NRGBA{176, 190, 208, 255}
	// Each cloud: its center x on the screen, its base row and its half width.
	for _, c := range [...][3]int{{40, 30, 34}, {150, 22, 22}, {262, 36, 44}, {380, 18, 18}, {462, 30, 30}} {
		base, half := c[1], c[2]
		for p := range 6 {
			// Puffs along the cloud, the middle ones taller.
			px := c[0] - half + (2*half)*p/5 + int(noise(c[0], p)%7) - 3
			rx := float64(half) / 2.4
			ry := rx * 0.8
			if p == 2 || p == 3 {
				ry *= 1.4
			}
			cy := float64(base) - ry*0.5
			for y := range height {
				for t := range texels {
					// The texel's center on the screen, nearest copy of the puff.
					x := float64(2*t+1) - float64(px)
					x -= stripWidth * math.Round(x/stripWidth)
					dx, dy := x/rx, (float64(y)+.5-cy)/ry
					if dx*dx+dy*dy >= 1 || y > base {
						continue
					}
					col := white
					switch {
					case y >= base-1:
						col = dark
					case dy > 0.35:
						col = shade
					}
					img.SetNRGBA(t, y, col)
				}
			}
		}
	}
	return img
}

// sonar draws the plot's rim with ticks every 30 degrees, its range rings and
// a dotted cross.
func sonar() image.Image {
	const size = 2*sonarRadius + 1
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	ring, dim := color.NRGBA{60, 200, 110, 255}, color.NRGBA{24, 80, 48, 255}
	for y := range size {
		for x := range size {
			dx, dy := float64(x-sonarRadius), float64(y-sonarRadius)
			d := math.Hypot(dx, dy)
			a := math.Atan2(dy, dx) * 180 / math.Pi
			tick := math.Abs(math.Remainder(a, 30)) < 90/(math.Pi*d+1)
			switch {
			case math.Abs(d-sonarRadius) < 0.6, d > sonarRadius-6 && d < sonarRadius && tick:
				img.SetNRGBA(x, y, ring)
			case math.Abs(d-sonarRadius/2) < 0.5 && (x+y)%2 == 0, math.Abs(d-sonarRadius/4) < 0.5 && (x+y)%2 == 0,
				d < sonarRadius && (x == sonarRadius || y == sonarRadius) && (x+y)%3 == 0:
				img.SetNRGBA(x, y, dim)
			}
		}
	}
	return img
}

func main() {
	for name, img := range map[string]image.Image{"waves.png": waves(), "clouds.png": clouds(), "sonar.png": sonar()} {
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
