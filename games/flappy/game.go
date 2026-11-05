// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file is modified from Ebitengine v2.10.5 examples/flappy/main.go for ikaten.
// The changes are listed in docs/porting.md.

package flappy

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/inpututil"
)

func floorDiv(x, y int) int {
	d := x / y
	if d*y == x || x >= 0 {
		return d
	}
	return d - 1
}

func floorMod(x, y int) int {
	return x - floorDiv(x, y)*y
}

const (
	screenWidth      = 240
	screenHeight     = 240
	tileSize         = 16
	titleFontSize    = fontSize * 1.5
	fontSize         = 12
	smallFontSize    = fontSize / 2
	pipeWidth        = tileSize * 2
	pipeStartOffsetX = 8
	pipeIntervalX    = 8
	pipeGapY         = 5
)

var (
	gopherImage *ebiten.Image
	groundTile  *ebiten.Image
	pipeCapTile *ebiten.Image
	pipeTile    *ebiten.Image
)

// The tiles are cut once because SubImage allocates and Draw runs every frame.
func init() {
	const (
		pipeTileSrcX = 64
		pipeTileSrcY = 96
	)
	var tilesImage *ebiten.Image
	gopherImage, tilesImage = loadImages()
	groundTile = tilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image)
	pipeCapTile = tilesImage.SubImage(image.Rect(pipeTileSrcX, pipeTileSrcY, pipeTileSrcX+pipeWidth, pipeTileSrcY+tileSize)).(*ebiten.Image)
	pipeTile = tilesImage.SubImage(image.Rect(pipeTileSrcX, pipeTileSrcY+tileSize, pipeTileSrcX+pipeWidth, pipeTileSrcY+tileSize*2)).(*ebiten.Image)
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	mode Mode

	// The gopher's position
	x16  int
	y16  int
	vy16 int

	// Camera
	cameraX int
	cameraY int

	// Pipes
	pipeTileYs []int

	gameoverCount int

	touchIDs   []ebiten.TouchID
	gamepadIDs []ebiten.GamepadID
}

func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.x16 = 0
	g.y16 = 50 * 16
	g.cameraX = -90
	g.cameraY = 0
	g.pipeTileYs = make([]int, 256)
	for i := range g.pipeTileYs {
		g.pipeTileYs[i] = rand.IntN(6) + 2
	}
}

func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		return true
	}
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, g := range g.gamepadIDs {
		if ebiten.IsStandardGamepadLayoutAvailable(g) {
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightBottom) {
				return true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightRight) {
				return true
			}
		} else {
			// The button 0/1 might not be A/B buttons.
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton0) {
				return true
			}
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton1) {
				return true
			}
		}
	}
	return false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if g.isKeyJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:
		g.x16 += 16
		g.cameraX += 1
		if g.isKeyJustPressed() {
			g.vy16 = -48
		}
		g.y16 += g.vy16

		// Gravity
		g.vy16 += 2
		if g.vy16 > 48 {
			g.vy16 = 48
		}

		if g.hit() {
			g.mode = ModeGameOver
			g.gameoverCount = 30
		}
	case ModeGameOver:
		if g.gameoverCount > 0 {
			g.gameoverCount--
		}
		if g.gameoverCount == 0 && g.isKeyJustPressed() {
			g.init()
			g.mode = ModeTitle
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	g.drawTiles(screen)
	if g.mode != ModeTitle {
		g.drawGopher(screen)
	}

	var titleTexts string
	var texts string
	switch g.mode {
	case ModeTitle:
		titleTexts = "FLAPPY GOPHER"
		texts = "\n\n\n\n\n\nPRESS SPACE KEY\n\nOR A/B BUTTON\n\nOR TOUCH SCREEN"
	case ModeGameOver:
		texts = "\nGAME OVER!"
	}

	drawText(screen, titleTexts, titleFontSize, screenWidth/2, 3*titleFontSize, alignCenter)
	drawText(screen, texts, fontSize, screenWidth/2, 3*titleFontSize, alignCenter)

	if g.mode == ModeTitle {
		const msg = "Go Gopher by Renee French is\nlicenced under CC BY 3.0."
		drawText(screen, msg, smallFontSize, screenWidth/2, screenHeight-smallFontSize/2-2*smallFontSize, alignCenter)
	}

	drawText(screen, fmt.Sprintf("%04d", g.score()), fontSize, screenWidth, 0, alignEnd)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) pipeAt(tileX int) (tileY int, ok bool) {
	if (tileX - pipeStartOffsetX) <= 0 {
		return 0, false
	}
	if floorMod(tileX-pipeStartOffsetX, pipeIntervalX) != 0 {
		return 0, false
	}
	idx := floorDiv(tileX-pipeStartOffsetX, pipeIntervalX)
	return g.pipeTileYs[idx%len(g.pipeTileYs)], true
}

func (g *Game) score() int {
	x := floorDiv(g.x16, 16) / tileSize
	if (x - pipeStartOffsetX) <= 0 {
		return 0
	}
	return floorDiv(x-pipeStartOffsetX, pipeIntervalX)
}

func (g *Game) hit() bool {
	if g.mode != ModeGame {
		return false
	}
	const (
		gopherWidth  = 15
		gopherHeight = 30
	)
	w, h := gopherImage.Bounds().Dx(), gopherImage.Bounds().Dy()
	x0 := floorDiv(g.x16, 16) + (w-gopherWidth)/2
	y0 := floorDiv(g.y16, 16) + (h-gopherHeight)/2
	x1 := x0 + gopherWidth
	y1 := y0 + gopherHeight
	if y0 < -tileSize*4 {
		return true
	}
	if y1 >= screenHeight-tileSize {
		return true
	}
	xMin := floorDiv(x0-pipeWidth, tileSize)
	xMax := floorDiv(x0+gopherWidth, tileSize)
	for x := xMin; x <= xMax; x++ {
		y, ok := g.pipeAt(x)
		if !ok {
			continue
		}
		if x0 >= x*tileSize+pipeWidth {
			continue
		}
		if x1 < x*tileSize {
			continue
		}
		if y0 < y*tileSize {
			return true
		}
		if y1 >= (y+pipeGapY)*tileSize {
			return true
		}
	}
	return false
}

func (g *Game) drawTiles(screen *ebiten.Image) {
	const (
		nx = screenWidth / tileSize
		ny = screenHeight / tileSize
	)

	op := &ebiten.DrawImageOptions{}
	for i := -2; i < nx+1; i++ {
		// ground
		op.GeoM.Reset()
		op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
			float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
		screen.DrawImage(groundTile, op)

		// pipe
		if tileY, ok := g.pipeAt(floorDiv(g.cameraX, tileSize) + i); ok {
			for j := range tileY {
				op.GeoM.Reset()
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
				op.GeoM.Translate(0, tileSize)
				if j == tileY-1 {
					screen.DrawImage(pipeCapTile, op)
				} else {
					screen.DrawImage(pipeTile, op)
				}
			}
			for j := tileY + pipeGapY; j < screenHeight/tileSize-1; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
				if j == tileY+pipeGapY {
					screen.DrawImage(pipeCapTile, op)
				} else {
					screen.DrawImage(pipeTile, op)
				}
			}
		}
	}
}

func (g *Game) drawGopher(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	w, h := gopherImage.Bounds().Dx(), gopherImage.Bounds().Dy()
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	op.GeoM.Rotate(float64(g.vy16) / 48.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.x16/16.0)-float64(g.cameraX), float64(g.y16/16.0)-float64(g.cameraY))
	screen.DrawImage(gopherImage, op)
}
