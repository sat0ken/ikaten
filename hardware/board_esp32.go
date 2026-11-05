//go:build tinygo && esp32

package hardware

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 120, 120

// Generic ESP32 example wiring. Buttons are opt-in; see docs/hardware.md.
var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI2, Frequency: 40_000_000, SCK: machine.GPIO18, MOSI: machine.GPIO23, MISO: machine.NoPin,
	CS: machine.GPIO5, DC: machine.GPIO2, Reset: machine.GPIO4, Backlight: machine.NoPin,
	Panel: st7789.Config{Width: 240, Height: 240}, Debounce: 5 * time.Millisecond,
}}
