package wm

import (
	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/workspace"
)

var (
	X *xgbutil.XUtil
	Root *xwindow.Window
	Clients ClientList
	Heads *heads.Heads
	Prompts AllPrompts
	Config *Configuration
	Theme *ThemeConfig
	gribbleEnv *gribble.Environment
	cmdHacks CommandHacks
)

func Init(x *xgbutil.XUtil, cmdEnv *gribble.Environment, hacks CommandHacks) {
	var err error

	X = x

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
		AddWorkspace(wrkName)
	}
	Heads.Initialize(Clients)

	keybindings()
	rootMouseSetup()
}

func AddClient(c Client) {
	if cliIndex(c, Clients) != -1 {
		panic("BUG: Cannot add client that is already managed.")
	}
	Clients = append(Clients, c)
}

func RemoveClient(c Client) {
	if i := cliIndex(c, Clients); i > -1 {
		Clients = append(Clients[:i], Clients[i+1:]...)
	}
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
	focus.Fallback(Workspace())
}

func Workspace() *workspace.Workspace {
	return Heads.ActiveWorkspace()
}

func AddWorkspace(name string) {
	wrk := Heads.NewWorkspace(name)
	wrk.PromptSlctGroup = Prompts.Slct.AddGroup(wrk)
	wrk.PromptSlctItem = Prompts.Slct.AddChoice(wrk)

	Heads.AddWorkspace(wrk)
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
