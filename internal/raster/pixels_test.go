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
