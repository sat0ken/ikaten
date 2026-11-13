package train

import (
	"testing"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/hardware"
)

func playing() *Game {
	g := NewGame()
	g.setMode(modePlay)
	return g
}

// autoRun drives the whole line with the autopilot, sounding the horn at the
// whistle posts, and departs as soon as the doors open.
func autoRun(t *testing.T, g *Game) {
	t.Helper()
	for range 60 * tps * 20 {
		switch g.mode {
		case modePlay:
			g.notch = g.autopilot()
			if g.whistleDue > 0 {
				g.soundHorn()
			}
			g.step()
		case modeStation:
			g.depart()
		case modeEnd:
			return
		}
	}
	t.Fatalf("the run did not end: mode %d pos %.1f v %.1f next %d notch %d", g.mode, g.pos, g.v*3.6, g.next, g.notch)
}

// The autopilot stops at every mark, keeps the limits and the timetable, and
// so shows that a driver can.
func TestAutopilotRun(t *testing.T) {
	g := playing()
	autoRun(t, g)
	for i, r := range g.results[1:] {
		st := stations[i+1]
		t.Logf("%-9s stop %+.2fm delay %+ds score %d (stop %d time %d ride %d speed %d horn %d)",
			st.name, r.stopError, r.delay, r.total, r.stopPts, r.timePts, r.ridePts, r.speedPts, r.hornPts)
		if r.stopError < -0.25 || r.stopError > 0.25 || r.overrun {
			t.Errorf("%s: stopped %.2fm from the mark", st.name, r.stopError)
		}
		if r.delay > 0 || r.delay < -8 {
			t.Errorf("%s: %+d s from the timetable", st.name, r.delay)
		}
		if r.speedPts != 0 || r.ridePts != 20 || r.hornPts < 0 {
			t.Errorf("%s: penalties %+v", st.name, r)
		}
	}
	if g.rank() != "A" {
		t.Errorf("rank %s, total %d", g.rank(), g.total())
	}
}

func TestPoints(t *testing.T) {
	for _, r := range "SABCD" {
		if imageGlyph[r] == nil {
			t.Errorf("no glyph for rank %c", r)
		}
	}
	for _, c := range []struct {
		err  float32
		want int
	}{{0, 50}, {0.25, 50}, {-0.25, 50}, {1, 42}, {-3, 26}, {6, 2}, {10, 0}} {
		if got := stopPoints(c.err); got != c.want {
			t.Errorf("stopPoints(%v) = %d, want %d", c.err, got, c.want)
		}
	}
	for _, c := range []struct{ delay, want int }{{0, 30}, {1, 30}, {-1, 30}, {2, 24}, {-5, 15}, {10, 0}, {60, 0}} {
		if got := timePoints(c.delay); got != c.want {
			t.Errorf("timePoints(%d) = %d, want %d", c.delay, got, c.want)
		}
	}
}

// The brakes take about a second to build up.
func TestBrakesBuildUp(t *testing.T) {
	g := playing()
	g.notch, g.brake, g.v = 0, 0, 60/3.6
	g.setNotch(-maxBrake)
	for range tps / 4 {
		g.step()
	}
	if g.brake > brakeStep*maxBrake/2 {
		t.Fatalf("after a quarter second the brakes give %.2f km/h/s", g.brake)
	}
	for range tps {
		g.step()
	}
	if g.brake != brakeStep*maxBrake {
		t.Fatalf("after 1.25 s the brakes give %.2f km/h/s", g.brake)
	}
}

// From 90 km/h, full service braking stops the train in about 300 m.
func TestBrakingDistance(t *testing.T) {
	g := playing()
	g.pos, g.notch, g.brake, g.v = 1500, -maxBrake, brakeStep*maxBrake, 90/3.6
	for g.v > 0 {
		g.physics()
	}
	if d := g.pos - 1500; d < 250 || d > 330 {
		t.Fatalf("stopped in %.0f m", d)
	}
}

