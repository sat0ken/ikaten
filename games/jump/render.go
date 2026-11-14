package jump

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

var (
	textWhite  = color.RGBA{250, 250, 250, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGray   = color.RGBA{170, 180, 200, 255}
	textShadow = color.RGBA{20, 24, 40, 255}
	boxColor   = color.RGBA{20, 24, 40, 255}
	boxEdge    = color.RGBA{110, 130, 190, 255}
	boxLight   = color.RGBA{170, 186, 230, 255}
	poleLight  = color.RGBA{240, 240, 248, 255}
	poleShade  = color.RGBA{170, 170, 190, 255}
	knobColor  = color.RGBA{255, 210, 60, 255}
	knobShade  = color.RGBA{200, 140, 30, 255}
	flagRed    = color.RGBA{230, 70, 60, 255}
	flagGreen  = color.RGBA{60, 200, 90, 255}
	flagDark   = color.RGBA{30, 140, 60, 255}
	dashLine   = color.RGBA{255, 255, 255, 255}
)

var line pixtext.Line

// T is a tile's size on the screen.
const T = tile * zoom

func (g *Game) Draw(screen *ebiten.Image) {
	lk := &looks[g.stage]
	cam := int(g.camera * zoom)
	mark(0)
	g.drawBackground(screen, lk, cam)
	mark(1)
	g.drawTiles(screen, cam)
	mark(2)
	g.drawEnemies(screen, cam)
	mark(3)
	g.drawGopher(screen, cam) // behind the title too, where it runs on its own
	mark(4)
	g.drawHUD(screen)
	mark(5)
	g.drawOverlay(screen)
	mark(6)
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

// skyColors are the stages' skies as color.Color, made once: making one
// each frame would allocate.
var skyColors [len(looks)]color.Color

func init() {
	for i := range looks {
		skyColors[i] = looks[i].sky
	}
}

// cloudSpan is how far apart the clouds far behind come round again.
const cloudSpan = 4 * 180

// A ridge of hills behind: its height over each column comes round every
// period columns, and is between low and high. The part above low is drawn
// into an image a period wide when the stage changes, already in the stage's
// height (div) and colors, and then drawn side by side as the view scrolls;
// below low is one rectangle. In the cave a squashed copy hangs from the
// roof too.
type ridge struct {
	period, low, high int
	img, roof         *ebiten.Image
	div               int // the image's heights are divided by this
}

func (r *ridge) height(x int) int {
	if r.period == farPeriod {
		return 68 + tri(x, farPeriod)*2/3 + tri(x+40, farPeriod/2)/3
	}
	return 36 + tri(x, nearPeriod)/2 + tri(x+22, nearPeriod/2)/4
}

const farPeriod, nearPeriod = 180, 140

// backdrop holds the large images drawn for a stage: the ridges, and the
// ground strips below. They belong to a Game, made in NewGame, and are
// dropped with it: in the launcher they would otherwise take some 70 KB of
// RAM while other games run.
type backdrop struct {
	far, near  ridge
	ridgeStage int // the stage plus one whose colors the ridges are in

	// strip holds a column of ground tiles for the stage being drawn,
	// starting with each of the two kinds, and cut its tops n tiles long:
	// a run of ground is drawn at once rather than a tile at a time.
	strip      [2]*ebiten.Image
	cut        [2][rows + 1]*ebiten.Image
	stripStage int // the stage plus one
}

func newBackdrop() *backdrop {
	b := &backdrop{
		far:  ridge{period: farPeriod, low: 68, high: 68 + 42 + 21},
		near: ridge{period: nearPeriod, low: 36, high: 36 + 32 + 16},
	}
	for _, r := range [2]*ridge{&b.far, &b.near} {
		r.img = ebiten.NewImage(r.period, r.high-r.low)
		r.roof = ebiten.NewImage(r.period, (r.high-r.low)/2+1)
	}
	for k := range b.strip {
		b.strip[k] = ebiten.NewImage(T, rows*T)
		b.strip[k].Fill(color.RGBA{0, 0, 0, 255}) // opaque: see solidArt
		// SubImage allocates, so the pieces are cut once here.
		for n := 1; n <= rows; n++ {
			b.cut[k][n] = b.strip[k].SubImage(image.Rect(0, 0, T, n*T)).(*ebiten.Image)
		}
	}
	return b
}

// paint draws the ridge in c with its heights divided by div, standing on
// the bottom of its image, and the roof's ridge at half height hanging from
// the top of its own.
func (r *ridge) paint(c color.RGBA, div int, roof bool) {
	r.div = div
	r.img.Clear()
	r.roof.Clear()
	var ink color.Color = c // made once, not for every pixel
	h := r.high - r.low
	for x := range r.period {
		// The image's bottom row is just above the rectangle below low.
		for y := h - (r.height(x)/div - r.low/div); y < h; y++ {
			r.img.Set(x, y, ink)
		}
		if roof {
			for y := 0; r.low/2+y < r.height(x)/2; y++ {
				r.roof.Set(x, y, ink)
			}
		}
	}
}

// draw draws the ridge scrolled by offset, standing on the bottom of the
// screen or hanging from the top.
func (r *ridge) draw(screen *ebiten.Image, c color.RGBA, offset int, roof bool) {
	img, low := r.img, r.low/r.div
	y := ScreenHeight - low - img.Bounds().Dy()
	if roof {
		img, low, y = r.roof, r.low/2, r.low/2
	}
	for x := -(offset % r.period); x < ScreenWidth; x += r.period {
		drawAt(screen, img, x, y)
	}
	if roof {
		fill(screen, c, 0, 0, ScreenWidth, low)
	} else {
		fill(screen, c, 0, ScreenHeight-low, ScreenWidth, low)
	}
}

// drawBackground draws the sky, clouds far behind, and two ridges that
// scroll slower than the ground: hills in the meadow and the sky, rock in the
// cave.
func (g *Game) drawBackground(screen *ebiten.Image, lk *look, cam int) {
	const bands = 12
	if g.filled != screen {
		// Filling the whole screen marks it opaque, after which drawing on
		// it need not mark what it covers. The sky covers all of it in any
		// case, so it is done once.
		screen.Fill(skyColors[g.stage])
		g.filled = screen
	}
	// Below the far ridge's foot the sky is always covered.
	div := [2]int{1, 1}
	if g.stage == 2 {
		div = [2]int{2, 3} // a sea of clouds far below
	}
	b := g.back
	foot := ScreenHeight - b.far.low/div[0]
	for i := range bands {
		y := i * ScreenHeight / bands
		h := min(ScreenHeight/bands, foot-y)
		if h <= 0 {
			break
		}
		fill(screen, lerp(lk.sky, lk.horizon, float32(i)/(bands-1)), 0, y, ScreenWidth, h)
	}
	if lk.clouds {
		for i := range 4 {
			x := (i*180+97*i%60-cam/8)%cloudSpan - 40
			if x < -40 {
				x += cloudSpan
			}
			drawAt(screen, g.art.farCloud, x, 22+i*29%70)
		}
	}
	if b.ridgeStage != g.stage+1 {
		b.far.paint(lk.far, div[0], g.stage == 1)
		b.near.paint(lk.near, div[1], g.stage == 1)
		b.ridgeStage = g.stage + 1
	}
	if g.stage == 1 {
		// Rock hangs from the cave's roof as well.
		b.far.draw(screen, lk.far, cam/4+70, true)
		b.near.draw(screen, lk.near, cam/2+30, true)
	}
	b.far.draw(screen, lk.far, cam/4, false)
	b.near.draw(screen, lk.near, cam/2, false)
}

func lerp(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{mix(a.R, b.R, t), mix(a.G, b.G, t), mix(a.B, b.B, t), 255}
}

func mix(x, y uint8, t float32) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }

// ground reports whether a tile is ground, for the edges and the grass.
// Off the map counts as ground.
func (g *Game) ground(c, r int) bool {
	if c < 0 || c >= g.width || r < 0 || r >= rows {
		return true // nothing to edge off the map
	}
	return solid(g.level[r][c])
}

// drawTiles draws the map in view: ground with grass where it meets the air
// and an edge on its open sides, blocks, dirt, clouds, spikes, gems and
// flags.
func (g *Game) drawTiles(screen *ebiten.Image, cam int) {
	s := g.stage
	edge := looks[s].ground[3]
	if g.back.stripStage != s+1 {
		g.back.paintStrips(g.art, s)
	}
	c0, c1 := cam/T, min((cam+ScreenWidth)/T+1, g.width)
	for c := c0; c < c1; c++ {
		x := c*T - cam
		for r := 0; r < rows; r++ { // a run of ground moves r on
			y := r * T
			switch t := g.level[r][c]; t {
			case '#':
				run := r
				for run < rows && g.level[run][c] == '#' {
					run++
				}
				top := !g.ground(c, r-1)
				start := r
				if top {
					drawAt(screen, g.art.groundTop[s], x, y)
					start++
				}
				if start < run {
					// The rest of the run in one piece.
					drawAt(screen, g.back.cut[(c+start)%2][run-start], x, start*T)
				}
				g.drawEdges(screen, edge, c, r, run, top, x)
				r = run - 1
			case '=':
				drawAt(screen, g.art.block[s], x, y)
			case 'd':
				drawAt(screen, g.art.dirt[s], x, y)
			case '-':
				drawAt(screen, g.art.cloud, x, y)
			case '^':
				drawAt(screen, g.art.spike, x, y)
			case 'o':
				drawAt(screen, g.art.gem[(g.ticks/12+c)%2], x+2, y+2)
			case 'C', 'c':
				fill(screen, poleShade, x+3, y-12, 2, T+12)
				fill(screen, poleLight, x+3, y-12, 1, T+12)
				flag := flagRed
				if t == 'c' {
					flag = flagGreen
				}
				for i := range 8 {
					fill(screen, flag, x+5, y-12+i, 10-abs(i-4)*2, 1)
				}
			case 'G':
				fill(screen, poleShade, x+6, y, 4, T)
				fill(screen, poleLight, x+6, y, 2, T)
				if r == 0 || g.level[r-1][c] != 'G' {
					g.drawPass(screen, x, y)
				}
			}
		}
	}
}

// drawEdges draws the edge of a run of ground in column c, rows r to run,
// where its sides or its bottom meet the air: a stretch at a time.
func (g *Game) drawEdges(screen *ebiten.Image, edge color.RGBA, c, r, run int, top bool, x int) {
	for side := -1; side <= 1; side += 2 {
		from := -1
		for rr := r; rr <= run; rr++ {
			open := rr < run && !g.ground(c+side, rr)
			if open && from < 0 {
				from = rr
			}
			if !open && from >= 0 {
				y0 := from * T
				if from == r && top {
					y0 += 4 // under the grass
				}
				ex := x
				if side > 0 {
					ex = x + T - 2
				}
				fill(screen, edge, ex, y0, 2, rr*T-y0)
				from = -1
			}
		}
	}
	if !g.ground(c, run) {
		fill(screen, edge, x, run*T-2, T, 2)
	}
}

// paintStrips draws the ground strips in the stage's colors.
func (b *backdrop) paintStrips(art *sprites, stage int) {
	for k := range b.strip {
		for i := range rows {
			drawAt(b.strip[k], art.ground[stage][(k+i)%2], 0, i*T)
		}
	}
	b.stripStage = stage + 1
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// drawPass draws the PASS flag waving on top of the goal's pole.
func (g *Game) drawPass(screen *ebiten.Image, x, y int) {
	for i := range 40 {
		wave := tri(i*4+g.ticks*3, 128) / 16
		h := 18
		if i == 39 {
			h = 16
		}
		fill(screen, flagGreen, x+10+i, y-2+wave, 1, h)
		fill(screen, flagDark, x+10+i, y+15+wave, 1, 1)
	}
	line.Reset().Str("PASS").Shade(flagDark).Draw(screen, x+19, y+3+tri(g.ticks*3+40, 128)/16, 1, textWhite)
	fill(screen, knobShade, x+5, y-6, 6, 6)
	fill(screen, knobColor, x+5, y-6, 4, 4)
}

func (g *Game) drawEnemies(screen *ebiten.Image, cam int) {
	frame := g.ticks / 10 % 2
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive {
			continue
		}
		x, y := int(e.x*zoom)-cam, int(e.y*zoom)
		if x < -T || x > ScreenWidth {
			continue
		}
		var img *ebiten.Image
		switch {
		case e.squashed > 0:
			img = g.art.squash
		case e.kind == moth:
			img = g.art.moth[frame]
		case e.kind == spiky:
			img = g.art.spiky[frame]
		default:
			img = g.art.bug[frame]
		}
		var op ebiten.DrawImageOptions
		if e.vx < 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(T, 0)
		}
		op.GeoM.Translate(float64(x), float64(y+T-img.Bounds().Dy()))
		screen.DrawImage(img, &op)
	}
}

