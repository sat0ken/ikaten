package uboat

import (
	"image"
	"math"
	"testing"
)

func playing() *Game {
	g := NewGame()
	g.setMode(modePlay)
	return g
}

// step runs the simulation without input, the way Update does in play.
func (g *Game) step() {
	g.ticks++
	if g.moving > 0 {
		g.moving--
	}
	if g.reload > 0 {
		g.reload--
	}
	g.updateSub()
	g.updateShips()
	g.updateTorps()
	g.updateCharges()
	g.updateDetection()
}

// A torpedo fired at the aim mark meets a moving merchant.
func TestLeadHits(t *testing.T) {
	g := playing()
	g.up = false // stay undetected
	s := &g.ships[0]
	for range 600 {
		g.step()
	}
	brg, ok := g.lead(s)
	if !ok {
		t.Fatal("no solution")
	}
	g.bearing = brg
	g.fire()
	for range 1200 {
		g.step()
		if s.sink > 0 {
			break
		}
	}
	if s.sink == 0 || g.tonnage != shipTypes[s.kind].tons || g.torpedoes != torpedoes-1 {
		t.Fatalf("sink=%d tonnage=%d torpedoes=%d", s.sink, g.tonnage, g.torpedoes)
	}
	// Aiming straight at the ship instead misses: it has moved on.
	g = playing()
	g.up = false
	s = &g.ships[0]
	for range 600 {
		g.step()
	}
	g.bearing = bearingTo(s.x, s.y)
	g.fire()
	for range 1200 {
		g.step()
	}
	if s.sink > 0 {
		t.Fatal("a torpedo aimed at the ship itself hit; the lead is pointless")
	}
}

func TestTankerTakesTwoHits(t *testing.T) {
	g := playing()
	s := &g.ships[1]
	if s.kind != tanker {
		t.Fatalf("ship 1 is kind %d", s.kind)
	}
	for hit := 1; hit <= 2; hit++ {
		g.torps[0] = torp{x: s.x, y: s.y, alive: true}
		g.updateTorps()
		if (s.sink > 0) != (hit == 2) {
			t.Fatalf("after hit %d: sink=%d", hit, s.sink)
		}
	}
}

// Keeping the periscope up brings the escort and its depth charges; lowering it
// lets the escort lose contact.
func TestDetectionAttackAndEscape(t *testing.T) {
	g := playing()
	for i := 0; i < 60*60 && !g.alert; i++ {
		g.step()
	}
	if !g.alert {
		t.Fatalf("not detected after a minute, detect=%.0f", g.detect)
	}
	hull := g.hull
	for i := 0; i < 60*90 && g.hull == hull; i++ {
		g.step()
	}
	if g.hull >= hull {
		t.Fatal("no depth charge damage while detected")
	}
	g.up = false
	for i := 0; i < 60*30 && g.alert; i++ {
		g.step()
	}
	if g.alert {
		t.Fatalf("still hunted after 30 s down, detect=%.0f", g.detect)
	}
}

// With the periscope down the whole time, the convoy passes and the mission
// ends without damage.
func TestConvoyPasses(t *testing.T) {
	g := playing()
	g.up = false
	ticks := 0
	for ; ticks < 60*400 && !g.convoyGone(); ticks++ {
		g.step()
	}
	if !g.convoyGone() || g.hull != 100 || g.alert {
		t.Fatalf("after %d ticks: gone=%v hull=%.0f alert=%v", ticks, g.convoyGone(), g.hull, g.alert)
	}
	t.Logf("the convoy passed in %.0f s", float64(ticks)/60)
}

func TestDepthChargeDamageFallsWithDistance(t *testing.T) {
	for _, c := range []struct {
		x       float32
		damaged bool
	}{{0, true}, {chargeReach / 2, true}, {chargeReach + 1, false}} {
		g := playing()
		g.drop(c.x, 0)
		for range chargeFuse {
			g.updateCharges()
		}
		if damaged := g.hull < 100; damaged != c.damaged {
			t.Fatalf("charge at %.0f m: hull %.1f", c.x, g.hull)
		}
	}
}

