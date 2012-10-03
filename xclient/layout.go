package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/layout"
)

func (c *Client) Layout() layout.Layout {
	return c.workspace.Layout(c)
}

func (c *Client) LayoutMROpt(flags, x, y, width, height int) {
	c.Layout().MROpt(c, flags, x, y, width, height)
}

func (c *Client) LayoutMoveResize(x, y, width, height int) {
	c.Layout().MoveResize(c, x, y, width, height)
}

func (c *Client) LayoutMove(x, y int) {
	c.Layout().Move(c, x, y)
}

func (c *Client) LayoutResize(width, height int) {
	c.Layout().Resize(c, width, height)
}

func (c *Client) Geom() xrect.Rect {
	return c.frame.Geom()
}

func (c *Client) FrameTile() {
	c.FrameBorders()
}

func (c *Client) MROpt(validate bool, flags, x, y, w, h int) {
	c.frame.MROpt(validate, flags, x, y, w, h)

	// As per ICCCM 4.1.5, a window that has been moved but not resized must
	// receive a synthetic ConfigureNotify event.
	if flags&xproto.ConfigWindowWidth == 0 &&
		flags&xproto.ConfigWindowHeight == 0 {

		c.sendConfigureNotify()
	}
}

func (c *Client) MoveResize(validate bool, x, y, width, height int) {
	c.frame.MoveResize(validate, x, y, width, height)
}

func (c *Client) Move(x, y int) {
	c.frame.Move(x, y)

	// As per ICCCM 4.1.5, a window that has been moved but not resized must
	// receive a synthetic ConfigureNotify event.
	c.sendConfigureNotify()
}

func (c *Client) Resize(validate bool, width, height int) {
	c.frame.Resize(validate, width, height)
}
