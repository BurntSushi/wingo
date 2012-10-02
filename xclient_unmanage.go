package main

import (
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/stack"
)

func (c *client) unmanage() {
	X.Grab()
	defer X.Ungrab()

	go func() {
		c.frames.destroy()
		c.prompts.destroy()
	}()

	c.Unmap()
	c.win.Detach()
	icccm.WmStateSet(c.X, c.Id(), &icccm.WmState{State: icccm.StateWithdrawn})
	focus.Remove(c)
	wingo.focusFallback()
	stack.Remove(c)
	c.workspace.Remove(c)
	wingo.removeClient(c)
}