func TestBearingConvention(t *testing.T) {
	for _, c := range []struct{ x, y, deg float32 }{{0, -1, 0}, {1, 0, 90}, {0, 1, 180}, {-1, 0, -90}} {
		if got := bearingTo(c.x, c.y) * 180 / math.Pi; abs(got-c.deg) > 0.01 && abs(abs(got)-180) > 0.01 {
			t.Fatalf("bearing to (%.0f,%.0f) = %.1f", c.x, c.y, got)
		}
	}
}

// How long the periscope can stay up before the escort attacks, by distance.
func TestDetectionTimes(t *testing.T) {
	for _, d := range []float32{1500, 1000, 600, 300} {
		g := playing()
		for i := range g.ships {
			if s := &g.ships[i]; s.kind == destroyer {
				s.x, s.y, s.speed = 0, -d, 0
			}
		}
		ticks := 0
		for ; !g.alert && ticks < 60*60; ticks++ {
			g.updateDetection()
		}
		down := 0
		g.up = false
		for ; g.alert; down++ {
			g.updateDetection()
		}
		t.Logf("destroyer at %4.0f m: detected after %4.1f s up, lost after %.1f s down", d, float64(ticks)/60, float64(down)/60)
		if ticks < 60*3 {
			t.Fatalf("detected too fast at %.0f m", d)
		}
	}
}

// The submarine moves only with the periscope down, and slows to a stop when it
// is raised.
func TestSubMovesOnlyDown(t *testing.T) {
	g := playing()
	g.throttle = 2
	for range 120 {
		g.step()
	}
	if g.subX != 0 || g.subY != 0 {
		t.Fatalf("moved with the periscope up: %.1f,%.1f", g.subX, g.subY)
	}
	g.up = false
	g.subHeading = math.Pi / 2 // east
	for range 300 {
		g.step()
	}
	if g.subX < 100 || abs(g.subY) > 1 || g.subSpeed != throttles[2] {
		t.Fatalf("down at full: %.1f,%.1f speed %.2f", g.subX, g.subY, g.subSpeed)
	}
	g.up = true
	for range 120 {
		g.step()
	}
	if g.subSpeed != 0 {
		t.Fatalf("speed %.2f with the periscope up", g.subSpeed)
	}
}

// A destroyer runs over where it last fixed the submarine. Staying put takes
// the depth charges; moving off at full speed after diving makes them miss.
func TestMovingDodgesDepthCharges(t *testing.T) {
	for _, move := range []bool{false, true} {
		g := playing()
		for i := range g.ships {
			if s := &g.ships[i]; s.kind == destroyer {
				s.x, s.y, s.heading = 0, -300, math.Pi
			}
		}
		g.detect = 100
		g.updateDetection()
		for i := range g.ships {
			if s := &g.ships[i]; s.kind == destroyer {
				s.errX, s.errY = 0, 0
				g.reaim(s)
			}
		}
		g.up = false
		if move {
			g.subHeading, g.throttle = math.Pi/2, 2
		}
		exploded := false
		for i := 0; i < 60*20 && !exploded; i++ {
			g.step()
			for _, c := range g.charges {
				exploded = exploded || (c.used && c.fuse <= 0)
			}
		}
		if !exploded {
			t.Fatalf("move=%v: no depth charge in 20 s", move)
		}
		if damaged := g.hull < 100; damaged == move {
			t.Fatalf("move=%v: hull %.0f after the first charge, sub at %.0f,%.0f", move, g.hull, g.subX, g.subY)
		}
	}
}

// The assets drawn by assets/gen.go match the constants it copies.
func TestAssetSizes(t *testing.T) {
	for _, c := range []struct {
		name string
		w, h int
		got  interface{ Bounds() image.Rectangle }
	}{
		{"mask", ScreenWidth, ScreenHeight, imageMask},
		{"waves", stripWidth, ScreenHeight - horizon, imageWaves},
		{"clouds", stripWidth, imageClouds.Bounds().Dy(), imageClouds},
		{"sonar", 2*sonarRadius + 1, 2*sonarRadius + 1, imageSonar},
	} {
		if b := c.got.Bounds(); b.Dx() != c.w || b.Dy() != c.h {
			t.Fatalf("%s is %dx%d, want %dx%d; run go generate", c.name, b.Dx(), b.Dy(), c.w, c.h)
		}
	}
}
