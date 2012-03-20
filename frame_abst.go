package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xrect"
)

type frameAbst struct {
    parent *frameParent
    clientPos clientPos
    moving moveState
    resizing resizeState
}

func newFrameAbst(c Client, cp clientPos) (*frameAbst, error) {
    geom, err := c.Win().geometry()
    if err != nil {
        return nil, err
    }

    f := &frameAbst {
        parent: newParent(c),
        clientPos: cp,
        moving: moveState{},
        resizing: resizeState{},
    }

    f.Moveresize(DoW | DoH, 0, 0, geom.Width(), geom.Height(), false)

    return f, nil
}

func (f *frameAbst) Destroy() {
    f.parent.window.destroy()
}

func (f *frameAbst) Map() {
    f.parent.window.map_()
}

func (f *frameAbst) Unmap() {
    f.parent.window.unmap()
}

func (f *frameAbst) Client() Client {
    return f.parent.client
}

func (f *frameAbst) Moveresize(flags uint16, x, y int16, w, h uint16,
                               ignoreHints bool) {
    f.Configure(flags, x, y, w, h, xgb.Id(0), 0, ignoreHints)
}

func (f *frameAbst) Configure(flags uint16, x, y int16, w, h uint16,
                              sibling xgb.Id, stackMode byte,
                              ignoreHints bool) {
    // This will change with other frames
    if DoW & flags > 0 {
        w += f.clientPos.w
    }
    if DoH & flags > 0 {
        h += f.clientPos.h
    }

    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameAbst) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool) {
    cw, ch := fw, fh
    framex, framey, _, _ := xrect.RectPieces(f.Geom())
    _, _, clientw, clienth := xrect.RectPieces(f.Client().Geom())

    if DoX & flags > 0 {
        framex = fx
    }
    if DoY & flags > 0 {
        framey = fy
    }
    if DoW & flags > 0 {
        cw -= f.clientPos.w
        if !ignoreHints {
            cw = f.Client().ValidateWidth(cw)
            fw = cw + f.clientPos.w
        }
        clientw = cw
    }
    if DoH & flags > 0 {
        ch -= f.clientPos.h
        if !ignoreHints {
            ch = f.Client().ValidateHeight(ch)
            fh = ch + f.clientPos.h
        }
        clienth = ch
    }

    configNotify := xevent.NewConfigureNotify(f.Client().Id(), f.Client().Id(),
                                              0, framex, framey,
                                              clientw, clienth, 0, false)
    X.Conn().SendEvent(false, f.Client().Id(), xgb.EventMaskStructureNotify,
                       configNotify.Bytes())

    f.Client().Win().moveresize(flags | DoX | DoY,
                                f.clientPos.x, f.clientPos.y, cw, ch)
    f.Parent().Win().configure(flags, fx, fy, fw, fh, sibling, stackMode)
}

func (f *frameAbst) Geom() xrect.Rect {
    return f.parent.window.geom
}

func (f *frameAbst) Moving() bool {
    return f.moving.moving
}

func (f *frameAbst) Parent() *frameParent {
    return f.parent
}

func (f *frameAbst) ParentId() xgb.Id {
    return f.parent.window.id
}

func (f *frameAbst) Resizing() bool {
    return f.resizing.resizing
}

// ValidateHeight validates a height of a *frame*, which is equivalent
// to validating the height of a client.
func (f *frameAbst) ValidateHeight(height uint16) uint16 {
    return f.Client().ValidateHeight(height - f.clientPos.h) + f.clientPos.h
}

// ValidateWidth validates a width of a *frame*, which is equivalent
// to validating the width of a client.
func (f *frameAbst) ValidateWidth(width uint16) uint16 {
    return f.Client().ValidateWidth(width - f.clientPos.w) + f.clientPos.w
}

