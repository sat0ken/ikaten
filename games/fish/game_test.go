package fish

import (
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

func started() *Game {
	g := NewGame()
	g.setMode(modeCast)
	return g
}

// cast holds A until the power reaches p, then lets go, and waits for the
// jig to land.
func cast(g *Game, p float32) {
	g.step(input{a: true, aHeld: true})
	for g.mode == modeCast && g.power < p {
		g.step(input{aHeld: true})
	}
	for g.mode != modeFishing {
		g.step(input{})
	}
}

func TestCastDistance(t *testing.T) {
	for _, p := range []float32{0.02, 0.5, 1} {
		g := started()
		cast(g, p)
		want := minCast + (maxCast-minCast)*g.power
		if abs(g.lureX-want) > 0.01 || g.lureDepth != 0 {
			t.Errorf("power %.2f: landed at %.1f m, depth %.1f, want %.1f m", g.power, g.lureX, g.lureDepth, want)
		}
	}
}

// The press that starts the day does not cast: the cast needs its own press.
func TestCastNeedsPress(t *testing.T) {
	g := started()
	for range 30 {
		g.step(input{aHeld: true})
	}
	g.step(input{})
	if g.mode != modeCast || g.power != 0 {
		t.Fatalf("a held start cast: mode %d power %.2f", g.mode, g.power)
	}
}

// After landing the jig falls to the seabed unless stopped; down lets out
// more line, up reels it in, and B winds it all in.
func TestSinkAndStop(t *testing.T) {
	g := started()
	cast(g, 0.5)
	for range 30 * tps {
		g.step(input{})
	}
	if floor := seabed(g.lureX); abs(g.lureDepth-(floor-0.2)) > 0.3 || g.onBottom == 0 {
		t.Fatalf("jig at %.1f m over a %.1f m bottom", g.lureDepth, floor)
	}
	g = started()
	cast(g, 0.5)
	for range 5 * tps {
		g.step(input{})
	}
	g.step(input{a: true})
	d := g.lureDepth
	for range 5 * tps {
		g.step(input{})
	}
	if abs(g.jigDepth()-d) > 0.05 {
		t.Fatalf("stopped jig moved from %.1f to %.1f m", d, g.jigDepth())
	}
	g.step(input{down: true})
	if g.lureDepth <= d {
		t.Fatal("down did not let out line")
	}
	for range 120 * tps {
		g.step(input{b: true})
		if g.mode != modeFishing {
			break
		}
	}
	if g.mode != modeCast {
		t.Fatalf("winding in ended in mode %d", g.mode)
	}
}

// placeAt puts fish i of a kind beside the stopped jig, ready to nibble.
func placeAt(g *Game, k species) *fish {
	f := &g.fish[0]
	*f = fish{kind: k, alive: true, x: g.lureX - 0.3, depth: g.jigDepth(), length: speciesTypes[k].maxLen, facing: 1, state: approach}
	g.suitor = 0
	return f
}

// fishingAt casts and stops the jig at depth d.
func fishingAt(d float32) *Game {
	g := started()
	cast(g, 0.6)
	for g.lureDepth < d {
		g.step(input{})
	}
	g.step(input{a: true})
	for i := range g.fish[1:] {
		g.fish[i+1].cool = 1 << 30 // the others leave the jig alone
	}
	return g
}

// Striking at a nibble scares the fish; striking at the bite hooks it; not
// striking lets it go.
func TestStrikeTiming(t *testing.T) {
	g := fishingAt(10)
	f := placeAt(g, aji)
	for f.state != nibble {
		g.step(input{})
	}
	g.step(input{a: true})
	if f.state != flee || g.suitor != -1 || g.message != msgTooEarly {
		t.Fatalf("early: state %d message %d", f.state, g.message)
	}

	g = fishingAt(10)
	f = placeAt(g, aji)
	for f.state != bite {
		g.step(input{})
	}
	g.step(input{a: true})
	if g.mode != modeFight || f.state != hooked {
		t.Fatalf("on time: mode %d state %d", g.mode, f.state)
	}

	g = fishingAt(10)
	f = placeAt(g, aji)
	for range 20 * tps {
		g.step(input{})
		if g.message == msgMissed {
			break
		}
	}
	if g.message != msgMissed || g.mode != modeFishing {
		t.Fatalf("late: mode %d message %d", g.mode, g.message)
	}
}

// hookedOn starts a fight with a big fish of a kind.
func hookedOn(k species) *Game {
	g := fishingAt(12)
	f := placeAt(g, k)
	for f.state != bite {
		g.step(input{})
	}
	g.step(input{a: true})
	return g
}

// skilled reels while the fish rests and the tension is safe.
func skilled(g *Game) input {
	return input{aHeld: !g.running && g.tension < 0.8 || g.tension < 0.25}
}

// A skilled angler lands even the biggest fish; one who reels through the
// runs breaks the line, and one who never reels loses the fish.
func TestFight(t *testing.T) {
	for k := range boot {
		g := hookedOn(k)
		ticks := 0
		for ; g.mode == modeFight && ticks < 300*tps; ticks++ {
			g.step(skilled(g))
		}
		t.Logf("%-8s landed in %.0f s", speciesTypes[k].name, float32(ticks)/tps)
		if g.mode != modeCatch {
			t.Errorf("%s: mode %d, message %d", speciesTypes[k].name, g.mode, g.message)
		}
	}
	g := hookedOn(buri)
	for g.mode == modeFight {
		g.step(input{aHeld: true})
	}
	if g.message != msgBroke {
		t.Errorf("reeling through the runs: message %d", g.message)
	}
	g = hookedOn(aji)
	for g.mode == modeFight {
		g.step(input{})
	}
	if g.message != msgGotAway {
		t.Errorf("never reeling: message %d", g.message)
	}
}

func TestCatchScore(t *testing.T) {
	g := hookedOn(madai)
	for g.mode == modeFight {
		g.step(skilled(g))
	}
	c := g.lastLand
	if c.kind != madai || c.length != 75 || c.weight < 6.2 || c.weight > 6.4 || c.points != int(c.weight*100)+100 {
		t.Fatalf("catch %+v", c)
	}
	if g.score != c.points || g.catches[madai] != 1 || g.best[madai] != 75 {
		t.Fatalf("score %d, catches %d", g.score, g.catches[madai])
	}
	for range 40 {
		g.step(input{})
	}
	g.step(input{a: true})
	g.step(input{})
	if g.mode != modeCast {
		t.Fatalf("after the catch: mode %d", g.mode)
	}
}

// Ribbonfish and squid come out at dusk.
func TestDusk(t *testing.T) {
	g := NewGame()
	count := func() (n int) {
		for range 200 {
			for i := range g.fish {
				g.spawn(i, true)
				if k := g.fish[i].kind; k == tachiuo || k == ika {
					n++
				}
			}
		}
		return n
	}
	g.clock = 12 * 3600
	noon := count()
	g.clock = 17 * 3600
	dusk := count()
	if dusk < noon*2 {
		t.Fatalf("ribbonfish and squid: %d at noon, %d at dusk", noon, dusk)
	}
}

// angler plays a whole day: it casts to a spot, stops the jig at a depth,
// jigs now and then, strikes at the bite and fights with skill.
type angler struct {
	depth float32
	n     int
}

func (a *angler) play(g *Game) input {
	switch g.mode {
	case modeCast:
		if !g.charging {
			return input{a: true, aHeld: true}
		}
		if g.power < 0.3+float32(a.n%5)*0.15 {
			return input{aHeld: true}
		}
		a.depth = 6 + float32(a.n%6)*7
		a.n++
		return input{}
	case modeFishing:
		if g.suitor >= 0 && g.fish[g.suitor].state == bite {
			return input{a: true}
		}
		if g.freeFall && g.lureDepth >= a.depth {
			return input{a: true}
		}
		nibbling := g.suitor >= 0 && g.fish[g.suitor].state != approach
		if !g.freeFall && !nibbling && g.modeTicks%180 == 0 {
			return input{a: true}
		}
		// Nothing for a while: try elsewhere.
		if g.modeTicks > 40*tps && g.suitor < 0 {
			return input{b: true}
		}
	case modeFight:
		return skilled(g)
	case modeCatch:
		return input{a: g.modeTicks > 40}
	}
	return input{}
}

func TestAnglerDay(t *testing.T) {
	g := started()
	a := &angler{}
	for g.mode != modeResult {
		g.step(a.play(g))
	}
	n, kinds := 0, 0
	for k, c := range g.catches {
		n += c
		if c > 0 && species(k) != boot {
			kinds++
		}
		if c > 0 {
			t.Logf("%-8s x%d, best %.0f cm", speciesTypes[k].name, c, g.best[k])
		}
	}
	t.Logf("%d catches, %.2f kg, %d points", n, g.total, g.score)
	if n < 8 || kinds < 3 {
		t.Fatalf("%d catches of %d kinds in a day", n, kinds)
	}
}

// The day ends at six in the evening.
func TestDayEnds(t *testing.T) {
	g := started()
	g.clock = dayEnd - 10
	for range 20 {
		g.step(input{})
	}
	if g.mode != modeResult {
		t.Fatalf("mode %d at %d", g.mode, g.clock)
	}
}

// keyAngler plays through RunGame with keys, from the title to a cast.
type keyAngler struct {
	*Game
	t    *testing.T
	tick int
}

func (k *keyAngler) Update() error {
	k.tick++
	a, down := false, false
	switch k.mode {
	case modeTitle:
		a = k.tick%2 == 0
	case modeCast:
		// Let go of the start, then press and hold.
		a = k.modeTicks > 3 && k.power < 0.5
	case modeFishing:
		down = true
		if k.modeTicks > 60 {
			return ebiten.Termination
		}
	}
	ebiten.SetKeyState(ebiten.KeyZ, a)
	ebiten.SetKeyState(ebiten.KeyArrowDown, down)
	if k.tick > 1200 {
		k.t.Error("no cast")
		return ebiten.Termination
	}
	return k.Game.Update()
}

func TestKeysCast(t *testing.T) {
	ebiten.SetTPS(1000)
	defer ebiten.SetTPS(60)
	ebiten.SetHardware(&hardware.Headless{Width: 240, Height: 240})
	k := &keyAngler{Game: NewGame(), t: t}
	if err := ebiten.RunGame(k); err != nil {
		t.Fatal(err)
	}
	ebiten.SetKeyState(ebiten.KeyZ, false)
	ebiten.SetKeyState(ebiten.KeyArrowDown, false)
	if k.mode != modeFishing || k.lureX < 20 || k.lureDepth < 1 {
		t.Fatalf("mode %d, jig at %.1f m, depth %.1f", k.mode, k.lureX, k.lureDepth)
	}
}

// Drawing a frame allocates nothing, on the title, casting, fishing deep,
// fighting, and the catch and the day's result.
func TestDrawAllocs(t *testing.T) {
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	check := func(name string, g *Game) {
		t.Helper()
		g.Draw(screen)
		if n := testing.AllocsPerRun(5, func() { g.ticks++; g.Draw(screen) }); n != 0 {
			t.Errorf("%s: %v allocations", name, n)
		}
	}
	check("title", NewGame())
	g := started()
	g.step(input{a: true, aHeld: true})
	check("cast", g)
	check("fishing", fishingAt(40))
	g = hookedOn(madai)
	check("fight", g)
	for g.mode == modeFight {
		g.step(skilled(g))
	}
	check("catch", g)
	g.setMode(modeResult)
	check("result", g)
}
