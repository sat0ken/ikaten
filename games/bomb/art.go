package bomb

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. In the
// gophers, 'B', 'D', 'L' and 'O' are the body, its shade, its light and its
// outline, which differ for each gopher; the rest is the same for all. Each
// view has a standing frame and two steps. The side view faces right and is
// flipped for left.
var gopherArt = [3][3][]string{
	down: {
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".OhhO.OOOOOO.OhhO.",
			".ODhLOLBBBBBOBhDO.",
			"..OLWWWBBBBWWWBO..",
			"...WWWWWBBWWWWW...",
			"..OWWkWKBBWWkWKO..",
			"..OWWKKKBBWWKKKO..",
			"..OWWKKkBBWWKKkO..",
			".OBBWWWBBBBWWWBBO.",
			".OBBBBBNNNNBBBBBO.",
			"hhhBBBhhNNhhBBBhhh",
			"hhhBBBhhhhhhBBBhhh",
			"hhhBBBBBwwBBBBDhhh",
			"..OBBBBBwwBBBDDO..",
			"...OBBBBBBBBDDO...",
			"...OBBBBBBBDDDO...",
			"....OBBBBBDDDO....",
			"...hhhhhDDhhhhh...",
			"...hhhhhOOhhhhh...",
		},
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".OhhO.OOOOOO.OhhO.",
			".ODhLOLBBBBBOBhDO.",
			"..OLWWWBBBBWWWBO..",
			"...WWWWWBBWWWWW...",
			"..OWWkWKBBWWkWKO..",
			"..OWWKKKBBWWKKKO..",
			"..OWWKKkBBWWKKkO..",
			".OBBWWWBBBBWWWBBO.",
			"hhhBBBBNNNNBBBBBO.",
			"hhhBBBhhNNhhBBBBO.",
			"hhhBBBhhhhhhBBBhhh",
			"..OBBBBBwwBBBBDhhh",
			"..OBBBBBwwBBBDDhhh",
			"...OBBBBBBBBDDO...",
			"...OBBBBBBBDDDO...",
			"....hhhBBBDDDO....",
			"...hhhhhDDDhhh....",
			"....hhh.OOhhhhh...",
		},
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".OhhO.OOOOOO.OhhO.",
			".ODhLOLBBBBBOBhDO.",
			"..OLWWWBBBBWWWBO..",
			"...WWWWWBBWWWWW...",
			"..OWWkWKBBWWkWKO..",
			"..OWWKKKBBWWKKKO..",
			"..OWWKKkBBWWKKkO..",
			".OBBWWWBBBBWWWBBO.",
			".OBBBBBNNNNBBBBhhh",
			".OBBBBhhNNhhBBBhhh",
			"hhhBBBhhhhhhBBBhhh",
			"hhhBBBBBwwBBBBDO..",
			"hhhBBBBBwwBBBDDO..",
			"...OBBBBBBBBDDO...",
			"...OBBBBBBBDDDO...",
			"....OBBBBBDhhh....",
			"....hhhBDDhhhhh...",
			"...hhhhhOO.hhh....",
		},
	},
	up: {
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".ODDO.OOOOOO.ODDO.",
			".ODDLOLBBBBBOBDDO.",
			"..OLLLBBBBBBBBBO..",
			"...OLBBBBBBBBBO...",
			"..OLLBBBBBBBBBBO..",
			"..OLBBBBBBBBBBBO..",
			"..OBBBBBBBBBBBBO..",
			".OBBBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBBBO.",
			"hBBBBBBBBBBBBBBBBh",
			"hhBBBBBBBBBBBBBDhh",
			"hhBBBBBBBBBBBBDDhh",
			"..OBBBBBBBBBBDDO..",
			"...OBBBBBBBBDDO...",
			"...OBBBBDDBDDDO...",
			"....OBBBDDDDDO....",
			"...hhhhhDDhhhhh...",
			"...hhhhhOOhhhhh...",
		},
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".ODDO.OOOOOO.ODDO.",
			".ODDLOLBBBBBOBDDO.",
			"..OLLLBBBBBBBBBO..",
			"...OLBBBBBBBBBO...",
			"..OLLBBBBBBBBBBO..",
			"..OLBBBBBBBBBBBO..",
			"..OBBBBBBBBBBBBO..",
			".OBBBBBBBBBBBBBBO.",
			"hBBBBBBBBBBBBBBBO.",
			"hBBBBBBBBBBBBBBBO.",
			"hhBBBBBBBBBBBBBDhh",
			"..OBBBBBBBBBBBDDhh",
			"..OBBBBBBBBBBDDOhh",
			"...OBBBBBBBBDDO...",
			"...OBBBBDDBDDDO...",
			"....hhhBDDDDDO....",
			"...hhhhhDDDhhh....",
			"....hhh.OOhhhhh...",
		},
		{
			"..................",
			"..................",
			"..................",
			"..OOO........OOO..",
			".ODDO.OOOOOO.ODDO.",
			".ODDLOLBBBBBOBDDO.",
			"..OLLLBBBBBBBBBO..",
			"...OLBBBBBBBBBO...",
			"..OLLBBBBBBBBBBO..",
			"..OLBBBBBBBBBBBO..",
			"..OBBBBBBBBBBBBO..",
			".OBBBBBBBBBBBBBBO.",
			".OBBBBBBBBBBBBBBBh",
			".OBBBBBBBBBBBBBBBh",
			"hhBBBBBBBBBBBBBDhh",
			"hhBBBBBBBBBBBBDO..",
			"hhOBBBBBBBBBBDDO..",
			"...OBBBBBBBBDDO...",
			"...OBBBBDDBDDDO...",
			"....OBBBDDDhhh....",
			"....hhhBDDhhhhh...",
			"...hhhhhOO.hhh....",
		},
	},
	left: {
		{
			"..................",
			"..................",
			".....OO...........",
			"....ODDO..........",
			"....OLLLOOOO......",
			"....OLLBBBBBO.....",
			"...OLLBBBBBWWW....",
			"...OLBBBBBWWWWW...",
			"..OLBBBBBBWWWWK...",
			"..OLBBBBBBWWWKK...",
			"..OBBBBBBBWWWKKO..",
			".OBBBBBBBBBWWBBON.",
			".OBBBBBBBBBBBhhNNN",
			".OBBBBBBBhhhBhhhhh",
			"..OBBBBBBhhhBhhhh.",
			"..OBBBBBBhhhBBOww.",
			"..OBBBBBBhhhDDOww.",
			"...OBBBBBBBDDDO...",
			"...OBBBBBBDDDO....",
			"....OBBBBDDDO.....",
			"....HHHHDDhhhhh...",
			"...HHHHHHhhhhhh...",
		},
		{
			"..................",
			"..................",
			".....OO...........",
			"....ODDO..........",
			"....OLLLOOOO......",
			"....OLLBBBBBO.....",
			"...OLLBBBBBWWW....",
			"...OLBBBBBWWWWW...",
			"..OLBBBBBBWWWWK...",
			"..OLBBBBBBWWWKK...",
			"..OBBBBBBBWWWKKO..",
			".OBBBBBBBBBWWBBON.",
			".OBBBBBBBBBBBhhNNN",
			".OBBBBBBBBBhhhhhhh",
			"..OBBBBBBBhhhhhhh.",
			"..OBBBBBBBhhhhOww.",
			"..OBBBBBBBBhhDOww.",
			"...OBBBBBBBDDDO...",
			"...OBBBBBBDDDO....",
			"....OBBBBDDDO.....",
			"..HHHHHODDOhhhhh..",
			"..HHHHH.OO.hhhhh..",
		},
		{
			"..................",
			"..................",
			".....OO...........",
			"....ODDO..........",
			"....OLLLOOOO......",
			"....OLLBBBBBO.....",
			"...OLLBBBBBWWW....",
			"...OLBBBBBWWWWW...",
			"..OLBBBBBBWWWWK...",
			"..OLBBBBBBWWWKK...",
			"..OBBBBBBBWWWKKO..",
			".OBBBBBBBBBWWBBON.",
			".OBBBBBBBBBBBhhNNN",
			".OBBBBBBhhBBBhhhhh",
			"..OBBBBhhhhBBhhhh.",
			"..OBBBBhhhhBBBOww.",
			"..OBBBBBhhBBDDOww.",
			"...OBBBBBBBDDDO...",
			"...OBBBBBBDDDO....",
			"....OBBBBDDDO.....",
			".....HHHhhhhhh....",
			".....HHHhhhhh.....",
		},
	},
}

