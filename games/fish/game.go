// Package fish is an original sea fishing game seen from the side: cast from
// a boat, sink the jig to the depth where the fish are, wait out the nibbles
// and strike on the bite, then reel while the fish rests and give line while
// it runs, so that the line neither breaks nor goes slack. Each kind of fish
// keeps to its depth and is most active at its hours, from dawn to dusk. It
// uses only the Ebitengine API, and every image is drawn by code, so it also
// builds with Ebitengine.
//
// Positions are in meters: x away from the boat and depth down from the
// surface; the air above the sea has negative depth.
package fish

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	tps          = 60
	dt           = 1.0 / tps

	dayStart    = 6 * 3600
	dayEnd      = 18 * 3600
	clockRate   = 2 // seconds of the day per tick: the day lasts six minutes
	maxFish     = 10
	maxCast     = 58
	minCast     = 6
	tipX        = 1.5 // the rod's tip, in meters from the boat
	tipDepth    = -3
	sinkRate    = 1.6 // m/s, after the jig lands
	dropRate    = 2.4 // m/s, letting line out
	reelRate    = 1.2 // m/s, the jig coming in
	fastReel    = 4.5
	jigLift     = 0.8 // meters a jig jumps
	jigTicks    = 90  // ticks a jig stays attractive
	noticeRange = 5
	jigRange    = 9
	maxLine     = 150
	landLine    = 3.5 // line length at which the fish is in the net
	breakTicks  = 15  // ticks over full tension that break the line
	slackTicks  = 90  // ticks of a slack line that let the hook fall out
)

type mode uint8

const (
	modeTitle   mode = iota
	modeCast         // holding A to choose the cast's power
	modeFlight       // the jig is in the air
	modeFishing      // the jig is in the water
	modeFight
	modeCatch // showing the catch
	modeResult
)

type species uint8

const (
	aji species = iota
	saba
	madai
	hirame
	ika
	tachiuo
	buri
	boot
	kinds
)

// speciesType holds a kind's depth band in meters (bottom fish keep within
// bottomBand of the seabed), lengths in cm, fight and bite, weight factor
// (kg for a fish 100 cm long), and how active it is at dawn, by day and at
// dusk.
type speciesType struct {
	name               string
	shallow, deep      float32
	bottom             bool
	minLen, maxLen     float32
	strength, runSpeed float32
	run, rest          [2]int // ticks
	nibbles            [2]int
	window             int // ticks the bite lasts
	weight             float32
	bonus              int
	dawn, day, dusk    float32
	spawn, speed       float32
	jig                float32 // how much a jigged lure draws it
}

const bottomBand = 2.5

var speciesTypes = [kinds]speciesType{
	aji:     {"AJI", 3, 15, false, 15, 32, 0.35, 2.0, [2]int{30, 70}, [2]int{60, 120}, [2]int{1, 2}, 30, 12, 0, 1.0, 0.6, 1.0, 3, 1.6, 1},
	saba:    {"SABA", 5, 20, false, 25, 45, 0.6, 3.5, [2]int{60, 140}, [2]int{50, 100}, [2]int{0, 1}, 24, 10, 20, 0.8, 0.6, 0.9, 2, 2.4, 1},
	madai:   {"MADAI", 22, 50, false, 30, 75, 0.85, 3.0, [2]int{90, 180}, [2]int{60, 120}, [2]int{2, 3}, 30, 15, 100, 1.0, 0.5, 0.8, 1.2, 1.2, 1},
	hirame:  {"HIRAME", 0, 0, true, 35, 80, 0.7, 2.2, [2]int{60, 120}, [2]int{90, 150}, [2]int{3, 5}, 45, 10, 80, 0.8, 0.5, 0.8, 1, 0.6, 0.7},
	ika:     {"IKA", 8, 28, false, 18, 40, 0.3, 2.5, [2]int{20, 40}, [2]int{80, 140}, [2]int{1, 1}, 36, 6, 60, 0.9, 0.2, 1.2, 1.5, 1.0, 2.5},
	tachiuo: {"TACHIUO", 12, 35, false, 70, 120, 0.55, 3.0, [2]int{40, 90}, [2]int{70, 120}, [2]int{2, 3}, 24, 0.8, 50, 0.3, 0.1, 1.5, 1, 1.4, 1.5},
	buri:    {"BURI", 10, 40, false, 60, 100, 1.15, 3.5, [2]int{90, 170}, [2]int{60, 110}, [2]int{0, 0}, 20, 12, 200, 0.9, 0.3, 0.7, 0.35, 3.0, 1.2},
	boot:    {"OLD BOOT", 0, 0, true, 26, 28, 0.1, 0, [2]int{1, 1}, [2]int{600, 600}, [2]int{0, 0}, 60, 0, 0, 0, 0, 0, 0, 0, 0},
}

