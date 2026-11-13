// Package air is an original air combat game seen from the cockpit of a
// carrier plane. Choose one of three: fly the Type 0 fighter over your fleet
// and shoot down the attacking bombers and their escorts; dive on an enemy
// fleet in the Type 99 dive bomber, dive brakes out, and pull out after the
// release; or come in low and slow in the Type 97 attack plane and aim a
// torpedo ahead of a turning ship. Every image is drawn by code, and it uses
// only the Ebitengine API, so it also builds with Ebitengine.
package air

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120

	maxPlanes  = 10
	maxShips   = 6
	maxBullets = 128
	maxPuffs   = 56
	maxBursts  = 16
	sorties    = 3
	waves      = 3

	bulletSpeed = 750 // m/s, added to the shooter's speed
	cannonSpeed = 600
	bulletLife  = 80  // ticks
	shellSpeed  = 700 // m/s, for the flak's time of flight
	torpSpeed   = 35  // m/s: faster than the real 42 knots, so that a run takes seconds, not a minute
	torpArm     = 150 // meters a torpedo runs before it can explode
	torpRange   = 4000
	torpMaxAlt  = 60
	torpMaxKmh  = 300
	fleetSpeed  = 12
	leaveRange  = 13000 // enemy planes this far from the fleet are gone
	sortieRange = 4500  // after the release, this far from the fleet ends the sortie
	sortieTime  = 40 * tps
	messageTime = 120
)

type mission uint8

const (
	missionFighter mission = iota // defend the fleet in the Zero
	missionDive                   // dive bomb the enemy fleet
	missionTorpedo                // torpedo the enemy fleet
	missions
)

var missionPlane = [missions]kind{zero, val, kate}

type mode uint8

const (
	modeSelect mode = iota
	modePlay
	modeDebrief // between sorties
	modeResult
)

type shipKind uint8

const (
	carrier shipKind = iota
	cruiser
	destroyer
	shipKinds
)

// shipTypes holds sizes in meters, hit points, points for a hit and for a
// sinking, the heavy and light guns, and the turn rate in rad/s.
var shipTypes = [shipKinds]struct {
	length, beam, freeboard, height float32
	hp, hitScore, sinkScore         int
	heavy, light                    int
	turn                            float32
}{
	carrier:   {250, 30, 14, 30, 6, 300, 1000, 2, 3, 0.012},
	cruiser:   {180, 18, 9, 26, 4, 200, 600, 3, 3, 0.016},
	destroyer: {110, 11, 6, 16, 2, 150, 400, 2, 2, 0.025},
}

type ship struct {
	kind              shipKind
	alive, friendly   bool
	pos               vec
	bearing, want     float32
	speed             float32
	hp, sink          int // sink counts ticks since it started to go down
	zig               int
	heavyCool, aaCool int
	fire              int // ticks of fire after a hit
}

func (s *ship) dir() vec { return heading(s.bearing) }

// inside reports whether a point on the sea is within the hull.
func (s *ship) inside(p vec, margin float32) bool {
	d := p.sub(s.pos).flat()
	h := s.dir()
	along, across := d.dot(h), d.x*h.y-d.y*h.x
	t := &shipTypes[s.kind]
	return abs(along) < t.length/2+margin && abs(across) < t.beam/2+margin
}

type bullet struct {
	pos, vel vec
	life     int
	enemy    bool // fired at the player
	dmg      int
}

type puffKind uint8

const (
	puffFlak puffKind = iota
	puffSmoke
	puffFire
	puffSplash
)

type puff struct {
	pos  vec
	kind puffKind
	age  int
	size float32 // meters
	life int
}

// burst is a flak shell on its way to the point where it will explode.
type burst struct {
	pos  vec
	time int
}

type bomb struct {
	pos, vel vec
	alive    bool
}

type torpedo struct {
	pos, dir vec
	run      float32
	alive    bool
}

const (
	msgNone = iota
	msgKill
	msgHit
	msgSunk
	msgMiss
	msgBroke
	msgOverspeed
	msgPullUp
	msgWave
	msgCarrierHit
	msgRelease
	msgNoAmmo
	msgStall
)

var messages = [...]string{"", "SHOT DOWN!", "HIT!", "SUNK!", "MISSED", "TORPEDO BROKE", "OVERSPEED", "PULL UP!",
	"ENEMY RAID", "CARRIER HIT", "AWAY!", "NO AMMO", "STALL"}