var faintArt = []string{
	"..................",
	"..................",
	"..................",
	"..OOO........OOO..",
	".OhhO.OOOOOO.OhhO.",
	".ODhLOLBBBBBOBhDO.",
	"..OLLLBBBBBBBBBO..",
	"...KLBBKBKBBBKO...",
	"..OLKBKBBBKBKBBO..",
	"..OLBKBBBBBKBBBO..",
	"..OBKBKBBBKBKBBO..",
	".OBKBBBKBKBBBKBBO.",
	".OBBBBBNNNNBBBBBO.",
	"hhhBBBhhNNhhBBBhhh",
	"hhhBBBhhhhhhBBBhhh",
	"hhhBBBBBwwBBBBDhhh",
	"..OBBBBBwwBBBDDO..",
	"...OBBBBBBBBDDO...",
	"...OBBBBBBBDDDO...",
	"....OBBBBBDDDO....",
	"...hhhhhDDhhhhh...",
	"...hhhhhOOhhhhh...",
}

var wallArt = []string{
	"LLLLLLLLLLLLLLLLLg",
	"LLLLLLLLLLLLLLLLgk",
	"LLGGGGGGGGGGGGGGgk",
	"LLGGGGGGGGgGGGGGgk",
	"LLGGGgLGGGGGGGGGgk",
	"LLGGGGGGGGGGGGGGgk",
	"LLGGGGGGGGGgLGGGgk",
	"LLGGGGGGGgGGGGGGgk",
	"LLGGGGGGGGgGGGgGgk",
	"LLGGGGGGGGgGGGGGgk",
	"LLGGGGGGGGGgGGGGgk",
	"LLGGGGGgLGGGGGGGgk",
	"LLGGGGGGGGGGGgGGgk",
	"LLGGgGGGGGGGGGGGgk",
	"LLGGGGGGGGGGGGGGgk",
	"LLGGGGGGGGGGGGGGgk",
	"LLgggggggggggggggk",
	"Lkkkkkkkkkkkkkkkkk",
}