// The fish's states.
const (
	swim     = iota
	approach // to the jig
	nibble
	bite
	flee
	hooked
)

type fish struct {
	kind        species
	alive       bool
	x, depth    float32
	vx          float32
	facing      float32 // 1 swimming away from the boat, -1 toward it
	length      float32 // cm
	targetDepth float32
	state       int
	timer       int
	nibbles     int
	cool        int // ticks before it may notice the jig again
}

// catch is a landed fish.
type catch struct {
	kind   species
	length float32
	weight float32 // kg
	points int
}

const (
	msgNone = iota
	msgFishOn
	msgTooEarly
	msgMissed
	msgBroke
	msgGotAway
	msgBottom
	msgSpool
)

var messages = [...]string{"", "FISH ON!", "TOO EARLY", "MISSED", "LINE BROKE", "IT GOT AWAY", "BOTTOM", "OUT OF LINE"}

type Game struct {
	mode             mode
	ticks, modeTicks int
	rng              uint32
	clock            int // seconds of the day

	power     float32 // of the cast, 0 to 1
	powerUp   bool
	charging  bool // A has been pressed to cast
	castX     float32
	lureX     float32
	lureDepth float32
	jigOff    float32 // how far a jig lifted the lure; it settles back
	jigTime   int
	freeFall  bool
	onBottom  int // ticks the jig has lain on the seabed
	twitch    int // ticks the rod's tip dips for a nibble

	fish   [maxFish]fish
	suitor int // the fish going for the jig, or -1

	// The fight.
	line     float32 // meters of line out to the hooked fish
	angle    float32 // the line's direction below the horizontal, radians
	tension  float32 // 1 breaks the line
	stamina  float32
	running  bool
	phase    int // ticks left of the run or the rest
	over     int
	slack    int
	reeling  bool
	lastLand catch

	catches [kinds]int
	best    [kinds]float32
	score   int
	total   float32 // kg
	newBest bool

	message     int
	messageTime int
}

func NewGame() *Game {
	g := &Game{rng: 2463534242}
	g.reset()
	return g
}

func (g *Game) reset() {
	g.clock = dayStart
	g.score, g.total = 0, 0
	g.catches, g.best = [kinds]int{}, [kinds]float32{}
	g.suitor = -1
	g.lureX, g.lureDepth = tipX, tipDepth
	for i := range g.fish {
		g.spawn(i, true)
	}
}

// random returns 0 to n-1 from a xorshift generator, the same on every platform.
func (g *Game) random(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}

func (g *Game) frand() float32 { return float32(g.random(10001)) / 10000 }

func (g *Game) between(r [2]int) int { return r[0] + g.random(r[1]-r[0]+1) }

func (g *Game) say(m int) { g.message, g.messageTime = m, 120 }

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

func sinf(x float32) float32 { return float32(math.Sin(float64(x))) }

// seabed is the depth of the bottom x meters from the boat: shallow near the
// shore side, deeper out, with banks.
func seabed(x float32) float32 {
	return 16 + 0.62*x + 3*sinf(x*0.19) + 1.5*sinf(x*0.53+1)
}

// hour is the time of day in hours.
func (g *Game) hour() float32 { return float32(g.clock) / 3600 }

