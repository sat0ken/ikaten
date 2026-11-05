//go:build !js && !tinygo

package hardware

import "github.com/sat0ken/ikaten"

const LogicalWidth, LogicalHeight = 240, 240

// Headless is a desktop test backend. Present retains the latest frame; it
// does not create a native window. Use the WASM target for interactive preview.
type Headless struct {
	Width, Height int
	Screen        *ikaten.Image
}

var Device = &Headless{Width: 240, Height: 240}

func (h *Headless) Init() error                        { h.Screen = nil; return nil }
func (h *Headless) Size() (int, int)                   { return h.Width, h.Height }
func (h *Headless) Present(screen *ikaten.Image) error { h.Screen = screen; return nil }
func (h *Headless) KeyUpdate() error                   { return nil }
