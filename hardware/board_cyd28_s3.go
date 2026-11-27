//go:build tinygo && esp32s3 && ikaten_esp32_2432s028r

package hardware

import (
	"device/esp"
	"machine"
)

const LogicalWidth, LogicalHeight = 240, 240

// The Sunton ESP32-2432S028R (2.8" 320x240 ILI9341, "CYD") with its module
// replaced by an ESP32-S3-WROOM-1. Pins, MADCTL and the joystick wiring are
// from github.com/sat0ken/tinygo-cyd (board/s028/pins_s3.go), checked there
// on a board with two USB connectors: without row/column exchange its panel
// is already landscape, MY turns it the right way up and the colours are
// RGB. The 240x240 screen is shown as it is, in the middle.
var Device = &SPILCD{Config: SPILCDConfig{
	SPI: machine.SPI0, Frequency: 40_000_000,
	SCK: machine.GPIO19, MOSI: machine.GPIO9, MISO: machine.GPIO20, CS: machine.GPIO47, DC: machine.GPIO48,
	Backlight: machine.GPIO42,
	Width:     320, Height: 240, Init: InitILI9341, MADCTL: 0x80,
	// The mini joystick module on connector CN1 (GND, IO22, IO27, 3.3V as
	// printed for the original module): SDA on old IO27, SCL on old IO22.
	Joystick: machine.I2C1, JoySDA: machine.GPIO8, JoySCL: machine.GPIO2,
	// SCK and MISO are the S3's USB pins: after reset the USB Serial/JTAG
	// controller owns them. The board's serial console is the CH340 on UART0.
	PrepareSPI: func() { esp.USB_DEVICE.SetCONF0_USB_PAD_ENABLE(0) },
}}
