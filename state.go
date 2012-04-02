package main

import (
    "fmt"
    "strings"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/xinerama"
    "github.com/BurntSushi/xgbutil/xrect"
)

// state is the master singleton the carries all window manager related state
type state struct {
    clients []*client // a list of clients in order of being added
    stack []*client // clients ordered by visual stack
    focus []*client // focus ordering of clients; may be smaller than 'clients'
    heads xinerama.Heads
    workspaces workspaces
}

func headsLoad() xinerama.Heads {
    heads, err := xinerama.PhysicalHeads(X)
    if err != nil || len(heads) == 0 {
        if err == nil {
            logWarning.Printf("Could not find any physical heads with the " +
                              "Xinerama extension.")
        } else {
            logWarning.Printf("Could not load physical heads via Xinerama: %s",
                              err)
        }
        logWarning.Printf("Assuming one head with size equivalent to the " +
                          "root window.")

        heads = xinerama.Heads{
            xrect.Make(ROOT.geom.X(), ROOT.geom.Y(),
                       ROOT.geom.Width(), ROOT.geom.Height()),
        }
    }

    return heads
}

func newState() *state {
    heads := headsLoad()

    // If for whatever reason there are no workspaces in the config,
    // make ONE and emit a warning.
    var wrks workspaces
    if len(CONF.workspaces) < len(heads) {
        logWarning.Println("There were not enough workspaces found in the " +
                           "configuration. Namely, there must be at least " +
                           "as many workspaces as there are phyiscal heads. " +
                           "We are forcefully making some and " +
                           "moving on. Please report this as a bug if you " +
                           "think you're configuration is correct.")

        for i := 0; i < len(heads); i++ {
            wrks = workspaces{
                &workspace{i, fmt.Sprintf("Default Workspace %d", i + 1), -1,
                           false},
            }
        }
    } else {
        wrks = make(workspaces, len(CONF.workspaces))
        for i, wrkName := range CONF.workspaces {
            wrks[i] = &workspace{i, wrkName, -1, false}
        }
    }

    // Make the first one active and the first N workspaces visible,
    // where N is the number of heads
    wrks[0].active = true
    for i := 0; i < len(heads); i++ {
        wrks[i].head = i
    }

    return &state{
        clients: make([]*client, 0),
        stack: make([]*client, 0),
        focus: make([]*client, 0),
        heads: heads,
        workspaces: wrks,
    }
}

func (wm *state) clientAdd(c *client) {
    if cliIndex(c, wm.clients) == -1 {
        logMessage.Println("Managing new client:", c)
        wm.clients = append(wm.clients, c)
        wm.updateEwmhClients()
    } else {
        logMessage.Println("Already managing client:", c)
    }
}

func (wm *state) clientRemove(c *client) {
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

// There can only ever be one focused client, so just find it
func (wm *state) focused() *client {
    for _, client := range wm.clients {
        if client.state == StateActive {
            return client
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

func (wm *state) focusAdd(c *client) {
    wm.focusRemove(c)
    wm.focus = append(wm.focus, c)
}

func (wm *state) focusRemove(c *client) {
    if i := cliIndex(c, wm.focus); i > -1 {
        wm.focus = append(wm.focus[:i], wm.focus[i+1:]...)
    }
}

func (wm *state) fallback() {
    var c *client
    for i := len(wm.focus) - 1; i >= 0; i-- {
        c = wm.focus[i]
        if c.Mapped() && c.Alive() && c.workspace == WM.WrkActiveInd() {
            logMessage.Printf("Focus falling back to %s", c)
            c.Focus()
            return
        }
    }

    // No windows to fall back on... root focus
    // this is IMPORTANT. if we fail here, we risk a lock-up
    logMessage.Printf("Focus falling back to ROOT")
    ROOT.focus()
    wm.unfocusExcept(0)
}

func (wm *state) logClientList() {
    list := make([]string, len(wm.clients))
    for i, c := range wm.clients {
        list[i] = c.String()
    }
    logMessage.Println(strings.Join(list, ", "))
}

