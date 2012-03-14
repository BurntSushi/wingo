package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xwindow"
)

type window struct {
    id xgb.Id
    geom xrect.Rect
}

const (
    DoX = xgb.ConfigWindowX
    DoY = xgb.ConfigWindowY
    DoW = xgb.ConfigWindowWidth
    DoH = xgb.ConfigWindowHeight
)

func newWindow(id xgb.Id) *window {
    return &window{
        id: id,
        geom: xrect.Make(0, 0, 1, 1),
    }
}

func createWindow(parent xgb.Id, mask uint32, vals []uint32) *window {
    wid := X.Conn().NewId()
    scrn := X.Screen()

    X.Conn().CreateWindow(scrn.RootDepth, wid, parent, 0, 0, 1, 1, 0,
                          xgb.WindowClassInputOutput, scrn.RootVisual,
                          mask, vals)

    return newWindow(wid)
}

func (w *window) listen(masks int) {
    xwindow.Listen(X, w.id, masks)
}

func (w *window) map_() {
    X.Conn().MapWindow(w.id)
}

func (w *window) unmap() {
    X.Conn().UnmapWindow(w.id)
}

func (w *window) geometry() (xrect.Rect, error) {
    var err error
    w.geom, err = xwindow.RawGeometry(X, w.id)
    if err != nil {
        return nil, err
    }
    return w.geom, nil
}

func (w *window) kill() {
    X.Conn().KillClient(uint32(w.id))
}

func (w *window) destroy() {
    X.Conn().DestroyWindow(w.id)
    xevent.Detach(X, w.id)
}

func (w *window) focus() {
    X.Conn().SetInputFocus(xgb.InputFocusPointerRoot, w.id, 0)
}

func (win *window) moveresize(flags uint16, x, y int16, w, h uint16) {
    vals := []uint32{}

    if DoX & flags > 0 {
        vals = append(vals, uint32(x))
        win.geom.XSet(x)
    }
    if DoY & flags > 0 {
        vals = append(vals, uint32(y))
        win.geom.YSet(y)
    }
    if DoW & flags > 0 {
        vals = append(vals, uint32(w))
        win.geom.WidthSet(w)
    }
    if DoH & flags > 0 {
        vals = append(vals, uint32(h))
        win.geom.HeightSet(h)
    }

    X.Conn().ConfigureWindow(win.id, flags, vals)
}

