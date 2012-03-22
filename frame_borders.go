package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameBorders struct {
    *abstFrame
}

func newFrameBorders(p *frameParent, c Client) *frameBorders {
    cp := clientPos{x: 20, y: 20, w: 40, h: 40}
    return &frameBorders{newFrameAbst(p, c, cp)}
}

func (f *frameBorders) Off() {
}

func (f *frameBorders) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }
}

func (f *frameBorders) StateActive() {
    f.state = StateActive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xbb0000))
    f.ParentWin().clear()
}

func (f *frameBorders) StateInactive() {
    f.state = StateInactive

    f.ParentWin().change(xgb.CWBackPixel, uint32(0xdfdcdf))
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
}

