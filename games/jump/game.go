// Package jump is an original side-scrolling platformer starring the Go
// gopher: run, jump and dash through a meadow, a cave and the sky, collect
// gems, stomp the bugs, and reach the PASS flag at the end of each stage. The
// dash breaks dirt blocks and knocks out even the spiky bugs that cannot be
// stomped. It uses only the Ebitengine API, and every image is drawn by code,
// so it also builds with Ebitengine.
//
// The Go gopher was designed by Renée French.
package jump

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	tps          = 60
	tile         = 8 // in the game's pixels
	rows         = 15

	// The game runs in pixels of its own, half the screen's: zoom screen
	// pixels to each way. view is how much of the stage shows across.
	zoom = 2
	view = ScreenWidth / zoom

	// The gopher's box, in pixels, from the left and the top of its
	// position.
	boxW, boxH = 6, 10

	runSpeed    = 1.3 // pixels per tick
	runAccel    = 0.09
	airAccel    = 0.06
	runDecel    = 0.14
	gravity     = 0.18
	maxFall     = 3.6
	jumpSpeed   = 3.6
	cutJump     = 0.45 // a released jump keeps this much of its rise
	coyote      = 6    // ticks after leaving an edge that a jump still works
	buffer      = 6    // ticks before landing that a jump press counts
	dashSpeed   = 3.0
	dashTicks   = 12
	stompBounce = 2.8
	hurtTicks   = 90 // ticks of blinking after a hit
	hearts      = 3
	lives       = 3
	maxEnemies  = 24
	gemLife     = 100 // gems for a life
)

type mode uint8

const (
	modeTitle mode = iota
	modeIntro      // the stage's name
	modePlay
	modeDead  // the gopher lost a life
	modeClear // the stage is done
	modeOver  // no lives left
	modeEnd   // all stages done
)

type enemyKind uint8

const (
	bug   enemyKind = iota // walks and turns at walls and edges
	moth                   // flies in a wave
	spiky                  // walks; spikes on its back hurt a stomping gopher
)

type enemy struct {
	kind     enemyKind
	x, y     float32
	vx       float32
	baseY    float32 // the moth's line
	alive    bool
	squashed int // ticks since it was beaten
}

type Game struct {
	art *sprites

	mode             mode
	ticks, modeTicks int

	stage int
	level [rows][maxWidth]byte
	width int

	// The gopher: the top left of its box, its speed, and its state.
	x, y, vx, vy float32
	facing       float32
	onGround     bool
	sinceGround  int // ticks since it stood on something
	jumpPressed  int // ticks since A was pressed, for the jump buffer
	jumping      bool
	dash         int // ticks left of a dash
	dashUsed     bool
	hurt         int
	heartsLeft   int
	livesLeft    int
	gems         int
	score        int
	checkX       float32 // where to come back after losing a life
	checkY       float32
	camera       float32

	enemies [maxEnemies]enemy

	// The screen the whole of which has been filled, so that it is opaque.
	filled *ebiten.Image
	back   *backdrop
}

func NewGame() *Game {
	// Draw covers the whole screen anyway.
	ebiten.SetScreenClearedEveryFrame(false)
	g := &Game{art: newSprites(), back: newBackdrop()}
	g.startStage(0)
	g.mode = modeTitle
	return g
}

func (g *Game) setMode(m mode) { g.mode, g.modeTicks = m, 0 }

// startStage loads a stage from its map: the gopher starts at 'P', enemies
// at 'b', 'm' and 's'.
func (g *Game) startStage(n int) {
	g.stage = n
	g.width = len(stages[n].rows[0])
	clear(g.enemies[:])
	k := 0
	for r := range rows {
		for c := range g.width {
			ch := stages[n].rows[r][c]
			g.level[r][c] = ch
			x, y := float32(c*tile), float32(r*tile)
			switch ch {
			case 'P':
				g.checkX, g.checkY = x+1, y+tile-boxH
				g.level[r][c] = ' '
			case 'b', 'm', 's':
				if k < maxEnemies {
					e := enemy{x: x, y: y, vx: -0.4, baseY: y, alive: true}
					switch ch {
					case 'm':
						e.kind, e.vx = moth, -0.5
					case 's':
						e.kind = spiky
					}
					g.enemies[k] = e
					k++
				}
				g.level[r][c] = ' '
			}
		}
	}
	g.heartsLeft = hearts
	g.respawn()
	g.setMode(modeIntro)
}

