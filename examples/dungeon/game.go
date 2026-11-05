//go:build !ikaten_benchmark

package main

import (
	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/dungeon"
)

func game() ebiten.Game { return dungeon.NewGame() }
