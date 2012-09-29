package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/layout"
)

func (c *client) Layout() layout.Layout {
	return c.workspace.Layout(c)
}

func (c *client) LayoutMROpt(flags, x, y, width, height int) {
	c.Layout().MROpt(c, flags, x, y, width, height)
}

func (c *client) LayoutMoveResize(x, y, width, height int) {
	c.Layout().MoveResize(c, x, y, width, height)
}

func (c *client) LayoutMove(x, y int) {
	c.Layout().Move(c, x, y)
}

func (c *client) LayoutResize(width, height int) {
	c.Layout().Resize(c, width, height)
}

func (c *client) FrameTile() {
	c.FrameBorders()
}

func (c *client) MROpt(validate bool, flags, x, y, w, h int) {
	c.frame.MROpt(validate, flags, x, y, w, h)

	// As per ICCCM 4.1.5, a window that has been moved but not resized must
	// receive a synthetic ConfigureNotify event.
	if flags&xproto.ConfigWindowWidth == 0 &&
		flags&xproto.ConfigWindowHeight == 0 {

		c.sendConfigureNotify()
	}
}

func (c *client) MoveResize(validate bool, x, y, width, height int) {
	c.frame.MoveResize(validate, x, y, width, height)
}

func (c *client) Move(x, y int) {
	c.frame.Move(x, y)

	// As per ICCCM 4.1.5, a window that has been moved but not resized must
	// receive a synthetic ConfigureNotify event.
	c.sendConfigureNotify()
}

func (c *client) Resize(validate bool, width, height int) {
	c.frame.Resize(validate, width, height)
}
