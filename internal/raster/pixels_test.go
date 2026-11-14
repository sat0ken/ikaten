package raster

import (
	"image"
	"testing"
)

func TestFillAndCopyPixels(t *testing.T) {
	buf := make([]uint16, 64)
	src := make([]uint16, 64)
	for i := range src {
		src[i] = uint16(i*977 + 1)
	}
	for start := range 4 {
		for n := range 20 {
			for _, from := range []int{0, 1, 2, 3} {
				clear(buf)
				copyPixels(buf[start:start+n], src[from:from+n])
				for i := range buf {
					want := uint16(0)
					if i >= start && i < start+n {
						want = src[from+i-start]
					}
					if buf[i] != want {
						t.Fatalf("copy start %d n %d from %d: pixel %d is %d, want %d", start, n, from, i, buf[i], want)
					}
				}
			}
			clear(buf)
			fillPixels(buf[start:start+n], 0xabcd)
			for i := range buf {
				want := uint16(0)
				if i >= start && i < start+n {
					want = 0xabcd
				}
				if buf[i] != want {
					t.Fatalf("fill start %d n %d: pixel %d is %x", start, n, i, buf[i])
				}
			}
		}
	}
}

// Masked sources whose mask has whole bytes set or clear take the byte at a
// time path; they must draw as the general path does, at every offset, onto
// opaque and see-through destinations, copying and blending.
func TestMaskedBytes(t *testing.T) {
	const w, h = 40, 3
	pixels := make([]uint16, w*h)
	mask := make([]byte, (w*h+7)/8)
	for i := range pixels {
		pixels[i] = uint16(i*7919 + 3)
	}
	for i := range mask {
		mask[i] = [...]byte{0xff, 0, 0xff, 0x5a, 0xff, 0}[i%6]
	}
	src := View{Buffer: &Buffer{Pixels: pixels, Mask: mask}, Bounds: image.Rect(0, 0, w, h), Stride: w}
	for _, opaqueDst := range []bool{false, true} {
		for _, blend := range []Blend{SourceOver, Copy} {
			for tx := -9; tx < 12; tx++ {
				var got, want View
				for k, v := range []*View{&got, &want} {
					dp := make([]uint16, 64*8)
					dm := make([]byte, 64*8/8)
					for i := range dp {
						dp[i] = 0x1111
					}
					if opaqueDst {
						for i := range dm {
							dm[i] = 0xff
						}
					}
					*v = View{Buffer: &Buffer{Pixels: dp, Mask: dm, Opaque: opaqueDst}, Bounds: image.Rect(0, 0, 64, 8), Stride: 64}
					o := Options{A: 1, D: 1, TX: float32(tx), TY: 2, R: 256, G: 256, Blue: 256, Alpha: 256, Blend: blend}
					draw(*v, src, o, k == 1)
				}
				for i := range got.Buffer.Pixels {
					if got.Buffer.Pixels[i] != want.Buffer.Pixels[i] || got.Buffer.Visible(i) != want.Buffer.Visible(i) {
						t.Fatalf("opaque %v blend %d tx %d: pixel %d is %x, want %x", opaqueDst, blend, tx, i, got.Buffer.Pixels[i], want.Buffer.Pixels[i])
					}
				}
			}
		}
	}
}

// Images flipped left to right take their own path; it must draw as the
// general path does, clipped on either side, scaled down, from opaque and
// masked images in RAM and in flash data, onto opaque and see-through
// destinations, copying and blending.
func TestMirrored(t *testing.T) {
	const w, h = 11, 4
	pixels := make([]uint16, w*h)
	data := make([]byte, 2*w*h)
	mask := make([]byte, (w*h+7)/8)
	for i := range pixels {
		pixels[i] = uint16(i*7919 + 3)
		data[2*i], data[2*i+1] = byte(pixels[i]), byte(pixels[i]>>8)
	}
	for i := range mask {
		mask[i] = byte(i*37 + 0x5b)
	}
	sources := []*Buffer{
		{Pixels: pixels, Mask: mask},
		{Pixels: pixels, Opaque: true},
		{Data: data, Mask: mask},
	}
	for si, sb := range sources {
		src := View{Buffer: sb, Bounds: image.Rect(0, 0, w, h), Stride: w}
		for _, opaqueDst := range []bool{false, true} {
			for _, blend := range []Blend{SourceOver, Copy} {
				for _, d := range []float32{1, 2} {
					for tx := -3; tx < 24; tx++ {
						var got, want View
						for k, v := range []*View{&got, &want} {
							dp := make([]uint16, 20*12)
							dm := make([]byte, 20*12/8)
							for i := range dp {
								dp[i] = 0x1111
							}
							if opaqueDst {
								for i := range dm {
									dm[i] = 0xff
								}
							}
							*v = View{Buffer: &Buffer{Pixels: dp, Mask: dm, Opaque: opaqueDst}, Bounds: image.Rect(0, 0, 20, 12), Stride: 20}
							o := Options{A: -1, D: d, TX: float32(tx), TY: 1, R: 256, G: 256, Blue: 256, Alpha: 256, Blend: blend}
							draw(*v, src, o, k == 1)
						}
						for i := range got.Buffer.Pixels {
							if got.Buffer.Pixels[i] != want.Buffer.Pixels[i] || got.Buffer.Visible(i) != want.Buffer.Visible(i) {
								t.Fatalf("source %d opaque %v blend %d scale %v tx %d: pixel %d is %x, want %x", si, opaqueDst, blend, d, tx, i, got.Buffer.Pixels[i], want.Buffer.Pixels[i])
							}
						}
					}
				}
			}
		}
	}
}

// A solid pixel drawn rotated, skewed and scaled fills a parallelogram a row
// at a time; it must cover exactly the pixels the per-pixel loop does. The
// same pixel with a mask that shows it takes that loop.
func TestSolidParallelogram(t *testing.T) {
	solid := View{Buffer: &Buffer{Pixels: []uint16{0xbeef}, Opaque: true}, Bounds: image.Rect(0, 0, 1, 1), Stride: 1}
	masked := View{Buffer: &Buffer{Pixels: []uint16{0xbeef}, Mask: []byte{1}}, Bounds: image.Rect(0, 0, 1, 1), Stride: 1}
	rng := uint32(12345)
	random := func() float32 {
		rng ^= rng << 13
		rng ^= rng >> 17
		rng ^= rng << 5
		return float32(rng%20000)/10000 - 1
	}
	for k := range 3000 {
		o := Options{A: random() * 90, B: random() * 90, C: random() * 90, D: random() * 90,
			TX: random()*60 + 30, TY: random()*60 + 30, R: 256, G: 256, Blue: 256, Alpha: 256}
		if k%7 == 0 {
			o.B, o.C = 0, 0 // axis-aligned, as when level
		}
		var got, want View
		for i, v := range []*View{&got, &want} {
			*v = View{Buffer: &Buffer{Pixels: make([]uint16, 64*64), Mask: make([]byte, 64*64/8)}, Bounds: image.Rect(0, 0, 64, 64), Stride: 64}
			src := solid
			if i == 1 {
				src = masked
			}
			if !affine(*v, src, o) {
				t.Skip("transform not handled")
			}
		}
		for i := range got.Buffer.Pixels {
			if got.Buffer.Pixels[i] != want.Buffer.Pixels[i] || got.Buffer.Visible(i) != want.Buffer.Visible(i) {
				t.Fatalf("case %d %+v: pixel %d,%d is %x, want %x", k, o, i%64, i/64, got.Buffer.Pixels[i], want.Buffer.Pixels[i])
			}
		}
	}
}
