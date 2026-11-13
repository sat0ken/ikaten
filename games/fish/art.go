package fish

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Fish face
// right.
var palette = map[byte]color.RGBA{
	'K': {16, 18, 22, 255},
	'w': {200, 210, 220, 255},
	'W': {245, 248, 250, 255},
	'g': {90, 110, 130, 255},
	'G': {60, 120, 90, 255},
	'B': {30, 60, 100, 255},
	'b': {60, 100, 150, 255},
	'r': {220, 70, 70, 255},
	'R': {250, 150, 160, 255},
	'y': {240, 210, 60, 255},
	'n': {140, 110, 70, 255},
	'N': {90, 70, 44, 255},
	'p': {240, 220, 225, 255},
	'P': {200, 120, 150, 255},
	'h': {236, 236, 228, 255},
	'H': {170, 60, 44, 255},
	'c': {240, 190, 150, 255},
	'u': {50, 90, 170, 255},
	'e': {50, 130, 70, 255},
	'E': {30, 90, 50, 255},
	'k': {110, 106, 100, 255},
	'o': {255, 170, 60, 255},
}

var fishArt = [kinds][]string{
	aji: {
		"...ggg....",
		".ggggggg..",
		"gwwwwwwwKw",
		".wwwwwww..",
		"..w...w...",
	},
	saba: {
		"G...GGGG....",
		"GGGBGBGBGG..",
		"GwwwwwwwwwKw",
		"G.wwwwwwww..",
		"......w.....",
	},
	madai: {
		"....rrr.....",
		"..rrrrrrr...",
		"r.rRrRrrrr..",
		"rrrrrrrrrrKr",
		"r.rrrrrrrrr.",
		"..rrrrrrr...",
		"....r.r.....",
	},
	hirame: {
		"..nnnnnnnnn...",
		"nnnNnnnNnnnnK.",
		"nnnnnNnnnnnnnK",
		"nnnnnnnnnnnnn.",
		"..nnnnnnnnn...",
	},
	ika: {
		".........pp.",
		"P.P.ppppppPp",
		"PPPPPKpppppW",
		"P.P.ppppppPp",
		".........pp.",
	},
	tachiuo: {
		"...wwwwwwwwwwwwwwww.",
		"wwwwwwwwwwwwwwwwwwKw",
		"....wwwwwwwwwwwww...",
	},
	buri: {
		"....bbbbbbbb....",
		"b.bbbbbbbbbbbb..",
		"bbyyyyyyyyyyyyKb",
		"b.wwwwwwwwwwww..",
		"....wwwwwwww....",
	},
	boot: {
		"..NNN..",
		"..NnN..",
		"..NnN..",
		"..NnN..",
		"NNNnN..",
		"NnnnnN.",
		"NNNNNNN",
	},
}

var boatArt = []string{
	"hhhhhhhhhhhhhhhhhhhhhhhhhh",
	".hhhhhhhhhhhhhhhhhhhhhhhh.",
	"..HHHHHHHHHHHHHHHHHHHHHH..",
	"...HHHHHHHHHHHHHHHHHHHH...",
}

var anglerArt = []string{
	"..yyy...",
	".yyyyy..",
	"..ccc...",
	"..ccc...",
	".uuuu...",
	"uuuuuuc.",
	".uuuu...",
	".uu.u...",
}

var weedArt = []string{
	".e.",
	"eE.",
	".eE",
	"Ee.",
	".eE",
	"eE.",
}

var rockArt = []string{
	"..kkk.",
	".kkkkk",
	"kkkkkk",
}

var lureArt = []string{"oWK", ".W."}

