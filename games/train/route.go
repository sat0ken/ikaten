package train

// The line, in meters from the stop mark of the first station. The train's
// front is at pos; the cab and the stop marks refer to the front.

// station holds a stop mark and the scheduled arrival in seconds after the
// first departure.
type station struct {
	name   string
	stop   float32
	arrive int
}

var stations = [...]station{
	{"IKAHAMA", 0, 0},
	{"SUNAMACHI", 1300, 100},
	{"MINATO", 2700, 255},
	{"TAKADAI", 4300, 390},
	{"KAIGAN", 5500, 505},
	{"SAKIYAMA", terminal, 660},
}

const (
	trainLength  = 80  // a raised limit applies once the whole train is past its start
	dwellSeconds = 20  // scheduled stop at each station
	stopWindow   = 40  // the train must stop within this many meters of the mark to open its doors
	platformBack = 100 // the platform reaches this far behind the stop mark
	platformFore = 20  // and this far beyond it
	canopyBack   = 70
	canopyFore   = 10
	bufferStop   = 35 // the terminal's buffer stop, beyond its stop mark
	startSeconds = 10 * 3600
)

// The terminal's stop mark, and where the rails end. Constants keep the
// package's initialization simple enough for TinyGo to run at compile time.
const (
	terminal = 6900
	lineEnd  = terminal + bufferStop + 2
)

// limit is a speed limit in km/h that starts at from.
type limit struct {
	from float32
	kmh  int
}

var limits = [...]limit{
	{-200, 75},
	{1400, 90},
	{2100, 60},
	{2450, 90},
	{2800, 80},
	{4400, 85},
	{5000, 65},
	{5600, 95},
	{6450, 45},
}

// curve bends the line between from and to; a positive radius turns right.
type curve struct {
	from, to, radius float32
}

var curves = [...]curve{
	{600, 900, 800},
	{2150, 2400, -450},
	{3300, 3600, 700},
	{4600, 4800, -900},
	{6150, 6400, -600},
}

// grade is the gradient in permille from a point on; positive climbs.
type grade struct {
	from, permil float32
}

var grades = [...]grade{
	{-200, 0},
	{3000, 15},
	{3900, 0},
	{4700, -15},
	{5400, 0},
}

type scenery uint8

const (
	field scenery = iota
	town
	coast // the sea on the left
	river // a bridge over water
)

type zone struct {
	from float32
	kind scenery
}

var zones = [...]zone{
	{-400, town},
	{500, field},
	{1000, town},
	{1600, coast},
	{3500, field},
	{4150, town},
	{4600, field},
	{5050, town},
	{5700, field},
	{5900, river},
	{6100, field},
	{6500, town},
}

// Level crossings, each with a whistle post before it. The driver sounds the
// horn after passing the post.
var crossings = [...]float32{1950, 4950}

const whistleAhead = 150 // the post stands this far before the crossing
const whistleWindow = 120

func limitAt(s float32) int {
	kmh := limits[0].kmh
	for _, l := range limits {
		if l.from > s {
			break
		}
		kmh = l.kmh
	}
	return kmh
}

func curvatureAt(s float32) float32 {
	for _, c := range curves {
		if s >= c.from && s < c.to {
			return 1 / c.radius
		}
	}
	return 0
}

func gradeAt(s float32) float32 {
	p := grades[0].permil
	for _, g := range grades {
		if g.from > s {
			break
		}
		p = g.permil
	}
	return p
}

func sceneryAt(s float32) scenery {
	k := zones[0].kind
	for _, z := range zones {
		if z.from > s {
			break
		}
		k = z.kind
	}
	return k
}

// onCrossing reports whether s is on a level crossing's road.
func onCrossing(s float32) bool {
	for _, c := range crossings {
		if s > c-3 && s < c+3 {
			return true
		}
	}
	return false
}

// atStation reports whether s is beside a platform.
func atStation(s float32) bool {
	for _, st := range stations {
		if s > st.stop-platformBack-30 && s < st.stop+platformFore+30 {
			return true
		}
	}
	return false
}
