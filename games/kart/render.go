package kart

import (
	"image"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawSky(screen)
	g.drawGround(screen)
	g.drawSprites(screen)
	g.drawHUD(screen)
}

func (g *Game) drawSky(screen *ebiten.Image) {
	fillRect(screen, imageSky[0], 0, 0, ScreenWidth, horizon-12)
	fillRect(screen, imageSky[1], 0, horizon-12, ScreenWidth, 12)
	// The mountains turn with the camera, half a strip per turn.
	width := imageMountains.Bounds().Dx()
	shift := int(-g.camAngle / (2 * math.Pi) * float32(width) * 2)
	shift = (shift%width + width) % width
	for x := shift - width; x < ScreenWidth; x += width {
		drawAt(screen, imageMountains, x, horizon-imageMountains.Bounds().Dy())
	}
}

// drawGround draws each row of the floor as a rotated and scaled row of the
// track: the farther the row, the more track it covers.
func (g *Game) drawGround(screen *ebiten.Image) {
	if g.rowsOf != screen {
		// SubImage allocates, so the rows are cut once per screen image.
		for y := horizon; y < ScreenHeight; y++ {
			g.rows[y] = screen.SubImage(image.Rect(0, y, ScreenWidth, y+1)).(*ebiten.Image)
		}
		g.rowsOf = screen
	}
	fillRect(screen, imageGrass, 0, horizon, ScreenWidth, ScreenHeight-horizon)
	var op ebiten.DrawImageOptions
	for y := horizon; y < ScreenHeight; y++ {
		z := camHeight * focal / (float32(y-horizon) + 0.5)
		if z > farthest {
			continue
		}
		// One screen pixel covers unit world units. The row starts at the
		// left edge of the view, z ahead of the camera.
		unit := z / focal
		left := float32(-ScreenWidth / 2)
		px := g.camX + g.fx*z + g.rx*unit*left
		py := g.camY + g.fy*z + g.ry*unit*left
		// Track coordinates map to screen x along the right vector and to
		// screen y along the forward vector; only the row's y is drawn.
		a, b := g.rx/unit, g.ry/unit
		c, d := g.fx/unit, g.fy/unit
		op.GeoM.SetElement(0, 0, float64(a))
		op.GeoM.SetElement(0, 1, float64(b))
		op.GeoM.SetElement(0, 2, float64(-(a*px + b*py)))
		op.GeoM.SetElement(1, 0, float64(c))
		op.GeoM.SetElement(1, 1, float64(d))
		op.GeoM.SetElement(1, 2, float64(float32(y)-(c*px+d*py)))
		g.rows[y].DrawImage(imageTrack, &op)
	}
	fillRect(screen, imageHaze, 0, horizon, ScreenWidth, 2)
}

type drawItem struct {
	z     float32
	x, y  float32 // screen position of the bottom center
	kart  int
	scale float32
}

// project returns where a world point lands on screen and its depth.
func (g *Game) project(x, y float32) (sx, sy, z float32, ok bool) {
	dx, dy := x-g.camX, y-g.camY
	z = dx*g.fx + dy*g.fy
	if z < 3 || z > farthest {
		return 0, 0, 0, false
	}
	side := dx*g.rx + dy*g.ry
	return ScreenWidth/2 + side*focal/z, horizon + camHeight*focal/z, z, true
}

func (g *Game) drawSprites(screen *ebiten.Image) {
	items := g.sprites[:0]
	for i := range g.karts {
		k := &g.karts[i]
		if x, y, z, ok := g.project(k.x, k.y); ok {
			scale := camDist / z
			if i == 0 {
				scale = 1 // the player's kart keeps its pixel art crisp
			}
			items = append(items, drawItem{z: z, x: x, y: y, kart: i, scale: scale})
		}
	}
	// Far to near, so near sprites cover far ones.
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].z > items[j-1].z; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	for _, it := range items {
		g.drawKart(screen, it)
	}
}

