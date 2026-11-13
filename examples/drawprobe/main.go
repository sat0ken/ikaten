//go:build tinygo && (esp32s3 || esp32 || rp2040)

// drawprobe times the drawing calls the 240 by 240 games are made of, one
// kind at a time, on images in RAM. It prints name,iterations,us_per_op to
// serial. Nothing is sent to the display.
package main

import (
	"image/color"
	"time"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

var screen = ebiten.NewImage(240, 240)

// op times f run n times.
func op(name string, n int, f func()) {
	start := time.Now()
	for range n {
		f()
	}
	us := time.Since(start).Microseconds()
	println(name, ",", n, ",", us/int64(n), ".", us*10/int64(n)%10)
}

func pixel(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

// shape is a ridge-like image: column x is filled from a wavy height down.
func shape(w, h int, opaque bool) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	c := color.RGBA{100, 170, 110, 255}
	if opaque {
		img.Fill(c)
		return img
	}
	for x := range w {
		top := (x*7)%h/2 + h/4
		for y := top; y < h; y++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func main() {
	time.Sleep(2 * time.Second)
	println("name,iterations,us_per_op")
	green := pixel(color.RGBA{90, 190, 80, 255})
	var o ebiten.DrawImageOptions
	at := func(img *ebiten.Image, sx, sy, x, y float64) {
		o.GeoM.Reset()
		o.GeoM.Scale(sx, sy)
		o.GeoM.Translate(x, y)
		screen.DrawImage(img, &o)
	}
	op("screen_clear", 50, func() { screen.Clear() })
	op("screen_fill", 50, func() { screen.Fill(color.RGBA{100, 170, 240, 255}) })
	// From here the screen is opaque, as jump's is.
	op("fill_1x1", 2000, func() { at(green, 1, 1, 10, 10) })
	op("fill_16x2", 2000, func() { at(green, 16, 2, 10, 10) })
	op("fill_240x1", 1000, func() { at(green, 240, 1, 0, 10) })
	op("fill_240x20", 500, func() { at(green, 240, 20, 0, 10) })
	op("fill_208x184", 100, func() { at(green, 208, 184, 16, 28) })
	tile := shape(16, 16, true)
	op("tile_16x16_opaque", 2000, func() { at(tile, 1, 1, 10, 10) })
	sprite := shape(16, 16, false)
	op("sprite_16x16_masked", 2000, func() { at(sprite, 1, 1, 10, 10) })
	op("sprite_16x16_flipped", 2000, func() { at(sprite, -1, 1, 26, 10) })
	op("sprite_16x16_x3", 500, func() { at(sprite, 3, 3, 10, 10) })
	column := shape(16, 240, true)
	op("column_16x240_opaque", 500, func() { at(column, 1, 1, 10, 0) })
	ridge := shape(180, 63, false)
	op("ridge_180x63_masked", 200, func() { at(ridge, 1, 1, 10, 100) })
	op("ridge_180x63_half", 200, func() { at(ridge, 1, 0.5, 10, 100) })
	op("ridge_180x63_flipped_half", 200, func() { at(ridge, 1, -0.5, 10, 100) })
	op("ridge_180x63_third", 200, func() { at(ridge, 1, 1.0/3, 10, 100) })
	var line pixtext.Line
	white := color.RGBA{250, 250, 250, 255}
	op("text_10_chars", 500, func() { line.Reset().Str("SCORE 1234").Draw(screen, 10, 10, 1, white) })
	op("text_10_chars_shaded", 500, func() {
		line.Reset().Str("SCORE 1234").Shade(color.RGBA{0, 0, 0, 255}).Draw(screen, 10, 10, 1, white)
	})
	op("text_10_chars_x3", 200, func() { line.Reset().Str("GOPHER JUM").Draw(screen, 10, 10, 3, white) })
	op("text_10_chars_tinted", 500, func() {
		line.Reset().Str("SCORE 1234").Draw(screen, 10, 10, 1, color.RGBA{250, 250, 250, 200})
	})
	println("done")
	for {
		time.Sleep(time.Second)
	}
}
