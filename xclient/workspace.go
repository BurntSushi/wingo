package xclient

import (
	"fmt"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo-conc/wm"
	"github.com/BurntSushi/wingo-conc/workspace"
)

// ShouldForceFloating returns true whenever a client should be floating.
// More specifically, it returns true when a client should NOT be added to
// a tiling layout even if a tiling layout is active.
func (c *Client) ShouldForceFloating() bool {
	return c.floating ||
		c.sticky ||
		c.fullscreen ||
		c.transientFor != nil ||
		c.PrimaryType() != TypeNormal ||
		c.isFixedSize() ||
		c.hasType("_NET_WM_WINDOW_TYPE_SPLASH")
}

func (c *Client) Workspace() workspace.Workspacer {
	return c.workspace
}

func (c *Client) WorkspaceSet(newWrk workspace.Workspacer) {
	c.workspace = newWrk

	switch wrk := c.workspace.(type) {
	case *workspace.Sticky:
		ewmh.WmDesktopSet(wm.X, c.Id(), 0xFFFFFFFF)
	case *workspace.Workspace:
		ewmh.WmDesktopSet(wm.X, c.Id(), uint(wm.Heads.GlobalIndex(wrk)))
	default:
		panic(fmt.Sprintf("Unknown workspace type: %T", wrk))
	}
}

func (c *Client) IconifyToggle() {
	c.Workspace().IconifyToggle(c)

	if c.Iconified() {
		c.addState("_NET_WM_STATE_HIDDEN")
	} else {
		c.removeState("_NET_WM_STATE_HIDDEN")
	}
}

func (c *Client) Iconify() {
	if c.Iconified() {
		return
	}
	c.IconifyToggle()
}

func (c *Client) Deiconify() {
	if !c.Iconified() {
		return
	}
	c.IconifyToggle()
}

func (c *Client) IconifiedSet(iconified bool) {
	c.iconified = iconified
}
