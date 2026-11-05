//go:build tinygo && gopher_badge

package hardware

import (
	"github.com/sat0ken/ikaten"
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 120, 120

var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI0, Frequency: 62_500_000, SCK: machine.TFT_SCL, MOSI: machine.TFT_SDI, MISO: machine.NoPin,
	CS: machine.TFT_CS, DC: machine.TFT_WRX, Reset: machine.TFT_RST, Backlight: machine.TFT_BACKLIGHT,
	Panel: st7789.Config{Width: 240, Height: 320, Rotation: st7789.ROTATION_270},
	Buttons: []Button{{Pin: machine.BUTTON_A, Key: ikaten.KeyZ}, {Pin: machine.BUTTON_B, Key: ikaten.KeyX},
		{Pin: machine.BUTTON_UP, Key: ikaten.KeyArrowUp}, {Pin: machine.BUTTON_DOWN, Key: ikaten.KeyArrowDown},
		{Pin: machine.BUTTON_LEFT, Key: ikaten.KeyArrowLeft}, {Pin: machine.BUTTON_RIGHT, Key: ikaten.KeyArrowRight}},
	Debounce: 5 * time.Millisecond,
}}
