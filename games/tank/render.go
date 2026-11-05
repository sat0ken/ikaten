package tank

import (
	"image"
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/games/tank/field"
)

const (
	maxItems  = 128
	radarSize = 31
	radarX    = 2
	radarY    = 87
	radarSpan = 800 // meters from the radar's center to its rim
)

// camera is the gunsight: at the player's tank, looking along the turret.
type camera struct {
	x, y           float32
	fx, fy, rx, ry float32 // forward and right
	f              float32 // pixels per radian
	cx, hz         float32 // the axis's screen x and the horizon's screen y
}

type drawItem struct {
	z, sx, sy, scale float32
	what             uint8
	index            int
}

// What a draw item is.
const (
	itemObstacle = iota
	itemTank
	itemPuff
	itemShell
)

func (g *Game) Draw(screen *ebiten.Image) {
	g.setCamera()
	g.drawSky(screen)
	g.drawGround(screen)
	g.drawWorld(screen)
	g.drawSight(screen)
	g.drawHUD(screen)
}

func (g *Game) setCamera() {
	c := &g.cam
	c.x, c.y = g.x, g.y
	s, co := sincos(g.turret)
	c.fx, c.fy, c.rx, c.ry = s, -co, co, s
	c.f = focalWide
	if g.zoomed {
		c.f = focalZoom
	}
	c.cx = centerX
	// Raising the gun lowers the view; firing kicks it up.
	c.hz = centerY + c.f*g.elev + float32(g.recoil)*0.6
	if g.shake > 0 {
		amp := float32(min(g.shake, 6)) / 3
		c.cx += amp * float32(g.ticks*7%5-2)
		c.hz += amp * float32(g.ticks*3%5-2)
	}
}

// project returns where a point at height h lands on screen and its depth.
func (g *Game) project(x, y, h float32) (sx, sy, z float32, ok bool) {
	c := &g.cam
	dx, dy := x-c.x, y-c.y
	z = dx*c.fx + dy*c.fy
	if z < 2 || z > farthest {
		return 0, 0, 0, false
	}
	side := dx*c.rx + dy*c.ry
	return c.cx + side*c.f/z, c.hz + (eyeHeight-h)*c.f/z, z, true
}

func (g *Game) drawSky(screen *ebiten.Image) {
	hz := int(math.Floor(float64(g.cam.hz)))
	if hz <= 0 {
		return
	}
	fillRect(screen, imageSky, 0, 0, ScreenWidth, hz-12)
	fillRect(screen, imageSkyLow, 0, max(hz-12, 0), ScreenWidth, min(hz, 12))
	// The hills turn with the turret, a whole number of strips per turn so
	// that they meet at the back.
	width := field.HillsW
	strips := float32(math.Round(float64(g.cam.f) * 2 * math.Pi / float64(width)))
	shift := int(-g.turret/(2*math.Pi)*strips*float32(width)) + int(g.cam.cx) - centerX
	shift = (shift%width + width) % width
	for x := shift - width; x < ScreenWidth; x += width {
		drawAt(screen, imageHills, x, hz-field.HillsH)
	}
}

