package render

import (
	"image"
)

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"
)

const (
	BorderTop = 1 << iota
	BorderRight
	BorderBottom
	BorderLeft
)

const (
	GradientHorz = iota
	GradientVert
)

const (
	GradientRegular = iota
	GradientReverse
)

const (
	DiagTopLeft = iota
	DiagTopRight
	DiagBottomLeft
	DiagBottomRight
)

type Image struct {
	*xgraphics.Image
}

func New(ximg *xgraphics.Image) *Image {
	return &Image{ximg}
}

func NewSolid(X *xgbutil.XUtil, bgColor Color, width, height int) *Image {
	img := New(xgraphics.New(X, image.Rect(0, 0, width, height)))

	r, g, b := bgColor.RGB8()
	img.ForExp(func(x, y int) (uint8, uint8, uint8, uint8) {
		return r, g, b, 0xff
	})
	return img
}

func NewBorder(X *xgbutil.XUtil, borderType int, borderColor,
	bgColor Color, width, height, gradientType, gradientDir int) *Image {

	img := New(xgraphics.New(X, image.Rect(0, 0, width, height)))

	// bgClr could be a gradient!
	if bgColor.IsGradient() {
		img.Gradient(gradientType, gradientDir, bgColor)
	} else {
		r, g, b := bgColor.RGB8()
		img.ForExp(func(x, y int) (uint8, uint8, uint8, uint8) {
			return r, g, b, 0xff
		})
	}

	img.ThinBorder(borderType, borderColor)
	return img
}

func NewCorner(X *xgbutil.XUtil, borderType int, borderColor,
	bgColor Color, width, height, diagonal int) *Image {

	// If bgColor isn't a gradient, then we can cheat
	if !bgColor.IsGradient() {
		return NewBorder(X, borderType, borderColor, bgColor,
			width, height, 0, 0)
	}

	img := New(xgraphics.New(X, image.Rect(0, 0, width, height)))

	// aliases for convenience
	vert, horz := GradientVert, GradientHorz
	reg, rev := GradientRegular, GradientReverse

	// for Top Left to Bottom Right diagonals
	belowTLDiag := func(x, y int) bool { return y > x }
	aboveTLDiag := func(x, y int) bool { return y <= x }

	// for Bottom Left to Top Right diagonals
	belowBLDiag := func(x, y int) bool { return y > (width - x) }
	aboveBLDiag := func(x, y int) bool { return y <= (width - x) }

	switch diagonal {
	case DiagBottomLeft:
		img.GradientFunc(horz, reg, bgColor, aboveBLDiag)
		img.GradientFunc(vert, rev, bgColor, belowBLDiag)
	case DiagTopRight:
		img.GradientFunc(vert, reg, bgColor, aboveBLDiag)
		img.GradientFunc(horz, rev, bgColor, belowBLDiag)
	case DiagBottomRight:
		img.GradientFunc(horz, rev, bgColor, aboveTLDiag)
		img.GradientFunc(vert, rev, bgColor, belowTLDiag)
	default: // DiagTopLeft:
		img.GradientFunc(vert, reg, bgColor, aboveTLDiag)
		img.GradientFunc(horz, reg, bgColor, belowTLDiag)
	}

	img.ThinBorder(borderType, borderColor)
	return img
}

// XXX: Optimize.
func (img *Image) ThinBorder(borderType int, borderColor Color) {
	borderClr := borderColor.ImageColor()
	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	// Now go through and add a "thin border."
	// It's always one pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if borderType&BorderTop > 0 && y == 0 {
				img.Set(x, y, borderClr)
			}
			if borderType&BorderRight > 0 && x == width-1 {
				img.Set(x, y, borderClr)
			}
			if borderType&BorderBottom > 0 && y == height-1 {
				img.Set(x, y, borderClr)
			}
			if borderType&BorderLeft > 0 && x == 0 {
				img.Set(x, y, borderClr)
			}
		}
	}
}

func (img *Image) Gradient(gradientType, gradientDir int, clr Color) {
	img.GradientFunc(gradientType, gradientDir, clr,
		func(x, y int) bool { return true })
}

// XXX: Optimize.
func (img *Image) GradientFunc(gradientType, gradientDir int, clr Color,
	pixel func(x, y int) bool) {

	width, height := img.Bounds().Dx(), img.Bounds().Dy()

	if gradientType == GradientVert {
		steps := clr.Steps(height)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				if pixel(x, y) {
					if gradientDir == GradientReverse {
						img.Set(x, y, steps[height-y-1])
					} else {
						img.Set(x, y, steps[y])
					}
				}
			}
		}
	} else {
		steps := clr.Steps(width)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				if pixel(x, y) {
					if gradientDir == GradientReverse {
						img.Set(x, y, steps[width-x-1])
					} else {
						img.Set(x, y, steps[x])
					}
				}
			}
		}
	}
}
