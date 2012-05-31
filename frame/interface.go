package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/cursors"
)

type Frame interface {
	Client() client
	Parent() *Parent

	Map()
	Unmap()
	Destroy()
	Geom() xrect.Rect

	Move(x, y int)
	Resize(validate bool, width, height int)
	MoveResize(validate bool, x, y, width, height int)
	MROpt(validate bool, flags, x, y, width, height int)

	Moving() bool
	MovingState() MoveState
	Resizing() bool
	ResizingState() ResizeState

	Current() bool

	Top() int
	Bottom() int
	Left() int
	Right() int

	Off()
	On()
	Active()
	Inactive()
	Maximize()
	Unmaximize()
}

func Reset(f Frame) {
	geom := f.Geom()
	resize(f, true, geom.Width(), geom.Height())
}

func ClientToFrame(f Frame, x, y, w, h int) (int, int, int, int) {
	return f.Client().GravitizeX(x, -1),
		f.Client().GravitizeY(y, -1),
		w + f.Left() + f.Right(),
		h + f.Top() + f.Bottom()
}

func validateWidthHeight(f Frame, validate bool,
	w, h int) (fw, fh, cw, ch int) {
	fw, fh, cw, ch = w, h, w, h

	cw -= f.Left() + f.Right()
	if validate {
		cw = f.Client().ValidateWidth(cw)
		fw = cw + f.Left() + f.Right()
	}

	ch -= f.Top() + f.Bottom()
	if validate {
		ch = f.Client().ValidateHeight(ch)
		fh = ch + f.Top() + f.Bottom()
	}

	return
}

