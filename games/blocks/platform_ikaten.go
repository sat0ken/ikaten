package blocks

import (
	_ "embed"
	"image/color"
	"strings"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/text"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/proggy"
)

//go:embed assets/background.ika
var backgroundData string

//go:embed assets/blocks.ika
var blocksData string

//go:embed assets/gamebg.ika
var gameBGData string

func loadImage(data string) *ebiten.Image {
	img, err := ebiten.NewImageFromString(data)
	if err != nil {
		panic(err)
	}
	return img
}

func loadImages() (background, blocks, gameBG *ebiten.Image) {
	return loadImage(backgroundData), loadImage(blocksData), loadImage(gameBGData)
}

type textAlign int

const (
	alignStart textAlign = iota
	alignCenter
	alignEnd
)

const arcadeFontBaseSize = 8

// The shadow alpha is binary on ikaten, so the shadow is opaque black.
var shadowColor = color.RGBA{0, 0, 0, 0x80}

// drawTextWithShadow mirrors the text/v2 version: lines are arcadeFontBaseSize*scale
// apart and (x, y) is aligned by primaryAlign and secondaryAlign.
func drawTextWithShadow(rt *ebiten.Image, str string, x, y, scale int, clr color.Color, primaryAlign, secondaryAlign textAlign) {
	var font tinyfont.Fonter = &proggy.TinySZ8pt7b
	ascent := 7
	if scale >= 4 {
		font, ascent = &freemono.Bold18pt7b, 22
	}
	lineHeight := arcadeFontBaseSize * scale
	lines := strings.Count(str, "\n") + 1
	switch secondaryAlign {
	case alignCenter:
		y -= lines * lineHeight / 2
	case alignEnd:
		y -= lines * lineHeight
	}
	for str != "" || lines > 0 {
		line := str
		if n := strings.IndexByte(str, '\n'); n >= 0 {
			line, str = str[:n], str[n+1:]
		} else {
			str = ""
		}
		lines--
		if line != "" {
			_, width := tinyfont.LineWidth(font, line)
			left := x
			switch primaryAlign {
			case alignCenter:
				left -= int(width) / 2
			case alignEnd:
				left -= int(width)
			}
			text.Draw(rt, line, font, left+1, y+ascent+1, shadowColor)
			text.Draw(rt, line, font, left, y+ascent, clr)
		}
		y += lineHeight
	}
}
