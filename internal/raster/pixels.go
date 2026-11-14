package raster

import "unsafe"

// The loops here keep their slices in locals and write two pixels at a time
// where they can. Reading a slice through a struct pointer inside a loop
// makes TinyGo load the pointer, check it and load the length again for
// every pixel, as the store might change them; and copy calls memmove, which
// on ESP32-S3 is in ROM and copies a byte at a time when the rows are not
// aligned. On the badge both cost several times what a plain loop does.

// FillPixels sets every pixel of p to v.
func FillPixels(p []uint16, v uint16) { fillPixels(p, v) }

// fillPixels sets every pixel of p to v.
func fillPixels(p []uint16, v uint16) {
	n := len(p)
	if n == 0 {
		return
	}
	i := 0
	if uintptr(unsafe.Pointer(&p[0]))&2 != 0 {
		p[0] = v
		i = 1
	}
	if m := (n - i) >> 1; m > 0 {
		w := unsafe.Slice((*uint32)(unsafe.Pointer(&p[i])), m)
		pair := uint32(v) | uint32(v)<<16 // the same pixel twice in either byte order
		for k := range w {
			w[k] = pair
		}
		i += m << 1
	}
	if i < n {
		p[n-1] = v
	}
}

// copyPixels copies src to dst, which must be as long and must not overlap.
func copyPixels(dst, src []uint16) {
	n := len(dst)
	if n == 0 {
		return
	}
	src = src[:n]
	i := 0
	d0, s0 := uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(&src[0]))
	if (d0^s0)&2 == 0 {
		// Both rows can be read and written a word at a time.
		if d0&2 != 0 {
			dst[0] = src[0]
			i = 1
		}
		if m := (n - i) >> 1; m > 0 {
			d := unsafe.Slice((*uint32)(unsafe.Pointer(&dst[i])), m)
			s := unsafe.Slice((*uint32)(unsafe.Pointer(&src[i])), m)
			for k := range d {
				d[k] = s[k]
			}
			i += m << 1
		}
	}
	for ; i < n; i++ {
		dst[i] = src[i]
	}
}

// On the badge every read from an image in flash is slow, about 0.1 µs, so
// rows are read into RAM a 32-bit word, two pixels or four mask bytes, at a
// time before they are drawn. flashChunk is the most pixels read at once.
const flashChunk = 256

var flashRow struct {
	pix  [flashChunk + 8]uint16
	mask [flashChunk/8 + 2]byte
	buf  Buffer
}

// loadFlash reads pixels p to p+n-1 of src, n at most flashChunk, with their
// mask bits, into flashRow. Pixel p is at the returned index of the returned
// buffer, and keeps its bit position within its mask byte.
func loadFlash(src *Buffer, p, n int) (*Buffer, int) {
	base := p & 7
	readPixels(flashRow.pix[base:base+n], src.Data, p)
	b := &flashRow.buf
	b.Pixels, b.Mask = flashRow.pix[:], nil
	if src.Mask != nil {
		first, last := p>>3, (p+n-1)>>3
		readBytes(flashRow.mask[:last-first+1], src.Mask[first:last+1])
		b.Mask = flashRow.mask[:]
	}
	return b, base
}

// readPixels reads len(out) pixels of data starting at pixel p.
func readPixels(out []uint16, data []byte, p int) {
	n := len(out)
	if n == 0 {
		return
	}
	in := data[2*p : 2*(p+n)]
	a := uintptr(unsafe.Pointer(&in[0]))
	if a&1 != 0 {
		for i := range out {
			out[i] = uint16(in[2*i]) | uint16(in[2*i+1])<<8
		}
		return
	}
	i := 0
	if a&2 != 0 {
		out[0] = *(*uint16)(unsafe.Pointer(&in[0]))
		i = 1
	}
	if m := (n - i) >> 1; m > 0 {
		w := unsafe.Slice((*uint32)(unsafe.Pointer(&in[2*i])), m)
		for k, v := range w {
			// Little-endian, like the 16-bit pixels.
			out[i+2*k], out[i+2*k+1] = uint16(v), uint16(v>>16)
		}
		i += m << 1
	}
	if i < n {
		out[i] = *(*uint16)(unsafe.Pointer(&in[2*i]))
	}
}

// readBytes reads len(out) bytes of in.
func readBytes(out, in []byte) {
	n := len(out)
	in = in[:n]
	i := 0
	for ; i < n && uintptr(unsafe.Pointer(&in[i]))&3 != 0; i++ {
		out[i] = in[i]
	}
	for ; i+4 <= n; i += 4 {
		v := *(*uint32)(unsafe.Pointer(&in[i]))
		out[i], out[i+1], out[i+2], out[i+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	for ; i < n; i++ {
		out[i] = in[i]
	}
}
