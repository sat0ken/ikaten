package bomb

import (
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

// empty is a round with no blocks and the computer gophers out of the way;
// with them gone, a step would end the round, so tests call the parts.
func empty() *Game {
	g := NewGame()
	g.newMatch()
	g.setMode(modePlay)
	for r := range rows {
		for c := range cols {
			if g.grid[r][c] == block {
				g.grid[r][c] = floor
				g.items[r][c] = noItem
			}
		}
	}
	for i := 1; i < players; i++ {
		g.gophers[i].alive = false
	}
	return g
}

func place(p *gopher, c, r int) { p.x, p.y = c*unit+half, r*unit+half }

func TestArena(t *testing.T) {
	g := NewGame()
	floors := 0
	for r := range rows {
		for c := range cols {
			edge := r == 0 || c == 0 || r == rows-1 || c == cols-1
			if (edge || r%2 == 0 && c%2 == 0) != (g.grid[r][c] == wall) {
				t.Fatalf("cell %d,%d is %d", c, r, g.grid[r][c])
			}
			if g.grid[r][c] != wall {
				floors++
			}
		}
	}
	for _, k := range corners {
		for _, d := range [3][2]int{{0, 0}, {1, 0}, {0, 1}} {
			c, r := k[0]+d[0]*sign(cols/2-k[0]), k[1]+d[1]*sign(rows/2-k[1])
			if g.grid[r][c] != floor {
				t.Fatalf("start corner %v is shut at %d,%d", k, c, r)
			}
		}
	}
	if g.spirals != floors {
		t.Fatalf("the closing walls cover %d of %d cells", g.spirals, floors)
	}
}

// Walls stop a gopher at the middle of its cell; pressing toward a gap a
// little off the line slides it into the gap.
func TestWalk(t *testing.T) {
	g := empty()
	p := &g.gophers[0]
	for range 60 {
		g.control(0, input{move: up})
	}
	if p.x != 1*unit+half || p.y != 1*unit+half {
		t.Fatalf("walked into the wall: %d,%d", p.x, p.y)
	}
	for range 30 {
		g.control(0, input{move: right})
	}
	if p.col() != 3 || p.row() != 1 {
		t.Fatalf("walking right ended at cell %d,%d", p.col(), p.row())
	}
	// Off the column of the gap below (column 3), down still finds it.
	place(p, 3, 1)
	p.x += 10
	for range 40 {
		g.control(0, input{move: down})
	}
	if p.col() != 3 || p.row() < 2 {
		t.Fatalf("corner help: at %d,%d (%d,%d)", p.col(), p.row(), p.x, p.y)
	}
}

func TestBlast(t *testing.T) {
	g := empty()
	g.grid[1][5] = block
	g.items[1][5] = fireUp
	g.grid[1][7] = block // behind the first block: untouched
	p := &g.gophers[0]
	place(p, 3, 1)
	g.drop(0)
	place(p, 1, 9)
	// A second bomb in the first one's reach goes off with it.
	g.bombs[1] = bomb{c: 3, r: 3, x: 3*unit + half, y: 3*unit + half, fuse: 999, fire: 2, owner: 1, slide: none, active: true}
	for range fuseTicks + 1 {
		g.updateBombs()
	}
	switch {
	case g.grid[1][5] != floor || !g.shown[1][5]:
		t.Fatal("the near block did not break and show its item")
	case g.grid[1][7] != block:
		t.Fatal("the fire passed a block")
	case g.flames[1][1] == 0 || g.flames[1][2] == 0 || g.flames[3][3] == 0:
		t.Fatal("no fire where it should reach")
	case g.bombs[1].active:
		t.Fatal("the bomb in the fire did not go off")
	case g.flames[5][3] == 0:
		t.Fatal("the chained bomb's fire is missing")
	}
	// A gopher in the fire is caught.
	g = empty()
	p = &g.gophers[0]
	place(p, 1, 1)
	g.drop(0)
	for range fuseTicks {
		g.updateBombs()
	}
	g.updateFire()
	if p.dying == 0 || p.caught != 0 {
		t.Fatalf("dying %d caught by %d", p.dying, p.caught)
	}
}

func TestBombCount(t *testing.T) {
	g := empty()
	p := &g.gophers[0]
	place(p, 1, 1)
	g.drop(0)
	place(p, 3, 1)
	g.drop(0)
	if n := g.countBombs(0); n != 1 {
		t.Fatalf("%d bombs with a count of 1", n)
	}
	p.count = 2
	g.drop(0)
	if n := g.countBombs(0); n != 2 {
		t.Fatalf("%d bombs with a count of 2", n)
	}
}

func (g *Game) countBombs(i int) int {
	n := 0
	for k := range g.bombs {
		if g.bombs[k].active && g.bombs[k].owner == i {
			n++
		}
	}
	return n
}

// A kicked bomb slides until something stops it.
func TestKick(t *testing.T) {
	g := empty()
	g.grid[1][8] = block
	p := &g.gophers[0]
	place(p, 2, 1)
	g.bombs[0] = bomb{c: 3, r: 1, x: 3*unit + half, y: 1*unit + half, fuse: 999, fire: 1, slide: none, active: true}
	p.facing = right
	g.control(0, input{move: none, kick: true})
	for range 120 {
		g.updateBombs()
	}
	b := &g.bombs[0]
	if b.c != 7 || b.x != 7*unit+half || b.slide != none {
		t.Fatalf("bomb at cell %d (%d), sliding %d", b.c, b.x, b.slide)
	}
}

func TestItems(t *testing.T) {
	g := empty()
	p := &g.gophers[0]
	for k, at := range [...]item{fireUp, bombUp, speedUp} {
		c := 2 + k
		g.items[1][c], g.shown[1][c] = at, true
	}
	for range 60 {
		g.control(0, input{move: right})
		g.pickUp()
	}
	if p.fire != 3 || p.count != 2 || p.speed != baseSpeed+1 {
		t.Fatalf("fire %d count %d speed %d", p.fire, p.count, p.speed)
	}
}

// When time runs short the walls close in and crush a gopher that stays.
func TestCloseIn(t *testing.T) {
	g := empty()
	p := &g.gophers[0]
	g.gophers[1].alive = true
	place(&g.gophers[1], cols/2, rows/2+1)
	g.clock = hurryTime
	for range 30 * tps {
		g.step(input{move: none})
		if g.mode != modePlay {
			break
		}
	}
	if p.caught != -1 || g.mode != modeRoundOver || g.winner != 1 {
		t.Fatalf("caught by %d, mode %d, winner %d", p.caught, g.mode, g.winner)
	}
}

// A computer gopher next to a bomb runs from it and lives.
func TestComputerEscapes(t *testing.T) {
	for _, at := range [][2]int{{1, 1}, {5, 3}, {11, 9}} {
		g := empty()
		g.gophers[0].alive = false
		q := &g.gophers[1]
		*q = gopher{alive: true, fire: 2, count: 1, speed: baseSpeed, plan: none}
		place(q, at[0], at[1])
		g.drop(1)
		for range fuseTicks + fireTicks {
			g.updateDanger()
			g.control(1, g.cpu(1))
			g.updateBombs()
			g.updateFire()
		}
		if q.dying > 0 {
			t.Errorf("caught by its own bomb at %v", at)
		}
	}
}

// Four computer gophers play rounds to the end: most have a winner, and
// rounds last about a minute or two.
func TestComputerRounds(t *testing.T) {
	for level := 1; level <= 3; level++ {
		g := NewGame()
		g.level = level
		won, ticks, rounds := 0, 0, 20
		for seed := range rounds {
			g.rng = uint32(seed*7919 + 1)
			g.newMatch()
			g.setMode(modePlay)
			for g.mode == modePlay {
				g.step(g.cpu(0))
				ticks++
			}
			if g.winner >= 0 {
				won++
			}
		}
		avg := float32(ticks) / float32(rounds) / tps
		t.Logf("level %d: %d of %d rounds won, %.0f s a round", level, won, rounds, avg)
		if won < rounds*7/10 || avg > 110 {
			t.Errorf("level %d: %d of %d won, %.0f s", level, won, rounds, avg)
		}
	}
}

// Two rounds won make the champion; the person's gopher goes up a level.
func TestMatch(t *testing.T) {
	g := NewGame()
	g.newMatch()
	for round := 1; round <= 2; round++ {
		g.setMode(modePlay)
		for i := 1; i < players; i++ {
			g.gophers[i].dying = 1
			g.gophers[i].alive = false
		}
		g.step(input{move: none})
		if g.mode != modeRoundOver || g.winner != 0 {
			t.Fatalf("round %d: mode %d winner %d", round, g.mode, g.winner)
		}
		g.modeTicks = 200
		g.step(input{move: none})
	}
	if g.mode != modeMatchOver || g.champion != 0 {
		t.Fatalf("mode %d champion %d", g.mode, g.champion)
	}
}

// keyGopher plays through RunGame with keys: it starts, walks and drops a
// bomb.
type keyGopher struct {
	*Game
	t    *testing.T
	tick int
}

func (k *keyGopher) Update() error {
	k.tick++
	a, rightKey := false, false
	switch k.mode {
	case modeTitle:
		a = k.tick%2 == 0
	case modePlay:
		rightKey = k.modeTicks < 20
		a = k.modeTicks == 30
		if k.modeTicks > 40 {
			return ebiten.Termination
		}
	}
	ebiten.SetKeyState(ebiten.KeyZ, a)
	ebiten.SetKeyState(ebiten.KeyArrowRight, rightKey)
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
	if k.gophers[0].x <= 1*unit+half || k.countBombs(0) != 1 {
		t.Fatalf("x %d, bombs %d", k.gophers[0].x, k.countBombs(0))
	}
}
