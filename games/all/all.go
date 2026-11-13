// Package all is a launcher: a menu of every ikaten game. Holding A and B (Z and
// X on a keyboard) together for a second returns to the menu. It uses only the
// Ebitengine API, so it also builds with Ebitengine.
package all

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/air"
	"github.com/sat0ken/ikaten/games/blocks"
	"github.com/sat0ken/ikaten/games/bomb"
	"github.com/sat0ken/ikaten/games/bounce"
	"github.com/sat0ken/ikaten/games/dungeon"
	"github.com/sat0ken/ikaten/games/fish"
	"github.com/sat0ken/ikaten/games/flappy"
	"github.com/sat0ken/ikaten/games/jump"
	"github.com/sat0ken/ikaten/games/kart"
	"github.com/sat0ken/ikaten/games/pixtext"
	"github.com/sat0ken/ikaten/games/shooter"
	"github.com/sat0ken/ikaten/games/snake"
	"github.com/sat0ken/ikaten/games/tank"
	"github.com/sat0ken/ikaten/games/train"
	"github.com/sat0ken/ikaten/games/uboat"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 240
	holdTicks    = 60 // A and B held this long return to the menu
	rowHeight    = 14
	listTop      = 34
	shownRows    = 11 // the list scrolls to keep the cursor in these rows
)

type entry struct {
	name, note string
	start      func() ebiten.Game
}

var entries = [...]entry{
	{"BOUNCE", "SPRITE DEMO", func() ebiten.Game { return bounce.New(ScreenWidth, ScreenHeight, 30) }},
	{"SNAKE", "EAT AND GROW", func() ebiten.Game { return snake.NewGame() }},
	{"FLAPPY", "FLY THROUGH PIPES", func() ebiten.Game { return flappy.NewGame() }},
	{"BLOCKS", "FALLING BLOCKS", func() ebiten.Game { return &blocks.Game{} }},
	{"DUNGEON", "3D MAZE SHOOTER", func() ebiten.Game { return dungeon.NewGame() }},
	{"SHOOTER", "VERTICAL SHOOTER", func() ebiten.Game { return shooter.NewGame() }},
	{"KART", "KART RACING", func() ebiten.Game { return kart.NewGame() }},
	{"U-BOAT", "SUBMARINE ATTACK", func() ebiten.Game { return uboat.NewGame() }},
	{"TANK", "TANK BATTLE", func() ebiten.Game { return tank.NewGame() }},
	{"TRAIN", "TRAIN DRIVING", func() ebiten.Game { return train.NewGame() }},
	{"AIR", "CARRIER AIR WAR", func() ebiten.Game { return air.NewGame() }},
	{"FISHING", "CAST, HOOK AND REEL", func() ebiten.Game { return fish.NewGame() }},
	{"GOPHER BOMB", "BOMB MAZE BATTLE", func() ebiten.Game { return bomb.NewGame() }},
	{"GOPHER JUMP", "RUN, JUMP AND DASH", func() ebiten.Game { return jump.NewGame() }},
}

type Game struct {
	cursor int
	game   ebiten.Game // the running game, or nil on the menu
	fresh  bool        // the game has not run Update yet
	held   int         // ticks A and B have been held together
	ticks  int
}

func NewGame() *Game { return &Game{} }

func (g *Game) Update() error {
	g.ticks++
	if g.game != nil {
		if ebiten.IsKeyPressed(ebiten.KeyZ) && ebiten.IsKeyPressed(ebiten.KeyX) {
			g.held++
		} else {
			g.held = 0
		}
		if g.held >= holdTicks {
			g.game, g.held = nil, 0
			// A game may have kept its screen between frames; the menu clears it.
			ebiten.SetScreenClearedEveryFrame(true)
			return nil
		}
		g.fresh = false
		return g.game.Update()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.cursor = (g.cursor + len(entries) - 1) % len(entries)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.cursor = (g.cursor + 1) % len(entries)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) || inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		// The game starts on the next tick, so it does not see this press as new.
		ebiten.SetScreenClearedEveryFrame(true)
		g.game, g.fresh = entries[g.cursor].start(), true
	}
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.game != nil {
		return g.game.Layout(outsideWidth, outsideHeight)
	}
	return ScreenWidth, ScreenHeight
}

var (
	menuBackground = color.RGBA{16, 24, 48, 255}
	imageBar       = pixel(color.RGBA{50, 90, 170, 255})
	imageLine      = pixel(color.RGBA{90, 120, 180, 255})
	textWhite      = color.RGBA{240, 244, 250, 255}
	textYellow     = color.RGBA{255, 220, 60, 255}
	textGray       = color.RGBA{150, 166, 200, 255}
	textShadow     = color.RGBA{50, 90, 170, 255}
)

var line pixtext.Line

func pixel(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.game != nil {
		// Like Ebitengine, a game runs Update before its first Draw; some
		// games set themselves up there.
		if g.fresh {
			screen.Clear()
			return
		}
		g.game.Draw(screen)
		if g.held > 0 {
			// A bar fills while A and B are held.
			fillRect(screen, imageBar, 0, 0, screen.Bounds().Dx()*g.held/holdTicks, 3)
		}
		return
	}
	const mid = ScreenWidth / 2
	screen.Fill(menuBackground)
	line.Reset().Str("IKATEN GAMES").Shade(textShadow).Center(screen, mid, 8, 2, textYellow)
	fillRect(screen, imageLine, 8, listTop-5, ScreenWidth-16, 1)
	// Scroll so that the cursor stays in view, a row from the edge when it can.
	top := max(0, min(g.cursor-shownRows/2, len(entries)-shownRows))
	for i := top; i < min(top+shownRows, len(entries)); i++ {
		e := &entries[i]
		y := listTop + (i-top)*rowHeight
		c := textGray
		if i == g.cursor {
			fillRect(screen, imageBar, 8, y, ScreenWidth-16, rowHeight-1)
			if g.ticks/20%2 == 0 {
				line.Reset().Str(">").Draw(screen, 16, y+3, 1, textYellow)
			}
			c = textWhite
		}
		line.Reset().Str(e.name).Draw(screen, 30, y+3, 1, c)
	}
	bottom := listTop + shownRows*rowHeight + 3
	fillRect(screen, imageLine, 8, bottom, ScreenWidth-16, 1)
	// Marks show that the list goes on.
	if top > 0 {
		arrow(screen, ScreenWidth-16, listTop-12, -1)
	}
	if top+shownRows < len(entries) {
		arrow(screen, ScreenWidth-16, bottom+4, 1)
	}
	line.Reset().Str(entries[g.cursor].note).Center(screen, mid, bottom+12, 1, textWhite)
	line.Reset().Str("HOLD A+B: MENU").Center(screen, mid, bottom+28, 1, textGray)
}

// arrow draws a small triangle pointing up (dir -1) or down (dir 1).
func arrow(screen *ebiten.Image, x, y, dir int) {
	for i := range 4 {
		r := i // widening downward: pointing up
		if dir > 0 {
			r = 3 - i
		}
		fillRect(screen, imageLine, x-r, y+i, 2*r+1, 1)
	}
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
