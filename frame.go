package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frame interface {
    destroy()
    map_()
    moveresize(flags uint16, x, y int16, w, h uint16)
    parentWin() *frameParent
    unmap()
}

type frameParent struct {
    window *window
    child *window
}

type frameNada struct {
    parent *frameParent
}

func newParent(child *window, x, y int16) *frameParent {
    mask := uint32(xgb.CWEventMask)
    val := []uint32{xgb.EventMaskSubstructureRedirect |
                    xgb.EventMaskButtonPress |
                    xgb.EventMaskButtonRelease}
    parent := createWindow(X.RootWin(), mask, val)
    p := &frameParent{
        window: parent,
        child: child,
    }

    X.Conn().ReparentWindow(child.id, parent.id, x, y)

    return p
}

func newFrameNada(child *window) (*frameNada, error) {
    geom, err := child.geometry()
    if err != nil {
        return nil, err
    }

    f := &frameNada {
        parent: newParent(child, 0, 0),
    }

    f.moveresize(DoW | DoH, 0, 0, geom.Width(), geom.Height())

    return f, nil
}

func (f *frameNada) destroy() {
    // Unparent before destroying the parent window so we don't kill
    // the client prematurely.
    X.Conn().ReparentWindow(f.parent.child.id, X.RootWin(),
                            f.parent.window.geom.X(),
                            f.parent.window.geom.Y())
    f.parent.window.destroy()
}

func (f *frameNada) map_() {
    f.parent.window.map_()
}

func (f *frameNada) unmap() {
    f.parent.window.unmap()
}

func (f *frameNada) moveresize(flags uint16, x, y int16, w, h uint16) {
    // This will change with other frames
    if DoW & flags > 0 {
        w -= 0
    }
    if DoH & flags > 0 {
        h -= 0
    }

    f.moveresizeFrame(flags, x, y, w, h)
}

func (f *frameNada) moveresizeFrame(flags uint16, x, y int16, w, h uint16) {
    // This will change with other frames
    // See line 185 in frame/__init__.py for Pyndow
    if DoW & flags > 0 {
        w += 0
    }
    if DoH & flags > 0 {
        h += 0
    }

    f.parent.child.moveresize(flags, 0, 0, w, h)
    f.parent.window.moveresize(flags, x, y, w, h)
}

func (f *frameNada) parentWin() *frameParent {
    return f.parent
}

