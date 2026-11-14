package dungeon

import (
	"testing"

	ebiten "github.com/sat0ken/ikaten"
)

// The screen is 240 by 240, and the view ends above the status bar.
func TestLayout(t *testing.T) {
	g := NewGame()
	if w, h := g.Layout(0, 0); w != 240 || h != 240 {
		t.Fatalf("layout %d by %d", w, h)
	}
}

// Every column of the view finds a wall, at a positive distance, and a slime
// ahead can be shot dead.
func TestWallsAndShooting(t *testing.T) {
	g := NewGame()
	g.mode = modePlay
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	g.Draw(screen)
	for x, d := range g.zbuf {
		if d <= 0 || d > 2*float32(mapSize) {
			t.Fatalf("column %d: distance %v", x, d)
		}
	}
	// A slime three cells straight ahead in the first room dies under the
	// crosshair after its hits.
	s := &g.slimes[0]
	s.x, s.y = g.px+3, g.py
	if x, depth := g.project(s); depth < 2.9 || x < ScreenWidth/2-1 || x > ScreenWidth/2+1 {
		t.Fatalf("slime at column %v, depth %v", x, depth)
	}
	for range s.hp {
		g.shoot()
	}
	if s.alive {
		t.Fatalf("the slime lives with %d hp", s.hp)
	}
	if g.kills != 1 {
		t.Fatalf("kills %d", g.kills)
	}
}

// Drawing a frame allocates nothing, on the title, in play with a slime hurt
// and the player hit, and after.
func TestDrawAllocs(t *testing.T) {
	screen := ebiten.NewImage(ScreenWidth, ScreenHeight)
	g := NewGame()
	for _, m := range []mode{modeTitle, modePlay, modeDead, modeClear} {
		g.mode = m
		g.damage, g.flash, g.slimes[0].hurt = 5, 3, 4
		g.Draw(screen)
		if n := testing.AllocsPerRun(5, func() { g.ticks++; g.Draw(screen) }); n != 0 {
			t.Errorf("mode %d: %v allocations", m, n)
		}
	}
}