// drawGround draws each row of the ground as a rotated and scaled row of the
// field's texture: the farther the row, the more ground it covers.
func (g *Game) drawGround(screen *ebiten.Image) {
	if g.rowsOf != screen {
		// SubImage allocates, so the rows are cut once per screen image.
		for y := range ScreenHeight {
			g.rows[y] = screen.SubImage(image.Rect(0, y, ScreenWidth, y+1)).(*ebiten.Image)
		}
		g.rowsOf = screen
	}
	c := &g.cam
	top := max(int(math.Floor(float64(c.hz)))+1, 0)
	if top >= ScreenHeight {
		return
	}
	fillRect(screen, imageGrass, 0, top, ScreenWidth, ScreenHeight-top)
	// In texture pixels.
	const m = field.TexelMeters
	camX, camY := (c.x+field.Half)/m, (c.y+field.Half)/m
	height := float32(eyeHeight) / m
	var op ebiten.DrawImageOptions
	for y := top; y < ScreenHeight; y++ {
		below := float32(y) + 0.5 - c.hz
		if below <= 0 {
			continue
		}
		z := height * c.f / below
		if z > farthest/m {
			continue
		}
		// One screen pixel covers unit texture pixels. The row starts at the
		// screen's left edge, z ahead of the camera.
		unit := z / c.f
		left := -c.cx
		px := camX + c.fx*z + c.rx*unit*left
		py := camY + c.fy*z + c.ry*unit*left
		a, b := c.rx/unit, c.ry/unit
		cc, d := c.fx/unit, c.fy/unit
		op.GeoM.SetElement(0, 0, float64(a))
		op.GeoM.SetElement(0, 1, float64(b))
		op.GeoM.SetElement(0, 2, float64(-(a*px + b*py)))
		op.GeoM.SetElement(1, 0, float64(cc))
		op.GeoM.SetElement(1, 1, float64(d))
		op.GeoM.SetElement(1, 2, float64(float32(y)-(cc*px+d*py)))
		g.rows[y].DrawImage(imageGround, &op)
	}
	if top > 0 {
		fillRect(screen, imageHaze, 0, top-1, ScreenWidth, 2)
	}
}

func (g *Game) addItem(n *int, it drawItem) {
	if *n < len(g.items) {
		g.items[*n] = it
		*n++
	}
}

// visible reports whether something halfWidth meters across from its center
// at sx may show on screen.
func (g *Game) visible(sx, z, halfWidth float32) bool {
	w := halfWidth * g.cam.f / z
	return sx+w >= 0 && sx-w < ScreenWidth
}

func (g *Game) drawWorld(screen *ebiten.Image) {
	n := 0
	for i := range field.Obstacles {
		o := &field.Obstacles[i]
		if sx, sy, z, ok := g.project(o.X, o.Y, 0); ok && g.visible(sx, z, 7) {
			g.addItem(&n, drawItem{z: z, sx: sx, sy: sy, what: itemObstacle, index: i})
		}
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		if !e.alive && e.dead == 0 {
			continue
		}
		if sx, sy, z, ok := g.project(e.x, e.y, 0); ok && g.visible(sx, z, 5) {
			g.addItem(&n, drawItem{z: z, sx: sx, sy: sy, what: itemTank, index: i})
		}
	}
	for i := range g.puffs {
		p := &g.puffs[i]
		if !p.alive {
			continue
		}
		if sx, sy, z, ok := g.project(p.x, p.y, p.h); ok && g.visible(sx, z, 6) {
			g.addItem(&n, drawItem{z: z, sx: sx, sy: sy, what: itemPuff, index: i})
		}
	}
	for i := range g.shells {
		s := &g.shells[i]
		if !s.alive {
			continue
		}
		if sx, sy, z, ok := g.project(s.x, s.y, s.h); ok && g.visible(sx, z, 1) {
			g.addItem(&n, drawItem{z: z, sx: sx, sy: sy, what: itemShell, index: i})
		}
	}
	items := g.items[:n]
	// Far to near, so near things cover far ones.
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].z > items[j-1].z; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	g.ranging = 0
	for _, it := range items {
		scale := g.cam.f / (it.z * field.PixelsPerMeter)
		switch it.what {
		case itemObstacle:
			g.drawObstacle(screen, &field.Obstacles[it.index], it, scale)
		case itemTank:
			g.drawTank(screen, it, scale)
		case itemPuff:
			p := &g.puffs[it.index]
			ps := puffSpecs[p.kind]
			size := ps.size * (1 + ps.grow*float32(p.age)/float32(p.life))
			s := size * g.cam.f / (it.z * puffSize)
			drawScaled(screen, imagePuff[p.kind], it.sx, it.sy, s, puffSize/2, puffSize/2, 0)
		case itemShell:
			size := int(min(max(0.25*g.cam.f/it.z, 1), 4))
			fillRect(screen, imageTracer, int(it.sx)-size/2, int(it.sy)-size/2, size, size)
		}
	}
	if !g.zoomed {
		// Mark the enemies in sight, which are hard to see in the wide view.
		for i := range g.enemies {
			e := &g.enemies[i]
			if !e.alive || !e.sees {
				continue
			}
			if sx, sy, _, ok := g.project(e.x, e.y, specs[e.kind].height+2); ok {
				drawAt(screen, imageMarker, int(sx)-2, int(sy)-4)
			}
		}
	}
}

