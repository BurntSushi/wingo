package main

import (
    "strings"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

// state is the master singleton the carries all window manager related state
type state struct {
    clients []client // a list of clients in order of being added
    stack []client // clients ordered by visual stack
    focus []client // focus ordering of clients; may be smaller than 'clients'
}

func newState() *state {
    return &state{
        clients: make([]client, 0),
        stack: make([]client, 0),
        focus: make([]client, 0),
    }
}

func (wm *state) clientAdd(c client) {
    if cliIndex(c, wm.clients) == -1 {
        wm.clients = append(wm.clients, c)
        logMessage.Println("Managing new client:", c)
    } else {
        logMessage.Println("Already managing client:", c)
    }
}

func (wm *state) clientRemove(c client) {
    if i := cliIndex(c, wm.clients); i > -1 {
        wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
        logMessage.Println("Unmanaging client:", c)
    }
}

func (wm *state) focused() client {
    for i := len(wm.focus) - 1; i >= 0; i-- {
        if wm.focus[i].mapped() {
            return wm.focus[i]
        }
    }
    return nil
}

func (wm *state) focusAdd(c client) {
    wm.focusRemove(c)
    wm.focus = append(wm.focus, c)
}

func (wm *state) focusRemove(c client) {
    if i := cliIndex(c, wm.focus); i > -1 {
        wm.focus = append(wm.focus[:i], wm.focus[i+1:]...)
    }
}

func (wm *state) fallback() {
    var c client
    for i := len(wm.focus) - 1; i >= 0; i-- {
        c = wm.focus[i]
        if c.mapped() {
            c.focus()
            return
        }
    }

    // No windows to fall back on... root focus
    // this is IMPORTANT. if we fail here, we risk a lock-up
    X.Conn().SetInputFocus(xgb.InputFocusPointerRoot, X.RootWin(), 0)
}

func (wm *state) logClientList() {
    list := make([]string, len(wm.clients))
    for i, c := range wm.clients {
        list[i] = c.String()
    }
    logMessage.Println(strings.Join(list, ", "))
}

