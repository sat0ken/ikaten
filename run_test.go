package ikaten

import (
	"errors"
	"github.com/sat0ken/ikaten/internal/inputstate"
	"image/color"
	"testing"
	"time"
)

func TestScheduleCatchup(t *testing.T) {
	step := time.Second / 60
	s := schedule{}
	if n := s.due(step/2, step); n != 0 {
		t.Fatal(n)
	}
	if n := s.due(step/2+2*step, step); n != 3 {
		t.Fatal(n)
	}
	if n := s.due(time.Hour, step); n != 3 {
		t.Fatal(n)
	}
	if n := s.due(0, step); n != 0 {
		t.Fatal("backlog was not discarded", n)
	}
}

type fakeHardware struct {
	presented int
	fail      error
}

func (h *fakeHardware) Init() error                 { return nil }
func (h *fakeHardware) Size() (int, int)            { return 10, 10 }
func (h *fakeHardware) KeyUpdate() error            { SetKeyState(KeyZ, true); return nil }
func (h *fakeHardware) Present(screen *Image) error { h.presented++; return h.fail }

type fakeGame struct {
	updates, draws int
	failure        error
}

func (g *fakeGame) Update() error {
	g.updates++
	if !IsKeyPressed(KeyZ) || inputstate.Duration(int(KeyZ)) != g.updates {
		return errors.New("incorrect tick input")
	}
	if g.updates == 3 {
		return g.failure
	}
	return nil
}
func (g *fakeGame) Layout(w, h int) (int, int) { return 5, 5 }
func (g *fakeGame) Draw(screen *Image) {
	if screen.RGBAAt(0, 0).A != 0 {
		panic("screen not cleared")
	}
	screen.Set(0, 0, red)
	g.draws++
}
func TestRunGameTerminationAndErrors(t *testing.T) {
	previousHW, previousTPS := currentHardware, TPS()
	defer func() { SetHardware(previousHW); SetTPS(previousTPS) }()
	SetTPS(1000)
	hw := &fakeHardware{}
	SetHardware(hw)
	g := &fakeGame{failure: Termination}
	if err := RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.updates != 3 || g.draws == 0 || hw.presented != g.draws {
		t.Fatal(g, hw)
	}
	failure := errors.New("present failed")
	hw.fail = failure
	if err := RunGame(&fakeGame{}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	SetHardware(nil)
	if err := RunGame(&fakeGame{}); err == nil {
		t.Fatal("nil hardware accepted")
	}
}

type layoutOrderGame struct {
	layouts, layoutsAtFirstUpdate int
	updated                       bool
}

func (g *layoutOrderGame) Update() error {
	if !g.updated {
		g.updated = true
		g.layoutsAtFirstUpdate = g.layouts
	}
	return Termination
}
func (g *layoutOrderGame) Draw(*Image)                {}
func (g *layoutOrderGame) Layout(w, h int) (int, int) { g.layouts++; return w, h }

func TestLayoutRunsBeforeFirstUpdate(t *testing.T) {
	previousHW := currentHardware
	defer SetHardware(previousHW)
	SetHardware(&fakeHardware{})
	g := &layoutOrderGame{}
	if err := RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.layoutsAtFirstUpdate != 1 {
		t.Fatalf("Layout calls before the first Update = %d, want 1", g.layoutsAtFirstUpdate)
	}
}

// Switching sizes reuses the largest screen's memory instead of allocating.
func TestResizeScreenReusesMemory(t *testing.T) {
	big := resizeScreen(nil, 240, 240)
	small := resizeScreen(big, 120, 120)
	again := resizeScreen(small, 240, 240)
	if &small.view.Buffer.Pixels[0] != &big.view.Buffer.Pixels[0] || &again.view.Buffer.Pixels[0] != &big.view.Buffer.Pixels[0] {
		t.Fatal("a resize allocated new pixels")
	}
	if b := small.Bounds(); b.Dx() != 120 || b.Dy() != 120 || small.view.Stride != 120 {
		t.Fatalf("small screen %v stride %d", b, small.view.Stride)
	}
	if grown := resizeScreen(resizeScreen(nil, 120, 120), 240, 240); len(grown.view.Buffer.Pixels) != 240*240 {
		t.Fatal("growing past the old screen did not allocate")
	}
	// The reused screen draws like a new one.
	small.Fill(color.RGBA{R: 255, A: 255})
	if r, _, _, a := small.At(119, 119).RGBA(); r>>8 != 255 || a == 0 {
		t.Fatal("the reused screen does not draw")
	}
}
