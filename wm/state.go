package wm

import (
	"fmt"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/workspace"
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

	Heads = heads.NewHeads(X)
	for _, wrkName := range Config.Workspaces {
		if err := AddWorkspace(wrkName); err != nil {
			logger.Error.Fatalf("Could not initialize workspaces: %s", err)
		}
	}
	Heads.Initialize(Clients)

	keybindings()
	rootMouseSetup()

	StickyWrk = Heads.Workspaces.NewSticky()

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

	ewmhCurrentDesktop()
	ewmhVisibleDesktops()
	Heads.EwmhWorkarea()
}

func AddWorkspace(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("Workspaces must have a name of length at least one.")
	}
	if Heads.Workspaces.Find(name) != nil {
		return fmt.Errorf("A workspace with name '%s' already exists.", name)
	}
	wrk := Heads.NewWorkspace(name)
	wrk.PromptSlctGroup = Prompts.Slct.AddGroup(wrk)
	wrk.PromptSlctItem = Prompts.Slct.AddChoice(wrk)

	Heads.AddWorkspace(wrk)

	ewmhDesktopNames()
	ewmhNumberOfDesktops()
	ewmhVisibleDesktops()
	Heads.EwmhWorkarea()
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