var blockArt = []string{
	"RRRRRRRmRRRRRRRmRR",
	"RrrrrrrmRrrrrrrmRr",
	"RrrqrrrmRrrrrrrmRr",
	"RrrrrrrmRrrrrrrqRr",
	"qqqqqqqmqqqqqqqmqq",
	"mmmmmmmmmmmmmmmmmm",
	"RRRmRRRRRRRmRRRRRR",
	"rrrmRrrrrrrmRrrrrr",
	"rrrmRrrrrrrmqrrrrr",
	"rrrmRrrrrrrmRrrrrr",
	"qqqmqqqqqqqmqqqqqq",
	"mmmmmmmmmmmmmmmmmm",
	"RRRRRRRmRRRRRRRmRR",
	"RrrrrrrmRrrrrrrmRr",
	"RrrrrrqmRrrrrrrmRr",
	"RrrrrrrmRqrrrrrmRr",
	"qqqqqqqmqqqqqqqmqq",
	"mmmmmmmmmmmmmmmmmm",
}

var crumbleArt = [2][]string{
	{
		"..R.RR.m.R..RR..RR",
		"Rrr.r....r.rrrr...",
		"Rr.qr.rm.rrrr.r.Rr",
		".r.rrrr.Rrrrrrrq.r",
		"qqq.qq.m...q..qm.q",
		"mmmm.mmmm....mm..m",
		"RRRmRRR.RRRmRR.R..",
		".......r.rr...r.rr",
		"rr....r..rr.q.rrrr",
		".r.mRr..rrrmRr.rr.",
		"...mqqqqqq....qqqq",
		"mm.mmmmm.m.mmmmmm.",
		"..RR.RRmRR..RRRmRR",
		"R....r.m.rrrrrrmRr",
		"R.r..r.mRr.rrr.m..",
		"RrrrrrrmRqrrrrrmR.",
		"qqq..q...qqq.qq.qq",
		".mmmmm.mm...m.mm..",
	},
	{
		".R.R..R.R....R....",
		".r...r.m.r.r.rr...",
		"R..q.rr..r.r.rrmR.",
		".....r.m.......qR.",
		"..qq.q..q.........",
		".....m.mm.....m.m.",
		"R.R.RRR..R..R.....",
		"...mR.......Rrr...",
		".rr.R...rr.mqrrr..",
		".r.mR.rrr.r..rr..r",
		"qqqmqq....q..q...q",
		"..m..m..mmm.m.m.mm",
		".RRRRR.mR..RRR.mRR",
		"..rrr.rmR.......R.",
		"....r....r...rr..r",
		"R..rrr....r.r.r.R.",
		"qqqq..q..q..q.q.q.",
		"..mmm....m.mm.m...",
	},
}

