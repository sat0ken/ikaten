// Package tank is an original tank battle seen through the gunsight: hunt the
// enemy tanks across fields and villages. Armor is thick in front and thin at
// the sides and rear, shells drop with range so the gun is laid with the
// sight's range marks, and houses and trees stop shells and hide tanks. The
// hull and the turret turn separately. It uses only the Ebitengine API, so it
// also builds with Ebitengine.
//
// The world is package field's: flat, in meters, x to the east and y to the
// south. Bearings are radians clockwise from north.
package tank

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/tank/field"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	centerX      = field.SightX // the gun's axis on screen
	centerY      = field.SightY
	focalWide    = 104 // pixels per radian: about 60 degrees across
	focalZoom    = 832 // eight times
	eyeHeight    = 2.3 // the sight and the gun, in meters above the ground
	farthest     = 1500

	shellSpeed = 10     // meters per tick
	gravity    = 0.0072 // meters per tick squared: three times real, so that the drop shows
	shellRange = 2000
	aimHeight  = 1.2 // enemies aim at the middle of the hull

	maxEnemies  = 8
	maxShells   = 12
	maxPuffs    = 32
	reloadTicks = 180
	startAmmo   = 40
	tankRadius  = 3.6
	bound       = field.Half - 24
	minElev     = -0.015
	maxElev     = 0.05
	tapTicks    = 20 // B released this soon, without the stick, toggles the zoom
	hitDamage   = 34
	accel       = 0.004
	spotRange   = 1100
	fireRange   = 1000
)

type kind uint8

const (
	light kind = iota
	medium
	heavy
	player
	kinds
)

// Armor faces.
const (
	front = iota
	side
	rear
	top
)

// spec holds a tank's size in meters, armor in millimeters, its gun's
// penetration at the muzzle and the loss per meter, top speed in meters per
// tick, turn and traverse rates in radians per tick, and reload ticks.
type spec struct {
	length, width, height float32
	armor                 [4]float32
	pen, penLoss          float32
	speed, turn, traverse float32
	reload                int
}

var specs = [kinds]spec{
	light:  {5.0, 2.6, 2.0, [4]float32{50, 30, 20, 15}, 95, 0.03, 0.30, 0.012, 0.012, 220},
	medium: {6.4, 3.0, 2.4, [4]float32{100, 55, 40, 20}, 135, 0.04, 0.22, 0.009, 0.008, 260},
	heavy:  {7.0, 3.4, 2.7, [4]float32{180, 90, 60, 30}, 175, 0.05, 0.14, 0.006, 0.005, 320},
	player: {6.4, 3.0, 2.4, [4]float32{120, 60, 40, 20}, 180, 0.05, 0.25, 0.009, 0.012, reloadTicks},
}

// throttles are the player's speeds for reverse, stop, half and full.
var (
	throttles     = [...]float32{-0.08, 0, 0.12, 0.25}
	throttleNames = [len(throttles)]string{"REV", "STOP", "HALF", "FULL"}
)

const throttleStop = 1

// Enemy states.
const (
	holding = iota
	driving
)

type tank struct {
	kind                     kind
	x, y, heading, turret    float32
	speed                    float32
	alive                    bool
	dead                     int // ticks since it was destroyed
	state, stateTicks, until int
	wpX, wpY, steer          float32
	blocked                  int
	reload                   int
	sees                     bool // a clear line to the player
	lastSeen                 int
	seenX, seenY             float32 // where the player was last seen
	errRange, errSide        float32 // the spread of the next shot, in meters
	shotX, shotY             float32 // where the player was at the last shot
	flash                    int
}

type shell struct {
	x, y, h, vx, vy, vh float32
	from                int // the enemy that fired, or -1 for the player
	pen, penLoss, run   float32
	alive               bool
}

type puffKind uint8

const (
	dust puffKind = iota
	smoke
	fire
	spark
	flash
	puffKinds
)

