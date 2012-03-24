package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameFull struct {
    *abstFrame
}

func newFrameFull(p *frameParent, c *client) *frameFull {
    cp := clientOffset{x: 20, y: 20, w: 40, h: 40}
    return &frameFull{newFrameAbst(p, c, cp)}
}

func (f *frameFull) Off() {
}

func (f *frameFull) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }
}

func (f *frameFull) StateActive() {
    f.state = StateActive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0x3366ff))
    f.ParentWin().clear()
}

func (f *frameFull) StateInactive() {
    f.state = StateInactive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xdfdcdf))
    f.ParentWin().clear()
}

func (f *frameFull) Top() int16 {
    return 20
}

func (f *frameFull) Bottom() int16 {
    return 20
}

func (f *frameFull) Left() int16 {
    return 20
}

func (f *frameFull) Right() int16 {
    return 20
}

func (f *frameFull) ConfigureClient(flags uint16, x, y int16, w, h uint16,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameFull) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints)
}