// respawn puts the gopher back at the start or the last checkpoint.
func (g *Game) respawn() {
	g.x, g.y, g.vx, g.vy = g.checkX, g.checkY, 0, 0
	g.facing = 1
	g.dash, g.hurt, g.jumpPressed = 0, 0, 99
	g.onGround, g.jumping = false, false
	g.camera = max(0, min(g.x-40, float32(g.width*tile-view)))
}

func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// input is what the player does in one tick.
type input struct {
	left, right    bool
	jump, jumpHeld bool
	dash           bool
}

func readInput() input {
	return input{
		left:     ebiten.IsKeyPressed(ebiten.KeyArrowLeft),
		right:    ebiten.IsKeyPressed(ebiten.KeyArrowRight),
		jump:     pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace) || pressed(ebiten.KeyArrowUp),
		jumpHeld: ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp),
		dash:     pressed(ebiten.KeyX),
	}
}

func startPressed() bool {
	return pressed(ebiten.KeyZ) || pressed(ebiten.KeySpace) || pressed(ebiten.KeyEnter)
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	switch g.mode {
	case modeTitle:
		if startPressed() {
			g.livesLeft, g.gems, g.score = lives, 0, 0
			g.startStage(0)
			break
		}
		g.demo()
	case modeIntro:
		if g.modeTicks > 90 {
			g.setMode(modePlay)
		}
	case modePlay:
		g.step(readInput())
	case modeDead:
		if g.modeTicks > 90 {
			if g.livesLeft == 0 {
				g.setMode(modeOver)
			} else {
				g.heartsLeft = hearts
				g.respawn()
				g.setMode(modePlay)
			}
		}
	case modeClear:
		if g.modeTicks > 120 {
			if g.stage+1 < len(stages) {
				g.startStage(g.stage + 1)
			} else {
				g.setMode(modeEnd)
			}
		}
	case modeOver, modeEnd:
		if g.modeTicks > 60 && startPressed() {
			g.setMode(modeTitle)
		}
	}
	return nil
}

// tileAt returns the map at a pixel; beyond the sides is wall, below the
// bottom is open (a pit).
func (g *Game) tileAt(px, py float32) byte {
	c, r := int(floor(px/tile)), int(floor(py/tile))
	if c < 0 || c >= g.width {
		return '#'
	}
	if r < 0 || r >= rows {
		return ' '
	}
	return g.level[r][c]
}

func floor(x float32) float32 {
	i := float32(int(x))
	if i > x {
		i--
	}
	return i
}

// solid reports whether a tile stops the gopher from every side.
func solid(t byte) bool { return t == '#' || t == 'd' || t == '=' }

// step advances play one tick.
func (g *Game) step(in input) {
	g.move(in)
	g.updateEnemies()
	g.touch()
	if g.mode != modePlay {
		return
	}
	// The camera keeps the gopher a little left of the middle.
	want := g.x - 48
	if g.facing < 0 {
		want = g.x - 66
	}
	g.camera += (want - g.camera) * 0.1
	g.camera = max(0, min(g.camera, float32(g.width*tile-view)))
	if g.hurt > 0 {
		g.hurt--
	}
	if g.y > rows*tile+16 {
		g.lose()
	}
}

