package main

type frameNada struct {
    *abstFrame
}

func newFrameNada(c Client) (*frameNada, error) {
    abst, err := newFrameAbst(c, clientPos{})
    if err != nil {
        return nil, err
    }

    return &frameNada{abst}, nil
}

