//go:build !ikaten_benchmark

package uboat

// mark times the sections of Draw in benchmark builds only.
func mark(int) {}
