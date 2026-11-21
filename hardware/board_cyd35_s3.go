//go:build tinygo && esp32s3 && ikaten_esp32_3248s035c

package hardware

import (
	"device/esp"
	"machine"
)

const LogicalWidth, LogicalHeight = 240, 240

// The Sunton ESP32-3248S035C (3.5" 480x320 ST7796, "CYD") with its module
// replaced by an ESP32-S3-WROOM-1. Pins, MADCTL and the joystick wiring are
// from github.com/sat0ken/tinygo-cyd (board/s035/pins_s3.go and s035.go),
// checked there on the board. The 240x240 screen is shown 4/3 bigger, 320x320
// in the middle.
var Device = &ST7796{Config: ST7796Config{
	SPI: machine.SPI0, Frequency: 40_000_000,
	SCK: machine.GPIO19, MOSI: machine.GPIO9, MISO: machine.GPIO20, CS: machine.GPIO47, DC: machine.GPIO48,
	Backlight: machine.GPIO8,
	Width:     480, Height: 320, MADCTL: 0x28, // landscape, BGR, USB connector on the right
	Fractional: true,
	// The mini joystick module on connector CN1 (GND, IO22, IO21, 3.3V as
	// printed for the original module): SDA on old IO21, SCL on old IO22.
	Joystick: machine.I2C1, JoySDA: machine.GPIO42, JoySCL: machine.GPIO2,
	// SCK and MISO are the S3's USB pins: after reset the USB Serial/JTAG
	// controller owns them. The board's serial console is the CH340 on UART0.
	PrepareSPI: func() { esp.USB_DEVICE.SetCONF0_USB_PAD_ENABLE(0) },
}}
