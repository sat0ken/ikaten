package train

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent.
var palette = map[byte]color.RGBA{
	'K': {24, 24, 28, 255},
	'k': {70, 70, 76, 255},
	'g': {130, 130, 136, 255},
	'W': {240, 240, 236, 255},
	'R': {220, 40, 30, 255},
	'r': {120, 30, 24, 255},
	'Y': {250, 210, 40, 255},
	'B': {40, 80, 170, 255},
	'b': {120, 160, 210, 255},
	'G': {50, 120, 50, 255},
	'd': {30, 80, 36, 255},
	'T': {110, 74, 40, 255},
	'S': {200, 190, 170, 255},
	's': {160, 150, 132, 255},
	'O': {230, 120, 40, 255},
	'C': {140, 60, 50, 255},
	'c': {90, 96, 120, 255},
}

var treeArt = []string{
	"...GGG...",
	"..GGdGG..",
	".GGGGGdG.",
	"GGdGGGGGG",
	"GGGGGdGGG",
	".GdGGGGG.",
	"..GGGdG..",
	"....T....",
	"....T....",
}

var pineArt = []string{
	"...d...",
	"..dGd..",
	"..GdG..",
	".dGGGd.",
	".GdGdG.",
	"dGGdGGd",
	"...T...",
	"...T...",
}

var houseArt = [2][]string{{
	"....CCCCCCCC....",
	"...CCCCCCCCCC...",
	"..CCCCCCCCCCCC..",
	".CCCCCCCCCCCCCC.",
	"..SSSSSSSSSSSS..",
	"..SbbSSSSSSbbS..",
	"..SbbSSSSSSbbS..",
	"..SSSSSTTSSSSS..",
	"..SSSSSTTSSSSS..",
}, {
	"cccccccccccc",
	"cSSSSSSSSSSc",
	"cSbbSbbSbbSc",
	"cSSSSSSSSSSc",
	"cSbbSbbSbbSc",
	"cSSSSSSSSSSc",
	"cSbbSbbSbbSc",
	"cSSSSSSSSSSc",
	"cSSSSTTSSSSc",
	"cSSSSTTSSSSc",
}}

// The board that marks where the front of the train stops.
var stopMarkArt = []string{
	"KKKKKKK",
	"KYYYYYK",
	"KYKKKYK",
	"KYYKYYK",
	"KYYKYYK",
	"KYYYYYK",
	"KKKKKKK",
	"...k...",
	"...k...",
	"...k...",
	"...k...",
}

var whistleArt = []string{
	"KKKKKKK",
	"KWWWWWK",
	"KWKWKWK",
	"KWKWKWK",
	"KWWKWWK",
	"KWWWWWK",
	"KKKKKKK",
	"...k...",
	"...k...",
	"...k...",
	"...k...",
	"...k...",
}

var crossingArt = [2][]string{{
	"WR.RW",
	"RRkRR",
	"RRkRR",
	"..k..",
	"YKYKY",
	"..k..",
	"..k..",
	"..k..",
	"..k..",
	"..k..",
}, {
	"Wr.rW",
	"rrkrr",
	"rrkrr",
	"..k..",
	"KYKYK",
	"..k..",
	"..k..",
	"..k..",
	"..k..",
	"..k..",
}}

// The roadside warning of a lit crossing: lights alternate.
var lampArt = [2][]string{{"RR..", "RR.."}, {"..RR", "..RR"}}

var bufferArt = []string{
	"RWRWRWRWRW",
	"WRWRWRWRWR",
	"kkkkkkkkkk",
	"k..k..k..k",
	"k..k..k..k",
}

var nameBoardArt = []string{
	"kkkkkkkkkkkkkkkkkk",
	"kWWWWWWWWWWWWWWWWk",
	"kWWKKWKKKWKWKKWWWk",
	"kWWWWWWWWWWWWWWWWk",
	"kBBBBBBBBBBBBBBBBk",
	"kWWWWWWWWWWWWWWWWk",
	"kkkkkkkkkkkkkkkkkk",
	"...k..........k...",
	"...k..........k...",
}

// glyphs is a 3x5 font for the gauges.
var glyphs = map[byte][5]string{
	'0':  {"###", "#.#", "#.#", "#.#", "###"},
	'1':  {".#.", "##.", ".#.", ".#.", "###"},
	'2':  {"###", "..#", "###", "#..", "###"},
	'3':  {"###", "..#", "###", "..#", "###"},
	'4':  {"#.#", "#.#", "###", "..#", "..#"},
	'5':  {"###", "#..", "###", "..#", "###"},
	'6':  {"###", "#..", "###", "#.#", "###"},
	'7':  {"###", "..#", "..#", ".#.", ".#."},
	'8':  {"###", "#.#", "###", "#.#", "###"},
	'9':  {"###", "#.#", "###", "..#", "###"},
	':':  {"...", ".#.", "...", ".#.", "..."},
	'.':  {"...", "...", "...", "...", ".#."},
	'+':  {"...", ".#.", "###", ".#.", "..."},
	'-':  {"...", "...", "###", "...", "..."},
	'm':  {"...", "#.#", "###", "#.#", "#.#"},
	's':  {"...", ".##", "##.", "..#", "##."},
	'A':  {".#.", "#.#", "###", "#.#", "#.#"},
	'C':  {".##", "#..", "#..", "#..", ".##"},
	'D':  {"##.", "#.#", "#.#", "#.#", "##."},
	'S':  {".##", "#..", ".#.", "..#", "##."},
	'P':  {"##.", "#.#", "##.", "#..", "#.."},
	'B':  {"##.", "#.#", "##.", "#.#", "##."},
	'N':  {"##.", "#.#", "#.#", "#.#", "#.#"},
	'E':  {"###", "#..", "##.", "#..", "###"},
	'>':  {"#..", ".#.", "..#", ".#.", "#.."},
	'/':  {"..#", "..#", ".#.", "#..", "#.."},
	'\\': {"#..", "#..", ".#.", "..#", "..#"},
	' ':  {"...", "...", "...", "...", "..."},
}

