package flappy

import (
	_ "embed"
	"image/color"
	"strings"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/text"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

//go:embed assets/gopher.ika
var gopherData string

//go:embed assets/tiles.ika
var tilesData string

func loadImages() (gopher, tiles *ebiten.Image) {
	gopher, err := ebiten.NewImageFromString(gopherData)
	if err != nil {
		panic(err)
	}
	tiles, err = ebiten.NewImageFromString(tilesData)
	if err != nil {
		panic(err)
	}
	return gopher, tiles
}

type textAlign int

const (
	alignCenter textAlign = iota
	alignEnd
)

var textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}

// drawText draws lines starting at top y, spaced by size, like text/v2.
func drawText(screen *ebiten.Image, str string, size, x, y int, align textAlign) {
	var font tinyfont.Fonter = &freemono.Bold9pt7b
	baseline := size
	switch {
	case size >= titleFontSize:
		font = &freemono.Bold12pt7b
	case size <= smallFontSize:
		font = &tinyfont.TomThumb
		baseline = size - 1
	}
	for str != "" {
		line := str
		if n := strings.IndexByte(str, '\n'); n >= 0 {
			line, str = str[:n], str[n+1:]
		} else {
			str = ""
		}
		if line != "" {
			_, width := tinyfont.LineWidth(font, line)
			left := x - int(width)
			if align == alignCenter {
				left = x - int(width)/2
			}
			text.Draw(screen, line, font, left, y+baseline, textColor)
		}
		y += size
	}
}
