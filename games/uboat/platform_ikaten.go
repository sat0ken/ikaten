package uboat

import (
	_ "embed"

	ebiten "github.com/sat0ken/ikaten"
)

// The large images stay in flash; drawn at start they took about 100 KiB of RAM.
var (
	//go:embed assets/waves.ika
	wavesData string
	//go:embed assets/clouds.ika
	cloudsData string
	//go:embed assets/sonar.ika
	sonarData string
)

func loadAssets() (waves, clouds, sonar *ebiten.Image) {
	load := func(data string) *ebiten.Image {
		img, err := ebiten.NewImageFromString(data)
		if err != nil {
			panic(err)
		}
		return img
	}
	return load(wavesData), load(cloudsData), load(sonarData)
}
