// Package bomb is an original bomb maze battle starring the Go gopher: four
// gophers in a walled arena drop bombs that blow up in a cross, break the
// blocks, pick up the items the blocks hid, and the last gopher standing wins
// the round. Win two rounds to win the match, and the next level's
// computer gophers play smarter. It uses only the Ebitengine API, and every
// image is drawn by code, so it also builds with Ebitengine.
//
// The Go gopher was designed by Renée French.
package bomb

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	tps          = 60

	cols, rows = 13, 11
	tile       = 9  // pixels
	unit       = 48 // position units per tile; a gopher's center is in units
	half       = unit / 2

	players   = 4
	maxBombs  = 32
	fuseTicks = 150 // 2.5 s
	fireTicks = 30
	kickSpeed = 6 // units per tick
	baseSpeed = 3
	maxSpeed  = 6
	maxFire   = 8
	maxCount  = 8
	roundTime = 120 * tps
	hurryTime = 30 * tps // the walls start closing in
	fallTicks = 16       // between falling walls
	winsToWin = 2
	dieTicks  = 60
)

type cell uint8

const (
	floor cell = iota
	wall       // never breaks
	block      // breaks, and may hide an item
)

type item uint8

const (
	noItem item = iota
	fireUp
	bombUp
	speedUp
)

type dir uint8

const (
	down dir = iota
	up
	left
	right
	none
)

var dirX = [5]int{0, 0, -1, 1, 0}
var dirY = [5]int{1, -1, 0, 0, 0}
var back = [5]dir{up, down, right, left, none}

// gopher is a player: the person's is gophers[0], the computer's the others.
type gopher struct {
	x, y   int // center, in units
	facing dir
	moving bool
	alive  bool
	dying  int // ticks since caught by the fire
	caught int // whose fire caught it
	fire   int // blast reach in tiles
	count  int // bombs at once
	speed  int // units per tick
	wins   int
	// The computer: where it is going and when it thinks again.
	plan  dir
	think int
}

func (p *gopher) col() int { return p.x / unit }
func (p *gopher) row() int { return p.y / unit }

type bomb struct {
	c, r   int
	x, y   int // center, in units; a kicked bomb slides
	fuse   int
	owner  int
	fire   int
	slide  dir
	active bool
}

type mode uint8

const (
	modeTitle mode = iota
	modeReady      // the round is about to start
	modePlay
	modeRoundOver
	modeMatchOver
)

type Game struct {
	mode             mode
	ticks, modeTicks int
	rng              uint32

	grid    [rows][cols]cell
	items   [rows][cols]item // under blocks, or lying on the floor
	shown   [rows][cols]bool // an item lies uncovered
	flames  [rows][cols]int  // ticks of fire left
	burner  [rows][cols]int8 // whose bomb set each fire
	crumble [rows][cols]int  // ticks of a breaking block left
	bombs   [maxBombs]bomb
	gophers [players]gopher

	level    int
	round    int
	clock    int // ticks left in the round
	winner   int // of the round, -1 for a draw
	champion int // of the match

	// The computer's view of the danger, rebuilt each tick: ticks until
	// the fire reaches each cell, or 0 if it does not.
	// The closing walls: the cells in the order they fill, how many have
	// fallen, and the ticks to the next.
	spiral  [cols * rows]int
	spirals int
	fallen  int
	fallIn  int

	danger [rows][cols]int
	trial  [rows][cols]int // the danger with a bomb the gopher thinks of dropping
	queue  [cols * rows]int
	dist   [rows][cols]int
	first  [rows][cols]dir // the first move on the way to each cell
	cells  [1 + 4*maxFire]int
}

func NewGame() *Game {
	g := &Game{rng: 2463534242, level: 1}
	g.newMatch()
	g.mode = modeTitle
	return g
}

