package xclient

import (
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/wm"
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
	c.frame.Active()
	c.state = frame.Active
	focus.SetFocus(c)
}

func (c *Client) Unfocused() {
	c.frame.Inactive()
	c.state = frame.Inactive
}

func (c *Client) PrepareForFocus() {
	// There are only two ways a *managed* client is not prepared for focus:
	// 1) It belongs to any workspace except for the active one.
	// 2) It is iconified.
	// It is possible to be both. Check for both and remedy the situation.
	// We must check for (1) before (2), since a window cannot toggle its
	// iconification status if its workspace is not the current workspace.
	if c.workspace != wm.Workspace() {
		c.workspace.Activate(false) // don't be 'greedy'
	}
	if c.iconified {
		c.workspace.IconifyToggle(c)
	}
}
