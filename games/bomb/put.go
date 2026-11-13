package bomb

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/ebitenutil"
)

// put draws img with op, in the game's coordinates: everything is scaled
// by pixelScale onto the screen.
func put(screen, img *ebiten.Image, op *ebiten.DrawImageOptions) {
	if pixelScale != 1 {
		op.GeoM.Scale(pixelScale, pixelScale)
	}
	screen.DrawImage(img, op)
}

// debugPrint prints at a place in the game's coordinates. The debug font
// keeps its size, so on a larger screen it is smaller next to the rest.
func debugPrint(screen *ebiten.Image, s string, x, y int) {
	ebitenutil.DebugPrintAt(screen, s, x*pixelScale, y*pixelScale)
}
