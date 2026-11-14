package fish

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Fish face
// right.
var palette = map[byte]color.RGBA{
	'B': {30, 70, 110, 255},
	'E': {30, 100, 50, 255},
	'F': {120, 150, 130, 255},
	'G': {80, 120, 90, 255},
	'H': {180, 64, 48, 255},
	'K': {16, 18, 22, 255},
	'M': {70, 52, 30, 255},
	'N': {100, 76, 48, 255},
	'O': {40, 60, 70, 255},
	'P': {236, 200, 210, 255},
	'Q': {220, 160, 170, 255},
	'R': {240, 120, 120, 255},
	'U': {36, 66, 130, 255},
	'W': {245, 248, 250, 255},
	'X': {130, 30, 40, 255},
	'Z': {50, 90, 130, 255},
	'b': {60, 110, 160, 255},
	'c': {120, 190, 240, 255},
	'e': {60, 150, 80, 255},
	'f': {240, 190, 150, 255},
	'g': {110, 150, 120, 255},
	'h': {236, 236, 228, 255},
	'i': {40, 60, 90, 255},
	'j': {130, 40, 34, 255},
	'k': {80, 90, 100, 255},
	'l': {150, 146, 140, 255},
	'm': {84, 80, 76, 255},
	'n': {150, 120, 80, 255},
	'o': {255, 150, 40, 255},
	'p': {250, 226, 230, 255},
	'q': {200, 80, 90, 255},
	'r': {220, 70, 80, 255},
	's': {170, 182, 196, 255},
	'u': {50, 90, 170, 255},
	'w': {200, 210, 220, 255},
	'y': {240, 210, 60, 255},
}

var fishArt = [kinds][]string{
	aji: {
		"F.....................",
		"FF.....OOOOOOOOOOOO...",
		"FFFF..OggggGGggggggO..",
		"..FFFOggggggggggggggO.",
		"...FsssssssssyyysWKssO",
		"...FsssskkkkkkkkkkksOO",
		"..FFFOWWWWWWWWWWWWWWO.",
		"FFFF..OWWWWWWWWWWWWO..",
		"FF.....OOOOOOOOOOOO...",
		"F.....................",
	},
	saba: {
		"Z...........O..O..........",
		"ZZZ......OOOBOOBOOOOOO....",
		"ZZZZZ..OOBBKBBBBKBBBBKOO..",
		"..ZZZZOBBKBBBBKBBBBKBBBBO.",
		"....ZssssssssssssssssWKssO",
		"....ZsssbbbbbbbbbbbbbbssOO",
		"..ZZZZOWWWWWWWWWWWWWWWWWO.",
		"ZZZZZ..OOWWWWWWWWWWWWWOO..",
		"ZZZ......OOOOOOWOOOOOO....",
		"Z..............O..........",
	},
	madai: {
		"q..........XXXXX..........",
		"qq......XXXrrrrrXXXX......",
		"qqq.....XrrrrrrrrrrrX.....",
		"qqqq..XXrrrrrcrrrrrrrXX...",
		".qqqqXrrrrcrrrrrrrcrrrrX..",
		"..qqqrrrrrrrrrrrcrrrrrrX..",
		"...qRRRRRRRRRRRRRRRWKRRRX.",
		"....XRRRRRRRcRRRRRRRRRRRX.",
		"....XRRRRcRRRRRcRRRRRRRRX.",
		"...qRRRRRRRRRRRRRRRRRRRXX.",
		"..qqqppppppppppppppppppX..",
		".qqqqXpppppppppppppppppX..",
		"qqqq..XXpppppppppppppXX...",
		"qqq.....XXpppppppppXX.....",
		"qq........XXXXXXXXX.......",
		"q.........................",
	},
	hirame: {
		"..............................",
		"......MMMMMMMMMMMMMMMMMMMM....",
		"...MMMnnnnnnnnnnnnnnnnnnnnMM..",
		"MMM.MnnnnnnnnnnnNNnnnnnnWKnM..",
		".MnMnnnnNNnnnnnnnnnnnnnNNnKnM.",
		".MnnnnnnnnnnnnNNnnnnnnnnnnnnnM",
		".MnMnnnnnnnnNNnnnnnnNNnnnnnnM.",
		"MMM.MnnnnnNNnnnnnnnnnnnnnnnM..",
		"...MMMnnnnnnnnnnnnnnnnnnnnMM..",
		"......MMMMMMMMMMMMMMMMMMMM....",
	},
	ika: {
		"......................M.....",
		"......................MM....",
		"..............MMMM.....MMM..",
		"MMMMMMMM.MMMMMQQQQMMMMM..MM.",
		"..MPPpppMQQQQQQQQQQQQQQMM.M.",
		".MPPPppWKpppppppppppppppM...",
		"..MPPpppKpppppppppppppppM...",
		"MMPPPpppMppppppppppppppMM.M.",
		"..MPPPpM.MMMMMppppMMMMM..MM.",
		"..MMMMM.......MMMM.....MMM..",
		"......................MM....",
		"......................M.....",
	},
	tachiuo: {
		"..........sssssssssssssssssssssssssssssss...",
		".........swwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwss.",
		"...sssssswwwwwwwwwwwwwwwwwwwwwwwwwwwwwkKwwws",
		"...ssssssWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWs",
		".........sssssWWWWWWWWWWWWWWWWWWWWWWWWWWWss.",
		"..............sssssssssssssssssssssssssss...",
	},
	buri: {
		"O................OO...............",
		"OOO.........OOOOOBBOOOOOOOOO......",
		"OyyOO....OOOBBBBBBBBBBBBBBBBOOO...",
		".OOyyOO.ObbbbbbbbbbbbbbbbbbbbbbO..",
		"...OOyyOybyybyybyybyybyybyybWKbbO.",
		".....OWWyyyyyyyyyyyyyyyyyyyyyyWWWO",
		".....OWWWWWWWWWWWWWWWWWWWWWWWWWWOO",
		"...OOyyOWWWWWWWWWWWWWWWWWWWWWWWWO.",
		".OOyyOO.OWWWWWWWWWWWWWWWWWWWWWWO..",
		"OyyOO....OOOWWWWWWWWWWWWWWWWOOO...",
		"OOO.........OOOOOOOOOOOOOOOO......",
		"O.................................",
	},
	boot: {
		"....KKKKKK....",
		"....KnnnnK....",
		"....KnnnnK....",
		"....KnhnhK....",
		"....KnnnnK....",
		"....KnhnhK....",
		"...eKnnnnK....",
		"..e.KnhnhK....",
		"..E.KnnnnK....",
		"KKKKnnnnnnKK..",
		"KnnnnnnnnnnK..",
		"KnnnnnnnnnnnK.",
		"KnnnnnnnnnnnK.",
		"KKKKKKKKKKKKK.",
	},
}

