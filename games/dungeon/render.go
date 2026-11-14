package dungeon

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

var (
	textWhite  = color.RGBA{240, 240, 240, 255}
	textGray   = color.RGBA{150, 150, 160, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 80, 70, 255}
	textGreen  = color.RGBA{110, 230, 130, 255}
	textShadow = color.RGBA{0, 0, 0, 255}
	barBack    = color.RGBA{50, 50, 60, 255}
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
	// The ceiling and the floor, darker toward the horizon.
	const half = viewHeight / 2
	mark(0)
	for i := range bands {
		fillRect(screen, imageCeiling[i], 0, i*half/bands, ScreenWidth, half/bands+1)
		fillRect(screen, imageFloor[i], 0, viewHeight-(i+1)*half/bands, ScreenWidth, half/bands+1)
	}
	mark(1)
	g.drawWalls(screen)
	mark(2)
	g.drawSlimes(screen)
	mark(3)
	if g.mode == modePlay {
		g.drawWeapon(screen)
	}
	if g.damage > 0 {
		// A red edge round the view: a see-through red over it all would be
		// worked out pixel by pixel.
		w := 2 + g.damage
		fillRect(screen, imageRed, 0, 0, ScreenWidth, w)
		fillRect(screen, imageRed, 0, viewHeight-w, ScreenWidth, w)
		fillRect(screen, imageRed, 0, w, w, viewHeight-2*w)
		fillRect(screen, imageRed, ScreenWidth-w, w, w, viewHeight-2*w)
	}
	mark(4)
	g.drawStatus(screen)
	mark(5)
}

// colW is the width of a wall's column in pixels: a ray is cast for every
// colW columns of the screen. On the badge a column costs some 50 us
// whatever its width, and 240 of them took 11 ms.
const colW = 2

func (g *Game) drawWalls(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	for x := 0; x < ScreenWidth; x += colW {
		cameraX := 2*(float32(x)+colW/2.0)/ScreenWidth - 1
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
		for i := range colW {
			g.zbuf[x+i] = dist
		}
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
		op.GeoM.Scale(colW, float64(height/texSize))
		op.GeoM.Translate(float64(x), float64(viewHeight/2-height/2))
		op.ColorScale.Reset()
		op.ColorScale.Scale(shade, shade, shade, 1)
		screen.DrawImage(g.art.wallColumns[wallIndex(wall)][texX], &op)
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
	for i := 1; i < n; i++ {
		for j := i; j > 0 && g.distance(g.order[j-1]) < g.distance(g.order[j]); j-- {
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
		for x := max(0, int(left)) &^ (colW - 1); x < min(ScreenWidth, int(left+size)+1); x += colW {
			if depth >= g.zbuf[x] {
				continue
			}
			texX := int((float32(x) + colW/2.0 - left) * texSize / size)
			if texX < 0 || texX >= texSize {
				continue
			}
			op.GeoM.Reset()
			op.GeoM.Scale(colW, float64(size/texSize))
			op.GeoM.Translate(float64(x), float64(top))
			op.ColorScale.Reset()
			if s.hurt > 0 {
				op.ColorScale.Scale(3, 0.5, 0.5, 1) // flash red; the slime is mostly green
			} else {
				op.ColorScale.Scale(shade, shade, shade, 1)
			}
			screen.DrawImage(g.art.slimeColumns[texX], &op)
		}
	}
}

// distance is the square of slime i's distance from the player.
func (g *Game) distance(i int) float32 {
	dx, dy := g.slimes[i].x-g.px, g.slimes[i].y-g.py
	return dx*dx + dy*dy
}

func (g *Game) drawWeapon(screen *ebiten.Image) {
	var op ebiten.DrawImageOptions
	bob := float64(0)
	if g.flash > 0 {
		bob = 4
	}
	op.GeoM.Translate(ScreenWidth/2-24, viewHeight-40+bob)
	screen.DrawImage(g.art.weapon, &op)
	if g.flash > 0 {
		op.GeoM.Reset()
		op.GeoM.Translate(ScreenWidth/2-12, viewHeight-62)
		screen.DrawImage(g.art.flash, &op)
	}
	// The crosshair, with a dark outline.
	const cx, cy = ScreenWidth / 2, viewHeight / 2
	fillRect(screen, imageBlack, cx-1, cy-6, 3, 13)
	fillRect(screen, imageBlack, cx-6, cy-1, 13, 3)
	fillRect(screen, imageWhite, cx, cy-5, 1, 11)
	fillRect(screen, imageWhite, cx-5, cy, 11, 1)
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	fillRect(screen, imageBlack, 0, viewHeight, ScreenWidth, ScreenHeight-viewHeight)
	y := viewHeight + 8
	hp := max(g.hp, 0)
	text.Reset().Str("HP").Draw(screen, 8, y+2, 1, textGray)
	c := textGreen
	if hp <= 30 {
		c = textRed
	}
	fillRect(screen, solidOf(barBack), 24, y, 100, 11)
	fillRect(screen, solidOf(c), 24, y, hp, 11)
	text.Reset().Num(hp).Shade(textShadow).Draw(screen, 28, y+2, 1, textWhite)
	text.Reset().Str("SLIMES ").Num(g.kills).Str("/").Num(g.slimeCount).Right(screen, ScreenWidth-8, y+2, 1, textWhite)
	const mid = ScreenWidth / 2
	switch g.mode {
	case modeTitle:
		text.Reset().Str("IKATEN").Shade(textShadow).Center(screen, mid, 44, 3, textYellow)
		text.Reset().Str("DUNGEON").Shade(textShadow).Center(screen, mid, 72, 3, textYellow)
		text.Reset().Str("DEFEAT ALL SLIMES").Shade(textShadow).Center(screen, mid, 112, 1, textWhite)
		text.Reset().Str("ARROWS: MOVE   A: SHOOT").Shade(textShadow).Center(screen, mid, 128, 1, textWhite)
		if g.ticks/30%2 == 0 {
			text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 152, 2, textGreen)
		}
	case modeDead:
		text.Reset().Str("YOU DIED").Shade(textShadow).Center(screen, mid, 80, 3, textRed)
		text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 120, 2, textWhite)
	case modeClear:
		text.Reset().Str("ALL CLEAR!").Shade(textShadow).Center(screen, mid, 80, 3, textYellow)
		text.Reset().Str("PRESS A").Shade(textShadow).Center(screen, mid, 120, 2, textWhite)
	}
}

// solids holds a one-pixel image for each color the status bar fills with.
var solids [4]struct {
	c   color.RGBA
	img *ebiten.Image
}

// solidOf returns a one-pixel image of c, made once.
func solidOf(c color.RGBA) *ebiten.Image {
	for i := range solids {
		s := &solids[i]
		if s.img == nil {
			s.c, s.img = c, solid(c)
			return s.img
		}
		if s.c == c {
			return s.img
		}
	}
	return imageWhite
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// fillRect draws a 1x1 image scaled to a rectangle; this avoids per-pixel Set.
func fillRect(screen, pixel *ebiten.Image, x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
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
