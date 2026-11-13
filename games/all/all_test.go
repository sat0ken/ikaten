package all

import (
	"runtime"
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

// script drives the launcher through RunGame, which updates inpututil: it
// opens every game in turn from the menu, plays it a little, and holds A and B
// to return.
type script struct {
	*Game
	t      *testing.T
	tick   int
	next   int // the entry to open next
	phase  int // ticks spent in the current step
	heap   [len(entries) + 1]uint64
	failed bool
}

func (s *script) press(key ebiten.Key, on bool) { ebiten.SetKeyState(key, on) }

func (s *script) Update() error {
	s.tick++
	s.phase++
	for _, k := range []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeySpace, ebiten.KeyZ, ebiten.KeyX, ebiten.KeyArrowLeft} {
		s.press(k, false)
	}
	switch {
	case s.Game.game == nil && s.next == len(entries):
		s.heap[len(entries)] = heapAfterGC()
		return ebiten.Termination
	case s.Game.game == nil && s.cursor != s.next:
		s.press(ebiten.KeyArrowDown, s.phase%2 == 0)
	case s.Game.game == nil:
		if s.phase%2 == 0 {
			s.heap[s.next] = heapAfterGC()
			s.press(ebiten.KeySpace, true)
			s.phase = 0
		}
	case s.phase < 300:
		// Play: start from the title and press now and then.
		s.press(ebiten.KeySpace, s.phase%30 == 0)
		s.press(ebiten.KeyArrowLeft, s.phase%60 < 30)
	default:
		s.press(ebiten.KeyZ, true)
		s.press(ebiten.KeyX, true)
		if s.phase > 300+holdTicks+5 {
			s.t.Errorf("%s: still running after holding A and B", entries[s.next].name)
			return ebiten.Termination
		}
	}
	before := s.Game.game != nil
	if err := s.Game.Update(); err != nil {
		return err
	}
	if before && s.Game.game == nil {
		s.next++
		s.phase = 0
	}
	if s.tick > 60*120 {
		s.t.Error("the script did not finish")
		return ebiten.Termination
	}
	return nil
}

func heapAfterGC() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

func TestEveryGameStartsAndReturns(t *testing.T) {
	ebiten.SetTPS(1000)
	defer ebiten.SetTPS(60)
	ebiten.SetHardware(&hardware.Headless{Width: 240, Height: 240})
	s := &script{Game: NewGame(), t: t}
	if err := ebiten.RunGame(s); err != nil {
		t.Fatal(err)
	}
	if s.next != len(entries) {
		t.Fatalf("opened %d of %d games", s.next, len(entries))
	}
	if w, h := s.Layout(240, 240); w != ScreenWidth || h != ScreenHeight {
		t.Fatalf("menu layout %dx%d", w, h)
	}
	if !ebiten.IsScreenClearedEveryFrame() {
		t.Fatal("back on the menu, the screen is not cleared every frame")
	}
	for i, e := range entries {
		t.Logf("heap on the menu before %-7s %7d B", e.name, s.heap[i])
	}
	t.Logf("heap on the menu at the end    %7d B", s.heap[len(entries)])
	// The screen keeps the memory of the largest game played (FLAPPY's
	// 240x240), so from the next game on the heap must stay level.
	afterFlappy := 3
	if entries[afterFlappy-1].name != "FLAPPY" {
		t.Fatalf("entry %d is %s", afterFlappy-1, entries[afterFlappy-1].name)
	}
	if grown := int64(s.heap[len(entries)]) - int64(s.heap[afterFlappy]); grown > 16<<10 {
		t.Fatalf("the heap grew %d B after the largest screen", grown)
	}
}