var boatArt = []string{
	"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
	"KhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhK",
	"KHiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiHK",
	"KKKKHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHKKKK",
	"....KHiHHHHHHHiHHHHHHHiHHHHHHHiHHHHHHHiHHHHHHHiK....",
	".....KHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHK.....",
	"......KjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjK......",
	".......KjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjK.......",
	"........KjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjK........",
	".........KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.........",
}

var anglerArt = []string{
	"................",
	"....KKKKKK......",
	"....KyyyyK......",
	"....KyyyyK......",
	"...KKyfffyKK....",
	".....KfffK......",
	".....KffKK......",
	".....KfffK......",
	".....KuuK.......",
	"....KuuuuK......",
	"...KuuuuuuKKKK..",
	"...KuuuuuuuKKKK.",
	"...KuuuuuUK.....",
	"...KuuuuuUK.....",
	"...KuuuuuUK.....",
	"...KuuoooUK.....",
	"....KuuuuK......",
	"....KUUKUUK.....",
	"....KUK.KUK.....",
	"....KKK.KKK.....",
}

var rockArt = []string{
	"..............",
	"......KK......",
	"...KKKlkKKK...",
	".KKlllkkkkkKK.",
	".KlllkkkkkkmK.",
	"KlllkkkkkkmmmK",
	"KllkkkkkkmmmmK",
	".KKKKKKKKKKKK.",
}

var lureArt = []string{
	"........",
	".yyyyKWK",
	"soooooW.",
	"s.......",
}

var bubbleArt = []string{
	"..WW..",
	".W..W.",
	"W.W..W",
	"W....W",
	".W..W.",
	"..WW..",
}

var weedArt = [2][]string{
	{
		"...ee...",
		"...eeE..",
		"....eE..",
		"...eee..",
		"..eE....",
		".eeE....",
		"...ee...",
		"....eEe.",
		"....eE..",
		"..eee...",
		"..eE....",
		"..eEe...",
		"...ee...",
		"...eeE..",
		"....eE..",
		"...eee..",
	},
	{
		"....ee..",
		"..eeE...",
		"..eE....",
		"..eee...",
		"...eE...",
		"...eeE..",
		"....ee..",
		"...eEe..",
		"..eE....",
		".eee....",
		"...eE...",
		"....eEe.",
		"....ee..",
		"..eeE...",
		"..eE....",
		"..eee...",
	},
}

// sprites holds a game's images. A Game makes them in NewGame and drops them
// with it, so that the launcher keeps only the running game's in RAM.
type sprites struct {
	fish         [kinds]*ebiten.Image
	boat, angler *ebiten.Image
	weed         [2]*ebiten.Image
	rock, lure   *ebiten.Image
	sun, bubble  *ebiten.Image
}

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

// sun is a white disc, tinted by the hour when drawn.
func sun() *ebiten.Image {
	const d = 18
	img := ebiten.NewImage(d, d)
	for y := range d {
		for x := range d {
			if dx, dy := 2*x+1-d, 2*y+1-d; dx*dx+dy*dy <= d*d-4 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	return img
}

func newSprites() *sprites {
	a := &sprites{}
	for k := range kinds {
		a.fish[k] = sprite(fishArt[k])
	}
	a.boat = sprite(boatArt)
	a.angler = sprite(anglerArt)
	for f := range a.weed {
		a.weed[f] = sprite(weedArt[f])
	}
	a.rock = sprite(rockArt)
	a.lure = sprite(lureArt)
	a.bubble = sprite(bubbleArt)
	a.sun = sun()
	return a
}
