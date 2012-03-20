package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type normalClient struct {
    *abstClient
}

func newNormalClient(id xgb.Id) (*normalClient, error) {
    absCli, err := newAbstractClient(id)
    if err != nil {
        return nil, err
    }

    return &normalClient{
        abstClient: absCli,
    }, nil
}