func mropt(f Frame, validate bool, flags, x, y, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().MROpt(flags, x, y, fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func moveresize(f Frame, validate bool, x, y, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().MoveResize(x, y, fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func move(f Frame, x, y int) {
	f.Parent().Move(x, y)
}

func resize(f Frame, validate bool, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().Resize(fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func Maximize(f Frame) {
	hg := xrect.New(xrect.Pieces(f.Client().HeadGeom()))
	moveresize(f, false, hg.X(), hg.Y(), hg.Width(), hg.Height())
}

func DragMoveBegin(f Frame, rx, ry, ex, ey int) {
	moving := f.MovingState()
	moving.Moving = true
	moving.RootX, moving.RootY = rx, ry

	// call for side-effect; makes sure parent window has a valid geometry
	f.Parent().Geometry()

	// unmax!
	f.Client().EnsureUnmax()
}

func DragMoveStep(f Frame, rx, ry, ex, ey int) {
	moving := f.MovingState()
	newx := f.Geom().X() + rx - moving.RootX
	newy := f.Geom().Y() + ry - moving.RootY
	moving.RootX, moving.RootY = rx, ry

	move(f, newx, newy)
}

func DragMoveEnd(f Frame, rx, ry, ex, ey int) {
	Reset(f)
	// WM.headChoose(f.Client(), f.Geom()) 

	moving := f.MovingState()
	moving.Moving = false
	moving.RootX, moving.RootY = 0, 0
}

func DragResizeBegin(f Frame, direction uint32,
	rx, ry, ex, ey int) (bool, xproto.Cursor) {

	resizing := f.ResizingState()
	dir := direction
	w, h := f.Geom().Width(), f.Geom().Height()

	// If we aren't forcing a direction, we need to infer it based on
	// where the mouse is in the window.
	// (ex, ey) is the position of the mouse.
	// We basically split the window into something like a tic-tac-toe board:
	// -------------------------
	// |       |       |       |
	// |   A   |       |   F   |
	// |       |   D   |       |
	// ---------       |--------
	// |       |       |       |
	// |   B   |-------|   G   |
	// |       |       |       |
	// ---------       |--------
	// |       |   E   |       |
	// |   C   |       |   H   |
	// |       |       |       |
	// -------------------------
	// Where A, B, C correspond to 'ex < w / 3'
	// and F, G, H correspond to 'ex > w * 2 / 3'
	// and D and E correspond to 'ex >= w / 3 && ex <= w * 2 / 3'
	// The direction is not only important for assigning which cursor to display
	// (where each of the above blocks gets its own cursor), but it is also
	// important for choosing which parts of the geometry to change.
	// For example, if the mouse is in 'H', then the width and height could
	// be changed, but x and y cannot. Conversely, if the mouse is in 'A',
	// all parts of the geometry can change: x, y, width and height.
	// As one last example, if the mouse is in 'D', only y and height of the
	// window can change.
	if dir == ewmh.Infer {
		if ex < w/3 {
			switch {
			case ey < h/3:
				dir = ewmh.SizeTopLeft
			case ey > h*2/3:
				dir = ewmh.SizeBottomLeft
			default: // ey >= h / 3 && ey <= h * 2 / 3
				dir = ewmh.SizeLeft
			}
		} else if ex > w*2/3 {
			switch {
			case ey < h/3:
				dir = ewmh.SizeTopRight
			case ey > h*2/3:
				dir = ewmh.SizeBottomRight
			default: // ey >= h / 3 && ey <= h * 2 / 3
				dir = ewmh.SizeRight
			}
		} else { // ex >= w / 3 && ex <= w * 2 / 3
			switch {
			case ey < h/2:
				dir = ewmh.SizeTop
			default: // ey >= h / 2
				dir = ewmh.SizeBottom
			}
		}
	}

	// Find the right cursor
	var cursor xproto.Cursor = 0
	switch dir {
	case ewmh.SizeTop:
		cursor = cursors.TopSide
	case ewmh.SizeTopRight:
		cursor = cursors.TopRightCorner
	case ewmh.SizeRight:
		cursor = cursors.RightSide
	case ewmh.SizeBottomRight:
		cursor = cursors.BottomRightCorner
	case ewmh.SizeBottom:
		cursor = cursors.BottomSide
	case ewmh.SizeBottomLeft:
		cursor = cursors.BottomLeftCorner
	case ewmh.SizeLeft:
		cursor = cursors.LeftSide
	case ewmh.SizeTopLeft:
		cursor = cursors.TopLeftCorner
	}

	// Save some state that we'll need when computing a window's new geometry
	resizing.Resizing = true
	resizing.RootX, resizing.RootY = rx, ry
	resizing.X, resizing.Y = f.Geom().X(), f.Geom().Y()
	resizing.Width, resizing.Height = f.Geom().Width(), f.Geom().Height()

	// Our geometry calculations depend upon which direction we're resizing.
	// Namely, the direction determines which parts of the geometry need to
	// be modified. Pre-compute those parts (i.e., x, y, width and/or height)
	resizing.Xs = dir == ewmh.SizeLeft || dir == ewmh.SizeTopLeft ||
		dir == ewmh.SizeBottomLeft
	resizing.Ys = dir == ewmh.SizeTop || dir == ewmh.SizeTopLeft ||
		dir == ewmh.SizeTopRight
	resizing.Ws = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTopRight ||
		dir == ewmh.SizeRight || dir == ewmh.SizeBottomRight ||
		dir == ewmh.SizeBottomLeft || dir == ewmh.SizeLeft
	resizing.Hs = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTop ||
		dir == ewmh.SizeTopRight || dir == ewmh.SizeBottomRight ||
		dir == ewmh.SizeBottom || dir == ewmh.SizeBottomLeft

	// call for side-effect; makes sure parent window has a valid geometry
	f.Parent().Geometry()

	// unmax!
	f.Client().EnsureUnmax()

	return true, cursor
}

func DragResizeStep(f Frame, rx, ry, ex, ey int) {
	resizing := f.ResizingState()

	diffx, diffy := rx-resizing.RootX, ry-resizing.RootY
	newx, newy, neww, newh := 0, 0, 0, 0
	validw, validh := 0, 0
	flags := 0

	if resizing.Xs {
		flags |= fX
		newx = resizing.X + diffx
	}
	if resizing.Ys {
		flags |= fY
		newy = resizing.Y + diffy
	}
	if resizing.Ws {
		flags |= fW
		if resizing.Xs {
			neww = resizing.Width - diffx
		} else {
			neww = resizing.Width + diffx
		}

		topBot := f.Top() + f.Bottom()
		validw = f.Client().ValidateWidth(neww-topBot) + topBot

		// If validation changed our width, we need to make sure
		// our x-value is appropriately changed
		if resizing.Xs && validw != neww {
			newx = resizing.X + resizing.Width - validw
		}
	}
	if resizing.Hs {
		flags |= fH
		if resizing.Ys {
			newh = resizing.Height - diffy
		} else {
			newh = resizing.Height + diffy
		}

		leftRight := f.Left() + f.Right()
		validh = f.Client().ValidateHeight(newh-leftRight) + leftRight

		// If validation changed our height, we need to make sure
		// our y-value is appropriately changed
		if resizing.Ys && validh != newh {
			newy = resizing.Y + resizing.Height - validh
		}
	}

	moveresize(f, false, newx, newy, validw, validh)
}

func DragResizeEnd(f Frame, rx, ry, ex, ey int) {
	// If windows are really slow to respond/resize, this may be necessary.
	// If we don't, it's possible for the client to be out of whack inside
	// the decorations.
	// Example: Libreoffice in Xephyr. Try resizing it with the mouse and
	// releasing the mouse button really quickly.
	Reset(f)
	// WM.headChoose(f.Client(), f.Geom()) 

	// just zero out the resizing state
	resizing := f.ResizingState()
	resizing.Resizing = false
	resizing.RootX, resizing.RootY = 0, 0
	resizing.X, resizing.Y = 0, 0
	resizing.Width, resizing.Height = 0, 0
	resizing.Xs, resizing.Ys = false, false
	resizing.Ws, resizing.Hs = false, false
}
