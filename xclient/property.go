package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/cshapeshifter/wingo/layout"
	"github.com/cshapeshifter/wingo/wm"
)

func (c *Client) handleProperty(name string) {
	switch name {
	case "_NET_WM_VISIBLE_NAME":
		fallthrough
	case "_NET_WM_NAME":
		fallthrough
	case "WM_NAME":
		c.refreshName()
	case "_NET_WM_ICON":
		c.refreshIcon()
	case "WM_HINTS":
		if hints, err := icccm.WmHintsGet(wm.X, c.Id()); err == nil {
			c.hints = hints
			c.refreshIcon()
		}
	case "WM_NORMAL_HINTS":
		if nhints, err := icccm.WmNormalHintsGet(wm.X, c.Id()); err == nil {
			c.nhints = nhints
		}
	case "WM_TRANSIENT_FOR":
		if trans, err := icccm.WmTransientForGet(wm.X, c.Id()); err == nil {
			if transCli := wm.FindManagedClient(trans); transCli != nil {
				c.transientFor = transCli.(*Client)
			}
		}
	case "_NET_WM_USER_TIME":
		if newTime, err := ewmh.WmUserTimeGet(wm.X, c.Id()); err == nil {
			c.time = xproto.Timestamp(newTime)
		}
	case "_NET_WM_STRUT_PARTIAL":
		c.maybeApplyStruts()
	case "_MOTIF_WM_HINTS":
		// This is a bit messed up. If a client is floating, we don't
		// really care what the decorations are, so we oblige blindly.
		// However, if we're tiling, then we don't want to mess with
		// the frames---but we also want to make sure that any states
		// the client might revert to have the proper frames.
		decor := c.shouldDecor()
		if _, ok := c.Layout().(layout.Floater); ok {
			if decor {
				c.FrameFull()
			} else {
				c.FrameNada()
			}
		} else {
			for k := range c.states {
				s := c.states[k]
				if decor {
					s.frame = c.frames.full
				} else {
					s.frame = c.frames.nada
				}
				c.states[k] = s
			}
		}
	}
}

func (c *Client) refreshIcon() {
	c.frames.full.UpdateIcon()
	c.prompts.updateIcon()
}

func (c *Client) refreshName() {
	var newName string

	defer func() {
		if newName != c.name {
			c.name = newName
			c.frames.full.UpdateTitle()
			c.prompts.updateName()
			ewmh.WmVisibleNameSet(wm.X, c.Id(), c.name)
		}
	}()

	newName, _ = ewmh.WmNameGet(wm.X, c.Id())
	if len(newName) > 0 {
		return
	}

	newName, _ = icccm.WmNameGet(wm.X, c.Id())
	if len(newName) > 0 {
		return
	}

	newName = "Unnamed Window"
}

func (c *Client) maybeApplyStruts() {
	if strut, _ := ewmh.WmStrutPartialGet(wm.X, c.Id()); strut != nil {
		c.hadStruts = true
		wm.Heads.ApplyStruts(wm.Clients)
	}
}
