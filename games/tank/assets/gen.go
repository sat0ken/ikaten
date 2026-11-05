//go:build ignore

// gen paints the tank game's assets from package field: ground.png, the
// battlefield seen from above, sprites.png, the sprite sheet, and sight.png, the
// zoomed gunsight's mask. Tanks and
// houses are boxes, ray cast from each direction.
// Usage: go run gen.go, then run go generate in games/tank.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/sat0ken/ikaten/games/tank/field"
)

func hash(x, y int) uint32 {
	h := uint32(x)*374761393 + uint32(y)*668265263
	h = (h ^ h>>13) * 1274126177
	return h ^ h>>16
}

func shade(c color.NRGBA, d int) color.NRGBA {
	f := func(v uint8) uint8 { return uint8(min(max(int(v)+d, 0), 255)) }
	return color.NRGBA{f(c.R), f(c.G), f(c.B), 255}
}

func scale(c color.NRGBA, s float64) color.NRGBA {
	f := func(v uint8) uint8 { return uint8(min(max(float64(v)*s, 0), 255)) }
	return color.NRGBA{f(c.R), f(c.G), f(c.B), 255}
}

// segmentDistance returns the distance from p to the segment ab.
func segmentDistance(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	t = min(max(t, 0), 1)
	return math.Hypot(px-ax-t*dx, py-ay-t*dy)
}

var fieldColors = [...]color.NRGBA{
	{84, 146, 58, 255},
	{100, 158, 62, 255},
	{116, 162, 66, 255},
	{148, 156, 78, 255}, // ripe
	{126, 104, 72, 255}, // plowed
}

func ground() image.Image {
	const n = field.TextureSize
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	road := color.NRGBA{156, 134, 96, 255}
	for ty := range n {
		for tx := range n {
			wx := float64(tx*field.TexelMeters - field.Half + field.TexelMeters/2)
			wy := float64(ty*field.TexelMeters - field.Half + field.TexelMeters/2)
			// Patchwork fields, with the rows staggered.
			row := int(math.Floor((wy + field.Half) / 72))
			col := int(math.Floor((wx + field.Half + float64(hash(row, 7)%60)) / 88))
			kind := hash(col, row) % uint32(len(fieldColors))
			c := fieldColors[kind]
			if kind == 4 && ty%2 == 0 {
				c = shade(c, -14)
			}
			if kind == 3 && (tx+ty)%3 == 0 {
				c = shade(c, 10)
			}
			c = shade(c, int(hash(tx, ty)%11)-5)
			for _, line := range field.Roads {
				for i := 1; i < len(line); i++ {
					d := segmentDistance(wx, wy, float64(line[i-1].X), float64(line[i-1].Y), float64(line[i].X), float64(line[i].Y))
					switch {
					case d < 4:
						c = shade(road, int(hash(ty, tx)%13)-6)
					case d < 6.5 && kind != 4:
						c = shade(c, -12)
					}
				}
			}
			img.SetNRGBA(tx, ty, c)
		}
	}
	// Shade under trees and yards around houses.
	for _, o := range field.Obstacles {
		r, c := float64(o.Radius)+2.5, color.NRGBA{52, 98, 44, 255}
		if o.Kind == field.House {
			r, c = 9, color.NRGBA{124, 118, 100, 255}
		}
		paintDisc(img, float64(o.X), float64(o.Y), r, c)
	}
	// Shell craters.
	for i := range 40 {
		x := float64(hash(i, 1)%uint32(field.Size-80)) - field.Half + 40
		y := float64(hash(i, 2)%uint32(field.Size-80)) - field.Half + 40
		paintDisc(img, x, y, 6, color.NRGBA{98, 84, 60, 255})
		paintDisc(img, x, y, 3, color.NRGBA{72, 62, 46, 255})
	}
	return img
}

func paintDisc(img *image.NRGBA, x, y, r float64, c color.NRGBA) {
	m := float64(field.TexelMeters)
	for ty := int((y - r + field.Half) / m); ty <= int((y+r+field.Half)/m); ty++ {
		for tx := int((x - r + field.Half) / m); tx <= int((x+r+field.Half)/m); tx++ {
			wx, wy := float64(tx)*m-field.Half+m/2, float64(ty)*m-field.Half+m/2
			if math.Hypot(wx-x, wy-y) < r && image.Pt(tx, ty).In(img.Rect) {
				img.SetNRGBA(tx, ty, shade(c, int(hash(tx, ty)%9)-4))
			}
		}
	}
}

// A box in model space: x forward, y right, z up, in meters.
type box struct {
	x0, y0, z0, x1, y1, z1 float64
	c                      color.NRGBA
}

type vec [3]float64

func (a vec) dot(b vec) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

const pitch = 8 * math.Pi / 180

