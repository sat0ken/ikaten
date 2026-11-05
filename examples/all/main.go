// all starts the launcher with every ikaten game.
package main

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/all"
	"github.com/sat0ken/ikaten/hardware"
)

func main() {
	ebiten.SetHardware(hardware.Device)
	if err := ebiten.RunGame(all.NewGame()); err != nil {
		println(err.Error())
	}
}
