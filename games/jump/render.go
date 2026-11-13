package jump

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

var (
	textWhite  = color.RGBA{250, 250, 250, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGray   = color.RGBA{170, 180, 200, 255}
	textDark   = color.RGBA{30, 34, 50, 255}
	boxColor   = color.RGBA{20, 24, 40, 255}
	boxEdge    = color.RGBA{110, 130, 190, 255}
	poleColor  = color.RGBA{220, 220, 230, 255}
	flagRed    = color.RGBA{230, 70, 60, 255}
	flagGreen  = color.RGBA{60, 200, 90, 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	lk := &looks[g.stage]
	g.drawBackground(screen, lk)
	g.drawTiles(screen, lk)
	g.drawEnemies(screen)
	if g.mode != modeTitle {
		g.drawGopher(screen)
	}
	g.drawHUD(screen)
	g.drawOverlay(screen)
}

// tri is a triangle wave of period p, from 0 to 64, in integers.
func tri(x, p int) int {
	x %= p
	if x < 0 {
		x += p
	}
	if x < p/2 {
		return x * 128 / p
	}
	return (p - x) * 128 / p
}

// drawBackground draws the sky and two layers behind that scroll slower
// than the ground: hills in the meadow and the sky, rock in the cave.
func (g *Game) drawBackground(screen *ebiten.Image, lk *look) {
	const bands = 6
	for i := range bands {
		fill(screen, lerp(lk.sky, lk.horizon, float32(i)/(bands-1)), 0, i*ScreenHeight/bands, ScreenWidth, ScreenHeight/bands+1)
	}
	cam := int(g.camera)
	for x := range ScreenWidth {
		g.far[x] = 34 + tri(x+cam/4, 90)/3 + tri(x+cam/4+20, 37)/6
		g.near[x] = 18 + tri(x+cam/2, 70)/4 + tri(x+cam/2+11, 23)/8
		if g.stage == 2 {
			g.far[x], g.near[x] = g.far[x]/2, g.near[x]/3 // a sea of clouds far below
		}
	}
	if g.stage == 1 {
		// Rock hangs from the cave's roof as well.
		skyline(screen, lk.far, &g.far, 2, true)
		skyline(screen, lk.near, &g.near, 2, true)
	}
	skyline(screen, lk.far, &g.far, 1, false)
	skyline(screen, lk.near, &g.near, 1, false)
}

// skyline fills the screen's columns up from the bottom, or down from the
// top, to the given heights divided by div. It fills a row at a time in spans
// and the part all columns cover at once: a column at a time would set up
// each row of each column apart, which costs far more.
func skyline(screen *ebiten.Image, c color.RGBA, h *[ScreenWidth]int, div int, top bool) {
	low, high := h[0]/div, h[0]/div
	for _, v := range h {
		low, high = min(low, v/div), max(high, v/div)
	}
	for d := low; d < high; d++ {
		y := ScreenHeight - 1 - d
		if top {
			y = d
		}
		x0 := -1
		for x := 0; x <= ScreenWidth; x++ {
			in := x < ScreenWidth && h[x]/div > d
			switch {
			case in && x0 < 0:
				x0 = x
			case !in && x0 >= 0:
				fill(screen, c, x0, y, x-x0, 1)
				x0 = -1
			}
		}
	}
	if top {
		fill(screen, c, 0, 0, ScreenWidth, low)
	} else {
		fill(screen, c, 0, ScreenHeight-low, ScreenWidth, low)
	}
}

func lerp(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{mix(a.R, b.R, t), mix(a.G, b.G, t), mix(a.B, b.B, t), 255}
}

func mix(x, y uint8, t float32) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }

// drawTiles draws the map in view. Ground is drawn a column of tiles at a
// time, with grass where it meets the air.
func (g *Game) drawTiles(screen *ebiten.Image, lk *look) {
	cam := int(g.camera)
	c0, c1 := cam/tile, min((cam+ScreenWidth)/tile+1, g.width)
	for c := c0; c < c1; c++ {
		x := c*tile - cam
		for r := 0; r < rows; r++ {
			y := r * tile
			switch t := g.level[r][c]; t {
			case '#':
				run := r
				for run < rows && g.level[run][c] == '#' {
					run++
				}
				fill(screen, lk.ground, x, y, tile, (run-r)*tile)
				if r == 0 || !solid(g.level[r-1][c]) {
					fill(screen, lk.top, x, y, tile, 2)
					fill(screen, lk.edge, x, y+2, tile, 1)
				}
				r = run - 1
			case '=':
				fill(screen, lk.blockEdge, x, y, tile, tile)
				fill(screen, lk.block, x+1, y+1, tile-2, tile-2)
			case 'd':
				fill(screen, lk.dirt, x, y, tile, tile)
				fill(screen, lk.dirtSpot, x+1, y+2, 2, 1)
				fill(screen, lk.dirtSpot, x+5, y+5, 2, 1)
				fill(screen, lk.dirtSpot, x+3, y+6, 1, 1)
			case '-':
				drawAt(screen, imageCloud, x, y)
			case '^':
				drawAt(screen, imageSpike, x, y)
			case 'o':
				drawAt(screen, imageGem[(g.ticks/12+c)%2], x+1, y+1)
			case 'C', 'c':
				fill(screen, poleColor, x+1, y-6, 1, tile+6)
				flag := flagRed
				if t == 'c' {
					flag = flagGreen
				}
				fill(screen, flag, x+2, y-6, 5, 4)
			case 'G':
				fill(screen, poleColor, x+3, y, 2, tile)
				if r == 0 || g.level[r-1][c] != 'G' {
					// The PASS flag on top of the pole.
					wave := g.ticks / 10 % 2
					fill(screen, flagGreen, x+5, y-1+wave, 19, 8)
					buf.reset().str("PASS").draw(screen, x+7, y+1+wave, textWhite)
					fill(screen, textYellow, x+3, y-2, 2, 2)
				}
			}
		}
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	cam := int(g.camera)
	frame := g.ticks / 10 % 2
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive {
			continue
		}
		x, y := int(e.x)-cam, int(e.y)
		if x < -8 || x > ScreenWidth {
			continue
		}
		var img *ebiten.Image
		switch {
		case e.squashed > 0:
			img = imageSquash
		case e.kind == moth:
			img = imageMoth[frame]
		case e.kind == spiky:
			img = imageSpiky[frame]
		default:
			img = imageBug[frame]
		}
		var op ebiten.DrawImageOptions
		if e.vx > 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(8, 0)
		}
		op.GeoM.Translate(float64(x), float64(y+tile-img.Bounds().Dy()))
		screen.DrawImage(img, &op)
	}
}

