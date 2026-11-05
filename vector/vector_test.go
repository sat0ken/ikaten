package vector

import (
	"github.com/sat0ken/ikaten"
	"image"
	"image/color"
	"testing"
)

func TestShapesClipToDestinationBounds(t *testing.T) {
	root := ikaten.NewImage(20, 20)
	sub := root.SubImage(image.Rect(3, 4, 15, 16)).(*ikaten.Image)
	DrawFilledRect(sub, -100, -100, 300, 300, color.White, false)
	if root.RGBAAt(3, 4).A != 255 || root.RGBAAt(2, 4).A != 0 || root.RGBAAt(15, 4).A != 0 {
		t.Fatal("rectangle escaped subimage")
	}
	root.Clear()
	DrawFilledCircle(root, 10, 10, 3, color.White, false)
	if root.RGBAAt(10, 10).A != 255 || root.RGBAAt(0, 0).A != 0 {
		t.Fatal("circle coverage")
	}
	root.Clear()
	StrokeLine(root, 1, .5, 19, .5, 1, color.White, false)
	if root.RGBAAt(5, 0).A != 255 || root.RGBAAt(5, 1).A != 0 {
		t.Fatal("line coverage")
	}
	var c color.Color = color.White
	if n := testing.AllocsPerRun(100, func() { DrawFilledCircle(root, 10, 10, 3, c, false); StrokeLine(root, 1, 1, 18, 18, 2, c, false) }); n != 0 {
		t.Fatal("shapes allocate", n)
	}
}

func TestDeprecatedNamesMatchFillFunctions(t *testing.T) {
	a, b := ikaten.NewImage(20, 20), ikaten.NewImage(20, 20)
	FillRect(a, 1.5, 2.25, 9, 7.5, color.White, false)
	FillCircle(a, 12, 11, 5.5, color.White, false)
	DrawFilledRect(b, 1.5, 2.25, 9, 7.5, color.White, false)
	DrawFilledCircle(b, 12, 11, 5.5, color.White, false)
	seen := false
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if a.RGBAAt(x, y) != b.RGBAAt(x, y) {
				t.Fatalf("pixel %d,%d differs", x, y)
			}
			seen = seen || a.RGBAAt(x, y).A != 0
		}
	}
	if !seen {
		t.Fatal("nothing was drawn")
	}
}
