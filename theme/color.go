package theme

import (
	"image/color"

	"github.com/BurntSushi/wingo/misc"
)


type Color struct {
	Start, End int
}

func NewColor(clr int) Color {
	return Color{Start: clr, End: -1}
}

func NewGradient(start, end int) Color {
	return Color{Start: start, End: end}
}

// IsGradient returns whether the theme color is a gradient or not.
// Basically, a themeColor can either be a regular color (like an int)
// or a gradient when both 'start' and 'end' have valid color values.
func (tc Color) IsGradient() bool {
	return tc.Start >= 0 && tc.Start <= 0xffffff &&
		tc.End >= 0 && tc.End <= 0xffffff
}

// steps returns a slice of colors corresponding to the gradient
// of colors from start to end. The size is determined by the 'size' parameter.
// The first and last colors in the slice are guaranteed to be
// tc.start and tc.end. (Unless the size is 1, in which case, the first and
// only color in the slice is tc.start.)
// XXX: Optimize.
func (tc Color) Steps(size int) []color.RGBA {
	// naughty
	if !tc.IsGradient() {
		stps := make([]color.RGBA, size)
		for i := 0; i < size; i++ {
			stps[i] = misc.ColorFromInt(tc.Start)
		}
	}

	// yikes
	if size == 0 || size == 1 {
		return []color.RGBA{misc.ColorFromInt(tc.Start)}
	}

	stps := make([]color.RGBA, size)
	stps[0] = misc.ColorFromInt(tc.Start)
	stps[size-1] = misc.ColorFromInt(tc.End)

	// no more?
	if size == 2 {
		return stps
	}

	sr, sg, sb := misc.RGBFromInt(tc.Start)
	er, eg, eb := misc.RGBFromInt(tc.End)

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

		stps[i] = misc.ColorFromInt(misc.IntFromRGB(nr, ng, nb))
	}

	return stps
}

