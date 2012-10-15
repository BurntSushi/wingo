package main

import (
	"fmt"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
)

func handleClientMessages(X *xgbutil.XUtil, ev xevent.ClientMessageEvent) {
	name, err := xprop.AtomName(X, ev.Type)
	if err != nil {
		logger.Warning.Printf("Could not get atom name for '%s': %s", ev, err)
		return
	}
	switch name {
	case "_NET_NUMBER_OF_DESKTOPS":
		logger.Warning.Printf("Wingo does not support adding/removing " +
			"desktops using the _NET_NUMBER_OF_DESKTOPS property. Please use " +
			"the Wingo commands 'AddWorkspace' and 'RemoveWorkspace' to add " +
			"or remove workspaces.")
	case "_NET_DESKTOP_GEOMETRY":
		logger.Warning.Printf("Wingo does not support the " +
			"_NET_DESKTOP_GEOMETRY property. Namely, more than one workspace " +
			"can be visible at a time, so different workspaces can have " +
			"different geometries.")
	case "_NET_DESKTOP_VIEWPORT":
		logger.Warning.Printf("Wingo does not use viewports, and therefore " +
			"does not support the _NET_DESKTOP_VIEWPORT property.")
	case "_NET_CURRENT_DESKTOP":
		index := int(ev.Data.Data32[0])
		if wrk := wm.Heads.Workspaces.Get(index); wrk != nil {
			wm.SetWorkspace(wrk, false)
			wm.FocusFallback()
		} else {
			logger.Warning.Printf("Desktop index %d is not in the range "+
				"[0, %d).", index, len(wm.Heads.Workspaces.Wrks))
		}
	}

	fmt.Println(name, ev)
}

func ignoreRootFocus(modeByte, detailByte byte) bool {
	mode, detail := focus.Modes[modeByte], focus.Details[detailByte]

	if mode == "NotifyGrab" || mode == "NotifyUngrab" {
		return true
	}
	if detail == "NotifyAncestor" ||
		detail == "NotifyInferior" ||
		detail == "NotifyVirtual" ||
		detail == "NotifyNonlinear" ||
		detail == "NotifyNonlinearVirtual" ||
		detail == "NotifyPointer" {

		return true
	}
	// Only accept modes: NotifyNormal and NotifyWhileGrabbed
	// Only accept details: NotifyPointerRoot, NotifyNone
	return false
}
