package main

import (
    "burntsushi.net/go/xgbutil/icccm"
)

// transient determines whether 'test' is a transient window of 'c'.
// This is tricky because the logic to find the transients of a window is so
// complex. Where C is the client we are trying find transients *for*, and
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
func (c *client) transient(test *client) bool {
    if c == test {
        return false
    }

    if test.transientFor == c.Id() {
        return true
    }

    // If transientFor exists, then we don't look at window group stuff
    if test.transientFor > 0 {
        return false
    }

    if c.hints.Flags & icccm.HintWindowGroup > 0 &&
       test.hints.Flags & icccm.HintWindowGroup > 0 &&
       c.hints.WindowGroup == test.hints.WindowGroup {
        return !c.transientType() && test.transientType()
    }

    return false
}

// transientTypes determines whether there is a transient type in the client.
func (c *client) transientType() bool {
    return strIndex("_NET_WM_WINDOW_TYPE_TOOLBAR", c.types) > -1 ||
           strIndex("_NET_WM_WINDOW_TYPE_MENU", c.types) > -1 ||
           strIndex("_NET_WM_WINDOW_TYPE_UTILITY", c.types) > -1 ||
           strIndex("_NET_WM_WINDOW_TYPE_DIALOG", c.types) > -1
}

