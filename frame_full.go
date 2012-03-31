package main

import "exp/norm"

import "code.google.com/p/jamslam-x-go-binding/xgb"

import "github.com/BurntSushi/xgbutil/xgraphics"

type frameFull struct {
    *abstFrame

    // pieces
    titleBar, titleText, icon framePiece
    buttonMinimize, buttonMaximize, buttonClose framePiece
    topSide, bottomSide, leftSide, rightSide framePiece
    topLeft, topRight, bottomLeft, bottomRight framePiece
    titleBottom framePiece
}

func newFrameFull(p *frameParent, c *client) *frameFull {
    cp := clientOffset{}
    f := &frameFull{abstFrame: newFrameAbst(p, c, cp)}
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()

    f.titleBar = f.newTitleBar()
    f.titleText = f.newTitleText()
    f.buttonClose = f.newButtonClose()
    f.buttonMaximize = f.newButtonMaximize()
    f.buttonMinimize = f.newButtonMinimize()
    f.icon = f.newIcon()

    if THEME.full.borderSize > 0 {
        f.topSide = f.newTopSide()
        f.bottomSide = f.newBottomSide()
        f.leftSide = f.newLeftSide()
        f.rightSide = f.newRightSide()
        f.titleBottom = f.newTitleBottom()

        f.topLeft = f.newTopLeft()
        f.topRight = f.newTopRight()
        f.bottomLeft = f.newBottomLeft()
        f.bottomRight = f.newBottomRight()
    }

    f.updateTitle()
    f.updateIcon()

    return f
}

func (f *frameFull) Current() bool {
    return f.Client().Frame() == f
}

func (f *frameFull) Destroy() {
    if THEME.full.borderSize > 0 {
        f.topSide.destroy()
        f.bottomSide.destroy()
        f.leftSide.destroy()
        f.rightSide.destroy()
        f.titleBottom.destroy()

        f.topLeft.destroy()
        f.topRight.destroy()
        f.bottomLeft.destroy()
        f.bottomRight.destroy()
    }

    f.titleBar.destroy()
    f.titleText.destroy()
    f.icon.destroy()
    f.buttonClose.destroy()
    f.buttonMaximize.destroy()
    f.buttonMinimize.destroy()

    f.abstFrame.Destroy()
}

func (f *frameFull) Off() {
    if THEME.full.borderSize > 0 {
        f.topSide.win.unmap()
        f.bottomSide.win.unmap()
        f.leftSide.win.unmap()
        f.rightSide.win.unmap()
        f.titleBottom.win.unmap()

        f.topLeft.win.unmap()
        f.topRight.win.unmap()
        f.bottomLeft.win.unmap()
        f.bottomRight.win.unmap()
    }

    f.titleBar.win.unmap()
    f.titleText.win.unmap()
    f.icon.win.unmap()
    f.buttonClose.win.unmap()
    f.buttonMaximize.win.unmap()
    f.buttonMinimize.win.unmap()
}

func (f *frameFull) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.State() == StateActive {
        f.Active()
    } else {
        f.Inactive()
    }

    if THEME.full.borderSize > 0 {
        f.titleBottom.win.map_()

        if !f.Client().maximized {
            f.topSide.win.map_()
            f.bottomSide.win.map_()
            f.leftSide.win.map_()
            f.rightSide.win.map_()

            f.topLeft.win.map_()
            f.topRight.win.map_()
            f.bottomLeft.win.map_()
            f.bottomRight.win.map_()
        }
    }

    f.titleBar.win.map_()
    f.titleText.win.map_()
    f.icon.win.map_()
    f.buttonClose.win.map_()
    f.buttonMaximize.win.map_()
    f.buttonMinimize.win.map_()
}

func (f *frameFull) Active() {
    if THEME.full.borderSize > 0 {
        f.topSide.active()
        f.bottomSide.active()
        f.leftSide.active()
        f.rightSide.active()
        f.titleBottom.active()

        f.topLeft.active()
        f.topRight.active()
        f.bottomLeft.active()
        f.bottomRight.active()
    }

    f.titleBar.active()
    f.titleText.active()
    f.icon.active()
    f.buttonClose.active()
    f.buttonMaximize.active()
    f.buttonMinimize.active()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xffffff))
    f.ParentWin().clear()
}

