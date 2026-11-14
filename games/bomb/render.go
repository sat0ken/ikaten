package bomb

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/pixtext"
)

// The arena fills the screen between the score bar and the item bar.
const (
	arenaX = 3
	arenaY = 22
	barY   = arenaY + rows*tile // the item bar
	// A gopher's sprite is a cell wide and stands on the cell's bottom.
	gopherW, gopherH = 18, 22
)

var (
	shadow     = color.RGBA{50, 112, 60, 255}
	barColor   = color.RGBA{20, 24, 40, 255}
	barLine    = color.RGBA{50, 58, 90, 255}
	boxColor   = color.RGBA{16, 20, 36, 255}
	boxEdge    = color.RGBA{90, 110, 170, 255}
	boxLight   = color.RGBA{150, 170, 220, 255}
	textWhite  = color.RGBA{240, 244, 248, 255}
	textYellow = color.RGBA{255, 220, 60, 255}
	textRed    = color.RGBA{255, 90, 70, 255}
	textGray   = color.RGBA{150, 160, 180, 255}
	textShadow = color.RGBA{0, 0, 0, 255}
	winLit     = color.RGBA{255, 210, 60, 255}
	winRim     = color.RGBA{170, 110, 20, 255}
	winOff     = color.RGBA{56, 62, 90, 255}
)

var line pixtext.Line

// Draw redraws only the cells that changed since the last frame, and the
// cells the gophers and bombs cover now or covered then, and then draws the
// gophers and bombs on top: the screen is not cleared between frames (see
// NewGame), and most of the arena stays the same from one frame to the next.
// It draws everything when it cannot rely on the last frame: on a new screen,
// one cleared every frame, or when a box over the arena comes or goes.
func (g *Game) Draw(screen *ebiten.Image) {
	key := g.overlayKey()
	full := g.paintedOn != screen || ebiten.IsScreenClearedEveryFrame() || key != g.paintedKey
	var covered [rows][cols]bool
	g.markSprites(&covered)
	if full {
		fill(screen, barColor, 0, arenaY, ScreenWidth, rows*tile)
	}
	for r := range rows {
		for c := range cols {
			sig := g.cellSig(c, r)
			if full || sig != g.painted[r][c] || covered[r][c] || g.covered[r][c] {
				g.drawCell(screen, c, r)
				g.painted[r][c] = sig
			}
		}
	}
	g.covered, g.paintedOn, g.paintedKey = covered, screen, key
	g.drawBombs(screen)
	g.drawGophers(screen)
	// The bars change at most once a second, when the clock ticks.
	if bars := g.barsKey(); full || bars != g.paintedBars {
		g.drawBars(screen)
		g.paintedBars = bars
	}
	g.drawOverlay(screen)
}

// barsKey sums up what the bars show.
func (g *Game) barsKey() uint64 {
	k := uint64(max(g.clock, 0) / tps)
	if g.clock <= hurryTime && g.mode == modePlay {
		k |= 1 << 12
	}
	for i := range g.gophers {
		k |= uint64(g.gophers[i].wins&7) << (16 + 3*i)
	}
	p := &g.gophers[0]
	k |= uint64(p.fire&15)<<28 | uint64(p.count&15)<<32 | uint64(p.speed&15)<<36 | uint64(g.level&0xffff)<<40
	return k
}

// overlayKey changes when a box or a message over the arena appears or goes.
func (g *Game) overlayKey() int {
	key := int(g.mode) * 2
	if g.hurryShown() {
		key++
	}
	return key
}

func (g *Game) hurryShown() bool {
	return g.mode == modePlay && g.clock <= hurryTime && g.clock > hurryTime-2*tps && g.ticks/10%2 == 0
}

// markSprites marks the cells the gophers and the bombs cover.
func (g *Game) markSprites(covered *[rows][cols]bool) {
	for i := range g.gophers {
		if p := &g.gophers[i]; p.alive {
			x, y := gopherAt(p)
			markBox(covered, x, y, gopherW, gopherH)
		}
	}
	for i := range g.bombs {
		if b := &g.bombs[i]; b.active {
			x, y := bombAt(b)
			markBox(covered, x-2, y-2, tile+4, tile+4) // with room to swell
		}
	}
}

