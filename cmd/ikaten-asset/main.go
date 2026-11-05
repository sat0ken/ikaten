// ikaten-asset converts a PNG to an embeddable, read-only IKA1 sprite.
package main

import (
	"flag"
	"fmt"
	"github.com/sat0ken/ikaten/internal/asset"
	"image/png"
	"os"
)

func run() error {
	out := flag.String("o", "", "output .ika file")
	flag.Parse()
	if flag.NArg() != 1 || *out == "" {
		return fmt.Errorf("usage: ikaten-asset -o sprite.ika sprite.png")
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return err
	}
	img, err := png.Decode(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	data, err := asset.Encode(img)
	if err != nil {
		return err
	}
	return os.WriteFile(*out, data, 0644)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
