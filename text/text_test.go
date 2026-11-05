package text

import (
	"github.com/sat0ken/ikaten"
	"image/color"
	"testing"
	"tinygo.org/x/tinyfont/proggy"
)

func TestTextPixelsAndAllocations(t *testing.T) {
	dst := ikaten.NewImage(100, 30)
	var white color.Color = color.White
	Draw(dst, "Hello\nTinyGo", &proggy.TinySZ8pt7b, 0, 10, white)
	pixels := make([]byte, 100*30*4)
	dst.ReadPixels(pixels)
	visible := 0
	for p := 3; p < len(pixels); p += 4 {
		if pixels[p] > 0 {
			visible++
		}
	}
	if visible < 30 {
		t.Fatal("text did not draw")
	}
	if allocs := testing.AllocsPerRun(100, func() { Draw(dst, "Hello", &proggy.TinySZ8pt7b, 0, 10, white) }); allocs != 0 {
		t.Fatal("text allocated", allocs)
	}
}
