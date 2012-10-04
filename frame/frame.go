package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
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

// frame is a type that provides mostly boiler-plate methods to all frames.
// It's appropriate to think of it as an abstract frame, as it does not
// satisfy the Frame interface by itself.
type frame struct {
	X *xgbutil.XUtil

	State  int
	parent *Parent
	client Client
}

func newFrame(X *xgbutil.XUtil, p *Parent, c Client) (*frame, error) {
	var err error
	if p == nil {
		p, err = newParent(X, c.Id())
		if err != nil {
			return nil, err
		}
	}

	return &frame{
		X:      X,
		parent: p,
		client: c,
		State:  Inactive,
	}, nil
}

func (f *frame) Client() Client {
	return f.client
}

func (f *frame) Parent() *Parent {
	return f.parent
}

// Destroy will check if the client window is still a sub-window of this frame,
// and reparent it to the root window if so.
//
// Destroy does *not* destroy the parent window! The caller must do that, since
// the parent window is shared across many frames.
func (f *frame) Destroy() {
	// Only re-parent if the current parent is this frame window.
	parent, err := f.client.Win().Parent()
	if err != nil {
		// We don't care about this error
		return
	}
	if parent.Id == f.parent.Id {
		err := xproto.ReparentWindowChecked(f.X.Conn(), f.client.Id(),
			f.X.RootWin(), 0, 0).Check()
		if err != nil {
			logger.Warning.Println(err)
		}
	}
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

func (f *frame) IsMapped() bool {
	return f.parent.isMapped
}

func (f *frame) Moving() bool {
	return f.parent.MoveState.Moving
}

func (f *frame) MovingState() *MoveState {
	return f.parent.MoveState
}

func (f *frame) Resizing() bool {
	return f.parent.ResizeState.Resizing
}

func (f *frame) ResizingState() *ResizeState {
	return f.parent.ResizeState
}
