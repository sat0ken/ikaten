package fish

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

// The view is a cross-section of the sea from the side. It scrolls down to
// follow the jig.
const (
	viewTop = 18 // rows above hold the clock and the score
	pxPerM  = 4  // vertical
	xScale  = 3.5
	xOrigin = 28
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
	const bands = 8
	h := surfaceY - viewTop
	for i := range bands {
		y0, y1 := viewTop+h*i/bands, viewTop+h*(i+1)/bands
		fill(screen, lerp(top, horizon, float32(i)/(bands-1)), 0, y0, ScreenWidth, y1-y0)
	}
	// The sun crosses the sky from dawn to dusk.
	t := (g.hour() - 5.5) / 13
	if t > 0 && t < 1 {
		x := 68 + t*152
		lift := 0.2 + 3.2*t*(1-t) // low at the ends, highest at noon
		y := g.sy(-1 - lift*11)
		c := lerp(color.RGBA{255, 150, 80, 255}, color.RGBA{255, 245, 210, 255}, min((lift-0.2)*1.8, 1))
		drawTinted(screen, imageSun, int(x)-9, int(y)-9, c)
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
	// Waves on the surface, with a light crest.
	if cam < 0 {
		y := g.sy(0)
		wave := lerp(surface, color.RGBA{230, 240, 250, 255}, 0.45)
		crest := lerp(surface, color.RGBA{250, 252, 255, 255}, 0.8)
		for x := 0; x < ScreenWidth; x += 8 {
			dy := sinf(float32(x)*0.17+float32(g.ticks)*0.06) * 2.4
			fill(screen, wave, x, int(y+dy)-1, 8, 2)
			if x%16 == 0 {
				fill(screen, crest, x+2, int(y+dy)-2, 4, 1)
			}
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
// floating-point unit. bedHeights is how high the bed stands from the
// bottom of the screen in each column this frame.
var (
	bedColumns [ScreenWidth]float32
	bedHeights [ScreenWidth]int
	bedReady   bool
)

var (
	sand     = color.RGBA{150, 130, 90, 255}
	sandDark = color.RGBA{100, 86, 60, 255}
)

func (g *Game) drawBottom(screen *ebiten.Image) {
	if !bedReady {
		for px := range bedColumns {
			bedColumns[px] = seabed((float32(px) - xOrigin) / xScale)
		}
		bedReady = true
	}
	for px := range bedHeights {
		y := max(int(g.sy(bedColumns[px])), viewTop)
		bedHeights[px] = max(ScreenHeight-y, 0)
	}
	// A dark edge two pixels deep, then the sand.
	skyline(screen, sandDark, &bedHeights, 0)
	skyline(screen, sand, &bedHeights, 2)
	// Weed and rocks at fixed places.
	for i := range 14 {
		h := hash(i)
		x := float32(4 + h%58)
		y := g.sy(seabed(x))
		img := imageWeed[(g.ticks/40+i)%2]
		if h>>8%3 == 0 {
			img = imageRock
		}
		b := img.Bounds()
		drawAt(screen, img, int(sx(x))-b.Dx()/2, int(y)-b.Dy()+2)
	}
}

// spanStart is where the span on each row began. It is kept here: on the
// stack of a microcontroller it would be too large and go to the heap.
var spanStart [ScreenHeight/bedStep + 1]int

// bedStep is the seabed's step in pixels: it is filled in rows this tall,
// half as many draws as rows one pixel tall.
const bedStep = 2

// skyline fills the screen's columns up from the bottom to their heights
// less drop, in rows bedStep pixels tall. It fills a row at a time in spans,
// and the part all columns cover at once as one rectangle: a column at a
// time would draw each pixel on its own. Going across once, a span starts on
// the rows a column rises past and ends on the rows it drops below, so only
// the edges are visited.
func skyline(screen *ebiten.Image, c color.RGBA, h *[ScreenWidth]int, drop int) {
	const rows = ScreenHeight / bedStep
	low := rows
	for _, v := range h {
		low = min(low, max(v-drop, 0)/bedStep)
	}
	prev := low
	for x := 0; x <= ScreenWidth; x++ {
		cur := low
		if x < ScreenWidth {
			cur = min(max(h[x]-drop, 0)/bedStep, rows)
		}
		for d := prev; d < cur; d++ {
			spanStart[d] = x
		}
		for d := cur; d < prev; d++ {
			fill(screen, c, spanStart[d], ScreenHeight-(d+1)*bedStep, x-spanStart[d], bedStep)
		}
		prev = cur
	}
	fill(screen, c, 0, ScreenHeight-low*bedStep, ScreenWidth, low*bedStep)
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
		if y < viewTop-16 || y > ScreenHeight+16 {
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
			wobble = sinf(float32(g.ticks)*0.8) * 3
		}
		op.GeoM.Translate(float64(int(x-w/2)), float64(int(y-h/2+wobble)))
		screen.DrawImage(img, &op)
	}
}

// rodTip is where the tip of the rod is drawn: it dips for a nibble and bends
// toward a fish on the line.
func (g *Game) rodTip() (float32, float32) {
	x, y := sx(tipX), g.sy(tipDepth)
	switch {
	case g.mode == modeFight:
		bend := 4 + g.tension*10
		x += bend * 0.4
		y += bend
	case g.twitch > 0 && g.suitor >= 0 && g.fish[g.suitor].state == bite:
		y += 8
	case g.twitch > 0:
		y += float32(g.twitch%4) * 2
	}
	return x, y
}

func (g *Game) drawBoat(screen *ebiten.Image) {
	y := g.sy(0)
	if y < viewTop-40 || y > ScreenHeight+20 {
		return
	}
	bob := sinf(float32(g.ticks)*0.04) * 1.6
	top := int(y+bob) - 4 // the deck
	drawAt(screen, imageAngler, int(sx(-1.6))-6, top-imageAngler.Bounds().Dy()+1)
	drawAt(screen, imageBoat, int(sx(-6.5)), top)
	// The rod, from the angler's hands to the tip.
	hx, hy := sx(-1.6)+8, float32(top-9)
	tx, ty := g.rodTip()
	line(screen, color.RGBA{70, 50, 30, 255}, hx, hy, tx, ty)
	line(screen, color.RGBA{120, 90, 50, 255}, hx, hy-1, tx, ty-1)
}

func (g *Game) drawLine(screen *ebiten.Image) {
	if g.mode != modeFlight && g.mode != modeFishing && g.mode != modeFight {
		return
	}
	tx, ty := g.rodTip()
	lx, ly := sx(g.lureX), g.sy(g.jigDepth())
	line(screen, color.RGBA{220, 225, 230, 255}, tx, ty, lx, ly)
	if g.mode != modeFight {
		drawAt(screen, imageLure, int(lx)-3, int(ly)-1)
	}
	if g.mode == modeFishing && (g.freeFall || g.jigOff > 0.1) && g.ticks%8 < 5 {
		drawAt(screen, imageBubble, int(lx)-2, int(ly)-10-g.ticks%8*2)
	}
	if g.mode == modeFishing && g.suitor >= 0 {
		switch g.fish[g.suitor].state {
		case bite:
			text.Reset().Str("!").Shade(textShadow).Draw(screen, int(lx)-5, int(ly)-32, 2, textYellow)
		case nibble:
			if g.twitch > 0 {
				text.Reset().Str(".").Draw(screen, int(lx)+6, int(ly)-12, 1, textWhite)
			}
		}
	}
}

// drawScale marks every ten meters of depth at the right edge.
func (g *Game) drawScale(screen *ebiten.Image) {
	for d := 10; d <= deepest; d += 10 {
		y := int(g.sy(float32(d)))
		if y < viewTop+6 || y > ScreenHeight-6 {
			continue
		}
		fill(screen, color.RGBA{200, 220, 240, 255}, ScreenWidth-6, y, 6, 1)
		text.Reset().Num(d).Right(screen, ScreenWidth-8, y-3, 1, textPale)
	}
}

var (
	textWhite  = color.RGBA{240, 244, 248, 255}
	textPale   = color.RGBA{170, 200, 225, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGreen  = color.RGBA{110, 230, 130, 255}
	textGray   = color.RGBA{150, 160, 170, 255}
	textShadow = color.RGBA{10, 16, 28, 255}
	barColor   = color.RGBA{12, 20, 34, 255}
	barLine    = color.RGBA{40, 60, 90, 255}
	boxColor   = color.RGBA{12, 18, 30, 255}
	boxEdge    = color.RGBA{80, 110, 150, 255}
	boxLight   = color.RGBA{140, 170, 210, 255}
)

var text pixtext.Line

// dec adds x with two decimals to the line.
func dec(l *pixtext.Line, x float32) *pixtext.Line {
	c := int(x*100 + 0.5)
	return l.Num(c/100).Str(".").Pad(c%100, 2)
}

// drawBar draws the clock and the score at the top, and the gauges.
func (g *Game) drawBar(screen *ebiten.Image) {
	fill(screen, barColor, 0, 0, ScreenWidth, viewTop-1)
	fill(screen, barLine, 0, viewTop-1, ScreenWidth, 1)
	h, m := g.clock/3600, g.clock/60%60
	text.Reset().Pad(h, 2).Str(":").Pad(m, 2).Draw(screen, 4, 5, 1, textWhite)
	text.Reset().Num(g.score).Str(" PTS").Right(screen, ScreenWidth-4, 5, 1, textYellow)
	switch g.mode {
	case modeCast:
		// The power gauge in the sky over the water.
		x, y := 60, viewTop+8
		fill(screen, boxEdge, x-2, y-2, 104, 12)
		fill(screen, boxColor, x-1, y-1, 102, 10)
		n := int(g.power * 100)
		c := lerp(color.RGBA{90, 200, 110, 255}, color.RGBA{255, 90, 60, 255}, g.power)
		fill(screen, c, x, y, n, 8)
		fill(screen, lerp(c, color.RGBA{255, 255, 255, 255}, 0.4), x, y, n, 2)
		text.Reset().Num(int(minCast+(maxCast-minCast)*g.power)).Str("M").Shade(textShadow).Draw(screen, x+110, y, 1, textWhite)
	case modeFishing:
		text.Reset().Str("DEPTH ").Num(int(g.jigDepth())).Str("M").Draw(screen, 76, 5, 1, textPale)
	case modeFight:
		g.drawTension(screen)
	}
}

// drawTension shows the line's tension at the bottom: green, then yellow,
// red near breaking. The line out is beside it.
func (g *Game) drawTension(screen *ebiten.Image) {
	const x, y, w = 8, ScreenHeight - 16, 160
	fill(screen, boxEdge, x-2, y-2, w+4, 12)
	fill(screen, boxColor, x-1, y-1, w+2, 10)
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
		fill(screen, c, x, y, n, 8)
		fill(screen, lerp(c, color.RGBA{255, 255, 255, 255}, 0.4), x, y, n, 2)
	}
	// The breaking point.
	fill(screen, color.RGBA{255, 255, 255, 255}, x+w*10/11, y-4, 2, 16)
	text.Reset().Num(int(g.line)).Str("M").Shade(textShadow).Draw(screen, x+w+8, y, 1, textWhite)
	if g.running {
		text.Reset().Str("RUN").Draw(screen, 116, 5, 1, textRed)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	switch g.mode {
	case modeTitle:
		box(screen, 20, 36, 200, 148)
		text.Reset().Str("SEA FISHING").Shade(boxEdge).Center(screen, mid, 46, 3, textYellow)
		text.Reset().Str("HOLD A: CAST").Center(screen, mid, 80, 1, textWhite)
		text.Reset().Str("DOWN: DROP   UP: REEL").Center(screen, mid, 94, 1, textWhite)
		text.Reset().Str("A: JIG / STRIKE").Center(screen, mid, 108, 1, textWhite)
		text.Reset().Str("FIGHT: HOLD A TO REEL").Center(screen, mid, 122, 1, textYellow)
		text.Reset().Str("B: WIND IN").Center(screen, mid, 136, 1, textWhite)
		if g.modeTicks/30%2 == 0 {
			text.Reset().Str("PRESS A").Center(screen, mid, 156, 2, textGreen)
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
		text.Reset().Str(messages[g.message])
		w := text.Width(2) + 16
		box(screen, (ScreenWidth-w)/2, viewTop+8, w, 22)
		text.Center(screen, mid, viewTop+12, 2, textWhite)
	}
}

func (g *Game) drawCatch(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	c := &g.lastLand
	box(screen, 24, 44, 192, 136)
	text.Reset().Str(speciesTypes[c.kind].name).Center(screen, mid, 52, 2, textWhite)
	img := imageFish[c.kind]
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(3, 3)
	op.GeoM.Translate(float64(mid-b.Dx()*3/2), float64(102-b.Dy()*3/2))
	screen.DrawImage(img, &op)
	if c.kind == boot {
		text.Reset().Str("NOT A FISH").Center(screen, mid, 140, 1, textGray)
	} else {
		dec(text.Reset().Num(int(c.length+0.5)).Str("CM   "), c.weight).Str("KG").Center(screen, mid, 134, 1, textWhite)
		text.Reset().Str("+").Num(c.points).Str(" PTS").Center(screen, mid, 148, 1, textYellow)
		if g.newBest && g.catches[c.kind] > 1 {
			text.Reset().Str("BIGGEST TODAY").Center(screen, mid, 162, 1, textGreen)
		}
	}
}

func (g *Game) drawResult(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	box(screen, 16, 20, 208, 204)
	text.Reset().Str("SUNSET").Shade(boxEdge).Center(screen, mid, 28, 2, textYellow)
	y := 54
	for k := range kinds {
		if g.catches[k] == 0 {
			continue
		}
		text.Reset().Str(speciesTypes[k].name).Draw(screen, 28, y, 1, textWhite)
		text.Reset().Str("x").Num(g.catches[k]).Draw(screen, 112, y, 1, textWhite)
		if k != boot {
			text.Reset().Num(int(g.best[k]+0.5)).Str("CM").Right(screen, 212, y, 1, textPale)
		}
		y += 14
	}
	if y == 54 {
		text.Reset().Str("NO FISH TODAY").Center(screen, mid, y, 1, textGray)
	}
	dec(text.Reset(), g.total).Str(" KG").Center(screen, mid, 172, 1, textWhite)
	text.Reset().Num(g.score).Str(" PTS").Center(screen, mid, 186, 2, textYellow)
	if g.modeTicks > 60 && g.modeTicks/30%2 == 0 {
		text.Reset().Str("PRESS A").Center(screen, mid, 208, 1, textGreen)
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

// line draws a one-pixel line as runs: each stretch along the line's long
// way that stays on one row or column is filled at once, as one pixel at a
// time would cost a draw call each.
func line(screen *ebiten.Image, c color.RGBA, x0, y0, x1, y1 float32) {
	dx, dy := x1-x0, y1-y0
	n := int(max(abs(dx), abs(dy))) + 1
	steep := abs(dy) > abs(dx)
	// The run from (ax, ay) to (bx, by), where only the long way changes.
	ax, ay := int(x0), int(y0)
	bx, by := ax, ay
	for i := 1; i <= n; i++ {
		t := float32(i) / float32(n)
		px, py := int(x0+dx*t), int(y0+dy*t)
		if (steep && px == ax) || (!steep && py == ay) {
			bx, by = px, py
			continue
		}
		fill(screen, c, min(ax, bx), min(ay, by), iabs(bx-ax)+1, iabs(by-ay)+1)
		ax, ay, bx, by = px, py, px, py
	}
	fill(screen, c, min(ax, bx), min(ay, by), iabs(bx-ax)+1, iabs(by-ay)+1)
}

func iabs(v int) int { return max(v, -v) }

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
