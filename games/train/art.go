package train

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

// Sprites are pixel art written as strings; '.' is transparent.
var palette = map[byte]color.RGBA{
	'B': {40, 80, 170, 255},
	'C': {150, 64, 52, 255},
	'D': {24, 60, 30, 255},
	'G': {60, 130, 56, 255},
	'K': {24, 24, 28, 255},
	'L': {110, 170, 90, 255},
	'R': {220, 40, 30, 255},
	'S': {200, 190, 170, 255},
	'T': {110, 74, 40, 255},
	'W': {240, 240, 236, 255},
	'Y': {250, 210, 40, 255},
	'b': {120, 160, 210, 255},
	'c': {90, 96, 120, 255},
	'd': {34, 90, 40, 255},
	'g': {130, 130, 136, 255},
	'k': {70, 70, 76, 255},
	'r': {120, 30, 24, 255},
	's': {160, 150, 132, 255},
	't': {140, 100, 60, 255},
}

var treeArt = []string{
	"......DDDDDD......",
	"....DDLGGGGGDD....",
	"...DLLLLLGGGGGD...",
	"..DGLLLLLGGGGGGD..",
	".DGLdLLLLLGGGGGGD.",
	".DGGLLLLLGdGGGGdD.",
	".DGdLLLLLGGGGGddD.",
	".DGGGGdGGGGdGdddD.",
	".DGGGGGGGGGGddddD.",
	".DGddGGGGGGdddddD.",
	".DGGGGGGGGddddddD.",
	"..DGGGGGGddGdddD..",
	"...DddddddddddD...",
	"....DDddddddDD....",
	"......DDtTDD......",
	"........tT........",
	"........TT........",
	"........TT........",
	"........TT........",
	"........TT........",
}

var pineArt = []string{
	"......DD......",
	"......DD......",
	"......DD......",
	".....DGdD.....",
	".....DGdD.....",
	".....DGdD.....",
	"....DGGddD....",
	".....DGdD.....",
	"....DGGddD....",
	"....DGGddD....",
	"...DGGGdddD...",
	"...DGGGdddD...",
	"...DGGGdddD...",
	"...DGGGdddD...",
	"..DGGGGddddD..",
	"..DGGGGddddD..",
	".DDDDDGdDDDDD.",
	"......TT......",
	"......TT......",
	"......TT......",
	"......TT......",
	"......TT......",
}

var houseArt = [2][]string{{
	"........CCCCCCCCCCCCCCCC........",
	".......CCCCCCCCCCCCCCCCCC.......",
	"......CrCCrCCrCCrCCrCCrCkkk.....",
	".....CCCCCCCCCCCCCCCCCCCkkk.....",
	"....rCCrCCrCCrCCrCCrCCrCkkkC....",
	"...CCCCCCCCCCCCCCCCCCCCCkkkCC...",
	"..CCrCCrCCrCCrCCrCCrCCrCkkkCrC..",
	".CCCCCCCCCCCCCCCCCCCCCCCCCCCCCC.",
	"rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr",
	"...ssssssssssssssssssssssssss...",
	"...SSSSSSSSSSSSSSSSSSSSSSSSSS...",
	"...SSSkkWkkkSSSSSSSSkkWkkkSSS...",
	"...SSSbbWbbbSSSSSSSSbbWbbbSSS...",
	"...SSSWWWWWWSSTTTTSSWWWWWWSSS...",
	"...SSSbbWbbbSSTTTTSSbbWbbbSSS...",
	"...SSSbbWbbbSSTTTTSSbbWbbbSSS...",
	"...SSSssssssSSTTTTSSssssssSSS...",
	"...SSSSSSSSSSSTTTYSSSSSSSSSSS...",
	"...SSSSSSSSSSSTTTTSSSSSSSSSSS...",
	"...SSSSSSSSSSSTTTTSSSSSSSSSSS...",
	"...SSSSSSSSSSSTTTTSSSSSSSSSSS...",
	"...SSSSSSSSSSSTTTTSSSSSSSSSSS...",
}, {
	"kkkkkkkkkkkkkkkkkkkkkkks",
	"kccccccccccccccccccccccs",
	"kSSSSSSSSSSSSSSSSSSSSSSs",
	"kSWbbSSSWbbSSSWbbSSSWbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"ksssssssssssssssssssssss",
	"kSSSSSSSSSSSSSSSSSSSSSSs",
	"kSWbbSSSWbbSSSWbbSSSWbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"ksssssssssssssssssssssss",
	"kSSSSSSSSSSSSSSSSSSSSSSs",
	"kSWbbSSSWbbSSSWbbSSSWbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"ksssssssssssssssssssssss",
	"kSSSSSSSSSSSSSSSSSSSSSSs",
	"kSWbbSSSWbbSSSWbbSSSWbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"ksssssssssssssssssssssss",
	"kSSSSSSSSSSSSSSSSSSSSSSs",
	"kSWbbSSSWbbSSSWbbSSSWbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"kSbbbSSSbbbSSSbbbSSSbbbs",
	"ksssssssssTTTTssssssssss",
	"kSSSSSSSSSTTTTSSSSSSSSSs",
	"kSSSSSSSSSTTTTSSSSSSSSSs",
	"kkkkkkkkkkTTTTkkkkkkkkks",
}}

