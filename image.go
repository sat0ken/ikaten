package ikaten

import (
	"github.com/sat0ken/ikaten/internal/asset"
	"github.com/sat0ken/ikaten/internal/raster"
	"image"
	"image/color"
	"math"
	"unsafe"
)

type Filter int

const FilterNearest Filter = 0

type Blend = raster.Blend

const (
	BlendSourceOver = raster.SourceOver
	BlendCopy       = raster.Copy
	BlendLighter    = raster.Lighter
)

type DrawImageOptions struct {
	GeoM       GeoM
	ColorScale ColorScale
	Blend      Blend
	Filter     Filter
}

// Image stores RGB565 with binary alpha. Subimages share pixels and masks.
// Image operations and the game loop must run on the same goroutine.
type Image struct {
	view     raster.View
	readOnly bool
}

func NewImage(width, height int) *Image {
	if width <= 0 || height <= 0 || width > 32767 || height > 32767 {
		panic("ikaten: invalid image dimensions")
	}
	n := width * height
	return &Image{view: raster.View{Buffer: &raster.Buffer{Pixels: make([]uint16, n), Mask: make([]byte, (n+7)/8)}, Bounds: image.Rect(0, 0, width, height), Stride: width}}
}

// NewImageFromBytes references an IKA1 asset without copying its pixels. Keep
// data unchanged for the image's lifetime. The result and its subimages are
// read-only; draw it onto a NewImage to obtain a mutable copy.
func NewImageFromBytes(data []byte) (*Image, error) {
	w, h, pixels, mask, err := asset.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Image{view: raster.View{Buffer: &raster.Buffer{Data: pixels, Mask: mask}, Bounds: image.Rect(0, 0, w, h), Stride: w}, readOnly: true}, nil
}

// NewImageFromString references an immutable embedded string without copying.
// Prefer this for flash-resident TinyGo assets: embedding into []byte can
// allocate writable static RAM before main. No code writes through this view.
func NewImageFromString(data string) (*Image, error) {
	return NewImageFromBytes(unsafe.Slice(unsafe.StringData(data), len(data)))
}
func NewImageFromImage(src image.Image) *Image {
	b := src.Bounds()
	dst := NewImage(b.Dx(), b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(x, y, src.At(x+b.Min.X, y+b.Min.Y))
		}
	}
	return dst
}
func (i *Image) Bounds() image.Rectangle { return i.view.Bounds }
func (i *Image) Size() (int, int)        { return i.view.Bounds.Dx(), i.view.Bounds.Dy() }
func (i *Image) ColorModel() color.Model { return color.RGBAModel }
func (i *Image) At(x, y int) color.Color { return i.RGBAAt(x, y) }
func (i *Image) RGBAAt(x, y int) color.RGBA {
	if !image.Pt(x, y).In(i.view.Bounds) {
		return color.RGBA{}
	}
	p := i.view.Index(x, y)
	if !i.view.Buffer.Visible(p) {
		return color.RGBA{}
	}
	r, g, b := raster.Unpack(i.view.Buffer.Pixel(p))
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
func (i *Image) mutable() {
	if i.readOnly {
		panic("ikaten: cannot modify a read-only asset")
	}
}
func quantize(clr color.Color) (uint16, bool) {
	r, g, b, a := clr.RGBA()
	if a < 32768 {
		return 0, false
	}
	// RGBA is premultiplied; stored RGB is straight after alpha thresholding.
	return raster.Pack(uint8(min(r*255/a, 255)), uint8(min(g*255/a, 255)), uint8(min(b*255/a, 255))), true
}
func (i *Image) Set(x, y int, clr color.Color) {
	i.mutable()
	if !image.Pt(x, y).In(i.view.Bounds) {
		return
	}
	v, visible := quantize(clr)
	i.view.Buffer.Set(i.view.Index(x, y), v, visible)
}

// SetRGBA is an allocation-free backend/helper equivalent of Set.
func (i *Image) SetRGBA(x, y int, c color.RGBA) {
	i.mutable()
	if !image.Pt(x, y).In(i.view.Bounds) {
		return
	}
	a := uint32(c.A)
	v := uint16(0)
	if a >= 128 {
		v = raster.Pack(uint8(min(uint32(c.R)*255/a, 255)), uint8(min(uint32(c.G)*255/a, 255)), uint8(min(uint32(c.B)*255/a, 255)))
	}
	i.view.Buffer.Set(i.view.Index(x, y), v, a >= 128)
}
func (i *Image) Fill(clr color.Color) {
	i.mutable()
	v, visible := quantize(clr)
	b := i.view.Bounds
	buf := i.view.Buffer
	if i.wholeBuffer() {
		raster.FillPixels(buf.Pixels, v)
		buf.SetMaskRange(0, len(buf.Pixels), visible)
		buf.Opaque = visible
		return
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		p := i.view.Index(b.Min.X, y)
		raster.FillPixels(buf.Pixels[p:p+b.Dx()], v)
		buf.SetMaskRange(p, b.Dx(), visible)
	}
}
func (i *Image) Clear() {
	i.mutable()
	if i.wholeBuffer() {
		clear(i.view.Buffer.Pixels)
		clear(i.view.Buffer.Mask)
		i.view.Buffer.Opaque = false
		return
	}
	b := i.view.Bounds
	for y := b.Min.Y; y < b.Max.Y; y++ {
		p := i.view.Index(b.Min.X, y)
		for x := 0; x < b.Dx(); x++ {
			i.view.Buffer.Set(p+x, 0, false)
		}
	}
}
func (i *Image) wholeBuffer() bool {
	return i.view.Offset == 0 && i.view.Stride == i.view.Bounds.Dx() && len(i.view.Buffer.Pixels) == i.view.Bounds.Dx()*i.view.Bounds.Dy()
}
func (i *Image) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(i.view.Bounds)
	v := i.view
	v.Bounds = r
	if !r.Empty() {
		v.Offset = i.view.Index(r.Min.X, r.Min.Y)
	}
	return &Image{view: v, readOnly: i.readOnly}
}
func (i *Image) WritePixels(pixels []byte) {
	i.mutable()
	b := i.view.Bounds
	if len(pixels) != 4*b.Dx()*b.Dy() {
		panic("ikaten: invalid WritePixels buffer length")
	}
	p := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			a := uint32(pixels[p+3])
			var v uint16
			if a >= 128 {
				v = raster.Pack(uint8(min(uint32(pixels[p])*255/a, 255)), uint8(min(uint32(pixels[p+1])*255/a, 255)), uint8(min(uint32(pixels[p+2])*255/a, 255)))
			}
			i.view.Buffer.Set(i.view.Index(x, y), v, a >= 128)
			p += 4
		}
	}
}
func (i *Image) ReadPixels(pixels []byte) {
	b := i.view.Bounds
	if len(pixels) != 4*b.Dx()*b.Dy() {
		panic("ikaten: invalid ReadPixels buffer length")
	}
	p := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := i.RGBAAt(x, y)
			pixels[p], pixels[p+1], pixels[p+2], pixels[p+3] = c.R, c.G, c.B, c.A
			p += 4
		}
	}
}

