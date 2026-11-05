package shooter

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent.
var palette = map[byte]color.RGBA{
	'W': {255, 255, 255, 255},
	'w': {170, 180, 200, 255},
	'g': {100, 110, 130, 255},
	'C': {80, 220, 255, 255},
	'B': {40, 90, 220, 255},
	'R': {230, 50, 50, 255},
	'r': {140, 20, 30, 255},
	'O': {255, 150, 30, 255},
	'Y': {255, 230, 60, 255},
	'G': {60, 220, 90, 255},
	'd': {20, 120, 50, 255},
	'P': {180, 80, 230, 255},
	'p': {100, 40, 140, 255},
	'K': {30, 30, 50, 255},
}

var playerArt = [2][]string{{
	".....W.....",
	"....wCw....",
	"....wCw....",
	"...wwWww...",
	"...wBBBw...",
	"..gwBCBwg..",
	".ggwBBBwgg.",
	"gggwwWwwggg",
	"gg..wWw..gg",
	"g...O.O...g",
	"....Y.Y....",
}, {
	".....W.....",
	"....wCw....",
	"....wCw....",
	"...wwWww...",
	"...wBBBw...",
	"..gwBCBwg..",
	".ggwBBBwgg.",
	"gggwwWwwggg",
	"gg..wWw..gg",
	"g...Y.Y...g",
	"....O.O....",
}}

var scoutArt = []string{
	".r.....r.",
	".rR...Rr.",
	".rRRRRRr.",
	"rRRYRYRRr",
	".RRRRRRR.",
	"..RRWRR..",
	"...RRR...",
	"....R....",
}

var weaverArt = []string{
	"...ddd...",
	"..dGGGd..",
	"d.GGWGG.d",
	"dGGWKWGGd",
	"dGGGWGGGd",
	"d.GGGGG.d",
	"..dGGGd..",
	"...ddd...",
}

var gunnerArt = []string{
	".pp.....pp.",
	"pPPp...pPPp",
	"pPPPpppPPPp",
	".pPPPYPPPp.",
	"..pPYKYPp..",
	"..pPPYPPp..",
	"...pPPPp...",
	"....pOp....",
	".....O.....",
}

// The boss is symmetric: each row is mirrored around its last character.
var bossHalf = []string{
	"........pppppppp",
	"......ppPPPPPPPP",
	"....ppPPPPPPPPPP",
	"..ppPPPgggPPPPrr",
	".pPPPPgwwwgPPrRR",
	"pPPPPPgwCwgPrRYY",
	"pPPgPPgwwwgPrRYW",
	"pPgggPPgggPPPrRR",
	"pPPgPPPPPPPPPPrr",
	".pPPPPpppPPPPPPP",
	"..pPPp...pPPPPPP",
	"..pOp.....pPPPPp",
	"...O.......pPPp.",
	"............pp..",
}

var shotArt = []string{"WW", "CC", "CC", "BB"}

var bulletArt = []string{".R.", "RWR", ".R."}

var itemArt = []string{
	".OOOOO.",
	"OYYYYOO",
	"OYOOYOO",
	"OYYYYOO",
	"OYOOOOO",
	"OYOOOOO",
	".OOOOO.",
}

var lifeArt = []string{"..W..", ".wCw.", "wwBww", "g.O.g"}

var bombArt = []string{".YO.", "OOOO", "OOOO", ".OO."}

// digitGlyphs is a 3x5 font for numbers, drawn as images without text rendering.
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

const blastFrames = 4

var (
	imagePlayer          [2]*ebiten.Image
	imageEnemy           [4]*ebiten.Image // indexed by kind
	imageShot            *ebiten.Image
	imageBullet          *ebiten.Image
	imageItem            *ebiten.Image
	imageLife            *ebiten.Image
	imageBomb            *ebiten.Image
	imageBlast           [blastFrames]*ebiten.Image
	imageDigits          [10]*ebiten.Image
	imageStar            [3]*ebiten.Image
	imageBar, imageBarBG *ebiten.Image
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

func mirrored(half []string) []string {
	rows := make([]string, len(half))
	for i, h := range half {
		b := []byte(h)
		for j := len(h) - 2; j >= 0; j-- {
			b = append(b, h[j])
		}
		rows[i] = string(b)
	}
	return rows
}

func pixel(c color.RGBA, w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// blast draws expanding rings: a white core that turns into red sparks.
func blast(frame int) *ebiten.Image {
	const size = 15
	img := ebiten.NewImage(size, size)
	colors := [blastFrames][2]color.RGBA{
		{{255, 255, 255, 255}, {255, 230, 60, 255}},
		{{255, 230, 60, 255}, {255, 150, 30, 255}},
		{{255, 150, 30, 255}, {230, 50, 50, 255}},
		{{230, 50, 50, 255}, {140, 20, 30, 255}},
	}
	outer := 2.5 + float64(frame)*1.7
	inner := outer - 2 - float64(frame)*0.3
	for y := range size {
		for x := range size {
			d := math.Hypot(float64(x-size/2), float64(y-size/2))
			switch {
			case d > outer:
			case d > inner:
				// Later frames break the ring into sparks.
				if frame < 2 || (x+y+frame)%3 != 0 {
					img.Set(x, y, colors[frame][1])
				}
			case frame < 2:
				img.Set(x, y, colors[frame][0])
			}
		}
	}
	return img
}

func init() {
	for i := range imagePlayer {
		imagePlayer[i] = sprite(playerArt[i])
	}
	imageEnemy[scout] = sprite(scoutArt)
	imageEnemy[weaver] = sprite(weaverArt)
	imageEnemy[gunner] = sprite(gunnerArt)
	imageEnemy[boss] = sprite(mirrored(bossHalf))
	imageShot = sprite(shotArt)
	imageBullet = sprite(bulletArt)
	imageItem = sprite(itemArt)
	imageLife = sprite(lifeArt)
	imageBomb = sprite(bombArt)
	for i := range imageBlast {
		imageBlast[i] = blast(i)
	}
	// Digits share one image; SubImage allocates, so it runs once here.
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
	for d := range imageDigits {
		imageDigits[d] = digits.SubImage(image.Rect(d*4, 0, d*4+3, 5)).(*ebiten.Image)
	}
	imageStar[0] = pixel(color.RGBA{70, 70, 100, 255}, 1, 1)
	imageStar[1] = pixel(color.RGBA{140, 140, 180, 255}, 1, 1)
	imageStar[2] = pixel(color.RGBA{255, 255, 255, 255}, 1, 2)
	imageBar = pixel(color.RGBA{230, 50, 50, 255}, 1, 1)
	imageBarBG = pixel(color.RGBA{60, 20, 30, 255}, 1, 1)
	for i := range dirs {
		a := 2 * math.Pi * float64(i) / float64(len(dirs))
		dirs[i] = vec{int(math.Round(math.Cos(a) * 256)), int(math.Round(math.Sin(a) * 256))}
	}
}
