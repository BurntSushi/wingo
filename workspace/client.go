package workspace

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/layout"
)

type Clients interface {
	Get(i int) Client
	Len() int
}

type Client interface {
	Id() xproto.Window
	String() string
	Workspace() *Workspace
	WorkspaceSet(wrk *Workspace)
	Layout() layout.Layout
	Map()
	Unmap()
	UnmapFallback()
	ForceFloating() bool

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