type Game struct {
	mode             mode
	ticks, modeTicks int
	clock            int // ticks of play, for the enemy pilots
	rng              uint32
	cursor           int
	climbUp          bool // the stick pushed up climbs, instead of diving as in a cockpit

	mission      mission
	sortie, wave int
	score        int
	outcome      string
	lastResult   string // what the sortie's weapon did

	planes  [maxPlanes]plane // the player is planes[0]
	ships   [maxShips]ship
	bullets [maxBullets]bullet
	puffs   [maxPuffs]puff
	bursts  [maxBursts]burst
	bomb    bomb
	torp    torpedo

	weapons     int // bombs or torpedoes left in this sortie
	lost        int // bombers lost
	ammo        int // 7.7 mm rounds
	cannon      int // 20 mm rounds
	gunCool     int
	cannonCool  int
	released    int // ticks since the weapon left the plane, or 0
	gunnerCool  int
	overspeed   int
	waveTicks   int
	dead        int // ticks since the player was lost
	target      int // the ship the markers point to, or -1
	enemy       int // the enemy plane the markers point to, or -1
	message     int
	messageTime int
	shake       int

	view view
}

func NewGame() *Game {
	g := &Game{rng: 2463534242, target: -1, enemy: -1}
	g.start(missionFighter)
	g.mode = modeSelect
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

func (g *Game) say(m int) { g.message, g.messageTime = m, messageTime }

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

func (g *Game) player() *plane { return &g.planes[0] }

// start sets up a mission from the beginning.
func (g *Game) start(m mission) {
	g.mission, g.score, g.sortie, g.wave = m, 0, 0, 0
	g.outcome, g.lastResult, g.lost = "", "", 0
	clear(g.planes[:])
	clear(g.ships[:])
	g.clearShots()
	if m == missionFighter {
		g.ships[0] = ship{kind: carrier, friendly: true, alive: true, bearing: 0, want: 0, speed: fleetSpeed, hp: 3}
		g.ships[1] = ship{kind: destroyer, friendly: true, alive: true, pos: vec{-900, 600, 0}, speed: fleetSpeed, hp: 2}
		g.ships[2] = ship{kind: destroyer, friendly: true, alive: true, pos: vec{900, 300, 0}, speed: fleetSpeed, hp: 2}
		g.planes[0] = newPlane(zero, vec{0, -1500, 1800}, 0, 120)
		g.ammo, g.cannon = 1400, 120
		g.wave, g.waveTicks = 0, 3*tps
	} else {
		b := float32(g.random(628)) / 100
		g.ships[0] = ship{kind: carrier, alive: true, bearing: b, want: b, speed: fleetSpeed}
		g.ships[1] = ship{kind: cruiser, alive: true, pos: heading(b).mul(1400), bearing: b, want: b, speed: fleetSpeed}
		side := heading(b + 1.5708)
		g.ships[2] = ship{kind: destroyer, alive: true, pos: side.mul(1200).add(heading(b).mul(400)), bearing: b, want: b, speed: fleetSpeed}
		g.ships[3] = ship{kind: destroyer, alive: true, pos: side.mul(-1200).add(heading(b).mul(400)), bearing: b, want: b, speed: fleetSpeed}
		for i := range 4 {
			g.ships[i].hp = shipTypes[g.ships[i].kind].hp
			g.ships[i].zig = 600 + g.random(900)
		}
		g.startSortie()
	}
}

func (g *Game) clearShots() {
	clear(g.bullets[:])
	clear(g.puffs[:])
	clear(g.bursts[:])
	g.bomb, g.torp = bomb{}, torpedo{}
	g.released, g.overspeed, g.dead, g.shake = 0, 0, 0, 0
	g.message, g.messageTime = msgNone, 0
}

// startSortie puts the bomber 8 km from the enemy carrier, heading in, with
// one bomb or torpedo, and the enemy's fighters over the fleet.
func (g *Game) startSortie() {
	g.sortie++
	g.clearShots()
	for i := 1; i < maxPlanes; i++ {
		g.planes[i] = plane{}
	}
	g.weapons = 1
	c := g.ships[0].pos
	from := float32(3.6 + 0.9*float32(g.sortie))
	alt, k := float32(3000), val
	if g.mission == missionTorpedo {
		alt, k = 800, kate
	}
	start := c.add(heading(from).mul(8000))
	start.z = alt
	g.planes[0] = newPlane(k, start, bearingOf(c.sub(start)), 90)
	for i := range g.sortie {
		p := c.add(heading(from + 0.6*float32(i-1)).mul(1500))
		p.z = 1500 + float32(i)*400
		g.planes[1+i] = newPlane(fighter, p, from+3.14, 100)
		g.planes[1+i].task = taskPatrol
	}
	g.target = -1
	g.setMode(modePlay)
}

func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

func startPressed() bool {
	return pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace) || pressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	if g.messageTime > 0 {
		g.messageTime--
	}
	switch g.mode {
	case modeSelect:
		if pressed(ebiten.KeyArrowUp) {
			g.cursor = (g.cursor + int(missions) - 1) % int(missions)
		}
		if pressed(ebiten.KeyArrowDown) {
			g.cursor = (g.cursor + 1) % int(missions)
		}
		if pressed(ebiten.KeyX) {
			g.climbUp = !g.climbUp
		}
		// The select screen shows the chosen plane's sky turning slowly.
		p := g.player()
		p.f = heading(float32(g.ticks) * 0.002)
		p.u = vec{0, 0, 1}
		if startPressed() {
			g.start(mission(g.cursor))
			g.setMode(modePlay)
		}
	case modePlay:
		g.step(g.controls())
	case modeDebrief:
		if g.modeTicks > 45 && startPressed() {
			if g.sortie >= sorties || g.fleetSunk() {
				g.finish("MISSION OVER")
			} else {
				g.startSortie()
			}
		}
	case modeResult:
		if g.modeTicks > 60 && startPressed() {
			g.setMode(modeSelect)
		}
	}
	return nil
}