const glyphOrder = "0123456789:.+-msACDSPBNE>/\\ "

const (
	hillWidth  = 160
	hillHeight = 12
)

var (
	imageTree, imagePine, imageStopMark   *ebiten.Image
	imageWhistle, imageBuffer, imageBoard *ebiten.Image
	imageHouse                            [2]*ebiten.Image
	imageCrossing, imageLamp              [2]*ebiten.Image
	imageHills                            *ebiten.Image
	imageGlyph                            [256]*ebiten.Image
	imageLimit                            [100]*ebiten.Image // speed limit signs by km/h
	imageDistance                         [4]*ebiten.Image   // boards 100, 200 and 300 m before a stop
	imagePixel                            *ebiten.Image      // white, tinted with ColorScale
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

// drawGlyph writes a character into img in the 3x5 font.
func drawGlyph(img *ebiten.Image, ch byte, x, y int, c color.RGBA) {
	for gy, row := range glyphs[ch] {
		for gx := range len(row) {
			if row[gx] == '#' {
				img.Set(x+gx, y+gy, c)
			}
		}
	}
}

// limitSign is a round sign, red ring on white, with the limit in km/h.
func limitSign(kmh int) *ebiten.Image {
	img := ebiten.NewImage(11, 16)
	for y := range 11 {
		for x := range 11 {
			switch d := (x-5)*(x-5) + (y-5)*(y-5); {
			case d <= 15:
				img.Set(x, y, palette['W'])
			case d <= 31:
				img.Set(x, y, palette['R'])
			}
		}
	}
	drawGlyph(img, byte('0'+kmh/10), 2, 3, palette['K'])
	drawGlyph(img, byte('0'+kmh%10), 6, 3, palette['K'])
	for y := 11; y < 16; y++ {
		img.Set(5, y, palette['k'])
	}
	return img
}

// distanceBoard counts down to a stop mark in hundreds of meters.
func distanceBoard(n int) *ebiten.Image {
	img := ebiten.NewImage(7, 12)
	for y := range 7 {
		for x := range 7 {
			c := palette['W']
			if x == 0 || y == 0 || x == 6 || y == 6 {
				c = palette['B']
			}
			img.Set(x, y, c)
		}
	}
	drawGlyph(img, byte('0'+n), 2, 1, palette['K'])
	for y := 7; y < 12; y++ {
		img.Set(3, y, palette['k'])
	}
	return img
}

// isin returns the sine of a degrees in thousandths, by Bhaskara's formula
// (within 0.002). The hills are drawn with integers: TinyGo then draws them
// while compiling, into the program, instead of into the heap at start.
func isin(a int) int {
	a %= 360
	if a < 0 {
		a += 360
	}
	sign := 1
	if a >= 180 {
		a, sign = a-180, -1
	}
	p := a * (180 - a)
	return sign * 4000 * p / (40500 - p)
}

// hills is a strip of distant hills that repeats every hillWidth pixels.
func hills() *ebiten.Image {
	img := ebiten.NewImage(hillWidth, hillHeight)
	// TinyGo runs a loop at compile time only up to 1000 iterations, so the
	// strip is drawn in parts.
	for x := 0; x < hillWidth; x += 40 {
		hillColumns(img, x, x+40)
	}
	return img
}

func hillColumns(img *ebiten.Image, from, to int) {
	far, near := color.RGBA{120, 150, 170, 255}, color.RGBA{96, 130, 120, 255}
	for x := from; x < to; x++ {
		t := x * 360 / hillWidth
		// Heights in thousandths of a pixel.
		h1 := 7000 + 3*isin(t*2+29) + 3*isin(t*5)/2
		h2 := 3000 + 2*isin(t*3+115) + isin(t*7+57)
		for y := range hillHeight {
			switch up := (hillHeight - y) * 1000; {
			case up <= h2:
				img.Set(x, y, near)
			case up <= h1:
				img.Set(x, y, far)
			}
		}
	}
}

func init() {
	imageTree = sprite(treeArt)
	imagePine = sprite(pineArt)
	imageStopMark = sprite(stopMarkArt)
	imageWhistle = sprite(whistleArt)
	imageBuffer = sprite(bufferArt)
	imageBoard = sprite(nameBoardArt)
	for i := range 2 {
		imageHouse[i] = sprite(houseArt[i])
		imageCrossing[i] = sprite(crossingArt[i])
		imageLamp[i] = sprite(lampArt[i])
	}
	imageHills = hills()
	font := ebiten.NewImage(len(glyphOrder)*4, 5)
	for i := range len(glyphOrder) {
		drawGlyph(font, glyphOrder[i], i*4, 0, color.RGBA{255, 255, 255, 255})
	}
	// SubImage allocates, so the glyphs are cut once here.
	for i := range len(glyphOrder) {
		imageGlyph[glyphOrder[i]] = font.SubImage(image.Rect(i*4, 0, i*4+3, 5)).(*ebiten.Image)
	}
	for _, l := range limits {
		if imageLimit[l.kmh] == nil {
			imageLimit[l.kmh] = limitSign(l.kmh)
		}
	}
	for n := 1; n <= 3; n++ {
		imageDistance[n] = distanceBoard(n)
	}
	imagePixel = ebiten.NewImage(1, 1)
	imagePixel.Fill(color.RGBA{255, 255, 255, 255})
}
