package bomb

// The computer gophers plan on the grid. Each tick the danger map gives,
// for each cell, the ticks until fire reaches it; a gopher in danger runs for
// the nearest cell the fire will miss, and otherwise goes for items, blocks
// to break and rival gophers, dropping a bomb only when it can still get
// away from it. Nothing here makes closures or slices, so planning does not
// allocate.

// aggression is the chance in percent that a gopher hunts its rivals, and
// blunders the chance that it wanders instead of following its plan; both
// grow harder with the level.
func (g *Game) aggression() int { return min(30+20*(g.level-1), 85) }
func (g *Game) blunders() int   { return max(20-8*(g.level-1), 3) }

// blast lists in g.cells the cells the fire of a bomb at (c, r) reaches and
// returns how many: the bomb's cell, and out in four directions up to the
// reach, through to the first block, not into a wall.
func (g *Game) blast(c, r, reach int) int {
	n := 0
	g.cells[n] = r*cols + c
	n++
	for d := down; d < none; d++ {
		for k := 1; k <= reach; k++ {
			x, y := c+dirX[d]*k, r+dirY[d]*k
			if g.grid[y][x] == wall {
				break
			}
			g.cells[n] = y*cols + x
			n++
			if g.grid[y][x] == block {
				break
			}
		}
	}
	return n
}

// updateDanger rebuilds the danger map. A bomb in another's fire goes off
// with it.
func (g *Game) updateDanger() {
	var fuse [maxBombs]int
	for i := range g.bombs {
		fuse[i] = g.bombs[i].fuse
	}
	for range 3 {
		for i := range g.bombs {
			b := &g.bombs[i]
			if !b.active {
				continue
			}
			n := g.blast(b.c, b.r, b.fire)
			for _, at := range g.cells[:n] {
				if k := g.bombAt(at%cols, at/cols); k >= 0 && fuse[k] > fuse[i] {
					fuse[k] = fuse[i]
				}
			}
		}
	}
	g.danger = [rows][cols]int{}
	for i := range g.bombs {
		b := &g.bombs[i]
		if b.active {
			g.addDanger(&g.danger, b.c, b.r, b.fire, max(fuse[i], 1))
		}
	}
	// The next walls to close in, once they start, or soon.
	if g.clock > hurryTime+5*tps {
		return
	}
	for k := range 8 {
		if g.fallen+k >= g.spirals {
			break
		}
		at := g.spiral[g.fallen+k]
		c, r := at%cols, at/cols
		t := max(g.clock-hurryTime, 0) + g.fallIn + k*fallTicks
		if g.danger[r][c] == 0 || t < g.danger[r][c] {
			g.danger[r][c] = max(t, 1)
		}
	}
}

// addDanger marks the fire of a bomb that goes off in t ticks.
func (g *Game) addDanger(danger *[rows][cols]int, c, r, reach, t int) {
	n := g.blast(c, r, reach)
	for _, at := range g.cells[:n] {
		x, y := at%cols, at/cols
		if danger[y][x] == 0 || t < danger[y][x] {
			danger[y][x] = t
		}
	}
}

// What a search looks for.
type goal uint8

const (
	goalSafe   goal = iota // a cell the fire misses
	goalItem               // an item lying uncovered
	goalRival              // a cell from which a bomb would catch a rival
	goalBlock              // a cell from which a bomb would break a block
	goalMiddle             // the middle of the arena
)

// reached reports whether gopher i at (c, r) finds what it looks for.
func (g *Game) reached(i int, k goal, c, r int, danger *[rows][cols]int) bool {
	if danger[r][c] > 0 || g.flames[r][c] > 0 {
		return false
	}
	p := &g.gophers[i]
	here := c == p.col() && r == p.row()
	switch k {
	case goalItem:
		return g.shown[r][c] && g.crumble[r][c] == 0
	case goalRival:
		return !here && g.rivalInLine(i, c, r)
	case goalBlock:
		return !here && g.blockInReach(i, c, r)
	case goalMiddle:
		return abs(c-cols/2)+abs(r-rows/2) <= 2
	}
	return true
}

