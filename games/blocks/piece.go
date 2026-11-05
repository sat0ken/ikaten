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
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

var (
	imageBlocks     *ebiten.Image
	imageWhiteBlock *ebiten.Image
	// The blocks are cut once because SubImage allocates and Draw runs every frame.
	blockImages [BlockTypeMax]*ebiten.Image
)

func init() {
	imageBackground, imageBlocks, imageGameBG = loadImages()
	for i := range blockImages {
		srcX := i * blockWidth
		blockImages[i] = imageBlocks.SubImage(image.Rect(srcX, 0, srcX+blockWidth, blockHeight)).(*ebiten.Image)
	}
	imageWhiteBlock = ebiten.NewImage(blockWidth, blockHeight)
	imageWhiteBlock.Fill(color.White)
	imageBlack = ebiten.NewImage(1, 1)
	imageBlack.Fill(color.Black)
}

type Angle int

const (
	Angle0 Angle = iota
	Angle90
	Angle180
	Angle270
)

func (a Angle) RotateRight() Angle {
	if a == Angle270 {
		return Angle0
	}
	return a + 1
}

func (a Angle) RotateLeft() Angle {
	if a == Angle0 {
		return Angle270
	}
	return a - 1
}

type BlockType int

const (
	BlockTypeNone BlockType = iota
	BlockType1
	BlockType2
	BlockType3
	BlockType4
	BlockType5
	BlockType6
	BlockType7
	BlockTypeMax = BlockType7
)

type Piece struct {
	blockType BlockType
	blocks    [][]bool
}

func transpose(bs [][]bool) [][]bool {
	blocks := make([][]bool, len(bs))
	for j, row := range bs {
		blocks[j] = make([]bool, len(row))
	}
	// Transpose the argument matrix.
	for i, col := range bs {
		for j, v := range col {
			blocks[j][i] = v
		}
	}
	return blocks
}

// Pieces is the set of all the possible pieces.
var Pieces map[BlockType]*Piece

func init() {
	const (
		f = false
		t = true
	)
	Pieces = map[BlockType]*Piece{
		BlockType1: {
			blockType: BlockType1,
			blocks: transpose([][]bool{
				{f, f, f, f},
				{t, t, t, t},
				{f, f, f, f},
				{f, f, f, f},
			}),
		},
		BlockType2: {
			blockType: BlockType2,
			blocks: transpose([][]bool{
				{t, f, f},
				{t, t, t},
				{f, f, f},
			}),
		},
		BlockType3: {
			blockType: BlockType3,
			blocks: transpose([][]bool{
				{f, t, f},
				{t, t, t},
				{f, f, f},
			}),
		},
		BlockType4: {
			blockType: BlockType4,
			blocks: transpose([][]bool{
				{f, f, t},
				{t, t, t},
				{f, f, f},
			}),
		},
		BlockType5: {
			blockType: BlockType5,
			blocks: transpose([][]bool{
				{t, t, f},
				{f, t, t},
				{f, f, f},
			}),
		},
		BlockType6: {
			blockType: BlockType6,
			blocks: transpose([][]bool{
				{f, t, t},
				{t, t, f},
				{f, f, f},
			}),
		},
		BlockType7: {
			blockType: BlockType7,
			blocks: transpose([][]bool{
				{t, t},
				{t, t},
			}),
		},
	}
}

const (
	blockWidth       = 10
	blockHeight      = 10
	fieldBlockCountX = 10
	fieldBlockCountY = 20
)

func drawBlock(r *ebiten.Image, block BlockType, x, y int, clr flushColor) {
	if block == BlockTypeNone {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	if clr.active {
		op.ColorScale.ScaleAlpha(clr.alpha)
	}
	r.DrawImage(blockImages[block-1], op)
	if clr.active && clr.red > 0 {
		// colorm added red to the block's color. Add the same red, weighted by
		// the block's alpha, with additive blending.
		op.ColorScale.Reset()
		op.ColorScale.Scale(clr.red*clr.alpha, 0, 0, clr.alpha)
		op.Blend = ebiten.BlendLighter
		r.DrawImage(imageWhiteBlock, op)
	}
}

func (p *Piece) InitialPosition() (int, int) {
	size := len(p.blocks)
	x := (fieldBlockCountX - size) / 2
	var y int
Loop:
	for j := range size {
		for i := range size {
			if p.blocks[i][j] {
				break Loop
			}
		}
		y--
	}
	return x, y
}

// isBlocked returns a boolean value indicating whether
// there is a block at the position (x, y) of the piece with the given angle.
func (p *Piece) isBlocked(i, j int, angle Angle) bool {
	size := len(p.blocks)
	i2, j2 := i, j
	switch angle {
	case Angle0:
	case Angle90:
		i2 = j
		j2 = size - 1 - i
	case Angle180:
		i2 = size - 1 - i
		j2 = size - 1 - j
	case Angle270:
		i2 = size - 1 - j
		j2 = i
	}
	return p.blocks[i2][j2]
}

// collides returns a boolean value indicating whether
// the piece at (x, y) with the given angle would collide with the field's blocks.
func (p *Piece) collides(field *Field, x, y int, angle Angle) bool {
	size := len(p.blocks)
	for i := range size {
		for j := range size {
			if field.IsBlocked(x+i, y+j) && p.isBlocked(i, j, angle) {
				return true
			}
		}
	}
	return false
}

func (p *Piece) AbsorbInto(field *Field, x, y int, angle Angle) {
	size := len(p.blocks)
	for i := range size {
		for j := range size {
			if p.isBlocked(i, j, angle) {
				field.setBlock(x+i, y+j, p.blockType)
			}
		}
	}
}

func (p *Piece) DrawAtCenter(r *ebiten.Image, x, y, width, height int, angle Angle) {
	x += (width - len(p.blocks[0])*blockWidth) / 2
	y += (height - len(p.blocks)*blockHeight) / 2
	p.Draw(r, x, y, angle)
}

func (p *Piece) Draw(r *ebiten.Image, x, y int, angle Angle) {
	for i := range p.blocks {
		for j := range p.blocks[i] {
			if p.isBlocked(i, j, angle) {
				drawBlock(r, p.blockType, i*blockWidth+x, j*blockHeight+y, flushColor{})
			}
		}
	}

}
