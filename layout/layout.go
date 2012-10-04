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
	Place(geom xrect.Rect)
	Unplace(geom xrect.Rect)
	Add(c Client)
	Remove(c Client)
	Exists(c Client) bool

	MROpt(c Client, flags, x, y, width, height int)
	MoveResize(c Client, x, y, width, height int)
	Move(c Client, x, y int)
	Resize(c Client, width, height int)
}

type Floater interface {
	Floater()
	Layout
	InitialPlacement(geom xrect.Rect, c Client)
	Save()
	Reposition(geom xrect.Rect)
}

type AutoTiler interface {
	AutoTiler()
	Layout
}
