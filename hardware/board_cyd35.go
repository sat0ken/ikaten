//go:build tinygo && esp32 && ikaten_esp32_3248s035c

package hardware

import "machine"

const LogicalWidth, LogicalHeight = 240, 240

// The Sunton ESP32-3248S035C (3.5" 480x320 ST7796, "CYD") with its original
// ESP32 module. Pins are from github.com/sat0ken/tinygo-cyd
// (board/s035/pins_esp32.go). The ESP32 has too little RAM for the 240x240
// games, like the esp32 target.
var Device = &ST7796{Config: ST7796Config{
	SPI: machine.SPI2, Frequency: 40_000_000,
	SCK: machine.GPIO14, MOSI: machine.GPIO13, MISO: machine.GPIO12, CS: machine.GPIO15, DC: machine.GPIO2,
	Backlight: machine.GPIO27,
	Width:     480, Height: 320, MADCTL: 0x28,
	Fractional: true,
	Joystick:   machine.I2C1, JoySDA: machine.GPIO21, JoySCL: machine.GPIO22,
}}
