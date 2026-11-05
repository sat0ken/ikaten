// Package vector provides aliased lines, filled rectangles and circles. The
// antialias argument must be false: binary-alpha images cannot represent
// antialiased edges. Use small sprites for complex or repeated shapes.
package vector

import (
	"github.com/sat0ken/ikaten"
	"image/color"
	"math"
)

func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: byte(r >> 8), G: byte(g >> 8), B: byte(b >> 8), A: byte(a >> 8)}
}
func check(aa bool) {
	if aa {
		panic("ikaten/vector: antialias is unsupported")
	}
}
func valid(values ...float32) bool {
	for _, v := range values {
		if v != v || v > 1e6 || v < -1e6 {
			return false
		}
	}
	return true
}
func limits(dst *ikaten.Image, x0, y0, x1, y1 float64) (int, int, int, int) {
	b := dst.Bounds()
	return int(math.Max(float64(b.Min.X), math.Min(float64(b.Max.X), math.Floor(x0)))), int(math.Max(float64(b.Min.Y), math.Min(float64(b.Max.Y), math.Floor(y0)))), int(math.Max(float64(b.Min.X), math.Min(float64(b.Max.X), math.Ceil(x1)))), int(math.Max(float64(b.Min.Y), math.Min(float64(b.Max.Y), math.Ceil(y1))))
}

// DrawFilledRect is the name used before Ebitengine v2.9.
//
// Deprecated: Use FillRect.
func DrawFilledRect(dst *ikaten.Image, x, y, width, height float32, clr color.Color, antialias bool) {
	FillRect(dst, x, y, width, height, clr, antialias)
}

// DrawFilledCircle is the name used before Ebitengine v2.9.
//
// Deprecated: Use FillCircle.
func DrawFilledCircle(dst *ikaten.Image, cx, cy, radius float32, clr color.Color, antialias bool) {
	FillCircle(dst, cx, cy, radius, clr, antialias)
}
func FillRect(dst *ikaten.Image, x, y, width, height float32, clr color.Color, antialias bool) {
	check(antialias)
	if width <= 0 || height <= 0 || !valid(x, y, width, height) {
		return
	}
	c := rgba(clr)
	x0, y0, x1, y1 := limits(dst, float64(x), float64(y), float64(x+width), float64(y+height))
	left, top, right, bottom := int64(x*256), int64(y*256), int64((x+width)*256), int64((y+height)*256)
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			if int64(px*256+128) >= left && int64(px*256+128) < right && int64(py*256+128) >= top && int64(py*256+128) < bottom {
				dst.SetRGBA(px, py, c)
			}
		}
	}
}
func FillCircle(dst *ikaten.Image, cx, cy, radius float32, clr color.Color, antialias bool) {
	check(antialias)
	if radius <= 0 || !valid(cx, cy, radius) {
		return
	}
	c := rgba(clr)
	x0, y0, x1, y1 := limits(dst, float64(cx-radius), float64(cy-radius), float64(cx+radius), float64(cy+radius))
	// The shape helpers run at pixel centers; all coverage arithmetic is fixed
	// point, avoiding a floating-point operation per pixel on RP2040.
	fx, fy, fr := int64(cx*256), int64(cy*256), int64(radius*256)
	for y := y0; y < y1; y++ {
		dy := int64(y*256+128) - fy
		for x := x0; x < x1; x++ {
			dx := int64(x*256+128) - fx
			if dx*dx+dy*dy <= fr*fr {
				dst.SetRGBA(x, y, c)
			}
		}
	}
}
func StrokeLine(dst *ikaten.Image, x0, y0, x1, y1, strokeWidth float32, clr color.Color, antialias bool) {
	check(antialias)
	if strokeWidth <= 0 || !valid(x0, y0, x1, y1, strokeWidth) {
		return
	}
	c := rgba(clr)
	ax, ay, bx, by := int64(x0*256), int64(y0*256), int64(x1*256), int64(y1*256)
	vx, vy := bx-ax, by-ay
	length := vx*vx + vy*vy
	if length == 0 {
		return
	}
	half := float64(strokeWidth) / 2
	left, top, right, bottom := limits(dst, math.Min(float64(x0), float64(x1))-half, math.Min(float64(y0), float64(y1))-half, math.Max(float64(x0), float64(x1))+half, math.Max(float64(y0), float64(y1))+half)
	// Take one square root per line. The coverage loop uses integer cross
	// products, without division or floating-point operations per pixel.
	limit := int64(float64(strokeWidth) * 128 * math.Sqrt(float64(length)))
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			dx, dy := int64(x*256+128)-ax, int64(y*256+128)-ay
			projection := dx*vx + dy*vy
			if projection < 0 || projection > length {
				continue
			}
			cross := dx*vy - dy*vx
			if cross >= -limit && cross <= limit {
				dst.SetRGBA(x, y, c)
			}
		}
	}
}