// render ray casts boxes into a cell of the sheet. rel is the model's heading
// relative to the line of sight: 0 shows its rear, pi its front.
func render(dst *image.NRGBA, ox, oy, w, h, ax, ay int, boxes []box, rel float64) {
	ppm := float64(field.PixelsPerMeter)
	sr, cr := math.Sincos(rel)
	fwd, right := vec{sr, cr, 0}, vec{cr, -sr, 0} // model axes in view space
	d := vec{0, math.Cos(pitch), -math.Sin(pitch)}
	up := vec{0, math.Sin(pitch), math.Cos(pitch)}
	light := vec{-0.45, -0.5, 0.74}
	toModel := func(v vec) vec { return vec{v.dot(fwd), v.dot(right), v[2]} }
	dir := toModel(d)
	for py := range h {
		for px := range w {
			sx, su := (float64(px)+.5-float64(ax))/ppm, -(float64(py)+.5-float64(ay))/ppm
			o := toModel(vec{sx + up[0]*su - d[0]*50, up[1]*su - d[1]*50, up[2]*su - d[2]*50})
			best, hit := math.Inf(1), (*box)(nil)
			var normal vec
			for i := range boxes {
				b := &boxes[i]
				lo, hi := vec{b.x0, b.y0, b.z0}, vec{b.x1, b.y1, b.z1}
				tmin, tmax := math.Inf(-1), math.Inf(1)
				var n vec
				for k := range 3 {
					if dir[k] == 0 {
						if o[k] < lo[k] || o[k] > hi[k] {
							tmin = math.Inf(1)
						}
						continue
					}
					t0, t1 := (lo[k]-o[k])/dir[k], (hi[k]-o[k])/dir[k]
					var face vec
					face[k] = -1
					if t0 > t1 {
						t0, t1 = t1, t0
						face[k] = 1
					}
					if t0 > tmin {
						tmin, n = t0, face
					}
					tmax = min(tmax, t1)
				}
				if tmin <= tmax && tmin < best {
					best, hit, normal = tmin, b, n
				}
			}
			if hit == nil {
				continue
			}
			nv := vec{normal[0]*fwd[0] + normal[1]*right[0], normal[0]*fwd[1] + normal[1]*right[1], normal[2]}
			lit := 0.55 + 0.5*max(0, nv.dot(light))
			dst.SetNRGBA(ox+px, oy+py, scale(hit.c, lit))
		}
	}
}

type tankModel struct {
	length, width, hullLow, hullTop float64
	turretBack, turretFront         float64
	turretWidth, turretTop          float64
	gun                             float64
	body                            color.NRGBA
}

// The enemy kinds: light, medium and heavy.
var tanks = [field.TankKinds]tankModel{
	{5.0, 2.6, 0.4, 1.3, -1.1, 0.9, 0.85, 1.9, 3.7, color.NRGBA{196, 170, 112, 255}},
	{6.4, 3.0, 0.45, 1.55, -1.4, 1.2, 1.1, 2.35, 4.6, color.NRGBA{138, 142, 128, 255}},
	{7.0, 3.4, 0.5, 1.75, -1.6, 1.4, 1.35, 2.65, 4.7, color.NRGBA{98, 104, 118, 255}},
}

func (m tankModel) hull() []box {
	track := color.NRGBA{48, 46, 44, 255}
	l, w := m.length/2, m.width/2
	return []box{
		{-l, -w, 0, l, -w + 0.7, m.hullLow + 0.6, track},
		{-l, w - 0.7, 0, l, w, m.hullLow + 0.6, track},
		{-l + 0.2, -w + 0.15, m.hullLow, l - 0.6, w - 0.15, m.hullTop, m.body},
		{l - 0.6, -w + 0.25, m.hullLow, l - 0.1, w - 0.25, m.hullTop - 0.35, scale(m.body, 0.9)},
		{-l + 0.1, -w + 0.4, m.hullTop, -l + 0.9, w - 0.4, m.hullTop + 0.15, scale(m.body, 0.8)},
	}
}

func (m tankModel) turret() []box {
	tw := m.turretWidth
	gunY, gunZ := 0.13, m.hullTop+(m.turretTop-m.hullTop)*0.45
	return []box{
		{m.turretBack, -tw, m.hullTop, m.turretFront, tw, m.turretTop, m.body},
		{m.turretFront, -tw * 0.55, m.hullTop + 0.15, m.turretFront + 0.35, tw * 0.55, m.turretTop - 0.15, scale(m.body, 0.85)},
		{m.turretFront + 0.35, -gunY, gunZ - gunY, m.gun, gunY, gunZ + gunY, scale(m.body, 0.7)},
		{m.turretBack + 0.3, -tw + 0.25, m.turretTop, m.turretBack + 0.9, -tw + 0.85, m.turretTop + 0.25, scale(m.body, 0.9)},
	}
}

