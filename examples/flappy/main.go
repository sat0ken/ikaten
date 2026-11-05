package main

import (
	"github.com/sat0ken/ikaten/hardware"

	ebiten "github.com/sat0ken/ikaten"
)

func main() {
	ebiten.SetHardware(hardware.Device)
	if err := ebiten.RunGame(game()); err != nil {
		println(err.Error())
	}
}
