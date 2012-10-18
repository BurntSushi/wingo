package xclient

import (
	"fmt"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
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

// isFixedSize returns true when the client has the minimum and maximum
// width equivalent AND has the minimum and maximum height equivalent.
func (c *Client) isFixedSize() bool {
	return c.nhints.Flags&icccm.SizeHintPMinSize > 0 &&
		c.nhints.Flags&icccm.SizeHintPMaxSize > 0 &&
		c.nhints.MinWidth == c.nhints.MaxWidth &&
		c.nhints.MinHeight == c.nhints.MaxHeight
}

func (c *Client) FloatingToggle() {
	// Doesn't work on sticky windows. They are already floating.
	if wrk, ok := c.Workspace().(*workspace.Workspace); ok {
		c.floating = !c.floating
		wrk.CheckFloatingStatus(c)
	}
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
		ewmh.WmDesktopSet(wm.X, c.Id(), int64(wm.Heads.GlobalIndex(wrk)))
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

func (c *Client) Iconified() bool {
	return c.iconified
}

func (c *Client) IconifiedSet(iconified bool) {
	c.iconified = iconified
}
