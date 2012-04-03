package main

import (
    "image"
    "image/color"
)

import (
    "github.com/BurntSushi/xgbutil/xgraphics"
)

const (
    renderBorderTop = 1 << iota
    renderBorderRight
    renderBorderBottom
    renderBorderLeft
)

const (
    renderGradientHorz = iota
    renderGradientVert
)

const (
    renderGradientRegular = iota
    renderGradientReverse
)

const (
    renderDiagTopLeft = iota
    renderDiagTopRight
    renderDiagBottomLeft
    renderDiagBottomRight
)

// ColorFromInt takes a hex number in the format of '0xrrggbb' and transforms 
// it to an RGBA color.
func ColorFromInt(clr int) color.RGBA {
    r, g, b := RGBFromInt(clr)
    return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(255)}
}

// RGBFromInt returns the R, G and B portions of an integer.
func RGBFromInt(clr int) (r int, g int, b int) {
    r = clr >> 16
    g = (clr - (r << 16)) >> 8
    b = clr - ((clr >> 8) << 8)
    return
}

// IntFromRGB returns an integer given R, G and B integers.
func IntFromRGB(r, g, b int) int {
    return (r << 16) + (g << 8) + b
}

// type WImg *image.RGBA 
type WImg struct {
    *image.RGBA
}

func newWImg(r image.Rectangle) WImg {
    return WImg{image.NewRGBA(r)}
}

func renderBorder(borderType, borderColor int, bgColor themeColor,
                  width, height, gradientType, gradientDir int) WImg {

    img := newWImg(image.Rect(0, 0, width, height))

    // bgClr could be a gradient!
    if bgColor.isGradient() {
        img.gradient(gradientType, gradientDir, bgColor)
    } else {
        bgClr := ColorFromInt(bgColor.start)
        for x := 0; x < width; x++ {
            for y := 0; y < height; y++ {
                img.SetRGBA(x, y, bgClr)
            }
        }
    }

    img.thinBorder(borderType, borderColor)
    return img
}

func renderCorner(borderType, borderColor int, bgColor themeColor,
                  width, height, diagonal int) WImg {
    // If bgColor isn't a gradient, then we can cheat
    if !bgColor.isGradient() {
        return renderBorder(borderType, borderColor, bgColor,
                            width, height, 0, 0)
    }

    img := newWImg(image.Rect(0, 0, width, height))

    // aliases for convenience
    vert, horz := renderGradientVert, renderGradientHorz
    reg, rev := renderGradientRegular, renderGradientReverse

    // for Top Left to Bottom Right diagonals
    belowTLDiag := func(x, y int) bool { return y > x }
    aboveTLDiag := func(x, y int) bool { return y <= x }

    // for Bottom Left to Top Right diagonals
    belowBLDiag := func(x, y int) bool { return y > (width - x) }
    aboveBLDiag := func(x, y int) bool { return y <= (width - x) }

    switch diagonal {
    case renderDiagBottomLeft:
        img.gradientFunc(horz, reg, bgColor, aboveBLDiag)
        img.gradientFunc(vert, rev, bgColor, belowBLDiag)
    case renderDiagTopRight:
        img.gradientFunc(vert, reg, bgColor, aboveBLDiag)
        img.gradientFunc(horz, rev, bgColor, belowBLDiag)
    case renderDiagBottomRight:
        img.gradientFunc(horz, rev, bgColor, aboveTLDiag)
        img.gradientFunc(vert, rev, bgColor, belowTLDiag)
    default: // renderDiagTopLeft:
        img.gradientFunc(vert, reg, bgColor, aboveTLDiag)
        img.gradientFunc(horz, reg, bgColor, belowTLDiag)
    }

    img.thinBorder(borderType, borderColor)
    return img
}

func (img WImg) thinBorder(borderType, borderColor int) {
    borderClr := ColorFromInt(borderColor)
    width, height := xgraphics.GetDim(img)

    // Now go through and add a "thin border."
    // It's always one pixel.
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            if borderType & renderBorderTop > 0 && y == 0 {
                img.SetRGBA(x, y, borderClr)
            }
            if borderType & renderBorderRight > 0 && x == width - 1 {
                img.SetRGBA(x, y, borderClr)
            }
            if borderType & renderBorderBottom > 0 && y == height - 1 {
                img.SetRGBA(x, y, borderClr)
            }
            if borderType & renderBorderLeft > 0 && x == 0 {
                img.SetRGBA(x, y, borderClr)
            }
        }
    }
}

func (img WImg) gradient(gradientType, gradientDir int, clr themeColor) {
    img.gradientFunc(gradientType, gradientDir, clr,
                     func(x, y int) bool { return true })
}

func (img WImg) gradientFunc(gradientType, gradientDir int, clr themeColor,
                             pixel func(x, y int) bool) {
    width, height := xgraphics.GetDim(img)

    if gradientType == renderGradientVert {
        steps := clr.steps(height)
        for x := 0; x < width; x++ {
            for y := 0; y < height; y++ {
                if pixel(x, y) {
                    if gradientDir == renderGradientReverse {
                        img.SetRGBA(x, y, steps[height - y - 1])
                    } else {
                        img.SetRGBA(x, y, steps[y])
                    }
                }
            }
        }
    } else {
        steps := clr.steps(width)
        for x := 0; x < width; x++ {
            for y := 0; y < height; y++ {
                if pixel(x, y) {
                    if gradientDir == renderGradientReverse {
                        img.SetRGBA(x, y, steps[width - x - 1])
                    } else {
                        img.SetRGBA(x, y, steps[x])
                    }
                }
            }
        }
    }
}

// isGradient returns whether the theme color is a gradient or not.
// Basically, a themeColor can either be a regular color (like an int)
// or a gradient when both 'start' and 'end' have valid color values.
func (tc themeColor) isGradient() bool {
    return tc.start >= 0 && tc.start <= 0xffffff &&
           tc.end >= 0 && tc.end <= 0xffffff
}

// steps returns a slice of colors corresponding to the gradient
// of colors from start to end. The size is determined by the 'size' parameter.
// The first and last colors in the slice are guaranteed to be
// tc.start and tc.end. (Unless the size is 1, in which case, the first and
// only color in the slice is tc.start.)
func (tc themeColor) steps(size int) []color.RGBA {
    // naughty
    if !tc.isGradient() {
        stps := make([]color.RGBA, size)
        for i := 0; i < size; i++ {
            stps[i] = ColorFromInt(tc.start)
        }
    }

    // yikes
    if size == 0 || size == 1 {
        return []color.RGBA{ColorFromInt(tc.start)}
    }

    stps := make([]color.RGBA, size)
    stps[0], stps[size - 1] = ColorFromInt(tc.start), ColorFromInt(tc.end)

    // no more?
    if size == 2 {
        return stps
    }

    sr, sg, sb := RGBFromInt(tc.start)
    er, eg, eb := RGBFromInt(tc.end)

    rinc := float64(er - sr) / float64(size)
    ginc := float64(eg - sg) / float64(size)
    binc := float64(eb - sb) / float64(size)

    doInc := func(inc float64, start, index int) int {
        return int(float64(start) + inc * float64(index))
    }

    var nr, ng, nb int
    for i := 1; i < size - 1; i++ {
        nr = doInc(rinc, sr, i)
        ng = doInc(ginc, sg, i)
        nb = doInc(binc, sb, i)

        stps[i] = ColorFromInt(IntFromRGB(nr, ng, nb))
    }

    return stps
}

