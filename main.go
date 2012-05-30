package main

import (
	// "log"
	// "os"
	// "runtime/pprof"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/logger"
)

// global variables!
var (
	X       *xgbutil.XUtil
	WM      *state
	ROOT    *window
	CONF    *conf
	THEME   *theme
	PROMPTS prompts
)

func main() {
	var err error

	// f, err := os.Create("zzz.prof") 
	// if err != nil { 
	// log.Fatal(err) 
	// } 
	// pprof.StartCPUProfile(f) 
	// defer pprof.StopCPUProfile() 

	X, err = xgbutil.NewConn()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("Error connecting to X, quitting...")
		return
	}
	defer X.Conn().Close()

	// Allow key and mouse bindings to do their thang
	keybind.Initialize(X)
	mousebind.Initialize(X)

	// Create a root window abstraction and load its geometry
	ROOT = newWindow(X.RootWin())
	_, err = ROOT.geometry()
	if err != nil {
		logger.Error.Println("Could not get ROOT window geometry because: %v",
			err)
		logger.Error.Println("Cannot continue. Quitting...")
		return
	}

	// Create the _NET_SUPPORTING_WM_CHECK window.
	WM.ewmhSupportingWmCheck()

	// Load configuration
	err = loadConfig()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No configuration found. Quitting...")
		return
	}

	// Load theme
	err = loadTheme()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No theme configuration found. Quitting...")
		return
	}

	// Initialize prompts
	promptsInitialize()

	// Create WM state
	WM = newState()
	WM.headsLoad()

	// Attach all global key bindings
	attachAllKeys()

	// Attach all root mouse bindings
	rootMouseConfig()

	// Setup some cursors we use
	setupCursors()

	// Listen to Root. It is all-important.
	ROOT.listen(xproto.EventMaskPropertyChange |
		xproto.EventMaskStructureNotify |
		xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect)

	// Update state when the root window changes size
	xevent.ConfigureNotifyFun(rootGeometryChange).Connect(X, ROOT.id)

	// Oblige map request events
	xevent.MapRequestFun(clientMapRequest).Connect(X, ROOT.id)

	// Oblige configure requests from windows we don't manage.
	xevent.ConfigureRequestFun(
		func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
			flags := int(ev.ValueMask) &
				^int(xproto.ConfigWindowSibling) &
				^int(xproto.ConfigWindowStack)
			xwindow.New(ev.Window).Configure(flags,
				int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height),
				ev.Sibling, ev.StackMode)
		}).Connect(X, ROOT.id)

	// Listen to Root client message events.
	// We satisfy EWMH with these AND it also provides a mechanism
	// to issue commands using wingo-cmd.
	xevent.ClientMessageFun(commandHandler).Connect(X, ROOT.id)

	xevent.Main(X)

	// println("Writing memory profile...") 
	// f, err = os.Create("zzz.mprof") 
	// if err != nil { 
	// log.Fatal(err) 
	// } 
	// pprof.WriteHeapProfile(f) 
	// f.Close() 
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
