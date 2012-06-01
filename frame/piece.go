package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type piece struct {
	*xwindow.Window
	active, inactive xproto.Pixmap
}

func newPiece(w *xwindow.Window, active, inactive xproto.Pixmap) piece {
	return piece{
		Window:   w,
		active:   active,
		inactive: inactive,
	}
}

func (p piece) Destroy() {
	p.Window.Destroy() // detaches all event handlers
	xgraphics.FreePixmap(p.X, p.active)
	xgraphics.FreePixmap(p.X, p.inactive)
}

func (p piece) Active() {
	p.Change(xproto.CwBackPixmap, uint32(p.active))
	p.ClearAll()
}

func (p piece) Inactive() {
	p.Change(xproto.CwBackPixmap, uint32(p.inactive))
	p.ClearAll()
}

func (p piece) x() int {
	return p.Geom.X()
}

func (p piece) y() int {
	return p.Geom.Y()
}

func (p piece) w() int {
	return p.Geom.Width()
}

func (p piece) h() int {
	return p.Geom.Height()
}
