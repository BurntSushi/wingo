package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type abstFrame struct {
	parent   *frameParent
	moving   *moveState
	resizing *resizeState
	state    int
}

func newFrameAbst(p *frameParent, c *client) *abstFrame {
	f := &abstFrame{
		parent:   p,
		moving:   &moveState{},
		resizing: &resizeState{},
	}

	return f
}

func (f *abstFrame) Destroy() {
	if f.Client().TrulyAlive() {
		X.Conn().ReparentWindow(f.Client().Id(), ROOT.id, 0, 0)
	}
	f.parent.window.destroy()
}

func (f *abstFrame) FrameState() int {
	return f.state
}

func (f *abstFrame) State() int {
	return f.Client().state
}

func (f *abstFrame) Map() {
	f.parent.window.map_()
}

func (f *abstFrame) Unmap() {
	f.parent.window.unmap()
}

func (f *abstFrame) Client() *client {
	return f.parent.client
}

func (f *abstFrame) Geom() xrect.Rect {
	return f.parent.window.geom
}

func (f *abstFrame) Moving() bool {
	return f.moving.moving
}

func (f *abstFrame) MovingState() *moveState {
	return f.moving
}

func (f *abstFrame) Parent() *frameParent {
	return f.parent
}

func (f *abstFrame) ParentId() xgb.Id {
	return f.parent.window.id
}

func (f *abstFrame) ParentWin() *window {
	return f.parent.window
}

func (f *abstFrame) Resizing() bool {
	return f.resizing.resizing
}

func (f *abstFrame) ResizingState() *resizeState {
	return f.resizing
}
