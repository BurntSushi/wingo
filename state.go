package main

import (
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xinerama"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
)

// state is the master singleton the carries all window manager related state
type state struct {
	clients    []*client // a list of clients in order of being added
	stack      []*client // clients ordered by visual stack
	focus      []*client // focus order of clients; may be smaller than clients
	headsRaw   xinerama.Heads
	heads      xinerama.Heads
	workspaces workspaces
	stickyWrk  *workspace
}

func newState() *state {
	wm := &state{
		clients:    make([]*client, 0),
		stack:      make([]*client, 0),
		focus:      make([]*client, 0),
		heads:      nil,
		workspaces: make(workspaces, 0, len(CONF.workspaces)),
		stickyWrk:  nil,
	}

	// Add the special workspace that holds windows that are always visible
	wm.stickyWrk = newWorkspace(-1)
	wm.stickyWrk.nameSet("Sticky")

	for i, wrkName := range CONF.workspaces {
		wm.workspaceAdd(i, wrkName)
	}

	return wm
}

func (wm *state) workspaceAdd(id int, name string) {
	wrk := newWorkspace(id)
	wrk.nameSet(name)
	wm.workspaces = append(wm.workspaces, wrk)

	ewmh.NumberOfDesktopsSet(X, len(wm.workspaces))
	wm.ewmhDesktopNames()
}

func (wm *state) clientAdd(c *client) {
	if cliIndex(c, wm.clients) == -1 {
		logger.Message.Println("Managing new client:", c)
		wm.clients = append(wm.clients, c)
		wm.updateEwmhClients()
	} else {
		logger.Message.Println("Already managing client:", c)
	}
}

func (wm *state) clientRemove(c *client) {
	if i := cliIndex(c, wm.clients); i > -1 {
		logger.Message.Println("Unmanaging client:", c)
		wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
		wm.updateEwmhClients()
	}
}

func (wm *state) updateEwmhClients() {
	numWins := len(wm.clients)
	winList := make([]xproto.Window, numWins)
	for i, c := range wm.clients {
		winList[i] = c.Id()
	}
	err := ewmh.ClientListSet(X, winList)
	if err != nil {
		logger.Warning.Printf("Could not update _NET_CLIENT_LIST "+
			"because %v", err)
	}
}

// There can only ever be one focused client, so just find it
func (wm *state) focused() *client {
	for _, client := range wm.clients {
		if client.normal && client.state == frame.Active {
			return client
		}
	}
	return nil
}

func (wm *state) unfocusExcept(id xproto.Window) {
	// Go in reverse to make switching appear quicker in the common case
	// if there are a lot of windows.
	for i := len(wm.focus) - 1; i >= 0; i-- {
		if wm.focus[i].Id() != id {
			wm.focus[i].Unfocused()
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
		if c.Mapped() && c.Alive() && c.workspace.id == WM.wrkActive().id {
			logger.Message.Printf("Focus falling back to %s", c)
			c.Focus()
			return
		}
	}

	// No windows to fall back on... root focus
	// this is IMPORTANT. if we fail here, we risk a lock-up
	logger.Message.Printf("Focus falling back to ROOT")
	ROOT.Focus()
	ewmh.ActiveWindowSet(X, 0x0)
	wm.unfocusExcept(0)
}

func (wm *state) logClientList() {
	list := make([]string, len(wm.clients))
	for i, c := range wm.clients {
		list[i] = c.String()
	}
	logger.Message.Println(strings.Join(list, ", "))
}
