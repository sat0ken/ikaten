package air

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Top views
// have the nose or bow to the right and the right wing or starboard side at
// the bottom; side views have the top up. The enemy's planes carry a white
// star on blue.
var palette = map[byte]color.RGBA{
	'B': {54, 74, 104, 255},
	'D': {160, 120, 76, 255},
	'G': {150, 154, 160, 255},
	'K': {20, 22, 26, 255},
	'O': {255, 150, 40, 255},
	'R': {220, 50, 30, 255},
	'S': {150, 150, 156, 255},
	'W': {244, 246, 250, 255},
	'Y': {255, 226, 90, 255},
	'b': {84, 104, 134, 255},
	'c': {150, 200, 230, 255},
	'd': {130, 96, 60, 255},
	'g': {120, 124, 130, 255},
	'k': {60, 62, 68, 255},
	'n': {40, 56, 80, 255},
	'r': {120, 40, 36, 255},
	's': {110, 110, 116, 255},
	'u': {30, 50, 120, 255},
	'v': {255, 255, 255, 255},
	'w': {200, 204, 210, 255},
}

var fighterTop = []string{
	"........K.............",
	"........KKuuu.........",
	"........KuuWuu........",
	"........KuWWWu........",
	"........KuuWuu........",
	"........KBuuu.........",
	"........KBBBK.........",
	"........KBBBK.........",
	".KKK....KBBBK.........",
	".KBK....KBBBK.........",
	".KBK....KBBBK.........",
	".KBBKKKKbbbbbKKKKKK..k",
	".KbbbbbbbbbbcccbbbbKKK",
	".KbbbbbbbbbbcccbbbbKKK",
	".KnnKKKKbbbbbKKKKKK..k",
	".KnK....KnnnK.........",
	".KnK....KnnnK.........",
	".KKK....KnnnK.........",
	"........KnnnK.........",
	"........KnnnK.........",
	"........Knuuu.........",
	"........KuuWuu........",
	"........KuWWWu........",
	"........KuuWuu........",
	"........KKuuu.........",
	"........K.............",
}

var fighterSide = []string{
	"KKK...................",
	"KbK.........cc........",
	"KbK.......ccccc.......",
	"KbbKKKuuuKcccccKKKK..K",
	"KbbBBuuWuuBBccBBBBBKKK",
	"KKKBBuWWWuBBBBBBBBBKKK",
	"...KKuuWuuKKKKKKKKK..K",
	"......uuu.............",
}

var bomberTop = []string{
	"...........K................",
	"...........KKuuuK...........",
	"...........KuuWuu...........",
	"...........KuWWWu...........",
	"...........KuuWuu...........",
	"...........KBuuuK...........",
	"...........KBBBBK...........",
	"...........KBBBBK...........",
	"...........KBBBBK...........",
	"...........KBBBBK...........",
	".KKKK......KBBBBK...........",
	".KBBK......KBBBBK...........",
	".KBBK......KBBBBK...........",
	".KBBK....KKbbbbbbKKKK.......",
	".KBbbKKKKbbbbbbbbbbbbKKKKK.k",
	".KbbbbbbbbbbbbbcccbbbbbbbbKK",
	".KbbbbbbbbbbbbbcccbbbbbbbbKK",
	".KnbbKKKKbbbbbbbbbbbbKKKKK.k",
	".KnnK....KKbbbbbbKKKK.......",
	".KnnK......KnnnnK...........",
	".KnnK......KnnnnK...........",
	".KKKK......KnnnnK...........",
	"...........KnnnnK...........",
	"...........KnnnnK...........",
	"...........KnnnnK...........",
	"...........KnnnnK...........",
	"...........KnuuuK...........",
	"...........KuuWuu...........",
	"...........KuWWWu...........",
	"...........KuuWuu...........",
	"...........KKuuuK...........",
	"...........K................",
}

var bomberSide = []string{
	"KKK.........................",
	"KbK.........................",
	"KbK............cc...........",
	"KbK...........ccccc.........",
	"KbK.KKKKuuuKKKcccccKKKKK...K",
	"KbbKBBBuuWuuBBBccBBBBBBBKKKK",
	"KKKKBBBuWWWuBBBBBBBBBBBBKKKK",
	"....KKKuuWuuKKKKKKKKKKKK...K",
	"........uuu.................",
}

