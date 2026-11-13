package fish

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

// The view is a cross-section of the sea from the side. It scrolls down to
// follow the jig.
const (
	viewTop = 9 // rows above hold the clock and the score
	pxPerM  = 2 // vertical
	xScale  = 1.75
	xOrigin = 14
	skyM    = 14 // meters of sky above the surface at the top of the view
	deepest = 56
)

func sx(x float32) float32 { return xOrigin + x*xScale }

func (g *Game) camTop() float32 {
	if g.mode != modeFishing && g.mode != modeFight {
		return -skyM
	}
	return min(max(g.lureDepth-24, -skyM), deepest-(ScreenHeight-viewTop)/pxPerM)
}

func (g *Game) sy(depth float32) float32 { return viewTop + (depth-g.camTop())*pxPerM }

// keyframes of the light through the day: sky at the top, sky at the
// horizon and the sea at the surface.
var dayLight = [...]struct {
	hour                  float32
	top, horizon, surface color.RGBA
}{
	{5.5, color.RGBA{40, 40, 90, 255}, color.RGBA{230, 130, 90, 255}, color.RGBA{60, 70, 110, 255}},
	{7, color.RGBA{90, 140, 210, 255}, color.RGBA{240, 200, 160, 255}, color.RGBA{50, 100, 150, 255}},
	{12, color.RGBA{80, 150, 230, 255}, color.RGBA{180, 215, 245, 255}, color.RGBA{40, 110, 170, 255}},
	{16, color.RGBA{90, 140, 210, 255}, color.RGBA{240, 210, 160, 255}, color.RGBA{45, 100, 155, 255}},
	{18.5, color.RGBA{60, 50, 110, 255}, color.RGBA{240, 120, 70, 255}, color.RGBA{50, 60, 110, 255}},
}

var deepWater = color.RGBA{8, 20, 40, 255}

func lerp(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{mix(a.R, b.R, t), mix(a.G, b.G, t), mix(a.B, b.B, t), 255}
}

func mix(x, y uint8, t float32) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }

// light returns the colors of the hour.
func (g *Game) light() (top, horizon, surface color.RGBA) {
	h := g.hour()
	k := dayLight[:]
	if h <= k[0].hour {
		return k[0].top, k[0].horizon, k[0].surface
	}
	for i := 1; i < len(k); i++ {
		if h <= k[i].hour {
			t := (h - k[i-1].hour) / (k[i].hour - k[i-1].hour)
			return lerp(k[i-1].top, k[i].top, t), lerp(k[i-1].horizon, k[i].horizon, t), lerp(k[i-1].surface, k[i].surface, t)
		}
	}
	l := k[len(k)-1]
	return l.top, l.horizon, l.surface
}

func (g *Game) Draw(screen *ebiten.Image) {
	top, horizon, surface := g.light()
	g.drawSky(screen, top, horizon)
	g.drawSea(screen, surface)
	g.drawBottom(screen)
	g.drawFish(screen)
	g.drawBoat(screen)
	g.drawLine(screen)
	g.drawScale(screen)
	g.drawBar(screen)
	g.drawOverlay(screen)
}

func (g *Game) drawSky(screen *ebiten.Image, top, horizon color.RGBA) {
	surfaceY := int(g.sy(0))
	if surfaceY <= viewTop {
		return
	}
	const bands = 6
	h := surfaceY - viewTop
	for i := range bands {
		y0, y1 := viewTop+h*i/bands, viewTop+h*(i+1)/bands
		fill(screen, lerp(top, horizon, float32(i)/(bands-1)), 0, y0, ScreenWidth, y1-y0)
	}
	// The sun crosses the sky from dawn to dusk.
	t := (g.hour() - 5.5) / 13
	if t > 0 && t < 1 {
		x := 34 + t*76
		lift := 0.2 + 3.2*t*(1-t) // low at the ends, highest at noon
		y := g.sy(-1 - lift*11)
		c := lerp(color.RGBA{255, 150, 80, 255}, color.RGBA{255, 245, 210, 255}, min((lift-0.2)*1.8, 1))
		drawTinted(screen, imageSun, int(x)-4, int(y)-4, c)
	}
}

func (g *Game) drawSea(screen *ebiten.Image, surface color.RGBA) {
	cam := g.camTop()
	// Bands of five meters, darker with depth.
	for d := float32(0); d < deepest+10; d += 5 {
		y0, y1 := int(g.sy(d)), int(g.sy(d+5))
		if y1 <= viewTop || y0 >= ScreenHeight {
			continue
		}
		y0 = max(y0, viewTop)
		fill(screen, lerp(surface, deepWater, min(d/50, 1)), 0, y0, ScreenWidth, y1-y0)
	}
	// Waves on the surface.
	if cam < 0 {
		y := g.sy(0)
		wave := lerp(surface, color.RGBA{230, 240, 250, 255}, 0.5)
		for x := 0; x < ScreenWidth; x += 2 {
			dy := sinf(float32(x)*0.35+float32(g.ticks)*0.06) * 1.2
			fill(screen, wave, x, int(y+dy)-1, 2, 1)
		}
	}
}

