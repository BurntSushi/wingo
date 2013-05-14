package xclient

import (
	"github.com/BurntSushi/xgbutil/icccm"

	"github.com/BurntSushi/wingo-conc/stack"
)

// Transient is a wrapper around transient that type switches an empty interface
// to a *Client type. This is used to satisfy Client interfaces is various
// sub-packages.
//
// Currently, only values that have type *Client can be transient to each other.
func (c *Client) Transient(test stack.Client) bool {
	if testClient, ok := test.(*Client); ok {
		return c.transient(testClient)
	}
	return false
}

// transient determines whether 'test' is a transient window of 'c'.
// This is tricky because the logic to find the transients of a window is
// convoluted. Where C is the client we are trying find transients *for*, and
// c is any *other* client, the logic is something like this:
// If c has WM_TRANSIENT_FOR equal to C, then c is a transient of C.
// If c has window group (in WM_HINTS) equal to the window group in C,
// *and* c is one of the following window types:
// _NET_WM_WINDOW_TYPE_TOOLBAR
// _NET_WM_WINDOW_TYPE_MENU
// _NET_WM_WINDOW_TYPE_UTILITY
// _NET_WM_WINDOW_TYPE_DIALOG
// then c is a transient of C.
// There is one exception: if c and C are both transients in the same group,
// then they cannot be transient to each other.
func (c *Client) transient(test *Client) bool {
	if c == test {
		return false
	}
	if test.transientFor == c {
		return true
	}

	// If transientFor exists, then we don't look at window group stuff
	if test.transientFor != nil {
		return false
	}
	if c.hints.Flags&icccm.HintWindowGroup > 0 &&
		test.hints.Flags&icccm.HintWindowGroup > 0 &&
		c.hints.WindowGroup == test.hints.WindowGroup {

		return !c.transientType() && test.transientType()
	}
	return false
}

// transientType determines whether there is a transient type in the client.
func (c *Client) transientType() bool {
	return strIndex("_NET_WM_WINDOW_TYPE_TOOLBAR", c.winTypes) > -1 ||
		strIndex("_NET_WM_WINDOW_TYPE_MENU", c.winTypes) > -1 ||
		strIndex("_NET_WM_WINDOW_TYPE_UTILITY", c.winTypes) > -1 ||
		strIndex("_NET_WM_WINDOW_TYPE_DIALOG", c.winTypes) > -1
}
