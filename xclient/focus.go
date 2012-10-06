package xclient

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
)

func (c *Client) CanFocus() bool {
	return c.hints.Flags&icccm.HintInput > 0 && c.hints.Input == 1
}

func (c *Client) SendFocusNotify() bool {
	return strIndex("WM_TAKE_FOCUS", c.protocols) > -1
}

func (c *Client) IsActive() bool {
	return c.state == frame.Active
}

func (c *Client) Focused() {
	c.attnStop()
	c.frame.Active()
	c.state = frame.Active
	focus.SetFocus(c)
	ewmh.ActiveWindowSet(wm.X, c.Id())
}

func (c *Client) Unfocused() {
	c.frame.Inactive()
	c.state = frame.Inactive
	ewmh.ActiveWindowSet(wm.X, 0)
}

func (c *Client) PrepareForFocus() {
	// There are only two ways a *managed* client is not prepared for focus:
	// 1) It belongs to any workspace except for the active one.
	// 2) It is iconified.
	// It is possible to be both. Check for both and remedy the situation.
	// We must check for (1) before (2), since a window cannot toggle its
	// iconification status if its workspace is not the current workspace.
	if c.workspace != wm.Workspace() {
		// This isn't applicable if we're sticky.
		if wrk, ok := c.workspace.(*workspace.Workspace); ok {
			wrk.Activate(false) // don't be 'greedy'
		}
	}
	if c.iconified {
		c.IconifyToggle()
	}
}
