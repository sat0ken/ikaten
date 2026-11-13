package air

import "math"

// The world is flat, in meters: x to the east, y to the north and z up. The
// sea is at z = 0.
type vec struct{ x, y, z float32 }

func (a vec) add(b vec) vec           { return vec{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a vec) sub(b vec) vec           { return vec{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a vec) mul(s float32) vec       { return vec{a.x * s, a.y * s, a.z * s} }
func (a vec) dot(b vec) float32       { return a.x*b.x + a.y*b.y + a.z*b.z }
func (a vec) length() float32         { return sqrt(a.dot(a)) }
func (a vec) flat() vec               { return vec{a.x, a.y, 0} }
func (a vec) norm() vec               { return a.mul(1 / max(a.length(), 1e-6)) }
func (a vec) cross(b vec) vec         { return vec{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x} }
func sqrt(x float32) float32          { return float32(math.Sqrt(float64(max(x, 0)))) }
func sin(x float32) float32           { return float32(math.Sin(float64(x))) }
func cos(x float32) float32           { return float32(math.Cos(float64(x))) }
func atan2(y, x float32) float32      { return float32(math.Atan2(float64(y), float64(x))) }
func abs(x float32) float32           { return max(x, -x) }
func clamp(x, lo, hi float32) float32 { return max(lo, min(hi, x)) }

// approach moves x toward target by at most step.
func approach(x, target, step float32) float32 {
	if x < target {
		return min(x+step, target)
	}
	return max(x-step, target)
}

// heading returns a horizontal unit vector for a bearing clockwise from north.
func heading(bearing float32) vec { return vec{sin(bearing), cos(bearing), 0} }

// bearingOf returns the bearing of a horizontal direction, 0 to 2π.
func bearingOf(d vec) float32 {
	b := atan2(d.x, d.y)
	if b < 0 {
		b += 2 * math.Pi
	}
	return b
}

const (
	tps     = 60
	dt      = 1.0 / tps
	gravity = 9.8
)

type kind uint8

const (
	zero    kind = iota // the Type 0 carrier fighter
	val                 // the Type 99 carrier bomber, a dive bomber
	kate                // the Type 97 carrier attack plane, a torpedo bomber
	fighter             // an enemy fighter
	bomber              // an enemy bomber
	kinds
)

// planeType holds speeds in m/s: the top level speed, the speed never to
// exceed, and the stall. Roll is the full roll rate in rad/s.
type planeType struct {
	vmax, vne, vstall float32
	roll, maxG        float32
	hp                int
	length, span      float32
	brakeDrag         float32 // added drag factor with the dive brakes out
}

var planeTypes = [kinds]planeType{
	zero:    {vmax: 147, vne: 186, vstall: 31, roll: 2.4, maxG: 7, hp: 100, length: 9, span: 12},
	val:     {vmax: 107, vne: 125, vstall: 33, roll: 1.5, maxG: 5, hp: 100, length: 10, span: 14, brakeDrag: 0.0004},
	kate:    {vmax: 105, vne: 115, vstall: 31, roll: 1.2, maxG: 4, hp: 100, length: 10, span: 16},
	fighter: {vmax: 140, vne: 190, vstall: 36, roll: 1.6, maxG: 5, hp: 12, length: 9, span: 12},
	bomber:  {vmax: 110, vne: 140, vstall: 35, roll: 1.0, maxG: 3, hp: 20, length: 10, span: 13},
}

// plane is an aircraft. f and u are unit vectors forward and up; right is
// their cross product.
type plane struct {
	kind   kind
	alive  bool
	pos    vec
	f, u   vec
	v      float32
	hp     int
	load   float32 // the load factor of the last tick, in g
	rollIn float32 // control inputs from -1 to 1, smoothed
	pullIn float32
	// throttle from 0 to 1 and the dive brakes.
	throttle float32
	brakes   bool
	// falling counts ticks since the plane was shot down; it falls burning.
	falling int
	// The AI: what it does and when it changes its mind.
	task, until int
	breakDir    float32
	gunCool     int
	burst       int
	aimErr      vec
	lastAim     vec // the aim point of the last tick, to see how it moves
	goal        vec
}

func (p *plane) right() vec { return p.f.cross(p.u) }

func newPlane(k kind, pos vec, bearing, speed float32) plane {
	return plane{kind: k, alive: true, pos: pos, f: heading(bearing), u: vec{0, 0, 1}, v: speed, hp: planeTypes[k].hp, throttle: 1}
}

// fly advances the plane one tick with roll and pull commands from -1 to 1.
// Lift along the plane's up axis and gravity bend the flight path; the stick
// sets the load factor, so a plane in a bank with the stick centered sinks
// and turns, and pulling in the bank turns harder.
func (p *plane) fly(roll, pull float32) {
	t := &planeTypes[p.kind]
	p.rollIn = approach(p.rollIn, clamp(roll, -1, 1), 0.12)
	p.pullIn = approach(p.pullIn, clamp(pull, -1, 1), 0.08)
	load := 1 + p.pullIn*(t.maxG-1)
	if p.pullIn < 0 {
		load = 1 + 2*p.pullIn
	}
	// The wings give less lift near the stall.
	lift := p.v * p.v / (t.vstall * t.vstall)
	load = clamp(load, -lift, lift)
	p.load = load
	down := vec{0, 0, -gravity}
	a := p.u.mul(gravity * load).add(down.sub(p.f.mul(down.dot(p.f))))
	w := p.f.cross(a).mul(1 / max(p.v, 10)).add(p.f.mul(p.rollIn * t.roll))
	p.f = p.f.add(w.cross(p.f).mul(dt)).norm()
	u := p.u.add(w.cross(p.u).mul(dt))
	p.u = u.sub(p.f.mul(u.dot(p.f))).norm()

	// Speed: propeller thrust falls with speed; drag grows with its square and
	// with the load; gravity speeds a dive.
	k := 6 / (t.vmax * t.vmax)
	if p.brakes {
		k += t.brakeDrag
	}
	thrust := min(6*t.vmax/max(p.v, 1), 7) * p.throttle
	// Induced drag grows with the load and is largest near the stall.
	induced := 1.5 * load * load * t.vstall * t.vstall / max(p.v*p.v, 1)
	acc := thrust - k*p.v*p.v - induced - gravity*p.f.z
	p.v = max(p.v+acc*dt, 8)
	p.pos = p.pos.add(p.f.mul(p.v * dt))
}

// bank is the bank angle: positive with the right wing down.
func (p *plane) bank() float32 { return atan2(-p.right().z, p.u.z) }

// steer returns the roll and pull commands that turn the nose toward a point,
// as a pilot does: work out the lift that would bring the nose onto the
// point in under half a second and also hold the plane up, roll the lift that way
// and pull for it. Small corrections may push instead of rolling over.
func (p *plane) steer(to vec) (roll, pull float32) {
	t := &planeTypes[p.kind]
	d := to.sub(p.pos)
	r := p.right()
	x, y, z := d.dot(r), d.dot(p.u), d.dot(p.f)
	across := sqrt(x*x + y*y)
	off := atan2(across, z)
	var ax, ay float32
	if across > 1e-3 {
		k := p.v * min(off, 1.5) / 0.4 / across
		ax, ay = x*k, y*k
	}
	ax += gravity * r.z
	ay += gravity * p.u.z
	phi := atan2(ax, ay)
	load := sqrt(ax*ax+ay*ay) / gravity
	if off < 0.5 && abs(phi) > 1.5708 {
		// Push rather than roll over.
		load = -load
		if phi > 0 {
			phi -= 3.14159
		} else {
			phi += 3.14159
		}
	}
	roll = clamp(phi*3, -1, 1)
	// Pull once the lift points about the right way.
	if load >= 1 {
		pull = (load - 1) / (t.maxG - 1) * max(0, 1-abs(phi))
	} else {
		pull = (load - 1) / 2
	}
	return roll, clamp(pull, -1, 1)
}

// level returns the roll command that levels the wings.
func (p *plane) level() float32 { return clamp(-p.bank()*2, -1, 1) }
