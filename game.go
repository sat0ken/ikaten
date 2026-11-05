package ikaten

import "errors"

type Game interface {
	Update() error
	Draw(screen *Image)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

// Termination stops RunGame successfully when returned by Update.
var Termination = errors.New("ikaten: game terminated")

func SetWindowSize(width, height int) {}
func SetWindowTitle(title string)     {}
func SetFullscreen(fullscreen bool)   {}
func SetVsyncEnabled(enabled bool)    {}
