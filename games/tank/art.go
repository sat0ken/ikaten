package tank

import (
	"image"
	"image/color"

	ebiten "github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/tank/field"
)

var (
	// From the sheet, which stays in flash.
	imageGround *ebiten.Image
	imageSight  *ebiten.Image
	imageTank   [field.TankKinds][2][field.TankFrames]*ebiten.Image // kind, hull or turret, frame
	imageHouse  [field.HouseFrames]*ebiten.Image
	imageTree   [2]*ebiten.Image // round and pine
	imageHills  *ebiten.Image

	// Small images drawn here.
	imagePuff   [puffKinds]*ebiten.Image
	imageDigits [10]*ebiten.Image
	imageRadar  *ebiten.Image
	imageHull   *ebiten.Image // the player's hull on the radar
	imageMarker *ebiten.Image // over enemies in the wide view

	imageSky     = pixel(color.RGBA{96, 146, 214, 255})
	imageSkyLow  = pixel(color.RGBA{172, 200, 228, 255})
	imageGrass   = pixel(color.RGBA{96, 150, 62, 255})
	imageHaze    = pixel(color.RGBA{160, 186, 190, 255})
	colorReticle = color.RGBA{16, 16, 16, 255}
	colorCross   = color.RGBA{240, 240, 230, 255}
	imageReticle = pixel(colorReticle)
	imageCross   = pixel(colorCross)
	imageRed     = pixel(color.RGBA{240, 50, 40, 255})
	imageYellow  = pixel(color.RGBA{250, 210, 50, 255})
	imageGreen   = pixel(color.RGBA{80, 220, 100, 255})
	imageDim     = pixel(color.RGBA{70, 120, 80, 255})
	imageGray    = pixel(color.RGBA{150, 150, 150, 255})
	imageShade   = pixel(color.RGBA{10, 14, 10, 255})
	imageTracer  = pixel(color.RGBA{255, 226, 130, 255})
	// imageDot is a far tank, too small for its sprite.
	imageDot [field.TankKinds]*ebiten.Image
)

func pixel(c color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	return img
}

// digitGlyphs is a 3x5 font for numbers.
var digitGlyphs = [10][5]string{
	{"###", "#.#", "#.#", "#.#", "###"},
	{".#.", "##.", ".#.", ".#.", "###"},
	{"###", "..#", "###", "#..", "###"},
	{"###", "..#", "###", "..#", "###"},
	{"#.#", "#.#", "###", "..#", "..#"},
	{"###", "#..", "###", "..#", "###"},
	{"###", "#..", "###", "#.#", "###"},
	{"###", "..#", "..#", ".#.", ".#."},
	{"###", "#.#", "###", "#.#", "###"},
	{"###", "#.#", "###", "..#", "###"},
}

const puffSize = 9 // pixels across a puff image

// puffColors are the edge and core colors of each puff.
var puffColors = [puffKinds][2]color.RGBA{
	dust:  {{140, 118, 84, 255}, {170, 148, 110, 255}},
	smoke: {{58, 58, 60, 255}, {88, 88, 90, 255}},
	fire:  {{240, 110, 30, 255}, {255, 220, 90, 255}},
	spark: {{255, 200, 80, 255}, {255, 250, 200, 255}},
	flash: {{255, 170, 60, 255}, {255, 245, 190, 255}},
}

func cut(sheet *ebiten.Image, x, y, w, h int) *ebiten.Image {
	return sheet.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image)
}

func init() {
	var sheet *ebiten.Image
	imageGround, sheet, imageSight = loadAssets()
	// SubImage allocates, so the frames are cut once here.
	for k := range field.TankKinds {
		for part := range 2 {
			for f := range field.TankFrames {
				x, y := field.TankCell(k, part, f)
				imageTank[k][part][f] = cut(sheet, x, y, field.TankW, field.TankH)
			}
		}
	}
	for f := range imageHouse {
		imageHouse[f] = cut(sheet, f*field.HouseW, field.HouseY, field.HouseW, field.HouseH)
	}
	for i := range imageTree {
		imageTree[i] = cut(sheet, field.TreeX+i*field.TreeW, field.HouseY, field.TreeW, field.TreeH)
	}
	imageHills = cut(sheet, 0, field.HillsY, field.HillsW, field.HillsH)
	imageDot[light] = pixel(color.RGBA{150, 132, 90, 255})
	imageDot[medium] = pixel(color.RGBA{100, 104, 96, 255})
	imageDot[heavy] = pixel(color.RGBA{70, 74, 86, 255})

	// Puffs are rough discs, brighter in the middle.
	for k := range imagePuff {
		img := ebiten.NewImage(puffSize, puffSize)
		for y := range puffSize {
			for x := range puffSize {
				dx, dy := x-puffSize/2, y-puffSize/2
				d := dx*dx + dy*dy
				rough := (x*7+y*13)%5 == 0
				switch {
				case d <= 4:
					img.Set(x, y, puffColors[k][1])
				case d <= 17 && !(rough && d > 10):
					img.Set(x, y, puffColors[k][0])
				}
			}
		}
		imagePuff[k] = img
	}

	digits := ebiten.NewImage(4*len(digitGlyphs), 5)
	for d, glyph := range digitGlyphs {
		for y, row := range glyph {
			for x := range len(row) {
				if row[x] == '#' {
					digits.Set(d*4+x, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
	for d := range imageDigits {
		imageDigits[d] = cut(digits, d*4, 0, 3, 5)
	}

	imageRadar = ebiten.NewImage(radarSize, radarSize)
	for y := range radarSize {
		for x := range radarSize {
			dx, dy := x-radarSize/2, y-radarSize/2
			d := dx*dx + dy*dy
			r := radarSize / 2
			switch {
			case d > r*r:
			case d > (r-1)*(r-1):
				imageRadar.Set(x, y, color.RGBA{70, 150, 90, 255})
			case (dx == 0 || dy == 0) && (x+y)%2 == 0:
				imageRadar.Set(x, y, color.RGBA{30, 70, 40, 255})
			default:
				imageRadar.Set(x, y, color.RGBA{14, 34, 20, 255})
			}
		}
	}
	imageHull = ebiten.NewImage(3, 5)
	imageHull.Fill(color.RGBA{150, 200, 160, 255})
	imageMarker = ebiten.NewImage(5, 3)
	for y, row := range [3]string{"#####", ".###.", "..#.."} {
		for x := range len(row) {
			if row[x] == '#' {
				imageMarker.Set(x, y, color.RGBA{255, 70, 50, 255})
			}
		}
	}
}
