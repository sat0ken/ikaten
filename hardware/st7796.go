//go:build tinygo && (esp32s3 || esp32)

package hardware

import (
	"errors"
	"machine"
	"time"

	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/minijoy"
	"github.com/sat0ken/ikaten/internal/present"
)

// ST7796Config can be adjusted before SetHardware/RunGame. Width and Height
// are the panel's size after MADCTL.
type ST7796Config struct {
	SPI                                *machine.SPI
	Frequency                          uint32 // for drawing; the panel is set up at 10MHz
	SCK, MOSI, MISO, CS, DC, Backlight machine.Pin
	Width, Height                      int
	MADCTL                             byte
	Invert                             bool
	// Fractional scales the logical image by a fraction to fill the panel's
	// height: 240x240 becomes 320x320 on a 480x320 panel.
	Fractional bool
	// The mini joystick module (package internal/minijoy), or nil.
	Joystick        *machine.I2C
	JoySDA, JoySCL  machine.Pin
	JoyFrequency    uint32
	PrepareJoystick func()
	PrepareSPI      func() // runs before the SPI bus is configured
}

// ST7796 is a 4-wire SPI panel with the ST7796S controller. Pixels go
// straight to the panel; the engine's screen image is the only frame buffer.
type ST7796 struct {
	Config    ST7796Config
	presenter present.Presenter
	joystick  *minijoy.Device
	cmd       [1]byte
	window    [4]byte
}

// st7796Init is the ST7796S init sequence of TFT_eSPI (ST7796_Init.h), as the
// github.com/sat0ken/tinygo-cyd sends it on the ESP32-3248S035C.
var st7796Init = []struct {
	cmd  byte
	data string
	wait time.Duration
}{
	{0x11, "", 120 * time.Millisecond}, // sleep out
	{0xF0, "\xC3", 0},                  // enable extension command 2 part I
	{0xF0, "\x96", 0},                  // part II
	{0x3A, "\x55", 0},                  // 16-bit colour
	{0xB4, "\x01", 0},                  // 1-dot inversion
	{0xB6, "\x80\x02\x3B", 0},          // display function control
	{0xE8, "\x40\x8A\x00\x00\x29\x19\xA5\x33", 0},
	{0xC1, "\x06", 0},                      // power control 2
	{0xC2, "\xA7", 0},                      // power control 3
	{0xC5, "\x18", 120 * time.Millisecond}, // VCOM
	{0xE0, "\xF0\x09\x0B\x06\x04\x15\x2F\x54\x42\x3C\x17\x14\x18\x1B", 0},
	{0xE1, "\xE0\x09\x0B\x06\x04\x03\x2B\x43\x42\x3B\x16\x14\x17\x1B", 120 * time.Millisecond},
	{0xF0, "\x3C", 0}, // disable extension command 2 part I
	{0xF0, "\x69", 120 * time.Millisecond},
}

func (d *ST7796) Init() error {
	c := &d.Config
	if c.SPI == nil || c.Frequency == 0 || c.Width <= 0 || c.Height <= 0 {
		return errors.New("ikaten: SPI, frequency and size are required")
	}
	if c.PrepareSPI != nil {
		c.PrepareSPI()
	}
	spi := machine.SPIConfig{Frequency: 10_000_000, SCK: c.SCK, SDO: c.MOSI, SDI: c.MISO, Mode: 0}
	if err := c.SPI.Configure(spi); err != nil {
		return err
	}
	c.DC.Configure(machine.PinConfig{Mode: machine.PinOutput})
	c.CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	c.CS.High()
	// The board has no reset line: reset with the command.
	d.command(0x01, "")
	time.Sleep(120 * time.Millisecond)
	for _, s := range st7796Init {
		d.command(s.cmd, s.data)
		time.Sleep(s.wait)
	}
	d.command(0x36, string([]byte{c.MADCTL}))
	if c.Invert {
		d.command(0x21, "")
	} else {
		d.command(0x20, "")
	}
	d.command(0x29, "") // display on
	time.Sleep(20 * time.Millisecond)
	spi.Frequency = c.Frequency
	if err := c.SPI.Configure(spi); err != nil {
		return err
	}
	d.presenter = present.Presenter{Width: c.Width, Height: c.Height, Fractional: c.Fractional}
	if c.Backlight != machine.NoPin {
		c.Backlight.Configure(machine.PinConfig{Mode: machine.PinOutput})
		c.Backlight.High()
	}
	if c.Joystick != nil {
		if c.PrepareJoystick != nil {
			c.PrepareJoystick()
		}
		freq := c.JoyFrequency
		if freq == 0 {
			freq = 100_000
		}
		if err := c.Joystick.Configure(machine.I2CConfig{SDA: c.JoySDA, SCL: c.JoySCL, Frequency: freq}); err != nil {
			println("ikaten: joystick I2C:", err.Error())
		} else {
			d.joystick = minijoy.New(c.Joystick)
			if !d.joystick.Probe() {
				println("ikaten: no mini joystick module on I2C")
			}
		}
	}
	return nil
}

// command sends a command and its parameters with CS low.
func (d *ST7796) command(cmd byte, data string) {
	c := &d.Config
	c.CS.Low()
	d.send(cmd, nil)
	if len(data) > 0 {
		c.SPI.Tx([]byte(data), nil)
	}
	c.CS.High()
}

// send sends a command byte and then data, with CS already low.
func (d *ST7796) send(cmd byte, data []byte) {
	c := &d.Config
	c.DC.Low()
	d.cmd[0] = cmd
	c.SPI.Tx(d.cmd[:], nil)
	c.DC.High()
	if len(data) > 0 {
		c.SPI.Tx(data, nil)
	}
}

func (d *ST7796) Size() (int, int)                   { return d.Config.Width, d.Config.Height }
func (d *ST7796) Present(screen *ikaten.Image) error { return d.presenter.Present(screen, d) }
func (d *ST7796) KeyUpdate() error {
	if d.joystick != nil {
		d.joystick.Update()
	}
	return nil
}

// WriteRect sends big-endian RGB565 pixels to a rectangle of the panel.
func (d *ST7796) WriteRect(x, y, w, h int, data []byte) error {
	if w <= 0 || h <= 0 || len(data) < w*h*2 {
		return errors.New("ikaten: bad rectangle")
	}
	c := &d.Config
	c.CS.Low()
	x1, y1 := x+w-1, y+h-1
	d.window = [4]byte{byte(x >> 8), byte(x), byte(x1 >> 8), byte(x1)}
	d.send(0x2A, d.window[:]) // column address
	d.window = [4]byte{byte(y >> 8), byte(y), byte(y1 >> 8), byte(y1)}
	d.send(0x2B, d.window[:]) // row address
	d.send(0x2C, nil)         // memory write
	err := c.SPI.Tx(data[:w*h*2], nil)
	c.CS.High()
	return err
}
func (d *ST7796) PresentStats() present.Stats { return d.presenter.Stats }
