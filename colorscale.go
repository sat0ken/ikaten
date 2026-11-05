package ikaten

import "image/color"

// ColorScale multiplies premultiplied RGBA channels. The zero value is identity.
type ColorScale struct{ r, g, b, a float32 }

func (c *ColorScale) R() float32               { return c.r + 1 }
func (c *ColorScale) G() float32               { return c.g + 1 }
func (c *ColorScale) B() float32               { return c.b + 1 }
func (c *ColorScale) A() float32               { return c.a + 1 }
func (c *ColorScale) Reset()                   { *c = ColorScale{} }
func (c *ColorScale) Set(r, g, b, a float32)   { c.r, c.g, c.b, c.a = r-1, g-1, b-1, a-1 }
func (c *ColorScale) SetR(r float32)           { c.r = r - 1 }
func (c *ColorScale) SetG(g float32)           { c.g = g - 1 }
func (c *ColorScale) SetB(b float32)           { c.b = b - 1 }
func (c *ColorScale) SetA(a float32)           { c.a = a - 1 }
func (c *ColorScale) Scale(r, g, b, a float32) { c.Set(c.R()*r, c.G()*g, c.B()*b, c.A()*a) }
func (c *ColorScale) ScaleAlpha(a float32)     { c.Scale(a, a, a, a) }
func (c *ColorScale) ScaleWithColor(clr color.Color) {
	r, g, b, a := clr.RGBA()
	c.Scale(float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(a)/65535)
}
func (c *ColorScale) ScaleWithColorScale(other ColorScale) {
	c.Scale(other.R(), other.G(), other.B(), other.A())
}