// frameOf returns the sprite frame, of n around a turn, for something facing
// heading seen along bearing.
func frameOf(heading, bearing float32, n int) int {
	rel := angleDiff(heading, bearing)
	f := int(math.Round(float64(rel) * float64(n) / (2 * math.Pi)))
	return (f%n + n) % n
}

func (g *Game) drawObstacle(screen *ebiten.Image, o *field.Obstacle, it drawItem, scale float32) {
	switch o.Kind {
	case field.House:
		// Houses look the same turned half around, so 8 frames cover half a turn.
		f := frameOf(o.Heading, bearingTo(o.X-g.cam.x, o.Y-g.cam.y), 2*field.HouseFrames) % field.HouseFrames
		drawScaled(screen, imageHouse[f], it.sx, it.sy, scale, field.HouseAX, field.HouseAY, 0)
	default:
		drawScaled(screen, imageTree[o.Kind], it.sx, it.sy, scale, field.TreeAX, field.TreeAY, 0)
	}
}

func (g *Game) drawTank(screen *ebiten.Image, it drawItem, scale float32) {
	e := &g.enemies[it.index]
	s := &specs[e.kind]
	if scale*field.TankW < 4 {
		// Too far for the sprite: a dot, or two for a big one.
		fillRect(screen, imageDot[e.kind], int(it.sx), int(it.sy)-1, 1, 1)
		if scale*field.TankW >= 2 {
			fillRect(screen, imageDot[e.kind], int(it.sx)-1, int(it.sy)-1, 2, 1)
		}
	} else {
		along := bearingTo(e.x-g.cam.x, e.y-g.cam.y)
		dark := float32(0)
		if !e.alive {
			dark = 0.35
		}
		drawScaled(screen, imageTank[e.kind][0][frameOf(e.heading, along, field.TankFrames)], it.sx, it.sy, scale, field.TankAX, field.TankAY, dark)
		drawScaled(screen, imageTank[e.kind][1][frameOf(e.turret, along, field.TankFrames)], it.sx, it.sy, scale, field.TankAX, field.TankAY, dark)
	}
	if !e.alive || !g.zoomed {
		return
	}
	// The range finder reads the nearest live tank in the column under the
	// reticle, where the range marks are.
	half := s.length / 2 * g.cam.f / it.z
	top := it.sy - s.height*g.cam.f/it.z
	if centerX >= it.sx-half && centerX <= it.sx+half && it.sy >= centerY && top <= centerY+markSpan {
		g.ranging = int(it.z+5) / 10 * 10
	}
}

// drawScaled draws img scaled with its anchor ax, ay at x, y, darkened when
// dark is above zero.
func drawScaled(screen, img *ebiten.Image, x, y, scale, ax, ay, dark float32) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(math.Round(float64(x-ax*scale)), math.Round(float64(y-ay*scale)))
	if dark > 0 {
		op.ColorScale.Scale(dark, dark*0.95, dark*0.9, 1)
	}
	screen.DrawImage(img, &op)
}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
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

// rangeMarks are the sight's marks, in hundreds of meters.
var rangeMarks = [...]int{2, 4, 6, 8, 10}

// markSpan is how far below the reticle's center the marks reach, in pixels.
const markSpan = 34

