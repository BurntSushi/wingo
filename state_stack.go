package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo/logger"
)

// The client stacking list is ordered for highest to lowest.
// This ordering simplifies the algorithm to restack a window, and we don't
// have to pay for it! (We just have to make sure we update
// _NET_CLIENT_LIST_STACKING in reverse.)

// Represents each discrete "stacking layer" maintained by Wingo.
// The layers listed here are in order from lowest to highest.
// Namely, a window in layer X can *never* be above a window in layer X + 1
// and can *never* be below a window in layer X - 1.
const (
	stackDesktop = iota
	stackBelow
	stackDefault
	stackAbove
	stackDock
	stackFullscreen
)

// updateEwmhStacking refreshes the _NET_CLIENT_LIST_STACKING property on the
// root window.
func (wm *state) updateEwmhStacking() {
	numWins := len(wm.stack)
	winList := make([]xproto.Window, numWins)
	for i, c := range wm.stack {
		winList[numWins-i-1] = c.Win().id
	}
	err := ewmh.ClientListStackingSet(X, winList)
	if err != nil {
		logger.Warning.Printf("Could not update _NET_CLIENT_LIST_STACKING "+
			"because %v", err)
	}
}

// stackUpdate forces the current state of the stack to be reality.
// This is useful when we want to make multiple modifications to the stack,
// and apply them all at once. This prevents window flashing when the stack
// is unchanged.
func (wm *state) stackUpdate(clients []*client) {
	if len(wm.stack) <= 1 {
		return
	}
	for i := len(wm.stack) - 1; i >= 0; i-- {
		if cliIndex(wm.stack[i], clients) > -1 {
			if i == len(wm.stack)-1 {
				wm.stack[i].Frame().ConfigureClient(
					DoSibling|DoStack, 0, 0, 0, 0,
					wm.stack[i-1].Frame().ParentId(),
					xproto.StackModeBelow, false)
			} else {
				wm.stack[i].Frame().ConfigureClient(
					DoSibling|DoStack, 0, 0, 0, 0,
					wm.stack[i+1].Frame().ParentId(),
					xproto.StackModeAbove, false)
			}
		}
	}
}

// stackRaise raises the given client to the top of its layer.
// If configure is false, this becomes a state-modifying function only.
// Which is used when first managing a window, or when complying with
// a user request to restack.
func (wm *state) stackRaise(c *client, configure bool) {
	// make sure we update the EWMH stacking list when we're done
	defer wm.updateEwmhStacking()

	// if we've stacked this client before, remove it from the list.
	// this allows us not to care whether the client has changed layers.
	wm.stackRemove(c)

	// A special case: when the stack is empty, just add the client
	// with no magic.
	if len(wm.stack) == 0 {
		wm.stack = append(wm.stack, c)
		return
	}

	// now find where we need to place the client into the stack
	// and issue the appropriate stacking request.
	// Remember, wm.stack is ordered by highest to lowest.
	// Therefore, the first client we find in c's layer, we stack it on top
	// of that. If we can't find a client but have hit a layer that is "lower"
	// than c's, then stack c above that client.
	for i, c2 := range wm.stack {
		if c == c2 {
			continue
		}

		if c2.Layer() <= c.Layer() {
			if configure {
				c.Frame().ConfigureClient(DoSibling|DoStack, 0, 0, 0, 0,
					c2.Frame().ParentId(),
					xproto.StackModeAbove, false)
			}
			wm.stack = append(wm.stack[:i],
				append([]*client{c}, wm.stack[i:]...)...)
			return
		}
	}

	// If we're here, that means we couldn't find any clients in the
	// stacking list that were in a layer below the client's desired layer.
	// Thus, it must be the lowest client and it gets added to the end.
	// We also must stack it below the lowest window in the list.
	if configure {
		c.Frame().ConfigureClient(
			DoSibling|DoStack, 0, 0, 0, 0,
			wm.stack[len(wm.stack)-1].Frame().ParentId(),
			xproto.StackModeBelow, false)
	}
	wm.stack = append(wm.stack, c)
}

// stackRemove removes a client from the stacking list.
// This is only done when we raise a client (in which case, the client is
// subsequently re-added to the stacking list) or when a client is unmanaged.
// We maintain a client's stacking position even when it is unmapped.
func (wm *state) stackRemove(c *client) {
	if i := cliIndex(c, wm.stack); i > -1 {
		wm.stack = append(wm.stack[:i], wm.stack[i+1:]...)
	}
}
