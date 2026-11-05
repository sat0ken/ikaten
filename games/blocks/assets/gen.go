//go:build ignore

// gen prepares the blocks images from Ebitengine's examples/resources/images.
// Usage: go run gen.go <images directory>, then run go generate in games/blocks.
//
// gamebg.png is gophers.jpg scaled to cover the screen (bilinear), turned light
// gray like the original colorm matrix, with the translucent windows baked in.
package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
)

const screenWidth, screenHeight = 240, 240

// windows mirrors the window rectangles in gamescene.go for a 240-pixel screen.
var windows = []image.Rectangle{
	image.Rect(20, 20, 120, 220),   // field
	image.Rect(140, 30, 190, 80),   // next
	image.Rect(140, 100, 220, 120), // score
	image.Rect(140, 140, 220, 160), // level
	image.Rect(140, 180, 220, 200), // lines
}

func copyFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		log.Fatal(err)
	}
}

func sample(img image.Image, x, y float64) [3]float64 {
	b := img.Bounds()
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	at := func(x, y int) [3]float64 {
		x = min(max(x, 0), b.Dx()-1)
		y = min(max(y, 0), b.Dy()-1)
		r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return [3]float64{float64(r) / 65535, float64(g) / 65535, float64(bl) / 65535}
	}
	var c [3]float64
	for i, w := range [4]float64{(1 - fx) * (1 - fy), fx * (1 - fy), (1 - fx) * fy, fx * fy} {
		p := at(x0+i%2, y0+i/2)
		for k := range c {
			c[k] += p[k] * w
		}
	}
	return c
}

func main() {
	dir := os.Args[1]
	copyFile(filepath.Join(dir, "blocks", "background.png"), "background.png")
	copyFile(filepath.Join(dir, "blocks", "blocks.png"), "blocks.png")

	f, err := os.Open(filepath.Join(dir, "gophers.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	src, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	w, h := float64(src.Bounds().Dx()), float64(src.Bounds().Dy())
	scale := max(screenWidth/w, screenHeight/h)
	dst := image.NewNRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			c := sample(src, (float64(x)+0.5-screenWidth/2)/scale+w/2-0.5, (float64(y)+0.5-screenHeight/2)/scale+h/2-0.5)
			// lightGray = 0.7 * grayscale + 0.3 * identity, then +0.3.
			luma := 0.299*c[0] + 0.587*c[1] + 0.114*c[2]
			for k := range c {
				c[k] = min(1, 0.7*luma+0.3*c[k]+0.3)
			}
			for _, r := range windows {
				if image.Pt(x, y).In(r) {
					for k := range c {
						c[k] *= 1 - 0xc0/255.0
					}
				}
			}
			dst.SetNRGBA(x, y, color.NRGBA{uint8(c[0]*255 + 0.5), uint8(c[1]*255 + 0.5), uint8(c[2]*255 + 0.5), 255})
		}
	}
	out, err := os.Create("gamebg.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, dst); err != nil {
		log.Fatal(err)
	}
}
