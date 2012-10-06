package main

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
)

// This file is responsible for implementing all the crud around getting
// and giving up the window manager's selection ownership. This is effectively
// how one window manager is able to "replace" another one. Aren't we so
// cooperative?

// announce sends a ClientMessage event to the root window to let everyone
// know that Wingo is the boss. (As per ICCCM 2.8.)
func announce() {
	typAtom, err := xprop.Atm(wm.X, "MANAGER")
	if err != nil {
		logger.Warning.Println(err)
		return
	}
	manSelAtom, err := managerAtom()
	if err != nil {
		logger.Warning.Println(err)
		return
	}
	cm, err := xevent.NewClientMessage(32, wm.X.RootWin(), typAtom,
		int(wm.X.TimeGet()), int(manSelAtom), int(wm.X.Dummy()))
	xproto.SendEvent(wm.X.Conn(), false, wm.X.RootWin(),
		xproto.EventMaskStructureNotify, string(cm.Bytes()))
}

// managerAtom returns an xproto.Atom of the manager selection atom.
// Usually it's "WM_S0", where "0" is the screen number.
func managerAtom() (xproto.Atom, error) {
	name := fmt.Sprintf("WM_S%d", wm.X.Conn().DefaultScreen)
	return xprop.Atm(wm.X, name)
}
