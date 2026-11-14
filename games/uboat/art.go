package uboat

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
)

// Sprites are pixel art written as strings; '.' is transparent. Ships are seen
// from the side with the bow to the right.
var palette = map[byte]color.RGBA{
	'K': {58, 62, 70, 255},    // merchant hull
	'k': {96, 102, 110, 255},  // the deck edge
	'n': {72, 82, 96, 255},    // destroyer hull
	'g': {110, 116, 124, 255}, // masts, funnels, guns
	'G': {160, 166, 172, 255},
	'w': {200, 204, 208, 255}, // superstructure
	'c': {40, 48, 60, 255},    // windows
	'b': {24, 24, 28, 255},
	'r': {120, 40, 40, 255}, // boot-top
	'B': {110, 80, 50, 255}, // hatches
	'Y': {255, 230, 80, 255},
	'O': {255, 140, 30, 255},
	'R': {220, 50, 30, 255},
	'S': {90, 90, 96, 255},
	's': {150, 150, 156, 255},
	'W': {235, 240, 245, 255},
}

var shipArt = [kinds][]string{
	cargo: {
		".................................bbbbb..................................",
		"..............g..................ggggg................g.................",
		"..............g..................RRRRR................g.................",
		"............ggggg............bbbbgggggbbb...........ggggg...............",
		".............bgb............bwwwwwwwwwwwwbbb.........bgb................",
		"............b.g.b...........bcwcwcwcwcwcwcwb........b.g.b...............",
		"...........b..g..b..........bwwwwwwwwwwwwwwb.......b..g..b..............",
		"..........b...g...b.......bbwwwwwwwwwwwwwwwwbb....b...g...BBBBBB........",
		".........b....g....b......bcwcwcwcwcwcwcwcwcwb...b....g....bkkkkkkkkkkkk",
		"kkkkBBBBBB....g...BBBBBBBBwwwwwwwwwwwwwwwwwwwwBBBBBB..g...BBBBBBKKKKKKKK",
		"KKKKKKKKKKkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkKKKKKKKKKKKK",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK..",
		".KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK...",
		".KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK....",
		"..rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr.....",
		"...rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr......",
	},
	tanker: {
		"....g.bbbbbb................................................g...........................",
		"....g.gggggg................................................g...........................",
		"..bbGbGGGGGGbbbbbb.....................................bbbbbGbbbbb......................",
		"...bwwwwwwwwwwwwb.......................................bwwwwwwwb.......................",
		"...bcwcwcwcwcwcwb.......................................bcwcwcwcb.......................",
		"...bwwwwwwwwwwwwb.......................................bwwwwwwwb.......................",
		"...bcwwcwwcwwcwwb.......................................bwwwwwwwb.......................",
		"kkkkkkkkkkkkkkkkkkggggggbgggbgggbgggbgggbgggbgggbgggbggggbbbbbbbbgggbgggbgggbgggbgggbg..",
		"KKKKKKKKKKKKKKKKKK..g.bGb....gGGb.....bGb.....bGb.......g........g........gkkkkkkkkgkkkk",
		"KKKKKKKKKKKKKKKKKKkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkKKKKKKKKKKKKKK",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK.",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK..",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK...",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK....",
		".rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr.....",
		"..rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr......",
	},
	destroyer: {
		"...............................................g....................",
		".............................................ggggg..................",
		"..............................................ggg...................",
		"..............................................ggg...................",
		"............................bbbb....bbbb....bbGGGbb.................",
		"............................ggggg...ggggg...bcGcGcb.................",
		"............................ggggg...ggggg..bGGGGGGGb...bbbbbgggg....",
		"............................gggg....gggg...bGGGGGGGb...bGGGb........",
		"..ggggbbbbb....bbbb.........gggg....gggg...bGGGGGGGb...bGGGb........",
		"......bGGGGggggGGGb...ggggg.gggg....ggggkkkkkkkkkkkkkkkkkkkkkkkkkkkk",
		".bbb..bGGGb....bGGb...ggggg...........nnnnnnnnnnnnnnnnnnnnnnnnnnnnnn",
		"kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkknnnnnnnnnnnnnnnnnnnnnnnnnnnnnn",
		"nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn.",
		"nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn..",
		"nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn...",
		"nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn.....",
		".rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr......",
		".rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr.......",
	},
}

var fireArt = []string{
	"..................",
	"........OO........",
	"........OO........",
	".......RYYR..R....",
	"....R..OYYO..O....",
	"....O.ROYYORROR...",
	"...RORROYYOROYO...",
	"...OYOROYYOROYO...",
	"..ROYOROYYOROYOR..",
	"..ROYOOOYYOOOYOR..",
	"..RYYYOOYYOOYYYR..",
	"..OYYYOYYYYOYYYO..",
	".ROYYYOYYYYOYYYOR.",
	".ROYYYOYYYYOYYYOR.",
}

var smokeArt = []string{
	"....SSSSSS....",
	"..SSSSSSSSSS..",
	".SSSSSSSSSSSS.",
	"SSSssssSSSSSSS",
	"SSsssssSSSSSSS",
	"SsssssssSSSSSS",
	"SssssssssSSSSS",
	".ssssssssSSSS.",
	"..ssssssSSSS..",
	"....SSSSSS....",
}

