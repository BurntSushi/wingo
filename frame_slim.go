package main

type frameSlim struct {
    *frameAbst
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

