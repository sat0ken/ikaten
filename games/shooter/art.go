package shooter

import (
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. They are
// drawn at twice the detail of the game's positions.
var palette = map[byte]color.RGBA{
	'B': {40, 90, 220, 255},
	'C': {80, 220, 255, 255},
	'G': {60, 220, 90, 255},
	'K': {20, 20, 36, 255},
	'O': {255, 150, 30, 255},
	'P': {180, 80, 230, 255},
	'R': {230, 50, 50, 255},
	'W': {255, 255, 255, 255},
	'Y': {255, 230, 60, 255},
	'd': {20, 120, 50, 255},
	'g': {100, 110, 130, 255},
	'p': {100, 40, 140, 255},
	'r': {140, 20, 30, 255},
	'w': {190, 200, 220, 255},
}

var playerArt = [2][]string{{
	"..........KK..........",
	"..........KK..........",
	"..........KK..........",
	"..........KK..........",
	".........KwwK.........",
	".........KWCK.........",
	".........KCCK.........",
	".........KCCK.........",
	"........KwCCwK........",
	"........KwwwwK........",
	".......KwwwwwwK.......",
	".....KKwwBBBBwwKK.....",
	"....KwwwwBCCBwwwwK....",
	"..KKwwwwwBBBBwwwwwKK..",
	".KwwwwwwwwwwwwwwwwwwK.",
	"KggwwRwwwwwwwwwwRwwggK",
	"KggggggggwwwwggggggggK",
	"KggKKKKKKwwwwKKKKKKggK",
	".KK......KwwK......KK.",
	".........KKKK.........",
	".........O..O.........",
	".........O..O.........",
	".........Y..Y.........",
	".........R..R.........",
}, {
	"..........KK..........",
	"..........KK..........",
	"..........KK..........",
	"..........KK..........",
	".........KwwK.........",
	".........KWCK.........",
	".........KCCK.........",
	".........KCCK.........",
	"........KwCCwK........",
	"........KwwwwK........",
	".......KwwwwwwK.......",
	".....KKwwBBBBwwKK.....",
	"....KwwwwBCCBwwwwK....",
	"..KKwwwwwBBBBwwwwwKK..",
	".KwwwwwwwwwwwwwwwwwwK.",
	"KggwwRwwwwwwwwwwRwwggK",
	"KggggggggwwwwggggggggK",
	"KggKKKKKKwwwwKKKKKKggK",
	".KK......KwwK......KK.",
	".........KKKK.........",
	".........Y..Y.........",
	".........Y..Y.........",
	".........O..O.........",
	"......................",
}}

var scoutArt = []string{
	"KKKKK........KKKKK",
	"KrrrrK......KrrrrK",
	".KrrrRK....KRrrrK.",
	".KrrrRRK..KRRrrrK.",
	"..KrrRRRKKRRRrrK..",
	"..KrrRRRRRRRRrrK..",
	"...KrRRRYYRRRrK...",
	"....KRRRWYRRRK....",
	"....KRRRYYRRRK....",
	".....KRRYYRRK.....",
	".....KRRRRRRK.....",
	"......KRRRRK......",
	"......KRRRRK......",
	".......KRRK.......",
	".......KRRK.......",
	"........KK........",
}

var weaverArt = []string{
	"..................",
	".....KKKKKKKK.....",
	"...KKddddddddKK...",
	"..KddddGGGGGdddK..",
	".KdddGGGGGGGGGddK.",
	".KddGGGWWWWGGGGdK.",
	".KddGGWWKKWWGGGdK.",
	".KdGGGWKWKKWGGGGK.",
	".KdGGGWKKKKWGGGGK.",
	".KdGGGWWKKWWGGGGK.",
	".KddGGGWWWWGGGGdK.",
	".KddGGGGGGGGGGGdK.",
	"..KddGGGGGGGGGdK..",
	"...KKddGGGGGdKK...",
	".....KKKKKKKK.....",
	"..................",
}

var gunnerArt = []string{
	"...KK............KK...",
	"..KppK..........KppK..",
	".KpPPpK.KKKKKK.KpPPpK.",
	".KPPPPpKPPPPPPKpPPPPK.",
	".KPPPPpPPPPPPPPpPPPPK.",
	".KpPPppPPPPPPPPppPPpK.",
	"..KpppPPPPYYPPPPpppK..",
	"...KpPPPPYYYYPPPPpK...",
	"....KPPPYYKKYYPPPK....",
	"....KPPPYYKKYYPPPK....",
	".....KPPPYYYYPPPK.....",
	".....KPPPPYYPPPPK.....",
	"......KKppppppKK......",
	"........KKppKK........",
	"..........KK..........",
	"..........KK..........",
	"..........OO..........",
	"..........OO..........",
}

// The boss is symmetric: each row is mirrored around its last character.
var bossHalf = []string{
	"................KKKKKKKKKKKKKKKKK",
	"..............KKpPPPPPPPPPPPPPPPP",
	"............KKpPPPPPPPPPPPPPPPPPP",
	"..........KKpPPPPPPPPPPPPPPPPPPPP",
	"........KKpPPPPPPPPPPPPPPPPPPPPPP",
	"......KKpPPPPPPPPPPPPPPPPPPPPgggg",
	"....KKpPppppggggppppppppppppggggg",
	"..KKpPPPPPPggRRggPPPPPPPPPPggwwww",
	".KppPPPPPPPgRRRRgPPPPPPPPPggwwwCC",
	".KppPPPPPPPgRRRRgPPPPPPPPPggwwCCC",
	".KppPPPPpppggRRggpppppppppggwwCCC",
	".KppPPPPPPPPggggPPPPPPPPPPggwwwCC",
	"..KppPPPPPPPPgggPPPPPPPPPPPggwwww",
	"..KppPPPPPPPPgggPPPPPPPPPPPPggggg",
	"...KppPPpppppgggpppppppppppppgggg",
	"....KKpPPPPPPgggPPPPPPPPPPPPPPPPP",
	"......KKpPPPPgggPPPPPPPPPPPPPPPPP",
	"........KKpPPgggPPPPPPPPPPPPPPPPP",
	"..........KKpPPPPPPPYYYYPPPPPPPPP",
	"............KKpPPPPPYYYYPPPPPPPPP",
	"..............KKKPPPYYWYPPPPPPPPP",
	".................KKKYYYYppppppppp",
	"....................KKppppppppppp",
	"......................KKppppppppp",
	"....OOOO................KKppppppp",
	"....OOOO..................KKppppp",
	"....OOOO....................KKppp",
	"....OOOO......................KKK",
}

var shotArt = []string{
	".WW.",
	".WW.",
	"CWWC",
	"CWWC",
	"CWWC",
	"CWWC",
	"CWWC",
	".BB.",
	".BB.",
	".BB.",
}

var bulletArt = []string{
	".RRRR.",
	"RRWWRR",
	"RWWWWR",
	"RWWWWR",
	"RRWWRR",
	".RRRR.",
}

var itemArt = []string{
	"....KKKKKK....",
	"...KOOYYOOK...",
	"..KOrrrrYYOK..",
	".KOYrrrrYYYOK.",
	"KOYYrrYYrrYYOK",
	"KOYYrrYYrrYYOK",
	"KYYYrrrrYYYYYK",
	"KYYYrrrrYYYYYK",
	"KOYYrrYYYYYYOK",
	"KOYYrrYYYYYYOK",
	".KOYrrYYYYYOK.",
	"..KOrrYYYYOK..",
	"...KOOYYOOK...",
	"....KKKKKK....",
}

var lifeArt = []string{
	"....KK....",
	"....KK....",
	"....KK....",
	"...KCwK...",
	"...KwwK...",
	".KKwwwwKK.",
	"KwwwwwwwwK",
	"KKKwwwwKKK",
	"...KwwK...",
	"....KK....",
}

var bombArt = []string{
	"....OO....",
	"....OO....",
	"...KBBK...",
	"..KBBBBK..",
	".KBBCBBBK.",
	".KBBWBBBK.",
	".KBBBBBBK.",
	".KBBBBBBK.",
	"..KBBBBK..",
	"...KKKK...",
}

const blastFrames = 4

// sprites holds a game's images. A Game makes them in NewGame and drops them
// with it, so that the launcher keeps only the running game's in RAM.
type sprites struct {
	player     [2]*ebiten.Image
	enemy      [4]*ebiten.Image // indexed by kind
	shot       *ebiten.Image
	bullet     *ebiten.Image
	item       *ebiten.Image
	life       *ebiten.Image
	bomb       *ebiten.Image
	blast      [blastFrames]*ebiten.Image
	star       [3]*ebiten.Image
	bar, barBG *ebiten.Image
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

func mirrored(half []string) []string {
	rows := make([]string, len(half))
	for i, h := range half {
		b := []byte(h)
		for j := len(h) - 2; j >= 0; j-- {
			b = append(b, h[j])
		}
		rows[i] = string(b)
	}
	return rows
}

func pixel(c color.RGBA, w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

// blast draws expanding rings: a white core that turns into red sparks.
func blast(frame int) *ebiten.Image {
	const size = 30
	img := ebiten.NewImage(size, size)
	colors := [blastFrames][3]color.RGBA{
		{{255, 255, 255, 255}, {255, 230, 60, 255}, {255, 250, 200, 255}},
		{{255, 230, 60, 255}, {255, 150, 30, 255}, {255, 240, 150, 255}},
		{{255, 150, 30, 255}, {230, 50, 50, 255}, {255, 200, 80, 255}},
		{{230, 50, 50, 255}, {140, 20, 30, 255}, {200, 80, 60, 255}},
	}
	outer := 5 + float64(frame)*3.4
	inner := outer - 4 - float64(frame)*0.6
	for y := range size {
		for x := range size {
			d := math.Hypot(float64(x)-size/2+0.5, float64(y)-size/2+0.5)
			switch {
			case d > outer:
			case d > outer-1.5:
				// Later frames break the ring into sparks.
				if frame < 2 || (x/2+y/2+frame)%3 != 0 {
					img.Set(x, y, colors[frame][1])
				}
			case d > inner:
				if frame < 3 || (x+y)%2 == 0 {
					img.Set(x, y, colors[frame][2])
				}
			case frame < 2:
				img.Set(x, y, colors[frame][0])
			}
		}
	}
	return img
}

func newSprites() *sprites {
	a := &sprites{}
	for i := range a.player {
		a.player[i] = sprite(playerArt[i])
	}
	a.enemy[scout] = sprite(scoutArt)
	a.enemy[weaver] = sprite(weaverArt)
	a.enemy[gunner] = sprite(gunnerArt)
	a.enemy[boss] = sprite(mirrored(bossHalf))
	a.shot = sprite(shotArt)
	a.bullet = sprite(bulletArt)
	a.item = sprite(itemArt)
	a.life = sprite(lifeArt)
	a.bomb = sprite(bombArt)
	for i := range a.blast {
		a.blast[i] = blast(i)
	}
	a.star[0] = pixel(color.RGBA{70, 70, 100, 255}, 1, 1)
	a.star[1] = pixel(color.RGBA{140, 140, 180, 255}, 2, 2)
	a.star[2] = pixel(color.RGBA{255, 255, 255, 255}, 2, 4)
	a.bar = pixel(color.RGBA{230, 50, 50, 255}, 1, 1)
	a.barBG = pixel(color.RGBA{60, 20, 30, 255}, 1, 1)
	return a
}

func init() {
	for i := range dirs {
		a := 2 * math.Pi * float64(i) / float64(len(dirs))
		dirs[i] = vec{int(math.Round(math.Cos(a) * 256)), int(math.Round(math.Sin(a) * 256))}
	}
}