// The gopher's sprite and where it sits on its box.
const gopherW, gopherH, gopherLeft = 18, 22, (gopherW - boxW*zoom) / 2

func (g *Game) drawGopher(screen *ebiten.Image, cam int) {
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
	x := int(g.x*zoom) - cam - gopherLeft
	y := int((g.y+boxH)*zoom) - gopherH
	if g.mode == modeDead {
		// Up and away.
		y -= g.modeTicks * 6 / 5
	}
	if g.dash > 0 {
		// Speed lines behind.
		for i := range 3 {
			l := 6 + (g.dash+i*3)%5*2
			lx := x - l - 2
			if g.facing < 0 {
				lx = x + gopherW + 2
			}
			fill(screen, dashLine, lx, y+7+i*5, l, 1)
		}
	}
	var op ebiten.DrawImageOptions
	if g.facing < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(gopherW, 0)
	}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(g.art.gopher[frame], &op)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	if g.mode == modeTitle {
		return
	}
	for i := range hearts {
		img := g.art.heart[0]
		if i >= g.heartsLeft {
			img = g.art.heart[1]
		}
		drawAt(screen, img, 4+i*16, 4)
	}
	drawAt(screen, g.art.gem[0], 56, 4)
	line.Reset().Str("x").Num(g.gems).Shade(textShadow).Draw(screen, 70, 7, 1, textWhite)
	line.Reset().Str("GOPHER x").Num(g.livesLeft).Shade(textShadow).Draw(screen, 104, 7, 1, textWhite)
	line.Reset().Num(g.score).Shade(textShadow).Right(screen, ScreenWidth-4, 7, 1, textWhite)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	switch g.mode {
	case modeTitle:
		box(screen, 16, 28, 208, 184)
		line.Reset().Str("GOPHER JUMP").Shade(boxEdge).Center(screen, mid, 38, 3, textYellow)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(3, 3)
		op.GeoM.Translate(mid-27, 64)
		screen.DrawImage(g.art.gopher[3*(g.modeTicks/30%2)], &op)
		line.Reset().Str("A: JUMP   B: DASH").Center(screen, mid, 140, 1, textWhite)
		line.Reset().Str("DASH BREAKS DIRT").Center(screen, mid, 154, 1, textGray)
		if g.modeTicks/30%2 == 0 {
			line.Reset().Str("PRESS A").Center(screen, mid, 172, 2, textYellow)
		}
		line.Reset().Str("GOPHER BY RENEE FRENCH").Center(screen, mid, 198, 1, textGray)
	case modeIntro:
		box(screen, 48, 88, 144, 56)
		line.Reset().Str("STAGE ").Num(g.stage+1).Center(screen, mid, 96, 1, textGray)
		line.Reset().Str(stages[g.stage].name).Center(screen, mid, 114, 2, textWhite)
	case modeDead:
		if g.modeTicks > 30 {
			box(screen, 60, 100, 120, 32)
			line.Reset().Str("GOPHER x").Num(g.livesLeft).Center(screen, mid, 112, 1, textWhite)
		}
	case modeClear:
		box(screen, 48, 88, 144, 56)
		line.Reset().Str("PASS!").Center(screen, mid, 98, 2, flagGreen)
		line.Reset().Str("SCORE ").Num(g.score).Center(screen, mid, 124, 1, textWhite)
	case modeOver:
		box(screen, 40, 88, 160, 64)
		line.Reset().Str("GAME OVER").Center(screen, mid, 100, 2, textRed)
		line.Reset().Str("SCORE ").Num(g.score).Center(screen, mid, 132, 1, textWhite)
	case modeEnd:
		box(screen, 28, 60, 184, 120)
		line.Reset().Str("ALL PASS!").Center(screen, mid, 72, 2, flagGreen)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(mid-18, 96)
		screen.DrawImage(g.art.gopher[3*(g.modeTicks/20%2)], &op)
		line.Reset().Str("SCORE ").Num(g.score).Center(screen, mid, 154, 1, textYellow)
	}
}

// box is a dark panel with a two-tone frame, drawn side by side rather than
// one over another.
func box(screen *ebiten.Image, x, y, w, h int) {
	fill(screen, boxLight, x-2, y-2, w+3, 1)
	fill(screen, boxEdge, x+w+1, y-2, 1, 1)
	fill(screen, boxEdge, x-2, y-1, w+4, 1)
	fill(screen, boxEdge, x-2, y, 2, h)
	fill(screen, boxEdge, x+w, y, 2, h)
	fill(screen, boxEdge, x-2, y+h, w+4, 2)
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
