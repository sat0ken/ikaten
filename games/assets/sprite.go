// Package assets contains a small generated sprite for asset pipeline examples.
package assets

import (
	_ "embed"
	"github.com/sat0ken/ikaten"
)

//go:generate go run ../../cmd/ikaten-asset -o sprite.ika sprite.png
//go:generate go run ../../cmd/ikaten-asset -o opaque.ika opaque.png
//go:embed sprite.ika
var data string

//go:embed opaque.ika
var opaqueData string

func NewSprite() (*ikaten.Image, error) { return ikaten.NewImageFromString(data) }

// NewOpaqueSprite returns a sprite without a mask, for measuring the opaque path.
func NewOpaqueSprite() (*ikaten.Image, error) { return ikaten.NewImageFromString(opaqueData) }
