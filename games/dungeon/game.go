// Package dungeon is an original first-person shooter in the style of early
// raycasting games. It uses only the Ebitengine API, and every image is drawn
// by code, so it also builds with Ebitengine.
package dungeon

import (
	"image"
	"image/color"
	"math"
	"strconv"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	viewHeight   = 104 // the rows below hold the status bar
	texSize      = 16
	planeLength  = 0.66 // camera plane half-width; about a 66 degree view
	moveSpeed    = 0.06
	turnSpeed    = 0.05
	playerRadius = 0.2
	slimeSpeed   = 0.02
	slimeReach   = 0.7
	maxSlimes    = 8
)

// '#' brick, '=' stone, '%' metal, 'P' player start, 'S' slime.
var worldMap = [...]string{
	"################",
	"#P.....#.......#",
	"#......#..S....#",
	"#..==..#.......#",
	"#..==.....%%...#",
	"#.........%%...#",
	"#####..#.......#",
	"#......#####.###",
	"#..S...#.......#",
	"#......#..S....#",
	"#..%...........#",
	"#..%...#####...#",
	"#......#...#...#",
	"#.S....#.S.#.S.#",
	"#..........#...#",
	"################",
}

const mapSize = len(worldMap)

func wallAt(x, y int) byte {
	if x < 0 || y < 0 || x >= mapSize || y >= mapSize {
		return '#'
	}
	switch c := worldMap[y][x]; c {
	case '#', '=', '%':
		return c
	}
	return 0
}

type mode int

const (
	modeTitle mode = iota
	modePlay
	modeDead
	modeClear
)

type slime struct {
	x, y     float32
	hp       int
	hurt     int // ticks of the red flash after a hit
	cooldown int // ticks until the next attack
	alive    bool
}

type Game struct {
	mode           mode
	px, py         float32 // player position
	dirX, dirY     float32
	planeX, planeY float32
	hp, kills      int
	flash          int // ticks of the muzzle flash
	damage         int // ticks of the red screen after being hit
	ticks          int
	slimes         [maxSlimes]slime
	slimeCount     int
	zbuf           [ScreenWidth]float32
	order          [maxSlimes]int
	hud            []byte
	hudText        string // rebuilt only when hp or kills change
	hudHP, hudKill int
}

func NewGame() *Game {
	g := &Game{}
	g.reset()
	return g
}

func (g *Game) reset() {
	*g = Game{mode: g.mode, hud: g.hud}
	g.hp = 100
	g.dirX, g.dirY = 1, 0
	g.planeX, g.planeY = 0, planeLength
	for y, row := range worldMap {
		for x := range len(row) {
			switch row[x] {
			case 'P':
				g.px, g.py = float32(x)+0.5, float32(y)+0.5
			case 'S':
				if g.slimeCount < maxSlimes {
					g.slimes[g.slimeCount] = slime{x: float32(x) + 0.5, y: float32(y) + 0.5, hp: 3, alive: true}
					g.slimeCount++
				}
			}
		}
	}
}

func firePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Update() error {
	g.ticks++
	switch g.mode {
	case modeTitle:
		if firePressed() {
			g.mode = modePlay
		}
	case modePlay:
		g.updatePlayer()
		g.updateSlimes()
		if g.hp <= 0 {
			g.mode = modeDead
		} else if g.kills == g.slimeCount {
			g.mode = modeClear
		}
	case modeDead, modeClear:
		if firePressed() {
			g.mode = modeTitle
			g.reset()
		}
	}
	if g.flash > 0 {
		g.flash--
	}
	if g.damage > 0 {
		g.damage--
	}
	return nil
}

func free(x, y float32) bool {
	return wallAt(int(x-playerRadius), int(y-playerRadius)) == 0 &&
		wallAt(int(x+playerRadius), int(y-playerRadius)) == 0 &&
		wallAt(int(x-playerRadius), int(y+playerRadius)) == 0 &&
		wallAt(int(x+playerRadius), int(y+playerRadius)) == 0
}

var turnCos, turnSin = float32(math.Cos(turnSpeed)), float32(math.Sin(turnSpeed))

func (g *Game) updatePlayer() {
	turn := func(s float32) {
		g.dirX, g.dirY = g.dirX*turnCos-g.dirY*s, g.dirX*s+g.dirY*turnCos
		g.planeX, g.planeY = g.planeX*turnCos-g.planeY*s, g.planeX*s+g.planeY*turnCos
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		turn(-turnSin)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		turn(turnSin)
	}
	step := float32(0)
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		step += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		step -= moveSpeed
	}
	// Move each axis separately so the player slides along walls.
	if nx := g.px + g.dirX*step; free(nx, g.py) {
		g.px = nx
	}
	if ny := g.py + g.dirY*step; free(g.px, ny) {
		g.py = ny
	}
	if firePressed() {
		g.flash = 5
		g.shoot()
	}
}

