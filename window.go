package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type window struct {
    id xgb.Id
}

func newWindow(id xgb.Id) *window {
    return &window{
        id: id,
    }
}

func (w *window) map_() {
    X.Conn().MapWindow(w.id)
}

func (w *window) focus() {
    X.Conn().SetInputFocus(xgb.InputFocusPointerRoot, w.id, 0)
}

