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
