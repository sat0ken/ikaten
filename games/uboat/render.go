package uboat

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

var (
	textWhite  = color.RGBA{240, 240, 240, 255}
	textYellow = color.RGBA{255, 220, 40, 255}
	textGreen  = color.RGBA{60, 200, 110, 255}
	textGray   = color.RGBA{150, 156, 166, 255}
	textRed    = color.RGBA{240, 70, 50, 255}
	textShadow = color.RGBA{0, 0, 0, 255}
)

var text pixtext.Line

func (g *Game) Draw(screen *ebiten.Image) {
	if g.filled != screen {
		// Filling the whole screen marks it opaque, after which drawing on
		// it need not mark what it covers. Every frame covers all of it in
		// any case, so it is done once and the screen is not cleared.
		screen.Fill(color.RGBA{0, 0, 0, 255})
		g.filled = screen
	}
	scope := g.up
	if g.moving > 0 {
		scope = true // the view shows while the periscope moves
	}
	if g.mode == modeTitle || g.mode == modeOver {
		scope = true
	}
	mark(0)
	if scope {
		g.drawScope(screen)
		if g.moving > 0 {
			// The view closes from the top when lowering and opens when raising.
			shown := g.moving
			if !g.up {
				shown = moveTicks - g.moving
			}
			fillRect(screen, g.art.black, 0, 0, ScreenWidth, ScreenHeight-shown*ScreenHeight/moveTicks)
		}
	} else {
		g.drawSonar(screen)
		mark(1)
		mark(2)
	}
	mark(3)
	g.drawHUD(screen)
	mark(4)
}

// shakeOffset jolts the view after a nearby depth charge. It does not use the
// game's random generator, so drawing does not change the game.
func (g *Game) shakeOffset() (int, int) {
	if g.shake == 0 {
		return 0, 0
	}
	return (g.ticks*7%5 - 2) * 2, (g.ticks*3%5 - 2) * 2
}

// pxPerRad converts angles to screen pixels for the current magnification.
func (g *Game) pxPerRad() float32 { return ScreenWidth / g.view() }

// project returns the screen x of a bearing and whether it is in view.
func (g *Game) project(bearing float32) (float32, bool) {
	rel := angleDiff(bearing, g.bearing)
	return ScreenWidth/2 + rel*g.pxPerRad(), abs(rel) < g.view()/2+0.15
}

// waterline returns the screen row where something at dist meters meets the sea.
func (g *Game) waterline(dist float32) float32 {
	return horizon + eyeHeight*g.pxPerRad()/max(dist, 1)
}

// clip returns the screen above row y, for drawing a ship that stands in the
// sea at that row.
func (g *Game) clip(screen *ebiten.Image, y int) *ebiten.Image {
	if g.clipsOf != screen {
		// SubImage allocates, so the clips are cut once per screen image.
		for i := range g.clips {
			g.clips[i] = screen.SubImage(image.Rect(0, 0, ScreenWidth, clipTop+i)).(*ebiten.Image)
		}
		g.clipsOf = screen
	}
	if y < clipTop || y > ScreenHeight {
		return screen
	}
	return g.clips[y-clipTop]
}