// markBox marks the cells a box of pixels touches.
func markBox(covered *[rows][cols]bool, x, y, w, h int) {
	c0, r0 := max((x-arenaX)/tile, 0), max((y-arenaY)/tile, 0)
	c1, r1 := min((x+w-1-arenaX)/tile, cols-1), min((y+h-1-arenaY)/tile, rows-1)
	for r := r0; r <= r1; r++ {
		for c := c0; c <= c1; c++ {
			covered[r][c] = true
		}
	}
}

// fireShape returns which piece of fire a cell shows and how it burns, or
// false for no fire.
func (g *Game) fireShape(c, r int) (piece, heat int, burning bool) {
	if g.flames[r][c] == 0 {
		return 0, 0, false
	}
	at := func(c, r int) bool { return c >= 0 && r >= 0 && c < cols && r < rows && g.flames[r][c] > 0 }
	l, rt, u, d := at(c-1, r), at(c+1, r), at(c, r-1), at(c, r+1)
	switch {
	case (l || rt) && (u || d), !l && !rt && !u && !d:
		piece = fireCenter
	case l && rt:
		piece = fireAcross
	case u && d:
		piece = fireAlong
	case rt:
		piece = fireEndL
	case l:
		piece = fireEndR
	case d:
		piece = fireEndU
	default:
		piece = fireEndD
	}
	heat = g.ticks / 3 % 2 // flickering
	if g.flames[r][c] < 8 {
		heat = 2 // dying down
	}
	return piece, heat, true
}

// cellSig sums up everything a cell's drawing depends on.
func (g *Game) cellSig(c, r int) uint32 {
	s := uint32(g.grid[r][c])
	if r > 0 && g.grid[r-1][c] != floor {
		s |= 1 << 2
	}
	if n := g.crumble[r][c]; n > 0 {
		s |= 1 << 3
		if n <= fireTicks/2 {
			s |= 1 << 14
		}
	}
	if g.shown[r][c] {
		s |= uint32(g.items[r][c]) << 4
	}
	if piece, heat, ok := g.fireShape(c, r); ok {
		s |= 1<<8 | uint32(piece)<<9 | uint32(heat)<<12
	}
	return s
}

func cellX(c int) int { return arenaX + c*tile }
func cellY(r int) int { return arenaY + r*tile }

// gopherAt is the top left of a gopher's sprite.
func gopherAt(p *gopher) (int, int) {
	return arenaX + p.x*tile/unit - tile/2, arenaY + p.y*tile/unit + tile/2 - gopherH
}

// bombAt is the top left of a bomb's sprite.
func bombAt(b *bomb) (int, int) {
	return arenaX + b.x*tile/unit - tile/2, arenaY + b.y*tile/unit - tile/2
}

// drawCell draws one cell of the arena: the wall, the block or the grass
// with its shade, the item, and the fire.
func (g *Game) drawCell(screen *ebiten.Image, c, r int) {
	x, y := cellX(c), cellY(r)
	switch g.grid[r][c] {
	case wall:
		drawAt(screen, g.art.wall, x, y)
		return
	case block:
		drawAt(screen, g.art.block, x, y)
		return
	}
	drawAt(screen, g.art.grass[(r+c)%2], x, y)
	// The shade of the wall or block above.
	if r > 0 && g.grid[r-1][c] != floor {
		fill(screen, shadow, x, y, tile, 4)
	}
	if n := g.crumble[r][c]; n > 0 {
		drawAt(screen, g.art.crumble[min(1, (fireTicks-n)*2/fireTicks)], x, y)
	} else if g.shown[r][c] {
		drawAt(screen, g.art.item[g.items[r][c]], x, y)
	}
	if piece, heat, ok := g.fireShape(c, r); ok {
		drawAt(screen, g.art.fire[heat][piece], x, y)
	}
}