// Decks, seen from above.
var deckArt = [shipKinds][]string{
	carrier: {
		".....kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkgGGGGGGgkkkkkkkkkkkkkkkkkk........",
		"..kkkDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdggggggggDDDDDdDDDDDDdDDDDDkkkkkk..",
		"kkDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdggggggggDDDDDdDDDDDDdDDDDDDdWDDDkk",
		"kDDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDDk",
		"kDDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDDk",
		"kDDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDDk",
		"WWWWWDWWWWDDWWWWDDWWWWDDWWWWdDWWWWDdWWWWDDWWWWDDWWWWDDWWWWDDWWWWDDWWWWdDWWWWDdWWWWDk",
		"kDDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDDk",
		"kDDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDDk",
		"kkDDWDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdWDDDkk",
		"..kkkDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDDdDDDDDkkkkkk..",
		".....kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk........",
	},
	cruiser: {
		"......kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk......",
		"..kkkkggggggggggggggggggggggggggggggggggggggggggggggggkkkk..",
		"kkggggGGGGggggGGGGggggggssssssssssssggggGGGGggggGGGGggggggkk",
		"kgggggGGGGggggGGGGggggggssKKKssKKKssggggGGGGggggGGGGgggggggk",
		"kgggggGGGGKKggGGGGKKggggssKKKssKKKssggggGGGGKKggGGGGKKgggggk",
		"kkggggGGGGggggGGGGggggggssssssssssssggggGGGGggggGGGGggggggkk",
		"..kkkkggggggggggggggggggggggggggggggggggggggggggggggggkkkk..",
		"......kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk......",
	},
	destroyer: {
		".....kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk.....",
		".kkkkgGGggggggggssssssssssgggggggggGGggkkkk.",
		"kggggGGGGgggggggssKKssKKssggggggggGGGGgggggk",
		"kggggGGGGKggggggssKKssKKssggggggggGGGGKggggk",
		".kkkkgGGggggggggssssssssssgggggggggGGggkkkk.",
		".....kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk.....",
	},
}

// Superstructures above the deck, seen from the side.
var upperArt = [shipKinds][]string{
	carrier: {
		"....................................................kk..............................",
		"....................................................kk..............................",
		"..................................................kkkkkk............................",
		"....................................................kk..............................",
		"................................................kkkksskkkk..........................",
		"................................................kKKKKKKKKk..........................",
		"................................................kssssssssSkk........................",
		"................................................kssssssssSSk........................",
		"................................................kssssssssSSk........................",
		"................................................kssssssssSSk........................",
		"................................................kssssssssSSk........................",
		"................................................kkkkkkkkkkkk........................",
	},
	cruiser: {
		"...........................k................................",
		"..........................ksk...............................",
		"..........................ksk..kkk..........................",
		"..........................ksk..ksk..........................",
		"........................kkssskkssskk........................",
		"........................kKKKKKKKKKKk........................",
		"........................kssssssssssk........................",
		".....kkkkkk..kkkkkk.....kssssssssssk...kkkkkk..kkkkkk.......",
		".....kGGGGGKKGGGGGGKKKK.kssssssssssk...kGGGGGKKGGGGGGKKKK...",
		".....kkkkkk..kkkkkk.....kkkkkkkkkkkk...kkkkkk..kkkkkk.......",
	},
	destroyer: {
		"..................k.........................",
		".................ksk........................",
		".................ksk..kk....................",
		".................ksk..kk....................",
		"................kssskksskk..................",
		".....kkkkk......kssssssssk........kkkkk.....",
		".....kGGGGKKK...kssssssssk........kGGGGKKK..",
		".....kkkkk......kkkkkkkkkk........kkkkk.....",
	},
}

// The hull's side above the water, stretched to each ship's length.
var hullArt = []string{
	"gggggggggggggggggggggggggggggggggggggggggggggggg",
	"kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk",
	".kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk.",
	"..kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk..",
	"...rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr...",
	"....rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr....",
}

// The wake behind a ship on the sea, stern to the right.
var wakeArt = []string{
	"wwww............................",
	".wwwwwwww.......................",
	"......wwwwwwww..................",
	"...........wwwwwwww.............",
	"................wwwwwwww........",
	".....................wwwwwwww...",
	"...........WWWWWWWWWWWWWWWWWWWWW",
	"...........WWWWWWWWWWWWWWWWWWWWW",
	".....................wwwwwwww...",
	"................wwwwwwww........",
	"...........wwwwwwww.............",
	"......wwwwwwww..................",
	".wwwwwwww.......................",
	"wwww............................",
}

