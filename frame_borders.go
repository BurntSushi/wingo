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
    f.Reset()

    mask := uint32(xgb.CWBackPixel)
    vals := []uint32{0xbb0000}
    X.Conn().ChangeWindowAttributes(f.ParentId(), mask, vals)
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

