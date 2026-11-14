package tank

import (
	"image/color"
	"math"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
	"github.com/sat0ken/ikaten/games/tank/field"
)

const (
	maxItems   = 128
	radarSize  = 62
	radarX     = 4
	radarY     = ScreenHeight - radarSize - 4
	radarSpan  = 800 // meters from the radar's center to its rim
	groundStep = 2   // screen rows a row of the ground covers
)

var text pixtext.Line

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
	if g.filled != screen {
		// Filling the whole screen marks it opaque, after which drawing on
		// it need not mark what it covers. Every frame covers all of it in
		// any case, so it is done once and the screen is not cleared.
		screen.Fill(colorSky)
		g.filled = screen
	}
	g.setCamera()
	mark(0)
	g.drawSky(screen)
	mark(1)
	g.drawGround(screen)
	mark(2)
	g.drawWorld(screen)
	mark(3)
	g.drawSight(screen)
	g.drawHUD(screen)
	mark(4)
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
	c.hz = centerY + c.f*g.elev + float32(g.recoil)*1.2
	if g.shake > 0 {
		amp := float32(min(g.shake, 6)) * 2 / 3
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
	a := g.art
	hz := int(math.Floor(float64(g.cam.hz)))
	top := hz - field.HillsH
	if top > 0 {
		fillRect(screen, a.sky, 0, 0, ScreenWidth, min(top, ScreenHeight))
	}
	if top >= ScreenHeight || hz <= 0 {
		return
	}
	// The hills, with the sky behind them, turn with the turret, a whole
	// number of strips per turn so that they meet at the back.
	width := field.HillsW
	strips := float32(math.Round(float64(g.cam.f) * 2 * math.Pi / float64(width)))
	shift := int(-g.turret/(2*math.Pi)*strips*float32(width)) + int(g.cam.cx) - centerX
	shift = (shift%width + width) % width
	for x := shift - width; x < ScreenWidth; x += width {
		drawAt(screen, a.hills, x, top)
	}
}

// drawGround draws the ground a band of rows at a time, each a rotated and
// scaled row of the field's texture: the farther the row, the more ground it
// covers. The row is drawn into a buffer and copied groundStep rows tall,
// which halves the texture reads, the slow part.
func (g *Game) drawGround(screen *ebiten.Image) {
	a := g.art
	c := &g.cam
	top := max(int(math.Floor(float64(c.hz)))+1, 0)
	if top >= ScreenHeight {
		return
	}
	// In texture pixels.
	const m = field.TexelMeters
	camX, camY := (c.x+field.Half)/m, (c.y+field.Half)/m
	height := float32(eyeHeight) / m
	row := a.row
	var op, copyOp ebiten.DrawImageOptions
	for y := top; y < ScreenHeight; y += groundStep {
		h := min(groundStep, ScreenHeight-y)
		row.Fill(color.RGBA{grassColor >> 16, grassColor >> 8 & 0xff, grassColor & 0xff, 255}) // also marks it opaque, so it is copied fast
		below := float32(y) + float32(h)/2 - c.hz
		if z := height * c.f / below; below > 0 && z <= farthest/m {
			// One screen pixel covers unit texture pixels. The row starts at
			// the screen's left edge, z ahead of the camera.
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
			op.GeoM.SetElement(1, 2, float64(-(cc*px + d*py)))
			row.DrawImage(g.art.ground, &op)
		}
		copyOp.GeoM.Reset()
		copyOp.GeoM.Scale(1, float64(h))
		copyOp.GeoM.Translate(0, float64(y))
		screen.DrawImage(row, &copyOp)
	}
	if top > 0 {
		fillRect(screen, a.haze, 0, top-1, ScreenWidth, 3)
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
			drawScaled(screen, g.art.puff[p.kind], it.sx, it.sy, s, puffSize/2, puffSize/2, 0)
		case itemShell:
			size := int(min(max(0.25*g.cam.f/it.z, 2), 8))
			fillRect(screen, g.art.tracer, int(it.sx)-size/2, int(it.sy)-size/2, size, size)
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
				drawAt(screen, g.art.marker, int(sx)-4, int(sy)-8)
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
		drawScaled(screen, g.art.house[f], it.sx, it.sy, scale, field.HouseAX, field.HouseAY, 0)
	default:
		drawScaled(screen, g.art.tree[o.Kind], it.sx, it.sy, scale, field.TreeAX, field.TreeAY, 0)
	}
}

