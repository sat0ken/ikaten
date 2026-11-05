//go:build tinygo && esp32s3 && ikaten_badge

package hardware

import (
	"github.com/sat0ken/ikaten"
	"machine"
	"time"
	"tinygo.org/x/drivers/st7789"
)

const LogicalWidth, LogicalHeight = 240, 240

// TinyGo Conference 2026 badge. Pin facts are from the user's badge project:
// firmware/badge/badge.go and display.go. No runtime dependency on that project.
var Device = &ST7789{Config: ST7789Config{
	SPI: machine.SPI0, Frequency: 80_000_000,
	SCK: machine.GPIO12, MOSI: machine.GPIO11, MISO: machine.NoPin,
	CS: machine.GPIO10, DC: machine.GPIO5, Reset: machine.GPIO4, Backlight: machine.NoPin,
	Panel:    st7789.Config{Width: 240, Height: 240, Rotation: st7789.NO_ROTATION, FrameRate: st7789.FRAMERATE_60, VSyncLines: st7789.MAX_VSYNC_SCANLINES},
	Buttons:  []Button{{Pin: machine.GPIO13, Key: ikaten.KeyZ}, {Pin: machine.GPIO14, Key: ikaten.KeyX}, {Pin: machine.GPIO15, Key: ikaten.KeyEnter}},
	Debounce: 5 * time.Millisecond, InputInit: initBadgeInput, PollInput: pollBadgeInput,
}}

var badgeX = machine.ADC{Pin: machine.GPIO7}
var badgeY = machine.ADC{Pin: machine.GPIO6}
var centerX, centerY, xDirection, yDirection int

func initBadgeInput() error {
	machine.InitADC()
	if err := badgeX.Configure(machine.ADCConfig{}); err != nil {
		return err
	}
	if err := badgeY.Configure(machine.ADCConfig{}); err != nil {
		return err
	}
	// Hold the stick neutral during the first 80 ms after display initialization.
	sumX, sumY := 0, 0
	for n := 0; n < 16; n++ {
		sumX += int(badgeX.Get())
		sumY += int(badgeY.Get())
		time.Sleep(5 * time.Millisecond)
	}
	centerX, centerY = sumX/16, sumY/16
	xDirection, yDirection = 0, 0
	return nil
}
func pollBadgeInput() {
	// This board's X voltage falls to the right; positive Y means up.
	xDirection = axisDirection(centerX-int(badgeX.Get()), xDirection, 6000, 4000)
	yDirection = axisDirection(int(badgeY.Get())-centerY, yDirection, 6000, 4000)
	ikaten.SetKeyState(ikaten.KeyArrowLeft, xDirection < 0)
	ikaten.SetKeyState(ikaten.KeyArrowRight, xDirection > 0)
	ikaten.SetKeyState(ikaten.KeyArrowDown, yDirection < 0)
	ikaten.SetKeyState(ikaten.KeyArrowUp, yDirection > 0)
}
