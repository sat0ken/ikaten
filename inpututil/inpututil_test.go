package inpututil

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/inputstate"
	"testing"
)

func TestInputEdgesCatchupAndVirtualGamepad(t *testing.T) {
	inputstate.Reset()
	defer inputstate.Reset()
	ebiten.SetKeyState(ebiten.KeyZ, true)
	inputstate.Advance()
	if !IsKeyJustPressed(ebiten.KeyZ) || !IsStandardGamepadButtonJustPressed(0, ebiten.StandardGamepadButtonRightBottom) {
		t.Fatal("first tick edge")
	}
	inputstate.Advance()
	if IsKeyJustPressed(ebiten.KeyZ) || KeyPressDuration(ebiten.KeyZ) != 2 {
		t.Fatal("edge repeated during catch-up")
	}
	ebiten.SetKeyState(ebiten.KeyZ, false)
	inputstate.Advance()
	if !IsKeyJustReleased(ebiten.KeyZ) || KeyPressDuration(ebiten.KeyZ) != 0 {
		t.Fatal("release")
	}
	inputstate.Advance()
	if IsKeyJustReleased(ebiten.KeyZ) {
		t.Fatal("repeated release")
	}
	ebiten.SetKeyState(ebiten.KeyArrowRight, true)
	inputstate.Advance()
	if ebiten.StandardGamepadAxisValue(0, ebiten.StandardGamepadAxisLeftStickHorizontal) != 1 || ebiten.GamepadAxisValue(99, 0) != 0 {
		t.Fatal("virtual axes")
	}
	keys := make([]ebiten.Key, 0, 10)
	if allocs := testing.AllocsPerRun(100, func() { _ = AppendPressedKeys(keys[:0]) }); allocs != 0 {
		t.Fatal("input scan allocates", allocs)
	}
}

func TestGamepadButtonEdges(t *testing.T) {
	inputstate.Reset()
	defer inputstate.Reset()
	buttons := make([]ebiten.GamepadButton, 0, 8)
	standard := make([]ebiten.StandardGamepadButton, 0, 8)
	ebiten.SetKeyState(ebiten.KeyZ, true) // GamepadButton0, StandardGamepadButtonRightBottom
	inputstate.Advance()
	if !IsGamepadButtonJustPressed(0, ebiten.GamepadButton0) || GamepadButtonPressDuration(0, ebiten.GamepadButton0) != 1 {
		t.Fatal("press")
	}
	if got := AppendJustPressedGamepadButtons(0, buttons[:0]); len(got) != 1 || got[0] != ebiten.GamepadButton0 {
		t.Fatal("just pressed list", got)
	}
	inputstate.Advance()
	if IsGamepadButtonJustPressed(0, ebiten.GamepadButton0) || GamepadButtonPressDuration(0, ebiten.GamepadButton0) != 2 {
		t.Fatal("hold")
	}
	if got := AppendPressedGamepadButtons(0, buttons[:0]); len(got) != 1 || got[0] != ebiten.GamepadButton0 {
		t.Fatal("pressed list", got)
	}
	if got := AppendPressedStandardGamepadButtons(0, standard[:0]); len(got) != 1 || got[0] != ebiten.StandardGamepadButtonRightBottom {
		t.Fatal("standard pressed list", got)
	}
	ebiten.SetKeyState(ebiten.KeyZ, false)
	inputstate.Advance()
	if !IsGamepadButtonJustReleased(0, ebiten.GamepadButton0) || GamepadButtonPressDuration(0, ebiten.GamepadButton0) != 0 {
		t.Fatal("release")
	}
	if got := AppendJustReleasedGamepadButtons(0, buttons[:0]); len(got) != 1 || got[0] != ebiten.GamepadButton0 {
		t.Fatal("just released list", got)
	}
	// An unconnected ID reports nothing, even while the key is held.
	ebiten.SetKeyState(ebiten.KeyZ, true)
	inputstate.Advance()
	if IsGamepadButtonJustPressed(1, ebiten.GamepadButton0) || GamepadButtonPressDuration(1, ebiten.GamepadButton0) != 0 ||
		len(AppendPressedGamepadButtons(1, buttons[:0])) != 0 || len(AppendPressedStandardGamepadButtons(1, standard[:0])) != 0 {
		t.Fatal("unconnected gamepad")
	}
	ebiten.SetKeyState(ebiten.KeyZ, false)
	inputstate.Advance()
	if IsGamepadButtonJustReleased(1, ebiten.GamepadButton0) || len(AppendJustReleasedGamepadButtons(1, buttons[:0])) != 0 {
		t.Fatal("unconnected gamepad release")
	}
}
