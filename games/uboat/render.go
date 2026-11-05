package uboat

import (
	"image"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

func (g *Game) Draw(screen *ebiten.Image) {
	scope := g.up
	if g.moving > 0 {
		scope = true // the view shows while the periscope moves
	}
	if g.mode == modeTitle || g.mode == modeOver {
		scope = true
	}
	if scope {
		g.drawScope(screen)
		if g.moving > 0 {
			// The view closes from the top when lowering and opens when raising.
			shown := g.moving
			if !g.up {
				shown = moveTicks - g.moving
			}
			fillRect(screen, imageBlack, 0, 0, ScreenWidth, ScreenHeight-shown*ScreenHeight/moveTicks)
		}
	} else {
		g.drawSonar(screen)
	}
	g.drawHUD(screen)
}

// shakeOffset jolts the view after a nearby depth charge. It does not use the
// game's random generator, so drawing does not change the game.
func (g *Game) shakeOffset() (int, int) {
	if g.shake == 0 {
		return 0, 0
	}
	return g.ticks*7%5 - 2, g.ticks*3%5 - 2
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

func (g *Game) drawScope(screen *ebiten.Image) {
	if g.clipsOf != screen {
		// SubImage allocates, so the clips are cut once per screen image.
		for y := 1; y <= ScreenHeight; y++ {
			g.clips[y] = screen.SubImage(image.Rect(0, 0, ScreenWidth, y)).(*ebiten.Image)
		}
		g.clipsOf = screen
	}
	ox, oy := g.shakeOffset()
	fillRect(screen, imageSky, 0, 0, ScreenWidth, horizon+oy)
	fillRect(screen, imageSea, 0, horizon+oy, ScreenWidth, ScreenHeight-horizon-oy)
	// The sky and the sea slide with the bearing, the sea faster for depth.
	shift := int(g.bearing * g.pxPerRad())
	tile(screen, imageClouds, -shift/3+ox, 8+oy)
	tile(screen, imageWaves, -shift+ox, horizon+oy)

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
			y := g.waterline(hypot(g.rel(t.x, t.y))) + 1
			fillRect(screen, imageWhite, int(x)-1+ox, int(y)+oy, 3, 1)
		}
	}
	for _, c := range g.charges {
		if c.used && c.fuse <= 0 {
			// A water column rises over an exploded depth charge.
			if x, ok := g.project(bearingTo(g.rel(c.x, c.y))); ok {
				d := hypot(g.rel(c.x, c.y))
				scale := min(max(40/max(d, 1)*g.pxPerRad()/float32(imageSplash.Bounds().Dy()), 0.3), 6)
				drawScaled(screen, imageSplash, x+float32(ox), g.waterline(d)+float32(oy), scale)
			}
		}
	}
	for _, b := range g.blasts {
		if !b.alive {
			continue
		}
		if x, ok := g.project(bearingTo(g.rel(b.x, b.y))); ok && b.age%6 < 4 {
			d := hypot(g.rel(b.x, b.y))
			scale := min(max(30/max(d, 1)*g.pxPerRad()/float32(imageFire.Bounds().Dy()), 0.4), 6)
			drawScaled(screen, imageFire, x+float32(ox), g.waterline(d)+float32(oy), scale)
		}
	}
	if g.mode == modePlay {
		g.drawReticle(screen)
	}
	drawAt(screen, imageMask, 0, 0)
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
	img := imageShip[s.kind]
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
	clip := screen
	if row := int(water + 1); row >= 1 && row <= ScreenHeight {
		clip = g.clips[row]
	}
	clip.DrawImage(img, &op)
	if s.sink > 0 && s.sink < sinkTicks*2/3 {
		scale := min(max(height/float32(imageSmoke.Bounds().Dy()), 0.4), 5)
		rise := float32(s.sink%40) * height / 40
		drawScaled(clip, imageSmoke, x+ox, top-rise, scale)
	}
}

func (g *Game) drawReticle(screen *ebiten.Image) {
	// A vertical line with a gap at the horizon, and bearing ticks every 5
	// degrees, longer every 10.
	fillRect(screen, imageBlack, ScreenWidth/2, 14, 1, horizon-18)
	fillRect(screen, imageBlack, ScreenWidth/2, horizon+4, 1, 32)
	step := float32(5 * math.Pi / 180)
	first := float32(math.Ceil(float64((g.bearing-g.view()/2)/step))) * step
	for a := first; a < g.bearing+g.view()/2; a += step {
		x, _ := g.project(a)
		h := 2
		if int(math.Round(float64(a/step)))%2 == 0 {
			h = 4
		}
		fillRect(screen, imageBlack, int(x), horizon-h, 1, h)
	}
	if s := g.target(); s != nil {
		if brg, ok := g.lead(s); ok {
			if x, ok := g.project(brg); ok {
				// The aim mark: where a torpedo fired now meets the target.
				fillRect(screen, imageYellow, int(x)-1, horizon-9, 3, 1)
				fillRect(screen, imageYellow, int(x), horizon-8, 1, 2)
			}
		}
	}
}

