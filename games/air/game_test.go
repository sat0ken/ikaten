package air

import (
	"math"
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

func playing(m mission) *Game {
	g := NewGame()
	g.start(m)
	g.setMode(modePlay)
	return g
}

// With the stick centered, a level plane keeps its height and heading, and
// settles at its top speed.
func TestLevelFlight(t *testing.T) {
	for k := range kinds {
		p := newPlane(k, vec{0, 0, 1000}, 0.5, 90)
		for range 60 * tps {
			p.fly(0, 0)
		}
		ty := &planeTypes[k]
		if abs(p.pos.z-1000) > 5 || abs(bearingOf(p.f)-0.5) > 0.01 || abs(p.v-ty.vmax) > 3 {
			t.Errorf("kind %d: altitude %.1f bearing %.3f speed %.1f", k, p.pos.z, bearingOf(p.f), p.v)
		}
	}
}

// Banked 77 degrees and pulling about 4 g to hold the height, the Zero turns
// a full circle in 12 to 25 seconds.
func TestTurn(t *testing.T) {
	for _, k := range []kind{zero, kate} {
		p := newPlane(k, vec{0, 0, 2000}, 0, planeTypes[k].vmax)
		turned, ticks := float32(0), 0
		last := bearingOf(p.f)
		for turned < 2*math.Pi && ticks < 120*tps {
			roll := clamp((1.35-p.bank())*3, -1, 1)
			pull := clamp(0.7+(2000-p.pos.z)*0.02-p.f.z*20, -1, 1)
			p.fly(roll, pull)
			b := bearingOf(p.f)
			turned += abs(wrap(b - last))
			last = b
			ticks++
		}
		t.Logf("kind %d: a full turn in %.1f s, altitude %.0f, speed %.0f km/h", k, float32(ticks)/tps, p.pos.z, p.v*3.6)
		if k == zero && (ticks < 12*tps || ticks > 25*tps) {
			t.Errorf("the Zero turned a circle in %.1f s", float32(ticks)/tps)
		}
		if abs(p.pos.z-2000) > 150 {
			t.Errorf("kind %d lost %.0f m in the turn", k, 2000-p.pos.z)
		}
	}
}

// Too slow, the wings cannot hold the plane up.
func TestStall(t *testing.T) {
	p := newPlane(zero, vec{0, 0, 2000}, 0, 40)
	p.throttle = 0
	top := p.pos.z
	for range 30 * tps {
		p.fly(0, 1)
		top = max(top, p.pos.z)
	}
	if top > 2100 || p.pos.z > 1950 {
		t.Fatalf("a stalled plane held up: top %.0f m, now %.0f m", top, p.pos.z)
	}
}

// Rolling right lowers the right wing, and pulling raises the nose.
func TestControlsSense(t *testing.T) {
	p := newPlane(zero, vec{0, 0, 2000}, 0, 120)
	for range 20 {
		p.fly(1, 0)
	}
	if p.bank() <= 0.2 {
		t.Fatalf("bank %.2f after rolling right", p.bank())
	}
	p = newPlane(zero, vec{0, 0, 2000}, 0, 120)
	for range 30 {
		p.fly(0, 1)
	}
	if p.f.z <= 0.05 {
		t.Fatalf("nose %.2f after pulling", p.f.z)
	}
}

// In a 70 degree dive the dive brakes hold the Type 99 under its limit;
// without them it overspeeds.
func TestDiveBrakes(t *testing.T) {
	for _, brakes := range []bool{true, false} {
		p := newPlane(val, vec{0, 0, 4000}, 0, 90)
		p.f = vec{0, cos(1.22), -sin(1.22)}
		p.u = vec{0, sin(1.22), cos(1.22)}
		p.brakes = brakes
		p.throttle = 0.25
		top := float32(0)
		for p.pos.z > 600 {
			// Hold the dive angle.
			_, pull := p.steer(p.pos.add(vec{0, cos(1.22), -sin(1.22)}.mul(1000)))
			p.fly(0, pull)
			top = max(top, p.v)
		}
		t.Logf("brakes %v: top speed %.0f km/h", brakes, top*3.6)
		if over := top > planeTypes[val].vne; over == brakes {
			t.Errorf("brakes %v: top speed %.0f km/h, limit %.0f", brakes, top*3.6, planeTypes[val].vne*3.6)
		}
	}
}

// diveBomb flies the Type 99 down onto the target ship, steering so that the
// predicted impact lies on the ship's future position, and releases at
// release meters. It returns the result.
func diveBomb(t *testing.T, g *Game, release float32) string {
	t.Helper()
	p := g.player()
	s := &g.ships[0]
	// Start 1200 m behind the carrier's beam at 3000 m: a 65 degree dive.
	p.pos = s.pos.add(heading(s.bearing + 1.5708).mul(1300))
	p.pos.z = 3000
	dir := s.pos.sub(p.pos).norm()
	p.f = dir
	p.u = vec{0, 0, 1}.sub(dir.mul(dir.z)).norm()
	p.v = 80
	p.brakes, p.throttle = true, 0.25
	aim := s.pos
	for range 120 * tps {
		if g.weapons > 0 {
			// Move the aim point until the predicted impact meets the ship
			// where it will be when the bomb falls.
			impact := g.bombImpact()
			future := s.pos.add(s.dir().mul(s.speed * g.bombFall()))
			aim = aim.add(future.sub(impact).mul(0.05))
			roll, pull := p.steer(aim)
			if p.pos.z < release {
				g.dropBomb()
			}
			p.fly(roll, pull)
		} else {
			// Pull out.
			p.fly(p.level(), 1)
		}
		g.stepNoPilot()
		if g.released > 0 && !g.bomb.alive {
			return g.lastResult
		}
	}
	return "TIMEOUT"
}

// stepNoPilot runs a tick of play with the player's plane already flown.
func (g *Game) stepNoPilot() {
	for i := 1; i < maxPlanes; i++ {
		if g.planes[i].alive {
			g.think(i)
		}
	}
	g.updateShips()
	g.updateBullets()
	g.updateWeapons()
	g.updatePuffs()
	g.updatePlayer()
	g.pickTargets()
}

func TestDiveBombHits(t *testing.T) {
	g := playing(missionDive)
	clear(g.planes[1:])
	if r := diveBomb(t, g, 500); r != "HIT CARRIER" {
		t.Fatalf("a careful dive: %s", r)
	}
	if g.ships[0].hp != shipTypes[carrier].hp-2 || g.score != shipTypes[carrier].hitScore {
		t.Fatalf("carrier hp %d, score %d", g.ships[0].hp, g.score)
	}
}

// A bomb aimed at the ship itself, without the bomb's fall, misses from high.
func TestBombNeedsAim(t *testing.T) {
	g := playing(missionDive)
	clear(g.planes[1:])
	p := g.player()
	s := &g.ships[0]
	p.pos = s.pos.add(vec{0, 0, 1200}).add(heading(s.bearing + 1.5708).mul(1100))
	p.f = s.pos.sub(p.pos).norm()
	p.u = vec{0, 0, 1}.sub(p.f.mul(p.f.z)).norm()
	p.v = 110
	g.dropBomb()
	for g.bomb.alive {
		g.updateWeapons()
	}
	if g.lastResult == "HIT CARRIER" {
		t.Fatal("a bomb dropped in a 45 degree dive at 1200 m hit the point under the nose")
	}
}

// A torpedo dropped low and slow at the lead mark hits a moving cruiser; a
// torpedo dropped high breaks.
func TestTorpedo(t *testing.T) {
	g := playing(missionTorpedo)
	clear(g.planes[1:])
	p := g.player()
	s := &g.ships[1]
	s.want = s.bearing
	g.ships[0].alive, g.ships[2].alive, g.ships[3].alive = false, false, false
	p.pos = s.pos.add(heading(s.bearing + 1.5708).mul(900))
	p.pos.z = 30
	p.v = 75
	p.throttle = 0.45
	aim, ok := g.lead(1)
	if !ok {
		t.Fatal("no lead")
	}
	p.f = aim.sub(p.pos.flat()).norm()
	p.u = vec{0, 0, 1}
	if !g.torpedoOK() {
		t.Fatalf("torpedo not ready at %.0f m, %.0f km/h", p.pos.z, p.v*3.6)
	}
	g.dropTorpedo()
	for range 60 * tps {
		g.updateWeapons()
		s.pos = s.pos.add(s.dir().mul(s.speed * dt))
		if !g.torp.alive {
			break
		}
	}
	if g.lastResult != "SUNK CRUISER" && g.lastResult != "HIT CRUISER" {
		t.Fatalf("torpedo: %q", g.lastResult)
	}
	g = playing(missionTorpedo)
	g.player().pos.z = 200
	g.dropTorpedo()
	if g.lastResult != "TORPEDO BROKE" || g.torp.alive || g.weapons != 0 {
		t.Fatalf("high drop: %q", g.lastResult)
	}
}

// Rounds from the Zero aimed at the lead shoot down a bomber flying straight.
func TestGunsShootDown(t *testing.T) {
	g := playing(missionFighter)
	p := g.player()
	e := &g.planes[1]
	*e = newPlane(bomber, p.pos.add(vec{0, 250, 20}), 0.4, 100)
	e.task, e.goal = taskRun, e.pos.add(heading(0.4).mul(50000))
	for range 20 * tps {
		d := e.pos.sub(p.pos).length()
		tof := d / (bulletSpeed + p.v)
		aim := e.pos.add(e.f.mul(e.v * tof)).add(vec{0, 0, gravity * tof * tof / 2})
		roll, pull := p.steer(aim)
		in := input{roll: roll, pull: pull, fire: d < 400 && aim.sub(p.pos).norm().dot(p.f) > 0.999}
		g.pilot(in)
		for i := 1; i < maxPlanes; i++ {
			if g.planes[i].alive {
				g.think(i)
			}
		}
		g.updateBullets()
		if e.falling > 0 {
			break
		}
	}
	if e.falling == 0 {
		t.Fatalf("the bomber flew on with %d hp", e.hp)
	}
	if g.score != 150 {
		t.Fatalf("score %d", g.score)
	}
}

// An enemy fighter chases a plane that flies straight and hits it.
func TestFighterAttacks(t *testing.T) {
	g := playing(missionDive)
	clear(g.planes[1:])
	p := g.player()
	g.planes[1] = newPlane(fighter, p.pos.add(vec{1500, -1500, 300}), 0, 120)
	g.planes[1].task = taskAttack
	ticks := 0
	for ; ticks < 90*tps && p.hp == planeTypes[val].hp; ticks++ {
		g.gunnerCool = 2 // keep the rear gunner quiet
		g.step(input{})
	}
	t.Logf("first hit after %.1f s", float32(ticks)/tps)
	if p.hp == planeTypes[val].hp {
		t.Fatal("the fighter never hit")
	}
}

// The fleet's guns hit a bomber that flies straight and level over it.
func TestFlak(t *testing.T) {
	g := playing(missionDive)
	clear(g.planes[1:])
	p := g.player()
	for range 180 * tps {
		// Circle over the fleet at 2000 m.
		c := g.ships[0].pos.add(heading(float32(g.clock) * 0.003).mul(1500))
		c.z = 2000
		roll, pull := p.steer(c)
		g.step(input{roll: roll, pull: pull})
		if p.hp < planeTypes[val].hp {
			break
		}
	}
	if p.hp == planeTypes[val].hp {
		t.Fatal("no damage from the flak")
	}
}

// Left alone, the raids hit the carrier until it is lost.
func TestRaidsSinkCarrier(t *testing.T) {
	g := playing(missionFighter)
	p := g.player()
	p.pos = vec{20000, 20000, 3000} // far from the fight
	for range 15 * 60 * tps {
		p.pos = vec{20000, 20000, 3000}
		g.step(input{})
		if g.mode != modePlay {
			break
		}
	}
	if g.mode != modeResult || g.outcome != "CARRIER LOST" {
		t.Fatalf("mode %d outcome %q carrier hp %d wave %d", g.mode, g.outcome, g.ships[0].hp, g.wave)
	}
}

// Flying into the sea ends the fighter mission, and a bomber's sortie.
func TestCrash(t *testing.T) {
	for _, m := range []mission{missionFighter, missionTorpedo} {
		g := playing(m)
		clear(g.planes[1:])
		p := g.player()
		p.f = vec{0, 0.5, -0.866}
		p.u = vec{0, 0.866, 0.5}
		for range 60 * tps {
			g.step(input{})
			if g.mode != modePlay {
				break
			}
		}
		switch {
		case m == missionFighter && (g.mode != modeResult || g.outcome != "CRASHED"):
			t.Errorf("fighter: mode %d outcome %q", g.mode, g.outcome)
		case m == missionTorpedo && (g.mode != modeDebrief || g.lastResult != "CRASHED" || g.lost != 1):
			t.Errorf("torpedo: mode %d result %q lost %d", g.mode, g.lastResult, g.lost)
		}
	}
}

// After the release and the run away from the fleet, the sortie ends, and
// the third ends the mission.
func TestSorties(t *testing.T) {
	g := playing(missionDive)
	for n := 1; n <= sorties; n++ {
		clear(g.planes[1:])
		if g.sortie != n {
			t.Fatalf("sortie %d, want %d", g.sortie, n)
		}
		g.dropBomb()
		for range 90 * tps {
			g.planes[0].hp = 100
			g.step(input{})
			if g.mode != modePlay {
				break
			}
		}
		if g.mode != modeDebrief {
			t.Fatalf("sortie %d: mode %d", n, g.mode)
		}
		g.modeTicks = 100
		if n < sorties {
			g.startSortie()
		}
	}
}

// keyPlayer drives the game through RunGame with keys: it picks a mission on
// the select screen and flies with the arrows.
type keyPlayer struct {
	*Game
	t      *testing.T
	tick   int
	choose int
}

func (k *keyPlayer) Update() error {
	k.tick++
	down, a, right := false, false, false
	switch k.mode {
	case modeSelect:
		if k.cursor != k.choose {
			down = k.tick%2 == 0
		} else {
			a = k.tick%2 == 0
		}
	case modePlay:
		right = k.modeTicks < 30
		if k.modeTicks > 60 {
			return ebiten.Termination
		}
	}
	ebiten.SetKeyState(ebiten.KeyArrowDown, down)
	ebiten.SetKeyState(ebiten.KeyArrowRight, right)
	ebiten.SetKeyState(ebiten.KeyZ, a)
	if k.tick > 600 {
		k.t.Error("did not start")
		return ebiten.Termination
	}
	return k.Game.Update()
}

func TestKeysSelectAndFly(t *testing.T) {
	ebiten.SetTPS(1000)
	defer ebiten.SetTPS(60)
	ebiten.SetHardware(&hardware.Headless{Width: 240, Height: 240})
	for m := range missions {
		k := &keyPlayer{Game: NewGame(), t: t, choose: int(m)}
		if err := ebiten.RunGame(k); err != nil {
			t.Fatal(err)
		}
		for _, key := range []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyArrowRight, ebiten.KeyZ} {
			ebiten.SetKeyState(key, false)
		}
		if k.mission != m || k.player().kind != missionPlane[m] || k.player().bank() < 0.3 {
			t.Errorf("mission %d: flying %d, bank %.2f", k.mission, k.player().kind, k.player().bank())
		}
	}
}

