package main

import (
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/stack"
)

func (c *client) unmanage() {
	X.Grab()
	defer X.Ungrab()

	c.win.Detach()
	icccm.WmStateSet(c.X, c.Id(), &icccm.WmState{State: icccm.StateWithdrawn})
	c.Unmap()
	c.workspace.Remove(c)
	c.frames.destroy()
	c.prompts.destroy()

	focus.Remove(c)
	stack.Remove(c)
	wingo.removeClient(c)
	wingo.focusFallback()
}
