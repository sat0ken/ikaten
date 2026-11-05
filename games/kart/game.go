// Package kart is an original kart racer drawn like early 16-bit racers: the
// track image is rotated and scaled row by row into a perspective floor. It
// uses only the Ebitengine API, and the track and sprites are drawn by code, so
// it also builds with Ebitengine.
package kart

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/kart/course"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	horizon      = 34 // the first ground row
	focal        = 60 // pixels; a 90 degree view
	camHeight    = 16 // world units
	camDist      = 13 // the camera follows this far behind the player
	farthest     = 320
	laps         = 3
	karts        = 4
	countdown    = 180 // ticks before the start

	maxSpeed   = 0.8 // world units per tick
	grassSpeed = 0.35
	boostSpeed = 1.25
	accel      = 0.012
	braking    = 0.03
	turnRate   = 0.045
	driftTurn  = 0.065
	grip       = 0.35 // share of the gap between motion and heading closed per tick
	driftGrip  = 0.08
	kartRadius = 1.6
	smallBoost = 18 // ticks after a drift of driftShort ticks
	bigBoost   = 35 // ticks after a drift of driftLong ticks
	driftShort = 35
	driftLong  = 70
	padBoost   = 40
)

type input struct {
	gas, brake, drift bool
	steer             int // -1 left, 1 right
}

type kart struct {
	x, y           float32
	heading, slide float32 // where it points and where it moves
	speed          float32
	steer          int // for the sprite
	drift          int // -1 or 1 while drifting
	charge         int // drift ticks
	boost          int // ticks left
	seg            int // nearest centerline segment
	s, offset      float32
	onRoad         bool
	lap            int // -1 until the first crossing of the line
	finished       bool
	finishTicks    int
	lane, laneGoal float32 // CPU: offset from the centerline
	skill          float32 // CPU: share of the top speed
}

func (k *kart) progress() float32 { return float32(k.lap)*course.Length + k.s }

type mode int

const (
	modeTitle mode = iota // CPU drivers race on their own
	modeCountdown
	modeRace
	modeResult
)

type Game struct {
	mode             mode
	ticks, modeTicks int
	raceTicks        int
	lapStart         int
	bestLap, lastLap int
	finishers        int
	karts            [karts]kart // karts[0] is the player
	order            [karts]int  // kart indices by position
	player           input
	// The camera: position, angle and its forward and right vectors.
	camX, camY, camAngle float32
	fx, fy, rx, ry       float32
	rows                 [ScreenHeight]*ebiten.Image // one-row subimages of the screen
	rowsOf               *ebiten.Image
	sprites              [karts]drawItem
}

func NewGame() *Game {
	g := &Game{}
	g.reset()
	return g
}

// grid holds starting places behind the line: distance back and offset.
var grid = [karts][2]float32{{26, 4}, {6, -4}, {12, 4}, {19, -4}}

var (
	skills = [karts]float32{1, 0.93, 0.96, 0.99}
	lanes  = [karts]float32{0, -4, 3, 0}
)

func (g *Game) reset() {
	for i := range g.karts {
		_, d := course.At(course.Length - grid[i][0])
		pos := course.Place(course.Length-grid[i][0], grid[i][1])
		heading := float32(math.Atan2(float64(d.Y), float64(d.X)))
		g.karts[i] = kart{x: pos.X, y: pos.Y, heading: heading, slide: heading, lap: -1,
			lane: lanes[i], laneGoal: lanes[i], skill: skills[i]}
		g.locate(&g.karts[i], -1)
		g.order[i] = i
	}
	g.raceTicks, g.lapStart, g.bestLap, g.lastLap, g.finishers = 0, 0, 0, 0, 0
	g.rank()
	g.camAngle = g.karts[0].heading
	g.follow(1)
}

func (g *Game) setMode(m mode) {
	g.mode, g.modeTicks = m, 0
}

func startPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	switch g.mode {
	case modeTitle:
		g.step(true)
		if startPressed() {
			g.reset()
			g.setMode(modeCountdown)
		}
	case modeCountdown:
		if g.modeTicks >= countdown {
			g.setMode(modeRace)
		}
	case modeRace, modeResult:
		g.raceTicks++
		g.step(g.karts[0].finished)
		if g.karts[0].finished && g.mode == modeRace {
			g.setMode(modeResult)
		}
		if g.mode == modeResult && g.modeTicks > 90 && startPressed() {
			g.reset()
			g.setMode(modeTitle)
		}
	}
	g.follow(0.25)
	return nil
}

func (g *Game) readPlayer() input {
	in := input{
		gas:   ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeySpace),
		brake: ebiten.IsKeyPressed(ebiten.KeyArrowDown),
		drift: ebiten.IsKeyPressed(ebiten.KeyX),
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		in.steer--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		in.steer++
	}
	return in
}

// step moves every kart one tick. With auto, the CPU also drives the player.
func (g *Game) step(auto bool) {
	for i := range g.karts {
		var in input
		if i == 0 && !auto {
			in = g.readPlayer()
			g.player = in
		} else {
			in = g.drive(i)
		}
		g.move(i, in)
	}
	g.collideKarts()
	g.rank()
}

