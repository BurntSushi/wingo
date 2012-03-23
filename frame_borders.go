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

func newFrameBorders(p *frameParent, c Client) *frameBorders {
    cp := clientOffset{x: 20, y: 20, w: 40, h: 40}
    f := &frameBorders{abstFrame: newFrameAbst(p, c, cp)}

    f.topSide = f.newTopSide()
    f.topSide.w.moveresize(DoX | DoY | DoH,
                           f.topSide.xoff, f.topSide.yoff,
                           0, f.topSide.hoff)

    return f
}

func (f *frameBorders) Off() {
    f.topSide.w.unmap()
}

func (f *frameBorders) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }

    f.topSide.w.map_()
}

func (f *frameBorders) StateActive() {
    f.state = StateActive

    f.topSide.StateActive()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
    f.ParentWin().clear()
}

func (f *frameBorders) StateInactive() {
    f.state = StateInactive

    f.topSide.StateInactive()

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

    f.topSide.w.moveresize(DoW, 0, 0, fg.Width() - f.topSide.woff, 0)
}

