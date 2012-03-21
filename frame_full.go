package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameFull struct {
    *abstFrame
}

func newFrameFull(p *frameParent, c Client) *frameFull {
    cp := clientPos{x: 20, y: 20, w: 40, h: 40}
    return &frameFull{newFrameAbst(p, c, cp)}
}

func (f *frameFull) Off() {
}

func (f *frameFull) On() {
    f.Reset()

    mask := uint32(xgb.CWBackPixel)
    vals := []uint32{0x3366ff}
    X.Conn().ChangeWindowAttributes(f.ParentId(), mask, vals)
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

