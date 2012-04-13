package main

import (
	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
)

const (
	StateActive = iota
	StateInactive
)

type Frame interface {
	Client() *client
	ConfigureClient(flags, x, y, w, h int,
		sibling xgb.Id, stackMode byte, ignoreHints bool)
	ConfigureFrame(flags, x, y, w, h int,
		sibling xgb.Id, stackMode byte, ignoreHints, sendNotify bool)
	Current() bool
	Destroy()
	Geom() xrect.Rect // the geometry of the parent window
	Map()
	Off()
	On()
	Parent() *frameParent
	ParentId() xgb.Id
	ParentWin() *window
	State() int
	Active()
	Inactive()
	Unmap()

	// The margins of this frame's decorations.
	Top() int
	Bottom() int
	Left() int
	Right() int

	// These are temporary. I think they will move to 'layout'
	Moving() bool
	MovingState() *moveState
	Resizing() bool
	ResizingState() *resizeState
	Maximize()
	Unmaximize()
}

type frameParent struct {
	window *window
	client *client
}

func newParent(c *client) *frameParent {
	mask := xgb.CWEventMask
	val := uint32(xgb.EventMaskSubstructureRedirect |
		xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease)
	if CONF.ffm {
		val |= xgb.EventMaskEnterWindow
	}

	parent := createWindow(X.RootWin(), mask, val)
	p := &frameParent{
		window: parent,
		client: c,
	}

	X.Conn().ReparentWindow(c.Id(), parent.id, 0, 0)

	return p
}

func (p *frameParent) Win() *window {
	return p.window
}

// framePiece contains the information required to show *any* piece of the
// decorations. Basically, it contains the raw X window and pixmaps for each
// of the available states for quick switching.
type framePiece struct {
	win         *window
	imgActive   xgb.Id
	imgInactive xgb.Id
}

func newFramePiece(win *window, imgA, imgI xgb.Id) framePiece {
	return framePiece{win: win, imgActive: imgA, imgInactive: imgI}
}

func (p *framePiece) destroy() {
	p.win.destroy() // also detaches all event handlers
	xgraphics.FreePixmap(X, p.imgActive)
	xgraphics.FreePixmap(X, p.imgInactive)
}

func (p *framePiece) active() {
	p.win.change(xgb.CWBackPixmap, uint32(p.imgActive))
	p.win.clear()
}

func (p *framePiece) inactive() {
	p.win.change(xgb.CWBackPixmap, uint32(p.imgInactive))
	p.win.clear()
}

func (p *framePiece) x() int {
	return p.win.geom.X()
}

func (p *framePiece) y() int {
	return p.win.geom.Y()
}

func (p *framePiece) w() int {
	return p.win.geom.Width()
}

func (p *framePiece) h() int {
	return p.win.geom.Height()
}

// The relative geometry of the client window in the frame parent window.
// x and y are relative to the top-left corner of the parent window.
// w and h are values that satisfy these properties:
// parent_width - w = client_width
// parent_height - h = client_height
// Where client_width and client_height is the width and height of the client
// window inside the frame.
type clientOffset struct {
	x, y int
	w, h int
}

type moveState struct {
	moving    bool
	lastRootX int
	lastRootY int
}

type resizeState struct {
	resizing       bool
	rootX, rootY   int
	x, y           int
	width, height  int
	xs, ys, ws, hs bool
}

// Frame related functions that can be defined using only the Frame interface.

func FrameClientReset(f Frame) {
	geom := f.Client().Geom()
	FrameMR(f, DoW|DoH, 0, 0, geom.Width(), geom.Height(), false)
}

func FrameReset(f Frame) {
	geom := f.Geom()
	f.ConfigureFrame(DoW|DoH, 0, 0, geom.Width(), geom.Height(), 0, 0,
		false, false)
}

// FrameMR is short for FrameMoveresize.
func FrameMR(f Frame, flags int, x, y int, w, h int, ignoreHints bool) {
	f.ConfigureClient(flags, x, y, w, h, xgb.Id(0), 0, ignoreHints)
}

// FrameValidateHeight validates a height of a *frame*, which is equivalent
// to validating the height of a client.
func FrameValidateHeight(f Frame, height int) int {
	frameTopBot := f.Top() + f.Bottom()
	return f.Client().ValidateHeight(height-frameTopBot) + frameTopBot
}

// validateWidth validates a width of a *frame*, which is equivalent
// to validating the width of a client.
func FrameValidateWidth(f Frame, width int) int {
	frameLeftRight := f.Left() + f.Right()
	return f.Client().ValidateWidth(width-frameLeftRight) + frameLeftRight
}

// Configure is from the perspective of the client.
// Namely, the width and height specified here will be precisely the width
// and height that the client itself ends up with, assuming it passes
// validation. (Therefore, the actual window itself will be bigger, because
// of decorations.)
// Moreover, the x and y coordinates are gravitized. Yuck.
func FrameConfigureClient(f Frame, flags, x, y, w, h int) (int, int, int, int) {
	// Defy gravity!
	if DoX&flags > 0 {
		x = f.Client().GravitizeX(x, -1)
	}
	if DoY&flags > 0 {
		y = f.Client().GravitizeY(y, -1)
	}

	// This will change with other frames
	if DoW&flags > 0 {
		w += f.Left() + f.Right()
	}
	if DoH&flags > 0 {
		h += f.Top() + f.Bottom()
	}

	return x, y, w, h
}

// ConfigureFrame is from the perspective of the frame.
// The fw and fh specify the width of the entire window, so that the client
// will end up slightly smaller than the width/height specified here.
// Also, the fx and fy coordinates are interpreted plainly as root window
// coordinates. (No gravitization.)
func FrameConfigureFrame(f Frame, flags, fx, fy, fw, fh int,
	sibling xgb.Id, stackMode byte, ignoreHints bool, sendNotify bool) {

	cw, ch := fw, fh
	framex, framey, _, _ := xrect.RectPieces(f.Geom())
	_, _, clientw, clienth := xrect.RectPieces(f.Client().Geom())

	if DoX&flags > 0 {
		framex = fx
	}
	if DoY&flags > 0 {
		framey = fy
	}
	if DoW&flags > 0 {
		cw -= f.Left() + f.Right()
		if !ignoreHints {
			cw = f.Client().ValidateWidth(cw)
			fw = cw + f.Left() + f.Right()
		}
		clientw = cw
	}
	if DoH&flags > 0 {
		ch -= f.Top() + f.Bottom()
		if !ignoreHints {
			ch = f.Client().ValidateHeight(ch)
			fh = ch + f.Top() + f.Bottom()
		}
		clienth = ch
	}

	if sendNotify {
		configNotify := xevent.NewConfigureNotify(f.Client().Id(),
			f.Client().Id(),
			0, framex, framey,
			clientw, clienth, 0, false)
		X.Conn().SendEvent(false, f.Client().Id(), xgb.EventMaskStructureNotify,
			configNotify.Bytes())
	}

	f.Parent().Win().configure(flags, fx, fy, fw, fh, sibling, stackMode)
	f.Client().Win().moveresize(flags|DoX|DoY,
		f.Left(), f.Top(), cw, ch)
}
