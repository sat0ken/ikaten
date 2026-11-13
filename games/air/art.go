package air

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Top views
// have the nose or bow to the right and the right wing or starboard side at
// the bottom; side views have the top up.
var palette = map[byte]color.RGBA{
	'K': {20, 22, 26, 255},
	'k': {60, 62, 68, 255},
	'g': {120, 124, 130, 255},
	'G': {150, 154, 160, 255},
	'w': {200, 204, 210, 255},
	'W': {244, 246, 250, 255},
	'B': {54, 74, 104, 255},
	'b': {84, 104, 134, 255},
	'c': {150, 200, 230, 255},
	'D': {160, 120, 76, 255},
	'd': {120, 88, 56, 255},
	'r': {120, 40, 36, 255},
	'R': {220, 50, 30, 255},
	'O': {255, 150, 40, 255},
	'Y': {255, 226, 90, 255},
	's': {110, 110, 116, 255},
	'S': {150, 150, 156, 255},
}

var fighterTop = []string{
	"....bbb..",
	"....BBB..",
	"....BBB..",
	".b..BBB..",
	".B..BBB..",
	"BBBBBcBBK",
	".B..BBB..",
	".b..BBB..",
	"....BBB..",
	"....BBB..",
	"....bbb..",
}

var fighterSide = []string{
	"B....c...",
	"BBBBBBBBK",
	".bBBBBBK.",
}

var bomberTop = []string{
	".....bbb..",
	".....BBB..",
	".....BBB..",
	".....BBB..",
	".b...BBB..",
	".B...BBB..",
	"BBBBcBcBBK",
	".B...BBB..",
	".b...BBB..",
	".....BBB..",
	".....BBB..",
	".....BBB..",
	".....bbb..",
}

var bomberSide = []string{
	"B...ccc...",
	"BBBBBBBBBK",
	".bBBBBBBK.",
}

// Decks, seen from above.
var deckArt = [shipKinds][]string{
	carrier: {
		"...DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD.....",
		".DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD..",
		"DDWDDDDDDdDDDDDDDDdDDDDDDDDdDDDDDDDDWDDDDD",
		"DDWDDDDDDdDDDDDDDDdDDDDDDDDdDDDDDDDDWDDDDD",
		".DDDDDDDDDDDDDDDDDDDDDDDggggDDDDDDDDDDDD..",
		"...DDDDDDDDDDDDDDDDDDDDDggggDDDDDDDDDD....",
	},
	cruiser: {
		"...ggggggggggggggggggggggggg..",
		".gggggKggKgggsssssggggKggKggg.",
		".gggggKggKgggsssssggggKggKgggg",
		"...ggggggggggggggggggggggggg..",
	},
	destroyer: {
		"..gggggggggggggggggg..",
		".ggKgggsssgsssgggKgggg",
		"..gggggggggggggggggg..",
	},
}

// Superstructures above the deck, seen from the side.
var upperArt = [shipKinds][]string{
	carrier: {
		"........................k.................",
		"........................k.................",
		".......................sss................",
		".......................sss................",
		".......................sss................",
	},
	cruiser: {
		"...............k..............",
		"..............kk..k...........",
		".............sss..kk..........",
		"......g.g...ssssssss...g.g....",
		".....gggggggssssssssgggggg....",
	},
	destroyer: {
		"...........k..........",
		"......s...kk..........",
		"...g.sss.sss.ss...g...",
		"..ggggssssssssssggggg.",
	},
}

// The hull's side above the water, stretched to each ship's length.
var hullArt = []string{
	"kkkkkkkkkkkkkkkkkkkkkkkk",
	".kkkkkkkkkkkkkkkkkkkkkk.",
	"..rrrrrrrrrrrrrrrrrrrr..",
}

// The wake behind a ship on the sea, stern to the right.
var wakeArt = []string{
	"w...............",
	".w..............",
	"..ww............",
	"....wwwwwwwWWWWW",
	"..ww............",
	".w..............",
	"w...............",
}

var cloudArt = []string{
	".........WWWW...........",
	"......WWWWWWWWWW........",
	"...WWWWWWWWWWWWWWWW.....",
	".WWWWWWWWWWWWWWWWWWWWW..",
	"WWWWWWWWWWWWWWWWWWWWWWWW",
	"wwwwwWWWWWWWWWWWWWWwwwww",
	"..wwwwwwwwwwwwwwwwwwww..",
}

var puffArt = [4][]string{
	puffFlak:   {".kKk.", "kKKKk", "KKKKK", "kKKKk", ".kKk."},
	puffSmoke:  {".sSs.", "sSSSs", "SSSSS", "sSSSs", ".sSs."},
	puffFire:   {"..Y..", ".YOY.", "YORrO", ".ORO.", "..R.."},
	puffSplash: {"..W..", ".WWW.", ".WwW.", "WWwWW", "WwWwW", "WWWWW"},
}

// glyphs is a 3x5 font for the instruments.
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
	'%': {"#.#", "..#", ".#.", "#..", "#.#"},
	'=': {"...", "###", "...", "###", "..."},
	'>': {"#..", ".#.", "..#", ".#.", "#.."},
	' ': {"...", "...", "...", "...", "..."},
}

const glyphOrder = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ:.+-/%=> "

var (
	imageFighter, imageBomber [2]*ebiten.Image // top and side
	imageDeck, imageUpper     [shipKinds]*ebiten.Image
	imageHull, imageWake      *ebiten.Image
	imageCloud, imageCowl     *ebiten.Image
	imagePuff                 [4]*ebiten.Image
	imageGlyph                [256]*ebiten.Image
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

// cowl is the top of the engine cowling seen from the cockpit: half an
// ellipse with a lighter rim. It is drawn with integers in rows, so that
// TinyGo draws it while compiling.
func cowl() *ebiten.Image {
	const w, h = 64, 9
	img := ebiten.NewImage(w, h)
	for y := range h {
		cowlRow(img, y)
	}
	return img
}

func cowlRow(img *ebiten.Image, y int) {
	const w, h = 64, 9
	dark, rim := color.RGBA{34, 38, 34, 255}, color.RGBA{70, 76, 66, 255}
	dy := h - y
	for x := range w {
		dx := x*2 - w + 1
		// Inside (dx/w)² + (dy/h)² < 1, in integers.
		v := dx*dx*h*h + 4*dy*dy*w*w
		switch {
		case v < 4*w*w*h*h*80/100:
			img.Set(x, y, dark)
		case v < 4*w*w*h*h:
			img.Set(x, y, rim)
		}
	}
}

func init() {
	imageFighter = [2]*ebiten.Image{sprite(fighterTop), sprite(fighterSide)}
	imageBomber = [2]*ebiten.Image{sprite(bomberTop), sprite(bomberSide)}
	for k := range shipKinds {
		imageDeck[k] = sprite(deckArt[k])
		imageUpper[k] = sprite(upperArt[k])
	}
	imageHull = sprite(hullArt)
	imageWake = sprite(wakeArt)
	imageCloud = sprite(cloudArt)
	for i := range imagePuff {
		imagePuff[i] = sprite(puffArt[i])
	}
	imageCowl = cowl()
	font := ebiten.NewImage(len(glyphOrder)*4, 5)
	for i := range len(glyphOrder) {
		drawGlyph(font, glyphOrder[i], i*4, 0, color.RGBA{255, 255, 255, 255})
	}
	// SubImage allocates, so the glyphs are cut once here.
	for i := range len(glyphOrder) {
		imageGlyph[glyphOrder[i]] = font.SubImage(image.Rect(i*4, 0, i*4+3, 5)).(*ebiten.Image)
	}
}