func hash(n int) uint32 {
	h := uint32(n) * 2654435761
	h ^= h >> 15
	h *= 2246822519
	h ^= h >> 13
	return h
}

// bedColumns holds the seabed's depth under each column of the screen,
// worked out at the first frame: sines are slow on boards without a
// floating-point unit.
var (
	bedColumns [ScreenWidth]float32
	bedReady   bool
)

func (g *Game) drawBottom(screen *ebiten.Image) {
	if !bedReady {
		for px := range bedColumns {
			bedColumns[px] = seabed((float32(px) - xOrigin) / xScale)
		}
		bedReady = true
	}
	sand, dark := color.RGBA{150, 130, 90, 255}, color.RGBA{100, 86, 60, 255}
	for px := 0; px < ScreenWidth; px++ {
		y := int(g.sy(bedColumns[px]))
		if y >= ScreenHeight {
			continue
		}
		y = max(y, viewTop)
		fill(screen, dark, px, y, 1, 1)
		fill(screen, sand, px, y+1, 1, ScreenHeight-y-1)
	}
	// Weed and rocks at fixed places.
	for i := range 14 {
		h := hash(i)
		x := float32(4 + h%58)
		y := g.sy(seabed(x))
		img := imageWeed
		if h>>8%3 == 0 {
			img = imageRock
		}
		b := img.Bounds()
		drawAt(screen, img, int(sx(x))-b.Dx()/2, int(y)-b.Dy()+1)
	}
}

func (g *Game) drawFish(screen *ebiten.Image) {
	for i := range g.fish {
		f := &g.fish[i]
		if !f.alive {
			continue
		}
		t := &speciesTypes[f.kind]
		img := imageFish[f.kind]
		// A little bigger for a big one of its kind.
		scale := 0.8 + 0.4*(f.length-t.minLen)/max(t.maxLen-t.minLen, 1)
		x, y := sx(f.x), g.sy(f.depth)
		if y < viewTop-8 || y > ScreenHeight+8 {
			continue
		}
		b := img.Bounds()
		w, h := float32(b.Dx())*scale, float32(b.Dy())*scale
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(float64(scale*f.facing), float64(scale))
		if f.facing < 0 {
			op.GeoM.Translate(float64(w), 0)
		}
		// Thrash while hooked.
		wobble := float32(0)
		if f.state == hooked && g.running {
			wobble = sinf(float32(g.ticks)*0.8) * 1.5
		}
		op.GeoM.Translate(float64(x-w/2), float64(y-h/2+wobble))
		screen.DrawImage(img, &op)
	}
}

// rodTip is where the tip of the rod is drawn: it dips for a nibble and bends
// toward a fish on the line.
func (g *Game) rodTip() (float32, float32) {
	x, y := sx(tipX), g.sy(tipDepth)
	switch {
	case g.mode == modeFight:
		bend := 2 + g.tension*5
		x += bend * 0.4
		y += bend
	case g.twitch > 0 && g.suitor >= 0 && g.fish[g.suitor].state == bite:
		y += 4
	case g.twitch > 0:
		y += float32(g.twitch % 4)
	}
	return x, y
}

func (g *Game) drawBoat(screen *ebiten.Image) {
	y := g.sy(0)
	if y < viewTop-20 || y > ScreenHeight+10 {
		return
	}
	bob := sinf(float32(g.ticks)*0.04) * 0.8
	bx := sx(-6.5)
	drawAt(screen, imageAngler, int(sx(-1.6)), int(y+bob)-9)
	drawAt(screen, imageBoat, int(bx), int(y+bob)-2)
	// The rod, from the angler's hands to the tip.
	hx, hy := sx(-1.6)+6, y+bob-4
	tx, ty := g.rodTip()
	line(screen, color.RGBA{70, 50, 30, 255}, hx, hy, tx, ty)
}