// The board that marks where the front of the train stops.
var stopMarkArt = []string{
	"KKKKKKKKKKKKKK",
	"KYYYYYYYYYYYYK",
	"KYYYYYYYYYYYYK",
	"KYYKKKKKKKKYYK",
	"KYYKKKKKKKKYYK",
	"KYYYYYKKYYYYYK",
	"KYYYYYKKYYYYYK",
	"KYYYYYKKYYYYYK",
	"KYYYYYKKYYYYYK",
	"KYYYYYYYYYYYYK",
	"KYYYYYYYYYYYYK",
	"KKKKKKKKKKKKKK",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
}

// The board that asks for the horn before a crossing.
var whistleArt = []string{
	"KKKKKKKKKKKKKK",
	"KWWWWWWWWWWWWK",
	"KWWWWWWWWWWWWK",
	"KWWWWWWWWWWWWK",
	"KWWKWWWWWWKWWK",
	"KWWKWWWWWWKWWK",
	"KWWKWWKKWWKWWK",
	"KWWKWKWWKWKWWK",
	"KWWWKWWWWKWWWK",
	"KWWWKWWWWKWWWK",
	"KWWWWWWWWWWWWK",
	"KWWWWWWWWWWWWK",
	"KWWWWWWWWWWWWK",
	"KKKKKKKKKKKKKK",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
	"......gk......",
}

// A crossing's signal: its lamps flash in turn.
var crossingArt = [2][]string{{
	"......kk......",
	".W....kk....W.",
	".WW...kk...WW.",
	"..WW..kk..WW..",
	"...WR.kk.RW...",
	"....RRkkRR....",
	".....RkkR.....",
	"......kk......",
	".....WkkW.....",
	"....WWkkWW....",
	"...RW.kk.WR...",
	"..RR..kk..RR..",
	".RR...kk...RR.",
	".R....kk....R.",
	"......kk......",
	".KKKK.kk.KKKK.",
	".KRRK.kk.KrrK.",
	".KRRK.kk.KrrK.",
	".KKKK.kk.KKKK.",
	"......kk......",
	"......kk......",
	".KYYKYYKYYKYY.",
	".YKYYKYYKYYKY.",
	"......kk......",
	"......kk......",
	"......kk......",
}, {
	"......kk......",
	".W....kk....W.",
	".WW...kk...WW.",
	"..WW..kk..WW..",
	"...WR.kk.RW...",
	"....RRkkRR....",
	".....RkkR.....",
	"......kk......",
	".....WkkW.....",
	"....WWkkWW....",
	"...RW.kk.WR...",
	"..RR..kk..RR..",
	".RR...kk...RR.",
	".R....kk....R.",
	"......kk......",
	".KKKK.kk.KKKK.",
	".KrrK.kk.KRRK.",
	".KrrK.kk.KRRK.",
	".KKKK.kk.KKKK.",
	"......kk......",
	"......kk......",
	".KYYKYYKYYKYY.",
	".YKYYKYYKYYKY.",
	"......kk......",
	"......kk......",
	"......kk......",
}}

var bufferArt = []string{
	"RRRWWWRRRWWWRRRWWWRR",
	"RRWWWRRRWWWRRRWWWRRR",
	"RWWWRRRWWWRRRWWWRRRW",
	"WWWRRRWWWRRRWWWRRRWW",
	"WWRRRWWWRRRWWWRRRWWW",
	"KKKKKKKKKKKKKKKKKKKK",
	".kk......kk.......kk",
	".kk......kk.......kk",
	".kk......kk.......kk",
	"kkkkkkkkkkkkkkkkkkkk",
	".kk......kk.......kk",
	".kk......kk.......kk",
}

