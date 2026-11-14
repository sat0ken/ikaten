package raster

import (
	"math"
	"unsafe"
)

type Blend uint8

const (
	SourceOver Blend = iota
	Copy
	Lighter
)

type Options struct {
	A, B, C, D, TX, TY float32
	R, G, Blue, Alpha  int32
	Blend              Blend
}

func Draw(dst, src View, o Options) { draw(dst, src, o, false) }

// draw's forceGeneral switch is used by reference-path tests.
func draw(dst, src View, o Options, forceGeneral bool) {
	if dst.Bounds.Empty() || src.Bounds.Empty() {
		return
	}
	vals := [6]float32{o.A, o.B, o.C, o.D, o.TX, o.TY}
	for _, v := range vals {
		if math.Float32bits(v)&0x7f800000 == 0x7f800000 { // NaN or ±Inf, without software float
			return
		}
	}
	if !forceGeneral && o.B == 0 && o.C == 0 && o.A >= 1 && o.D >= 1 && o.A <= 32767 && o.D <= 32767 &&
		o.A == float32(int(o.A)) && o.D == float32(int(o.D)) && o.TX >= -1e7 && o.TX <= 1e7 && o.TY >= -1e7 && o.TY <= 1e7 &&
		o.TX == float32(int(o.TX)) && o.TY == float32(int(o.TY)) {
		integer(dst, src, o)
		return
	}
	if !forceGeneral && o.A == -1 && o.B == 0 && o.C == 0 && o.D >= 1 && o.D <= 32767 && o.D == float32(int(o.D)) &&
		o.TX >= -1e7 && o.TX <= 1e7 && o.TY >= -1e7 && o.TY <= 1e7 && o.TX == float32(int(o.TX)) && o.TY == float32(int(o.TY)) &&
		o.R == 256 && o.G == 256 && o.Blue == 256 && o.Alpha == 256 && o.Blend != Lighter {
		mirrored(dst, src, o)
		return
	}
	if !forceGeneral && o.B == 0 && o.C == 0 && src.Bounds.Dx() == 1 && o.A >= 1 && o.A <= 64 && o.A == float32(int(o.A)) &&
		o.TX >= -1e7 && o.TX <= 1e7 && o.TX == float32(int(o.TX)) && o.D > 1e-3 && o.TY > -8192 && o.TY < 8192 &&
		float32(src.Bounds.Dy())*o.D < 8192 && o.Alpha == 256 && o.Blend != Lighter {
		column(dst, src, o)
		return
	}
	if !forceGeneral && o.B == 0 && o.C == 0 && scaled(dst, src, o) {
		return
	}
	if !forceGeneral && affine(dst, src, o) {
		return
	}
	det := float64(o.A)*float64(o.D) - float64(o.B)*float64(o.C)
	if det == 0 {
		return
	}
	w, h := float64(src.Bounds.Dx()), float64(src.Bounds.Dy())
	a, b, c, d, tx, ty := float64(o.A), float64(o.B), float64(o.C), float64(o.D), float64(o.TX), float64(o.TY)
	xs := [4]float64{tx, a*w + tx, b*h + tx, a*w + b*h + tx}
	ys := [4]float64{ty, c*w + ty, d*h + ty, c*w + d*h + ty}
	minx, maxx, miny, maxy := xs[0], xs[0], ys[0], ys[0]
	for i := 1; i < 4; i++ {
		minx = math.Min(minx, xs[i])
		maxx = math.Max(maxx, xs[i])
		miny = math.Min(miny, ys[i])
		maxy = math.Max(maxy, ys[i])
	}
	minx = math.Max(minx, float64(dst.Bounds.Min.X))
	maxx = math.Min(maxx, float64(dst.Bounds.Max.X))
	miny = math.Max(miny, float64(dst.Bounds.Min.Y))
	maxy = math.Min(maxy, float64(dst.Bounds.Max.Y))
	if minx >= maxx || miny >= maxy {
		return
	}
	x0, x1, y0, y1 := int(math.Floor(minx)), int(math.Ceil(maxx)), int(math.Floor(miny)), int(math.Ceil(maxy))
	ia, ib, ic, id := d/det, -b/det, -c/det, a/det
	// 32 fractional bits keep accumulated sampling error below a pixel for small
	// displays. Coordinates and increments are integers inside the pixel loop.
	const unit = 4294967296.0
	if math.Abs(ia) > 1e6 || math.Abs(ib) > 1e6 || math.Abs(ic) > 1e6 || math.Abs(id) > 1e6 {
		return
	}
	dx, dy := int64(math.Round(ia*unit)), int64(math.Round(ic*unit))
	// Row starts also advance in fixed point: float64 is software-emulated on
	// RP2040 and ESP32, and one-pixel-wide draws are all row setup. A row start is
	// linear in y, so the first and last rows bound the rest; beyond them the
	// source coordinate is over 10^9 pixels away and nothing is visible.
	fx, fy := float64(x0)+.5-tx, float64(y0)+.5-ty
	u, v := (ia*fx+ib*fy)*unit, (ic*fx+id*fy)*unit
	du, dv := ib*unit, id*unit
	rows := float64(y1 - 1 - y0)
	if math.Abs(u) > 8e18 || math.Abs(v) > 8e18 || math.Abs(u+du*rows) > 8e18 || math.Abs(v+dv*rows) > 8e18 {
		return
	}
	rowX, rowY := int64(math.Round(u)), int64(math.Round(v))
	stepX, stepY := int64(math.Round(du)), int64(math.Round(dv))
	for y := y0; y < y1; y, rowX, rowY = y+1, rowX+stepX, rowY+stepY {
		sx, sy := rowX, rowY
		for x := x0; x < x1; x++ {
			px, py := int(sx>>32), int(sy>>32)
			if px >= 0 && px < src.Bounds.Dx() && py >= 0 && py < src.Bounds.Dy() {
				shade(dst.Buffer, dst.Index(x, y), src.Buffer, src.Offset+py*src.Stride+px, &o)
			}
			sx += dx
			sy += dy
		}
	}
}

