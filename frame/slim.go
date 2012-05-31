package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"

	"github.com/BurntSushi/wingo/theme"
)

type Slim struct {
	*frame
}

func NewSlim(X *xgbutil.XUtil,
	t *theme.Theme, p *Parent, c client) (*Slim, error) {

	f, err := newFrame(X, t, p, c)
	if err != nil {
		return nil, err
	}
	return &Slim{f}, nil
}

func (f *Slim) Current() bool {
	return f.client.Frame() == f
}

func (f *Slim) Off() {}

func (f *Slim) On() {
	Reset(f)

	if f.client.State() == Active {
		f.Active()
	} else {
		f.Inactive()
	}
}

func (f *Slim) Active() {
	f.parent.Change(xproto.CwBackPixel, uint32(f.theme.Slim.ABorderColor))
	f.parent.ClearAll()
}

func (f *Slim) Inactive() {
	f.parent.Change(xproto.CwBackPixel, uint32(f.theme.Slim.IBorderColor))
	f.parent.ClearAll()
}

func (f *Slim) Maximize() {}
func (f *Slim) Unmaximize() {}

func (f *Slim) Top() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Slim.BorderSize
}

func (f *Slim) Bottom() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Slim.BorderSize
}

func (f *Slim) Left() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Slim.BorderSize
}

func (f *Slim) Right() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Slim.BorderSize
}

func (f *Slim) MROpt(validate bool, flags, x, y, w, h int) {
	mropt(f, validate, flags, x, y, w, h)
}

func (f *Slim) MoveResize(validate bool, x, y, w, h int) {
	moveresize(f, validate, x, y, w, h)
}

func (f *Slim) Move(x, y int) {
	move(f, x, y)
}

func (f *Slim) Resize(validate bool, w, h int) {
	resize(f, validate, w, h)
}
