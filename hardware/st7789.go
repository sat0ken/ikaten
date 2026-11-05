//go:build tinygo && (esp32s3 || esp32 || rp2040)

package hardware

import (
	"errors"
	"github.com/sat0ken/ikaten"
	"github.com/sat0ken/ikaten/internal/present"
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

type Button struct {
	Pin        machine.Pin
	Key        ikaten.Key
	ActiveHigh bool
}

// ST7789Config can be adjusted before SetHardware/RunGame. Panel dimensions,
// rotation and offsets use the driver's convention. SPI frequency is a request;
// the machine implementation chooses the achievable clock.
type ST7789Config struct {
	SPI                                       *machine.SPI
	Frequency                                 uint32
	SCK, MOSI, MISO, CS, DC, Reset, Backlight machine.Pin
	Panel                                     st7789.Config
	Buttons                                   []Button
	Debounce                                  time.Duration
	InputInit                                 func() error
	PollInput                                 func()
}
type ST7789 struct {
	Config    ST7789Config
	display   st7789.Device
	presenter present.Presenter
	buttons   []debouncer
}

func (d *ST7789) Init() error {
	c := &d.Config
	if c.SPI == nil || c.Frequency == 0 {
		return errors.New("ikaten: SPI and frequency are required")
	}
	if err := c.SPI.Configure(machine.SPIConfig{Frequency: c.Frequency, SCK: c.SCK, SDO: c.MOSI, SDI: c.MISO, Mode: 0}); err != nil {
		return err
	}
	d.display = st7789.New(c.SPI, c.Reset, c.DC, c.CS, c.Backlight)
	d.display.Configure(c.Panel)
	w, h := d.display.Size()
	d.presenter = present.Presenter{Width: int(w), Height: int(h)}
	d.buttons = make([]debouncer, len(c.Buttons))
	for _, b := range c.Buttons {
		if b.Pin != machine.NoPin {
			mode := machine.PinInputPullup
			if b.ActiveHigh {
				mode = machine.PinInputPulldown
			}
			b.Pin.Configure(machine.PinConfig{Mode: mode})
		}
	}
	if c.InputInit != nil {
		return c.InputInit()
	}
	return nil
}
func (d *ST7789) Size() (int, int)                   { w, h := d.display.Size(); return int(w), int(h) }
func (d *ST7789) Present(screen *ikaten.Image) error { return d.presenter.Present(screen, d) }
func (d *ST7789) KeyUpdate() error {
	now := time.Now()
	for n, b := range d.Config.Buttons {
		pressed := false
		if b.Pin != machine.NoPin {
			pressed = b.Pin.Get() == b.ActiveHigh
		}
		ikaten.SetKeyState(b.Key, d.buttons[n].update(pressed, now, d.Config.Debounce))
	}
	if d.Config.PollInput != nil {
		d.Config.PollInput()
	}
	return nil
}
func (d *ST7789) WriteRect(x, y, w, h int, data []byte) error {
	// v0.37.0 supports direct, synchronous big-endian row writes. It internally
	// ignores SPI Tx errors; see docs/hardware.md for this driver limitation.
	return d.display.DrawRGBBitmap8(int16(x), int16(y), data, int16(w), int16(h))
}
func (d *ST7789) PresentStats() present.Stats { return d.presenter.Stats }