func (f *frameFull) Inactive() {
    if THEME.full.borderSize > 0 {
        f.topSide.inactive()
        f.bottomSide.inactive()
        f.leftSide.inactive()
        f.rightSide.inactive()
        f.titleBottom.inactive()

        f.topLeft.inactive()
        f.topRight.inactive()
        f.bottomLeft.inactive()
        f.bottomRight.inactive()
    }

    f.titleBar.inactive()
    f.titleText.inactive()
    f.icon.inactive()
    f.buttonClose.inactive()
    f.buttonMaximize.inactive()
    f.buttonMinimize.inactive()

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xffffff))
    f.ParentWin().clear()
}

func (f *frameFull) Maximize() {
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()

    f.buttonClose.win.moveresize(DoY, 0, 0, 0, 0)
    f.buttonMaximize.win.moveresize(DoY, 0, 0, 0, 0)
    f.buttonMinimize.win.moveresize(DoY, 0, 0, 0, 0)
    f.titleBar.win.moveresize(DoX | DoY, 0, 0, 0, 0)
    f.titleText.win.moveresize(DoX | DoY, THEME.full.titleSize, 0, 0, 0)
    f.icon.win.moveresize(DoX | DoY, 0, 0, 0, 0)
    f.titleBottom.win.moveresize(DoX | DoY, 0, THEME.full.titleSize, 0, 0)

    if THEME.full.borderSize > 0 && f.Current() {
        f.topSide.win.unmap()
        f.bottomSide.win.unmap()
        f.leftSide.win.unmap()
        f.rightSide.win.unmap()

        f.topLeft.win.unmap()
        f.topRight.win.unmap()
        f.bottomLeft.win.unmap()
        f.bottomRight.win.unmap()

        FrameReset(f)
    }
}

func (f *frameFull) Unmaximize() {
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()

    f.buttonClose.win.moveresize(DoY, 0, THEME.full.borderSize, 0, 0)
    f.buttonMaximize.win.moveresize(DoY, 0, THEME.full.borderSize, 0, 0)
    f.buttonMinimize.win.moveresize(DoY, 0, THEME.full.borderSize, 0, 0)
    f.titleBar.win.moveresize(DoX | DoY, THEME.full.borderSize,
                              THEME.full.borderSize, 0, 0)
    f.titleText.win.moveresize(DoX | DoY,
                               THEME.full.borderSize + THEME.full.titleSize,
                               THEME.full.borderSize, 0, 0)
    f.icon.win.moveresize(DoX | DoY, THEME.full.borderSize,
                          THEME.full.borderSize, 0, 0)
    f.titleBottom.win.moveresize(DoX | DoY, THEME.full.borderSize,
                                 THEME.full.borderSize + THEME.full.titleSize,
                                 0, 0)

    if THEME.full.borderSize > 0 && f.Current() {
        f.topSide.win.map_()
        f.bottomSide.win.map_()
        f.leftSide.win.map_()
        f.rightSide.win.map_()

        f.topLeft.win.map_()
        f.topRight.win.map_()
        f.bottomLeft.win.map_()
        f.bottomRight.win.map_()

        FrameReset(f)
    }

}

func (f *frameFull) Top() int {
    if f.Client().maximized {
        return THEME.full.borderSize + THEME.full.titleSize
    }
    return (THEME.full.borderSize * 2) + THEME.full.titleSize
}

func (f *frameFull) Bottom() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.full.borderSize
}

func (f *frameFull) Left() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.full.borderSize
}

func (f *frameFull) Right() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.full.borderSize
}

func (f *frameFull) ConfigureClient(flags, x, y, w, h int,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints,
                     true)
}

