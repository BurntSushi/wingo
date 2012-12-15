package wm

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/workspace"
)

type Client interface {
	Id() xproto.Window
	Frame() frame.Frame
	IsMapped() bool
	Iconified() bool
	IsSkipPager() bool
	IsSkipTaskbar() bool
	Workspace() workspace.Workspacer
	ImminentDestruction() bool
	IsMaximized() bool
	Remaximize()

	CycleItem() *prompt.CycleItem
	SelectItem() *prompt.SelectItem

	DragMoveBegin(rx, ry, ex, ey int) bool
	DragMoveStep(rx, ry, ex, ey int)
	DragMoveEnd(rx, ry, ex, ey int)

	DragResizeBegin(direction uint32, rx, ry, ex, ey int) (bool, xproto.Cursor)
	DragResizeStep(rx, ry, ex, ey int)
	DragResizeEnd(rx, ry, ex, ey int)
}

type ClientList []Client

func (cls ClientList) Get(i int) heads.Client {
	return cls[i]
}

func (cls ClientList) Len() int {
	return len(cls)
}
