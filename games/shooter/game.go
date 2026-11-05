// Package shooter is an original vertical-scrolling shooter. It uses only the
// Ebitengine API, and every image is drawn by code, so it also builds with
// Ebitengine. Positions are integers in 1/16 pixels and angles come from a
// table, because float arithmetic is software on RP2040.
package shooter

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	fp           = 16 // positions are in 1/16 pixels

	maxEnemies = 16
	maxShots   = 24
	maxBullets = 48
	maxBlasts  = 16
	maxItems   = 4
	maxStars   = 24

	playerSpeed = 24 // 1.5 pixels per tick
	shotSpeed   = 64
	fireTicks   = 6
	startLives  = 3
	startBombs  = 2
	maxBombs    = 5
	maxPower    = 3
	blastTicks  = 4 // ticks per blast frame

	// Difficulty.
	respawnTicks    = 180 // invincibility after a respawn
	hitRadius       = 2   // pixels from the cockpit that a bullet must reach
	bossFireTicks   = 60
	bulletBaseSpeed = 14 // plus bulletLevelSpeed per stage, in 1/16 pixels per tick
	bulletLevelStep = 2
)

type kind uint8

const (
	scout kind = iota
	weaver
	gunner
	boss
)

// bodies holds each kind's half size in pixels for collisions, its hit points
// on the first stage and its points.
var bodies = [...]struct{ w, h, hp, points int }{
	scout:  {4, 3, 1, 100},
	weaver: {4, 4, 2, 150},
	gunner: {5, 4, 8, 300},
	boss:   {15, 6, 80, 5000},
}

type vec struct{ x, y int }

// dirs holds unit vectors scaled by 256 for 64 angles; index 16 points down.
var dirs [64]vec

const down = 16

type enemy struct {
	kind           kind
	x, y, vy       int
	baseX          int
	hp, age, flash int
	alive          bool
}

type mover struct {
	x, y, vx, vy int
	alive        bool
}

// A blast with a negative age waits before it appears.
type blastState struct {
	x, y, age int
	alive     bool
}

type star struct{ x, y, layer int }

var starSpeeds = [3]int{4, 8, 16}

// A wave spawns count enemies, one every interval ticks, starting at x pixels
// and moving dx pixels for each spawn.
type wave struct {
	kind            kind
	count, interval int
	x, dx           int
}

var waves = [...]wave{
	{scout, 6, 12, 15, 18},
	{weaver, 6, 20, 40, 0},
	{scout, 8, 8, 10, 14},
	{gunner, 2, 70, 35, 50},
	{weaver, 6, 20, 80, 0},
	{scout, 8, 10, 110, -13},
	{gunner, 3, 50, 25, 35},
	{weaver, 8, 14, 60, 0},
}

type mode int

const (
	modeTitle mode = iota
	modePlay
	modeOver
)

type Game struct {
	mode                          mode
	ticks, modeTicks              int
	rng                           uint32
	px, py                        int
	lives, bombs, power           int
	invincible, respawn, fireCool int
	score, hiScore, kills         int
	level, wave, waveTick         int
	bossSpawned                   bool
	bombFlash                     int
	enemies                       [maxEnemies]enemy
	shots                         [maxShots]mover
	bullets                       [maxBullets]mover
	items                         [maxItems]mover
	blasts                        [maxBlasts]blastState
	stars                         [maxStars]star
}

func NewGame() *Game {
	g := &Game{rng: 2463534242}
	for i := range g.stars {
		g.stars[i] = star{x: g.random(ScreenWidth), y: g.random(ScreenHeight) * fp, layer: i % 3}
	}
	g.reset()
	return g
}

