package frame

import (
	"github.com/BurntSushi/xgbutil"
)

type Nada struct {
	*frame
}

func NewNada(X *xgbutil.XUtil, p *Parent, c Client) (*Nada, error) {
	f, err := newFrame(X, p, c)
	if err != nil {
		return nil, err
	}
	return &Nada{f}, nil
}

func (f *Nada) Current() bool {
	return f.client.Frame() == f
}

func (f *Nada) Off() {}

func (f *Nada) On() {
	Reset(f)
}

func (f *Nada) Active() {
	f.State = Active
}

func (f *Nada) Inactive() {
	f.State = Inactive
}

func (f *Nada) Maximize()   {}
func (f *Nada) Unmaximize() {}

func (f *Nada) Top() int    { return 0 }
func (f *Nada) Bottom() int { return 0 }
func (f *Nada) Left() int   { return 0 }
func (f *Nada) Right() int  { return 0 }

func (f *Nada) MROpt(validate bool, flags, x, y, w, h int) {
	mropt(f, validate, flags, x, y, w, h)
}

func (f *Nada) MoveResize(validate bool, x, y, w, h int) {
	moveresize(f, validate, x, y, w, h)
}

func (f *Nada) Move(x, y int) {
	move(f, x, y)
}

func (f *Nada) Resize(validate bool, w, h int) {
	resize(f, validate, w, h)
}
