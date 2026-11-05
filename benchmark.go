//go:build !ikaten_benchmark

package ikaten

import "time"

func benchmarkStart() time.Time  { return time.Time{} }
func benchmarkUpdate(time.Time)  {}
func benchmarkDraw(time.Time)    {}
func benchmarkPresent(time.Time) {}
func resetBenchmark()            {}
