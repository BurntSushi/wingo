package wm

import (
	"fmt"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/shape"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo-conc/event"
	"github.com/BurntSushi/wingo-conc/focus"
	"github.com/BurntSushi/wingo-conc/heads"
	"github.com/BurntSushi/wingo-conc/logger"
	"github.com/BurntSushi/wingo-conc/workspace"
)

var (
	X          *xgbutil.XUtil
	Root       *xwindow.Window
	Startup    bool
	Clients    ClientList
	Heads      *heads.Heads
	Prompts    AllPrompts
	Config     *Configuration
	Theme      *ThemeConfig
	StickyWrk  *workspace.Sticky
	gribbleEnv *gribble.Environment
	cmdHacks   CommandHacks
	ShapeExt   bool
	Restart    bool
)

func Initialize(x *xgbutil.XUtil,
	cmdEnv *gribble.Environment, hacks CommandHacks) {

	var err error

	X = x
	Startup = true

	gribbleEnv = cmdEnv
	cmdHacks = hacks

	Root = xwindow.New(X, X.RootWin())
	if _, err = Root.Geometry(); err != nil {
		logger.Error.Fatalf("Could not get ROOT window geometry: %s", err)
	}

	if Config, err = loadConfig(); err != nil {
		logger.Error.Fatalf("Could not load configuration: %s", err)
	}
	if Theme, err = loadTheme(); err != nil {
		logger.Error.Fatalf("Could not load theme: %s", err)
	}

	Clients = make(ClientList, 0, 50)
	Prompts = newPrompts()

	Heads = heads.NewHeads(X, Config.DefaultLayout)

	// If _NET_DESKTOP_NAMES is set, let's use workspaces from that instead.
	if names, _ := ewmh.DesktopNamesGet(X); len(names) > 0 {
		for _, wrkName := range names {
			if err := AddWorkspace(wrkName); err != nil {
				logger.Warning.Printf("Could not add workspace %s: %s",
					wrkName, err)
			}
		}
	} else {
		for _, wrkName := range Config.Workspaces {
			if err := AddWorkspace(wrkName); err != nil {
				logger.Error.Fatalf("Could not initialize workspaces: %s", err)
			}
		}
	}
	Heads.Initialize(Clients)

	keybindings()
	rootMouseSetup()

	StickyWrk = Heads.Workspaces.NewSticky()

	err = shape.Init(X.Conn())
	if err != nil {
		ShapeExt = false
		logger.Warning.Printf("The X SHAPE extension could not be loaded. " +
			"Google Chrome might look ugly.")
	} else {
		ShapeExt = true
	}

	Restart = false

	ewmhClientList()
	ewmhNumberOfDesktops()
	ewmhCurrentDesktop()
	ewmhVisibleDesktops()
	ewmhDesktopNames()
	ewmhDesktopGeometry()
}

func AddClient(c Client) {
	if cliIndex(c, Clients) != -1 {
		panic("BUG: Cannot add client that is already managed.")
	}
	Clients = append(Clients, c)

	ewmhClientList()
}

func RemoveClient(c Client) {
	if i := cliIndex(c, Clients); i > -1 {
		Clients = append(Clients[:i], Clients[i+1:]...)
	}

	ewmhClientList()
}

func FindManagedClient(id xproto.Window) Client {
	for _, client := range Clients {
		if client.Id() == id {
			return client
		}
	}
	return nil
}

func FocusFallback() {
	focus.Fallback(focusable)
}

func LastFocused() Client {
	if c := focus.LastFocused(focusable); c != nil {
		return c.(Client)
	}
	return nil
}

func focusable(client focus.Client) bool {
	c := client.(Client)
	wrk := Workspace()
	return c.IsMapped() &&
		(c.Workspace() == wrk || c.Workspace() == StickyWrk) &&
		!c.ImminentDestruction()
}

func Workspace() *workspace.Workspace {
	return Heads.ActiveWorkspace()
}

func SetWorkspace(wrk *workspace.Workspace, greedy bool) {
	old := Workspace()
	wrk.Activate(greedy)
	if old != Workspace() {
		FYI("%s", wrk)
	}

	ewmhVisibleDesktops()
	ewmhCurrentDesktop()
	Heads.EwmhWorkarea()
}