func (g *Game) drawSonar(screen *ebiten.Image) {
	fillRect(screen, imageBlack, 0, 0, ScreenWidth, ScreenHeight)
	ox, oy := g.shakeOffset()
	cx, cy := ScreenWidth/2+ox, ScreenHeight/2+oy
	drawAt(screen, imageSonar, cx-sonarRadius, cy-sonarRadius)
	scale := float32(sonarRadius) / sonarRange
	// The periscope's bearing as a dotted line.
	s, c := sincos(g.bearing)
	for r := float32(6); r < sonarRadius; r += 4 {
		fillRect(screen, imageDim, cx+int(s*r), cy-int(c*r), 1, 1)
	}
	for _, sh := range g.ships {
		x, y := g.rel(sh.x, sh.y)
		if !sh.alive || sh.sink > 0 || hypot(x, y) > sonarRange {
			continue
		}
		if sh.kind == destroyer && g.alert && g.ticks%20 < 6 {
			continue // an attacking destroyer blinks
		}
		img := imageContact[sh.kind]
		b := img.Bounds()
		drawAt(screen, img, cx+int(x*scale)-b.Dx()/2, cy+int(y*scale)-b.Dy()/2)
	}
	for _, t := range g.torps {
		if t.alive {
			x, y := g.rel(t.x, t.y)
			fillRect(screen, imageWhite, cx+int(x*scale), cy+int(y*scale), 1, 1)
		}
	}
	for _, ch := range g.charges {
		if ch.used && ch.fuse <= 0 && ch.fuse > -12 {
			r := 1 - ch.fuse/4
			x, y := g.rel(ch.x, ch.y)
			fillRect(screen, imageYellow, cx+int(x*scale)-r, cy+int(y*scale)-r, 2*r+1, 2*r+1)
		}
	}
	// The submarine with a line toward its heading, and the throttle.
	fillRect(screen, imageGreen, cx-1, cy-1, 3, 3)
	hs, hc := sincos(g.subHeading)
	for r := float32(3); r <= 7; r++ {
		fillRect(screen, imageGreen, cx+int(hs*r), cy-int(hc*r), 1, 1)
	}
	name := throttleNames[g.throttle]
	ebitenutil.DebugPrintAt(screen, name, ScreenWidth/2-len(name)*3, 100)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	switch g.mode {
	case modeTitle:
		ebitenutil.DebugPrintAt(screen, "IKATEN U-BOAT", 21, 32)
		if g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 70)
		}
		ebitenutil.DebugPrintAt(screen, "<> SCOPE  ^v ZOOM", 9, 96)
		ebitenutil.DebugPrintAt(screen, "Z FIRE  X DIVE", 18, 106)
		return
	case modeOver:
		ebitenutil.DebugPrintAt(screen, "YOUR BOAT IS LOST", 9, 30)
		g.drawTonnage(screen, 50)
		if g.modeTicks > 60 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 80)
		}
		return
	}
	// The corners outside the circle hold the gauges.
	// The periscope's bearing, or the submarine's heading with it down.
	heading := g.bearing
	if !g.up && g.moving == 0 {
		heading = g.subHeading
	}
	drawNumber(screen, int(math.Round(float64(heading*180/math.Pi)+360))%360, 3, 2, 2)
	for i := range g.torpedoes {
		fillRect(screen, imageWhite, ScreenWidth-3-i*3, 2, 2, 5)
	}
	if g.reload > 0 {
		fillRect(screen, imageGray, ScreenWidth-30, 9, 28*g.reload/reloadTicks, 1)
	}
	// Hull at the bottom left, detection at the bottom right.
	fillRect(screen, imageGray, 2, ScreenHeight-4, 24, 2)
	fillRect(screen, imageGreen, 2, ScreenHeight-4, int(g.hull)*24/100, 2)
	fillRect(screen, imageGray, ScreenWidth-26, ScreenHeight-4, 24, 2)
	if !g.alert || g.ticks%20 < 14 {
		fillRect(screen, imageRed, ScreenWidth-26, ScreenHeight-4, int(g.detect)*24/100, 2)
	}
	if g.zoomed && g.up {
		ebitenutil.DebugPrintAt(screen, "x3", 2, 6)
	}
	if g.messageTicks > 0 && g.message != msgNone && g.messageTicks%30 < 22 {
		msg := messages[g.message]
		ebitenutil.DebugPrintAt(screen, msg, ScreenWidth/2-len(msg)*3, 14)
	}
	if g.mode == modeResult {
		ebitenutil.DebugPrintAt(screen, "MISSION OVER", 24, 26)
		g.drawTonnage(screen, 40)
		if g.modeTicks > 60 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 76)
		}
	}
}

func (g *Game) drawTonnage(screen *ebiten.Image, y int) {
	ebitenutil.DebugPrintAt(screen, "TONS", 24, y-5)
	drawNumber(screen, g.tonnage, 6, 54, y)
	ebitenutil.DebugPrintAt(screen, "SHIPS", 24, y+5)
	drawNumber(screen, g.sunk, 2, 54, y+10)
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

// drawNumber draws n with the given number of digits, padded with zeros.
func drawNumber(screen *ebiten.Image, n, digits, x, y int) {
	for i := digits - 1; i >= 0; i-- {
		drawAt(screen, imageDigits[n%10], x+i*4, y)
		n /= 10
	}
}
