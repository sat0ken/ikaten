package jump

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent.
// The gopher faces right and is flipped for left: standing, two steps,
// jumping with its arms up, and dashing leaning in.
var gopherArt = [5][]string{
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
	{
		"..............OO..",
		".OO..........OhhO.",
		".OhO.OO......OhhO.",
		"OhhhODDOOO...OhhO.",
		"OhhhOLLBBBOO.OhhO.",
		"OhhOOLLBBBBBOOOOO.",
		".OOOLLBBBBWWWWW...",
		"..OLLBBBBBWWWWK...",
		"..OLBBBBBBWWWKK...",
		"..OBBBBBBBWWWKK...",
		"..OBBBBBBBWWWWBO..",
		".OBBBBBBBBBBBBhNNN",
		".OBBBBBBBBBBBhhNNN",
		"..OBBBBBBBBBBhhhhh",
		"..OBBBBBBBBBBBhww.",
		"..OBBBBBBBBBBDOww.",
		"..OBBBBBBBBBDDO...",
		"...OBBBBBBDDDO....",
		"....OBBBBDDDO.....",
		".....HHHDhhh......",
		"....HHHHhhhhh.....",
		".....HHH.hhhh.....",
	},
	{
		"..................",
		"..................",
		".......OO.........",
		"......ODDO........",
		"......OLLLOOOO....",
		"......OLLBBBBBO...",
		".....OLLBBBBBWWW..",
		".....OLBBBBBWWWWW.",
		"....OLBBBBBBWWWWK.",
		"...OLBBBBBBWWWKK..",
		"...OBBBBBBBWWWKKO.",
		"..OBBBBBBBBBWWBBON",
		"..OBBBBBBBBBBBhhNN",
		"..OBBBBBBBBBhhhhhh",
		"...OBBBBBBBhhhhhhh",
		"...OBBBBBBBhhhhOww",
		"..OBBBBBBBBhhDOww.",
		"...OBBBBBBBDDDO...",
		"...OBBBBBBDDDO....",
		"....OBBBBDDDO.....",
		"..HHHHHODDOhhhhh..",
		"..HHHHH.OO.hhhhh..",
	},
}

// The bugs face right and are flipped for left.
var bugArt = [2][]string{
	{
		"................",
		"................",
		".....kkkkk......",
		"...kkPPPppkk....",
		"..kPPPPPppppk..k",
		".kpPPPPPppppkk.k",
		".kpppppppppkkWW.",
		".kppppppppkkkkKk",
		".kqqqqqqqqqkkkk.",
		"..kqqqqqqqqqkk..",
		".kkkkkkkkkkkkkk.",
		"..k...k...k.....",
	},
	{
		"................",
		"................",
		".....kkkkk......",
		"...kkPPPppkk....",
		"..kPPPPPppppk..k",
		".kpPPPPPppppkk.k",
		".kpppppppppkkWW.",
		".kppppppppkkkkKk",
		".kqqqqqqqqqkkkk.",
		"..kqqqqqqqqqkk..",
		".kkkkkkkkkkkkkk.",
		"...k...k...k....",
	},
}

var spikyArt = [2][]string{
	{
		"................",
		"..y..y..y..y....",
		".yYyyYyyYyyYy...",
		"...kkoooOOkk....",
		"..koooooOOOOk..k",
		".kOoooooOOOOkk.k",
		".kOOOOOOOOOkkWW.",
		".kOOOOOOOOkkkkKk",
		".krrrrrrrrrkkkk.",
		"..krrrrrrrrrkk..",
		".kkkkkkkkkkkkkk.",
		"..k...k...k.....",
	},
	{
		"................",
		"..y..y..y..y....",
		".yYyyYyyYyyYy...",
		"...kkoooOOkk....",
		"..koooooOOOOk..k",
		".kOoooooOOOOkk.k",
		".kOOOOOOOOOkkWW.",
		".kOOOOOOOOkkkkKk",
		".krrrrrrrrrkkkk.",
		"..krrrrrrrrrkk..",
		".kkkkkkkkkkkkkk.",
		"...k...k...k....",
	},
}

var mothArt = [2][]string{
	{
		"..MMMM....MMMM..",
		".Mmmmmk..kmmmmM.",
		"MmmmmmmkkmmmmmmM",
		"MmmvvmmkkmmvvmmM",
		"MmmvKmmkkmmKvmmM",
		"MmmmmmmkkmmmmmmM",
		".MmnnnMkkMnnnmM.",
		"..MnnnMkkMnnnM..",
		"..MnnnMkkMnnnM..",
		"..MnnnMkkMnnnM..",
		"...MMM.kk.MMM...",
		"................",
	},
	{
		"................",
		"......k..k......",
		".......kk.......",
		".......kk.......",
		"..MM...kk...MM..",
		".MmmM..kk..MmmM.",
		"MmmmmM.kk.MmmmmM",
		"MmmmmM.kk.MmmmmM",
		".MmnnM.kk.MnnmM.",
		"..MnnnMkkMnnnM..",
		"...MMM.kk.MMM...",
		"................",
	},
}

