package ikaten

import (
	"math"
	"testing"
)

func near(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-4 {
		t.Fatalf("got %g, want %g", got, want)
	}
}
func TestGeoMOrderInverseAndZero(t *testing.T) {
	var g GeoM
	x, y := g.Apply(2, 3)
	near(t, x, 2)
	near(t, y, 3)
	g.Translate(4, 5)
	g.Scale(2, 3)
	g.Rotate(math.Pi / 2)
	x, y = g.Apply(2, 3)
	near(t, x, -24)
	near(t, y, 12)
	original := g
	g.Invert()
	x, y = g.Apply(x, y)
	near(t, x, 2)
	near(t, y, 3)
	original.Concat(g)
	x, y = original.Apply(17, -2)
	near(t, x, 17)
	near(t, y, -2)
	g.Reset()
	g.SetElement(0, 0, 0)
	if g.IsInvertible() {
		t.Fatal("zero determinant is invertible")
	}
}
func TestColorScale(t *testing.T) {
	var c ColorScale
	c.Scale(2, 3, 4, 1)
	c.ScaleAlpha(.5)
	near(t, float64(c.R()), 1)
	near(t, float64(c.G()), 1.5)
	near(t, float64(c.B()), 2)
	near(t, float64(c.A()), .5)
	c.Reset()
	near(t, float64(c.R()), 1)
	near(t, float64(c.A()), 1)
}
