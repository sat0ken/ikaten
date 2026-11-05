// Package present scales logical RGB565 rows and sends only changed rows.
package present

import (
	"errors"
	"image"
	"unsafe"
)

type Source interface {
	Bounds() image.Rectangle
	ReadRGB565Row(dst []byte, y int)
}

// DirectSource exposes pixels already in ST7789 byte order, so rows can be
// hashed and sent without copying. pixels starts at the top-left of Bounds.
type DirectSource interface {
	RGB565Rows() (pixels []uint16, stride int, ok bool)
}

// Writer must consume data synchronously, without retaining it.
type Writer interface {
	WriteRect(x, y, width, height int, data []byte) error
}
type Stats struct{ ChangedRows, Bytes int }
type Presenter struct {
	Width, Height                              int
	logicalWidth, logicalHeight, scale, ox, oy int
	words                                      []uint16
	row                                        []byte // words viewed as bytes
	hashes                                     []uint32
	valid                                      []bool
	cleared                                    bool
	Stats                                      Stats
}

func (p *Presenter) configure(w, h int) error {
	if w <= 0 || h <= 0 || p.Width <= 0 || p.Height <= 0 || w > p.Width || h > p.Height {
		return errors.New("ikaten: logical display does not fit")
	}
	if p.logicalWidth == w && p.logicalHeight == h {
		return nil
	}
	p.logicalWidth, p.logicalHeight = w, h
	p.scale = min(p.Width/w, p.Height/h)
	p.ox, p.oy = (p.Width-w*p.scale)/2, (p.Height-h*p.scale)/2
	p.words = make([]uint16, max(p.Width, w*p.scale*min(p.scale, 2)))
	p.row = wordBytes(p.words)
	p.hashes = make([]uint32, h)
	p.valid = make([]bool, h)
	p.cleared = false
	return nil
}

// wordBytes views pixels as bytes. Stored pixels are byte-swapped for the
// little-endian targets, so this memory order is the ST7789 wire order.
func wordBytes(words []uint16) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(words))), 2*len(words))
}

// hashRow is FNV-1a over 16-bit pixels.
func hashRow(row []uint16) uint32 {
	hash := uint32(2166136261)
	for _, v := range row {
		hash = (hash ^ uint32(v)) * 16777619
	}
	return hash
}

func (p *Presenter) Present(src Source, writer Writer) error {
	b := src.Bounds()
	if err := p.configure(b.Dx(), b.Dy()); err != nil {
		return err
	}
	p.Stats = Stats{}
	if !p.cleared {
		clear(p.row)
		for y := 0; y < p.Height; y++ {
			if err := writer.WriteRect(0, y, p.Width, 1, p.row[:p.Width*2]); err != nil {
				return err
			}
			p.Stats.Bytes += p.Width * 2
		}
		p.cleared = true
	}
	var pixels []uint16
	stride, direct := 0, false
	if d, ok := src.(DirectSource); ok {
		pixels, stride, direct = d.RGB565Rows()
	}
	if direct && p.scale == 1 && stride == p.logicalWidth {
		return p.presentBands(pixels, writer)
	}
	return p.presentRows(src, pixels, stride, direct, writer)
}

// presentBands sends runs of changed rows from contiguous pixels in one write.
func (p *Presenter) presentBands(pixels []uint16, writer Writer) error {
	w, start := p.logicalWidth, -1
	for y := 0; y <= p.logicalHeight; y++ {
		if y < p.logicalHeight {
			hash := hashRow(pixels[y*w : (y+1)*w])
			if !p.valid[y] || p.hashes[y] != hash {
				p.hashes[y] = hash
				p.valid[y] = false
				if start < 0 {
					start = y
				}
				continue
			}
		}
		if start >= 0 {
			// Rows stay invalid until the write succeeds, so failures are retried.
			if err := writer.WriteRect(p.ox, p.oy+start, w, y-start, wordBytes(pixels[start*w:y*w])); err != nil {
				return err
			}
			for n := start; n < y; n++ {
				p.valid[n] = true
			}
			p.Stats.ChangedRows += y - start
			p.Stats.Bytes += (y - start) * w * 2
			start = -1
		}
	}
	return nil
}

// presentRows handles scaling and sources that must be copied, one row at a time.
func (p *Presenter) presentRows(src Source, pixels []uint16, stride int, direct bool, writer Writer) error {
	b := src.Bounds()
	w, s := p.logicalWidth, p.scale
	rowWords := w * s
	for y := 0; y < p.logicalHeight; y++ {
		var row []uint16
		if direct {
			row = pixels[y*stride : y*stride+w]
		} else {
			src.ReadRGB565Row(p.row[:w*2], b.Min.Y+y)
			row = p.words[:w]
		}
		hash := hashRow(row)
		if p.valid[y] && p.hashes[y] == hash {
			continue
		}
		out := p.words[:rowWords]
		// Expanding backwards also works when row is the start of out.
		for x := w - 1; x >= 0; x-- {
			v := row[x]
			for j := s - 1; j >= 0; j-- {
				out[x*s+j] = v
			}
		}
		for j := 1; j < min(s, 2); j++ {
			copy(p.words[j*rowWords:(j+1)*rowWords], out)
		}
		for offset := 0; offset < s; offset += 2 {
			h := min(2, s-offset)
			if err := writer.WriteRect(p.ox, p.oy+y*s+offset, rowWords, h, p.row[:rowWords*2*h]); err != nil {
				return err
			}
		}
		// Commit a hash only after a successful write, so failed rows are retried.
		p.hashes[y] = hash
		p.valid[y] = true
		p.Stats.ChangedRows++
		p.Stats.Bytes += rowWords * 2 * s
	}
	return nil
}
func (p *Presenter) Invalidate() { clear(p.valid); p.cleared = false }
