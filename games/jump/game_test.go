package jump

import (
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

func playing(stage int) *Game {
	g := NewGame()
	g.livesLeft = lives
	g.startStage(stage)
	g.setMode(modePlay)
	return g
}

// standOn puts the gopher on the ground at a column, settled.
func standOn(g *Game, c int) {
	// From under the cave's roof, so as not to land on top of it.
	g.x, g.y, g.vx, g.vy = float32(c*tile)+1, 3*tile, 0, 0
	for range 200 {
		g.move(input{})
		if g.onGround {
			break
		}
	}
}

// A full jump rises about four and a half tiles; a quick tap less; a
// running jump crosses five tiles of gap, and with a dash eight.
func TestJump(t *testing.T) {
	g := playing(0)
	standOn(g, 3)
	top := g.y
	g.move(input{jump: true, jumpHeld: true})
	for g.vy < 0 {
		g.move(input{jumpHeld: true})
		top = min(top, g.y)
	}
	full := (g.checkY - top) / tile
	g = playing(0)
	standOn(g, 3)
	low := g.y
	g.move(input{jump: true, jumpHeld: true})
	for g.vy < 0 {
		g.move(input{})
		low = min(low, g.y)
	}
	tap := (g.checkY - low) / tile
	t.Logf("full jump %.1f tiles, tap %.1f tiles", full, tap)
	if full < 4 || full > 5 || tap > 2 {
		t.Fatalf("full jump %.1f tiles, tap %.1f", full, tap)
	}
}

// Walls stop the gopher; a dash breaks dirt.
func TestWallsAndDirt(t *testing.T) {
	g := playing(0)
	// The dirt wall in the meadow, at column 56.
	standOn(g, 55)
	for range 60 {
		g.move(input{right: true})
	}
	if g.x+boxW > 56*tile+0.01 {
		t.Fatalf("walked into the dirt wall: x %.1f", g.x)
	}
	g.move(input{right: true, dash: true})
	for range dashTicks {
		g.move(input{right: true})
	}
	if g.level[11][56] != ' ' || g.x < 56*tile {
		t.Fatalf("dash did not break through: tile %q, x %.1f", g.level[11][56], g.x)
	}
}

// Landing on a bug beats it and bounces; walking into one costs a heart;
// a spiky bug hurts a stomper; a dash beats even a spiky bug.
func TestEnemies(t *testing.T) {
	g := playing(0)
	g.enemies[0] = enemy{kind: bug, x: 100, y: 88, alive: true}
	g.x, g.y, g.vy = 101, 82, 2
	g.touch()
	if g.enemies[0].squashed == 0 || g.vy >= 0 || g.heartsLeft != hearts {
		t.Fatalf("stomp: squashed %d vy %.1f hearts %d", g.enemies[0].squashed, g.vy, g.heartsLeft)
	}
	g = playing(0)
	g.enemies[0] = enemy{kind: bug, x: 100, y: 88, alive: true}
	g.x, g.y, g.vy = 96, 86, 0
	g.touch()
	if g.heartsLeft != hearts-1 || g.enemies[0].squashed != 0 {
		t.Fatalf("side: hearts %d", g.heartsLeft)
	}
	g = playing(0)
	g.enemies[0] = enemy{kind: spiky, x: 100, y: 88, alive: true}
	g.x, g.y, g.vy = 101, 82, 2
	g.touch()
	if g.heartsLeft != hearts-1 {
		t.Fatalf("spiky stomp: hearts %d", g.heartsLeft)
	}
	g = playing(0)
	g.enemies[0] = enemy{kind: spiky, x: 100, y: 88, alive: true}
	g.x, g.y, g.dash = 96, 86, 5
	g.touch()
	if g.enemies[0].squashed == 0 || g.heartsLeft != hearts {
		t.Fatalf("dash: squashed %d hearts %d", g.enemies[0].squashed, g.heartsLeft)
	}
}

// Three hits, a fall into a pit, or spikes cost lives; the checkpoint is
// where the gopher comes back.
func TestLivesAndCheckpoint(t *testing.T) {
	g := playing(0)
	for range hearts {
		g.hurt = 0
		g.hit()
	}
	if g.mode != modeDead || g.livesLeft != lives-1 {
		t.Fatalf("mode %d lives %d", g.mode, g.livesLeft)
	}
	g = playing(0)
	g.x, g.y = 78*tile+1, 11*tile-boxH+tile
	g.touch()
	if g.level[11][78] != 'c' || g.checkX != 78*tile+1 {
		t.Fatalf("checkpoint %q at %.0f", g.level[11][78], g.checkX)
	}
	// Fall into the first pit.
	g.x, g.y, g.vx = 16*tile+1, 90, 0
	for g.mode == modePlay {
		g.step(input{})
	}
	if g.mode != modeDead {
		t.Fatalf("mode %d after the pit", g.mode)
	}
	for g.mode == modeDead {
		g.Update()
	}
	if g.x != 78*tile+1 {
		t.Fatalf("came back at %.0f", g.x)
	}
}

// Gems count, and a hundred give a life.
func TestGems(t *testing.T) {
	g := playing(0)
	g.gems = gemLife - 1
	g.level[5][10] = 'o'
	g.x, g.y = 10*tile+1, 5*tile-2
	g.touch()
	if g.gems != gemLife || g.livesLeft != lives+1 || g.level[5][10] != ' ' {
		t.Fatalf("gems %d lives %d", g.gems, g.livesLeft)
	}
}

// spot is a place the gopher can stand: its column and the row of its feet.
type spot struct{ c, r int }

// reachable searches every stage from its start with the real physics:
// from each spot it walks, and jumps with the button held for several
// lengths, with and without a running start and a dash, and follows where it
// lands. Enemies are left out; spikes and pits end a try.
func reachable(t *testing.T, base *Game) bool {
	clear(base.enemies[:])
	start := spot{int(base.checkX) / tile, int(base.checkY+boxH-1) / tile}
	seen := map[spot]bool{start: true}
	queue := []spot{start}
	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		for _, dir := range [2]float32{-1, 1} {
			for _, run := range [2]float32{0, runSpeed} {
				for _, hold := range [6]int{0, 4, 8, 14, 22, 60} {
					for _, dashAt := range [5]int{-1, 6, 14, 22, 30} {
						if hold == 0 && dashAt >= 0 && run == 0 {
							continue
						}
						land, goal := try(base, s, dir, run*dir, hold, dashAt)
						if goal {
							return true
						}
						if land != nil && !seen[*land] {
							seen[*land] = true
							queue = append(queue, *land)
						}
					}
				}
			}
		}
	}
	t.Logf("stage %d: %d spots reached", base.stage, len(seen))
	return false
}

