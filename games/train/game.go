// Package train is an original train driving game seen from the cab: drive a
// local train along a coastal line and stop at each station's mark on time.
// One handle sets power and brakes, and both take a moment to build up, so
// the driver plans the braking ahead. Speed limits, gradients, curves and
// level crossings come along the way. It uses only the Ebitengine API, and
// every image is drawn by code, so it also builds with Ebitengine.
//
// Distances are meters along the line, speeds are meters per second inside
// and km/h on screen, and rates are in km/h per second, as drivers count them.
package train

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	tps          = 60

	// The handle: EB, B8 to B1, N, P1 to P5.
	notchEB    = -9
	maxBrake   = 8
	maxPower   = 5
	repeatWait = 18 // ticks a held direction waits before it repeats
	repeatRate = 6

	powerAccel   = 3.0  // km/h/s at P5 below constantFrom
	constantFrom = 35   // km/h; above this the motors give constant power
	brakeStep    = 0.45 // km/h/s per brake notch
	emergency    = 4.5
	brakeRise    = 4.0 // km/h/s per second: how fast the brakes build up and release
	powerRise    = 3.0
	powerFall    = 8.0
	gravity      = 9.8 * 3.6 / 1000 // km/h/s per permille

	atsMargin     = 10  // km/h over the limit that trips the automatic brake
	hardStop      = 2.7 // km/h/s of braking at the moment of stopping that jolts the passengers
	hornTicks     = 40
	messageLength = 120
)

type mode int

const (
	modeTitle   mode = iota
	modePlay         // between stations
	modeStation      // stopped at a platform with the doors open
	modeEnd          // after the terminal
)

const (
	msgNone = iota
	msgDepart
	msgOverSpeed
	msgATS
	msgOverrun
	msgBuffer
	msgWhistle
	msgShort
)

var messages = [...]string{"", "DEPART", "OVER SPEED", "ATS BRAKE", "OVERRUN", "BUFFER STOP", "SOUND HORN", "STOP SHORT"}

// result is the score of the run to one station.
type result struct {
	stopError float32 // meters past the mark; negative is short
	delay     int     // seconds late; negative is early
	stopPts   int
	timePts   int
	ridePts   int
	speedPts  int // a penalty, zero or negative
	hornPts   int
	total     int
	overrun   bool
}

type Game struct {
	mode             mode
	ticks, modeTicks int

	// The train.
	pos, v       float32 // front position and speed in m/s
	notch        int
	power, brake float32 // the force built up so far, in km/h/s
	heading      float32 // radians turned since the start, for the distant hills
	held         int     // ticks the handle direction has been held
	ats          bool    // the automatic brake holds the train until it stops

	clock int // ticks since the first departure
	next  int // the station ahead
	horn  int // ticks left of the horn

	// The run to the next station.
	overTicks  int
	ebUsed     int
	atsTripped bool
	whistles   int // whistle posts passed
	sounded    int // of those, the ones answered with the horn
	whistleDue float32
	overrun    bool

	results      [len(stations)]result
	message      int
	messageTicks int

	// The autopilot, for the title screen, the benchmark and tests.
	apBraking bool
	wait      int // ticks the title's train waits at a station

	view view
}

func NewGame() *Game {
	g := &Game{}
	g.reset()
	return g
}

// fresh is a new game to copy from: a composite literal of this size would be
// built on the heap each time.
var fresh Game

func (g *Game) reset() {
	m, t := g.mode, g.ticks
	*g = fresh
	g.mode, g.ticks = m, t
	g.next = 1
	g.notch = -maxBrake
	g.brake = brakeStep * maxBrake
	g.startRun()
}

func (g *Game) startRun() {
	g.overTicks, g.ebUsed, g.atsTripped = 0, 0, false
	g.whistles, g.sounded, g.whistleDue, g.overrun = 0, 0, 0, false
	g.apBraking = false
}

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

func (g *Game) say(m int) { g.message, g.messageTicks = m, messageLength }

func startPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	if g.messageTicks > 0 {
		g.messageTicks--
	}
	switch g.mode {
	case modeTitle:
		// The title runs the autopilot, waits at each station, and starts
		// over at the end of the line.
		if g.wait > 0 {
			g.wait--
			g.drive(-maxBrake / 2)
		} else {
			g.drive(g.autopilot())
		}
		if g.v == 0 && g.wait == 0 && g.pos > stations[g.next].stop-stopWindow {
			g.next++
			g.startRun()
			g.wait = 3 * tps
			if g.next == len(stations) {
				g.reset()
			}
		}
		if startPressed() {
			g.reset()
			g.setMode(modePlay)
			g.say(msgDepart)
		}
	case modePlay:
		g.handleInput()
		g.step()
	case modeStation:
		if g.modeTicks > 30 && startPressed() {
			g.depart()
		}
	case modeEnd:
		if g.modeTicks > 60 && startPressed() {
			g.reset()
			g.setMode(modeTitle)
		}
	}
	return nil
}