func angleDiff(a, b float32) float32 {
	d := a - b
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

func sincos(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

// wrapped returns b-a along the loop, between -Length/2 and Length/2.
func wrapped(a, b float32) float32 {
	d := b - a
	if d > course.Length/2 {
		d -= course.Length
	} else if d < -course.Length/2 {
		d += course.Length
	}
	return d
}

// drive is the CPU driver: it aims at a point ahead on its lane, moves its lane
// onto boost pads, and lifts in sharp turns.
func (g *Game) drive(i int) input {
	k := &g.karts[i]
	k.laneGoal = lanes[i]
	for _, p := range course.Pads {
		if d := wrapped(k.s, p.S); d > 0 && d < 30 {
			k.laneGoal = p.Offset
		}
	}
	k.lane += min(max(k.laneGoal-k.lane, -0.15), 0.15)
	target := course.Place(k.s+12+k.speed*8, k.lane)
	want := float32(math.Atan2(float64(target.Y-k.y), float64(target.X-k.x)))
	diff := angleDiff(want, k.heading)
	in := input{gas: abs(diff) < 0.7 || k.speed < 0.3} // steering needs some speed
	if diff > 0.03 {
		in.steer = 1
	} else if diff < -0.03 {
		in.steer = -1
	}
	return in
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// topSpeed is a kart's limit. CPU drivers close or open the gap to the player
// a little so that the race stays close.
func (g *Game) topSpeed(i int) float32 {
	k := &g.karts[i]
	if k.boost > 0 {
		return boostSpeed
	}
	if !k.onRoad {
		return grassSpeed
	}
	top := maxSpeed * k.skill
	if i != 0 && g.mode != modeTitle {
		gap := (g.karts[0].progress() - k.progress()) / 400
		top *= 1 + min(max(gap, -0.08), 0.08)
	}
	return top
}

func (g *Game) move(i int, in input) {
	k := &g.karts[i]
	k.steer = in.steer
	if in.gas {
		k.speed += accel
	} else {
		k.speed *= 0.985
	}
	if in.brake {
		k.speed -= braking
	}
	top := g.topSpeed(i)
	if k.boost > 0 {
		k.boost--
		k.speed = max(k.speed, top*0.9)
	}
	if k.speed > top {
		k.speed = max(top, k.speed-0.02)
	}
	k.speed = max(k.speed, 0)

	turn := float32(in.steer) * turnRate * min(1, k.speed/0.3)
	if in.drift && in.steer != 0 && k.drift == 0 && k.speed > 0.5 {
		k.drift = in.steer
	}
	if k.drift != 0 {
		if !in.drift || k.speed < 0.3 {
			// Releasing a long drift gives a short boost.
			switch {
			case k.charge >= driftLong:
				k.boost = bigBoost
			case k.charge >= driftShort:
				k.boost = smallBoost
			}
			k.drift, k.charge = 0, 0
		} else {
			k.charge++
			turn = float32(k.drift)*driftTurn*0.6 + float32(in.steer)*driftTurn*0.5
		}
	}
	k.heading = angleDiff(k.heading+turn, 0)
	gripNow := float32(grip)
	if k.drift != 0 {
		gripNow = driftGrip
	}
	k.slide = angleDiff(k.slide+angleDiff(k.heading, k.slide)*gripNow, 0)
	s, c := sincos(k.slide)
	k.x += c * k.speed
	k.y += s * k.speed
	if lo, hi := float32(course.Wall), float32(course.Size-course.Wall); k.x < lo || k.x > hi || k.y < lo || k.y > hi {
		k.x, k.y = min(max(k.x, lo), hi), min(max(k.y, lo), hi)
		k.speed *= 0.5
	}
	g.locate(k, 6)
	for _, p := range course.Pads {
		if abs(wrapped(p.S, k.s)) < course.PadLength/2 && abs(k.offset-p.Offset) < course.PadHalfWidth {
			k.boost = max(k.boost, padBoost)
		}
	}
}

// locate updates the kart's place on the course, and counts laps when it
// crosses the line between the last and the first segment.
func (g *Game) locate(k *kart, window int) {
	w := course.Locate(k.x, k.y, k.seg, window)
	if w.Dist > course.HalfWidth*3 && window >= 0 {
		w = course.Locate(k.x, k.y, 0, -1) // lost after a long excursion
	}
	n := len(course.Path)
	if window >= 0 {
		if k.seg > n*3/4 && w.Index < n/4 {
			g.crossed(k)
		} else if k.seg < n/4 && w.Index > n*3/4 {
			k.lap--
		}
	}
	k.seg, k.s, k.offset = w.Index, w.S, w.Offset
	k.onRoad = w.Dist < course.HalfWidth+course.CurbWidth
}

func (g *Game) crossed(k *kart) {
	k.lap++
	if g.mode == modeTitle || k.finished {
		return
	}
	if k == &g.karts[0] && k.lap >= 1 {
		g.lastLap = g.raceTicks - g.lapStart
		if g.bestLap == 0 || g.lastLap < g.bestLap {
			g.bestLap = g.lastLap
		}
	}
	if k == &g.karts[0] {
		g.lapStart = g.raceTicks
	}
	if k.lap >= laps {
		k.finished = true
		k.finishTicks = g.raceTicks
		g.finishers++
	}
}

func (g *Game) collideKarts() {
	const reach = 2 * kartRadius
	for i := range g.karts {
		for j := i + 1; j < len(g.karts); j++ {
			a, b := &g.karts[i], &g.karts[j]
			dx, dy := b.x-a.x, b.y-a.y
			d2 := dx*dx + dy*dy
			if d2 >= reach*reach || d2 == 0 {
				continue
			}
			d := float32(math.Sqrt(float64(d2)))
			push := (reach - d) / d / 2
			a.x, a.y = a.x-dx*push, a.y-dy*push
			b.x, b.y = b.x+dx*push, b.y+dy*push
			a.speed *= 0.97
			b.speed *= 0.97
		}
	}
}

// rank orders the karts: finishers by time, then by distance covered.
func (g *Game) rank() {
	ahead := func(a, b int) bool {
		ka, kb := &g.karts[a], &g.karts[b]
		if ka.finished != kb.finished {
			return ka.finished
		}
		if ka.finished {
			return ka.finishTicks < kb.finishTicks
		}
		return ka.progress() > kb.progress()
	}
	for i := 1; i < len(g.order); i++ {
		for j := i; j > 0 && ahead(g.order[j], g.order[j-1]); j-- {
			g.order[j], g.order[j-1] = g.order[j-1], g.order[j]
		}
	}
}

func (g *Game) position(i int) int {
	for p, k := range g.order {
		if k == i {
			return p + 1
		}
	}
	return len(g.order)
}

// follow turns the camera toward the player's heading by rate and places it
// behind the player.
func (g *Game) follow(rate float32) {
	k := &g.karts[0]
	g.camAngle = angleDiff(g.camAngle+angleDiff(k.heading, g.camAngle)*rate, 0)
	s, c := sincos(g.camAngle)
	g.fx, g.fy, g.rx, g.ry = c, s, -s, c
	g.camX, g.camY = k.x-c*camDist, k.y-s*camDist
}
