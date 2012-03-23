package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameNada struct {
    *abstFrame
}

func newFrameNada(p *frameParent, c Client) *frameNada {
    return &frameNada{newFrameAbst(p, c, clientOffset{})}
}

func (f *frameNada) Off() {
}

func (f *frameNada) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    // Although, this probably isn't necessary for the Nada frame...
    if f.state == StateActive {
        f.StateActive()
    } else {
        f.StateInactive()
    }
}

func (f *frameNada) StateActive() {
    f.state = StateActive
}

func (f *frameNada) StateInactive() {
    f.state = StateInactive
}

func (f *frameNada) Top() int16 {
    return 0
}

func (f *frameNada) Bottom() int16 {
    return 0
}

func (f *frameNada) Left() int16 {
    return 0
}

func (f *frameNada) Right() int16 {
    return 0
}

func (f *frameNada) ConfigureClient(flags uint16, x, y int16, w, h uint16,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

func (f *frameNada) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints)
}

