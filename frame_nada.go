package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameNada struct {
    *abstFrame
}

func newFrameNada(p *frameParent, c *client) *frameNada {
    return &frameNada{newFrameAbst(p, c, clientOffset{})}
}

func (f *frameNada) Off() {
}

func (f *frameNada) On() {
    FrameReset(f)

    // Make sure the current state is properly shown
    // Although, this probably isn't necessary for the Nada frame...
    if f.State() == StateActive {
        f.Active()
    } else {
        f.Inactive()
    }
}

func (f *frameNada) Active() {
}

func (f *frameNada) Inactive() {
}

func (f *frameNada) Top() int {
    return 0
}

func (f *frameNada) Bottom() int {
    return 0
}

func (f *frameNada) Left() int {
    return 0
}

func (f *frameNada) Right() int {
    return 0
}

func (f *frameNada) ConfigureClient(flags, x, y, w, h int,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    x, y, w, h = f.configureClient(flags, x, y, w, h)
    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints, true)
}

func (f *frameNada) ConfigureFrame(flags, fx, fy, fw, fh int,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool, sendNotify bool) {
    f.configureFrame(flags, fx, fy, fw, fh, sibling, stackMode, ignoreHints,
                     sendNotify)
}

