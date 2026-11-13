// Package raster implements RGB565 rendering without hardware dependencies.
package raster

import "image"

// Buffer owns pixels. Pixels contains byte-swapped RGB565 (little-endian targets
// have ST7789 byte order in memory). Data is the zero-copy, big-endian asset form.
// A nil Mask denotes an entirely opaque image; other masks are LSB first.
// Image keeps transparent Pixels at 0, which backends rely on when sending rows.
type Buffer struct {
	Pixels     []uint16
	Data, Mask []byte
	Opaque     bool
}
type View struct {
	Buffer         *Buffer
	Bounds         image.Rectangle
	Stride, Offset int
}

func Pack(r, g, b uint8) uint16 {
	v := uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
	return v<<8 | v>>8
}
func Unpack(v uint16) (r, g, b uint8) {
	v = v<<8 | v>>8
	r5, g6, b5 := uint8(v>>11), uint8(v>>5)&63, uint8(v)&31
	return r5<<3 | r5>>2, g6<<2 | g6>>4, b5<<3 | b5>>2
}
func (b *Buffer) Pixel(i int) uint16 {
	if b.Data != nil {
		return uint16(b.Data[2*i]) | uint16(b.Data[2*i+1])<<8
	}
	return b.Pixels[i]
}
func (b *Buffer) Visible(i int) bool {
	return b.Opaque || b.Mask == nil || b.Mask[i>>3]&(1<<uint(i&7)) != 0
}
func (b *Buffer) Set(i int, v uint16, visible bool) {
	b.Pixels[i] = v
	if !visible {
		b.Opaque = false
	}
	if b.Mask != nil {
		if visible {
			b.Mask[i>>3] |= 1 << uint(i&7)
		} else {
			b.Mask[i>>3] &^= 1 << uint(i&7)
		}
	}
}

// SetMaskRange changes a span without touching adjacent subimage pixels.
func (b *Buffer) SetMaskRange(start, count int, visible bool) {
	if !visible {
		b.Opaque = false
	}
	mask := b.Mask
	if mask == nil || count == 0 {
		return
	}
	end := start + count
	for start < end && start&7 != 0 {
		if visible {
			mask[start>>3] |= 1 << uint(start&7)
		} else {
			mask[start>>3] &^= 1 << uint(start&7)
		}
		start++
	}
	value := byte(0)
	if visible {
		value = 255
	}
	for start+8 <= end {
		mask[start>>3] = value
		start += 8
	}
	for start < end {
		if visible {
			mask[start>>3] |= 1 << uint(start&7)
		} else {
			mask[start>>3] &^= 1 << uint(start&7)
		}
		start++
	}
}
func (v View) Index(x, y int) int { return v.Offset + (y-v.Bounds.Min.Y)*v.Stride + x - v.Bounds.Min.X }
