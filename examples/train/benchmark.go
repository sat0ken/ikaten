//go:build ikaten_benchmark

package main

import (
	"runtime"
	"time"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/train"
)

// autoplay stays on the title, whose demo drives the whole line with the
// autopilot and starts over, so every scene is drawn without input. It prints
// timings every 10 seconds.
type autoplay struct {
	*train.Game
	last, prev    time.Time
	stats         ebiten.BenchmarkStats
	longestUpdate time.Duration // longest gap between Update calls, which shows GC pauses
}

func (g *autoplay) Update() error {
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
		if n > 0 {
			// Where Draw's time went, in microseconds a frame.
			t := train.SectionTimes()
			println("  sky=", t[0].Microseconds()/int64(n), " ground=", t[1].Microseconds()/int64(n), " stations=", t[2].Microseconds()/int64(n), " objects=", t[3].Microseconds()/int64(n), " wire=", t[4].Microseconds()/int64(n), " gauges=", t[5].Microseconds()/int64(n))
		}
		g.stats = current
		g.last = now
		g.longestUpdate = 0
	}
	return g.Game.Update()
}

func game() ebiten.Game { return &autoplay{Game: train.NewGame()} }
