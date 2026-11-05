package ebitenutil

import (
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/text"
	"image/color"
	"tinygo.org/x/tinyfont/proggy"
)

var white = color.RGBA{R: 255, G: 255, B: 255, A: 255}

func DebugPrint(dst *ikaten.Image, str string) { DebugPrintAt(dst, str, 0, 0) }
func DebugPrintAt(dst *ikaten.Image, str string, x, y int) {
	text.Draw(dst, str, &proggy.TinySZ8pt7b, x, y+10, white)
}