func (g *Game) drawSight(screen *ebiten.Image) {
	cx, cy := centerX, centerY
	if !g.zoomed {
		for _, d := range [...]int{-6, 3} {
			fillRect(screen, imageCross, cx+d, cy, 4, 1)
			fillRect(screen, imageCross, cx, cy+d, 1, 4)
		}
		return
	}
	drawAt(screen, imageSight, 0, 0)
	// The aiming chevron, with the horizontal line and lead ticks beside it.
	fillRect(screen, imageReticle, cx, cy, 1, 1)
	fillRect(screen, imageReticle, cx-1, cy+1, 3, 1)
	fillRect(screen, imageReticle, cx-2, cy+2, 1, 1)
	fillRect(screen, imageReticle, cx+2, cy+2, 1, 1)
	fillRect(screen, imageReticle, cx-34, cy, 30, 1)
	fillRect(screen, imageReticle, cx+5, cy, 30, 1)
	for _, d := range [...]int{-24, -14, 14, 24} {
		fillRect(screen, imageReticle, cx+d, cy-2, 1, 2)
	}
	// Range marks: put the target between the pair of ticks for its range.
	for _, r := range rangeMarks {
		y := cy + int(math.Round(float64(dropAngle(float32(r)*100)*focalZoom)))
		fillRect(screen, imageReticle, cx-6, y, 3, 1)
		fillRect(screen, imageReticle, cx+4, y, 3, 1)
		drawNumber(screen, colorReticle, r, 1+r/10, cx+9, y-2, 1)
	}
	if g.ranging > 0 {
		drawNumber(screen, colorCross, g.ranging, 3+g.ranging/1000, cx+20, cy-9, 1)
	}
}

var enemyLabels = [maxEnemies + 1]string{"ENEMY 0", "ENEMY 1", "ENEMY 2", "ENEMY 3", "ENEMY 4", "ENEMY 5", "ENEMY 6", "ENEMY 7", "ENEMY 8"}

func (g *Game) drawHUD(screen *ebiten.Image) {
	if g.mode == modeTitle {
		g.drawTitle(screen)
		return
	}
	// Armor left, top left.
	ebitenutil.DebugPrintAt(screen, "HP", 1, -3)
	bar := imageGreen
	switch {
	case g.hp <= hitDamage:
		bar = imageRed
	case g.hp <= 2*hitDamage:
		bar = imageYellow
	}
	fillRect(screen, imageShade, 14, 2, 32, 5)
	fillRect(screen, bar, 15, 3, int(g.hp*30/100), 3)
	ebitenutil.DebugPrintAt(screen, enemyLabels[g.enemiesLeft()], 77, -3)
	if g.messageTicks > 0 && g.message != msgNone {
		msg := messages[g.message]
		ebitenutil.DebugPrintAt(screen, msg, ScreenWidth/2-len(msg)*3, 8)
	}
	g.drawRadar(screen)
	// Throttle, shells left and the reload, bottom right.
	ebitenutil.DebugPrintAt(screen, throttleNames[g.throttle], 92, 92)
	ebitenutil.DebugPrintAt(screen, "AP", 92, 103)
	drawNumber(screen, colorCross, g.ammo, 2, 106, 109, 1)
	load := imageGreen
	if g.reload > 0 {
		load = imageYellow
	}
	fillRect(screen, imageShade, 92, 116, 26, 3)
	fillRect(screen, load, 93, 117, 24*(reloadTicks-g.reload)/reloadTicks, 1)
	if g.hurt > 0 && g.hurt%8 < 4 {
		fillRect(screen, imageRed, 0, 0, ScreenWidth, 2)
		fillRect(screen, imageRed, 0, ScreenHeight-2, ScreenWidth, 2)
		fillRect(screen, imageRed, 0, 0, 2, ScreenHeight)
		fillRect(screen, imageRed, ScreenWidth-2, 0, 2, ScreenHeight)
	}
	switch g.mode {
	case modeClear:
		g.drawBanner(screen, "AREA CLEAR!")
	case modeOver:
		g.drawBanner(screen, "TANK DESTROYED")
	}
}

