package kart

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

// trackScale is the track image's texels a world unit (see assets/gen.go).
const trackScale = 2

var (
	textWhite  = color.RGBA{250, 250, 250, 255}
	textYellow = color.RGBA{255, 230, 60, 255}
	textGray   = color.RGBA{200, 210, 230, 255}
	textShadow = color.RGBA{20, 24, 40, 255}
)

var text pixtext.Line

func (g *Game) Draw(screen *ebiten.Image) {
	mark(0)
	g.drawSky(screen)
	mark(1)
	g.drawGround(screen)
	mark(2)
	g.drawSprites(screen)
	mark(3)
	g.drawHUD(screen)
	mark(4)
}

func (g *Game) drawSky(screen *ebiten.Image) {
	// Bands lighter toward the horizon.
	sky := &g.art.sky
	band := (horizon - 12) / (len(sky) - 1)
	for i := range len(sky) - 1 {
		fillRect(screen, sky[i], 0, i*band, ScreenWidth, band)
	}
	fillRect(screen, sky[len(sky)-1], 0, (len(sky)-1)*band, ScreenWidth, horizon-(len(sky)-1)*band)
	// The mountains turn with the camera, half a strip per turn.
	m := g.art.mountains
	width := m.Bounds().Dx()
	shift := int(-g.camAngle / (2 * math.Pi) * float32(width) * 2)
	shift = (shift%width + width) % width
	for x := shift - width; x < ScreenWidth; x += width {
		drawAt(screen, m, x, horizon-m.Bounds().Dy())
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
	fillRect(screen, g.art.grass, 0, horizon, ScreenWidth, ScreenHeight-horizon)
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
		// screen y along the forward vector; only the row's y is drawn. The
		// track image has trackScale texels a world unit.
		a, b := g.rx/unit/trackScale, g.ry/unit/trackScale
		c, d := g.fx/unit/trackScale, g.fy/unit/trackScale
		px, py = px*trackScale, py*trackScale
		op.GeoM.SetElement(0, 0, float64(a))
		op.GeoM.SetElement(0, 1, float64(b))
		op.GeoM.SetElement(0, 2, float64(-(a*px + b*py)))
		op.GeoM.SetElement(1, 0, float64(c))
		op.GeoM.SetElement(1, 1, float64(d))
		op.GeoM.SetElement(1, 2, float64(float32(y)-(c*px+d*py)))
		g.rows[y].DrawImage(g.art.track, &op)
	}
	fillRect(screen, g.art.haze, 0, horizon, ScreenWidth, 3)
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
	img := g.art.kart[it.kart][frame]
	drawScaled(screen, img, it.x, it.y, it.scale)
	if it.kart != 0 {
		return
	}
	w := float32(img.Bounds().Dx())
	if k.boost > 0 && g.ticks%4 < 3 {
		drawScaled(screen, g.art.flame, it.x, it.y+6, 1)
	}
	if k.drift != 0 {
		spark := g.art.spark[0]
		switch {
		case k.charge >= driftLong:
			spark = g.art.spark[2]
		case k.charge >= driftShort:
			spark = g.art.spark[1]
		}
		jump := float32(g.ticks%2) * 2
		drawScaled(screen, spark, it.x-w/2+4, it.y-jump, 1)
		drawScaled(screen, spark, it.x+w/2-4, it.y-2+jump, 1)
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

// clock adds ticks as m:ss.cc to the line.
func clock(l *pixtext.Line, ticks int) *pixtext.Line {
	cs := ticks * 100 / 60
	return l.Num(cs/6000%10).Str(":").Pad(cs/100%60, 2).Str(".").Pad(cs%100, 2)
}

var (
	ordinals  = [karts]string{"1ST", "2ND", "3RD", "4TH"}
	lapLabels = [laps]string{"LAP 1/3", "LAP 2/3", "LAP 3/3"}
)

func (g *Game) drawHUD(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	switch g.mode {
	case modeTitle:
		text.Reset().Str("IKATEN KART").Shade(textShadow).Center(screen, mid, 10, 3, textYellow)
		if g.modeTicks/30%2 == 0 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 90, 2, textWhite)
		}
		text.Reset().Str("A: GAS   B: DRIFT").Shade(textShadow).Center(screen, mid, 222, 1, textWhite)
		return
	case modeCountdown:
		if n := 3 - g.modeTicks/60; n > 0 {
			text.Reset().Num(n).Shade(textShadow).Center(screen, mid, 84, 6, textYellow)
		}
	}
	k := &g.karts[0]
	text.Reset().Str(lapLabels[min(max(k.lap+1, 1), laps)-1]).Shade(textShadow).Draw(screen, 4, 4, 1, textWhite)
	clock(text.Reset(), g.raceTicks).Shade(textShadow).Draw(screen, 4, 16, 1, textWhite)
	text.Reset().Str(ordinals[g.position(0)-1]).Shade(textShadow).Right(screen, ScreenWidth-4, 4, 2, textYellow)
	switch {
	case g.mode == modeRace && g.modeTicks < 50:
		text.Reset().Str("GO!").Shade(textShadow).Center(screen, mid, 84, 4, textYellow)
	case g.mode == modeResult:
		text.Reset().Str("FINISH!").Shade(textShadow).Center(screen, mid, 70, 3, textWhite)
		text.Reset().Str(ordinals[g.position(0)-1]).Shade(textShadow).Center(screen, mid, 100, 3, textYellow)
		clock(text.Reset().Str("BEST "), g.bestLap).Shade(textShadow).Center(screen, mid, 132, 1, textWhite)
		if g.modeTicks > 90 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 152, 2, textGray)
		}
	}
}
