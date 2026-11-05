// Package all is a launcher: a menu of every ikaten game. Holding A and B (Z and
// X on a keyboard) together for a second returns to the menu. It uses only the
// Ebitengine API, so it also builds with Ebitengine.
package all

import (
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/games/blocks"
	"github.com/sat0ken/ikaten/games/bounce"
	"github.com/sat0ken/ikaten/games/dungeon"
	"github.com/sat0ken/ikaten/games/flappy"
	"github.com/sat0ken/ikaten/games/kart"
	"github.com/sat0ken/ikaten/games/shooter"
	"github.com/sat0ken/ikaten/games/snake"
	"github.com/sat0ken/ikaten/games/tank"
	"github.com/sat0ken/ikaten/games/uboat"
	"github.com/sat0ken/ikaten/inpututil"
)

const (
	ScreenWidth  = 120
	ScreenHeight = 120
	holdTicks    = 60 // A and B held this long return to the menu
	rowHeight    = 9
	listTop      = 15
)

type entry struct {
	name, note string
	start      func() ebiten.Game
}

var entries = [...]entry{
	{"BOUNCE", "SPRITE DEMO", func() ebiten.Game { return bounce.New(ScreenWidth, ScreenHeight, 15) }},
	{"SNAKE", "EAT AND GROW", func() ebiten.Game { return snake.NewGame() }},
	{"FLAPPY", "FLY THROUGH PIPES", func() ebiten.Game { return flappy.NewGame() }},
	{"BLOCKS", "FALLING BLOCKS", func() ebiten.Game { return &blocks.Game{} }},
	{"DUNGEON", "3D MAZE SHOOTER", func() ebiten.Game { return dungeon.NewGame() }},
	{"SHOOTER", "VERTICAL SHOOTER", func() ebiten.Game { return shooter.NewGame() }},
	{"KART", "KART RACING", func() ebiten.Game { return kart.NewGame() }},
	{"U-BOAT", "SUBMARINE ATTACK", func() ebiten.Game { return uboat.NewGame() }},
	{"TANK", "TANK BATTLE", func() ebiten.Game { return tank.NewGame() }},
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
)

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
			fillRect(screen, imageBar, 0, 0, screen.Bounds().Dx()*g.held/holdTicks, 2)
		}
		return
	}
	screen.Fill(menuBackground)
	ebitenutil.DebugPrintAt(screen, "IKATEN GAMES", 24, -1)
	fillRect(screen, imageLine, 4, listTop-3, ScreenWidth-8, 1)
	for i, e := range entries {
		y := listTop + i*rowHeight
		if i == g.cursor {
			fillRect(screen, imageBar, 4, y, ScreenWidth-8, rowHeight)
			if g.ticks/20%2 == 0 {
				ebitenutil.DebugPrintAt(screen, ">", 8, y-2)
			}
		}
		ebitenutil.DebugPrintAt(screen, e.name, 18, y-2)
	}
	fillRect(screen, imageLine, 4, listTop+len(entries)*rowHeight+2, ScreenWidth-8, 1)
	note := entries[g.cursor].note
	ebitenutil.DebugPrintAt(screen, note, ScreenWidth/2-len(note)*3, 98)
	ebitenutil.DebugPrintAt(screen, "HOLD A+B: MENU", 18, 108)
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