func (g *Game) drawGopher(screen *ebiten.Image) {
	if g.mode == modePlay && g.hurt > 0 && g.hurt%6 < 3 {
		return
	}
	frame := 0
	switch {
	case g.dash > 0:
		frame = 4
	case !g.onGround:
		frame = 3
	case g.vx > 0.1 || g.vx < -0.1:
		frame = 1 + g.ticks/6%2
	}
	x := g.x - g.camera - 1.5
	y := g.y + boxH - 11
	if g.mode == modeDead {
		// Up and away.
		y -= float32(g.modeTicks) * 0.6
	}
	var op ebiten.DrawImageOptions
	if g.facing < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(9, 0)
	}
	op.GeoM.Translate(float64(int(x)), float64(int(y)))
	screen.DrawImage(imageGopher[frame], &op)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	if g.mode == modeTitle {
		return
	}
	for i := range hearts {
		img := imageHeart[0]
		if i >= g.heartsLeft {
			img = imageHeart[1]
		}
		drawAt(screen, img, 2+i*8, 2)
	}
	drawAt(screen, imageGem[0], 30, 2)
	buf.reset().str("X").num(g.gems).draw(screen, 37, 3, textWhite)
	buf.reset().num(g.score).draw(screen, -(ScreenWidth - 2), 3, textWhite)
	buf.reset().str("GOPHER X").num(g.livesLeft).draw(screen, 62, 3, textWhite)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	switch g.mode {
	case modeTitle:
		box(screen, 12, 16, 96, 88)
		ebitenutil.DebugPrintAt(screen, "GOPHER JUMP", 27, 14)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(3, 3)
		op.GeoM.Translate(46, 30)
		screen.DrawImage(imageGopher[3*(g.modeTicks/30%2)], &op)
		buf.reset().str("A: JUMP  B: DASH").center(screen, 68, textWhite)
		buf.reset().str("DASH BREAKS DIRT").center(screen, 76, textGray)
		if g.modeTicks/30%2 == 0 {
			buf.reset().str("PRESS A").center(screen, 86, textYellow)
		}
		buf.reset().str("GOPHER BY RENEE FRENCH").center(screen, 96, textGray)
	case modeIntro:
		box(screen, 24, 44, 72, 28)
		buf.reset().str("STAGE ").num(g.stage+1).center(screen, 50, textGray)
		buf.reset().str(stages[g.stage].name).center2(screen, 58, textWhite)
	case modeDead:
		if g.modeTicks > 30 {
			box(screen, 30, 50, 60, 16)
			buf.reset().str("GOPHER X").num(g.livesLeft).center(screen, 56, textWhite)
		}
	case modeClear:
		box(screen, 24, 44, 72, 28)
		buf.reset().str("PASS!").center2(screen, 50, flagGreen)
		buf.reset().str("SCORE ").num(g.score).center(screen, 64, textWhite)
	case modeOver:
		box(screen, 20, 44, 80, 32)
		buf.reset().str("GAME OVER").center2(screen, 50, textRed)
		buf.reset().str("SCORE ").num(g.score).center(screen, 66, textWhite)
	case modeEnd:
		box(screen, 14, 30, 92, 60)
		buf.reset().str("ALL PASS!").center2(screen, 36, flagGreen)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(51, 50)
		screen.DrawImage(imageGopher[3*(g.modeTicks/20%2)], &op)
		buf.reset().str("SCORE ").num(g.score).center(screen, 76, textYellow)
	}
}

func box(screen *ebiten.Image, x, y, w, h int) {
	fill(screen, boxEdge, x-1, y-1, w+2, h+2)
	fill(screen, boxColor, x, y, w, h)
}

// paints holds one-pixel images of the colors in use: a pixel of the right
// color is copied, where a tinted white pixel would be shaded pixel by pixel.
var paints [48]struct {
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