var splashArt = []string{
	"..W.WW..W.",
	".W.sWWs...",
	"...sWWsW..",
	"..ssWWss.W",
	"W.sWWWWs..",
	"..sWWWWs..",
	"..sWWWWs..",
	".sWWWWWWs.",
	".sWWWWWWs.",
	".sWWWWWWs.",
	"sWWWWWWWWs",
	"sWWWWWWWWs",
}

// The sky and sea strips repeat every stripWidth pixels. They are opaque,
// with the sky's and the sea's colors behind; see assets/gen.go.
const stripWidth = 512

// sprites are the game's images. They belong to the Game, so that the
// launcher frees them on returning to its menu.
type sprites struct {
	ship                [kinds]*ebiten.Image
	fire, smoke, splash *ebiten.Image
	clouds, waves       *ebiten.Image    // in flash, drawn by assets/gen.go
	frame               [frameBands]band // black outside the periscope's circle
	sonar               *ebiten.Image    // rings and cross of the sonar plot
	sky, sea, black     *ebiten.Image
	white, yellow, red  *ebiten.Image
	green, dim, gray    *ebiten.Image
	contact             [kinds]*ebiten.Image
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

func pixel(c color.RGBA, w, h int) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	img.Fill(c)
	return img
}

func newSprites() *sprites {
	a := &sprites{}
	for k := range a.ship {
		a.ship[k] = sprite(shipArt[k])
	}
	a.fire = sprite(fireArt)
	a.smoke = sprite(smokeArt)
	a.splash = sprite(splashArt)
	a.waves, a.clouds, a.sonar = loadAssets()
	a.frame = periscopeFrame()
	a.sky = pixel(color.RGBA{150, 182, 214, 255}, 1, 1)
	a.sea = pixel(color.RGBA{40, 70, 108, 255}, 1, 1)
	a.black = pixel(color.RGBA{0, 0, 0, 255}, 1, 1)
	a.white = pixel(color.RGBA{240, 240, 240, 255}, 1, 1)
	a.yellow = pixel(color.RGBA{255, 220, 40, 255}, 1, 1)
	a.red = pixel(color.RGBA{230, 40, 30, 255}, 1, 1)
	a.green = pixel(color.RGBA{60, 200, 110, 255}, 1, 1)
	a.dim = pixel(color.RGBA{24, 80, 48, 255}, 1, 1)
	a.gray = pixel(color.RGBA{70, 70, 76, 255}, 1, 1)
	a.contact[cargo] = pixel(color.RGBA{150, 230, 170, 255}, 4, 4)
	a.contact[tanker] = pixel(color.RGBA{150, 230, 170, 255}, 6, 4)
	a.contact[destroyer] = pixel(color.RGBA{240, 70, 50, 255}, 5, 5)
	return a
}

// The periscope's frame is black outside the circle with a rim just outside
// it. It used to be one 240x240 image in flash, but flash reads are slow on
// the badge: drawing it took about 5 ms a frame. Now each band of rows is a
// black rectangle at each side, plus a narrow piece in RAM for the curve and
// the rim, drawn as is on the left and flipped on the right.
const (
	rimWidth   = 4
	pieceWidth = 24 // the most the curve's x changes over a band
	frameBands = 20
)

type band struct {
	y0, y1 int // rows y0 to y1-1
	black  int // the black rectangle's width at each side
	piece  *ebiten.Image
}

// periscopeFrame splits the rows into bands, top to bottom.
func periscopeFrame() [frameBands]band {
	const outer = scopeRadius + rimWidth
	// first returns the first x of row y within radius r of the center, or the
	// center if none is.
	first := func(y int, r float64) int {
		for x := range ScreenWidth / 2 {
			if math.Hypot(float64(x)+.5-scopeX, float64(y)+.5-scopeY) <= r {
				return x
			}
		}
		return ScreenWidth / 2
	}
	pieces := ebiten.NewImage(pieceWidth, ScreenHeight)
	var bands [frameBands]band
	n := 0
	for y := 0; y < ScreenHeight; n++ {
		if n == frameBands {
			panic("uboat: too many frame bands")
		}
		b := &bands[n]
		b.y0, b.black = y, first(y, outer)
		inner := first(y, scopeRadius)
		for y++; y < ScreenHeight; y++ {
			lo, hi := min(b.black, first(y, outer)), max(inner, first(y, scopeRadius))
			if hi-lo > pieceWidth {
				break
			}
			b.black, inner = lo, hi
		}
		b.y1 = y
		for py := b.y0; py < b.y1; py++ {
			for px := range pieceWidth {
				dx, dy := float64(b.black+px)+.5-scopeX, float64(py)+.5-scopeY
				d := math.Hypot(dx, dy)
				switch {
				case d > outer:
					pieces.Set(px, py, color.RGBA{0, 0, 0, 255})
				case d > scopeRadius:
					// The light falls from above onto the rim.
					l := 0.5 - 0.5*dy/d
					v := uint8(22 + 40*l*(outer-d)/rimWidth)
					pieces.Set(px, py, color.RGBA{v, v + 4, v + 10, 255})
				}
			}
		}
		b.piece = pieces.SubImage(image.Rect(0, b.y0, inner-b.black, b.y1)).(*ebiten.Image)
	}
	return bands
}