var squashArt = []string{
	"................",
	"................",
	"................",
	"................",
	"................",
	"................",
	"................",
	"................",
	"...kkkkkkkkk....",
	".kkppWKppWKpkk..",
	"kqqqqqqqqqqqqqk.",
	".kkkkkkkkkkkkk..",
}

var gemArt = [2][]string{
	{
		"............",
		"...cCCCc....",
		"..cCScccc...",
		".cCCccccce..",
		"cCcccccccee.",
		"eccccccceee.",
		".ecccccceee.",
		"..eccccee...",
		"...eccee....",
		"....eee.....",
		".....e......",
		"............",
	},
	{
		"............",
		"...cCCCc....",
		"..cSCcccc...",
		".cSSScccce..",
		"cCcScccccee.",
		"eccccccceee.",
		".ecccccceee.",
		"..eccccee...",
		"...eccee....",
		"....eee.....",
		".....e......",
		"............",
	},
}

var heartArt = [2][]string{
	{
		"..............",
		"..kkkk..kkkk..",
		".krRRrkkrrrrk.",
		".kRCRRrrrrrrk.",
		"krrRRrrrrrrrXk",
		".krrrrrrrrrXk.",
		"..krrrrrrrXk..",
		"...krrrrrXk...",
		"....krrrXk....",
		".....krXk.....",
		"......kk......",
		"..............",
	},
	{
		"..............",
		"..kkkk..kkkk..",
		".kggggkkggggk.",
		".kggggggggggk.",
		"kggggggggggggk",
		".kggggggggggk.",
		"..kggggggggk..",
		"...kggggggk...",
		"....kggggk....",
		".....kggk.....",
		"......kk......",
		"..............",
	},
}

var spikeArt = []string{
	"................",
	"................",
	"................",
	"................",
	"...v.......v....",
	"...Ss......Ss...",
	"...Ss......Ss...",
	"...Ss......Ss...",
	"..SSss....SSss..",
	"..SSss....SSss..",
	"..SSss....SSss..",
	".SSSsss..SSSsss.",
	".SSSsss..SSSsss.",
	".SSSsss..SSSsss.",
	"ssssssssssssssss",
	"kkkkkkkkkkkkkkkk",
}

var cloudArt = []string{
	"................",
	"........CCCC....",
	"......CCCCCCCC..",
	".CCCCCCCCCCCCCC.",
	"vvvvvvvvvvvvvvv.",
	"vvvvvvvvvvvvvvv.",
	"vvvvvvvvvvvvvvv.",
	"vvvvvvvvvvvvvvv.",
	".VVVVVVVVVVVVV..",
	"......VVVV......",
	"................",
	"................",
	"................",
	"................",
	"................",
	"................",
}

var farCloudArt = []string{
	"...............vvvvvv...........",
	".............vvvvvvvvvv.........",
	"............vvvvvvvvvvvv........",
	"......vvvvvvvvvvvvvvvvvvv.......",
	"....vvvvvvvvvvvvvvvvvvvvvv......",
	"...vvvvvvvvvvvvvvvvvvvvvvvvv....",
	"..vvvvvvvvvvvvvvvvvvvvvvvvvvvv..",
	"..vvvvvvvvvvvvvvvvvvvvvvvvvvvvv.",
	"..vvvvvvvvvvvvvvvvvvvvvvvvvvvvv.",
	"..vvvvvvvvvvvvvvvvvvvvvvvvvvvvv.",
	"..vvvvvvvvvvvvvvvvvvvvvvvvvvvvv.",
	"...VVVVVVVVVVVVVVVVVVVVVVVVVVV..",
	"......VVVVVVVVVVVVVVVVVVVVVV....",
	"................................",
}

// The ground, the blocks and the dirt are in letters that each stage colors:
// see look.
var groundTopArt = []string{
	"TUUTUUTUUTUUTUUT",
	"TTTTTTTTTTTTTTTT",
	"TTTTTTTTTTTTTTTT",
	"TTTTTTTTTTTTTTTT",
	"ETEtEETEtEETEtEE",
	"GtGEGGtGEGGtGEGG",
	"GEGGGGEGGGGEGGGG",
	"GGGGGGGGGGGGGHGG",
	"GGGHGGGGGGGGGggG",
	"GGGggGGGGGGGGGGG",
	"GGGGGGGGGGGGGGGG",
	"GGGGGGGGGGHGGGGG",
	"GGGGGGHGGGggGGGG",
	"GGGGGGggGGGGGGGG",
	"GGGGGGGGGGGGGGGG",
	"GGGGGGGGGGGGGGGG",
}

