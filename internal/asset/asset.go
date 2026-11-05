// Package asset defines IKA1: a 12-byte header, big-endian RGB565 pixels, and
// an optional LSB-first visibility mask. Header integers are little-endian.
package asset

import (
	"encoding/binary"
	"errors"
	"github.com/sat0ken/ikaten/internal/raster"
	"image"
)

const HeaderSize = 12

func Parse(data []byte) (width, height int, pixels, mask []byte, err error) {
	invalid := errors.New("ikaten: invalid IKA1 asset")
	if len(data) < HeaderSize || string(data[:4]) != "IKA1" {
		err = invalid
		return
	}
	width, height = int(binary.LittleEndian.Uint16(data[4:6])), int(binary.LittleEndian.Uint16(data[6:8]))
	flags := binary.LittleEndian.Uint16(data[8:10])
	if width == 0 || height == 0 || width > 32767 || height > 32767 || flags > 1 || binary.LittleEndian.Uint16(data[10:12]) != 0 {
		err = invalid
		return
	}
	n := width * height
	size := HeaderSize + 2*n
	if flags == 1 {
		size += (n + 7) / 8
	}
	if len(data) != size {
		err = invalid
		return
	}
	pixels = data[HeaderSize : HeaderSize+2*n]
	if flags == 1 {
		mask = data[HeaderSize+2*n:]
	}
	return
}

// Encode thresholds alpha at 128 and rounds color down to RGB565.
func Encode(src image.Image) ([]byte, error) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || w > 32767 || h > 32767 {
		return nil, errors.New("ikaten: invalid asset dimensions")
	}
	n := w * h
	data := make([]byte, HeaderSize+2*n+(n+7)/8)
	copy(data, "IKA1")
	binary.LittleEndian.PutUint16(data[4:6], uint16(w))
	binary.LittleEndian.PutUint16(data[6:8], uint16(h))
	mask := data[HeaderSize+2*n:]
	opaque := true
	p := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			v := uint16(0)
			if a >= 32768 {
				v = raster.Pack(uint8(min(r*255/a, 255)), uint8(min(g*255/a, 255)), uint8(min(bl*255/a, 255)))
				mask[p>>3] |= 1 << uint(p&7)
			} else {
				opaque = false
			}
			data[HeaderSize+2*p], data[HeaderSize+2*p+1] = byte(v), byte(v>>8)
			p++
		}
	}
	if opaque {
		return data[:HeaderSize+2*n], nil
	}
	binary.LittleEndian.PutUint16(data[8:10], 1)
	return data, nil
}
