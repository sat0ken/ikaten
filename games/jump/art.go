package jump

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. The gopher
// faces right and is flipped for left.
var gopherArt = [5][]string{
	{ // standing
		"..D......",
		".BBBBBB..",
		"BBBBBWWB.",
		"BBBBBWKB.",
		"BBBBBBBBN",
		"BBBBBBBw.",
		"BBBhBBBB.",
		".BBBBBBB.",
		".BBBBBBB.",
		"..BBBBB..",
		".f....f..",
	},
	{ // walking
		"..D......",
		".BBBBBB..",
		"BBBBBWWB.",
		"BBBBBWKB.",
		"BBBBBBBBN",
		"BBBBBBBw.",
		"BBBBBBBB.",
		".BBBhBBB.",
		".BBBBBBB.",
		"..BBBBB..",
		"..f..f...",
	},
	{ // walking
		"..D......",
		".BBBBBB..",
		"BBBBBWWB.",
		"BBBBBWKB.",
		"BBBBBBBBN",
		"BBBBBBBw.",
		"BBhBBBBB.",
		".BBBBBBB.",
		".BBBBBBB.",
		"..BBBBB..",
		"f......f.",
	},
	{ // jumping: arms up
		"h.D....h.",
		"hBBBBBBh.",
		"BBBBBWWB.",
		"BBBBBWKB.",
		"BBBBBBBBN",
		"BBBBBBBw.",
		"BBBBBBBB.",
		".BBBBBBB.",
		".BBBBBBB.",
		"..BBBBB..",
		"..f.f....",
	},
	{ // dashing: leaning in
		".........",
		"...D.....",
		"..BBBBBB.",
		".BBBBBWWB",
		"BBBBBBWKB",
		"BBBBBBBBN",
		"BBBBBBBwh",
		"BBBBBBBB.",
		".BBBBBB..",
		"..BBBB...",
		"ff.......",
	},
}

var palette = map[byte]color.RGBA{
	'B': {214, 186, 150, 255},
	'D': {176, 146, 110, 255},
	'K': {20, 20, 24, 255},
	'W': {250, 250, 250, 255},
	'w': {245, 245, 240, 255},
	'N': {60, 40, 34, 255},
	'h': {238, 214, 186, 255},
	'f': {238, 214, 186, 255},
	'r': {200, 60, 60, 255},
	'R': {240, 110, 100, 255},
	'k': {30, 30, 40, 255},
	'p': {130, 80, 170, 255},
	'P': {170, 120, 210, 255},
	'y': {255, 220, 70, 255},
	'o': {255, 150, 40, 255},
	'O': {200, 100, 30, 255},
	'g': {140, 140, 150, 255},
	'G': {60, 200, 90, 255},
	'c': {80, 220, 255, 255},
	'C': {200, 250, 255, 255},
	'e': {40, 150, 200, 255},
	'S': {210, 214, 222, 255},
	's': {150, 154, 164, 255},
	'm': {250, 246, 236, 255},
	'n': {210, 200, 180, 255},
	'v': {245, 248, 252, 255},
	'V': {200, 214, 232, 255},
}

var bugArt = [2][]string{{
	"...kk...",
	"..pPPp..",
	".pPpPPp.",
	"kpPPpPpk",
	".pppppp.",
	".k.k.k..",
}, {
	"...kk...",
	"..pPPp..",
	".pPpPPp.",
	"kpPPpPpk",
	".pppppp.",
	"..k.k.k.",
}}

var spikyArt = [2][]string{{
	".y.y.y..",
	"yoyoyoy.",
	"oOOOOOo.",
	"kOOOOOOk",
	".oooooo.",
	".k.k.k..",
}, {
	".y.y.y..",
	"yoyoyoy.",
	"oOOOOOo.",
	"kOOOOOOk",
	".oooooo.",
	"..k.k.k.",
}}

var mothArt = [2][]string{{
	"mm....mm",
	"mnm..mnm",
	".mnkkmn.",
	"..mkkm..",
	"...kk...",
}, {
	"........",
	"........",
	"mmnkknmm",
	".mmkkmm.",
	"...kk...",
}}

var squashArt = []string{
	"........",
	"........",
	"........",
	"........",
	".kpppk..",
	"kpppppk.",
}

var gemArt = [2][]string{{
	"..CC..",
	".cCcc.",
	"cccccc",
	"eccce.",
	".eee..",
	"..e...",
}, {
	"..CC..",
	".CCcc.",
	"cCCccc",
	"ecccee",
	".eee..",
	"..e...",
}}

var spikeArt = []string{
	"........",
	"........",
	"..S...S.",
	".SSs.SSs",
	".SSs.SSs",
	"SSSsSSSs",
	"SSSsSSSs",
	"ssssssss",
}

var cloudArt = []string{
	".vvvvvv.",
	"vvvvvvvv",
	"VVVVVVVV",
	"........",
	"........",
	"........",
	"........",
	"........",
}

