package train

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

// The view from the cab: a camera camHeight meters above the rails looks
// along the track. Each screen row below the horizon shows the ground at one
// distance; the track's sideways offset at that distance comes from the
// curves ahead.
const (
	viewHeight = 164 // rows of the view; the gauges are below
	horizon    = 72
	focal      = 200 // pixels per meter at one meter
	camHeight  = 3.0
	near       = 3.0
	offStep    = 5 // meters between samples of the track's offset
	offSamples = 72
	farObjects = 300

	platformHeight = 1.1
	canopyHeight   = 4.0
	wireHeight     = 5.5
	gauge          = 1.067
)

// rowZ is the distance shown by each row below the horizon.
var rowZ [viewHeight + 1]float32

func init() {
	for y := horizon + 1; y <= viewHeight; y++ {
		rowZ[y] = camHeight * focal / float32(y-horizon)
	}
}

func rgb(r, g, b uint8) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.RGBA{r, g, b, 255})
	return img
}

var (
	colorSky     = [4]*ebiten.Image{rgb(96, 150, 214), rgb(122, 170, 224), rgb(150, 190, 232), rgb(180, 210, 238)}
	colorGround  = [4][3]*ebiten.Image{}
	colorBallast = [3]*ebiten.Image{rgb(128, 124, 118), rgb(112, 108, 102), rgb(120, 116, 110)}
	colorSleeper = rgb(84, 70, 60)
	colorRail    = rgb(200, 200, 206)
	colorWater   = [3]*ebiten.Image{rgb(46, 100, 168), rgb(56, 112, 180), rgb(50, 106, 174)}
	colorSand    = rgb(214, 198, 150)
	colorDeck    = rgb(86, 78, 72)
	colorGirder  = rgb(150, 54, 40)
	colorRoad    = rgb(70, 70, 74)
	colorYard    = [3]*ebiten.Image{rgb(118, 112, 104), rgb(108, 102, 96), rgb(113, 107, 100)}
	colorPlatTop = rgb(198, 196, 188)
	colorPlatEdg = rgb(232, 232, 226)
	colorTactile = rgb(240, 200, 40)
	colorWall    = rgb(118, 114, 106)
	colorRoof    = rgb(92, 88, 84)
	colorRoofEdg = rgb(60, 58, 56)
	colorPole    = rgb(72, 72, 78)
	colorWire    = rgb(44, 44, 48)
	colorFrame   = rgb(30, 30, 34)
	colorPanel   = rgb(26, 28, 34)
	colorPanelHi = rgb(70, 74, 84)
	colorBarBack = rgb(50, 54, 62)
	colorGreen   = rgb(80, 200, 110)
	colorRed     = rgb(230, 50, 40)
	colorYellow  = rgb(255, 220, 40)
	colorCell    = rgb(48, 50, 58)
	colorPowerOn = rgb(90, 160, 255)
	colorBrakeOn = rgb(255, 150, 50)
	colorWhite   = rgb(240, 240, 240)
	colorBox     = rgb(16, 20, 32)
	colorBoxLit  = rgb(120, 126, 140)
	colorRailLit = rgb(236, 236, 242)
)

