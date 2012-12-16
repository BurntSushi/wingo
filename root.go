package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

func rootInit(X *xgbutil.XUtil) {
	var err error

	// Listen to Root. It is all-important.
	evMasks := xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect
	if wm.Config.FfmHead {
		evMasks |= xproto.EventMaskPointerMotion
	}
	err = xwindow.New(X, X.RootWin()).Listen(evMasks)
	if err != nil {
		logger.Error.Fatalf("Could not listen to Root window events: %s", err)
	}

	// Update state when the root window changes size
	wm.RootGeomChangeFun().Connect(X, wm.Root.Id)

	// Oblige map request events
	xevent.MapRequestFun(
		func(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
			xclient.New(ev.Window)
		}).Connect(X, wm.Root.Id)

	// Oblige configure requests from windows we don't manage.
	xevent.ConfigureRequestFun(
		func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
			// Make sure we aren't managing this client.
			if wm.FindManagedClient(ev.Window) != nil {
				return
			}

			xwindow.New(X, ev.Window).Configure(int(ev.ValueMask),
				int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height),
				ev.Sibling, ev.StackMode)
		}).Connect(X, wm.Root.Id)

	xevent.FocusInFun(
		func(X *xgbutil.XUtil, ev xevent.FocusInEvent) {
			if ignoreRootFocus(ev.Mode, ev.Detail) {
				return
			}
			if len(wm.Workspace().Clients) == 0 {
				return
			}
			wm.FocusFallback()
		}).Connect(X, wm.Root.Id)

	// Listen to Root client message events. This is how we handle all
	// of the EWMH bullshit.
	xevent.ClientMessageFun(handleClientMessages).Connect(X, wm.Root.Id)

	// Check where the pointer is on motion events. If it's crossed a monitor
	// boundary, switch the focus of the head.
	if wm.Config.FfmHead {
		xevent.MotionNotifyFun(handleMotionNotify).Connect(X, wm.Root.Id)
	}
}

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
	default:
		logger.Warning.Printf("Unknown root client message: %s", name)
	}
}

func handleMotionNotify(X *xgbutil.XUtil, ev xevent.MotionNotifyEvent) {
	qp, err := xproto.QueryPointer(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		logger.Warning.Printf("Could not query pointer: %s", err)
		return
	}

	geom := xrect.New(int(qp.RootX), int(qp.RootY), 1, 1)
	if wrk := wm.Heads.FindMostOverlap(geom); wrk != nil {
		if wrk != wm.Workspace() {
			wm.SetWorkspace(wrk, false)
			wm.FocusFallback()
		}
	}
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