// ReadRGB565Row writes one row in ST7789 byte order, flattening transparency
// onto black. y uses Bounds coordinates. This is intended for hardware backends.
func (i *Image) ReadRGB565Row(dst []byte, y int) {
	b := i.view.Bounds
	if y < b.Min.Y || y >= b.Max.Y || len(dst) != 2*b.Dx() {
		panic("ikaten: invalid RGB565 row")
	}
	p := i.view.Index(b.Min.X, y)
	for x := 0; x < b.Dx(); x++ {
		v := uint16(0)
		if i.view.Buffer.Visible(p + x) {
			v = i.view.Buffer.Pixel(p + x)
		}
		dst[2*x], dst[2*x+1] = byte(v), byte(v>>8)
	}
}

// RGB565Rows exposes a mutable image's pixels for hardware backends. Transparent
// pixels hold 0, so the data equals ReadRGB565Row's flattening onto black.
func (i *Image) RGB565Rows() (pixels []uint16, stride int, ok bool) {
	if i.view.Buffer.Pixels == nil || i.view.Bounds.Empty() {
		return nil, 0, false
	}
	return i.view.Buffer.Pixels[i.view.Offset:], i.view.Stride, true
}

// fixed returns the scale in 1/256 units. The zero value, the identity, skips
// float arithmetic, which is software on RP2040.
func (c *ColorScale) fixed() (r, g, b, a int32) {
	if math.Float32bits(c.r)|math.Float32bits(c.g)|math.Float32bits(c.b)|math.Float32bits(c.a) == 0 {
		return 256, 256, 256, 256
	}
	return scaleFixed(c.R()), scaleFixed(c.G()), scaleFixed(c.B()), min(scaleFixed(c.A()), 256)
}

// scaleFixed converts a scale to 1/256 units, rounding half up. It reads the
// float's bits because float arithmetic is software on RP2040.
func scaleFixed(v float32) int32 {
	bits := math.Float32bits(v)
	if bits == 0 || bits > 0x7f800000 { // zero, NaN or negative
		return 0
	}
	if bits >= 0x43800000 { // 256 and above
		return 65536
	}
	// v*256 is the 24-bit mantissa shifted right by 142-exponent.
	shift := 142 - bits>>23
	if shift > 24 {
		return 0
	}
	mantissa := bits&0x7fffff | 0x800000
	return int32((mantissa + 1<<(shift-1)) >> shift)
}
func (i *Image) DrawImage(src *Image, options *DrawImageOptions) {
	i.mutable()
	if src == nil {
		panic("ikaten: nil source image")
	}
	if i.view.Buffer == src.view.Buffer {
		panic("ikaten: DrawImage source and destination share pixels")
	}
	var op DrawImageOptions
	if options != nil {
		op = *options
	}
	if op.Filter != FilterNearest || op.Blend > BlendLighter {
		panic("ikaten: unsupported filter or blend")
	}
	a, b, c, d, tx, ty := op.GeoM.elements()
	r, g, bl, al := op.ColorScale.fixed()
	raster.Draw(i.view, src.view, raster.Options{A: a, B: b, C: c, D: d, TX: tx, TY: ty, R: r, G: g, Blue: bl, Alpha: al, Blend: op.Blend})
}
