package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xrect"
)

type Client interface {
    Alive() bool
    Close()
    Focus()
    Frame() Frame
    FrameNada()
    FrameSlim()
    FrameBorders()
    FrameFull()
    Geom() xrect.Rect
    GravitizeX(x int16) int16
    GravitizeY(y int16) int16
    Id() xgb.Id
    Layer() int
    Map()
    Mapped() bool
    String() string
    ValidateHeight(height uint16) uint16
    ValidateWidth(width uint16) uint16
    Win() *window
}

func clientMapRequest(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
    X.Grab()
    defer X.Ungrab()

    client, err := newNormalClient(ev.Window)
    if err != nil {
        logWarning.Printf("Could not manage window %X because: %v\n",
                          ev.Window, err)
        return
    }

    client.manage()
}