var groundArt = [2][]string{
	{
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGHGGG",
		"GGGGGGGHGGGGggGG",
		"GGGGGGGggGGGGHGG",
		"GGGGGGGGGGGGGggG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGHHGGGGGGG",
		"GGGGGGGgggGGGGGG",
		"GGGGGGGGGGGGGGGG",
	},
	{
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGHGGGGGGGG",
		"HGGGGGGggGHGGGGG",
		"ggHGGGGGGGggGGGG",
		"GGggGGGGGGGGGGGG",
		"GGGGGGGGGGHGGGGG",
		"GGGGGGGGGGggGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
		"GGGGGGGGGGGGGGGG",
	},
}

var blockArt = []string{
	"IIIIIIIIIIIIIIII",
	"IIIIIIIIIIIIIIIj",
	"IJIJJJJJJJJIJJJj",
	"IJJjJJJJJJJJjJJj",
	"IJJJJJJJJJJJJJJj",
	"IJJJJjjjjjjJJJJj",
	"IJJJJjIIIIjJJJJj",
	"IJJJJjJJJJjJJJJj",
	"IJJJJjJJJJjJJJJj",
	"IJJJJjJJJJjJJJJj",
	"IJJJJjjjjjjJJJJj",
	"IJIJJJJJJJJIJJJj",
	"IJJjJJJJJJJJjJJj",
	"IJJJJJJJJJJJJJJj",
	"IJJJJJJJJJJJJJJj",
	"Ijjjjjjjjjjjjjjj",
}

var dirtArt = []string{
	"ZZZZZZZZZZZZZZZZ",
	"QQQQQQQQQQQQQQQq",
	"QQQQQQQQQqqQQQQq",
	"QQQqqQQQQqQQQQQq",
	"QQQqQQQQQQQQQQQq",
	"QQQQQQzQQQQQQQQq",
	"QQQQQQQzQQQQQQQq",
	"QQQQQQQzQQQQqqQq",
	"QQQQQQQQzzQQqQQq",
	"QQQQQqqQQQQQQQQq",
	"QQQQQqQQQQQQQQQq",
	"QQQQQQQQQQQQQQQq",
	"QQQQQQQQQQqqQQQq",
	"QQqqQQQQQQqQQqqq",
	"QQqQQQQQQQQQQqQq",
	"qqqqqqqqqqqqqqqq",
}

var gopherPalette = map[byte]color.RGBA{
	'B': {214, 186, 150, 255},
	'D': {176, 146, 110, 255},
	'L': {236, 214, 184, 255},
	'O': {112, 86, 60, 255},
	'K': {20, 20, 24, 255},
	'W': {250, 250, 250, 255},
	'N': {60, 40, 34, 255},
	'w': {255, 255, 248, 255},
	'h': {240, 218, 190, 255},
	'H': {214, 190, 160, 255},
}

var palette = map[byte]color.RGBA{
	'p': {130, 80, 170, 255},
	'P': {176, 128, 214, 255},
	'q': {96, 56, 130, 255},
	'k': {30, 30, 40, 255},
	'K': {20, 20, 24, 255},
	'W': {250, 250, 250, 255},
	'O': {220, 110, 40, 255},
	'o': {250, 160, 70, 255},
	'r': {200, 60, 60, 255},
	'R': {255, 140, 130, 255},
	'X': {160, 40, 48, 255},
	'y': {255, 226, 80, 255},
	'Y': {255, 250, 200, 255},
	'm': {250, 246, 236, 255},
	'n': {214, 204, 184, 255},
	'M': {150, 140, 130, 255},
	'v': {245, 248, 252, 255},
	'V': {200, 214, 232, 255},
	'c': {80, 220, 255, 255},
	'C': {220, 252, 255, 255},
	'e': {30, 140, 200, 255},
	'S': {220, 224, 232, 255},
	's': {150, 154, 164, 255},
	'g': {80, 80, 96, 255},
}

// Each stage colors its sky, the hills or rock behind, and its tiles. The
// tiles' letters are, in order: groundLetters for the grass on top, its
// shade and light, the edge under it, the ground and its dark and light
// spots; blockLetters for a block, its light and its shade; dirtLetters for
// dirt, its spots, its light and its crack.
type look struct {
	sky, horizon color.RGBA
	far, near    color.RGBA // the hills or the rocks behind
	clouds       bool       // clouds drift far behind
	ground       [7]color.RGBA
	block        [3]color.RGBA
	dirt         [4]color.RGBA
}

const (
	groundLetters = "TtUEGgH"
	blockLetters  = "JIj"
	dirtLetters   = "QqZz"
)