// A raised limit applies when the rear of the train has passed its start.
func TestLimitForWholeTrain(t *testing.T) {
	g := playing()
	g.pos = 2460 // just past the end of the 60 km/h curve
	if g.limitNow() != 60 {
		t.Fatalf("limit %d with the rear still in the curve", g.limitNow())
	}
	g.pos = 2450 + trainLength + 1
	if g.limitNow() != 90 {
		t.Fatalf("limit %d with the whole train past", g.limitNow())
	}
	g.pos = 2101 // a lower limit applies at once
	if g.limitNow() != 60 {
		t.Fatalf("limit %d at the start of the curve", g.limitNow())
	}
}

// Running 10 km/h over the limit trips the automatic brake, which holds
// until the train stops and costs points.
func TestATS(t *testing.T) {
	g := playing()
	g.notch = maxPower
	for range 60 * tps {
		g.step()
		if g.ats {
			break
		}
	}
	if !g.ats || g.notch != notchEB || g.v*3.6 < 75 {
		t.Fatalf("ats=%v notch=%d at %.1f km/h", g.ats, g.notch, g.v*3.6)
	}
	g.setNotch(maxPower)
	if g.notch != notchEB {
		t.Fatal("the handle moved while the automatic brake held")
	}
	for g.v > 0 {
		g.step()
	}
	if g.ats {
		t.Fatal("the automatic brake held after the stop")
	}
	g.setNotch(1)
	if g.notch != 1 {
		t.Fatal("the handle did not move after the stop")
	}
	// Stopped short of the station, the train drives on and is scored there.
	for g.mode == modePlay {
		g.notch = g.autopilot()
		g.step()
	}
	if r := g.results[1]; r.speedPts > -20 || !g.atsTripped {
		t.Fatalf("speed points %d", r.speedPts)
	}
}

// Passing the mark by more than the window is an overrun: no stop points and
// a penalty, scored when the train stops.
func TestOverrun(t *testing.T) {
	g := playing()
	g.pos, g.v, g.notch, g.brake = 1200, 70/3.6, 0, 0
	for g.mode == modePlay {
		if g.pos > stations[1].stop+stopWindow+5 {
			g.notch = notchEB
		}
		g.step()
	}
	r := g.results[1]
	if !r.overrun || r.stopPts != 0 || r.speedPts > -20 || r.ridePts > 10 {
		t.Fatalf("result %+v", r)
	}
}

// The terminal's buffer stop ends a run that does not stop.
func TestBufferStop(t *testing.T) {
	g := playing()
	g.next = len(stations) - 1
	last := stations[g.next]
	g.pos, g.v, g.notch, g.brake = last.stop-50, 40/3.6, 0, 0
	for range 30 * tps {
		g.step()
		if g.mode != modePlay {
			break
		}
	}
	if g.mode != modeStation || g.pos != last.stop+bufferStop || g.v != 0 || !g.results[g.next].overrun {
		t.Fatalf("mode %d at %.1f, %.1f km/h", g.mode, g.pos, g.v*3.6)
	}
	g.depart()
	if g.mode != modeEnd {
		t.Fatalf("mode %d after the terminal", g.mode)
	}
}

// Stopping short of the platform keeps the doors shut.
func TestStopShort(t *testing.T) {
	g := playing()
	g.pos, g.v, g.notch, g.brake = 1180, 30/3.6, notchEB, emergency
	for g.v > 0 {
		g.step()
	}
	if g.mode != modePlay || g.message != msgShort {
		t.Fatalf("mode %d message %d at %.0f m", g.mode, g.message, g.pos)
	}
}