func (g *Game) drawBombs(screen *ebiten.Image) {
	for i := range g.bombs {
		b := &g.bombs[i]
		if !b.active {
			continue
		}
		x, y := bombAt(b)
		// The fuse sparks, and the bomb swells as it burns down.
		frame := g.ticks / 6 % 2
		var op ebiten.DrawImageOptions
		if b.fuse < 50 && g.ticks/4%2 == 0 {
			op.GeoM.Translate(-tile/2, -tile/2)
			op.GeoM.Scale(1.15, 1.15)
			op.GeoM.Translate(tile/2, tile/2)
		}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(g.art.bomb[frame], &op)
	}
}

// walkFrames steps from foot to foot with a stand between.
var walkFrames = [4]int{1, 0, 2, 0}

// drawGophers draws the gophers from the back row forward.
func (g *Game) drawGophers(screen *ebiten.Image) {
	var order [players]int
	for i := range order {
		order[i] = i
	}
	for i := 1; i < players; i++ {
		for j := i; j > 0 && g.gophers[order[j]].y < g.gophers[order[j-1]].y; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	for _, i := range order {
		p := &g.gophers[i]
		if !p.alive {
			continue
		}
		x, y := gopherAt(p)
		if p.dying > 0 {
			// Blink and sink away.
			if p.dying%8 < 5 {
				var op ebiten.DrawImageOptions
				s := 1 - float64(p.dying)/dieTicks*0.7
				op.GeoM.Translate(-gopherW/2, -gopherH)
				op.GeoM.Scale(s, s)
				op.GeoM.Translate(float64(x)+gopherW/2, float64(y)+gopherH)
				screen.DrawImage(g.art.faint[i], &op)
			}
			continue
		}
		view, flip := p.facing, false
		switch p.facing {
		case right:
			view, flip = left, false
		case left:
			view, flip = left, true
		}
		frame := 0
		if p.moving {
			frame = walkFrames[g.ticks/6%4]
		}
		img := g.art.gopher[i][min(int(view), 2)][frame]
		var op ebiten.DrawImageOptions
		if flip {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(gopherW, 0)
		}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(img, &op)
	}
}

// drawBars shows the wins and the clock above the arena and the person's
// powers below.
func (g *Game) drawBars(screen *ebiten.Image) {
	fill(screen, barColor, 0, 0, ScreenWidth, arenaY)
	fill(screen, barLine, 0, arenaY-2, ScreenWidth, 1)
	for i := range g.gophers {
		x := 4 + i*44
		drawAt(screen, g.art.head[i], x, 5)
		for w := range winsToWin {
			c, rim := winOff, barLine
			if w < g.gophers[i].wins {
				c, rim = winLit, winRim
			}
			fill(screen, rim, x+21+w*8, 7, 7, 7)
			fill(screen, c, x+22+w*8, 8, 5, 5)
		}
	}
	secs := max(g.clock, 0) / tps
	c := textWhite
	if g.clock <= hurryTime && g.mode == modePlay {
		c = textRed
	}
	line.Reset().Num(secs/60).Str(":").Pad(secs%60, 2).Right(screen, ScreenWidth-4, 3, 2, c)

	fill(screen, barColor, 0, barY, ScreenWidth, ScreenHeight-barY)
	fill(screen, barLine, 0, barY+1, ScreenWidth, 1)
	p := &g.gophers[0]
	for i, n := range [3]int{p.fire, p.count, p.speed - baseSpeed + 1} {
		x := 4 + i*46
		drawAt(screen, g.art.item[fireUp+item(i)], x, barY+2)
		line.Reset().Str("x").Num(n).Draw(screen, x+21, barY+7, 1, textWhite)
	}
	line.Reset().Str("LEVEL ").Num(g.level).Right(screen, ScreenWidth-4, barY+7, 1, textGray)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	switch g.mode {
	case modeTitle:
		g.drawTitle(screen)
	case modeReady:
		box(screen, 60, 90, 120, 50)
		line.Reset().Str("ROUND ").Num(g.round).Center(screen, mid, 98, 1, textGray)
		if g.modeTicks > 45 {
			line.Reset().Str("GO!").Center(screen, mid, 112, 2, textYellow)
		} else {
			line.Reset().Str("READY").Center(screen, mid, 112, 2, textWhite)
		}
	case modePlay:
		if g.hurryShown() {
			line.Reset().Str("HURRY!").Shade(textShadow).Center(screen, mid, 98, 3, textRed)
		}
	case modeRoundOver:
		box(screen, 36, 86, 168, 56)
		if g.winner < 0 {
			line.Reset().Str("DRAW").Center(screen, mid, 104, 3, textWhite)
			break
		}
		drawScaled(screen, g.art.gopher[g.winner][down][g.modeTicks/10%2], 46, 92, 2)
		if g.winner == 0 {
			line.Reset().Str("YOU WIN!").Draw(screen, 92, 107, 2, textYellow)
		} else {
			line.Reset().Str(gopherNames[g.winner]).Str(" WINS").Draw(screen, 92, 107, 2, textWhite)
		}
	case modeMatchOver:
		box(screen, 28, 56, 184, 120)
		drawScaled(screen, g.art.gopher[g.champion][down][walkFrames[g.modeTicks/15%4]], mid-27, 62, 3)
		if g.champion == 0 {
			line.Reset().Str("CHAMPION!").Center(screen, mid, 134, 2, textYellow)
			if g.modeTicks > 60 {
				line.Reset().Str("A: LEVEL ").Num(g.level+1).Center(screen, mid, 158, 1, textWhite)
			}
		} else {
			line.Reset().Str("GAME OVER").Center(screen, mid, 134, 2, textRed)
			if g.modeTicks > 60 {
				line.Reset().Str("REACHED LEVEL ").Num(g.level).Center(screen, mid, 158, 1, textWhite)
			}
		}
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	const mid = ScreenWidth / 2
	box(screen, 12, 24, 216, 192)
	line.Reset().Str("GOPHER BOMB").Shade(boxEdge).Center(screen, mid, 34, 3, textYellow)
	drawScaled(screen, g.art.gopher[0][down][walkFrames[g.modeTicks/20%4]], mid-27, 66, 3)
	line.Reset().Str("A: BOMB   B: KICK").Center(screen, mid, 142, 1, textWhite)
	line.Reset().Str("LAST GOPHER STANDING WINS").Center(screen, mid, 156, 1, textGray)
	if g.modeTicks/30%2 == 0 {
		line.Reset().Str("PRESS A").Center(screen, mid, 174, 2, textYellow)
	}
	line.Reset().Str("GOPHER BY RENEE FRENCH").Center(screen, mid, 202, 1, textGray)
}

// box is a dark panel with a two-tone frame, drawn side by side rather than
// one over another.
func box(screen *ebiten.Image, x, y, w, h int) {
	fill(screen, boxLight, x-2, y-2, w+3, 1)
	fill(screen, boxEdge, x+w+1, y-2, 1, 1)
	fill(screen, boxEdge, x-2, y-1, w+4, 1)
	fill(screen, boxEdge, x-2, y, 2, h)
	fill(screen, boxEdge, x+w, y, 2, h)
	fill(screen, boxEdge, x-2, y+h, w+4, 2)
	fill(screen, boxColor, x, y, w, h)
}

// paints holds one-pixel images of the colors in use: a pixel of the right
// color is copied, where a tinted white pixel would be shaded pixel by pixel.
var paints [32]struct {
	img  *ebiten.Image
	c    color.RGBA
	used int
}

var paintTick int

func init() {
	for i := range paints {
		paints[i].img = ebiten.NewImage(1, 1)
	}
}

// paint returns a one-pixel image of c.
func paint(c color.RGBA) *ebiten.Image {
	paintTick++
	h := (int(c.R)*31 + int(c.G)*7 + int(c.B)) % len(paints)
	oldest := h
	for i := range 8 {
		p := &paints[(h+i)%len(paints)]
		if p.used > 0 && p.c == c {
			p.used = paintTick
			return p.img
		}
		if p.used < paints[oldest].used {
			oldest = (h + i) % len(paints)
		}
	}
	p := &paints[oldest]
	p.c, p.used = c, paintTick
	p.img.Fill(c) // Fill marks the image opaque, which draws faster than Set
	return p.img
}

func fill(screen *ebiten.Image, c color.RGBA, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(paint(c), &op)
}

func drawAt(screen, img *ebiten.Image, x, y int) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

func drawScaled(screen, img *ebiten.Image, x, y int, s float64) {
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}
