package render

import (
	"fmt"

	"image/color"
)

type Color struct {
	start, end int
}

var NoColor = Color{}

func NewColor(clr int) Color {
	return Color{start: clr, end: -1}
}

func NewImageColor(clr color.Color) Color {
	return NewColor(intFromImageColor(clr))
}

func NewGradient(start, end int) Color {
	return Color{start: start, end: end}
}

func NewImageGradient(startClr, endClr color.Color) Color {
	return NewGradient(intFromImageColor(startClr), intFromImageColor(endClr))
}

func (c Color) String() string {
	if c.IsGradient() {
		return fmt.Sprintf("(%x, %x)", c.start, c.end)
	}
	return fmt.Sprintf("%x", c.start)
}

func (c *Color) ColorSet(clr int) {
	c.start = clr
}

func (c *Color) GradientSet(start, end int) {
	c.start, c.end = start, end
}

func (c Color) Int() int {
	return c.start
}

func (c Color) Uint32() uint32 {
	return uint32(c.start)
}

func (c Color) ImageColor() color.RGBA {
	return imageColorFromInt(c.start)
}

func (c Color) RGB() (r, g, b int) {
	return rgbFromInt(c.start)
}

func (c Color) RGB8() (r, g, b uint8) {
	r32, g32, b32 := c.RGB()
	r, g, b = uint8(r32), uint8(g32), uint8(b32)
	return
}

// IsGradient returns whether the theme color is a gradient or not.
// Basically, a themeColor can either be a regular color (like an int)
// or a gradient when both 'start' and 'end' have valid color values.
func (c Color) IsGradient() bool {
	return c.start >= 0 && c.start <= 0xffffff &&
		c.end >= 0 && c.end <= 0xffffff
}

// steps returns a slice of colors corresponding to the gradient
// of colors from start to end. The size is determined by the 'size' parameter.
// The first and last colors in the slice are guaranteed to be
// c.start and c.end. (Unless the size is 1, in which case, the first and
// only color in the slice is c.start.)
// XXX: Optimize.
func (c Color) Steps(size int) []color.RGBA {
	// naughty
	if !c.IsGradient() {
		stps := make([]color.RGBA, size)
		for i := 0; i < size; i++ {
			stps[i] = imageColorFromInt(c.start)
		}
	}

	// yikes
	if size == 0 || size == 1 {
		return []color.RGBA{imageColorFromInt(c.start)}
	}

	stps := make([]color.RGBA, size)
	stps[0] = imageColorFromInt(c.start)
	stps[size-1] = imageColorFromInt(c.end)

	// no more?
	if size == 2 {
		return stps
	}

	sr, sg, sb := rgbFromInt(c.start)
	er, eg, eb := rgbFromInt(c.end)

	rinc := float64(er-sr) / float64(size)
	ginc := float64(eg-sg) / float64(size)
	binc := float64(eb-sb) / float64(size)

	doInc := func(inc float64, start, index int) int {
		return int(float64(start) + inc*float64(index))
	}

	var nr, ng, nb int
	for i := 1; i < size-1; i++ {
		nr = doInc(rinc, sr, i)
		ng = doInc(ginc, sg, i)
		nb = doInc(binc, sb, i)

		stps[i] = imageColorFromInt(intFromRgb(nr, ng, nb))
	}

	return stps
}

func intFromImageColor(clr color.Color) int {
	r, g, b, _ := clr.RGBA()
	return intFromRgb(int(r>>8), int(g>>8), int(b>>8))
}

func imageColorFromInt(clr int) color.RGBA {
	r, g, b := rgbFromInt(clr)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func intFromRgb(r, g, b int) int {
	return (r << 16) + (g << 8) + b
}

func rgbFromInt(clr int) (r, g, b int) {
	r = clr >> 16
	g = (clr & 0x00ff00) >> 8
	b = clr & 0x0000ff
	return
}
