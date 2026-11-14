// Package uboat is an original submarine game: from a periscope, sink the
// ships of a passing convoy with torpedoes, and lower the periscope before the
// escort detects the submarine and attacks with depth charges. It uses only the
// Ebitengine API, and every image is drawn by code, so it also builds with
// Ebitengine. It draws at 240x240.
//
// The world is flat, in meters, with x to the east and y to the south. The
// submarine starts at the origin and moves only with the periscope down.
// Bearings are radians clockwise from north.
package uboat

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	horizon      = 116
	scopeX       = 120.0 // the periscope's circle
	scopeY       = 120.0
	scopeRadius  = 108.0
	sonarRadius  = 104
	sonarRange   = 2600 // meters at the sonar plot's rim

	wideView  = 40 * math.Pi / 180
	zoomView  = 12 * math.Pi / 180
	eyeHeight = 8 // meters; near ships sit a little below the horizon

	maxShips    = 10
	maxTorps    = 4
	maxCharges  = 16
	maxBlasts   = 12
	torpedoes   = 10
	reloadTicks = 120
	torpSpeed   = 3 // meters per tick
	torpRange   = 3200
	beam        = 16
	sinkTicks   = 360
	moveTicks   = 18 // raising or lowering the periscope
	subTurn     = 0.6 * math.Pi / 180
	subAccel    = 0.01
	subBound    = 3000 // the submarine stays within this many meters of the start

	// Detection, in percent per tick.
	detectBase   = 0.03
	detectNear   = 0.5 // added in full when a destroyer is within detectClose
	detectClose  = 300
	detectFire   = 10
	detectFall   = 0.25
	detectEngine = 0.2 // subtracted from detectFall in proportion to the submarine's speed
	sonarPing    = 0.1 // even with the periscope down, within pingRange
	pingRange    = 150
	contactLost  = 35
	chargeFuse   = 70
	chargeReach  = 90 // meters within which a depth charge damages the hull
	chargeDamage = 35
)

type kind uint8

const (
	cargo kind = iota
	tanker
	destroyer
	kinds
)

// shipTypes holds length and height in meters, hits to sink, and tonnage.
var shipTypes = [kinds]struct {
	length, height float32
	hp, tons       int
}{
	cargo:     {120, 40, 1, 5000},
	tanker:    {160, 30, 2, 8000},
	destroyer: {100, 34, 1, 1500},
}

type ship struct {
	kind                 kind
	x, y, heading, speed float32
	hp                   int
	sink                 int // ticks since it started sinking
	alive                bool
	// A destroyer's attack: where it thinks the submarine is, the error of
	// that guess, ticks of the straight run over it, and depth charges left in
	// this run.
	aimX, aimY, errX, errY float32
	runTicks, charges      int
	chargeCool, slot       int
}

type torp struct {
	x, y, vx, vy, run float32
	alive             bool
}

type charge struct {
	x, y float32
	fuse int // ticks until it explodes; negative counts ticks since
	used bool
}

type blast struct {
	x, y  float32
	age   int
	alive bool
}

type mode int

const (
	modeTitle mode = iota
	modePlay
	modeResult // the convoy has passed or is sunk
	modeOver   // the submarine was sunk
)

// Messages shown at the top of the view.
const (
	msgNone = iota
	msgDetected
	msgCharges
	msgLost
	msgHit
	msgSunk
	msgEmpty
)

var messages = [...]string{"", "DETECTED!", "DEPTH CHARGES!", "CONTACT LOST", "HIT!", "SUNK!", "NO TORPEDOES"}

type Game struct {
	mode             mode
	ticks, modeTicks int
	rng              uint32
	level            int
	bearing          float32
	turnHeld         int
	zoomed, up       bool
	moving           int // ticks left of raising or lowering
	detect           float32
	// The submarine: position, heading, speed and the chosen throttle step.
	subX, subY           float32
	subHeading, subSpeed float32
	throttle             int
	alert                bool
	hull                 float32
	torpedoes            int
	reload               int
	tonnage, sunk        int
	shake                int
	message              int
	messageTicks         int
	ships                [maxShips]ship
	torps                [maxTorps]torp
	charges              [maxCharges]charge
	blasts               [maxBlasts]blast
	order                [maxShips]int
	clips                [ScreenHeight - clipTop + 1]*ebiten.Image // screen rows above clipTop+i, for ships at the waterline
	clipsOf              *ebiten.Image
	art                  *sprites
	filled               *ebiten.Image // the screen filled whole, so opaque
}

// clipTop is the highest waterline row, with the view shaken up.
const clipTop = horizon - 2

