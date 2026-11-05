//go:build ikaten_benchmark

package ikaten

import "time"

// BenchmarkStats contains cumulative frame timings. Subtract two snapshots for
// a measurement interval. Timings include screen clearing in Draw.
type BenchmarkStats struct {
	Frames                uint64
	Update, Draw, Present time.Duration
}

var benchmarkStats BenchmarkStats

func ReadBenchmarkStats() BenchmarkStats { return benchmarkStats }
func resetBenchmark()                    { benchmarkStats = BenchmarkStats{} }
func benchmarkStart() time.Time          { return time.Now() }
func benchmarkUpdate(start time.Time)    { benchmarkStats.Update += time.Since(start) }
func benchmarkDraw(start time.Time)      { benchmarkStats.Draw += time.Since(start) }
func benchmarkPresent(start time.Time) {
	benchmarkStats.Present += time.Since(start)
	benchmarkStats.Frames++
}