var (
	textWhite  = color.RGBA{240, 240, 240, 255}
	textYellow = color.RGBA{255, 220, 40, 255}
	textGray   = color.RGBA{150, 154, 166, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGreen  = color.RGBA{110, 230, 140, 255}
	textPower  = color.RGBA{110, 180, 255, 255}
	textBrake  = color.RGBA{255, 160, 60, 255}
	textBlack  = color.RGBA{0, 0, 0, 255}
	textShadow = color.RGBA{10, 12, 20, 255}
)

var text pixtext.Line

func init() {
	colorGround[field] = [3]*ebiten.Image{rgb(92, 152, 70), rgb(80, 138, 62), rgb(86, 145, 66)}
	colorGround[town] = [3]*ebiten.Image{rgb(150, 144, 128), rgb(138, 132, 118), rgb(144, 138, 123)}
	colorGround[coast] = [3]*ebiten.Image{rgb(110, 160, 80), rgb(98, 148, 72), rgb(104, 154, 76)}
	colorGround[river] = colorGround[field]
}

type objKind uint8

const (
	objTree objKind = iota
	objPine
	objHouse
	objFlats
	objPole
	objLimit
	objWhistle
	objCrossing
	objDistance
	objStopMark
	objBoard
	objPillar
	objBuffer
)

// spriteHeight is the height of each kind drawn from an image, in meters.
var spriteHeight = [...]float32{
	objTree:     7,
	objPine:     9,
	objHouse:    6,
	objFlats:    11,
	objLimit:    3.2,
	objWhistle:  2.6,
	objCrossing: 3.6,
	objDistance: 2.2,
	objStopMark: 1.7,
	objBoard:    2.0,
	objBuffer:   1.3,
}

type object struct {
	z, x, base float32 // distance, sideways offset and the height it stands on, in meters
	kind       objKind
	arg        int
}

const maxObjects = 64

// view is the per-frame state of drawing.
type view struct {
	off     [offSamples + 1]float32
	objects [maxObjects]object
	n       int
}

func (g *Game) Draw(screen *ebiten.Image) {
	v := &g.view
	mark(0)
	v.trackOffsets(g.pos)
	g.drawSky(screen)
	mark(1)
	g.drawGround(screen)
	mark(2)
	g.drawStations(screen)
	mark(3)
	g.collectObjects()
	g.drawObjects(screen)
	mark(4)
	g.drawWire(screen)
	fillRect(screen, colorFrame, 0, 0, ScreenWidth, 6)
	g.drawNotch(screen)
	mark(5)
	g.drawGauges(screen)
	g.drawOverlay(screen)
	mark(6)
}

// trackOffsets integrates the curves ahead into the track's sideways offset.
func (v *view) trackOffsets(pos float32) {
	var x, heading float32
	v.off[0] = 0
	for i := 1; i <= offSamples; i++ {
		k := curvatureAt(pos + (float32(i)-0.5)*offStep)
		x += heading*offStep + k*offStep*offStep/2
		heading += k * offStep
		v.off[i] = x
	}
}

// offset is the track's sideways offset at distance z.
func (v *view) offset(z float32) float32 {
	f := z / offStep
	i := int(f)
	if i >= offSamples {
		return v.off[offSamples]
	}
	if i < 0 {
		return 0
	}
	t := f - float32(i)
	return v.off[i]*(1-t) + v.off[i+1]*t
}

// sx is the screen column of a point x meters beside the track at distance z.
func (v *view) sx(x, z float32) float32 {
	return v.at(z).x(x)
}

// columns maps meters beside the track to screen columns at one distance.
type columns struct{ center, ppm float32 }

// at returns the columns at distance z.
func (v *view) at(z float32) columns {
	ppm := focal / z
	return columns{ScreenWidth/2 + v.offset(z)*ppm, ppm}
}

func (c columns) x(x float32) float32 { return c.center + x*c.ppm }

// sy is the screen row of a point h meters above the rails at distance z.
func sy(h, z float32) float32 {
	return horizon + (camHeight-h)*focal/z
}

func (g *Game) drawSky(screen *ebiten.Image) {
	band := horizon / len(colorSky)
	for i, c := range colorSky {
		fillRect(screen, c, 0, i*band, ScreenWidth, band+1)
	}
	fillRect(screen, colorSky[3], 0, band*len(colorSky), ScreenWidth, horizon-band*len(colorSky)+1)
	// The hills are far away: they move only as the line turns.
	x := -int(g.heading*focal) % hillWidth
	if x > 0 {
		x -= hillWidth
	}
	for ; x < ScreenWidth; x += hillWidth {
		drawAt(screen, imageHills, x, horizon+1-hillHeight)
	}
}

// span fills a row from x0 to x1, in either order.
func span(screen, c *ebiten.Image, x0, x1 float32, y int) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	l, r := int(x0), int(x1+0.999)
	if r == l {
		r++
	}
	l, r = max(l, 0), min(r, ScreenWidth)
	fillRect(screen, c, l, y, r-l, 1)
}

