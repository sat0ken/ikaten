package main

import (
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/bounce"
	"github.com/sat0ken/ikaten/hardware"
)

func main() {
	ikaten.SetHardware(hardware.Device)
	count := 15
	if hardware.LogicalWidth >= 240 {
		count = 30
	}
	if err := ikaten.RunGame(bounce.New(hardware.LogicalWidth, hardware.LogicalHeight, count)); err != nil {
		println(err.Error())
	}
}
