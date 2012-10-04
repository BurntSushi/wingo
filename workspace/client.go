package workspace

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xrect"

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
	Geom() xrect.Rect
	DragGeom() xrect.Rect

	Iconified() bool
	IconifiedSet(iconified bool)

	HasState(name string) bool
	SaveState(name string)
	CopyState(src, dest string)
	LoadState(name string)
	DeleteState(name string)

	MROpt(validate bool, flags, x, y, width, height int)
	MoveResize(validate bool, x, y, width, height int)
	Move(x, y int)
	Resize(validate bool, width, height int)

	FrameTile()
}