// random returns 0 to n-1 from a xorshift generator, the same on every platform.
func (g *Game) random(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

// corners are where the gophers start.
var corners = [players][2]int{{1, 1}, {cols - 2, rows - 2}, {cols - 2, 1}, {1, rows - 2}}

func (g *Game) newMatch() {
	g.round = 0
	for i := range g.gophers {
		g.gophers[i].wins = 0
	}
	g.newRound()
}

// newRound builds the arena: walls around and on every other cell, blocks on
// most of the rest, but none in the corners where the gophers start.
func (g *Game) newRound() {
	g.round++
	g.clock = roundTime
	g.winner = -1
	clear(g.bombs[:])
	g.flames, g.crumble = [rows][cols]int{}, [rows][cols]int{}
	g.items, g.shown = [rows][cols]item{}, [rows][cols]bool{}
	for r := range rows {
		for c := range cols {
			switch {
			case r == 0 || c == 0 || r == rows-1 || c == cols-1 || r%2 == 0 && c%2 == 0:
				g.grid[r][c] = wall
			case g.nearCorner(c, r) || g.random(100) >= 72:
				g.grid[r][c] = floor
			default:
				g.grid[r][c] = block
				if n := g.random(100); n < 16 {
					g.items[r][c] = fireUp
				} else if n < 30 {
					g.items[r][c] = bombUp
				} else if n < 36 {
					g.items[r][c] = speedUp
				}
			}
		}
	}
	g.makeSpiral()
	for i := range g.gophers {
		c, r := corners[i][0], corners[i][1]
		w := g.gophers[i].wins
		g.gophers[i] = gopher{x: c*unit + half, y: r*unit + half, alive: true, fire: 2, count: 1, speed: baseSpeed, wins: w, plan: none}
	}
	g.setMode(modeReady)
}

// nearCorner reports whether a cell is within a step of a starting corner.
func (g *Game) nearCorner(c, r int) bool {
	for _, k := range corners {
		dc, dr := c-k[0], r-k[1]
		if dc*dc+dr*dr <= 1 || (dc == 0 && abs(dr) == 2) || (dr == 0 && abs(dc) == 2) {
			return true
		}
	}
	return false
}

func abs(x int) int { return max(x, -x) }

func (g *Game) bombAt(c, r int) int {
	for i := range g.bombs {
		if b := &g.bombs[i]; b.active && b.c == c && b.r == r {
			return i
		}
	}
	return -1
}

// open reports whether a gopher may enter a cell.
func (g *Game) open(c, r int) bool {
	return c >= 0 && r >= 0 && c < cols && r < rows && g.grid[r][c] == floor && g.bombAt(c, r) < 0
}

func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// input is what the person does in one tick.
type input struct {
	move       dir
	drop, kick bool
}

func readInput() input {
	in := input{move: none}
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		in.move = up
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		in.move = down
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		in.move = left
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		in.move = right
	}
	in.drop = pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace)
	in.kick = pressed(ebiten.KeyX)
	return in
}

func startPressed() bool {
	return pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace) || pressed(ebiten.KeyEnter)
}

func (g *Game) Update() error {
	g.ticks++
	switch g.mode {
	case modeTitle:
		g.modeTicks++
		if startPressed() {
			g.level = 1
			g.newMatch()
		}
	case modeMatchOver:
		g.modeTicks++
		if g.modeTicks > 60 && startPressed() {
			if g.champion == 0 {
				g.level++
				g.newMatch()
			} else {
				g.setMode(modeTitle)
			}
		}
	default:
		g.step(readInput())
	}
	return nil
}

// step advances the round one tick.
func (g *Game) step(in input) {
	g.modeTicks++
	switch g.mode {
	case modeReady:
		if g.modeTicks > 90 {
			g.setMode(modePlay)
		}
		return
	case modeRoundOver:
		g.updateFire()
		if g.modeTicks > 150 {
			g.nextRound()
		}
		return
	}
	g.clock--
	if g.clock <= hurryTime {
		g.closeIn()
	}
	g.updateDanger()
	for i := range g.gophers {
		p := &g.gophers[i]
		if !p.alive {
			continue
		}
		if p.dying > 0 {
			p.dying++
			if p.dying > dieTicks {
				p.alive = false
			}
			continue
		}
		if i == 0 {
			g.control(i, in)
		} else {
			g.control(i, g.cpu(i))
		}
	}
	g.updateBombs()
	g.updateFire()
	g.pickUp()
	g.checkRound()
}

// control moves a gopher and works its bomb and kick.
func (g *Game) control(i int, in input) {
	p := &g.gophers[i]
	if in.drop {
		g.drop(i)
	}
	if in.kick {
		g.kick(i)
	}
	p.moving = in.move != none
	if p.moving {
		p.facing = in.move
		g.walk(p, in.move)
	}
}

