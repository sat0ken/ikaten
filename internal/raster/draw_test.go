package raster

import (
	"image"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func newView(w, h int, opaque bool) View {
	b := &Buffer{Pixels: make([]uint16, w*h)}
	if !opaque {
		b.Mask = make([]byte, (w*h+7)/8)
	}
	return View{Buffer: b, Bounds: image.Rect(0, 0, w, h), Stride: w}
}
func TestIntegerMatchesGeneral(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for _, opaque := range []bool{true, false} {
		for scale := 1; scale <= 5; scale++ {
			for _, blend := range []Blend{SourceOver, Copy, Lighter} {
				for tx := -10; tx <= 12; tx += 3 {
					src := newView(7, 5, opaque)
					for n := range src.Buffer.Pixels {
						src.Buffer.Set(n, uint16(rng.Uint32()), n%3 != 0 || opaque)
					}
					fast, general := newView(23, 21, false), newView(23, 21, false)
					for n := range fast.Buffer.Pixels {
						v := uint16(rng.Uint32())
						fast.Buffer.Set(n, v, n%2 == 0)
						general.Buffer.Set(n, v, n%2 == 0)
					}
					// Nonzero destination origin and parent stride must agree on both paths.
					fast.Bounds, general.Bounds = image.Rect(3, 2, 19, 18), image.Rect(3, 2, 19, 18)
					fast.Offset, general.Offset = 49, 49
					op := Options{A: float32(scale), D: float32(scale), TX: float32(tx), TY: 4, R: 256, G: 256, Blue: 256, Alpha: 256, Blend: blend}
					draw(fast, src, op, false)
					draw(general, src, op, true)
					if !reflect.DeepEqual(fast.Buffer, general.Buffer) {
						t.Fatalf("opaque=%v scale=%d blend=%d tx=%d", opaque, scale, blend, tx)
					}
				}
			}
		}
	}
}

// sourceVariants returns the same pixels in every storage form DrawImage accepts.
func sourceVariants(rng *rand.Rand, w, h int) map[string]View {
	n := w * h
	pixels, mask := make([]uint16, n), make([]byte, (n+7)/8)
	data := make([]byte, 2*n)
	for i := range pixels {
		pixels[i] = uint16(rng.Uint32())
		data[2*i], data[2*i+1] = byte(pixels[i]), byte(pixels[i]>>8)
		if i%3 != 0 {
			mask[i>>3] |= 1 << uint(i&7)
		}
	}
	all := make([]byte, len(mask))
	for i := range all {
		all[i] = 255
	}
	view := func(b *Buffer) View { return View{Buffer: b, Bounds: image.Rect(0, 0, w, h), Stride: w} }
	return map[string]View{
		"ram-nomask":   view(&Buffer{Pixels: pixels}),
		"ram-opaque":   view(&Buffer{Pixels: pixels, Mask: all, Opaque: true}),
		"ram-masked":   view(&Buffer{Pixels: pixels, Mask: mask}),
		"flash-nomask": view(&Buffer{Data: data}),
		"flash-masked": view(&Buffer{Data: data, Mask: mask}),
	}
}
func TestIntegerMatchesGeneralForAllSources(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	scales := [][2]int{{1, 1}, {1, 2}, {2, 1}, {3, 2}, {4, 4}}
	for name, src := range sourceVariants(rng, 7, 5) {
		// A nested source must keep its parent's stride and offset.
		for _, nested := range []bool{false, true} {
			s := src
			if nested {
				s.Bounds, s.Offset = image.Rect(2, 1, 6, 4), 9
			}
			for _, scale := range scales {
				for _, blend := range []Blend{SourceOver, Copy, Lighter} {
					for _, alpha := range []int32{256, 128} {
						for _, opaqueDst := range []bool{false, true} {
							for _, tx := range []int{-7, 0, 5} {
								fast, general := newView(23, 21, false), newView(23, 21, false)
								for n := range fast.Buffer.Pixels {
									v := uint16(rng.Uint32())
									visible := opaqueDst || n%2 == 0
									fast.Buffer.Set(n, v, visible)
									general.Buffer.Set(n, v, visible)
								}
								fast.Buffer.Opaque, general.Buffer.Opaque = opaqueDst, opaqueDst
								fast.Bounds, general.Bounds = image.Rect(3, 2, 19, 18), image.Rect(3, 2, 19, 18)
								fast.Offset, general.Offset = 49, 49
								op := Options{A: float32(scale[0]), D: float32(scale[1]), TX: float32(tx), TY: 4, R: 256, G: 256, Blue: 256, Alpha: alpha, Blend: blend}
								draw(fast, s, op, false)
								draw(general, s, op, true)
								if !reflect.DeepEqual(fast.Buffer, general.Buffer) {
									t.Fatalf("%s nested=%v scale=%v blend=%d alpha=%d opaqueDst=%v tx=%d", name, nested, scale, blend, alpha, opaqueDst, tx)
								}
							}
						}
					}
				}
			}
		}
	}
}

// Power-of-two scales have exact reciprocals, so they must match the general path.
func TestScaledMatchesGeneral(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	scales := [][2]float32{{0.5, 0.5}, {4, 0.25}, {0.25, 2}, {1, 1}, {2, 0.125}, {8, 0.5}}
	offsets := [][2]float32{{0.5, 4}, {-3.25, 1.5}, {2.75, -0.5}, {5, 2}}
	for name, src := range sourceVariants(rng, 7, 5) {
		for _, scale := range scales {
			for _, offset := range offsets {
				for _, blend := range []Blend{SourceOver, Copy, Lighter} {
					for _, alpha := range []int32{256, 128} {
						fast, general := filledPair(rng)
						op := Options{A: scale[0], D: scale[1], TX: offset[0], TY: offset[1], R: 256, G: 256, Blue: 256, Alpha: alpha, Blend: blend}
						draw(fast, src, op, false)
						draw(general, src, op, true)
						if !reflect.DeepEqual(fast.Buffer, general.Buffer) {
							t.Fatalf("%s scale=%v offset=%v blend=%d alpha=%d", name, scale, offset, blend, alpha)
						}
					}
				}
			}
		}
	}
}

// Arbitrary scales may differ only where a pixel center sits on a texel edge.
func TestScaledNearlyMatchesGeneral(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	src := sourceVariants(rng, 16, 16)["ram-masked"]
	differing, total := 0, 0
	for range 2000 {
		fast, general := filledPair(rng)
		op := Options{A: 0.2 + rng.Float32()*3, D: 0.2 + rng.Float32()*8, TX: rng.Float32()*30 - 5, TY: rng.Float32()*30 - 5, R: 200, G: 256, Blue: 100, Alpha: 256}
		draw(fast, src, op, false)
		draw(general, src, op, true)
		for i := range fast.Buffer.Pixels {
			total++
			if fast.Buffer.Pixels[i] != general.Buffer.Pixels[i] {
				differing++
			}
		}
	}
	if differing*10000 > total {
		t.Fatalf("%d of %d pixels differ", differing, total)
	}
}

// Quarter turns and flips with power-of-two scales are exact in float32, so
// they must match the general path for every source and blend.
func TestAffineMatchesGeneral(t *testing.T) {
	rng := rand.New(rand.NewSource(19))
	matrices := [][4]float32{{0, -1, 1, 0}, {0, 1, -1, 0}, {-1, 0, 0, -1}, {-1, 0, 0, 1}, {1, 0, 0, -2}, {0, 2, -0.5, 0}, {1, 0.5, 0, 1}}
	offsets := [][2]float32{{12, 10}, {7.5, 3.25}, {20, 18}, {-2, 25}}
	for name, src := range sourceVariants(rng, 7, 5) {
		for _, m := range matrices {
			for _, offset := range offsets {
				for _, blend := range []Blend{SourceOver, Copy, Lighter} {
					for _, alpha := range []int32{256, 128} {
						fast, general := filledPair(rng)
						op := Options{A: m[0], B: m[1], C: m[2], D: m[3], TX: offset[0], TY: offset[1], R: 256, G: 256, Blue: 256, Alpha: alpha, Blend: blend}
						draw(fast, src, op, false)
						draw(general, src, op, true)
						if !reflect.DeepEqual(fast.Buffer, general.Buffer) {
							t.Fatalf("%s matrix=%v offset=%v blend=%d alpha=%d", name, m, offset, blend, alpha)
						}
					}
				}
			}
		}
	}
}

// Arbitrary rotations may differ only where a pixel center sits on a texel edge.
func TestAffineNearlyMatchesGeneral(t *testing.T) {
	rng := rand.New(rand.NewSource(23))
	for _, name := range []string{"ram-masked", "flash-nomask"} {
		src := sourceVariants(rng, 16, 16)[name]
		differing, total := 0, 0
		for range 2000 {
			fast, general := filledPair(rng)
			angle := rng.Float64() * 2 * math.Pi
			scale := 0.3 + rng.Float64()*3
			s, c := math.Sincos(angle)
			op := Options{A: float32(c * scale), B: float32(-s * scale), C: float32(s * scale), D: float32(c * scale),
				TX: rng.Float32()*30 - 5, TY: rng.Float32()*30 - 5, R: 256, G: 256, Blue: 256, Alpha: 256}
			if rng.Intn(2) == 0 {
				op.R, op.Blue = 200, 100
			}
			draw(fast, src, op, false)
			draw(general, src, op, true)
			for i := range fast.Buffer.Pixels {
				total++
				if fast.Buffer.Pixels[i] != general.Buffer.Pixels[i] {
					differing++
				}
			}
		}
		if differing*10000 > total {
			t.Fatalf("%s: %d of %d pixels differ", name, differing, total)
		}
		t.Logf("%s: %d of %d pixels differ", name, differing, total)
	}
}

func filledPair(rng *rand.Rand) (View, View) {
	fast, general := newView(23, 21, false), newView(23, 21, false)
	for n := range fast.Buffer.Pixels {
		v := uint16(rng.Uint32())
		fast.Buffer.Set(n, v, n%2 == 0)
		general.Buffer.Set(n, v, n%2 == 0)
	}
	fast.Bounds, general.Bounds = image.Rect(3, 2, 19, 18), image.Rect(3, 2, 19, 18)
	fast.Offset, general.Offset = 49, 49
	return fast, general
}

func TestRGB565WireOrder(t *testing.T) {
	if got := Pack(255, 0, 0); got != 0x00f8 {
		t.Fatalf("red swapped word = %04x", got)
	}
	r, g, b := Unpack(Pack(255, 255, 255))
	if r != 255 || g != 255 || b != 255 {
		t.Fatal(r, g, b)
	}
}