func house() []box {
	wall, roof, dark := color.NRGBA{214, 204, 184, 255}, color.NRGBA{156, 70, 52, 255}, color.NRGBA{60, 52, 50, 255}
	boxes := []box{{-4.8, -3.4, 0, 4.8, 3.4, 4, wall}}
	for i := range 6 {
		hw := 3.9 * (1 - float64(i)/6.2)
		boxes = append(boxes, box{-5.2, -hw, 4 + float64(i)*0.5, 5.2, hw, 4.5 + float64(i)*0.5, roof})
	}
	for _, x := range []float64{-3, 0, 3} {
		boxes = append(boxes, box{x - 0.5, -3.5, 1.6, x + 0.5, 3.5, 2.6, dark})
	}
	boxes = append(boxes, box{-4.9, -0.6, 0, 4.9, 0.6, 2.2, scale(dark, 1.5)}) // doors
	boxes = append(boxes, box{2, 1, 6, 2.7, 1.7, 7.4, scale(wall, 0.7)})       // chimney
	return boxes
}

func tree(dst *image.NRGBA, ox, oy int, pine bool) {
	trunk := color.NRGBA{92, 66, 44, 255}
	for y := 30; y < field.TreeH; y++ {
		for x := 9; x < 11; x++ {
			dst.SetNRGBA(ox+x, oy+y, trunk)
		}
	}
	for y := range 34 {
		for x := range field.TreeW {
			fx, fy := float64(x)+.5-10, float64(y)+.5
			var in bool
			var lit float64
			if pine {
				// Three tiers, each wider at its foot.
				tier := math.Mod(fy, 11)
				half := 2 + tier*0.55 + fy*0.12
				in = fy > 0 && math.Abs(fx) < half
				lit = -fx / half
			} else {
				dx, dy := fx, (fy-17)*1.05
				r := 9 + 1.2*math.Sin(math.Atan2(dy, dx)*5)
				in = fy > 4 && math.Hypot(dx, dy) < r
				lit = (-dx - dy) / 12
			}
			if !in {
				continue
			}
			c := color.NRGBA{58, 118, 52, 255}
			if pine {
				c = color.NRGBA{44, 98, 62, 255}
			}
			d := int(lit*18) + int(hash(x, y)%15) - 7
			dst.SetNRGBA(ox+x, oy+y, shade(c, d))
		}
	}
}

func hills(dst *image.NRGBA, ox, oy int) {
	far, near := color.NRGBA{142, 160, 184, 255}, color.NRGBA{76, 112, 82, 255}
	for x := range field.HillsW {
		t := 2 * math.Pi * float64(x) / field.HillsW
		hill := 8 + 3*math.Sin(2*t) + 2*math.Sin(5*t+1) + math.Sin(11*t+2)
		trees := 2.5 + math.Sin(9*t) + float64(hash(x/2, 0)%3)
		for y := range field.HillsH {
			h := float64(field.HillsH - y)
			switch {
			case h <= trees:
				dst.SetNRGBA(ox+x, oy+y, shade(near, int(hash(x, y)%9)-4))
			case h <= hill:
				dst.SetNRGBA(ox+x, oy+y, far)
			}
		}
	}
}

func sheet() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, field.SheetW, field.SheetH))
	for k, m := range tanks {
		for f := range field.TankFrames {
			rel := float64(f) * 2 * math.Pi / field.TankFrames
			for part, boxes := range [][]box{m.hull(), m.turret()} {
				x, y := field.TankCell(k, part, f)
				render(img, x, y, field.TankW, field.TankH, field.TankAX, field.TankAY, boxes, rel)
			}
		}
	}
	for f := range field.HouseFrames {
		render(img, f*field.HouseW, field.HouseY, field.HouseW, field.HouseH, field.HouseAX, field.HouseAY, house(), float64(f)*math.Pi/field.HouseFrames)
	}
	tree(img, field.TreeX, field.HouseY, false)
	tree(img, field.TreeX+field.TreeW, field.HouseY, true)
	hills(img, 0, field.HillsY)
	return img
}

// sight is the gunsight's eyepiece: black outside a circle with a dark rim.
func sight() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, field.SightW, field.SightH))
	for y := range field.SightH {
		for x := range field.SightW {
			d := math.Hypot(float64(x)+.5-field.SightX, float64(y)+.5-field.SightY)
			switch {
			case d > field.SightRadius+2:
				img.SetNRGBA(x, y, color.NRGBA{8, 10, 8, 255})
			case d > field.SightRadius:
				img.SetNRGBA(x, y, color.NRGBA{40, 44, 40, 255})
			}
		}
	}
	return img
}

func main() {
	for name, img := range map[string]image.Image{"ground.png": ground(), "sprites.png": sheet(), "sight.png": sight()} {
		out, err := os.Create(name)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, img); err != nil {
			log.Fatal(err)
		}
		if err := out.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
