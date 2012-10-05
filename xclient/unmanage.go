package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
)

func (c *Client) unmanage() {
	wm.X.Grab()
	defer wm.X.Ungrab()

	go func() {
		c.frames.destroy()
		c.prompts.destroy()
	}()

	logger.Message.Printf("Unmanaging client: %s", c)

	c.frame.Unmap()
	c.win.Detach()
	icccm.WmStateSet(wm.X, c.Id(), &icccm.WmState{State: icccm.StateWithdrawn})
	focus.Remove(c)
	wm.FocusFallback()
	stack.Remove(c)
	c.workspace.Remove(c)
	wm.RemoveClient(c)

	if c.hadStruts {
		wm.Heads.ApplyStruts(wm.Clients)
	}
}

func (c *Client) ImminentDestruction() bool {
	toIgnore := c.unmapIgnore
	for _, evOrErr := range xevent.Peek(wm.X) {
		ev := evOrErr.Event
		if ev == nil {
			continue
		}

		evUnmap, ok := ev.(xproto.UnmapNotifyEvent)
		if !ok {
			continue
		}

		if evUnmap.Window == c.Id() {
			toIgnore--
			if toIgnore <= 0 {
				return true
			}
		}
	}
	return false
}