func NewGame() *Game {
	// Draw covers the whole screen anyway.
	ebiten.SetScreenClearedEveryFrame(false)
	g := &Game{rng: 2463534242, art: newSprites()}
	g.newCampaign()
	return g
}

func (g *Game) newCampaign() {
	g.level, g.hull, g.tonnage, g.sunk = 1, 100, 0, 0
	g.newMission()
}

// newMission sets up a convoy that comes from the west, two columns of
// merchants with destroyers on the near side.
func (g *Game) newMission() {
	clear(g.ships[:])
	clear(g.torps[:])
	clear(g.charges[:])
	clear(g.blasts[:])
	g.detect, g.alert, g.shake, g.message = 0, false, 0, msgNone
	g.up, g.moving, g.zoomed = true, 0, false
	g.torpedoes, g.reload = torpedoes, 0
	g.bearing = -50 * math.Pi / 180
	g.subX, g.subY, g.subHeading, g.subSpeed, g.throttle = 0, 0, 0, 0, 0
	merchants := 4 + min(g.level-1, 3)
	escorts := min(1+(g.level-1)/2, 3)
	convoyY := -1100 - float32(g.random(300))
	speed := 0.6 + 0.05*float32(min(g.level-1, 4))
	n := 0
	for i := range merchants {
		k := cargo
		if i%3 == 1 {
			k = tanker
		}
		g.ships[n] = ship{kind: k, x: -1500 - float32(i/2)*300 - float32(i%2)*120, y: convoyY + float32(i%2)*140,
			heading: math.Pi / 2, speed: speed, hp: shipTypes[k].hp, alive: true}
		n++
	}
	for i := range escorts {
		g.ships[n] = ship{kind: destroyer, x: -1600 + float32(i)*500, y: convoyY + 300,
			heading: math.Pi / 2, speed: speed, hp: shipTypes[destroyer].hp, alive: true, slot: i}
		n++
	}
}

