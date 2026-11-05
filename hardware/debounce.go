package hardware

import "time"

type debouncer struct {
	stable, candidate bool
	since             time.Time
}

func (d *debouncer) update(raw bool, now time.Time, delay time.Duration) bool {
	if raw != d.candidate {
		d.candidate = raw
		d.since = now
	}
	if d.candidate != d.stable && now.Sub(d.since) >= delay {
		d.stable = d.candidate
	}
	return d.stable
}

// axisDirection adds hysteresis around the joystick's calibrated center.
func axisDirection(value, previous, press, release int) int {
	if value >= press {
		return 1
	}
	if value <= -press {
		return -1
	}
	if previous > 0 && value > release {
		return 1
	}
	if previous < 0 && value < -release {
		return -1
	}
	return 0
}
