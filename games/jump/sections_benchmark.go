//go:build ikaten_benchmark

package jump

import "time"

// Sections of Draw, timed in benchmark builds: the background, the tiles,
// the enemies, the gopher, the board and the boxes over the play.
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