func (g *Game) drawLine(screen *ebiten.Image) {
	if g.mode != modeFlight && g.mode != modeFishing && g.mode != modeFight {
		return
	}
	tx, ty := g.rodTip()
	lx, ly := sx(g.lureX), g.sy(g.jigDepth())
	line(screen, color.RGBA{220, 225, 230, 255}, tx, ty, lx, ly)
	if g.mode != modeFight {
		drawAt(screen, imageLure, int(lx)-1, int(ly)-1)
	}
	if g.mode == modeFishing && (g.freeFall || g.jigOff > 0.1) && g.ticks%8 < 5 {
		drawAt(screen, imageBubble, int(lx)-1, int(ly)-5-g.ticks%8)
	}
	if g.mode == modeFishing && g.suitor >= 0 {
		switch g.fish[g.suitor].state {
		case bite:
			buf.reset().str("!").draw2(screen, int(lx)-2, int(ly)-16, textYellow)
		case nibble:
			if g.twitch > 0 {
				buf.reset().str(".").draw(screen, int(lx)+3, int(ly)-6, textWhite)
			}
		}
	}
}

// drawScale marks every ten meters of depth at the right edge.
func (g *Game) drawScale(screen *ebiten.Image) {
	for d := 10; d <= deepest; d += 10 {
		y := int(g.sy(float32(d)))
		if y < viewTop+3 || y > ScreenHeight-3 {
			continue
		}
		fill(screen, color.RGBA{200, 220, 240, 255}, ScreenWidth-3, y, 3, 1)
		buf.reset().num(d).draw(screen, -(ScreenWidth - 4), y-2, textPale)
	}
}