func (g *Game) drawBanner(screen *ebiten.Image, title string) {
	fillRect(screen, imageShade, 10, 30, ScreenWidth-20, 40)
	ebitenutil.DebugPrintAt(screen, title, ScreenWidth/2-len(title)*3, 30)
	ebitenutil.DebugPrintAt(screen, "KILLS", 34, 42)
	drawNumber(screen, colorCross, g.kills, 2, 72, 48, 1)
	if g.modeTicks > 90 && g.modeTicks/30%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 54)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	fillRect(screen, imageShade, 0, 0, ScreenWidth, 12)
	ebitenutil.DebugPrintAt(screen, "IKATEN TANK", 27, -2)
	if g.modeTicks/30%2 == 0 {
		ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 20)
	}
	fillRect(screen, imageShade, 0, 84, ScreenWidth, 36)
	ebitenutil.DebugPrintAt(screen, "STICK:AIM  A:FIRE", 9, 82)
	ebitenutil.DebugPrintAt(screen, "B+STICK:DRIVE", 21, 93)
	ebitenutil.DebugPrintAt(screen, "TAP B:ZOOM", 30, 104)
}

func (g *Game) drawRadar(screen *ebiten.Image) {
	drawAt(screen, imageRadar, radarX, radarY)
	cx, cy := float32(radarX+radarSize/2), float32(radarY+radarSize/2)
	// The view's edges on the rim.
	half := float32(math.Atan(float64(centerX / g.cam.f)))
	for _, a := range [...]float32{-half, half} {
		x, y := g.radarDot(g.turret+a, radarSpan)
		fillRect(screen, imageDim, x, y, 1, 1)
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		dx, dy := e.x-g.x, e.y-g.y
		switch {
		case !e.alive && e.dead > 0:
			x, y := g.radarDot(bearingTo(dx, dy), hypot(dx, dy))
			fillRect(screen, imageGray, x, y, 1, 1)
		case e.alive && g.ticks-e.lastSeen < 120:
			x, y := g.radarDot(bearingTo(dx, dy), hypot(dx, dy))
			size := 2
			if e.flash > 0 {
				size = 3
			}
			fillRect(screen, imageRed, x-size/2, y-size/2, size, size)
		}
	}
	// The hull turns under the turret, which always points up.
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-1.5, -2.5)
	op.GeoM.Rotate(float64(angleDiff(g.heading, g.turret)))
	op.GeoM.Translate(float64(cx)+0.5, float64(cy)+0.5)
	screen.DrawImage(imageHull, &op)
	fillRect(screen, imageGreen, int(cx), int(cy)-5, 1, 5)
	if g.hitTicks > 0 && g.hitTicks%10 < 6 {
		x, y := g.radarDot(g.hitBearing, radarSpan)
		fillRect(screen, imageYellow, x-1, y-1, 3, 3)
	}
}

// radarDot returns where a bearing and distance from the player show on the
// radar, which turns with the turret.
func (g *Game) radarDot(bearing, dist float32) (int, int) {
	const r = radarSize/2 - 1
	s, c := sincos(bearing - g.turret)
	d := min(dist*r/radarSpan, r)
	return int(radarX + radarSize/2 + s*d + 0.5), int(radarY + radarSize/2 - c*d + 0.5)
}

// drawNumber draws n with the given number of digits, padded with zeros, at
// scale times the 3x5 font size, in a color.
func drawNumber(screen *ebiten.Image, clr color.RGBA, n, digits, x, y, scale int) {
	for i := digits - 1; i >= 0; i-- {
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x+i*4*scale), float64(y))
		op.ColorScale.Scale(float32(clr.R)/255, float32(clr.G)/255, float32(clr.B)/255, 1)
		screen.DrawImage(imageDigits[n%10], &op)
		n /= 10
	}
}
