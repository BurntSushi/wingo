package commands

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

type GetActive struct {
	Help string `
Returns the id of the currently active window. If there is no active window,
0 is returned.
`
}

func (cmd GetActive) Run() gribble.Value {
	reply, err := xproto.GetInputFocus(wm.X.Conn()).Reply()
	if err != nil {
		logger.Warning.Printf("Could not get input focus: %s", err)
		return 0
	}

	// If our dummy window has focus, then it's equivalent to having root
	// window focus.
	// XXX: This may not be right if we're in a DE with desktop windows...
	if reply.Focus == wm.X.Dummy() {
		return 0
	}
	if focused := wm.LastFocused(); focused != nil {
		client := focused.(*xclient.Client)
		return int(client.Id())
	}
	return 0
}

type GetWorkspace struct {
	Help string `
Returns the name of the current workspace.
`
}

func (cmd GetWorkspace) Run() gribble.Value {
	return wm.Workspace().Name
}

type GetWorkspaceNext struct {
	Help string `
Returns the name of the "next" workspace. The ordering of workspaces is
the order in which they were added. This might cause confusing behavior in
multi-head setups, since multiple workspaces can be viewable at one time.
`
}

func (cmd GetWorkspaceNext) Run() gribble.Value {
	return wm.Heads.NextWorkspace().Name
}

type GetWorkspacePrefix struct {
	Prefix string `param:"1"`
	Help   string `
Returns the first workspace starting with Prefix. If the current workspace
starts with Prefix, then the first workspace *after* the current workspace
starting with Prefix will be returned.
`
}

func (cmd GetWorkspacePrefix) Run() gribble.Value {
	return nil
}

type GetWorkspacePrev struct {
	Help string `
Returns the name of the "previous" workspace. The ordering of workspaces is
the order in which they were added. This might cause confusing behavior in
multi-head setups, since multiple workspaces can be viewable at one time.
`
}

func (cmd GetWorkspacePrev) Run() gribble.Value {
	return wm.Heads.PrevWorkspace().Name
}