// move runs, jumps and dashes the gopher, then moves it through the map.
func (g *Game) move(in input) {
	if in.jump {
		g.jumpPressed = 0
	} else {
		g.jumpPressed++
	}
	if in.dash && g.dash == 0 && !g.dashUsed {
		g.dash, g.dashUsed = dashTicks, true
		g.vy = 0
	}
	if g.dash > 0 {
		g.dash--
		g.vx = g.facing * dashSpeed
		g.vy = 0
	} else {
		want := float32(0)
		if in.left {
			want, g.facing = -runSpeed, -1
		}
		if in.right {
			want, g.facing = runSpeed, 1
		}
		accel := float32(runAccel)
		if !g.onGround {
			accel = airAccel
		}
		if want == 0 && g.onGround {
			accel = runDecel
		}
		g.vx = approach(g.vx, want, accel)
		// Jump from the ground, or just after leaving it, or when landing
		// just after the press.
		if g.jumpPressed < buffer && g.sinceGround < coyote && !g.jumping {
			g.vy, g.jumping = -jumpSpeed, true
			g.sinceGround, g.jumpPressed = coyote, buffer
		}
		if g.jumping && !in.jumpHeld && g.vy < 0 {
			g.vy *= cutJump
			g.jumping = false
		}
		g.vy = min(g.vy+gravity, maxFall)
	}
	g.moveX(g.vx)
	g.moveY(g.vy)
	if g.onGround {
		g.sinceGround, g.dashUsed, g.jumping = 0, false, false
	} else {
		g.sinceGround++
	}
}

func approach(x, target, step float32) float32 {
	if x < target {
		return min(x+step, target)
	}
	return max(x-step, target)
}

// moveX moves the gopher across and stops it at a wall. A dash breaks the
// dirt it runs into.
func (g *Game) moveX(dx float32) {
	g.x += dx
	edge := g.x
	if dx > 0 {
		edge = g.x + boxW
	}
	for _, py := range [3]float32{g.y + 1, g.y + boxH/2, g.y + boxH - 1} {
		t := g.tileAt(edge, py)
		if !solid(t) {
			continue
		}
		if t == 'd' && g.dash > 0 {
			g.breakDirt(edge, py)
			continue
		}
		if dx > 0 {
			g.x = floor(edge/tile)*tile - boxW
		} else {
			g.x = floor(edge/tile)*tile + tile
		}
		g.vx = 0
		return
	}
}

// moveY moves the gopher up or down, lands it on the ground and on clouds
// from above, and stops it at a ceiling.
func (g *Game) moveY(dy float32) {
	feet := g.y + boxH
	g.y += dy
	g.onGround = false
	if dy >= 0 {
		for _, px := range [2]float32{g.x + 1, g.x + boxW - 1} {
			t := g.tileAt(px, g.y+boxH)
			top := floor((g.y+boxH)/tile) * tile
			if solid(t) || t == '-' && feet <= top+0.01 {
				g.y = top - boxH
				g.vy = 0
				g.onGround = true
				return
			}
		}
		return
	}
	for _, px := range [2]float32{g.x + 1, g.x + boxW - 1} {
		if solid(g.tileAt(px, g.y)) {
			g.y = floor(g.y/tile)*tile + tile
			g.vy = 0
			return
		}
	}
}

func (g *Game) breakDirt(px, py float32) {
	c, r := int(px/tile), int(py/tile)
	if r >= 0 && r < rows && c >= 0 && c < g.width {
		g.level[r][c] = ' '
		g.score += 5
	}
}

// updateEnemies walks the bugs and flies the moths.
func (g *Game) updateEnemies() {
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive {
			continue
		}
		if e.squashed > 0 {
			e.squashed++
			if e.squashed > 30 {
				e.alive = false
			}
			continue
		}
		// Enemies far from the view wait.
		if e.x < g.camera-40 || e.x > g.camera+view+40 {
			continue
		}
		switch e.kind {
		case moth:
			e.x += e.vx
			e.y = e.baseY + sin(float32(g.ticks+i*40)*0.05)*10
			if solid(g.tileAt(e.x+4+e.vx*8, e.y+4)) {
				e.vx = -e.vx
			}
		default:
			e.x += e.vx
			ahead := e.x + 4 + sign(e.vx)*5
			// Turn at a wall or at the edge of the ground.
			if solid(g.tileAt(ahead, e.y+4)) || !solid(g.tileAt(ahead, e.y+tile+1)) && g.tileAt(ahead, e.y+tile+1) != '-' {
				e.vx = -e.vx
			}
		}
	}
}