// ground is what lies under a row before the track: its color across the
// whole row.
func (g *Game) groundRow(y int) *ebiten.Image {
	z := rowZ[y]
	dz := z - rowZ[y+1] // meters covered by the row
	s := g.pos + z
	shade := 2
	if dz < 4 {
		shade = int(s/12) & 1
	}
	kind := sceneryAt(s)
	switch {
	case onCrossing(s):
		return colorRoad
	case kind == river:
		return colorWater[shade]
	case atStation(s):
		return colorYard[shade]
	}
	return colorGround[kind][shade]
}

// drawGround fills the ground row by row from the horizon down, the rows
// alike at once, and then draws the track and the shore over it.
func (g *Game) drawGround(screen *ebiten.Image) {
	v := &g.view
	from, last := horizon+1, g.groundRow(horizon+1)
	for y := horizon + 2; y <= viewHeight; y++ {
		var c *ebiten.Image
		if y < viewHeight {
			c = g.groundRow(y)
		}
		if c != last {
			fillRect(screen, last, 0, from, ScreenWidth, y-from)
			from, last = y, c
		}
	}
	for y := horizon + 1; y < viewHeight; y++ {
		z := rowZ[y]
		dz := z - rowZ[y+1]
		s := g.pos + z
		shade := 2
		if dz < 4 {
			shade = int(s/12) & 1
		}
		col := v.at(z)
		kind := sceneryAt(s)
		crossing := onCrossing(s)
		if kind == coast && !crossing {
			span(screen, colorWater[shade], 0, col.x(-7), y)
			span(screen, colorSand, col.x(-7), col.x(-2.2), y)
		}
		if s > lineEnd {
			continue
		}
		switch {
		case kind == river:
			span(screen, colorDeck, col.x(-2), col.x(2), y)
			span(screen, colorGirder, col.x(-2.2), col.x(-1.9), y)
			span(screen, colorGirder, col.x(1.9), col.x(2.2), y)
		case !crossing:
			span(screen, colorBallast[shade], col.x(-1.6), col.x(1.6), y)
		}
		if dz < 0.5 && !crossing {
			// Sleepers show only where a row covers less than one.
			if f := s / 0.7; f-float32(int(f)) < 0.4 {
				span(screen, colorSleeper, col.x(-1.1), col.x(1.1), y)
			}
		}
		w := max(float32(1), 0.07*col.ppm)
		for _, side := range [2]float32{-gauge / 2, gauge / 2} {
			x := col.x(side)
			span(screen, colorRail, x-w/2, x+w/2, y)
			if w >= 3 {
				// The rail's head catches the light.
				span(screen, colorRailLit, x-w/2, x, y)
			}
		}
	}
}

