//go:build ikaten_benchmark

package main

import (
	"runtime"
	"time"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/air"
)

// autoplay starts the fighter mission and flies straight, the enemy
// attacking, then starts again from the select screen. It prints timings
// every 10 seconds. It presses Space: boards read their buttons into KeyZ
// and KeyX every tick, which would hide autoplay's presses.
type autoplay struct {
	*air.Game
	ticks         int
	last, prev    time.Time
	stats         ebiten.BenchmarkStats
	longestUpdate time.Duration // longest gap between Update calls, which shows GC pauses
}

func (g *autoplay) Update() error {
	g.ticks++
	ebiten.SetKeyState(ebiten.KeySpace, g.ticks%90 == 0) // start, fire and restart
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
			t := air.SectionTimes()
			println("  sky=", t[0].Microseconds()/int64(n), " items=", t[1].Microseconds()/int64(n), " sight=", t[2].Microseconds()/int64(n), " panel=", t[3].Microseconds()/int64(n), " boxes=", t[4].Microseconds()/int64(n))
		}
		g.stats = current
		g.last = now
		g.longestUpdate = 0
	}
	return g.Game.Update()
}

func game() ebiten.Game { return &autoplay{Game: air.NewGame()} }
