package hardware

import (
	"testing"
	"time"
)

func TestDebounceAndJoystickHysteresis(t *testing.T) {
	now := time.Unix(0, 0)
	var d debouncer
	delay := 5 * time.Millisecond
	if d.update(true, now, delay) || d.update(false, now.Add(time.Millisecond), delay) || d.update(true, now.Add(2*time.Millisecond), delay) {
		t.Fatal("bounce accepted")
	}
	if !d.update(true, now.Add(7*time.Millisecond), delay) {
		t.Fatal("stable press ignored")
	}
	if !d.update(false, now.Add(8*time.Millisecond), delay) || d.update(false, now.Add(13*time.Millisecond), delay) {
		t.Fatal("release debounce")
	}
	if axisDirection(6500, 0, 6000, 4000) != 1 || axisDirection(4500, 1, 6000, 4000) != 1 || axisDirection(3500, 1, 6000, 4000) != 0 || axisDirection(-6500, 1, 6000, 4000) != -1 {
		t.Fatal("axis hysteresis")
	}
}
