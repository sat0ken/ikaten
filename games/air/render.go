package air

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

// The camera is the pilot's eye, looking along the plane's nose. A point d
// meters from the eye along right, up and forward (x, y, z) shows at
// (cx + focal*x/z, cy - focal*y/z).
const (
	viewHeight = 94 // rows of the view; the instruments are below
	cx, cy     = 60, 46
	focal      = 100 // pixels per radian near the center: about 62 degrees across
	nearZ      = 2
	reach      = 82 // pixels from the center that cover the view in any direction

	cloudCell = 3000 // meters between cloud slots
	capCell   = 50   // meters between whitecap slots
	capRange  = 500
	maxDraw   = 96
)

func rgb(r, g, b uint8) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.RGBA{r, g, b, 255})
	return img
}

var (
	colorSky = [3]*ebiten.Image{rgb(190, 214, 236), rgb(140, 182, 228), rgb(92, 140, 210)}
	// The sea from the horizon in: hazy far away, darker near.
	colorSea    = [6]*ebiten.Image{rgb(150, 176, 196), rgb(96, 132, 164), rgb(66, 108, 150), rgb(50, 92, 138), rgb(42, 82, 128), rgb(38, 76, 120)}
	colorCap    = rgb(200, 216, 230)
	colorFrame  = rgb(28, 30, 28)
	colorPanel  = rgb(24, 26, 24)
	colorEdge   = rgb(70, 74, 66)
	colorSight  = rgb(255, 150, 40)
	colorRed    = rgb(240, 60, 40)
	colorYellow = rgb(255, 220, 60)
	colorWhite  = rgb(236, 240, 244)
	colorGreen  = rgb(90, 210, 110)
	colorBar    = rgb(56, 60, 56)
	colorTracer = rgb(255, 240, 160)
	colorEnemyT = rgb(255, 130, 60)
	colorDark   = rgb(30, 32, 36)
	colorBox    = rgb(14, 18, 26)
	colorBoxHi  = rgb(70, 80, 100)
)