var (
	textWhite  = color.RGBA{240, 244, 248, 255}
	textPale   = color.RGBA{170, 200, 225, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGreen  = color.RGBA{110, 230, 130, 255}
	textGray   = color.RGBA{150, 160, 170, 255}
	barColor   = color.RGBA{12, 20, 34, 255}
	boxColor   = color.RGBA{12, 18, 30, 255}
	boxEdge    = color.RGBA{80, 110, 150, 255}
)

// drawBar draws the clock and the score at the top, and the gauges.
func (g *Game) drawBar(screen *ebiten.Image) {
	fill(screen, barColor, 0, 0, ScreenWidth, viewTop)
	h, m := g.clock/3600, g.clock/60%60
	buf.reset().two(h).str(":").two(m).draw(screen, 2, 2, textWhite)
	buf.reset().num(g.score).str(" PTS").draw(screen, -(ScreenWidth - 2), 2, textYellow)
	switch g.mode {
	case modeCast:
		// The power gauge in the sky over the water.
		x, y := 30, viewTop+4
		fill(screen, boxColor, x-1, y-1, 52, 6)
		n := int(g.power * 50)
		c := lerp(color.RGBA{90, 200, 110, 255}, color.RGBA{255, 90, 60, 255}, g.power)
		fill(screen, c, x, y, n, 4)
		buf.reset().num(int(minCast+(maxCast-minCast)*g.power)).str("M").draw(screen, x+55, y-1, textWhite)
	case modeFishing:
		buf.reset().str("DEPTH ").num(int(g.jigDepth())).str("M").draw(screen, 38, 2, textPale)
	case modeFight:
		g.drawTension(screen)
	}
}

// drawTension shows the line's tension at the bottom: green, then yellow,
// red near breaking. The line out is beside it.
func (g *Game) drawTension(screen *ebiten.Image) {
	const x, y, w = 4, ScreenHeight - 8, 80
	fill(screen, boxColor, x-1, y-1, w+2, 6)
	n := int(min(g.tension, 1.1) / 1.1 * w)
	c := color.RGBA{90, 210, 110, 255}
	switch {
	case g.tension > 0.85:
		c = color.RGBA{255, 70, 50, 255}
	case g.tension > 0.6:
		c = color.RGBA{255, 210, 60, 255}
	case g.tension < 0.15:
		c = color.RGBA{120, 140, 160, 255}
	}
	if g.over == 0 || g.ticks%6 < 3 {
		fill(screen, c, x, y, n, 4)
	}
	// The breaking point.
	fill(screen, color.RGBA{255, 255, 255, 255}, x+w*10/11, y-2, 1, 8)
	buf.reset().num(int(g.line)).str("M").draw(screen, x+w+4, y-1, textWhite)
	if g.running {
		buf.reset().str("RUN").draw(screen, 58, 2, textRed)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	switch g.mode {
	case modeTitle:
		box(screen, 12, 22, 96, 66)
		ebitenutil.DebugPrintAt(screen, "SEA FISHING", 27, 20)
		buf.reset().str("HOLD A: CAST").center(screen, 38, textWhite)
		buf.reset().str("V: DROP  ^: REEL").center(screen, 46, textWhite)
		buf.reset().str("A: JIG / STRIKE").center(screen, 54, textWhite)
		buf.reset().str("FIGHT: HOLD A TO REEL").center(screen, 62, textYellow)
		buf.reset().str("B: WIND IN").center(screen, 70, textWhite)
		if g.modeTicks/30%2 == 0 {
			buf.reset().str("PRESS A").center(screen, 80, textGreen)
		}
		return
	case modeCatch:
		g.drawCatch(screen)
		return
	case modeResult:
		g.drawResult(screen)
		return
	}
	if g.messageTime > 0 && g.message != msgNone && g.messageTime%30 < 22 {
		msg := messages[g.message]
		w := len(msg)*6 + 6
		box(screen, (ScreenWidth-w)/2-1, viewTop+4, w, 10)
		ebitenutil.DebugPrintAt(screen, msg, (ScreenWidth-len(msg)*6)/2+2, viewTop+1)
	}
}

func (g *Game) drawCatch(screen *ebiten.Image) {
	c := &g.lastLand
	box(screen, 12, 22, 96, 66)
	name := speciesTypes[c.kind].name
	ebitenutil.DebugPrintAt(screen, name, (ScreenWidth-len(name)*6)/2, 21)
	img := imageFish[c.kind]
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(3, 3)
	op.GeoM.Translate(float64(ScreenWidth/2-b.Dx()*3/2), float64(52-b.Dy()*3/2))
	screen.DrawImage(img, &op)
	if c.kind == boot {
		buf.reset().str("NOT A FISH").center(screen, 70, textGray)
	} else {
		buf.reset().num(int(c.length+0.5)).str("CM  ").dec(c.weight).str("KG").center(screen, 66, textWhite)
		buf.reset().str("+").num(c.points).str(" PTS").center(screen, 74, textYellow)
		if g.newBest && g.catches[c.kind] > 1 {
			buf.reset().str("BIGGEST TODAY").center(screen, 82, textGreen)
		}
	}
}

func (g *Game) drawResult(screen *ebiten.Image) {
	box(screen, 8, 12, 104, 100)
	ebitenutil.DebugPrintAt(screen, "SUNSET", 42, 10)
	y := 26
	for k := range kinds {
		if g.catches[k] == 0 {
			continue
		}
		buf.reset().str(speciesTypes[k].name).draw(screen, 14, y, textWhite)
		buf.reset().str("X").num(g.catches[k]).draw(screen, 60, y, textWhite)
		if k != boot {
			buf.reset().num(int(g.best[k]+0.5)).str("CM").draw(screen, -106, y, textPale)
		}
		y += 7
	}
	if y == 26 {
		buf.reset().str("NO FISH TODAY").center(screen, y, textGray)
		y += 7
	}
	buf.reset().dec(g.total).str(" KG").center(screen, 88, textWhite)
	buf.reset().num(g.score).str(" PTS").center(screen, 96, textYellow)
	if g.modeTicks > 60 && g.modeTicks/30%2 == 0 {
		buf.reset().str("PRESS A").center(screen, 104, textGreen)
	}
}

func box(screen *ebiten.Image, x, y, w, h int) {
	fill(screen, boxEdge, x-1, y-1, w+2, h+2)
	fill(screen, boxColor, x, y, w, h)
}

// line plots a one-pixel line.
func line(screen *ebiten.Image, c color.RGBA, x0, y0, x1, y1 float32) {
	dx, dy := x1-x0, y1-y0
	n := int(max(abs(dx), abs(dy))) + 1
	for i := 0; i <= n; i++ {
		t := float32(i) / float32(n)
		fill(screen, c, int(x0+dx*t), int(y0+dy*t), 1, 1)
	}
}

func abs(x float32) float32 { return max(x, -x) }

// paints holds one-pixel images of the colors in use. A tinted white pixel
// would be shaded pixel by pixel; a pixel of the right color is copied.
// The colors of the sky and the sea change with the hour, so the slots are
// repainted as colors come and go.
var paints [64]struct {
	img  *ebiten.Image
	c    color.RGBA
	used int
}

func init() {
	for i := range paints {
		paints[i].img = ebiten.NewImage(1, 1)
	}
}

var paintTick int

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

// fill draws a rectangle of a color.
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

func drawTinted(screen, img *ebiten.Image, x, y int, c color.RGBA) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.Scale(float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, 1)
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
	if n < 0 {
		t.str("-")
		n = -n
	}
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

// dec writes x with two decimals.
func (t *text) dec(x float32) *text {
	c := int(x*100 + 0.5)
	t.num(c / 100).str(".")
	return t.two(c % 100)
}

// draw draws the line with its left edge at x, or its right edge at -x.
func (t *text) draw(screen *ebiten.Image, x, y int, c color.RGBA) {
	if x < 0 {
		x = -x - t.n*4 + 1
	}
	drawText(screen, t.b[:t.n], x, y, 1, c)
}

func (t *text) draw2(screen *ebiten.Image, x, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], x, y, 2, c)
}

func (t *text) center(screen *ebiten.Image, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], (ScreenWidth-t.n*4)/2, y, 1, c)
}

// drawText draws text in the 3x5 font at the given scale and color.
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
