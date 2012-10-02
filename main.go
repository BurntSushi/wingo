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

	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
)

// global variables!
var (
	X     *xgbutil.XUtil
	wingo *wingoState
)

func main() {
	var err error

	// giggity
	runtime.GOMAXPROCS(runtime.NumCPU())

	X, err = xgbutil.NewConn()
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

	wingo = newWingoState()

	// Create a root window abstraction and load its geometry
	wingo.root = xwindow.New(X, X.RootWin())
	_, err = wingo.root.Geometry()
	if err != nil {
		logger.Error.Printf("Could not get ROOT window geometry because: %v\n",
			err)
		logger.Error.Println("Cannot continue. Quitting...")
		return
	}

	// Load configuration
	wingo.conf, err = loadConfig()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No configuration found. Quitting...")
		return
	}

	// Load theme
	wingo.theme, err = loadTheme(X)
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No theme configuration found. Quitting...")
		return
	}

	// Initialize prompts
	wingo.prompts = newPrompts()

	wingo.initializeHeads()

	// Attach all global key bindings
	attachAllKeys()

	// Attach all root mouse bindings
	rootMouseConfig()

	// Setup some cursors we use
	cursors.Setup(X)

	// Listen to Root. It is all-important.
	wingo.root.Listen(xproto.EventMaskPropertyChange |
		xproto.EventMaskFocusChange |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect)

	// Update state when the root window changes size
	// xevent.ConfigureNotifyFun(rootGeometryChange).Connect(X, wingo.root.Id) 

	// Oblige map request events
	xevent.MapRequestFun(
		func(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
			newClient(X, ev.Window)
		}).Connect(X, wingo.root.Id)

	// Oblige configure requests from windows we don't manage.
	xevent.ConfigureRequestFun(
		func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
			flags := int(ev.ValueMask) &
				^int(xproto.ConfigWindowSibling) &
				^int(xproto.ConfigWindowStackMode)
			xwindow.New(X, ev.Window).Configure(flags,
				int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height),
				ev.Sibling, ev.StackMode)
		}).Connect(X, wingo.root.Id)

	xevent.FocusInFun(
		func(X *xgbutil.XUtil, ev xevent.FocusInEvent) {
			if ignoreRootFocus(ev.Mode, ev.Detail) {
				return
			}
			if len(wingo.workspace().Clients) == 0 {
				return
			}
			wingo.focusFallback()
		}).Connect(X, wingo.root.Id)

	// Listen to Root client message events.
	// We satisfy EWMH with these AND it also provides a mechanism
	// to issue commands using wingo-cmd.
	// xevent.ClientMessageFun(commandHandler).Connect(X, wingo.root.Id) 

	xevent.Main(X)
}

func set_supported() {
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
	ewmh.SupportedSet(X, supported)
}
