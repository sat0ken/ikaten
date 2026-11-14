//go:build ikaten_benchmark

package kart

import "time"

// Sections of Draw, timed in benchmark builds: the sky and the mountains,
// the ground, the karts, and the board.
var (
	sectionStart time.Time
	sectionTimes [4]time.Duration
)

// mark ends section i-1 and starts section i.
func mark(i int) {
	now := time.Now()
	if i > 0 {
		sectionTimes[i-1] += now.Sub(sectionStart)
	}
	sectionStart = now
}

// SectionTimes returns the time spent in each section of Draw since it was
// last called.
func SectionTimes() [4]time.Duration {
	t := sectionTimes
	sectionTimes = [4]time.Duration{}
	return t
}
