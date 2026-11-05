package kart

import (
	_ "embed"

	ebiten "github.com/sat0ken/ikaten"
)

// The track stays in flash: 256x256 RGB565 would take 128 KiB of RAM.
//
//go:embed assets/track.ika
var trackData string

func loadTrack() *ebiten.Image {
	img, err := ebiten.NewImageFromString(trackData)
	if err != nil {
		panic(err)
	}
	return img
}