// try plays one move from a spot and returns where the gopher lands, if it
// lands safely somewhere else, and whether it reached the goal.
func try(base *Game, s spot, dir, vx float32, hold, dashAt int) (*spot, bool) {
	g := *base
	g.x, g.y = float32(s.c*tile)+1, float32((s.r+1)*tile-boxH)
	g.vx, g.vy, g.onGround, g.sinceGround, g.jumpPressed = vx, 0, true, 0, 99
	left := false
	for tick := range 240 {
		in := input{left: dir < 0, right: dir > 0, jump: tick == 0 && hold > 0, jumpHeld: tick < hold, dash: tick == dashAt}
		g.move(in)
		for _, p := range [4][2]float32{{1, 1}, {boxW - 1, 1}, {1, boxH - 1}, {boxW - 1, boxH - 1}} {
			switch g.tileAt(g.x+p[0], g.y+p[1]) {
			case 'G':
				return nil, true
			case '^':
				return nil, false
			}
		}
		if g.y > rows*tile {
			return nil, false
		}
		if !g.onGround {
			left = true
		}
		c := int(g.x+boxW/2) / tile
		if g.onGround && (left || c != s.c) {
			return &spot{c, int(g.y+boxH-1) / tile}, false
		}
	}
	return nil, false
}

func TestStagesReachable(t *testing.T) {
	for n := range stages {
		if !reachable(t, playing(n)) {
			t.Errorf("stage %d (%s): the goal cannot be reached", n+1, stages[n].name)
		}
	}
	// The search does find a stage that cannot be finished: a pit of
	// fifteen tiles is too wide even with a dash.
	g := playing(0)
	for c := 20; c < 35; c++ {
		for r := range rows {
			g.level[r][c] = ' '
		}
	}
	if reachable(t, g) {
		t.Error("the search crossed a pit of fifteen tiles")
	}
}

