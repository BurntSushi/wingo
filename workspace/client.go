package workspace

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/layout"
)

type Client interface {
	Id() xproto.Window
	String() string
	Workspace() *Workspace
	WorkspaceSet(wrk *Workspace)
	Layout() layout.Layout
	Map()
	Unmap()
	ShouldForceFloating() bool

	Iconified() bool
	IconifiedSet(iconified bool)

	HasState(name string) bool
	SaveState(name string)
	LoadState(name string)
	DeleteState(name string)

	MoveResize(validate bool, x, y, width, height int)
	Move(x, y int)
	Resize(validate bool, width, height int)

	FrameTile()
}
