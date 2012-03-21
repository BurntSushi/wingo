package main

type frameNada struct {
    *abstFrame
}

func newFrameNada(p *frameParent, c Client) *frameNada {
    return &frameNada{newFrameAbst(p, c, clientPos{})}
}

func (f *frameNada) Off() {
}

func (f *frameNada) On() {
    f.Reset()
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