// input is what the player does in one tick.
type input struct {
	roll, pull   float32
	fire, second bool // held
	release      bool // pressed this tick
	toggle       bool
}

func (g *Game) controls() input {
	var in input
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		in.roll--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		in.roll++
	}
	// As with a control stick, pulling back (down) raises the nose.
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		in.pull++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		in.pull--
	}
	if g.climbUp {
		in.pull = -in.pull
	}
	in.fire = ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeySpace)
	in.second = ebiten.IsKeyPressed(ebiten.KeyX)
	in.release = pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace)
	in.toggle = pressed(ebiten.KeyX)
	return in
}

// step advances play one tick.
func (g *Game) step(in input) {
	g.clock++
	if g.shake > 0 {
		g.shake--
	}
	p := g.player()
	if p.alive && p.falling == 0 {
		g.pilot(in)
	} else if p.alive {
		p.fly(0.5, -0.2)
	}
	for i := 1; i < maxPlanes; i++ {
		if g.planes[i].alive {
			g.think(i)
		}
	}
	g.updateShips()
	g.updateBullets()
	g.updateWeapons()
	g.updateFlak()
	g.updatePuffs()
	g.updatePlayer()
	g.updateMission()
	g.pickTargets()
}

// pilot flies the player's plane and works its weapons.
func (g *Game) pilot(in input) {
	p := g.player()
	switch p.kind {
	case zero:
		if g.gunCool > 0 {
			g.gunCool--
		}
		if g.cannonCool > 0 {
			g.cannonCool--
		}
		if in.fire && g.gunCool == 0 {
			if g.ammo > 0 {
				g.ammo -= 2
				g.shoot(p, p.f, bulletSpeed, 1, 0.004, false)
				g.gunCool = 4
			} else if pressed(ebiten.KeyZ) {
				g.say(msgNoAmmo)
			}
		}
		if in.second && g.cannonCool == 0 && g.cannon > 0 {
			g.cannon -= 2
			g.shoot(p, p.f, cannonSpeed, 5, 0.006, false)
			g.cannonCool = 10
		}
	case val:
		if in.toggle {
			p.brakes = !p.brakes
		}
		p.throttle = 1
		if p.brakes {
			p.throttle = 0.25
		}
		if in.release {
			g.dropBomb()
		}
	case kate:
		if in.toggle {
			if p.throttle == 1 {
				p.throttle = 0.45
			} else {
				p.throttle = 1
			}
		}
		if in.release {
			g.dropTorpedo()
		}
	}
	p.fly(in.roll, in.pull)
	if p.kind != zero {
		g.gunner(p, true)
	}
}