func sign(x float32) float32 {
	if x < 0 {
		return -1
	}
	return 1
}

// touch handles what the gopher meets: gems, spikes, checkpoints, the goal
// and the enemies.
func (g *Game) touch() {
	for _, p := range [4][2]float32{{1, 1}, {boxW - 1, 1}, {1, boxH - 1}, {boxW - 1, boxH - 1}} {
		px, py := g.x+p[0], g.y+p[1]
		c, r := int(floor(px/tile)), int(floor(py/tile))
		if r < 0 || r >= rows || c < 0 || c >= g.width {
			continue
		}
		switch g.level[r][c] {
		case 'o':
			g.level[r][c] = ' '
			g.gems++
			g.score += 10
			if g.gems%gemLife == 0 {
				g.livesLeft++
			}
		case '^':
			g.hit()
		case 'C':
			g.level[r][c] = 'c' // a reached checkpoint
			g.checkX, g.checkY = float32(c*tile)+1, float32(r*tile+tile-boxH)
		case 'G':
			g.score += 500 + 100*g.heartsLeft
			g.setMode(modeClear)
			return
		}
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive || e.squashed > 0 {
			continue
		}
		if g.x+boxW < e.x+1 || g.x > e.x+7 || g.y+boxH < e.y+2 || g.y > e.y+7 {
			continue
		}
		switch {
		case g.dash > 0:
			e.squashed = 1
			g.score += 100
		case g.vy > 0 && g.y+boxH < e.y+6 && e.kind != spiky:
			// A stomp from above.
			e.squashed = 1
			g.score += 100
			g.vy, g.jumping = -stompBounce, true
			g.dashUsed = false
		default:
			g.hit()
		}
	}
}

// hit costs a heart, unless the gopher is still blinking from the last.
func (g *Game) hit() {
	if g.hurt > 0 {
		return
	}
	g.heartsLeft--
	g.hurt = hurtTicks
	g.vy = -2
	g.vx = -g.facing * 1.5
	if g.heartsLeft <= 0 {
		g.lose()
	}
}

// lose costs a life.
func (g *Game) lose() {
	g.livesLeft--
	g.setMode(modeDead)
}

func sin(x float32) float32 { return float32(math.Sin(float64(x))) }

// demo runs the meadow behind the title with the gopher on its own: it runs
// right, jumps at walls, gaps, spikes and bugs, and dashes at dirt and at
// gaps it is falling into. A lost gopher starts over.
func (g *Game) demo() {
	ticks := g.modeTicks
	g.mode = modePlay
	g.step(g.autopilot())
	if g.mode != modePlay {
		g.livesLeft = lives
		g.startStage(0)
	}
	g.mode, g.modeTicks = modeTitle, ticks
}

func (g *Game) autopilot() input {
	in := input{right: true, jumpHeld: g.vy < 0}
	front := g.x + boxW + 6
	feet := g.y + boxH
	wall := solid(g.tileAt(front, feet-4)) || solid(g.tileAt(front, g.y))
	gap := !solid(g.tileAt(front, feet+2)) && g.tileAt(front, feet+2) != '-'
	spikes := g.tileAt(front+4, feet-2) == '^' || g.tileAt(front+4, feet+2) == '^'
	bugAhead := false
	for i := range g.enemies {
		e := &g.enemies[i]
		if e.alive && e.squashed == 0 && e.x > g.x && e.x-g.x < 20 && e.y-g.y < 12 && g.y-e.y < 12 {
			bugAhead = true
		}
	}
	switch {
	case g.tileAt(front, feet-4) == 'd' || g.tileAt(front, g.y+2) == 'd':
		in.dash = true
	case g.onGround && (wall || gap || spikes || bugAhead):
		in.jump, in.jumpHeld = true, true
	case !g.onGround && g.vy > 1 && gap && !solid(g.tileAt(g.x+boxW/2, feet+24)):
		in.dash = true
	}
	return in
}