var bombArt = [2][]string{
	{
		"..............my..",
		".............mo.o.",
		"..........Ggg.....",
		".......bbbGgg.....",
		".....bbbbbbb......",
		"....bxxbbbbbbB....",
		"...bXXXxbbbbBB....",
		"..bbXXxxbbbBBBB...",
		"..bbxxxxbbbbbbB...",
		"..bbbxxbbbbbbbbB..",
		"..bbbbbbbbbbbbbb..",
		"..bbbbbbbbbbbbbb..",
		"..bbbbbbbbbbbbb...",
		"..bbbbbbbbbbbbb...",
		"...bBbbbbbbbbb....",
		"....BBbbbbbbb.....",
		".....Bbbbbbb......",
		"..................",
	},
	{
		"..............oOy.",
		".............m.y.o",
		"..........Ggg.....",
		".......bbbGgg.....",
		".....bbbbbbb......",
		"....bxxbbbbbbB....",
		"...bXXXxbbbbBB....",
		"..bbXXxxbbbBBBB...",
		"..bbxxxxbbbbbbB...",
		"..bbbxxbbbbbbbbB..",
		"..bbbbbbbbbbbbbb..",
		"..bbbbbbbbbbbbbb..",
		"..bbbbbbbbbbbbb...",
		"..bbbbbbbbbbbbb...",
		"...bBbbbbbbbbb....",
		"....BBbbbbbbb.....",
		".....Bbbbbbb......",
		"..................",
	},
}

var itemArt = [4][]string{
	fireUp: {
		"PPPPPPPPPPPPPPPPPP",
		"Ppppppppppppppppps",
		"Ppppppppppppppppps",
		"PpppppppOOppppppps",
		"PpppppOOOOOpppppps",
		"PppppppOOOOpppppps",
		"PpppppOOOOOOOpppps",
		"PpppppOOooOOpOppps",
		"PppppOOooooOOpppps",
		"PpppOOOooooOOOppps",
		"PpppOOooooooOOppps",
		"PpppOOooyyooOOppps",
		"PpppOOoyyyyoOOppps",
		"PpppOOoyyyyoOOppps",
		"PppppOoyyyyoOpppps",
		"PpppppOoyyoOppppps",
		"PppppppOOOOpppppps",
		"Psssssssssssssssss",
	},
	bombUp: {
		"PPPPPPPPPPPPPPPPPP",
		"Ppppppppppppppppps",
		"Pppppppppppppoopps",
		"Pppppppppppppyppps",
		"Ppppppppppppmpppps",
		"Ppppppppppggppppps",
		"Ppppppbbbbggppppps",
		"Pppppbbbbbbpppppps",
		"Ppppbxxxbbbbppppps",
		"Pppbbxxxbbbbbpppps",
		"Pppbbxxbbbbbbpppps",
		"Pppbbbbbbbbbbpppps",
		"Pppbbbbbbbbbbpppps",
		"Ppppbbbbbbbbppppps",
		"Pppppbbbbbbpppppps",
		"Ppppppbbbbppppppps",
		"Ppppppppppppppppps",
		"Psssssssssssssssss",
	},
	speedUp: {
		"PPPPPPPPPPPPPPPPPP",
		"Pppppppppppccccpps",
		"PpppppppppcccCppps",
		"PppppppppcccCpppps",
		"PpppppppcccCppppps",
		"PppppppcccCpppppps",
		"PpppppcccCppppppps",
		"Pppppcccccccccppps",
		"PppppCCCCcccceppps",
		"Pppppppppcccepppps",
		"Ppppppppccceppppps",
		"Pppppppcccepppppps",
		"Ppppppccceppppppps",
		"Pppppcceppppppppps",
		"Ppppccepppppppppps",
		"Ppppceppppppppppps",
		"Ppppppppppppppppps",
		"Psssssssssssssssss",
	},
}