// shoot fires a round from a plane, or from a ship when from is nil, along
// dir with a random spread in radians.
func (g *Game) shoot(from *plane, dir vec, speed float32, dmg int, spread float32, enemy bool) {
	g.fire(from.pos.add(dir.mul(8)), dir, from.f.mul(from.v), speed, dmg, spread, enemy)
}

func (g *Game) fire(pos, dir, carried vec, speed float32, dmg int, spread float32, enemy bool) {
	for i := range g.bullets {
		b := &g.bullets[i]
		if b.life > 0 {
			continue
		}
		d := dir.add(vec{g.spread(), g.spread(), g.spread()}.mul(spread)).norm()
		*b = bullet{pos: pos, vel: d.mul(speed).add(carried), life: bulletLife, enemy: enemy, dmg: dmg}
		return
	}
}

// updateBullets moves the rounds and checks each tick's path against the
// planes it could hit.
func (g *Game) updateBullets() {
	for i := range g.bullets {
		b := &g.bullets[i]
		if b.life <= 0 {
			continue
		}
		b.life--
		from := b.pos
		b.vel.z -= gravity * dt
		b.pos = b.pos.add(b.vel.mul(dt))
		if b.pos.z < 0 {
			b.life = 0
			continue
		}
		for j := range g.planes {
			q := &g.planes[j]
			if !q.alive || q.falling > 0 || (j == 0) != b.enemy {
				continue
			}
			radius := planeTypes[q.kind].span * 0.4
			if j == 0 {
				radius = 5
			}
			if segmentNear(from, b.pos, q.pos, radius) {
				b.life = 0
				g.damage(j, b.dmg)
				break
			}
		}
	}
}

// segmentNear reports whether the segment from a to b passes within r of p.
func segmentNear(a, b, p vec, r float32) bool {
	ab := b.sub(a)
	t := clamp(p.sub(a).dot(ab)/max(ab.dot(ab), 1e-6), 0, 1)
	d := a.add(ab.mul(t)).sub(p)
	return d.dot(d) < r*r
}

// damage hits plane i; the player's plane counts hit points, the enemy's fall.
func (g *Game) damage(i, n int) {
	q := &g.planes[i]
	if q.falling > 0 {
		return
	}
	q.hp -= n
	if i == 0 {
		g.shake = 8
		if q.hp <= 0 && q.falling == 0 {
			q.falling = 1
			g.outcome = "SHOT DOWN"
		}
		return
	}
	g.addPuff(q.pos, puffSmoke, 4, 30)
	if q.hp <= 0 && q.falling == 0 {
		q.falling = 1
		if q.kind == bomber {
			g.score += 150
		} else {
			g.score += 100
		}
		g.say(msgKill)
	}
}

func (g *Game) addPuff(pos vec, k puffKind, size float32, life int) {
	oldest := 0
	for i := range g.puffs {
		if g.puffs[i].life == 0 {
			oldest = i
			break
		}
		if g.puffs[i].age > g.puffs[oldest].age {
			oldest = i
		}
	}
	g.puffs[oldest] = puff{pos: pos, kind: k, size: size, life: life}
}

func (g *Game) updatePuffs() {
	for i := range g.puffs {
		p := &g.puffs[i]
		if p.life == 0 {
			continue
		}
		p.age++
		if p.age >= p.life {
			p.life = 0
		}
		if p.kind == puffSmoke || p.kind == puffFlak {
			p.size += 0.05
		}
	}
}

// dropBomb releases the dive bomber's bomb with the plane's velocity.
func (g *Game) dropBomb() {
	p := g.player()
	if g.weapons == 0 {
		return
	}
	g.weapons--
	g.bomb = bomb{pos: p.pos.sub(p.u.mul(2)), vel: p.f.mul(p.v), alive: true}
	g.released = 1
	g.say(msgRelease)
}

// torpedoOK reports whether a torpedo dropped now would run: the plane must be
// low, slow and level.
func (g *Game) torpedoOK() bool {
	p := g.player()
	return p.pos.z <= torpMaxAlt && p.v*3.6 <= torpMaxKmh && abs(p.f.z) < 0.1 && abs(p.bank()) < 0.35
}

