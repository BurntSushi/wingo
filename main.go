package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/BurntSushi/xgb"
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

var (
	flagGoMaxProcs = runtime.NumCPU()
	flagLogLevel   = 3
	flagLogColors  = true
	flagReplace    = false
)

func init() {
	flag.IntVar(&flagGoMaxProcs, "p", flagGoMaxProcs,
		"The maximum number of CPUs that can be executing simultaneously.")
	flag.IntVar(&flagLogLevel, "log-level", flagLogLevel,
		"The logging level of Wingo. Valid values are 0, 1, 2, 3, or 4.\n"+
			"Higher numbers result in Wingo producing more output.")
	flag.BoolVar(&flagLogColors, "log-colors", flagLogColors,
		"Whether to output logging data with terminal colors.")
	flag.BoolVar(&flagReplace, "replace", flagReplace,
		"When set, Wingo will attempt to replace a currently running\n"+
			"window manager. If this is not set, and another window manager\n"+
			"is running, Wingo will exit.")

	flag.Usage = usage
	flag.Parse()

	runtime.GOMAXPROCS(flagGoMaxProcs)
	logger.Colors(flagLogColors)
	logger.LevelSet(flagLogLevel)

	// If the log level is 0, don't show XGB log output either.
	if flagLogLevel == 0 {
		xgb.PrintLog = false
	}
}

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Fatalln("Error connecting to X, quitting...")
	}
	defer X.Conn().Close()

	// Do this first! Attempt to retrieve window manager ownership.
	// This includes waiting for any existing window manager to die.
	// 'own' also sets up handlers for quitting when a window manager tries
	// to replace *us*.
	if err := own(X, flagReplace); err != nil {
		logger.Error.Fatalf(
			"Could not establish window manager ownership: %s", err)
	}

	keybind.Initialize(X)
	mousebind.Initialize(X)
	focus.Initialize(X)
	stack.Initialize(X)
	cursors.Initialize(X)
	wm.Initialize(X, commands.Env, newHacks())

	// Listen to Root. It is all-important.
	err = xwindow.New(X, X.RootWin()).Listen(
		xproto.EventMaskPropertyChange |
			xproto.EventMaskFocusChange |
			xproto.EventMaskButtonPress |
			xproto.EventMaskButtonRelease |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskSubstructureNotify |
			xproto.EventMaskSubstructureRedirect)
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

	// Tell everyone what we support.
	setSupported()

	// Just before starting the main event loop, check to see if there are
	// any clients that already exist that we should manage.
	manageExistingClients()

	// Now make sure that clients are in the appropriate visible state.
	for _, wrk := range wm.Heads.Workspaces.Wrks {
		if wrk.IsVisible() {
			wrk.Show()
		} else {
			wrk.Hide()
		}
	}
	wm.Heads.ApplyStruts(wm.Clients)

	wm.FocusFallback()
	wm.Startup = false
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
	// _NET_WM_ICON_NAME, _NET_WM_VISIBLE_ICON_NAME
	// _NET_WM_STATE_MODAL, _NET_WM_STATE_SHADED, _NET_WM_STATE_FOCUSED,
	// _NET_WM_ACTION_SHADE,
	// _NET_WM_OPAQUE_REGION

	// Purposefully PARTIALLY supported
	// _NET_NUMBER_OF_DESKTOPS
	//	Read support only. Wingo ignores client messages to add/remove
	//	desktops.
	//	Wingo provides facilities to add/remove any desktop using commands.

	// Some day...
	// _NET_DESKTOP_LAYOUT, _NET_SHOWING_DESKTOP,
	// _NET_WM_ICON_GEOMETRY, _NET_WM_PID,
	// _NET_WM_HANDLED_ICONS, _NET_WM_USER_TIME,
	// _NET_WM_USER_TIME_WINDOW

	// breadcrumb: _NET_ACTIVE_WINDOW, _NET_WORKAREA,
	// _NET_SUPPORTING_WM_CHECK, 

	supported := []string{
		"_NET_SUPPORTED", "_NET_CLIENT_LIST", "_NET_CLIENT_LIST_STACKING",
		"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_GEOMETRY",
		"_NET_CURRENT_DESKTOP", "_NET_DESKTOP_NAMES", "_NET_ACTIVE_WINDOW",

		"_NET_WM_NAME", "_NET_WM_VISIBLE_NAME", "_NET_WM_DESKTOP",
		"_NET_WM_WINDOW_TYPE",
		"_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK",
		"_NET_WM_WINDOW_TYPE_TOOLBAR", "_NET_WM_WINDOW_TYPE_UTILITY",
		"_NET_WM_WINDOW_TYPE_SPLASH", "_NET_WM_WINDOW_TYPE_DIALOG",
		"_NET_WM_WINDOW_TYPE_DROPDOWN_MENU", "_NET_WM_WINDOW_TYPE_POPUP_MENU",
		"_NET_WM_WINDOW_TYPE_TOOLTIP", "_NET_WM_WINDOW_TYPE_NOTIFICATION",
		"_NET_WM_WINDOW_TYPE_COMBO", "_NET_WM_WINDOW_TYPE_DND",
		"_NET_WM_WINDOW_TYPE_NORMAL",
		"_NET_WM_STATE",
		"_NET_WM_STATE_STICKY", "_NET_WM_STATE_MAXIMIZED_VERT",
		"_NET_WM_STATE_MAXIMIZED_HORZ", "_NET_WM_STATE_SKIP_TASKBAR",
		"_NET_WM_STATE_SKIP_PAGER", "_NET_WM_STATE_HIDDEN",
		"_NET_WM_STATE_FULLSCREEN", "_NET_WM_STATE_ABOVE",
		"_NET_WM_STATE_BELOW", "_NET_WM_STATE_DEMANDS_ATTENTION",
		"_NET_WM_ALLOWED_ACTIONS",
		"_NET_WM_ACTION_MOVE", "_NET_WM_ACTION_RESIZE",
		"_NET_WM_ACTION_MINIMIZE", "_NET_WM_ACTION_STICK",
		"_NET_WM_ACTION_MAXMIZE_HORZ", "_NET_WM_ACTION_MAXIMIZE_VERT",
		"_NET_WM_ACTION_FULLSCREEN", "_NET_WM_ACTION_CHANGE_DESKTOP",
		"_NET_WM_ACTION_CLOSE", "_NET_WM_ACTION_ABOVE", "_NET_WM_ACTION_BELOW",
		"_NET_WM_STRUT_PARTIAL",
		"_NET_WM_ICON",
		"_NET_FRAME_EXTENTS",
	}

	// Set supported atoms
	ewmh.SupportedSet(wm.X, supported)

	// While we're at it, set the supporting wm hint too.
	ewmh.SupportingWmCheckSet(wm.X, wm.X.RootWin(), wm.X.Dummy())
	ewmh.SupportingWmCheckSet(wm.X, wm.X.Dummy(), wm.X.Dummy())
	ewmh.WmNameSet(wm.X, wm.X.Dummy(), "Wingo")
}

// manageExistingClients traverse the window tree and tries to manage all
// top-level clients. Clients that are not in the Unmapped state will be
// managed.
func manageExistingClients() {
	tree, err := xproto.QueryTree(wm.X.Conn(), wm.Root.Id).Reply()
	if err != nil {
		logger.Warning.Printf("Could not issue QueryTree request: %s", err)
		return
	}
	for _, potential := range tree.Children {
		// Ignore our own dummy window...
		if potential == wm.X.Dummy() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), potential).Reply()
		if err != nil {
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}
		logger.Message.Printf("Managing existing client %d", potential)
		xclient.New(potential)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [flags]\n", path.Base(os.Args[0]))
	flag.VisitAll(func(fg *flag.Flag) {
		fmt.Printf("--%s=\"%s\"\n\t%s\n", fg.Name, fg.DefValue,
			strings.Replace(fg.Usage, "\n", "\n\t", -1))
	})
	os.Exit(1)
}