func (g *Game) drawTank(screen *ebiten.Image, it drawItem, scale float32) {
	e := &g.enemies[it.index]
	s := &specs[e.kind]
	if scale*field.TankW < 8 {
		// Too far for the sprite: a dot, or a wider one for a big one.
		w := 2
		if scale*field.TankW >= 4 {
			w = 4
		}
		fillRect(screen, g.art.dot[e.kind], int(it.sx)-w/2, int(it.sy)-2, w, 2)
	} else {
		along := bearingTo(e.x-g.cam.x, e.y-g.cam.y)
		dark := float32(0)
		if !e.alive {
			dark = 0.35
		}
		drawScaled(screen, g.art.tank[e.kind][0][frameOf(e.heading, along, field.TankFrames)], it.sx, it.sy, scale, field.TankAX, field.TankAY, dark)
		drawScaled(screen, g.art.tank[e.kind][1][frameOf(e.turret, along, field.TankFrames)], it.sx, it.sy, scale, field.TankAX, field.TankAY, dark)
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
const markSpan = 68

func (g *Game) drawSight(screen *ebiten.Image) {
	a := g.art
	cx, cy := centerX, centerY
	if !g.zoomed {
		for _, d := range [...]int{-13, 6} {
			fillRect(screen, a.cross, cx+d, cy, 8, 1)
			fillRect(screen, a.cross, cx, cy+d, 1, 8)
		}
		return
	}
	for _, b := range a.sight {
		if b.piece == nil {
			break
		}
		h := b.y1 - b.y0
		fillRect(screen, a.black, 0, b.y0, b.black, h)
		fillRect(screen, a.black, ScreenWidth-b.black, b.y0, b.black, h)
		drawAt(screen, b.piece, b.black, b.y0)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(ScreenWidth-b.black), float64(b.y0))
		screen.DrawImage(b.piece, &op)
	}
	// The aiming chevron, with the horizontal line and lead ticks beside it.
	fillRect(screen, a.reticle, cx, cy, 1, 1)
	for i := 1; i <= 4; i++ {
		fillRect(screen, a.reticle, cx-i, cy+i, 1, 1)
		fillRect(screen, a.reticle, cx+i, cy+i, 1, 1)
	}
	fillRect(screen, a.reticle, cx-68, cy, 60, 1)
	fillRect(screen, a.reticle, cx+9, cy, 60, 1)
	for _, d := range [...]int{-48, -28, 28, 48} {
		fillRect(screen, a.reticle, cx+d, cy-4, 1, 4)
	}
	// Range marks: put the target between the pair of ticks for its range.
	for _, r := range rangeMarks {
		y := cy + int(math.Round(float64(dropAngle(float32(r)*100)*focalZoom)))
		fillRect(screen, a.reticle, cx-12, y, 6, 1)
		fillRect(screen, a.reticle, cx+7, y, 6, 1)
		text.Reset().Num(r).Draw(screen, cx+17, y-3, 1, colorReticle)
	}
	if g.ranging > 0 {
		text.Reset().Num(g.ranging).Str("M").Shade(colorShadow).Draw(screen, cx+40, cy-18, 1, colorCross)
	}
}

var enemyLabels = [maxEnemies + 1]string{"ENEMY 0", "ENEMY 1", "ENEMY 2", "ENEMY 3", "ENEMY 4", "ENEMY 5", "ENEMY 6", "ENEMY 7", "ENEMY 8"}

func (g *Game) drawHUD(screen *ebiten.Image) {
	a := g.art
	if g.mode == modeTitle {
		g.drawTitle(screen)
		return
	}
	// Armor left, top left.
	text.Reset().Str("HP").Shade(colorShadow).Draw(screen, 4, 5, 1, colorCross)
	bar := a.green
	switch {
	case g.hp <= hitDamage:
		bar = a.red
	case g.hp <= 2*hitDamage:
		bar = a.yellow
	}
	fillRect(screen, a.shade, 20, 4, 66, 9)
	fillRect(screen, bar, 21, 5, int(g.hp*64/100), 7)
	text.Reset().Str(enemyLabels[g.enemiesLeft()]).Shade(colorShadow).Right(screen, ScreenWidth-4, 5, 1, colorCross)
	if g.messageTicks > 0 && g.message != msgNone {
		text.Reset().Str(messages[g.message]).Shade(colorShadow).Center(screen, ScreenWidth/2, 20, 2, colorYellow)
	}
	g.drawRadar(screen)
	// Throttle, shells left and the reload, bottom right.
	text.Reset().Str(throttleNames[g.throttle]).Shade(colorShadow).Right(screen, ScreenWidth-4, 188, 1, colorCross)
	text.Reset().Str("AP ").Pad(g.ammo, 2).Shade(colorShadow).Right(screen, ScreenWidth-4, 202, 2, colorCross)
	load := a.green
	if g.reload > 0 {
		load = a.yellow
	}
	fillRect(screen, a.shade, ScreenWidth-58, 224, 54, 8)
	fillRect(screen, load, ScreenWidth-57, 225, 52*(reloadTicks-g.reload)/reloadTicks, 6)
	if g.hurt > 0 && g.hurt%8 < 4 {
		fillRect(screen, a.red, 0, 0, ScreenWidth, 4)
		fillRect(screen, a.red, 0, ScreenHeight-4, ScreenWidth, 4)
		fillRect(screen, a.red, 0, 0, 4, ScreenHeight)
		fillRect(screen, a.red, ScreenWidth-4, 0, 4, ScreenHeight)
	}
	switch g.mode {
	case modeClear:
		g.drawBanner(screen, "AREA CLEAR!")
	case modeOver:
		g.drawBanner(screen, "TANK DESTROYED")
	}
}

func (g *Game) drawBanner(screen *ebiten.Image, title string) {
	const mid = ScreenWidth / 2
	fillRect(screen, g.art.shade, 16, 56, ScreenWidth-32, 84)
	text.Reset().Str(title).Center(screen, mid, 66, 2, colorYellow)
	text.Reset().Str("KILLS ").Pad(g.kills, 2).Center(screen, mid, 94, 2, colorCross)
	if g.modeTicks > 90 && g.modeTicks/30%2 == 0 {
		text.Reset().Str("PRESS A").Center(screen, mid, 120, 1, colorCross)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	fillRect(screen, g.art.shade, 0, 0, ScreenWidth, 32)
	text.Reset().Str("IKATEN TANK").Center(screen, mid, 6, 3, colorYellow)
	if g.modeTicks/30%2 == 0 {
		text.Reset().Str("PRESS A").Shade(colorShadow).Center(screen, mid, 44, 2, colorCross)
	}
	fillRect(screen, g.art.shade, 0, 176, ScreenWidth, 64)
	text.Reset().Str("STICK: AIM   A: FIRE").Center(screen, mid, 186, 1, colorCross)
	text.Reset().Str("B + STICK: DRIVE").Center(screen, mid, 202, 1, colorCross)
	text.Reset().Str("TAP B: ZOOM").Center(screen, mid, 218, 1, colorCross)
}

func (g *Game) drawRadar(screen *ebiten.Image) {
	a := g.art
	drawAt(screen, a.radar, radarX, radarY)
	cx, cy := float32(radarX+radarSize/2), float32(radarY+radarSize/2)
	// The view's edges on the rim.
	half := float32(math.Atan(float64(centerX / g.cam.f)))
	for _, t := range [...]float32{-half, half} {
		x, y := g.radarDot(g.turret+t, radarSpan)
		fillRect(screen, a.dim, x-1, y-1, 2, 2)
	}
	for i := range g.enemies {
		e := &g.enemies[i]
		dx, dy := e.x-g.x, e.y-g.y
		switch {
		case !e.alive && e.dead > 0:
			x, y := g.radarDot(bearingTo(dx, dy), hypot(dx, dy))
			fillRect(screen, a.gray, x-1, y-1, 2, 2)
		case e.alive && g.ticks-e.lastSeen < 120:
			x, y := g.radarDot(bearingTo(dx, dy), hypot(dx, dy))
			size := 4
			if e.flash > 0 {
				size = 6
			}
			fillRect(screen, a.red, x-size/2, y-size/2, size, size)
		}
	}
	// The hull turns under the turret, which always points up.
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-3, -5)
	op.GeoM.Rotate(float64(angleDiff(g.heading, g.turret)))
	op.GeoM.Translate(float64(cx), float64(cy))
	screen.DrawImage(a.hull, &op)
	fillRect(screen, a.green, int(cx)-1, int(cy)-12, 2, 12)
	if g.hitTicks > 0 && g.hitTicks%10 < 6 {
		x, y := g.radarDot(g.hitBearing, radarSpan)
		fillRect(screen, a.yellow, x-3, y-3, 6, 6)
	}
}

// radarDot returns where a bearing and distance from the player show on the
// radar, which turns with the turret.
func (g *Game) radarDot(bearing, dist float32) (int, int) {
	const r = radarSize/2 - 3
	s, c := sincos(bearing - g.turret)
	d := min(dist*r/radarSpan, r)
	return int(radarX + radarSize/2 + s*d + 0.5), int(radarY + radarSize/2 - c*d + 0.5)
}