// project returns a slime's horizontal screen position and its depth.
func (g *Game) project(s *slime) (screenX, depth float32) {
	sx, sy := s.x-g.px, s.y-g.py
	invDet := 1 / (g.planeX*g.dirY - g.dirX*g.planeY)
	tx := invDet * (g.dirY*sx - g.dirX*sy)
	depth = invDet * (-g.planeY*sx + g.planeX*sy)
	return ScreenWidth / 2 * (1 + tx/depth), depth
}

func spriteSize(depth float32) float32 { return viewHeight / depth * 0.7 }

// shoot hits the nearest slime under the crosshair that is in front of the wall.
func (g *Game) shoot() {
	best, bestDepth := -1, g.zbuf[ScreenWidth/2]
	for i := range g.slimeCount {
		s := &g.slimes[i]
		if !s.alive {
			continue
		}
		x, depth := g.project(s)
		if depth < 0.2 || depth >= bestDepth {
			continue
		}
		if half := spriteSize(depth) / 2; x-half <= ScreenWidth/2 && ScreenWidth/2 <= x+half {
			best, bestDepth = i, depth
		}
	}
	if best < 0 {
		return
	}
	s := &g.slimes[best]
	s.hp--
	s.hurt = 8
	if s.hp <= 0 {
		s.alive = false
		g.kills++
	}
}

