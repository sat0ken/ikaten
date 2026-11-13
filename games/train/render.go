package train

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

// The view from the cab: a camera camHeight meters above the rails looks
// along the track. Each screen row below the horizon shows the ground at one
// distance; the track's sideways offset at that distance comes from the
// curves ahead.
const (
	viewHeight = 82 // rows of the view; the gauges are below
	horizon    = 36
	focal      = 100 // pixels per meter at one meter
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
)

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
	v.trackOffsets(g.pos)
	g.drawSky(screen)
	g.drawGround(screen)
	g.drawStations(screen)
	g.collectObjects()
	g.drawObjects(screen)
	g.drawWire(screen)
	fillRect(screen, colorFrame, 0, 0, ScreenWidth, 3)
	g.drawNotch(screen)
	g.drawGauges(screen)
	g.drawOverlay(screen)
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

func (g *Game) drawGround(screen *ebiten.Image) {
	v := &g.view
	for y := horizon + 1; y < viewHeight; y++ {
		z := rowZ[y]
		dz := z - rowZ[y+1] // meters covered by the row
		s := g.pos + z
		shade := 2
		if dz < 4 {
			shade = int(s/12) & 1
		}
		col := v.at(z)
		kind := sceneryAt(s)
		ground := colorGround[kind][shade]
		if atStation(s) && kind != river {
			ground = colorYard[shade]
		}
		switch kind {
		case river:
			fillRect(screen, colorWater[shade], 0, y, ScreenWidth, 1)
		case coast:
			fillRect(screen, ground, 0, y, ScreenWidth, 1)
			span(screen, colorWater[shade], 0, col.x(-7), y)
			span(screen, colorSand, col.x(-7), col.x(-2.2), y)
		default:
			fillRect(screen, ground, 0, y, ScreenWidth, 1)
		}
		crossing := onCrossing(s)
		if crossing {
			fillRect(screen, colorRoad, 0, y, ScreenWidth, 1)
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
		for y := horizon - 1; y >= 3; y-- {
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

// drawWire draws the contact wire over the track, one pixel on each row above
// the horizon.
func (g *Game) drawWire(screen *ebiten.Image) {
	v := &g.view
	for y := horizon - 1; y >= 3; y-- {
		z := (wireHeight - camHeight) * focal / float32(horizon-y)
		if z < near || g.pos+z > lineEnd {
			continue
		}
		fillRect(screen, colorWire, int(v.sx(0, z)), y, 1, 1)
	}
}

// drawNotch shows the handle: P5 at the top, then N, the brakes and EB.
func (g *Game) drawNotch(screen *ebiten.Image) {
	const x, top, cell = 112, 6, 5
	fillRect(screen, colorFrame, x-2, top-2, 9, (maxPower+maxBrake+2)*cell+3)
	for i := maxPower; i >= notchEB; i-- {
		y := top + (maxPower-i)*cell
		if i == notchEB {
			y++
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
		fillRect(screen, c, x, y, 5, cell-1)
	}
}

// text builds a line of the 3x5 font in a fixed buffer, so that drawing
// numbers makes no strings and does not allocate.
type text struct {
	b [24]byte
	n int
}

var buf text

func (t *text) reset() *text { t.n = 0; return t }

func (t *text) str(s string) *text {
	for i := range len(s) {
		t.b[t.n] = s[i]
		t.n++
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
	for ; i < len(d); i++ {
		t.b[t.n] = d[i]
		t.n++
	}
	return t
}

// signed writes n with a plus sign when it is positive.
func (t *text) signed(n int) *text {
	if n > 0 {
		t.str("+")
	}
	return t.num(n)
}

func (t *text) two(n int) *text {
	t.b[t.n], t.b[t.n+1] = byte('0'+n/10), byte('0'+n%10)
	t.n += 2
	return t
}

func (t *text) clock(seconds int) *text {
	return t.two(seconds / 3600 % 24).str(":").two(seconds / 60 % 60).str(":").two(seconds % 60)
}

// meters writes a distance, with tenths below ten meters.
func (t *text) meters(d float32, plus bool) *text {
	tenths := int(max(d, -d)*10 + 0.5)
	switch {
	case d < 0 && tenths > 0:
		t.str("-")
	case plus && tenths > 0:
		t.str("+")
	}
	if tenths < 100 {
		t.num(tenths / 10).str(".").num(tenths % 10)
	} else {
		t.num((tenths + 5) / 10)
	}
	return t.str("m")
}

func (t *text) notch(n int) *text {
	switch {
	case n == notchEB:
		return t.str("EB")
	case n > 0:
		return t.str("P").num(n)
	case n < 0:
		return t.str("B").num(-n)
	}
	return t.str("N")
}

// draw draws the line with its left edge at x, or its right edge at -x.
func (t *text) draw(screen *ebiten.Image, x, y, scale int, c color.RGBA) {
	if x < 0 {
		x = -x - t.n*4*scale + scale
	}
	drawText(screen, t.b[:t.n], x, y, scale, c)
}

func (g *Game) drawGauges(screen *ebiten.Image) {
	fillRect(screen, colorPanel, 0, viewHeight, ScreenWidth, ScreenHeight-viewHeight)
	fillRect(screen, colorPanelHi, 0, viewHeight, ScreenWidth, 1)
	kmh := int(g.v * 3.6)
	lim := g.limitNow()
	c := textWhite
	if kmh > lim {
		c = textRed
	}
	buf.reset().num(kmh).draw(screen, -27, 85, 2, c)
	ebitenutil.DebugPrintAt(screen, "km/h", 30, 84)
	// The speed bar, half a pixel per km/h, with the limit marked.
	fillRect(screen, colorBarBack, 4, 97, 60, 3)
	bar := colorGreen
	if kmh > lim {
		bar = colorRed
	}
	fillRect(screen, bar, 4, 97, min(kmh, 120)/2, 3)
	fillRect(screen, colorYellow, 4+lim/2, 95, 1, 7)
	// The limit, as a sign: black on white with a red rim.
	fillRect(screen, colorRed, 64, 84, 19, 14)
	fillRect(screen, colorWhite, 65, 85, 17, 12)
	buf.reset().num(lim).draw(screen, 66, 86, 2, textBlack)
	nc := textWhite
	switch {
	case g.notch == notchEB:
		nc = textRed
	case g.notch > 0:
		nc = textPower
	case g.notch < 0:
		nc = textBrake
	}
	buf.reset().notch(g.notch).draw(screen, 88, 86, 2, nc)
	if p := int(gradeAt(g.pos - trainLength/2)); p > 0 {
		buf.reset().str("/").num(p).draw(screen, 88, 98, 1, textGray)
	} else if p < 0 {
		buf.reset().str("\\").num(-p).draw(screen, 88, 98, 1, textGray)
	}
	st := stations[min(g.next, len(stations)-1)]
	ebitenutil.DebugPrintAt(screen, st.name, 3, 99)
	dc := textWhite
	if st.stop-g.pos < 100 {
		dc = textYellow
	}
	buf.reset().meters(st.stop-g.pos, false).draw(screen, -(ScreenWidth - 3), 104, 2, dc)
	// The time now, and the scheduled arrival.
	buf.reset().clock(startSeconds+g.clock/tps).draw(screen, 4, 113, 1, textWhite)
	buf.reset().str(">").draw(screen, 38, 113, 1, textGray)
	buf.reset().clock(startSeconds+st.arrive).draw(screen, 44, 113, 1, textYellow)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	if g.horn > 0 {
		ebitenutil.DebugPrintAt(screen, "PAAN!", 45, 2)
	}
	switch g.mode {
	case modeTitle:
		box(screen, 10, 14, 100, 62)
		ebitenutil.DebugPrintAt(screen, "IKA LINE", 36, 12)
		ebitenutil.DebugPrintAt(screen, "TRAIN DRIVER", 24, 22)
		if g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 36)
		}
		ebitenutil.DebugPrintAt(screen, "^ POWER v BRAKE", 15, 50)
		ebitenutil.DebugPrintAt(screen, "A HORN  B N", 27, 60)
		return
	case modeStation:
		g.drawResult(screen)
		return
	case modeEnd:
		box(screen, 10, 10, 100, 64)
		ebitenutil.DebugPrintAt(screen, "LAST STOP", 33, 8)
		ebitenutil.DebugPrintAt(screen, "SCORE", 22, 24)
		buf.reset().num(g.total()).draw(screen, -97, 30, 2, textYellow)
		ebitenutil.DebugPrintAt(screen, "RANK", 22, 40)
		buf.reset().str(g.rank()).draw(screen, -97, 46, 2, textYellow)
		if g.modeTicks > 60 && g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 58)
		}
		return
	}
	if g.messageTicks > 0 && g.message != msgNone && g.messageTicks%30 < 22 {
		msg := messages[g.message]
		w := len(msg)*6 + 6
		box(screen, (ScreenWidth-w)/2-4, 44, w, 12)
		ebitenutil.DebugPrintAt(screen, msg, (ScreenWidth-len(msg)*6)/2-4, 42)
	}
}

// drawResult shows the score of the run to the station.
func (g *Game) drawResult(screen *ebiten.Image) {
	r := &g.results[g.next]
	box(screen, 8, 6, 98, 74)
	name := stations[g.next].name
	ebitenutil.DebugPrintAt(screen, name, (ScreenWidth-len(name)*6)/2-4, 4)
	sc := textGreen
	if r.stopPts < 40 {
		sc = textYellow
	}
	buf.reset().meters(r.stopError, true).draw(screen, 48, 21, 1, sc)
	tc := textGreen
	if r.timePts < 20 {
		tc = textYellow
	}
	buf.reset().signed(r.delay).str("s").draw(screen, 48, 30, 1, tc)
	for i, label := range [...]string{"STOP", "TIME", "RIDE", "SPEED", "HORN"} {
		y := 16 + i*9
		ebitenutil.DebugPrintAt(screen, label, 12, y)
		pts := [...]int{r.stopPts, r.timePts, r.ridePts, r.speedPts, r.hornPts}[i]
		if i > 2 {
			buf.reset().signed(pts)
		} else {
			buf.reset().num(pts)
		}
		buf.draw(screen, -100, y+5, 1, textWhite)
	}
	fillRect(screen, colorPanelHi, 12, 64, 90, 1)
	ebitenutil.DebugPrintAt(screen, "TOTAL", 12, 63)
	buf.reset().num(r.total).draw(screen, -100, 67, 2, textYellow)
}

func box(screen *ebiten.Image, x, y, w, h int) {
	fillRect(screen, colorPanelHi, x-1, y-1, w+2, h+2)
	fillRect(screen, colorBox, x, y, w, h)
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