var grassArt = [2][]string{
	{
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaataaaaaaa",
		"aaaaaaaaatauaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaataaaaaa",
		"aaaaaaaaaatauaaaaa",
		"aaatataaaaaataaaaa",
		"aatauauaaaatauaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
		"aataaaataaaaaaaaaa",
		"atauaatauaaaaaaaaa",
		"aaaaaaaaaaaaaaaaaa",
	},
	{
		"AAAAAAAAAAAAAAAAAA",
		"AAAtAAAAAAAAAAAAAA",
		"AAtAuAAAAAAAAAAAAA",
		"AAAtAuAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
		"AAAAAtAAAAAAAAAAtA",
		"AAAAtAuAAAAAAAAtAu",
		"AAAAAAAAAAAAAAAAAA",
		"AAAAAAAAtAAAAAAAAA",
		"AAAAAAAttuAAAAAAAA",
		"AAAAAAAtAuAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
		"AAAtAAAAAAAAAAAAAA",
		"AAtAuAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
		"AAAAAAAAAAAAAAAAAA",
	},
}

// The gophers' colors: body, shade, light and outline. The person's is the
// tan gopher.
var gopherColors = [players][4]color.RGBA{
	{{214, 186, 150, 255}, {176, 146, 110, 255}, {236, 214, 184, 255}, {112, 86, 60, 255}},
	{{120, 200, 230, 255}, {80, 160, 196, 255}, {176, 228, 246, 255}, {36, 92, 128, 255}},
	{{246, 160, 190, 255}, {210, 120, 150, 255}, {255, 204, 222, 255}, {136, 60, 92, 255}},
	{{160, 214, 120, 255}, {116, 172, 84, 255}, {202, 238, 172, 255}, {62, 106, 44, 255}},
}

var gopherNames = [players]string{"YOU", "BLUE", "PINK", "LIME"}

// gopherPalette colors the parts all gophers share.
var gopherPalette = map[byte]color.RGBA{
	'K': {20, 20, 24, 255},
	'k': {90, 90, 96, 255},
	'W': {250, 250, 250, 255},
	'N': {60, 40, 34, 255},
	'w': {255, 255, 248, 255},
	'h': {240, 218, 190, 255},
	'H': {214, 190, 160, 255},
}

// palette colors the arena, the bombs and the items.
var palette = map[byte]color.RGBA{
	'G': {150, 154, 164, 255},
	'L': {206, 210, 218, 255},
	'g': {110, 114, 124, 255},
	'k': {66, 68, 78, 255},
	'R': {222, 132, 84, 255},
	'r': {194, 98, 56, 255},
	'q': {160, 76, 44, 255},
	'm': {104, 56, 40, 255},
	'b': {32, 32, 40, 255},
	'x': {86, 90, 104, 255},
	'X': {170, 176, 190, 255},
	'B': {16, 16, 22, 255},
	'y': {255, 236, 90, 255},
	'o': {255, 150, 40, 255},
	'O': {232, 64, 32, 255},
	'p': {250, 244, 220, 255},
	'P': {255, 255, 250, 255},
	's': {196, 186, 160, 255},
	'c': {90, 190, 250, 255},
	'C': {200, 240, 255, 255},
	'e': {40, 110, 200, 255},
	'a': {76, 156, 84, 255},
	'A': {68, 146, 76, 255},
	't': {100, 180, 100, 255},
	'u': {56, 126, 64, 255},
}

// The head shown in the score bar is the top of the gopher facing down.
const headH = 13

var (
	imageGopher           [players][3][3]*ebiten.Image
	imageFaint            [players]*ebiten.Image
	imageHead             [players]*ebiten.Image
	imageWall, imageBlock *ebiten.Image
	imageCrumble          [2]*ebiten.Image
	imageBomb             [2]*ebiten.Image
	imageItem             [4]*ebiten.Image
	imageGrass            [2]*ebiten.Image
	imageFire             [3][7]*ebiten.Image // two flickers and dying down
)

