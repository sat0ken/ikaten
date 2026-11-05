// Package bounce is a small shared-source compatibility and stress scene.
package bounce

import "image/color"

type body struct{ x, y, vx, vy int }
type Game struct {
	width, height, count int
	sprite               *gameImage
	bodies               [30]body
	tint, paused         bool
}

var background = color.RGBA{R: 16, G: 24, B: 40, A: 255}
var spriteColor = color.RGBA{R: 72, G: 208, B: 232, A: 255}
var highlight = color.RGBA{R: 232, G: 248, B: 255, A: 255}

func New(width, height, count int) *Game {
	if width < 16 || height < 16 {
		panic("bounce: display must be at least 16x16")
	}
	g := &Game{width: width, height: height, count: min(max(count, 1), 30), sprite: newGameImage(16, 16)}
	g.sprite.Fill(spriteColor)
	for y := 3; y < 7; y++ {
		for x := 3; x < 7; x++ {
			g.sprite.Set(x, y, highlight)
		}
	}
	for n := 0; n < g.count; n++ {
		g.bodies[n] = body{x: n * 19 % (width - 15), y: n * 31 % (height - 15), vx: 1 + n%2, vy: 1 + (n/2)%2}
	}
	return g
}

// UseMaskedSprite makes the sprite's corners transparent for masked-path tests.
func (g *Game) UseMaskedSprite() {
	for _, y := range []int{0, 15} {
		for _, x := range []int{0, 15} {
			g.sprite.Set(x, y, color.RGBA{})
		}
	}
}
func (g *Game) Update() error {
	if justPressed(keyZ) {
		g.tint = !g.tint
	}
	if justPressed(keyX) {
		g.paused = !g.paused
	}
	if pressed(keyEscape) {
		return termination
	}
	if !g.paused {
		for n := 0; n < g.count; n++ {
			p := &g.bodies[n]
			p.x += p.vx
			p.y += p.vy
			if p.x < 0 || p.x > g.width-16 {
				p.vx = -p.vx
				p.x = min(max(p.x, 0), g.width-16)
			}
			if p.y < 0 || p.y > g.height-16 {
				p.vy = -p.vy
				p.y = min(max(p.y, 0), g.height-16)
			}
		}
	}
	p := &g.bodies[0]
	if pressed(keyLeft) {
		p.x -= 3
	}
	if pressed(keyRight) {
		p.x += 3
	}
	if pressed(keyUp) {
		p.y -= 3
	}
	if pressed(keyDown) {
		p.y += 3
	}
	p.x = min(max(p.x, 0), g.width-16)
	p.y = min(max(p.y, 0), g.height-16)
	return nil
}
func (g *Game) Draw(screen *gameImage) {
	screen.Fill(background)
	for n := 0; n < g.count; n++ {
		var op drawImageOptions
		op.GeoM.Translate(float64(g.bodies[n].x), float64(g.bodies[n].y))
		if g.tint {
			op.ColorScale.Scale(1, .5, .5, 1)
		}
		screen.DrawImage(g.sprite, &op)
	}
	debugPrint(screen, "ikaten  Z:color X:pause")
}
func (g *Game) Layout(width, height int) (int, int) { return g.width, g.height }