// handleInput moves the handle: up toward power, down toward the brakes. A
// held direction repeats. A sounds the horn and B returns the handle to N.
func (g *Game) handleInput() {
	up, down := ebiten.IsKeyPressed(ebiten.KeyArrowUp), ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	if up != down {
		g.held++
	} else {
		g.held = 0
	}
	if g.held == 1 || g.held > repeatWait && (g.held-repeatWait)%repeatRate == 0 {
		if up {
			g.setNotch(g.notch + 1)
		} else {
			g.setNotch(g.notch - 1)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.setNotch(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.soundHorn()
	}
}

func (g *Game) setNotch(n int) {
	if g.ats {
		return
	}
	n = max(notchEB, min(maxPower, n))
	if n == notchEB && g.notch != notchEB && g.v > 0 {
		g.ebUsed++
	}
	g.notch = n
}

func (g *Game) soundHorn() {
	g.horn = hornTicks
	if g.whistleDue > 0 {
		g.sounded++
		g.whistleDue = 0
	}
}

// targets returns the power and brake rates the handle asks for.
func (g *Game) targets() (power, brake float32) {
	switch {
	case g.notch == notchEB:
		return 0, emergency
	case g.notch < 0:
		return 0, brakeStep * float32(-g.notch)
	}
	kmh := g.v * 3.6
	p := powerAccel * float32(g.notch) / maxPower
	if kmh > constantFrom {
		p *= constantFrom / kmh
	}
	return p, 0
}

func approach(x, target, rate float32) float32 {
	if x < target {
		return min(x+rate, target)
	}
	return max(x-rate, target)
}

// resistance is the running resistance in km/h/s.
func resistance(kmh float32) float32 {
	if kmh <= 0 {
		return 0
	}
	return 0.05 + 0.00004*kmh*kmh
}

// physics advances the train one tick.
func (g *Game) physics() {
	pt, bt := g.targets()
	if pt > g.power {
		g.power = approach(g.power, pt, powerRise/tps)
	} else {
		g.power = approach(g.power, pt, powerFall/tps)
	}
	g.brake = approach(g.brake, bt, brakeRise/tps)
	kmh := g.v * 3.6
	a := g.power - resistance(kmh) - gravity*gradeAt(g.pos-trainLength/2)
	if g.v > 0 || a > g.brake {
		a -= g.brake
	} else {
		a = 0 // brakes hold a standing train; they do not push it backward
	}
	v := g.v + a/3.6/tps
	if v < 0 {
		v = 0
	}
	ds := (g.v + v) / 2 / tps
	g.v = v
	g.heading += curvatureAt(g.pos) * ds
	g.pos += ds
	if g.horn > 0 {
		g.horn--
	}
}

// drive runs the train with the handle at notch, without scoring.
func (g *Game) drive(notch int) {
	g.notch = notch
	g.physics()
	g.clock++
}

// limitNow is the limit for the whole train: a lower limit applies when the
// front reaches it and a higher one when the rear has passed.
func (g *Game) limitNow() int {
	return min(limitAt(g.pos), limitAt(g.pos-trainLength))
}

// step advances play one tick: the train, the limits and the stations.
func (g *Game) step() {
	moving := g.v > 0
	before := g.pos
	g.physics()
	g.clock++
	kmh := g.v * 3.6
	lim := float32(g.limitNow())
	if kmh > lim+0.5 {
		g.overTicks++
		if g.messageTicks == 0 {
			g.say(msgOverSpeed)
		}
	}
	if kmh > lim+atsMargin && !g.ats {
		g.ats, g.atsTripped = true, true
		g.notch = notchEB
		g.say(msgATS)
	}
	if g.ats && g.v == 0 {
		g.ats = false
	}
	for _, c := range crossings {
		post := c - whistleAhead
		if before < post && g.pos >= post {
			g.whistles++
			g.whistleDue = post + whistleWindow
			g.say(msgWhistle)
		}
	}
	if g.whistleDue > 0 && g.pos > g.whistleDue {
		g.whistleDue = 0
	}
	st := stations[g.next]
	if g.next == len(stations)-1 && g.pos >= st.stop+bufferStop {
		// The buffer stop ends the line.
		g.pos, g.v, g.power, g.brake = st.stop+bufferStop, 0, 0, emergency
		g.overrun = true
		g.say(msgBuffer)
		g.arrive()
		return
	}
	if !g.overrun && g.pos > st.stop+stopWindow {
		g.overrun = true
		g.say(msgOverrun)
	}
	if moving && g.v == 0 {
		switch {
		case g.pos >= st.stop-stopWindow:
			g.arrive()
		case g.pos >= st.stop-platformBack-stopWindow:
			g.say(msgShort)
		}
	}
}

// arrive scores the run, opens the doors and waits for the driver.
func (g *Game) arrive() {
	st := stations[g.next]
	var r result
	r.stopError = g.pos - st.stop
	r.overrun = g.overrun
	r.delay = g.clock/tps - st.arrive
	r.stopPts = stopPoints(r.stopError)
	if g.overrun {
		r.stopPts = 0
	}
	r.timePts = timePoints(r.delay)
	r.ridePts = 20
	if g.brake >= hardStop {
		r.ridePts -= 10
	}
	r.ridePts = max(0, r.ridePts-10*g.ebUsed)
	r.speedPts = -g.overTicks / (tps / 2)
	if g.atsTripped {
		r.speedPts -= 20
	}
	if g.overrun {
		r.speedPts -= 20
	}
	r.hornPts = 5*g.sounded - 5*(g.whistles-g.sounded)
	r.total = max(0, r.stopPts+r.timePts+r.ridePts+r.speedPts+r.hornPts)
	g.results[g.next] = r
	g.ats = false
	g.setMode(modeStation)
	g.messageTicks = 0
}

// depart closes the doors. The clock moves on by the time the passengers took,
// or to the scheduled departure if that is later.
func (g *Game) depart() {
	st := stations[g.next]
	if g.next == len(stations)-1 {
		g.setMode(modeEnd)
		return
	}
	g.clock = max(g.clock+dwellSeconds*tps, (st.arrive+dwellSeconds)*tps)
	g.next++
	g.startRun()
	g.setMode(modePlay)
	g.say(msgDepart)
}

// stopPoints gives 50 within 25 cm of the mark and 8 fewer per meter beyond.
func stopPoints(err float32) int {
	err = max(err, -err)
	if err <= 0.25 {
		return 50
	}
	return max(0, 50-int(err*8+0.5))
}

// timePoints gives 30 within a second of the timetable and 3 fewer per second beyond.
func timePoints(delay int) int {
	delay = max(delay, -delay)
	if delay <= 1 {
		return 30
	}
	return max(0, 30-3*delay)
}

// total is the score so far.
func (g *Game) total() int {
	t := 0
	for _, r := range g.results {
		t += r.total
	}
	return t
}

// rank grades the whole run.
func (g *Game) rank() string {
	avg := g.total() / (len(stations) - 1)
	switch {
	case avg >= 90:
		return "S"
	case avg >= 75:
		return "A"
	case avg >= 55:
		return "B"
	case avg >= 35:
		return "C"
	}
	return "D"
}

// autopilot returns the handle position of a steady driver: run near the
// limit, brake for lower limits ahead and stop at the next mark.
func (g *Game) autopilot() int {
	kmh := g.v * 3.6
	st := stations[min(g.next, len(stations)-1)]
	slope := gravity * gradeAt(g.pos-trainLength/2) // helps braking when it climbs
	// need is the braking rate, in km/h/s, that meets every target ahead.
	need := float32(0)
	if d := st.stop - g.pos; d > 0.05 {
		need = g.v * g.v / (2 * d) * 3.6
	} else {
		return -maxBrake / 2
	}
	for _, l := range limits {
		d := l.from - g.pos - 20
		if l.from <= g.pos || d > 800 {
			continue
		}
		vl := float32(l.kmh-3) / 3.6
		if g.v > vl {
			need = max(need, (g.v*g.v-vl*vl)/(2*max(d, 1))*3.6)
		}
	}
	if need > 2.0 {
		g.apBraking = true
	}
	if g.apBraking {
		if need < 0.4 && kmh > 5 {
			g.apBraking = false
		} else {
			b := need + slope - resistance(kmh) - g.power
			// The brakes build up slowly, so ask a little ahead of the need.
			b += (b - g.brake) * 0.6
			n := int(b/brakeStep + 0.5)
			return -max(1, min(maxBrake, n))
		}
	}
	target := float32(g.limitNow() - 3)
	switch {
	case kmh > target+1:
		return -2
	case kmh < target-4:
		return maxPower
	case kmh < target:
		return 2
	}
	return 0
}