// walk moves a gopher along the grid. Off the line of the move it first
// slides back to the middle of its row or column, so that turning into a gap
// is forgiving; it stops at the middle of a cell whose next cell is shut.
func (g *Game) walk(p *gopher, d dir) {
	c, r := p.col(), p.row()
	cx, cy := c*unit+half, r*unit+half
	dx, dy := dirX[d], dirY[d]
	step := p.speed
	if dx != 0 && p.y != cy {
		// Corner help: slide toward the row's middle if the way on is open.
		if !g.open(c+dx, r) {
			return
		}
		p.y += sign(cy-p.y) * min(step, abs(cy-p.y))
		return
	}
	if dy != 0 && p.x != cx {
		if !g.open(c, r+dy) {
			return
		}
		p.x += sign(cx-p.x) * min(step, abs(cx-p.x))
		return
	}
	pos, mid := p.x, cx
	if dy != 0 {
		pos, mid = p.y, cy
	}
	n := pos + (dx+dy)*step
	ahead := dx + dy
	if !g.open(c+dx, r+dy) && (n-mid)*ahead > 0 {
		n = mid // stop at the middle
		if (pos-mid)*ahead > 0 {
			n = pos
		}
	}
	if dy != 0 {
		p.y = n
	} else {
		p.x = n
	}
}

func sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	}
	return 0
}

// drop puts a bomb where the gopher stands, if it has one left.
func (g *Game) drop(i int) {
	p := &g.gophers[i]
	c, r := p.col(), p.row()
	if g.bombAt(c, r) >= 0 || g.grid[r][c] != floor {
		return
	}
	n := 0
	for k := range g.bombs {
		if g.bombs[k].active && g.bombs[k].owner == i {
			n++
		}
	}
	if n >= p.count {
		return
	}
	for k := range g.bombs {
		if !g.bombs[k].active {
			g.bombs[k] = bomb{c: c, r: r, x: c*unit + half, y: r*unit + half, fuse: fuseTicks, owner: i, fire: p.fire, slide: none, active: true}
			return
		}
	}
}

// kick sends the bomb just ahead of the gopher sliding until it hits
// something.
func (g *Game) kick(i int) {
	p := &g.gophers[i]
	d := p.facing
	k := g.bombAt(p.col()+dirX[d], p.row()+dirY[d])
	if k < 0 {
		return
	}
	g.bombs[k].slide = d
}

// updateBombs slides kicked bombs, burns the fuses and sets off the bombs.
func (g *Game) updateBombs() {
	for i := range g.bombs {
		b := &g.bombs[i]
		if !b.active {
			continue
		}
		if b.slide != none {
			g.slideBomb(b)
		}
		b.fuse--
		if b.fuse <= 0 {
			g.explode(i)
		}
	}
}

func (g *Game) slideBomb(b *bomb) {
	dx, dy := dirX[b.slide], dirY[b.slide]
	cx, cy := b.c*unit+half, b.r*unit+half
	// Stop in the middle of the cell when the next one is shut.
	past := (b.x-cx)*dx + (b.y-cy)*dy
	if past+kickSpeed > 0 && (!g.open(b.c+dx, b.r+dy) || g.gopherAt(b.c+dx, b.r+dy) >= 0) {
		b.x, b.y, b.slide = cx, cy, none
		return
	}
	b.x += dx * kickSpeed
	b.y += dy * kickSpeed
	b.c, b.r = b.x/unit, b.y/unit
}

// gopherAt returns a living gopher standing in a cell, or -1.
func (g *Game) gopherAt(c, r int) int {
	for i := range g.gophers {
		if p := &g.gophers[i]; p.alive && p.dying == 0 && p.col() == c && p.row() == r {
			return i
		}
	}
	return -1
}

// explode sets off bomb i: fire runs out in four directions until a wall,
// burns the first block it meets, and sets off the bombs it reaches.
func (g *Game) explode(i int) {
	b := &g.bombs[i]
	b.active = false
	g.burn(b.c, b.r, b.owner)
	for d := down; d < none; d++ {
		for n := 1; n <= b.fire; n++ {
			c, r := b.c+dirX[d]*n, b.r+dirY[d]*n
			if g.grid[r][c] == wall {
				break
			}
			if g.grid[r][c] == block {
				g.grid[r][c] = floor
				g.crumble[r][c] = fireTicks
				g.shown[r][c] = g.items[r][c] != noItem
				break
			}
			g.burn(c, r, b.owner)
			if k := g.bombAt(c, r); k >= 0 {
				g.bombs[k].fuse = min(g.bombs[k].fuse, 1)
			}
		}
	}
}

