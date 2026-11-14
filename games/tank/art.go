package tank

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/tank/field"
)

var (
	colorSky     = color.RGBA{96, 146, 214, 255}
	colorReticle = color.RGBA{16, 16, 16, 255}
	colorCross   = color.RGBA{240, 240, 230, 255}
	colorRed     = color.RGBA{240, 50, 40, 255}
	colorYellow  = color.RGBA{250, 210, 50, 255}
	colorGreen   = color.RGBA{80, 220, 100, 255}
	colorGray    = color.RGBA{170, 176, 170, 255}
	colorShadow  = color.RGBA{10, 14, 10, 255}
)

// skyLowBand is how many rows above the horizon the sky is pale; the hills
// strip has the same sky behind it (see assets/gen.go).
const skyLowBand = 24

// sprites are the game's images. They belong to the Game, so that the
// launcher frees them on returning to its menu.
type sprites struct {
	// In flash, drawn by assets/gen.go.
	ground *ebiten.Image
	tank   [field.TankKinds][2][field.TankFrames]*ebiten.Image // kind, hull or turret, frame
	house  [field.HouseFrames]*ebiten.Image
	tree   [2]*ebiten.Image // round and pine
	hills  *ebiten.Image

	// Drawn here.
	puff   [puffKinds]*ebiten.Image
	radar  *ebiten.Image
	hull   *ebiten.Image // the player's hull on the radar
	marker *ebiten.Image // over enemies in the wide view
	row    *ebiten.Image // one row of the ground, drawn and then copied
	sight  [sightBands]band

	sky, grass, haze          *ebiten.Image
	reticle, cross            *ebiten.Image
	red, yellow, green, gray  *ebiten.Image
	dim, shade, tracer, black *ebiten.Image
	dot                       [field.TankKinds]*ebiten.Image // a far tank, too small for its sprite
}

const grassColor = 0x60963e // under the ground texture, beyond the field

func pixel(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

const puffSize = 18 // pixels across a puff image

// puffColors are the edge and core colors of each puff.
var puffColors = [puffKinds][2]color.RGBA{
	dust:  {{140, 118, 84, 255}, {170, 148, 110, 255}},
	smoke: {{58, 58, 60, 255}, {88, 88, 90, 255}},
	fire:  {{240, 110, 30, 255}, {255, 220, 90, 255}},
	spark: {{255, 200, 80, 255}, {255, 250, 200, 255}},
	flash: {{255, 170, 60, 255}, {255, 245, 190, 255}},
}

func cut(sheet *ebiten.Image, x, y, w, h int) *ebiten.Image {
	return sheet.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image)
}

