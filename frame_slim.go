package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameSlim struct {
    *abstFrame
}

func newFrameSlim(p *frameParent, c Client) *frameSlim {
    cp := clientPos{x: 20, y: 20, w: 40, h: 40}
    return &frameSlim{newFrameAbst(p, c, cp)}
}

func (f *frameSlim) Off() {
}

func (f *frameSlim) On() {
    f.Reset()

    mask := uint32(xgb.CWBackPixel)
    vals := []uint32{0xff7f00}
    X.Conn().ChangeWindowAttributes(f.ParentId(), mask, vals)
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

