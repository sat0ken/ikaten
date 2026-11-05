package ikaten

import "github.com/sat0ken/ikaten/internal/inputstate"

type Key int

const (
	KeyA Key = iota
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyDigit0
	KeyDigit1
	KeyDigit2
	KeyDigit3
	KeyDigit4
	KeyDigit5
	KeyDigit6
	KeyDigit7
	KeyDigit8
	KeyDigit9
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeySpace
	KeyEnter
	KeyEscape
	KeyTab
	KeyBackspace
	KeyShiftLeft
	KeyShiftRight
	KeyControlLeft
	KeyControlRight
	KeyAltLeft
	KeyAltRight
	KeyMax = KeyAltRight
)
const (
	KeyUp    = KeyArrowUp
	KeyDown  = KeyArrowDown
	KeyLeft  = KeyArrowLeft
	KeyRight = KeyArrowRight
)

// SetKeyState is the backend input hook. The snapshot becomes visible at the
// next Update, so edge events occur exactly once even during catch-up ticks.
func SetKeyState(key Key, pressed bool) { inputstate.Set(int(key), pressed) }
func IsKeyPressed(key Key) bool {
	return key >= 0 && key <= KeyMax && inputstate.Duration(int(key)) > 0
}

type GamepadID int
type GamepadButton int
type StandardGamepadButton int
type StandardGamepadAxis int

// GamepadAxisType is an alias of int, as in Ebitengine.
type GamepadAxisType = int

const (
	StandardGamepadButtonRightBottom StandardGamepadButton = iota
	StandardGamepadButtonRightRight
	StandardGamepadButtonRightLeft
	StandardGamepadButtonRightTop
	StandardGamepadButtonFrontTopLeft
	StandardGamepadButtonFrontTopRight
	StandardGamepadButtonFrontBottomLeft
	StandardGamepadButtonFrontBottomRight
	StandardGamepadButtonCenterLeft
	StandardGamepadButtonCenterRight
	StandardGamepadButtonLeftStick
	StandardGamepadButtonRightStick
	StandardGamepadButtonLeftTop
	StandardGamepadButtonLeftBottom
	StandardGamepadButtonLeftLeft
	StandardGamepadButtonLeftRight
	StandardGamepadButtonCenterCenter
	StandardGamepadButtonMax = StandardGamepadButtonCenterCenter
)
const (
	StandardGamepadAxisLeftStickHorizontal StandardGamepadAxis = iota
	StandardGamepadAxisLeftStickVertical
	StandardGamepadAxisRightStickHorizontal
	StandardGamepadAxisRightStickVertical
)
const (
	GamepadButton0 GamepadButton = iota
	GamepadButton1
	GamepadButton2
	GamepadButton3
	GamepadButton4
	GamepadButton5
	GamepadButtonMax = GamepadButton5
)

var gamepadKeys = [6]Key{KeyZ, KeyX, KeyArrowUp, KeyArrowDown, KeyArrowLeft, KeyArrowRight}

func AppendGamepadIDs(ids []GamepadID) []GamepadID { return append(ids, 0) }
func GamepadName(id GamepadID) string {
	if id == 0 {
		return "ikaten GPIO"
	}
	return ""
}
func GamepadSDLID(id GamepadID) string { return "" }
func GamepadButtonCount(id GamepadID) int {
	if id == 0 {
		return 6
	}
	return 0
}
func GamepadAxisCount(id GamepadID) int {
	if id == 0 {
		return 2
	}
	return 0
}

// GamepadButtonKey returns the digital key bound to a virtual button.
func GamepadButtonKey(button GamepadButton) (Key, bool) {
	if button < 0 || int(button) >= len(gamepadKeys) {
		return 0, false
	}
	return gamepadKeys[button], true
}
func IsGamepadButtonPressed(id GamepadID, button GamepadButton) bool {
	key, ok := GamepadButtonKey(button)
	return id == 0 && ok && IsKeyPressed(key)
}
func GamepadAxisValue(id GamepadID, axis GamepadAxisType) float64 {
	if id != 0 {
		return 0
	}
	switch axis {
	case 0:
		return axisValue(KeyArrowLeft, KeyArrowRight)
	case 1:
		return axisValue(KeyArrowUp, KeyArrowDown)
	}
	return 0
}
func axisValue(negative, positive Key) float64 {
	v := 0.0
	if IsKeyPressed(negative) {
		v--
	}
	if IsKeyPressed(positive) {
		v++
	}
	return v
}
func IsStandardGamepadLayoutAvailable(id GamepadID) bool { return id == 0 }

// StandardGamepadButtonKey returns the digital key bound to a virtual button.
func StandardGamepadButtonKey(button StandardGamepadButton) (Key, bool) {
	switch button {
	case StandardGamepadButtonRightBottom:
		return KeyZ, true
	case StandardGamepadButtonRightRight:
		return KeyX, true
	case StandardGamepadButtonLeftTop:
		return KeyArrowUp, true
	case StandardGamepadButtonLeftBottom:
		return KeyArrowDown, true
	case StandardGamepadButtonLeftLeft:
		return KeyArrowLeft, true
	case StandardGamepadButtonLeftRight:
		return KeyArrowRight, true
	}
	return 0, false
}
func IsStandardGamepadButtonAvailable(id GamepadID, button StandardGamepadButton) bool {
	_, ok := StandardGamepadButtonKey(button)
	return id == 0 && ok
}
func IsStandardGamepadButtonPressed(id GamepadID, button StandardGamepadButton) bool {
	key, ok := StandardGamepadButtonKey(button)
	return id == 0 && ok && IsKeyPressed(key)
}
func StandardGamepadButtonValue(id GamepadID, button StandardGamepadButton) float64 {
	if IsStandardGamepadButtonPressed(id, button) {
		return 1
	}
	return 0
}
func IsStandardGamepadAxisAvailable(id GamepadID, axis StandardGamepadAxis) bool {
	return id == 0 && axis >= 0 && axis <= StandardGamepadAxisLeftStickVertical
}
func StandardGamepadAxisValue(id GamepadID, axis StandardGamepadAxis) float64 {
	return GamepadAxisValue(id, GamepadAxisType(axis))
}

type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

type TouchID int

func IsMouseButtonPressed(button MouseButton) bool { return false }
func CursorPosition() (int, int)                   { return 0, 0 }
func AppendTouchIDs(ids []TouchID) []TouchID       { return ids }
func TouchPosition(id TouchID) (int, int)          { return 0, 0 }
func Wheel() (float64, float64)                    { return 0, 0 }