// puffSpecs holds a puff's size in meters, its growth over its life, how fast it
// rises in meters per tick, and its life in ticks.
var puffSpecs = [puffKinds]struct {
	size, grow, rise float32
	life             int
}{
	dust:  {6, 0.8, 0.03, 50},
	smoke: {5, 1.5, 0.06, 110},
	fire:  {3, 0.3, 0.02, 24},
	spark: {1.6, 0, 0, 8},
	flash: {4, 0.5, 0, 8},
}

type puff struct {
	x, y, h   float32
	kind      puffKind
	age       int
	alive     bool
	life      int
	sizeScale float32
}

type mode int

const (
	modeTitle mode = iota
	modePlay
	modeClear
	modeOver
)

// Messages shown at the top of the view.
const (
	msgNone = iota
	msgKill
	msgRicochet
	msgNoPen
	msgWreck
	msgEmpty
	msgHitFront
	msgHitLeft
	msgHitRight
	msgHitRear
	msgHeldFront
	msgHeldLeft
	msgHeldRight
	msgHeldRear
)

var messages = [...]string{"", "TARGET DESTROYED", "RICOCHET", "NO PENETRATION", "WRECK HIT", "NO AMMO",
	"HIT FRONT!", "HIT LEFT!", "HIT RIGHT!", "HIT REAR!",
	"BOUNCED: FRONT", "BOUNCED: LEFT", "BOUNCED: RIGHT", "BOUNCED: REAR"}

type Game struct {
	mode             mode
	ticks, modeTicks int
	rng              uint32
	level, kills     int
	// The player's tank: hull, gun and controls.
	x, y, heading, speed float32
	throttle             int
	turret, elev         float32 // the turret's bearing in the world and the gun's elevation
	zoomed               bool
	turnHeld, elevHeld   int
	bHeld                int
	bUsed                bool
	reload, ammo         int
	hp                   float32
	shake, hurt, recoil  int
	hitBearing           float32
	hitTicks             int
	message              int
	messageTicks         int

	enemies [maxEnemies]tank
	shells  [maxShells]shell
	puffs   [maxPuffs]puff

	// Drawing state: the camera and the cached screen rows.
	cam     camera
	items   [maxItems]drawItem
	rows    [ScreenHeight]*ebiten.Image
	rowsOf  *ebiten.Image
	ranging int // the range under the reticle in meters, or 0
}

func NewGame() *Game {
	g := &Game{rng: 2463534242}
	g.newBattle(2)
	return g
}

// random returns 0 to n-1 from a xorshift generator, the same on every platform.
func (g *Game) random(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}

// spread returns a random number from -1 to 1.
func (g *Game) spread() float32 { return float32(g.random(2001)-1000) / 1000 }

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

func (g *Game) say(m int) { g.message, g.messageTicks = m, 120 }

func startPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// lanes are where the enemies start, clear of obstacles.
var lanes = [...]float32{-357, -227, -97, 33, 163, 293}

