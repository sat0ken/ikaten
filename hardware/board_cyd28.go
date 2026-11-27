//go:build tinygo && esp32 && ikaten_esp32_2432s028r

package hardware

import "machine"

const LogicalWidth, LogicalHeight = 240, 240

// The Sunton ESP32-2432S028R (2.8" 320x240 ILI9341, "CYD") with its original
// ESP32 module. Pins and MADCTL are from github.com/sat0ken/tinygo-cyd
// (board/s028/pins_esp32.go), checked there on a board with one USB
// connector. The ESP32 has too little RAM for the 240x240 games, like the
// esp32 target.
var Device = &SPILCD{Config: SPILCDConfig{
	SPI: machine.SPI2, Frequency: 40_000_000,
	SCK: machine.GPIO14, MOSI: machine.GPIO13, MISO: machine.GPIO12, CS: machine.GPIO15, DC: machine.GPIO2,
	Backlight: machine.GPIO21,
	Width:     320, Height: 240, Init: InitILI9341, MADCTL: 0x28, // landscape, BGR
	Joystick: machine.I2C1, JoySDA: machine.GPIO27, JoySCL: machine.GPIO22,
}}