func (g *Game) drawScope(screen *ebiten.Image) {
	a := g.art
	ox, oy := g.shakeOffset()
	fillRect(screen, a.sky, 0, 0, ScreenWidth, horizon+oy)
	fillRect(screen, a.sea, 0, horizon+oy, ScreenWidth, ScreenHeight-horizon-oy)
	// The sky and the sea slide with the bearing, the sea faster for depth.
	shift := int(g.bearing * g.pxPerRad())
	tile(screen, a.clouds, -shift/3+ox, 16+oy)
	tile(screen, a.waves, -shift+ox, horizon+oy)
	mark(1)

	// Ships far to near, so that near ones cover far ones.
	n := 0
	for i := range g.ships {
		if g.ships[i].alive {
			g.order[n] = i
			n++
		}
	}
	dist := func(i int) float32 { return hypot(g.rel(g.ships[i].x, g.ships[i].y)) }
	for i := 1; i < n; i++ {
		for j := i; j > 0 && dist(g.order[j]) > dist(g.order[j-1]); j-- {
			g.order[j], g.order[j-1] = g.order[j-1], g.order[j]
		}
	}
	for _, i := range g.order[:n] {
		g.drawShip(screen, &g.ships[i], float32(ox), float32(oy))
	}
	for _, t := range g.torps {
		if !t.alive {
			continue
		}
		if x, ok := g.project(bearingTo(g.rel(t.x, t.y))); ok {
			// The torpedo's wake: a white streak on the water.
			y := g.waterline(hypot(g.rel(t.x, t.y))) + 2
			fillRect(screen, a.white, int(x)-3+ox, int(y)+oy, 7, 1)
			fillRect(screen, a.white, int(x)-1+ox, int(y)+1+oy, 3, 1)
		}
	}
	for _, c := range g.charges {
		if c.used && c.fuse <= 0 {
			// A water column rises over an exploded depth charge.
			if x, ok := g.project(bearingTo(g.rel(c.x, c.y))); ok {
				d := hypot(g.rel(c.x, c.y))
				scale := min(max(40/max(d, 1)*g.pxPerRad()/float32(a.splash.Bounds().Dy()), 0.3), 6)
				drawScaled(screen, a.splash, x+float32(ox), g.waterline(d)+float32(oy), scale)
			}
		}
	}
	for _, b := range g.blasts {
		if !b.alive {
			continue
		}
		if x, ok := g.project(bearingTo(g.rel(b.x, b.y))); ok && b.age%6 < 4 {
			d := hypot(g.rel(b.x, b.y))
			scale := min(max(30/max(d, 1)*g.pxPerRad()/float32(a.fire.Bounds().Dy()), 0.4), 6)
			drawScaled(screen, a.fire, x+float32(ox), g.waterline(d)+float32(oy), scale)
		}
	}
	if g.mode == modePlay {
		g.drawReticle(screen)
	}
	mark(2)
	drawAt(screen, a.mask, 0, 0)
}

// tile draws a repeating strip shifted by x.
func tile(screen, img *ebiten.Image, x, y int) {
	w := img.Bounds().Dx()
	x %= w
	if x > 0 {
		x -= w
	}
	for ; x < ScreenWidth; x += w {
		drawAt(screen, img, x, y)
	}
}

// drawShip scales the side view by distance and shortens it by the angle
// between the ship's course and the line of sight. A sinking ship slides down
// behind its waterline.
func (g *Game) drawShip(screen *ebiten.Image, s *ship, ox, oy float32) {
	brg := bearingTo(g.rel(s.x, s.y))
	x, ok := g.project(brg)
	if !ok {
		return
	}
	d := hypot(g.rel(s.x, s.y))
	t := shipTypes[s.kind]
	aspect, _ := sincos(angleDiff(s.heading, brg))
	length := (t.length*abs(aspect) + beam*(1-abs(aspect))) / d * g.pxPerRad()
	height := t.height / d * g.pxPerRad()
	img := g.art.ship[s.kind]
	b := img.Bounds()
	sx, sy := length/float32(b.Dx()), height/float32(b.Dy())
	water := g.waterline(d) + oy
	top := water - height
	if s.sink > 0 {
		top += height * 1.3 * float32(s.sink) / sinkTicks
	}
	var op ebiten.DrawImageOptions
	left := x + ox - length/2
	if aspect < 0 {
		// Moving to the left on screen: flip so that the bow leads.
		op.GeoM.Scale(-float64(sx), float64(sy))
		left += length
	} else {
		op.GeoM.Scale(float64(sx), float64(sy))
	}
	op.GeoM.Translate(float64(left), float64(top))
	clip := g.clip(screen, int(water+1))
	clip.DrawImage(img, &op)
	if s.sink > 0 && s.sink < sinkTicks*2/3 {
		smoke := g.art.smoke
		scale := min(max(height/float32(smoke.Bounds().Dy()), 0.4), 5)
		rise := float32(s.sink%40) * height / 40
		drawScaled(clip, smoke, x+ox, top-rise, scale)
	}
}