func (f *frameFull) ConfigureFrame(flags, fx, fy, fw, fh int,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool, sendNotify bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints,
                     sendNotify)
    fg := f.Geom()

    if THEME.full.borderSize > 0 {
        f.topSide.win.moveresize(
            DoW, 0, 0, fg.Width() - f.topLeft.w() - f.topRight.w(), 0)
        f.bottomSide.win.moveresize(
            DoY | DoW, 0, fg.Height() - f.bottomSide.h(), f.topSide.w(), 0)
        f.leftSide.win.moveresize(
            DoH, 0, 0, 0, fg.Height() - f.topLeft.h() - f.bottomLeft.h())
        f.rightSide.win.moveresize(
            DoX | DoH, fg.Width() - f.rightSide.w(), 0, 0, f.leftSide.h())
        f.titleBottom.win.moveresize(DoW, 0, 0,
                                     fg.Width() - f.Left() - f.Right(), 0)

        f.topRight.win.moveresize(DoX, f.topLeft.w() + f.topSide.w(), 0, 0, 0)
        f.bottomLeft.win.moveresize(DoY, 0, f.bottomSide.y(), 0, 0)
        f.bottomRight.win.moveresize(
            DoX | DoY,
            f.bottomLeft.w() + f.bottomSide.w(), f.bottomSide.y(),
            0, 0)
    }

    f.titleBar.win.moveresize(
        DoW, 0, 0, fg.Width() - f.Left() - f.Right(), 0)
    f.buttonClose.win.moveresize(
        DoX, fg.Width() - f.Right() - f.buttonClose.w(), 0, 0, 0)
    f.buttonMaximize.win.moveresize(
        DoX, f.buttonClose.x() - f.buttonMinimize.w(), 0, 0, 0)
    f.buttonMinimize.win.moveresize(
        DoX, f.buttonMaximize.x() - f.buttonMinimize.w(), 0, 0, 0)
}

func (f *frameFull) updateIcon() {
    imgA := renderBorder(0, 0, THEME.full.aTitleColor,
                         THEME.full.titleSize, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)
    imgI := renderBorder(0, 0, THEME.full.iTitleColor,
                         THEME.full.titleSize, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)

    img, msk := f.Client().iconImage(THEME.full.titleSize - 4,
                                     THEME.full.titleSize - 4)
    xgraphics.Blend(imgA, img, msk, 100, 2, 2)
    xgraphics.Blend(imgI, img, msk, 100, 2, 2)

    if f.icon.imgActive > 0 {
        xgraphics.FreePixmap(X, f.icon.imgActive)
    }
    if f.icon.imgInactive > 0 {
        xgraphics.FreePixmap(X, f.icon.imgInactive)
    }

    f.icon.imgActive = xgraphics.CreatePixmap(X, imgA)
    f.icon.imgInactive = xgraphics.CreatePixmap(X, imgI)

    if f.State() == StateActive {
        f.icon.active()
    } else {
        f.icon.inactive()
    }
}

func (f *frameFull) updateTitle() {
    title := f.Client().Name()
    font := THEME.full.font
    fontSize := THEME.full.fontSize
    aFontColor := ColorFromInt(THEME.full.aFontColor)
    iFontColor := ColorFromInt(THEME.full.iFontColor)

    // Try to normalize the window name so freetype can handle it.
    title = norm.NFD.String(title)

    ew, eh, err := xgraphics.TextExtents(font, fontSize, title)
    if err != nil {
        logWarning.Printf("Could not get text extents for name '%s' on " +
                          "window %s because: %v",
                          title, f.Client(), err)
        logWarning.Printf("Resorting to default with of 300.")
        ew = 300
    }

    imgA := renderBorder(0, 0, THEME.full.aTitleColor,
                         ew, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)
    imgI := renderBorder(0, 0, THEME.full.iTitleColor,
                         ew, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)

    y := (THEME.full.titleSize - eh) / 2 - 1

    err = xgraphics.DrawText(imgA, 0, y, aFontColor, fontSize, font, title)
    if err != nil {
        logWarning.Printf("Could not draw window title for window %s " +
                          "because: %v", f.Client(), err)
    }

    err = xgraphics.DrawText(imgI, 0, y, iFontColor, fontSize, font, title)
    if err != nil {
        logWarning.Printf("Could not draw window title for window %s " +
                          "because: %v", f.Client(), err)
    }

    if f.titleText.imgActive > 0 {
        xgraphics.FreePixmap(X, f.titleText.imgActive)
    }
    if f.titleText.imgInactive > 0 {
        xgraphics.FreePixmap(X, f.titleText.imgInactive)
    }

    f.titleText.imgActive = xgraphics.CreatePixmap(X, imgA)
    f.titleText.imgInactive = xgraphics.CreatePixmap(X, imgI)

    f.titleText.win.moveresize(DoW, 0, 0, ew, 0)
    if f.State() == StateActive {
        f.titleText.active()
    } else {
        f.titleText.inactive()
    }
}

