//go:build ikaten_benchmark

package uboat

import "time"

// Sections of Draw, timed in benchmark builds: the sky and the sea filled (or
// the sonar), the clouds and the waves, the ships and the reticle, the
// periscope's frame, and the gauges.
var (
	sectionStart time.Time
	sectionTimes [5]time.Duration
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
func SectionTimes() [5]time.Duration {
	t := sectionTimes
	sectionTimes = [5]time.Duration{}
	return t
}
