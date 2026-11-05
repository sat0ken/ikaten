//go:build tinygo && rp2040 && !gopher_badge && !ikaten_gopher_board

package hardware

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 120, 120

var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI0, Frequency: 62_500_000, SCK: machine.GPIO18, MOSI: machine.GPIO19, MISO: machine.NoPin,
	CS: machine.GPIO17, DC: machine.GPIO20, Reset: machine.GPIO21, Backlight: machine.NoPin,
	Panel: st7789.Config{Width: 240, Height: 240}, Debounce: 5 * time.Millisecond,
}}