func (g *Game) drawReticle(screen *ebiten.Image) {
	a := g.art
	// A vertical line with a gap at the horizon, and bearing ticks every 5
	// degrees, longer every 10, with a finer tick between when zoomed.
	fillRect(screen, a.black, ScreenWidth/2, 28, 1, horizon-36)
	fillRect(screen, a.black, ScreenWidth/2, horizon+8, 1, 64)
	fillRect(screen, a.black, ScreenWidth/2-6, horizon+40, 13, 1)
	step := float32(5 * math.Pi / 180)
	if g.zoomed {
		step /= 5
	}
	first := float32(math.Ceil(float64((g.bearing-g.view()/2)/step))) * step
	for t := first; t < g.bearing+g.view()/2; t += step {
		x, _ := g.project(t)
		h := 3
		if int(math.Round(float64(t/step)))%2 == 0 {
			h = 7
		}
		fillRect(screen, a.black, int(x), horizon-h, 1, h)
	}
	if s := g.target(); s != nil {
		if brg, ok := g.lead(s); ok {
			if x, ok := g.project(brg); ok {
				// The aim mark: where a torpedo fired now meets the target.
				fillRect(screen, a.yellow, int(x)-3, horizon-18, 7, 2)
				fillRect(screen, a.yellow, int(x)-1, horizon-16, 3, 2)
				fillRect(screen, a.yellow, int(x), horizon-14, 1, 3)
			}
		}
	}
}