var cloudArt = []string{
	"....................vvvvvvvvvvvv................",
	"..................vvvvvvvvvvvvvvvv..............",
	"................vvvvvvvvvvvvvvvvvvvv............",
	"...............vvvvvvvvvvvvvvvvvvvvvv...........",
	"..........WWWWWWWWWWWWWWWWWWWWWWWWWWW...........",
	".......WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW......",
	"......WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW....",
	".....WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW...",
	"....WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW..",
	"....WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW..",
	"....WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW..",
	"...WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW..",
	"..wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww..",
	"...wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww...",
	"......wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww......",
	"..........wwwwwwwwwwwwwwwwwwwwwwwww.............",
}

var puffArt = [4][]string{
	puffFlak: {
		"....kk....",
		"..kkkkkk..",
		".kKKKKkkk.",
		"kKKgKKKkkk",
		"kKKKKKKkkk",
		"kkKKKKkkkk",
		"kkkkkkkkkk",
		".kkkkkkkk.",
		"..kkkkkk..",
		"....kk....",
	},
	puffSmoke: {
		"....ss....",
		"..ssssss..",
		".sSSSSsss.",
		"sSSSSSSsss",
		"sSSSSSSsss",
		"ssSSSSssss",
		"ssssssssss",
		".ssssssss.",
		"..ssssss..",
		"....ss....",
	},
	puffFire: {
		"..........",
		".....O....",
		"...RRRR...",
		"..RROORR..",
		".RROOOORR.",
		"RROOYYOORR",
		"RROYYYYORR",
		".RROYYORR.",
		"..RROORR..",
		"...RRRR...",
	},
	puffSplash: {
		"..........",
		"....Ww....",
		"....wW....",
		"...wWWw...",
		"...WWwW...",
		"..WWwWWw..",
		"..WwWWwW..",
		".WwWWwWWw.",
		".wWWwWWwW.",
		"wWWwWWwWWw",
	},
}

// sprites holds a game's images. A Game makes them in NewGame and drops them
// with it, so that the launcher keeps only the running game's in RAM.
type sprites struct {
	fighter, bomber [2]*ebiten.Image // top and side
	deck, upper     [shipKinds]*ebiten.Image
	hull, wake      *ebiten.Image
	cloud, cowl     *ebiten.Image
	puff            [4]*ebiten.Image
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

// The cowling ahead of the pilot: the top of a dark ellipse with a rim and
// a light along its crest.
const cowlW, cowlH = 128, 18

func cowl() *ebiten.Image {
	img := ebiten.NewImage(cowlW, cowlH)
	for y := range cowlH {
		cowlRow(img, y)
	}
	return img
}

func cowlRow(img *ebiten.Image, y int) {
	const w, h = cowlW, cowlH
	dark, rim, light := color.RGBA{34, 38, 34, 255}, color.RGBA{70, 76, 66, 255}, color.RGBA{58, 64, 56, 255}
	dy := h - y
	for x := range w {
		dx := x*2 - w + 1
		// Inside (dx/w)² + (dy/h)² < 1, in integers.
		v := dx*dx*h*h + 4*dy*dy*w*w
		switch {
		case v < 4*w*w*h*h*70/100 && dx > -w/3 && dx < w/3 && v > 4*w*w*h*h*55/100:
			img.Set(x, y, light)
		case v < 4*w*w*h*h*86/100:
			img.Set(x, y, dark)
		case v < 4*w*w*h*h:
			img.Set(x, y, rim)
		}
	}
}

func newSprites() *sprites {
	a := &sprites{}
	a.fighter = [2]*ebiten.Image{sprite(fighterTop), sprite(fighterSide)}
	a.bomber = [2]*ebiten.Image{sprite(bomberTop), sprite(bomberSide)}
	for k := range shipKinds {
		a.deck[k] = sprite(deckArt[k])
		a.upper[k] = sprite(upperArt[k])
	}
	a.hull = sprite(hullArt)
	a.wake = sprite(wakeArt)
	a.cloud = sprite(cloudArt)
	for i := range a.puff {
		a.puff[i] = sprite(puffArt[i])
	}
	a.cowl = cowl()
	return a
}
