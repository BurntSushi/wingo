package main

import (
	"github.com/BurntSushi/wingo/layout"
)

func (c *client) Layout() layout.Layout {
	return c.workspace.Layout(c)
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

func (c *client) MoveResize(validate bool, x, y, width, height int) {
	c.frame.MoveResize(validate, x, y, width, height)
}

func (c *client) Move(x, y int) {
	c.frame.Move(x, y)
}

func (c *client) Resize(validate bool, width, height int) {
	c.frame.Resize(validate, width, height)
}
