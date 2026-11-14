//go:build tinygo && (esp32s3 || esp32 || rp2040)

package main

import "unsafe"

// Ways to copy and fill a row of pixels, timed against each other: the
// built-in copy, loops over 16-bit pixels, and loops two pixels at a time
// over 32-bit words, which need both rows 4-byte aligned.

//go:noinline
func copyBuiltin(dst, src []uint16) { copy(dst, src) }

//go:noinline
func copyLoop16(dst, src []uint16) {
	src = src[:len(dst)]
	for i := range dst {
		dst[i] = src[i]
	}
}

//go:noinline
func copyLoop32(dst, src []uint16) {
	n := len(dst) / 2
	d := unsafe.Slice((*uint32)(unsafe.Pointer(&dst[0])), n)
	s := unsafe.Slice((*uint32)(unsafe.Pointer(&src[0])), n)
	for i := range d {
		d[i] = s[i]
	}
}

//go:noinline
func copyUnrolled32(dst, src []uint16) {
	n := len(dst) / 2
	d := unsafe.Slice((*uint32)(unsafe.Pointer(&dst[0])), n)
	s := unsafe.Slice((*uint32)(unsafe.Pointer(&src[0])), n)
	i := 0
	for ; i+4 <= n; i += 4 {
		a, b, c, e := s[i], s[i+1], s[i+2], s[i+3]
		d[i], d[i+1], d[i+2], d[i+3] = a, b, c, e
	}
	for ; i < n; i++ {
		d[i] = s[i]
	}
}

//go:noinline
func fillLoop16(dst []uint16, v uint16) {
	for i := range dst {
		dst[i] = v
	}
}

//go:noinline
func fillLoop32(dst []uint16, v uint16) {
	n := len(dst) / 2
	d := unsafe.Slice((*uint32)(unsafe.Pointer(&dst[0])), n)
	w := uint32(v) | uint32(v)<<16
	for i := range d {
		d[i] = w
	}
}

//go:noinline
func fillUnrolled32(dst []uint16, v uint16) {
	n := len(dst) / 2
	d := unsafe.Slice((*uint32)(unsafe.Pointer(&dst[0])), n)
	w := uint32(v) | uint32(v)<<16
	i := 0
	for ; i+4 <= n; i += 4 {
		d[i], d[i+1], d[i+2], d[i+3] = w, w, w, w
	}
	for ; i < n; i++ {
		d[i] = w
	}
}

// loops times each way on rows of 240 and of 16 pixels.
func loops() {
	a := make([]uint16, 512)
	b := make([]uint16, 512)
	for i := range b {
		b[i] = uint16(i * 40503)
	}
	for _, n := range []int{240, 16} {
		suffix := "_240"
		if n == 16 {
			suffix = "_16"
		}
		op("copy_builtin_aligned"+suffix, 2000, func() { copyBuiltin(a[:n], b[:n]) })
		op("copy_builtin_offset1"+suffix, 2000, func() { copyBuiltin(a[1:1+n], b[3:3+n]) })
		op("copy_builtin_odd"+suffix, 2000, func() { copyBuiltin(a[1:1+n], b[:n]) })
		op("copy_loop16"+suffix, 2000, func() { copyLoop16(a[1:1+n], b[:n]) })
		op("copy_loop32"+suffix, 2000, func() { copyLoop32(a[:n], b[:n]) })
		op("copy_unrolled32"+suffix, 2000, func() { copyUnrolled32(a[:n], b[:n]) })
		op("fill_loop16"+suffix, 2000, func() { fillLoop16(a[1:1+n], 0x1234) })
		op("fill_loop32"+suffix, 2000, func() { fillLoop32(a[:n], 0x1234) })
		op("fill_unrolled32"+suffix, 2000, func() { fillUnrolled32(a[:n], 0x1234) })
		op("clear_builtin"+suffix, 2000, func() { clear(a[1 : 1+n]) })
	}
}

// floatSum and intSum do 1000 multiplies and adds, and floatToInt 1000
// conversions, to show what float32 arithmetic costs next to integers.
//
//go:noinline
func floatSum(a, b float32) float32 {
	s := float32(0)
	for range 1000 {
		s = s*a + b
	}
	return s
}

//go:noinline
func intSum(a, b int32) int32 {
	s := int32(0)
	for range 1000 {
		s = s*a + b
	}
	return s
}

//go:noinline
func floatToInt(a float32) int {
	s := 0
	for i := range 1000 {
		s += int(a * float32(i))
	}
	return s
}

var sink float32
var sinkInt int

func arithmetic() {
	op("float32_mul_add_x1000", 200, func() { sink += floatSum(0.999, 0.5) })
	op("int32_mul_add_x1000", 200, func() { sinkInt += int(intSum(3, 1)) })
	op("float32_to_int_x1000", 200, func() { sinkInt += floatToInt(0.37) })
}
