package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xrect"
    // "github.com/BurntSushi/xgbutil" 
    // "github.com/BurntSushi/xgbutil/mousebind" 
    // "github.com/BurntSushi/xgbutil/xevent" 
)

type Frame interface {
    Client() Client
    Configure(flags uint16, x, y int16, w, h uint16,
              sibling xgb.Id, stackMode byte)
    Destroy()
    Geom() xrect.Rect // the geometry of the parent window
    Map()
    Moveresize(flags uint16, x, y int16, w, h uint16)
    Parent() *frameParent
    Unmap()

    // These are temporary. I think they will move to 'layout'
    moveBegin(rx, ry, ex, ey int16)
    moveStep(rx, ry, ex, ey int16)
    moveEnd(rx, ry, ex, ey int16)
    resizeBegin(direction uint32, rx, ry, ex, ey int16) (bool, xgb.Id)
    resizeStep(rx, ry, ex, ey int16)
    resizeEnd(rx, ry, ex, ey int16)
}

type frameParent struct {
    window *window
    client Client
}

type frameNada struct {
    parent *frameParent
    moving moveState
    resizing resizeState
}

type moveState struct {
    lastRootX int16
    lastRootY int16
}

type resizeState struct {
    rootX, rootY int16
    x, y int16
    width, height uint16
    direction uint32
}

func newParent(c Client, x, y int16) *frameParent {
    mask := uint32(xgb.CWEventMask)
    val := []uint32{xgb.EventMaskSubstructureRedirect |
                    xgb.EventMaskButtonPress |
                    xgb.EventMaskButtonRelease}
    parent := createWindow(X.RootWin(), mask, val)
    p := &frameParent{
        window: parent,
        client: c,
    }

    X.Conn().ReparentWindow(c.Id(), parent.id, x, y)

    return p
}

func newFrameNada(c Client) (*frameNada, error) {
    geom, err := c.Win().geometry()
    if err != nil {
        return nil, err
    }

    f := &frameNada {
        parent: newParent(c, 0, 0),
        moving: moveState{},
        resizing: resizeState{},
    }

    f.Moveresize(DoW | DoH, 0, 0, geom.Width(), geom.Height())

    return f, nil
}

func (f *frameNada) Destroy() {
    // Unparent before destroying the parent window so we don't kill
    // the client prematurely.
    // X.Conn().ReparentWindow(f.parent.client.id(), X.RootWin(), 
                            // f.parent.window.geom.X(), 
                            // f.parent.window.geom.Y()) 
    f.parent.window.destroy()
}

func (f *frameNada) Map() {
    f.parent.window.map_()
}

func (f *frameNada) Unmap() {
    f.parent.window.unmap()
}

func (f *frameNada) Client() Client {
    return f.parent.client
}

func (f *frameNada) Moveresize(flags uint16, x, y int16, w, h uint16) {
    f.Configure(flags, x, y, w, h, xgb.Id(0), 0)
}

func (f *frameNada) Configure(flags uint16, x, y int16, w, h uint16,
                              sibling xgb.Id, stackMode byte) {
    // This will change with other frames
    if DoW & flags > 0 {
        w -= 0
    }
    if DoH & flags > 0 {
        h -= 0
    }

    f.configureFrame(flags, x, y, w, h, sibling, stackMode)
}

func (f *frameNada) configureFrame(flags uint16, x, y int16, w, h uint16,
                                   sibling xgb.Id, stackMode byte) {
    // This will change with other frames
    // See line 185 in frame/__init__.py for Pyndow
    if DoW & flags > 0 {
        w += 0
    }
    if DoH & flags > 0 {
        h += 0
    }

    f.parent.client.Win().moveresize(flags, 0, 0, w, h)
    f.parent.window.configure(flags, x, y, w, h, sibling, stackMode)
}

func (f *frameNada) Geom() xrect.Rect {
    return f.parent.window.geom
}

func (f *frameNada) Parent() *frameParent {
    return f.parent
}

