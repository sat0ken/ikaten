//go:build tinygo && esp32s3 && !ikaten_badge && !ikaten_esp32_3248s035c && !ikaten_esp32_2432s028r

package hardware

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 240, 240

// Generic S3 defaults use the badge's LCD wiring. Configure Buttons for a
// different board before RunGame. The dedicated badge target enables its ADC.
var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI0, Frequency: 40_000_000, SCK: machine.GPIO12, MOSI: machine.GPIO11, MISO: machine.NoPin,
	CS: machine.GPIO10, DC: machine.GPIO5, Reset: machine.GPIO4, Backlight: machine.NoPin,
	Panel: st7789.Config{Width: 240, Height: 240}, Debounce: 5 * time.Millisecond,
}}
