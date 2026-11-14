//go:build ikaten_benchmark

package dungeon

import "time"

// Sections of Draw, timed in benchmark builds: the ceiling and the floor,
// the walls, the slimes, the weapon, and the status bar.
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