// scaled draws positive axis-aligned scales. A raycaster's one-pixel-wide
// columns spend most of their time in setup, so it uses float32: float64 is
// software on RP2040 and ESP32, and ESP32-S3 has only a float32 unit. Sampling
// can differ from the general path only where a pixel center falls within
// float32 rounding of a texel edge. It reports false for values it does not
// handle, which then take the general path.
func scaled(dst, src View, o Options) bool {
	if o.A < 1e-6 || o.D < 1e-6 || o.A > 32767 || o.D > 32767 || o.TX < -1e7 || o.TX > 1e7 || o.TY < -1e7 || o.TY > 1e7 {
		return false
	}
	w, h := src.Bounds.Dx(), src.Bounds.Dy()
	if o.A*float32(w) > 1e9 || o.D*float32(h) > 1e9 {
		return false
	}
	x0, x1 := span(o.TX, o.A*float32(w), dst.Bounds.Min.X, dst.Bounds.Max.X)
	y0, y1 := span(o.TY, o.D*float32(h), dst.Bounds.Min.Y, dst.Bounds.Max.Y)
	if x0 >= x1 || y0 >= y1 {
		return true
	}
	ia, id := 1/o.A, 1/o.D
	// Source coordinates of the first and last pixel centers. They increase with
	// a positive scale, so a span that misses the source draws nothing.
	u0, u1 := (float32(x0)+.5-o.TX)*ia, (float32(x1)-.5-o.TX)*ia
	v0, v1 := (float32(y0)+.5-o.TY)*id, (float32(y1)-.5-o.TY)*id
	if u1 < 0 || u0 >= float32(w) || v1 < 0 || v0 >= float32(h) {
		return true
	}
	// Past 1e9 source pixels the 32.32 fixed-point coordinates could overflow.
	if u1-u0 > 1e9 || v1-v0 > 1e9 {
		return false
	}
	const unit = 4294967296.0
	dx, dy := int64(ia*unit), int64(id*unit)
	rowX, sy := int64(u0*unit), int64(v0*unit)
	// At full alpha without Lighter a pixel depends only on its texel, so the
	// rows that sample one source row (several, when magnified) are filled
	// together: each texel is shaded once and written down the run of rows.
	direct := o.Alpha == 256 && o.Blend != Lighter
	db, sb := dst.Buffer, src.Buffer
	pixels, mask, spix, smask, sdata := db.Pixels, db.Mask, sb.Pixels, sb.Mask, sb.Data
	sopaque := sb.Opaque || smask == nil
	stride := dst.Stride
	tinted := o.R != 256 || o.G != 256 || o.Blue != 256
	// An opaque destination's mask bits are already set.
	marking := !db.Opaque && mask != nil
	// The source column of each destination column is the same on every
	// row, so it is worked out once, without 64-bit sums in the pixel loop.
	n := x1 - x0
	var cols []int32
	if direct && n <= len(columnTable) {
		cols = columnTable[:n]
		sx := rowX
		for i := range cols {
			if px := int(sx >> 32); px >= 0 && px < w {
				cols[i] = int32(px)
			} else {
				cols[i] = -1
			}
			sx += dx
		}
	}
	for y := y0; y < y1; {
		py := int(sy >> 32)
		end := y + 1
		sy += dy
		for direct && end < y1 && int(sy>>32) == py {
			end++
			sy += dy
		}
		if py >= 0 && py < h {
			si := src.Offset + py*src.Stride
			di := dst.Index(x0, y)
			last := di + (end-y-1)*dst.Stride
			if cols != nil && !tinted && sdata == nil && end == y+1 && o.Blend != Copy {
				// One row from one source row in RAM, as when shrinking.
				out := pixels[di : di+n]
				for i, c := range cols {
					if c < 0 {
						continue
					}
					p := si + int(c)
					if !sopaque && smask[p>>3]&(1<<uint(p&7)) == 0 {
						continue
					}
					out[i] = spix[p]
					if marking {
						q := di + i
						mask[q>>3] |= 1 << uint(q&7)
					}
				}
				y = end
				continue
			}
			sx := rowX
			for i := 0; i < n; i, sx = i+1, sx+dx {
				px := int(sx >> 32)
				if cols != nil {
					px = int(cols[i])
				}
				if px < 0 || px >= w {
					continue
				}
				if !direct {
					shade(db, di+i, sb, si+px, &o)
					continue
				}
				p := si + px
				if !sopaque && smask[p>>3]&(1<<uint(p&7)) == 0 {
					if o.Blend == Copy {
						for d := di + i; d <= last+i; d += stride {
							db.Set(d, 0, false)
						}
						marking = mask != nil
					}
					continue
				}
				var v uint16
				if sdata != nil {
					v = uint16(sdata[2*p]) | uint16(sdata[2*p+1])<<8
				} else {
					v = spix[p]
				}
				if tinted {
					v = tint(v, &o)
				}
				for d := di + i; d <= last+i; d += stride {
					pixels[d] = v
					if marking {
						mask[d>>3] |= 1 << uint(d&7)
					}
				}
			}
		}
		y = end
	}
	return true
}