// random returns 0 to n-1 from a xorshift generator, the same on every platform.
func (g *Game) random(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

func (g *Game) say(m int) { g.message, g.messageTicks = m, 120 }

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
	if g.shake > 0 {
		g.shake--
	}
	switch g.mode {
	case modeTitle:
		g.bearing = angleDiff(g.bearing+0.002, 0)
		if startPressed() {
			g.newCampaign()
			g.setMode(modePlay)
		}
	case modePlay:
		g.updateScope()
		g.updateSub()
		g.updateShips()
		g.updateTorps()
		g.updateCharges()
		g.updateDetection()
		if g.hull <= 0 {
			g.hull = 0
			g.setMode(modeOver)
		} else if g.convoyGone() {
			g.setMode(modeResult)
		}
	case modeResult:
		g.updateShips()
		g.updateCharges()
		if g.modeTicks > 60 && startPressed() {
			g.level++
			g.hull = min(g.hull+30, 100) // repaired between missions
			g.newMission()
			g.setMode(modePlay)
		}
	case modeOver:
		if g.modeTicks > 60 && startPressed() {
			g.newCampaign()
			g.setMode(modeTitle)
		}
	}
	for i := range g.blasts {
		if b := &g.blasts[i]; b.alive {
			b.age++
			b.alive = b.age < 40
		}
	}
	return nil
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

// rel returns a point relative to the submarine.
func (g *Game) rel(x, y float32) (float32, float32) { return x - g.subX, y - g.subY }

func hypot(x, y float32) float32 { return float32(math.Sqrt(float64(x*x + y*y))) }

// bearingTo returns the bearing from the submarine to a point.
func bearingTo(x, y float32) float32 { return float32(math.Atan2(float64(x), float64(-y))) }

func (g *Game) view() float32 {
	if g.zoomed {
		return zoomView
	}
	return wideView
}

func (g *Game) updateScope() {
	if g.moving > 0 {
		g.moving--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) && g.moving == 0 {
		g.up = !g.up
		g.moving = moveTicks
	}
	if g.reload > 0 {
		g.reload--
	}
	if g.moving > 0 {
		g.turnHeld = 0
		return
	}
	if !g.up {
		g.turnHeld = 0
		g.steerSub()
		return
	}
	turn := 0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		turn--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		turn++
	}
	if turn == 0 {
		g.turnHeld = 0
	} else {
		// Turning speeds up while held: fine aim at first, then a fast sweep.
		g.turnHeld++
		speed := (0.15 + float32(min(g.turnHeld, 60))/60*1.35) * math.Pi / 180
		if g.zoomed {
			speed /= 3
		}
		g.bearing = angleDiff(g.bearing+float32(turn)*speed, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.zoomed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.zoomed = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.fire()
	}
}

// throttles are the submarine's speeds in meters per tick: stop, slow, full.
var throttles = [...]float32{0, 0.35, 0.7}

var throttleNames = [len(throttles)]string{"STOP", "SLOW", "FULL"}

// steerSub turns the submarine and changes its throttle with the periscope
// down.
func (g *Game) steerSub() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.subHeading = angleDiff(g.subHeading-subTurn, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.subHeading = angleDiff(g.subHeading+subTurn, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.throttle = min(g.throttle+1, len(throttles)-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.throttle = max(g.throttle-1, 0)
	}
}

// updateSub moves the submarine. With the periscope up it slows to a stop,
// which keeps the view steady for aiming.
func (g *Game) updateSub() {
	want := throttles[g.throttle]
	if g.up {
		want = 0
	}
	g.subSpeed += min(max(want-g.subSpeed, -subAccel), subAccel)
	s, c := sincos(g.subHeading)
	g.subX = min(max(g.subX+s*g.subSpeed, -subBound), subBound)
	g.subY = min(max(g.subY-c*g.subSpeed, -subBound), subBound)
}

func (g *Game) fire() {
	if g.torpedoes == 0 {
		g.say(msgEmpty)
		return
	}
	if g.reload > 0 {
		return
	}
	for i := range g.torps {
		if t := &g.torps[i]; !t.alive {
			s, c := sincos(g.bearing)
			*t = torp{x: g.subX, y: g.subY, vx: s * torpSpeed, vy: -c * torpSpeed, alive: true}
			g.torpedoes--
			g.reload = reloadTicks
			g.detect += detectFire // the launch is loud
			return
		}
	}
}

func (g *Game) updateShips() {
	cx, cy, n := float32(0), float32(0), 0
	for i := range g.ships {
		if s := &g.ships[i]; s.alive && s.kind != destroyer && s.sink == 0 {
			cx, cy, n = cx+s.x, cy+s.y, n+1
		}
	}
	if n > 0 {
		cx, cy = cx/float32(n), cy/float32(n)
	}
	for i := range g.ships {
		s := &g.ships[i]
		if !s.alive {
			continue
		}
		if s.sink > 0 {
			if s.sink++; s.sink > sinkTicks {
				s.alive = false
			}
			continue
		}
		if s.kind == destroyer {
			g.steerDestroyer(s, cx, cy, n > 0)
		}
		sn, cs := sincos(s.heading)
		s.x += sn * s.speed
		s.y -= cs * s.speed
		if s.kind != destroyer && s.x > sonarRange+200 {
			s.alive = false // escaped
		}
	}
}

// steerDestroyer escorts the convoy, weaving on its near side, or, when the
// submarine is detected, runs over where it thinks the submarine is and drops
// depth charges. Each run narrows the guess.
func (g *Game) steerDestroyer(s *ship, cx, cy float32, convoy bool) {
	if s.chargeCool > 0 {
		s.chargeCool--
	}
	var tx, ty float32
	switch {
	case g.alert:
		s.speed = 1.3
		if s.runTicks > 0 {
			s.runTicks--
			if s.charges > 0 && s.chargeCool == 0 {
				g.drop(s.x, s.y)
				s.charges--
				s.chargeCool = 18
			}
			if s.runTicks == 0 {
				g.reaim(s)
			}
			return // keep straight through the run
		}
		tx, ty = s.aimX, s.aimY
		if hypot(tx-s.x, ty-s.y) < 30 {
			s.runTicks, s.charges = 150, 5
			return
		}
	case convoy:
		s.speed = 0.85
		w := float32(g.ticks)/400 + float32(s.slot)*2
		ws, wc := sincos(w)
		tx, ty = cx+400*ws-float32(s.slot)*350, cy+320+120*wc
	default:
		s.speed = 0.6
		s.heading = math.Pi / 2
		return
	}
	want := bearingTo(tx-s.x, ty-s.y)
	turn := angleDiff(want, s.heading)
	s.heading = angleDiff(s.heading+min(max(turn, -0.025), 0.025), 0)
}

// reaim takes a new fix on the submarine for the next run: its position now,
// off by the guess's error, which halves with each run.
func (g *Game) reaim(s *ship) {
	s.aimX, s.aimY = g.subX+s.errX, g.subY+s.errY
	s.errX *= 0.5
	s.errY *= 0.5
}

func (g *Game) drop(x, y float32) {
	for i := range g.charges {
		if c := &g.charges[i]; !c.used {
			*c = charge{x: x, y: y, fuse: chargeFuse, used: true}
			return
		}
	}
}

func (g *Game) updateCharges() {
	for i := range g.charges {
		c := &g.charges[i]
		if !c.used {
			continue
		}
		if c.fuse--; c.fuse == 0 {
			d := hypot(g.rel(c.x, c.y))
			if d < chargeReach {
				g.hull -= (1 - d/chargeReach) * chargeDamage
			}
			if d < chargeReach*2 {
				g.shake = 16
			}
			g.say(msgCharges)
		}
		if c.fuse < -40 {
			c.used = false
		}
	}
}

func (g *Game) updateTorps() {
	for i := range g.torps {
		t := &g.torps[i]
		if !t.alive {
			continue
		}
		t.x += t.vx
		t.y += t.vy
		if t.run += torpSpeed; t.run > torpRange {
			t.alive = false
			continue
		}
		for j := range g.ships {
			s := &g.ships[j]
			if !s.alive || s.sink > 0 {
				continue
			}
			// The torpedo in the ship's frame: along the keel and across it.
			sn, cs := sincos(s.heading)
			dx, dy := t.x-s.x, t.y-s.y
			along, across := dx*sn-dy*cs, dx*cs+dy*sn
			if abs(along) < shipTypes[s.kind].length/2 && abs(across) < beam/2 {
				t.alive = false
				g.addBlast(t.x, t.y)
				if s.hp--; s.hp > 0 {
					g.say(msgHit)
				} else {
					s.sink = 1
					g.tonnage += shipTypes[s.kind].tons
					g.sunk++
					g.say(msgSunk)
				}
				break
			}
		}
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func (g *Game) addBlast(x, y float32) {
	for i := range g.blasts {
		if b := &g.blasts[i]; !b.alive {
			*b = blast{x: x, y: y, alive: true}
			return
		}
	}
}

// nearestDestroyer returns the distance to the closest destroyer afloat.
func (g *Game) nearestDestroyer() float32 {
	nearest := float32(math.MaxFloat32)
	for i := range g.ships {
		if s := &g.ships[i]; s.alive && s.sink == 0 && s.kind == destroyer {
			nearest = min(nearest, hypot(g.rel(s.x, s.y)))
		}
	}
	return nearest
}

// updateDetection raises the detection while the periscope is up, faster with
// a destroyer near, and lowers it while the periscope is down. At 100 the
// destroyers attack; they give up below contactLost.
func (g *Game) updateDetection() {
	nearest := g.nearestDestroyer()
	if g.up {
		g.detect += detectBase + detectNear*min(1, detectClose/nearest)
	} else {
		g.detect -= detectFall - detectEngine*g.subSpeed/throttles[len(throttles)-1]
	}
	if nearest < pingRange {
		g.detect += sonarPing
	}
	g.detect = min(max(g.detect, 0), 100)
	if !g.alert && g.detect >= 100 {
		g.alert = true
		g.say(msgDetected)
		for i := range g.ships {
			if s := &g.ships[i]; s.kind == destroyer {
				// The first guess is up to 80 meters off.
				s.errX, s.errY = float32(g.random(161)-80), float32(g.random(161)-80)
				s.runTicks, s.charges = 0, 0
				g.reaim(s)
			}
		}
	} else if g.alert && g.detect < contactLost {
		g.alert = false
		g.say(msgLost)
	}
}

// convoyGone reports whether every merchant has sunk or escaped.
func (g *Game) convoyGone() bool {
	for i := range g.ships {
		if s := &g.ships[i]; s.alive && s.kind != destroyer {
			return false
		}
	}
	return true
}

// lead returns the bearing at which a torpedo fired now meets the ship, if it
// can: |P + V t| = torpSpeed t for the smallest positive t, with P relative to
// the submarine.
func (g *Game) lead(s *ship) (float32, bool) {
	sn, cs := sincos(s.heading)
	vx, vy := sn*s.speed, -cs*s.speed
	px, py := g.rel(s.x, s.y)
	a := vx*vx + vy*vy - torpSpeed*torpSpeed
	b := 2 * (px*vx + py*vy)
	c := px*px + py*py
	disc := b*b - 4*a*c
	if disc < 0 || a == 0 {
		return 0, false
	}
	root := float32(math.Sqrt(float64(disc)))
	t := (-b - root) / (2 * a)
	if t < 0 {
		t = (-b + root) / (2 * a)
	}
	if t < 0 || t*torpSpeed > torpRange {
		return 0, false
	}
	return bearingTo(px+vx*t, py+vy*t), true
}

// target returns the ship afloat closest to the crosshair, if one is in view.
func (g *Game) target() *ship {
	var best *ship
	bestDiff := g.view() / 2
	for i := range g.ships {
		s := &g.ships[i]
		if !s.alive || s.sink > 0 {
			continue
		}
		if d := abs(angleDiff(bearingTo(g.rel(s.x, s.y)), g.bearing)); d < bestDiff {
			best, bestDiff = s, d
		}
	}
	return best
}
