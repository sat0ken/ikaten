package tank

import (
	"math"
	"testing"

	ebiten "github.com/sat0ken/ikaten"

	"github.com/sat0ken/ikaten/games/tank/field"
)

// battle returns a game in play with a single enemy at a range straight north
// of the player, on open ground, facing heading.
func battle(k kind, r, heading float32) (*Game, *tank) {
	g := NewGame()
	g.hp = 100
	g.newBattle(1)
	clear(g.enemies[:])
	g.setMode(modePlay)
	g.x, g.y = 0, 0
	e := &g.enemies[0]
	*e = tank{kind: k, x: 0, y: -r, heading: heading, turret: heading, alive: true, until: 1 << 30}
	return g, e
}

// step runs the simulation without input, the way Update does in play.
func (g *Game) step() {
	g.ticks++
	if g.reload > 0 {
		g.reload--
	}
	g.updatePlayer()
	g.updateEnemies()
	g.updateShells()
	g.updatePuffs()
}

// layGun aims the player's gun the way the sight's marks do: the mark for the
// range on the middle of the target.
func (g *Game) layGun(e *tank) {
	r := hypot(e.x-g.x, e.y-g.y)
	g.turret = bearingTo(e.x-g.x, e.y-g.y)
	g.elev = dropAngle(r) + (aimHeight-eyeHeight)/r
}

func (g *Game) shootAndWait() {
	g.fire()
	for range 300 {
		g.step()
	}
}

// The test field is open between the player at the origin and the enemy.
func TestOpenGround(t *testing.T) {
	g, _ := battle(light, 1000, 0)
	if !g.clear(0, 0, 0, -1000) {
		t.Fatal("the test line is blocked")
	}
}

// A gun laid with the range marks hits at every range.
func TestRangeMarksHit(t *testing.T) {
	for _, r := range []float32{150, 300, 500, 700, 900, 1000} {
		g, e := battle(light, r, math.Pi/2) // the light's side, which any hit penetrates
		g.layGun(e)
		g.shootAndWait()
		if e.alive || g.kills != 1 {
			t.Errorf("range %v: the shell missed", r)
		}
		// Without the drop, the shell falls short.
		g, e = battle(light, r, math.Pi/2)
		g.layGun(e)
		g.elev -= dropAngle(r)
		g.shootAndWait()
		if r >= 300 && !e.alive {
			t.Errorf("range %v: hit without allowing for the drop", r)
		}
	}
}

func TestArmorFaces(t *testing.T) {
	for _, c := range []struct {
		heading float32
		alive   bool
	}{
		{math.Pi, true},      // facing the player: the heavy's front holds
		{math.Pi / 2, false}, // the side gives
		{0, false},           // and the rear
	} {
		g, e := battle(heavy, 600, c.heading)
		g.layGun(e)
		g.shootAndWait()
		if e.alive != c.alive {
			t.Errorf("heading %v: alive=%v, message %q", c.heading, e.alive, messages[g.message])
		}
	}
}

func TestPenetrates(t *testing.T) {
	if penetrates(100, 100, 1) != penetrated || penetrates(99, 100, 1) != noPenetration {
		t.Fatal("straight on")
	}
	if penetrates(150, 100, 0.6) != noPenetration || penetrates(170, 100, 0.6) != penetrated {
		t.Fatal("at an angle the armor is thicker")
	}
	if penetrates(1000, 10, 0.2) != ricochet {
		t.Fatal("past about 72 degrees the shell glances off")
	}
}

// Houses and trees stop shells.
func TestObstaclesStopShells(t *testing.T) {
	var o field.Obstacle
	for _, c := range field.Obstacles {
		if c.Kind == field.House {
			o = c
			break
		}
	}
	g, e := battle(light, 0, math.Pi/2)
	// Put the enemy behind the house, the player 300 m in front of it.
	e.x, e.y = o.X, o.Y-20
	g.x, g.y = o.X, o.Y+300
	if g.clear(g.x, g.y, e.x, e.y) {
		t.Fatal("the house does not block the line")
	}
	g.layGun(e)
	g.shootAndWait()
	if !e.alive {
		t.Fatal("the shell went through the house")
	}
}

// The turret keeps its bearing while the hull turns, and the hull drives
// along its own heading.
func TestHullAndTurret(t *testing.T) {
	g, _ := battle(light, 900, 0)
	g.turret = 0.5
	g.heading = 0
	g.throttle = len(throttles) - 1
	for range 200 {
		g.heading = angleDiff(g.heading+specs[player].turn, 0)
		g.step()
	}
	if g.turret != 0.5 {
		t.Fatalf("turret %v", g.turret)
	}
	if g.speed <= 0 || hypot(g.x, g.y) < 10 {
		t.Fatalf("speed %v at %v, %v", g.speed, g.x, g.y)
	}
}

func TestBlockedByObstacles(t *testing.T) {
	o := field.Obstacles[0]
	g, _ := battle(light, 900, 0)
	if !g.blocked(o.X+o.Radius, o.Y, -1) {
		t.Fatal("drove into an obstacle")
	}
	if !g.blocked(field.Half, 0, -1) {
		t.Fatal("drove off the field")
	}
}

// Enemies find, approach and finally destroy a player who does nothing, but
// not too soon.
func TestEnemiesFight(t *testing.T) {
	for level := 1; level <= 4; level++ {
		g := NewGame()
		g.hp, g.level = 100, level
		g.newBattle(level)
		g.setMode(modePlay)
		ticks := 0
		for ; ticks < 60*600 && g.hp > 0; ticks++ {
			g.step()
		}
		t.Logf("level %d: the idle player lasted %d s", level, ticks/60)
		if g.hp > 0 {
			t.Errorf("level %d: the enemies never destroyed the idle player", level)
		}
		if ticks < 60*30 {
			t.Errorf("level %d: the idle player lasted only %d s", level, ticks/60)
		}
	}
}

func TestStartsClear(t *testing.T) {
	for level := 1; level <= 8; level++ {
		g := NewGame()
		g.newBattle(level)
		if g.blocked(g.x, g.y, -1) {
			t.Fatalf("level %d: the player starts blocked", level)
		}
		for i := range g.enemies {
			if e := &g.enemies[i]; e.alive && g.blocked(e.x, e.y, i) {
				t.Fatalf("level %d: enemy %d starts blocked at %v, %v", level, i, e.x, e.y)
			}
		}
	}
}

// Drawing allocates nothing, in the wide and the zoomed view, with puffs,
// shells and the banners.
func TestDrawAllocs(t *testing.T) {
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	g := NewGame()
	check := func(name string) {
		t.Helper()
		g.Draw(screen)
		if n := testing.AllocsPerRun(5, func() { g.ticks++; g.Draw(screen) }); n != 0 {
			t.Errorf("%s: %v allocations", name, n)
		}
	}
	check("title")
	g, e := battle(medium, 400, 0)
	g.layGun(e)
	g.fire()
	for range 20 {
		g.step()
	}
	g.say(msgRicochet)
	g.shake, g.hurt, g.hitTicks = 6, 6, 6
	check("wide")
	g.zoomed = true
	check("zoomed")
	g.setMode(modeClear)
	g.modeTicks = 120
	check("clear")
	g.setMode(modeOver)
	g.modeTicks = 120
	check("over")
}
