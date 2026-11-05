package present

import (
	"errors"
	"image"
	"reflect"
	"testing"
)

type source struct {
	bounds image.Rectangle
	pixels []byte
}

func (s *source) Bounds() image.Rectangle { return s.bounds }
func (s *source) ReadRGB565Row(dst []byte, y int) {
	start := (y - s.bounds.Min.Y) * s.bounds.Dx() * 2
	copy(dst, s.pixels[start:start+len(dst)])
}

type writer struct {
	width, height, calls int
	pixels               []byte
	fail                 bool
}

func (w *writer) WriteRect(x, y, width, height int, data []byte) error {
	if w.fail {
		return errors.New("transfer failed")
	}
	if x < 0 || y < 0 || x+width > w.width || y+height > w.height || len(data) != width*height*2 {
		panic("invalid rectangle")
	}
	for n := 0; n < height; n++ {
		copy(w.pixels[((y+n)*w.width+x)*2:((y+n)*w.width+x+width)*2], data[n*width*2:(n+1)*width*2])
	}
	w.calls++
	return nil
}
func TestScaleDirtyRowsAndRetry(t *testing.T) {
	src := &source{image.Rect(5, 6, 7, 8), []byte{0xf8, 0, 0, 31, 7, 224, 255, 255}}
	dst := &writer{width: 6, height: 5, pixels: make([]byte, 60)}
	p := Presenter{Width: 6, Height: 5}
	if err := p.Present(src, dst); err != nil {
		t.Fatal(err)
	}
	want := make([]byte, 60)
	for y := 0; y < 4; y++ {
		for x := 1; x < 5; x++ {
			copy(want[(y*6+x)*2:], src.pixels[((y/2)*2+(x-1)/2)*2:((y/2)*2+(x-1)/2)*2+2])
		}
	}
	if !reflect.DeepEqual(dst.pixels, want) {
		t.Fatalf("scale/letterbox\n%x\n%x", dst.pixels, want)
	}
	dst.calls = 0
	if err := p.Present(src, dst); err != nil {
		t.Fatal(err)
	}
	if dst.calls != 0 || p.Stats.ChangedRows != 0 {
		t.Fatal("unchanged frame transferred")
	}
	src.pixels[4] = 0
	dst.fail = true
	if err := p.Present(src, dst); err == nil {
		t.Fatal("write error swallowed")
	}
	dst.fail = false
	if err := p.Present(src, dst); err != nil {
		t.Fatal(err)
	}
	if dst.calls != 1 || p.Stats.ChangedRows != 1 {
		t.Fatal("failed row not retried", dst.calls, p.Stats)
	}
	if allocs := testing.AllocsPerRun(100, func() {
		if err := p.Present(src, dst); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatal("present allocates", allocs)
	}
	// Resizing must clear old letterboxing and invalidate every row.
	src.bounds = image.Rect(0, 0, 3, 1)
	src.pixels = []byte{0, 31, 0, 31, 0, 31}
	if err := p.Present(src, dst); err != nil {
		t.Fatal(err)
	}
	if dst.pixels[2] != 0 || dst.pixels[3] != 0 {
		t.Fatal("old viewport survived layout resize")
	}
}

// directSource stores wire-order pixels the way ikaten.Image does.
type directSource struct {
	bounds image.Rectangle
	pixels []uint16
	stride int
}

func (s *directSource) Bounds() image.Rectangle { return s.bounds }
func (s *directSource) ReadRGB565Row(dst []byte, y int) {
	panic("direct source must not be copied")
}
func (s *directSource) RGB565Rows() ([]uint16, int, bool) { return s.pixels, s.stride, true }

func wire(v uint16) (byte, byte) { return byte(v), byte(v >> 8) }

func newDirect(w, h, stride int) *directSource {
	s := &directSource{bounds: image.Rect(0, 0, w, h), pixels: make([]uint16, stride*h), stride: stride}
	for n := range s.pixels {
		s.pixels[n] = uint16(n*257 + 1)
	}
	return s
}

// checkFrame compares the panel with the source scaled by nearest neighbour.
func checkFrame(t *testing.T, dst *writer, s *directSource, scale, ox, oy int) {
	t.Helper()
	b := s.bounds
	for y := 0; y < b.Dy()*scale; y++ {
		for x := 0; x < b.Dx()*scale; x++ {
			hi, lo := wire(s.pixels[(y/scale)*s.stride+x/scale])
			at := ((y+oy)*dst.width + x + ox) * 2
			if dst.pixels[at] != hi || dst.pixels[at+1] != lo {
				t.Fatalf("pixel %d,%d: %02x%02x, want %02x%02x", x, y, dst.pixels[at], dst.pixels[at+1], hi, lo)
			}
		}
	}
}
func TestDirectBandsMergeRowsAndRetry(t *testing.T) {
	s := newDirect(6, 8, 6)
	dst := &writer{width: 6, height: 8, pixels: make([]byte, 96)}
	p := Presenter{Width: 6, Height: 8}
	if err := p.Present(s, dst); err != nil {
		t.Fatal(err)
	}
	checkFrame(t, dst, s, 1, 0, 0)
	dst.calls = 0
	if err := p.Present(s, dst); err != nil || dst.calls != 0 {
		t.Fatal("unchanged frame transferred", err, dst.calls)
	}
	// Rows 1-3 form one band and row 6 another.
	for _, y := range []int{1, 2, 3, 6} {
		s.pixels[y*6+2] ^= 0xffff
	}
	if err := p.Present(s, dst); err != nil {
		t.Fatal(err)
	}
	if dst.calls != 2 || p.Stats.ChangedRows != 4 || p.Stats.Bytes != 48 {
		t.Fatal("bands", dst.calls, p.Stats)
	}
	checkFrame(t, dst, s, 1, 0, 0)
	s.pixels[0] ^= 0xffff
	s.pixels[7*6] ^= 0xffff
	dst.fail = true
	if err := p.Present(s, dst); err == nil {
		t.Fatal("write error swallowed")
	}
	dst.fail = false
	dst.calls = 0
	if err := p.Present(s, dst); err != nil {
		t.Fatal(err)
	}
	if dst.calls != 2 || p.Stats.ChangedRows != 2 {
		t.Fatal("failed bands not retried", dst.calls, p.Stats)
	}
	checkFrame(t, dst, s, 1, 0, 0)
	if allocs := testing.AllocsPerRun(100, func() {
		s.pixels[3*6] ^= 0xffff
		if err := p.Present(s, dst); err != nil {
			panic(err)
		}
	}); allocs != 0 {
		t.Fatal("present allocates", allocs)
	}
}
func TestDirectRowsWithStrideAndScale(t *testing.T) {
	for _, c := range []struct{ w, h, stride, width, height, scale, ox, oy int }{
		{4, 3, 7, 4, 3, 1, 0, 0}, // subimage: rows are not contiguous
		{3, 2, 3, 7, 5, 2, 0, 0}, // 2x with letterboxing
		{3, 2, 5, 9, 6, 3, 0, 0}, // 3x splits into 2+1 rows
		{2, 2, 2, 5, 5, 2, 0, 0},
	} {
		s := newDirect(c.w, c.h, c.stride)
		dst := &writer{width: c.width, height: c.height, pixels: make([]byte, c.width*c.height*2)}
		p := Presenter{Width: c.width, Height: c.height}
		if err := p.Present(s, dst); err != nil {
			t.Fatal(err)
		}
		checkFrame(t, dst, s, c.scale, (c.width-c.w*c.scale)/2, (c.height-c.h*c.scale)/2)
		dst.calls = 0
		s.pixels[c.stride] ^= 0xffff
		if err := p.Present(s, dst); err != nil {
			t.Fatal(err)
		}
		if p.Stats.ChangedRows != 1 {
			t.Fatal("changed rows", c, p.Stats)
		}
		checkFrame(t, dst, s, c.scale, (c.width-c.w*c.scale)/2, (c.height-c.h*c.scale)/2)
		if allocs := testing.AllocsPerRun(100, func() {
			s.pixels[0] ^= 0xffff
			if err := p.Present(s, dst); err != nil {
				panic(err)
			}
		}); allocs != 0 {
			t.Fatal("present allocates", c, allocs)
		}
	}
}
