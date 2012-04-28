package main

import (
	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"
)

func frameMaximize(f Frame) {
	headGeom := xrect.New(xrect.Pieces(f.Client().workspace.headGeom()))
	f.ConfigureFrame(DoX|DoY|DoW|DoH,
		headGeom.X(), headGeom.Y(), headGeom.Width(), headGeom.Height(),
		0, 0, true, true)
}

func frameMoveBegin(f Frame, rx, ry, ex, ey int) {
	moving := f.MovingState()
	moving.moving = true
	moving.lastRootX, moving.lastRootY = rx, ry

	// call for side-effect; makes sure parent window has a valid geometry
	f.ParentWin().geometry()

	// unmax!
	f.Client().EnsureUnmax()
}

func frameMoveStep(f Frame, rx, ry, ex, ey int) {
	moving := f.MovingState()
	newx := f.Geom().X() + rx - moving.lastRootX
	newy := f.Geom().Y() + ry - moving.lastRootY
	moving.lastRootX, moving.lastRootY = rx, ry

	f.ConfigureFrame(DoX|DoY, newx, newy, 0, 0, 0, 0, false, false)
}

func frameMoveEnd(f Frame, rx, ry, ex, ey int) {
	FrameReset(f)
	WM.headChoose(f.Client(), f.Geom())

	moving := f.MovingState()
	moving.moving = false
	moving.lastRootX, moving.lastRootY = 0, 0
}

func frameResizeBegin(f Frame, direction uint32,
	rx, ry, ex, ey int) (bool, xgb.Id) {

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
			default:
				dir = ewmh.SizeLeft // ey >= h / 3 && ey <= h * 2 / 3
			}
		} else if ex > w*2/3 {
			switch {
			case ey < h/3:
				dir = ewmh.SizeTopRight
			case ey > h*2/3:
				dir = ewmh.SizeBottomRight
			default:
				dir = ewmh.SizeRight // ey >= h / 3 && ey <= h * 2 / 3
			}
		} else { // ex >= w / 3 && ex <= w * 2 / 3
			switch {
			case ey < h/2:
				dir = ewmh.SizeTop
			default:
				dir = ewmh.SizeBottom // ey >= h / 2
			}
		}
	}

	// Find the right cursor
	var cursor xgb.Id = 0
	switch dir {
	case ewmh.SizeTop:
		cursor = cursorTopSide
	case ewmh.SizeTopRight:
		cursor = cursorTopRightCorner
	case ewmh.SizeRight:
		cursor = cursorRightSide
	case ewmh.SizeBottomRight:
		cursor = cursorBottomRightCorner
	case ewmh.SizeBottom:
		cursor = cursorBottomSide
	case ewmh.SizeBottomLeft:
		cursor = cursorBottomLeftCorner
	case ewmh.SizeLeft:
		cursor = cursorLeftSide
	case ewmh.SizeTopLeft:
		cursor = cursorTopLeftCorner
	}

	// Save some state that we'll need when computing a window's new geometry
	resizing.resizing = true
	resizing.rootX, resizing.rootY = rx, ry
	resizing.x, resizing.y = f.Geom().X(), f.Geom().Y()
	resizing.width, resizing.height = f.Geom().Width(), f.Geom().Height()

	// Our geometry calculations depend upon which direction we're resizing.
	// Namely, the direction determines which parts of the geometry need to
	// be modified. Pre-compute those parts (i.e., x, y, width and/or height)
	resizing.xs = dir == ewmh.SizeLeft || dir == ewmh.SizeTopLeft ||
		dir == ewmh.SizeBottomLeft
	resizing.ys = dir == ewmh.SizeTop || dir == ewmh.SizeTopLeft ||
		dir == ewmh.SizeTopRight
	resizing.ws = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTopRight ||
		dir == ewmh.SizeRight || dir == ewmh.SizeBottomRight ||
		dir == ewmh.SizeBottomLeft || dir == ewmh.SizeLeft
	resizing.hs = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTop ||
		dir == ewmh.SizeTopRight || dir == ewmh.SizeBottomRight ||
		dir == ewmh.SizeBottom || dir == ewmh.SizeBottomLeft

	// call for side-effect; makes sure parent window has a valid geometry
	f.ParentWin().geometry()

	// unmax!
	f.Client().EnsureUnmax()

	return true, cursor
}

func frameResizeStep(f Frame, rx, ry, ex, ey int) {
	resizing := f.ResizingState()

	diffx, diffy := rx-resizing.rootX, ry-resizing.rootY
	newx, newy, neww, newh := 0, 0, 0, 0
	validw, validh := 0, 0
	flags := 0

	if resizing.xs {
		flags |= DoX
		newx = resizing.x + diffx
	}
	if resizing.ys {
		flags |= DoY
		newy = resizing.y + diffy
	}
	if resizing.ws {
		flags |= DoW
		if resizing.xs {
			neww = resizing.width - diffx
		} else {
			neww = resizing.width + diffx
		}
		validw = FrameValidateWidth(f, neww)

		// If validation changed our width, we need to make sure
		// our x-value is appropriately changed
		if resizing.xs && validw != neww {
			newx = resizing.x + resizing.width - validw
		}
	}
	if resizing.hs {
		flags |= DoH
		if resizing.ys {
			newh = resizing.height - diffy
		} else {
			newh = resizing.height + diffy
		}
		validh = FrameValidateHeight(f, newh)

		// If validation changed our height, we need to make sure
		// our y-value is appropriately changed
		if resizing.ys && validh != newh {
			newy = resizing.y + resizing.height - validh
		}
	}

	f.ConfigureFrame(flags, newx, newy, validw, validh, 0, 0, true, false)
}

func frameResizeEnd(f Frame, rx, ry, ex, ey int) {
	// If windows are really slow to respond/resize, this may be necessary.
	// If we don't, it's possible for the client to be out of whack inside
	// the decorations.
	// Example: Libreoffice in Xephyr. Try resizing it with the mouse and
	// releasing the mouse button really quickly.
	FrameReset(f)
	WM.headChoose(f.Client(), f.Geom())

	// just zero out the resizing state
	resizing := f.ResizingState()
	resizing.resizing = false
	resizing.rootX, resizing.rootY = 0, 0
	resizing.x, resizing.y = 0, 0
	resizing.width, resizing.height = 0, 0
	resizing.xs, resizing.ys = false, false
	resizing.ws, resizing.hs = false, false
}
