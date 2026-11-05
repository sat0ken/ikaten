package ikaten

import (
	"github.com/sat0ken/ikaten/internal/asset"
	"image"
	"image/color"
	"testing"
)

var red = color.RGBA{R: 255, A: 255}
var blue = color.RGBA{B: 255, A: 255}

func TestImageSubimageAndPixels(t *testing.T) {
	root := NewImage(7, 5)
	root.Fill(blue)
	sub := root.SubImage(image.Rect(2, 1, 6, 4)).(*Image)
	sub.Clear()
	sub.Set(3, 2, red)
	if got := root.RGBAAt(3, 2); got != red {
		t.Fatal(got)
	}
	if root.RGBAAt(2, 1).A != 0 || root.RGBAAt(1, 1) != blue {
		t.Fatal("mask or stride escaped the subimage")
	}
	sub.Set(-1, 0, red)
	if sub.RGBAAt(0, 0).A != 0 {
		t.Fatal("At outside bounds")
	}
	nested := sub.SubImage(image.Rect(3, 2, 5, 3)).(*Image)
	pixels := []byte{255, 0, 0, 255, 0, 255, 0, 255}
	nested.WritePixels(pixels)
	actual := make([]byte, 8)
	nested.ReadPixels(actual)
	for n := range actual {
		if actual[n] != pixels[n] {
			t.Fatalf("pixel %d: %d != %d", n, actual[n], pixels[n])
		}
	}
	dst := NewImage(4, 2)
	dst.DrawImage(nested, nil)
	if dst.RGBAAt(0, 0) != red || dst.RGBAAt(1, 0).G != 255 {
		t.Fatal("source subimage origin was not normalized")
	}
	empty := root.SubImage(image.Rect(50, 50, 51, 51)).(*Image)
	dst.DrawImage(empty, nil)
	empty.Clear()
}
func TestDrawImageClippingRotationAndBlend(t *testing.T) {
	src := NewImage(2, 2)
	src.Set(0, 0, red)
	src.Set(1, 1, blue)
	dst := NewImage(4, 4)
	dst.Fill(blue)
	var op DrawImageOptions
	op.GeoM.Translate(-1, -1)
	dst.DrawImage(src, &op)
	if dst.RGBAAt(0, 0) != blue {
		t.Fatal("clipped corner")
	}
	op = DrawImageOptions{Blend: BlendCopy}
	dst.DrawImage(src, &op)
	if dst.RGBAAt(1, 0).A != 0 || dst.RGBAAt(0, 0) != red {
		t.Fatal("BlendCopy did not copy mask")
	}
	dst.Fill(blue)
	op = DrawImageOptions{}
	op.ColorScale.ScaleAlpha(.5)
	dst.DrawImage(src, &op)
	got := dst.RGBAAt(0, 0)
	if got.R < 120 || got.R > 132 || got.B < 120 || got.B > 132 {
		t.Fatal("source over", got)
	}
	dst.Fill(blue)
	op = DrawImageOptions{Blend: BlendLighter}
	dst.DrawImage(src, &op)
	if dst.RGBAAt(0, 0) != (color.RGBA{R: 255, B: 255, A: 255}) {
		t.Fatal("lighter", dst.RGBAAt(0, 0))
	}
	dst.Clear()
	op = DrawImageOptions{}
	op.GeoM.Rotate(1.5707963267948966)
	op.GeoM.Translate(2, 0)
	dst.DrawImage(src, &op)
	if dst.RGBAAt(1, 0) != red || dst.RGBAAt(0, 1) != blue {
		t.Fatal("rotation")
	}
}
func TestAssetZeroCopyAndBinaryAlpha(t *testing.T) {
	original := image.NewNRGBA(image.Rect(3, 4, 5, 5))
	original.SetNRGBA(3, 4, color.NRGBA{R: 255, A: 128})
	data, err := asset.Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	img, err := NewImageFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	immutable, err := NewImageFromString(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if immutable.RGBAAt(0, 0) != img.RGBAAt(0, 0) {
		t.Fatal("string asset disagrees")
	}
	if img.RGBAAt(0, 0) != red || img.RGBAAt(1, 0).A != 0 {
		t.Fatal("asset alpha")
	}
	// Changing backing bytes is forbidden for users, but proves this is a view.
	data[12], data[13] = 0, 31
	if img.RGBAAt(0, 0) != blue {
		t.Fatal("asset was copied")
	}
	copy := NewImage(2, 1)
	copy.DrawImage(img, nil)
	if copy.RGBAAt(0, 0) != blue {
		t.Fatal("asset draw")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("read-only image accepted mutation")
		}
	}()
	img.Clear()
}
func TestRendererAllocations(t *testing.T) {
	src, dst := NewImage(16, 16), NewImage(120, 120)
	src.Fill(red)
	masked := NewImage(16, 16)
	masked.Set(3, 3, red)
	sprite := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	sprite.SetNRGBA(1, 1, color.NRGBA{R: 255, A: 255})
	data, err := asset.Encode(sprite)
	if err != nil {
		t.Fatal(err)
	}
	flash, err := NewImageFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	var scaled DrawImageOptions
	scaled.GeoM.Scale(2, 2)
	row := make([]byte, 240)
	var op DrawImageOptions
	op.GeoM.Rotate(.3)
	op.GeoM.Translate(40, 40)
	var fill color.Color = blue
	if n := testing.AllocsPerRun(100, func() {
		dst.Clear()
		dst.Fill(fill)
		dst.DrawImage(src, nil)
		dst.DrawImage(src, &op)
		dst.DrawImage(masked, nil)
		dst.DrawImage(flash, nil)
		dst.DrawImage(masked, &scaled)
		dst.ReadRGB565Row(row, 0)
	}); n != 0 {
		t.Fatalf("render hot path allocated %g times", n)
	}
}
func TestRGB565RowsMatchFlattenedRows(t *testing.T) {
	root := NewImage(9, 7)
	root.Fill(blue)
	sub := root.SubImage(image.Rect(1, 1, 8, 6)).(*Image)
	sub.Clear()
	sub.Set(2, 2, red)
	sub.Set(3, 2, color.NRGBA{R: 255, A: 10})
	sprite := NewImage(3, 3)
	sprite.Fill(red)
	sprite.Set(1, 1, color.RGBA{})
	root.DrawImage(sprite, nil)
	var op DrawImageOptions
	op.GeoM.Translate(4, 3)
	op.Blend = BlendCopy
	root.DrawImage(sprite, &op)
	part := root.SubImage(image.Rect(6, 0, 8, 2)).(*Image)
	part.WritePixels([]byte{255, 0, 0, 255, 200, 200, 200, 1, 0, 255, 0, 255, 0, 0, 0, 0})
	for _, img := range []*Image{root, sub} {
		b := img.Bounds()
		pixels, stride, ok := img.RGB565Rows()
		if !ok {
			t.Fatal("mutable image has no direct rows")
		}
		flat := make([]byte, 2*b.Dx())
		for y := 0; y < b.Dy(); y++ {
			img.ReadRGB565Row(flat, b.Min.Y+y)
			for x := 0; x < b.Dx(); x++ {
				v := pixels[y*stride+x]
				if flat[2*x] != byte(v) || flat[2*x+1] != byte(v>>8) {
					t.Fatalf("pixel %d,%d: direct %04x, flattened %02x%02x", x, y, v, flat[2*x], flat[2*x+1])
				}
			}
		}
	}
	original := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	data, err := asset.Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	flash, err := NewImageFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, ok := flash.RGB565Rows(); ok {
		t.Fatal("read-only asset exposed direct rows")
	}
}
func BenchmarkDrawImage16(b *testing.B) {
	src, dst := NewImage(16, 16), NewImage(240, 240)
	src.Fill(red)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dst.DrawImage(src, nil)
	}
}
func FuzzDrawImageClipping(f *testing.F) {
	f.Add(float32(1), float32(0), float32(0), float32(1), float32(-2), float32(3))
	f.Add(float32(1e30), float32(1e-20), float32(-1), float32(0), float32(0), float32(0))
	f.Fuzz(func(t *testing.T, a, b, c, d, x, y float32) {
		src, dst := NewImage(3, 2), NewImage(8, 8)
		var op DrawImageOptions
		op.GeoM.SetElement(0, 0, float64(a))
		op.GeoM.SetElement(0, 1, float64(b))
		op.GeoM.SetElement(1, 0, float64(c))
		op.GeoM.SetElement(1, 1, float64(d))
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(src, &op)
	})
}
