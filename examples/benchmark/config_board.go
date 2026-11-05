//go:build ikaten_benchmark && tinygo && (esp32s3 || esp32 || rp2040)

package main

import (
	"github.com/sat0ken/ikaten/hardware"
	"strconv"
)

// spiHz overrides the board default with -ldflags="-X main.spiHz=80000000".
var spiHz string

func configure() {
	if hz, err := strconv.Atoi(spiHz); err == nil && hz > 0 {
		hardware.Device.Config.Frequency = uint32(hz)
	}
}
