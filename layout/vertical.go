package layout

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type Vertical struct {
	store *store
	proportion float64
}

func NewVertical() *Vertical {
	return &Vertical{
		store: newStore(),
		proportion: 0.5,
	}
}

func (v *Vertical) Place(geom xrect.Rect) {
	if geom == nil {
		return
	}

	msize, ssize := len(v.store.masters), len(v.store.slaves)
	mx, mw := geom.X(), int(float64(geom.Width()) * v.proportion)
	sx, sw := mx+mw, geom.Width() - mw

	// If there are zero widths or they are too big, don't do anything.
	if mw <= 0 || mw > geom.Width() || sw <= 0 || sw > geom.Width() {
		return
	}

	if msize > 0 {
		mh := geom.Height() / msize
		if ssize == 0 {
			mw = geom.Width()
		}
		for i, item := range v.store.masters {
			item.client.FrameTile()
			item.client.MoveResize(false, mx, geom.Y() + i*mh, mw, mh)
		}
	}
	if ssize > 0 {
		if msize == 0 {
			sx, sw = geom.X(), geom.Width()
		}
		sy := geom.Y()
		for _, item := range v.store.slaves {
			sh := int(float64(geom.Height()) * item.proportion)
			item.client.FrameTile()
			item.client.MoveResize(false, sx, sy, sw, sh)
			sy += sh
		}
	}
}

func (v *Vertical) Unplace(geom xrect.Rect) {
	if geom == nil {
		return
	}
}

func (v *Vertical) Exists(c Client) bool {
	return v.store.mFindClient(c) >= 0 || v.store.sFindClient(c) >= 0
}

func (v *Vertical) Add(c Client) {
	v.store.add(c)
}

func (v *Vertical) Remove(c Client) {
	v.store.remove(c)
}

func (v *Vertical) MoveResize(c Client, x, y, width, height int) {
	c.MoveResize(true, x, y, width, height)
}

func (v *Vertical) Move(c Client, x, y int) {
	c.Move(x, y)
}

func (v *Vertical) Resize(c Client, width, height int) {
	c.Resize(true, width, height)
}