// reset prepares a new game. The high score and the stars stay.
func (g *Game) reset() {
	g.px, g.py = ScreenWidth/2*fp, 104*fp
	g.lives, g.bombs, g.power = startLives, startBombs, 1
	g.invincible, g.respawn, g.fireCool = 0, 0, 0
	g.score, g.kills = 0, 0
	g.level, g.wave, g.waveTick, g.bossSpawned = 1, 0, -90, false
	g.bombFlash = 0
	clear(g.enemies[:])
	clear(g.shots[:])
	clear(g.bullets[:])
	clear(g.items[:])
	clear(g.blasts[:])
}

// random returns 0 to n-1 from a xorshift generator, the same on every platform.
func (g *Game) random(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}

func (g *Game) setMode(m mode) {
	g.mode, g.modeTicks = m, 0
}

func startPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	g.modeTicks++
	for i := range g.stars {
		s := &g.stars[i]
		if s.y += starSpeeds[s.layer]; s.y >= ScreenHeight*fp {
			s.y -= ScreenHeight * fp
			s.x = g.random(ScreenWidth)
		}
	}
	switch g.mode {
	case modeTitle:
		if startPressed() {
			g.setMode(modePlay)
		}
	case modePlay:
		g.updatePlayer()
		g.updateWaves()
		g.updateEnemies()
		g.updateMovers()
		g.collide()
	case modeOver:
		g.updateEnemies()
		g.updateMovers()
		if g.modeTicks > 60 && startPressed() {
			g.reset()
			g.setMode(modeTitle)
		}
	}
	for i := range g.blasts {
		b := &g.blasts[i]
		if b.alive {
			b.age++
			b.alive = b.age < blastFrames*blastTicks
		}
	}
	if g.bombFlash > 0 {
		g.bombFlash--
	}
	return nil
}

func (g *Game) updatePlayer() {
	if g.respawn > 0 {
		if g.respawn--; g.respawn == 0 {
			if g.lives == 0 {
				g.setMode(modeOver)
				return
			}
			g.px, g.py = ScreenWidth/2*fp, 104*fp
			g.invincible = respawnTicks
		}
		return
	}
	if g.invincible > 0 {
		g.invincible--
	}
	dx, dy := 0, 0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy++
	}
	g.px = min(max(g.px+dx*playerSpeed, 6*fp), (ScreenWidth-6)*fp)
	g.py = min(max(g.py+dy*playerSpeed, 14*fp), (ScreenHeight-7)*fp)
	if g.fireCool > 0 {
		g.fireCool--
	}
	if g.fireCool == 0 && (ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeySpace)) {
		g.fire()
		g.fireCool = fireTicks
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) && g.bombs > 0 {
		g.bomb()
	}
}

func (g *Game) fire() {
	switch g.power {
	case 1:
		g.addShot(0, 0)
	case 2:
		g.addShot(-3*fp, 0)
		g.addShot(3*fp, 0)
	default:
		g.addShot(-3*fp, 0)
		g.addShot(3*fp, 0)
		g.addShot(-4*fp, -16)
		g.addShot(4*fp, 16)
	}
}

func (g *Game) addShot(offset, vx int) {
	for i := range g.shots {
		if s := &g.shots[i]; !s.alive {
			*s = mover{x: g.px + offset, y: g.py - 5*fp, vx: vx, vy: -shotSpeed, alive: true}
			return
		}
	}
}

// bomb clears every enemy bullet and damages every enemy on screen.
func (g *Game) bomb() {
	g.bombs--
	g.bombFlash = 16
	clear(g.bullets[:])
	for i := range g.enemies {
		if e := &g.enemies[i]; e.alive && e.y > 0 {
			g.damage(e, 12)
		}
	}
	g.invincible = max(g.invincible, 30)
}

func maxHP(k kind, level int) int {
	hp := bodies[k].hp
	return hp + (level-1)*hp/2
}

