package main

import (
	"runtime"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/commands"
	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

func main() {
	// giggity
	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1) 

	X, err := xgbutil.NewConn()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("Error connecting to X, quitting...")
		return
	}
	defer X.Conn().Close()

	keybind.Initialize(X)
	mousebind.Initialize(X)
	focus.Initialize(X)
	stack.Initialize(X)

	wm.Init(X, commands.Env, newHacks())

	// Setup some cursors we use
	cursors.Setup(X)

	// Listen to Root. It is all-important.
	wm.Root.Listen(xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect)

	// Update state when the root window changes size
	// xevent.ConfigureNotifyFun(rootGeometryChange).Connect(X, wm.Root.Id) 

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

	// Listen to Root client message events.
	// We satisfy EWMH with these AND it also provides a mechanism
	// to issue commands using wingo-cmd.
	// xevent.ClientMessageFun(commandHandler).Connect(X, wm.Root.Id) 

	pingBefore, pingAfter, pingQuit := xevent.MainPing(X)
	for {
		select {
		case <-pingBefore:
			// Wait for the event to finish processing.
			<-pingAfter
		case f := <-commands.SafeExec:
			commands.SafeReturn <- f()
		case <-pingQuit:
			return
		}
	}
}

func setSupported() {
	// Purposefully NOT supported
	// _NET_DESKTOP_GEOMETRY, _NET_DESKTOP_VIEWPORT, _NET_VIRTUAL_ROOTS
	// _NET_WORKAREA

	// Purposefully PARTIALLY supported
	// _NET_NUMBER_OF_DESKTOPS
	//	Read support only. Wingo ignores client messages to add/remove
	//	desktops.
	//	Wingo provides facilities to add/remove any desktop using commands.

	// Some day...
	// _NET_DESKTOP_LAYOUT, _NET_SHOWING_DESKTOP

	// breadcrumb: _NET_ACTIVE_WINDOW, _NET_WORKAREA,
	// _NET_SUPPORTING_WM_CHECK, 

	supported := []string{
		"_NET_SUPPORTED", "_NET_CLIENT_LIST", "_NET_CLIENT_LIST_STACKING",
		"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_GEOMETRY",
		"_NET_CURRENT_DESKTOP", "_NET_DESKTOP_NAMES", "_NET_ACTIVE_WINDOW",

		"_NET_WM_ICON",
	}
	// Set supported atoms
	ewmh.SupportedSet(wm.X, supported)
}
