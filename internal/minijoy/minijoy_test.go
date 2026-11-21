package minijoy

import (
	"errors"
	"testing"

	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/inputstate"
)

// module answers like the firmware: a write sets the register, a read
// returns bytes from it, moving on after each.
type module struct {
	regs    [0x30]byte
	ptr     byte
	missing bool
}

func (m *module) Tx(addr uint16, w, r []byte) error {
	if m.missing || addr != Address {
		return errors.New("nack")
	}
	if len(w) > 0 {
		m.ptr = w[0]
	}
	for i := range r {
		r[i] = m.regs[m.ptr]
		m.ptr++
	}
	return nil
}

func TestUpdate(t *testing.T) {
	m := &module{}
	m.regs[regStick], m.regs[regStick+1] = 250, 10 // right and down
	for i := range 5 {
		m.regs[regButtons+i] = 8 // none
	}
	m.regs[regButtons+1] = pressDown     // A
	m.regs[regButtons+3] = longPressHold // B
	d := New(m)
	if !d.Probe() {
		t.Fatal("not found")
	}
	d.Update()
	inputstate.Advance() // what RunGame does after KeyUpdate
	for k, want := range map[ikaten.Key]bool{
		ikaten.KeyArrowRight: true, ikaten.KeyArrowLeft: false, ikaten.KeyArrowDown: true, ikaten.KeyArrowUp: false,
		ikaten.KeyZ: true, ikaten.KeyX: true, ikaten.KeyEnter: false, ikaten.KeySpace: false,
	} {
		if ikaten.IsKeyPressed(k) != want {
			t.Errorf("key %v: %v", k, !want)
		}
	}
	// A module that stops answering releases everything.
	m.missing = true
	d.Update()
	inputstate.Advance()
	if ikaten.IsKeyPressed(ikaten.KeyZ) || ikaten.IsKeyPressed(ikaten.KeyArrowRight) {
		t.Fatal("keys stay down without the module")
	}
}
