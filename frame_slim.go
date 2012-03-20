package main

type frameSlim struct {
    *abstFrame
}

func newFrameSlim(c Client) (*frameSlim, error) {
    cp := clientPos{
        x: 20, y: 20, w: 40, h: 40}
    abst, err := newFrameAbst(c, cp)
    if err != nil {
        return nil, err
    }

    return &frameSlim{abst}, nil
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

