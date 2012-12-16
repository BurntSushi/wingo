package xclient

import (
	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/cshapeshifter/wingo/focus"
	"github.com/cshapeshifter/wingo/frame"
	"github.com/cshapeshifter/wingo/hook"
	"github.com/cshapeshifter/wingo/wm"
	"github.com/cshapeshifter/wingo/workspace"
)

func (c *Client) Focus() {
	c.attnStop()
	focus.Focus(c)
}

func (c *Client) Focused() {
	c.attnStop()
	c.frame.Active()
	c.state = frame.Active
	focus.SetFocus(c)
	ewmh.ActiveWindowSet(wm.X, c.Id())
	c.addState("_NET_WM_STATE_FOCUSED")

	c.FireHook(hook.Focused)
}

func (c *Client) Unfocused() {
	c.frame.Inactive()
	c.state = frame.Inactive
	ewmh.ActiveWindowSet(wm.X, 0)
	c.removeState("_NET_WM_STATE_FOCUSED")

	c.FireHook(hook.Unfocused)
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
			wm.SetWorkspace(wrk, false)
		}
	}
	if c.iconified {
		c.IconifyToggle()
	}
}
