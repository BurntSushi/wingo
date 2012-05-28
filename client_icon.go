package main

import (
	"image/draw"

	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/logger"
)

func (c *client) iconImage(width, height int) draw.Image {
	ximg, err := xgraphics.FindIcon(X, c.Id(), width, height)
	if err != nil {
		logger.Message.Printf("Could not find icon for '%s': %s", c, err)
		ximg = xgraphics.NewConvert(X, THEME.defaultIcon)
	}

	return ximg
}
