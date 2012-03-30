package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameSlim struct {
    *abstFrame
}

func newFrameSlim(p *frameParent, c *client) *frameSlim {
    cp := clientOffset{
        x: THEME.slim.borderSize,
        y: THEME.slim.borderSize,
        w: THEME.slim.borderSize * 2,
        h: THEME.slim.borderSize * 2,
    }
    return &frameSlim{newFrameAbst(p, c, cp)}
}

func (f *frameSlim) Current() bool {
    return f.Client().Frame() == f
}

func (f *frameSlim) Off() {
}

func (f *frameSlim) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.State() == StateActive {
        f.Active()
    } else {
        f.Inactive()
    }
}

func (f *frameSlim) Active() {
    f.ParentWin().change(xgb.CWBackPixel, uint32(THEME.slim.aBorderColor))
    f.ParentWin().clear()
}

func (f *frameSlim) Inactive() {
    f.ParentWin().change(xgb.CWBackPixel, uint32(THEME.slim.iBorderColor))
    f.ParentWin().clear()
}

func (f *frameSlim) Maximize() {
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()
}

func (f *frameSlim) Unmaximize() {
    f.clientOffset.x = f.Left()
    f.clientOffset.y = f.Top()
    f.clientOffset.w = f.Left() + f.Right()
    f.clientOffset.h = f.Top() + f.Bottom()
}

func (f *frameSlim) Top() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.slim.borderSize
}

func (f *frameSlim) Bottom() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.slim.borderSize
}

func (f *frameSlim) Left() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.slim.borderSize
}

func (f *frameSlim) Right() int {
    if f.Client().maximized {
        return 0
    }
    return THEME.slim.borderSize
}

func (f *frameSlim) ConfigureClient(flags, x, y, w, h int,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints, true)
}

func (f *frameSlim) ConfigureFrame(flags, fx, fy, fw, fh int,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool, sendNotify bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints,
                     sendNotify)
}

