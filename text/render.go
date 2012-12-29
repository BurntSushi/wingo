package text

import (
	"image"

	"code.google.com/p/jamslam-freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/render"
)

// BUG(burntsushi): I don't think freetype-go has a way to compute text extents
// before rendering text to an image. To work-around this, DrawText will over
// estimate the extents by assuming each character has a width equal to 1 em,
// and draw the text on an image with width = len(text) * pixels-per-em and
// a height = pixels-per-em. freetype-go then returns the point advanced by the
// drawn text, and this function uses it to take a sub-image of the image which
// is then drawn to the window. Unfortunately, this point doesn't reflect the
// true bounding box of the text (it cuts off the part of the text that dips
// below the base line). So to work-around this, the height of the extents is
// padded by a fixed pixel amount. This is wrong and will break if the font
// size is too large.

// DrawText is a convenience function that will create a new image, render
// the provided text to it, paint the image to the provided window, and resize
// the window to fit the text snugly.
//
// An error can occur when rendering the text to an image.
func DrawText(win *xwindow.Window, font *truetype.Font, size float64,
	fontClr, bgClr render.Color, text string) error {

	// Over estimate the extents.
	ew, eh := xgraphics.Extents(font, size, text)

	// Create an image using the over estimated extents.
	img := xgraphics.New(win.X, image.Rect(0, 0, ew, eh))
	xgraphics.BlendBgColor(img, bgClr.ImageColor())

	// Now draw the text, grab the (x, y) position advanced by the text, and
	// check for an error in rendering.
	_, _, err := img.Text(0, 0, fontClr.ImageColor(), size, font, text)
	if err != nil {
		return err
	}

	// Resize the window to the geometry determined by (x, y).
	win.Resize(ew, eh)

	// Now draw the image to the window and destroy it.
	img.XSurfaceSet(win.Id)
	// subimg := img.SubImage(image.Rect(0, 0, ew, eh))
	img.XDraw()
	img.XPaint(win.Id)
	img.Destroy()

	return nil
}
