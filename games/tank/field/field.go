// Package field is the battlefield of the tank game: the obstacles that block
// tanks and shells, the roads, and the layout of the sprite sheet. The asset
// generator paints the ground and the sprites from it, and the game uses it for
// collisions and drawing.
//
// The field is flat, in meters, with x to the east and y to the south, and
// centered on the origin. Headings are radians clockwise from north.
package field

import "math"

const (
	Size        = 1024 // meters on each side
	Half        = Size / 2
	TexelMeters = 4 // the ground texture covers 4 m with each pixel
	TextureSize = Size / TexelMeters
	CellMeters  = 64 // the obstacle grid
	Cells       = Size / CellMeters

	// The player starts in the south facing north; enemies come from the north.
	StartX, StartY = 0, 400
	EnemyMinY      = -440
	EnemyMaxY      = -300
)

type Kind uint8

const (
	Tree Kind = iota
	Pine
	House
)

// An Obstacle blocks tanks within Radius and shells below Height.
type Obstacle struct {
	X, Y, Radius, Height float32
	Heading              float32 // houses only
	Kind                 Kind
}

type Point struct{ X, Y float32 }

// Roads are dirt tracks, painted on the ground only.
var Roads = [][]Point{
	{{0, 512}, {-20, 300}, {-90, 120}, {-190, -20}, {-150, -250}, {-80, 512 - Size}},
	{{-512, 150}, {-300, 60}, {-190, -20}, {20, 40}, {230, 100}, {512, 70}},
	{{230, 100}, {180, -120}, {60, -260}, {40, -512}},
}

var (
	// Obstacles are sorted by grid cell; the ones in cell c are
	// Obstacles[CellStart[c]:CellStart[c+1]].
	Obstacles []Obstacle
	CellStart [Cells*Cells + 1]uint16
)

// CellOf returns the grid cell of a point, clamped to the field.
func CellOf(x, y float32) (cx, cy int) {
	cx = min(max(int((x+Half)/CellMeters), 0), Cells-1)
	cy = min(max(int((y+Half)/CellMeters), 0), Cells-1)
	return cx, cy
}

// InCell returns the obstacles in a cell; an index outside the grid has none.
func InCell(cx, cy int) []Obstacle {
	if cx < 0 || cy < 0 || cx >= Cells || cy >= Cells {
		return nil
	}
	c := cy*Cells + cx
	return Obstacles[CellStart[c]:CellStart[c+1]]
}

type builder struct {
	rng  uint32
	list []Obstacle
}

func (b *builder) random() float32 {
	b.rng ^= b.rng << 13
	b.rng ^= b.rng >> 17
	b.rng ^= b.rng << 5
	return float32(b.rng%10000) / 10000
}

// free reports whether a tree or house fits at x, y: inside the field, apart
// from other obstacles, and off the start areas.
func (b *builder) free(x, y, spacing float32) bool {
	if x < -Half+20 || x > Half-20 || y < -Half+20 || y > Half-20 {
		return false
	}
	if dx, dy := x-StartX, y-StartY; dx*dx+dy*dy < 90*90 {
		return false
	}
	if y > EnemyMinY-20 && y < EnemyMaxY+20 && x > -420 && x < 420 && math.Mod(float64(x+Half), 130) < 50 {
		return false // lanes for the enemies to start in
	}
	for _, o := range b.list {
		dx, dy := x-o.X, y-o.Y
		if dx*dx+dy*dy < spacing*spacing {
			return false
		}
	}
	return true
}

func (b *builder) tree(x, y, spacing float32) {
	if !b.free(x, y, spacing) {
		return
	}
	if b.random() < 0.4 {
		b.list = append(b.list, Obstacle{X: x, Y: y, Radius: 1.8, Height: 10, Kind: Pine})
	} else {
		b.list = append(b.list, Obstacle{X: x, Y: y, Radius: 2.2, Height: 9, Kind: Tree})
	}
}