// newBattle puts the player at the start and the enemies across the north.
func (g *Game) newBattle(level int) {
	clear(g.enemies[:])
	clear(g.shells[:])
	clear(g.puffs[:])
	g.x, g.y, g.heading, g.speed, g.throttle = field.StartX, field.StartY, 0, 0, throttleStop
	g.turret, g.elev, g.zoomed = 0, 0, false
	g.reload, g.ammo, g.shake, g.hurt, g.recoil, g.hitTicks, g.message = 0, startAmmo, 0, 0, 0, 0, msgNone
	// The first battle is a duel with a medium tank; from the second on, three
	// enemies and one more each time.
	n := 1
	if level > 1 {
		n = min(level+1, maxEnemies)
	}
	heavies := min((level-1)/2, 2)
	mediums := min(max(level-1, 1), n-heavies)
	for i := range n {
		k := light
		switch {
		case i < heavies:
			k = heavy
		case i < heavies+mediums:
			k = medium
		}
		lane := (i*5 + level) % len(lanes)
		e := &g.enemies[i]
		*e = tank{kind: k, x: lanes[lane] + g.spread()*8, y: field.EnemyMinY + 20 + float32(i/len(lanes))*80 + g.spread()*10,
			heading: math.Pi, turret: math.Pi, alive: true, until: 60 + g.random(240), lastSeen: -1000}
	}
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	countDown(&g.messageTicks)
	countDown(&g.shake)
	countDown(&g.hurt)
	countDown(&g.recoil)
	countDown(&g.hitTicks)
	switch g.mode {
	case modeTitle:
		// The attract mode sweeps the turret and zooms now and then.
		g.turret = angleDiff(g.turret+0.0015, 0)
		g.elev = 0.012
		g.zoomed = g.modeTicks/360%2 == 1
		g.updateEnemies()
		if startPressed() {
			g.level, g.kills, g.hp = 1, 0, 100
			g.newBattle(1)
			g.setMode(modePlay)
		}
	case modePlay:
		g.control()
		g.updatePlayer()
		g.updateEnemies()
		if g.hp <= 0 {
			g.hp = 0
			g.burst(fire, g.x, g.y, 2)
			g.setMode(modeOver)
		} else if g.enemiesLeft() == 0 {
			g.setMode(modeClear)
		}
	case modeClear:
		g.throttle = throttleStop
		g.updatePlayer()
		g.updateEnemies()
		if g.modeTicks > 90 && startPressed() {
			g.level++
			g.hp = min(g.hp+hitDamage, 100) // repaired between battles
			g.newBattle(g.level)
			g.setMode(modePlay)
		}
	case modeOver:
		g.updateEnemies()
		if g.modeTicks > 90 && startPressed() {
			g.newBattle(2)
			g.setMode(modeTitle)
		}
	}
	g.updateShells()
	g.updatePuffs()
	return nil
}

func countDown(n *int) {
	if *n > 0 {
		*n--
	}
}

