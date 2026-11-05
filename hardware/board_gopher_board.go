//go:build tinygo && rp2040 && ikaten_gopher_board

package hardware

import (
	"github.com/sat0ken/ikaten"
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 120, 120

// Gopher Board (RP2040-Zero). Pins are from github.com/sat0ken/gopher-board-workshop:
// the schematic and 08_st7789_txt. Switches pull to GND.
var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI1, Frequency: 62_500_000, SCK: machine.GPIO10, MOSI: machine.GPIO11, MISO: machine.NoPin,
	CS: machine.GPIO13, DC: machine.GPIO12, Reset: machine.GPIO9, Backlight: machine.GPIO14,
	Panel: st7789.Config{Width: 240, Height: 240, Rotation: st7789.ROTATION_90},
	Buttons: []Button{{Pin: machine.GPIO26, Key: ikaten.KeyZ}, {Pin: machine.GPIO15, Key: ikaten.KeyX},
		{Pin: machine.GPIO3, Key: ikaten.KeyArrowUp}, {Pin: machine.GPIO6, Key: ikaten.KeyArrowDown},
		{Pin: machine.GPIO4, Key: ikaten.KeyArrowLeft}, {Pin: machine.GPIO5, Key: ikaten.KeyArrowRight}},
	Debounce: 5 * time.Millisecond,
}}
