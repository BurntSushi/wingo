package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameBorders struct {
    *abstFrame

    // pieces
    topSide, topLeft, topRight framePiece
    bottomSide, bottomLeft, bottomRight framePiece
    leftSide, leftTop, leftBottom framePiece
    rightSide, rightTop, rightBottom framePiece
}

func newFrameBorders(p *frameParent, c *client) *frameBorders {
    cp := clientOffset{x: 20, y: 20, w: 40, h: 40}
    f := &frameBorders{abstFrame: newFrameAbst(p, c, cp)}

    f.topSide = f.newTopSide()
    f.topLeft = f.newTopLeft()
    f.topRight = f.newTopRight()

    f.topSide.initialGeom(DoX | DoY | DoH)
    f.topLeft.initialGeom(DoX | DoY | DoW | DoH)
    f.topRight.initialGeom(DoY | DoW | DoH)

    return f
}

func (f *frameBorders) Destroy() {
    f.topSide.destroy()
    f.topLeft.destroy()
    f.topRight.destroy()

    f.abstFrame.Destroy()
}

func (f *frameBorders) Off() {
    f.topSide.win.unmap()
    f.topLeft.win.unmap()
    f.topRight.win.unmap()
}

func (f *frameBorders) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }

    f.topSide.win.map_()
    f.topLeft.win.map_()
    f.topRight.win.map_()
}

func (f *frameBorders) StateActive() {
    f.state = StateActive

    f.topSide.active()
    f.topLeft.active()
    f.topRight.active()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
    f.ParentWin().clear()
}

func (f *frameBorders) StateInactive() {
    f.state = StateInactive

    f.topSide.inactive()
    f.topLeft.inactive()
    f.topRight.inactive()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
    f.ParentWin().clear()
}

func (f *frameBorders) Top() int16 {
    return 20
}

func (f *frameBorders) Bottom() int16 {
    return 20
}

func (f *frameBorders) Left() int16 {
    return 20
}

func (f *frameBorders) Right() int16 {
    return 20
}

func (f *frameBorders) ConfigureClient(flags uint16, x, y int16, w, h uint16,
                                       sibling xgb.Id, stackMode byte,
                                       ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameBorders) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                      sibling xgb.Id, stackMode byte,
                                      ignoreHints bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints)
    fg, _ := f.Geom(), f.Client().Geom()

    f.topSide.win.moveresize(DoW, 0, 0, fg.Width() - f.topSide.woff, 0)
    f.topRight.win.moveresize(DoX, f.topSide.x() + int16(f.topSide.w()),
                              0, 0, 0)
}

