package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameSlim struct {
    *abstFrame
}

func newFrameSlim(p *frameParent, c *client) *frameSlim {
    cp := clientOffset{x: 20, y: 20, w: 40, h: 40}
    return &frameSlim{newFrameAbst(p, c, cp)}
}

func (f *frameSlim) Off() {
}

func (f *frameSlim) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }
}

func (f *frameSlim) StateActive() {
    f.state = StateActive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xff7f00))
    f.ParentWin().clear()
}

func (f *frameSlim) StateInactive() {
    f.state = StateInactive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xdfdcdf))
    f.ParentWin().clear()
}

func (f *frameSlim) Top() int16 {
    return 20
}

func (f *frameSlim) Bottom() int16 {
    return 20
}

func (f *frameSlim) Left() int16 {
    return 20
}

func (f *frameSlim) Right() int16 {
    return 20
}

func (f *frameSlim) ConfigureClient(flags uint16, x, y int16, w, h uint16,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameSlim) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints)
}