func (g *Game) drawSonar(screen *ebiten.Image) {
	a := g.art
	fillRect(screen, a.black, 0, 0, ScreenWidth, ScreenHeight)
	ox, oy := g.shakeOffset()
	cx, cy := ScreenWidth/2+ox, ScreenHeight/2+oy
	drawAt(screen, a.sonar, cx-sonarRadius, cy-sonarRadius)
	scale := float32(sonarRadius) / sonarRange
	// The periscope's bearing as a dotted line.
	s, c := sincos(g.bearing)
	for r := float32(12); r < sonarRadius; r += 6 {
		fillRect(screen, a.dim, cx+int(s*r), cy-int(c*r), 2, 2)
	}
	for _, sh := range g.ships {
		x, y := g.rel(sh.x, sh.y)
		if !sh.alive || sh.sink > 0 || hypot(x, y) > sonarRange {
			continue
		}
		if sh.kind == destroyer && g.alert && g.ticks%20 < 6 {
			continue // an attacking destroyer blinks
		}
		img := a.contact[sh.kind]
		b := img.Bounds()
		drawAt(screen, img, cx+int(x*scale)-b.Dx()/2, cy+int(y*scale)-b.Dy()/2)
	}
	for _, t := range g.torps {
		if t.alive {
			x, y := g.rel(t.x, t.y)
			fillRect(screen, a.white, cx+int(x*scale), cy+int(y*scale), 2, 2)
		}
	}
	for _, ch := range g.charges {
		if ch.used && ch.fuse <= 0 && ch.fuse > -12 {
			r := 2 - ch.fuse/2
			x, y := g.rel(ch.x, ch.y)
			fillRect(screen, a.yellow, cx+int(x*scale)-r, cy+int(y*scale)-r, 2*r+1, 2*r+1)
		}
	}
	// The submarine with a line toward its heading, and the throttle.
	fillRect(screen, a.green, cx-2, cy-2, 5, 5)
	hs, hc := sincos(g.subHeading)
	for r := float32(5); r <= 14; r++ {
		fillRect(screen, a.green, cx+int(hs*r), cy-int(hc*r), 1, 1)
	}
	text.Reset().Str("SONAR").Draw(screen, 6, 26, 1, textGreen)
	text.Reset().Str(throttleNames[g.throttle]).Center(screen, ScreenWidth/2, 196, 2, textGreen)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	a := g.art
	switch g.mode {
	case modeTitle:
		text.Reset().Str("IKATEN").Shade(textShadow).Center(screen, mid, 44, 2, textWhite)
		text.Reset().Str("U-BOAT").Shade(textShadow).Center(screen, mid, 62, 4, textYellow)
		if g.modeTicks/30%2 == 0 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 150, 2, textWhite)
		}
		text.Reset().Str("LEFT/RIGHT: TURN  UP/DOWN: ZOOM").Shade(textShadow).Center(screen, mid, 214, 1, textWhite)
		text.Reset().Str("A: FIRE   B: DIVE").Shade(textShadow).Center(screen, mid, 226, 1, textWhite)
		return
	case modeOver:
		text.Reset().Str("YOUR BOAT").Shade(textShadow).Center(screen, mid, 50, 3, textRed)
		text.Reset().Str("IS LOST").Shade(textShadow).Center(screen, mid, 76, 3, textRed)
		g.drawTonnage(screen, 120)
		if g.modeTicks > 60 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 176, 2, textWhite)
		}
		return
	}
	// The corners outside the circle hold the gauges.
	// The periscope's bearing, or the submarine's heading with it down.
	heading := g.bearing
	if !g.up && g.moving == 0 {
		heading = g.subHeading
	}
	text.Reset().Pad(int(math.Round(float64(heading*180/math.Pi)+360))%360, 3).Draw(screen, 5, 5, 2, textWhite)
	if g.zoomed && g.up {
		text.Reset().Str("x3").Draw(screen, 5, 24, 2, textYellow)
	}
	// Torpedoes left at the top right, with the reload under them.
	for i := range torpedoes {
		img := a.white
		if i >= g.torpedoes {
			img = a.gray
		}
		fillRect(screen, img, ScreenWidth-8-i*6, 5, 4, 12)
		fillRect(screen, img, ScreenWidth-7-i*6, 4, 2, 1)
	}
	if g.reload > 0 {
		fillRect(screen, a.gray, ScreenWidth-62, 21, 58, 3)
		fillRect(screen, a.yellow, ScreenWidth-62, 21, 58-58*g.reload/reloadTicks, 3)
	}
	// Hull at the bottom left, detection at the bottom right.
	text.Reset().Str("HULL").Draw(screen, 4, ScreenHeight-20, 1, textGray)
	fillRect(screen, a.gray, 4, ScreenHeight-10, 48, 5)
	fillRect(screen, a.green, 4, ScreenHeight-10, int(g.hull)*48/100, 5)
	text.Reset().Str("DETECT").Right(screen, ScreenWidth-4, ScreenHeight-20, 1, textGray)
	fillRect(screen, a.gray, ScreenWidth-52, ScreenHeight-10, 48, 5)
	if !g.alert || g.ticks%20 < 14 {
		fillRect(screen, a.red, ScreenWidth-52, ScreenHeight-10, int(g.detect)*48/100, 5)
	}
	if g.messageTicks > 0 && g.message != msgNone && g.messageTicks%30 < 22 {
		c := textWhite
		if g.message == msgDetected || g.message == msgCharges {
			c = textRed
		}
		text.Reset().Str(messages[g.message]).Shade(textShadow).Center(screen, mid, 34, 2, c)
	}
	if g.mode == modeResult {
		text.Reset().Str("MISSION OVER").Shade(textShadow).Center(screen, mid, 56, 3, textYellow)
		g.drawTonnage(screen, 92)
		if g.modeTicks > 60 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 150, 2, textWhite)
		}
	}
}

func (g *Game) drawTonnage(screen *ebiten.Image, y int) {
	text.Reset().Str("TONS  ").Pad(g.tonnage, 6).Shade(textShadow).Center(screen, ScreenWidth/2, y, 2, textWhite)
	text.Reset().Str("SHIPS     ").Pad(g.sunk, 2).Shade(textShadow).Center(screen, ScreenWidth/2, y+20, 2, textWhite)
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