func WorkspaceToHead(headIndex int, wrk *workspace.Workspace) {
	if headIndex == Heads.VisibleIndex(wrk) {
		return
	}

	// If headIndex is the currently active head, then just activate 'wrk'
	// greedily.
	if headIndex == Heads.VisibleIndex(Workspace()) {
		SetWorkspace(wrk, true)
		return
	}

	// Now we know that we're setting the workspace of a head that isn't
	// active. The last special case to check for is whether the workspace
	// we're setting is the currently activate workspace. If it is, then we
	// simply activate the workspace at headIndex greedily.
	if wrk.IsActive() {
		Heads.WithVisibleWorkspace(headIndex, func(w *workspace.Workspace) {
			SetWorkspace(w, true)
		})
		return
	}

	// Finally, we can just swap workspaces now without worrying about the
	// active workspace changing.
	Heads.WithVisibleWorkspace(headIndex, func(w *workspace.Workspace) {
		Heads.SwitchWorkspaces(wrk, w)
	})
	ewmhVisibleDesktops()
	ewmhCurrentDesktop()
	Heads.EwmhWorkarea()
}

func AddWorkspace(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("workspaces must have a name of length at least one.")
	}
	if Heads.Workspaces.Find(name) != nil {
		return fmt.Errorf("a workspace with name '%s' already exists.", name)
	}
	wrk := Heads.NewWorkspace(name)
	wrk.PromptSlctGroup = Prompts.Slct.AddGroup(wrk)
	wrk.PromptSlctItem = Prompts.Slct.AddChoice(wrk)

	Heads.AddWorkspace(wrk)

	ewmhDesktopNames()
	ewmhNumberOfDesktops()
	ewmhVisibleDesktops()
	Heads.EwmhWorkarea()

	event.Notify(event.AddedWorkspace{wrk.Name})
	return nil
}

func RemoveWorkspace(wrk *workspace.Workspace) error {
	if len(Heads.Workspaces.Wrks) == Heads.NumHeads() {
		return fmt.Errorf("Cannot have fewer workspaces than active monitors.")
	}
	if len(wrk.Clients) > 0 {
		return fmt.Errorf("Non-empty workspace '%s' cannot be removed.", wrk)
	}
	Heads.RemoveWorkspace(wrk)

	ewmhDesktopNames()
	ewmhNumberOfDesktops()
	ewmhVisibleDesktops()
	Heads.EwmhWorkarea()
	event.Notify(event.RemovedWorkspace{wrk.Name})
	return nil
}

func RenameWorkspace(wrk *workspace.Workspace, newName string) error {
	if len(newName) == 0 {
		return fmt.Errorf("workspaces must have a name of length at least one.")
	}
	if Heads.Workspaces.Find(newName) != nil {
		return fmt.Errorf("a workspace with name '%s' already exists.", newName)
	}
	wrk.Rename(newName)

	ewmhDesktopNames()
	return nil
}

func RootGeomChangeFun() xevent.ConfigureNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		// Before trying to reload, make sure we have enough workspaces...
		// We don't want to die here like we might on start up.
		for i := len(Heads.Workspaces.Wrks); i < Heads.NumConnected(); i++ {
			AddWorkspace(uniqueWorkspaceName())
		}
		Heads.Reload(Clients)
		FocusFallback()
		ewmhVisibleDesktops()
		ewmhDesktopGeometry()
	}
	return xevent.ConfigureNotifyFun(f)
}

// uniqueWorkspaceName returns a workspace name that is guaranteed to be of
// non-zero length and unique with respect to all other workspaces.
func uniqueWorkspaceName() string {
	// Simple... try "1", "2", ... until we get a unique workspace.
	for i := 1; true; i++ {
		try := fmt.Sprintf("%d", i)
		if Heads.Workspaces.Find(try) == nil {
			return try
		}
	}
	panic("unreachable")
}

// cliIndex returns the index of the first occurrence of needle in haystack.
// Returns -1 if needle is not in haystack.
func cliIndex(needle Client, haystack []Client) int {
	for i, possible := range haystack {
		if needle.Id() == possible.Id() {
			return i
		}
	}
	return -1
}
