package xclient

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/cshapeshifter/wingo/logger"
	"github.com/cshapeshifter/wingo/wm"
)

func (c *Client) Icon(width, height int) *xgraphics.Image {
	ximg, err := xgraphics.FindIcon(wm.X, c.Id(), width, height)
	if err != nil {
		logger.Message.Printf("Could not find icon for '%s': %s", c, err)
		ximg = xgraphics.NewConvert(wm.X, wm.Theme.DefaultIcon)
		ximg = ximg.Scale(width, height)
	}

	return ximg
}
