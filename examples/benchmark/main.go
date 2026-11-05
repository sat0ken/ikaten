//go:build ikaten_benchmark

// benchmark runs the planned stress scene for 30 minutes and prints cumulative
// stage timings and heap statistics every 10 seconds. It exits on Termination.
package main

import (
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/bounce"
	"github.com/sat0ken/ikaten/hardware"
	"runtime"
	"time"
)

type game struct {
	*bounce.Game
	start, last time.Time
	stats       ikaten.BenchmarkStats
}

func (g *game) Update() error {
	now := time.Now()
	if g.start.IsZero() {
		g.start = now
		g.last = now
	}
	if now.Sub(g.start) >= 30*time.Minute {
		return ikaten.Termination
	}
	if elapsed := now.Sub(g.last); elapsed >= 10*time.Second {
		current := ikaten.ReadBenchmarkStats()
		n := current.Frames - g.stats.Frames
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		if n > 0 {
			println("frames=", n, " tps=", int(ikaten.ActualTPS()), " fps=", int(ikaten.ActualFPS()), " update_us/frame=", (current.Update-g.stats.Update).Microseconds()/int64(n), " draw_us/frame=", (current.Draw-g.stats.Draw).Microseconds()/int64(n), " present_us/frame=", (current.Present-g.stats.Present).Microseconds()/int64(n), " heap=", mem.HeapAlloc)
		}
		g.stats = current
		g.last = now
	}
	return g.Game.Update()
}

// masked is set with -ldflags="-X main.masked=1" to measure masked sprites.
var masked string

func main() {
	configure()
	count := 15
	if hardware.LogicalWidth >= 240 {
		count = 30
	}
	ikaten.SetHardware(hardware.Device)
	scene := bounce.New(hardware.LogicalWidth, hardware.LogicalHeight, count)
	if masked != "" {
		scene.UseMaskedSprite()
	}
	if err := ikaten.RunGame(&game{Game: scene}); err != nil {
		println(err.Error())
	}
}