func (g *Game) damage(e *enemy, n int) {
	e.hp -= n
	e.flash = 4
	if e.hp > 0 {
		return
	}
	e.alive = false
	g.score += bodies[e.kind].points * g.level
	g.hiScore = max(g.hiScore, g.score)
	g.kills++
	g.addBlast(e.x, e.y, 0)
	if e.kind == gunner || g.kills%15 == 0 {
		g.addItem(e.x, e.y)
	}
	if e.kind == boss {
		for i := range 6 {
			g.addBlast(e.x+(g.random(31)-15)*fp, e.y+(g.random(13)-6)*fp, -(i+1)*5)
		}
		clear(g.bullets[:])
		g.level++
		g.wave, g.waveTick, g.bossSpawned = 0, -150, false
		g.bombs = min(g.bombs+1, maxBombs)
	}
}

func (g *Game) addBlast(x, y, age int) {
	for i := range g.blasts {
		if b := &g.blasts[i]; !b.alive {
			*b = blastState{x: x, y: y, age: age, alive: true}
			return
		}
	}
}

func (g *Game) addItem(x, y int) {
	for i := range g.items {
		if it := &g.items[i]; !it.alive {
			*it = mover{x: x, y: y, vy: 8, alive: true}
			return
		}
	}
}

func (g *Game) updateWaves() {
	if g.waveTick < 0 {
		g.waveTick++
		return
	}
	if g.wave < len(waves) {
		w := waves[g.wave]
		if n := g.waveTick / w.interval; g.waveTick%w.interval == 0 && n < w.count {
			g.spawn(w.kind, w.x+n*w.dx)
		}
		if g.waveTick++; g.waveTick >= w.count*w.interval+60 {
			g.wave++
			g.waveTick = 0
		}
		return
	}
	if !g.bossSpawned && g.enemyCount() == 0 {
		g.spawn(boss, ScreenWidth/2)
		g.bossSpawned = true
	}
}

func (g *Game) enemyCount() int {
	n := 0
	for i := range g.enemies {
		if g.enemies[i].alive {
			n++
		}
	}
	return n
}

func (g *Game) spawn(k kind, x int) {
	if k == weaver {
		x = min(max(x, 26), ScreenWidth-26) // keep the sway on screen
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		if e.alive {
			continue
		}
		*e = enemy{kind: k, x: x * fp, baseX: x * fp, y: -8 * fp, hp: maxHP(k, g.level), alive: true}
		switch k {
		case scout:
			e.vy = 16 + 3*g.level
		case weaver:
			e.vy = 10 + 2*g.level
		case gunner:
			e.vy = 14
		case boss:
			e.y, e.vy = -16*fp, 8
		}
		return
	}
}

func (g *Game) bulletSpeed() int { return bulletBaseSpeed + bulletLevelStep*g.level }

func (g *Game) updateEnemies() {
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive {
			continue
		}
		e.age++
		if e.flash > 0 {
			e.flash--
		}
		switch e.kind {
		case scout:
			e.y += e.vy
			if g.level >= 2 && e.age == 30 && g.random(3) == 0 {
				g.fan(e.x, e.y, g.aim(e.x, e.y), 1, 0, g.bulletSpeed())
			}
		case weaver:
			e.y += e.vy
			e.x = e.baseX + dirs[e.age%len(dirs)].y*20*fp/256 // a sway of about a second
			if g.level >= 2 && e.age%90 == 45 {
				g.fan(e.x, e.y, g.aim(e.x, e.y), 1, 0, g.bulletSpeed())
			}
		case gunner:
			if e.age >= 200 {
				// Back off upward rather than diving through the player.
				if e.y -= 20; e.y < -10*fp {
					e.alive = false
				}
			} else if e.y < 30*fp {
				e.y += e.vy
			}
			if e.age == 70 || e.age == 110 || e.age == 150 {
				g.fan(e.x, e.y+4*fp, g.aim(e.x, e.y), 2*min(g.level, 3)-1, 3, g.bulletSpeed())
			}
		case boss:
			g.updateBoss(e)
		}
		if e.y > (ScreenHeight+12)*fp {
			e.alive = false
		}
	}
}

