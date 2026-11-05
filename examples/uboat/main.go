package main

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

func main() {
	ebiten.SetHardware(hardware.Device)
	if err := ebiten.RunGame(game()); err != nil {
		println(err.Error())
	}
}
