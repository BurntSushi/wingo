package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
)

var newx, newy int16 // prevent memory allocation in 'step' functions

func (f *frameAbst) moveBegin(rx, ry, ex, ey int16) {
    f.moving.moving = true
    f.moving.lastRootX, f.moving.lastRootY = rx, ry

    // call for side-effect; makes sure parent window has a valid geometry
    f.parent.window.geometry()
}

func (f *frameAbst) moveStep(rx, ry, ex, ey int16) {
    newx = f.Geom().X() + rx - f.moving.lastRootX
    newy = f.Geom().Y() + ry - f.moving.lastRootY
    f.moving.lastRootX, f.moving.lastRootY = rx, ry

    f.ConfigureFrame(DoX | DoY, newx, newy, 0, 0, 0, 0, false)
}

func (f *frameAbst) moveEnd(rx, ry, ex, ey int16) {
    f.moving.moving = false
    f.moving.lastRootX, f.moving.lastRootY = 0, 0
}

func (f *frameAbst) resizeBegin(direction uint32,
                                rx, ry, ex, ey int16) (bool, xgb.Id) {
    dir := direction
    w, h := f.Geom().Width(), f.Geom().Height()
    uex, uey := uint16(ex), uint16(ey)

    // If we aren't forcing a direction, we need to infer it based on
    // where the mouse is in the window.
    // (uex, uey) is the position of the mouse.
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
    // Where A, B, C correspond to 'uex < w / 3'
    // and F, G, H correspond to 'uex > w * 2 / 3'
    // and D and E correspond to 'uex >= w / 3 && uex <= w * 2 / 3'
    // The direction is not only important for assigning which cursor to display
    // (where each of the above blocks gets its own cursor), but it is also
    // important for choosing which parts of the geometry to change.
    // For example, if the mouse is in 'H', then the width and height could
    // be changed, but x and y cannot. Conversely, if the mouse is in 'A',
    // all parts of the geometry can change: x, y, width and height.
    // As one last example, if the mouse is in 'D', only y and height of the
    // window can change.
    if dir == ewmh.Infer {
        if uex < w / 3 {
            switch {
            case uey < h / 3: dir = ewmh.SizeTopLeft
            case uey > h * 2 / 3: dir = ewmh.SizeBottomLeft
            default: dir = ewmh.SizeLeft // uey >= h / 3 && uey <= h * 2 / 3
            }
        } else if uex > w * 2 / 3 {
            switch {
            case uey < h / 3: dir = ewmh.SizeTopRight
            case uey > h * 2 / 3: dir = ewmh.SizeBottomRight
            default: dir = ewmh.SizeRight // uey >= h / 3 && uey <= h * 2 / 3
            }
        } else { // uex >= w / 3 && uex <= w * 2 / 3
            switch {
            case uey < h / 2: dir = ewmh.SizeTop
            default: dir = ewmh.SizeBottom // uey >= h / 2
            }
        }
    }

    // Find the right cursor
    var cursor xgb.Id = 0
    switch dir {
    case ewmh.SizeTop: cursor = cursorTopSide
    case ewmh.SizeTopRight: cursor = cursorTopRightCorner
    case ewmh.SizeRight: cursor = cursorRightSide
    case ewmh.SizeBottomRight: cursor = cursorBottomRightCorner
    case ewmh.SizeBottom: cursor = cursorBottomSide
    case ewmh.SizeBottomLeft: cursor = cursorBottomLeftCorner
    case ewmh.SizeLeft: cursor = cursorLeftSide
    case ewmh.SizeTopLeft: cursor = cursorTopLeftCorner
    }

    // Save some state that we'll need when computing a window's new geometry
    f.resizing.resizing = true
    f.resizing.rootX, f.resizing.rootY = rx, ry
    f.resizing.x, f.resizing.y = f.Geom().X(), f.Geom().Y()
    f.resizing.width, f.resizing.height = f.Geom().Width(), f.Geom().Height()

    // Our geometry calculations depend upon which direction we're resizing.
    // Namely, the direction determines which parts of the geometry need to
    // be modified. Pre-compute those parts (i.e., x, y, width and/or height)
    f.resizing.xs = dir == ewmh.SizeLeft || dir == ewmh.SizeTopLeft ||
                    dir == ewmh.SizeBottomLeft
    f.resizing.ys = dir == ewmh.SizeTop || dir == ewmh.SizeTopLeft ||
                    dir == ewmh.SizeTopRight
    f.resizing.ws = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTopRight ||
                    dir == ewmh.SizeRight || dir == ewmh.SizeBottomRight ||
                    dir == ewmh.SizeBottomLeft || dir == ewmh.SizeLeft
    f.resizing.hs = dir == ewmh.SizeTopLeft || dir == ewmh.SizeTop ||
                    dir == ewmh.SizeTopRight || dir == ewmh.SizeBottomRight ||
                    dir == ewmh.SizeBottom || dir == ewmh.SizeBottomLeft

    // call for side-effect; makes sure parent window has a valid geometry
    f.parent.window.geometry()

    return true, cursor
}

func (f *frameAbst) resizeStep(rx, ry, ex, ey int16) {
    var diffx, diffy int16 = rx - f.resizing.rootX, ry - f.resizing.rootY
    var newx, newy int16 = 0, 0
    var neww, newh uint16 = 0, 0
    var validw, validh uint16 = 0, 0
    var flags uint16 = 0

    if f.resizing.xs {
        flags |= DoX
        newx = f.resizing.x + diffx
    }
    if f.resizing.ys {
        flags |= DoY
        newy = f.resizing.y + diffy
    }
    if f.resizing.ws {
        flags |= DoW
        if f.resizing.xs {
            neww = f.resizing.width - uint16(diffx)
        } else {
            neww = f.resizing.width + uint16(diffx)
        }
        validw = f.ValidateWidth(neww)

        // If validation changed our width, we need to make sure
        // our x-value is appropriately changed
        if f.resizing.xs && validw != neww {
            newx = f.resizing.x + int16(f.resizing.width - validw)
        }
    }
    if f.resizing.hs {
        flags |= DoH
        if f.resizing.ys {
            newh = f.resizing.height - uint16(diffy)
        } else {
            newh = f.resizing.height + uint16(diffy)
        }
        validh = f.ValidateHeight(newh)

        // If validation changed our height, we need to make sure
        // our y-value is appropriately changed
        if f.resizing.ys && validh != newh {
            newy = f.resizing.y + int16(f.resizing.height - validh)
        }
    }

    f.ConfigureFrame(flags, newx, newy, validw, validh, 0, 0, true)
}

func (f *frameAbst) resizeEnd(rx, ry, ex, ey int16) {
    // just zero out the resizing state
    f.resizing.resizing = false
    f.resizing.rootX, f.resizing.rootY = 0, 0
    f.resizing.x, f.resizing.y = 0, 0
    f.resizing.width, f.resizing.height = 0, 0
    f.resizing.xs, f.resizing.ys = false, false
    f.resizing.ws, f.resizing.hs = false, false
}

