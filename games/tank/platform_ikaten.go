package tank

import (
	_ "embed"

	ebiten "github.com/sat0ken/ikaten"
)

// The images stay in flash: the ground alone would take 128 KiB of RAM.
var (
	//go:embed assets/ground.ika
	groundData string
	//go:embed assets/sprites.ika
	spritesData string
	//go:embed assets/hills.ika
	hillsData string
)

func loadAssets() (ground, sprites, hills *ebiten.Image) {
	load := func(data string) *ebiten.Image {
		img, err := ebiten.NewImageFromString(data)
		if err != nil {
			panic(err)
		}
		return img
	}
	return load(groundData), load(spritesData), load(hillsData)
}
