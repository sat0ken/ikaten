package ikaten

import (
	"errors"
	"github.com/sat0ken/ikaten/internal/inputstate"
	"github.com/sat0ken/ikaten/internal/raster"
	"image"
	"time"
)

var targetTPS = 60
var actualTPS, actualFPS float64
var running bool

func SetTPS(tps int) {
	if tps <= 0 || tps > 1000 {
		panic("ikaten: TPS must be between 1 and 1000")
	}
	targetTPS = tps
}
func TPS() int           { return targetTPS }
func ActualTPS() float64 { return actualTPS }
func ActualFPS() float64 { return actualFPS }

// schedule bounds catch-up to three ticks. Excess time is discarded. Running
// several updates before drawing skips intermediate frames, keeping game time
// independent of a slow SPI transfer up to that catch-up limit.
type schedule struct{ accumulated time.Duration }

func (s *schedule) due(elapsed, step time.Duration) int {
	if elapsed < 0 {
		elapsed = 0
	}
	s.accumulated += min(elapsed, step*4)
	n := int(s.accumulated / step)
	if n > 3 {
		n = 3
		s.accumulated %= step
	} else {
		s.accumulated -= time.Duration(n) * step
	}
	return n
}

func layout(game Game, width, height int) (int, int, error) {
	w, h := game.Layout(width, height)
	if w <= 0 || h <= 0 || w > width || h > height {
		return 0, 0, errors.New("ikaten: Layout must fit the physical display at integer scale")
	}
	return w, h, nil
}

func RunGame(game Game) error {
	if game == nil {
		return errors.New("ikaten: nil game")
	}
	if running {
		return errors.New("ikaten: RunGame is already active")
	}
	hw := currentHardware
	if hw == nil {
		return errors.New("ikaten: call SetHardware before RunGame")
	}
	running = true
	defer func() { running = false }()
	inputstate.Reset()
	actualTPS, actualFPS = 0, 0
	resetBenchmark()
	if err := hw.Init(); err != nil {
		return err
	}
	if closer, ok := hw.(interface{ Close() }); ok {
		defer closer.Close()
	}
	width, height := hw.Size()
	if width <= 0 || height <= 0 {
		return errors.New("ikaten: invalid hardware size")
	}
	// Ebitengine guarantees Layout runs before the first Update.
	if _, _, err := layout(game, width, height); err != nil {
		return err
	}
	var screen *Image
	previous := time.Now()
	metricStart := previous
	updates, frames := 0, 0
	sched := schedule{accumulated: time.Second / time.Duration(targetTPS)}
	for {
		now := time.Now()
		step := time.Second / time.Duration(targetTPS)
		ticks := sched.due(now.Sub(previous), step)
		previous = now
		if ticks == 0 {
			time.Sleep(step - sched.accumulated)
			continue
		}
		if err := hw.KeyUpdate(); err != nil {
			return err
		}
		begin := benchmarkStart()
		for n := 0; n < ticks; n++ {
			inputstate.Advance()
			if err := game.Update(); err != nil {
				if errors.Is(err, Termination) {
					return nil
				}
				return err
			}
			updates++
		}
		benchmarkUpdate(begin)
		w, h, err := layout(game, width, height)
		if err != nil {
			return err
		}
		if screen == nil || screen.view.Bounds.Dx() != w || screen.view.Bounds.Dy() != h {
			screen = resizeScreen(screen, w, h)
		}
		begin = benchmarkStart()
		screen.Clear()
		game.Draw(screen)
		benchmarkDraw(begin)
		begin = benchmarkStart()
		err = hw.Present(screen)
		benchmarkPresent(begin)
		if err != nil {
			return err
		}
		frames++
		end := time.Now()
		if elapsed := end.Sub(metricStart); elapsed >= time.Second {
			actualTPS = float64(updates) / elapsed.Seconds()
			actualFPS = float64(frames) / elapsed.Seconds()
			updates, frames = 0, 0
			metricStart = end
		}
	}
}

// resizeScreen returns a screen of a new size, reusing the old screen's memory
// when it is large enough. A launcher that switches between games of different
// sizes then keeps only the largest screen. Allocating a new one instead relies
// on the garbage collector freeing the old screen, and TinyGo's conservative
// collector can keep it alive through a stale word on the stack: on the badge,
// going from a 240x240 game to the menu and into another 240x240 game ran out
// of memory.
func resizeScreen(old *Image, w, h int) *Image {
	n := w * h
	if old == nil || cap(old.view.Buffer.Pixels) < n || cap(old.view.Buffer.Mask) < (n+7)/8 {
		return NewImage(w, h)
	}
	b := old.view.Buffer
	buffer := &raster.Buffer{Pixels: b.Pixels[:n], Mask: b.Mask[:(n+7)/8]}
	return &Image{view: raster.View{Buffer: buffer, Bounds: image.Rect(0, 0, w, h), Stride: w}}
}
