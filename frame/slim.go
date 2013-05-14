package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"

	"github.com/BurntSushi/wingo-conc/render"
)

type Slim struct {
	*frame
	theme *SlimTheme
}

func NewSlim(X *xgbutil.XUtil,
	t *SlimTheme, p *Parent, c Client) (*Slim, error) {

	f, err := newFrame(X, p, c)
	if err != nil {
		return nil, err
	}
	return &Slim{frame: f, theme: t}, nil
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
	f.State = Active

	f.parent.Change(xproto.CwBackPixel, f.theme.ABorderColor.Uint32())
	f.parent.ClearAll()
}

func (f *Slim) Inactive() {
	f.State = Inactive

	f.parent.Change(xproto.CwBackPixel, f.theme.IBorderColor.Uint32())
	f.parent.ClearAll()
}

func (f *Slim) Maximize()   {}
func (f *Slim) Unmaximize() {}

func (f *Slim) Top() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Slim) Bottom() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Slim) Left() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Slim) Right() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
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

type SlimTheme struct {
	BorderSize                 int
	ABorderColor, IBorderColor render.Color
}

func DefaultSlimTheme() *SlimTheme {
	return &SlimTheme{
		BorderSize:   10,
		ABorderColor: render.NewColor(0x3366ff),
		IBorderColor: render.NewColor(0xdfdcdf),
	}
}
