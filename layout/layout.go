package layout

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

const (
	FloatFloating = iota
)

const (
	AutoTileVertical = iota
)

type Layout interface {
	Name() string
	SetGeom(geom xrect.Rect)
	Place()
	Unplace()
	Add(c Client)
	Remove(c Client)
	Exists(c Client) bool
	Destroy()

	MROpt(c Client, flags, x, y, width, height int)
	MoveResize(c Client, x, y, width, height int)
	Move(c Client, x, y int)
	Resize(c Client, width, height int)
}

type Floater interface {
	Layout
	InitialPlacement(c Client)
	Save()
	Reposition()
}

type AutoTiler interface {
	Layout
	ResizeMaster(amount float64)
	ResizeWindow(amount float64)
	Next()
	Prev()
	SwitchNext()
	SwitchPrev()
	FocusMaster()
	MakeMaster()
	MastersMore()
	MastersFewer()
}