// The pieces of a blast: the middle where it crosses or starts, the arms
// along a row or a column, and the ends, named for the side they end on.
const (
	fireCenter = iota
	fireAcross
	fireAlong
	fireEndL
	fireEndR
	fireEndU
	fireEndD
)

// The fire's colors from the outside in.
var fireColors = [4]color.RGBA{
	{226, 62, 28, 255},
	{255, 140, 36, 255},
	{255, 226, 90, 255},
	{255, 250, 222, 255},
}

// fireReach is how far from the middle line each color reaches, in half
// pixels, burning and dying down.
var fireReach = [2][4]int{{15, 12, 9, 5}, {7, 5, 3, 0}}

// fireWave ripples the edges, in half pixels.
var fireWave = [8]int{0, 1, 2, 2, 1, 0, 1, 1}

func iabs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// fireImage draws a piece of fire. Coordinates are in half pixels from the
// cell's middle, so that the middle falls between pixels.
func fireImage(piece, heat int) *ebiten.Image {
	img := ebiten.NewImage(tile, tile)
	reach := &fireReach[heat/2]
	for y := range tile {
		for x := range tile {
			hx, hy := 2*x+1-tile, 2*y+1-tile
			across := iabs(hy) + fireWave[(x+heat*3)%8]
			along := iabs(hx) + fireWave[(y+heat*3+4)%8]
			var d int
			switch piece {
			case fireCenter:
				d = min(across, along, (iabs(hx)+iabs(hy))*2/3)
			case fireAcross:
				d = across
			case fireAlong:
				d = along
			case fireEndL:
				d = across + max(0, -hx-4)*3/4
			case fireEndR:
				d = across + max(0, hx-4)*3/4
			case fireEndU:
				d = along + max(0, -hy-4)*3/4
			case fireEndD:
				d = along + max(0, hy-4)*3/4
			}
			for i := len(fireColors) - 1; i >= 0; i-- {
				if d < reach[i] {
					img.Set(x, y, fireColors[i])
					break
				}
			}
		}
	}
	return img
}

// gopherSprite draws pixel art in the colors of gopher i.
func gopherSprite(rows []string, i int) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	for y, row := range rows {
		for x := range len(row) {
			switch ch := row[x]; ch {
			case 'B':
				img.Set(x, y, gopherColors[i][0])
			case 'D':
				img.Set(x, y, gopherColors[i][1])
			case 'L':
				img.Set(x, y, gopherColors[i][2])
			case 'O':
				img.Set(x, y, gopherColors[i][3])
			default:
				if c, ok := gopherPalette[ch]; ok {
					img.Set(x, y, c)
				}
			}
		}
	}
	return img
}

func sprite(rows []string) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	solidArt(img, rows)
	for y, row := range rows {
		for x := range len(row) {
			if c, ok := palette[row[x]]; ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

func init() {
	for i := range players {
		for v := range 3 {
			for f := range 3 {
				imageGopher[i][v][f] = gopherSprite(gopherArt[v][f], i)
			}
		}
		imageFaint[i] = gopherSprite(faintArt, i)
		// SubImage allocates, so the heads are cut once here.
		imageHead[i] = imageGopher[i][down][0].SubImage(image.Rect(0, 0, tile, headH)).(*ebiten.Image)
	}
	imageWall = sprite(wallArt)
	imageBlock = sprite(blockArt)
	for f := range 2 {
		imageCrumble[f] = sprite(crumbleArt[f])
		imageBomb[f] = sprite(bombArt[f])
		imageGrass[f] = sprite(grassArt[f])
	}
	for k := fireUp; k <= speedUp; k++ {
		imageItem[k] = sprite(itemArt[k])
	}
	for heat := range imageFire {
		for piece := range imageFire[heat] {
			imageFire[heat][piece] = fireImage(piece, heat)
		}
	}
}

// solidArt fills an image whose art has no transparent pixel before it is
// drawn: filling marks the image opaque, and an opaque image is copied
// faster, without looking at its mask.
func solidArt(img *ebiten.Image, rows []string) {
	for _, row := range rows {
		for i := range len(row) {
			if row[i] == '.' {
				return
			}
		}
	}
	img.Fill(color.RGBA{0, 0, 0, 255})
}