// The board with the station's name, before the platform.
var nameBoardArt = []string{
	"kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kWWWWKWWWKWKWWWKWKWWWKWKWWWKWKWWWWWk",
	"kWWWWKWWWKWKWWWKWKWWWKWKWWWKWKWWWWWk",
	"kWWWWKWWWKWKWWWKWKWWWKWKWWWKWKWWWWWk",
	"kWWWWKWWWKWKWWWKWKWWWKWKWWWKWKWWWWWk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBk",
	"kBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWk",
	"kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk",
	"......kk....................kk......",
	"......kk....................kk......",
	"......kk....................kk......",
	"......kk....................kk......",
}

const (
	hillWidth  = 320
	hillHeight = 24
)

var (
	imageTree, imagePine, imageStopMark   *ebiten.Image
	imageWhistle, imageBuffer, imageBoard *ebiten.Image
	imageHouse                            [2]*ebiten.Image
	imageCrossing                         [2]*ebiten.Image
	imageHills                            *ebiten.Image
	imageLimit                            [100]*ebiten.Image // speed limit signs by km/h
	imageDistance                         [4]*ebiten.Image   // boards 100, 200 and 300 m before a stop
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

var signText pixtext.Line

// limitSign is a round sign, red ring on white, with the limit in km/h, on
// a post.
func limitSign(kmh int) *ebiten.Image {
	img := ebiten.NewImage(22, 32)
	for y := range 22 {
		for x := range 22 {
			switch d := (2*x-21)*(2*x-21) + (2*y-21)*(2*y-21); {
			case d <= 15*15:
				img.Set(x, y, palette['W'])
			case d <= 21*21:
				img.Set(x, y, palette['R'])
			}
		}
	}
	signText.Reset().Num(kmh).Center(img, 11, 8, 1, palette['K'])
	for y := 22; y < 32; y++ {
		img.Set(10, y, palette['g'])
		img.Set(11, y, palette['k'])
	}
	return img
}

// distanceBoard counts down to a stop mark in hundreds of meters.
func distanceBoard(n int) *ebiten.Image {
	img := ebiten.NewImage(14, 26)
	for y := range 14 {
		for x := range 14 {
			c := palette['W']
			if x < 2 || y < 2 || x > 11 || y > 11 {
				c = palette['B']
			}
			img.Set(x, y, c)
		}
	}
	signText.Reset().Num(n).Center(img, 7, 4, 1, palette['K'])
	for y := 14; y < 26; y++ {
		img.Set(6, y, palette['g'])
		img.Set(7, y, palette['k'])
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

// hills is a strip of distant hills that repeats every hillWidth pixels:
// far blue ridges, and green ones nearer with a lighter crest.
func hills() *ebiten.Image {
	img := ebiten.NewImage(hillWidth, hillHeight)
	// TinyGo runs a loop at compile time only up to 1000 iterations, so the
	// strip is drawn in parts.
	for x := 0; x < hillWidth; x += 20 {
		hillColumns(img, x, x+20)
	}
	return img
}

func hillColumns(img *ebiten.Image, from, to int) {
	far, farTop := color.RGBA{120, 150, 170, 255}, color.RGBA{150, 176, 194, 255}
	near, nearTop := color.RGBA{96, 130, 120, 255}, color.RGBA{120, 154, 130, 255}
	for x := from; x < to; x++ {
		t := x * 360 / hillWidth
		// Heights in thousandths of a pixel.
		h1 := 14000 + 6*isin(t*2+29) + 3*isin(t*5)
		h2 := 6000 + 4*isin(t*3+115) + 2*isin(t*7+57)
		for y := range hillHeight {
			switch up := (hillHeight - y) * 1000; {
			case up <= h2-1000:
				img.Set(x, y, near)
			case up <= h2:
				img.Set(x, y, nearTop)
			case up <= h1-1000:
				img.Set(x, y, far)
			case up <= h1:
				img.Set(x, y, farTop)
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
	}
	imageHills = hills()
	for _, l := range limits {
		if imageLimit[l.kmh] == nil {
			imageLimit[l.kmh] = limitSign(l.kmh)
		}
	}
	for n := 1; n <= 3; n++ {
		imageDistance[n] = distanceBoard(n)
	}
}
