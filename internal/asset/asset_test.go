package asset

import (
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

func TestParseRejectsInvalidData(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.Set(0, 0, color.White)
	good, err := Encode(img)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < len(good); n++ {
		if _, _, _, _, err := Parse(good[:n]); err == nil {
			t.Fatalf("accepted %d bytes", n)
		}
	}
	for _, offset := range []int{4, 6, 8, 10} {
		data := append([]byte(nil), good...)
		binary.LittleEndian.PutUint16(data[offset:], 65535)
		if _, _, _, _, err := Parse(data); err == nil {
			t.Fatalf("accepted corrupt header at %d", offset)
		}
	}
}
func FuzzParse(f *testing.F) {
	f.Add([]byte("IKA1\x01\x00\x01\x00\x00\x00\x00\x00\xf8\x00"))
	f.Fuzz(func(t *testing.T, data []byte) {
		w, h, pixels, mask, err := Parse(data)
		if err != nil {
			return
		}
		if len(pixels) != 2*w*h || len(mask) > 0 && len(mask) != (w*h+7)/8 {
			t.Fatal("inconsistent asset")
		}
	})
}