// activity is how eager a kind is at this hour: the hours around sunrise and
// sunset are the best.
func (g *Game) activity(k species) float32 {
	t := &speciesTypes[k]
	switch h := g.hour(); {
	case h < 8:
		return t.dawn
	case h < 15.5:
		return t.day
	}
	return t.dusk
}

// spawn puts a new fish in slot i, at either end of the water or anywhere.
func (g *Game) spawn(i int, anywhere bool) {
	sum := float32(0)
	for k := range boot {
		sum += speciesTypes[k].spawn * (0.3 + g.activity(k))
	}
	r := g.frand() * sum
	k := aji
	for ; k < boot-1; k++ {
		r -= speciesTypes[k].spawn * (0.3 + g.activity(k))
		if r < 0 {
			break
		}
	}
	t := &speciesTypes[k]
	f := fish{kind: k, alive: true, facing: 1}
	f.x = -4
	if anywhere {
		f.x = 6 + g.frand()*54
	} else if g.random(2) == 0 {
		f.x, f.facing = 66, -1
	}
	f.length = t.minLen + (t.maxLen-t.minLen)*g.frand()*g.frand() // big ones are rare
	f.depth = g.newDepth(&f)
	f.targetDepth = f.depth
	g.fish[i] = f
}

// newDepth picks a depth within the fish's band and above the seabed.
func (g *Game) newDepth(f *fish) float32 {
	t := &speciesTypes[f.kind]
	floor := seabed(f.x)
	if t.bottom {
		return floor - 0.5 - g.frand()*(bottomBand-0.5)
	}
	return min(t.shallow+(t.deep-t.shallow)*g.frand(), floor-1)
}

// inBand reports whether the jig is where the fish keeps.
func (g *Game) inBand(f *fish) bool {
	t := &speciesTypes[f.kind]
	d := g.jigDepth()
	if t.bottom {
		return d > seabed(g.lureX)-bottomBand-1
	}
	return d > t.shallow-4 && d < t.deep+4
}

func (g *Game) jigDepth() float32 { return g.lureDepth - g.jigOff }

func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

func aPressed() bool {
	return pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace) || pressed(ebiten.KeyEnter)
}

