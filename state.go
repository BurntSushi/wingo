package main

import (
    "strings"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
    // "github.com/BurntSushi/xgbutil/xevent" 
)

// state is the master singleton the carries all window manager related state
type state struct {
    clients []Client // a list of clients in order of being added
    stack []Client // clients ordered by visual stack
    focus []Client // focus ordering of clients; may be smaller than 'clients'
}

func newState() *state {
    return &state{
        clients: make([]Client, 0),
        stack: make([]Client, 0),
        focus: make([]Client, 0),
    }
}

func (wm *state) clientAdd(c Client) {
    if cliIndex(c, wm.clients) == -1 {
        logMessage.Println("Managing new client:", c)
        wm.clients = append(wm.clients, c)
        wm.updateEwmhClients()
    } else {
        logMessage.Println("Already managing client:", c)
    }
}

func (wm *state) clientRemove(c Client) {
    if i := cliIndex(c, wm.clients); i > -1 {
        logMessage.Println("Unmanaging client:", c)
        wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
        wm.updateEwmhClients()
    }
}

func (wm *state) updateEwmhClients() {
    numWins := len(wm.clients)
    winList := make([]xgb.Id, numWins)
    for i, c := range wm.clients {
        winList[i] = c.Win().id
    }
    err := ewmh.ClientListSet(X, winList)
    if err != nil {
        logWarning.Printf("Could not update _NET_CLIENT_LIST " +
                          "because %v", err)
    }
}

func (wm *state) focused() Client {
    for i := len(wm.focus) - 1; i >= 0; i-- {
        if wm.focus[i].Mapped() {
            return wm.focus[i]
        }
    }
    return nil
}

func (wm *state) unfocusExcept(id xgb.Id) {
    for _, c := range wm.focus {
        if c.Id() != id {
            c.Unfocused()
        }
    }
}

func (wm *state) focusAdd(c Client) {
    wm.focusRemove(c)
    wm.focus = append(wm.focus, c)
}

func (wm *state) focusRemove(c Client) {
    if i := cliIndex(c, wm.focus); i > -1 {
        wm.focus = append(wm.focus[:i], wm.focus[i+1:]...)
    }
}

func (wm *state) fallback() {
    var c Client
    for i := len(wm.focus) - 1; i >= 0; i-- {
        c = wm.focus[i]
        if c.Mapped() && c.Alive() {
            logMessage.Printf("Focus falling back to %s", c)
            c.Focus()
            return
        }
    }

    // No windows to fall back on... root focus
    // this is IMPORTANT. if we fail here, we risk a lock-up
    logMessage.Printf("Focus falling back to ROOT")
    ROOT.focus()
    X.Conn().SetInputFocus(xgb.InputFocusPointerRoot, X.RootWin(), 0)
}

func (wm *state) logClientList() {
    list := make([]string, len(wm.clients))
    for i, c := range wm.clients {
        list[i] = c.String()
    }
    logMessage.Println(strings.Join(list, ", "))
}

