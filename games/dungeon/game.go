// Package dungeon is an original first-person shooter in the style of early
// raycasting games, drawn for a 240 by 240 screen. It uses only the
// Ebitengine API, and every image is drawn by code, so it also builds with
// Ebitengine.
package dungeon

import (
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	viewHeight   = 208 // the rows below hold the status bar
	texSize      = 32
	planeLength  = 0.66 // camera plane half-width; about a 66 degree view
	moveSpeed    = 0.06
	turnSpeed    = 0.05
	playerRadius = 0.2
	slimeSpeed   = 0.02
	slimeReach   = 0.7
	maxSlimes    = 8
)

// '#' brick, '=' stone, '%' metal, 'P' player start, 'S' slime.
var worldMap = [...]string{
	"################",
	"#P.....#.......#",
	"#......#..S....#",
	"#..==..#.......#",
	"#..==.....%%...#",
	"#.........%%...#",
	"#####..#.......#",
	"#......#####.###",
	"#..S...#.......#",
	"#......#..S....#",
	"#..%...........#",
	"#..%...#####...#",
	"#......#...#...#",
	"#.S....#.S.#.S.#",
	"#..........#...#",
	"################",
}

const mapSize = len(worldMap)

func wallAt(x, y int) byte {
	if x < 0 || y < 0 || x >= mapSize || y >= mapSize {
		return '#'
	}
	switch c := worldMap[y][x]; c {
	case '#', '=', '%':
		return c
	}
	return 0
}

type mode int

const (
	modeTitle mode = iota
	modePlay
	modeDead
	modeClear
)

type slime struct {
	x, y     float32
	hp       int
	hurt     int // ticks of the red flash after a hit
	cooldown int // ticks until the next attack
	alive    bool
}

type Game struct {
	mode           mode
	px, py         float32 // player position
	dirX, dirY     float32
	planeX, planeY float32
	hp, kills      int
	flash          int // ticks of the muzzle flash
	damage         int // ticks of the red screen after being hit
	ticks          int
	slimes         [maxSlimes]slime
	slimeCount     int
	zbuf           [ScreenWidth]float32
	order          [maxSlimes]int
	art            *sprites
}

func NewGame() *Game {
	g := &Game{art: newSprites()}
	g.reset()
	return g
}

func (g *Game) reset() {
	*g = Game{mode: g.mode, art: g.art}
	g.hp = 100
	g.dirX, g.dirY = 1, 0
	g.planeX, g.planeY = 0, planeLength
	for y, row := range worldMap {
		for x := range len(row) {
			switch row[x] {
			case 'P':
				g.px, g.py = float32(x)+0.5, float32(y)+0.5
			case 'S':
				if g.slimeCount < maxSlimes {
					g.slimes[g.slimeCount] = slime{x: float32(x) + 0.5, y: float32(y) + 0.5, hp: 3, alive: true}
					g.slimeCount++
				}
			}
		}
	}
}

func firePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	switch g.mode {
	case modeTitle:
		if firePressed() {
			g.mode = modePlay
		}
	case modePlay:
		g.updatePlayer()
		g.updateSlimes()
		if g.hp <= 0 {
			g.mode = modeDead
		} else if g.kills == g.slimeCount {
			g.mode = modeClear
		}
	case modeDead, modeClear:
		if firePressed() {
			g.mode = modeTitle
			g.reset()
		}
	}
	if g.flash > 0 {
		g.flash--
	}
	if g.damage > 0 {
		g.damage--
	}
	return nil
}

func free(x, y float32) bool {
	return wallAt(int(x-playerRadius), int(y-playerRadius)) == 0 &&
		wallAt(int(x+playerRadius), int(y-playerRadius)) == 0 &&
		wallAt(int(x-playerRadius), int(y+playerRadius)) == 0 &&
		wallAt(int(x+playerRadius), int(y+playerRadius)) == 0
}

var turnCos, turnSin = float32(math.Cos(turnSpeed)), float32(math.Sin(turnSpeed))

func (g *Game) updatePlayer() {
	turn := func(s float32) {
		g.dirX, g.dirY = g.dirX*turnCos-g.dirY*s, g.dirX*s+g.dirY*turnCos
		g.planeX, g.planeY = g.planeX*turnCos-g.planeY*s, g.planeX*s+g.planeY*turnCos
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		turn(-turnSin)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		turn(turnSin)
	}
	step := float32(0)
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		step += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		step -= moveSpeed
	}
	// Move each axis separately so the player slides along walls.
	if nx := g.px + g.dirX*step; free(nx, g.py) {
		g.px = nx
	}
	if ny := g.py + g.dirY*step; free(g.px, ny) {
		g.py = ny
	}
	if firePressed() {
		g.flash = 5
		g.shoot()
	}
}

// project returns a slime's horizontal screen position and its depth.
func (g *Game) project(s *slime) (screenX, depth float32) {
	sx, sy := s.x-g.px, s.y-g.py
	invDet := 1 / (g.planeX*g.dirY - g.dirX*g.planeY)
	tx := invDet * (g.dirY*sx - g.dirX*sy)
	depth = invDet * (-g.planeY*sx + g.planeX*sy)
	return ScreenWidth / 2 * (1 + tx/depth), depth
}

func spriteSize(depth float32) float32 { return viewHeight / depth * 0.7 }

// shoot hits the nearest slime under the crosshair that is in front of the wall.
func (g *Game) shoot() {
	best, bestDepth := -1, g.zbuf[ScreenWidth/2]
	for i := range g.slimeCount {
		s := &g.slimes[i]
		if !s.alive {
			continue
		}
		x, depth := g.project(s)
		if depth < 0.2 || depth >= bestDepth {
			continue
		}
		if half := spriteSize(depth) / 2; x-half <= ScreenWidth/2 && ScreenWidth/2 <= x+half {
			best, bestDepth = i, depth
		}
	}
	if best < 0 {
		return
	}
	s := &g.slimes[best]
	s.hp--
	s.hurt = 8
	if s.hp <= 0 {
		s.alive = false
		g.kills++
	}
}

func (g *Game) updateSlimes() {
	for i := range g.slimeCount {
		s := &g.slimes[i]
		if s.hurt > 0 {
			s.hurt--
		}
		if !s.alive {
			continue
		}
		if s.cooldown > 0 {
			s.cooldown--
		}
		dx, dy := g.px-s.x, g.py-s.y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist < slimeReach {
			if s.cooldown == 0 {
				g.hp -= 10
				g.damage = 10
				s.cooldown = 45
			}
			continue
		}
		if dist > 8 {
			continue
		}
		nx, ny := s.x+dx/dist*slimeSpeed, s.y+dy/dist*slimeSpeed
		if wallAt(int(nx), int(s.y)) == 0 {
			s.x = nx
		}
		if wallAt(int(s.x), int(ny)) == 0 {
			s.y = ny
		}
	}
}