func (g *Game) updateSlimes() {
	for i := range g.slimeCount {
		s := &g.slimes[i]
		if s.hurt > 0 {
			s.hurt--
		}
		if !s.alive {
			continue
		}
		if s.cooldown > 0 {
			s.cooldown--
		}
		dx, dy := g.px-s.x, g.py-s.y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist < slimeReach {
			if s.cooldown == 0 {
				g.hp -= 10
				g.damage = 10
				s.cooldown = 45
			}
			continue
		}
		if dist > 8 {
			continue
		}
		nx, ny := s.x+dx/dist*slimeSpeed, s.y+dy/dist*slimeSpeed
		if wallAt(int(nx), int(s.y)) == 0 {
			s.x = nx
		}
		if wallAt(int(s.x), int(ny)) == 0 {
			s.y = ny
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	fillRect(screen, imageCeiling, 0, 0, ScreenWidth, viewHeight/2)
	fillRect(screen, imageFloor, 0, viewHeight/2, ScreenWidth, viewHeight/2)
	g.drawWalls(screen)
	g.drawSlimes(screen)
	if g.mode == modePlay {
		g.drawWeapon(screen)
	}
	if g.damage > 0 {
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(ScreenWidth, viewHeight)
		op.ColorScale.ScaleAlpha(0.4)
		screen.DrawImage(imageRed, &op)
	}
	g.drawStatus(screen)
}

func (g *Game) drawWalls(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	for x := range ScreenWidth {
		cameraX := 2*float32(x)/ScreenWidth - 1
		rayX, rayY := g.dirX+g.planeX*cameraX, g.dirY+g.planeY*cameraX
		mapX, mapY := int(g.px), int(g.py)
		deltaX, deltaY := float32(1e30), float32(1e30)
		if rayX != 0 {
			deltaX = abs(1 / rayX)
		}
		if rayY != 0 {
			deltaY = abs(1 / rayY)
		}
		stepX, stepY := 1, 1
		sideX := (float32(mapX) + 1 - g.px) * deltaX
		if rayX < 0 {
			stepX, sideX = -1, (g.px-float32(mapX))*deltaX
		}
		sideY := (float32(mapY) + 1 - g.py) * deltaY
		if rayY < 0 {
			stepY, sideY = -1, (g.py-float32(mapY))*deltaY
		}
		// Digital differential analysis: step cell by cell until a wall.
		var wall byte
		side := 0
		for wall == 0 {
			if sideX < sideY {
				sideX += deltaX
				mapX += stepX
				side = 0
			} else {
				sideY += deltaY
				mapY += stepY
				side = 1
			}
			wall = wallAt(mapX, mapY)
		}
		dist := sideX - deltaX
		wallX := g.py + dist*rayY
		if side == 1 {
			dist = sideY - deltaY
			wallX = g.px + dist*rayX
		}
		dist = max(dist, 0.05)
		g.zbuf[x] = dist
		texX := int((wallX - float32(int(wallX))) * texSize)
		if (side == 0 && rayX > 0) || (side == 1 && rayY < 0) {
			texX = texSize - 1 - texX
		}
		height := viewHeight / dist
		shade := min(1, max(0.25, 1.2-dist*0.12))
		if side == 1 {
			shade *= 0.75
		}
		op.GeoM.Reset()
		op.GeoM.Scale(1, float64(height/texSize))
		op.GeoM.Translate(float64(x), float64(viewHeight/2-height/2))
		op.ColorScale.Reset()
		op.ColorScale.Scale(shade, shade, shade, 1)
		screen.DrawImage(wallColumns[wallIndex(wall)][texX], &op)
	}
}

func (g *Game) drawSlimes(screen *ebiten.Image) {
	// Painter's algorithm: sort the visible slimes far to near.
	n := 0
	for i := range g.slimeCount {
		if g.slimes[i].alive || g.slimes[i].hurt > 0 {
			g.order[n] = i
			n++
		}
	}
	dist := func(i int) float32 {
		dx, dy := g.slimes[i].x-g.px, g.slimes[i].y-g.py
		return dx*dx + dy*dy
	}
	for i := 1; i < n; i++ {
		for j := i; j > 0 && dist(g.order[j-1]) < dist(g.order[j]); j-- {
			g.order[j-1], g.order[j] = g.order[j], g.order[j-1]
		}
	}
	var op ebiten.DrawImageOptions
	for _, i := range g.order[:n] {
		s := &g.slimes[i]
		centerX, depth := g.project(s)
		if depth < 0.2 {
			continue
		}
		size := spriteSize(depth)
		left := centerX - size/2
		top := viewHeight/2 + viewHeight/depth/2 - size
		shade := min(1, max(0.3, 1.2-depth*0.12))
		for x := max(0, int(left)); x < min(ScreenWidth, int(left+size)+1); x++ {
			if depth >= g.zbuf[x] {
				continue
			}
			texX := int((float32(x) - left) * texSize / size)
			if texX < 0 || texX >= texSize {
				continue
			}
			op.GeoM.Reset()
			op.GeoM.Scale(1, float64(size/texSize))
			op.GeoM.Translate(float64(x), float64(top))
			op.ColorScale.Reset()
			if s.hurt > 0 {
				op.ColorScale.Scale(3, 0.5, 0.5, 1) // flash red; the slime is mostly green
			} else {
				op.ColorScale.Scale(shade, shade, shade, 1)
			}
			screen.DrawImage(slimeColumns[texX], &op)
		}
	}
}

func (g *Game) drawWeapon(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	bob := float64(0)
	if g.flash > 0 {
		bob = 2
	}
	op.GeoM.Translate(ScreenWidth/2-12, viewHeight-20+bob)
	screen.DrawImage(imageWeapon, &op)
	if g.flash > 0 {
		op.GeoM.Reset()
		op.GeoM.Translate(ScreenWidth/2-6, viewHeight-32)
		screen.DrawImage(imageFlash, &op)
	}
	// Crosshair.
	fillRect(screen, imageWhite, ScreenWidth/2, viewHeight/2-2, 1, 5)
	fillRect(screen, imageWhite, ScreenWidth/2-2, viewHeight/2, 5, 1)
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	fillRect(screen, imageBlack, 0, viewHeight, ScreenWidth, ScreenHeight-viewHeight)
	if g.hudText == "" || g.hudHP != g.hp || g.hudKill != g.kills {
		g.hud = append(g.hud[:0], "HP "...)
		g.hud = strconv.AppendInt(g.hud, int64(max(g.hp, 0)), 10)
		g.hud = append(g.hud, "  KILL "...)
		g.hud = strconv.AppendInt(g.hud, int64(g.kills), 10)
		g.hud = append(g.hud, '/')
		g.hud = strconv.AppendInt(g.hud, int64(g.slimeCount), 10)
		g.hudText, g.hudHP, g.hudKill = string(g.hud), g.hp, g.kills
	}
	ebitenutil.DebugPrintAt(screen, g.hudText, 2, viewHeight+2)
	switch g.mode {
	case modeTitle:
		ebitenutil.DebugPrintAt(screen, "IKATEN DUNGEON\n\nDEFEAT ALL SLIMES\nPRESS Z", 14, 20)
	case modeDead:
		ebitenutil.DebugPrintAt(screen, "YOU DIED\n\nPRESS Z", 34, 36)
	case modeClear:
		ebitenutil.DebugPrintAt(screen, "ALL CLEAR!\n\nPRESS Z", 30, 36)
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// fillRect draws a 1x1 image scaled to a rectangle; this avoids per-pixel Set.
func fillRect(screen, pixel *ebiten.Image, x, y, width, height int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(width), float64(height))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(pixel, &op)
}

func wallIndex(wall byte) int {
	switch wall {
	case '=':
		return 1
	case '%':
		return 2
	}
	return 0
}

var (
	imageCeiling, imageFloor, imageBlack, imageWhite, imageRed *ebiten.Image
	imageWeapon, imageFlash                                    *ebiten.Image
	// Textures are cut into one-pixel columns once; SubImage allocates.
	wallColumns  [3][texSize]*ebiten.Image
	slimeColumns [texSize]*ebiten.Image
)

func solid(c color.Color) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

// hash gives repeatable texture noise without math/rand.
func hash(x, y, seed int) int {
	h := uint32(x*374761393 + y*668265263 + seed*1442695041)
	h = (h ^ h>>13) * 1274126177
	return int((h ^ h>>16) & 0xff)
}

func brick(x, y int) color.RGBA {
	row := y / 4
	if y%4 == 3 || (x+row%2*4)%8 == 7 {
		return color.RGBA{96, 88, 80, 255} // mortar
	}
	n := uint8(hash(x/8, row, 1) / 8)
	return color.RGBA{140 + n, 52 + n/2, 36, 255}
}

func stone(x, y int) color.RGBA {
	n := uint8(hash(x, y, 2) / 6)
	v := 92 + n
	if (x/4+y/4)%2 == 0 {
		v += 12
	}
	return color.RGBA{v, v, v + 8, 255}
}

func metal(x, y int) color.RGBA {
	if x%8 == 0 || y%8 == 0 {
		return color.RGBA{40, 52, 64, 255} // panel seams
	}
	if (x%8 == 2 || x%8 == 6) && (y%8 == 2 || y%8 == 6) {
		return color.RGBA{200, 210, 220, 255} // rivets
	}
	return color.RGBA{80, 104, 128, 255}
}

func texture(f func(x, y int) color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(texSize, texSize)
	for y := range texSize {
		for x := range texSize {
			img.Set(x, y, f(x, y))
		}
	}
	return img
}

// slimeImage is an original green blob with two eyes; the rest is transparent.
func slimeImage() *ebiten.Image {
	img := ebiten.NewImage(texSize, texSize)
	for y := range texSize {
		for x := range texSize {
			// A dome over a flat base: an ellipse above row 11, a rectangle below.
			dx, dy := (float64(x)-7.5)/7, (float64(y)-11)/9
			if y < 11 && dx*dx+dy*dy > 1 || y >= 11 && dx*dx > 1 {
				continue
			}
			c := color.RGBA{60, 190, 80, 255}
			if y > 12 {
				c = color.RGBA{40, 140, 60, 255}
			}
			if (x == 5 || x == 10) && (y == 7 || y == 8) {
				c = color.RGBA{250, 250, 250, 255}
			}
			if (x == 5 || x == 10) && y == 8 {
				c = color.RGBA{20, 20, 30, 255}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func weaponImage() *ebiten.Image {
	img := ebiten.NewImage(24, 20)
	for y := range 20 {
		for x := range 24 {
			switch {
			case x >= 9 && x < 15 && y < 12: // barrel
				v := uint8(70 + (x-9)*12)
				img.Set(x, y, color.RGBA{v, v, v + 10, 255})
			case y >= 10 && x >= 4 && x < 20: // body and grip
				img.Set(x, y, color.RGBA{50, 44, 40, 255})
			}
		}
	}
	return img
}

func flashImage() *ebiten.Image {
	img := ebiten.NewImage(12, 12)
	for y := range 12 {
		for x := range 12 {
			d := (x-6)*(x-6) + (y-6)*(y-6)
			switch {
			case d < 8:
				img.Set(x, y, color.RGBA{255, 250, 200, 255})
			case d < 30 && (x+y)%2 == 0:
				img.Set(x, y, color.RGBA{255, 160, 40, 255})
			}
		}
	}
	return img
}

func init() {
	imageCeiling = solid(color.RGBA{36, 36, 48, 255})
	imageFloor = solid(color.RGBA{64, 52, 40, 255})
	imageBlack = solid(color.RGBA{0, 0, 0, 255})
	imageWhite = solid(color.RGBA{255, 255, 255, 255})
	imageRed = solid(color.RGBA{255, 0, 0, 255})
	imageWeapon = weaponImage()
	imageFlash = flashImage()
	for i, f := range [...]func(x, y int) color.RGBA{brick, stone, metal} {
		tex := texture(f)
		for x := range texSize {
			wallColumns[i][x] = tex.SubImage(image.Rect(x, 0, x+1, texSize)).(*ebiten.Image)
		}
	}
	tex := slimeImage()
	for x := range texSize {
		slimeColumns[x] = tex.SubImage(image.Rect(x, 0, x+1, texSize)).(*ebiten.Image)
	}
}
