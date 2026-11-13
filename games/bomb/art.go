package bomb

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. In the
// gophers, 'B' and 'D' are the body and its shade, which differ for each
// gopher; the rest is the same for all.
var gopherArt = [3][2][]string{
	down: {{
		".D.....D.",
		".BBBBBBB.",
		"BWWBBBWWB",
		"BWKBBBKWB",
		"BBBBNBBBB",
		"BBBBwBBBB",
		"hBBBBBBBh",
		".BBBBBBB.",
		".BBBBBBB.",
		"..BBBBB..",
		".f.....f.",
	}, {
		".D.....D.",
		".BBBBBBB.",
		"BWWBBBWWB",
		"BWKBBBKWB",
		"BBBBNBBBB",
		"BBBBwBBBB",
		"BBBBBBBBB",
		"hBBBBBBBh",
		".BBBBBBB.",
		"..BBBBB..",
		"..f...f..",
	}},
	up: {{
		".D.....D.",
		".BBBBBBB.",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"hBBBBBBBh",
		".BBBBBBB.",
		".BBBBBBB.",
		"..BBDBB..",
		".f.....f.",
	}, {
		".D.....D.",
		".BBBBBBB.",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"BBBBBBBBB",
		"hBBBBBBBh",
		".BBBBBBB.",
		"..BBDBB..",
		"..f...f..",
	}},
	left: {{ // drawn facing right, flipped for left
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
	}, {
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
	}},
}

var faintArt = []string{
	".D.....D.",
	".BBBBBBB.",
	"BBBBBBBBB",
	"BKKBBBKKB",
	"BBBBNBBBB",
	"BBBBwBBBB",
	"hBBBBBBBh",
	".BBBBBBB.",
	".BBBBBBB.",
	"..BBBBB..",
	".f.....f.",
}

// The gophers' colors: body and shade. The person's is the tan gopher.
var gopherColors = [players][2]color.RGBA{
	{{214, 186, 150, 255}, {176, 146, 110, 255}},
	{{120, 200, 230, 255}, {80, 160, 196, 255}},
	{{246, 160, 190, 255}, {210, 120, 150, 255}},
	{{160, 214, 120, 255}, {116, 172, 84, 255}},
}

var gopherNames = [players]string{"YOU", "BLUE", "PINK", "LIME"}

var palette = map[byte]color.RGBA{
	'K': {20, 20, 24, 255},
	'W': {250, 250, 250, 255},
	'w': {245, 245, 240, 255},
	'N': {60, 40, 34, 255},
	'h': {238, 214, 186, 255},
	'f': {238, 214, 186, 255},
	'g': {120, 124, 132, 255},
	'G': {170, 174, 182, 255},
	'k': {70, 72, 80, 255},
	'L': {212, 216, 222, 255},
	'r': {180, 90, 50, 255},
	'R': {220, 130, 70, 255},
	'm': {120, 60, 36, 255},
	'b': {30, 30, 36, 255},
	'x': {90, 90, 100, 255},
	'y': {255, 230, 80, 255},
	'o': {255, 140, 30, 255},
	'O': {230, 60, 30, 255},
	'c': {120, 200, 255, 255},
	'p': {250, 244, 220, 255},
}

var wallArt = []string{
	"LLLLLLLLk",
	"LGGGGGGgk",
	"LGGGGGGgk",
	"LGGgGGGgk",
	"LGGGGGGgk",
	"LGGGGgGgk",
	"LGGGGGGgk",
	"Lgggggggk",
	"kkkkkkkkk",
}

var blockArt = []string{
	"RRRRmRRRR",
	"rrrrmrrrr",
	"mmmmmmmmm",
	"RRmRRRRmR",
	"rrmrrrrmr",
	"mmmmmmmmm",
	"RRRRmRRRR",
	"rrrrmrrrr",
	"mmmmmmmmm",
}

var crumbleArt = []string{
	"R.R.m.R.R",
	".r.rm.r..",
	"m.m.m.m.m",
	"R.m.R.R.R",
	".rm.r.rm.",
	"m.m.m.m.m",
	"R.R.m.R.R",
	".r.rm.r..",
	"m.m.m.m.m",
}

