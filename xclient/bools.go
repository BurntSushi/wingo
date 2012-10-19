package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
)

func (c *Client) IsActive() bool {
	return c.state == frame.Active
}

func (c *Client) IsMapped() bool {
	return c.frame.IsMapped()
}

func (c *Client) IsMaximized() bool {
	return c.maximized
}

func (c *Client) IsSticky() bool {
	return c.sticky
}

func (c *Client) Iconified() bool {
	return c.iconified
}

func (c *Client) hasType(atom string) bool {
	return strIndex(atom, c.winTypes) > -1
}

func (c *Client) CanFocus() bool {
	return c.hints.Flags&icccm.HintInput > 0 && c.hints.Input == 1
}

func (c *Client) SendFocusNotify() bool {
	return strIndex("WM_TAKE_FOCUS", c.protocols) > -1
}

func (c *Client) IsTransient() bool {
	return c.transientFor != nil
}

func (c *Client) IsSkipPager() bool {
	return c.skipPager
}

func (c *Client) IsSkipTaskbar() bool {
	return c.skipTaskbar
}

// shouldDecor returns false if the client has requested no frames or
// has a type that implies it shouldn't be decorated.
func (c *Client) shouldDecor() bool {
	if c.PrimaryType() != TypeNormal {
		return false
	}
	if c.hasType("_NET_WM_WINDOW_TYPE_SPLASH") {
		return false
	}

	mh, err := motif.WmHintsGet(wm.X, c.Id())
	if err == nil && !motif.Decor(mh) {
		return false
	}

	return true
}

func (c *Client) isAttrsUnmapped() bool {
	attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), c.Id()).Reply()
	if err != nil {
		logger.Warning.Printf(
			"Could not get window attributes for '%s': %s.", c, err)
		return false
	}
	return attrs.MapState == xproto.MapStateUnmapped
}

// isFixedSize returns true when the client has the minimum and maximum
// width equivalent AND has the minimum and maximum height equivalent.
func (c *Client) isFixedSize() bool {
	return c.nhints.Flags&icccm.SizeHintPMinSize > 0 &&
		c.nhints.Flags&icccm.SizeHintPMaxSize > 0 &&
		c.nhints.MinWidth == c.nhints.MaxWidth &&
		c.nhints.MinHeight == c.nhints.MaxHeight
}
