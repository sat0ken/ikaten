//go:build ikaten_benchmark

package train

import "time"

// Sections of Draw, timed in benchmark builds: the sky, the ground, the
// stations, what stands beside the line, the wire and the handle, and the
// gauges with the boxes over the view.
var (
	sectionStart time.Time
	sectionTimes [6]time.Duration
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
func SectionTimes() [6]time.Duration {
	t := sectionTimes
	sectionTimes = [6]time.Duration{}
	return t
}