// The dirt wall in the meadow cannot be jumped: the stage teaches the dash.
func TestMeadowNeedsDash(t *testing.T) {
	base := playing(0)
	start := spot{130, 11}
	for _, run := range [2]float32{0, runSpeed} {
		for _, hold := range [6]int{4, 8, 14, 22, 40, 60} {
			if land, _ := try(base, start, 1, run, hold, -1); land != nil && land.c > 134 {
				t.Fatalf("jumped the dirt wall to column %d", land.c)
			}
		}
	}
	g := playing(0)
	standOn(g, 132)
	g.move(input{right: true, dash: true})
	for range 30 {
		g.move(input{right: true})
	}
	if g.x < 135*tile {
		t.Fatalf("a dash did not go through: x %.0f", g.x)
	}
}

// Reaching the flag clears the stage, and the last stage ends the game.
func TestClear(t *testing.T) {
	g := playing(2)
	for c := range g.width {
		if g.level[11][c] == 'G' {
			g.x, g.y = float32(c*tile), 11*tile
		}
	}
	g.touch()
	if g.mode != modeClear {
		t.Fatalf("mode %d at the flag", g.mode)
	}
	for range 200 {
		g.Update()
	}
	if g.mode != modeEnd {
		t.Fatalf("mode %d after the last stage", g.mode)
	}
}

// keyGopher plays through RunGame with keys: it starts and runs right.
type keyGopher struct {
	*Game
	t    *testing.T
	tick int
}

func (k *keyGopher) Update() error {
	k.tick++
	a, right := false, false
	switch k.mode {
	case modeTitle:
		a = k.tick%2 == 0
	case modePlay:
		right = true
		a = k.modeTicks == 20
		if k.modeTicks > 60 {
			return ebiten.Termination
		}
	}
	ebiten.SetKeyState(ebiten.KeyZ, a)
	ebiten.SetKeyState(ebiten.KeyArrowRight, right)
	if k.tick > 600 {
		k.t.Error("never played")
		return ebiten.Termination
	}
	return k.Game.Update()
}

func TestKeys(t *testing.T) {
	ebiten.SetTPS(1000)
	defer ebiten.SetTPS(60)
	ebiten.SetHardware(&hardware.Headless{Width: 240, Height: 240})
	k := &keyGopher{Game: NewGame(), t: t}
	if err := ebiten.RunGame(k); err != nil {
		t.Fatal(err)
	}
	ebiten.SetKeyState(ebiten.KeyZ, false)
	ebiten.SetKeyState(ebiten.KeyArrowRight, false)
	if k.x < k.checkX+30 {
		t.Fatalf("x %.0f from %.0f", k.x, k.checkX)
	}
}

// Every letter the screens use has a glyph.
func TestGlyphs(t *testing.T) {
	for _, s := range []string{"GOPHER JUMP", "A: JUMP  B: DASH", "DASH BREAKS DIRT", "PRESS A", "GOPHER BY RENEE FRENCH",
		"STAGE ", "PASS!", "SCORE ", "GAME OVER", "ALL PASS!", "GOPHER X", "X"} {
		for i := range len(s) {
			if imageGlyph[s[i]] == nil {
				t.Errorf("no glyph for %q in %q", s[i], s)
			}
		}
	}
	for _, st := range stages {
		for i := range len(st.name) {
			if imageGlyph[st.name[i]] == nil {
				t.Errorf("no glyph for %q in %q", st.name[i], st.name)
			}
		}
	}
}

// Behind the title the gopher runs the meadow on its own, for the benchmark
// and for show, and the title stays.
func TestTitleDemo(t *testing.T) {
	g := NewGame()
	far := float32(0)
	for range 60 * tps {
		g.Update()
		far = max(far, g.x)
	}
	if g.mode != modeTitle || far < 60*tile {
		t.Fatalf("mode %d, the demo got to column %d", g.mode, int(far)/tile)
	}
}