func (g *Game) updateBoss(e *enemy) {
	if e.y < 26*fp {
		e.y += e.vy
		e.age = 0
		return
	}
	e.x = ScreenWidth/2*fp + dirs[e.age/2%len(dirs)].y*34*fp/256
	if e.age%bossFireTicks != 0 {
		return
	}
	speed := g.bulletSpeed()
	switch e.age / bossFireTicks % 3 {
	case 0:
		g.fan(e.x, e.y+6*fp, down, 5, 3, speed)
	case 1:
		g.fan(e.x, e.y+6*fp, g.aim(e.x, e.y), 3, 2, speed+6)
	default:
		if g.level >= 2 {
			g.fan(e.x, e.y, 0, 16, 4, speed) // a full ring
		} else {
			g.fan(e.x, e.y+6*fp, down, 7, 3, speed)
		}
	}
}

// aim returns the table angle closest to the direction of the player.
func (g *Game) aim(x, y int) int {
	dx, dy := g.px-x, g.py-y
	best, bestDot := 0, dirs[0].x*dx+dirs[0].y*dy
	for i, d := range dirs {
		if dot := d.x*dx + d.y*dy; dot > bestDot {
			best, bestDot = i, dot
		}
	}
	return best
}

// fan fires count bullets centered on angle, step table angles apart.
func (g *Game) fan(x, y, angle, count, step, speed int) {
	for i := range count {
		d := dirs[(angle+(i-count/2)*step+len(dirs))%len(dirs)]
		for j := range g.bullets {
			if b := &g.bullets[j]; !b.alive {
				*b = mover{x: x, y: y, vx: d.x * speed / 256, vy: d.y * speed / 256, alive: true}
				break
			}
		}
	}
}

func move(ms []mover) {
	for i := range ms {
		m := &ms[i]
		if !m.alive {
			continue
		}
		m.x += m.vx
		m.y += m.vy
		if m.x < -4*fp || m.x > (ScreenWidth+4)*fp || m.y < -4*fp || m.y > (ScreenHeight+4)*fp {
			m.alive = false
		}
	}
}

