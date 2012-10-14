package commands

import (
	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/wm"
)

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
