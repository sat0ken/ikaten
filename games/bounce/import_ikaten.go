//go:build !ebitengine

package bounce

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
	"github.com/sat0ken/ikaten/inpututil"
)

type gameImage = ebiten.Image
type drawImageOptions = ebiten.DrawImageOptions

var newGameImage = ebiten.NewImage
var pressed = ebiten.IsKeyPressed
var justPressed = inpututil.IsKeyJustPressed
var debugPrint = ebitenutil.DebugPrint
var termination = ebiten.Termination

const (
	keyZ      = ebiten.KeyZ
	keyX      = ebiten.KeyX
	keyEscape = ebiten.KeyEscape
	keyLeft   = ebiten.KeyArrowLeft
	keyRight  = ebiten.KeyArrowRight
	keyUp     = ebiten.KeyArrowUp
	keyDown   = ebiten.KeyArrowDown
)
