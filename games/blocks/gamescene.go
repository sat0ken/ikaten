// Copyright 2014 Hajime Hoshi
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

// This file is modified from Ebitengine v2.10.5 examples/blocks/blocks for ikaten.
// The changes are listed in docs/porting.md.

package blocks

import (
	"image/color"
	"math/rand/v2"
	"strconv"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/inpututil"
)

// imageGameBG is the gray background with the translucent windows baked in
// (see assets/gen.go). Images of the screen size do not fit in the RAM, and
// binary-alpha images cannot hold translucent windows.
var (
	imageGameBG *ebiten.Image
	imageBlack  *ebiten.Image
)

func fieldWindowPosition() (x, y int) {
	return 20, 20
}

func nextWindowLabelPosition() (x, y int) {
	x, y = fieldWindowPosition()
	return x + fieldWidth + 2*blockWidth, y
}

func nextWindowPosition() (x, y int) {
	x, y = nextWindowLabelPosition()
	return x, y + blockHeight
}

func textBoxWidth() int {
	x, _ := nextWindowPosition()
	return ScreenWidth - 2*blockWidth - x
}

func scoreTextBoxPosition() (x, y int) {
	x, y = nextWindowPosition()
	return x, y + 6*blockHeight
}

func levelTextBoxPosition() (x, y int) {
	x, y = scoreTextBoxPosition()
	return x, y + 4*blockHeight
}

func linesTextBoxPosition() (x, y int) {
	x, y = levelTextBoxPosition()
	return x, y + 4*blockHeight
}

func drawWindowLabels(r *ebiten.Image) {
	// Windows: Next
	x, y := nextWindowLabelPosition()
	drawTextWithShadow(r, "NEXT", x, y, 1, fontColor, alignStart, alignStart)

	// Windows: Score
	x, y = scoreTextBoxPosition()
	drawTextBox(r, "SCORE", x, y, textBoxWidth())

	// Windows: Level
	x, y = levelTextBoxPosition()
	drawTextBox(r, "LEVEL", x, y, textBoxWidth())

	// Windows: Lines
	x, y = linesTextBoxPosition()
	drawTextBox(r, "LINES", x, y, textBoxWidth())
}

func drawGameover(r *ebiten.Image) {
	fillTranslucent(r, 0, 0, ScreenWidth, ScreenHeight, 0x80/255.0)
	y := (ScreenHeight - blockHeight) / 2
	drawTextWithShadow(r, "GAME OVER\n\nPRESS SPACE", ScreenWidth/2, y, 1, color.White, alignCenter, alignStart)
}

// fillTranslucent darkens a rectangle by drawing black with the given alpha.
func fillTranslucent(r *ebiten.Image, x, y, width, height int, alpha float32) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(width), float64(height))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleAlpha(alpha)
	r.DrawImage(imageBlack, op)
}

var fontColor = color.RGBA{0x40, 0x40, 0xff, 0xff}

func drawTextBox(r *ebiten.Image, label string, x, y, width int) {
	drawTextWithShadow(r, label, x, y, 1, fontColor, alignStart, alignStart)
}

func drawTextBoxContent(r *ebiten.Image, content string, x, y, width int) {
	y += blockHeight
	drawTextWithShadow(r, content, x+width-2*blockHeight/4, y+2*blockHeight/2, 1, color.White, alignEnd, alignCenter)
}

type GameScene struct {
	field              *Field
	currentPiece       *Piece
	currentPieceX      int
	currentPieceY      int
	currentPieceYCarry int
	currentPieceAngle  Angle
	nextPiece          *Piece
	landingCount       int
	score              int
	lines              int
	gameover           bool
}

func NewGameScene() *GameScene {
	return &GameScene{
		field: &Field{},
	}
}

func (s *GameScene) drawBackground(r *ebiten.Image) {
	r.DrawImage(imageGameBG, nil)
}

const (
	fieldWidth  = blockWidth * fieldBlockCountX
	fieldHeight = blockHeight * fieldBlockCountY
)

func (s *GameScene) choosePiece() *Piece {
	num := int(BlockTypeMax)
	blockType := BlockType(rand.IntN(num) + 1)
	return Pieces[blockType]
}

func (s *GameScene) initCurrentPiece(piece *Piece) {
	s.currentPiece = piece
	x, y := s.currentPiece.InitialPosition()
	s.currentPieceX = x
	s.currentPieceY = y
	s.currentPieceYCarry = 0
	s.currentPieceAngle = Angle0
}

func (s *GameScene) level() int {
	return s.lines / 10
}

func (s *GameScene) addScore(lines int) {
	var base int
	switch lines {
	case 1:
		base = 100
	case 2:
		base = 300
	case 3:
		base = 600
	case 4:
		base = 1000
	default:
		panic("not reach")
	}
	s.score += (s.level() + 1) * base
}

