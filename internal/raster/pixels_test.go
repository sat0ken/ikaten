package raster

import (
	"image"
	"math/rand"
	"reflect"
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
	// The same image in memory, and in flash form at even and odd addresses.
	raw := make([]byte, 2*w*h+1)
	for i, p := range pixels {
		raw[1+2*i], raw[2+2*i] = byte(p), byte(p>>8)
	}
	even := append([]byte(nil), raw[1:]...)
	for _, src := range []View{
		{Buffer: &Buffer{Pixels: pixels, Mask: mask}, Bounds: image.Rect(0, 0, w, h), Stride: w},
		{Buffer: &Buffer{Data: even, Mask: mask}, Bounds: image.Rect(0, 0, w, h), Stride: w},
		{Buffer: &Buffer{Data: raw[1:], Mask: mask}, Bounds: image.Rect(0, 0, w, h), Stride: w},
	} {
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
							t.Fatalf("data %v opaque %v blend %d tx %d: pixel %d is %x, want %x", src.Buffer.Data != nil, opaqueDst, blend, tx, i, got.Buffer.Pixels[i], want.Buffer.Pixels[i])
						}
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

// A one-pixel-wide image stretched or shrunk down, a whole number of pixels
// wide, as a raycaster's walls, takes the column path. It may differ from
// the general path only where a row's center sits on a texel's edge.
func TestColumnNearlyMatchesGeneral(t *testing.T) {
	rng := uint32(99)
	random := func() float32 {
		rng ^= rng << 13
		rng ^= rng >> 17
		rng ^= rng << 5
		return float32(rng%100000) / 100000
	}
	const w, h = 5, 32
	pixels, mask := make([]uint16, w*h), make([]byte, (w*h+7)/8)
	for i := range pixels {
		pixels[i] = uint16(i*7919 + 3)
		if i%3 != 0 {
			mask[i>>3] |= 1 << uint(i&7)
		}
	}
	differing, total := 0, 0
	for k := range 3000 {
		// One column of a wider image, as SubImage gives.
		col := 1 + k%3
		buf := &Buffer{Pixels: pixels, Mask: mask}
		if k%2 == 0 {
			buf = &Buffer{Pixels: pixels, Opaque: true}
		}
		src := View{Buffer: buf, Bounds: image.Rect(col, 0, col+1, h), Stride: w, Offset: col}
		o := Options{A: float32(1 + k%3), D: 0.1 + random()*12, TX: float32(int(random()*80) - 10), TY: random()*120 - 40,
			R: 256, G: 256, Blue: 256, Alpha: 256}
		if k%5 == 0 {
			o.R, o.Blue = 180, 90 // shaded
		}
		if k%7 == 0 {
			o.Blend = Copy
		}
		var got, want View
		for i, v := range []*View{&got, &want} {
			dp := make([]uint16, 64*96)
			dm := make([]byte, 64*96/8)
			for j := range dp {
				dp[j] = 0x1111
				if j%2 == 0 {
					dm[j>>3] |= 1 << uint(j&7)
				}
			}
			*v = View{Buffer: &Buffer{Pixels: dp, Mask: dm}, Bounds: image.Rect(0, 0, 64, 96), Stride: 64}
			draw(*v, src, o, i == 1)
		}
		for i := range got.Buffer.Pixels {
			total++
			if got.Buffer.Pixels[i] != want.Buffer.Pixels[i] || got.Buffer.Visible(i) != want.Buffer.Visible(i) {
				differing++
			}
		}
	}
	if differing*10000 > total {
		t.Fatalf("%d of %d pixels differ", differing, total)
	}
	t.Logf("%d of %d pixels differ", differing, total)
}

// Opaque images rotated and scaled with steps under a pixel, from RAM and
// from flash data aligned to two bytes or not, take affineOpaque. It may
// differ from the general path only on texel edges.
func TestAffineOpaqueNearlyMatchesGeneral(t *testing.T) {
	const w, h = 40, 30
	pixels := make([]uint16, w*h)
	raw := make([]byte, 2*w*h+1)
	for i := range pixels {
		pixels[i] = uint16(i*7919 + 3)
	}
	sources := map[string]View{"ram": {Buffer: &Buffer{Pixels: pixels}, Bounds: image.Rect(0, 0, w, h), Stride: w}}
	for _, shift := range []int{0, 1} {
		data := raw[shift : shift+2*w*h]
		for i, p := range pixels {
			data[2*i], data[2*i+1] = byte(p), byte(p>>8)
		}
		sources[[]string{"flash-aligned", "flash-odd"}[shift]] = View{Buffer: &Buffer{Data: data}, Bounds: image.Rect(0, 0, w, h), Stride: w}
	}
	rng := uint32(7)
	random := func() float32 {
		rng ^= rng << 13
		rng ^= rng >> 17
		rng ^= rng << 5
		return float32(rng%20000)/10000 - 1
	}
	for name, src := range sources {
		differing, total := 0, 0
		for range 1000 {
			// Rows of a floor: magnified, turned any way.
			o := Options{A: 1 + random()*2, B: random() * 2, C: random() * 2, D: 1 + random()*2,
				TX: random() * 30, TY: random() * 30, R: 256, G: 256, Blue: 256, Alpha: 256}
			var got, want View
			for i, v := range []*View{&got, &want} {
				*v = View{Buffer: &Buffer{Pixels: make([]uint16, 64*64), Mask: make([]byte, 64*64/8)}, Bounds: image.Rect(0, 0, 64, 64), Stride: 64}
				draw(*v, src, o, i == 1)
			}
			for i := range got.Buffer.Pixels {
				total++
				if got.Buffer.Pixels[i] != want.Buffer.Pixels[i] || got.Buffer.Visible(i) != want.Buffer.Visible(i) {
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

// Rows of flash images are read into RAM in chunks before drawing; wide rows,
// every alignment of the data and the mask, and magnified draws must match
// the general path.
func TestFlashRows(t *testing.T) {
	const w, h = 700, 3
	rng := rand.New(rand.NewSource(5))
	for shift := range 4 {
		raw := make([]byte, 2*w*h+shift)
		for i := range raw {
			raw[i] = byte(rng.Uint32())
		}
		rawMask := make([]byte, (w*h+7)/8+shift)
		for i := range rawMask {
			rawMask[i] = [...]byte{0xff, 0, 0xff, 0x5a, byte(rng.Uint32())}[i%5]
		}
		for _, masked := range []bool{false, true} {
			b := &Buffer{Data: raw[shift:]}
			if masked {
				b.Mask = rawMask[shift:]
			}
			src := View{Buffer: b, Bounds: image.Rect(0, 0, w, h), Stride: w}
			for _, sx := range []int{1, 2, 3} {
				for _, c := range [][2]int{{-5, 500}, {0, 500}, {3, 500}, {-5, 800}, {3, 800}} {
					tx := c[0]
					var got, want View
					for k, v := range []*View{&got, &want} {
						*v = newView(c[1], 4, false)
						draw(*v, src, Options{A: float32(sx), D: 1, TX: float32(tx), TY: 1, R: 256, G: 256, Blue: 256, Alpha: 256}, k == 1)
					}
					if !reflect.DeepEqual(got.Buffer, want.Buffer) {
						t.Fatalf("shift %d masked %v scale %d tx %d", shift, masked, sx, tx)
					}
				}
			}
		}
	}
}