// Sounding the horn after a whistle post earns points; not sounding costs.
func TestWhistle(t *testing.T) {
	for _, sound := range []bool{true, false} {
		g := playing()
		g.next = 2
		post := crossings[0] - whistleAhead
		g.pos, g.v, g.notch, g.brake = post-20, 60/3.6, 0, 0
		for g.pos < post+whistleWindow+10 {
			if sound && g.pos > post+30 {
				g.soundHorn()
			}
			g.step()
		}
		for g.mode == modePlay {
			g.notch = g.autopilot()
			g.step()
		}
		want := -5
		if sound {
			want = 5
		}
		if got := g.results[2].hornPts; got != want {
			t.Errorf("sound=%v: horn points %d, want %d", sound, got, want)
		}
	}
}

// Running late at a station shortens the stop: the doors close after the
// dwell time, not at the scheduled departure.
func TestDepartLate(t *testing.T) {
	g := playing()
	g.mode, g.next = modeStation, 1
	g.clock = (stations[1].arrive + 30) * tps
	g.depart()
	if want := (stations[1].arrive + 30 + dwellSeconds) * tps; g.clock != want {
		t.Fatalf("clock %d, want %d", g.clock, want)
	}
	g.mode, g.next = modeStation, 1
	g.clock = (stations[1].arrive - 30) * tps
	g.depart()
	if want := (stations[1].arrive + dwellSeconds) * tps; g.clock != want {
		t.Fatalf("early: clock %d, want %d", g.clock, want)
	}
}

// keyDriver plays through RunGame with key presses, as a person would:
// it starts from the title and moves the handle toward the autopilot's.
type keyDriver struct {
	*Game
	t    *testing.T
	tick int
}

func (d *keyDriver) Update() error {
	d.tick++
	up, down, a := false, false, false
	switch d.mode {
	case modeTitle:
		a = d.tick%2 == 0
	case modePlay:
		want := d.autopilot()
		// A press moves one notch; release between presses.
		if d.tick%2 == 0 {
			up, down = want > d.notch, want < d.notch
		}
	case modeStation:
		return ebiten.Termination
	}
	ebiten.SetKeyState(ebiten.KeyArrowUp, up)
	ebiten.SetKeyState(ebiten.KeyArrowDown, down)
	ebiten.SetKeyState(ebiten.KeyZ, a)
	if d.tick > 200*tps {
		d.t.Error("no station after 200 s")
		return ebiten.Termination
	}
	return d.Game.Update()
}

func TestKeysDrive(t *testing.T) {
	ebiten.SetTPS(1000)
	defer ebiten.SetTPS(60)
	ebiten.SetHardware(&hardware.Headless{Width: 240, Height: 240})
	d := &keyDriver{Game: NewGame(), t: t}
	if err := ebiten.RunGame(d); err != nil {
		t.Fatal(err)
	}
	for _, k := range []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyZ} {
		ebiten.SetKeyState(k, false)
	}
	r := d.results[1]
	t.Logf("by keys: stop %+.2f m, delay %+d s, total %d", r.stopError, r.delay, r.total)
	if d.mode != modeStation || r.stopError < -1 || r.stopError > 1 {
		t.Fatalf("mode %d, stop %+.2f m", d.mode, r.stopError)
	}
}

// Everything beside the line fits the list of objects to draw, all along it.
func TestObjectsFit(t *testing.T) {
	g := playing()
	most := 0
	for pos := float32(-50); pos < lineEnd; pos += 3 {
		g.pos = pos
		g.collectObjects()
		most = max(most, g.view.n)
	}
	t.Logf("at most %d objects", most)
	if most >= maxObjects {
		t.Fatalf("%d objects fill the list of %d", most, maxObjects)
	}
}

// The title's demo drives the whole line and starts over.
func TestTitleDemoLoops(t *testing.T) {
	g := NewGame()
	stops := 0
	for range 20 * 60 * tps {
		before := g.next
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
		if g.next != before {
			stops++
		}
		if stops > len(stations) {
			break
		}
	}
	if stops <= len(stations) || g.mode != modeTitle {
		t.Fatalf("%d stops in 20 minutes", stops)
	}
}