func (g *Game) drawKart(screen *ebiten.Image, it drawItem) {
	k := &g.karts[it.kart]
	frame := k.steer + 1
	if k.drift != 0 {
		frame = k.drift + 1
	}
	img := imageKart[it.kart][frame]
	drawScaled(screen, img, it.x, it.y, it.scale)
	if it.kart != 0 {
		return
	}
	w := float32(img.Bounds().Dx())
	if k.boost > 0 && g.ticks%4 < 3 {
		drawScaled(screen, imageFlame, it.x, it.y+3, 1)
	}
	if k.drift != 0 {
		spark := imageSpark[0]
		switch {
		case k.charge >= driftLong:
			spark = imageSpark[2]
		case k.charge >= driftShort:
			spark = imageSpark[1]
		}
		jump := float32(g.ticks % 2)
		drawScaled(screen, spark, it.x-w/2+2, it.y-jump, 1)
		drawScaled(screen, spark, it.x+w/2-2, it.y-1+jump, 1)
	}
}

// drawScaled draws img scaled, with its bottom center at x, y.
func drawScaled(screen, img *ebiten.Image, x, y, scale float32) {
	b := img.Bounds()
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(math.Round(float64(x-float32(b.Dx())*scale/2)), math.Round(float64(y-float32(b.Dy())*scale)))
	screen.DrawImage(img, &op)
}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

// fillRect draws a 1x1 image scaled to a rectangle.
func fillRect(screen, pixel *ebiten.Image, x, y, width, height int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(width), float64(height))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(pixel, &op)
}

// drawNumber draws n with the given number of digits, padded with zeros, at
// scale times the 3x5 font size.
func drawNumber(screen *ebiten.Image, n, digits, x, y, scale int) {
	for i := digits - 1; i >= 0; i-- {
		drawGlyph(screen, n%10, x+i*4*scale, y, scale)
		n /= 10
	}
}

func drawGlyph(screen *ebiten.Image, glyph, x, y, scale int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(imageDigits[glyph], &op)
}

// drawTime draws ticks as m:ss.cc.
func drawTime(screen *ebiten.Image, ticks, x, y int) {
	cs := ticks * 100 / 60
	drawNumber(screen, cs/6000%10, 1, x, y, 1)
	drawGlyph(screen, glyphColon, x+3, y, 1)
	drawNumber(screen, cs/100%60, 2, x+6, y, 1)
	drawGlyph(screen, glyphDot, x+13, y, 1)
	drawNumber(screen, cs%100, 2, x+16, y, 1)
}

var (
	ordinals  = [karts]string{"1ST", "2ND", "3RD", "4TH"}
	lapLabels = [laps]string{"LAP 1/3", "LAP 2/3", "LAP 3/3"}
)

func (g *Game) drawHUD(screen *ebiten.Image) {
	switch g.mode {
	case modeTitle:
		ebitenutil.DebugPrintAt(screen, "IKATEN KART", 27, 2)
		if g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 40)
		}
		ebitenutil.DebugPrintAt(screen, "Z:GAS X:DRIFT", 21, 106)
		return
	case modeCountdown:
		if n := 3 - g.modeTicks/60; n > 0 {
			drawNumber(screen, n, 1, 54, 46, 4)
		}
	}
	k := &g.karts[0]
	ebitenutil.DebugPrintAt(screen, lapLabels[min(max(k.lap+1, 1), laps)-1], 2, -2)
	drawTime(screen, g.raceTicks, 2, 10)
	ebitenutil.DebugPrintAt(screen, ordinals[g.position(0)-1], 98, -2)
	switch {
	case g.mode == modeRace && g.modeTicks < 50:
		ebitenutil.DebugPrintAt(screen, "GO!", 51, 40)
	case g.mode == modeResult:
		ebitenutil.DebugPrintAt(screen, "FINISH!", 39, 36)
		ebitenutil.DebugPrintAt(screen, ordinals[g.position(0)-1], 51, 48)
		ebitenutil.DebugPrintAt(screen, "BEST", 30, 62)
		drawTime(screen, g.bestLap, 60, 67)
		if g.modeTicks > 90 {
			ebitenutil.DebugPrintAt(screen, "PRESS Z", 39, 76)
		}
	}
}
