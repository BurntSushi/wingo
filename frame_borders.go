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
    cp := clientOffset{}
    f := &frameBorders{abstFrame: newFrameAbst(p, c, cp)}
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()

    f.topSide = f.newTopSide()
    f.topLeft = f.newTopLeft()
    f.topRight = f.newTopRight()

    f.leftSide = f.newLeftSide()
    f.leftTop = f.newLeftTop()
    f.leftBottom = f.newLeftBottom()

    f.bottomSide = f.newBottomSide()
    f.bottomLeft = f.newBottomLeft()
    f.bottomRight = f.newBottomRight()

    f.rightSide = f.newRightSide()
    f.rightTop = f.newRightTop()
    f.rightBottom = f.newRightBottom()

    f.topSide.initialGeom(DoX | DoY | DoH)
    f.topLeft.initialGeom(DoX | DoY | DoW | DoH)
    f.topRight.initialGeom(DoY | DoW | DoH)

    f.leftSide.initialGeom(DoX | DoY | DoW)
    f.leftTop.initialGeom(DoX | DoY | DoW | DoH)
    f.leftBottom.initialGeom(DoX | DoW | DoH)

    f.bottomSide.initialGeom(DoX | DoH)
    f.bottomLeft.initialGeom(DoX | DoW | DoH)
    f.bottomRight.initialGeom(DoW | DoH)

    f.rightSide.initialGeom(DoY | DoW)
    f.rightTop.initialGeom(DoY | DoW | DoH)
    f.rightBottom.initialGeom(DoW | DoH)

    return f
}

func (f *frameBorders) Destroy() {
    f.topSide.destroy()
    f.topLeft.destroy()
    f.topRight.destroy()

    f.leftSide.destroy()
    f.leftTop.destroy()
    f.leftBottom.destroy()

    f.bottomSide.destroy()
    f.bottomLeft.destroy()
    f.bottomRight.destroy()

    f.rightSide.destroy()
    f.rightTop.destroy()
    f.rightBottom.destroy()

    f.abstFrame.Destroy()
}

func (f *frameBorders) Off() {
    f.topSide.win.unmap()
    f.topLeft.win.unmap()
    f.topRight.win.unmap()

    f.leftSide.win.unmap()
    f.leftTop.win.unmap()
    f.leftBottom.win.unmap()

    f.bottomSide.win.unmap()
    f.bottomLeft.win.unmap()
    f.bottomRight.win.unmap()

    f.rightSide.win.unmap()
    f.rightTop.win.unmap()
    f.rightBottom.win.unmap()
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

    f.leftSide.win.map_()
    f.leftTop.win.map_()
    f.leftBottom.win.map_()

    f.bottomSide.win.map_()
    f.bottomLeft.win.map_()
    f.bottomRight.win.map_()

    f.rightSide.win.map_()
    f.rightTop.win.map_()
    f.rightBottom.win.map_()
}

func (f *frameBorders) StateActive() {
    f.state = StateActive

    f.topSide.active()
    f.topLeft.active()
    f.topRight.active()

    f.leftSide.active()
    f.leftTop.active()
    f.leftBottom.active()

    f.bottomSide.active()
    f.bottomLeft.active()
    f.bottomRight.active()

    f.rightSide.active()
    f.rightTop.active()
    f.rightBottom.active()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
    f.ParentWin().clear()
}

func (f *frameBorders) StateInactive() {
    f.state = StateInactive

    f.topSide.inactive()
    f.topLeft.inactive()
    f.topRight.inactive()

    f.leftSide.inactive()
    f.leftTop.inactive()
    f.leftBottom.inactive()

    f.bottomSide.inactive()
    f.bottomLeft.inactive()
    f.bottomRight.inactive()

    f.rightSide.inactive()
    f.rightTop.inactive()
    f.rightBottom.inactive()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
    f.ParentWin().clear()
}

func (f *frameBorders) Top() int {
    return THEME.borders.borderSize
}

func (f *frameBorders) Bottom() int {
    return THEME.borders.borderSize
}

func (f *frameBorders) Left() int {
    return THEME.borders.borderSize
}

func (f *frameBorders) Right() int {
    return THEME.borders.borderSize
}

func (f *frameBorders) ConfigureClient(flags, x, y, w, h int,
                                       sibling xgb.Id, stackMode byte,
                                       ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameBorders) ConfigureFrame(flags, fx, fy, fw, fh int,
                                      sibling xgb.Id, stackMode byte,
                                      ignoreHints bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints)
    fg := f.Geom()

    f.topSide.win.moveresize(DoW, 0, 0, fg.Width() - f.topSide.woff, 0)
    f.topRight.win.moveresize(DoX, f.topRight.xoff + f.topSide.w(),
                              0, 0, 0)

    f.leftSide.win.moveresize(DoH, 0, 0, 0, fg.Height() - f.leftSide.hoff)
    f.leftBottom.win.moveresize(DoY, 0,
                                f.leftBottom.yoff + f.leftSide.h(),
                                0, 0)

    f.bottomSide.win.moveresize(DoY | DoW, 0,
                                fg.Height() - f.bottomSide.yoff,
                                fg.Width() - f.bottomSide.woff, 0)
    f.bottomLeft.win.moveresize(DoY, 0, f.bottomSide.y(), 0, 0)
    f.bottomRight.win.moveresize(DoX | DoY,
                                 f.bottomRight.xoff + f.bottomSide.w(),
                                 f.bottomSide.y(), 0, 0)

    f.rightSide.win.moveresize(DoX | DoH,
                               fg.Width() - f.rightSide.xoff, 0, 0,
                               f.leftSide.h())
    f.rightTop.win.moveresize(DoX, f.rightSide.x(), 0, 0, 0)
    f.rightBottom.win.moveresize(DoX | DoY, f.rightSide.x(),
                                 f.rightTop.h() + f.rightSide.h(),
                                 0, 0)
}

