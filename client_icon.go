package main

import (
	"image"
	"image/color"
	"image/draw"
)

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xgraphics"
)

func (c *client) iconImage(width, height int) (draw.Image, draw.Image) {
	var img, mask draw.Image
	var iok, mok bool

	img, mask, iok, mok = c.iconTryEwmh(width, height)
	if iok {
		goto DONE
	}

	img, mask, iok, mok = c.iconTryIcccm()
	if iok {
		goto DONE
	}

	iok, mok = false, false
	img = THEME.defaultIcon
	if img != nil {
		iok = true
		goto DONE
	}

DONE:
	// If we've got an image, great! If not, just create a completely
	// transparent image. It will take up space, but it won't look
	// horrendous and we won't crash.
	if iok {
		img = xgraphics.Scale(img, width, height)
	} else {
		uni := image.NewUniform(color.RGBA{0, 0, 0, 0})
		img = image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(img, img.Bounds(), uni, image.ZP, draw.Src)
	}

	// Just as we did for img above, if we don't have a mask, create a benign
	// mask in its stead.
	if mok {
		mask = xgraphics.Scale(mask, width, height)
	} else {
		uni := image.NewUniform(color.RGBA{0, 0, 0, 255})
		mask = image.NewRGBA(img.Bounds())
		draw.Draw(mask, mask.Bounds(), uni, image.ZP, draw.Src)
	}
	return img, mask
}

func (c *client) iconTryEwmh(width, height int) (
	*image.RGBA, *image.RGBA, bool, bool) {

	icons, err := ewmh.WmIconGet(X, c.Id())
	if err != nil {
		logWarning.Printf("Could not get EWMH icon for window %s because: %v",
			c, err)
		return nil, nil, false, false
	}

	icon := xgraphics.FindBestIcon(width, height, icons)
	if icon == nil {
		logWarning.Printf("Could not find any decent icon for size (%d, %d) "+
			" on window %s.", width, height, c)
		return nil, nil, false, false
	}

	img, mask := xgraphics.EwmhIconToImage(icon)
	return img, mask, true, true
}

func (c *client) iconTryIcccm() (*image.RGBA, *image.RGBA, bool, bool) {
	if c.hints.Flags&icccm.HintIconPixmap == 0 ||
		c.hints.IconPixmap == 0 || c.hints.IconMask == 0 {
		return nil, nil, false, false
	}

	img, err := xgraphics.PixmapToImage(X, c.hints.IconPixmap)
	if err != nil {
		logWarning.Printf("Could not get IconPixmap from window %s "+
			"because: %v", err)
		return nil, nil, false, false
	}

	mask, err := xgraphics.BitmapToImage(X, c.hints.IconMask)
	if err != nil {
		logWarning.Printf("Could not get IconMask from window %s "+
			"because: %v", err)
		return img, nil, true, false
	}

	return img, mask, true, true
}
