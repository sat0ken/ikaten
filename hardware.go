package ikaten

// Hardware is the boundary between the portable engine and a display/input
// backend. Size reports physical pixels. Present performs integer scaling and
// centering. KeyUpdate calls SetKeyState for each supported key.
type Hardware interface {
	Init() error
	Size() (width, height int)
	Present(screen *Image) error
	KeyUpdate() error
}

var currentHardware Hardware

// SetHardware selects a backend before RunGame. It must not be called while
// RunGame is active.
func SetHardware(hw Hardware) {
	if running {
		panic("ikaten: cannot change hardware during RunGame")
	}
	currentHardware = hw
}