var (
	textWhite  = color.RGBA{236, 240, 244, 255}
	textGray   = color.RGBA{140, 150, 140, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGreen  = color.RGBA{110, 230, 130, 255}
	textAmber  = color.RGBA{255, 170, 60, 255}
)

// Sea bands begin at these distances in meters, from the horizon in.
var seaBands = [len(colorSea) - 1]float32{30000, 12000, 5000, 2000, 700}

type drawKind uint8

const (
	drawShip drawKind = iota
	drawPlane
	drawCloud
	drawPuff
)

type drawItem struct {
	z    float32
	kind drawKind
	i    int
	pos  vec // for clouds
	size float32
}

// view is the camera of one frame and what it will draw.
type view struct {
	eye, r, u, f vec
	items        [maxDraw]drawItem
	n            int
}

// project returns the screen position and distance along the view of p.
func (v *view) project(p vec) (sx, sy, z float32) {
	d := p.sub(v.eye)
	z = d.dot(v.f)
	if z < nearZ {
		return 0, 0, z
	}
	return cx + focal*d.dot(v.r)/z, cy - focal*d.dot(v.u)/z, z
}

// onScreen reports whether a projected point is in the view.
func onScreen(sx, sy, z float32) bool {
	return z >= nearZ && sx >= 0 && sx < ScreenWidth && sy >= 0 && sy < viewHeight
}

// screenVec is how a short world vector a at p shows on screen.
func (v *view) screenVec(p, a vec) (float32, float32) {
	d := p.sub(v.eye)
	z := d.dot(v.f)
	x, y := d.dot(v.r), d.dot(v.u)
	az := a.dot(v.f)
	return focal / z * (a.dot(v.r) - x/z*az), -focal / z * (a.dot(v.u) - y/z*az)
}

// drawFlat draws img on the plane through center spanned by a (the image's
// width) and b (its height), as an affine image of the perspective there.
func (g *Game) drawFlat(screen, img *ebiten.Image, center, a, b vec) {
	v := &g.view
	sx, sy, z := v.project(center)
	// Far off the view's axis the affine image is a poor fit, and the
	// object is out of view anyway.
	if z < nearZ || z < 0.6*center.sub(v.eye).length() {
		return
	}
	ax, ay := v.screenVec(center, a)
	bx, by := v.screenVec(center, b)
	if ext := abs(ax) + abs(ay) + abs(bx) + abs(by); sx < -ext || sx > ScreenWidth+ext || sy < -ext || sy > viewHeight+ext {
		return
	}
	size := img.Bounds().Size()
	w, h := float64(size.X), float64(size.Y)
	var op ebiten.DrawImageOptions
	op.GeoM.SetElement(0, 0, float64(ax)/w)
	op.GeoM.SetElement(0, 1, float64(bx)/h)
	op.GeoM.SetElement(1, 0, float64(ay)/w)
	op.GeoM.SetElement(1, 1, float64(by)/h)
	op.GeoM.SetElement(0, 2, float64(sx-ax/2-bx/2))
	op.GeoM.SetElement(1, 2, float64(sy-ay/2-by/2))
	screen.DrawImage(img, &op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	v := &g.view
	p := g.player()
	v.eye, v.f, v.u = p.pos, p.f, p.u
	v.r = p.right()
	if g.shake > 0 {
		v.eye = v.eye.add(v.u.mul(float32(g.ticks%3-1) * 0.3))
	}
	if !p.alive {
		v.eye.z = max(v.eye.z, 3)
	}
	g.drawSkyAndSea(screen)
	if v.eye.z < capRange {
		g.drawWhitecaps(screen)
	}
	g.collect()
	g.drawItems(screen)
	g.drawShots(screen)
	if g.mode == modePlay && p.alive {
		g.drawMarkers(screen)
		g.drawSight(screen)
	}
	drawAt(screen, imageCowl, cx-32, viewHeight-9)
	fillRect(screen, colorFrame, 0, 0, ScreenWidth, 2)
	g.drawPanel(screen)
	g.drawOverlay(screen)
}

// band fills the part of the view between two lines parallel to the
// horizon, d0 and d1 pixels from the center along n (a unit vector toward
// the sky), as a rotated rectangle.
func band(screen, c *ebiten.Image, nx, ny, d0, d1 float32) {
	d0, d1 = max(d0, -reach), min(d1, reach)
	if d1 <= d0 {
		return
	}
	tx, ty := ny, -nx // along the lines
	var op ebiten.DrawImageOptions
	op.GeoM.SetElement(0, 0, float64(tx*2*reach))
	op.GeoM.SetElement(1, 0, float64(ty*2*reach))
	op.GeoM.SetElement(0, 1, float64(nx*(d1-d0)))
	op.GeoM.SetElement(1, 1, float64(ny*(d1-d0)))
	op.GeoM.SetElement(0, 2, float64(cx+nx*d0-tx*reach))
	op.GeoM.SetElement(1, 2, float64(cy+ny*d0-ty*reach))
	screen.DrawImage(c, &op)
}

// drawSkyAndSea draws bands parallel to the horizon. A screen point (px,
// py) from the center looks along px*r - py*u + focal*f, which climbs by
// s = px*r.z - py*u.z + focal*f.z: zero on the horizon and, below it, about
// -focal times the slope down to the sea, which gives the distance.
func (g *Game) drawSkyAndSea(screen *ebiten.Image) {
	v := &g.view
	nx, ny := v.r.z, -v.u.z
	m := sqrt(nx*nx + ny*ny)
	if m < 1e-3 {
		// Straight up or down.
		c := colorSky[2]
		if v.f.z < 0 {
			c = colorSea[len(colorSea)-1]
		}
		fillRect(screen, c, 0, 0, ScreenWidth, viewHeight)
		return
	}
	nx, ny = nx/m, ny/m
	// A climb of s is d = (s - focal*f.z)/m pixels from the center along n.
	off := focal * v.f.z
	sky := [...]float32{0, 0.05 * focal, 0.25 * focal, 1e6}
	for i := range colorSky {
		band(screen, colorSky[i], nx, ny, (sky[i]-off)/m, (sky[i+1]-off)/m)
	}
	alt := max(v.eye.z, 1)
	upper := float32(0)
	for i := range colorSea {
		lower := float32(-1e6)
		if i < len(seaBands) {
			lower = -focal * alt / seaBands[i]
		}
		band(screen, colorSea[i], nx, ny, (lower-off)/m, (upper-off)/m)
		upper = lower
	}
}

func hash(a, b int) uint32 {
	h := uint32(a)*2654435761 ^ uint32(b)*2246822519
	h ^= h >> 15
	h *= 2246822519
	h ^= h >> 13
	return h
}

// drawWhitecaps scatters foam on the sea near a low plane, for the sense of
// speed.
func (g *Game) drawWhitecaps(screen *ebiten.Image) {
	v := &g.view
	x0, y0 := int(v.eye.x/capCell), int(v.eye.y/capCell)
	const n = capRange / capCell
	for i := x0 - n; i <= x0+n; i++ {
		for j := y0 - n; j <= y0+n; j++ {
			h := hash(i, j)
			if h%3 != 0 {
				continue
			}
			p := vec{float32(i*capCell) + float32(h>>8%capCell), float32(j*capCell) + float32(h>>16%capCell), 0}
			if p.sub(v.eye).dot(v.f) < nearZ {
				continue
			}
			sx, sy, z := v.project(p)
			if !onScreen(sx, sy, z) {
				continue
			}
			w := max(1, int(4*focal/z))
			fillRect(screen, colorCap, int(sx), int(sy), w, max(1, w/3))
		}
	}
}

func (g *Game) add(item drawItem) {
	if g.view.n < maxDraw && item.z >= nearZ {
		g.view.items[g.view.n] = item
		g.view.n++
	}
}

// collect lists the ships, planes, clouds and puffs in front of the eye.
func (g *Game) collect() {
	v := &g.view
	v.n = 0
	for i := range g.ships {
		if s := &g.ships[i]; s.alive {
			g.add(drawItem{z: s.pos.sub(v.eye).dot(v.f), kind: drawShip, i: i})
		}
	}
	for i := 1; i < maxPlanes; i++ {
		if q := &g.planes[i]; q.alive {
			g.add(drawItem{z: q.pos.sub(v.eye).dot(v.f), kind: drawPlane, i: i})
		}
	}
	for i := range g.puffs {
		if pf := &g.puffs[i]; pf.life > 0 {
			g.add(drawItem{z: pf.pos.sub(v.eye).dot(v.f), kind: drawPuff, i: i})
		}
	}
	x0, y0 := int(v.eye.x/cloudCell), int(v.eye.y/cloudCell)
	for i := x0 - 2; i <= x0+2; i++ {
		for j := y0 - 2; j <= y0+2; j++ {
			h := hash(i+7777, j)
			if h%2 == 0 {
				continue
			}
			p := vec{float32(i*cloudCell) + float32(h>>4%cloudCell), float32(j*cloudCell) + float32(h>>12%cloudCell), 1400 + float32(h>>20%900)}
			g.add(drawItem{z: p.sub(v.eye).dot(v.f), kind: drawCloud, pos: p, size: 300 + float32(h>>8%300)})
		}
	}
	// Far to near.
	items := v.items[:v.n]
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].z > items[j-1].z; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func (g *Game) drawItems(screen *ebiten.Image) {
	v := &g.view
	for k := range v.n {
		it := &v.items[k]
		switch it.kind {
		case drawShip:
			g.drawShip(screen, &g.ships[it.i])
		case drawPlane:
			g.drawPlane(screen, &g.planes[it.i])
		case drawCloud:
			g.drawBillboard(screen, imageCloud, it.pos, it.size)
		case drawPuff:
			pf := &g.puffs[it.i]
			if pf.kind == puffSplash {
				g.drawStanding(screen, imagePuff[pf.kind], pf.pos, pf.size*min(1, float32(pf.age)/20+0.3))
			} else {
				g.drawBillboard(screen, imagePuff[pf.kind], pf.pos, pf.size)
			}
		}
	}
}

// drawBillboard draws img centered at p, size meters wide.
func (g *Game) drawBillboard(screen, img *ebiten.Image, p vec, size float32) {
	sx, sy, z := g.view.project(p)
	if z < nearZ {
		return
	}
	w := size * focal / z
	if w < 1 {
		fillRect(screen, imagePuffDot(img), int(sx), int(sy), 1, 1)
		return
	}
	drawScaled(screen, img, sx, sy, w/float32(img.Bounds().Dx()), true)
}

// drawStanding draws img standing on p, size meters tall.
func (g *Game) drawStanding(screen, img *ebiten.Image, p vec, size float32) {
	sx, sy, z := g.view.project(p)
	if z < nearZ {
		return
	}
	h := size * focal / z
	drawScaled(screen, img, sx, sy, max(h, 1)/float32(img.Bounds().Dy()), false)
}

// imagePuffDot is the color of a far billboard.
func imagePuffDot(img *ebiten.Image) *ebiten.Image {
	if img == imageCloud {
		return colorWhite
	}
	return colorDark
}

func (g *Game) drawShip(screen *ebiten.Image, s *ship) {
	v := &g.view
	t := &shipTypes[s.kind]
	h := s.dir()
	side := vec{h.y, -h.x, 0} // starboard
	sink := float32(s.sink) / 900 * (t.freeboard + t.height)
	if _, _, z := v.project(s.pos); z > 0 && t.length*focal/z < 2 {
		sx, sy, _ := v.project(s.pos)
		if onScreen(sx, sy, z) {
			fillRect(screen, colorDark, int(sx), int(sy), 1, 1)
		}
		return
	}
	if s.sink == 0 && s.speed > 1 {
		g.drawFlat(screen, imageWake, s.pos.sub(h.mul(t.length*1.2)), h.mul(t.length*1.4), side.mul(t.beam*3))
	}
	up := vec{0, 0, 1}
	hull := s.pos.add(up.mul(t.freeboard/2 - sink))
	g.drawFlat(screen, imageHull, hull, h.mul(t.length), up.mul(-t.freeboard))
	if s.sink == 0 {
		g.drawFlat(screen, imageDeck[s.kind], s.pos.add(up.mul(t.freeboard)), h.mul(t.length), side.mul(t.beam))
	}
	upper := t.height - t.freeboard
	g.drawFlat(screen, imageUpper[s.kind], s.pos.add(up.mul(t.freeboard+upper/2-sink)), h.mul(t.length), up.mul(-upper))
}

func (g *Game) drawPlane(screen *ebiten.Image, q *plane) {
	v := &g.view
	t := &planeTypes[q.kind]
	sx, sy, z := v.project(q.pos)
	if z < nearZ {
		return
	}
	if t.span*focal/z < 2.5 {
		if onScreen(sx, sy, z) {
			fillRect(screen, colorDark, int(sx), int(sy), 1, 1)
		}
		return
	}
	img := &imageFighter
	if q.kind == bomber {
		img = &imageBomber
	}
	r := q.right()
	g.drawFlat(screen, img[0], q.pos, q.f.mul(t.length), r.mul(t.span))
	g.drawFlat(screen, img[1], q.pos.add(q.u.mul(0.5)), q.f.mul(t.length), q.u.mul(-t.length/3))
}

// drawShots draws rounds, the bomb and the torpedo as dots and streaks.
func (g *Game) drawShots(screen *ebiten.Image) {
	v := &g.view
	for i := range g.bullets {
		b := &g.bullets[i]
		if b.life <= 0 {
			continue
		}
		c := colorTracer
		if b.enemy {
			c = colorEnemyT
		}
		for _, back := range [2]float32{0, 0.015} {
			sx, sy, z := v.project(b.pos.sub(b.vel.mul(back)))
			if onScreen(sx, sy, z) {
				fillRect(screen, c, int(sx), int(sy), 1, 1)
			}
		}
	}
	if g.bomb.alive {
		if sx, sy, z := v.project(g.bomb.pos); onScreen(sx, sy, z) {
			fillRect(screen, colorDark, int(sx), int(sy), 2, 2)
		}
	}
	if g.torp.alive {
		for k := range 6 {
			p := g.torp.pos.sub(g.torp.dir.mul(float32(k) * 12))
			if sx, sy, z := v.project(p); onScreen(sx, sy, z) {
				fillRect(screen, colorCap, int(sx), int(sy), 1, 1)
			}
		}
	}
}

// drawSight draws the gunsight or the bombsight at the center.
func (g *Game) drawSight(screen *ebiten.Image) {
	fillRect(screen, colorSight, cx-6, cy, 4, 1)
	fillRect(screen, colorSight, cx+3, cy, 4, 1)
	fillRect(screen, colorSight, cx, cy-6, 1, 4)
	fillRect(screen, colorSight, cx, cy+3, 1, 4)
	if g.player().kind == zero {
		// A ring of about two degrees.
		for _, d := range [...][2]int{{-4, -1}, {-4, 0}, {-4, 1}, {4, -1}, {4, 0}, {4, 1}, {-1, -4}, {0, -4}, {1, -4}, {-1, 4}, {0, 4}, {1, 4},
			{-3, -3}, {3, -3}, {-3, 3}, {3, 3}} {
			fillRect(screen, colorSight, cx+d[0], cy+d[1], 1, 1)
		}
	}
}

// brackets marks a point with four corners.
func brackets(screen, c *ebiten.Image, x, y, r int) {
	for _, d := range [4][2]int{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
		fillRect(screen, c, x+d[0]*r-max(d[0], 0), y+d[1]*r, 2, 1)
		fillRect(screen, c, x+d[0]*r, y+d[1]*r-max(d[1], 0), 1, 2)
	}
}

// edgeArrow points from the view's edge toward a point out of view.
func (g *Game) edgeArrow(screen, c *ebiten.Image, p vec) {
	v := &g.view
	d := p.sub(v.eye)
	x, y, z := d.dot(v.r), -d.dot(v.u), d.dot(v.f)
	if z > nearZ {
		sx, sy := cx+focal*x/z, cy+focal*y/z
		if onScreen(sx, sy, z) {
			return
		}
	}
	// Keep clear of the cowling at the bottom.
	down := float32(cy - 4)
	if y > 0 {
		down = viewHeight - 13 - cy
	}
	m := max(abs(x)/(cx-4), abs(y)/down)
	if m == 0 {
		return
	}
	ax, ay := int(cx+x/m), int(cy+y/m)
	fillRect(screen, c, ax-1, ay-1, 3, 3)
	fillRect(screen, colorDark, ax, ay, 1, 1)
}

// drawMarkers marks the enemy planes and ships, and the aim points.
func (g *Game) drawMarkers(screen *ebiten.Image) {
	v := &g.view
	p := g.player()
	for i := 1; i < maxPlanes; i++ {
		q := &g.planes[i]
		if !q.alive || q.falling > 0 {
			continue
		}
		sx, sy, z := v.project(q.pos)
		if onScreen(sx, sy, z) && z < 4000 {
			c := colorRed
			if i != g.enemy {
				c = colorEnemyT
			}
			brackets(screen, c, int(sx), int(sy), 4)
		}
	}
	if g.enemy >= 0 {
		q := &g.planes[g.enemy]
		g.edgeArrow(screen, colorRed, q.pos)
		if p.kind == zero {
			// The lead: where to put the sight for the rounds to meet it.
			d := q.pos.sub(p.pos).length()
			if d < 900 {
				tof := d / (bulletSpeed + p.v)
				aim := q.pos.add(q.f.mul(q.v * tof)).add(vec{0, 0, gravity * tof * tof / 2})
				if sx, sy, z := v.project(aim); onScreen(sx, sy, z) {
					fillRect(screen, colorYellow, int(sx)-1, int(sy), 3, 1)
					fillRect(screen, colorYellow, int(sx), int(sy)-1, 1, 3)
				}
			}
		}
	}
	if g.mission == missionFighter {
		g.edgeArrow(screen, colorGreen, g.ships[0].pos)
		return
	}
	if g.target < 0 {
		return
	}
	s := &g.ships[g.target]
	if sx, sy, z := v.project(s.pos); onScreen(sx, sy, z) {
		brackets(screen, colorYellow, int(sx), int(sy), 5)
	}
	g.edgeArrow(screen, colorYellow, s.pos)
	if g.weapons == 0 {
		return
	}
	switch p.kind {
	case val:
		if p.f.z < -0.3 {
			if sx, sy, z := v.project(g.bombImpact()); onScreen(sx, sy, z) {
				fillRect(screen, colorYellow, int(sx)-2, int(sy), 5, 1)
				fillRect(screen, colorYellow, int(sx), int(sy)-2, 1, 5)
			}
		}
	case kate:
		if aim, ok := g.lead(g.target); ok {
			if sx, sy, z := v.project(aim); onScreen(sx, sy, z) {
				fillRect(screen, colorYellow, int(sx)-1, int(sy)-3, 3, 1)
				fillRect(screen, colorYellow, int(sx), int(sy)-5, 1, 5)
			}
		}
	}
}

func (g *Game) drawPanel(screen *ebiten.Image) {
	p := g.player()
	fillRect(screen, colorPanel, 0, viewHeight, ScreenWidth, ScreenHeight-viewHeight)
	fillRect(screen, colorEdge, 0, viewHeight, ScreenWidth, 1)
	if g.mode == modeSelect {
		return
	}
	kmh := int(p.v * 3.6)
	sc := textWhite
	if p.v > planeTypes[p.kind].vne {
		sc = textRed
	}
	buf.reset().str("SPD").draw(screen, 3, 97, textGray)
	buf.reset().num(kmh).draw(screen, -35, 97, sc)
	ac := textWhite
	if p.pos.z < 100 {
		ac = textAmber
	}
	buf.reset().str("ALT").draw(screen, 40, 97, textGray)
	buf.reset().num(int(max(p.pos.z, 0))).draw(screen, -76, 97, ac)
	buf.reset().str("HDG").draw(screen, 82, 97, textGray)
	hdg := int(bearingOf(p.f)*57.3+0.5) % 360
	buf.reset().pad(hdg, 3).draw(screen, 98, 97, textWhite)

	switch p.kind {
	case zero:
		buf.reset().str("7.7").draw(screen, 3, 105, textGray)
		buf.reset().num(g.ammo).draw(screen, 18, 105, textWhite)
		buf.reset().str("20MM").draw(screen, 40, 105, textGray)
		buf.reset().num(g.cannon).draw(screen, 58, 105, textWhite)
	case val:
		buf.reset().str("BOMB").draw(screen, 3, 105, textGray)
		buf.reset().num(g.weapons).draw(screen, 22, 105, textWhite)
		if p.brakes {
			buf.reset().str("BRAKE").draw(screen, 30, 105, textGreen)
		}
	case kate:
		buf.reset().str("TORP").draw(screen, 3, 105, textGray)
		buf.reset().num(g.weapons).draw(screen, 22, 105, textWhite)
		if p.throttle < 1 {
			buf.reset().str("THR LOW").draw(screen, 30, 105, textGreen)
		}
		if g.weapons > 0 {
			switch {
			case g.torpedoOK():
				buf.reset().str("OK").draw(screen, 62, 105, textGreen)
			case p.pos.z > torpMaxAlt:
				buf.reset().str("HIGH").draw(screen, 62, 105, textRed)
			case p.v*3.6 > torpMaxKmh:
				buf.reset().str("FAST").draw(screen, 62, 105, textRed)
			default:
				buf.reset().str("LEVEL").draw(screen, 62, 105, textRed)
			}
		}
	}
	if g.mission != missionFighter && g.target >= 0 {
		d := g.ships[g.target].pos.sub(p.pos.flat()).length()
		buf.reset().num(int(d)).str("M").draw(screen, -117, 105, textYellow)
	}
	// Damage left, the score, and the raid or the sortie.
	fillRect(screen, colorBar, 3, 113, 30, 5)
	hc := colorGreen
	if p.hp < 40 {
		hc = colorRed
	}
	fillRect(screen, hc, 3, 113, max(p.hp, 0)*30/100, 5)
	buf.reset().num(g.score).draw(screen, -76, 113, textYellow)
	if g.mission == missionFighter {
		buf.reset().str("CV").draw(screen, 82, 113, textGray)
		for i := range 3 {
			c := colorBar
			if i < g.ships[0].hp {
				c = colorGreen
			}
			fillRect(screen, c, 92+i*5, 113, 4, 5)
		}
	} else {
		buf.reset().num(g.sortie).str("/").num(sorties).draw(screen, -117, 113, textWhite)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	switch g.mode {
	case modeSelect:
		g.drawSelect(screen)
		return
	case modeDebrief:
		box(screen, 10, 16, 100, 58)
		buf.reset().str("SORTIE ").num(g.sortie).str(" OF ").num(sorties).draw(screen, 30, 22, textWhite)
		r := g.lastResult
		if r == "" {
			r = "NO RELEASE"
		}
		buf.reset().str(r).center(screen, 36, textYellow)
		buf.reset().str("SCORE ").num(g.score).center(screen, 48, textWhite)
		if g.modeTicks > 45 && g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 56)
		}
		return
	case modeResult:
		box(screen, 10, 16, 100, 58)
		buf.reset().str(g.outcome).center(screen, 24, textYellow)
		buf.reset().str("SCORE").center(screen, 36, textGray)
		buf.reset().num(g.score).center2(screen, 44, textWhite)
		if g.mission != missionFighter {
			buf.reset().str("PLANES LOST ").num(g.lost).center(screen, 56, textGray)
		}
		if g.modeTicks > 60 && g.modeTicks/30%2 == 0 {
			ebitenutil.DebugPrintAt(screen, "PRESS A", 39, 62)
		}
		return
	}
	if g.messageTime > 0 && g.message != msgNone && g.messageTime%30 < 22 {
		msg := messages[g.message]
		if g.message == msgWave {
			buf.reset().str("RAID ").num(g.wave).str(" OF ").num(waves).center(screen, 18, textRed)
		}
		w := len(msg)*6 + 6
		box(screen, (ScreenWidth-w)/2-1, 4, w, 10)
		ebitenutil.DebugPrintAt(screen, msg, (ScreenWidth-len(msg)*6)/2+2, 1)
	}
}

var missionLines = [missions][2]string{
	{"TYPE 0 FIGHTER", "SHOOT DOWN THE RAIDERS"},
	{"TYPE 99 DIVE BOMBER", "DIVE AND BOMB THE FLEET"},
	{"TYPE 97 ATTACK PLANE", "TORPEDO LOW AND SLOW"},
}

var missionNames = [missions]string{"FIGHTER", "DIVE BOMB", "TORPEDO"}

func (g *Game) drawSelect(screen *ebiten.Image) {
	box(screen, 8, 6, 104, 84)
	ebitenutil.DebugPrintAt(screen, "IKATEN AIR", 30, 3)
	for i, name := range missionNames {
		y := 20 + i*11
		if i == g.cursor {
			fillRect(screen, colorBoxHi, 10, y-1, 100, 10)
			if g.ticks/20%2 == 0 {
				ebitenutil.DebugPrintAt(screen, ">", 14, y-3)
			}
		}
		ebitenutil.DebugPrintAt(screen, name, 24, y-3)
	}
	l := &missionLines[g.cursor]
	buf.reset().str(l[0]).center(screen, 56, textYellow)
	buf.reset().str(l[1]).center(screen, 63, textWhite)
	stick := "B: STICK UP = DIVE"
	if g.climbUp {
		stick = "B: STICK UP = CLIMB"
	}
	buf.reset().str(stick).center(screen, 75, textGray)
	buf.reset().str("A: START").center(screen, 82, textGray)
	// The instruments show the controls.
	buf.reset().str("ARROWS FLY").draw(screen, 3, 99, textGray)
	switch g.cursor {
	case int(missionFighter):
		buf.reset().str("A 7.7MM  B 20MM CANNON").draw(screen, 3, 107, textWhite)
	case int(missionDive):
		buf.reset().str("A BOMB  B DIVE BRAKES").draw(screen, 3, 107, textWhite)
	case int(missionTorpedo):
		buf.reset().str("A TORPEDO  B THROTTLE").draw(screen, 3, 107, textWhite)
	}
}

func box(screen *ebiten.Image, x, y, w, h int) {
	fillRect(screen, colorBoxHi, x-1, y-1, w+2, h+2)
	fillRect(screen, colorBox, x, y, w, h)
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
	for ; i < len(d); i++ {
		t.b[t.n] = d[i]
		t.n++
	}
	return t
}

// pad writes n with leading zeros to the given width.
func (t *text) pad(n, width int) *text {
	for w, m := 1, 10; w < width; w, m = w+1, m*10 {
		if n < m {
			t.str("0")
		}
	}
	return t.num(n)
}

// draw draws the line with its left edge at x, or its right edge at -x.
func (t *text) draw(screen *ebiten.Image, x, y int, c color.RGBA) {
	if x < 0 {
		x = -x - t.n*4 + 1
	}
	drawText(screen, t.b[:t.n], x, y, 1, c)
}

func (t *text) center(screen *ebiten.Image, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], (ScreenWidth-t.n*4)/2, y, 1, c)
}

func (t *text) center2(screen *ebiten.Image, y int, c color.RGBA) {
	drawText(screen, t.b[:t.n], (ScreenWidth-t.n*8)/2, y, 2, c)
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

// drawScaled draws img scaled, centered at x, y or standing on it.
func drawScaled(screen, img *ebiten.Image, x, y, scale float32, centered bool) {
	b := img.Bounds()
	w, h := float32(b.Dx())*scale, float32(b.Dy())*scale
	top := y - h
	if centered {
		top = y - h/2
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x-w/2), float64(top))
	screen.DrawImage(img, &op)
}

// fillRect draws a 1x1 image scaled to a rectangle.
func fillRect(screen, c *ebiten.Image, x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(width), float64(height))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(c, &op)
}