// columnTable holds scaled's source columns. It is kept here: on the stack
// of a microcontroller it would be too large and go to the heap.
var columnTable [512]int32

// tint scales a pixel's color. It matches shade at full alpha without Lighter.
func tint(p uint16, o *Options) uint16 {
	if o.R == 256 && o.G == 256 && o.Blue == 256 {
		return p
	}
	r, g, b := Unpack(p)
	return Pack(clamp(int32(r)*o.R/256), clamp(int32(g)*o.G/256), clamp(int32(b)*o.Blue/256))
}

// affine draws rotations, flips and skews like the general path, with float32
// setup and, without a color scale, a pixel loop that copies directly. A
// rotated floor drawn row by row spends most of its time here: float64 is
// software on every microcontroller target, and shade is a call per pixel.
// Sampling can differ from the general path only where a pixel center falls
// within float32 rounding of a texel edge. It reports false for values it does
// not handle, which then take the general path.
func affine(dst, src View, o Options) bool {
	const limit = 32767
	if o.A < -limit || o.A > limit || o.B < -limit || o.B > limit || o.C < -limit || o.C > limit ||
		o.D < -limit || o.D > limit || o.TX < -1e7 || o.TX > 1e7 || o.TY < -1e7 || o.TY > 1e7 {
		return false
	}
	det := o.A*o.D - o.B*o.C
	if det > -1e-12 && det < 1e-12 {
		return false
	}
	sw, sh := src.Bounds.Dx(), src.Bounds.Dy()
	w, h := float32(sw), float32(sh)
	xs := [4]float32{o.TX, o.A*w + o.TX, o.B*h + o.TX, o.A*w + o.B*h + o.TX}
	ys := [4]float32{o.TY, o.C*w + o.TY, o.D*h + o.TY, o.C*w + o.D*h + o.TY}
	minx, maxx, miny, maxy := xs[0], xs[0], ys[0], ys[0]
	for i := 1; i < 4; i++ {
		minx, maxx = min(minx, xs[i]), max(maxx, xs[i])
		miny, maxy = min(miny, ys[i]), max(maxy, ys[i])
	}
	// Clip before converting to int: the corners can be far outside.
	minx, maxx = max(minx, float32(dst.Bounds.Min.X)), min(maxx, float32(dst.Bounds.Max.X))
	miny, maxy = max(miny, float32(dst.Bounds.Min.Y)), min(maxy, float32(dst.Bounds.Max.Y))
	if minx >= maxx || miny >= maxy {
		return true
	}
	x0, x1 := span(minx, maxx-minx, dst.Bounds.Min.X, dst.Bounds.Max.X)
	y0, y1 := span(miny, maxy-miny, dst.Bounds.Min.Y, dst.Bounds.Max.Y)
	if x0 >= x1 || y0 >= y1 {
		return true
	}
	ia, ib, ic, id := o.D/det, -o.B/det, -o.C/det, o.A/det
	// Source coordinates are linear in the destination, so the corners of the
	// clipped area bound them. Past 1e9 the 32.32 fixed point could overflow.
	fx0, fx1 := float32(x0)+.5-o.TX, float32(x1)-.5-o.TX
	fy0, fy1 := float32(y0)+.5-o.TY, float32(y1)-.5-o.TY
	// small reports whether every source coordinate is within 16.16 fixed
	// point, with room for a pixel's step.
	small := sw < 16384 && sh < 16384
	for _, c := range [4][2]float32{{fx0, fy0}, {fx1, fy0}, {fx0, fy1}, {fx1, fy1}} {
		u, v := ia*c[0]+ib*c[1], ic*c[0]+id*c[1]
		if u < -1e9 || u > 1e9 || v < -1e9 || v > 1e9 {
			return false
		}
		if u < -16384 || u > 16384 || v < -16384 || v > 16384 {
			small = false
		}
	}
	const unit = 4294967296.0
	dx, dy := int64(ia*unit), int64(ic*unit)
	rowX, rowY := int64((ia*fx0+ib*fy0)*unit), int64((ic*fx0+id*fy0)*unit)
	stepX, stepY := int64(ib*unit), int64(id*unit)
	plain := o.R == 256 && o.G == 256 && o.Blue == 256 && o.Alpha == 256 && o.Blend != Lighter
	opaque := src.Buffer.Opaque || src.Buffer.Mask == nil
	pixels, mask := dst.Buffer.Pixels, dst.Buffer.Mask
	// An opaque destination's mask bits are already set.
	marking := !dst.Buffer.Opaque && mask != nil
	data, spix, smask := src.Buffer.Data, src.Buffer.Pixels, src.Buffer.Mask
	if plain && opaque && sw == 1 && sh == 1 {
		// A solid parallelogram, such as a band of sky: on each row the
		// pixels that sample the one source pixel are a span, found from
		// the same fixed-point coordinates the loop below steps through.
		v := src.Buffer.Pixel(src.Offset)
		n := x1 - x0
		for y := y0; y < y1; y, rowX, rowY = y+1, rowX+stepX, rowY+stepY {
			lo, hi := within(rowX, dx, 0, n)
			lo, hi = within(rowY, dy, lo, hi)
			if lo >= hi {
				continue
			}
			di := dst.Index(x0, y)
			fillPixels(pixels[di+lo:di+hi], v)
			if marking {
				dst.Buffer.SetMaskRange(di+lo, hi-lo, true)
			}
		}
		return true
	}
	if plain && opaque && small && (ia > -1 && ia < 1 && ic > -1 && ic < 1) {
		affineOpaque(dst, src, x0, x1, y0, y1, rowX, rowY, stepX, stepY, dx, dy, marking)
		return true
	}
	for y := y0; y < y1; y, rowX, rowY = y+1, rowX+stepX, rowY+stepY {
		di := dst.Index(x0, y)
		sx, sy := rowX, rowY
		for i := 0; i < x1-x0; i, sx, sy = i+1, sx+dx, sy+dy {
			px, py := int(sx>>32), int(sy>>32)
			if uint(px) >= uint(sw) || uint(py) >= uint(sh) {
				continue
			}
			p := src.Offset + py*src.Stride + px
			if !plain {
				shade(dst.Buffer, di+i, src.Buffer, p, &o)
				continue
			}
			if !opaque && smask[p>>3]&(1<<uint(p&7)) == 0 {
				if o.Blend == Copy {
					dst.Buffer.Set(di+i, 0, false)
					marking = mask != nil
				}
				continue
			}
			q := di + i
			if data != nil {
				pixels[q] = uint16(data[2*p]) | uint16(data[2*p+1])<<8
			} else {
				pixels[q] = spix[p]
			}
			if marking {
				mask[q>>3] |= 1 << uint(q&7)
			}
		}
	}
	return true
}