func (g *Game) enemiesLeft() int {
	n := 0
	for i := range g.enemies {
		if g.enemies[i].alive {
			n++
		}
	}
	return n
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

// turnToward turns a toward b by at most rate.
func turnToward(a, b, rate float32) float32 {
	return angleDiff(a+min(max(angleDiff(b, a), -rate), rate), 0)
}

func sincos(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func hypot(x, y float32) float32 { return float32(math.Sqrt(float64(x*x + y*y))) }

// bearingTo returns the bearing of a vector.
func bearingTo(x, y float32) float32 { return float32(math.Atan2(float64(x), float64(-y))) }

// control reads the stick and buttons. The stick aims; with B held it drives.
// A tap of B, or Enter, toggles the zoom.
func (g *Game) control() {
	left, right := ebiten.IsKeyPressed(ebiten.KeyArrowLeft), ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	up, down := ebiten.IsKeyPressed(ebiten.KeyArrowUp), ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		g.bHeld++
		g.bUsed = g.bUsed || left || right || up || down
		g.turnHeld, g.elevHeld = 0, 0
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			g.throttle = min(g.throttle+1, len(throttles)-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			g.throttle = max(g.throttle-1, 0)
		}
		// The turret is stabilized: it keeps its bearing as the hull turns.
		rate := specs[player].turn
		if left {
			g.heading = angleDiff(g.heading-rate, 0)
		}
		if right {
			g.heading = angleDiff(g.heading+rate, 0)
		}
	} else {
		if g.bHeld > 0 && g.bHeld <= tapTicks && !g.bUsed {
			g.zoomed = !g.zoomed
		}
		g.bHeld, g.bUsed = 0, false
		g.aim(left, right, up, down)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.zoomed = !g.zoomed
	}
	if g.reload > 0 {
		g.reload--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.fire()
	}
}

// aim traverses the turret and elevates the gun, slowly at first for fine
// aim and faster while held.
func (g *Game) aim(left, right, up, down bool) {
	turn := 0
	if left {
		turn--
	}
	if right {
		turn++
	}
	if turn == 0 {
		g.turnHeld = 0
	} else {
		g.turnHeld++
		rate := (0.03 + float32(min(g.turnHeld, 50))/50*0.67) * math.Pi / 180
		if g.zoomed {
			rate /= 3
		}
		g.turret = angleDiff(g.turret+float32(turn)*rate, 0)
	}
	lift := 0
	if up {
		lift++
	}
	if down {
		lift--
	}
	if lift == 0 {
		g.elevHeld = 0
		return
	}
	// In pixels of the zoomed sight per tick.
	g.elevHeld++
	rate := (0.2 + float32(min(g.elevHeld, 40))/40*1.3) / focalZoom
	g.elev = min(max(g.elev+float32(lift)*rate, minElev), maxElev)
}

// dropAngle is how far a shell falls below the gun's axis at a range, as an
// angle: the sight's range marks sit this far below its center.
func dropAngle(r float32) float32 { return gravity * r / (2 * shellSpeed * shellSpeed) }

func (g *Game) fire() {
	if g.reload > 0 {
		return
	}
	if g.ammo == 0 {
		g.say(msgEmpty)
		return
	}
	g.ammo--
	g.reload = reloadTicks
	g.shoot(-1, g.x, g.y, g.turret, g.elev, eyeHeight, specs[player])
	g.recoil, g.shake = 12, 6
}

// shoot fires a shell from a tank at x, y along a bearing and elevation.
func (g *Game) shoot(from int, x, y, bearing, elev, height float32, s spec) {
	for i := range g.shells {
		sh := &g.shells[i]
		if sh.alive {
			continue
		}
		sb, cb := sincos(bearing)
		se, ce := sincos(elev)
		muzzle := s.length * 0.6
		*sh = shell{x: x + sb*muzzle, y: y - cb*muzzle, h: height,
			vx: sb * ce * shellSpeed, vy: -cb * ce * shellSpeed, vh: se * shellSpeed,
			from: from, pen: s.pen, penLoss: s.penLoss, alive: true}
		if from >= 0 {
			// The player's own smoke would fill the sight; it shows as recoil.
			g.puff(smoke, sh.x, sh.y, height)
			g.puff(flash, sh.x, sh.y, height)
		}
		return
	}
}

func (g *Game) updatePlayer() {
	s := specs[player]
	want := throttles[g.throttle] * s.speed / throttles[len(throttles)-1]
	g.speed += min(max(want-g.speed, -accel), accel)
	sh, ch := sincos(g.heading)
	nx, ny := g.x+sh*g.speed, g.y-ch*g.speed
	if g.blocked(nx, ny, -1) {
		if g.speed > 0.1 || g.speed < -0.05 {
			g.shake = 4
		}
		g.speed = 0
		return
	}
	g.x, g.y = nx, ny
}

// blocked reports whether a tank other than self (-1 for the player) would hit
// something at x, y.
func (g *Game) blocked(x, y float32, self int) bool {
	if x < -bound || x > bound || y < -bound || y > bound {
		return true
	}
	cx, cy := field.CellOf(x, y)
	for j := cy - 1; j <= cy+1; j++ {
		for i := cx - 1; i <= cx+1; i++ {
			for _, o := range field.InCell(i, j) {
				dx, dy, r := x-o.X, y-o.Y, o.Radius+tankRadius
				if dx*dx+dy*dy < r*r {
					return true
				}
			}
		}
	}
	const r = 2 * tankRadius
	if self >= 0 {
		if dx, dy := x-g.x, y-g.y; dx*dx+dy*dy < r*r {
			return true
		}
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		if i == self || (!e.alive && e.dead == 0) {
			continue
		}
		if dx, dy := x-e.x, y-e.y; dx*dx+dy*dy < r*r {
			return true
		}
	}
	return false
}

// clear reports whether nothing blocks the line between two points.
func (g *Game) clear(ax, ay, bx, by float32) bool {
	x0, x1 := min(ax, bx), max(ax, bx)
	y0, y1 := min(ay, by), max(ay, by)
	dx, dy := bx-ax, by-ay
	length2 := dx*dx + dy*dy
	for _, o := range field.Obstacles {
		r := o.Radius * 0.9
		if o.X+r < x0 || o.X-r > x1 || o.Y+r < y0 || o.Y-r > y1 {
			continue
		}
		t := min(max(((o.X-ax)*dx+(o.Y-ay)*dy)/length2, 0), 1)
		px, py := ax+t*dx-o.X, ay+t*dy-o.Y
		if px*px+py*py < r*r {
			return false
		}
	}
	return true
}

func (g *Game) updateEnemies() {
	for i := range g.enemies {
		e := &g.enemies[i]
		switch {
		case e.alive:
			g.updateEnemy(i, e)
		case e.dead > 0:
			e.dead++
			if e.dead%24 == 0 && e.dead < 60*40 {
				g.puff(smoke, e.x+g.spread(), e.y+g.spread(), specs[e.kind].height)
			}
		}
		if e.flash > 0 {
			e.flash--
		}
	}
}

func (g *Game) updateEnemy(i int, e *tank) {
	s := specs[e.kind]
	dx, dy := g.x-e.x, g.y-e.y
	dist := hypot(dx, dy)
	toPlayer := bearingTo(dx, dy)
	// Look for the player now and then; tanks take turns.
	if (g.ticks+i*3)%12 == 0 {
		was := e.sees
		e.sees = g.mode != modeOver && dist < spotRange && g.clear(e.x, e.y, g.x, g.y)
		if e.sees {
			if !was {
				e.reload = max(e.reload, 90+g.random(90)) // time to lay the gun
				e.errRange, e.errSide = 40+dist*0.05, 4+dist*0.006
				e.shotX, e.shotY = g.x, g.y
			}
			e.lastSeen, e.seenX, e.seenY = g.ticks, g.x, g.y
		}
	}
	// The turret follows the player while seen and for a while after.
	aim := e.heading
	if g.ticks-e.lastSeen < 300 {
		aim = bearingTo(e.seenX-e.x, e.seenY-e.y)
		if e.sees {
			aim = toPlayer
		}
	}
	e.turret = turnToward(e.turret, aim, s.traverse)
	e.stateTicks++
	want := float32(0)
	switch e.state {
	case holding:
		if e.kind == heavy && e.sees {
			e.heading = turnToward(e.heading, toPlayer, s.turn) // face the thick front armor
		}
		if e.stateTicks > e.until {
			g.pickWaypoint(i, e)
			e.state, e.stateTicks = driving, 0
		}
	case driving:
		if (g.ticks+i*5)%15 == 0 {
			g.steer(i, e)
		}
		wx, wy := e.wpX-e.x, e.wpY-e.y
		goal := angleDiff(bearingTo(wx, wy)+e.steer, 0)
		e.heading = turnToward(e.heading, goal, s.turn)
		want = s.speed
		if d := angleDiff(goal, e.heading); d > 0.4 || d < -0.4 {
			want = s.speed * 0.3
		}
		if wx*wx+wy*wy < 15*15 || e.stateTicks > 1500 || e.blocked > 90 {
			e.state, e.stateTicks, e.blocked = holding, 0, 0
			e.until = 200 + g.random(300)
		}
	}
	e.speed += min(max(want-e.speed, -accel), accel)
	sh, ch := sincos(e.heading)
	if nx, ny := e.x+sh*e.speed, e.y-ch*e.speed; g.blocked(nx, ny, i) {
		e.speed = 0
		e.blocked++
	} else {
		e.x, e.y = nx, ny
	}
	if e.reload > 0 {
		e.reload--
	}
	moving := e.speed > 0.05 && e.kind != light // lights fire on the move
	if g.mode == modePlay && e.sees && e.reload == 0 && dist < fireRange && !moving {
		if d := angleDiff(e.turret, toPlayer); d < 0.02 && d > -0.02 {
			g.enemyFire(i, e, dist)
		}
	}
}

// pickWaypoint chooses where to drive next: around the player at a fighting
// range.
func (g *Game) pickWaypoint(i int, e *tank) {
	spreadAngle, near, far := float32(1.1), float32(300), float32(700)
	// Lights, and the last enemy left, close in and go for the flanks.
	if e.kind == light || g.enemiesLeft() == 1 {
		spreadAngle, near, far = 1.7, 200, 500
	}
	from := bearingTo(e.x-g.x, e.y-g.y)
	for try := range 8 {
		a := from + g.spread()*spreadAngle
		r := near + float32(g.random(int(far-near)))
		s, c := sincos(a)
		x := min(max(g.x+s*r, -bound+10), bound-10)
		y := min(max(g.y-c*r, -bound+10), bound-10)
		if !g.blocked(x, y, i) || try == 7 {
			e.wpX, e.wpY, e.steer, e.blocked = x, y, 0, 0
			return
		}
	}
}

// steer looks 14 m ahead toward the waypoint and swerves around obstacles.
func (g *Game) steer(i int, e *tank) {
	goal := bearingTo(e.wpX-e.x, e.wpY-e.y)
	for _, off := range [...]float32{0, 0.4, -0.4, 0.8, -0.8, 1.3, -1.3, 2} {
		s, c := sincos(goal + off)
		if !g.blocked(e.x+s*14, e.y-c*14, i) && !g.blocked(e.x+s*7, e.y-c*7, i) {
			e.steer = off
			return
		}
	}
	e.steer = 2.5
}

func (g *Game) enemyFire(i int, e *tank, dist float32) {
	s := specs[e.kind]
	// A player who has moved spoils the enemy's ranging.
	if moved := hypot(g.x-e.shotX, g.y-e.shotY); moved > 30 {
		e.errRange = min(e.errRange*1.7, 40+dist*0.05)
		e.errSide = min(e.errSide*1.7, 4+dist*0.006)
	}
	e.shotX, e.shotY = g.x, g.y
	// Lead the player a little.
	sh, ch := sincos(g.heading)
	t := dist / shellSpeed * 0.7
	px, py := g.x+sh*g.speed*t, g.y-ch*g.speed*t
	ux, uy := (px-e.x)/dist, (py-e.y)/dist
	r, w := g.spread()*e.errRange, g.spread()*e.errSide
	ax, ay := px+ux*r-uy*w, py+uy*r+ux*w
	d := hypot(ax-e.x, ay-e.y)
	muzzle := s.height * 0.8
	elev := (aimHeight-muzzle)/d + dropAngle(d)
	g.shoot(i, e.x, e.y, bearingTo(ax-e.x, ay-e.y), elev, muzzle, s)
	e.errRange = max(e.errRange*0.6, 4)
	e.errSide = max(e.errSide*0.6, 0.8)
	e.reload = s.reload + g.random(90)
	e.flash = 6
}

// body is a tank's box for hits.
type body struct {
	x, y, heading float32
	s             *spec
}

// hitBox returns where the segment from p by d first enters a tank's box, as
// a fraction of d, with the face it enters and the cosine of the angle
// between the shell's path and that face's normal.
func hitBox(b body, px, py, ph, dx, dy, dh float32) (t float32, face int, cos float32, ok bool) {
	fs, fc := sincos(b.heading)
	fx, fy := fs, -fc // forward
	rx, ry := fc, fs  // right
	ox, oy := px-b.x, py-b.y
	p := [3]float32{ox*fx + oy*fy, ox*rx + oy*ry, ph}
	d := [3]float32{dx*fx + dy*fy, dx*rx + dy*ry, dh}
	lo := [3]float32{-b.s.length / 2, -b.s.width / 2, 0}
	hi := [3]float32{b.s.length / 2, b.s.width / 2, b.s.height}
	tmin, tmax, axis := float32(0), float32(1), -1
	for k := range 3 {
		if d[k] == 0 {
			if p[k] < lo[k] || p[k] > hi[k] {
				return 0, 0, 0, false
			}
			continue
		}
		t0, t1 := (lo[k]-p[k])/d[k], (hi[k]-p[k])/d[k]
		if t0 > t1 {
			t0, t1 = t1, t0
		}
		if t0 > tmin {
			tmin, axis = t0, k
		}
		tmax = min(tmax, t1)
		if tmin > tmax {
			return 0, 0, 0, false
		}
	}
	flat := hypot(d[0], d[1])
	switch axis {
	case 0:
		face, cos = rear, d[0]/flat
		if d[0] < 0 {
			face, cos = front, -d[0]/flat
		}
	case 2:
		face, cos = top, 1
	default:
		face, cos = side, d[1]/flat
		if cos < 0 {
			cos = -cos
		}
	}
	return tmin, face, cos, true
}

// leftSide reports whether a shell along dx, dy struck the left of a hull.
func leftSide(b body, dx, dy float32) bool {
	fs, fc := sincos(b.heading)
	return dx*fc+dy*fs > 0 // moving right in the hull's frame: it came from the left
}

func (g *Game) updateShells() {
	for i := range g.shells {
		if s := &g.shells[i]; s.alive {
			g.updateShell(s)
		}
	}
}

// Things a shell can hit.
const (
	hitNothing = iota
	hitGround
	hitObstacle
	hitTank
	hitPlayer
)

func (g *Game) updateShell(s *shell) {
	s.vh -= gravity
	dx, dy, dh := s.vx, s.vy, s.vh
	best, what, which, face, cos := float32(2), hitNothing, 0, 0, float32(0)
	if s.h+dh < 0 {
		best, what = s.h/(s.h-(s.h+dh)), hitGround
	}
	// Obstacles near the shell: it moves 10 m per tick, well within a cell.
	cx, cy := field.CellOf(s.x, s.y)
	for j := cy - 1; j <= cy+1; j++ {
		for k := cx - 1; k <= cx+1; k++ {
			for _, o := range field.InCell(k, j) {
				if t, ok := hitCylinder(o, s.x, s.y, s.h, dx, dy, dh); ok && t < best {
					best, what = t, hitObstacle
				}
			}
		}
	}
	for n := range g.enemies {
		e := &g.enemies[n]
		if n == s.from || (!e.alive && e.dead == 0) {
			continue
		}
		if t, f, c, ok := hitBox(body{e.x, e.y, e.heading, &specs[e.kind]}, s.x, s.y, s.h, dx, dy, dh); ok && t < best {
			best, what, which, face, cos = t, hitTank, n, f, c
		}
	}
	me := body{g.x, g.y, g.heading, &specs[player]}
	if s.from >= 0 && g.mode == modePlay {
		if t, f, c, ok := hitBox(me, s.x, s.y, s.h, dx, dy, dh); ok && t < best {
			best, what, face, cos = t, hitPlayer, f, c
		}
	}
	if what == hitNothing {
		s.x, s.y, s.h = s.x+dx, s.y+dy, s.h+dh
		s.run += shellSpeed
		if s.run > shellRange {
			s.alive = false
		}
		return
	}
	s.alive = false
	x, y, h := s.x+dx*best, s.y+dy*best, s.h+dh*best
	run := s.run + shellSpeed*best
	pen := s.pen - s.penLoss*run
	switch what {
	case hitGround:
		g.puff(dust, x, y, 0)
	case hitObstacle:
		g.puff(dust, x, y, h)
		g.puff(spark, x, y, h)
	case hitTank:
		e := &g.enemies[which]
		if !e.alive {
			g.puff(spark, x, y, h)
			if s.from < 0 {
				g.say(msgWreck)
			}
			return
		}
		result := penetrates(pen, specs[e.kind].armor[face], cos)
		if result != penetrated {
			g.puff(spark, x, y, h)
			if s.from < 0 && result == ricochet {
				g.say(msgRicochet)
			} else if s.from < 0 {
				g.say(msgNoPen)
			}
			return
		}
		e.alive, e.dead, e.speed = false, 1, 0
		g.burst(fire, e.x, e.y, specs[e.kind].height)
		if s.from < 0 {
			g.kills++
			g.say(msgKill)
		}
	case hitPlayer:
		var dir int
		switch {
		case face == front:
			dir = 0
		case face == rear:
			dir = 3
		case leftSide(me, dx, dy):
			dir = 1
		default:
			dir = 2
		}
		g.hitBearing, g.hitTicks = bearingTo(-dx, -dy), 90
		g.puff(spark, x, y, h)
		if penetrates(pen, specs[player].armor[face], cos) == penetrated {
			g.hp -= hitDamage
			g.hurt, g.shake = 30, 20
			g.say(msgHitFront + dir)
		} else {
			g.shake = 8
			g.say(msgHeldFront + dir)
		}
	}
}

// Results of a shell striking armor.
const (
	penetrated = iota
	noPenetration
	ricochet
)

// penetrates compares a shell's penetration with the armor along its path:
// the more oblique the hit, the thicker the armor, and past about 72 degrees
// the shell glances off.
func penetrates(pen, armor, cos float32) int {
	if cos < 0.3 {
		return ricochet
	}
	if pen >= armor/cos {
		return penetrated
	}
	return noPenetration
}

// hitCylinder returns where a segment first enters an obstacle below its top.
func hitCylinder(o field.Obstacle, px, py, ph, dx, dy, dh float32) (float32, bool) {
	ox, oy := px-o.X, py-o.Y
	a := dx*dx + dy*dy
	b := ox*dx + oy*dy
	c := ox*ox + oy*oy - o.Radius*o.Radius
	disc := b*b - a*c
	if disc < 0 {
		return 0, false
	}
	root := float32(math.Sqrt(float64(disc)))
	t0, t1 := (-b-root)/a, (-b+root)/a
	t0, t1 = max(t0, 0), min(t1, 1)
	if t0 > t1 {
		return 0, false
	}
	// The height is linear along the segment: find where it is below the top.
	h0, h1 := ph+dh*t0, ph+dh*t1
	switch {
	case h0 < o.Height:
		return t0, true
	case h1 < o.Height:
		return t0 + (t1-t0)*(h0-o.Height)/(h0-h1), true
	}
	return 0, false
}

func (g *Game) puff(k puffKind, x, y, h float32) {
	for i := range g.puffs {
		if p := &g.puffs[i]; !p.alive {
			*p = puff{x: x, y: y, h: h, kind: k, alive: true, life: puffSpecs[k].life, sizeScale: 1}
			return
		}
	}
}

// burst is a tank blowing up: fire and smoke.
func (g *Game) burst(k puffKind, x, y, h float32) {
	for n := range 3 {
		g.puff(k, x+g.spread()*1.5, y+g.spread()*1.5, h*float32(n)/2)
	}
	g.puff(smoke, x, y, h)
	g.puff(flash, x, y, h)
}

func (g *Game) updatePuffs() {
	for i := range g.puffs {
		p := &g.puffs[i]
		if !p.alive {
			continue
		}
		p.age++
		p.h += puffSpecs[p.kind].rise
		if p.age >= p.life {
			p.alive = false
		}
	}
}
