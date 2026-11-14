package dungeon

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
)

// sprites holds a game's images. A Game makes them in NewGame and drops them
// with it, so that the launcher keeps only the running game's in RAM.
type sprites struct {
	weapon, flash *ebiten.Image
	// Textures are cut into one-pixel columns once; SubImage allocates.
	wallColumns  [3][texSize]*ebiten.Image
	slimeColumns [texSize]*ebiten.Image
}

// Solid colors, made once: one-pixel images scaled to rectangles.
var (
	imageBlack, imageWhite, imageRed *ebiten.Image
	imageCeiling, imageFloor         [bands]*ebiten.Image
)

// The ceiling and the floor are bands that darken toward the horizon.
const bands = 6

func solid(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

func init() {
	imageBlack = solid(color.RGBA{0, 0, 0, 255})
	imageWhite = solid(color.RGBA{255, 255, 255, 255})
	imageRed = solid(color.RGBA{230, 30, 30, 255})
	for i := range bands {
		// i counts from the top of the view's half, away from the horizon.
		f := uint8((bands - 1 - i) * 6)
		imageCeiling[i] = solid(color.RGBA{44 - f, 44 - f, 58 - f, 255})
		imageFloor[i] = solid(color.RGBA{58 + f, 46 + f, 34 + f/2, 255})
	}
}

// hash gives repeatable texture noise without math/rand.
func hash(x, y, seed int) int {
	h := uint32(x*374761393 + y*668265263 + seed*1442695041)
	h = (h ^ h>>13) * 1274126177
	return int((h ^ h>>16) & 0xff)
}

// brick is red brick in courses of eight pixels with lit tops and shaded
// bottoms, and mortar between.
func brick(x, y int) color.RGBA {
	row := y / 8
	bx := (x + row%2*8) % 16
	switch {
	case y%8 == 7 || bx == 15:
		return color.RGBA{86, 78, 72, 255} // mortar
	case y%8 == 0 || bx == 0:
		return color.RGBA{186, 92, 64, 255} // the lit edge
	case y%8 == 6:
		return color.RGBA{104, 40, 28, 255} // the shade
	}
	n := uint8(hash(x/16, row, 1) / 10)
	d := uint8(hash(x, y, 9) % 12)
	return color.RGBA{140 + n - d, 54 + n/2, 38, 255}
}

// stone is grey blocks with a bevel and grain.
func stone(x, y int) color.RGBA {
	n := uint8(hash(x, y, 2) / 8)
	v := 88 + n
	if (x/8+y/8)%2 == 0 {
		v += 12
	}
	switch {
	case x%8 == 0 || y%8 == 0:
		v += 30
	case x%8 == 7 || y%8 == 7:
		v -= 30
	}
	return color.RGBA{v, v, v + 10, 255}
}

// metal is blue panels with seams, a highlight and rivets.
func metal(x, y int) color.RGBA {
	px, py := x%16, y%16
	switch {
	case px == 0 || py == 0:
		return color.RGBA{30, 40, 52, 255} // panel seams
	case px == 1 || py == 1:
		return color.RGBA{120, 146, 172, 255}
	case (px == 3 || px == 13) && (py == 3 || py == 13):
		return color.RGBA{210, 220, 230, 255} // rivets
	case (px == 4 || px == 14) && (py == 4 || py == 14):
		return color.RGBA{40, 52, 64, 255}
	}
	v := uint8(hash(x/2, y, 3) % 10)
	return color.RGBA{76 + v, 100 + v, 124 + v, 255}
}

func texture(f func(x, y int) color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(texSize, texSize)
	// Filling first marks the texture opaque, which draws faster.
	img.Fill(color.RGBA{0, 0, 0, 255})
	for y := range texSize {
		for x := range texSize {
			img.Set(x, y, f(x, y))
		}
	}
	return img
}

// slimeImage is an original green blob with a shine, two eyes and a mouth;
// the rest is transparent.
func slimeImage() *ebiten.Image {
	img := ebiten.NewImage(texSize, texSize)
	for y := range texSize {
		slimeRow(img, y)
	}
	return img
}

func slimeRow(img *ebiten.Image, y int) {
	for x := range texSize {
		// A dome over a flat base, in sixteenths: an ellipse above row 22,
		// a band below that bulges a little.
		dx, dy := (2*x+1-texSize)*8/14, (y-22)*16/18
		if y < 22 && dx*dx+dy*dy > 256 || y >= 22 && dx*dx > 256+16*(y-22) || y > 29 {
			continue
		}
		c := color.RGBA{64, 196, 86, 255}
		switch {
		case y >= 26:
			c = color.RGBA{36, 130, 56, 255}
		case dx*dx+dy*dy > 200 && (dx > 0 || y > 18):
			c = color.RGBA{44, 156, 66, 255} // the rim in shade
		case (x-10)*(x-10)+(y-8)*(y-8) < 6:
			c = color.RGBA{180, 250, 190, 255} // the shine
		}
		for _, ex := range [2]int{11, 20} {
			switch d := (x-ex)*(x-ex) + (y-15)*(y-15); {
			case d <= 2:
				c = color.RGBA{20, 20, 30, 255}
			case d <= 7:
				c = color.RGBA{250, 250, 250, 255}
			}
		}
		if y == 21 && x >= 13 && x <= 18 || y == 20 && (x == 12 || x == 19) {
			c = color.RGBA{24, 80, 40, 255} // the grin
		}
		img.Set(x, y, c)
	}
}

// weaponImage is the pistol held at the bottom of the view.
func weaponImage() *ebiten.Image {
	img := ebiten.NewImage(48, 40)
	for y := range 40 {
		for x := range 48 {
			if c, ok := weaponPixel(x, y); ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

func weaponPixel(x, y int) (color.RGBA, bool) {
	switch {
	case x >= 19 && x < 29 && y < 24: // the barrel, lit on its left
		if x == 19 || x == 28 {
			return color.RGBA{30, 30, 36, 255}, true
		}
		v := uint8(150 - (x-20)*12)
		if y < 2 {
			v = 60 // the muzzle
		}
		return color.RGBA{v, v, v + 12, 255}, true
	case y >= 20 && y < 30 && x >= 8 && x < 40: // the slide and frame
		if y == 20 || x == 8 || x == 39 {
			return color.RGBA{30, 30, 36, 255}, true
		}
		return color.RGBA{70, 70, 80, 255}, true
	case y >= 30 && x >= 14 && x < 34: // the grip
		if (x+y)%4 == 0 {
			return color.RGBA{70, 52, 40, 255}, true
		}
		return color.RGBA{96, 70, 50, 255}, true
	case y >= 30 && (x >= 6 && x < 14 || x >= 34 && x < 42): // the hands
		return color.RGBA{220, 170, 130, 255}, true
	}
	return color.RGBA{}, false
}

func flashImage() *ebiten.Image {
	img := ebiten.NewImage(24, 24)
	for y := range 24 {
		for x := range 24 {
			d := (x-12)*(x-12) + (y-12)*(y-12)
			switch {
			case d < 20:
				img.Set(x, y, color.RGBA{255, 252, 220, 255})
			case d < 60:
				img.Set(x, y, color.RGBA{255, 210, 80, 255})
			case d < 130 && (x+y)%2 == 0:
				img.Set(x, y, color.RGBA{255, 140, 30, 255})
			}
		}
	}
	return img
}

func newSprites() *sprites {
	a := &sprites{weapon: weaponImage(), flash: flashImage()}
	for i, f := range [...]func(x, y int) color.RGBA{brick, stone, metal} {
		tex := texture(f)
		for x := range texSize {
			a.wallColumns[i][x] = tex.SubImage(image.Rect(x, 0, x+1, texSize)).(*ebiten.Image)
		}
	}
	tex := slimeImage()
	for x := range texSize {
		a.slimeColumns[x] = tex.SubImage(image.Rect(x, 0, x+1, texSize)).(*ebiten.Image)
	}
	return a
}
