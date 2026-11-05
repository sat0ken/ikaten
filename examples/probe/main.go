//go:build tinygo && (esp32s3 || esp32 || rp2040)

// probe measures raw SPI throughput, row-command overhead, and sprite copying.
// It prints CSV to serial; no measurements are fabricated by a host build.
package main

import (
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/games/assets"
	"github.com/sat0ken/ikaten/hardware"
	"image/color"
	"runtime"
	"time"
)

func must(err error) {
	if err != nil {
		println(err.Error())
		for {
			time.Sleep(time.Second)
		}
	}
}
func main() {
	time.Sleep(2 * time.Second)
	d := hardware.Device
	must(d.Init())
	w, h := d.Size()
	// Repeating a 4-row strip avoids allocating a physical framebuffer on ESP32.
	rows := 4
	for h%rows != 0 {
		rows--
	}
	strip := make([]byte, w*rows*2)
	for p := 0; p < len(strip); p += 2 {
		strip[p] = 0xf8
	}
	println("kind,requested_hz,width,height,band_rows,iterations,total_us,bytes")
	for _, hz := range frequencies {
		d.Config.Frequency = hz
		must(d.Init())
		for _, band := range []int{1, rows} {
			const frames = 10
			start := time.Now()
			for frame := 0; frame < frames; frame++ {
				for y := 0; y < h; y += band {
					must(d.WriteRect(0, y, w, band, strip[:w*band*2]))
				}
			}
			println("spi", ",", hz, ",", w, ",", h, ",", band, ",", frames, ",", time.Since(start).Microseconds(), ",", w*h*2*frames)
			time.Sleep(500 * time.Millisecond)
		}
	}
	flash, err := assets.NewSprite()
	must(err)
	src := ikaten.NewImage(16, 16)
	src.DrawImage(flash, nil)
	flashOpaque, err := assets.NewOpaqueSprite()
	must(err)
	dst := ikaten.NewImage(hardware.LogicalWidth, hardware.LogicalHeight)
	// A filled destination matches the game loop, which draws over a background.
	dst.Fill(color.RGBA{A: 255})
	var op ikaten.DrawImageOptions
	op.GeoM.Translate(10, 10)
	start := time.Now()
	const copies = 10000
	for n := 0; n < copies; n++ {
		dst.DrawImage(src, &op)
	}
	println("sprite_ram,0,16,16,0,", copies, ",", time.Since(start).Microseconds(), ",0")
	opaque := ikaten.NewImage(16, 16)
	opaque.Fill(color.RGBA{R: 255, A: 255})
	start = time.Now()
	for n := 0; n < copies; n++ {
		dst.DrawImage(opaque, &op)
	}
	println("sprite_ram_opaque,0,16,16,0,", copies, ",", time.Since(start).Microseconds(), ",0")
	start = time.Now()
	for n := 0; n < copies; n++ {
		dst.DrawImage(flash, &op)
	}
	println("sprite_flash,0,16,16,0,", copies, ",", time.Since(start).Microseconds(), ",0")
	start = time.Now()
	for n := 0; n < copies; n++ {
		dst.DrawImage(flashOpaque, &op)
	}
	println("sprite_flash_opaque,0,16,16,0,", copies, ",", time.Since(start).Microseconds(), ",0")
	// The second Present of an unchanged image costs row reading and hashing only.
	must(d.Present(dst))
	start = time.Now()
	const presents = 10
	for n := 0; n < presents; n++ {
		must(d.Present(dst))
	}
	println("present_unchanged,0,", hardware.LogicalWidth, ",", hardware.LogicalHeight, ",0,", presents, ",", time.Since(start).Microseconds(), ",0")
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	println("heap_alloc_bytes,", stats.HeapAlloc)
	println("probe complete; check panel color/stability at every clock")
	for {
		time.Sleep(time.Second)
	}
}