var bombArt = [2][]string{{
	".....y...",
	"....m....",
	"..bbbb...",
	".bxbbbb..",
	"bxbbbbbb.",
	"bbbbbbbb.",
	"bbbbbbbb.",
	".bbbbbb..",
	"..bbbb...",
}, {
	"....o.o..",
	"....my...",
	"..bbbb...",
	".bxbbbb..",
	"bxbbbbbb.",
	"bbbbbbbb.",
	"bbbbbbbb.",
	".bbbbbb..",
	"..bbbb...",
}}

var itemArt = [4][]string{
	fireUp: {
		"ppppppppp",
		"pppppOppp",
		"ppppOOppp",
		"pppOOoOpp",
		"ppOOooOpp",
		"ppOoyyOpp",
		"ppOoyyOpp",
		"pppOOOppp",
		"ppppppppp",
	},
	bombUp: {
		"ppppppppp",
		"ppppyppbp",
		"pppmpppbp",
		"ppbbbpbbb",
		"pbxbbbpbp",
		"pbbbbbpbp",
		"pbbbbbppp",
		"ppbbbpppp",
		"ppppppppp",
	},
	speedUp: {
		"ppppppppp",
		"pppppcccp",
		"ppppcccpp",
		"pppcccppp",
		"ppcccccpp",
		"pppppccpp",
		"ppppccppp",
		"pppccpppp",
		"ppppppppp",
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
	' ': {"...", "...", "...", "...", "..."},
}

const glyphOrder = "0123456789ABCDEFGHIKLMNOPRSTUVWXY:!. "

var (
	imageGopher           [players][3][2]*ebiten.Image
	imageFaint            [players]*ebiten.Image
	imageHead             [players]*ebiten.Image // the top of each gopher, for the score bar
	imageWall, imageBlock *ebiten.Image
	imageCrumble          *ebiten.Image
	imageBomb             [2]*ebiten.Image
	imageItem             [4]*ebiten.Image
	imageGlyph            [256]*ebiten.Image
)

// sprite draws pixel art, with the body colors of gopher i.
func sprite(rows []string, i int) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	for y, row := range rows {
		spriteRow(img, row, y, i)
	}
	return img
}

func spriteRow(img *ebiten.Image, row string, y, i int) {
	for x := range len(row) {
		switch row[x] {
		case 'B':
			img.Set(x, y, gopherColors[i][0])
		case 'D':
			img.Set(x, y, gopherColors[i][1])
		default:
			if c, ok := palette[row[x]]; ok {
				img.Set(x, y, c)
			}
		}
	}
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
	for i := range players {
		for v := range 3 {
			for f := range 2 {
				imageGopher[i][v][f] = sprite(gopherArt[v][f], i)
			}
		}
		imageFaint[i] = sprite(faintArt, i)
		// SubImage allocates, so the heads are cut once here.
		imageHead[i] = imageGopher[i][down][0].SubImage(image.Rect(0, 0, 9, 9)).(*ebiten.Image)
	}
	imageWall = sprite(wallArt, 0)
	imageBlock = sprite(blockArt, 0)
	imageCrumble = sprite(crumbleArt, 0)
	for f := range 2 {
		imageBomb[f] = sprite(bombArt[f], 0)
	}
	for k := fireUp; k <= speedUp; k++ {
		imageItem[k] = sprite(itemArt[k], 0)
	}
	font := ebiten.NewImage(len(glyphOrder)*4, 5)
	for i := range len(glyphOrder) {
		drawGlyph(font, glyphOrder[i], i*4, 0, color.RGBA{255, 255, 255, 255})
	}
	// SubImage allocates, so the glyphs are cut once here.
	for i := range len(glyphOrder) {
		imageGlyph[glyphOrder[i]] = font.SubImage(image.Rect(i*4, 0, i*4+3, 5)).(*ebiten.Image)
	}
}