// within narrows lo to hi to the steps i at which the 32.32 fixed-point
// coordinate start+i*step is in the first source pixel, 0 to 1 exclusive.
func within(start, step int64, lo, hi int) (int, int) {
	const one = 1 << 32
	switch {
	case step == 0:
		if start < 0 || start >= one {
			return 0, 0
		}
		return lo, hi
	case step > 0:
		// start+i*step >= 0 and start+i*step <= one-1.
		lo = max(lo, int(max(ceilDiv(-start, step), int64(lo))))
		hi = min(hi, int(min(floorDiv(one-1-start, step)+1, int64(hi))))
	default:
		// start+i*step <= one-1 and start+i*step >= 0, the step negative.
		lo = max(lo, int(max(ceilDiv(one-1-start, step), int64(lo))))
		hi = min(hi, int(min(floorDiv(-start, step)+1, int64(hi))))
	}
	return lo, hi
}

// floorDiv and ceilDiv divide rounding down and up, for either sign.
func floorDiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func ceilDiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) == (b < 0)) {
		q++
	}
	return q
}

// affineOpaque is affine's pixel loop for an opaque source without a color
// scale, a step of under a source pixel along the row and coordinates within
// 16.16 fixed point: as a row of a floor or a road. Its 32-bit sums cost
// less than 32.32 ones on a 32-bit processor; they can drift from the 32.32
// ones by a 65536th of a pixel a step, so a sample can differ from the
// general path only within that of a texel's edge. An image in flash is read
// a pixel at a time where its data is aligned to two bytes.
func affineOpaque(dst, src View, x0, x1, y0, y1 int, rowX, rowY, stepX, stepY, dx, dy int64, marking bool) {
	sw, sh := uint32(src.Bounds.Dx()), uint32(src.Bounds.Dy())
	pixels, mask := dst.Buffer.Pixels, dst.Buffer.Mask
	data, spix := src.Buffer.Data, src.Buffer.Pixels
	aligned := data != nil && uintptr(unsafe.Pointer(&data[0]))&1 == 0
	// Steps rounded to the nearest 65536th, so the drift has no bias.
	du, dv := int32((dx+1<<15)>>16), int32((dy+1<<15)>>16)
	n := x1 - x0
	for y := y0; y < y1; y, rowX, rowY = y+1, rowX+stepX, rowY+stepY {
		di := dst.Index(x0, y)
		out := pixels[di : di+n]
		u, v := int32(rowX>>16), int32(rowY>>16)
		for i := range out {
			px, py := uint32(u>>16), uint32(v>>16)
			u, v = u+du, v+dv
			if px >= sw || py >= sh {
				continue
			}
			p := src.Offset + int(py)*src.Stride + int(px)
			switch {
			case data == nil:
				out[i] = spix[p]
			case aligned:
				out[i] = *(*uint16)(unsafe.Pointer(&data[2*p]))
			default:
				out[i] = uint16(data[2*p]) | uint16(data[2*p+1])<<8
			}
			if marking {
				q := di + i
				mask[q>>3] |= 1 << uint(q&7)
			}
		}
	}
}

