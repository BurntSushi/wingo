package main

import (
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
)

func (c *client) CanFocus() bool {
	return c.hints.Flags&icccm.HintInput > 0 && c.hints.Input == 1
}

func (c *client) SendFocusNotify() bool {
	return strIndex("WM_TAKE_FOCUS", c.protocols) > -1
}

func (c *client) Focused() {
	c.frame.Active()
	c.state = frame.Active
	focus.UnfocusExcept(c)
}

func (c *client) Unfocused() {
	if c.state == frame.Inactive {
		return
	}
	c.frame.Inactive()
	c.state = frame.Inactive
}

func (c *client) PrepareForFocus() {
	// There are only two ways a *managed* client is not prepared for focus:
	// 1) It belongs to any workspace except for the active one.
	// 2) It is iconified.
	// It is possible to be both. Check for both and remedy the situation.
	// We must check for (1) before (2), since a window cannot toggle its
	// iconification status if its workspace is not the current workspace.
	if c.workspace != wingo.workspace() {
		c.workspace.Activate(false) // don't be 'greedy'
	}
	if c.iconified {
		c.workspace.IconifyToggle(c)
	}
}
