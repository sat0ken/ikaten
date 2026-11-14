//go:build !ikaten_benchmark

package kart

// mark times the sections of Draw in benchmark builds only.
func mark(int) {}
