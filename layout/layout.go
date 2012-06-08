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

	MoveResize(c Client, x, y, width, height int)
	Move(c Client, x, y int)
	Resize(c Client, width, height int)
}

type Floater interface {
	Layout
}

type AutoTiler interface {
	Layout
}
