package commands

import (
	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/wm"
)

type GetWorkspace struct {}

func (cmd GetWorkspace) Run() gribble.Value {
	return wm.Workspace().Name
}

type GetWorkspaceNext struct{}

func (cmd GetWorkspaceNext) Run() gribble.Value {
	return wm.Heads.NextWorkspace().Name
}

type GetWorkspacePrefix struct {
	Prefix string `param:"1"`
	Help   string `
Some documentation.
`
}

func (cmd GetWorkspacePrefix) Run() gribble.Value {
	return nil
}

type GetWorkspacePrev struct{}

func (cmd GetWorkspacePrev) Run() gribble.Value {
	return wm.Heads.PrevWorkspace().Name
}