func (g *Game) dropTorpedo() {
	p := g.player()
	if g.weapons == 0 {
		return
	}
	g.weapons--
	g.released = 1
	if !g.torpedoOK() {
		g.lastResult = "TORPEDO BROKE"
		g.say(msgBroke)
		g.addPuff(p.pos.flat().add(p.f.flat().norm().mul(80)), puffSplash, 12, 40)
		return
	}
	dir := p.f.flat().norm()
	g.torp = torpedo{pos: p.pos.flat().add(dir.mul(p.pos.z * 1.5)), dir: dir, alive: true}
	g.addPuff(g.torp.pos, puffSplash, 10, 40)
	g.say(msgRelease)
}

// bombFall is how long a bomb dropped now takes to reach the sea.
func (g *Game) bombFall() float32 {
	p := g.player()
	vz := p.f.z * p.v
	return (vz + sqrt(vz*vz+2*gravity*max(p.pos.z, 0))) / gravity
}

// bombImpact predicts where a bomb dropped now hits the sea.
func (g *Game) bombImpact() vec {
	p := g.player()
	return p.pos.add(p.f.mul(p.v * g.bombFall())).flat()
}

func (g *Game) updateWeapons() {
	if g.released > 0 {
		g.released++
	}
	if b := &g.bomb; b.alive {
		b.vel.z -= gravity * dt
		b.pos = b.pos.add(b.vel.mul(dt))
		if b.pos.z <= 0 {
			b.alive = false
			b.pos.z = 0
			if i := g.shipAt(b.pos, 0); i >= 0 {
				g.hitShip(i, b.pos, 2)
			} else {
				g.addPuff(b.pos, puffSplash, 30, 60)
				if g.shipAt(b.pos, 25) >= 0 {
					g.lastResult = "NEAR MISS"
				} else {
					g.lastResult = "MISSED"
				}
				g.say(msgMiss)
			}
		}
	}
	if t := &g.torp; t.alive {
		step := float32(torpSpeed * dt)
		t.pos = t.pos.add(t.dir.mul(step))
		t.run += step
		if t.run > torpArm {
			if i := g.shipAt(t.pos, 2); i >= 0 {
				t.alive = false
				g.hitShip(i, t.pos, 3)
			}
		}
		if t.alive && t.run > torpRange {
			t.alive = false
			g.lastResult = "MISSED"
			g.say(msgMiss)
		}
	}
}

// shipAt returns the enemy ship whose hull is at p, or -1.
func (g *Game) shipAt(p vec, margin float32) int {
	for i := range g.ships {
		s := &g.ships[i]
		if s.alive && !s.friendly && s.sink == 0 && s.inside(p, margin) {
			return i
		}
	}
	return -1
}

var (
	hitNames  = [shipKinds]string{"HIT CARRIER", "HIT CRUISER", "HIT DESTROYER"}
	sunkNames = [shipKinds]string{"SUNK CARRIER", "SUNK CRUISER", "SUNK DESTROYER"}
)

func (g *Game) hitShip(i int, at vec, dmg int) {
	s := &g.ships[i]
	t := &shipTypes[s.kind]
	s.hp -= dmg
	s.fire = 900
	g.score += t.hitScore
	g.addPuff(at, puffFire, 25, 60)
	g.addPuff(at.add(vec{0, 0, 20}), puffSmoke, 30, 120)
	g.lastResult = hitNames[s.kind]
	g.say(msgHit)
	if s.hp <= 0 {
		s.sink = 1
		g.score += t.sinkScore
		g.lastResult = sunkNames[s.kind]
		g.say(msgSunk)
	}
}

func (g *Game) fleetSunk() bool {
	for i := range g.ships {
		if s := &g.ships[i]; s.alive && !s.friendly && s.sink == 0 {
			return false
		}
	}
	return true
}

