package workspace

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/cshapeshifter/wingo/layout"
)

type Client interface {
	Id() xproto.Window
	String() string
	Workspace() Workspacer
	WorkspaceSet(wrk Workspacer)
	Layout() layout.Layout
	Map()
	Unmap()
	ShouldForceFloating() bool
	Focus()
	Raise()
	Geom() xrect.Rect
	DragGeom() xrect.Rect

	Iconified() bool
	IconifiedSet(iconified bool)
	IsSticky() bool
	IsActive() bool

	HasState(name string) bool
	SaveState(name string)
	CopyState(src, dest string)
	LoadState(name string)
	DeleteState(name string)

	MROpt(validate bool, flags, x, y, width, height int)
	MoveResize(x, y, width, height int)
	MoveResizeValid(x, y, width, height int)
	Move(x, y int)
	Resize(validate bool, width, height int)

	FrameTile()
}
