package main

import (
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/games/assets"
	"github.com/sat0ken/ikaten/hardware"
	"github.com/sat0ken/ikaten/vector"
	"image/color"
)

type game struct {
	sprite *ikaten.Image
	angle  float64
}

var background = color.RGBA{R: 16, G: 24, B: 40, A: 255}
var accent = color.RGBA{R: 48, G: 120, B: 160, A: 255}

func (g *game) Update() error              { g.angle += .02; return nil }
func (g *game) Layout(w, h int) (int, int) { return 120, 120 }
func (g *game) Draw(screen *ikaten.Image) {
	screen.Fill(background)
	vector.FillRect(screen, 5, 20, 110, 80, accent, false)
	vector.FillCircle(screen, 60, 60, 30, background, false)
	var op ikaten.DrawImageOptions
	op.GeoM.Translate(-8, -8)
	op.GeoM.Rotate(g.angle)
	op.GeoM.Scale(3, 3)
	op.GeoM.Translate(60, 60)
	screen.DrawImage(g.sprite, &op)
	vector.StrokeLine(screen, 5, 105, 115, 105, 2, accent, false)
	ebitenutil.DebugPrint(screen, "RGB565 + mask")
}
func main() {
	sprite, err := assets.NewSprite()
	if err != nil {
		println(err.Error())
		return
	}
	ikaten.SetHardware(hardware.Device)
	if err := ikaten.RunGame(&game{sprite: sprite}); err != nil {
		println(err.Error())
	}
}