// updateShips steers the ships: they zigzag, and turn to comb a torpedo or
// away from a dive.
func (g *Game) updateShips() {
	p := g.player()
	for i := range g.ships {
		s := &g.ships[i]
		if !s.alive {
			continue
		}
		t := &shipTypes[s.kind]
		if s.sink > 0 {
			s.sink++
			s.speed = approach(s.speed, 0, 0.02)
			if s.sink > 900 {
				s.alive = false
			}
		} else if !s.friendly {
			s.zig--
			if s.zig <= 0 {
				s.want = g.ships[0].want + float32(g.random(3)-1)*0.6
				s.zig = 900 + g.random(900)
			}
			if g.torp.alive && g.torp.pos.sub(s.pos).length() < 600 {
				// Turn the bow or the stern toward the torpedo's track.
				tb := bearingOf(g.torp.dir)
				if abs(wrap(tb-s.bearing)) < 1.5708 {
					s.want = tb
				} else {
					s.want = tb + 3.14159
				}
			}
			if p.alive && p.kind == val && p.f.z < -0.5 && p.pos.flat().sub(s.pos).length() < 3000 {
				s.want = s.bearing + 1
			}
		}
		if s.fire > 0 {
			s.fire--
			if s.fire%20 == 0 {
				g.addPuff(s.pos.add(vec{0, 0, t.height}), puffSmoke, 20, 150)
			}
		}
		d := wrap(s.want - s.bearing)
		s.bearing = wrap(s.bearing + clamp(d, -t.turn*dt, t.turn*dt))
		s.pos = s.pos.add(s.dir().mul(s.speed * dt))
	}
}

// wrap brings an angle into -π to π.
func wrap(a float32) float32 {
	for a > 3.14159 {
		a -= 6.28318
	}
	for a < -3.14159 {
		a += 6.28318
	}
	return a
}

// updateFlak fires the enemy ships' guns: heavy shells burst near where the
// plane will be, light guns fire tracers at close range.
func (g *Game) updateFlak() {
	p := g.player()
	if !p.alive || p.falling > 0 || g.mission == missionFighter {
		return
	}
	for i := range g.ships {
		s := &g.ships[i]
		if !s.alive || s.sink > 0 || s.friendly {
			continue
		}
		t := &shipTypes[s.kind]
		gun := s.pos.add(vec{0, 0, 10})
		d := p.pos.sub(gun)
		dist := d.length()
		if s.heavyCool > 0 {
			s.heavyCool--
		}
		if s.heavyCool == 0 && dist < 6000 && p.pos.z > 150 {
			s.heavyCool = 240/t.heavy + g.random(60)
			tof := dist / shellSpeed
			miss := 40 + dist*0.02
			aim := p.pos.add(p.f.mul(p.v * tof)).add(vec{g.spread(), g.spread(), g.spread()}.mul(miss))
			g.addBurst(aim, int(tof*tps))
		}
		if s.aaCool > 0 {
			s.aaCool--
		}
		if s.aaCool == 0 && dist < 2200 {
			s.aaCool = 30/t.light + 6
			tof := dist / 900
			lead := p.pos.add(p.f.mul(p.v * tof)).sub(gun).norm()
			g.fire(gun, lead, vec{}, 900, 1, 0.02, true)
		}
	}
	for i := range g.bursts {
		b := &g.bursts[i]
		if b.time <= 0 {
			continue
		}
		b.time--
		if b.time == 0 {
			g.addPuff(b.pos, puffFlak, 14, 150)
			if d := b.pos.sub(p.pos).length(); d < 25 {
				g.damage(0, int((25-d)*0.6)+2)
			}
		}
	}
}

func (g *Game) addBurst(pos vec, time int) {
	for i := range g.bursts {
		if g.bursts[i].time <= 0 {
			g.bursts[i] = burst{pos: pos, time: max(time, 1)}
			return
		}
	}
}

// updatePlayer checks the sea, overspeed and the end of a lost plane.
func (g *Game) updatePlayer() {
	p := g.player()
	if !p.alive {
		g.dead++
		if g.dead > 2*tps {
			if g.outcome == "" {
				g.outcome = "LOST"
			}
			if g.mission != missionFighter {
				// The bomber is lost, but the next sortie flies.
				g.lastResult, g.outcome = g.outcome, ""
				g.lost++
				g.setMode(modeDebrief)
				return
			}
			g.finish(g.outcome)
		}
		return
	}
	if p.falling > 0 {
		p.falling++
		if p.falling%4 == 0 {
			g.addPuff(p.pos, puffSmoke, 6, 90)
		}
	}
	if p.pos.z <= 0 {
		p.alive = false
		p.pos.z = 0
		g.addPuff(p.pos, puffSplash, 30, 120)
		if g.outcome == "" {
			g.outcome = "CRASHED"
		}
		return
	}
	t := &planeTypes[p.kind]
	if p.v > t.vne {
		g.overspeed++
		if g.overspeed%8 == 0 {
			g.damage(0, 2)
		}
		if g.messageTime == 0 {
			g.say(msgOverspeed)
		}
	} else {
		g.overspeed = 0
	}
	// Warn when the sea is less than four seconds away.
	if p.f.z < 0 && p.pos.z < -p.f.z*p.v*4 && g.messageTime == 0 {
		g.say(msgPullUp)
	}
	if p.v < t.vstall*1.05 && g.messageTime == 0 {
		g.say(msgStall)
	}
}