// span returns the destination pixels covering start to start+length, clipped
// to lo and hi. Pixels whose centers fall outside are rejected while sampling.
func span(start, length float32, lo, hi int) (int, int) {
	first, end := int(start), int(start+length)
	if float32(first) > start {
		first--
	}
	if float32(end) < start+length {
		end++
	}
	return max(lo, first), min(hi, end)
}

// integer draws axis-aligned integer scales row by row. Source columns advance
// by stepping, so the pixel loops do not divide.
func integer(dst, src View, o Options) {
	sx, sy, tx, ty := int(o.A), int(o.D), int(o.TX), int(o.TY)
	x0, y0 := max(dst.Bounds.Min.X, tx), max(dst.Bounds.Min.Y, ty)
	x1, y1 := min(dst.Bounds.Max.X, tx+src.Bounds.Dx()*sx), min(dst.Bounds.Max.Y, ty+src.Bounds.Dy()*sy)
	if x0 >= x1 || y0 >= y1 {
		return
	}
	n := x1 - x0
	plain := o.R == 256 && o.G == 256 && o.Blue == 256 && o.Alpha == 256 && o.Blend != Lighter
	opaque := src.Buffer.Opaque || src.Buffer.Mask == nil
	firstColumn, firstRepeat := (x0-tx)/sx, (x0-tx)%sx
	db, sb := dst.Buffer, src.Buffer
	dpix, spix, sdata, smask := db.Pixels, sb.Pixels, sb.Data, sb.Mask
	for y := y0; y < y1; y++ {
		si := src.Offset + (y-ty)/sy*src.Stride
		di := dst.Index(x0, y)
		switch {
		case plain && sx == 1 && opaque:
			opaqueRow(db, di, sb, si+firstColumn, n)
		case plain && sx == 1:
			maskedRow(db, di, sb, si+firstColumn, n, o.Blend)
		case plain && opaque:
			// Each source pixel fills a run, such as a 1x1 image scaled to a rectangle.
			out := dpix[di : di+n]
			if src.Bounds.Dx() == 1 {
				fillPixels(out, sb.Pixel(si))
			} else {
				column, run := firstColumn, sx-firstRepeat
				for i := 0; i < n; column, run = column+1, sx {
					v := sb.Pixel(si + column)
					fillPixels(out[i:min(n, i+run)], v)
					i += run
				}
			}
			// An opaque destination's mask bits are already set.
			if !db.Opaque {
				db.SetMaskRange(di, n, true)
			}
		case plain && o.Blend != Copy:
			// Magnified with a mask: each source pixel's bit is read once,
			// and a shown pixel is written as a run sx wide. From flash this
			// also reads each pixel once, and flash reads are slow.
			dmask := db.Mask
			marking := !db.Opaque && dmask != nil
			out := dpix[di : di+n]
			column, i := firstColumn, 0
			for run := sx - firstRepeat; i < n; column, run = column+1, sx {
				end := min(n, i+run)
				p := si + column
				if smask[p>>3]&(1<<uint(p&7)) != 0 {
					var v uint16
					if sdata != nil {
						v = uint16(sdata[2*p]) | uint16(sdata[2*p+1])<<8
					} else {
						v = spix[p]
					}
					for j := i; j < end; j++ {
						out[j] = v
					}
					if marking {
						db.SetMaskRange(di+i, end-i, true)
					}
				}
				i = end
			}
		case plain:
			column, repeat := firstColumn, firstRepeat
			dmask := db.Mask
			marking := !db.Opaque && dmask != nil
			for i := 0; i < n; i++ {
				if p := si + column; smask[p>>3]&(1<<uint(p&7)) != 0 {
					q := di + i
					if sdata != nil {
						dpix[q] = uint16(sdata[2*p]) | uint16(sdata[2*p+1])<<8
					} else {
						dpix[q] = spix[p]
					}
					if marking {
						dmask[q>>3] |= 1 << uint(q&7)
					}
				} else if o.Blend == Copy {
					db.Set(di+i, 0, false)
					marking = dmask != nil
				}
				if repeat++; repeat == sx {
					repeat = 0
					column++
				}
			}
		default:
			column, repeat := firstColumn, firstRepeat
			for i := 0; i < n; i++ {
				shade(db, di+i, sb, si+column, &o)
				if repeat++; repeat == sx {
					repeat = 0
					column++
				}
			}
		}
	}
}