func (s *GameScene) Update(state *GameState) error {
	s.field.Update()

	if s.gameover {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			anyGamepadVirtualButtonJustPressed(state.Input) {
			state.SceneManager.GoTo(&TitleScene{})
		}
		return nil
	}

	maxLandingCount := ebiten.TPS()
	if s.currentPiece == nil {
		s.initCurrentPiece(s.choosePiece())
	}
	if s.nextPiece == nil {
		s.nextPiece = s.choosePiece()
	}

	var moved bool
	piece := s.currentPiece
	angle := s.currentPieceAngle

	// Move piece by user input.
	if !s.field.IsFlushAnimating() {
		piece := s.currentPiece
		x := s.currentPieceX
		y := s.currentPieceY
		if state.Input.IsRotateRightJustPressed() {
			s.currentPieceAngle = s.field.RotatePieceRight(piece, x, y, angle)
			moved = angle != s.currentPieceAngle
		} else if state.Input.IsRotateLeftJustPressed() {
			s.currentPieceAngle = s.field.RotatePieceLeft(piece, x, y, angle)
			moved = angle != s.currentPieceAngle
		} else if l := state.Input.StateForLeft(); l == 1 || (10 <= l && l%2 == 0) {
			s.currentPieceX = s.field.MovePieceToLeft(piece, x, y, angle)
			moved = x != s.currentPieceX
		} else if r := state.Input.StateForRight(); r == 1 || (10 <= r && r%2 == 0) {
			s.currentPieceX = s.field.MovePieceToRight(piece, x, y, angle)
			moved = y != s.currentPieceX
		} else if d := state.Input.StateForDown(); (d-1)%2 == 0 {
			s.currentPieceY = s.field.DropPiece(piece, x, y, angle)
			moved = y != s.currentPieceY
			if moved {
				s.score++
			}
		}
	}

	// Drop the current piece with gravity.
	if !s.field.IsFlushAnimating() {
		angle := s.currentPieceAngle
		s.currentPieceYCarry += 2*s.level() + 1
		const maxCarry = 60
		for maxCarry <= s.currentPieceYCarry {
			s.currentPieceYCarry -= maxCarry
			s.currentPieceY = s.field.DropPiece(piece, s.currentPieceX, s.currentPieceY, angle)
		}
	}

	if !s.field.IsFlushAnimating() && !s.field.PieceDroppable(piece, s.currentPieceX, s.currentPieceY, angle) {
		if 0 < state.Input.StateForDown() {
			s.landingCount += 10
		} else {
			s.landingCount++
		}
		if maxLandingCount <= s.landingCount {
			s.field.AbsorbPiece(piece, s.currentPieceX, s.currentPieceY, angle)
			if s.field.IsFlushAnimating() {
				s.field.SetEndFlushAnimating(func(lines int) {
					s.lines += lines
					if 0 < lines {
						s.addScore(lines)
					}
					s.goNextPiece()
				})
			} else {
				s.goNextPiece()
			}

		}
	}
	return nil
}

func (s *GameScene) goNextPiece() {
	s.initCurrentPiece(s.nextPiece)
	s.nextPiece = s.choosePiece()
	s.landingCount = 0
	if s.currentPiece.collides(s.field, s.currentPieceX, s.currentPieceY, s.currentPieceAngle) {
		s.gameover = true
	}
}

func (s *GameScene) Draw(r *ebiten.Image) {
	s.drawBackground(r)

	drawWindowLabels(r)

	// Draw score
	x, y := scoreTextBoxPosition()
	drawTextBoxContent(r, strconv.Itoa(s.score), x, y, textBoxWidth())

	// Draw level
	x, y = levelTextBoxPosition()
	drawTextBoxContent(r, strconv.Itoa(s.level()), x, y, textBoxWidth())

	// Draw lines
	x, y = linesTextBoxPosition()
	drawTextBoxContent(r, strconv.Itoa(s.lines), x, y, textBoxWidth())

	// Draw blocks
	fieldX, fieldY := fieldWindowPosition()
	s.field.Draw(r, fieldX, fieldY)
	if s.currentPiece != nil && !s.field.IsFlushAnimating() {
		x := fieldX + s.currentPieceX*blockWidth
		y := fieldY + s.currentPieceY*blockHeight
		s.currentPiece.Draw(r, x, y, s.currentPieceAngle)
	}
	if s.nextPiece != nil {
		// TODO: Make functions to get these values.
		x := fieldX + fieldWidth + blockWidth*2
		y := fieldY + blockHeight
		s.nextPiece.DrawAtCenter(r, x, y, blockWidth*5, blockHeight*5, 0)
	}

	if s.gameover {
		drawGameover(r)
	}
}
