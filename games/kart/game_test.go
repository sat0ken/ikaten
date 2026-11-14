package kart

import (
	"math"
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/kart/course"
)

func racing() *Game {
	g := NewGame()
	g.setMode(modeRace)
	return g
}

// The CPU driving every kart, including the player's, finishes the race.
func TestRaceFinishes(t *testing.T) {
	g := racing()
	for g.finishers < karts && g.raceTicks < 60*180 {
		g.raceTicks++
		g.step(true)
	}
	if g.finishers != karts {
		t.Fatalf("%d of %d finished after %d ticks", g.finishers, karts, g.raceTicks)
	}
	if g.bestLap == 0 || g.bestLap > g.karts[0].finishTicks/2 {
		t.Fatalf("best lap %d of %d ticks", g.bestLap, g.karts[0].finishTicks)
	}
	for p, i := range g.order {
		if p > 0 && g.karts[i].finishTicks < g.karts[g.order[p-1]].finishTicks {
			t.Fatalf("order %v", g.order)
		}
	}
	t.Logf("winner %d in %.1f s, the player's best lap %.2f s", g.order[0], float64(g.karts[g.order[0]].finishTicks)/60, float64(g.bestLap)/60)
}

// placeAt puts the player on the course heading along it.
func placeAt(g *Game, s, offset float32) *kart {
	k := &g.karts[0]
	p := course.Place(s, offset)
	_, d := course.At(s)
	k.x, k.y = p.X, p.Y
	k.heading = float32(math.Atan2(float64(d.Y), float64(d.X)))
	k.slide = k.heading
	g.locate(k, -1)
	return k
}

func TestBoostPad(t *testing.T) {
	g := racing()
	p := course.Pads[0]
	k := placeAt(g, p.S-2, p.Offset)
	k.speed = 0.5
	g.move(0, input{gas: true})
	if k.boost == 0 {
		t.Fatal("no boost on the pad")
	}
	for range 10 {
		g.move(0, input{gas: true})
	}
	if k.speed <= maxSpeed {
		t.Fatalf("speed %.2f with a boost", k.speed)
	}
}

func TestLongDriftBoosts(t *testing.T) {
	g := racing()
	k := placeAt(g, course.Pads[0].S+20, 0)
	k.speed = maxSpeed
	g.move(0, input{gas: true, drift: true, steer: -1}) // toward the infield, away from the wall
	for range driftLong {
		g.move(0, input{gas: true, drift: true})
	}
	if k.drift == 0 || k.charge < driftLong {
		t.Fatalf("drift=%d charge=%d", k.drift, k.charge)
	}
	g.move(0, input{gas: true})
	if k.boost < bigBoost-1 {
		t.Fatalf("boost %d after a long drift", k.boost)
	}
}

// Crossing the line backward takes a lap away, so it cannot be farmed.
func TestLapsCountForward(t *testing.T) {
	g := racing()
	k := placeAt(g, course.Length-3, 0)
	k.lap = 0
	placeAtKeep := func(s float32) {
		p := course.Place(s, 0)
		k.x, k.y = p.X, p.Y
		g.locate(k, 6)
	}
	placeAtKeep(2)
	if k.lap != 1 {
		t.Fatalf("forward crossing: lap %d", k.lap)
	}
	placeAtKeep(course.Length - 2)
	if k.lap != 0 {
		t.Fatalf("backward crossing: lap %d", k.lap)
	}
}

// Drawing a frame allocates nothing: on the title, in the countdown, racing
// with a drift and a boost, and at the finish.
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
	g.setMode(modeCountdown)
	check("countdown")
	g.setMode(modeRace)
	for range 200 {
		g.raceTicks++
		g.step(true)
		g.follow(0.25)
	}
	k := &g.karts[0]
	k.drift, k.charge, k.boost = 1, 80, 10
	check("race")
	g.setMode(modeResult)
	g.modeTicks = 100
	check("result")
}
