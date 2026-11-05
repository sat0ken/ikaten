//go:build js && wasm

package hardware

import (
	"errors"
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/present"
	"syscall/js"
)

const LogicalWidth, LogicalHeight = 240, 240

type WASM struct {
	pixels    []byte
	buffer    js.Value
	presenter present.Presenter
	callbacks [3]js.Func
	active    bool
}

var Device = &WASM{}

func (d *WASM) Init() error {
	g := js.Global()
	if g.Get("ikatenInit").Type() != js.TypeFunction || g.Get("ikatenPresent").Type() != js.TypeFunction {
		return errors.New("ikaten: load static/index.html before WASM")
	}
	if d.active {
		d.Close()
	}
	g.Call("ikatenInit", 240, 240)
	d.buffer = g.Get("ikatenPixels")
	d.pixels = make([]byte, 240*240*4)
	d.presenter = present.Presenter{Width: 240, Height: 240}
	for n := 0; n < 2; n++ {
		pressed := n == 0
		d.callbacks[n] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) == 0 {
				return nil
			}
			event := args[0]
			if key, ok := browserKey(event.Get("code").String()); ok {
				ikaten.SetKeyState(key, pressed)
				event.Call("preventDefault")
			}
			return nil
		})
	}
	d.callbacks[2] = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for k := ikaten.Key(0); k <= ikaten.KeyMax; k++ {
			ikaten.SetKeyState(k, false)
		}
		return nil
	})
	g.Call("addEventListener", "keydown", d.callbacks[0])
	g.Call("addEventListener", "keyup", d.callbacks[1])
	g.Call("addEventListener", "blur", d.callbacks[2])
	d.active = true
	return nil
}
func (d *WASM) Close() {
	if !d.active {
		return
	}
	g := js.Global()
	g.Call("removeEventListener", "keydown", d.callbacks[0])
	g.Call("removeEventListener", "keyup", d.callbacks[1])
	g.Call("removeEventListener", "blur", d.callbacks[2])
	for _, f := range d.callbacks {
		f.Release()
	}
	d.active = false
}
func browserKey(code string) (ikaten.Key, bool) {
	if len(code) == 4 && code[:3] == "Key" && code[3] >= 'A' && code[3] <= 'Z' {
		return ikaten.KeyA + ikaten.Key(code[3]-'A'), true
	}
	if len(code) == 6 && code[:5] == "Digit" && code[5] >= '0' && code[5] <= '9' {
		return ikaten.KeyDigit0 + ikaten.Key(code[5]-'0'), true
	}
	switch code {
	case "ArrowUp":
		return ikaten.KeyArrowUp, true
	case "ArrowDown":
		return ikaten.KeyArrowDown, true
	case "ArrowLeft":
		return ikaten.KeyArrowLeft, true
	case "ArrowRight":
		return ikaten.KeyArrowRight, true
	case "Space":
		return ikaten.KeySpace, true
	case "Enter":
		return ikaten.KeyEnter, true
	case "Escape":
		return ikaten.KeyEscape, true
	case "Tab":
		return ikaten.KeyTab, true
	case "Backspace":
		return ikaten.KeyBackspace, true
	case "ShiftLeft":
		return ikaten.KeyShiftLeft, true
	case "ShiftRight":
		return ikaten.KeyShiftRight, true
	case "ControlLeft":
		return ikaten.KeyControlLeft, true
	case "ControlRight":
		return ikaten.KeyControlRight, true
	case "AltLeft":
		return ikaten.KeyAltLeft, true
	case "AltRight":
		return ikaten.KeyAltRight, true
	}
	return 0, false
}
func (d *WASM) Size() (int, int) { return 240, 240 }
func (d *WASM) KeyUpdate() error { return nil }
func (d *WASM) Present(screen *ikaten.Image) error {
	if err := d.presenter.Present(screen, d); err != nil {
		return err
	}
	js.CopyBytesToJS(d.buffer, d.pixels)
	js.Global().Call("ikatenPresent")
	return nil
}
func (d *WASM) WriteRect(x, y, w, h int, data []byte) error {
	p := 0
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			v := uint16(data[p])<<8 | uint16(data[p+1])
			p += 2
			r, g, b := byte(v>>11), byte(v>>5)&63, byte(v)&31
			i := ((y+dy)*240 + x + dx) * 4
			d.pixels[i], d.pixels[i+1], d.pixels[i+2], d.pixels[i+3] = r<<3|r>>2, g<<2|g>>4, b<<3|b>>2, 255
		}
	}
	return nil
}
