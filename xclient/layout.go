package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo-conc/layout"
	"github.com/BurntSushi/wingo-conc/wm"
	"github.com/BurntSushi/wingo-conc/workspace"
)

func (c *Client) CheckNewWorkspace() {
	var newWrk *workspace.Workspace = nil
	curWrk := c.Workspace()

	if dragGeom := c.DragGeom(); dragGeom != nil {
		newWrk = wm.Heads.FindMostOverlap(dragGeom)
	} else {
		newWrk = wm.Heads.FindMostOverlap(c.frame.Geom())
	}
	if newWrk == nil || curWrk == newWrk {
		return
	}

	newWrk.Add(c)

	// If this is the active window, switch to this workspace too.
	if c.IsActive() {
		wm.SetWorkspace(newWrk, false)
	}
}

func (c *Client) Layout() layout.Layout {
	return c.workspace.Layout(c)
}

func (c *Client) LayoutMROpt(flags, x, y, width, height int) {
	c.resizing = true
	c.Layout().MROpt(c, flags, x, y, width, height)
	c.CheckNewWorkspace()
	c.resizing = false
}

func (c *Client) LayoutMoveResize(x, y, width, height int) {
	c.resizing = true
	c.Layout().MoveResize(c, x, y, width, height)
	c.CheckNewWorkspace()
	c.resizing = false
}

func (c *Client) LayoutMove(x, y int) {
	c.moving = true
	c.Layout().Move(c, x, y)
	c.CheckNewWorkspace()
	c.moving = false
}

func (c *Client) LayoutResize(width, height int) {
	c.resizing = true
	c.Layout().Resize(c, width, height)
	c.CheckNewWorkspace()
	c.resizing = false
}

func (c *Client) Geom() xrect.Rect {
	return c.frame.Geom()
}

func (c *Client) FrameTile() {
	c.EnsureUnmax()
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

func (c *Client) MoveResize(x, y, width, height int) {
	c.frame.MoveResize(false, x, y, width, height)
}

func (c *Client) MoveResizeValid(x, y, width, height int) {
	c.frame.MoveResize(true, x, y, width, height)
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
