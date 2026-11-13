package bomb

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

// The arena fills the screen between the score bar and the item bar.
const (
	arenaX = 1
	arenaY = 11
	barH   = 10
)

var (
	grassA     = color.RGBA{70, 150, 80, 255}
	grassB     = color.RGBA{62, 138, 72, 255}
	shadow     = color.RGBA{48, 110, 58, 255}
	barColor   = color.RGBA{20, 24, 40, 255}
	boxColor   = color.RGBA{16, 20, 36, 255}
	boxEdge    = color.RGBA{90, 110, 170, 255}
	fireOuter  = color.RGBA{240, 90, 30, 255}
	fireInner  = color.RGBA{255, 220, 80, 255}
	fireCore   = color.RGBA{255, 250, 220, 255}
	textWhite  = color.RGBA{240, 244, 248, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGray   = color.RGBA{150, 160, 180, 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawArena(screen)
	g.drawFire(screen)
	g.drawBombs(screen)
	g.drawGophers(screen)
	g.drawBars(screen)
	g.drawOverlay(screen)
}

func cellX(c int) int { return arenaX + c*tile }
func cellY(r int) int { return arenaY + r*tile }

func (g *Game) drawArena(screen *ebiten.Image) {
	fill(screen, barColor, 0, arenaY, ScreenWidth, rows*tile)
	for r := range rows {
		for c := range cols {
			x, y := cellX(c), cellY(r)
			switch g.grid[r][c] {
			case wall:
				drawAt(screen, imageWall, x, y)
				continue
			case block:
				drawAt(screen, imageBlock, x, y)
				continue
			}
			grass := grassA
			if (r+c)%2 == 1 {
				grass = grassB
			}
			fill(screen, grass, x, y, tile, tile)
			// The shade of the wall or block above.
			if r > 0 && g.grid[r-1][c] != floor {
				fill(screen, shadow, x, y, tile, 2)
			}
			if g.crumble[r][c] > 0 {
				drawAt(screen, imageCrumble, x, y)
			} else if g.shown[r][c] {
				drawAt(screen, imageItem[g.items[r][c]], x, y)
			}
		}
	}
}

// drawFire draws each burning cell as bands along the lines the fire runs.
func (g *Game) drawFire(screen *ebiten.Image) {
	flick := g.ticks / 3 % 2
	for r := range rows {
		for c := range cols {
			if g.flames[r][c] == 0 {
				continue
			}
			x, y := cellX(c), cellY(r)
			across := c > 0 && g.flames[r][c-1] > 0 || c < cols-1 && g.flames[r][c+1] > 0
			along := r > 0 && g.flames[r-1][c] > 0 || r < rows-1 && g.flames[r+1][c] > 0
			if !across && !along {
				across, along = true, true
			}
			w := 7 - flick
			if g.flames[r][c] < 8 {
				w = 3 // dying down
			}
			if across {
				fill(screen, fireOuter, x, y+(tile-w)/2, tile, w)
				fill(screen, fireInner, x, y+(tile-w)/2+1, tile, w-2)
			}
			if along {
				fill(screen, fireOuter, x+(tile-w)/2, y, w, tile)
				fill(screen, fireInner, x+(tile-w)/2+1, y, w-2, tile)
			}
			if across && along {
				fill(screen, fireCore, x+3, y+3, 3, 3)
			}
		}
	}
}

func (g *Game) drawBombs(screen *ebiten.Image) {
	for i := range g.bombs {
		b := &g.bombs[i]
		if !b.active {
			continue
		}
		x := arenaX + b.x*tile/unit - tile/2
		y := arenaY + b.y*tile/unit - tile/2
		// It swells as the fuse burns down.
		frame := g.ticks / 8 % 2
		var op ebiten.DrawImageOptions
		if b.fuse < 50 && g.ticks/4%2 == 0 {
			op.GeoM.Translate(-4.5, -4.5)
			op.GeoM.Scale(1.15, 1.15)
			op.GeoM.Translate(4.5, 4.5)
		}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(imageBomb[frame], &op)
	}
}

// drawGophers draws the gophers from the back row forward.
func (g *Game) drawGophers(screen *ebiten.Image) {
	var order [players]int
	for i := range order {
		order[i] = i
	}
	for i := 1; i < players; i++ {
		for j := i; j > 0 && g.gophers[order[j]].y < g.gophers[order[j-1]].y; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	for _, i := range order {
		p := &g.gophers[i]
		if !p.alive {
			continue
		}
		x := arenaX + p.x*tile/unit - tile/2
		y := arenaY + p.y*tile/unit + tile/2 - 11
		if p.dying > 0 {
			// Blink and sink away.
			if p.dying%8 < 5 {
				var op ebiten.DrawImageOptions
				s := 1 - float64(p.dying)/dieTicks*0.7
				op.GeoM.Translate(-4.5, -11)
				op.GeoM.Scale(s, s)
				op.GeoM.Translate(float64(x)+4.5, float64(y)+11)
				screen.DrawImage(imageFaint[i], &op)
			}
			continue
		}
		view, flip := p.facing, false
		switch p.facing {
		case right:
			view, flip = left, false
		case left:
			view, flip = left, true
		}
		frame := 0
		if p.moving {
			frame = g.ticks / 8 % 2
		}
		img := imageGopher[i][min(int(view), 2)][frame]
		var op ebiten.DrawImageOptions
		if flip {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(9, 0)
		}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(img, &op)
	}
}

// drawBars shows the wins and the clock above the arena and the person's
// powers below.
func (g *Game) drawBars(screen *ebiten.Image) {
	fill(screen, barColor, 0, 0, ScreenWidth, arenaY)
	for i := range g.gophers {
		x := 2 + i*24
		drawAt(screen, imageHead[i], x, 1)
		for w := range g.gophers[i].wins {
			fill(screen, textYellow, x+11+w*4, 4, 3, 3)
		}
	}
	secs := max(g.clock, 0) / tps
	c := textWhite
	if g.clock <= hurryTime && g.mode == modePlay {
		c = textRed
	}
	buf.reset().num(secs/60).str(":").two(secs%60).draw(screen, -(ScreenWidth - 2), 3, c)

	y := arenaY + rows*tile
	fill(screen, barColor, 0, y, ScreenWidth, ScreenHeight-y)
	p := &g.gophers[0]
	drawAt(screen, imageItem[fireUp], 2, y+1)
	buf.reset().num(p.fire).draw(screen, 13, y+3, textWhite)
	drawAt(screen, imageItem[bombUp], 22, y+1)
	buf.reset().num(p.count).draw(screen, 33, y+3, textWhite)
	drawAt(screen, imageItem[speedUp], 42, y+1)
	buf.reset().num(p.speed-baseSpeed+1).draw(screen, 53, y+3, textWhite)
	buf.reset().str("LEVEL ").num(g.level).draw(screen, -(ScreenWidth - 2), y+3, textGray)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	switch g.mode {
	case modeTitle:
		g.drawTitle(screen)
	case modeReady:
		box(screen, 30, 46, 60, 22)
		buf.reset().str("ROUND ").num(g.round).center(screen, 50, textWhite)
		if g.modeTicks > 45 {
			buf.reset().str("GO!").center2(screen, 58, textYellow)
		} else {
			buf.reset().str("READY").center(screen, 60, textYellow)
		}
	case modePlay:
		if g.clock <= hurryTime && g.clock > hurryTime-2*tps && g.ticks/10%2 == 0 {
			buf.reset().str("HURRY!").center2(screen, 50, textRed)
		}
	case modeRoundOver:
		box(screen, 22, 44, 76, 26)
		if g.winner < 0 {
			buf.reset().str("DRAW").center2(screen, 52, textWhite)
			break
		}
		drawScaled(screen, imageGopher[g.winner][down][g.modeTicks/10%2], 28, 46, 2)
		if g.winner == 0 {
			buf.reset().str("YOU WIN!").draw(screen, 50, 54, textYellow)
		} else {
			buf.reset().str(gopherNames[g.winner]).str(" WINS").draw(screen, 50, 54, textWhite)
		}
	case modeMatchOver:
		box(screen, 14, 30, 92, 56)
		drawScaled(screen, imageGopher[g.champion][down][g.modeTicks/15%2], 46, 34, 3)
		if g.champion == 0 {
			buf.reset().str("CHAMPION!").center2(screen, 70, textYellow)
			if g.modeTicks > 60 {
				buf.reset().str("A: LEVEL ").num(g.level+1).center(screen, 80, textWhite)
			}
		} else {
			buf.reset().str("GAME OVER").center2(screen, 70, textRed)
			if g.modeTicks > 60 {
				buf.reset().str("REACHED LEVEL ").num(g.level).center(screen, 80, textWhite)
			}
		}
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	box(screen, 10, 18, 100, 84)
	ebitenutil.DebugPrintAt(screen, "GOPHER BOMB", 27, 16)
	drawScaled(screen, imageGopher[0][down][g.modeTicks/20%2], 47, 32, 3)
	buf.reset().str("A: BOMB  B: KICK").center(screen, 70, textWhite)
	buf.reset().str("LAST GOPHER STANDING WINS").center(screen, 78, textGray)
	if g.modeTicks/30%2 == 0 {
		buf.reset().str("PRESS A").center(screen, 88, textYellow)
	}
	buf.reset().str("GOPHER BY RENEE FRENCH").center(screen, 96, textGray)
}

func box(screen *ebiten.Image, x, y, w, h int) {
	fill(screen, boxEdge, x-1, y-1, w+2, h+2)
	fill(screen, boxColor, x, y, w, h)
}

// paints holds one-pixel images of the colors in use: a pixel of the right
// color is copied, where a tinted white pixel would be shaded pixel by pixel.
var paints [32]struct {
	img  *ebiten.Image
	c    color.RGBA
	used int
}

var paintTick int

func init() {
	for i := range paints {
		paints[i].img = ebiten.NewImage(1, 1)
	}
}

// paint returns a one-pixel image of c.
func paint(c color.RGBA) *ebiten.Image {
	paintTick++
	h := (int(c.R)*31 + int(c.G)*7 + int(c.B)) % len(paints)
	oldest := h
	for i := range 8 {
		p := &paints[(h+i)%len(paints)]
		if p.used > 0 && p.c == c {
			p.used = paintTick
			return p.img
		}
		if p.used < paints[oldest].used {
			oldest = (h + i) % len(paints)
		}
	}
	p := &paints[oldest]
	p.c, p.used = c, paintTick
	p.img.Fill(c) // Fill marks the image opaque, which draws faster than Set
	return p.img
}

func fill(screen *ebiten.Image, c color.RGBA, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(paint(c), &op)
}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

func drawScaled(screen, img *ebiten.Image, x, y int, s float64) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

// text builds a line of the 3x5 font in a fixed buffer, so that drawing
// numbers makes no strings and does not allocate.
type text struct {
	b [32]byte
	n int
}

var buf text

func (t *text) reset() *text { t.n = 0; return t }

func (t *text) str(s string) *text {
	for i := range len(s) {
		if t.n < len(t.b) {
			t.b[t.n] = s[i]
			t.n++
		}
	}
	return t
}

func (t *text) num(n int) *text {
	var d [10]byte
	i := len(d)
	for {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
		if n == 0 {
			break
		}
	}
	for ; i < len(d) && t.n < len(t.b); i++ {
		t.b[t.n] = d[i]
		t.n++
	}
	return t
}

func (t *text) two(n int) *text {
	if n < 10 {
		t.str("0")
	}
	return t.num(n)
}

// draw draws the line with its left edge at x, or its right edge at -x.
func (t *text) draw(screen *ebiten.Image, x, y int, c color.RGBA) {
	if x < 0 {
		x = -x - t.n*4 + 1
	}
	drawText(screen, t.b[:t.n], x, y, 1, c)
}

func (t *text) center(screen *ebiten.Image, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], (ScreenWidth-t.n*4)/2, y, 1, c)
}

func (t *text) center2(screen *ebiten.Image, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], (ScreenWidth-t.n*8)/2, y, 2, c)
}

func drawText(screen *ebiten.Image, text []byte, x, y, scale int, c color.RGBA) {
	var op ebiten.DrawImageOptions
	for i, ch := range text {
		img := imageGlyph[ch]
		if img == nil {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x+i*4*scale), float64(y))
		op.ColorScale.Reset()
		op.ColorScale.Scale(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, 1)
		screen.DrawImage(img, &op)
	}
}