// drawStations draws the platforms on the left, their faces toward the
// track, and the canopies over them. A raised surface at height h shows at
// distance (camHeight-h)*focal/dy on the row dy below the horizon.
func (g *Game) drawStations(screen *ebiten.Image) {
	v := &g.view
	for _, st := range stations {
		z0, z1 := st.stop-platformBack-g.pos, st.stop+platformFore-g.pos
		if z1 < near || z0 > farObjects+100 {
			continue
		}
		for y := horizon + 1; y < viewHeight; y++ {
			dy := float32(y - horizon)
			zt := (camHeight - platformHeight) * focal / dy
			lo, hi := max(zt, z0, near), min(rowZ[y], z1)
			if lo < hi {
				span(screen, colorWall, v.sx(-1.6, lo), v.sx(-1.6, hi), y)
			}
			if zt >= max(z0, near) && zt <= z1 {
				col := v.at(zt)
				edge := col.x(-1.6)
				span(screen, colorPlatTop, col.x(-9), edge, y)
				span(screen, colorTactile, col.x(-2.5), col.x(-2.1), y)
				span(screen, colorPlatEdg, col.x(-1.75), edge, y)
			}
		}
		c0, c1 := st.stop-canopyBack-g.pos, st.stop-canopyFore-g.pos
		for y := horizon - 1; y >= 6; y-- {
			zr := (canopyHeight - camHeight) * focal / float32(horizon-y)
			if zr >= max(c0, near) && zr <= c1 {
				col := v.at(zr)
				edge := col.x(-2.4)
				span(screen, colorRoof, col.x(-9), edge, y)
				span(screen, colorRoofEdg, edge-1, edge, y)
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

func (g *Game) add(z, x, base float32, kind objKind, arg int) {
	v := &g.view
	if z < near || z > farObjects || v.n == maxObjects {
		return
	}
	v.objects[v.n] = object{z: z, x: x, base: base, kind: kind, arg: arg}
	v.n++
}

// collectObjects lists what stands beside the line within view.
func (g *Game) collectObjects() {
	g.view.n = 0
	from, to := g.pos+near, g.pos+farObjects
	// Catenary poles on the right.
	for k := int(from/50) + 1; float32(k*50) < to; k++ {
		if s := float32(k * 50); s < lineEnd {
			g.add(s-g.pos, 2.6, 0, objPole, 0)
		}
	}
	// Trees and houses, placed by a hash of 23 m slots.
	for k := int(from / 23); float32(k*23) < to; k++ {
		for side := -1; side <= 1; side += 2 {
			h := hash(k*2 + side)
			s := float32(k*23) + float32(h%13)
			kind := sceneryAt(s)
			if h%5 == 0 || kind == river || side < 0 && (kind == coast || atStation(s)) {
				continue
			}
			obj := objTree
			switch {
			case kind == town && h&0x100 != 0:
				obj = objFlats
			case kind == town:
				obj = objHouse
			case h&0x100 != 0:
				obj = objPine
			}
			x := float32(side) * float32(7+h>>9%14)
			if obj == objFlats {
				x *= 1.6
			}
			g.add(s-g.pos, x, 0, obj, int(h>>4&1))
		}
	}
	for i, l := range limits {
		if i > 0 {
			g.add(l.from-g.pos, 2.2, 0, objLimit, l.kmh)
		}
	}
	for _, c := range crossings {
		g.add(c-whistleAhead-g.pos, 2.2, 0, objWhistle, 0)
		g.add(c-4-g.pos, -3, 0, objCrossing, 0)
		g.add(c-4-g.pos, 3, 0, objCrossing, 1)
	}
	for i, st := range stations {
		if i > 0 {
			for n := 1; n <= 3; n++ {
				g.add(st.stop-float32(n*100)-g.pos, -2.4, 0, objDistance, n)
			}
		}
		g.add(st.stop-g.pos, -2.0, platformHeight, objStopMark, 0)
		g.add(st.stop-45-g.pos, -5.2, platformHeight, objBoard, 0)
		for s := st.stop - canopyBack + 5; s <= st.stop-canopyFore; s += 15 {
			g.add(s-g.pos, -6, platformHeight, objPillar, 0)
		}
	}
	last := stations[len(stations)-1]
	g.add(last.stop+bufferStop+1-g.pos, 0, 0, objBuffer, 0)
}

// drawObjects draws far to near, so that near objects cover far ones.
func (g *Game) drawObjects(screen *ebiten.Image) {
	v := &g.view
	objs := v.objects[:v.n]
	for i := 1; i < len(objs); i++ {
		for j := i; j > 0 && objs[j].z > objs[j-1].z; j-- {
			objs[j], objs[j-1] = objs[j-1], objs[j]
		}
	}
	for i := range objs {
		g.drawObject(screen, &objs[i])
	}
}

func (g *Game) drawObject(screen *ebiten.Image, o *object) {
	v := &g.view
	col := v.at(o.z)
	x, bottom, ppm := col.x(o.x), sy(o.base, o.z), col.ppm
	switch o.kind {
	case objPole:
		top := sy(6.2, o.z)
		w := max(1, int(0.25*ppm))
		fillRect(screen, colorPole, int(x)-w/2, int(top), w, int(bottom-top)+1)
		arm := sy(5.8, o.z)
		end := col.x(-0.3)
		span(screen, colorPole, end, x, int(arm))
		if h := int(0.12 * ppm); h > 1 {
			fillRect(screen, colorPole, int(end), int(arm), int(x-end), h)
		}
		return
	case objPillar:
		top := sy(canopyHeight, o.z)
		w := max(1, int(0.3*ppm))
		fillRect(screen, colorRoofEdg, int(x)-w/2, int(top), w, int(bottom-top)+1)
		return
	}
	var img *ebiten.Image
	switch o.kind {
	case objTree:
		img = imageTree
	case objPine:
		img = imagePine
	case objHouse:
		img = imageHouse[0]
	case objFlats:
		img = imageHouse[1]
	case objLimit:
		img = imageLimit[o.arg]
	case objWhistle:
		img = imageWhistle
	case objCrossing:
		img = imageCrossing[(g.ticks/20+o.arg)%2]
	case objDistance:
		img = imageDistance[o.arg]
	case objStopMark:
		img = imageStopMark
	case objBoard:
		img = imageBoard
	case objBuffer:
		img = imageBuffer
	}
	scale := spriteHeight[o.kind] * ppm / float32(img.Bounds().Dy())
	if scale*float32(img.Bounds().Dy()) < 1 {
		return
	}
	drawScaled(screen, img, x, bottom, scale)
}

// drawWire draws the contact wire over the track, one pixel wide, a run of
// rows at a time where it stays in one column.
func (g *Game) drawWire(screen *ebiten.Image) {
	v := &g.view
	runX, runY, runN := 0, 0, 0
	for y := horizon - 1; y >= 6; y-- {
		z := (wireHeight - camHeight) * focal / float32(horizon-y)
		if z < near || g.pos+z > lineEnd {
			continue
		}
		x := int(v.sx(0, z))
		if runN > 0 && x == runX && y == runY-1 {
			runY, runN = y, runN+1
			continue
		}
		fillRect(screen, colorWire, runX, runY, 1, runN)
		runX, runY, runN = x, y, 1
	}
	fillRect(screen, colorWire, runX, runY, 1, runN)
}

// drawNotch shows the handle: P5 at the top, then N, the brakes and EB.
func (g *Game) drawNotch(screen *ebiten.Image) {
	const x, top, cell = 222, 12, 9
	fillRect(screen, colorFrame, x-3, top-3, 16, (maxPower+maxBrake+2)*cell+6)
	for i := maxPower; i >= notchEB; i-- {
		y := top + (maxPower-i)*cell
		if i == notchEB {
			y += 2
		}
		c := colorCell
		if i == 0 {
			c = colorPanelHi
		}
		switch {
		case i > 0 && g.notch >= i:
			c = colorPowerOn
		case i == 0 && g.notch == 0:
			c = colorWhite
		case i < 0 && i > notchEB && g.notch <= i:
			c = colorBrakeOn
		case i == notchEB && g.notch == notchEB:
			c = colorRed
		}
		fillRect(screen, c, x, y, 10, cell-2)
	}
}

// clock adds a time of day.
func clock(l *pixtext.Line, seconds int) *pixtext.Line {
	return l.Pad(seconds/3600%24, 2).Str(":").Pad(seconds/60%60, 2).Str(":").Pad(seconds%60, 2)
}

// signed adds n with a plus sign when it is positive.
func signed(l *pixtext.Line, n int) *pixtext.Line {
	if n > 0 {
		l.Str("+")
	}
	return l.Num(n)
}

// meters adds a distance, with tenths below ten meters.
func meters(l *pixtext.Line, d float32, plus bool) *pixtext.Line {
	tenths := int(max(d, -d)*10 + 0.5)
	switch {
	case d < 0 && tenths > 0:
		l.Str("-")
	case plus && tenths > 0:
		l.Str("+")
	}
	if tenths < 100 {
		l.Num(tenths / 10).Str(".").Num(tenths % 10)
	} else {
		l.Num((tenths + 5) / 10)
	}
	return l.Str("M")
}

func notch(l *pixtext.Line, n int) *pixtext.Line {
	switch {
	case n == notchEB:
		return l.Str("EB")
	case n > 0:
		return l.Str("P").Num(n)
	case n < 0:
		return l.Str("B").Num(-n)
	}
	return l.Str("N")
}

func (g *Game) drawGauges(screen *ebiten.Image) {
	fillRect(screen, colorPanel, 0, viewHeight, ScreenWidth, ScreenHeight-viewHeight)
	fillRect(screen, colorPanelHi, 0, viewHeight, ScreenWidth, 2)
	kmh := int(g.v * 3.6)
	lim := g.limitNow()
	c := textWhite
	if kmh > lim {
		c = textRed
	}
	text.Reset().Num(kmh).Right(screen, 62, 170, 3, c)
	text.Reset().Str("KM/H").Draw(screen, 68, 184, 1, textGray)
	// The speed bar, a pixel per km/h, with the limit marked.
	fillRect(screen, colorBarBack, 8, 197, 120, 6)
	bar := colorGreen
	if kmh > lim {
		bar = colorRed
	}
	fillRect(screen, bar, 8, 197, min(kmh, 120), 6)
	fillRect(screen, colorYellow, 8+lim, 193, 2, 14)
	// The limit, as a sign: black on white with a red rim.
	fillRect(screen, colorRed, 136, 169, 40, 30)
	fillRect(screen, colorWhite, 139, 172, 34, 24)
	text.Reset().Num(lim).Center(screen, 156, 174, 3, textBlack)
	nc := textWhite
	switch {
	case g.notch == notchEB:
		nc = textRed
	case g.notch > 0:
		nc = textPower
	case g.notch < 0:
		nc = textBrake
	}
	notch(text.Reset(), g.notch).Draw(screen, 186, 172, 3, nc)
	if p := int(gradeAt(g.pos - trainLength/2)); p > 0 {
		text.Reset().Str("/").Num(p).Draw(screen, 186, 200, 1, textGray)
	} else if p < 0 {
		text.Reset().Str("\\").Num(-p).Draw(screen, 186, 200, 1, textGray)
	}
	st := stations[min(g.next, len(stations)-1)]
	text.Reset().Str(st.name).Draw(screen, 8, 211, 2, textWhite)
	dc := textWhite
	if st.stop-g.pos < 100 {
		dc = textYellow
	}
	meters(text.Reset(), st.stop-g.pos, false).Right(screen, ScreenWidth-6, 211, 2, dc)
	// The time now, and the scheduled arrival.
	clock(text.Reset(), startSeconds+g.clock/tps).Draw(screen, 8, 229, 1, textWhite)
	text.Reset().Str(">").Draw(screen, 62, 229, 1, textGray)
	clock(text.Reset(), startSeconds+st.arrive).Draw(screen, 74, 229, 1, textYellow)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	if g.horn > 0 {
		text.Reset().Str("PAAN!").Shade(textShadow).Center(screen, mid, 12, 2, textWhite)
	}
	switch g.mode {
	case modeTitle:
		box(screen, 20, 24, 200, 132)
		text.Reset().Str("IKA LINE").Shade(colorShade).Center(screen, mid, 34, 3, textYellow)
		text.Reset().Str("TRAIN DRIVER").Center(screen, mid, 64, 2, textWhite)
		if g.modeTicks/30%2 == 0 {
			text.Reset().Str("PRESS A").Center(screen, mid, 90, 2, textGreen)
		}
		text.Reset().Str("UP: POWER   DOWN: BRAKE").Center(screen, mid, 120, 1, textWhite)
		text.Reset().Str("A: HORN   B: NEUTRAL").Center(screen, mid, 136, 1, textWhite)
		return
	case modeStation:
		g.drawResult(screen)
		return
	case modeEnd:
		box(screen, 20, 20, 200, 140)
		text.Reset().Str("LAST STOP").Shade(colorShade).Center(screen, mid, 30, 2, textYellow)
		text.Reset().Str("SCORE").Draw(screen, 40, 66, 1, textWhite)
		text.Reset().Num(g.total()).Right(screen, 200, 58, 3, textYellow)
		text.Reset().Str("RANK").Draw(screen, 40, 104, 1, textWhite)
		text.Reset().Str(g.rank()).Right(screen, 200, 96, 3, textYellow)
		if g.modeTicks > 60 && g.modeTicks/30%2 == 0 {
			text.Reset().Str("PRESS A").Center(screen, mid, 136, 1, textGreen)
		}
		return
	}
	if g.messageTicks > 0 && g.message != msgNone && g.messageTicks%30 < 22 {
		text.Reset().Str(messages[g.message])
		w := text.Width(2) + 16
		box(screen, (ScreenWidth-w)/2-8, 88, w, 22)
		text.Center(screen, mid-8, 92, 2, textWhite)
	}
}

// drawResult shows the score of the run to the station.
func (g *Game) drawResult(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	r := &g.results[g.next]
	box(screen, 16, 12, 196, 148)
	text.Reset().Str(stations[g.next].name).Shade(colorShade).Center(screen, mid-6, 20, 2, textYellow)
	for i, label := range [...]string{"STOP", "TIME", "RIDE", "SPEED", "HORN"} {
		y := 46 + i*16
		text.Reset().Str(label).Draw(screen, 26, y, 1, textWhite)
		pts := [...]int{r.stopPts, r.timePts, r.ridePts, r.speedPts, r.hornPts}[i]
		if i > 2 {
			signed(text.Reset(), pts)
		} else {
			text.Reset().Num(pts)
		}
		text.Right(screen, 202, y, 1, textWhite)
	}
	sc := textGreen
	if r.stopPts < 40 {
		sc = textYellow
	}
	meters(text.Reset(), r.stopError, true).Draw(screen, 96, 46, 1, sc)
	tc := textGreen
	if r.timePts < 20 {
		tc = textYellow
	}
	signed(text.Reset(), r.delay).Str("S").Draw(screen, 96, 62, 1, tc)
	fillRect(screen, colorPanelHi, 24, 128, 180, 1)
	text.Reset().Str("TOTAL").Draw(screen, 26, 138, 1, textWhite)
	text.Reset().Num(r.total).Right(screen, 202, 134, 2, textYellow)
}

// box is a dark panel with a two-tone frame, drawn side by side rather than
// one over another.
func box(screen *ebiten.Image, x, y, w, h int) {
	fillRect(screen, colorBoxLit, x-2, y-2, w+3, 1)
	fillRect(screen, colorPanelHi, x+w+1, y-2, 1, 1)
	fillRect(screen, colorPanelHi, x-2, y-1, w+4, 1)
	fillRect(screen, colorPanelHi, x-2, y, 2, h)
	fillRect(screen, colorPanelHi, x+w, y, 2, h)
	fillRect(screen, colorPanelHi, x-2, y+h, w+4, 2)
	fillRect(screen, colorBox, x, y, w, h)
}

var colorShade = color.RGBA{70, 74, 84, 255}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

// drawScaled draws img scaled with its bottom center at x, y.
func drawScaled(screen, img *ebiten.Image, x, y, scale float32) {
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x-float32(b.Dx())*scale/2), float64(y-float32(b.Dy())*scale))
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