// burn sets a cell on fire; an item lying there is lost.
func (g *Game) burn(c, r, owner int) {
	g.flames[r][c] = fireTicks
	g.burner[r][c] = int8(owner)
	if g.shown[r][c] && g.crumble[r][c] == 0 {
		g.shown[r][c], g.items[r][c] = false, noItem
	}
}

// updateFire burns out the flames and catches gophers in them.
func (g *Game) updateFire() {
	for r := range rows {
		for c := range cols {
			if g.flames[r][c] > 0 {
				g.flames[r][c]--
			}
			if g.crumble[r][c] > 0 {
				g.crumble[r][c]--
			}
		}
	}
	for i := range g.gophers {
		p := &g.gophers[i]
		if p.alive && p.dying == 0 && g.flames[p.row()][p.col()] > 0 {
			p.dying = 1
			p.caught = int(g.burner[p.row()][p.col()])
		}
	}
}

// pickUp gives a gopher the item it walks onto.
func (g *Game) pickUp() {
	for i := range g.gophers {
		p := &g.gophers[i]
		if !p.alive || p.dying > 0 {
			continue
		}
		c, r := p.col(), p.row()
		if !g.shown[r][c] || g.crumble[r][c] > 0 {
			continue
		}
		switch g.items[r][c] {
		case fireUp:
			p.fire = min(p.fire+1, maxFire)
		case bombUp:
			p.count = min(p.count+1, maxCount)
		case speedUp:
			p.speed = min(p.speed+1, maxSpeed)
		}
		g.shown[r][c], g.items[r][c] = false, noItem
	}
}

// standing counts the gophers not caught, and returns one of them.
func (g *Game) standing() (n, last int) {
	last = -1
	for i := range g.gophers {
		if p := &g.gophers[i]; p.alive && p.dying == 0 {
			n++
			last = i
		}
	}
	return n, last
}

// checkRound ends the round when one gopher or none is left, once the
// fallen have fallen, or when time runs out.
func (g *Game) checkRound() {
	n, last := g.standing()
	falling := false
	for i := range g.gophers {
		if g.gophers[i].dying > 0 && g.gophers[i].alive {
			falling = true
		}
	}
	switch {
	case n <= 1 && !falling:
		g.winner = last
	case g.clock <= 0:
		g.winner = -1
	default:
		return
	}
	if g.winner >= 0 {
		g.gophers[g.winner].wins++
	}
	g.setMode(modeRoundOver)
}

// makeSpiral lists the floor cells from the outside in, clockwise from the
// top left: the order in which the walls close in.
func (g *Game) makeSpiral() {
	g.spirals = 0
	left, top, right, bottom := 1, 1, cols-2, rows-2
	for left <= right && top <= bottom {
		for c := left; c <= right; c++ {
			g.addSpiral(c, top)
		}
		for r := top + 1; r <= bottom; r++ {
			g.addSpiral(right, r)
		}
		if top < bottom {
			for c := right - 1; c >= left; c-- {
				g.addSpiral(c, bottom)
			}
		}
		if left < right {
			for r := bottom - 1; r > top; r-- {
				g.addSpiral(left, r)
			}
		}
		left, top, right, bottom = left+1, top+1, right-1, bottom-1
	}
	g.fallen, g.fallIn = 0, fallTicks
}

func (g *Game) addSpiral(c, r int) {
	if g.grid[r][c] != wall {
		g.spiral[g.spirals] = r*cols + c
		g.spirals++
	}
}

// closeIn drops the next wall: it crushes a gopher, a bomb or an item there.
func (g *Game) closeIn() {
	g.fallIn--
	if g.fallIn > 0 || g.fallen >= g.spirals {
		return
	}
	g.fallIn = fallTicks
	at := g.spiral[g.fallen]
	g.fallen++
	c, r := at%cols, at/cols
	g.grid[r][c] = wall
	g.shown[r][c], g.items[r][c], g.flames[r][c] = false, noItem, 0
	if k := g.bombAt(c, r); k >= 0 {
		g.bombs[k].active = false
	}
	for i := range g.gophers {
		if p := &g.gophers[i]; p.alive && p.dying == 0 && p.col() == c && p.row() == r {
			p.dying, p.caught = 1, -1
		}
	}
}

// nextRound starts the next round, or ends the match.
func (g *Game) nextRound() {
	for i := range g.gophers {
		if g.gophers[i].wins >= winsToWin {
			g.champion = i
			g.setMode(modeMatchOver)
			return
		}
	}
	g.newRound()
}
