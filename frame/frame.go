package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/theme"
)

const (
	Active = iota
	Inactive
)

const (
	fX = xproto.ConfigWindowX
	fY = xproto.ConfigWindowY
	fW = xproto.ConfigWindowWidth
	fH = xproto.ConfigWindowHeight
)

type MoveState struct {
	Moving bool
	RootX, RootY int
}

type ResizeState struct {
	Resizing bool
	RootX, RootY int
	X, Y, Width, Height int
	Xs, Ys, Ws, Hs bool
}

// frame is a type that provides mostly boiler-plate methods to all frames.
// It's appropriate to think of it as an abstract frame, as it does not
// satisfy the Frame interface by itself.
type frame struct {
	X *xgbutil.XUtil
	theme *theme.Theme

	MoveState MoveState
	ResizeState ResizeState
	State int
	parent *Parent
	client client
}

func newFrame(X *xgbutil.XUtil,
	t *theme.Theme, p *Parent, c client) (*frame, error) {

	var err error
	if p == nil {
		p, err = newParent(X, c.Id())
		if err != nil {
			return nil, err
		}
	}

	return &frame{
		X: X,
		theme: t,
		parent: p,
		client: c,
	}, nil
}

func (f *frame) Client() client {
	return f.client
}

func (f *frame) Parent() *Parent {
	return f.parent
}

func (f *frame) Destroy() {
	err := xproto.ReparentWindowChecked(f.X.Conn(), f.client.Id(),
		f.X.RootWin(), 0, 0).Check()
	if err != nil {
		logger.Warning.Println(err)
	}
	f.parent.Destroy()
}

func (f *frame) Map() {
	f.parent.Map()
}

func (f *frame) Unmap() {
	f.parent.Unmap()
}

func (f *frame) Geom() xrect.Rect {
	return f.parent.Geom
}

func (f *frame) Moving() bool {
	return f.MoveState.Moving
}

func (f *frame) MovingState() MoveState {
	return f.MoveState
}

func (f *frame) Resizing() bool {
	return f.ResizeState.Resizing
}

func (f *frame) ResizingState() ResizeState {
	return f.ResizeState
}