// mirrored draws an image flipped left to right, at a whole pixel and
// scaled up a whole number of times down, without a color scale: a sprite
// facing the other way. Destination x samples source column tx-x-1.
func mirrored(dst, src View, o Options) {
	sy, tx, ty := int(o.D), int(o.TX), int(o.TY)
	x0, x1 := max(dst.Bounds.Min.X, tx-src.Bounds.Dx()), min(dst.Bounds.Max.X, tx)
	y0, y1 := max(dst.Bounds.Min.Y, ty), min(dst.Bounds.Max.Y, ty+src.Bounds.Dy()*sy)
	if x0 >= x1 || y0 >= y1 {
		return
	}
	db, sb := dst.Buffer, src.Buffer
	dpix, dmask, spix, smask, sdata := db.Pixels, db.Mask, sb.Pixels, sb.Mask, sb.Data
	opaque := sb.Opaque || smask == nil
	// An opaque destination's mask bits are already set.
	marking := !db.Opaque && dmask != nil
	n := x1 - x0
	for y := y0; y < y1; y++ {
		// The source pixel for the row's first destination pixel; the
		// following ones go leftward from it.
		p0 := src.Offset + (y-ty)/sy*src.Stride + tx - x0 - 1
		di := dst.Index(x0, y)
		out := dpix[di : di+n]
		for i := range out {
			p := p0 - i
			if !opaque && smask[p>>3]&(1<<uint(p&7)) == 0 {
				if o.Blend == Copy {
					db.Set(di+i, 0, false)
					marking = dmask != nil
				}
				continue
			}
			if sdata != nil {
				out[i] = uint16(sdata[2*p]) | uint16(sdata[2*p+1])<<8
			} else {
				out[i] = spix[p]
			}
			if marking {
				q := di + i
				dmask[q>>3] |= 1 << uint(q&7)
			}
		}
	}
}

