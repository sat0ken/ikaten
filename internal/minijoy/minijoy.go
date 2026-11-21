// Package minijoy reads the nulllab mini joystick module: a stick with a
// push button and four buttons, read by a CH32V003 that answers on I2C at
// address 0x5A. The register map and the button letters (Y at the top, A at
// the bottom, X on the left, B on the right, as printed on the module) follow
// github.com/sat0ken/tinygo-cyd (minijoy), which checked them on the module.
//
// It does not import machine, so it is tested on the host.
package minijoy

import "github.com/sat0ken/ikaten"

// Bus is an I2C bus (machine.I2C).
type Bus interface {
	Tx(addr uint16, w, r []byte) error
}

const (
	Address    = 0x5A
	regStick   = 0x10 // X, then Y: 0..255, 128 in the middle, X to the right, Y up
	regButtons = 0x20 // OK (the stick), A, Y, B, X: one event each
	dead       = 48   // the stick counts as pushed this far from the middle
)

// Events the module reports for a button while it is held.
const (
	pressDown      = 0
	pressRepeat    = 2
	longPressStart = 5
	longPressHold  = 6
)

func held(event byte) bool {
	switch event {
	case pressDown, pressRepeat, longPressStart, longPressHold:
		return true
	}
	return false
}

// Keys are the keys the buttons press, in the order of the registers from
// regButtons: the stick's button, A, Y, B and X. A and B are Z and X, like
// the badge's SW1 and SW2.
var Keys = [5]ikaten.Key{ikaten.KeyEnter, ikaten.KeyZ, ikaten.KeySpace, ikaten.KeyX, ikaten.KeyC}

// Device is the module.
type Device struct {
	bus Bus
	w   [1]byte
	r   [5]byte
	ok  bool
}

func New(bus Bus) *Device { return &Device{bus: bus} }

// Probe reports whether the module answers.
func (d *Device) Probe() bool {
	d.ok = d.bus.Tx(Address, nil, d.r[:1]) == nil
	return d.ok
}

// read reads n bytes from reg: the register in one transaction and the bytes
// in another, as the module's firmware expects.
func (d *Device) read(reg byte, n int) bool {
	d.w[0] = reg
	return d.bus.Tx(Address, d.w[:], nil) == nil && d.bus.Tx(Address, nil, d.r[:n]) == nil
}

// Update reads the module and sets the arrow keys and the button keys. If the
// module is missing or does not answer, everything is released.
func (d *Device) Update() {
	var x, y byte = 128, 128
	var buttons [5]bool
	if d.ok && d.read(regStick, 2) {
		x, y = d.r[0], d.r[1]
		if d.read(regButtons, 5) {
			for i := range buttons {
				buttons[i] = held(d.r[i])
			}
		}
	}
	ikaten.SetKeyState(ikaten.KeyArrowRight, int(x) > 128+dead)
	ikaten.SetKeyState(ikaten.KeyArrowLeft, int(x) < 128-dead)
	ikaten.SetKeyState(ikaten.KeyArrowUp, int(y) > 128+dead)
	ikaten.SetKeyState(ikaten.KeyArrowDown, int(y) < 128-dead)
	for i, k := range Keys {
		ikaten.SetKeyState(k, buttons[i])
	}
}
