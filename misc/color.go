package misc

import (
	"image/color"
)

// ColorFromInt takes a hex number in the format of '0xrrggbb' and transforms 
// it to an RGBA color.
func ColorFromInt(clr int) color.RGBA {
	r, g, b := RGBFromInt(clr)
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(255)}
}

// IntFromColor takes a color and returns an RGB integer (i.e., 0x3366ff).
func IntFromColor(clr color.Color) int {
	r, g, b, _ := clr.RGBA()
	return IntFromRGB(int(r>>8), int(g>>8), int(b>>8))
}

// RGBFromInt returns the R, G and B portions of an integer.
func RGBFromInt(clr int) (r int, g int, b int) {
	r = clr >> 16
	g = (clr & 0x00ff00) >> 8
	b = clr & 0x0000ff
	return
}

// IntFromRGB returns an integer given R, G and B integers.
func IntFromRGB(r, g, b int) int {
	return (r << 16) + (g << 8) + b
}