var looks = [3]look{
	{ // meadow
		sky: color.RGBA{100, 172, 240, 255}, horizon: color.RGBA{196, 230, 252, 255},
		far: color.RGBA{150, 204, 180, 255}, near: color.RGBA{100, 172, 112, 255},
		clouds: true,
		ground: [7]color.RGBA{{90, 190, 80, 255}, {60, 150, 60, 255}, {140, 220, 110, 255}, {96, 64, 36, 255}, {150, 100, 60, 255}, {118, 78, 44, 255}, {176, 126, 82, 255}},
		block:  [3]color.RGBA{{200, 150, 90, 255}, {236, 196, 136, 255}, {130, 90, 50, 255}},
		dirt:   [4]color.RGBA{{170, 120, 70, 255}, {120, 80, 44, 255}, {196, 146, 94, 255}, {100, 64, 36, 255}},
	},
	{ // cave
		sky: color.RGBA{30, 22, 36, 255}, horizon: color.RGBA{64, 46, 60, 255},
		far: color.RGBA{50, 40, 56, 255}, near: color.RGBA{72, 56, 70, 255},
		ground: [7]color.RGBA{{130, 120, 146, 255}, {100, 92, 116, 255}, {170, 160, 186, 255}, {50, 44, 60, 255}, {86, 76, 96, 255}, {66, 58, 76, 255}, {110, 100, 124, 255}},
		block:  [3]color.RGBA{{130, 120, 140, 255}, {170, 160, 182, 255}, {80, 70, 90, 255}},
		dirt:   [4]color.RGBA{{150, 100, 60, 255}, {100, 66, 40, 255}, {176, 126, 82, 255}, {80, 52, 30, 255}},
	},
	{ // sky
		sky: color.RGBA{64, 132, 228, 255}, horizon: color.RGBA{176, 214, 250, 255},
		far: color.RGBA{220, 236, 252, 255}, near: color.RGBA{242, 248, 255, 255},
		clouds: true,
		ground: [7]color.RGBA{{110, 210, 120, 255}, {80, 170, 90, 255}, {170, 240, 170, 255}, {120, 86, 56, 255}, {170, 130, 90, 255}, {140, 100, 66, 255}, {196, 156, 112, 255}},
		block:  [3]color.RGBA{{230, 210, 130, 255}, {250, 236, 180, 255}, {170, 140, 70, 255}},
		dirt:   [4]color.RGBA{{190, 140, 90, 255}, {140, 96, 56, 255}, {214, 166, 112, 255}, {120, 80, 46, 255}},
	},
}

var (
	imageGopher            [5]*ebiten.Image
	imageBug, imageSpiky   [2]*ebiten.Image
	imageMoth, imageGem    [2]*ebiten.Image
	imageHeart             [2]*ebiten.Image // full and empty
	imageSquash            *ebiten.Image
	imageSpike, imageCloud *ebiten.Image
	imageFarCloud          *ebiten.Image

	// Each stage's tiles.
	imageGroundTop [3]*ebiten.Image
	imageGround    [3][2]*ebiten.Image
	imageBlock     [3]*ebiten.Image
	imageDirt      [3]*ebiten.Image
)

func sprite(rows []string, pal map[byte]color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	solidArt(img, rows)
	for y, row := range rows {
		for x := range len(row) {
			if c, ok := pal[row[x]]; ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

// tileSprite draws a tile in a stage's colors: each letter of letters takes
// the color at the same place in colors.
func tileSprite(rows []string, letters string, colors []color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(len(rows[0]), len(rows))
	solidArt(img, rows)
	for y, row := range rows {
		for x := range len(row) {
			for i := range len(letters) {
				if row[x] == letters[i] {
					img.Set(x, y, colors[i])
				}
			}
		}
	}
	return img
}

func init() {
	for i := range imageGopher {
		imageGopher[i] = sprite(gopherArt[i], gopherPalette)
	}
	for f := range 2 {
		imageBug[f] = sprite(bugArt[f], palette)
		imageSpiky[f] = sprite(spikyArt[f], palette)
		imageMoth[f] = sprite(mothArt[f], palette)
		imageGem[f] = sprite(gemArt[f], palette)
		imageHeart[f] = sprite(heartArt[f], palette)
	}
	imageSquash = sprite(squashArt, palette)
	imageSpike = sprite(spikeArt, palette)
	imageCloud = sprite(cloudArt, palette)
	imageFarCloud = sprite(farCloudArt, palette)
	for s := range looks {
		lk := &looks[s]
		imageGroundTop[s] = tileSprite(groundTopArt, groundLetters, lk.ground[:])
		for v := range 2 {
			imageGround[s][v] = tileSprite(groundArt[v], groundLetters, lk.ground[:])
		}
		imageBlock[s] = tileSprite(blockArt, blockLetters, lk.block[:])
		imageDirt[s] = tileSprite(dirtArt, dirtLetters, lk.dirt[:])
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
