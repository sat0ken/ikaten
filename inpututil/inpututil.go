// Package inpututil tracks input in Update ticks, as in Ebitengine.
package inpututil

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/inputstate"
)

func KeyPressDuration(key ebiten.Key) int {
	if key < 0 || key > ebiten.KeyMax {
		return 0
	}
	return inputstate.Duration(int(key))
}
func IsKeyJustPressed(key ebiten.Key) bool { return KeyPressDuration(key) == 1 }
func IsKeyJustReleased(key ebiten.Key) bool {
	return key >= 0 && key <= ebiten.KeyMax && inputstate.Released(int(key))
}
func AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			keys = append(keys, k)
		}
	}
	return keys
}
func AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key {
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if IsKeyJustPressed(k) {
			keys = append(keys, k)
		}
	}
	return keys
}
func AppendJustReleasedKeys(keys []ebiten.Key) []ebiten.Key {
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if IsKeyJustReleased(k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Mouse and touch input does not exist on the boards, so these report nothing.
func IsMouseButtonJustPressed(button ebiten.MouseButton) bool  { return false }
func IsMouseButtonJustReleased(button ebiten.MouseButton) bool { return false }
func MouseButtonPressDuration(button ebiten.MouseButton) int   { return 0 }
func AppendJustPressedTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID {
	return ids
}
func AppendJustReleasedTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID {
	return ids
}
func IsTouchJustReleased(id ebiten.TouchID) bool { return false }
func TouchPressDuration(id ebiten.TouchID) int   { return 0 }
func GamepadButtonPressDuration(id ebiten.GamepadID, b ebiten.GamepadButton) int {
	k, ok := ebiten.GamepadButtonKey(b)
	if id != 0 || !ok {
		return 0
	}
	return KeyPressDuration(k)
}
func IsGamepadButtonJustPressed(id ebiten.GamepadID, b ebiten.GamepadButton) bool {
	return GamepadButtonPressDuration(id, b) == 1
}
func IsGamepadButtonJustReleased(id ebiten.GamepadID, b ebiten.GamepadButton) bool {
	k, ok := ebiten.GamepadButtonKey(b)
	return id == 0 && ok && IsKeyJustReleased(k)
}
func AppendPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	for b := ebiten.GamepadButton(0); b <= ebiten.GamepadButtonMax; b++ {
		if GamepadButtonPressDuration(id, b) > 0 {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
func AppendJustPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	for b := ebiten.GamepadButton(0); b <= ebiten.GamepadButtonMax; b++ {
		if IsGamepadButtonJustPressed(id, b) {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
func AppendJustReleasedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	for b := ebiten.GamepadButton(0); b <= ebiten.GamepadButtonMax; b++ {
		if IsGamepadButtonJustReleased(id, b) {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
func AppendPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	for b := ebiten.StandardGamepadButton(0); b <= ebiten.StandardGamepadButtonMax; b++ {
		if StandardGamepadButtonPressDuration(id, b) > 0 {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
func StandardGamepadButtonPressDuration(id ebiten.GamepadID, b ebiten.StandardGamepadButton) int {
	k, ok := ebiten.StandardGamepadButtonKey(b)
	if id != 0 || !ok {
		return 0
	}
	return KeyPressDuration(k)
}
func IsStandardGamepadButtonJustPressed(id ebiten.GamepadID, b ebiten.StandardGamepadButton) bool {
	return StandardGamepadButtonPressDuration(id, b) == 1
}
func IsStandardGamepadButtonJustReleased(id ebiten.GamepadID, b ebiten.StandardGamepadButton) bool {
	k, ok := ebiten.StandardGamepadButtonKey(b)
	return id == 0 && ok && IsKeyJustReleased(k)
}
func AppendJustPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	for b := ebiten.StandardGamepadButton(0); b <= ebiten.StandardGamepadButtonMax; b++ {
		if IsStandardGamepadButtonJustPressed(id, b) {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
func AppendJustReleasedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	for b := ebiten.StandardGamepadButton(0); b <= ebiten.StandardGamepadButtonMax; b++ {
		if IsStandardGamepadButtonJustReleased(id, b) {
			buttons = append(buttons, b)
		}
	}
	return buttons
}