func (g *Game) updateMovers() {
	move(g.shots[:])
	move(g.bullets[:])
	move(g.items[:])
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (g *Game) collide() {
	for i := range g.shots {
		s := &g.shots[i]
		if !s.alive {
			continue
		}
		for j := range g.enemies {
			e := &g.enemies[j]
			if !e.alive || e.y < -4*fp {
				continue
			}
			if b := bodies[e.kind]; abs(s.x-e.x) < (b.w+1)*fp && abs(s.y-e.y) < (b.h+2)*fp {
				s.alive = false
				g.damage(e, 1)
				break
			}
		}
	}
	if g.respawn > 0 {
		return
	}
	for i := range g.items {
		if it := &g.items[i]; it.alive && abs(it.x-g.px) < 7*fp && abs(it.y-g.py) < 7*fp {
			it.alive = false
			if g.power < maxPower {
				g.power++
			} else {
				g.score += 1000
				g.hiScore = max(g.hiScore, g.score)
			}
		}
	}
	if g.invincible > 0 {
		return
	}
	// The player's hitbox is the cockpit, much smaller than the ship.
	for i := range g.bullets {
		if b := &g.bullets[i]; b.alive && abs(b.x-g.px) < hitRadius*fp && abs(b.y-g.py) < hitRadius*fp {
			g.die()
			return
		}
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		// Touching counts only when the cockpit enters the enemy's body.
		if b := bodies[e.kind]; e.alive && abs(e.x-g.px) < b.w*fp && abs(e.y-g.py) < b.h*fp {
			g.die()
			return
		}
	}
}

func (g *Game) die() {
	g.addBlast(g.px, g.py, 0)
	g.addBlast(g.px-4*fp, g.py+2*fp, -6)
	g.addBlast(g.px+4*fp, g.py-2*fp, -10)
	g.lives--
	g.power = max(1, g.power-1)
	g.bombs = max(g.bombs, startBombs)
	g.respawn = 90
}

var colorSpace = color.RGBA{6, 6, 20, 255}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.bombFlash > 0 {
		v := uint8(6 + g.bombFlash*10)
		screen.Fill(color.RGBA{v, v, v + 14, 255})
	} else {
		screen.Fill(colorSpace)
	}
	for _, s := range g.stars {
		drawAt(screen, imageStar[s.layer], s.x, s.y/fp)
	}
	for _, it := range g.items {
		if it.alive {
			drawCentered(screen, imageItem, it.x, it.y)
		}
	}
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		if e.flash > 0 {
			drawFlashing(screen, imageEnemy[e.kind], e.x, e.y)
		} else {
			drawCentered(screen, imageEnemy[e.kind], e.x, e.y)
		}
	}
	for _, s := range g.shots {
		if s.alive {
			drawCentered(screen, imageShot, s.x, s.y)
		}
	}
	if g.mode == modePlay && g.respawn == 0 && (g.invincible == 0 || g.ticks/3%2 == 0) {
		drawCentered(screen, imagePlayer[g.ticks/4%2], g.px, g.py)
	}
	for _, b := range g.bullets {
		if b.alive {
			drawCentered(screen, imageBullet, b.x, b.y)
		}
	}
	for _, b := range g.blasts {
		if b.alive && b.age >= 0 {
			drawCentered(screen, imageBlast[b.age/blastTicks], b.x, b.y)
		}
	}
	g.drawHUD(screen)
	switch g.mode {
	case modeTitle:
		ebitenutil.DebugPrintAt(screen, "IKATEN SHOOTER", 18, 26)
		ebitenutil.DebugPrintAt(screen, "HI", 34, 44)
		drawNumber(screen, g.hiScore, 6, 50, 49)
		ebitenutil.DebugPrintAt(screen, "Z:SHOT  X:BOMB", 18, 66)
		if g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 86)
		}
	case modeOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER", 33, 40)
		if g.modeTicks > 60 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 66)
		}
	case modePlay:
		if g.waveTick < 0 && g.wave == 0 {
			ebitenutil.DebugPrintAt(screen, "STAGE", 42, 44)
			drawNumber(screen, g.level, 2, 76, 49)
		}
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	if g.mode == modeTitle {
		return
	}
	drawNumber(screen, g.score, 6, 2, 2)
	for i := range max(g.lives-1, 0) {
		drawAt(screen, imageLife, ScreenWidth-7-i*6, 2)
	}
	for i := range g.bombs {
		drawAt(screen, imageBomb, 2+i*5, ScreenHeight-6)
	}
	for _, e := range g.enemies {
		if e.alive && e.kind == boss && e.y > 0 {
			const x, width = 20, 80
			fillRect(screen, imageBarBG, x, 9, width, 2)
			fillRect(screen, imageBar, x, 9, max(e.hp, 0)*width/maxHP(boss, g.level), 2)
		}
	}
}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

// drawCentered draws img centered on a position in 1/16 pixels.
func drawCentered(screen, img *ebiten.Image, x, y int) {
	b := img.Bounds()
	drawAt(screen, img, x>>4-b.Dx()/2, y>>4-b.Dy()/2)
}

// drawFlashing brightens a sprite that was just hit.
func drawFlashing(screen, img *ebiten.Image, x, y int) {
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x>>4-b.Dx()/2), float64(y>>4-b.Dy()/2))
	op.ColorScale.Scale(2.5, 2.5, 2.5, 1)
	screen.DrawImage(img, &op)
}

// fillRect draws a 1x1 image scaled to a rectangle.
func fillRect(screen, pixel *ebiten.Image, x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(width), float64(height))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(pixel, &op)
}

// drawNumber draws n with the given number of digits, padded with zeros.
func drawNumber(screen *ebiten.Image, n, digits, x, y int) {
	for i := digits - 1; i >= 0; i-- {
		drawAt(screen, imageDigits[n%10], x+i*4, y)
		n /= 10
	}
}