// column draws a one-pixel-wide image stretched or shrunk down, a whole
// number of pixels wide at a whole pixel across, at full alpha: a
// raycaster's wall. The rows of each source pixel, those whose centers fall
// within it, are worked out once and written as a run, where the scaled
// path worked out every row. Sampling can differ from the general path only
// where a row's center falls within 1/65536 of a texel's edge.
func column(dst, src View, o Options) {
	tx := int(o.TX)
	x0, x1 := max(dst.Bounds.Min.X, tx), min(dst.Bounds.Max.X, tx+int(o.A))
	if x0 >= x1 {
		return
	}
	n := x1 - x0
	db, sb := dst.Buffer, src.Buffer
	pixels, mask, smask := db.Pixels, db.Mask, sb.Mask
	opaque := sb.Opaque || smask == nil
	// An opaque destination's mask bits are already set.
	marking := !db.Opaque && mask != nil
	tinted := o.R != 256 || o.G != 256 || o.Blue != 256
	ymin, ymax := dst.Bounds.Min.Y, dst.Bounds.Max.Y
	h := src.Bounds.Dy()
	// Row y's center samples texel t where t*D <= y+0.5-TY < (t+1)*D. The
	// edges t*D+TY-0.5 are stepped through in 16.16 fixed point: float32
	// arithmetic in the loop cost about a microsecond a texel on the badge.
	// The checks in draw keep them within int32.
	step := int32(o.D*65536 + 0.5)
	edge := int32((o.TY-0.5)*65536 + 0.5)
	top := int((edge + 0xffff) >> 16) // rounded up
	for t := range h {
		edge += step
		next := int((edge + 0xffff) >> 16)
		a, b := max(top, ymin), min(next, ymax)
		top = next
		if a >= b {
			if a >= ymax {
				return
			}
			continue
		}
		p := src.Offset + t*src.Stride
		if !opaque && smask[p>>3]&(1<<uint(p&7)) == 0 {
			if o.Blend == Copy {
				for y := a; y < b; y++ {
					di := dst.Index(x0, y)
					for i := range n {
						db.Set(di+i, 0, false)
					}
				}
				marking = mask != nil
			}
			continue
		}
		v := sb.Pixel(p)
		if tinted {
			v = tint(v, &o)
		}
		di := dst.Index(x0, a)
		for y := a; y < b; y, di = y+1, di+dst.Stride {
			for i := range n {
				pixels[di+i] = v
			}
			if marking {
				db.SetMaskRange(di, n, true)
			}
		}
	}
}

