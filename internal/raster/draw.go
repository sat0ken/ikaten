package raster

import "math"

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
			sx := rowX
			for i := 0; i < x1-x0; i, sx = i+1, sx+dx {
				px := int(sx >> 32)
				if px < 0 || px >= w {
					continue
				}
				if !direct {
					shade(dst.Buffer, di+i, src.Buffer, si+px, &o)
					continue
				}
				visible := src.Buffer.Visible(si + px)
				if !visible && o.Blend != Copy {
					continue
				}
				v := uint16(0)
				if visible {
					v = tint(src.Buffer.Pixel(si+px), &o)
				}
				if !visible {
					for d := di + i; d <= last+i; d += dst.Stride {
						dst.Buffer.Set(d, 0, false)
					}
					continue
				}
				// An opaque destination's mask bits are already set.
				pixels, mask := dst.Buffer.Pixels, dst.Buffer.Mask
				marking := !dst.Buffer.Opaque && mask != nil
				for d := di + i; d <= last+i; d += dst.Stride {
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
	for _, c := range [4][2]float32{{fx0, fy0}, {fx1, fy0}, {fx0, fy1}, {fx1, fy1}} {
		u, v := ia*c[0]+ib*c[1], ic*c[0]+id*c[1]
		if u < -1e9 || u > 1e9 || v < -1e9 || v > 1e9 {
			return false
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
	data := src.Buffer.Data
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
			if !opaque && !src.Buffer.Visible(p) {
				if o.Blend == Copy {
					dst.Buffer.Set(di+i, 0, false)
				}
				continue
			}
			q := di + i
			if data != nil {
				pixels[q] = uint16(data[2*p]) | uint16(data[2*p+1])<<8
			} else {
				pixels[q] = src.Buffer.Pixels[p]
			}
			if marking {
				mask[q>>3] |= 1 << uint(q&7)
			}
		}
	}
	return true
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
	for y := y0; y < y1; y++ {
		si := src.Offset + (y-ty)/sy*src.Stride
		di := dst.Index(x0, y)
		switch {
		case plain && sx == 1 && opaque:
			opaqueRow(dst.Buffer, di, src.Buffer, si+firstColumn, n)
		case plain && sx == 1:
			maskedRow(dst.Buffer, di, src.Buffer, si+firstColumn, n, o.Blend)
		case plain && opaque:
			// Each source pixel fills a run, such as a 1x1 image scaled to a rectangle.
			out := dst.Buffer.Pixels[di : di+n]
			if y > y0 && (y-ty)/sy == (y-1-ty)/sy {
				// The same source row as the row above: copy that.
				copy(out, dst.Buffer.Pixels[di-dst.Stride:])
			} else if src.Bounds.Dx() == 1 {
				// One color across: double what is filled until it is full.
				out[0] = src.Buffer.Pixel(si)
				for k := 1; k < n; k *= 2 {
					copy(out[k:], out[:k])
				}
			} else {
				column, run := firstColumn, sx-firstRepeat
				for i := 0; i < n; column, run = column+1, sx {
					v := src.Buffer.Pixel(si + column)
					for end := min(n, i+run); i < end; i++ {
						out[i] = v
					}
				}
			}
			// An opaque destination's mask bits are already set.
			if !dst.Buffer.Opaque {
				dst.Buffer.SetMaskRange(di, n, true)
			}
		default:
			column, repeat := firstColumn, firstRepeat
			for i := 0; i < n; i++ {
				if plain {
					if p := si + column; src.Buffer.Visible(p) {
						dst.Buffer.Set(di+i, src.Buffer.Pixel(p), true)
					} else if o.Blend == Copy {
						dst.Buffer.Set(di+i, 0, false)
					}
				} else {
					shade(dst.Buffer, di+i, src.Buffer, si+column, &o)
				}
				if repeat++; repeat == sx {
					repeat = 0
					column++
				}
			}
		}
	}
}

func opaqueRow(dst *Buffer, di int, src *Buffer, si, n int) {
	out := dst.Pixels[di : di+n]
	if src.Data == nil {
		copy(out, src.Pixels[si:si+n])
	} else {
		data := src.Data[2*si : 2*(si+n)]
		for i := range out {
			out[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
		}
	}
	dst.SetMaskRange(di, n, true)
}

func maskedRow(dst *Buffer, di int, src *Buffer, si, n int, blend Blend) {
	out := dst.Pixels[di : di+n]
	mask := src.Mask
	// An opaque destination's mask bits are already set.
	marking := !dst.Opaque && dst.Mask != nil
	for i := range out {
		p := si + i
		if mask[p>>3]&(1<<uint(p&7)) == 0 {
			if blend == Copy {
				dst.Set(di+i, 0, false)
			}
			continue
		}
		if src.Data == nil {
			out[i] = src.Pixels[p]
		} else {
			out[i] = uint16(src.Data[2*p]) | uint16(src.Data[2*p+1])<<8
		}
		if marking {
			q := di + i
			dst.Mask[q>>3] |= 1 << uint(q&7)
		}
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