// Each stage colors its ground, its blocks and its dirt.
type look struct {
	sky, horizon      color.RGBA
	far, near         color.RGBA // the hills or the rocks behind
	top, ground, edge color.RGBA
	block, blockEdge  color.RGBA
	dirt, dirtSpot    color.RGBA
}

var looks = [3]look{
	{ // meadow
		sky: color.RGBA{110, 180, 240, 255}, horizon: color.RGBA{190, 226, 250, 255},
		far: color.RGBA{140, 200, 170, 255}, near: color.RGBA{96, 170, 110, 255},
		top: color.RGBA{90, 190, 80, 255}, ground: color.RGBA{150, 100, 60, 255}, edge: color.RGBA{110, 70, 40, 255},
		block: color.RGBA{200, 150, 90, 255}, blockEdge: color.RGBA{130, 90, 50, 255},
		dirt: color.RGBA{170, 120, 70, 255}, dirtSpot: color.RGBA{120, 80, 44, 255},
	},
	{ // cave
		sky: color.RGBA{34, 26, 40, 255}, horizon: color.RGBA{60, 44, 56, 255},
		far: color.RGBA{50, 40, 54, 255}, near: color.RGBA{70, 54, 66, 255},
		top: color.RGBA{120, 110, 130, 255}, ground: color.RGBA{86, 76, 96, 255}, edge: color.RGBA{56, 48, 64, 255},
		block: color.RGBA{130, 120, 140, 255}, blockEdge: color.RGBA{80, 70, 90, 255},
		dirt: color.RGBA{150, 100, 60, 255}, dirtSpot: color.RGBA{100, 66, 40, 255},
	},
	{ // sky
		sky: color.RGBA{70, 140, 230, 255}, horizon: color.RGBA{170, 210, 250, 255},
		far: color.RGBA{220, 236, 252, 255}, near: color.RGBA{240, 246, 255, 255},
		top: color.RGBA{110, 210, 120, 255}, ground: color.RGBA{170, 130, 90, 255}, edge: color.RGBA{120, 86, 56, 255},
		block: color.RGBA{230, 210, 130, 255}, blockEdge: color.RGBA{170, 140, 70, 255},
		dirt: color.RGBA{190, 140, 90, 255}, dirtSpot: color.RGBA{140, 96, 56, 255},
	},
}

// glyphs is a 3x5 font for the board.
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
	'R': {"##.", "#.#", "##.", "#.#", "#.#"},
	'S': {".##", "#..", ".#.", "..#", "##."},
	'T': {"###", ".#.", ".#.", ".#.", ".#."},
	'U': {"#.#", "#.#", "#.#", "#.#", "###"},
	'V': {"#.#", "#.#", "#.#", "#.#", ".#."},
	'W': {"#.#", "#.#", "###", "###", "#.#"},
	'X': {"#.#", "#.#", ".#.", "#.#", "#.#"},
	'Y': {"#.#", "#.#", ".#.", ".#.", ".#."},
	':': {"...", ".#.", "...", ".#.", "..."},
	'!': {".#.", ".#.", ".#.", "...", ".#."},
	'.': {"...", "...", "...", "...", ".#."},
	'-': {"...", "...", "###", "...", "..."},
	' ': {"...", "...", "...", "...", "..."},
}

const glyphOrder = "0123456789ABCDEFGHIJKLMNOPRSTUVWXY:!.- "

var (
	imageGopher            [5]*ebiten.Image
	imageBug, imageSpiky   [2]*ebiten.Image
	imageMoth, imageGem    [2]*ebiten.Image
	imageSquash            *ebiten.Image
	imageSpike, imageCloud *ebiten.Image
	imageHeart             [2]*ebiten.Image // full and empty
	imageGlyph             [256]*ebiten.Image
)

var heartArt = [2][]string{{
	".rr.rr.",
	"rRrrrrr",
	"rrrrrrr",
	".rrrrr.",
	"..rrr..",
	"...r...",
}, {
	".kk.kk.",
	"k..k..k",
	"k.....k",
	".k...k.",
	"..k.k..",
	"...k...",
}}

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

func init() {
	for i := range imageGopher {
		imageGopher[i] = sprite(gopherArt[i])
	}
	for f := range 2 {
		imageBug[f] = sprite(bugArt[f])
		imageSpiky[f] = sprite(spikyArt[f])
		imageMoth[f] = sprite(mothArt[f])
		imageGem[f] = sprite(gemArt[f])
		imageHeart[f] = sprite(heartArt[f])
	}
	imageSquash = sprite(squashArt)
	imageSpike = sprite(spikeArt)
	imageCloud = sprite(cloudArt)
	font := ebiten.NewImage(len(glyphOrder)*4, 5)
	for i := range len(glyphOrder) {
		drawGlyph(font, glyphOrder[i], i*4, 0, color.RGBA{255, 255, 255, 255})
	}
	// SubImage allocates, so the glyphs are cut once here.
	for i := range len(glyphOrder) {
		imageGlyph[glyphOrder[i]] = font.SubImage(image.Rect(i*4, 0, i*4+3, 5)).(*ebiten.Image)
	}
}