func opaqueRow(dst *Buffer, di int, src *Buffer, si, n int) {
	out := dst.Pixels[di : di+n]
	if data := src.Data; data == nil {
		copyPixels(out, src.Pixels[si:si+n])
	} else {
		data = data[2*si : 2*(si+n)]
		for i := range out {
			out[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
		}
	}
	// An opaque destination's mask bits are already set.
	if !dst.Opaque {
		dst.SetMaskRange(di, n, true)
	}
}

func maskedRow(dst *Buffer, di int, src *Buffer, si, n int, blend Blend) {
	out := dst.Pixels[di : di+n]
	mask, pix, data, dmask := src.Mask, src.Pixels, src.Data, dst.Mask
	// An opaque destination's mask bits are already set.
	marking := !dst.Opaque && dmask != nil
	if blend != Copy {
		// Only shown pixels are written, so the mask is read a byte, eight
		// pixels, at a time: a clear byte is skipped and a full one copied.
		if data != nil {
			maskedData(out, di, src, si, n, dmask, marking)
			return
		}
		pix = pix[si : si+n]
		for i := 0; i < n; {
			p := si + i
			k := min(8-p&7, n-i) // the pixels this byte of the mask covers
			full := byte(1<<uint(k) - 1)
			m := mask[p>>3] >> uint(p&7) & full
			if m == 0 {
				i += k
				continue
			}
			o, s := out[i:i+k], pix[i:i+k]
			if m == full {
				for j := range o {
					o[j] = s[j]
				}
			} else {
				for j := 0; m != 0; j, m = j+1, m>>1 {
					if m&1 != 0 {
						o[j] = s[j]
					}
				}
				m = mask[p>>3] >> uint(p&7) & full
			}
			if marking {
				for j := 0; m != 0; j, m = j+1, m>>1 {
					if m&1 != 0 {
						q := di + i + j
						dmask[q>>3] |= 1 << uint(q&7)
					}
				}
			}
			i += k
		}
		return
	}
	for i := range out {
		p := si + i
		if mask[p>>3]&(1<<uint(p&7)) == 0 {
			if blend == Copy {
				dst.Set(di+i, 0, false)
				marking = dmask != nil
			}
			continue
		}
		if data == nil {
			out[i] = pix[p]
		} else {
			out[i] = uint16(data[2*p]) | uint16(data[2*p+1])<<8
		}
		if marking {
			q := di + i
			dmask[q>>3] |= 1 << uint(q&7)
		}
	}
}

// maskedData is maskedRow's byte-at-a-time loop for an image in flash, whose
// pixels are read as 16-bit words when the data is aligned.
func maskedData(out []uint16, di int, src *Buffer, si, n int, dmask []byte, marking bool) {
	mask, data := src.Mask, src.Data
	aligned := uintptr(unsafe.Pointer(&data[0]))&1 == 0
	for i := 0; i < n; {
		p := si + i
		k := min(8-p&7, n-i)
		full := byte(1<<uint(k) - 1)
		m := mask[p>>3] >> uint(p&7) & full
		if m == 0 {
			i += k
			continue
		}
		o := out[i : i+k]
		switch {
		case m == full && aligned:
			for j := range o {
				o[j] = *(*uint16)(unsafe.Pointer(&data[2*(p+j)]))
			}
		case m == full:
			for j := range o {
				o[j] = uint16(data[2*(p+j)]) | uint16(data[2*(p+j)+1])<<8
			}
		default:
			for j, b := 0, m; b != 0; j, b = j+1, b>>1 {
				if b&1 != 0 {
					o[j] = uint16(data[2*(p+j)]) | uint16(data[2*(p+j)+1])<<8
				}
			}
		}
		if marking {
			for j := 0; m != 0; j, m = j+1, m>>1 {
				if m&1 != 0 {
					q := di + i + j
					dmask[q>>3] |= 1 << uint(q&7)
				}
			}
		}
		i += k
	}
}

func shade(dst *Buffer, di int, src *Buffer, si int, o *Options) {
	if !src.Visible(si) {
		if o.Blend == Copy {
			dst.Set(di, 0, false)
		}
		return
	}
	if o.Alpha == 0 {
		if o.Blend == Copy {
			dst.Set(di, 0, false)
		}
		return
	}
	sr, sg, sb := Unpack(src.Pixel(si))
	r, g, b := int32(sr)*o.R/256, int32(sg)*o.G/256, int32(sb)*o.Blue/256
	a := o.Alpha
	if o.Blend != Copy && dst.Visible(di) {
		dr, dg, db := Unpack(dst.Pixel(di))
		f := int32(256) - a
		if o.Blend == Lighter {
			f = 256
		}
		r += int32(dr) * f / 256
		g += int32(dg) * f / 256
		b += int32(db) * f / 256
	}
	// Binary destination alpha: partial results are flattened to opaque RGB565.
	dst.Set(di, Pack(clamp(r), clamp(g), clamp(b)), true)
}
func clamp(v int32) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