// ace flies the player's Zero like the enemy pilots: it chases the bombers
// first, leads them and fires when the lead is in the sight.
func ace(g *Game) input {
	p := g.player()
	target := -1
	best := float32(1e9)
	for i := 1; i < maxPlanes; i++ {
		q := &g.planes[i]
		if !q.alive || q.falling > 0 {
			continue
		}
		d := q.pos.sub(p.pos).length()
		if q.kind == bomber {
			d -= 3000
		}
		if d < best {
			best, target = d, i
		}
	}
	if target < 0 {
		// Wait over the fleet.
		c := g.ships[0].pos.add(vec{0, 0, 2500})
		roll, pull := p.steer(c.add(heading(float32(g.clock) * 0.004).mul(2000)))
		return input{roll: roll, pull: pull}
	}
	q := &g.planes[target]
	d := q.pos.sub(p.pos).length()
	tof := d / (bulletSpeed + p.v)
	aim := q.pos.add(q.f.mul(q.v * tof)).add(vec{0, 0, gravity * tof * tof / 2})
	roll, pull := p.steer(aim)
	// Break when a fighter sits behind.
	for i := 1; i < maxPlanes; i++ {
		e := &g.planes[i]
		if e.alive && e.falling == 0 && e.kind == fighter {
			to := p.pos.sub(e.pos)
			if to.length() < 500 && to.norm().dot(e.f) > 0.95 && to.dot(p.f) > 0 {
				roll, pull = clamp(1-p.bank(), -1, 1), 1
			}
		}
	}
	if p.pos.z < 400+max(0, -p.f.z)*p.v*4 && p.f.z < 0.2 {
		roll, pull = p.level(), 1
	}
	c := aim.sub(p.pos).norm().dot(p.f)
	return input{roll: roll, pull: pull, fire: d < 450 && c > 0.9995, second: d < 350 && c > 0.9995}
}