// path searches outward from gopher i's cell, moving a cell every step
// ticks through cells it may enter and will not burn in while it passes,
// for the nearest cell with what it looks for. It returns the first move and
// the number of cells, or none and -1.
func (g *Game) path(i int, k goal, danger *[rows][cols]int) (dir, int) {
	p := &g.gophers[i]
	c, r := p.col(), p.row()
	step := unit / p.speed
	for y := range rows {
		for x := range cols {
			g.dist[y][x] = -1
		}
	}
	g.dist[r][c] = 0
	g.first[r][c] = none
	head, tail := 0, 0
	g.queue[tail] = r*cols + c
	tail++
	for head < tail {
		at := g.queue[head]
		head++
		x, y := at%cols, at/cols
		if g.reached(i, k, x, y, danger) {
			return g.first[y][x], g.dist[y][x]
		}
		for d := down; d < none; d++ {
			nx, ny := x+dirX[d], y+dirY[d]
			if !g.open(nx, ny) || g.dist[ny][nx] >= 0 {
				continue
			}
			t := (g.dist[y][x] + 1) * step
			// Not through fire that still burns, nor where the fire comes
			// before the gopher is through.
			if g.flames[ny][nx] > t-step || danger[ny][nx] > 0 && danger[ny][nx] <= t+step {
				continue
			}
			g.dist[ny][nx] = g.dist[y][x] + 1
			g.first[ny][nx] = g.first[y][x]
			if g.first[ny][nx] == none {
				g.first[ny][nx] = d
			}
			g.queue[tail] = ny*cols + nx
			tail++
		}
	}
	return none, -1
}

// rivalInLine reports whether another gopher stands where a bomb dropped at
// (c, r) by gopher i would burn it.
func (g *Game) rivalInLine(i, c, r int) bool {
	n := g.blast(c, r, g.gophers[i].fire)
	for _, at := range g.cells[:n] {
		if k := g.gopherAt(at%cols, at/cols); k >= 0 && k != i {
			return true
		}
	}
	return false
}

// blockInReach reports whether a bomb dropped at (c, r) would break a block.
func (g *Game) blockInReach(i, c, r int) bool {
	n := g.blast(c, r, g.gophers[i].fire)
	for _, at := range g.cells[:n] {
		if g.grid[at/cols][at%cols] == block {
			return true
		}
	}
	return false
}

// canEscape reports whether gopher i could drop a bomb where it stands and
// still reach a cell the fire misses.
func (g *Game) canEscape(i int) bool {
	p := &g.gophers[i]
	c, r := p.col(), p.row()
	g.trial = g.danger
	t := fuseTicks
	if g.trial[r][c] > 0 {
		t = g.trial[r][c] // an earlier bomb sets it off
	}
	g.addDanger(&g.trial, c, r, p.fire, t)
	// The new bomb shuts its cell to the search, but the gopher is already
	// on it.
	// Keep time in hand: another bomb may shut the way or set this one off.
	_, n := g.path(i, goalSafe, &g.trial)
	return n >= 0 && (n+3)*unit/p.speed < t
}

// cpu decides what computer gopher i does this tick.
func (g *Game) cpu(i int) input {
	p := &g.gophers[i]
	c, r := p.col(), p.row()
	in := input{move: none}
	if (p.x != c*unit+half || p.y != r*unit+half) && p.plan != none {
		// Finish the step to the middle of a cell before thinking again,
		// or turn back if a bomb dropped ahead shuts the way.
		d := p.plan
		past := (p.x-c*unit-half)*dirX[d] + (p.y-r*unit-half)*dirY[d]
		if past >= 0 && !g.open(c+dirX[d], r+dirY[d]) {
			p.plan = back[d]
		}
		in.move = p.plan
		return in
	}
	if g.danger[r][c] > 0 || g.flames[r][c] > 0 {
		p.plan, _ = g.path(i, goalSafe, &g.danger)
		in.move = p.plan
		return in
	}
	// Drop a bomb at a rival or a block, if there is a way out.
	hunt := g.random(100) < g.aggression()
	if (g.rivalInLine(i, c, r) || g.blockInReach(i, c, r)) && g.canEscape(i) {
		in.drop = true
		// Move on the next tick, which sees the new bomb.
		p.plan = none
		return in
	}
	if g.random(100) < g.blunders() {
		// Wander to a safe neighbor.
		d := dir(g.random(4))
		p.plan = none
		if nc, nr := c+dirX[d], r+dirY[d]; g.open(nc, nr) && g.danger[nr][nc] == 0 && g.flames[nr][nc] == 0 {
			p.plan = d
		}
		in.move = p.plan
		return in
	}
	// Items close by first, then rivals, then blocks, else the middle.
	d, n := g.path(i, goalItem, &g.danger)
	if n < 0 || n > 8 {
		n = -1
		if hunt {
			d, n = g.path(i, goalRival, &g.danger)
		}
		if n < 0 {
			d, n = g.path(i, goalBlock, &g.danger)
		}
		if n < 0 {
			// No blocks left in reach: go after the rivals.
			d, n = g.path(i, goalRival, &g.danger)
		}
		if n < 0 {
			d, _ = g.path(i, goalMiddle, &g.danger)
		}
	}
	p.plan, in.move = d, d
	return in
}