// glyphs is a 3x5 font for the gauges and the board.
var glyphs = map[byte][5]string{
	'0': {"###", "#.#", "#.#", "#.#", "###"},
	'1': {".#.", "##.", ".#.", ".#.", "###"},
	'2': {"###", "..#", "###", "#..", "###"},
	'3': {"###", "..#", "###", "..#", "###"},
	'4': {"#.#", "#.#", "###", "..#", "..#"},
	'5': {"###", "#..", "###", "..#", "###"},
	'6': {"###", "#..", "###", "#.#", "###"},
	'7': {"###", "..#", "..#", ".#.", ".#."},
	'8': {"###", "#.#", "###", "#.#", "###"},
	'9': {"###", "#.#", "###", "..#", "###"},
	'A': {".#.", "#.#", "###", "#.#", "#.#"},
	'B': {"##.", "#.#", "##.", "#.#", "##."},
	'C': {".##", "#..", "#..", "#..", ".##"},
	'D': {"##.", "#.#", "#.#", "#.#", "##."},
	'E': {"###", "#..", "##.", "#..", "###"},
	'F': {"###", "#..", "##.", "#..", "#.."},
	'G': {".##", "#..", "#.#", "#.#", ".##"},
	'H': {"#.#", "#.#", "###", "#.#", "#.#"},
	'I': {"###", ".#.", ".#.", ".#.", "###"},
	'J': {"..#", "..#", "..#", "#.#", ".#."},
	'K': {"#.#", "#.#", "##.", "#.#", "#.#"},
	'L': {"#..", "#..", "#..", "#..", "###"},
	'M': {"#.#", "###", "###", "#.#", "#.#"},
	'N': {"##.", "#.#", "#.#", "#.#", "#.#"},
	'O': {".#.", "#.#", "#.#", "#.#", ".#."},
	'P': {"##.", "#.#", "##.", "#..", "#.."},
	'Q': {".#.", "#.#", "#.#", "##.", ".##"},
	'R': {"##.", "#.#", "##.", "#.#", "#.#"},
	'S': {".##", "#..", ".#.", "..#", "##."},
	'T': {"###", ".#.", ".#.", ".#.", ".#."},
	'U': {"#.#", "#.#", "#.#", "#.#", "###"},
	'V': {"#.#", "#.#", "#.#", "#.#", ".#."},
	'W': {"#.#", "#.#", "###", "###", "#.#"},
	'X': {"#.#", "#.#", ".#.", "#.#", "#.#"},
	'Y': {"#.#", "#.#", ".#.", ".#.", ".#."},
	'Z': {"###", "..#", ".#.", "#..", "###"},
	':': {"...", ".#.", "...", ".#.", "..."},
	'.': {"...", "...", "...", "...", ".#."},
	'+': {"...", ".#.", "###", ".#.", "..."},
	'-': {"...", "...", "###", "...", "..."},
	'/': {"..#", "..#", ".#.", "#..", "#.."},
	'!': {".#.", ".#.", ".#.", "...", ".#."},
	'^': {".#.", "#.#", "...", "...", "..."},
	'v': {"...", "...", "...", "#.#", ".#."},
	'>': {"#..", ".#.", "..#", ".#.", "#.."},
	' ': {"...", "...", "...", "...", "..."},
}

const glyphOrder = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ:.+-/!^v> "

var (
	imageFish               [kinds]*ebiten.Image
	imageBoat, imageAngler  *ebiten.Image
	imageWeed, imageRock    *ebiten.Image
	imageLure, imageSun     *ebiten.Image
	imageGlyph              [256]*ebiten.Image
	imagePixel, imageBubble *ebiten.Image
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

func drawGlyph(img *ebiten.Image, ch byte, x, y int, c color.RGBA) {
	for gy, row := range glyphs[ch] {
		for gx := range len(row) {
			if row[gx] == '#' {
				img.Set(x+gx, y+gy, c)
			}
		}
	}
}

// sun is a white disc, tinted by the hour when drawn.
func sun() *ebiten.Image {
	img := ebiten.NewImage(9, 9)
	for y := range 9 {
		for x := range 9 {
			if (x-4)*(x-4)+(y-4)*(y-4) <= 17 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	return img
}

func init() {
	for k := range kinds {
		imageFish[k] = sprite(fishArt[k])
	}
	imageBoat = sprite(boatArt)
	imageAngler = sprite(anglerArt)
	imageWeed = sprite(weedArt)
	imageRock = sprite(rockArt)
	imageLure = sprite(lureArt)
	imageSun = sun()
	imageBubble = sprite([]string{".W.", "W.W", ".W."})
	imagePixel = ebiten.NewImage(1, 1)
	imagePixel.Fill(color.RGBA{255, 255, 255, 255})
	font := ebiten.NewImage(len(glyphOrder)*4, 5)
	for i := range len(glyphOrder) {
		drawGlyph(font, glyphOrder[i], i*4, 0, color.RGBA{255, 255, 255, 255})
	}
	// SubImage allocates, so the glyphs are cut once here.
	for i := range len(glyphOrder) {
		imageGlyph[glyphOrder[i]] = font.SubImage(image.Rect(i*4, 0, i*4+3, 5)).(*ebiten.Image)
	}
}
