package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
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
    layerDesktop = iota
    layerBelow
    layerDefault
    layerDock
    layerAbove
    layerFullscreen
)

// updateEwmhStacking refreshes the _NET_CLIENT_LIST_STACKING property on the
// root window.
func (wm *state) updateEwmhStacking() {
    numWins := len(wm.stack)
    winList := make([]xgb.Id, numWins)
    for i, c := range wm.stack {
        winList[numWins - i - 1] = c.Win().id
    }
    err := ewmh.ClientListStackingSet(X, winList)
    if err != nil {
        logWarning.Printf("Could not update _NET_CLIENT_LIST_STACKING " +
                          "because %v", err)
    }
}

// stackRaise raises the given client to the top of its layer.
// If configure is false, this becomes a state-modifying function only.
// Which is used when first managing a window, or when complying with
// a user request to restack.
func (wm *state) stackRaise(c Client, configure bool) {
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
            // When configuring, we either stack it below the client that
            // came before this one, or stack it above everything else if
            // it's c2 is the top client.
            if configure {
                logMessage.Printf("Stacking %s above %s", c, c2)
                c.Frame().ConfigureClient(DoSibling | DoStack, 0, 0, 0, 0,
                                          c2.Frame().ParentId(),
                                          xgb.StackModeAbove, false)
            }
            wm.stack = append(wm.stack[:i],
                              append([]Client{c}, wm.stack[i:]...)...)
            return
        }
    }

    // If we're here, that means we couldn't find any clients in the
    // stacking list that were in a layer below the client's desired layer.
    // Thus, it must be the lowest client and it gets added to the end.
    // We also must stack it below the lowest window in the list.
    if configure {
        c.Frame().ConfigureClient(
            DoSibling | DoStack, 0, 0, 0, 0,
            wm.stack[len(wm.stack) - 1].Frame().ParentId(),
            xgb.StackModeBelow, false)
    }
    wm.stack = append(wm.stack, c)
}

// stackRemove removes a client from the stacking list.
// This is only done when we raise a client (in which case, the client is
// subsequently re-added to the stacking list) or when a client is unmanaged.
// We maintain a client's stacking position even when it is unmapped.
func (wm *state) stackRemove(c Client) {
    if i := cliIndex(c, wm.stack); i > -1 {
        wm.stack = append(wm.stack[:i], wm.stack[i+1:]...)
    }
}

