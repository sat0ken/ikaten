package kart

import (
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. In the kart,
// 'M' and 'm' take each driver's colors.
var palette = map[byte]color.RGBA{
	'W': {240, 240, 240, 255},
	'K': {30, 30, 40, 255},
	'k': {70, 70, 80, 255},
	'g': {110, 110, 120, 255},
	'G': {170, 170, 180, 255},
	'c': {120, 200, 255, 255},
	'Y': {255, 230, 60, 255},
	'R': {255, 60, 60, 255},
	'O': {255, 130, 20, 255},
}

// kartArt is the rear view. Turning frames move the driver two pixels.
var kartArt = []string{
	"..............KKKK..............",
	".............KMMMMK.............",
	"............KWWWWWWK............",
	"............KWWWWWWK............",
	"...........KKcccccccK...........",
	"............KKKKKKKKK...........",
	"............KWWWWWWK............",
	".............KWWWWK.............",
	"...........KKmmmmmmKK...........",
	"....KKKKKKKggggggggggKKKKKKK....",
	"....KKKKKKKKKKKKKKKKKKKKKKKK....",
	".......KMMMMMMMMMMMMMMMMK.......",
	".KKKKKKMMMMMMMMMMMMMMMMMMKKKKKK.",
	".KKKKKKMMMMMMMYYYYMMMMMMMKKKKKK.",
	".KkkkkKMMMMMMMYYYYMMMMMMMKkkkkK.",
	".KKKKKKMMMMMMMMMMMMMMMMMMKKKKKK.",
	".KkkkkKKKmmmmmmmmmmmmmmKKKkkkkK.",
	".KKKKKK..KggggggggggggK..KKKKKK.",
	".KkkkkK..KkkkkkkkkkkkkK..KkkkkK.",
	".KKKKKK..KKKgggKKKgggKK..KKKKKK.",
	".KKKKKK.....KGK...KGK....KKKKKK.",
	"............KKK...KKK...........",
}

const driverRows = 11

var flameArt = []string{
	"....ROOR....",
	"..ROOYYOOR..",
	".ROOYYYYOOR.",
	".ROYYYYYYOR.",
	".ROOYYYYOOR.",
	"..ROOYYOOR..",
	"...ROOOOR...",
	"....ROOR....",
}

// kartColors holds the main and dark colors: the player is blue.
var kartColors = [karts][2]color.RGBA{
	{{40, 120, 255, 255}, {20, 60, 160, 255}},
	{{230, 50, 50, 255}, {140, 20, 30, 255}},
	{{60, 200, 80, 255}, {20, 110, 40, 255}},
	{{250, 200, 40, 255}, {170, 120, 20, 255}},
}

// sprites holds a game's images. A Game makes them in NewGame and drops them
// with it, so that the launcher keeps only the running game's in RAM. The
// track stays in flash.
type sprites struct {
	track     *ebiten.Image
	kart      [karts][3]*ebiten.Image // left, straight, right
	flame     *ebiten.Image
	spark     [3]*ebiten.Image // dust, blue, orange
	mountains *ebiten.Image
	sky       [4]*ebiten.Image // from the top down
	grass     *ebiten.Image
	haze      *ebiten.Image
}

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

// mountains draws a strip that tiles every 512 pixels: snowy peaks behind green
// hills. Only whole periods of 512 pixels are used so that the ends meet.
func mountains() *ebiten.Image {
	const width, height = 512, 36
	img := ebiten.NewImage(width, height)
	rock, rockLit, snow := color.RGBA{110, 128, 170, 255}, color.RGBA{136, 152, 190, 255}, color.RGBA{232, 238, 248, 255}
	hill, hillTop := color.RGBA{60, 140, 84, 255}, color.RGBA{86, 170, 100, 255}
	for x := range width {
		t := 2 * math.Pi * float64(x) / width
		peak := 18 + 8*math.Sin(3*t) + 6*math.Sin(7*t+1) + 3*math.Sin(13*t+2)
		ridge := 8 + 3*math.Sin(5*t+0.5) + 2*math.Sin(11*t)
		for y := range height {
			h := float64(height - y)
			switch {
			case h <= ridge-1:
				img.Set(x, y, hill)
			case h <= ridge:
				img.Set(x, y, hillTop)
			case h <= peak-4:
				// The rock is lit on the slopes that face left.
				if math.Cos(3*t)+math.Cos(7*t+1) > 0 {
					img.Set(x, y, rockLit)
				} else {
					img.Set(x, y, rock)
				}
			case h <= peak:
				img.Set(x, y, snow)
			}
		}
	}
	return img
}

func newSprites() *sprites {
	a := &sprites{track: loadTrack()}
	for k := range karts {
		colors := map[byte]color.RGBA{'M': kartColors[k][0], 'm': kartColors[k][1]}
		for c, v := range palette {
			colors[c] = v
		}
		for frame, dx := range [3]int{-2, 0, 2} {
			a.kart[k][frame] = sprite(shifted(kartArt, dx), colors)
		}
	}
	a.flame = sprite(flameArt, palette)
	for i, c := range [3]color.RGBA{{210, 210, 200, 255}, {90, 170, 255, 255}, {255, 150, 30, 255}} {
		img := ebiten.NewImage(4, 4)
		img.Fill(c)
		a.spark[i] = img
	}
	a.mountains = mountains()
	for i, c := range [4]color.RGBA{{60, 116, 220, 255}, {80, 140, 232, 255}, {110, 166, 240, 255}, {150, 196, 248, 255}} {
		a.sky[i] = pixel(c)
	}
	a.grass = pixel(color.RGBA{58, 160, 66, 255})
	a.haze = pixel(color.RGBA{170, 205, 220, 255})
	return a
}