func newSprites() *sprites {
	a := &sprites{}
	var sheet *ebiten.Image
	a.ground, sheet, a.hills = loadAssets()
	// SubImage allocates, so the frames are cut once here.
	for k := range field.TankKinds {
		for part := range 2 {
			for f := range field.TankFrames {
				x, y := field.TankCell(k, part, f)
				a.tank[k][part][f] = cut(sheet, x, y, field.TankW, field.TankH)
			}
		}
	}
	for f := range a.house {
		a.house[f] = cut(sheet, f*field.HouseW, field.HouseY, field.HouseW, field.HouseH)
	}
	for i := range a.tree {
		a.tree[i] = cut(sheet, field.TreeX+i*field.TreeW, field.HouseY, field.TreeW, field.TreeH)
	}
	a.dot[light] = pixel(color.RGBA{150, 132, 90, 255})
	a.dot[medium] = pixel(color.RGBA{100, 104, 96, 255})
	a.dot[heavy] = pixel(color.RGBA{70, 74, 86, 255})
	a.sky = pixel(colorSky)
	a.grass = pixel(color.RGBA{grassColor >> 16, grassColor >> 8 & 0xff, grassColor & 0xff, 255})
	a.haze = pixel(color.RGBA{160, 186, 190, 255})
	a.reticle = pixel(colorReticle)
	a.cross = pixel(colorCross)
	a.red = pixel(colorRed)
	a.yellow = pixel(colorYellow)
	a.green = pixel(colorGreen)
	a.dim = pixel(color.RGBA{70, 120, 80, 255})
	a.gray = pixel(color.RGBA{150, 150, 150, 255})
	a.shade = pixel(color.RGBA{10, 14, 10, 255})
	a.tracer = pixel(color.RGBA{255, 226, 130, 255})
	a.black = pixel(color.RGBA{8, 10, 8, 255})
	a.row = ebiten.NewImage(ScreenWidth, 1)

	// Puffs are rough discs, brighter in the middle, with a speckled edge.
	for k := range a.puff {
		img := ebiten.NewImage(puffSize, puffSize)
		for y := range puffSize {
			for x := range puffSize {
				dx, dy := float64(x)+.5-puffSize/2, float64(y)+.5-puffSize/2
				d := math.Hypot(dx, dy)
				rough := (x*7+y*13)%5 == 0
				edge, core := puffColors[k][0], puffColors[k][1]
				switch {
				case d <= 3.5:
					img.Set(x, y, core)
				case d <= 5.5 && !rough:
					img.Set(x, y, core)
				case d <= 8.6 && !(rough && d > 6.5):
					c := edge
					if dx+dy > 4 { // shaded at the bottom right
						c = color.RGBA{edge.R * 7 / 8, edge.G * 7 / 8, edge.B * 7 / 8, 255}
					}
					img.Set(x, y, c)
				}
			}
		}
		a.puff[k] = img
	}

	a.radar = ebiten.NewImage(radarSize, radarSize)
	r := float64(radarSize) / 2
	for y := range radarSize {
		for x := range radarSize {
			dx, dy := float64(x)+.5-r, float64(y)+.5-r
			d := math.Hypot(dx, dy)
			switch {
			case d > r:
			case d > r-1.5:
				a.radar.Set(x, y, color.RGBA{70, 150, 90, 255})
			case math.Abs(d-r/2) < 0.5 && (x+y)%2 == 0:
				a.radar.Set(x, y, color.RGBA{30, 70, 40, 255})
			case (x == radarSize/2 || y == radarSize/2) && (x+y)%3 == 0:
				a.radar.Set(x, y, color.RGBA{30, 70, 40, 255})
			default:
				a.radar.Set(x, y, color.RGBA{14, 34, 20, 255})
			}
		}
	}
	a.hull = ebiten.NewImage(6, 10)
	a.hull.Fill(color.RGBA{150, 200, 160, 255})
	for y := range 10 {
		// The tracks along the sides.
		a.hull.Set(0, y, color.RGBA{90, 130, 100, 255})
		a.hull.Set(5, y, color.RGBA{90, 130, 100, 255})
	}
	a.marker = ebiten.NewImage(9, 6)
	for y, row := range [6]string{"#########", "#########", ".#######.", "..#####..", "...###...", "....#...."} {
		for x := range len(row) {
			if row[x] == '#' {
				a.marker.Set(x, y, color.RGBA{255, 70, 50, 255})
			}
		}
	}
	a.sight = sightFrame()
	return a
}

// The zoomed sight is dark outside a circle with a rim just outside it. A
// masked image the size of the screen is slow to draw on the badge, so each
// band of rows is a dark rectangle at each side, plus a narrow piece in RAM
// for the curve and the rim, drawn as is on the left and flipped on the right.
const (
	rimWidth   = 3
	pieceWidth = 24 // the most the curve's x changes over a band
	sightBands = 24
)

type band struct {
	y0, y1 int // rows y0 to y1-1
	black  int // the dark rectangle's width at each side
	piece  *ebiten.Image
}

// sightFrame splits the rows into bands, top to bottom.
func sightFrame() [sightBands]band {
	const (
		cx, cy = field.SightX, field.SightY
		inner  = field.SightRadius
		outer  = inner + rimWidth
	)
	// first returns the first x of row y within radius r of the center, or the
	// center if none is.
	first := func(y int, r float64) int {
		for x := range cx {
			if math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy) <= r {
				return x
			}
		}
		return cx
	}
	pieces := ebiten.NewImage(pieceWidth, ScreenHeight)
	var bands [sightBands]band
	n := 0
	for y := 0; y < ScreenHeight; n++ {
		if n == sightBands {
			panic("tank: too many sight bands")
		}
		b := &bands[n]
		b.y0, b.black = y, first(y, outer)
		in := first(y, inner)
		for y++; y < ScreenHeight; y++ {
			lo, hi := min(b.black, first(y, outer)), max(in, first(y, inner))
			if hi-lo > pieceWidth {
				break
			}
			b.black, in = lo, hi
		}
		b.y1 = y
		for py := b.y0; py < b.y1; py++ {
			for px := range pieceWidth {
				d := math.Hypot(float64(b.black+px)+.5-cx, float64(py)+.5-cy)
				switch {
				case d > outer:
					pieces.Set(px, py, color.RGBA{8, 10, 8, 255})
				case d > inner:
					pieces.Set(px, py, color.RGBA{40, 44, 40, 255})
				}
			}
		}
		b.piece = pieces.SubImage(image.Rect(0, b.y0, in-b.black, b.y1)).(*ebiten.Image)
	}
	return bands
}
