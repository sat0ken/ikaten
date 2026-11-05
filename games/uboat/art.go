package uboat

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Ships are seen
// from the side with the bow to the right.
var palette = map[byte]color.RGBA{
	'K': {52, 56, 64, 255},
	'g': {110, 116, 124, 255},
	'w': {180, 184, 190, 255},
	'b': {24, 24, 28, 255},
	'r': {120, 40, 40, 255},
	'Y': {255, 230, 80, 255},
	'O': {255, 140, 30, 255},
	'R': {220, 50, 30, 255},
	'S': {90, 90, 96, 255},
	's': {150, 150, 156, 255},
	'W': {235, 240, 245, 255},
}

var shipArt = [kinds][]string{
	cargo: {
		"....................b...............",
		"....................g..........g....",
		"......g............ggg.........g....",
		"......g...........gwwwg........g....",
		"......g.....gg....gwwwg...gg...g....",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		".KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"..KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK..",
		"...rrrrrrrrrrrrrrrrrrrrrrrrrrrrrr...",
	},
	tanker: {
		"..b.........................................",
		"..g..ww.....................................",
		".gggwwww............................g.......",
		".gggwwww...........................ggg......",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		".KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"..KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"...rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr...",
	},
	destroyer: {
		"..............g...................",
		"..............g...................",
		"...........gg.gg..g...............",
		"......gg..gggggg.gbg....gg........",
		"..gg.gggggggggggggggggggggg..gg...",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		".KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"...KKKKKKKKKKKKKKKKKKKKKKKKKKKKK..",
		"....rrrrrrrrrrrrrrrrrrrrrrrrrr....",
	},
}

var fireArt = []string{
	"...Y.Y...",
	"..YOYOY..",
	".YORRROY.",
	"YORRYRROY",
	"ORRYYYRRO",
	".ORRYRRO.",
	"..ORRRO..",
}

var smokeArt = []string{
	"..sss..",
	".sSSSs.",
	"sSSSSSs",
	".sSSSs.",
	"..sSs..",
}

var splashArt = []string{
	"..W..",
	".WWW.",
	".WsW.",
	"WWsWW",
	"WsWsW",
	"WWWWW",
}

// digitGlyphs is a 3x5 font for numbers.
var digitGlyphs = [10][5]string{
	{"###", "#.#", "#.#", "#.#", "###"},
	{".#.", "##.", ".#.", ".#.", "###"},
	{"###", "..#", "###", "#..", "###"},
	{"###", "..#", "###", "..#", "###"},
	{"#.#", "#.#", "###", "..#", "..#"},
	{"###", "#..", "###", "..#", "###"},
	{"###", "#..", "###", "#.#", "###"},
	{"###", "..#", "..#", ".#.", ".#."},
	{"###", "#.#", "###", "#.#", "###"},
	{"###", "#.#", "###", "..#", "###"},
}

const stripWidth = 256 // the sky and sea strips repeat every stripWidth pixels; see assets/gen.go

var (
	imageShip                          [kinds]*ebiten.Image
	imageFire, imageSmoke, imageSplash *ebiten.Image
	imageDigits                        [10]*ebiten.Image
	imageClouds, imageWaves            *ebiten.Image // in flash, drawn by assets/gen.go
	imageMask                          *ebiten.Image // black outside the periscope's circle
	imageSonar                         *ebiten.Image // rings and cross of the sonar plot
	imageSky, imageSea, imageBlack     *ebiten.Image
	imageWhite, imageYellow, imageRed  *ebiten.Image
	imageGreen, imageDim, imageGray    *ebiten.Image
	imageContact                       [kinds]*ebiten.Image
)

func sprite(rows []string) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	for y, row := range rows {
		for x := range len(row) {
			if c, ok := palette[row[x]]; ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

func pixel(c color.RGBA, w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

func init() {
	for k := range imageShip {
		imageShip[k] = sprite(shipArt[k])
	}
	imageFire = sprite(fireArt)
	imageSmoke = sprite(smokeArt)
	imageSplash = sprite(splashArt)
	digits := ebiten.NewImage(40, 5)
	for d, glyph := range digitGlyphs {
		for y, row := range glyph {
			for x := range len(row) {
				if row[x] == '#' {
					digits.Set(d*4+x, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
	// SubImage allocates, so the glyphs are cut once here.
	for d := range imageDigits {
		imageDigits[d] = digits.SubImage(image.Rect(d*4, 0, d*4+3, 5)).(*ebiten.Image)
	}
	imageWaves, imageClouds, imageMask, imageSonar = loadAssets()
	imageSky = pixel(color.RGBA{150, 182, 214, 255}, 1, 1)
	imageSea = pixel(color.RGBA{40, 70, 108, 255}, 1, 1)
	imageBlack = pixel(color.RGBA{0, 0, 0, 255}, 1, 1)
	imageWhite = pixel(color.RGBA{240, 240, 240, 255}, 1, 1)
	imageYellow = pixel(color.RGBA{255, 220, 40, 255}, 1, 1)
	imageRed = pixel(color.RGBA{230, 40, 30, 255}, 1, 1)
	imageGreen = pixel(color.RGBA{60, 200, 110, 255}, 1, 1)
	imageDim = pixel(color.RGBA{24, 80, 48, 255}, 1, 1)
	imageGray = pixel(color.RGBA{70, 70, 76, 255}, 1, 1)
	imageContact[cargo] = pixel(color.RGBA{150, 230, 170, 255}, 2, 2)
	imageContact[tanker] = pixel(color.RGBA{150, 230, 170, 255}, 3, 2)
	imageContact[destroyer] = pixel(color.RGBA{240, 70, 50, 255}, 3, 3)
}
