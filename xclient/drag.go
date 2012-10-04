package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/frame"
)

func (c *Client) DragGeom() xrect.Rect {
	return c.dragGeom
}

func (c *Client) DragMoveBegin(rx, ry, ex, ey int) {
	f := c.frame
	moving := f.MovingState()
	moving.Moving = true
	moving.RootX, moving.RootY = rx, ry

	// call for side-effect; makes sure parent window has a valid geometry
	f.Parent().Geometry()

	// unmax!
	c.EnsureUnmax()

	c.dragGeom = xrect.New(xrect.Pieces(f.Geom()))
}

func (c *Client) DragMoveStep(rx, ry, ex, ey int) {
	f := c.frame
	moving := f.MovingState()
	newx := c.dragGeom.X() + rx - moving.RootX
	newy := c.dragGeom.Y() + ry - moving.RootY
	moving.RootX, moving.RootY = rx, ry

	c.dragGeom.XSet(newx)
	c.dragGeom.YSet(newy)
	c.LayoutMove(newx, newy)
}

func (c *Client) DragMoveEnd(rx, ry, ex, ey int) {
	f := c.frame
	frame.Reset(f)

	moving := f.MovingState()
	moving.Moving = false
	moving.RootX, moving.RootY = 0, 0
	c.dragGeom = nil
}

func (c *Client) DragResizeBegin(direction uint32,
	rx, ry, ex, ey int) (bool, xproto.Cursor) {

	f := c.frame

	// call for side-effect; makes sure parent window has a valid geometry
	f.Parent().Geometry()

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

	// unmax!
	c.EnsureUnmax()
	c.dragGeom = xrect.New(xrect.Pieces(f.Geom()))

	return true, cursor
}

func (c *Client) DragResizeStep(rx, ry, ex, ey int) {
	f := c.frame
	resizing := f.ResizingState()

	diffx, diffy := rx-resizing.RootX, ry-resizing.RootY
	newx, newy := c.dragGeom.X(), c.dragGeom.Y()
	neww, newh := c.dragGeom.Width(), c.dragGeom.Height()
	validw, validh := neww, newh

	if resizing.Xs {
		newx = resizing.X + diffx
	}
	if resizing.Ys {
		newy = resizing.Y + diffy
	}
	if resizing.Ws {
		if resizing.Xs {
			neww = resizing.Width - diffx
		} else {
			neww = resizing.Width + diffx
		}

		leftRight := f.Left() + f.Right()
		validw = c.ValidateWidth(neww-leftRight) + leftRight

		// If validation changed our width, we need to make sure
		// our x-value is appropriately changed
		if resizing.Xs && validw != neww {
			newx = resizing.X + resizing.Width - validw
		}
	}
	if resizing.Hs {
		if resizing.Ys {
			newh = resizing.Height - diffy
		} else {
			newh = resizing.Height + diffy
		}

		topBot := f.Top() + f.Bottom()
		validh = c.ValidateHeight(newh-topBot) + topBot

		// If validation changed our height, we need to make sure
		// our y-value is appropriately changed
		if resizing.Ys && validh != newh {
			newy = resizing.Y + resizing.Height - validh
		}
	}

	c.dragGeom.XSet(newx)
	c.dragGeom.YSet(newy)
	c.dragGeom.WidthSet(validw)
	c.dragGeom.HeightSet(validh)
	c.LayoutMoveResize(newx, newy, validw, validh)
}

func (c *Client) DragResizeEnd(rx, ry, ex, ey int) {
	f := c.frame

	// If windows are really slow to respond/resize, this may be necessary.
	// If we don't, it's possible for the client to be out of whack inside
	// the decorations.
	// Example: Libreoffice in Xephyr. Try resizing it with the mouse and
	// releasing the mouse button really quickly.
	frame.Reset(f)

	// just zero out the resizing state
	resizing := f.ResizingState()
	resizing.Resizing = false
	resizing.RootX, resizing.RootY = 0, 0
	resizing.X, resizing.Y = 0, 0
	resizing.Width, resizing.Height = 0, 0
	resizing.Xs, resizing.Ys = false, false
	resizing.Ws, resizing.Hs = false, false
	c.dragGeom = nil
}
