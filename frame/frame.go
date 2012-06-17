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
	Moving       bool
	RootX, RootY int
}

type ResizeState struct {
	Resizing            bool
	RootX, RootY        int
	X, Y, Width, Height int
	Xs, Ys, Ws, Hs      bool
}

// frame is a type that provides mostly boiler-plate methods to all frames.
// It's appropriate to think of it as an abstract frame, as it does not
// satisfy the Frame interface by itself.
type frame struct {
	X     *xgbutil.XUtil
	theme *theme.Theme

	MoveState   *MoveState
	ResizeState *ResizeState
	State       int
	parent      *Parent
	client      Client
	isMapped    bool
}

func newFrame(X *xgbutil.XUtil,
	t *theme.Theme, p *Parent, c Client) (*frame, error) {

	var err error
	if p == nil {
		p, err = newParent(X, c.Id())
		if err != nil {
			return nil, err
		}
	}

	return &frame{
		X:           X,
		theme:       t,
		parent:      p,
		client:      c,
		MoveState:   &MoveState{},
		ResizeState: &ResizeState{},
		State:       Inactive,
		isMapped:    false,
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
	f.isMapped = true
}

func (f *frame) Unmap() {
	f.parent.Unmap()
	f.isMapped = false
}

func (f *frame) Geom() xrect.Rect {
	return f.parent.Geom
}

func (f *frame) IsMapped() bool {
	return f.isMapped
}

func (f *frame) Moving() bool {
	return f.MoveState.Moving
}

func (f *frame) MovingState() *MoveState {
	return f.MoveState
}

func (f *frame) Resizing() bool {
	return f.ResizeState.Resizing
}

func (f *frame) ResizingState() *ResizeState {
	return f.ResizeState
}
