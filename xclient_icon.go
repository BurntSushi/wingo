package main

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/logger"
)

func (c *client) Icon(width, height int) *xgraphics.Image {
	ximg, err := xgraphics.FindIcon(X, c.Id(), width, height)
	if err != nil {
		logger.Message.Printf("Could not find icon for '%s': %s", c, err)
		ximg = xgraphics.NewConvert(X, wingo.theme.defaultIcon)
		ximg = ximg.Scale(width, height)
	}

	return ximg
}
