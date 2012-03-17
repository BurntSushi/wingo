package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

// import ( 
    // "github.com/BurntSushi/xgbutil" 
    // "github.com/BurntSushi/xgbutil/mousebind" 
    // "github.com/BurntSushi/xgbutil/xevent" 
// ) 

type frame interface {
    client() client
    configure(flags uint16, x, y int16, w, h uint16,
              sibling xgb.Id, stackMode byte)
    destroy()
    map_()
    moveresize(flags uint16, x, y int16, w, h uint16)
    parentWin() *frameParent
    unmap()

    // These are temporary. I think they will move to 'layout'
    moveBegin(rx, ry, ex, ey int16)
    moveStep(rx, ry, ex, ey int16)
    moveEnd(rx, ry, ex, ey int16)
    resizeBegin(rx, ry, ex, ey int16)
    resizeStep(rx, ry, ex, ey int16)
    resizeEnd(rx, ry, ex, ey int16)
}

type frameParent struct {
    window *window
    client client
}

type frameNada struct {
    parent *frameParent
    moving moveState
}

type moveState struct {
    lastRootX int16
    lastRootY int16
}

func newParent(c client, x, y int16) *frameParent {
    mask := uint32(xgb.CWEventMask)
    val := []uint32{xgb.EventMaskSubstructureRedirect |
                    xgb.EventMaskButtonPress |
                    xgb.EventMaskButtonRelease}
    parent := createWindow(X.RootWin(), mask, val)
    p := &frameParent{
        window: parent,
        client: c,
    }

    X.Conn().ReparentWindow(c.id(), parent.id, x, y)

    return p
}

func newFrameNada(c client) (*frameNada, error) {
    geom, err := c.win().geometry()
    if err != nil {
        return nil, err
    }

    f := &frameNada {
        parent: newParent(c, 0, 0),
        moving: moveState{
            lastRootX: 0,
            lastRootY: 0,
        },
    }

    f.moveresize(DoW | DoH, 0, 0, geom.Width(), geom.Height())

    return f, nil
}

func (f *frameNada) destroy() {
    // Unparent before destroying the parent window so we don't kill
    // the client prematurely.
    // X.Conn().ReparentWindow(f.parent.client.id(), X.RootWin(), 
                            // f.parent.window.geom.X(), 
                            // f.parent.window.geom.Y()) 
    f.parent.window.destroy()
}

func (f *frameNada) map_() {
    f.parent.window.map_()
}

func (f *frameNada) unmap() {
    f.parent.window.unmap()
}

func (f *frameNada) client() client {
    return f.parent.client
}

func (f *frameNada) moveresize(flags uint16, x, y int16, w, h uint16) {
    f.configure(flags, x, y, w, h, xgb.Id(0), 0)
}

func (f *frameNada) configure(flags uint16, x, y int16, w, h uint16,
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

    f.parent.client.win().moveresize(flags, 0, 0, w, h)
    f.parent.window.configure(flags, x, y, w, h, sibling, stackMode)
}

func (f *frameNada) parentWin() *frameParent {
    return f.parent
}

