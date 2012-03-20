package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xrect"
)

type Frame interface {
    Client() Client
    ConfigureClient(flags uint16, x, y int16, w, h uint16,
                    sibling xgb.Id, stackMode byte, ignoreHints bool)
    ConfigureFrame(flags uint16, x, y int16, w, h uint16,
                   sibling xgb.Id, stackMode byte, ignoreHints bool)
    Destroy()
    Geom() xrect.Rect // the geometry of the parent window
    Map()
    Moveresize(flags uint16, x, y int16, w, h uint16, ignoreHints bool)
    Moving() bool
    Parent() *frameParent
    ParentId() xgb.Id
    Resizing() bool
    Unmap()
    ValidateHeight(height uint16) uint16
    ValidateWidth(width uint16) uint16

    // The margins of this frame's decorations.
    Top() int16
    Bottom() int16
    Left() int16
    Right() int16

    // These are temporary. I think they will move to 'layout'
    moveBegin(rx, ry, ex, ey int16)
    moveStep(rx, ry, ex, ey int16)
    moveEnd(rx, ry, ex, ey int16)
    resizeBegin(direction uint32, rx, ry, ex, ey int16) (bool, xgb.Id)
    resizeStep(rx, ry, ex, ey int16)
    resizeEnd(rx, ry, ex, ey int16)
}

// The relative geometry of the client window in the frame parent window.
// x and y are relative to the top-left corner of the parent window.
// w and h are values that satisfy these properties:
// parent_width - w = client_width
// parent_height - h = client_height
// Where client_width and client_height is the width and height of the client
// window inside the frame.
type clientPos struct {
    x, y int16
    w, h uint16
}

type moveState struct {
    moving bool
    lastRootX int16
    lastRootY int16
}

type resizeState struct {
    resizing bool
    rootX, rootY int16
    x, y int16
    width, height uint16
    xs, ys, ws, hs bool
}

type frameParent struct {
    window *window
    client Client
}

func newParent(c Client) *frameParent {
    mask := uint32(xgb.CWBackPixel | xgb.CWEventMask)
    val := []uint32{0xff7f00,
                    xgb.EventMaskSubstructureRedirect |
                    xgb.EventMaskButtonPress |
                    xgb.EventMaskButtonRelease}
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