func aHeld() bool {
	return ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// input is what the player does in one tick.
type input struct {
	a, aHeld, b, up, down bool
}

func (g *Game) Update() error {
	g.ticks++
	in := input{a: aPressed(), aHeld: aHeld(), b: ebiten.IsKeyPressed(ebiten.KeyX),
		up: ebiten.IsKeyPressed(ebiten.KeyArrowUp), down: ebiten.IsKeyPressed(ebiten.KeyArrowDown)}
	switch g.mode {
	case modeTitle:
		g.modeTicks++
		g.swimAll()
		if in.a {
			g.reset()
			g.setMode(modeCast)
		}
	case modeResult:
		g.modeTicks++
		g.swimAll()
		if g.modeTicks > 60 && in.a {
			g.setMode(modeTitle)
		}
	default:
		g.step(in)
	}
	return nil
}

// step advances the day one tick.
func (g *Game) step(in input) {
	g.modeTicks++
	if g.messageTime > 0 {
		g.messageTime--
	}
	if g.twitch > 0 {
		g.twitch--
	}
	if g.mode != modeCatch && g.mode != modeFight {
		g.clock += clockRate
	}
	g.swimAll()
	switch g.mode {
	case modeCast:
		// The power swings while A is held, from a fresh press: the press
		// that started the day or closed the catch does not cast.
		if in.a {
			g.charging = true
		}
		if g.charging && in.aHeld {
			if g.powerUp {
				g.power += 0.02
			} else {
				g.power -= 0.02
			}
			if g.power >= 1 || g.power <= 0 {
				g.power = min(max(g.power, 0), 1)
				g.powerUp = !g.powerUp
			}
		} else if g.charging {
			g.castX = minCast + (maxCast-minCast)*g.power
			g.charging = false
			g.setMode(modeFlight)
		}
		if g.clock >= dayEnd {
			g.setMode(modeResult)
		}
	case modeFlight:
		// A one-second arc from the tip to the water.
		t := float32(g.modeTicks) / tps
		g.lureX = tipX + (g.castX-tipX)*t
		g.lureDepth = tipDepth - 12*t + 15*t*t
		if t >= 1 {
			g.lureX, g.lureDepth = g.castX, 0
			g.freeFall, g.onBottom = true, 0
			g.setMode(modeFishing)
		}
	case modeFishing:
		g.fishing(in)
	case modeFight:
		g.fight(in)
	case modeCatch:
		if g.modeTicks > 30 && in.a {
			g.recast()
		}
	}
}

// recast starts the next cast, or ends the day.
func (g *Game) recast() {
	g.power, g.powerUp, g.charging = 0, true, false
	g.lureX, g.lureDepth, g.jigOff = tipX, tipDepth, 0
	g.suitor = -1
	if g.clock >= dayEnd {
		g.setMode(modeResult)
		return
	}
	g.setMode(modeCast)
}

// fishing works the jig: it falls after landing until stopped; down lets out
// line, up reels in, A jigs it, B winds it all in. Fish come to it.
func (g *Game) fishing(in input) {
	floor := seabed(g.lureX)
	nibbling := g.suitor >= 0 && (g.fish[g.suitor].state == nibble || g.fish[g.suitor].state == bite)
	if in.a {
		switch {
		case nibbling && g.fish[g.suitor].state == bite:
			g.strike()
			return
		case nibbling:
			// Striking at a nibble scares the fish.
			f := &g.fish[g.suitor]
			f.state, f.timer, f.cool = flee, 120, 600
			g.suitor = -1
			g.say(msgTooEarly)
		case !g.freeFall:
			g.jigOff, g.jigTime = jigLift, jigTicks
		}
		g.freeFall = false
	}
	if in.up {
		g.freeFall = false
	}
	switch {
	case in.b:
		g.reelIn(fastReel)
	case in.up:
		g.reelIn(reelRate)
	case in.down || g.freeFall:
		rate := float32(dropRate)
		if g.freeFall {
			rate = sinkRate
		}
		g.lureDepth = min(g.lureDepth+rate*dt, floor-0.2)
	}
	if g.lureDepth >= floor-0.25 {
		if g.freeFall {
			g.say(msgBottom)
		}
		g.freeFall = false
		g.onBottom++
	} else {
		g.onBottom = 0
	}
	g.jigOff = max(0, g.jigOff-0.03)
	if g.jigTime > 0 {
		g.jigTime--
	}
	if g.mode != modeFishing {
		return
	}
	// An old boot lies on the bottom here and there.
	if g.onBottom > 6*tps && g.suitor < 0 && g.random(900) == 0 {
		g.hook(-1)
		return
	}
	if g.clock >= dayEnd && g.suitor < 0 {
		g.recast()
	}
}

// reelIn brings the jig toward the rod's tip; at the boat it is out.
func (g *Game) reelIn(speed float32) {
	dx, dd := tipX-g.lureX, -g.lureDepth
	if g.lureDepth <= 0.3 {
		dd = 0
	}
	l := float32(math.Sqrt(float64(dx*dx + dd*dd)))
	step := speed * dt
	if l <= step || g.lureX <= tipX+0.5 && g.lureDepth <= 0.3 {
		g.recast()
		return
	}
	g.lureX += dx / l * step
	g.lureDepth += dd / l * step
}

// strike sets the hook in the biting fish.
func (g *Game) strike() {
	g.hook(g.suitor)
}

// hook starts the fight with fish i, or with an old boot for -1.
func (g *Game) hook(i int) {
	if i >= 0 {
		g.fish[i].state = hooked
	} else {
		// The boot takes a fish slot for the fight.
		i = g.random(maxFish)
		g.fish[i] = fish{kind: boot, alive: true, x: g.lureX, depth: g.lureDepth, length: 27, state: hooked, facing: 1}
	}
	g.suitor = i
	f := &g.fish[i]
	g.line = g.lineLength()
	g.angle = float32(math.Atan2(float64(g.lureDepth-tipDepth), float64(g.lureX-tipX)))
	g.tension, g.stamina = 0.3, 1
	g.running, g.phase = true, g.between(speciesTypes[f.kind].run)
	g.over, g.slack = 0, 0
	g.say(msgFishOn)
	g.setMode(modeFight)
}

func (g *Game) lineLength() float32 {
	dx, dd := g.lureX-tipX, g.lureDepth-tipDepth
	return float32(math.Sqrt(float64(dx*dx + dd*dd)))
}

// pull is how hard the hooked fish pulls, 0 to about 1.2.
func (g *Game) pull() float32 {
	f := &g.fish[g.suitor]
	t := &speciesTypes[f.kind]
	size := f.length / t.maxLen
	p := t.strength * (0.7 + 0.6*size) * (0.35 + 0.65*g.stamina)
	if !g.running {
		p *= 0.4
	}
	return p
}

// fight plays the hooked fish. Holding A reels: the tension climbs, more so
// against a running fish. Letting go gives line to a run. Full tension for a
// quarter second breaks the line; a slack line lets the hook fall out.
func (g *Game) fight(in input) {
	f := &g.fish[g.suitor]
	t := &speciesTypes[f.kind]
	g.reeling = in.aHeld
	g.phase--
	if g.phase <= 0 {
		g.running = !g.running
		// A tired fish runs shorter and rests longer.
		if g.running {
			g.phase = int(float32(g.between(t.run)) * (0.3 + 0.7*g.stamina))
		} else {
			g.phase = int(float32(g.between(t.rest)) * (2 - g.stamina))
		}
	}
	p := g.pull()
	want := p * 0.55
	if g.reeling {
		want = p + 0.22
	}
	g.tension += (want - g.tension) * 0.15
	if g.reeling {
		g.line -= 3 * (1 - min(p, 1)*0.75) * dt
	} else if g.running {
		g.line += p * t.runSpeed * dt
	}
	size := max(f.length/t.maxLen, 0.3)
	drain := g.tension * 0.12
	if g.running {
		drain += 0.1
	}
	g.stamina = max(0, g.stamina-drain*dt/size)
	// The fish swims away and down while it runs, and swings about.
	if g.running {
		g.angle += (0.9 - g.angle) * 0.004
	}
	g.angle += sinf(float32(g.modeTicks)*0.05) * 0.003
	g.placeHooked()

	switch {
	case g.tension > 1:
		g.over++
	default:
		g.over = 0
	}
	if g.tension < 0.12 {
		g.slack++
	} else {
		g.slack = 0
	}
	switch {
	case g.over > breakTicks:
		g.lose(msgBroke)
	case g.slack > slackTicks:
		g.lose(msgGotAway)
	case g.line > maxLine:
		g.lose(msgSpool)
	case g.line < landLine:
		g.land()
	}
}

// placeHooked puts the hooked fish and the jig at the end of the line.
func (g *Game) placeHooked() {
	f := &g.fish[g.suitor]
	c, s := float32(math.Cos(float64(g.angle))), float32(math.Sin(float64(g.angle)))
	x := tipX + c*g.line
	d := tipDepth + s*g.line
	d = min(max(d, 0.4), seabed(x)-0.4)
	if g.line < landLine+3 {
		d = max(d*0.9, 0) // up to the surface for the net
	}
	f.x, f.depth = x, d
	f.facing = 1
	if g.reeling && !g.running {
		f.facing = -1
	}
	g.lureX, g.lureDepth = x, d
}

// lose ends the fight without the fish.
func (g *Game) lose(m int) {
	f := &g.fish[g.suitor]
	if f.kind == boot {
		f.alive = false
	} else {
		f.state, f.timer, f.cool = flee, 180, 1200
	}
	g.suitor = -1
	g.say(m)
	g.lureX, g.lureDepth = tipX, tipDepth
	g.recast()
}

// land nets the fish and scores it.
func (g *Game) land() {
	f := &g.fish[g.suitor]
	t := &speciesTypes[f.kind]
	l := f.length / 100
	c := catch{kind: f.kind, length: f.length, weight: t.weight * l * l * l}
	if f.kind != boot {
		c.points = int(c.weight*100) + t.bonus
	}
	g.lastLand = c
	g.catches[f.kind]++
	g.newBest = f.kind != boot && f.length > g.best[f.kind]
	if g.newBest {
		g.best[f.kind] = f.length
	}
	g.score += c.points
	if f.kind != boot {
		g.total += c.weight
	}
	f.alive = false
	g.spawn(g.suitor, false)
	g.suitor = -1
	g.setMode(modeCatch)
}

// swimAll moves the free fish and lets them find the jig.
func (g *Game) swimAll() {
	inWater := (g.mode == modeFishing) && g.lureDepth > 0
	for i := range g.fish {
		f := &g.fish[i]
		if !f.alive {
			g.spawn(i, false)
			continue
		}
		if f.state == hooked {
			continue
		}
		g.swim(i, f, inWater)
		if f.x < -6 || f.x > 68 {
			if i == g.suitor {
				g.suitor = -1
			}
			g.spawn(i, false)
		}
	}
}

func (g *Game) swim(i int, f *fish, inWater bool) {
	t := &speciesTypes[f.kind]
	if f.cool > 0 {
		f.cool--
	}
	if f.kind == boot {
		f.alive = false
		return
	}
	switch f.state {
	case swim:
		if g.random(240) == 0 {
			f.facing = -f.facing
		}
		if g.random(180) == 0 {
			f.targetDepth = g.newDepth(f)
		}
		f.vx = f.facing * t.speed * 0.5
		f.depth += (f.targetDepth - f.depth) * 0.01
		if inWater && g.suitor < 0 && f.cool == 0 && g.inBand(f) {
			r, chance := float32(noticeRange), float32(0.02)
			if g.jigTime > 0 {
				r, chance = jigRange, chance*t.jig
			}
			dx, dd := g.lureX-f.x, g.jigDepth()-f.depth
			if dx*dx+dd*dd < r*r && g.frand() < chance*g.activity(f.kind) {
				f.state = approach
				g.suitor = i
			}
		}
	case approach:
		if !inWater || i != g.suitor {
			f.state, f.cool = swim, 300
			break
		}
		dx, dd := g.lureX-f.x, g.jigDepth()-f.depth
		d := float32(math.Sqrt(float64(dx*dx + dd*dd)))
		if d < 0.6 {
			f.state, f.nibbles = nibble, g.between(t.nibbles)
			f.timer = 0
			f.vx = 0
			break
		}
		step := t.speed * 0.9 * dt
		f.vx = dx / d * t.speed * 0.9
		f.depth += dd / d * step
		if dx > 0 {
			f.facing = 1
		} else {
			f.facing = -1
		}
	case nibble, bite:
		if !inWater || i != g.suitor {
			f.state, f.cool = swim, 300
			break
		}
		// Stay at the jig.
		f.vx = (g.lureX - f.x - f.facing*0.5) * 2
		f.depth += (g.jigDepth() - f.depth) * 0.2
		f.timer--
		if f.timer > 0 {
			break
		}
		if f.state == bite {
			// Too late: it spat the hook.
			f.state, f.timer, f.cool = flee, 120, 600
			g.suitor = -1
			g.say(msgMissed)
			break
		}
		if f.nibbles > 0 {
			f.nibbles--
			f.timer = 30 + g.random(50)
			g.twitch = 8
		} else {
			f.state, f.timer = bite, t.window
			g.twitch = t.window
		}
	case flee:
		f.vx = f.facing * t.speed * 2.5
		f.timer--
		if f.timer <= 0 {
			f.state = swim
		}
	}
	f.x += f.vx * dt
	f.depth = min(f.depth, seabed(f.x)-0.3)
	f.depth = max(f.depth, 0.5)
}