func (b *builder) village(x, y float32, houses int) {
	for i := range houses {
		a := float32(i)*2.4 + b.random()
		r := 18 + 22*b.random()
		hx := x + r*float32(math.Sin(float64(a)))
		hy := y - r*float32(math.Cos(float64(a)))
		if b.free(hx, hy, 16) {
			b.list = append(b.list, Obstacle{X: hx, Y: hy, Radius: 5.5, Height: 7, Heading: b.random() * math.Pi, Kind: House})
		}
	}
}

func (b *builder) line(x0, y0, x1, y1, step float32) {
	n := int(float32(math.Hypot(float64(x1-x0), float64(y1-y0))) / step)
	for i := range n + 1 {
		t := float32(i) / float32(n)
		b.tree(x0+(x1-x0)*t+(b.random()-.5)*4, y0+(y1-y0)*t+(b.random()-.5)*4, step-3)
	}
}

// wood scatters trees in a circle, close enough that tanks cannot pass.
func (b *builder) wood(x, y, r float32, n int) {
	for range n * 3 {
		a, d := b.random()*2*math.Pi, r*float32(math.Sqrt(float64(b.random())))
		b.tree(x+d*float32(math.Sin(float64(a))), y-d*float32(math.Cos(float64(a))), 9)
	}
}

func build() []Obstacle {
	b := &builder{rng: 2463534242}
	b.village(-190, -20, 6)
	b.village(230, 100, 5)
	b.village(60, -250, 3)
	b.village(-260, 330, 3)
	b.line(-460, 210, -130, 250, 13)
	b.line(110, -110, 460, -60, 13)
	b.line(-420, -160, -250, -210, 13)
	b.line(90, 260, 330, 230, 13)
	b.wood(330, -380, 55, 16)
	b.wood(-380, -360, 50, 14)
	b.wood(380, 360, 50, 14)
	b.wood(10, 60, 28, 8)
	b.wood(-60, -150, 24, 6)
	for range 40 {
		b.tree(b.random()*(Size-60)-Half+30, b.random()*(Size-60)-Half+30, 30)
	}
	// Sort by grid cell: count each cell, then place.
	cells := make([]uint16, len(b.list))
	for i, o := range b.list {
		cx, cy := CellOf(o.X, o.Y)
		cells[i] = uint16(cy*Cells + cx)
		CellStart[cells[i]+1]++
	}
	for c := range Cells * Cells {
		CellStart[c+1] += CellStart[c]
	}
	var next [Cells * Cells]uint16
	copy(next[:], CellStart[:Cells*Cells])
	list := make([]Obstacle, len(b.list))
	for i, o := range b.list {
		list[next[cells[i]]] = o
		next[cells[i]]++
	}
	return list
}

func init() { Obstacles = build() }

// The sprite sheet: tank hulls and turrets from 16 directions for each of the
// three enemy kinds, houses from 8 (they look the same turned half around),
// two trees and the distant hills. Sprites have PixelsPerMeter, and an anchor:
// the pixel over the object's center on the ground.
const (
	PixelsPerMeter = 4
	TankKinds      = 3
	TankFrames     = 16
	TankW, TankH   = 40, 18
	TankAX, TankAY = 20, 14
	HouseFrames    = 8
	HouseW, HouseH = 52, 36
	HouseAX        = 26
	HouseAY        = 32
	HouseY         = TankKinds * 2 * TankH
	TreeW, TreeH   = 20, 40
	TreeAX, TreeAY = 10, 39
	TreeX          = HouseFrames * HouseW
	HillsW, HillsH = 256, 16
	HillsY         = HouseY + TreeH
	SheetW         = TankFrames * TankW
	SheetH         = HillsY + HillsH
)

// TankCell returns the sheet position of a tank frame; part 0 is the hull and
// 1 the turret.
func TankCell(kind, part, frame int) (x, y int) {
	return frame * TankW, (kind*2 + part) * TankH
}

// The zoomed gunsight is a circle around the gun's axis.
const (
	SightW, SightH = 120, 120
	SightX, SightY = 60, 54
	SightRadius    = 51
)