// updateMission runs the raids of the fighter mission and the end of a
// bomber's sortie.
func (g *Game) updateMission() {
	p := g.player()
	if !p.alive || p.falling > 0 {
		return
	}
	if g.mission == missionFighter {
		if g.ships[0].hp <= 0 {
			g.finish("CARRIER LOST")
			return
		}
		if g.raidOver() {
			g.waveTicks--
			if g.waveTicks <= 0 {
				if g.wave == waves {
					g.score += 300 * g.ships[0].hp
					g.finish("FLEET DEFENDED")
					return
				}
				g.wave++
				g.spawnRaid()
				g.say(msgWave)
				if g.wave > 1 {
					// Between the raids the Zero lands, rearms and is repaired.
					g.ammo, g.cannon = 1400, 120
					p.hp = planeTypes[zero].hp
				}
				g.waveTicks = 5 * tps
			}
		}
		return
	}
	if g.released > 0 && !g.bomb.alive && !g.torp.alive {
		far := p.pos.flat().sub(g.ships[0].pos).length() > sortieRange
		if far || g.released > sortieTime || g.fleetSunk() {
			g.setMode(modeDebrief)
		}
	}
}

func (g *Game) raidOver() bool {
	for i := 1; i < maxPlanes; i++ {
		if g.planes[i].alive {
			return false
		}
	}
	return true
}

// spawnRaid sends bombers with escorts at the carrier from the north: two
// bombers and a fighter, then three and one, then four and two.
func (g *Game) spawnRaid() {
	c := g.ships[0].pos
	from := -0.8 + float32(g.random(160))/100
	bombers, fighters := 1+g.wave, (g.wave+1)/2
	n := 1
	for i := range bombers {
		p := c.add(heading(from).mul(11000)).add(heading(from + 1.5708).mul(float32(i-bombers/2) * 60))
		p.z = 3000 - float32(i%2)*40
		g.planes[n] = newPlane(bomber, p, from+3.14159, 100)
		g.planes[n].task, g.planes[n].goal = taskRun, c
		n++
	}
	for i := range fighters {
		p := c.add(heading(from).mul(11300)).add(heading(from + 1.5708).mul(float32(i*2-1) * 300))
		p.z = 3400
		g.planes[n] = newPlane(fighter, p, from+3.14159, 110)
		g.planes[n].task = taskEscort
		n++
	}
}

func (g *Game) finish(outcome string) {
	g.outcome = outcome
	g.setMode(modeResult)
}

// pickTargets chooses the ship and the enemy plane the markers point to:
// the ones closest to the nose.
func (g *Game) pickTargets() {
	p := g.player()
	g.target, g.enemy = -1, -1
	best := float32(-2)
	for i := range g.ships {
		s := &g.ships[i]
		if !s.alive || s.friendly || s.sink > 0 {
			continue
		}
		if c := s.pos.sub(p.pos).norm().dot(p.f); c > best {
			best, g.target = c, i
		}
	}
	best = -2
	for i := 1; i < maxPlanes; i++ {
		q := &g.planes[i]
		if !q.alive || q.falling > 0 {
			continue
		}
		d := q.pos.sub(p.pos)
		if c := d.norm().dot(p.f) - d.length()/20000; c > best {
			best, g.enemy = c, i
		}
	}
}

// lead returns where to aim a torpedo dropped here to meet ship i.
func (g *Game) lead(i int) (vec, bool) {
	s := &g.ships[i]
	p := g.player().pos.flat()
	v := s.dir().mul(s.speed)
	d := s.pos.sub(p)
	// |d + v t| = torpSpeed t
	a := v.dot(v) - torpSpeed*torpSpeed
	b := 2 * d.dot(v)
	c := d.dot(d)
	disc := b*b - 4*a*c
	if disc < 0 {
		return vec{}, false
	}
	t := (-b - sqrt(disc)) / (2 * a)
	if t < 0 {
		return vec{}, false
	}
	return s.pos.add(v.mul(t)), true
}
