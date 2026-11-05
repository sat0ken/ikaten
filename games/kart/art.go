package kart

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. In the kart,
// 'M' and 'm' take each driver's colors.
var palette = map[byte]color.RGBA{
	'W': {240, 240, 240, 255},
	'K': {30, 30, 40, 255},
	'g': {110, 110, 120, 255},
	'Y': {255, 230, 60, 255},
	'R': {255, 60, 60, 255},
	'O': {255, 130, 20, 255},
}

// kartArt is the rear view. Turning frames move the driver one pixel.
var kartArt = []string{
	"......WWWW......",
	".....WWWWWW.....",
	".....WMMMMW.....",
	"......WWWW......",
	"....mMMMMMMm....",
	"..KKmMMMMMMmKK..",
	".KKKmMMYYMMmKKK.",
	".KKKmmmmmmmmKKK.",
	".KKKggggggggKKK.",
	"..KK.gRggRg.KK..",
	"....gggggggg....",
}

const driverRows = 4

var flameArt = []string{".YOOY.", "YOOOOY", ".OYYO.", "..OO.."}

// kartColors holds the main and dark colors: the player is blue.
var kartColors = [karts][2]color.RGBA{
	{{40, 120, 255, 255}, {20, 60, 160, 255}},
	{{230, 50, 50, 255}, {140, 20, 30, 255}},
	{{60, 200, 80, 255}, {20, 110, 40, 255}},
	{{250, 200, 40, 255}, {170, 120, 20, 255}},
}

// digitGlyphs is a 3x5 font for numbers; 10 is ':' and 11 is '.'.
var digitGlyphs = [12][5]string{
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
	{"...", ".#.", "...", ".#.", "..."},
	{"...", "...", "...", "...", ".#."},
}

const (
	glyphColon = 10
	glyphDot   = 11
)

var (
	imageTrack     *ebiten.Image
	imageKart      [karts][3]*ebiten.Image // left, straight, right
	imageFlame     *ebiten.Image
	imageSpark     [3]*ebiten.Image // dust, blue, orange
	imageMountains *ebiten.Image
	imageDigits    [12]*ebiten.Image
	imageSky       [2]*ebiten.Image
	imageGrass     *ebiten.Image
	imageHaze      *ebiten.Image
)

func sprite(rows []string, colors map[byte]color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	for y, row := range rows {
		for x := range len(row) {
			if c, ok := colors[row[x]]; ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

// shifted moves the driver rows by dx pixels for the turning frames.
func shifted(rows []string, dx int) []string {
	out := make([]string, len(rows))
	copy(out, rows)
	for y := range driverRows {
		row := []byte(rows[y])
		moved := make([]byte, len(row))
		for x := range row {
			moved[x] = '.'
			if src := x - dx; src >= 0 && src < len(row) {
				moved[x] = row[src]
			}
		}
		out[y] = string(moved)
	}
	return out
}

func pixel(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

// mountains draws a strip that tiles every 256 pixels: snowy peaks behind green
// hills. Only whole periods of 256 pixels are used so that the ends meet.
func mountains() *ebiten.Image {
	const width, height = 256, 18
	img := ebiten.NewImage(width, height)
	rock, snow, hill := color.RGBA{110, 128, 170, 255}, color.RGBA{232, 238, 248, 255}, color.RGBA{60, 140, 84, 255}
	for x := range width {
		t := 2 * math.Pi * float64(x) / width
		peak := 9 + 4*math.Sin(3*t) + 3*math.Sin(7*t+1) + 1.5*math.Sin(13*t+2)
		ridge := 4 + 1.5*math.Sin(5*t+0.5) + math.Sin(11*t)
		for y := range height {
			h := float64(height - y)
			switch {
			case h <= ridge:
				img.Set(x, y, hill)
			case h <= peak-2:
				img.Set(x, y, rock)
			case h <= peak:
				img.Set(x, y, snow)
			}
		}
	}
	return img
}

func init() {
	imageTrack = loadTrack()
	for k := range karts {
		colors := map[byte]color.RGBA{'M': kartColors[k][0], 'm': kartColors[k][1]}
		for c, v := range palette {
			colors[c] = v
		}
		for frame, dx := range [3]int{-1, 0, 1} {
			imageKart[k][frame] = sprite(shifted(kartArt, dx), colors)
		}
	}
	imageFlame = sprite(flameArt, palette)
	for i, c := range [3]color.RGBA{{210, 210, 200, 255}, {90, 170, 255, 255}, {255, 150, 30, 255}} {
		img := ebiten.NewImage(2, 2)
		img.Fill(c)
		imageSpark[i] = img
	}
	imageMountains = mountains()
	digits := ebiten.NewImage(4*len(digitGlyphs), 5)
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
	imageSky[0] = pixel(color.RGBA{70, 130, 230, 255})
	imageSky[1] = pixel(color.RGBA{140, 190, 245, 255})
	imageGrass = pixel(color.RGBA{58, 160, 66, 255})
	imageHaze = pixel(color.RGBA{170, 205, 220, 255})
}
