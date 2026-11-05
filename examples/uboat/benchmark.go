//go:build ikaten_benchmark

package main

import (
	"runtime"
	"time"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/uboat"
)

// autoplay keeps the periscope up and turning, the worst case for drawing,
// fires now and then, and starts again after the boat is lost. It prints
// timings every 10 seconds. It presses Space: boards read their buttons into
// KeyZ and KeyX every tick, which would hide autoplay's presses.
type autoplay struct {
	*uboat.Game
	ticks         int
	last, prev    time.Time
	stats         ebiten.BenchmarkStats
	longestUpdate time.Duration // longest gap between Update calls, which shows GC pauses
}

func (g *autoplay) Update() error {
	g.ticks++
	ebiten.SetKeyState(ebiten.KeySpace, g.ticks%90 == 0) // start, fire and restart
	ebiten.SetKeyState(ebiten.KeyArrowLeft, true)
	now := time.Now()
	if g.last.IsZero() {
		g.last, g.prev = now, now
	}
	g.longestUpdate = max(g.longestUpdate, now.Sub(g.prev))
	g.prev = now
	if now.Sub(g.last) >= 10*time.Second {
		current := ebiten.ReadBenchmarkStats()
		n := current.Frames - g.stats.Frames
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		if n > 0 {
			println("frames=", n, " tps=", int(ebiten.ActualTPS()), " update_us/frame=", (current.Update-g.stats.Update).Microseconds()/int64(n), " draw_us/frame=", (current.Draw-g.stats.Draw).Microseconds()/int64(n), " present_us/frame=", (current.Present-g.stats.Present).Microseconds()/int64(n), " heap=", mem.HeapAlloc, " gc=", mem.NumGC, " longest_gap_us=", g.longestUpdate.Microseconds())
		}
		g.stats = current
		g.last = now
		g.longestUpdate = 0
	}
	return g.Game.Update()
}

func game() ebiten.Game { return &autoplay{Game: uboat.NewGame()} }
