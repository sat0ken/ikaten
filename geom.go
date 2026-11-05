package ikaten

import "math"

// GeoM is an affine matrix. Its zero value is the identity. Operations are
// applied in call order. Storage uses float32 on microcontrollers.
type GeoM struct{ a, b, c, d, tx, ty float32 }

func (g *GeoM) elements() (a, b, c, d, tx, ty float32) { return g.a + 1, g.b, g.c, g.d + 1, g.tx, g.ty }
func (g *GeoM) Reset()                                 { *g = GeoM{} }
func (g *GeoM) Apply(x, y float64) (float64, float64) {
	a, b, c, d, tx, ty := g.elements()
	return float64(a)*x + float64(b)*y + float64(tx), float64(c)*x + float64(d)*y + float64(ty)
}
func (g *GeoM) Translate(tx, ty float64) { g.tx += float32(tx); g.ty += float32(ty) }
func (g *GeoM) Scale(x, y float64) {
	a, b, c, d, tx, ty := g.elements()
	sx, sy := float32(x), float32(y)
	g.a, g.b, g.c, g.d, g.tx, g.ty = a*sx-1, b*sx, c*sy, d*sy-1, tx*sx, ty*sy
}
func (g *GeoM) Rotate(theta float64) {
	s, c := math.Sincos(theta)
	g.Concat(GeoM{a: float32(c) - 1, b: -float32(s), c: float32(s), d: float32(c) - 1})
}
func (g *GeoM) Skew(x, y float64) { g.Concat(GeoM{b: float32(math.Tan(x)), c: float32(math.Tan(y))}) }
func (g *GeoM) Concat(other GeoM) {
	a, b, c, d, tx, ty := g.elements()
	e, f, h, i, u, v := other.elements()
	g.a, g.b, g.c, g.d, g.tx, g.ty = e*a+f*c-1, e*b+f*d, h*a+i*c, h*b+i*d-1, e*tx+f*ty+u, h*tx+i*ty+v
}
func (g *GeoM) Element(i, j int) float64 {
	a, b, c, d, tx, ty := g.elements()
	if i == 0 {
		switch j {
		case 0:
			return float64(a)
		case 1:
			return float64(b)
		case 2:
			return float64(tx)
		}
	}
	if i == 1 {
		switch j {
		case 0:
			return float64(c)
		case 1:
			return float64(d)
		case 2:
			return float64(ty)
		}
	}
	panic("ikaten: invalid GeoM element")
}
func (g *GeoM) SetElement(i, j int, element float64) {
	v := float32(element)
	if i == 0 {
		switch j {
		case 0:
			g.a = v - 1
			return
		case 1:
			g.b = v
			return
		case 2:
			g.tx = v
			return
		}
	}
	if i == 1 {
		switch j {
		case 0:
			g.c = v
			return
		case 1:
			g.d = v - 1
			return
		case 2:
			g.ty = v
			return
		}
	}
	panic("ikaten: invalid GeoM element")
}
func (g *GeoM) IsInvertible() bool {
	a, b, c, d, _, _ := g.elements()
	det := a*d - b*c
	return det != 0 && !math.IsNaN(float64(det)) && !math.IsInf(float64(det), 0)
}
func (g *GeoM) Invert() {
	if !g.IsInvertible() {
		panic("ikaten: singular GeoM")
	}
	a, b, c, d, tx, ty := g.elements()
	det := a*d - b*c
	g.a, g.b, g.c, g.d, g.tx, g.ty = d/det-1, -b/det, -c/det, a/det-1, (b*ty-d*tx)/det, (c*tx-a*ty)/det
}