// A good fighter pilot can defend the fleet through the three raids.
func TestFighterMissionWinnable(t *testing.T) {
	wins := 0
	for seed := range 5 {
		g := playing(missionFighter)
		g.rng += uint32(seed) * 7919
		for range 20 * 60 * tps {
			g.step(ace(g))
			if g.mode != modePlay {
				break
			}
		}
		t.Logf("seed %d: %s, score %d, carrier %d, hp %d, wave %d", seed, g.outcome, g.score, g.ships[0].hp, g.player().hp, g.wave)
		if g.outcome == "FLEET DEFENDED" {
			wins++
		}
	}
	if wins < 2 {
		t.Fatalf("%d wins of 5", wins)
	}
}

// bomberAce flies a sortie of the dive bomber or the torpedo bomber: in at
// height, dive or let down, aim with the sight's marks, release and leave.
type bomberAce struct {
	aim    vec
	diving bool
}

func (a *bomberAce) fly(g *Game) input {
	p := g.player()
	s := &g.ships[max(g.target, 0)]
	if g.target < 0 || g.weapons == 0 {
		// Away from the fleet, low over the sea if already low.
		away := p.pos.add(p.pos.sub(g.ships[0].pos).flat().norm().mul(3000))
		away.z = max(min(p.pos.z, 1500), 300)
		roll, pull := p.steer(away)
		if p.pos.z < 250+max(0, -p.f.z)*p.v*4 && p.f.z < 0.2 {
			roll, pull = p.level(), 1
		}
		return input{roll: roll, pull: pull, toggle: p.brakes}
	}
	flat := s.pos.sub(p.pos.flat()).length()
	if p.kind == val {
		if !a.diving {
			level := s.pos
			level.z = p.pos.z
			roll, pull := p.steer(level)
			if flat < 1400 {
				a.diving, a.aim = true, s.pos
				return input{roll: roll, pull: pull, toggle: !p.brakes}
			}
			return input{roll: roll, pull: pull}
		}
		future := s.pos.add(s.dir().mul(s.speed * g.bombFall()))
		a.aim = a.aim.add(future.sub(g.bombImpact()).mul(0.05))
		roll, pull := p.steer(a.aim)
		return input{roll: roll, pull: pull, release: p.pos.z < 550}
	}
	// The torpedo run: down to 40 m and slow from 5 km out.
	want := s.pos
	if aim, ok := g.lead(g.target); ok {
		want = aim
	}
	want.z = 40
	if flat > 7500 {
		want.z = p.pos.z
	}
	roll, pull := p.steer(want)
	toggle := (flat < 6000) == (p.throttle == 1)
	return input{roll: roll, pull: pull, toggle: toggle, release: flat < 1000 && g.torpedoOK()}
}

// A good bomber pilot comes through the flak and hits in some sorties.
func TestBomberMissionsWinnable(t *testing.T) {
	for _, m := range []mission{missionDive, missionTorpedo} {
		hits, lost := 0, 0
		for seed := range 3 {
			g := playing(m)
			g.rng += uint32(seed) * 7919
			for g.mode != modeResult {
				a := &bomberAce{}
				for g.mode == modePlay {
					g.step(a.fly(g))
				}
				if g.mode == modeDebrief {
					t.Logf("mission %d seed %d sortie %d: %s, hp %d", m, seed, g.sortie, g.lastResult, g.player().hp)
					if len(g.lastResult) > 4 && (g.lastResult[:4] == "HIT " || g.lastResult[:4] == "SUNK") {
						hits++
					}
					if g.sortie >= sorties || g.fleetSunk() {
						g.finish("MISSION OVER")
					} else {
						g.startSortie()
					}
				}
			}
			lost += g.lost
		}
		t.Logf("mission %d: %d hits and %d planes lost in 9 sorties", m, hits, lost)
		if hits < 3 || lost > 4 {
			t.Errorf("mission %d: %d hits in 9 sorties, %d planes lost", m, hits, lost)
		}
	}
}
