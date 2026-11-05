// Package course defines the kart circuit: a closed centerline, the road
// width and boost pads. The asset generator paints the track image from
// it and the game uses it for laps, the road surface and the CPU drivers.
package course

import "math"

const (
	Size      = 256 // the track image and the world are Size x Size units
	HalfWidth = 10  // road half width
	CurbWidth = 2
	Wall      = 6 // karts stay this far from the world's edge
	samples   = 8 // centerline samples per control point
)

// controls is the circuit, counterclockwise on screen. Catmull-Rom splines join them.
var controls = [...][2]float64{
	{60, 222}, {150, 224}, {205, 212}, {228, 172}, {220, 122}, {186, 98},
	{176, 62}, {196, 34}, {150, 26}, {90, 30}, {40, 42}, {28, 82},
	{55, 112}, {86, 132}, {70, 166}, {34, 186},
}

type Point struct{ X, Y float32 }

// A Pad boosts a kart that drives over it. S is the distance along the
// centerline and Offset is to the right of it.
type Pad struct{ S, Offset float32 }

const PadLength, PadHalfWidth = 8, 4

var (
	Path   []Point   // centerline samples, closed: the last joins the first
	Arc    []float32 // distance along the centerline at each sample
	Length float32
	Pads   []Pad
)

func init() {
	n := len(controls)
	for i := range n {
		p0, p1, p2, p3 := controls[(i+n-1)%n], controls[i], controls[(i+1)%n], controls[(i+2)%n]
		for k := range samples {
			t := float64(k) / samples
			t2, t3 := t*t, t*t*t
			at := func(a, b, c, d float64) float32 {
				return float32(0.5 * (2*b + (c-a)*t + (2*a-5*b+4*c-d)*t2 + (3*b-a-3*c+d)*t3))
			}
			Path = append(Path, Point{at(p0[0], p1[0], p2[0], p3[0]), at(p0[1], p1[1], p2[1], p3[1])})
		}
	}
	Arc = make([]float32, len(Path))
	for i := 1; i < len(Path); i++ {
		Arc[i] = Arc[i-1] + dist(Path[i-1], Path[i])
	}
	Length = Arc[len(Path)-1] + dist(Path[len(Path)-1], Path[0])
	l := Length
	Pads = []Pad{{0.22 * l, 0}, {0.47 * l, -5}, {0.70 * l, 5}, {0.90 * l, 0}}
}

func dist(a, b Point) float32 {
	return float32(math.Hypot(float64(b.X-a.X), float64(b.Y-a.Y)))
}

// Where is a position relative to the centerline.
type Where struct {
	Index  int     // the centerline segment from Path[Index] to the next sample
	S      float32 // distance along the centerline
	Offset float32 // signed distance to the right of the centerline
	Dist   float32 // distance to the centerline
}

// Locate finds the nearest centerline segment within window segments of hint,
// or among all of them when window is negative.
func Locate(x, y float32, hint, window int) Where {
	n := len(Path)
	first, count := 0, n
	if window >= 0 && window*2+1 < n {
		first, count = hint-window, window*2+1
	}
	best := Where{Dist: math.MaxFloat32}
	bestSq := float32(math.MaxFloat32)
	for k := range count {
		i := ((first+k)%n + n) % n
		a, b := Path[i], Path[(i+1)%n]
		dx, dy := b.X-a.X, b.Y-a.Y
		lengthSq := dx*dx + dy*dy
		t := ((x-a.X)*dx + (y-a.Y)*dy) / lengthSq
		t = min(max(t, 0), 1)
		px, py := a.X+dx*t-x, a.Y+dy*t-y
		if d := px*px + py*py; d < bestSq {
			bestSq = d
			length := float32(math.Sqrt(float64(lengthSq)))
			// Right of the direction of travel, with y pointing down.
			cross := dx*(y-a.Y) - dy*(x-a.X)
			best = Where{Index: i, S: Arc[i] + t*length, Offset: cross / length}
		}
	}
	best.Dist = float32(math.Sqrt(float64(bestSq)))
	return best
}

// At returns the centerline point at distance s and the direction of travel.
func At(s float32) (p, dir Point) {
	s = float32(math.Mod(float64(s), float64(Length)))
	if s < 0 {
		s += Length
	}
	n := len(Path)
	i := 0
	for i+1 < n && Arc[i+1] <= s {
		i++
	}
	a, b := Path[i], Path[(i+1)%n]
	length := dist(a, b)
	t := (s - Arc[i]) / length
	dir = Point{(b.X - a.X) / length, (b.Y - a.Y) / length}
	return Point{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t}, dir
}

// Place returns the world position offset to the right of the centerline.
func Place(s, offset float32) Point {
	p, d := At(s)
	return Point{p.X - d.Y*offset, p.Y + d.X*offset}
}
